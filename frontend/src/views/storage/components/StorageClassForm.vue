<script setup lang="ts">
import { ref, reactive } from 'vue'
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
  parameters: KVPair[]
  labels: KVPair[]
}

const form = reactive<FormData>({
  name: '',
  provisioner: '',
  reclaimPolicy: 'Delete',
  volumeBindingMode: 'Immediate',
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
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="140px"
      style="max-width: 700px;"
    >
      <div class="form-section">
        <div class="section-title">基本信息</div>
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" :disabled="isEdit" placeholder="例如: fast-ssd" />
        </el-form-item>

        <el-form-item label="Provisioner" prop="provisioner">
          <el-input v-model="form.provisioner" :disabled="isEdit" placeholder="例如: kubernetes.io/aws-ebs" />
        </el-form-item>

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

        <el-form-item label="参数">
          <div style="width: 100%;">
            <div v-for="(p, i) in form.parameters" :key="i" style="display: flex; gap: 8px; margin-bottom: 8px;">
              <el-input v-model="p.key" placeholder="键" style="flex: 1;" />
              <el-input v-model="p.value" placeholder="值" style="flex: 1;" />
              <el-button type="danger" circle :disabled="form.parameters.length <= 1" @click="removeParam(i)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <el-button @click="addParam" size="small">
              <el-icon><Plus /></el-icon> 添加参数
            </el-button>
          </div>
        </el-form-item>

        <el-form-item label="标签">
          <div style="width: 100%;">
            <div v-for="(l, i) in form.labels" :key="i" style="display: flex; gap: 8px; margin-bottom: 8px;">
              <el-input v-model="l.key" placeholder="键" style="flex: 1;" />
              <el-input v-model="l.value" placeholder="值" style="flex: 1;" />
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
    </el-form>

    <!-- Actions -->
    <div class="form-actions">
      <el-button @click="handleCancel">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ isEdit ? '更新' : '创建' }}</el-button>
    </div>
  </div>
</template>

<style scoped>
.sc-form {
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

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 24px;
  border-top: 1px solid var(--el-border-color-lighter);
  margin-top: 24px;
}
</style>
