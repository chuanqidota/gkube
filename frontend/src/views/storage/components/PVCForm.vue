<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Delete, Plus } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'
import { createPvc, updatePvcYaml, getNamespaceList, extractNamespaceNames } from '@/api/resource'

const props = defineProps<{
  isEdit?: boolean
  initialData?: any
}>()

const emit = defineEmits<{
  success: []
  cancel: []
}>()

const router = useRouter()
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
  volumeMode: string
  volumeName: string
  dataSourceType: string
  dataSourceName: string
  dataSourceKind: string
}

const form = reactive<FormData>({
  name: '',
  namespace: 'default',
  labels: [{ key: '', value: '' }],
  storageClassName: '',
  accessModes: ['ReadWriteOnce'],
  storageRequestSize: '10',
  storageRequestUnit: 'Gi',
  volumeMode: 'Filesystem',
  volumeName: '',
  dataSourceType: '',
  dataSourceName: '',
  dataSourceKind: 'VolumeSnapshot',
})

// ---- Parse initial data for edit mode ----

function parseInitialData(data: any) {
  if (!data) return
  const spec = data.spec || {}
  const meta = data.metadata || {}

  form.name = meta.name || ''
  form.namespace = meta.namespace || 'default'
  form.storageClassName = spec.storageClassName || ''
  form.accessModes = spec.accessModes || ['ReadWriteOnce']

  // Storage size
  const storage = spec.resources?.requests?.storage || ''
  if (storage) {
    const match = storage.match(/^(\d+(?:\.\d+)?)\s*(Mi|Gi|Ti|M|G|T)?$/i)
    if (match) {
      form.storageRequestSize = match[1]
      form.storageRequestUnit = (match[2] || 'Gi').charAt(0).toUpperCase() + (match[2] || 'Gi').slice(1).toLowerCase()
      // Normalize units
      if (form.storageRequestUnit === 'M') form.storageRequestUnit = 'Mi'
      if (form.storageRequestUnit === 'G') form.storageRequestUnit = 'Gi'
      if (form.storageRequestUnit === 'T') form.storageRequestUnit = 'Ti'
    }
  }

  // Labels
  const labels = meta.labels || {}
  form.labels = Object.entries(labels).map(([k, v]) => ({ key: k, value: String(v) }))
  if (form.labels.length === 0) form.labels = [{ key: '', value: '' }]

  // Volume mode
  form.volumeMode = spec.volumeMode || 'Filesystem'

  // Volume name (binding to specific PV)
  form.volumeName = spec.volumeName || ''

  // Data source
  if (spec.dataSource) {
    form.dataSourceType = spec.dataSource.kind === 'PersistentVolumeClaim' ? 'pvc' : 'snapshot'
    form.dataSourceName = spec.dataSource.name || ''
    form.dataSourceKind = spec.dataSource.kind || 'VolumeSnapshot'
  }
}

if (props.isEdit && props.initialData) {
  parseInitialData(props.initialData)
}

// ---- Validation ----

const formRef = ref<FormInstance>()

const formRules: FormRules = {
  name: [
    { required: true, message: '请输入名称', trigger: 'blur' },
    { pattern: /^[a-z][a-z0-9-]*[a-z0-9]$/, message: '仅支持小写字母、数字和连字符，必须以字母开头，以字母或数字结尾', trigger: 'blur' },
    { max: 253, message: '最多253个字符', trigger: 'blur' },
  ],
  namespace: [{ required: true, message: '请选择命名空间', trigger: 'change' }],
  storageRequestSize: [
    { required: true, message: '请输入存储大小', trigger: 'blur' },
  ],
}

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

// ---- Build & Submit ----

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

  if (form.volumeMode && form.volumeMode !== 'Filesystem') {
    spec.volumeMode = form.volumeMode
  }

  if (form.volumeName) {
    spec.volumeName = form.volumeName
  }

  if (form.dataSourceType && form.dataSourceName) {
    spec.dataSource = {
      name: form.dataSourceName,
      kind: form.dataSourceType === 'pvc' ? 'PersistentVolumeClaim' : 'VolumeSnapshot',
      apiGroup: form.dataSourceType === 'pvc' ? '' : 'snapshot.storage.k8s.io',
    }
  }

  return {
    apiVersion: 'v1',
    kind: 'PersistentVolumeClaim',
    metadata: {
      name: form.name,
      namespace: form.namespace,
      labels: { ...labels },
    },
    spec,
  }
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  if (!form.accessModes.length) {
    ElMessage.error('至少需要选择一种访问模式')
    return
  }

  submitting.value = true
  try {
    const yamlStr = (await import('js-yaml')).default.dump(buildK8sPVC(), { indent: 2, lineWidth: -1, noRefs: true })
    if (props.isEdit) {
      await updatePvcYaml({ namespace: form.namespace, yaml: yamlStr })
      ElMessage.success('PVC更新成功')
      emit('success')
    } else {
      await createPvc({ namespace: form.namespace, yaml: yamlStr })
      ElMessage.success('PVC创建成功')
      router.push('/storage/pvcs')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || (props.isEdit ? '更新失败' : '创建失败'))
  } finally {
    submitting.value = false
  }
}

function handleCancel() {
  if (props.isEdit) {
    emit('cancel')
  } else {
    router.push('/storage/pvcs')
  }
}
</script>

<template>
  <div class="pvc-form">
    <el-form ref="formRef" :model="form" :rules="formRules" label-position="top">
      <!-- Section 1: Basic Info -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">基本信息</div>
        </div>
        <div class="section-content">
          <div class="fields-grid">
            <el-form-item label="名称" prop="name">
              <el-input v-model="form.name" :disabled="isEdit" placeholder="例如: my-pvc" />
            </el-form-item>
            <el-form-item label="命名空间" prop="namespace">
              <el-select v-model="form.namespace" :disabled="isEdit" filterable placeholder="选择命名空间" style="width: 100%;" :loading="namespaceLoading">
                <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
              </el-select>
            </el-form-item>
          </div>
        </div>
      </div>

      <!-- Section 2: Labels -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">标签</div>
        </div>
        <div class="section-content">
          <el-form-item label="标签">
            <div style="width: 100%;">
              <div v-for="(label, i) in form.labels" :key="i" class="kv-row">
                <el-input v-model="label.key" placeholder="键" />
                <el-input v-model="label.value" placeholder="值" />
                <el-button type="danger" text circle :disabled="form.labels.length <= 1" @click="removeLabel(i)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
              <el-button text type="primary" @click="addLabel" size="small">
                <el-icon><Plus /></el-icon> 添加标签
              </el-button>
            </div>
          </el-form-item>
        </div>
      </div>

      <!-- Section 3: Storage Config -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">存储配置</div>
        </div>
        <div class="section-content">
          <div class="fields-grid">
            <el-form-item label="存储类名">
              <el-input v-model="form.storageClassName" placeholder="留空使用默认存储类" />
            </el-form-item>
            <el-form-item label="存储大小" prop="storageRequestSize">
              <div style="display: flex; gap: 8px; width: 100%;">
                <el-input v-model="form.storageRequestSize" placeholder="例如: 10" style="flex: 1;" />
                <el-select v-model="form.storageRequestUnit" style="width: 100px;">
                  <el-option label="Mi" value="Mi" />
                  <el-option label="Gi" value="Gi" />
                  <el-option label="Ti" value="Ti" />
                </el-select>
              </div>
            </el-form-item>
          </div>
          <el-form-item label="访问模式" required>
            <el-checkbox-group v-model="form.accessModes">
              <el-checkbox label="ReadWriteOnce" value="ReadWriteOnce" />
              <el-checkbox label="ReadOnlyMany" value="ReadOnlyMany" />
              <el-checkbox label="ReadWriteMany" value="ReadWriteMany" />
            </el-checkbox-group>
          </el-form-item>

          <div class="fields-grid">
            <el-form-item label="卷模式">
              <el-select v-model="form.volumeMode" style="width: 100%;">
                <el-option label="Filesystem (文件系统)" value="Filesystem" />
                <el-option label="Block (块设备)" value="Block" />
              </el-select>
            </el-form-item>
            <el-form-item label="指定 PV 名称">
              <el-input v-model="form.volumeName" placeholder="留空不绑定特定 PV" />
            </el-form-item>
          </div>

          <el-divider content-position="left">数据源 (从快照或 PVC 克隆)</el-divider>
          <div class="fields-grid">
            <el-form-item label="数据源类型">
              <el-select v-model="form.dataSourceType" style="width: 100%;" clearable placeholder="不使用数据源">
                <el-option label="VolumeSnapshot (快照)" value="snapshot" />
                <el-option label="PVC (克隆)" value="pvc" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="form.dataSourceType" label="数据源名称">
              <el-input v-model="form.dataSourceName" placeholder="快照或 PVC 名称" />
            </el-form-item>
          </div>
        </div>
      </div>

      <!-- Submit -->
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
.pvc-form {
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
</style>
