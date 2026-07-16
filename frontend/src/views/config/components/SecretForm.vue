<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Delete, Plus, Upload } from '@element-plus/icons-vue'
import yaml from 'js-yaml'
import type { FormInstance, FormRules } from 'element-plus'
import { createSecret, updateSecret, getNamespaceList, extractNamespaceNames } from '@/api/resource'

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
  type: string
  labels: Array<{ key: string; value: string }>
  data: DataEntry[]
}

const form = reactive<FormData>({
  name: '',
  namespace: 'default',
  type: 'Opaque',
  labels: [],
  data: [],
})

// ---- Secret Types ----

const secretTypes = [
  { label: 'Opaque', value: 'Opaque' },
  { label: 'TLS', value: 'kubernetes.io/tls' },
  { label: 'Docker Config JSON', value: 'kubernetes.io/dockerconfigjson' },
]

// TLS specific data
const tlsData = ref({ cert: '', key: '' })

// Docker Config JSON specific data
const dockerConfig = ref({ server: '', username: '', password: '', email: '' })

// File upload
function handleFileUpload(entry: DataEntry, event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  const reader = new FileReader()
  reader.onload = (e) => {
    entry.value = e.target?.result as string
  }
  reader.readAsText(file)
  input.value = ''
}

function handleTlsFileUpload(field: 'cert' | 'key', event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  const reader = new FileReader()
  reader.onload = (e) => {
    tlsData.value[field] = e.target?.result as string
  }
  reader.readAsText(file)
  input.value = ''
}

function buildDockerConfigJson(): string {
  const config = {
    auths: {
      [dockerConfig.value.server]: {
        username: dockerConfig.value.username,
        password: dockerConfig.value.password,
        email: dockerConfig.value.email || undefined,
        auth: btoa(`${dockerConfig.value.username}:${dockerConfig.value.password}`),
      },
    },
  }
  return JSON.stringify(config)
}

// ---- Parse initial data for edit mode ----

function base64Decode(str: string): string {
  try {
    return atob(str)
  } catch {
    return str
  }
}

function base64Encode(str: string): string {
  try {
    return btoa(str)
  } catch {
    return btoa(unescape(encodeURIComponent(str)))
  }
}

function parseInitialData(data: any) {
  if (!data) return
  const meta = data.metadata || {}

  form.name = meta.name || data.name || ''
  form.namespace = meta.namespace || data.namespace || 'default'
  form.type = data.type || 'Opaque'

  // Labels
  const labels = meta.labels || data.labels || {}
  form.labels = Object.entries(labels).map(([k, v]) => ({ key: k, value: String(v) }))
  if (form.labels.length === 0) form.labels.push({ key: '', value: '' })

  // Data entries (decode base64)
  const entries = data.data || {}

  if (form.type === 'kubernetes.io/tls') {
    tlsData.value.cert = entries['tls.crt'] ? base64Decode(entries['tls.crt']) : ''
    tlsData.value.key = entries['tls.key'] ? base64Decode(entries['tls.key']) : ''
  } else if (form.type === 'kubernetes.io/dockerconfigjson') {
    try {
      const configJson = JSON.parse(base64Decode(entries['.dockerconfigjson'] || ''))
      const auths = configJson.auths || {}
      const server = Object.keys(auths)[0] || ''
      const auth = auths[server] || {}
      dockerConfig.value = {
        server,
        username: auth.username || '',
        password: auth.password || '',
        email: auth.email || '',
      }
    } catch { /* ignore parse errors */ }
  } else {
    form.data = Object.entries(entries).map(([k, v]) => ({ key: k, value: base64Decode(String(v ?? '')) }))
    if (form.data.length === 0) form.data.push({ key: '', value: '' })
  }
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
  type: [{ required: true, message: '请选择Secret类型', trigger: 'change' }],
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
  const labels: Record<string, string> = {}
  form.labels.forEach((l) => {
    if (l.key.trim()) labels[l.key.trim()] = l.value
  })

  let data: Record<string, string> = {}

  if (form.type === 'kubernetes.io/tls') {
    // TLS: use dedicated cert/key fields
    if (tlsData.value.cert) data['tls.crt'] = base64Encode(tlsData.value.cert)
    if (tlsData.value.key) data['tls.key'] = base64Encode(tlsData.value.key)
  } else if (form.type === 'kubernetes.io/dockerconfigjson') {
    // Docker Config JSON: build from form fields
    if (dockerConfig.value.server && dockerConfig.value.username && dockerConfig.value.password) {
      data['.dockerconfigjson'] = base64Encode(buildDockerConfigJson())
    }
  } else {
    // Opaque: use generic data entries
    form.data.forEach((entry) => {
      if (entry.key.trim()) data[entry.key.trim()] = base64Encode(entry.value)
    })
  }

  const obj = {
    apiVersion: 'v1',
    kind: 'Secret',
    metadata: {
      name: form.name,
      namespace: form.namespace,
      ...(Object.keys(labels).length > 0 ? { labels } : {}),
    },
    type: form.type,
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
      await updateSecret({ namespace: form.namespace, name: form.name, yaml: yamlStr })
      ElMessage.success('Secret 更新成功')
      emit('success')
    } else {
      await createSecret({ namespace: form.namespace, yaml: yamlStr })
      ElMessage.success('Secret 创建成功')
      router.push('/config/secrets')
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
    router.push('/config/secrets')
  }
}
</script>

<template>
  <div class="secret-form">
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
          <el-input v-model="form.name" :disabled="isEdit" placeholder="例如: my-secret" />
        </el-form-item>
        <el-form-item label="命名空间" prop="namespace">
          <el-select v-model="form.namespace" :disabled="isEdit" filterable placeholder="选择命名空间" style="width: 100%;" :loading="namespaceLoading">
            <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
          </el-select>
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-select v-model="form.type" :disabled="isEdit" style="width: 100%;">
            <el-option v-for="t in secretTypes" :key="t.value" :label="t.label" :value="t.value" />
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

        <!-- TLS 专用表单 -->
        <template v-if="form.type === 'kubernetes.io/tls'">
          <el-alert title="TLS Secret 需要证书 (PEM) 和私钥 (PEM) 两个字段。" type="info" :closable="false" show-icon style="margin-bottom: 16px;" />
          <el-form-item label="证书 (tls.crt)" required>
            <div style="width: 100%;">
              <el-input v-model="tlsData.cert" type="textarea" :rows="6" placeholder="-----BEGIN CERTIFICATE-----&#10;...&#10;-----END CERTIFICATE-----" />
              <div style="margin-top: 8px;">
                <label class="file-upload-btn">
                  <input type="file" accept=".pem,.crt,.cer,.txt" style="display: none;" @change="handleTlsFileUpload('cert', $event)" />
                  <el-button size="small" type="primary" plain>上传证书文件</el-button>
                </label>
              </div>
            </div>
          </el-form-item>
          <el-form-item label="私钥 (tls.key)" required>
            <div style="width: 100%;">
              <el-input v-model="tlsData.key" type="textarea" :rows="6" placeholder="-----BEGIN PRIVATE KEY-----&#10;...&#10;-----END PRIVATE KEY-----" />
              <div style="margin-top: 8px;">
                <label class="file-upload-btn">
                  <input type="file" accept=".pem,.key,.txt" style="display: none;" @change="handleTlsFileUpload('key', $event)" />
                  <el-button size="small" type="primary" plain>上传私钥文件</el-button>
                </label>
              </div>
            </div>
          </el-form-item>
        </template>

        <!-- Docker Config JSON 专用表单 -->
        <template v-else-if="form.type === 'kubernetes.io/dockerconfigjson'">
          <el-alert title="Docker Registry 认证信息，将自动编码为 .dockerconfigjson 格式。" type="info" :closable="false" show-icon style="margin-bottom: 16px;" />
          <el-form-item label="Registry 地址" required>
            <el-input v-model="dockerConfig.server" placeholder="https://index.docker.io/v1/" />
          </el-form-item>
          <el-form-item label="用户名" required>
            <el-input v-model="dockerConfig.username" placeholder="用户名" />
          </el-form-item>
          <el-form-item label="密码" required>
            <el-input v-model="dockerConfig.password" type="password" show-password placeholder="密码" />
          </el-form-item>
          <el-form-item label="邮箱">
            <el-input v-model="dockerConfig.email" placeholder="user@example.com" />
          </el-form-item>
        </template>

        <!-- 通用数据表单 (Opaque) -->
        <template v-else>
          <el-alert title="值将自动进行 Base64 编码后写入 YAML。" type="info" :closable="false" show-icon style="margin-bottom: 16px;" />
          <el-form-item label="数据项">
            <div style="width: 100%;">
              <div v-for="(entry, i) in form.data" :key="i" class="data-entry-row">
                <el-input v-model="entry.key" placeholder="Key" style="width: 200px;" />
                <el-input v-model="entry.value" type="textarea" :rows="2" placeholder="Value" style="flex: 1;" />
                <label class="file-upload-btn">
                  <input type="file" style="display: none;" @change="handleFileUpload(entry, $event)" />
                  <el-button type="primary" text circle>
                    <el-icon><Upload /></el-icon>
                  </el-button>
                </label>
                <el-button type="danger" circle :disabled="form.data.length <= 1" @click="removeEntry(i)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
              <el-button @click="addEntry" size="small">
                <el-icon><Plus /></el-icon> 添加数据项
              </el-button>
            </div>
          </el-form-item>
        </template>
      </div>
    </el-form>

    <div class="form-actions">
      <el-button @click="handleCancel">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ isEdit ? '更新' : '创建' }}</el-button>
    </div>
  </div>
</template>

<style scoped>
.secret-form {
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

.file-upload-btn {
  cursor: pointer;
  display: inline-flex;
  align-items: center;
}
</style>
