<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import yaml from 'js-yaml'
import type { FormInstance, FormRules } from 'element-plus'
import { getNamespaceList, extractNamespaceNames } from '@/api/resource'
import { createDeployment, createStatefulSet, createDaemonSet, updateDeploymentYaml, updateStatefulSetYaml, updateDaemonSetYaml } from '@/api/resource'
import SchedulingForm from './form/SchedulingForm.vue'
import LabelsAnnotationsForm from './form/LabelsAnnotationsForm.vue'
import ContainerConfigForm from './form/ContainerConfigForm.vue'
import StorageConfigForm from './form/StorageConfigForm.vue'
import HealthCheckForm from './HealthCheckForm.vue'
import SecurityContextForm from './form/SecurityContextForm.vue'
import type { WorkloadFormData, Container, Probe, LifecycleHandler, EnvVar, AffinityRule } from './form-types'
import { createEmptyContainer } from './form-types'

const props = withDefaults(defineProps<{
  kind: 'Deployment' | 'StatefulSet' | 'DaemonSet'
  isEdit?: boolean
  initialData?: any
  onSubmit?: (yaml: string) => Promise<void>
}>(), {
  isEdit: false,
  initialData: undefined,
  onSubmit: undefined,
})

const emit = defineEmits<{
  success: []
  cancel: []
}>()

const router = useRouter()
const submitting = ref(false)
const namespaceLoading = ref(false)
const namespaces = ref<string[]>([])
const { t } = useI18n()

const form = reactive<WorkloadFormData>({
  name: '', namespace: 'default', replicas: 1,
  labels: [{ key: 'app', value: '' }],
  containers: [createEmptyContainer()],
  initContainers: [],
  volumes: [],
  strategyType: 'RollingUpdate', maxSurge: '25%', maxUnavailable: '25%',
  serviceName: '', updateStrategy: 'RollingUpdate', dsUpdateStrategy: 'RollingUpdate',
  podManagementPolicy: 'OrderedReady',
  nodeSelector: [], tolerations: [], annotations: [],
  serviceAccountName: '', terminationGracePeriodSeconds: null, imagePullSecrets: [],
  volumeClaimTemplates: [],
  podAffinityRules: [], podAntiAffinityRules: [],
  topologySpreadConstraints: [],
  dnsPolicy: 'ClusterFirst', hostNetwork: false, priorityClassName: '',
})

const formRef = ref<FormInstance>()
const formRules: FormRules = {
  name: [
    { required: true, message: '请输入名称', trigger: 'blur' },
    { pattern: /^[a-z][a-z0-9-]*[a-z0-9]$/, message: '仅支持小写字母、数字和连字符', trigger: 'blur' },
  ],
  namespace: [{ required: true, message: '请选择命名空间', trigger: 'change' }],
  replicas: [{ required: true, message: '请输入副本数', trigger: 'change' }],
}

async function fetchNamespaces() {
  namespaceLoading.value = true
  try {
    const res: any = await getNamespaceList()
    namespaces.value = extractNamespaceNames(res.data)
  } catch { namespaces.value = ['default'] }
  finally { namespaceLoading.value = false }
}

function parseProbe(probeData: any): Probe | null {
  if (!probeData) return null
  return {
    type: probeData.httpGet ? 'httpGet' : probeData.tcpSocket ? 'tcpSocket' : 'exec',
    httpGetPath: probeData.httpGet?.path ?? '/',
    httpGetPort: probeData.httpGet?.port ?? 80,
    tcpSocketPort: probeData.tcpSocket?.port ?? null,
    execCommand: probeData.exec?.command?.join('\n') ?? '',
    initialDelaySeconds: probeData.initialDelaySeconds ?? 15,
    periodSeconds: probeData.periodSeconds ?? 10,
    timeoutSeconds: probeData.timeoutSeconds ?? 5,
    failureThreshold: probeData.failureThreshold ?? 3,
  }
}

function parseLifecycleHandler(data: any): LifecycleHandler | null {
  if (!data) return null
  return {
    type: data.exec ? 'exec' : data.httpGet ? 'httpGet' : 'tcpSocket',
    execCommand: data.exec?.command?.join('\n') || '',
    httpGetPath: data.httpGet?.path || '/',
    httpGetPort: data.httpGet?.port || 80,
    tcpSocketPort: data.tcpSocket?.port || null,
  }
}

function parseEnvVar(e: any): EnvVar {
  if (e.valueFrom?.configMapKeyRef) {
    return { name: e.name, value: '', type: 'configMapKeyRef', configMapName: e.valueFrom.configMapKeyRef.name || '', configMapKey: e.valueFrom.configMapKeyRef.key || '', secretName: '', secretKey: '', fieldPath: '' }
  }
  if (e.valueFrom?.secretKeyRef) {
    return { name: e.name, value: '', type: 'secretKeyRef', configMapName: '', configMapKey: '', secretName: e.valueFrom.secretKeyRef.name || '', secretKey: e.valueFrom.secretKeyRef.key || '', fieldPath: '' }
  }
  if (e.valueFrom?.fieldRef) {
    return { name: e.name, value: '', type: 'fieldRef', configMapName: '', configMapKey: '', secretName: '', secretKey: '', fieldPath: e.valueFrom.fieldRef.fieldPath || '' }
  }
  return { name: e.name, value: e.value || '', type: 'plain', configMapName: '', configMapKey: '', secretName: '', secretKey: '', fieldPath: '' }
}

function parseContainer(c: any): Container {
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

function parseInitialData(data: any) {
  if (!data) return

  const metadata = data.metadata || {}
  const spec = data.spec || {}
  const template = spec.template || {}
  const podSpec = template.spec || {}

  // Basic info
  form.name = metadata.name || ''
  form.namespace = metadata.namespace || 'default'
  form.replicas = spec.replicas || 1

  // Labels
  const labels = metadata.labels || {}
  form.labels = Object.entries(labels).map(([key, value]) => ({ key, value: value as string }))
  if (form.labels.length === 0) form.labels.push({ key: 'app', value: '' })

  // Annotations
  const annotations = template.metadata?.annotations || {}
  form.annotations = Object.entries(annotations).map(([key, value]) => ({ key, value: value as string }))

  // Containers
  const containers = podSpec.containers || []
  form.containers = containers.map(parseContainer)
  if (form.containers.length === 0) form.containers.push(createEmptyContainer())

  // Init Containers
  const initContainers = podSpec.initContainers || []
  form.initContainers = initContainers.map(parseContainer)

  // Volumes
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

  // Node selector
  const nodeSelector = podSpec.nodeSelector || {}
  form.nodeSelector = Object.entries(nodeSelector).map(([key, value]) => ({ key, value: value as string }))

  // Tolerations
  const tolerations = podSpec.tolerations || []
  form.tolerations = tolerations.map((t: any) => ({
    key: t.key || '',
    operator: t.operator || 'Equal',
    value: t.value || '',
    effect: t.effect || 'NoSchedule',
    tolerationSeconds: t.tolerationSeconds || null,
  }))

  // Service account
  form.serviceAccountName = podSpec.serviceAccountName || ''

  // Termination grace period
  form.terminationGracePeriodSeconds = podSpec.terminationGracePeriodSeconds || null

  // DNS policy & host network
  form.dnsPolicy = podSpec.dnsPolicy || 'ClusterFirst'
  form.hostNetwork = podSpec.hostNetwork || false
  form.priorityClassName = podSpec.priorityClassName || ''

  // Image pull secrets
  const imagePullSecrets = podSpec.imagePullSecrets || []
  form.imagePullSecrets = imagePullSecrets.map((s: any) => s.name || '')

  // Pod Affinity
  const affinity = podSpec.affinity || {}
  const parseAffinityRules = (rules: any[]): AffinityRule[] => {
    return (rules || []).map((r: any) => ({
      weight: r.weight || 1,
      topologyKey: r.podAffinityTerm?.topologyKey || r.topologyKey || '',
      namespaces: r.podAffinityTerm?.namespaces?.join(', ') || r.namespaces?.join(', ') || '',
      labelKey: r.podAffinityTerm?.labelSelector?.matchLabels ? Object.keys(r.podAffinityTerm.labelSelector.matchLabels)[0] || '' : r.labelSelector?.matchLabels ? Object.keys(r.labelSelector.matchLabels)[0] || '' : '',
      labelValue: r.podAffinityTerm?.labelSelector?.matchLabels ? Object.values(r.podAffinityTerm.labelSelector.matchLabels)[0] as string || '' : r.labelSelector?.matchLabels ? Object.values(r.labelSelector.matchLabels)[0] as string || '' : '',
    }))
  }
  form.podAffinityRules = parseAffinityRules(affinity.podAffinity?.preferredDuringSchedulingIgnoredDuringExecution || [])
  if (affinity.podAffinity?.requiredDuringSchedulingIgnoredDuringExecution) {
    form.podAffinityRules.push(...affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution.map((r: any) => ({
      weight: 0, topologyKey: r.topologyKey || '', namespaces: r.namespaces?.join(', ') || '',
      labelKey: r.labelSelector?.matchLabels ? Object.keys(r.labelSelector.matchLabels)[0] || '' : '',
      labelValue: r.labelSelector?.matchLabels ? Object.values(r.labelSelector.matchLabels)[0] as string || '' : '',
    })))
  }
  form.podAntiAffinityRules = parseAffinityRules(affinity.podAntiAffinity?.preferredDuringSchedulingIgnoredDuringExecution || [])
  if (affinity.podAntiAffinity?.requiredDuringSchedulingIgnoredDuringExecution) {
    form.podAntiAffinityRules.push(...affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution.map((r: any) => ({
      weight: 0, topologyKey: r.topologyKey || '', namespaces: r.namespaces?.join(', ') || '',
      labelKey: r.labelSelector?.matchLabels ? Object.keys(r.labelSelector.matchLabels)[0] || '' : '',
      labelValue: r.labelSelector?.matchLabels ? Object.values(r.labelSelector.matchLabels)[0] as string || '' : '',
    })))
  }

  // Topology Spread Constraints
  form.topologySpreadConstraints = (podSpec.topologySpreadConstraints || []).map((t: any) => ({
    maxSkew: t.maxSkew || 1,
    topologyKey: t.topologyKey || '',
    whenUnsatisfiable: t.whenUnsatisfiable || 'DoNotSchedule',
    labelKey: t.labelSelector?.matchLabels ? Object.keys(t.labelSelector.matchLabels)[0] || '' : '',
    labelValue: t.labelSelector?.matchLabels ? Object.values(t.labelSelector.matchLabels)[0] as string || '' : '',
  }))

  // Strategy (Deployment)
  if (props.kind === 'Deployment') {
    form.strategyType = spec.strategy?.type || 'RollingUpdate'
    form.maxSurge = spec.strategy?.rollingUpdate?.maxSurge?.toString() || '25%'
    form.maxUnavailable = spec.strategy?.rollingUpdate?.maxUnavailable?.toString() || '25%'
  }

  // Service name (StatefulSet)
  if (props.kind === 'StatefulSet') {
    form.serviceName = spec.serviceName || ''
    form.updateStrategy = spec.updateStrategy?.type || 'RollingUpdate'
    form.podManagementPolicy = spec.podManagementPolicy || 'OrderedReady'
  }

  // Update strategy (DaemonSet)
  if (props.kind === 'DaemonSet') {
    form.dsUpdateStrategy = spec.updateStrategy?.type || 'RollingUpdate'
  }

  // Volume claim templates (StatefulSet)
  if (props.kind === 'StatefulSet') {
    const vcts = spec.volumeClaimTemplates || []
    form.volumeClaimTemplates = vcts.map((v: any) => ({
      name: v.metadata?.name || '',
      storageSize: v.spec?.resources?.requests?.storage || '1Gi',
      storageClassName: v.spec?.storageClassName || '',
      accessModes: v.spec?.accessModes || ['ReadWriteOnce'],
    }))
  }
}

onMounted(() => {
  fetchNamespaces()
  if (props.isEdit && props.initialData) {
    parseInitialData(props.initialData)
  }
})

// 监听 initialData 变化：支持克隆场景下动态填充表单（克隆时 isEdit=false）
// 不用 deep：克隆是整体引用替换，浅 watch 即可检测，避免递归遍历大 K8s 对象
watch(
  () => props.initialData,
  (newData) => {
    if (newData) parseInitialData(newData)
  }
)

function addImagePullSecret() { form.imagePullSecrets.push('') }
function removeImagePullSecret(i: number) { form.imagePullSecrets.splice(i, 1) }

const generatedYaml = computed(() => yaml.dump(buildK8sResource(), { indent: 2, lineWidth: -1, noRefs: true }))

function buildProbe(probe: Probe | null): any {
  if (!probe) return undefined
  const p: any = { initialDelaySeconds: probe.initialDelaySeconds, periodSeconds: probe.periodSeconds, timeoutSeconds: probe.timeoutSeconds, failureThreshold: probe.failureThreshold }
  if (probe.type === 'httpGet') { p.httpGet = { path: probe.httpGetPath, port: probe.httpGetPort } }
  else if (probe.type === 'tcpSocket') { p.tcpSocket = { port: probe.tcpSocketPort } }
  else if (probe.type === 'exec') { p.exec = { command: probe.execCommand.split('\n').map(s => s.trim()).filter(Boolean) } }
  return p
}

function buildK8sResource(): Record<string, any> {
  const labels: Record<string, string> = {}
  form.labels.forEach(l => { if (l.key.trim()) labels[l.key.trim()] = l.value })

  // spec.selector 创建后不可变：编辑模式沿用初始 selector，避免改标签时 selector 随之变化被 K8s 拒绝（field is immutable）
  const selectorLabels: Record<string, string> = props.isEdit && props.initialData?.spec?.selector?.matchLabels
    ? { ...props.initialData.spec.selector.matchLabels }
    : { ...labels }

  function buildContainer(c: Container): Record<string, any> {
    const container: Record<string, any> = { name: c.name, image: c.image, imagePullPolicy: c.imagePullPolicy }
    if (c.command) container.command = c.command.split('\n').map(s => s.trim()).filter(Boolean)
    if (c.args) container.args = c.args.split('\n').map(s => s.trim()).filter(Boolean)
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
    if (c.lifecycle.preStop || c.lifecycle.postStart) {
      container.lifecycle = {}
      if (c.lifecycle.preStop) {
        if (c.lifecycle.preStop.type === 'exec') container.lifecycle.preStop = { exec: { command: c.lifecycle.preStop.execCommand.split('\n').map(s => s.trim()).filter(Boolean) } }
        else if (c.lifecycle.preStop.type === 'httpGet') container.lifecycle.preStop = { httpGet: { path: c.lifecycle.preStop.httpGetPath, port: c.lifecycle.preStop.httpGetPort } }
        else if (c.lifecycle.preStop.type === 'tcpSocket') container.lifecycle.preStop = { tcpSocket: { port: c.lifecycle.preStop.tcpSocketPort } }
      }
      if (c.lifecycle.postStart) {
        if (c.lifecycle.postStart.type === 'exec') container.lifecycle.postStart = { exec: { command: c.lifecycle.postStart.execCommand.split('\n').map(s => s.trim()).filter(Boolean) } }
        else if (c.lifecycle.postStart.type === 'httpGet') container.lifecycle.postStart = { httpGet: { path: c.lifecycle.postStart.httpGetPath, port: c.lifecycle.postStart.httpGetPort } }
        else if (c.lifecycle.postStart.type === 'tcpSocket') container.lifecycle.postStart = { tcpSocket: { port: c.lifecycle.postStart.tcpSocketPort } }
      }
    }
    const sc: any = {}
    if (c.securityContext.runAsUser !== null) sc.runAsUser = c.securityContext.runAsUser
    if (c.securityContext.runAsNonRoot) sc.runAsNonRoot = true
    if (c.securityContext.readOnlyRootFilesystem) sc.readOnlyRootFilesystem = true
    if (c.securityContext.privileged) sc.privileged = true
    if (c.securityContext.capabilitiesAdd.length > 0 || c.securityContext.capabilitiesDrop.length > 0) {
      sc.capabilities = {}
      if (c.securityContext.capabilitiesAdd.length > 0) sc.capabilities.add = c.securityContext.capabilitiesAdd.filter(Boolean)
      if (c.securityContext.capabilitiesDrop.length > 0) sc.capabilities.drop = c.securityContext.capabilitiesDrop.filter(Boolean)
    }
    if (Object.keys(sc).length > 0) container.securityContext = sc
    return container
  }

  const containers = form.containers.map(buildContainer)
  const initContainers = form.initContainers.filter(c => c.name).map(buildContainer)

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

  const podSpec: any = { containers }
  if (initContainers.length > 0) podSpec.initContainers = initContainers
  if (volumes.length > 0) podSpec.volumes = volumes
  if (Object.keys(nodeSelector).length > 0) podSpec.nodeSelector = nodeSelector
  if (tolerations.length > 0) podSpec.tolerations = tolerations
  if (form.serviceAccountName) podSpec.serviceAccountName = form.serviceAccountName
  if (form.terminationGracePeriodSeconds) podSpec.terminationGracePeriodSeconds = form.terminationGracePeriodSeconds
  if (imagePullSecrets.length > 0) podSpec.imagePullSecrets = imagePullSecrets
  if (form.dnsPolicy && form.dnsPolicy !== 'ClusterFirst') podSpec.dnsPolicy = form.dnsPolicy
  if (form.hostNetwork) podSpec.hostNetwork = true
  if (form.priorityClassName) podSpec.priorityClassName = form.priorityClassName

  // Pod Affinity
  const buildAffinityTerm = (rule: AffinityRule) => {
    const term: any = { topologyKey: rule.topologyKey }
    if (rule.labelKey) term.labelSelector = { matchLabels: { [rule.labelKey]: rule.labelValue } }
    if (rule.namespaces) term.namespaces = rule.namespaces.split(',').map(s => s.trim()).filter(Boolean)
    return term
  }
  if (form.podAffinityRules.length > 0) {
    const preferred = form.podAffinityRules.filter(r => r.weight > 0 && r.topologyKey).map(r => ({ weight: r.weight, podAffinityTerm: buildAffinityTerm(r) }))
    const required = form.podAffinityRules.filter(r => r.weight === 0 && r.topologyKey).map(r => buildAffinityTerm(r))
    if (!podSpec.affinity) podSpec.affinity = {}
    if (preferred.length > 0 || required.length > 0) {
      podSpec.affinity.podAffinity = {}
      if (preferred.length > 0) podSpec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution = preferred
      if (required.length > 0) podSpec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution = required
    }
  }
  if (form.podAntiAffinityRules.length > 0) {
    const preferred = form.podAntiAffinityRules.filter(r => r.weight > 0 && r.topologyKey).map(r => ({ weight: r.weight, podAffinityTerm: buildAffinityTerm(r) }))
    const required = form.podAntiAffinityRules.filter(r => r.weight === 0 && r.topologyKey).map(r => buildAffinityTerm(r))
    if (!podSpec.affinity) podSpec.affinity = {}
    if (preferred.length > 0 || required.length > 0) {
      podSpec.affinity.podAntiAffinity = {}
      if (preferred.length > 0) podSpec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution = preferred
      if (required.length > 0) podSpec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution = required
    }
  }

  // Topology Spread Constraints
  if (form.topologySpreadConstraints.length > 0) {
    podSpec.topologySpreadConstraints = form.topologySpreadConstraints.filter(t => t.topologyKey).map(t => {
      const tc: any = { maxSkew: t.maxSkew, topologyKey: t.topologyKey, whenUnsatisfiable: t.whenUnsatisfiable }
      if (t.labelKey) tc.labelSelector = { matchLabels: { [t.labelKey]: t.labelValue } }
      return tc
    })
  }

  const podTemplate: any = { metadata: { labels: { ...labels } }, spec: podSpec }
  if (Object.keys(annotations).length > 0) podTemplate.metadata.annotations = annotations

  // 编辑模式保留资源级 annotations（全量 PUT 替换，否则会被清空）；表单的"注解"栏只编辑 template 级 annotations
  const metadata: any = { name: form.name, namespace: form.namespace, labels: { ...labels } }
  if (props.isEdit && props.initialData?.metadata?.annotations) {
    metadata.annotations = { ...props.initialData.metadata.annotations }
  }
  const resource: any = { apiVersion: 'apps/v1', kind: props.kind, metadata, spec: {} }

  if (props.kind === 'Deployment') {
    resource.spec = { replicas: form.replicas, selector: { matchLabels: selectorLabels }, template: podTemplate, strategy: { type: form.strategyType } }
    if (form.strategyType === 'RollingUpdate') resource.spec.strategy.rollingUpdate = { maxSurge: form.maxSurge, maxUnavailable: form.maxUnavailable }
  } else if (props.kind === 'StatefulSet') {
    resource.spec = { replicas: form.replicas, selector: { matchLabels: selectorLabels }, template: podTemplate, serviceName: form.serviceName || form.name, updateStrategy: { type: form.updateStrategy }, podManagementPolicy: form.podManagementPolicy }
    const vcts = form.volumeClaimTemplates.filter(v => v.name).map(v => {
      const vct: any = { metadata: { name: v.name }, spec: { accessModes: v.accessModes, resources: { requests: { storage: v.storageSize } } } }
      if (v.storageClassName) vct.spec.storageClassName = v.storageClassName
      return vct
    })
    if (vcts.length > 0) resource.spec.volumeClaimTemplates = vcts
  } else if (props.kind === 'DaemonSet') {
    resource.spec = { selector: { matchLabels: selectorLabels }, template: podTemplate, updateStrategy: { type: form.dsUpdateStrategy } }
  }

  return resource
}

// Resource validation helpers
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
  // Validate basic fields
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  // Validate containers
  for (let i = 0; i < form.containers.length; i++) {
    if (!form.containers[i].name) { ElMessage.error(t('workload.containerNameRequired', { n: i + 1 })); return }
    if (!form.containers[i].image) { ElMessage.error(t('workload.containerImageRequired', { n: i + 1 })); return }
    // Validate resource requests <= limits
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
      // Custom submit handler (for edit mode)
      await props.onSubmit(generatedYaml.value)
    } else if (props.isEdit) {
      // Edit mode - call update API based on kind
      const updateFn = props.kind === 'Deployment' ? updateDeploymentYaml
        : props.kind === 'StatefulSet' ? updateStatefulSetYaml
        : props.kind === 'DaemonSet' ? updateDaemonSetYaml
        : null
      if (!updateFn) {
        ElMessage.error(t('workload.unsupportedResourceType', { type: props.kind }))
        return
      }
      await updateFn({ namespace: form.namespace, name: form.name, yaml: generatedYaml.value })
      ElMessage.success(t('common.updateSuccess'))
      emit('success')
    } else {
      // Create mode
      await (props.kind === 'Deployment' ? createDeployment : props.kind === 'StatefulSet' ? createStatefulSet : createDaemonSet)({ namespace: form.namespace, yaml: generatedYaml.value })
      ElMessage.success(t('common.createSuccess'))
      router.push(getListRoute())
    }
  } catch (e: any) { ElMessage.error(e?.message || (props.isEdit ? t('common.updateFailed') : t('common.createFailed'))) }
  finally { submitting.value = false }
}

function getListRoute(): string {
  return props.kind === 'Deployment' ? '/workloads/deployments' : props.kind === 'StatefulSet' ? '/workloads/statefulsets' : '/workloads/daemonsets'
}

function handleCancel() {
  if (props.isEdit) {
    emit('cancel')
  } else {
    router.push(getListRoute())
  }
}
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
              <el-input v-model="form.name" :disabled="isEdit" placeholder="my-app" />
            </el-form-item>
            <el-form-item :label="t('common.namespace_label')" prop="namespace">
              <el-select v-model="form.namespace" :disabled="isEdit" filterable placeholder="选择命名空间" style="width: 100%;" :loading="namespaceLoading">
                <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="kind !== 'DaemonSet'" label="副本数" prop="replicas">
              <el-input-number v-model="form.replicas" :min="1" :max="1000" style="width: 100%;" />
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
          <LabelsAnnotationsForm v-model:labels="form.labels" v-model:annotations="form.annotations" />
        </div>
      </div>

      <!-- Section 2: Container Config -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">容器配置</div>
        </div>
        <div class="section-content">
          <ContainerConfigForm :containers="form.containers" title="容器" />
        </div>
      </div>

      <!-- Section: Init Containers -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">初始化容器</div>
        </div>
        <div class="section-content">
          <el-alert type="info" :closable="false" style="margin-bottom: var(--gk-space-4);">
            初始化容器在主容器启动之前运行，常用于数据迁移、依赖检查等场景。
          </el-alert>
          <ContainerConfigForm :containers="form.initContainers" title="初始化容器" :show-pull-policy="false" :min-containers="0" index-color="var(--el-color-warning)" />
        </div>
      </div>

      <!-- Section 3: Storage -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">存储配置</div>
        </div>
        <div class="section-content">
          <StorageConfigForm :volumes="form.volumes" :containers="form.containers" :volume-claim-templates="form.volumeClaimTemplates" :kind="kind" />
        </div>
      </div>

      <!-- Section 4: Health Probes -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">健康检查</div>
        </div>
        <div class="section-content">
          <HealthCheckForm :containers="form.containers" />
        </div>
      </div>

      <!-- Section 5: Security -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">安全设置</div>
        </div>
        <div class="section-content">
          <SecurityContextForm :containers="form.containers" />
        </div>
      </div>

      <!-- Section 6: Scheduling -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">调度配置</div>
        </div>
        <div class="section-content">
          <SchedulingForm
            v-model:nodeSelector="form.nodeSelector"
            v-model:tolerations="form.tolerations"
            v-model:podAffinityRules="form.podAffinityRules"
            v-model:podAntiAffinityRules="form.podAntiAffinityRules"
            v-model:topologySpreadConstraints="form.topologySpreadConstraints"
          />
        </div>
      </div>

      <!-- Section 7: Advanced -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">高级配置</div>
        </div>
        <div class="section-content">

          <template v-if="kind === 'Deployment'">
            <div class="fields-grid">
              <el-form-item label="更新策略">
                <el-select v-model="form.strategyType" style="width: 100%;">
                  <el-option label="RollingUpdate" value="RollingUpdate" />
                  <el-option label="Recreate" value="Recreate" />
                </el-select>
              </el-form-item>
              <template v-if="form.strategyType === 'RollingUpdate'">
                <el-form-item label="Max Surge">
                  <el-input v-model="form.maxSurge" placeholder="25% 或 2" />
                </el-form-item>
                <el-form-item label="Max Unavailable">
                  <el-input v-model="form.maxUnavailable" placeholder="25% 或 1" />
                </el-form-item>
              </template>
            </div>
          </template>
          <template v-if="kind === 'StatefulSet'">
            <div class="fields-grid">
              <el-form-item label="服务名称">
                <el-input v-model="form.serviceName" placeholder="Headless service 名称" />
              </el-form-item>
              <el-form-item label="Pod 管理策略">
                <el-select v-model="form.podManagementPolicy" style="width: 100%;">
                  <el-option label="OrderedReady - 顺序创建/删除" value="OrderedReady" />
                  <el-option label="Parallel - 并行创建/删除" value="Parallel" />
                </el-select>
              </el-form-item>
              <el-form-item label="更新策略">
                <el-select v-model="form.updateStrategy" style="width: 100%;">
                  <el-option label="RollingUpdate" value="RollingUpdate" />
                  <el-option label="OnDelete" value="OnDelete" />
                </el-select>
              </el-form-item>
            </div>
          </template>
          <template v-if="kind === 'DaemonSet'">
            <el-form-item label="更新策略">
              <el-select v-model="form.dsUpdateStrategy" style="width: 100%;">
                <el-option label="RollingUpdate" value="RollingUpdate" />
                <el-option label="OnDelete" value="OnDelete" />
              </el-select>
            </el-form-item>
            <el-form-item label="主机网络">
              <el-switch v-model="form.hostNetwork" />
              <div class="form-help">启用后 Pod 直接使用主机网络命名空间，常用于需要访问主机网络的 DaemonSet（如监控 agent、日志收集器）</div>
            </el-form-item>
          </template>

          <el-divider />

          <div class="fields-grid">
            <el-form-item label="DNS 策略">
              <el-select v-model="form.dnsPolicy" style="width: 100%;">
                <el-option label="ClusterFirst - 集群 DNS 优先（默认）" value="ClusterFirst" />
                <el-option label="Default - 继承节点 DNS" value="Default" />
                <el-option label="ClusterFirstWithHostNet - 主机网络下仍用集群 DNS" value="ClusterFirstWithHostNet" />
                <el-option label="None - 必须手动配置 dnsConfig" value="None" />
              </el-select>
            </el-form-item>
            <el-form-item label="优先级类名 (PriorityClassName)">
              <el-input v-model="form.priorityClassName" placeholder="留空使用默认优先级" />
              <div class="form-help">如 system-node-critical, system-cluster-critical 等，决定资源不足时的驱逐顺序</div>
            </el-form-item>
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
            <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ isEdit ? t('common.update') : t('common.create') }}</el-button>
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
