<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import yaml from 'js-yaml'
import type { FormInstance, FormRules } from 'element-plus'
import YamlEditor from '@/components/YamlEditor.vue'
import { createPv } from '@/api/resource'

const router = useRouter()
const currentStep = ref(0)
const submitting = ref(false)

// ---- Form Data ----

interface Label {
  key: string
  value: string
}

interface FormData {
  name: string
  capacity: string
  accessModes: string[]
  storageClassName: string
  reclaimPolicy: string
  storageType: string
  nfsServer: string
  nfsPath: string
  hostPath: string
  localPath: string
  labels: Label[]
}

const form = reactive<FormData>({
  name: '',
  capacity: '10Gi',
  accessModes: ['ReadWriteOnce'],
  storageClassName: '',
  reclaimPolicy: 'Retain',
  storageType: 'nfs',
  nfsServer: '',
  nfsPath: '',
  hostPath: '',
  localPath: '',
  labels: [{ key: '', value: '' }],
})

// ---- Validation ----

const step0FormRef = ref<FormInstance>()

const step0Rules: FormRules = {
  name: [
    { required: true, message: 'Name is required', trigger: 'blur' },
    { pattern: /^[a-z][a-z0-9-]*[a-z0-9]$/, message: 'Lowercase letters, numbers, hyphens only. Must start with letter, end with alphanumeric.', trigger: 'blur' },
    { max: 253, message: 'Max 253 characters', trigger: 'blur' },
  ],
  capacity: [
    { required: true, message: 'Capacity is required', trigger: 'blur' },
  ],
  accessModes: [
    { type: 'array', required: true, message: 'At least one access mode is required', trigger: 'change' },
  ],
}

// ---- Steps ----

const steps = [
  { title: 'Basic Info', icon: 'Document' },
  { title: 'Storage Source', icon: 'FolderOpened' },
  { title: 'YAML Preview', icon: 'View' },
]

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
    if (!validateStorageSource()) return
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

function validateStorageSource(): boolean {
  if (form.storageType === 'nfs') {
    if (!form.nfsServer.trim()) {
      ElMessage.error('NFS server is required')
      return false
    }
    if (!form.nfsPath.trim()) {
      ElMessage.error('NFS path is required')
      return false
    }
  } else if (form.storageType === 'hostPath') {
    if (!form.hostPath.trim()) {
      ElMessage.error('Host path is required')
      return false
    }
  } else if (form.storageType === 'local') {
    if (!form.localPath.trim()) {
      ElMessage.error('Local path is required')
      return false
    }
  }
  return true
}

// ---- YAML Generation ----

const generatedYaml = computed(() => {
  const resource = buildK8sPV()
  return yaml.dump(resource, { indent: 2, lineWidth: -1, noRefs: true })
})

function buildK8sPV(): Record<string, any> {
  const labels: Record<string, string> = {}
  form.labels.forEach((l) => {
    if (l.key.trim()) labels[l.key.trim()] = l.value
  })

  // Parse capacity - extract numeric value and unit
  const capacityStr = form.capacity.trim()
  const capacity: Record<string, string> = {}
  if (capacityStr) {
    capacity['storage'] = capacityStr
  }

  const spec: Record<string, any> = {
    capacity,
    accessModes: form.accessModes,
    persistentVolumeReclaimPolicy: form.reclaimPolicy,
  }

  if (form.storageClassName) {
    spec.storageClassName = form.storageClassName
  }

  // Storage source
  if (form.storageType === 'nfs') {
    spec.nfs = {
      server: form.nfsServer.trim(),
      path: form.nfsPath.trim(),
    }
  } else if (form.storageType === 'hostPath') {
    spec.hostPath = {
      path: form.hostPath.trim(),
    }
  } else if (form.storageType === 'local') {
    spec.local = {
      path: form.localPath.trim(),
    }
  }

  const resource: Record<string, any> = {
    apiVersion: 'v1',
    kind: 'PersistentVolume',
    metadata: {
      name: form.name,
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
    await createPv({ yaml: yamlContent })
    ElMessage.success('PersistentVolume created successfully')
    router.push('/storage/pvs')
  } catch (e: any) {
    ElMessage.error(e?.message || 'Create failed')
  } finally {
    submitting.value = false
  }
}

function handleCancel() {
  router.push('/storage/pvs')
}
</script>

<template>
  <div class="pv-form">
    <!-- Header -->
    <div class="form-header">
      <h2>Create PersistentVolume</h2>
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
            <el-input v-model="form.name" placeholder="e.g. my-pv" />
          </el-form-item>

          <el-form-item label="Capacity" prop="capacity">
            <el-input v-model="form.capacity" placeholder="e.g. 10Gi" />
          </el-form-item>

          <el-form-item label="Access Modes" prop="accessModes">
            <el-checkbox-group v-model="form.accessModes">
              <el-checkbox label="ReadWriteOnce" value="ReadWriteOnce" />
              <el-checkbox label="ReadOnlyMany" value="ReadOnlyMany" />
              <el-checkbox label="ReadWriteMany" value="ReadWriteMany" />
            </el-checkbox-group>
          </el-form-item>

          <el-form-item label="Storage Class Name">
            <el-input v-model="form.storageClassName" placeholder="Leave empty for no storage class" />
          </el-form-item>

          <el-form-item label="Reclaim Policy">
            <el-select v-model="form.reclaimPolicy" style="width: 100%;">
              <el-option label="Retain" value="Retain" />
              <el-option label="Recycle" value="Recycle" />
              <el-option label="Delete" value="Delete" />
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

      <!-- Step 1: Storage Source -->
      <div v-show="currentStep === 1">
        <el-form label-width="180px" style="max-width: 700px;">
          <el-form-item label="Storage Type" required>
            <el-select v-model="form.storageType" style="width: 100%;">
              <el-option label="NFS" value="nfs" />
              <el-option label="Host Path" value="hostPath" />
              <el-option label="Local" value="local" />
            </el-select>
          </el-form-item>

          <!-- NFS -->
          <template v-if="form.storageType === 'nfs'">
            <el-form-item label="NFS Server" required>
              <el-input v-model="form.nfsServer" placeholder="e.g. 10.0.0.1" />
            </el-form-item>
            <el-form-item label="NFS Path" required>
              <el-input v-model="form.nfsPath" placeholder="e.g. /exports/data" />
            </el-form-item>
          </template>

          <!-- Host Path -->
          <template v-if="form.storageType === 'hostPath'">
            <el-form-item label="Host Path" required>
              <el-input v-model="form.hostPath" placeholder="e.g. /mnt/data" />
            </el-form-item>
          </template>

          <!-- Local -->
          <template v-if="form.storageType === 'local'">
            <el-form-item label="Local Path" required>
              <el-input v-model="form.localPath" placeholder="e.g. /mnt/disks/ssd1" />
            </el-form-item>
          </template>
        </el-form>
      </div>

      <!-- Step 2: YAML Preview -->
      <div v-show="currentStep === 2">
        <el-alert
          title="Generated PersistentVolume YAML"
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
        Create PersistentVolume
      </el-button>
    </div>
  </div>
</template>

<style scoped>
.pv-form {
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
