<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import yaml from 'js-yaml'
import type { FormInstance, FormRules } from 'element-plus'
import { getNamespaceList, extractNamespaceNames, createJob, updateJobYaml } from '@/api/resource'
import type { JobFormData } from './form-types'
import { createEmptyContainer } from './form-types'
import LabelsAnnotationsForm from './form/LabelsAnnotationsForm.vue'
import ContainerConfigForm from './form/ContainerConfigForm.vue'
import StorageConfigForm from './form/StorageConfigForm.vue'
import HealthCheckForm from './HealthCheckForm.vue'
import SecurityContextForm from './form/SecurityContextForm.vue'
import SchedulingForm from './form/SchedulingForm.vue'

const props = withDefaults(defineProps<{
  isEdit?: boolean
  initialData?: any
  onSubmit?: (yaml: string) => Promise<void>
}>(), {
  isEdit: false,
  initialData: undefined,
  onSubmit: undefined,
})

const emit = defineEmits<{ success: []; cancel: [] }>()

const router = useRouter()
const submitting = ref(false)
const namespaceLoading = ref(false)
const namespaces = ref<string[]>([])
const { t } = useI18n()

const form = reactive<JobFormData>({
  name: '', namespace: 'default',
  labels: [{ key: 'app', value: '' }],
  completions: 1, parallelism: 1, backoffLimit: 6, activeDeadlineSeconds: null,
  ttlSecondsAfterFinished: null, completionMode: 'NonIndexed', restartPolicy: 'Never',
  containers: [createEmptyContainer()], initContainers: [], volumes: [],
  nodeSelector: [], tolerations: [], annotations: [],
  serviceAccountName: '', terminationGracePeriodSeconds: null, imagePullSecrets: [],
  podAffinityRules: [], podAntiAffinityRules: [],
  topologySpreadConstraints: [],
})

const formRef = ref<FormInstance>()
const formRules: FormRules = {
  name: [
    { required: true, message: '请输入名称', trigger: 'blur' },
    { pattern: /^[a-z][a-z0-9-]*[a-z0-9]$/, message: '仅支持小写字母、数字和连字符', trigger: 'blur' },
  ],
  namespace: [{ required: true, message: '请选择命名空间', trigger: 'change' }],
}

async function fetchNamespaces() {
  namespaceLoading.value = true
  try {
    const res: any = await getNamespaceList()
    namespaces.value = extractNamespaceNames(res.data)
  } catch { namespaces.value = ['default'] }
  finally { namespaceLoading.value = false }
}

onMounted(() => {
  fetchNamespaces()
  if (props.isEdit && props.initialData) parseInitialData(props.initialData)
})

watch(() => props.initialData, (newData) => {
  if (newData) parseInitialData(newData)
})

// ============ Parse helpers ============

function parseEnvVar(e: any) {
  if (e.valueFrom?.configMapKeyRef) {
    return { name: e.name, value: '', type: 'configMapKeyRef' as const, configMapName: e.valueFrom.configMapKeyRef.name || '', configMapKey: e.valueFrom.configMapKeyRef.key || '', secretName: '', secretKey: '', fieldPath: '' }
  }
  if (e.valueFrom?.secretKeyRef) {
    return { name: e.name, value: '', type: 'secretKeyRef' as const, configMapName: '', configMapKey: '', secretName: e.valueFrom.secretKeyRef.name || '', secretKey: e.valueFrom.secretKeyRef.key || '', fieldPath: '' }
  }
  if (e.valueFrom?.fieldRef) {
    return { name: e.name, value: '', type: 'fieldRef' as const, configMapName: '', configMapKey: '', secretName: '', secretKey: '', fieldPath: e.valueFrom.fieldRef.fieldPath || '' }
  }
  return { name: e.name, value: e.value || '', type: 'plain' as const, configMapName: '', configMapKey: '', secretName: '', secretKey: '', fieldPath: '' }
}

function parseContainer(c: any) {
  return {
    name: c.name || '',
    image: c.image || '',
    imagePullPolicy: c.imagePullPolicy || 'IfNotPresent',
    ports: (c.ports || []).map((p: any) => ({ name: p.name || '', containerPort: p.containerPort || null, protocol: p.protocol || 'TCP' })),
    env: (c.env || []).map(parseEnvVar),
    resources: {
      requests: { cpu: c.resources?.requests?.cpu || '', memory: c.resources?.requests?.memory || '' },
      limits: { cpu: c.resources?.limits?.cpu || '', memory: c.resources?.limits?.memory || '' },
    },
    volumeMounts: (c.volumeMounts || []).map((m: any) => ({ name: m.name || '', mountPath: m.mountPath || '', subPath: m.subPath || '', readOnly: m.readOnly || false })),
    livenessProbe: parseProbe(c.livenessProbe),
    readinessProbe: parseProbe(c.readinessProbe),
    startupProbe: parseProbe(c.startupProbe),
    command: c.command?.join('\n') || '',
    args: c.args?.join('\n') || '',
    lifecycle: { preStop: parseLifecycleHandler(c.lifecycle?.preStop), postStart: parseLifecycleHandler(c.lifecycle?.postStart) },
    securityContext: {
      runAsUser: c.securityContext?.runAsUser ?? null,
      runAsNonRoot: c.securityContext?.runAsNonRoot || false,
      readOnlyRootFilesystem: c.securityContext?.readOnlyRootFilesystem || false,
      privileged: c.securityContext?.privileged || false,
      capabilitiesAdd: c.securityContext?.capabilities?.add || [],
      capabilitiesDrop: c.securityContext?.capabilities?.drop || [],
    },
  }
}

function parseProbe(probeData: any) {
  if (!probeData) return null
  return {
    type: probeData.httpGet ? 'httpGet' : probeData.tcpSocket ? 'tcpSocket' : 'exec',
    httpGetPath: probeData.httpGet?.path || '/',
    httpGetPort: probeData.httpGet?.port || 80,
    tcpSocketPort: probeData.tcpSocket?.port || null,
    execCommand: probeData.exec?.command?.join('\n') || '',
    initialDelaySeconds: probeData.initialDelaySeconds || 15,
    periodSeconds: probeData.periodSeconds || 10,
    timeoutSeconds: probeData.timeoutSeconds || 5,
    failureThreshold: probeData.failureThreshold || 3,
  }
}

function parseLifecycleHandler(data: any) {
  if (!data) return null
  return {
    type: data.exec ? 'exec' : data.httpGet ? 'httpGet' : 'tcpSocket',
    execCommand: data.exec?.command?.join('\n') || '',
    httpGetPath: data.httpGet?.path || '/',
    httpGetPort: data.httpGet?.port || 80,
    tcpSocketPort: data.tcpSocket?.port || null,
  }
}

function parseInitialData(data: any) {
  if (!data) return

  const metadata = data.metadata || {}
  const spec = data.spec || {}
  const template = spec.template || {}
  const podSpec = template.spec || {}

  form.name = metadata.name || ''
  form.namespace = metadata.namespace || 'default'

  const labels = metadata.labels || {}
  form.labels = Object.entries(labels).map(([key, value]) => ({ key, value: value as string }))
  if (form.labels.length === 0) form.labels.push({ key: 'app', value: '' })

  const annotations = template.metadata?.annotations || {}
  form.annotations = Object.entries(annotations).map(([key, value]) => ({ key, value: value as string }))

  const containers = podSpec.containers || []
  form.containers = containers.map(parseContainer)
  if (form.containers.length === 0) form.containers.push(createEmptyContainer())

  const initContainers = podSpec.initContainers || []
  form.initContainers = initContainers.map(parseContainer)

  const volumes = podSpec.volumes || []
  form.volumes = volumes.map((v: any) => ({
    name: v.name || '',
    type: v.emptyDir ? 'emptyDir' : v.hostPath ? 'hostPath' : v.configMap ? 'configMap' : v.secret ? 'secret' : v.persistentVolumeClaim ? 'pvc' : 'emptyDir',
    hostPath: v.hostPath?.path || '',
    hostPathType: v.hostPath?.type || 'DirectoryOrCreate',
    configMapName: v.configMap?.name || '',
    secretName: v.secret?.secretName || '',
    pvcName: v.persistentVolumeClaim?.claimName || '',
  }))

  const nodeSelector = podSpec.nodeSelector || {}
  form.nodeSelector = Object.entries(nodeSelector).map(([key, value]) => ({ key, value: value as string }))

  const tolerations = podSpec.tolerations || []
  form.tolerations = tolerations.map((t: any) => ({
    key: t.key || '', operator: t.operator || 'Equal', value: t.value || '',
    effect: t.effect || 'NoSchedule', tolerationSeconds: t.tolerationSeconds || null,
  }))

  form.serviceAccountName = podSpec.serviceAccountName || ''
  form.terminationGracePeriodSeconds = podSpec.terminationGracePeriodSeconds || null

  const imagePullSecrets = podSpec.imagePullSecrets || []
  form.imagePullSecrets = imagePullSecrets.map((s: any) => s.name || '')

  // Job-specific
  form.completions = spec.completions || 1
  form.parallelism = spec.parallelism || 1
  form.backoffLimit = spec.backoffLimit ?? 6
  form.activeDeadlineSeconds = spec.activeDeadlineSeconds || null
  form.ttlSecondsAfterFinished = spec.ttlSecondsAfterFinished || null
  form.completionMode = spec.completionMode || 'NonIndexed'
  form.restartPolicy = podSpec.restartPolicy || 'Never'

  // Pod Affinity
  const affinity = podSpec.affinity || {}
  const parseAffinityRules = (rules: any[]) => {
    return (rules || []).map((r: any) => ({
      weight: r.weight || 1,
      topologyKey: r.podAffinityTerm?.topologyKey || r.topologyKey || '',
      namespaces: r.namespaces?.join(', ') || '',
      labelKey: r.labelSelector?.matchExpressions?.[0]?.key || '',
      labelValue: r.labelSelector?.matchExpressions?.[0]?.values?.[0] || '',
    }))
  }
  form.podAffinityRules = parseAffinityRules(affinity.podAffinity?.preferredDuringSchedulingIgnoredDuringExecution || [])
  if (affinity.podAffinity?.requiredDuringSchedulingIgnoredDuringExecution) {
    form.podAffinityRules.push(...affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution.map((r: any) => ({ ...parseAffinityRules([r])[0], weight: 0 })))
  }
  form.podAntiAffinityRules = parseAffinityRules(affinity.podAntiAffinity?.preferredDuringSchedulingIgnoredDuringExecution || [])
  if (affinity.podAntiAffinity?.requiredDuringSchedulingIgnoredDuringExecution) {
    form.podAntiAffinityRules.push(...affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution.map((r: any) => ({ ...parseAffinityRules([r])[0], weight: 0 })))
  }

  form.topologySpreadConstraints = (podSpec.topologySpreadConstraints || []).map((t: any) => ({
    maxSkew: t.maxSkew || 1, topologyKey: t.topologyKey || '',
    whenUnsatisfiable: t.whenUnsatisfiable || 'DoNotSchedule',
    labelKey: t.labelSelector?.matchExpressions?.[0]?.key || '',
    labelValue: t.labelSelector?.matchExpressions?.[0]?.values?.[0] || '',
  }))
}

// ============ Build helpers ============

function buildProbe(probe: any): any {
  if (!probe) return undefined
  const p: any = { initialDelaySeconds: probe.initialDelaySeconds, periodSeconds: probe.periodSeconds, timeoutSeconds: probe.timeoutSeconds, failureThreshold: probe.failureThreshold }
  if (probe.type === 'httpGet') { p.httpGet = { path: probe.httpGetPath, port: probe.httpGetPort } }
  else if (probe.type === 'tcpSocket') { p.tcpSocket = { port: probe.tcpSocketPort } }
  else if (probe.type === 'exec') { p.exec = { command: probe.execCommand.split('\n').map((s: string) => s.trim()).filter(Boolean) } }
  return p
}

function buildLifecycleHandler(handler: any): any {
  if (!handler) return undefined
  if (handler.type === 'exec') return { exec: { command: handler.execCommand.split('\n').map((s: string) => s.trim()).filter(Boolean) } }
  if (handler.type === 'httpGet') return { httpGet: { path: handler.httpGetPath, port: handler.httpGetPort } }
  if (handler.type === 'tcpSocket') return { tcpSocket: { port: handler.tcpSocketPort } }
  return undefined
}

function buildContainer(c: any): Record<string, any> {
  const container: Record<string, any> = { name: c.name, image: c.image, imagePullPolicy: c.imagePullPolicy }
  if (c.command) container.command = c.command.split('\n').map((s: string) => s.trim()).filter(Boolean)
  if (c.args) container.args = c.args.split('\n').map((s: string) => s.trim()).filter(Boolean)
  const ports = c.ports.filter((p: any) => p.containerPort).map((p: any) => { const port: any = { containerPort: p.containerPort, protocol: p.protocol }; if (p.name) port.name = p.name; return port })
  if (ports.length > 0) container.ports = ports
  const env = c.env.filter((e: any) => e.name.trim()).map((e: any) => {
    if (e.type === 'configMapKeyRef' && e.configMapName && e.configMapKey) return { name: e.name, valueFrom: { configMapKeyRef: { name: e.configMapName, key: e.configMapKey } } }
    if (e.type === 'secretKeyRef' && e.secretName && e.secretKey) return { name: e.name, valueFrom: { secretKeyRef: { name: e.secretName, key: e.secretKey } } }
    if (e.type === 'fieldRef' && e.fieldPath) return { name: e.name, valueFrom: { fieldRef: { fieldPath: e.fieldPath } } }
    return { name: e.name, value: e.value }
  })
  if (env.length > 0) container.env = env
  const resources: any = {}; const requests: any = {}; const limits: any = {}
  if (c.resources.requests.cpu) requests.cpu = c.resources.requests.cpu
  if (c.resources.requests.memory) requests.memory = c.resources.requests.memory
  if (c.resources.limits.cpu) limits.cpu = c.resources.limits.cpu
  if (c.resources.limits.memory) limits.memory = c.resources.limits.memory
  if (Object.keys(requests).length > 0) resources.requests = requests
  if (Object.keys(limits).length > 0) resources.limits = limits
  if (Object.keys(resources).length > 0) container.resources = resources
  const mounts = c.volumeMounts.filter((m: any) => m.name && m.mountPath).map((m: any) => { const vm: any = { name: m.name, mountPath: m.mountPath }; if (m.subPath) vm.subPath = m.subPath; if (m.readOnly) vm.readOnly = true; return vm })
  if (mounts.length > 0) container.volumeMounts = mounts
  const liveness = buildProbe(c.livenessProbe)
  if (liveness) container.livenessProbe = liveness
  const readiness = buildProbe(c.readinessProbe)
  if (readiness) container.readinessProbe = readiness
  const startup = buildProbe(c.startupProbe)
  if (startup) container.startupProbe = startup
  const lifecycle: any = {}
  if (c.lifecycle.postStart) lifecycle.postStart = buildLifecycleHandler(c.lifecycle.postStart)
  if (c.lifecycle.preStop) lifecycle.preStop = buildLifecycleHandler(c.lifecycle.preStop)
  if (Object.keys(lifecycle).length > 0) container.lifecycle = lifecycle
  const sc: any = {}
  if (c.securityContext.runAsUser !== null) sc.runAsUser = c.securityContext.runAsUser
  if (c.securityContext.runAsNonRoot) sc.runAsNonRoot = true
  if (c.securityContext.readOnlyRootFilesystem) sc.readOnlyRootFilesystem = true
  if (c.securityContext.privileged) sc.privileged = true
  const caps: any = {}
  if (c.securityContext.capabilitiesAdd.length > 0) caps.add = c.securityContext.capabilitiesAdd
  if (c.securityContext.capabilitiesDrop.length > 0) caps.drop = c.securityContext.capabilitiesDrop
  if (Object.keys(caps).length > 0) sc.capabilities = caps
  if (Object.keys(sc).length > 0) container.securityContext = sc
  return container
}

function buildAffinity(rules: any[]): any {
  const validRules = rules.filter((r: any) => r.labelKey)
  if (validRules.length === 0) return undefined
  const preferred = validRules.filter((r: any) => r.weight > 0)
  const required = validRules.filter((r: any) => r.weight === 0)
  const result: any = {}
  if (preferred.length > 0) {
    result.preferredDuringSchedulingIgnoredDuringExecution = preferred.map((r: any) => ({
      weight: r.weight,
      podAffinityTerm: {
        labelSelector: { matchExpressions: [{ key: r.labelKey, operator: 'In', values: [r.labelValue] }] },
        topologyKey: r.topologyKey,
        ...(r.namespaces ? { namespaces: r.namespaces.split(',').map((s: string) => s.trim()) } : {}),
      },
    }))
  }
  if (required.length > 0) {
    result.requiredDuringSchedulingIgnoredDuringExecution = required.map((r: any) => ({
      labelSelector: { matchExpressions: [{ key: r.labelKey, operator: 'In', values: [r.labelValue] }] },
      topologyKey: r.topologyKey,
      ...(r.namespaces ? { namespaces: r.namespaces.split(',').map((s: string) => s.trim()) } : {}),
    }))
  }
  return result
}

function buildK8sResource(): Record<string, any> {
  const labels: Record<string, string> = {}
  form.labels.forEach(l => { if (l.key.trim()) labels[l.key.trim()] = l.value })

  const containers = form.containers.map(buildContainer)
  const initContainers = form.initContainers.map(buildContainer)

  const volumes = form.volumes.filter(v => v.name).map(v => {
    const vol: any = { name: v.name }
    if (v.type === 'emptyDir') vol.emptyDir = {}
    else if (v.type === 'hostPath') vol.hostPath = { path: v.hostPath, type: v.hostPathType || 'DirectoryOrCreate' }
    else if (v.type === 'configMap') vol.configMap = { name: v.configMapName || v.name }
    else if (v.type === 'secret') vol.secret = { secretName: v.secretName || v.name }
    else if (v.type === 'pvc') vol.persistentVolumeClaim = { claimName: v.pvcName || v.name }
    return vol
  })

  const nodeSelector: Record<string, string> = {}
  form.nodeSelector.forEach(ns => { if (ns.key.trim()) nodeSelector[ns.key.trim()] = ns.value })

  const annotations: Record<string, string> = {}
  form.annotations.forEach(a => { if (a.key.trim()) annotations[a.key.trim()] = a.value })

  const tolerations = form.tolerations.filter(t => t.key).map(t => {
    const tol: any = { key: t.key, operator: t.operator, effect: t.effect }
    if (t.value) tol.value = t.value
    if (t.tolerationSeconds) tol.tolerationSeconds = t.tolerationSeconds
    return tol
  })

  const imagePullSecrets = form.imagePullSecrets.filter(s => s).map(s => ({ name: s }))

  const podSpec: any = { containers, restartPolicy: form.restartPolicy || 'Never' }
  if (initContainers.length > 0) podSpec.initContainers = initContainers
  if (volumes.length > 0) podSpec.volumes = volumes
  if (Object.keys(nodeSelector).length > 0) podSpec.nodeSelector = nodeSelector
  if (tolerations.length > 0) podSpec.tolerations = tolerations
  if (form.serviceAccountName) podSpec.serviceAccountName = form.serviceAccountName
  if (form.terminationGracePeriodSeconds) podSpec.terminationGracePeriodSeconds = form.terminationGracePeriodSeconds
  if (imagePullSecrets.length > 0) podSpec.imagePullSecrets = imagePullSecrets

  const affinity: any = {}
  const podAffinity = buildAffinity(form.podAffinityRules)
  if (podAffinity) affinity.podAffinity = podAffinity
  const podAntiAffinity = buildAffinity(form.podAntiAffinityRules)
  if (podAntiAffinity) affinity.podAntiAffinity = podAntiAffinity
  if (Object.keys(affinity).length > 0) podSpec.affinity = affinity

  if (form.topologySpreadConstraints.length > 0) {
    podSpec.topologySpreadConstraints = form.topologySpreadConstraints
      .filter(t => t.topologyKey)
      .map(t => ({
        maxSkew: t.maxSkew, topologyKey: t.topologyKey, whenUnsatisfiable: t.whenUnsatisfiable,
        labelSelector: { matchExpressions: [{ key: t.labelKey || 'app', operator: 'In', values: [t.labelValue || ''] }] },
      }))
  }

  const podTemplate: any = { metadata: { labels: { ...labels } }, spec: podSpec }
  if (Object.keys(annotations).length > 0) podTemplate.metadata.annotations = annotations

  const resource: any = {
    apiVersion: 'batch/v1',
    kind: 'Job',
    metadata: { name: form.name, namespace: form.namespace, labels: { ...labels } },
    spec: { template: podTemplate },
  }

  if (form.completions !== null && form.completions !== undefined) resource.spec.completions = form.completions
  if (form.parallelism !== null && form.parallelism !== undefined) resource.spec.parallelism = form.parallelism
  if (form.backoffLimit !== null && form.backoffLimit !== undefined) resource.spec.backoffLimit = form.backoffLimit
  if (form.activeDeadlineSeconds !== null && form.activeDeadlineSeconds !== undefined) resource.spec.activeDeadlineSeconds = form.activeDeadlineSeconds
  if (form.ttlSecondsAfterFinished !== null && form.ttlSecondsAfterFinished !== undefined) resource.spec.ttlSecondsAfterFinished = form.ttlSecondsAfterFinished
  if (form.completionMode && form.completionMode !== 'NonIndexed') resource.spec.completionMode = form.completionMode

  return resource
}

const generatedYaml = computed(() => yaml.dump(buildK8sResource(), { indent: 2, lineWidth: -1, noRefs: true }))

function parseCpuToMillicores(cpu: string): number | null {
  if (!cpu) return null
  cpu = cpu.trim()
  if (cpu.endsWith('m')) return parseInt(cpu.slice(0, -1), 10)
  const val = parseFloat(cpu)
  return isNaN(val) ? null : Math.round(val * 1000)
}

function parseMemoryToBytes(mem: string): number | null {
  if (!mem) return null
  mem = mem.trim()
  const units: Record<string, number> = { 'Ki': 1024, 'Mi': 1024**2, 'Gi': 1024**3, 'Ti': 1024**4, 'K': 1000, 'M': 1000**2, 'G': 1000**3, 'T': 1000**4 }
  for (const [suffix, multiplier] of Object.entries(units)) {
    if (mem.endsWith(suffix)) {
      const val = parseFloat(mem.slice(0, -suffix.length))
      return isNaN(val) ? null : Math.round(val * multiplier)
    }
  }
  const val = parseFloat(mem)
  return isNaN(val) ? null : val
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  for (let i = 0; i < form.containers.length; i++) {
    if (!form.containers[i].name) { ElMessage.error(t('workload.containerNameRequired', { n: i + 1 })); return }
    if (!form.containers[i].image) { ElMessage.error(t('workload.containerImageRequired', { n: i + 1 })); return }
    const c = form.containers[i]
    if (c.resources.requests.cpu && c.resources.limits.cpu) {
      const reqCpu = parseCpuToMillicores(c.resources.requests.cpu)
      const limCpu = parseCpuToMillicores(c.resources.limits.cpu)
      if (reqCpu !== null && limCpu !== null && reqCpu > limCpu) {
        ElMessage.error(t('workload.cpuRequestsExceedLimits', { n: i + 1 })); return
      }
    }
    if (c.resources.requests.memory && c.resources.limits.memory) {
      const reqMem = parseMemoryToBytes(c.resources.requests.memory)
      const limMem = parseMemoryToBytes(c.resources.limits.memory)
      if (reqMem !== null && limMem !== null && reqMem > limMem) {
        ElMessage.error(t('workload.memoryRequestsExceedLimits', { n: i + 1 })); return
      }
    }
  }

  submitting.value = true
  try {
    if (props.onSubmit) {
      await props.onSubmit(generatedYaml.value)
    } else if (props.isEdit) {
      await updateJobYaml({ namespace: form.namespace, name: form.name, yaml: generatedYaml.value })
      ElMessage.success(t('common.updateSuccess'))
      emit('success')
    } else {
      await createJob({ namespace: form.namespace, yaml: generatedYaml.value })
      ElMessage.success(t('common.createSuccess'))
      router.push('/workloads/jobs')
    }
  } catch (e: any) { ElMessage.error(e?.message || (props.isEdit ? t('common.updateFailed') : t('common.createFailed'))) }
  finally { submitting.value = false }
}

function handleCancel() {
  if (props.isEdit) emit('cancel')
  else router.push('/workloads/jobs')
}

function addImagePullSecret() { form.imagePullSecrets.push('') }
function removeImagePullSecret(i: number) { form.imagePullSecrets.splice(i, 1) }
</script>

<template>
  <div class="workload-form">
    <el-form ref="formRef" :model="form" :rules="formRules" label-position="top">
      <!-- Section 1: Basic Info -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">{{ t('config.basicInfo') }}</div>
        </div>
        <div class="section-content">
          <div class="fields-grid">
            <el-form-item :label="t('common.name')" prop="name">
              <el-input v-model="form.name" placeholder="my-job" />
            </el-form-item>
            <el-form-item :label="t('common.namespace_label')" prop="namespace">
              <el-select v-model="form.namespace" filterable placeholder="选择命名空间" style="width: 100%;" :loading="namespaceLoading">
                <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
              </el-select>
            </el-form-item>
            <el-form-item label="服务账号">
              <el-input v-model="form.serviceAccountName" placeholder="default" />
            </el-form-item>
          </div>
        </div>
      </div>

      <!-- Section: Labels & Annotations -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">标签与注解</div>
        </div>
        <div class="section-content">
          <LabelsAnnotationsForm :labels="form.labels" :annotations="form.annotations" />
        </div>
      </div>

      <!-- Section 2: Job Config -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">任务配置</div>
        </div>
        <div class="section-content">
          <div class="fields-grid">
            <el-form-item label="完成数 (Completions)">
              <el-input-number v-model="form.completions" :min="1" style="width: 100%;" />
            </el-form-item>
            <el-form-item label="并行度 (Parallelism)">
              <el-input-number v-model="form.parallelism" :min="1" style="width: 100%;" />
            </el-form-item>
            <el-form-item label="重试次数 (Backoff Limit)">
              <el-input-number v-model="form.backoffLimit" :min="0" style="width: 100%;" />
            </el-form-item>
            <el-form-item label="超时时间(秒)">
              <el-input-number v-model="form.activeDeadlineSeconds" :min="1" placeholder="无限制" style="width: 100%;" />
            </el-form-item>
            <el-form-item label="完成后 TTL(秒)">
              <el-input-number v-model="form.ttlSecondsAfterFinished" :min="0" placeholder="不自动清理" style="width: 100%;" />
            </el-form-item>
            <el-form-item label="完成模式">
              <el-select v-model="form.completionMode" style="width: 100%;">
                <el-option label="NonIndexed (默认)" value="NonIndexed" />
                <el-option label="Indexed (索引模式)" value="Indexed" />
              </el-select>
            </el-form-item>
            <el-form-item label="重启策略 (Restart Policy)">
              <el-select v-model="form.restartPolicy" style="width: 100%;">
                <el-option label="Never - 不重启，创建新 Pod" value="Never" />
                <el-option label="OnFailure - 容器内重启" value="OnFailure" />
              </el-select>
              <div class="form-help">Never: 失败后创建新Pod；OnFailure: 在容器内重启</div>
            </el-form-item>
          </div>
        </div>
      </div>

      <!-- Section 3: Container Config -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">容器配置</div>
        </div>
        <div class="section-content">
          <ContainerConfigForm :containers="form.containers" :min-containers="1" />
        </div>
      </div>

      <!-- Section: Init Containers -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">初始化容器</div>
        </div>
        <div class="section-content">
          <ContainerConfigForm :containers="form.initContainers" title="初始化容器" :min-containers="0" />
        </div>
      </div>

      <!-- Section 4: Storage -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">存储配置</div>
        </div>
        <div class="section-content">
          <StorageConfigForm :volumes="form.volumes" :containers="form.containers" />
        </div>
      </div>

      <!-- Section 5: Health Probes -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">健康检查</div>
        </div>
        <div class="section-content">
          <HealthCheckForm :containers="form.containers" />
        </div>
      </div>

      <!-- Section 6: Security -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">安全设置</div>
        </div>
        <div class="section-content">
          <SecurityContextForm :containers="form.containers" />
        </div>
      </div>

      <!-- Section 7: Scheduling -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">调度配置</div>
        </div>
        <div class="section-content">
          <SchedulingForm
            v-model:node-selector="form.nodeSelector"
            v-model:tolerations="form.tolerations"
            v-model:pod-affinity-rules="form.podAffinityRules"
            v-model:pod-anti-affinity-rules="form.podAntiAffinityRules"
            v-model:topology-spread-constraints="form.topologySpreadConstraints"
          />
        </div>
      </div>

      <!-- Section 8: Advanced -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">高级配置</div>
        </div>
        <div class="section-content">
          <div class="fields-grid">
            <el-form-item label="优雅终止时间(秒)">
              <el-input-number v-model="form.terminationGracePeriodSeconds" :min="0" :max="300" style="width: 100%;" />
            </el-form-item>
          </div>

          <el-form-item label="镜像拉取密钥">
            <div style="width: 100%;">
              <div v-for="(_s, i) in form.imagePullSecrets" :key="i" class="kv-row">
                <el-input v-model="form.imagePullSecrets[i]" placeholder="Secret 名称" />
                <el-button type="danger" text circle @click="removeImagePullSecret(i)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
              <el-button text type="primary" @click="addImagePullSecret" size="small">
                <el-icon><Plus /></el-icon> 添加密钥
              </el-button>
            </div>
          </el-form-item>
        </div>
      </div>

      <!-- Submit Button -->
      <div class="form-section">
        <div class="section-sidebar"></div>
        <div class="section-content">
          <div class="form-actions">
            <el-button @click="handleCancel">{{ t('common.cancel') }}</el-button>
            <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ t('common.create') }}</el-button>
          </div>
        </div>
      </div>
    </el-form>
  </div>
</template>

<style scoped>
.workload-form {
  padding: 0 40px;
  max-width: 1000px;
  margin: 0 auto;
}

/* Section layout with sidebar titles */
.form-section {
  display: flex;
  gap: 24px;
  margin-bottom: 32px;
  align-items: flex-start;
}

.section-sidebar {
  width: 120px;
  flex-shrink: 0;
  position: sticky;
  top: 20px;
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--el-color-primary);
  padding: 12px 16px;
  background: var(--el-fill-color-lighter);
  border-left: 3px solid var(--el-color-primary);
  border-radius: 0 4px 4px 0;
}

.section-content {
  flex: 1;
  min-width: 0;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 24px;
  border-top: 1px solid var(--el-border-color-light);
}

.fields-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0 32px;
}

.fields-grid :deep(.el-form-item) {
  margin-bottom: 16px;
}

.fields-grid :deep(.el-form-item:last-child) {
  margin-bottom: 0;
}

/* Key-value rows */
.kv-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: center;
}

.kv-row :deep(.el-input) {
  flex: 1;
}
</style>
