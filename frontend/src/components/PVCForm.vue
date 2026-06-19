<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import yaml from 'js-yaml'
import type { FormInstance, FormRules } from 'element-plus'
import YamlEditor from '@/components/YamlEditor.vue'
import { createPvc, getNamespaceList, extractNamespaceNames } from '@/api/resource'

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

interface FormData {
  name: string
  namespace: string
  labels: Label[]
  storageClassName: string
  accessModes: string[]
  storageRequestSize: string
  storageRequestUnit: string
}

const form = reactive<FormData>({
  name: '',
  namespace: 'default',
  labels: [{ key: '', value: '' }],
  storageClassName: '',
  accessModes: ['ReadWriteOnce'],
  storageRequestSize: '10',
  storageRequestUnit: 'Gi',
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
  { title: 'Storage Request', icon: 'Coin' },
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

// ---- Step Navigation ----

async function handleNext() {
  if (currentStep.value === 0) {
    const valid = await step0FormRef.value?.validate().catch(() => false)
    if (!valid) return
  }
  if (currentStep.value === 1) {
    if (!validateStorageRequest()) return
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

function validateStorageRequest(): boolean {
  if (!form.storageRequestSize.trim()) {
    ElMessage.error('Storage request size is required')
    return false
  }
  const num = parseFloat(form.storageRequestSize)
  if (isNaN(num) || num <= 0) {
    ElMessage.error('Storage request size must be a positive number')
    return false
  }
  if (form.accessModes.length === 0) {
    ElMessage.error('At least one access mode is required')
    return false
  }
  return true
}

// ---- YAML Generation ----

const generatedYaml = computed(() => {
  const resource = buildK8sPVC()
  return yaml.dump(resource, { indent: 2, lineWidth: -1, noRefs: true })
})

function buildK8sPVC(): Record<string, any> {
  const labels: Record<string, string> = {}
  form.labels.forEach((l) => {
    if (l.key.trim()) labels[l.key.trim()] = l.value
  })

  const storageSize = `${form.storageRequestSize}${form.storageRequestUnit}`

  const spec: Record<string, any> = {
    accessModes: form.accessModes,
    resources: {
      requests: {
        storage: storageSize,
      },
    },
  }

  if (form.storageClassName) {
    spec.storageClassName = form.storageClassName
  }

  const resource: Record<string, any> = {
    apiVersion: 'v1',
    kind: 'PersistentVolumeClaim',
    metadata: {
      name: form.name,
      namespace: form.namespace,
      labels: { ...labels },
    },
    spec,
  }

  return resource
}

// ---- Submit ----

async function handleSubmit() {
  submitting.value = true
  try {
    const yamlContent = generatedYaml.value
    await createPvc({ namespace: form.namespace, yamlContent })
    ElMessage.success('PersistentVolumeClaim created successfully')
    router.push('/storage/pvcs')
  } catch (e: any) {
    ElMessage.error(e?.message || 'Create failed')
  } finally {
    submitting.value = false
  }
}

function handleCancel() {
  router.push('/storage/pvcs')
}
</script>

<template>
  <div class="pvc-form">
    <!-- Header -->
    <div class="form-header">
      <h2>Create PersistentVolumeClaim</h2>
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
          label-width="160px"
          style="max-width: 700px;"
        >
          <el-form-item label="Name" prop="name">
            <el-input v-model="form.name" placeholder="e.g. my-pvc" />
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
        </el-form>
      </div>

      <!-- Step 1: Storage Request -->
      <div v-show="currentStep === 1">
        <el-form label-width="160px" style="max-width: 700px;">
          <el-form-item label="Storage Class Name">
            <el-input v-model="form.storageClassName" placeholder="Leave empty for default storage class" />
          </el-form-item>

          <el-form-item label="Access Modes" required>
            <el-checkbox-group v-model="form.accessModes">
              <el-checkbox label="ReadWriteOnce" value="ReadWriteOnce" />
              <el-checkbox label="ReadOnlyMany" value="ReadOnlyMany" />
              <el-checkbox label="ReadWriteMany" value="ReadWriteMany" />
            </el-checkbox-group>
          </el-form-item>

          <el-form-item label="Storage Request" required>
            <div style="display: flex; gap: 8px; width: 100%;">
              <el-input
                v-model="form.storageRequestSize"
                placeholder="e.g. 10"
                style="flex: 1;"
              />
              <el-select v-model="form.storageRequestUnit" style="width: 100px;">
                <el-option label="Mi" value="Mi" />
                <el-option label="Gi" value="Gi" />
                <el-option label="Ti" value="Ti" />
              </el-select>
            </div>
          </el-form-item>
        </el-form>
      </div>

      <!-- Step 2: YAML Preview -->
      <div v-show="currentStep === 2">
        <el-alert
          title="Generated PersistentVolumeClaim YAML"
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
        Create PersistentVolumeClaim
      </el-button>
    </div>
  </div>
</template>

<style scoped>
.pvc-form {
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

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 24px;
  border-top: 1px solid #e4e7ed;
  margin-top: 24px;
}
</style>
