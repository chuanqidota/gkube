<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import yaml from 'js-yaml'
import type { FormInstance, FormRules } from 'element-plus'
import YamlEditor from '@/components/YamlEditor.vue'
import { getNamespaceList, extractNamespaceNames } from '@/api/resource'
import { createDeployment, createStatefulSet, createDaemonSet } from '@/api/resource'

const props = defineProps<{
  kind: 'Deployment' | 'StatefulSet' | 'DaemonSet'
}>()

const router = useRouter()
const currentStep = ref(0)
const submitting = ref(false)
const namespaceLoading = ref(false)
const namespaces = ref<string[]>([])

interface Label { key: string; value: string }
interface Port { name: string; containerPort: number | null; protocol: string }
interface EnvVar { name: string; value: string }
interface Resources { requests: { cpu: string; memory: string }; limits: { cpu: string; memory: string } }
interface VolumeMount { name: string; mountPath: string; subPath: string; readOnly: boolean }
interface Probe { type: string; httpGetPath: string; httpGetPort: number | null; tcpSocketPort: number | null; execCommand: string; initialDelaySeconds: number; periodSeconds: number; timeoutSeconds: number; failureThreshold: number }
interface Volume { name: string; type: string; hostPath: string; configMapName: string; secretName: string; pvcName: string }
interface Container {
  name: string; image: string; imagePullPolicy: string
  ports: Port[]; env: EnvVar[]; resources: Resources
  volumeMounts: VolumeMount[]; livenessProbe: Probe | null; readinessProbe: Probe | null
  securityContext: { runAsUser: number | null; runAsNonRoot: boolean; readOnlyRootFilesystem: boolean; privileged: boolean }
}
interface Tolerance { key: string; operator: string; value: string; effect: string; tolerationSeconds: number | null }
interface Annotation { key: string; value: string }

interface FormData {
  name: string; namespace: string; replicas: number; labels: Label[]
  containers: Container[]; volumes: Volume[]
  strategyType: string; maxSurge: string; maxUnavailable: string
  serviceName: string; updateStrategy: string; dsUpdateStrategy: string
  nodeSelector: Label[]; tolerations: Tolerance[]; annotations: Annotation[]
  serviceAccountName: string; terminationGracePeriodSeconds: number | null
  imagePullSecrets: string[]
}

function createEmptyProbe(): Probe {
  return { type: 'httpGet', httpGetPath: '/', httpGetPort: 80, tcpSocketPort: null, execCommand: '', initialDelaySeconds: 15, periodSeconds: 10, timeoutSeconds: 5, failureThreshold: 3 }
}

function createEmptyContainer(): Container {
  return {
    name: '', image: '', imagePullPolicy: 'IfNotPresent',
    ports: [], env: [],
    resources: { requests: { cpu: '', memory: '' }, limits: { cpu: '', memory: '' } },
    volumeMounts: [], livenessProbe: null, readinessProbe: null,
    securityContext: { runAsUser: null, runAsNonRoot: false, readOnlyRootFilesystem: false, privileged: false },
  }
}

const form = reactive<FormData>({
  name: '', namespace: 'default', replicas: 1,
  labels: [{ key: 'app', value: '' }],
  containers: [createEmptyContainer()],
  volumes: [],
  strategyType: 'RollingUpdate', maxSurge: '25%', maxUnavailable: '25%',
  serviceName: '', updateStrategy: 'RollingUpdate', dsUpdateStrategy: 'RollingUpdate',
  nodeSelector: [], tolerations: [], annotations: [],
  serviceAccountName: '', terminationGracePeriodSeconds: null, imagePullSecrets: [],
})

const step0FormRef = ref<FormInstance>()
const step0Rules: FormRules = {
  name: [
    { required: true, message: 'Name is required', trigger: 'blur' },
    { pattern: /^[a-z][a-z0-9-]*[a-z0-9]$/, message: 'Lowercase letters, numbers, hyphens only.', trigger: 'blur' },
  ],
  namespace: [{ required: true, message: 'Namespace is required', trigger: 'change' }],
  replicas: [{ required: true, message: 'Replicas is required', trigger: 'change' }],
}

const steps = [
  { title: 'Basic Info' },
  { title: 'Containers' },
  { title: 'Volumes & Mounts' },
  { title: 'Probes & Security' },
  { title: 'Advanced' },
  { title: 'YAML Preview' },
]

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
function addPort(ci: number) { form.containers[ci].ports.push({ name: '', containerPort: null, protocol: 'TCP' }) }
function removePort(ci: number, pi: number) { form.containers[ci].ports.splice(pi, 1) }
function addEnv(ci: number) { form.containers[ci].env.push({ name: '', value: '' }) }
function removeEnv(ci: number, ei: number) { form.containers[ci].env.splice(ei, 1) }
function addNodeSelector() { form.nodeSelector.push({ key: '', value: '' }) }
function removeNodeSelector(i: number) { form.nodeSelector.splice(i, 1) }

// Volume management
function addVolume() { form.volumes.push({ name: '', type: 'emptyDir', hostPath: '', configMapName: '', secretName: '', pvcName: '' }) }
function removeVolume(i: number) { form.volumes.splice(i, 1) }

// Volume mount management
function addVolumeMount(ci: number) { form.containers[ci].volumeMounts.push({ name: '', mountPath: '', subPath: '', readOnly: false }) }
function removeVolumeMount(ci: number, mi: number) { form.containers[ci].volumeMounts.splice(mi, 1) }

// Probe management
function enableLivenessProbe(ci: number) { form.containers[ci].livenessProbe = createEmptyProbe() }
function disableLivenessProbe(ci: number) { form.containers[ci].livenessProbe = null }
function enableReadinessProbe(ci: number) { form.containers[ci].readinessProbe = createEmptyProbe() }
function disableReadinessProbe(ci: number) { form.containers[ci].readinessProbe = null }

// Toleration management
function addToleration() { form.tolerations.push({ key: '', operator: 'Equal', value: '', effect: 'NoSchedule', tolerationSeconds: null }) }
function removeToleration(i: number) { form.tolerations.splice(i, 1) }

// Annotation management
function addAnnotation() { form.annotations.push({ key: '', value: '' }) }
function removeAnnotation(i: number) { form.annotations.splice(i, 1) }

// Image pull secrets
function addImagePullSecret() { form.imagePullSecrets.push('') }
function removeImagePullSecret(i: number) { form.imagePullSecrets.splice(i, 1) }

async function handleNext() {
  if (currentStep.value === 0) {
    const valid = await step0FormRef.value?.validate().catch(() => false)
    if (!valid) return
  }
  if (currentStep.value === 1) {
    for (let i = 0; i < form.containers.length; i++) {
      if (!form.containers[i].name) { ElMessage.error(`Container ${i + 1}: name is required`); return }
      if (!form.containers[i].image) { ElMessage.error(`Container ${i + 1}: image is required`); return }
    }
  }
  if (currentStep.value < steps.length - 1) currentStep.value++
}

function handlePrev() { if (currentStep.value > 0) currentStep.value-- }
function handleStepClick(step: number) { if (step <= currentStep.value) currentStep.value = step }

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

  const containers = form.containers.map(c => {
    const container: Record<string, any> = { name: c.name, image: c.image, imagePullPolicy: c.imagePullPolicy }
    const ports = c.ports.filter(p => p.containerPort).map(p => { const port: any = { containerPort: p.containerPort, protocol: p.protocol }; if (p.name) port.name = p.name; return port })
    if (ports.length > 0) container.ports = ports
    const env = c.env.filter(e => e.name.trim()).map(e => ({ name: e.name, value: e.value }))
    if (env.length > 0) container.env = env
    const resources: any = {}; const requests: any = {}; const limits: any = {}
    if (c.resources.requests.cpu) requests.cpu = c.resources.requests.cpu
    if (c.resources.requests.memory) requests.memory = c.resources.requests.memory
    if (c.resources.limits.cpu) limits.cpu = c.resources.limits.cpu
    if (c.resources.limits.memory) limits.memory = c.resources.limits.memory
    if (Object.keys(requests).length > 0) resources.requests = requests
    if (Object.keys(limits).length > 0) resources.limits = limits
    if (Object.keys(resources).length > 0) container.resources = resources
    // Volume mounts
    const mounts = c.volumeMounts.filter(m => m.name && m.mountPath).map(m => { const vm: any = { name: m.name, mountPath: m.mountPath }; if (m.subPath) vm.subPath = m.subPath; if (m.readOnly) vm.readOnly = true; return vm })
    if (mounts.length > 0) container.volumeMounts = mounts
    // Probes
    const liveness = buildProbe(c.livenessProbe)
    if (liveness) container.livenessProbe = liveness
    const readiness = buildProbe(c.readinessProbe)
    if (readiness) container.readinessProbe = readiness
    // Security context
    const sc: any = {}
    if (c.securityContext.runAsUser !== null) sc.runAsUser = c.securityContext.runAsUser
    if (c.securityContext.runAsNonRoot) sc.runAsNonRoot = true
    if (c.securityContext.readOnlyRootFilesystem) sc.readOnlyRootFilesystem = true
    if (c.securityContext.privileged) sc.privileged = true
    if (Object.keys(sc).length > 0) container.securityContext = sc
    return container
  })

  // Volumes
  const volumes = form.volumes.filter(v => v.name).map(v => {
    const vol: any = { name: v.name }
    if (v.type === 'emptyDir') vol.emptyDir = {}
    else if (v.type === 'hostPath') vol.hostPath = { path: v.hostPath, type: 'DirectoryOrCreate' }
    else if (v.type === 'configMap') vol.configMap = { name: v.configMapName || v.name }
    else if (v.type === 'secret') vol.secret = { secretName: v.secretName || v.name }
    else if (v.type === 'pvc') vol.persistentVolumeClaim = { claimName: v.pvcName || v.name }
    return vol
  })

  const nodeSelector: Record<string, string> = {}
  form.nodeSelector.forEach(ns => { if (ns.key.trim()) nodeSelector[ns.key.trim()] = ns.value })

  // Annotations
  const annotations: Record<string, string> = {}
  form.annotations.forEach(a => { if (a.key.trim()) annotations[a.key.trim()] = a.value })

  // Tolerations
  const tolerations = form.tolerations.filter(t => t.key).map(t => {
    const tol: any = { key: t.key, operator: t.operator, effect: t.effect }
    if (t.value) tol.value = t.value
    if (t.tolerationSeconds) tol.tolerationSeconds = t.tolerationSeconds
    return tol
  })

  // Image pull secrets
  const imagePullSecrets = form.imagePullSecrets.filter(s => s).map(s => ({ name: s }))

  const podSpec: any = { containers }
  if (volumes.length > 0) podSpec.volumes = volumes
  if (Object.keys(nodeSelector).length > 0) podSpec.nodeSelector = nodeSelector
  if (tolerations.length > 0) podSpec.tolerations = tolerations
  if (form.serviceAccountName) podSpec.serviceAccountName = form.serviceAccountName
  if (form.terminationGracePeriodSeconds) podSpec.terminationGracePeriodSeconds = form.terminationGracePeriodSeconds
  if (imagePullSecrets.length > 0) podSpec.imagePullSecrets = imagePullSecrets

  const podTemplate: any = { metadata: { labels: { ...labels } }, spec: podSpec }
  if (Object.keys(annotations).length > 0) podTemplate.metadata.annotations = annotations

  const resource: any = { apiVersion: 'apps/v1', kind: props.kind, metadata: { name: form.name, namespace: form.namespace, labels: { ...labels } }, spec: {} }

  if (props.kind === 'Deployment') {
    resource.spec = { replicas: form.replicas, selector: { matchLabels: { ...labels } }, template: podTemplate, strategy: { type: form.strategyType } }
    if (form.strategyType === 'RollingUpdate') resource.spec.strategy.rollingUpdate = { maxSurge: form.maxSurge, maxUnavailable: form.maxUnavailable }
  } else if (props.kind === 'StatefulSet') {
    resource.spec = { replicas: form.replicas, selector: { matchLabels: { ...labels } }, template: podTemplate, serviceName: form.serviceName || form.name, updateStrategy: { type: form.updateStrategy } }
  } else if (props.kind === 'DaemonSet') {
    resource.spec = { selector: { matchLabels: { ...labels } }, template: podTemplate, updateStrategy: { type: form.dsUpdateStrategy } }
  }

  return resource
}

async function handleSubmit() {
  submitting.value = true
  try {
    await (props.kind === 'Deployment' ? createDeployment : props.kind === 'StatefulSet' ? createStatefulSet : createDaemonSet)({ namespace: form.namespace, yaml: generatedYaml.value })
    ElMessage.success(`${props.kind} created successfully`)
    router.push(getListRoute())
  } catch (e: any) { ElMessage.error(e?.message || 'Create failed') }
  finally { submitting.value = false }
}

function getListRoute(): string {
  return props.kind === 'Deployment' ? '/workloads/deployments' : props.kind === 'StatefulSet' ? '/workloads/statefulsets' : '/workloads/daemonsets'
}

function handleCancel() { router.push(getListRoute()) }
</script>

<template>
  <div class="workload-form">
    <div class="form-header"><h2>Create {{ kind }}</h2></div>
    <el-steps :active="currentStep" finish-status="success" align-center style="margin-bottom: 32px;">
      <el-step v-for="(step, index) in steps" :key="index" :title="step.title" @click="handleStepClick(index)" style="cursor: pointer;" />
    </el-steps>

    <div class="step-content">
      <!-- Step 0: Basic Info -->
      <div v-show="currentStep === 0">
        <el-form ref="step0FormRef" :model="form" :rules="step0Rules" label-width="140px" style="max-width: 700px;">
          <el-form-item label="Name" prop="name"><el-input v-model="form.name" placeholder="e.g. my-app" /></el-form-item>
          <el-form-item label="Namespace" prop="namespace">
            <el-select v-model="form.namespace" filterable placeholder="Select namespace" style="width: 100%;" :loading="namespaceLoading">
              <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
            </el-select>
          </el-form-item>
          <el-form-item label="Replicas" prop="replicas"><el-input-number v-model="form.replicas" :min="1" :max="1000" /></el-form-item>
          <el-form-item label="Labels">
            <div style="width: 100%;">
              <div v-for="(label, i) in form.labels" :key="i" style="display: flex; gap: 8px; margin-bottom: 8px;">
                <el-input v-model="label.key" placeholder="Key" style="flex: 1;" />
                <el-input v-model="label.value" placeholder="Value" style="flex: 1;" />
                <el-button type="danger" circle :disabled="form.labels.length <= 1" @click="removeLabel(i)">X</el-button>
              </div>
              <el-button @click="addLabel" size="small">+ Add Label</el-button>
            </div>
          </el-form-item>
          <el-form-item label="Annotations">
            <div style="width: 100%;">
              <div v-for="(ann, i) in form.annotations" :key="i" style="display: flex; gap: 8px; margin-bottom: 8px;">
                <el-input v-model="ann.key" placeholder="Key" style="flex: 1;" />
                <el-input v-model="ann.value" placeholder="Value" style="flex: 1;" />
                <el-button type="danger" circle size="small" @click="removeAnnotation(i)">X</el-button>
              </div>
              <el-button @click="addAnnotation" size="small">+ Add Annotation</el-button>
            </div>
          </el-form-item>
          <el-form-item label="Service Account">
            <el-input v-model="form.serviceAccountName" placeholder="default" />
          </el-form-item>
        </el-form>
      </div>

      <!-- Step 1: Containers -->
      <div v-show="currentStep === 1">
        <div v-for="(container, ci) in form.containers" :key="ci" class="container-card">
          <div class="container-card-header">
            <h4>Container {{ ci + 1 }}: {{ container.name || '(unnamed)' }}</h4>
            <el-button v-if="form.containers.length > 1" type="danger" size="small" @click="removeContainer(ci)">Remove</el-button>
          </div>
          <el-form label-width="140px" style="max-width: 700px;">
            <el-form-item label="Container Name" required><el-input v-model="container.name" placeholder="e.g. nginx" /></el-form-item>
            <el-form-item label="Image" required><el-input v-model="container.image" placeholder="e.g. nginx:1.25" /></el-form-item>
            <el-form-item label="Pull Policy">
              <el-select v-model="container.imagePullPolicy" style="width: 100%;">
                <el-option label="Always" value="Always" /><el-option label="IfNotPresent" value="IfNotPresent" /><el-option label="Never" value="Never" />
              </el-select>
            </el-form-item>
            <el-form-item label="Ports">
              <div style="width: 100%;">
                <div v-for="(port, pi) in container.ports" :key="pi" style="display: flex; gap: 8px; margin-bottom: 8px; align-items: center;">
                  <el-input v-model="port.name" placeholder="Name" style="width: 120px;" />
                  <el-input-number v-model="port.containerPort" :min="1" :max="65535" placeholder="Port" style="width: 160px;" />
                  <el-select v-model="port.protocol" style="width: 110px;"><el-option label="TCP" value="TCP" /><el-option label="UDP" value="UDP" /></el-select>
                  <el-button type="danger" circle size="small" @click="removePort(ci, pi)">X</el-button>
                </div>
                <el-button size="small" @click="addPort(ci)">+ Add Port</el-button>
              </div>
            </el-form-item>
            <el-form-item label="Env Variables">
              <div style="width: 100%;">
                <div v-for="(env, ei) in container.env" :key="ei" style="display: flex; gap: 8px; margin-bottom: 8px;">
                  <el-input v-model="env.name" placeholder="Name" style="flex: 1;" />
                  <el-input v-model="env.value" placeholder="Value" style="flex: 1;" />
                  <el-button type="danger" circle size="small" @click="removeEnv(ci, ei)">X</el-button>
                </div>
                <el-button size="small" @click="addEnv(ci)">+ Add Env</el-button>
              </div>
            </el-form-item>
            <el-form-item label="Requests">
              <div style="display: flex; gap: 16px; width: 100%;">
                <div style="flex: 1;"><div class="resource-label">CPU</div><el-input v-model="container.resources.requests.cpu" placeholder="e.g. 100m" /></div>
                <div style="flex: 1;"><div class="resource-label">Memory</div><el-input v-model="container.resources.requests.memory" placeholder="e.g. 128Mi" /></div>
              </div>
            </el-form-item>
            <el-form-item label="Limits">
              <div style="display: flex; gap: 16px; width: 100%;">
                <div style="flex: 1;"><div class="resource-label">CPU</div><el-input v-model="container.resources.limits.cpu" placeholder="e.g. 500m" /></div>
                <div style="flex: 1;"><div class="resource-label">Memory</div><el-input v-model="container.resources.limits.memory" placeholder="e.g. 512Mi" /></div>
              </div>
            </el-form-item>
          </el-form>
        </div>
        <el-button @click="addContainer">+ Add Container</el-button>
      </div>

      <!-- Step 2: Volumes & Mounts -->
      <div v-show="currentStep === 2">
        <h4>Volumes</h4>
        <div v-for="(vol, vi) in form.volumes" :key="vi" class="container-card" style="margin-bottom: 12px;">
          <div style="display: flex; gap: 8px; align-items: center; margin-bottom: 8px;">
            <el-input v-model="vol.name" placeholder="Volume Name" style="flex: 1;" />
            <el-select v-model="vol.type" style="width: 160px;">
              <el-option label="emptyDir" value="emptyDir" /><el-option label="hostPath" value="hostPath" />
              <el-option label="ConfigMap" value="configMap" /><el-option label="Secret" value="secret" />
              <el-option label="PVC" value="pvc" />
            </el-select>
            <el-button type="danger" circle size="small" @click="removeVolume(vi)">X</el-button>
          </div>
          <el-input v-if="vol.type === 'hostPath'" v-model="vol.hostPath" placeholder="Host path (e.g. /data)" style="margin-bottom: 8px;" />
          <el-input v-if="vol.type === 'configMap'" v-model="vol.configMapName" placeholder="ConfigMap name" style="margin-bottom: 8px;" />
          <el-input v-if="vol.type === 'secret'" v-model="vol.secretName" placeholder="Secret name" style="margin-bottom: 8px;" />
          <el-input v-if="vol.type === 'pvc'" v-model="vol.pvcName" placeholder="PVC name" style="margin-bottom: 8px;" />
        </div>
        <el-button @click="addVolume">+ Add Volume</el-button>

        <h4 style="margin-top: 24px;">Volume Mounts (per container)</h4>
        <div v-for="(container, ci) in form.containers" :key="ci" style="margin-bottom: 16px;">
          <h5>{{ container.name || `Container ${ci + 1}` }}</h5>
          <div v-for="(mount, mi) in container.volumeMounts" :key="mi" style="display: flex; gap: 8px; margin-bottom: 8px; align-items: center;">
            <el-select v-model="mount.name" placeholder="Volume" style="width: 160px;">
              <el-option v-for="v in form.volumes.filter(v => v.name)" :key="v.name" :label="v.name" :value="v.name" />
            </el-select>
            <el-input v-model="mount.mountPath" placeholder="Mount path" style="flex: 1;" />
            <el-input v-model="mount.subPath" placeholder="Sub path" style="width: 120px;" />
            <el-checkbox v-model="mount.readOnly">RO</el-checkbox>
            <el-button type="danger" circle size="small" @click="removeVolumeMount(ci, mi)">X</el-button>
          </div>
          <el-button size="small" @click="addVolumeMount(ci)">+ Add Mount</el-button>
        </div>
      </div>

      <!-- Step 3: Probes & Security -->
      <div v-show="currentStep === 3">
        <div v-for="(container, ci) in form.containers" :key="ci" style="margin-bottom: 24px;">
          <h4>{{ container.name || `Container ${ci + 1}` }}</h4>

          <!-- Liveness Probe -->
          <el-card shadow="never" style="margin-bottom: 12px;">
            <template #header>
              <div style="display: flex; justify-content: space-between; align-items: center;">
                <span>Liveness Probe</span>
                <el-button v-if="!container.livenessProbe" size="small" type="primary" @click="enableLivenessProbe(ci)">Enable</el-button>
                <el-button v-else size="small" type="danger" @click="disableLivenessProbe(ci)">Disable</el-button>
              </div>
            </template>
            <template v-if="container.livenessProbe">
              <el-form label-width="120px" size="small">
                <el-form-item label="Type">
                  <el-select v-model="container.livenessProbe.type"><el-option label="HTTP GET" value="httpGet" /><el-option label="TCP Socket" value="tcpSocket" /><el-option label="Exec" value="exec" /></el-select>
                </el-form-item>
                <el-form-item v-if="container.livenessProbe.type === 'httpGet'" label="Path"><el-input v-model="container.livenessProbe.httpGetPath" placeholder="/" /></el-form-item>
                <el-form-item v-if="container.livenessProbe.type === 'httpGet'" label="Port"><el-input-number v-model="container.livenessProbe.httpGetPort" :min="1" :max="65535" /></el-form-item>
                <el-form-item v-if="container.livenessProbe.type === 'tcpSocket'" label="Port"><el-input-number v-model="container.livenessProbe.tcpSocketPort" :min="1" :max="65535" /></el-form-item>
                <el-form-item v-if="container.livenessProbe.type === 'exec'" label="Command"><el-input v-model="container.livenessProbe.execCommand" placeholder="e.g. cat /tmp/healthy" /></el-form-item>
                <el-form-item label="Initial Delay"><el-input-number v-model="container.livenessProbe.initialDelaySeconds" :min="0" />s</el-form-item>
                <el-form-item label="Period"><el-input-number v-model="container.livenessProbe.periodSeconds" :min="1" />s</el-form-item>
              </el-form>
            </template>
          </el-card>

          <!-- Readiness Probe -->
          <el-card shadow="never" style="margin-bottom: 12px;">
            <template #header>
              <div style="display: flex; justify-content: space-between; align-items: center;">
                <span>Readiness Probe</span>
                <el-button v-if="!container.readinessProbe" size="small" type="primary" @click="enableReadinessProbe(ci)">Enable</el-button>
                <el-button v-else size="small" type="danger" @click="disableReadinessProbe(ci)">Disable</el-button>
              </div>
            </template>
            <template v-if="container.readinessProbe">
              <el-form label-width="120px" size="small">
                <el-form-item label="Type">
                  <el-select v-model="container.readinessProbe.type"><el-option label="HTTP GET" value="httpGet" /><el-option label="TCP Socket" value="tcpSocket" /><el-option label="Exec" value="exec" /></el-select>
                </el-form-item>
                <el-form-item v-if="container.readinessProbe.type === 'httpGet'" label="Path"><el-input v-model="container.readinessProbe.httpGetPath" placeholder="/" /></el-form-item>
                <el-form-item v-if="container.readinessProbe.type === 'httpGet'" label="Port"><el-input-number v-model="container.readinessProbe.httpGetPort" :min="1" :max="65535" /></el-form-item>
                <el-form-item v-if="container.readinessProbe.type === 'tcpSocket'" label="Port"><el-input-number v-model="container.readinessProbe.tcpSocketPort" :min="1" :max="65535" /></el-form-item>
                <el-form-item v-if="container.readinessProbe.type === 'exec'" label="Command"><el-input v-model="container.readinessProbe.execCommand" placeholder="e.g. cat /tmp/healthy" /></el-form-item>
                <el-form-item label="Initial Delay"><el-input-number v-model="container.readinessProbe.initialDelaySeconds" :min="0" />s</el-form-item>
                <el-form-item label="Period"><el-input-number v-model="container.readinessProbe.periodSeconds" :min="1" />s</el-form-item>
              </el-form>
            </template>
          </el-card>

          <!-- Security Context -->
          <el-card shadow="never">
            <template #header><span>Security Context</span></template>
            <el-form label-width="160px" size="small">
              <el-form-item label="Run as User"><el-input-number v-model="container.securityContext.runAsUser" :min="0" placeholder="UID" /></el-form-item>
              <el-form-item label="Run as Non-Root"><el-switch v-model="container.securityContext.runAsNonRoot" /></el-form-item>
              <el-form-item label="Read-Only Root FS"><el-switch v-model="container.securityContext.readOnlyRootFilesystem" /></el-form-item>
              <el-form-item label="Privileged"><el-switch v-model="container.securityContext.privileged" /></el-form-item>
            </el-form>
          </el-card>
        </div>
      </div>

      <!-- Step 4: Advanced -->
      <div v-show="currentStep === 4">
        <el-form label-width="180px" style="max-width: 700px;">
          <template v-if="kind === 'Deployment'">
            <el-form-item label="Strategy Type"><el-select v-model="form.strategyType" style="width: 100%;"><el-option label="RollingUpdate" value="RollingUpdate" /><el-option label="Recreate" value="Recreate" /></el-select></el-form-item>
            <el-form-item v-if="form.strategyType === 'RollingUpdate'" label="Max Surge"><el-input v-model="form.maxSurge" placeholder="e.g. 25% or 2" /></el-form-item>
            <el-form-item v-if="form.strategyType === 'RollingUpdate'" label="Max Unavailable"><el-input v-model="form.maxUnavailable" placeholder="e.g. 25% or 1" /></el-form-item>
          </template>
          <template v-if="kind === 'StatefulSet'">
            <el-form-item label="Service Name"><el-input v-model="form.serviceName" placeholder="Headless service name" /></el-form-item>
            <el-form-item label="Update Strategy"><el-select v-model="form.updateStrategy" style="width: 100%;"><el-option label="RollingUpdate" value="RollingUpdate" /><el-option label="OnDelete" value="OnDelete" /></el-select></el-form-item>
          </template>
          <template v-if="kind === 'DaemonSet'">
            <el-form-item label="Update Strategy"><el-select v-model="form.dsUpdateStrategy" style="width: 100%;"><el-option label="RollingUpdate" value="RollingUpdate" /><el-option label="OnDelete" value="OnDelete" /></el-select></el-form-item>
          </template>
          <el-form-item label="Node Selector">
            <div style="width: 100%;">
              <div v-for="(ns, i) in form.nodeSelector" :key="i" style="display: flex; gap: 8px; margin-bottom: 8px;">
                <el-input v-model="ns.key" placeholder="Key" style="flex: 1;" /><el-input v-model="ns.value" placeholder="Value" style="flex: 1;" />
                <el-button type="danger" circle size="small" @click="removeNodeSelector(i)">X</el-button>
              </div>
              <el-button size="small" @click="addNodeSelector">+ Add Node Selector</el-button>
            </div>
          </el-form-item>
          <el-form-item label="Tolerations">
            <div style="width: 100%;">
              <div v-for="(tol, i) in form.tolerations" :key="i" style="display: flex; gap: 8px; margin-bottom: 8px; align-items: center;">
                <el-input v-model="tol.key" placeholder="Key" style="flex: 1;" />
                <el-select v-model="tol.operator" style="width: 100px;"><el-option label="Equal" value="Equal" /><el-option label="Exists" value="Exists" /></el-select>
                <el-input v-model="tol.value" placeholder="Value" style="flex: 1;" />
                <el-select v-model="tol.effect" style="width: 140px;"><el-option label="NoSchedule" value="NoSchedule" /><el-option label="PreferNoSchedule" value="PreferNoSchedule" /><el-option label="NoExecute" value="NoExecute" /></el-select>
                <el-button type="danger" circle size="small" @click="removeToleration(i)">X</el-button>
              </div>
              <el-button size="small" @click="addToleration">+ Add Toleration</el-button>
            </div>
          </el-form-item>
          <el-form-item label="Termination Grace">
            <el-input-number v-model="form.terminationGracePeriodSeconds" :min="0" :max="300" placeholder="seconds" />
          </el-form-item>
          <el-form-item label="Image Pull Secrets">
            <div style="width: 100%;">
              <div v-for="(_s, i) in form.imagePullSecrets" :key="i" style="display: flex; gap: 8px; margin-bottom: 8px;">
                <el-input v-model="form.imagePullSecrets[i]" placeholder="Secret name" style="flex: 1;" />
                <el-button type="danger" circle size="small" @click="removeImagePullSecret(i)">X</el-button>
              </div>
              <el-button size="small" @click="addImagePullSecret">+ Add Secret</el-button>
            </div>
          </el-form-item>
        </el-form>
      </div>

      <!-- Step 5: YAML Preview -->
      <div v-show="currentStep === 5">
        <el-alert :title="`Generated ${kind} YAML`" description="Review the generated YAML before creating." type="info" :closable="false" show-icon style="margin-bottom: 16px;" />
        <YamlEditor :model-value="generatedYaml" height="500px" read-only />
      </div>
    </div>

    <div class="form-actions">
      <el-button @click="handleCancel">Cancel</el-button>
      <el-button v-if="currentStep > 0" @click="handlePrev">Previous</el-button>
      <el-button v-if="currentStep < steps.length - 1" type="primary" @click="handleNext">Next</el-button>
      <el-button v-if="currentStep === steps.length - 1" type="primary" :loading="submitting" @click="handleSubmit">Create {{ kind }}</el-button>
    </div>
  </div>
</template>

<style scoped>
.workload-form { max-width: 900px; margin: 0 auto; padding: 20px 0; }
.form-header { margin-bottom: 24px; }
.form-header h2 { margin: 0; font-size: 20px; font-weight: 600; }
.step-content { min-height: 400px; padding: 16px 0; }
.container-card { border: 1px solid #e4e7ed; border-radius: 8px; padding: 20px; margin-bottom: 16px; background: #fafafa; }
.container-card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; padding-bottom: 12px; border-bottom: 1px solid #e4e7ed; }
.container-card-header h4 { margin: 0; font-size: 15px; font-weight: 600; color: #303133; }
.resource-label { font-size: 12px; color: #909399; margin-bottom: 4px; }
.form-actions { display: flex; justify-content: flex-end; gap: 12px; padding-top: 24px; border-top: 1px solid #e4e7ed; margin-top: 24px; }
</style>
