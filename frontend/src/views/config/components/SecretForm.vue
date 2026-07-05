<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import yaml from 'js-yaml'
import type { FormInstance, FormRules } from 'element-plus'
import YamlEditor from '@/components/YamlEditor.vue'
import { getNamespaceList, createSecret, extractNamespaceNames } from '@/api/resource'

const router = useRouter()
const currentStep = ref(0)
const submitting = ref(false)
const namespaceLoading = ref(false)
const namespaces = ref<string[]>([])

// ---- Form Data ----

interface DataEntry {
  key: string
  value: string
}

interface FormData {
  name: string
  namespace: string
  type: string
  data: DataEntry[]
}

const form = reactive<FormData>({
  name: '',
  namespace: 'default',
  type: 'Opaque',
  data: [{ key: '', value: '' }],
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
  type: [{ required: true, message: 'Secret type is required', trigger: 'change' }],
}

// ---- Steps ----

const steps = [
  { title: 'Basic Info', icon: 'Document' },
  { title: 'Data', icon: 'Coin' },
  { title: 'YAML Preview', icon: 'View' },
]

// ---- Secret Types ----

const secretTypes = [
  { label: 'Opaque', value: 'Opaque' },
  { label: 'TLS', value: 'kubernetes.io/tls' },
  { label: 'Docker Config JSON', value: 'kubernetes.io/dockerconfigjson' },
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

// ---- Data Entry Management ----

function addEntry() {
  form.data.push({ key: '', value: '' })
}

function removeEntry(index: number) {
  if (form.data.length <= 1) {
    ElMessage.warning('At least one data entry is required')
    return
  }
  form.data.splice(index, 1)
}

// ---- Step Navigation ----

async function handleNext() {
  if (currentStep.value === 0) {
    const valid = await step0FormRef.value?.validate().catch(() => false)
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

// ---- Base64 Encoding ----

function base64Encode(str: string): string {
  try {
    return btoa(str)
  } catch {
    // Handle UTF-8
    return btoa(unescape(encodeURIComponent(str)))
  }
}

// ---- YAML Generation ----

const generatedYaml = computed(() => {
  const resource = buildK8sSecret()
  return yaml.dump(resource, { indent: 2, lineWidth: -1, noRefs: true })
})

function buildK8sSecret(): Record<string, any> {
  const data: Record<string, string> = {}
  form.data.forEach((entry) => {
    if (entry.key.trim()) {
      data[entry.key.trim()] = base64Encode(entry.value)
    }
  })

  return {
    apiVersion: 'v1',
    kind: 'Secret',
    metadata: {
      name: form.name,
      namespace: form.namespace,
    },
    type: form.type,
    data,
  }
}

// ---- Submit ----

async function handleSubmit() {
  submitting.value = true
  try {
    const yaml = generatedYaml.value
    await createSecret({
      namespace: form.namespace,
      yaml,
    })
    ElMessage.success('Secret created successfully')
    router.push('/config/secrets')
  } catch (e: any) {
    ElMessage.error(e?.message || 'Create failed')
  } finally {
    submitting.value = false
  }
}

function handleCancel() {
  router.push('/config/secrets')
}
</script>

<template>
  <div class="secret-form">
    <!-- Header -->
    <div class="form-header">
      <h2>创建 Secret</h2>
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
            <el-input v-model="form.name" placeholder="e.g. my-secret" />
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
              <el-option
                v-for="t in secretTypes"
                :key="t.value"
                :label="t.label"
                :value="t.value"
              />
            </el-select>
          </el-form-item>
        </el-form>
      </div>

      <!-- Step 1: Data -->
      <div v-show="currentStep === 1">
        <el-form label-width="140px" style="max-width: 700px;">
          <el-form-item label="Data Entries">
            <div style="width: 100%;">
              <el-alert
                title="Values will be Base64 encoded automatically when generating the YAML."
                type="info"
                :closable="false"
                show-icon
                style="margin-bottom: 12px;"
              />
              <div
                v-for="(entry, index) in form.data"
                :key="index"
                class="data-entry-card"
              >
                <div style="display: flex; gap: 8px; align-items: flex-start;">
                  <el-input v-model="entry.key" placeholder="Key" style="width: 200px;" />
                  <el-input
                    v-model="entry.value"
                    type="textarea"
                    :rows="3"
                    placeholder="Value"
                    style="flex: 1;"
                  />
                  <el-button
                    type="danger"
                    circle
                    size="small"
                    :disabled="form.data.length <= 1"
                    @click="removeEntry(index)"
                    style="margin-top: 4px;"
                  >
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
              </div>
              <el-button size="small" @click="addEntry">
                <el-icon><Plus /></el-icon> Add Entry
              </el-button>
            </div>
          </el-form-item>
        </el-form>
      </div>

      <!-- Step 2: YAML Preview -->
      <div v-show="currentStep === 2">
        <el-alert
          title="Generated Secret YAML"
          description="Review the generated YAML below before creating the resource. Values are Base64 encoded."
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
        Create Secret
      </el-button>
    </div>
  </div>
</template>

<style scoped>
.secret-form {
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

.data-entry-card {
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
