<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Delete, Plus } from '@element-plus/icons-vue'
import yaml from 'js-yaml'
import type { FormInstance, FormRules } from 'element-plus'
import { getNamespaceList, createIngress, updateIngress, extractNamespaceNames } from '@/api/resource'

const props = withDefaults(defineProps<{
  isEdit?: boolean
  initialData?: any
}>(), {
  isEdit: false,
  initialData: undefined,
})

const emit = defineEmits<{
  success: []
  cancel: []
}>()

const router = useRouter()
const submitting = ref(false)
const namespaceLoading = ref(false)
const namespaces = ref<string[]>([])

// ---- Form Data ----

interface Label { key: string; value: string }

interface IngressRule {
  host: string
  path: string
  pathType: string
  backendService: string
  backendPort: number | null
}

interface TlsConfig {
  hosts: string
  secretName: string
}

interface FormData {
  name: string
  namespace: string
  labels: Label[]
  ingressClassName: string
  rules: IngressRule[]
  tlsEnabled: boolean
  tls: TlsConfig[]
}

const form = reactive<FormData>({
  name: '',
  namespace: 'default',
  labels: [{ key: 'app', value: '' }],
  ingressClassName: 'nginx',
  rules: [{ host: '', path: '/', pathType: 'Prefix', backendService: '', backendPort: 80 }],
  tlsEnabled: false,
  tls: [{ hosts: '', secretName: '' }],
})

// ---- Validation ----

const formRef = ref<FormInstance>()

const formRules: FormRules = {
  name: [
    { required: true, message: '请输入名称', trigger: 'blur' },
    { pattern: /^[a-z][a-z0-9-]*[a-z0-9]$/, message: '仅支持小写字母、数字和连字符，以字母开头', trigger: 'blur' },
    { max: 253, message: '最长 253 个字符', trigger: 'blur' },
  ],
  namespace: [{ required: true, message: '请选择命名空间', trigger: 'change' }],
  ingressClassName: [{ required: true, message: '请输入 Ingress Class 名称', trigger: 'blur' }],
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

onMounted(() => {
  fetchNamespaces()
  if (props.isEdit && props.initialData) {
    parseInitialData(props.initialData)
  }
})

// ---- Label Management ----

function addLabel() { form.labels.push({ key: '', value: '' }) }
function removeLabel(i: number) { form.labels.splice(i, 1) }

// ---- Rule Management ----

function addRule() {
  form.rules.push({ host: '', path: '/', pathType: 'Prefix', backendService: '', backendPort: 80 })
}
function removeRule(i: number) {
  if (form.rules.length <= 1) { ElMessage.warning('至少需要一条规则'); return }
  form.rules.splice(i, 1)
}

// ---- TLS Management ----

function addTls() { form.tls.push({ hosts: '', secretName: '' }) }
function removeTls(i: number) {
  if (form.tls.length <= 1) { ElMessage.warning('至少需要一条 TLS 配置'); return }
  form.tls.splice(i, 1)
}

// ---- YAML Generation ----

const generatedYaml = computed(() => {
  const resource = buildK8sIngress()
  return yaml.dump(resource, { indent: 2, lineWidth: -1, noRefs: true })
})

function buildK8sIngress(): Record<string, any> {
  const labels: Record<string, string> = {}
  form.labels.forEach(l => { if (l.key.trim()) labels[l.key.trim()] = l.value })

  const rules = form.rules
    .filter(r => r.host.trim())
    .map(r => ({
      host: r.host.trim(),
      http: {
        paths: [{
          path: r.path,
          pathType: r.pathType,
          backend: {
            service: {
              name: r.backendService,
              port: { number: r.backendPort },
            },
          },
        }],
      },
    }))

  const resource: Record<string, any> = {
    apiVersion: 'networking.k8s.io/v1',
    kind: 'Ingress',
    metadata: { name: form.name, namespace: form.namespace, labels: { ...labels } },
    spec: {
      ingressClassName: form.ingressClassName,
      rules,
    },
  }

  if (form.tlsEnabled) {
    const tls = form.tls
      .filter(t => t.hosts.trim())
      .map(t => ({
        hosts: t.hosts.split(',').map(h => h.trim()).filter(Boolean),
        secretName: t.secretName,
      }))
    if (tls.length > 0) resource.spec.tls = tls
  }

  return resource
}

// ---- Parse Initial Data (Edit Mode) ----

function parseInitialData(data: any) {
  const meta = data.metadata || {}
  const spec = data.spec || {}

  form.name = meta.name || ''
  form.namespace = meta.namespace || 'default'
  form.ingressClassName = spec.ingressClassName || 'nginx'

  // Labels
  const labels = meta.labels || {}
  form.labels = Object.keys(labels).length > 0
    ? Object.entries(labels).map(([k, v]) => ({ key: k, value: v as string }))
    : [{ key: 'app', value: '' }]

  // Rules
  const rules = spec.rules || []
  form.rules = rules.length > 0
    ? rules.flatMap((rule: any) =>
        (rule.http?.paths || []).map((p: any) => ({
          host: rule.host || '',
          path: p.path || '/',
          pathType: p.pathType || 'Prefix',
          backendService: p.backend?.service?.name || '',
          backendPort: p.backend?.service?.port?.number ?? null,
        }))
      )
    : [{ host: '', path: '/', pathType: 'Prefix', backendService: '', backendPort: 80 }]

  // TLS
  const tls = spec.tls || []
  form.tlsEnabled = tls.length > 0
  form.tls = tls.length > 0
    ? tls.map((t: any) => ({
        hosts: (t.hosts || []).join(', '),
        secretName: t.secretName || '',
      }))
    : [{ hosts: '', secretName: '' }]
}

// ---- Submit ----

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  for (let i = 0; i < form.rules.length; i++) {
    const r = form.rules[i]
    if (!r.host.trim()) { ElMessage.error(`规则 ${i + 1}: Host 不能为空`); return }
    if (!r.backendService.trim()) { ElMessage.error(`规则 ${i + 1}: 后端 Service 名称不能为空`); return }
    if (!r.backendPort) { ElMessage.error(`规则 ${i + 1}: 后端端口不能为空`); return }
  }

  submitting.value = true
  try {
    if (props.isEdit) {
      await updateIngress({ namespace: form.namespace, name: form.name, yaml: generatedYaml.value })
      ElMessage.success('Ingress 更新成功')
      emit('success')
    } else {
      await createIngress({ namespace: form.namespace, yaml: generatedYaml.value })
      ElMessage.success('Ingress 创建成功')
      router.push('/network/ingresses')
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
    router.push('/network/ingresses')
  }
}
</script>

<template>
  <div class="ingress-form">
    <el-form ref="formRef" :model="form" :rules="formRules" label-position="top">

      <!-- Section 1: Basic Info -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">基本信息</div>
        </div>
        <div class="section-content">
          <div class="fields-grid">
            <el-form-item label="名称" prop="name">
              <el-input v-model="form.name" placeholder="my-ingress" />
            </el-form-item>
            <el-form-item label="命名空间" prop="namespace">
              <el-select v-model="form.namespace" filterable placeholder="选择命名空间" style="width: 100%;" :loading="namespaceLoading">
                <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
              </el-select>
            </el-form-item>
            <el-form-item label="Ingress Class Name" prop="ingressClassName">
              <el-input v-model="form.ingressClassName" placeholder="nginx" />
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

      <!-- Section 3: Rules -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">路由规则</div>
        </div>
        <div class="section-content">
          <el-form-item label="规则" required>
            <div style="width: 100%;">
              <div v-for="(rule, ri) in form.rules" :key="ri" class="rule-card">
                <div class="rule-row-top">
                  <el-input v-model="rule.host" placeholder="Host (如 example.com)" style="flex: 2;" />
                  <el-input v-model="rule.path" placeholder="Path" style="flex: 1;" />
                  <el-select v-model="rule.pathType" style="width: 160px;">
                    <el-option label="Prefix" value="Prefix" />
                    <el-option label="Exact" value="Exact" />
                    <el-option label="ImplementationSpecific" value="ImplementationSpecific" />
                  </el-select>
                  <el-button type="danger" text circle @click="removeRule(ri)">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
                <div class="rule-row-bottom">
                  <span class="backend-label">后端:</span>
                  <el-input v-model="rule.backendService" placeholder="Service 名称" style="flex: 1;" />
                  <el-input-number v-model="rule.backendPort" :min="1" :max="65535" placeholder="端口" style="width: 140px;" />
                </div>
              </div>
              <el-button text type="primary" size="small" @click="addRule">
                <el-icon><Plus /></el-icon> 添加规则
              </el-button>
            </div>
          </el-form-item>
        </div>
      </div>

      <!-- Section 4: TLS -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">TLS 配置</div>
        </div>
        <div class="section-content">
          <el-form-item label="启用 TLS">
            <el-switch v-model="form.tlsEnabled" />
          </el-form-item>
          <template v-if="form.tlsEnabled">
            <el-form-item label="TLS 配置">
              <div style="width: 100%;">
                <div v-for="(t, ti) in form.tls" :key="ti" class="tls-card">
                  <div class="tls-row">
                    <el-input v-model="t.hosts" placeholder="Hosts (逗号分隔)" style="flex: 1;" />
                    <el-input v-model="t.secretName" placeholder="Secret 名称" style="width: 200px;" />
                    <el-button type="danger" text circle @click="removeTls(ti)">
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </div>
                </div>
                <el-button text type="primary" size="small" @click="addTls">
                  <el-icon><Plus /></el-icon> 添加 TLS
                </el-button>
              </div>
            </el-form-item>
          </template>
        </div>
      </div>

      <!-- Submit Button -->
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
.ingress-form {
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

/* Rule cards */
.rule-card {
  border: 1px solid var(--el-border-color-extra-light);
  border-radius: 8px;
  padding: 14px;
  margin-bottom: 8px;
  background: var(--el-fill-color-lighter);
}

.rule-row-top {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
  margin-bottom: 8px;
}

.rule-row-bottom {
  display: flex;
  gap: 8px;
  align-items: center;
}

.backend-label {
  font-size: 13px;
  color: var(--el-text-color-regular);
  white-space: nowrap;
}

/* TLS cards */
.tls-card {
  border: 1px solid var(--el-border-color-extra-light);
  border-radius: 8px;
  padding: 14px;
  margin-bottom: 8px;
  background: var(--el-fill-color-lighter);
}

.tls-row {
  display: flex;
  gap: 8px;
  align-items: center;
}
</style>
