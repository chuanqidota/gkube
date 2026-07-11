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
}

const form = reactive<FormData>({
  name: '',
  namespace: 'default',
  labels: [],
  data: [],
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

  const labels: Record<string, string> = {}
  form.labels.forEach((l) => {
    if (l.key.trim()) labels[l.key.trim()] = l.value
  })

  const obj = {
    apiVersion: 'v1',
    kind: 'ConfigMap',
    metadata: {
      name: form.name,
      namespace: form.namespace,
      ...(Object.keys(labels).length > 0 ? { labels } : {}),
    },
    data,
  }
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
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="120px"
      style="max-width: 700px;"
    >
      <div class="form-section">
        <div class="section-title">基本信息</div>
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" :disabled="isEdit" placeholder="例如: my-config" />
        </el-form-item>
        <el-form-item label="命名空间" prop="namespace">
          <el-select v-model="form.namespace" :disabled="isEdit" filterable placeholder="选择命名空间" style="width: 100%;" :loading="namespaceLoading">
            <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
          </el-select>
        </el-form-item>
        <el-form-item label="标签">
          <div style="width: 100%;">
            <div v-for="(label, i) in form.labels" :key="i" style="display: flex; gap: 8px; margin-bottom: 8px;">
              <el-input v-model="label.key" placeholder="键" style="flex: 1;" />
              <el-input v-model="label.value" placeholder="值" style="flex: 1;" />
              <el-button type="danger" circle :disabled="form.labels.length <= 1" @click="removeLabel(i)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <el-button @click="addLabel" size="small">
              <el-icon><Plus /></el-icon> 添加标签
            </el-button>
          </div>
        </el-form-item>
      </div>

      <div class="form-section">
        <div class="section-title">数据</div>
        <el-form-item label="数据项">
          <div style="width: 100%;">
            <div v-for="(entry, i) in form.data" :key="i" class="data-entry-row">
              <el-input v-model="entry.key" placeholder="Key" style="width: 200px;" />
              <el-input v-model="entry.value" type="textarea" :rows="2" placeholder="Value" style="flex: 1;" />
              <el-button type="danger" circle :disabled="form.data.length <= 1" @click="removeEntry(i)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <el-button @click="addEntry" size="small">
              <el-icon><Plus /></el-icon> 添加数据项
            </el-button>
          </div>
        </el-form-item>
      </div>
    </el-form>

    <div class="form-actions">
      <el-button @click="handleCancel">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ isEdit ? '更新' : '创建' }}</el-button>
    </div>
  </div>
</template>

<style scoped>
.cm-form {
  max-width: 900px;
  margin: 0 auto;
  padding: 20px 0;
}

.form-section {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.data-entry-row {
  display: flex;
  gap: 8px;
  align-items: flex-start;
  margin-bottom: 8px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 24px;
  border-top: 1px solid var(--el-border-color-lighter);
  margin-top: 24px;
}
</style>
