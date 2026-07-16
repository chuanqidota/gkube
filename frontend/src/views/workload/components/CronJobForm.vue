<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import yaml from 'js-yaml'
import type { FormInstance, FormRules } from 'element-plus'
import { getNamespaceList, extractNamespaceNames, createCronJob } from '@/api/resource'

const router = useRouter()
const submitting = ref(false)
const namespaceLoading = ref(false)
const namespaces = ref<string[]>([])

interface Label { key: string; value: string }
interface Port { name: string; containerPort: number | null; protocol: string }
interface EnvVar {
  name: string; value: string; type: 'plain' | 'configMapKeyRef' | 'secretKeyRef' | 'fieldRef'
  configMapName: string; configMapKey: string
  secretName: string; secretKey: string
  fieldPath: string
}
interface Resources { requests: { cpu: string; memory: string }; limits: { cpu: string; memory: string } }
interface VolumeMount { name: string; mountPath: string; subPath: string; readOnly: boolean }
interface Probe { type: string; httpGetPath: string; httpGetPort: number | null; tcpSocketPort: number | null; execCommand: string; initialDelaySeconds: number; periodSeconds: number; timeoutSeconds: number; failureThreshold: number }
interface LifecycleHandler { type: 'exec' | 'httpGet' | 'tcpSocket'; execCommand: string; httpGetPath: string; httpGetPort: number | null; tcpSocketPort: number | null }
interface Volume { name: string; type: string; hostPath: string; hostPathType: string; configMapName: string; secretName: string; pvcName: string }
interface Tolerance { key: string; operator: string; value: string; effect: string; tolerationSeconds: number | null }
interface Annotation { key: string; value: string }
interface Container {
  name: string; image: string; imagePullPolicy: string
  ports: Port[]; env: EnvVar[]; resources: Resources
  volumeMounts: VolumeMount[]; livenessProbe: Probe | null; readinessProbe: Probe | null; startupProbe: Probe | null
  command: string; args: string
  lifecycle: { preStop: LifecycleHandler | null; postStart: LifecycleHandler | null }
  securityContext: { runAsUser: number | null; runAsNonRoot: boolean; readOnlyRootFilesystem: boolean; privileged: boolean; capabilitiesAdd: string[]; capabilitiesDrop: string[] }
}
interface AffinityRule { weight: number; topologyKey: string; namespaces: string; labelKey: string; labelValue: string }
interface TopologySpreadConstraint { maxSkew: number; topologyKey: string; whenUnsatisfiable: string; labelKey: string; labelValue: string }

interface FormData {
  name: string; namespace: string; labels: Label[]
  schedule: string; concurrencyPolicy: string; suspend: boolean
  successfulJobsHistoryLimit: number | null; failedJobsHistoryLimit: number | null
  completions: number | null; parallelism: number | null; backoffLimit: number | null
  containers: Container[]; initContainers: Container[]; volumes: Volume[]
  nodeSelector: Label[]; tolerations: Tolerance[]; annotations: Annotation[]
  serviceAccountName: string; terminationGracePeriodSeconds: number | null
  imagePullSecrets: string[]
  podAffinityRules: AffinityRule[]; podAntiAffinityRules: AffinityRule[]
  topologySpreadConstraints: TopologySpreadConstraint[]
}

function createEmptyProbe(): Probe {
  return { type: 'httpGet', httpGetPath: '/', httpGetPort: 80, tcpSocketPort: null, execCommand: '', initialDelaySeconds: 15, periodSeconds: 10, timeoutSeconds: 5, failureThreshold: 3 }
}

function createEmptyEnv(): EnvVar {
  return { name: '', value: '', type: 'plain', configMapName: '', configMapKey: '', secretName: '', secretKey: '', fieldPath: '' }
}

function createEmptyContainer(): Container {
  return {
    name: '', image: '', imagePullPolicy: 'IfNotPresent',
    ports: [], env: [],
    resources: { requests: { cpu: '', memory: '' }, limits: { cpu: '', memory: '' } },
    volumeMounts: [], livenessProbe: null, readinessProbe: null, startupProbe: null,
    command: '', args: '',
    lifecycle: { preStop: null, postStart: null },
    securityContext: { runAsUser: null, runAsNonRoot: false, readOnlyRootFilesystem: false, privileged: false, capabilitiesAdd: [], capabilitiesDrop: [] },
  }
}

const form = reactive<FormData>({
  name: '', namespace: 'default',
  labels: [{ key: 'app', value: '' }],
  schedule: '', concurrencyPolicy: 'Allow', suspend: false,
  successfulJobsHistoryLimit: 3, failedJobsHistoryLimit: 1,
  completions: 1, parallelism: 1, backoffLimit: 6,
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
  schedule: [{ required: true, message: '请输入调度表达式', trigger: 'blur' }],
}

async function fetchNamespaces() {
  namespaceLoading.value = true
  try {
    const res: any = await getNamespaceList()
    namespaces.value = extractNamespaceNames(res.data)
  } catch { namespaces.value = ['default'] }
  finally { namespaceLoading.value = false }
}

onMounted(fetchNamespaces)

function addLabel() { form.labels.push({ key: '', value: '' }) }
function removeLabel(i: number) { form.labels.splice(i, 1) }
function addContainer() { form.containers.push(createEmptyContainer()) }
function removeContainer(i: number) { if (form.containers.length > 1) form.containers.splice(i, 1) }
function addInitContainer() { form.initContainers.push(createEmptyContainer()) }
function removeInitContainer(i: number) { form.initContainers.splice(i, 1) }
function addPort(ci: number, isInit?: boolean) { const list = isInit ? form.initContainers : form.containers; list[ci].ports.push({ name: '', containerPort: null, protocol: 'TCP' }) }
function removePort(ci: number, pi: number, isInit?: boolean) { const list = isInit ? form.initContainers : form.containers; list[ci].ports.splice(pi, 1) }
function addEnv(ci: number, isInit?: boolean) { const list = isInit ? form.initContainers : form.containers; list[ci].env.push(createEmptyEnv()) }
function removeEnv(ci: number, ei: number, isInit?: boolean) { const list = isInit ? form.initContainers : form.containers; list[ci].env.splice(ei, 1) }
function addVolumeMount(ci: number, isInit?: boolean) { const list = isInit ? form.initContainers : form.containers; list[ci].volumeMounts.push({ name: '', mountPath: '', subPath: '', readOnly: false }) }
function removeVolumeMount(ci: number, mi: number, isInit?: boolean) { const list = isInit ? form.initContainers : form.containers; list[ci].volumeMounts.splice(mi, 1) }
function addNodeSelector() { form.nodeSelector.push({ key: '', value: '' }) }
function removeNodeSelector(i: number) { form.nodeSelector.splice(i, 1) }
function addVolume() { form.volumes.push({ name: '', type: 'emptyDir', hostPath: '', hostPathType: 'DirectoryOrCreate', configMapName: '', secretName: '', pvcName: '' }) }
function removeVolume(i: number) { form.volumes.splice(i, 1) }
function enableProbe(ci: number, probeType: 'livenessProbe' | 'readinessProbe' | 'startupProbe', isInit?: boolean) { const list = isInit ? form.initContainers : form.containers; list[ci][probeType] = createEmptyProbe() }
function disableProbe(ci: number, probeType: 'livenessProbe' | 'readinessProbe' | 'startupProbe', isInit?: boolean) { const list = isInit ? form.initContainers : form.containers; list[ci][probeType] = null }
function addToleration() { form.tolerations.push({ key: '', operator: 'Equal', value: '', effect: 'NoSchedule', tolerationSeconds: null }) }
function removeToleration(i: number) { form.tolerations.splice(i, 1) }
function addAnnotation() { form.annotations.push({ key: '', value: '' }) }
function removeAnnotation(i: number) { form.annotations.splice(i, 1) }
function addImagePullSecret() { form.imagePullSecrets.push('') }
function removeImagePullSecret(i: number) { form.imagePullSecrets.splice(i, 1) }
function addAffinityRule(type: 'podAffinityRules' | 'podAntiAffinityRules') { form[type].push({ weight: 1, topologyKey: 'kubernetes.io/hostname', namespaces: '', labelKey: '', labelValue: '' }) }
function removeAffinityRule(type: 'podAffinityRules' | 'podAntiAffinityRules', i: number) { form[type].splice(i, 1) }
function addTopologySpreadConstraint() { form.topologySpreadConstraints.push({ maxSkew: 1, topologyKey: 'kubernetes.io/hostname', whenUnsatisfiable: 'DoNotSchedule', labelKey: '', labelValue: '' }) }
function removeTopologySpreadConstraint(i: number) { form.topologySpreadConstraints.splice(i, 1) }
function addCapability(ci: number, type: 'capabilitiesAdd' | 'capabilitiesDrop', isInit?: boolean) { const list = isInit ? form.initContainers : form.containers; list[ci].securityContext[type].push('') }
function removeCapability(ci: number, type: 'capabilitiesAdd' | 'capabilitiesDrop', capIdx: number, isInit?: boolean) { const list = isInit ? form.initContainers : form.containers; list[ci].securityContext[type].splice(capIdx, 1) }

const generatedYaml = computed(() => yaml.dump(buildK8sResource(), { indent: 2, lineWidth: -1, noRefs: true }))

function buildProbe(probe: Probe | null): any {
  if (!probe) return undefined
  const p: any = { initialDelaySeconds: probe.initialDelaySeconds, periodSeconds: probe.periodSeconds, timeoutSeconds: probe.timeoutSeconds, failureThreshold: probe.failureThreshold }
  if (probe.type === 'httpGet') { p.httpGet = { path: probe.httpGetPath, port: probe.httpGetPort } }
  else if (probe.type === 'tcpSocket') { p.tcpSocket = { port: probe.tcpSocketPort } }
  else if (probe.type === 'exec') { p.exec = { command: probe.execCommand.split(' ').filter(Boolean) } }
  return p
}

function buildLifecycleHandler(handler: LifecycleHandler | null): any {
  if (!handler) return undefined
  if (handler.type === 'exec') return { exec: { command: handler.execCommand.split(' ').filter(Boolean) } }
  if (handler.type === 'httpGet') return { httpGet: { path: handler.httpGetPath, port: handler.httpGetPort } }
  if (handler.type === 'tcpSocket') return { tcpSocket: { port: handler.tcpSocketPort } }
  return undefined
}

function buildContainer(c: Container): Record<string, any> {
  const container: Record<string, any> = { name: c.name, image: c.image, imagePullPolicy: c.imagePullPolicy }
  if (c.command) container.command = c.command.split(' ').filter(Boolean)
  if (c.args) container.args = c.args.split(' ').filter(Boolean)
  const ports = c.ports.filter(p => p.containerPort).map(p => { const port: any = { containerPort: p.containerPort, protocol: p.protocol }; if (p.name) port.name = p.name; return port })
  if (ports.length > 0) container.ports = ports
  const env = c.env.filter(e => e.name.trim()).map(e => {
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
  const mounts = c.volumeMounts.filter(m => m.name && m.mountPath).map(m => { const vm: any = { name: m.name, mountPath: m.mountPath }; if (m.subPath) vm.subPath = m.subPath; if (m.readOnly) vm.readOnly = true; return vm })
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

function buildAffinity(rules: AffinityRule[]): any {
  const validRules = rules.filter(r => r.labelKey)
  if (validRules.length === 0) return undefined
  const preferred = validRules.filter(r => r.weight > 0)
  const required = validRules.filter(r => r.weight === 0)
  const result: any = {}
  if (preferred.length > 0) {
    result.preferredDuringSchedulingIgnoredDuringExecution = preferred.map(r => ({
      weight: r.weight,
      podAffinityTerm: {
        labelSelector: { matchExpressions: [{ key: r.labelKey, operator: 'In', values: [r.labelValue] }] },
        topologyKey: r.topologyKey,
        ...(r.namespaces ? { namespaces: r.namespaces.split(',').map(s => s.trim()) } : {}),
      },
    }))
  }
  if (required.length > 0) {
    result.requiredDuringSchedulingIgnoredDuringExecution = required.map(r => ({
      labelSelector: { matchExpressions: [{ key: r.labelKey, operator: 'In', values: [r.labelValue] }] },
      topologyKey: r.topologyKey,
      ...(r.namespaces ? { namespaces: r.namespaces.split(',').map(s => s.trim()) } : {}),
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

  const podSpec: any = { containers, restartPolicy: 'Never' }
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
        maxSkew: t.maxSkew,
        topologyKey: t.topologyKey,
        whenUnsatisfiable: t.whenUnsatisfiable,
        labelSelector: { matchExpressions: [{ key: t.labelKey || 'app', operator: 'In', values: [t.labelValue || ''] }] },
      }))
  }

  const podTemplate: any = { metadata: { labels: { ...labels } }, spec: podSpec }
  if (Object.keys(annotations).length > 0) podTemplate.metadata.annotations = annotations

  const jobSpec: Record<string, any> = { template: podTemplate }
  if (form.completions !== null && form.completions !== undefined) jobSpec.completions = form.completions
  if (form.parallelism !== null && form.parallelism !== undefined) jobSpec.parallelism = form.parallelism
  if (form.backoffLimit !== null && form.backoffLimit !== undefined) jobSpec.backoffLimit = form.backoffLimit

  const resource: any = {
    apiVersion: 'batch/v1',
    kind: 'CronJob',
    metadata: { name: form.name, namespace: form.namespace, labels: { ...labels } },
    spec: {
      schedule: form.schedule,
      concurrencyPolicy: form.concurrencyPolicy,
      suspend: form.suspend,
      jobTemplate: { spec: jobSpec },
    },
  }

  if (form.successfulJobsHistoryLimit !== null && form.successfulJobsHistoryLimit !== undefined) resource.spec.successfulJobsHistoryLimit = form.successfulJobsHistoryLimit
  if (form.failedJobsHistoryLimit !== null && form.failedJobsHistoryLimit !== undefined) resource.spec.failedJobsHistoryLimit = form.failedJobsHistoryLimit

  return resource
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  for (let i = 0; i < form.containers.length; i++) {
    if (!form.containers[i].name) { ElMessage.error(`容器 ${i + 1}: 名称不能为空`); return }
    if (!form.containers[i].image) { ElMessage.error(`容器 ${i + 1}: 镜像不能为空`); return }
  }

  submitting.value = true
  try {
    await createCronJob({ namespace: form.namespace, yaml: generatedYaml.value })
    ElMessage.success('CronJob 创建成功')
    router.push('/workloads/cronjobs')
  } catch (e: any) { ElMessage.error(e?.message || '创建失败') }
  finally { submitting.value = false }
}

function handleCancel() { router.push('/workloads/cronjobs') }
</script>

<template>
  <div class="workload-form">
    <el-form ref="formRef" :model="form" :rules="formRules" label-position="top">
      <!-- Section 1: Basic Info -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">基本信息</div>
        </div>
        <div class="section-content">
          <div class="fields-grid">
            <el-form-item label="名称" prop="name">
              <el-input v-model="form.name" placeholder="my-cronjob" />
            </el-form-item>
            <el-form-item label="命名空间" prop="namespace">
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
          <el-form-item label="标签">
            <div style="width: 100%;">
              <div v-for="(label, i) in form.labels" :key="i" class="kv-row">
                <el-input v-model="label.key" placeholder="Key" />
                <el-input v-model="label.value" placeholder="Value" />
                <el-button type="danger" text circle :disabled="form.labels.length <= 1" @click="removeLabel(i)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
              <el-button text type="primary" @click="addLabel" size="small">
                <el-icon><Plus /></el-icon> 添加标签
              </el-button>
            </div>
          </el-form-item>
          <el-form-item label="注解">
            <div style="width: 100%;">
              <div v-for="(ann, i) in form.annotations" :key="i" class="kv-row">
                <el-input v-model="ann.key" placeholder="Key" />
                <el-input v-model="ann.value" placeholder="Value" />
                <el-button type="danger" text circle @click="removeAnnotation(i)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
              <el-button text type="primary" @click="addAnnotation" size="small">
                <el-icon><Plus /></el-icon> 添加注解
              </el-button>
            </div>
          </el-form-item>
        </div>
      </div>

      <!-- Section 2: CronJob Config -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">定时任务配置</div>
        </div>
        <div class="section-content">
          <div class="fields-grid">
            <el-form-item label="调度表达式 (Schedule)" prop="schedule" class="full-width">
              <el-input v-model="form.schedule" placeholder="*/5 * * * *" />
              <div class="form-help">Cron 表达式，例如 "0 */6 * * *" 表示每6小时执行一次</div>
            </el-form-item>
            <el-form-item label="并发策略 (Concurrency Policy)">
              <el-select v-model="form.concurrencyPolicy" style="width: 100%;">
                <el-option label="Allow - 允许并发" value="Allow" />
                <el-option label="Forbid - 禁止并发" value="Forbid" />
                <el-option label="Replace - 替换旧任务" value="Replace" />
              </el-select>
            </el-form-item>
            <el-form-item label="暂停 (Suspend)">
              <el-switch v-model="form.suspend" />
            </el-form-item>
            <el-form-item label="成功任务历史限制">
              <el-input-number v-model="form.successfulJobsHistoryLimit" :min="0" style="width: 100%;" />
            </el-form-item>
            <el-form-item label="失败任务历史限制">
              <el-input-number v-model="form.failedJobsHistoryLimit" :min="0" style="width: 100%;" />
            </el-form-item>
          </div>
          <el-divider />
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
          </div>
        </div>
      </div>

      <!-- Section 3: Container Config -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">容器配置</div>
        </div>
        <div class="section-content">
          <div v-for="(container, ci) in form.containers" :key="ci" class="container-card">
            <div class="container-card-header">
              <div class="container-title">
                <span class="container-index">{{ ci + 1 }}</span>
                <span>{{ container.name || '未命名容器' }}</span>
              </div>
              <el-button v-if="form.containers.length > 1" type="danger" text size="small" @click="removeContainer(ci)">
                <el-icon><Delete /></el-icon> 移除
              </el-button>
            </div>
            <div class="fields-grid">
              <el-form-item label="容器名称" required>
                <el-input v-model="container.name" placeholder="nginx" />
              </el-form-item>
              <el-form-item label="镜像" required>
                <el-input v-model="container.image" placeholder="nginx:1.25" />
              </el-form-item>
              <el-form-item label="拉取策略">
                <el-select v-model="container.imagePullPolicy" style="width: 100%;">
                  <el-option label="Always" value="Always" />
                  <el-option label="IfNotPresent" value="IfNotPresent" />
                  <el-option label="Never" value="Never" />
                </el-select>
              </el-form-item>
              <el-form-item label="启动命令 (Command)">
                <el-input v-model="container.command" placeholder="/bin/sh -c" />
              </el-form-item>
              <el-form-item label="命令参数 (Args)">
                <el-input v-model="container.args" placeholder="arg1 arg2" />
              </el-form-item>
            </div>

            <!-- Ports -->
            <el-divider content-position="left">端口</el-divider>
            <div v-for="(port, pi) in container.ports" :key="pi" class="kv-row">
              <el-input v-model="port.name" placeholder="名称" style="width: 120px;" />
              <el-input-number v-model="port.containerPort" :min="1" :max="65535" placeholder="端口" style="flex: 1;" />
              <el-select v-model="port.protocol" style="width: 100px;">
                <el-option label="TCP" value="TCP" /><el-option label="UDP" value="UDP" />
              </el-select>
              <el-button type="danger" text circle @click="removePort(ci, pi)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <el-button text type="primary" size="small" @click="addPort(ci)">
              <el-icon><Plus /></el-icon> 添加端口
            </el-button>

            <!-- Env -->
            <el-divider content-position="left">环境变量</el-divider>
            <div v-for="(env, ei) in container.env" :key="ei" class="env-row">
              <el-input v-model="env.name" placeholder="名称" style="width: 140px;" />
              <el-select v-model="env.type" style="width: 130px;" @change="() => { env.value = ''; env.configMapName = ''; env.configMapKey = ''; env.secretName = ''; env.secretKey = ''; env.fieldPath = '' }">
                <el-option label="直接值" value="plain" />
                <el-option label="ConfigMap" value="configMapKeyRef" />
                <el-option label="Secret" value="secretKeyRef" />
                <el-option label="字段引用" value="fieldRef" />
              </el-select>
              <el-input v-if="env.type === 'plain'" v-model="env.value" placeholder="值" style="flex: 1;" />
              <template v-if="env.type === 'configMapKeyRef'">
                <el-input v-model="env.configMapName" placeholder="ConfigMap名" style="flex: 1;" />
                <el-input v-model="env.configMapKey" placeholder="Key" style="width: 120px;" />
              </template>
              <template v-if="env.type === 'secretKeyRef'">
                <el-input v-model="env.secretName" placeholder="Secret名" style="flex: 1;" />
                <el-input v-model="env.secretKey" placeholder="Key" style="width: 120px;" />
              </template>
              <el-input v-if="env.type === 'fieldRef'" v-model="env.fieldPath" placeholder="metadata.name" style="flex: 1;" />
              <el-button type="danger" text circle @click="removeEnv(ci, ei)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <el-button text type="primary" size="small" @click="addEnv(ci)">
              <el-icon><Plus /></el-icon> 添加环境变量
            </el-button>

            <!-- Resources -->
            <el-divider content-position="left">资源配额</el-divider>
            <div class="resources-grid">
              <div class="resource-group">
                <div class="resource-group-title">Requests</div>
                <div class="resource-fields">
                  <el-form-item label="CPU"><el-input v-model="container.resources.requests.cpu" placeholder="100m" /></el-form-item>
                  <el-form-item label="Memory"><el-input v-model="container.resources.requests.memory" placeholder="128Mi" /></el-form-item>
                </div>
              </div>
              <div class="resource-group">
                <div class="resource-group-title">Limits</div>
                <div class="resource-fields">
                  <el-form-item label="CPU"><el-input v-model="container.resources.limits.cpu" placeholder="500m" /></el-form-item>
                  <el-form-item label="Memory"><el-input v-model="container.resources.limits.memory" placeholder="512Mi" /></el-form-item>
                </div>
              </div>
            </div>
          </div>
          <el-button text type="primary" @click="addContainer" class="add-container-btn">
            <el-icon><Plus /></el-icon> 添加容器
          </el-button>
        </div>
      </div>

      <!-- Section: Init Containers -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">初始化容器</div>
        </div>
        <div class="section-content">
          <div v-for="(container, ci) in form.initContainers" :key="ci" class="container-card">
            <div class="container-card-header">
              <div class="container-title">
                <span class="container-index">{{ ci + 1 }}</span>
                <span>{{ container.name || '未命名容器' }}</span>
                <el-tag size="small" type="warning">Init</el-tag>
              </div>
              <el-button type="danger" text size="small" @click="removeInitContainer(ci)">
                <el-icon><Delete /></el-icon> 移除
              </el-button>
            </div>
            <div class="fields-grid">
              <el-form-item label="容器名称" required>
                <el-input v-model="container.name" placeholder="init-container" />
              </el-form-item>
              <el-form-item label="镜像" required>
                <el-input v-model="container.image" placeholder="busybox:1.36" />
              </el-form-item>
              <el-form-item label="拉取策略">
                <el-select v-model="container.imagePullPolicy" style="width: 100%;">
                  <el-option label="Always" value="Always" />
                  <el-option label="IfNotPresent" value="IfNotPresent" />
                  <el-option label="Never" value="Never" />
                </el-select>
              </el-form-item>
              <el-form-item label="启动命令 (Command)">
                <el-input v-model="container.command" placeholder="/bin/sh -c" />
              </el-form-item>
              <el-form-item label="命令参数 (Args)">
                <el-input v-model="container.args" placeholder="arg1 arg2" />
              </el-form-item>
            </div>
            <el-divider content-position="left">环境变量</el-divider>
            <div v-for="(env, ei) in container.env" :key="ei" class="env-row">
              <el-input v-model="env.name" placeholder="名称" style="width: 140px;" />
              <el-select v-model="env.type" style="width: 130px;" @change="() => { env.value = ''; env.configMapName = ''; env.configMapKey = ''; env.secretName = ''; env.secretKey = ''; env.fieldPath = '' }">
                <el-option label="直接值" value="plain" />
                <el-option label="ConfigMap" value="configMapKeyRef" />
                <el-option label="Secret" value="secretKeyRef" />
                <el-option label="字段引用" value="fieldRef" />
              </el-select>
              <el-input v-if="env.type === 'plain'" v-model="env.value" placeholder="值" style="flex: 1;" />
              <template v-if="env.type === 'configMapKeyRef'">
                <el-input v-model="env.configMapName" placeholder="ConfigMap名" style="flex: 1;" />
                <el-input v-model="env.configMapKey" placeholder="Key" style="width: 120px;" />
              </template>
              <template v-if="env.type === 'secretKeyRef'">
                <el-input v-model="env.secretName" placeholder="Secret名" style="flex: 1;" />
                <el-input v-model="env.secretKey" placeholder="Key" style="width: 120px;" />
              </template>
              <el-input v-if="env.type === 'fieldRef'" v-model="env.fieldPath" placeholder="metadata.name" style="flex: 1;" />
              <el-button type="danger" text circle @click="removeEnv(ci, ei, true)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <el-button text type="primary" size="small" @click="addEnv(ci, true)">
              <el-icon><Plus /></el-icon> 添加环境变量
            </el-button>
            <el-divider content-position="left">资源配额</el-divider>
            <div class="resources-grid">
              <div class="resource-group">
                <div class="resource-group-title">Requests</div>
                <div class="resource-fields">
                  <el-form-item label="CPU"><el-input v-model="container.resources.requests.cpu" placeholder="100m" /></el-form-item>
                  <el-form-item label="Memory"><el-input v-model="container.resources.requests.memory" placeholder="128Mi" /></el-form-item>
                </div>
              </div>
              <div class="resource-group">
                <div class="resource-group-title">Limits</div>
                <div class="resource-fields">
                  <el-form-item label="CPU"><el-input v-model="container.resources.limits.cpu" placeholder="500m" /></el-form-item>
                  <el-form-item label="Memory"><el-input v-model="container.resources.limits.memory" placeholder="512Mi" /></el-form-item>
                </div>
              </div>
            </div>
          </div>
          <el-button text type="primary" @click="addInitContainer" class="add-container-btn">
            <el-icon><Plus /></el-icon> 添加初始化容器
          </el-button>
        </div>
      </div>

      <!-- Section 4: Storage -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">存储配置</div>
        </div>
        <div class="section-content">

          <el-form-item label="数据卷">
            <div style="width: 100%;">
              <div v-for="(vol, vi) in form.volumes" :key="vi" class="volume-card">
                <div class="volume-row">
                  <el-input v-model="vol.name" placeholder="卷名称" />
                  <el-select v-model="vol.type" style="width: 160px;">
                    <el-option label="emptyDir" value="emptyDir" />
                    <el-option label="hostPath" value="hostPath" />
                    <el-option label="ConfigMap" value="configMap" />
                    <el-option label="Secret" value="secret" />
                    <el-option label="PVC" value="pvc" />
                  </el-select>
                  <el-button type="danger" text circle @click="removeVolume(vi)">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
                <el-input v-if="vol.type === 'hostPath'" v-model="vol.hostPath" placeholder="主机路径 (e.g. /data)" style="margin-top: 8px;" />
                <el-select v-if="vol.type === 'hostPath'" v-model="vol.hostPathType" style="width: 100%; margin-top: 8px;">
                  <el-option label="DirectoryOrCreate" value="DirectoryOrCreate" />
                  <el-option label="Directory" value="Directory" />
                  <el-option label="FileOrCreate" value="FileOrCreate" />
                  <el-option label="File" value="File" />
                  <el-option label="Socket" value="Socket" />
                  <el-option label="CharDevice" value="CharDevice" />
                  <el-option label="BlockDevice" value="BlockDevice" />
                </el-select>
                <el-input v-if="vol.type === 'configMap'" v-model="vol.configMapName" placeholder="ConfigMap 名称" style="margin-top: 8px;" />
                <el-input v-if="vol.type === 'secret'" v-model="vol.secretName" placeholder="Secret 名称" style="margin-top: 8px;" />
                <el-input v-if="vol.type === 'pvc'" v-model="vol.pvcName" placeholder="PVC 名称" style="margin-top: 8px;" />
              </div>
              <el-button text type="primary" @click="addVolume" size="small">
                <el-icon><Plus /></el-icon> 添加数据卷
              </el-button>
            </div>
          </el-form-item>

          <el-divider v-if="form.volumes.length > 0" />

          <el-form-item v-if="form.volumes.length > 0" label="卷挂载">
            <div style="width: 100%;">
              <div v-for="(container, ci) in form.containers" :key="ci" style="margin-bottom: 16px;">
                <div class="mount-container-name">{{ container.name || `容器 ${ci + 1}` }}</div>
                <div v-for="(mount, mi) in container.volumeMounts" :key="mi" class="kv-row">
                  <el-select v-model="mount.name" placeholder="选择卷" style="width: 160px;">
                    <el-option v-for="v in form.volumes.filter(v => v.name)" :key="v.name" :label="v.name" :value="v.name" />
                  </el-select>
                  <el-input v-model="mount.mountPath" placeholder="挂载路径" />
                  <el-input v-model="mount.subPath" placeholder="子路径" style="width: 120px;" />
                  <el-checkbox v-model="mount.readOnly">只读</el-checkbox>
                  <el-button type="danger" text circle @click="removeVolumeMount(ci, mi)">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
                <el-button text type="primary" size="small" @click="addVolumeMount(ci)">
                  <el-icon><Plus /></el-icon> 添加挂载
                </el-button>
              </div>
            </div>
          </el-form-item>
        </div>
      </div>

      <!-- Section 5: Health Probes -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">健康检查</div>
        </div>
        <div class="section-content">

          <div v-for="(container, ci) in form.containers" :key="ci" style="margin-bottom: 24px;">
            <div class="mount-container-name">{{ container.name || `容器 ${ci + 1}` }}</div>

            <!-- Liveness -->
            <div class="probe-card">
              <div class="probe-header">
                <div>
                  <span class="probe-label">存活探针</span>
                  <span class="probe-desc">容器是否正在运行</span>
                </div>
                <el-switch :model-value="!!container.livenessProbe" @update:model-value="(v: boolean) => v ? enableProbe(ci, 'livenessProbe') : disableProbe(ci, 'livenessProbe')" />
              </div>
              <template v-if="container.livenessProbe">
                <div class="fields-grid" style="margin-top: 16px;">
                  <el-form-item label="检测类型">
                    <el-select v-model="container.livenessProbe.type" style="width: 100%;">
                      <el-option label="HTTP GET" value="httpGet" />
                      <el-option label="TCP Socket" value="tcpSocket" />
                      <el-option label="Exec" value="exec" />
                    </el-select>
                  </el-form-item>
                  <el-form-item v-if="container.livenessProbe.type === 'httpGet'" label="路径">
                    <el-input v-model="container.livenessProbe.httpGetPath" placeholder="/" />
                  </el-form-item>
                  <el-form-item v-if="container.livenessProbe.type === 'httpGet'" label="端口">
                    <el-input-number v-model="container.livenessProbe.httpGetPort" :min="1" :max="65535" style="width: 100%;" />
                  </el-form-item>
                  <el-form-item v-if="container.livenessProbe.type === 'tcpSocket'" label="端口">
                    <el-input-number v-model="container.livenessProbe.tcpSocketPort" :min="1" :max="65535" style="width: 100%;" />
                  </el-form-item>
                  <el-form-item v-if="container.livenessProbe.type === 'exec'" label="命令">
                    <el-input v-model="container.livenessProbe.execCommand" placeholder="cat /tmp/healthy" />
                  </el-form-item>
                  <el-form-item label="初始延迟(秒)">
                    <el-input-number v-model="container.livenessProbe.initialDelaySeconds" :min="0" style="width: 100%;" />
                  </el-form-item>
                  <el-form-item label="检测周期(秒)">
                    <el-input-number v-model="container.livenessProbe.periodSeconds" :min="1" style="width: 100%;" />
                  </el-form-item>
                </div>
              </template>
            </div>

            <!-- Readiness -->
            <div class="probe-card">
              <div class="probe-header">
                <div>
                  <span class="probe-label">就绪探针</span>
                  <span class="probe-desc">容器是否准备好接收流量</span>
                </div>
                <el-switch :model-value="!!container.readinessProbe" @update:model-value="(v: boolean) => v ? enableProbe(ci, 'readinessProbe') : disableProbe(ci, 'readinessProbe')" />
              </div>
              <template v-if="container.readinessProbe">
                <div class="fields-grid" style="margin-top: 16px;">
                  <el-form-item label="检测类型">
                    <el-select v-model="container.readinessProbe.type" style="width: 100%;">
                      <el-option label="HTTP GET" value="httpGet" />
                      <el-option label="TCP Socket" value="tcpSocket" />
                      <el-option label="Exec" value="exec" />
                    </el-select>
                  </el-form-item>
                  <el-form-item v-if="container.readinessProbe.type === 'httpGet'" label="路径">
                    <el-input v-model="container.readinessProbe.httpGetPath" placeholder="/" />
                  </el-form-item>
                  <el-form-item v-if="container.readinessProbe.type === 'httpGet'" label="端口">
                    <el-input-number v-model="container.readinessProbe.httpGetPort" :min="1" :max="65535" style="width: 100%;" />
                  </el-form-item>
                  <el-form-item v-if="container.readinessProbe.type === 'tcpSocket'" label="端口">
                    <el-input-number v-model="container.readinessProbe.tcpSocketPort" :min="1" :max="65535" style="width: 100%;" />
                  </el-form-item>
                  <el-form-item v-if="container.readinessProbe.type === 'exec'" label="命令">
                    <el-input v-model="container.readinessProbe.execCommand" placeholder="cat /tmp/healthy" />
                  </el-form-item>
                  <el-form-item label="初始延迟(秒)">
                    <el-input-number v-model="container.readinessProbe.initialDelaySeconds" :min="0" style="width: 100%;" />
                  </el-form-item>
                  <el-form-item label="检测周期(秒)">
                    <el-input-number v-model="container.readinessProbe.periodSeconds" :min="1" style="width: 100%;" />
                  </el-form-item>
                </div>
              </template>
            </div>

            <!-- Startup -->
            <div class="probe-card">
              <div class="probe-header">
                <div>
                  <span class="probe-label">启动探针</span>
                  <span class="probe-desc">容器是否已启动完成</span>
                </div>
                <el-switch :model-value="!!container.startupProbe" @update:model-value="(v: boolean) => v ? enableProbe(ci, 'startupProbe') : disableProbe(ci, 'startupProbe')" />
              </div>
              <template v-if="container.startupProbe">
                <div class="fields-grid" style="margin-top: 16px;">
                  <el-form-item label="检测类型">
                    <el-select v-model="container.startupProbe.type" style="width: 100%;">
                      <el-option label="HTTP GET" value="httpGet" />
                      <el-option label="TCP Socket" value="tcpSocket" />
                      <el-option label="Exec" value="exec" />
                    </el-select>
                  </el-form-item>
                  <el-form-item v-if="container.startupProbe.type === 'httpGet'" label="路径">
                    <el-input v-model="container.startupProbe.httpGetPath" placeholder="/" />
                  </el-form-item>
                  <el-form-item v-if="container.startupProbe.type === 'httpGet'" label="端口">
                    <el-input-number v-model="container.startupProbe.httpGetPort" :min="1" :max="65535" style="width: 100%;" />
                  </el-form-item>
                  <el-form-item v-if="container.startupProbe.type === 'tcpSocket'" label="端口">
                    <el-input-number v-model="container.startupProbe.tcpSocketPort" :min="1" :max="65535" style="width: 100%;" />
                  </el-form-item>
                  <el-form-item v-if="container.startupProbe.type === 'exec'" label="命令">
                    <el-input v-model="container.startupProbe.execCommand" placeholder="cat /tmp/healthy" />
                  </el-form-item>
                  <el-form-item label="初始延迟(秒)">
                    <el-input-number v-model="container.startupProbe.initialDelaySeconds" :min="0" style="width: 100%;" />
                  </el-form-item>
                  <el-form-item label="检测周期(秒)">
                    <el-input-number v-model="container.startupProbe.periodSeconds" :min="1" style="width: 100%;" />
                  </el-form-item>
                  <el-form-item label="失败阈值">
                    <el-input-number v-model="container.startupProbe.failureThreshold" :min="1" style="width: 100%;" />
                  </el-form-item>
                </div>
              </template>
            </div>
          </div>
        </div>
      </div>

      <!-- Section 6: Security -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">安全设置</div>
        </div>
        <div class="section-content">

          <div v-for="(container, ci) in form.containers" :key="ci" style="margin-bottom: 24px;">
            <div class="mount-container-name">{{ container.name || `容器 ${ci + 1}` }}</div>
            <div class="security-grid">
              <div class="security-item">
                <div class="security-item-label">运行用户 ID</div>
                <el-input-number v-model="container.securityContext.runAsUser" :min="0" placeholder="UID" style="width: 100%;" />
              </div>
              <div class="security-item">
                <div class="security-item-label">非 Root 运行</div>
                <el-switch v-model="container.securityContext.runAsNonRoot" />
              </div>
              <div class="security-item">
                <div class="security-item-label">只读根文件系统</div>
                <el-switch v-model="container.securityContext.readOnlyRootFilesystem" />
              </div>
              <div class="security-item">
                <div class="security-item-label">特权模式</div>
                <el-switch v-model="container.securityContext.privileged" />
              </div>
            </div>
            <el-divider content-position="left">Linux Capabilities</el-divider>
            <div class="capabilities-section">
              <div class="cap-group">
                <div class="cap-group-title">添加 (Add)</div>
                <div v-for="(_cap, cai) in container.securityContext.capabilitiesAdd" :key="cai" class="cap-row">
                  <el-select v-model="container.securityContext.capabilitiesAdd[cai]" filterable allow-create placeholder="选择 capability" style="flex: 1;">
                    <el-option label="SYS_ADMIN" value="SYS_ADMIN" /><el-option label="NET_ADMIN" value="NET_ADMIN" />
                    <el-option label="SYS_TIME" value="SYS_TIME" /><el-option label="SYS_PTRACE" value="SYS_PTRACE" />
                    <el-option label="NET_RAW" value="NET_RAW" /><el-option label="CHOWN" value="CHOWN" />
                    <el-option label="DAC_OVERRIDE" value="DAC_OVERRIDE" /><el-option label="FOWNER" value="FOWNER" />
                    <el-option label="SETUID" value="SETUID" /><el-option label="SETGID" value="SETGID" />
                  </el-select>
                  <el-button type="danger" text circle @click="removeCapability(ci, 'capabilitiesAdd', cai)"><el-icon><Delete /></el-icon></el-button>
                </div>
                <el-button text type="primary" size="small" @click="addCapability(ci, 'capabilitiesAdd')"><el-icon><Plus /></el-icon> 添加</el-button>
              </div>
              <div class="cap-group">
                <div class="cap-group-title">移除 (Drop)</div>
                <div v-for="(_cap, cdi) in container.securityContext.capabilitiesDrop" :key="cdi" class="cap-row">
                  <el-select v-model="container.securityContext.capabilitiesDrop[cdi]" filterable allow-create placeholder="选择 capability" style="flex: 1;">
                    <el-option label="ALL" value="ALL" /><el-option label="SYS_ADMIN" value="SYS_ADMIN" />
                    <el-option label="NET_ADMIN" value="NET_ADMIN" /><el-option label="NET_RAW" value="NET_RAW" />
                    <el-option label="CHOWN" value="CHOWN" /><el-option label="DAC_OVERRIDE" value="DAC_OVERRIDE" />
                    <el-option label="FOWNER" value="FOWNER" /><el-option label="SETUID" value="SETUID" /><el-option label="SETGID" value="SETGID" />
                  </el-select>
                  <el-button type="danger" text circle @click="removeCapability(ci, 'capabilitiesDrop', cdi)"><el-icon><Delete /></el-icon></el-button>
                </div>
                <el-button text type="primary" size="small" @click="addCapability(ci, 'capabilitiesDrop')"><el-icon><Plus /></el-icon> 添加</el-button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Section 7: Scheduling -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">调度配置</div>
        </div>
        <div class="section-content">

          <el-form-item label="节点选择器">
            <div style="width: 100%;">
              <div v-for="(ns, i) in form.nodeSelector" :key="i" class="kv-row">
                <el-input v-model="ns.key" placeholder="Key" />
                <el-input v-model="ns.value" placeholder="Value" />
                <el-button type="danger" text circle @click="removeNodeSelector(i)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
              <el-button text type="primary" @click="addNodeSelector" size="small">
                <el-icon><Plus /></el-icon> 添加节点选择器
              </el-button>
            </div>
          </el-form-item>

          <el-form-item label="容忍规则">
            <div style="width: 100%;">
              <div v-for="(tol, i) in form.tolerations" :key="i" class="toleration-row">
                <el-input v-model="tol.key" placeholder="Key" />
                <el-select v-model="tol.operator" style="width: 100px;">
                  <el-option label="Equal" value="Equal" /><el-option label="Exists" value="Exists" />
                </el-select>
                <el-input v-model="tol.value" placeholder="Value" />
                <el-select v-model="tol.effect" style="width: 150px;">
                  <el-option label="NoSchedule" value="NoSchedule" />
                  <el-option label="PreferNoSchedule" value="PreferNoSchedule" />
                  <el-option label="NoExecute" value="NoExecute" />
                </el-select>
                <el-button type="danger" text circle @click="removeToleration(i)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
              <el-button text type="primary" @click="addToleration" size="small">
                <el-icon><Plus /></el-icon> 添加容忍规则
              </el-button>
            </div>
          </el-form-item>

          <!-- Pod Affinity -->
          <el-divider content-position="left">Pod 亲和性</el-divider>
          <div class="affinity-section">
            <div class="affinity-section-title">亲和规则</div>
            <div v-for="(rule, i) in form.podAffinityRules" :key="i" class="affinity-row">
              <el-input-number v-model="rule.weight" :min="0" :max="100" placeholder="权重" style="width: 100px;" />
              <el-input v-model="rule.topologyKey" placeholder="topologyKey" style="flex: 1;" />
              <el-input v-model="rule.labelKey" placeholder="标签Key" style="flex: 1;" />
              <el-input v-model="rule.labelValue" placeholder="标签Value" style="flex: 1;" />
              <el-input v-model="rule.namespaces" placeholder="命名空间(逗号分隔)" style="flex: 1;" />
              <el-button type="danger" text circle @click="removeAffinityRule('podAffinityRules', i)"><el-icon><Delete /></el-icon></el-button>
            </div>
            <el-button text type="primary" size="small" @click="addAffinityRule('podAffinityRules')"><el-icon><Plus /></el-icon> 添加亲和规则</el-button>
            <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-top: 4px;">权重 0 = 必须满足 (required)，1-100 = 优先满足 (preferred)</div>
          </div>
          <div class="affinity-section" style="margin-top: 16px;">
            <div class="affinity-section-title">反亲和规则</div>
            <div v-for="(rule, i) in form.podAntiAffinityRules" :key="i" class="affinity-row">
              <el-input-number v-model="rule.weight" :min="0" :max="100" placeholder="权重" style="width: 100px;" />
              <el-input v-model="rule.topologyKey" placeholder="topologyKey" style="flex: 1;" />
              <el-input v-model="rule.labelKey" placeholder="标签Key" style="flex: 1;" />
              <el-input v-model="rule.labelValue" placeholder="标签Value" style="flex: 1;" />
              <el-input v-model="rule.namespaces" placeholder="命名空间(逗号分隔)" style="flex: 1;" />
              <el-button type="danger" text circle @click="removeAffinityRule('podAntiAffinityRules', i)"><el-icon><Delete /></el-icon></el-button>
            </div>
            <el-button text type="primary" size="small" @click="addAffinityRule('podAntiAffinityRules')"><el-icon><Plus /></el-icon> 添加反亲和规则</el-button>
          </div>

          <!-- Topology Spread Constraints -->
          <el-divider content-position="left">拓扑分布约束</el-divider>
          <div v-for="(tsc, i) in form.topologySpreadConstraints" :key="i" class="topology-row">
            <el-input-number v-model="tsc.maxSkew" :min="1" placeholder="maxSkew" style="width: 100px;" />
            <el-input v-model="tsc.topologyKey" placeholder="topologyKey" style="flex: 1;" />
            <el-select v-model="tsc.whenUnsatisfiable" style="width: 160px;">
              <el-option label="DoNotSchedule" value="DoNotSchedule" />
              <el-option label="ScheduleAnyway" value="ScheduleAnyway" />
            </el-select>
            <el-input v-model="tsc.labelKey" placeholder="标签Key" style="flex: 1;" />
            <el-input v-model="tsc.labelValue" placeholder="标签Value" style="flex: 1;" />
            <el-button type="danger" text circle @click="removeTopologySpreadConstraint(i)"><el-icon><Delete /></el-icon></el-button>
          </div>
          <el-button text type="primary" size="small" @click="addTopologySpreadConstraint"><el-icon><Plus /></el-icon> 添加拓扑分布约束</el-button>
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
            <el-button @click="handleCancel">取消</el-button>
            <el-button type="primary" :loading="submitting" @click="handleSubmit">创建</el-button>
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

.form-help {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
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

.fields-grid :deep(.el-form-item.full-width) {
  grid-column: 1 / -1;
}

/* Container cards */
.container-card {
  border: 1px solid var(--el-border-color-extra-light);
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 12px;
  background: var(--el-fill-color-blank);
  transition: box-shadow 0.2s;
}

.container-card:hover {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);
}

.container-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--el-border-color-extra-light);
}

.container-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 600;
  font-size: 15px;
  color: var(--el-text-color-primary);
}

.container-index {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: var(--el-color-primary);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 600;
}

.add-container-btn {
  margin-top: 8px;
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

/* Resources */
.resources-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.resource-group {
  border: 1px solid var(--el-border-color-extra-light);
  border-radius: 8px;
  padding: 14px;
  background: var(--el-fill-color-lighter);
}

.resource-group-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-regular);
  margin-bottom: 12px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.resource-fields {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.resource-fields :deep(.el-form-item) {
  margin-bottom: 0;
}

/* Volume cards */
.volume-card {
  border: 1px solid var(--el-border-color-extra-light);
  border-radius: 8px;
  padding: 14px;
  margin-bottom: 8px;
  background: var(--el-fill-color-lighter);
}

.volume-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.volume-row :deep(.el-input) {
  flex: 1;
}

.mount-container-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-regular);
  margin-bottom: 10px;
  padding: 4px 10px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
  display: inline-block;
}

/* Probe cards */
.probe-card {
  border: 1px solid var(--el-border-color-extra-light);
  border-radius: 8px;
  padding: 14px;
  margin-bottom: 10px;
  background: var(--el-fill-color-blank);
}

.probe-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.probe-label {
  font-weight: 600;
  font-size: 14px;
  color: var(--el-text-color-primary);
}

.probe-desc {
  display: block;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 2px;
}

/* Security grid */
.security-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.security-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border: 1px solid var(--el-border-color-extra-light);
  border-radius: 8px;
  background: var(--el-fill-color-lighter);
}

.security-item-label {
  font-size: 14px;
  color: var(--el-text-color-regular);
}

/* Toleration row */
.toleration-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: center;
}

.toleration-row :deep(.el-input) {
  flex: 1;
}

/* Env row */
.env-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: center;
}

/* Capabilities */
.capabilities-section {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  margin-top: 12px;
}

.cap-group {
  border: 1px solid var(--el-border-color-extra-light);
  border-radius: 8px;
  padding: 12px;
  background: var(--el-fill-color-lighter);
}

.cap-group-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-regular);
  margin-bottom: 10px;
}

.cap-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: center;
}

/* Affinity */
.affinity-section {
  margin-top: 8px;
}

.affinity-section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin-bottom: 12px;
}

.affinity-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: center;
}

.topology-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: center;
}

</style>
