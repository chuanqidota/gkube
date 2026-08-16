<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
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
  volumeMode: string
  mountOptions: string[]
  storageType: string
  nfsServer: string
  nfsPath: string
  hostPath: string
  localPath: string
  nodeAffinityRequired: string
  labels: Label[]
}

const form = reactive<FormData>({
  name: '',
  capacity: '10Gi',
  accessModes: ['ReadWriteOnce'],
  storageClassName: '',
  reclaimPolicy: 'Retain',
  volumeMode: 'Filesystem',
  mountOptions: [],
  storageType: 'nfs',
  nfsServer: '',
  nfsPath: '',
  hostPath: '',
  localPath: '',
  nodeAffinityRequired: '',
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
  form.volumeMode = spec.volumeMode || 'Filesystem'

  // Mount options
  const mountOpts = spec.mountOptions || []
  form.mountOptions = mountOpts.length > 0 ? [...mountOpts] : []

  // Node affinity (for local volumes)
  const nodeAffinity = spec.nodeAffinity?.required?.nodeSelectorTerms
  if (nodeAffinity && nodeAffinity.length > 0) {
    const expr = nodeAffinity[0]?.matchExpressions?.[0] || nodeAffinity[0]?.matchFields?.[0]
    if (expr) {
      form.nodeAffinityRequired = expr.values ? expr.values.join(',') : ''
    }
  }

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

// 克隆流入（创建模式 isEdit=false，上面不会触发 parseInitialData，故用 watch 兜底）
watch(() => props.initialData, (newData) => {
  if (newData) parseInitialData(newData)
})

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

  if (form.volumeMode && form.volumeMode !== 'Filesystem') {
    spec.volumeMode = form.volumeMode
  }

  if (form.mountOptions.length > 0) {
    spec.mountOptions = form.mountOptions
  }

  // Node affinity for local volumes
  if (form.storageType === 'local' && form.nodeAffinityRequired) {
    const values = form.nodeAffinityRequired.split(',').map(s => s.trim()).filter(Boolean)
    if (values.length > 0) {
      spec.nodeAffinity = {
        required: {
          nodeSelectorTerms: [{
            matchExpressions: [{
              key: 'kubernetes.io/hostname',
              operator: 'In',
              values,
            }],
          }],
        },
      }
    }
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
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
      <!-- Section 1: Basic Info -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">基本信息</div>
        </div>
        <div class="section-content">
          <div class="fields-grid">
            <el-form-item label="名称" prop="name">
              <el-input v-model="form.name" :disabled="isEdit" placeholder="例如: my-pv" />
            </el-form-item>
            <el-form-item label="容量" prop="capacity">
              <el-input v-model="form.capacity" placeholder="例如: 10Gi" />
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
            <el-form-item label="卷模式">
              <el-select v-model="form.volumeMode" style="width: 100%;">
                <el-option label="Filesystem (文件系统)" value="Filesystem" />
                <el-option label="Block (块设备)" value="Block" />
              </el-select>
            </el-form-item>
          </div>
          <el-form-item label="访问模式" prop="accessModes">
            <el-checkbox-group v-model="form.accessModes">
              <el-checkbox label="ReadWriteOnce" value="ReadWriteOnce" />
              <el-checkbox label="ReadOnlyMany" value="ReadOnlyMany" />
              <el-checkbox label="ReadWriteMany" value="ReadWriteMany" />
            </el-checkbox-group>
          </el-form-item>
          <el-form-item label="挂载选项">
            <div style="width: 100%;">
              <div v-for="(_opt, i) in form.mountOptions" :key="i" class="kv-row">
                <el-input v-model="form.mountOptions[i]" placeholder="例如: hard,nfsvers=4.1" />
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

      <!-- Section 2: Labels -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">标签</div>
        </div>
        <div class="section-content">
          <el-form-item label="标签">
            <div style="width: 100%;">
              <div v-for="(label, index) in form.labels" :key="index" class="kv-row">
                <el-input v-model="label.key" placeholder="键" />
                <el-input v-model="label.value" placeholder="值" />
                <el-button type="danger" text circle :disabled="form.labels.length <= 1" @click="removeLabel(index)">
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

      <!-- Section 3: Storage Source -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">存储源</div>
        </div>
        <div class="section-content">
          <el-form-item label="存储类型" prop="storageType" required>
            <el-select v-model="form.storageType" style="width: 100%;">
              <el-option label="NFS" value="nfs" />
              <el-option label="Host Path" value="hostPath" />
              <el-option label="Local" value="local" />
            </el-select>
          </el-form-item>

          <!-- NFS -->
          <div v-if="form.storageType === 'nfs'" class="fields-grid">
            <el-form-item label="NFS 服务器" required>
              <el-input v-model="form.nfsServer" placeholder="例如: 10.0.0.1" />
            </el-form-item>
            <el-form-item label="NFS 路径" required>
              <el-input v-model="form.nfsPath" placeholder="例如: /exports/data" />
            </el-form-item>
          </div>

          <!-- Host Path -->
          <el-form-item v-if="form.storageType === 'hostPath'" label="主机路径" required>
            <el-input v-model="form.hostPath" placeholder="例如: /mnt/data" />
          </el-form-item>

          <!-- Local -->
          <template v-if="form.storageType === 'local'">
            <el-form-item label="本地路径" required>
              <el-input v-model="form.localPath" placeholder="例如: /mnt/disks/ssd1" />
            </el-form-item>
            <el-form-item label="节点亲和性">
              <el-input v-model="form.nodeAffinityRequired" placeholder="节点名称，多个用逗号分隔" />
              <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-top: 4px;">Local PV 必须指定节点亲和性，用逗号分隔多个节点名</div>
            </el-form-item>
          </template>
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
.pv-form {
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
