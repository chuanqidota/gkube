<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import yaml from 'js-yaml'
import type { FormInstance, FormRules } from 'element-plus'
import YamlEditor from '@/components/YamlEditor.vue'
import { getNamespaceList, createJob, extractNamespaceNames } from '@/api/resource'

const router = useRouter()
const currentStep = ref(0)
const submitting = ref(false)
const namespaceLoading = ref(false)
const namespaces = ref<string[]>([])

// ---- Form Data ----

interface Label {
  key: string
  value: string
}

interface Port {
  name: string
  containerPort: number | null
  protocol: string
}

interface EnvVar {
  name: string
  value: string
}

interface Resources {
  requests: { cpu: string; memory: string }
  limits: { cpu: string; memory: string }
}

interface Container {
  name: string
  image: string
  imagePullPolicy: string
  ports: Port[]
  env: EnvVar[]
  resources: Resources
}

interface FormData {
  name: string
  namespace: string
  labels: Label[]
  completions: number | null
  parallelism: number | null
  backoffLimit: number | null
  activeDeadlineSeconds: number | null
  containers: Container[]
}

function createEmptyContainer(): Container {
  return {
    name: '',
    image: '',
    imagePullPolicy: 'IfNotPresent',
    ports: [],
    env: [],
    resources: {
      requests: { cpu: '', memory: '' },
      limits: { cpu: '', memory: '' },
    },
  }
}

const form = reactive<FormData>({
  name: '',
  namespace: 'default',
  labels: [{ key: 'app', value: '' }],
  completions: 1,
  parallelism: 1,
  backoffLimit: 6,
  activeDeadlineSeconds: null,
  containers: [createEmptyContainer()],
})

// ---- Validation ----

const step0FormRef = ref<FormInstance>()

const step0Rules: FormRules = {
  name: [
    { required: true, message: 'Name is required', trigger: 'blur' },
    { pattern: /^[a-z][a-z0-9-]*[a-z0-9]$/, message: 'Lowercase letters, numbers, hyphens only. Must start with letter, end with alphanumeric.', trigger: 'blur' },
    { max: 253, message: 'Max 253 characters', trigger: 'blur' },
  ],
  namespace: [{ required: true, message: 'Namespace is required', trigger: 'change' }],
}

// ---- Steps ----

const steps = [
  { title: 'Basic Info', icon: 'Document' },
  { title: 'Container Config', icon: 'Box' },
  { title: 'YAML Preview', icon: 'View' },
]

// ---- Namespace Fetch ----

async function fetchNamespaces() {
  namespaceLoading.value = true
  try {
    const res: any = await getNamespaceList()
    namespaces.value = extractNamespaceNames(res.data)
  } catch {
    namespaces.value = ['default']
  } finally {
    namespaceLoading.value = false
  }
}

onMounted(fetchNamespaces)

// ---- Label Management ----

function addLabel() {
  form.labels.push({ key: '', value: '' })
}

function removeLabel(index: number) {
  form.labels.splice(index, 1)
}

// ---- Container Management ----

function addContainer() {
  form.containers.push(createEmptyContainer())
}

function removeContainer(index: number) {
  if (form.containers.length <= 1) {
    ElMessage.warning('At least one container is required')
    return
  }
  form.containers.splice(index, 1)
}

// ---- Port Management ----

function addPort(containerIndex: number) {
  form.containers[containerIndex].ports.push({ name: '', containerPort: null, protocol: 'TCP' })
}

function removePort(containerIndex: number, portIndex: number) {
  form.containers[containerIndex].ports.splice(portIndex, 1)
}

// ---- Env Management ----

function addEnv(containerIndex: number) {
  form.containers[containerIndex].env.push({ name: '', value: '' })
}

function removeEnv(containerIndex: number, envIndex: number) {
  form.containers[containerIndex].env.splice(envIndex, 1)
}

// ---- Step Navigation ----

async function handleNext() {
  if (currentStep.value === 0) {
    const valid = await step0FormRef.value?.validate().catch(() => false)
    if (!valid) return
  }
  if (currentStep.value === 1) {
    const valid = await validateContainers()
    if (!valid) return
  }
  if (currentStep.value < 2) {
    currentStep.value++
  }
}

function handlePrev() {
  if (currentStep.value > 0) {
    currentStep.value--
  }
}

function handleStepClick(step: number) {
  if (step <= currentStep.value) {
    currentStep.value = step
  }
}

async function validateContainers(): Promise<boolean> {
  for (let i = 0; i < form.containers.length; i++) {
    const c = form.containers[i]
    if (!c.name) {
      ElMessage.error(`Container ${i + 1}: name is required`)
      return false
    }
    if (!c.image) {
      ElMessage.error(`Container ${i + 1}: image is required`)
      return false
    }
  }
  return true
}

// ---- YAML Generation ----

const generatedYaml = computed(() => {
  const resource = buildK8sResource()
  return yaml.dump(resource, { indent: 2, lineWidth: -1, noRefs: true })
})

function buildK8sResource(): Record<string, any> {
  // Build labels
  const labels: Record<string, string> = {}
  form.labels.forEach((l) => {
    if (l.key.trim()) labels[l.key.trim()] = l.value
  })

  // Build containers
  const containers = form.containers.map((c) => {
    const container: Record<string, any> = {
      name: c.name,
      image: c.image,
      imagePullPolicy: c.imagePullPolicy,
    }

    // Ports
    const ports = c.ports
      .filter((p) => p.containerPort)
      .map((p) => {
        const port: Record<string, any> = { containerPort: p.containerPort, protocol: p.protocol }
        if (p.name) port.name = p.name
        return port
      })
    if (ports.length > 0) container.ports = ports

    // Env
    const env = c.env
      .filter((e) => e.name.trim())
      .map((e) => ({ name: e.name, value: e.value }))
    if (env.length > 0) container.env = env

    // Resources
    const resources: Record<string, any> = {}
    const requests: Record<string, string> = {}
    const limits: Record<string, string> = {}
    if (c.resources.requests.cpu) requests.cpu = c.resources.requests.cpu
    if (c.resources.requests.memory) requests.memory = c.resources.requests.memory
    if (c.resources.limits.cpu) limits.cpu = c.resources.limits.cpu
    if (c.resources.limits.memory) limits.memory = c.resources.limits.memory
    if (Object.keys(requests).length > 0) resources.requests = requests
    if (Object.keys(limits).length > 0) resources.limits = limits
    if (Object.keys(resources).length > 0) container.resources = resources

    return container
  })

  // Build resource
  const resource: Record<string, any> = {
    apiVersion: 'batch/v1',
    kind: 'Job',
    metadata: {
      name: form.name,
      namespace: form.namespace,
      labels: { ...labels },
    },
    spec: {
      template: {
        metadata: {
          labels: { ...labels },
        },
        spec: {
          containers,
          restartPolicy: 'Never',
        },
      },
    },
  }

  if (form.completions !== null && form.completions !== undefined) {
    resource.spec.completions = form.completions
  }
  if (form.parallelism !== null && form.parallelism !== undefined) {
    resource.spec.parallelism = form.parallelism
  }
  if (form.backoffLimit !== null && form.backoffLimit !== undefined) {
    resource.spec.backoffLimit = form.backoffLimit
  }
  if (form.activeDeadlineSeconds !== null && form.activeDeadlineSeconds !== undefined) {
    resource.spec.activeDeadlineSeconds = form.activeDeadlineSeconds
  }

  return resource
}

// ---- Submit ----

async function handleSubmit() {
  submitting.value = true
  try {
    const yaml = generatedYaml.value
    const data = {
      namespace: form.namespace,
      yaml,
    }

    await createJob(data)

    ElMessage.success('Job created successfully')
    router.push('/workloads/jobs')
  } catch (e: any) {
    ElMessage.error(e?.message || 'Create failed')
  } finally {
    submitting.value = false
  }
}

function handleCancel() {
  router.push('/workloads/jobs')
}
</script>

<template>
  <div class="workload-form">
    <!-- Header -->
    <div class="form-header">
      <h2>Create Job</h2>
    </div>

    <!-- Steps -->
    <el-steps :active="currentStep" finish-status="success" align-center style="margin-bottom: 32px;">
      <el-step
        v-for="(step, index) in steps"
        :key="index"
        :title="step.title"
        :icon="step.icon"
        @click="handleStepClick(index)"
        style="cursor: pointer;"
      />
    </el-steps>

    <!-- Step Content -->
    <div class="step-content">

      <!-- Step 0: Basic Info -->
      <div v-show="currentStep === 0">
        <el-form
          ref="step0FormRef"
          :model="form"
          :rules="step0Rules"
          label-width="180px"
          style="max-width: 700px;"
        >
          <el-form-item label="Name" prop="name">
            <el-input v-model="form.name" placeholder="e.g. my-job" />
          </el-form-item>

          <el-form-item label="Namespace" prop="namespace">
            <el-select
              v-model="form.namespace"
              filterable
              placeholder="Select namespace"
              style="width: 100%;"
              :loading="namespaceLoading"
            >
              <el-option
                v-for="ns in namespaces"
                :key="ns"
                :label="ns"
                :value="ns"
              />
            </el-select>
          </el-form-item>

          <el-form-item label="Labels">
            <div style="width: 100%;">
              <div
                v-for="(label, index) in form.labels"
                :key="index"
                style="display: flex; gap: 8px; margin-bottom: 8px;"
              >
                <el-input v-model="label.key" placeholder="Key" style="flex: 1;" />
                <el-input v-model="label.value" placeholder="Value" style="flex: 1;" />
                <el-button
                  type="danger"
                  circle
                  :disabled="form.labels.length <= 1"
                  @click="removeLabel(index)"
                >
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
              <el-button @click="addLabel" size="small">
                <el-icon><Plus /></el-icon> Add Label
              </el-button>
            </div>
          </el-form-item>

          <el-form-item label="Completions">
            <el-input-number v-model="form.completions" :min="1" />
          </el-form-item>

          <el-form-item label="Parallelism">
            <el-input-number v-model="form.parallelism" :min="1" />
          </el-form-item>

          <el-form-item label="Backoff Limit">
            <el-input-number v-model="form.backoffLimit" :min="0" />
          </el-form-item>

          <el-form-item label="Active Deadline (s)">
            <el-input-number v-model="form.activeDeadlineSeconds" :min="1" placeholder="No limit" />
          </el-form-item>
        </el-form>
      </div>

      <!-- Step 1: Container Config -->
      <div v-show="currentStep === 1">
        <div
          v-for="(container, ci) in form.containers"
          :key="ci"
          class="container-card"
        >
          <div class="container-card-header">
            <h4>Container {{ ci + 1 }}: {{ container.name || '(unnamed)' }}</h4>
            <el-button
              v-if="form.containers.length > 1"
              type="danger"
              size="small"
              @click="removeContainer(ci)"
            >
              Remove
            </el-button>
          </div>

          <el-form label-width="140px" style="max-width: 700px;">
            <el-form-item label="Container Name" required>
              <el-input v-model="container.name" placeholder="e.g. my-task" />
            </el-form-item>

            <el-form-item label="Image" required>
              <el-input v-model="container.image" placeholder="e.g. busybox:1.36" />
            </el-form-item>

            <el-form-item label="Pull Policy">
              <el-select v-model="container.imagePullPolicy" style="width: 100%;">
                <el-option label="Always" value="Always" />
                <el-option label="IfNotPresent" value="IfNotPresent" />
                <el-option label="Never" value="Never" />
              </el-select>
            </el-form-item>

            <!-- Ports -->
            <el-form-item label="Ports">
              <div style="width: 100%;">
                <div
                  v-for="(port, pi) in container.ports"
                  :key="pi"
                  style="display: flex; gap: 8px; margin-bottom: 8px; align-items: center;"
                >
                  <el-input v-model="port.name" placeholder="Name (opt)" style="width: 120px;" />
                  <el-input-number v-model="port.containerPort" :min="1" :max="65535" placeholder="Port" style="width: 160px;" />
                  <el-select v-model="port.protocol" style="width: 110px;">
                    <el-option label="TCP" value="TCP" />
                    <el-option label="UDP" value="UDP" />
                    <el-option label="SCTP" value="SCTP" />
                  </el-select>
                  <el-button type="danger" circle size="small" @click="removePort(ci, pi)">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
                <el-button size="small" @click="addPort(ci)">
                  <el-icon><Plus /></el-icon> Add Port
                </el-button>
              </div>
            </el-form-item>

            <!-- Environment Variables -->
            <el-form-item label="Env Variables">
              <div style="width: 100%;">
                <div
                  v-for="(env, ei) in container.env"
                  :key="ei"
                  style="display: flex; gap: 8px; margin-bottom: 8px;"
                >
                  <el-input v-model="env.name" placeholder="Name" style="flex: 1;" />
                  <el-input v-model="env.value" placeholder="Value" style="flex: 1;" />
                  <el-button type="danger" circle size="small" @click="removeEnv(ci, ei)">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
                <el-button size="small" @click="addEnv(ci)">
                  <el-icon><Plus /></el-icon> Add Env
                </el-button>
              </div>
            </el-form-item>

            <!-- Resource Requests -->
            <el-form-item label="Requests">
              <div style="display: flex; gap: 16px; width: 100%;">
                <div style="flex: 1;">
                  <div class="resource-label">CPU</div>
                  <el-input v-model="container.resources.requests.cpu" placeholder="e.g. 100m" />
                </div>
                <div style="flex: 1;">
                  <div class="resource-label">Memory</div>
                  <el-input v-model="container.resources.requests.memory" placeholder="e.g. 128Mi" />
                </div>
              </div>
            </el-form-item>

            <!-- Resource Limits -->
            <el-form-item label="Limits">
              <div style="display: flex; gap: 16px; width: 100%;">
                <div style="flex: 1;">
                  <div class="resource-label">CPU</div>
                  <el-input v-model="container.resources.limits.cpu" placeholder="e.g. 500m" />
                </div>
                <div style="flex: 1;">
                  <div class="resource-label">Memory</div>
                  <el-input v-model="container.resources.limits.memory" placeholder="e.g. 512Mi" />
                </div>
              </div>
            </el-form-item>
          </el-form>
        </div>

        <el-button @click="addContainer" style="margin-top: 8px;">
          <el-icon><Plus /></el-icon> Add Container
        </el-button>
      </div>

      <!-- Step 2: YAML Preview -->
      <div v-show="currentStep === 2">
        <el-alert
          title="Generated Job YAML"
          description="Review the generated YAML below before creating the resource."
          type="info"
          :closable="false"
          show-icon
          style="margin-bottom: 16px;"
        />
        <YamlEditor
          :model-value="generatedYaml"
          height="500px"
          read-only
        />
      </div>
    </div>

    <!-- Navigation Buttons -->
    <div class="form-actions">
      <el-button @click="handleCancel">Cancel</el-button>
      <el-button v-if="currentStep > 0" @click="handlePrev">Previous</el-button>
      <el-button
        v-if="currentStep < 2"
        type="primary"
        @click="handleNext"
      >
        Next
      </el-button>
      <el-button
        v-if="currentStep === 2"
        type="primary"
        :loading="submitting"
        @click="handleSubmit"
      >
        Create Job
      </el-button>
    </div>
  </div>
</template>

<style scoped>
.workload-form {
  max-width: 900px;
  margin: 0 auto;
  padding: 20px 0;
}

.form-header {
  margin-bottom: 24px;
}

.form-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.step-content {
  min-height: 400px;
  padding: 16px 0;
}

.container-card {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 16px;
  background: #fafafa;
}

.container-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e4e7ed;
}

.container-card-header h4 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: #303133;
}

.resource-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 24px;
  border-top: 1px solid #e4e7ed;
  margin-top: 24px;
}
</style>
