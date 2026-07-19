<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import yaml from 'js-yaml'
import type { FormInstance, FormRules } from 'element-plus'
import { getNamespaceList, extractNamespaceNames } from '@/api/resource'
import { createDeployment, createStatefulSet, createDaemonSet, updateDeploymentYaml, updateStatefulSetYaml, updateDaemonSetYaml, updateJobYaml, updateCronJobYaml } from '@/api/resource'

const props = withDefaults(defineProps<{
  kind: 'Deployment' | 'StatefulSet' | 'DaemonSet' | 'Job' | 'CronJob'
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
interface Volume { name: string; type: string; hostPath: string; hostPathType: string; configMapName: string; secretName: string; pvcName: string }
interface LifecycleHandler { type: 'exec' | 'httpGet' | 'tcpSocket'; execCommand: string; httpGetPath: string; httpGetPort: number | null; tcpSocketPort: number | null }
interface Container {
  name: string; image: string; imagePullPolicy: string
  ports: Port[]; env: EnvVar[]; resources: Resources
  volumeMounts: VolumeMount[]; livenessProbe: Probe | null; readinessProbe: Probe | null; startupProbe: Probe | null
  command: string; args: string
  lifecycle: { preStop: LifecycleHandler | null; postStart: LifecycleHandler | null }
  securityContext: { runAsUser: number | null; runAsNonRoot: boolean; readOnlyRootFilesystem: boolean; privileged: boolean; capabilitiesAdd: string[]; capabilitiesDrop: string[] }
}
interface Tolerance { key: string; operator: string; value: string; effect: string; tolerationSeconds: number | null }
interface Annotation { key: string; value: string }
interface VolumeClaimTemplate {
  name: string; storageSize: string; storageClassName: string; accessModes: string[]
}
interface AffinityRule { weight: number; topologyKey: string; namespaces: string; labelKey: string; labelValue: string }
interface TopologySpreadConstraint { maxSkew: number; topologyKey: string; whenUnsatisfiable: string; labelKey: string; labelValue: string }

interface FormData {
  name: string; namespace: string; replicas: number; labels: Label[]
  containers: Container[]; initContainers: Container[]; volumes: Volume[]
  strategyType: string; maxSurge: string; maxUnavailable: string
  serviceName: string; updateStrategy: string; dsUpdateStrategy: string
  nodeSelector: Label[]; tolerations: Tolerance[]; annotations: Annotation[]
  serviceAccountName: string; terminationGracePeriodSeconds: number | null
  imagePullSecrets: string[]
  volumeClaimTemplates: VolumeClaimTemplate[]
  podAffinityRules: AffinityRule[]; podAntiAffinityRules: AffinityRule[]
  topologySpreadConstraints: TopologySpreadConstraint[]
}

function createEmptyProbe(): Probe {
  return { type: 'httpGet', httpGetPath: '/', httpGetPort: 80, tcpSocketPort: null, execCommand: '', initialDelaySeconds: 15, periodSeconds: 10, timeoutSeconds: 5, failureThreshold: 3 }
}

function createEmptyLifecycleHandler(): LifecycleHandler {
  return { type: 'exec', execCommand: '', httpGetPath: '/', httpGetPort: 80, tcpSocketPort: null }
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
  name: '', namespace: 'default', replicas: 1,
  labels: [{ key: 'app', value: '' }],
  containers: [createEmptyContainer()],
  initContainers: [],
  volumes: [],
  strategyType: 'RollingUpdate', maxSurge: '25%', maxUnavailable: '25%',
  serviceName: '', updateStrategy: 'RollingUpdate', dsUpdateStrategy: 'RollingUpdate',
  nodeSelector: [], tolerations: [], annotations: [],
  serviceAccountName: '', terminationGracePeriodSeconds: null, imagePullSecrets: [],
  volumeClaimTemplates: [],
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
    httpGetPath: probeData.httpGet?.path || '/',
    httpGetPort: probeData.httpGet?.port || 80,
    tcpSocketPort: probeData.tcpSocket?.port || null,
    execCommand: probeData.exec?.command?.join(' ') || '',
    initialDelaySeconds: probeData.initialDelaySeconds || 15,
    periodSeconds: probeData.periodSeconds || 10,
    timeoutSeconds: probeData.timeoutSeconds || 5,
    failureThreshold: probeData.failureThreshold || 3,
  }
}

function parseLifecycleHandler(data: any): LifecycleHandler | null {
  if (!data) return null
  return {
    type: data.exec ? 'exec' : data.httpGet ? 'httpGet' : 'tcpSocket',
    execCommand: data.exec?.command?.join(' ') || '',
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
    command: c.command?.join(' ') || '',
    args: c.args?.join(' ') || '',
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

function addLabel() { form.labels.push({ key: '', value: '' }) }
function removeLabel(i: number) { form.labels.splice(i, 1) }
function addContainer() { form.containers.push(createEmptyContainer()) }
function removeContainer(i: number) { if (form.containers.length > 1) form.containers.splice(i, 1) }
function addInitContainer() { form.initContainers.push(createEmptyContainer()) }
function removeInitContainer(i: number) { form.initContainers.splice(i, 1) }
function addPort(ci: number, isInit?: boolean) { (isInit ? form.initContainers[ci] : form.containers[ci]).ports.push({ name: '', containerPort: null, protocol: 'TCP' }) }
function removePort(ci: number, pi: number, isInit?: boolean) { (isInit ? form.initContainers[ci] : form.containers[ci]).ports.splice(pi, 1) }
function addEnv(ci: number, isInit?: boolean) { (isInit ? form.initContainers[ci] : form.containers[ci]).env.push(createEmptyEnv()) }
function removeEnv(ci: number, ei: number, isInit?: boolean) { (isInit ? form.initContainers[ci] : form.containers[ci]).env.splice(ei, 1) }
function addNodeSelector() { form.nodeSelector.push({ key: '', value: '' }) }
function removeNodeSelector(i: number) { form.nodeSelector.splice(i, 1) }
function addVolume() { form.volumes.push({ name: '', type: 'emptyDir', hostPath: '', hostPathType: 'DirectoryOrCreate', configMapName: '', secretName: '', pvcName: '' }) }
function removeVolume(i: number) { form.volumes.splice(i, 1) }
function addVolumeMount(ci: number, isInit?: boolean) { (isInit ? form.initContainers[ci] : form.containers[ci]).volumeMounts.push({ name: '', mountPath: '', subPath: '', readOnly: false }) }
function removeVolumeMount(ci: number, mi: number, isInit?: boolean) { (isInit ? form.initContainers[ci] : form.containers[ci]).volumeMounts.splice(mi, 1) }
function enableProbe(ci: number, probeType: 'livenessProbe' | 'readinessProbe' | 'startupProbe', isInit?: boolean) { (isInit ? form.initContainers[ci] : form.containers[ci])[probeType] = createEmptyProbe() }
function disableProbe(ci: number, probeType: 'livenessProbe' | 'readinessProbe' | 'startupProbe', isInit?: boolean) { (isInit ? form.initContainers[ci] : form.containers[ci])[probeType] = null }
function enableLifecycle(ci: number, hookType: 'preStop' | 'postStart', isInit?: boolean) { (isInit ? form.initContainers[ci] : form.containers[ci]).lifecycle[hookType] = createEmptyLifecycleHandler() }
function disableLifecycle(ci: number, hookType: 'preStop' | 'postStart', isInit?: boolean) { (isInit ? form.initContainers[ci] : form.containers[ci]).lifecycle[hookType] = null }
function addToleration() { form.tolerations.push({ key: '', operator: 'Equal', value: '', effect: 'NoSchedule', tolerationSeconds: null }) }
function removeToleration(i: number) { form.tolerations.splice(i, 1) }
function addAnnotation() { form.annotations.push({ key: '', value: '' }) }
function removeAnnotation(i: number) { form.annotations.splice(i, 1) }
function addImagePullSecret() { form.imagePullSecrets.push('') }
function removeImagePullSecret(i: number) { form.imagePullSecrets.splice(i, 1) }
function addVolumeClaimTemplate() { form.volumeClaimTemplates.push({ name: '', storageSize: '1Gi', storageClassName: '', accessModes: ['ReadWriteOnce'] }) }
function removeVolumeClaimTemplate(i: number) { form.volumeClaimTemplates.splice(i, 1) }
function addAffinityRule(type: 'podAffinity' | 'podAntiAffinity') { (type === 'podAffinity' ? form.podAffinityRules : form.podAntiAffinityRules).push({ weight: 1, topologyKey: '', namespaces: '', labelKey: '', labelValue: '' }) }
function removeAffinityRule(type: 'podAffinity' | 'podAntiAffinity', i: number) { (type === 'podAffinity' ? form.podAffinityRules : form.podAntiAffinityRules).splice(i, 1) }
function addTopologySpreadConstraint() { form.topologySpreadConstraints.push({ maxSkew: 1, topologyKey: '', whenUnsatisfiable: 'DoNotSchedule', labelKey: '', labelValue: '' }) }
function removeTopologySpreadConstraint(i: number) { form.topologySpreadConstraints.splice(i, 1) }
function addCapability(sc: Container['securityContext'], type: 'add' | 'drop') { (type === 'add' ? sc.capabilitiesAdd : sc.capabilitiesDrop).push('') }
function removeCapability(sc: Container['securityContext'], type: 'add' | 'drop', i: number) { (type === 'add' ? sc.capabilitiesAdd : sc.capabilitiesDrop).splice(i, 1) }

const generatedYaml = computed(() => yaml.dump(buildK8sResource(), { indent: 2, lineWidth: -1, noRefs: true }))

function buildProbe(probe: Probe | null): any {
  if (!probe) return undefined
  const p: any = { initialDelaySeconds: probe.initialDelaySeconds, periodSeconds: probe.periodSeconds, timeoutSeconds: probe.timeoutSeconds, failureThreshold: probe.failureThreshold }
  if (probe.type === 'httpGet') { p.httpGet = { path: probe.httpGetPath, port: probe.httpGetPort } }
  else if (probe.type === 'tcpSocket') { p.tcpSocket = { port: probe.tcpSocketPort } }
  else if (probe.type === 'exec') { p.exec = { command: probe.execCommand.split(' ').filter(Boolean) } }
  return p
}

function buildK8sResource(): Record<string, any> {
  const labels: Record<string, string> = {}
  form.labels.forEach(l => { if (l.key.trim()) labels[l.key.trim()] = l.value })

  function buildContainer(c: Container): Record<string, any> {
    const container: Record<string, any> = { name: c.name, image: c.image, imagePullPolicy: c.imagePullPolicy }
    if (c.command) container.command = c.command.split(/\s+/).filter(Boolean)
    if (c.args) container.args = c.args.split(/\s+/).filter(Boolean)
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
        if (c.lifecycle.preStop.type === 'exec') container.lifecycle.preStop = { exec: { command: c.lifecycle.preStop.execCommand.split(/\s+/).filter(Boolean) } }
        else if (c.lifecycle.preStop.type === 'httpGet') container.lifecycle.preStop = { httpGet: { path: c.lifecycle.preStop.httpGetPath, port: c.lifecycle.preStop.httpGetPort } }
        else if (c.lifecycle.preStop.type === 'tcpSocket') container.lifecycle.preStop = { tcpSocket: { port: c.lifecycle.preStop.tcpSocketPort } }
      }
      if (c.lifecycle.postStart) {
        if (c.lifecycle.postStart.type === 'exec') container.lifecycle.postStart = { exec: { command: c.lifecycle.postStart.execCommand.split(/\s+/).filter(Boolean) } }
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

  const resource: any = { apiVersion: 'apps/v1', kind: props.kind, metadata: { name: form.name, namespace: form.namespace, labels: { ...labels } }, spec: {} }

  if (props.kind === 'Deployment') {
    resource.spec = { replicas: form.replicas, selector: { matchLabels: { ...labels } }, template: podTemplate, strategy: { type: form.strategyType } }
    if (form.strategyType === 'RollingUpdate') resource.spec.strategy.rollingUpdate = { maxSurge: form.maxSurge, maxUnavailable: form.maxUnavailable }
  } else if (props.kind === 'StatefulSet') {
    resource.spec = { replicas: form.replicas, selector: { matchLabels: { ...labels } }, template: podTemplate, serviceName: form.serviceName || form.name, updateStrategy: { type: form.updateStrategy } }
    const vcts = form.volumeClaimTemplates.filter(v => v.name).map(v => {
      const vct: any = { metadata: { name: v.name }, spec: { accessModes: v.accessModes, resources: { requests: { storage: v.storageSize } } } }
      if (v.storageClassName) vct.spec.storageClassName = v.storageClassName
      return vct
    })
    if (vcts.length > 0) resource.spec.volumeClaimTemplates = vcts
  } else if (props.kind === 'DaemonSet') {
    resource.spec = { selector: { matchLabels: { ...labels } }, template: podTemplate, updateStrategy: { type: form.dsUpdateStrategy } }
  }

  return resource
}

async function handleSubmit() {
  // Validate basic fields
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  // Validate containers
  for (let i = 0; i < form.containers.length; i++) {
    if (!form.containers[i].name) { ElMessage.error(`容器 ${i + 1}: 名称不能为空`); return }
    if (!form.containers[i].image) { ElMessage.error(`容器 ${i + 1}: 镜像不能为空`); return }
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
        : props.kind === 'Job' ? updateJobYaml
        : props.kind === 'CronJob' ? updateCronJobYaml
        : null
      if (!updateFn) {
        ElMessage.error(`不支持的资源类型: ${props.kind}`)
        return
      }
      await updateFn({ namespace: form.namespace, name: form.name, yaml: generatedYaml.value })
      ElMessage.success(`${props.kind} 更新成功`)
      emit('success')
    } else {
      // Create mode
      await (props.kind === 'Deployment' ? createDeployment : props.kind === 'StatefulSet' ? createStatefulSet : createDaemonSet)({ namespace: form.namespace, yaml: generatedYaml.value })
      ElMessage.success(`${props.kind} 创建成功`)
      router.push(getListRoute())
    }
  } catch (e: any) { ElMessage.error(e?.message || (props.isEdit ? '更新失败' : '创建失败')) }
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
          <div class="section-title">基本信息</div>
        </div>
        <div class="section-content">
          <div class="fields-grid">
            <el-form-item label="名称" prop="name">
              <el-input v-model="form.name" :disabled="isEdit" placeholder="my-app" />
            </el-form-item>
            <el-form-item label="命名空间" prop="namespace">
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

      <!-- Section 2: Container Config -->
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
              <el-form-item label="启动命令 (command)">
                <el-input v-model="container.command" placeholder="多个参数用空格分隔，如: /bin/sh -c" />
              </el-form-item>
              <el-form-item label="启动参数 (args)" class="full-width">
                <el-input v-model="container.args" placeholder="多个参数用空格分隔，如: echo hello" />
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
              <el-select v-model="env.type" style="width: 140px;" @change="env.value = ''; env.configMapName = ''; env.configMapKey = ''; env.secretName = ''; env.secretKey = ''; env.fieldPath = ''">
                <el-option label="直接值" value="plain" />
                <el-option label="ConfigMap" value="configMapKeyRef" />
                <el-option label="Secret" value="secretKeyRef" />
                <el-option label="字段引用" value="fieldRef" />
              </el-select>
              <el-input v-if="env.type === 'plain'" v-model="env.value" placeholder="值" style="flex: 1;" />
              <template v-if="env.type === 'configMapKeyRef'">
                <el-input v-model="env.configMapName" placeholder="ConfigMap 名称" style="flex: 1;" />
                <el-input v-model="env.configMapKey" placeholder="Key" style="width: 140px;" />
              </template>
              <template v-if="env.type === 'secretKeyRef'">
                <el-input v-model="env.secretName" placeholder="Secret 名称" style="flex: 1;" />
                <el-input v-model="env.secretKey" placeholder="Key" style="width: 140px;" />
              </template>
              <el-input v-if="env.type === 'fieldRef'" v-model="env.fieldPath" placeholder="如: metadata.name" style="flex: 1;" />
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
          <el-alert type="info" :closable="false" style="margin-bottom: 16px;">
            初始化容器在主容器启动之前运行，常用于数据迁移、依赖检查等场景。
          </el-alert>
          <div v-for="(container, ci) in form.initContainers" :key="ci" class="container-card">
            <div class="container-card-header">
              <div class="container-title">
                <span class="container-index" style="background: var(--el-color-warning);">{{ ci + 1 }}</span>
                <span>{{ container.name || '未命名初始化容器' }}</span>
              </div>
              <el-button type="danger" text size="small" @click="removeInitContainer(ci)">
                <el-icon><Delete /></el-icon> 移除
              </el-button>
            </div>
            <div class="fields-grid">
              <el-form-item label="容器名称" required>
                <el-input v-model="container.name" placeholder="init-mysql" />
              </el-form-item>
              <el-form-item label="镜像" required>
                <el-input v-model="container.image" placeholder="busybox:1.36" />
              </el-form-item>
              <el-form-item label="启动命令">
                <el-input v-model="container.command" placeholder="多个参数用空格分隔" />
              </el-form-item>
              <el-form-item label="启动参数">
                <el-input v-model="container.args" placeholder="多个参数用空格分隔" />
              </el-form-item>
            </div>
            <!-- Env for init container -->
            <el-divider content-position="left">环境变量</el-divider>
            <div v-for="(env, ei) in container.env" :key="ei" class="env-row">
              <el-input v-model="env.name" placeholder="名称" style="width: 140px;" />
              <el-select v-model="env.type" style="width: 140px;" @change="env.value = ''; env.configMapName = ''; env.configMapKey = ''; env.secretName = ''; env.secretKey = ''; env.fieldPath = ''">
                <el-option label="直接值" value="plain" />
                <el-option label="ConfigMap" value="configMapKeyRef" />
                <el-option label="Secret" value="secretKeyRef" />
                <el-option label="字段引用" value="fieldRef" />
              </el-select>
              <el-input v-if="env.type === 'plain'" v-model="env.value" placeholder="值" style="flex: 1;" />
              <template v-if="env.type === 'configMapKeyRef'">
                <el-input v-model="env.configMapName" placeholder="ConfigMap 名称" style="flex: 1;" />
                <el-input v-model="env.configMapKey" placeholder="Key" style="width: 140px;" />
              </template>
              <template v-if="env.type === 'secretKeyRef'">
                <el-input v-model="env.secretName" placeholder="Secret 名称" style="flex: 1;" />
                <el-input v-model="env.secretKey" placeholder="Key" style="width: 140px;" />
              </template>
              <el-input v-if="env.type === 'fieldRef'" v-model="env.fieldPath" placeholder="如: metadata.name" style="flex: 1;" />
              <el-button type="danger" text circle @click="removeEnv(ci, ei, true)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <el-button text type="primary" size="small" @click="addEnv(ci, true)">
              <el-icon><Plus /></el-icon> 添加环境变量
            </el-button>
            <!-- Resources for init container -->
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

      <!-- Section 3: Storage -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">存储配置</div>
        </div>
        <div class="section-content">

          <!-- Volume Claim Templates (StatefulSet only) -->
          <template v-if="kind === 'StatefulSet'">
            <el-form-item label="持久卷声明模板 (VolumeClaimTemplates)">
              <div style="width: 100%;">
                <div v-for="(vct, vi) in form.volumeClaimTemplates" :key="vi" class="volume-card">
                  <div class="volume-row">
                    <el-input v-model="vct.name" placeholder="模板名称 (如: data)" />
                    <el-button type="danger" text circle @click="removeVolumeClaimTemplate(vi)">
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </div>
                  <div class="fields-grid" style="margin-top: 8px;">
                    <el-form-item label="存储大小">
                      <el-input v-model="vct.storageSize" placeholder="1Gi" />
                    </el-form-item>
                    <el-form-item label="存储类名">
                      <el-input v-model="vct.storageClassName" placeholder="留空使用默认 StorageClass" />
                    </el-form-item>
                    <el-form-item label="访问模式" class="full-width">
                      <el-checkbox-group v-model="vct.accessModes">
                        <el-checkbox label="ReadWriteOnce" />
                        <el-checkbox label="ReadOnlyMany" />
                        <el-checkbox label="ReadWriteMany" />
                      </el-checkbox-group>
                    </el-form-item>
                  </div>
                </div>
                <el-button text type="primary" @click="addVolumeClaimTemplate" size="small">
                  <el-icon><Plus /></el-icon> 添加持久卷声明模板
                </el-button>
              </div>
            </el-form-item>
            <el-divider />
          </template>

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
                <template v-if="vol.type === 'hostPath'">
                  <el-input v-model="vol.hostPath" placeholder="主机路径 (e.g. /data)" style="margin-top: 8px;" />
                  <el-select v-model="vol.hostPathType" style="margin-top: 8px; width: 100%;">
                    <el-option label="DirectoryOrCreate" value="DirectoryOrCreate" />
                    <el-option label="Directory" value="Directory" />
                    <el-option label="FileOrCreate" value="FileOrCreate" />
                    <el-option label="File" value="File" />
                    <el-option label="Socket" value="Socket" />
                    <el-option label="CharDevice" value="CharDevice" />
                    <el-option label="BlockDevice" value="BlockDevice" />
                  </el-select>
                </template>
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

      <!-- Section 4: Health Probes -->
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

            <!-- Startup Probe -->
            <div class="probe-card">
              <div class="probe-header">
                <div>
                  <span class="probe-label">启动探针</span>
                  <span class="probe-desc">慢启动应用专用，成功后切换到存活探针</span>
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

            <!-- Lifecycle Hooks -->
            <div class="probe-card">
              <div class="probe-header">
                <div>
                  <span class="probe-label">生命周期钩子</span>
                  <span class="probe-desc">容器启动后/停止前执行的操作</span>
                </div>
              </div>
              <div style="margin-top: 12px;">
                <div style="margin-bottom: 12px;">
                  <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 8px;">
                    <span style="font-size: 13px; font-weight: 600; color: var(--el-text-color-regular);">postStart（启动后）</span>
                    <el-switch :model-value="!!container.lifecycle.postStart" @update:model-value="(v: boolean) => v ? enableLifecycle(ci, 'postStart') : disableLifecycle(ci, 'postStart')" />
                  </div>
                  <template v-if="container.lifecycle.postStart">
                    <div class="fields-grid">
                      <el-form-item label="类型">
                        <el-select v-model="container.lifecycle.postStart.type" style="width: 100%;">
                          <el-option label="Exec" value="exec" />
                          <el-option label="HTTP GET" value="httpGet" />
                          <el-option label="TCP Socket" value="tcpSocket" />
                        </el-select>
                      </el-form-item>
                      <el-form-item v-if="container.lifecycle.postStart.type === 'exec'" label="命令">
                        <el-input v-model="container.lifecycle.postStart.execCommand" placeholder="多个参数用空格分隔" />
                      </el-form-item>
                      <el-form-item v-if="container.lifecycle.postStart.type === 'httpGet'" label="路径">
                        <el-input v-model="container.lifecycle.postStart.httpGetPath" placeholder="/" />
                      </el-form-item>
                      <el-form-item v-if="container.lifecycle.postStart.type === 'httpGet'" label="端口">
                        <el-input-number v-model="container.lifecycle.postStart.httpGetPort" :min="1" :max="65535" style="width: 100%;" />
                      </el-form-item>
                      <el-form-item v-if="container.lifecycle.postStart.type === 'tcpSocket'" label="端口">
                        <el-input-number v-model="container.lifecycle.postStart.tcpSocketPort" :min="1" :max="65535" style="width: 100%;" />
                      </el-form-item>
                    </div>
                  </template>
                </div>
                <div>
                  <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 8px;">
                    <span style="font-size: 13px; font-weight: 600; color: var(--el-text-color-regular);">preStop（停止前）</span>
                    <el-switch :model-value="!!container.lifecycle.preStop" @update:model-value="(v: boolean) => v ? enableLifecycle(ci, 'preStop') : disableLifecycle(ci, 'preStop')" />
                  </div>
                  <template v-if="container.lifecycle.preStop">
                    <div class="fields-grid">
                      <el-form-item label="类型">
                        <el-select v-model="container.lifecycle.preStop.type" style="width: 100%;">
                          <el-option label="Exec" value="exec" />
                          <el-option label="HTTP GET" value="httpGet" />
                          <el-option label="TCP Socket" value="tcpSocket" />
                        </el-select>
                      </el-form-item>
                      <el-form-item v-if="container.lifecycle.preStop.type === 'exec'" label="命令">
                        <el-input v-model="container.lifecycle.preStop.execCommand" placeholder="多个参数用空格分隔" />
                      </el-form-item>
                      <el-form-item v-if="container.lifecycle.preStop.type === 'httpGet'" label="路径">
                        <el-input v-model="container.lifecycle.preStop.httpGetPath" placeholder="/" />
                      </el-form-item>
                      <el-form-item v-if="container.lifecycle.preStop.type === 'httpGet'" label="端口">
                        <el-input-number v-model="container.lifecycle.preStop.httpGetPort" :min="1" :max="65535" style="width: 100%;" />
                      </el-form-item>
                      <el-form-item v-if="container.lifecycle.preStop.type === 'tcpSocket'" label="端口">
                        <el-input-number v-model="container.lifecycle.preStop.tcpSocketPort" :min="1" :max="65535" style="width: 100%;" />
                      </el-form-item>
                    </div>
                  </template>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Section 5: Security -->
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
            <!-- Capabilities -->
            <div style="margin-top: 16px;">
              <el-divider content-position="left">Linux Capabilities</el-divider>
              <div class="fields-grid">
                <el-form-item label="添加 (Add)">
                  <div style="width: 100%;">
                    <div v-for="(_cap, i) in container.securityContext.capabilitiesAdd" :key="i" class="kv-row">
                      <el-select v-model="container.securityContext.capabilitiesAdd[i]" filterable allow-create placeholder="如: NET_ADMIN" style="flex: 1;">
                        <el-option label="NET_ADMIN" value="NET_ADMIN" />
                        <el-option label="NET_RAW" value="NET_RAW" />
                        <el-option label="SYS_ADMIN" value="SYS_ADMIN" />
                        <el-option label="SYS_PTRACE" value="SYS_PTRACE" />
                        <el-option label="SYS_TIME" value="SYS_TIME" />
                        <el-option label="SYS_RESOURCE" value="SYS_RESOURCE" />
                        <el-option label="DAC_OVERRIDE" value="DAC_OVERRIDE" />
                        <el-option label="DAC_READ_SEARCH" value="DAC_READ_SEARCH" />
                        <el-option label="SETUID" value="SETUID" />
                        <el-option label="SETGID" value="SETGID" />
                        <el-option label="CHOWN" value="CHOWN" />
                        <el-option label="FOWNER" value="FOWNER" />
                        <el-option label="KILL" value="KILL" />
                        <el-option label="MKNOD" value="MKNOD" />
                      </el-select>
                      <el-button type="danger" text circle @click="removeCapability(container.securityContext, 'add', i)">
                        <el-icon><Delete /></el-icon>
                      </el-button>
                    </div>
                    <el-button text type="primary" size="small" @click="addCapability(container.securityContext, 'add')">
                      <el-icon><Plus /></el-icon> 添加
                    </el-button>
                  </div>
                </el-form-item>
                <el-form-item label="移除 (Drop)">
                  <div style="width: 100%;">
                    <div v-for="(_cap, i) in container.securityContext.capabilitiesDrop" :key="i" class="kv-row">
                      <el-select v-model="container.securityContext.capabilitiesDrop[i]" filterable allow-create placeholder="如: ALL" style="flex: 1;">
                        <el-option label="ALL" value="ALL" />
                        <el-option label="NET_ADMIN" value="NET_ADMIN" />
                        <el-option label="NET_RAW" value="NET_RAW" />
                        <el-option label="SYS_ADMIN" value="SYS_ADMIN" />
                        <el-option label="SYS_PTRACE" value="SYS_PTRACE" />
                        <el-option label="SYS_TIME" value="SYS_TIME" />
                        <el-option label="SYS_RESOURCE" value="SYS_RESOURCE" />
                        <el-option label="DAC_OVERRIDE" value="DAC_OVERRIDE" />
                        <el-option label="DAC_READ_SEARCH" value="DAC_READ_SEARCH" />
                        <el-option label="SETUID" value="SETUID" />
                        <el-option label="SETGID" value="SETGID" />
                        <el-option label="CHOWN" value="CHOWN" />
                        <el-option label="FOWNER" value="FOWNER" />
                        <el-option label="KILL" value="KILL" />
                        <el-option label="MKNOD" value="MKNOD" />
                      </el-select>
                      <el-button type="danger" text circle @click="removeCapability(container.securityContext, 'drop', i)">
                        <el-icon><Delete /></el-icon>
                      </el-button>
                    </div>
                    <el-button text type="primary" size="small" @click="addCapability(container.securityContext, 'drop')">
                      <el-icon><Plus /></el-icon> 添加
                    </el-button>
                  </div>
                </el-form-item>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Section 6: Scheduling -->
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
          <el-divider />
          <el-form-item label="Pod 亲和性">
            <div style="width: 100%;">
              <div class="affinity-section">
                <div class="affinity-section-title">亲和规则（Pod Affinity）</div>
                <div v-for="(rule, i) in form.podAffinityRules" :key="i" class="affinity-row">
                  <el-input v-model="rule.topologyKey" placeholder="topologyKey" style="flex: 1;" />
                  <el-input v-model="rule.labelKey" placeholder="标签 Key" style="width: 140px;" />
                  <el-input v-model="rule.labelValue" placeholder="标签 Value" style="width: 140px;" />
                  <el-input-number v-model="rule.weight" :min="0" :max="100" placeholder="权重" style="width: 100px;" />
                  <el-button type="danger" text circle @click="removeAffinityRule('podAffinity', i)">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
                <el-button text type="primary" size="small" @click="addAffinityRule('podAffinity')">
                  <el-icon><Plus /></el-icon> 添加亲和规则
                </el-button>
              </div>
              <div class="affinity-section" style="margin-top: 16px;">
                <div class="affinity-section-title">反亲和规则（Pod Anti-Affinity）</div>
                <div v-for="(rule, i) in form.podAntiAffinityRules" :key="i" class="affinity-row">
                  <el-input v-model="rule.topologyKey" placeholder="topologyKey" style="flex: 1;" />
                  <el-input v-model="rule.labelKey" placeholder="标签 Key" style="width: 140px;" />
                  <el-input v-model="rule.labelValue" placeholder="标签 Value" style="width: 140px;" />
                  <el-input-number v-model="rule.weight" :min="0" :max="100" placeholder="权重" style="width: 100px;" />
                  <el-button type="danger" text circle @click="removeAffinityRule('podAntiAffinity', i)">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
                <el-button text type="primary" size="small" @click="addAffinityRule('podAntiAffinity')">
                  <el-icon><Plus /></el-icon> 添加反亲和规则
                </el-button>
              </div>
              <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-top: 8px;">
                权重 0 = required（必须满足），权重 1-100 = preferred（优先满足）。topologyKey 常用: kubernetes.io/hostname, topology.kubernetes.io/zone
              </div>
            </div>
          </el-form-item>

          <!-- Topology Spread Constraints -->
          <el-divider />
          <el-form-item label="拓扑分布约束">
            <div style="width: 100%;">
              <div v-for="(tc, i) in form.topologySpreadConstraints" :key="i" class="topology-row">
                <el-input v-model="tc.topologyKey" placeholder="topologyKey" style="flex: 1;" />
                <el-input v-model="tc.labelKey" placeholder="标签 Key" style="width: 140px;" />
                <el-input v-model="tc.labelValue" placeholder="标签 Value" style="width: 140px;" />
                <el-input-number v-model="tc.maxSkew" :min="1" placeholder="maxSkew" style="width: 100px;" />
                <el-select v-model="tc.whenUnsatisfiable" style="width: 150px;">
                  <el-option label="DoNotSchedule" value="DoNotSchedule" />
                  <el-option label="ScheduleAnyway" value="ScheduleAnyway" />
                </el-select>
                <el-button type="danger" text circle @click="removeTopologySpreadConstraint(i)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
              <el-button text type="primary" size="small" @click="addTopologySpreadConstraint">
                <el-icon><Plus /></el-icon> 添加拓扑约束
              </el-button>
              <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-top: 8px;">
                控制 Pod 在不同拓扑域（节点、可用区等）间的分布。maxSkew 表示最大偏差，topologyKey 常用: kubernetes.io/hostname, topology.kubernetes.io/zone
              </div>
            </div>
          </el-form-item>
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
          </template>

          <el-divider />

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
            <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ isEdit ? '更新' : '创建' }}</el-button>
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

.env-row :deep(.el-input) {
  flex: 1;
}

/* Affinity section */
.affinity-section {
  padding: 12px;
  border: 1px solid var(--el-border-color-extra-light);
  border-radius: 8px;
  background: var(--el-fill-color-lighter);
}

.affinity-section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-regular);
  margin-bottom: 12px;
}

.affinity-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: center;
}

.affinity-row :deep(.el-input) {
  flex: 1;
}

/* Topology row */
.topology-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: center;
}

.topology-row :deep(.el-input) {
  flex: 1;
}

</style>
