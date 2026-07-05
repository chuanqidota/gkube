<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import yaml from 'js-yaml'
import type { FormInstance, FormRules } from 'element-plus'
import YamlEditor from '@/components/YamlEditor.vue'
import { getNamespaceList, createService, extractNamespaceNames } from '@/api/resource'

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

interface ServicePort {
  name: string
  port: number | null
  targetPort: number | null
  protocol: string
  nodePort: number | null
}

interface Selector {
  key: string
  value: string
}

interface FormData {
  name: string
  namespace: string
  type: string
  labels: Label[]
  ports: ServicePort[]
  selectors: Selector[]
}

const form = reactive<FormData>({
  name: '',
  namespace: 'default',
  type: 'ClusterIP',
  labels: [{ key: 'app', value: '' }],
  ports: [{ name: '', port: 80, targetPort: 80, protocol: 'TCP', nodePort: null }],
  selectors: [{ key: 'app', value: '' }],
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
  type: [{ required: true, message: 'Service type is required', trigger: 'change' }],
}

// ---- Steps ----

const steps = [
  { title: 'Basic Info', icon: 'Document' },
  { title: 'Ports & Selector', icon: 'Connection' },
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

// ---- Port Management ----

function addPort() {
  form.ports.push({ name: '', port: null, targetPort: null, protocol: 'TCP', nodePort: null })
}

function removePort(index: number) {
  if (form.ports.length <= 1) {
    ElMessage.warning('At least one port is required')
    return
  }
  form.ports.splice(index, 1)
}

// ---- Selector Management ----

function addSelector() {
  form.selectors.push({ key: '', value: '' })
}

function removeSelector(index: number) {
  form.selectors.splice(index, 1)
}

// ---- Step Navigation ----

async function handleNext() {
  if (currentStep.value === 0) {
    const valid = await step0FormRef.value?.validate().catch(() => false)
    if (!valid) return
  }
  if (currentStep.value === 1) {
    const valid = validatePorts()
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

function validatePorts(): boolean {
  for (let i = 0; i < form.ports.length; i++) {
    const p = form.ports[i]
    if (!p.port) {
      ElMessage.error(`Port ${i + 1}: port number is required`)
      return false
    }
    if (!p.targetPort) {
      ElMessage.error(`Port ${i + 1}: targetPort is required`)
      return false
    }
  }
  return true
}

// ---- YAML Generation ----

const generatedYaml = computed(() => {
  const resource = buildK8sService()
  return yaml.dump(resource, { indent: 2, lineWidth: -1, noRefs: true })
})

function buildK8sService(): Record<string, any> {
  const labels: Record<string, string> = {}
  form.labels.forEach((l) => {
    if (l.key.trim()) labels[l.key.trim()] = l.value
  })

  const selector: Record<string, string> = {}
  form.selectors.forEach((s) => {
    if (s.key.trim()) selector[s.key.trim()] = s.value
  })

  const ports = form.ports
    .filter((p) => p.port && p.targetPort)
    .map((p) => {
      const port: Record<string, any> = {
        port: p.port,
        targetPort: p.targetPort,
        protocol: p.protocol,
      }
      if (p.name) port.name = p.name
      if (form.type === 'NodePort' && p.nodePort) port.nodePort = p.nodePort
      return port
    })

  const resource: Record<string, any> = {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: {
      name: form.name,
      namespace: form.namespace,
      labels: { ...labels },
    },
    spec: {
      type: form.type,
      selector,
      ports,
    },
  }

  return resource
}

// ---- Submit ----

async function handleSubmit() {
  submitting.value = true
  try {
    const yaml = generatedYaml.value
    await createService({
      namespace: form.namespace,
      yaml,
    })
    ElMessage.success('Service created successfully')
    router.push('/services')
  } catch (e: any) {
    ElMessage.error(e?.message || 'Create failed')
  } finally {
    submitting.value = false
  }
}

function handleCancel() {
  router.push('/services')
}
</script>

<template>
  <div class="service-form">
    <!-- Header -->
    <div class="form-header">
      <h2>创建 Service</h2>
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
          label-width="140px"
          style="max-width: 700px;"
        >
          <el-form-item label="Name" prop="name">
            <el-input v-model="form.name" placeholder="e.g. my-service" />
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

          <el-form-item label="Type" prop="type">
            <el-select v-model="form.type" style="width: 100%;">
              <el-option label="ClusterIP" value="ClusterIP" />
              <el-option label="NodePort" value="NodePort" />
              <el-option label="LoadBalancer" value="LoadBalancer" />
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
        </el-form>
      </div>

      <!-- Step 1: Ports & Selector -->
      <div v-show="currentStep === 1">
        <el-form label-width="140px" style="max-width: 700px;">
          <!-- Ports -->
          <el-form-item label="Ports" required>
            <div style="width: 100%;">
              <div
                v-for="(port, pi) in form.ports"
                :key="pi"
                class="port-card"
              >
                <div style="display: flex; gap: 8px; flex-wrap: wrap; align-items: center;">
                  <el-input v-model="port.name" placeholder="Name (opt)" style="width: 120px;" />
                  <el-input-number v-model="port.port" :min="1" :max="65535" placeholder="Port" style="width: 140px;" />
                  <el-input-number v-model="port.targetPort" :min="1" :max="65535" placeholder="Target Port" style="width: 140px;" />
                  <el-select v-model="port.protocol" style="width: 100px;">
                    <el-option label="TCP" value="TCP" />
                    <el-option label="UDP" value="UDP" />
                    <el-option label="SCTP" value="SCTP" />
                  </el-select>
                  <el-input-number
                    v-if="form.type === 'NodePort'"
                    v-model="port.nodePort"
                    :min="30000"
                    :max="32767"
                    placeholder="NodePort"
                    style="width: 140px;"
                  />
                  <el-button type="danger" circle size="small" @click="removePort(pi)">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
              </div>
              <el-button size="small" @click="addPort">
                <el-icon><Plus /></el-icon> Add Port
              </el-button>
            </div>
          </el-form-item>

          <!-- Selector -->
          <el-form-item label="Selector">
            <div style="width: 100%;">
              <div
                v-for="(sel, index) in form.selectors"
                :key="index"
                style="display: flex; gap: 8px; margin-bottom: 8px;"
              >
                <el-input v-model="sel.key" placeholder="Key" style="flex: 1;" />
                <el-input v-model="sel.value" placeholder="Value" style="flex: 1;" />
                <el-button
                  type="danger"
                  circle
                  :disabled="form.selectors.length <= 1"
                  @click="removeSelector(index)"
                >
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
              <el-button @click="addSelector" size="small">
                <el-icon><Plus /></el-icon> Add Selector
              </el-button>
            </div>
          </el-form-item>
        </el-form>
      </div>

      <!-- Step 2: YAML Preview -->
      <div v-show="currentStep === 2">
        <el-alert
          title="Generated Service YAML"
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
      <el-button @click="handleCancel">取消</el-button>
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
        Create Service
      </el-button>
    </div>
  </div>
</template>

<style scoped>
.service-form {
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

.port-card {
  border: 1px solid var(--gk-color-border);
  border-radius: 6px;
  padding: 12px;
  margin-bottom: 8px;
  background: var(--gk-neutral-50);
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 24px;
  border-top: 1px solid var(--gk-color-border);
  margin-top: 24px;
}
</style>
