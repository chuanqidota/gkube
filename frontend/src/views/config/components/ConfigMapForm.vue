<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Delete, Plus } from '@element-plus/icons-vue'
import yaml from 'js-yaml'
import type { FormInstance, FormRules } from 'element-plus'
import { createConfigMap, updateConfigMap, getNamespaceList, extractNamespaceNames } from '@/api/resource'

const props = defineProps<{
  isEdit?: boolean
  initialData?: any
}>()

const emit = defineEmits<{
  success: []
  cancel: []
}>()

const router = useRouter()
const formRef = ref<FormInstance>()
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
  labels: Array<{ key: string; value: string }>
  data: DataEntry[]
  binaryData: DataEntry[]
  immutable: boolean
}

const form = reactive<FormData>({
  name: '',
  namespace: 'default',
  labels: [],
  data: [],
  binaryData: [],
  immutable: false,
})

// ---- Parse initial data for edit mode ----

function parseInitialData(data: any) {
  if (!data) return
  const meta = data.metadata || {}

  form.name = meta.name || data.name || ''
  form.namespace = meta.namespace || data.namespace || 'default'

  // Labels
  const labels = meta.labels || data.labels || {}
  form.labels = Object.entries(labels).map(([k, v]) => ({ key: k, value: String(v) }))
  if (form.labels.length === 0) form.labels.push({ key: '', value: '' })

  // Data entries
  const entries = data.data || {}
  form.data = Object.entries(entries).map(([k, v]) => ({ key: k, value: String(v ?? '') }))
  if (form.data.length === 0) form.data.push({ key: '', value: '' })

  // Binary data entries
  const binEntries = data.binaryData || {}
  form.binaryData = Object.entries(binEntries).map(([k, v]) => ({ key: k, value: String(v ?? '') }))
  if (form.binaryData.length === 0) form.binaryData.push({ key: '', value: '' })

  // Immutable
  form.immutable = data.immutable || false
}

if (props.isEdit && props.initialData) {
  parseInitialData(props.initialData)
} else {
  form.labels = [{ key: '', value: '' }]
  form.data = [{ key: '', value: '' }]
}

// ---- Validation ----

const rules: FormRules = {
  name: [
    { required: true, message: '请输入名称', trigger: 'blur' },
    { pattern: /^[a-z][a-z0-9-]*[a-z0-9]$/, message: '仅支持小写字母、数字和连字符', trigger: 'blur' },
    { max: 253, message: '最多253个字符', trigger: 'blur' },
  ],
  namespace: [{ required: true, message: '请选择命名空间', trigger: 'change' }],
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

// ---- Data Entry Management ----

function addEntry() {
  form.data.push({ key: '', value: '' })
}

function removeEntry(index: number) {
  if (form.data.length <= 1) return
  form.data.splice(index, 1)
}

// ---- Build & Submit ----

function buildYamlStr(): string {
  const data: Record<string, string> = {}
  form.data.forEach((entry) => {
    if (entry.key.trim()) data[entry.key.trim()] = entry.value
  })

  const binaryData: Record<string, string> = {}
  form.binaryData.forEach((entry) => {
    if (entry.key.trim()) binaryData[entry.key.trim()] = entry.value
  })

  const labels: Record<string, string> = {}
  form.labels.forEach((l) => {
    if (l.key.trim()) labels[l.key.trim()] = l.value
  })

  const obj: any = {
    apiVersion: 'v1',
    kind: 'ConfigMap',
    metadata: {
      name: form.name,
      namespace: form.namespace,
      ...(Object.keys(labels).length > 0 ? { labels } : {}),
    },
    data,
  }

  if (Object.keys(binaryData).length > 0) obj.binaryData = binaryData
  if (form.immutable) obj.immutable = true

  return yaml.dump(obj, { indent: 2, lineWidth: -1, noRefs: true })
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    const yamlStr = buildYamlStr()
    if (props.isEdit) {
      await updateConfigMap({ namespace: form.namespace, name: form.name, yaml: yamlStr })
      ElMessage.success('ConfigMap 更新成功')
      emit('success')
    } else {
      await createConfigMap({ namespace: form.namespace, yaml: yamlStr })
      ElMessage.success('ConfigMap 创建成功')
      router.push('/config/configmaps')
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
    router.push('/config/configmaps')
  }
}
</script>

<template>
  <div class="cm-form">
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
      <!-- Section 1: Basic Info -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">基本信息</div>
        </div>
        <div class="section-content">
          <div class="fields-grid">
            <el-form-item label="名称" prop="name">
              <el-input v-model="form.name" :disabled="isEdit" placeholder="my-config" />
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
        </div>
      </div>

      <!-- Section 3: Data -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">数据</div>
        </div>
        <div class="section-content">
          <el-form-item label="不可变">
            <div style="width: 100%;">
              <el-switch v-model="form.immutable" />
              <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-top: 4px;">设置后不可修改，只能删除重建</div>
            </div>
          </el-form-item>
          <el-form-item label="数据项">
            <div style="width: 100%;">
              <div v-for="(entry, i) in form.data" :key="i" class="data-row">
                <el-input v-model="entry.key" placeholder="Key" style="width: 220px;" />
                <el-input v-model="entry.value" type="textarea" :rows="2" placeholder="Value" style="flex: 1;" />
                <el-button type="danger" text circle :disabled="form.data.length <= 1" @click="removeEntry(i)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
              <el-button text type="primary" @click="addEntry" size="small">
                <el-icon><Plus /></el-icon> 添加数据项
              </el-button>
            </div>
          </el-form-item>
          <el-form-item label="二进制数据">
            <div style="width: 100%;">
              <div v-for="(entry, i) in form.binaryData" :key="i" class="data-row">
                <el-input v-model="entry.key" placeholder="Key" style="width: 220px;" />
                <el-input v-model="entry.value" type="textarea" :rows="2" placeholder="Base64 编码值" style="flex: 1;" />
                <el-button type="danger" text circle :disabled="form.binaryData.length <= 1" @click="form.binaryData.splice(i, 1)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
              <el-button text type="primary" @click="form.binaryData.push({ key: '', value: '' })" size="small">
                <el-icon><Plus /></el-icon> 添加二进制数据项
              </el-button>
            </div>
          </el-form-item>
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
.cm-form {
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

/* Data rows (key + textarea value) */
.data-row {
  display: flex;
  gap: 8px;
  align-items: flex-start;
  margin-bottom: 8px;
}
</style>
