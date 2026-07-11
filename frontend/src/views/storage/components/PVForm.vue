<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Delete, Plus } from '@element-plus/icons-vue'
import yaml from 'js-yaml'
import type { FormInstance, FormRules } from 'element-plus'
import { createPv, updatePvYaml } from '@/api/resource'

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

// ---- Parse initial data for edit mode ----

function parseInitialData(data: any) {
  if (!data) return
  const spec = data.spec || {}
  const meta = data.metadata || {}

  form.name = meta.name || ''
  form.capacity = spec.capacity?.storage || '10Gi'
  form.accessModes = spec.accessModes || ['ReadWriteOnce']
  form.storageClassName = spec.storageClassName || ''
  form.reclaimPolicy = spec.persistentVolumeReclaimPolicy || 'Retain'

  // Storage source
  if (spec.nfs) {
    form.storageType = 'nfs'
    form.nfsServer = spec.nfs.server || ''
    form.nfsPath = spec.nfs.path || ''
  } else if (spec.hostPath) {
    form.storageType = 'hostPath'
    form.hostPath = spec.hostPath.path || ''
  } else if (spec.local) {
    form.storageType = 'local'
    form.localPath = spec.local.path || ''
  } else {
    form.storageType = 'nfs'
  }

  // Labels
  const labels = meta.labels || {}
  form.labels = Object.entries(labels).map(([k, v]) => ({ key: k, value: String(v) }))
  if (form.labels.length === 0) form.labels = [{ key: '', value: '' }]
}

if (props.isEdit && props.initialData) {
  parseInitialData(props.initialData)
}

// ---- Validation ----

const rules: FormRules = {
  name: [
    { required: true, message: '请输入名称', trigger: 'blur' },
    { pattern: /^[a-z][a-z0-9-]*[a-z0-9]$/, message: '只能包含小写字母、数字和连字符，必须以字母开头，以字母或数字结尾', trigger: 'blur' },
    { max: 253, message: '最大长度为253个字符', trigger: 'blur' },
  ],
  capacity: [
    { required: true, message: '请输入容量', trigger: 'blur' },
  ],
  accessModes: [
    { type: 'array', required: true, message: '请至少选择一种访问模式', trigger: 'change' },
  ],
  storageType: [
    { required: true, message: '请选择存储类型', trigger: 'change' },
  ],
}

// ---- Label Management ----

function addLabel() {
  form.labels.push({ key: '', value: '' })
}

function removeLabel(index: number) {
  form.labels.splice(index, 1)
}

// ---- YAML Generation ----

function buildK8sPV(): Record<string, any> {
  const labels: Record<string, string> = {}
  form.labels.forEach((l) => {
    if (l.key.trim()) labels[l.key.trim()] = l.value
  })

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
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  // Validate storage source
  if (form.storageType === 'nfs') {
    if (!form.nfsServer.trim()) {
      ElMessage.error('请输入 NFS 服务器地址')
      return
    }
    if (!form.nfsPath.trim()) {
      ElMessage.error('请输入 NFS 路径')
      return
    }
  } else if (form.storageType === 'hostPath') {
    if (!form.hostPath.trim()) {
      ElMessage.error('请输入主机路径')
      return
    }
  } else if (form.storageType === 'local') {
    if (!form.localPath.trim()) {
      ElMessage.error('请输入本地路径')
      return
    }
  }

  submitting.value = true
  try {
    const resource = buildK8sPV()
    const yamlContent = yaml.dump(resource, { indent: 2, lineWidth: -1, noRefs: true })
    if (props.isEdit) {
      await updatePvYaml({ name: form.name, yaml: yamlContent })
      ElMessage.success('持久卷更新成功')
      emit('success')
    } else {
      await createPv({ yaml: yamlContent })
      ElMessage.success('持久卷创建成功')
      router.push('/storage/pvs')
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
    router.push('/storage/pvs')
  }
}
</script>

<template>
  <div class="pv-form">
    <!-- Form -->
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="140px"
      style="max-width: 700px;"
    >
      <!-- 基本信息 -->
      <div class="form-section">
        <div class="section-title">基本信息</div>
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" :disabled="isEdit" placeholder="例如: my-pv" />
        </el-form-item>

        <el-form-item label="容量" prop="capacity">
          <el-input v-model="form.capacity" placeholder="例如: 10Gi" />
        </el-form-item>

        <el-form-item label="访问模式" prop="accessModes">
          <el-checkbox-group v-model="form.accessModes">
            <el-checkbox label="ReadWriteOnce" value="ReadWriteOnce" />
            <el-checkbox label="ReadOnlyMany" value="ReadOnlyMany" />
            <el-checkbox label="ReadWriteMany" value="ReadWriteMany" />
          </el-checkbox-group>
        </el-form-item>

        <el-form-item label="存储类名称">
          <el-input v-model="form.storageClassName" placeholder="留空表示不指定存储类" />
        </el-form-item>

        <el-form-item label="回收策略">
          <el-select v-model="form.reclaimPolicy" style="width: 100%;">
            <el-option label="Retain" value="Retain" />
            <el-option label="Recycle" value="Recycle" />
            <el-option label="Delete" value="Delete" />
          </el-select>
        </el-form-item>

        <el-form-item label="标签">
          <div style="width: 100%;">
            <div
              v-for="(label, index) in form.labels"
              :key="index"
              style="display: flex; gap: 8px; margin-bottom: 8px;"
            >
              <el-input v-model="label.key" placeholder="键" style="flex: 1;" />
              <el-input v-model="label.value" placeholder="值" style="flex: 1;" />
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
              <el-icon><Plus /></el-icon> 添加标签
            </el-button>
          </div>
        </el-form-item>
      </div>

      <!-- 存储源 -->
      <div class="form-section">
        <div class="section-title">存储源</div>
        <el-form-item label="存储类型" prop="storageType" required>
          <el-select v-model="form.storageType" style="width: 100%;">
            <el-option label="NFS" value="nfs" />
            <el-option label="Host Path" value="hostPath" />
            <el-option label="Local" value="local" />
          </el-select>
        </el-form-item>

        <!-- NFS -->
        <template v-if="form.storageType === 'nfs'">
          <el-form-item label="NFS 服务器" required>
            <el-input v-model="form.nfsServer" placeholder="例如: 10.0.0.1" />
          </el-form-item>
          <el-form-item label="NFS 路径" required>
            <el-input v-model="form.nfsPath" placeholder="例如: /exports/data" />
          </el-form-item>
        </template>

        <!-- Host Path -->
        <template v-if="form.storageType === 'hostPath'">
          <el-form-item label="主机路径" required>
            <el-input v-model="form.hostPath" placeholder="例如: /mnt/data" />
          </el-form-item>
        </template>

        <!-- Local -->
        <template v-if="form.storageType === 'local'">
          <el-form-item label="本地路径" required>
            <el-input v-model="form.localPath" placeholder="例如: /mnt/disks/ssd1" />
          </el-form-item>
        </template>
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
.pv-form {
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
