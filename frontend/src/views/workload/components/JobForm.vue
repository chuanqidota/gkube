<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import yaml from 'js-yaml'
import type { FormInstance, FormRules } from 'element-plus'
import YamlEditor from '@/components/YamlEditor.vue'
import { getNamespaceList, extractNamespaceNames, createJob } from '@/api/resource'

const router = useRouter()
const submitting = ref(false)
const namespaceLoading = ref(false)
const namespaces = ref<string[]>([])
const showYamlPreview = ref(false)

interface Label { key: string; value: string }
interface Port { name: string; containerPort: number | null; protocol: string }
interface EnvVar { name: string; value: string }
interface Resources { requests: { cpu: string; memory: string }; limits: { cpu: string; memory: string } }
interface VolumeMount { name: string; mountPath: string; subPath: string; readOnly: boolean }
interface Probe { type: string; httpGetPath: string; httpGetPort: number | null; tcpSocketPort: number | null; execCommand: string; initialDelaySeconds: number; periodSeconds: number; timeoutSeconds: number; failureThreshold: number }
interface Volume { name: string; type: string; hostPath: string; configMapName: string; secretName: string; pvcName: string }
interface Tolerance { key: string; operator: string; value: string; effect: string; tolerationSeconds: number | null }
interface Annotation { key: string; value: string }
interface Container {
  name: string; image: string; imagePullPolicy: string
  ports: Port[]; env: EnvVar[]; resources: Resources
  volumeMounts: VolumeMount[]; livenessProbe: Probe | null; readinessProbe: Probe | null
  securityContext: { runAsUser: number | null; runAsNonRoot: boolean; readOnlyRootFilesystem: boolean; privileged: boolean }
}

interface FormData {
  name: string; namespace: string; labels: Label[]
  completions: number | null; parallelism: number | null; backoffLimit: number | null; activeDeadlineSeconds: number | null
  containers: Container[]; volumes: Volume[]
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
  name: '', namespace: 'default',
  labels: [{ key: 'app', value: '' }],
  completions: 1, parallelism: 1, backoffLimit: 6, activeDeadlineSeconds: null,
  containers: [createEmptyContainer()], volumes: [],
  nodeSelector: [], tolerations: [], annotations: [],
  serviceAccountName: '', terminationGracePeriodSeconds: null, imagePullSecrets: [],
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
function addVolume() { form.volumes.push({ name: '', type: 'emptyDir', hostPath: '', configMapName: '', secretName: '', pvcName: '' }) }
function removeVolume(i: number) { form.volumes.splice(i, 1) }
function addVolumeMount(ci: number) { form.containers[ci].volumeMounts.push({ name: '', mountPath: '', subPath: '', readOnly: false }) }
function removeVolumeMount(ci: number, mi: number) { form.containers[ci].volumeMounts.splice(mi, 1) }
function enableLivenessProbe(ci: number) { form.containers[ci].livenessProbe = createEmptyProbe() }
function disableLivenessProbe(ci: number) { form.containers[ci].livenessProbe = null }
function enableReadinessProbe(ci: number) { form.containers[ci].readinessProbe = createEmptyProbe() }
function disableReadinessProbe(ci: number) { form.containers[ci].readinessProbe = null }
function addToleration() { form.tolerations.push({ key: '', operator: 'Equal', value: '', effect: 'NoSchedule', tolerationSeconds: null }) }
function removeToleration(i: number) { form.tolerations.splice(i, 1) }
function addAnnotation() { form.annotations.push({ key: '', value: '' }) }
function removeAnnotation(i: number) { form.annotations.splice(i, 1) }
function addImagePullSecret() { form.imagePullSecrets.push('') }
function removeImagePullSecret(i: number) { form.imagePullSecrets.splice(i, 1) }

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
    const mounts = c.volumeMounts.filter(m => m.name && m.mountPath).map(m => { const vm: any = { name: m.name, mountPath: m.mountPath }; if (m.subPath) vm.subPath = m.subPath; if (m.readOnly) vm.readOnly = true; return vm })
    if (mounts.length > 0) container.volumeMounts = mounts
    const liveness = buildProbe(c.livenessProbe)
    if (liveness) container.livenessProbe = liveness
    const readiness = buildProbe(c.readinessProbe)
    if (readiness) container.readinessProbe = readiness
    const sc: any = {}
    if (c.securityContext.runAsUser !== null) sc.runAsUser = c.securityContext.runAsUser
    if (c.securityContext.runAsNonRoot) sc.runAsNonRoot = true
    if (c.securityContext.readOnlyRootFilesystem) sc.readOnlyRootFilesystem = true
    if (c.securityContext.privileged) sc.privileged = true
    if (Object.keys(sc).length > 0) container.securityContext = sc
    return container
  })

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
  if (volumes.length > 0) podSpec.volumes = volumes
  if (Object.keys(nodeSelector).length > 0) podSpec.nodeSelector = nodeSelector
  if (tolerations.length > 0) podSpec.tolerations = tolerations
  if (form.serviceAccountName) podSpec.serviceAccountName = form.serviceAccountName
  if (form.terminationGracePeriodSeconds) podSpec.terminationGracePeriodSeconds = form.terminationGracePeriodSeconds
  if (imagePullSecrets.length > 0) podSpec.imagePullSecrets = imagePullSecrets

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
    await createJob({ namespace: form.namespace, yaml: generatedYaml.value })
    ElMessage.success('Job 创建成功')
    router.push('/workloads/jobs')
  } catch (e: any) { ElMessage.error(e?.message || '创建失败') }
  finally { submitting.value = false }
}

function handleCancel() { router.push('/workloads/jobs') }
</script>

<template>
  <div class="workload-form">
    <!-- Page Header -->
    <div class="form-header">
      <div class="form-header-left">
        <el-button text @click="handleCancel" class="back-btn">
          <el-icon><ArrowLeft /></el-icon>
        </el-button>
        <div>
          <h2>创建 Job</h2>
          <p class="form-subtitle">填写以下信息来创建一个新的 Job 任务</p>
        </div>
      </div>
      <el-button @click="showYamlPreview = !showYamlPreview" class="yaml-preview-btn">
        <el-icon><Document /></el-icon>
        {{ showYamlPreview ? '隐藏 YAML' : '查看 YAML' }}
      </el-button>
    </div>

    <!-- YAML Preview Drawer -->
    <el-drawer v-model="showYamlPreview" title="YAML 预览" size="560px" direction="rtl">
      <YamlEditor :model-value="generatedYaml" height="calc(100vh - 120px)" read-only />
    </el-drawer>

    <el-form ref="formRef" :model="form" :rules="formRules" label-position="top">
      <div class="form-grid">

        <!-- Section 1: Basic Info -->
        <el-card shadow="never" class="form-section">
          <template #header>
            <div class="section-header">
              <div class="section-icon basic"><el-icon size="16"><Box /></el-icon></div>
              <div>
                <h3>基本信息</h3>
                <p>设置名称、命名空间和任务参数</p>
              </div>
            </div>
          </template>
          <div class="fields-grid">
            <el-form-item label="名称" prop="name">
              <el-input v-model="form.name" placeholder="my-job" />
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
          <el-divider />
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
        </el-card>

        <!-- Section 2: Job Config -->
        <el-card shadow="never" class="form-section">
          <template #header>
            <div class="section-header">
              <div class="section-icon scheduling"><el-icon size="16"><Setting /></el-icon></div>
              <div>
                <h3>任务配置</h3>
                <p>设置完成数、并行度和重试策略</p>
              </div>
            </div>
          </template>
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
          </div>
        </el-card>

        <!-- Section 3: Container Config -->
        <el-card shadow="never" class="form-section">
          <template #header>
            <div class="section-header">
              <div class="section-icon container"><el-icon size="16"><Cpu /></el-icon></div>
              <div>
                <h3>容器配置</h3>
                <p>定义镜像、端口、环境变量和资源</p>
              </div>
            </div>
          </template>
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
            <div v-for="(env, ei) in container.env" :key="ei" class="kv-row">
              <el-input v-model="env.name" placeholder="名称" />
              <el-input v-model="env.value" placeholder="值" />
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
        </el-card>

        <!-- Section 4: Storage -->
        <el-card shadow="never" class="form-section">
          <template #header>
            <div class="section-header">
              <div class="section-icon storage"><el-icon size="16"><Coin /></el-icon></div>
              <div>
                <h3>存储配置</h3>
                <p>配置数据卷和挂载</p>
              </div>
            </div>
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
                <el-input v-if="vol.type === 'hostPath'" v-model="vol.hostPath" placeholder="主机路径 (e.g. /data)" style="margin-top: 8px;" />
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
        </el-card>

        <!-- Section 5: Health Probes -->
        <el-card shadow="never" class="form-section">
          <template #header>
            <div class="section-header">
              <div class="section-icon probe"><el-icon size="16"><CircleCheck /></el-icon></div>
              <div>
                <h3>健康检查</h3>
                <p>配置存活探针和就绪探针</p>
              </div>
            </div>
          </template>

          <div v-for="(container, ci) in form.containers" :key="ci" style="margin-bottom: 24px;">
            <div class="mount-container-name">{{ container.name || `容器 ${ci + 1}` }}</div>

            <!-- Liveness -->
            <div class="probe-card">
              <div class="probe-header">
                <div>
                  <span class="probe-label">存活探针</span>
                  <span class="probe-desc">容器是否正在运行</span>
                </div>
                <el-switch :model-value="!!container.livenessProbe" @update:model-value="(v: boolean) => v ? enableLivenessProbe(ci) : disableLivenessProbe(ci)" />
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
                <el-switch :model-value="!!container.readinessProbe" @update:model-value="(v: boolean) => v ? enableReadinessProbe(ci) : disableReadinessProbe(ci)" />
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
          </div>
        </el-card>

        <!-- Section 6: Security -->
        <el-card shadow="never" class="form-section">
          <template #header>
            <div class="section-header">
              <div class="section-icon security"><el-icon size="16"><Lock /></el-icon></div>
              <div>
                <h3>安全设置</h3>
                <p>配置安全上下文</p>
              </div>
            </div>
          </template>

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
          </div>
        </el-card>

        <!-- Section 7: Scheduling -->
        <el-card shadow="never" class="form-section">
          <template #header>
            <div class="section-header">
              <div class="section-icon scheduling"><el-icon size="16"><Location /></el-icon></div>
              <div>
                <h3>调度配置</h3>
                <p>节点选择器和容忍规则</p>
              </div>
            </div>
          </template>

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
        </el-card>

        <!-- Section 8: Advanced -->
        <el-card shadow="never" class="form-section">
          <template #header>
            <div class="section-header">
              <div class="section-icon advanced"><el-icon size="16"><Setting /></el-icon></div>
              <div>
                <h3>高级配置</h3>
                <p>优雅终止时间、镜像密钥等</p>
              </div>
            </div>
          </template>

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
        </el-card>

      </div>
    </el-form>

    <!-- Bottom Action Bar -->
    <div class="form-actions">
      <el-button @click="handleCancel" size="large">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit" size="large">
        创建 Job
      </el-button>
    </div>
  </div>
</template>

<style scoped>
.workload-form {
  max-width: 1000px;
  margin: 0 auto;
  padding: 24px 0 100px;
}

.form-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding-bottom: 20px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.form-header-left {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.back-btn {
  margin-top: 2px;
}

.form-header h2 {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.form-subtitle {
  margin: 4px 0 0;
  font-size: 14px;
  color: var(--el-text-color-secondary);
}

.yaml-preview-btn {
  flex-shrink: 0;
}

.form-grid {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.form-section {
  border-radius: 12px;
  border: 1px solid var(--el-border-color-lighter);
}

.form-section :deep(.el-card__header) {
  padding: 14px 20px;
  border-bottom: 1px solid var(--el-border-color-extra-light);
  background: var(--el-fill-color-blank);
  border-radius: 12px 12px 0 0;
}

.form-section :deep(.el-card__body) {
  padding: 20px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.section-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.section-icon.basic {
  background: #ecf5ff;
  color: #409eff;
}

.section-icon.container {
  background: #f0f9eb;
  color: #67c23a;
}

.section-icon.storage {
  background: #fdf6ec;
  color: #e6a23c;
}

.section-icon.probe {
  background: #f0f9eb;
  color: #67c23a;
}

.section-icon.security {
  background: #fef0f0;
  color: #f56c6c;
}

.section-icon.scheduling {
  background: #f4f4f5;
  color: #909399;
}

.section-icon.advanced {
  background: #f4ecff;
  color: #9b59b6;
}

.section-header h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.section-header p {
  margin: 1px 0 0;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.fields-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0 24px;
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
  gap: 16px;
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
  gap: 12px;
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
  gap: 16px;
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

/* Bottom action bar */
.form-actions {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 32px;
  background: var(--el-bg-color);
  border-top: 1px solid var(--el-border-color-lighter);
  box-shadow: 0 -2px 12px rgba(0, 0, 0, 0.04);
  z-index: 100;
}
</style>
