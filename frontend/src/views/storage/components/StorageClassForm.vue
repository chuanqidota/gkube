<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Delete, Plus } from '@element-plus/icons-vue'
import yaml from 'js-yaml'
import type { FormInstance, FormRules } from 'element-plus'
import { createStorageClass, updateStorageClass } from '@/api/resource'

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

interface KVPair {
  key: string
  value: string
}

interface FormData {
  name: string
  provisioner: string
  reclaimPolicy: string
  volumeBindingMode: string
  allowVolumeExpansion: boolean
  mountOptions: string[]
  parameters: KVPair[]
  labels: KVPair[]
}

const form = reactive<FormData>({
  name: '',
  provisioner: '',
  reclaimPolicy: 'Delete',
  volumeBindingMode: 'Immediate',
  allowVolumeExpansion: false,
  mountOptions: [],
  parameters: [],
  labels: [],
})

// Parse initial data for edit mode
function parseInitialData(data: any) {
  if (!data) return
  form.name = data.metadata?.name || data.name || ''
  form.provisioner = data.provisioner || ''
  form.reclaimPolicy = data.reclaimPolicy || data.reclaim_policy || 'Delete'
  form.volumeBindingMode = data.volumeBindingMode || data.volume_binding_mode || 'Immediate'
  form.allowVolumeExpansion = data.allowVolumeExpansion || data.allow_volume_expansion || false

  // Mount Options
  const mountOpts = data.mountOptions || data.mount_options || []
  form.mountOptions = mountOpts.length > 0 ? [...mountOpts] : []

  // Parameters
  const params = data.parameters || {}
  form.parameters = Object.entries(params).map(([k, v]) => ({ key: k, value: String(v) }))
  if (form.parameters.length === 0) form.parameters.push({ key: '', value: '' })

  // Labels
  const labels = data.metadata?.labels || data.labels || {}
  form.labels = Object.entries(labels).map(([k, v]) => ({ key: k, value: String(v) }))
  if (form.labels.length === 0) form.labels.push({ key: '', value: '' })
}

// Initialize
if (props.isEdit && props.initialData) {
  parseInitialData(props.initialData)
} else {
  form.parameters = [{ key: '', value: '' }]
  form.labels = [{ key: '', value: '' }]
}

// 克隆流入（创建模式 isEdit=false，上面不会触发 parseInitialData，故用 watch 兜底）
watch(() => props.initialData, (newData) => {
  if (newData) parseInitialData(newData)
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  provisioner: [{ required: true, message: '请输入 Provisioner', trigger: 'blur' }],
}

function addParam() { form.parameters.push({ key: '', value: '' }) }
function removeParam(i: number) { form.parameters.splice(i, 1) }
function addLabel() { form.labels.push({ key: '', value: '' }) }
function removeLabel(i: number) { form.labels.splice(i, 1) }

function buildYamlStr(): string {
  const parameters: Record<string, string> = {}
  form.parameters.forEach((p) => { if (p.key.trim()) parameters[p.key.trim()] = p.value })
  const labels: Record<string, string> = {}
  form.labels.forEach((l) => { if (l.key.trim()) labels[l.key.trim()] = l.value })

  const obj: any = {
    apiVersion: 'storage.k8s.io/v1',
    kind: 'StorageClass',
    metadata: {
      name: form.name,
    },
    provisioner: form.provisioner,
    reclaimPolicy: form.reclaimPolicy,
    volumeBindingMode: form.volumeBindingMode,
  }
  if (Object.keys(parameters).length > 0) obj.parameters = parameters
  if (Object.keys(labels).length > 0) obj.metadata.labels = labels
  if (form.allowVolumeExpansion) obj.allowVolumeExpansion = true
  if (form.mountOptions.length > 0) obj.mountOptions = form.mountOptions
  return yaml.dump(obj, { indent: 2, lineWidth: -1, noRefs: true })
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    const yamlStr = buildYamlStr()
    if (props.isEdit) {
      await updateStorageClass({ name: form.name, yaml: yamlStr })
      ElMessage.success('StorageClass 更新成功')
      emit('success')
    } else {
      await createStorageClass({ yaml: yamlStr })
      ElMessage.success('StorageClass 创建成功')
      router.push('/storage/storageclasses')
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
    router.push('/storage/storageclasses')
  }
}
</script>

<template>
  <div class="sc-form">
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
      <!-- Section 1: Basic Info -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">基本信息</div>
        </div>
        <div class="section-content">
          <div class="fields-grid">
            <el-form-item label="名称" prop="name">
              <el-input v-model="form.name" :disabled="isEdit" placeholder="例如: fast-ssd" />
            </el-form-item>
            <el-form-item label="Provisioner" prop="provisioner">
              <el-input v-model="form.provisioner" :disabled="isEdit" placeholder="例如: kubernetes.io/aws-ebs" />
            </el-form-item>
          </div>
          <el-form-item label="回收策略">
            <el-radio-group v-model="form.reclaimPolicy">
              <el-radio value="Delete">Delete</el-radio>
              <el-radio value="Retain">Retain</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="卷绑定模式">
            <el-radio-group v-model="form.volumeBindingMode">
              <el-radio value="Immediate">Immediate</el-radio>
              <el-radio value="WaitForFirstConsumer">WaitForFirstConsumer</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="允许卷扩展">
            <el-switch v-model="form.allowVolumeExpansion" />
          </el-form-item>
          <el-form-item label="挂载选项">
            <div style="width: 100%;">
              <div v-for="(_opt, i) in form.mountOptions" :key="i" class="kv-row">
                <el-input v-model="form.mountOptions[i]" placeholder="例如: nfsvers=4.1" />
                <el-button type="danger" text circle @click="form.mountOptions.splice(i, 1)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
              <el-button text type="primary" @click="form.mountOptions.push('')" size="small">
                <el-icon><Plus /></el-icon> 添加挂载选项
              </el-button>
            </div>
          </el-form-item>
        </div>
      </div>

      <!-- Section 2: Parameters -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">参数</div>
        </div>
        <div class="section-content">
          <el-form-item label="参数">
            <div style="width: 100%;">
              <div v-for="(p, i) in form.parameters" :key="i" class="kv-row">
                <el-input v-model="p.key" placeholder="键" />
                <el-input v-model="p.value" placeholder="值" />
                <el-button type="danger" text circle :disabled="form.parameters.length <= 1" @click="removeParam(i)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
              <el-button text type="primary" @click="addParam" size="small">
                <el-icon><Plus /></el-icon> 添加参数
              </el-button>
            </div>
          </el-form-item>
        </div>
      </div>

      <!-- Section 3: Labels -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">标签</div>
        </div>
        <div class="section-content">
          <el-form-item label="标签">
            <div style="width: 100%;">
              <div v-for="(l, i) in form.labels" :key="i" class="kv-row">
                <el-input v-model="l.key" placeholder="键" />
                <el-input v-model="l.value" placeholder="值" />
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
.sc-form {
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
