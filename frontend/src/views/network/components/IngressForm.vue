<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Delete, Plus } from '@element-plus/icons-vue'
import yaml from 'js-yaml'
import type { FormInstance, FormRules } from 'element-plus'
import {
  getNamespaceList,
  createIngress,
  updateIngress,
  extractNamespaceNames,
  getIngressClassList,
} from '@/api/resource'

const props = withDefaults(
  defineProps<{
    isEdit?: boolean
    initialData?: any
  }>(),
  {
    isEdit: false,
    initialData: undefined,
  },
)

const emit = defineEmits<{
  success: []
  cancel: []
}>()

const { t } = useI18n()
const router = useRouter()
const submitting = ref(false)
const namespaceLoading = ref(false)
const namespaces = ref<string[]>([])
const ingressClasses = ref<string[]>([])
const ingressClassLoading = ref(false)

// ---- Form Data ----

interface Label {
  key: string
  value: string
}

interface IngressPath {
  path: string
  pathType: string
  backendService: string
  backendPort: number | null
}

interface IngressHostRule {
  host: string
  paths: IngressPath[]
}

interface TlsConfig {
  hosts: string
  secretName: string
}

interface Annotation {
  key: string
  value: string
}

interface FormData {
  name: string
  namespace: string
  labels: Label[]
  annotations: Annotation[]
  ingressClassName: string
  defaultBackendEnabled: boolean
  defaultBackendService: string
  defaultBackendPort: number | null
  rules: IngressHostRule[]
  tlsEnabled: boolean
  tls: TlsConfig[]
}

const form = reactive<FormData>({
  name: '',
  namespace: 'default',
  labels: [{ key: 'app', value: '' }],
  annotations: [],
  ingressClassName: '',
  defaultBackendEnabled: false,
  defaultBackendService: '',
  defaultBackendPort: null,
  rules: [
    { host: '', paths: [{ path: '/', pathType: 'Prefix', backendService: '', backendPort: 80 }] },
  ],
  tlsEnabled: false,
  tls: [{ hosts: '', secretName: '' }],
})

// ---- Validation ----

const formRef = ref<FormInstance>()

const formRules: FormRules = {
  name: [
    { required: true, message: '请输入名称', trigger: 'blur' },
    {
      pattern: /^[a-z][a-z0-9-]*[a-z0-9]$/,
      message: '仅支持小写字母、数字和连字符，以字母开头',
      trigger: 'blur',
    },
    { max: 253, message: '最长 253 个字符', trigger: 'blur' },
  ],
  namespace: [{ required: true, message: '请选择命名空间', trigger: 'change' }],
  ingressClassName: [{ required: true, message: '请选择 IngressClass', trigger: 'change' }],
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

async function fetchIngressClasses() {
  ingressClassLoading.value = true
  try {
    const res: any = await getIngressClassList({})
    ingressClasses.value = res.data || []
  } catch {
    ingressClasses.value = []
  } finally {
    ingressClassLoading.value = false
  }
}

onMounted(() => {
  fetchNamespaces()
  fetchIngressClasses()
  if (props.isEdit && props.initialData) {
    parseInitialData(props.initialData)
  }
})

// 克隆流入（创建模式 isEdit=false，onMounted 不会触发 parseInitialData，故用 watch 兜底）
watch(
  () => props.initialData,
  (newData) => {
    if (newData) parseInitialData(newData)
  },
)

// ---- Label Management ----

function addLabel() {
  form.labels.push({ key: '', value: '' })
}
function removeLabel(i: number) {
  form.labels.splice(i, 1)
}

// ---- Rule Management ----

function addRule() {
  form.rules.push({
    host: '',
    paths: [{ path: '/', pathType: 'Prefix', backendService: '', backendPort: 80 }],
  })
}
function removeRule(i: number) {
  if (form.rules.length <= 1) {
    ElMessage.warning(t('network.atLeastOneRule'))
    return
  }
  form.rules.splice(i, 1)
}
function addPath(ruleIdx: number) {
  form.rules[ruleIdx].paths.push({
    path: '/',
    pathType: 'Prefix',
    backendService: '',
    backendPort: 80,
  })
}
function removePath(ruleIdx: number, pathIdx: number) {
  if (form.rules[ruleIdx].paths.length <= 1) {
    ElMessage.warning(t('network.hostPathRequired'))
    return
  }
  form.rules[ruleIdx].paths.splice(pathIdx, 1)
}

// ---- TLS Management ----

function addTls() {
  form.tls.push({ hosts: '', secretName: '' })
}
function removeTls(i: number) {
  if (form.tls.length <= 1) {
    ElMessage.warning(t('network.atLeastOneTls'))
    return
  }
  form.tls.splice(i, 1)
}

// ---- YAML Generation ----

const generatedYaml = computed(() => {
  const resource = buildK8sIngress()
  return yaml.dump(resource, { indent: 2, lineWidth: -1, noRefs: true })
})

function buildK8sIngress(): Record<string, any> {
  const labels: Record<string, string> = {}
  form.labels.forEach((l) => {
    if (l.key.trim()) labels[l.key.trim()] = l.value
  })

  const rules = form.rules
    .filter((r) => r.host.trim())
    .map((r) => ({
      host: r.host.trim(),
      http: {
        paths: r.paths
          .filter((p) => p.backendService.trim())
          .map((p) => ({
            path: p.path,
            pathType: p.pathType,
            backend: {
              service: {
                name: p.backendService,
                port: { number: p.backendPort },
              },
            },
          })),
      },
    }))

  const annotations: Record<string, string> = {}
  form.annotations.forEach((a) => {
    if (a.key.trim()) annotations[a.key.trim()] = a.value
  })

  const metadata: Record<string, any> = {
    name: form.name,
    namespace: form.namespace,
    labels: { ...labels },
  }
  if (Object.keys(annotations).length > 0) metadata.annotations = annotations

  const spec: Record<string, any> = {
    ingressClassName: form.ingressClassName,
    rules,
  }

  // Default backend
  if (form.defaultBackendEnabled && form.defaultBackendService && form.defaultBackendPort) {
    spec.defaultBackend = {
      service: {
        name: form.defaultBackendService,
        port: { number: form.defaultBackendPort },
      },
    }
  }

  const resource: Record<string, any> = {
    apiVersion: 'networking.k8s.io/v1',
    kind: 'Ingress',
    metadata,
    spec,
  }

  if (form.tlsEnabled) {
    const tls = form.tls
      .filter((t) => t.hosts.trim())
      .map((t) => ({
        hosts: t.hosts
          .split(',')
          .map((h) => h.trim())
          .filter(Boolean),
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
  form.ingressClassName = spec.ingressClassName || ''

  // Labels
  const labels = meta.labels || {}
  form.labels =
    Object.keys(labels).length > 0
      ? Object.entries(labels).map(([k, v]) => ({ key: k, value: v as string }))
      : [{ key: 'app', value: '' }]

  // Annotations
  const annotations = meta.annotations || {}
  form.annotations =
    Object.keys(annotations).length > 0
      ? Object.entries(annotations).map(([k, v]) => ({ key: k, value: v as string }))
      : []

  // Default backend
  const defaultBackend = spec.defaultBackend
  if (defaultBackend) {
    form.defaultBackendEnabled = true
    form.defaultBackendService = defaultBackend.service?.name || ''
    form.defaultBackendPort = defaultBackend.service?.port?.number ?? null
  }

  // Rules — group by host to preserve multi-path structure
  const specRules = spec.rules || []
  if (specRules.length > 0) {
    const hostMap = new Map<string, IngressPath[]>()
    for (const rule of specRules) {
      const host = rule.host || ''
      for (const p of rule.http?.paths || []) {
        const path: IngressPath = {
          path: p.path || '/',
          pathType: p.pathType || 'Prefix',
          backendService: p.backend?.service?.name || '',
          backendPort: p.backend?.service?.port?.number ?? null,
        }
        const existing = hostMap.get(host)
        if (existing) existing.push(path)
        else hostMap.set(host, [path])
      }
    }
    form.rules = Array.from(hostMap.entries()).map(([host, paths]) => ({ host, paths }))
  } else {
    form.rules = [
      { host: '', paths: [{ path: '/', pathType: 'Prefix', backendService: '', backendPort: 80 }] },
    ]
  }

  // TLS
  const tls = spec.tls || []
  form.tlsEnabled = tls.length > 0
  form.tls =
    tls.length > 0
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
    if (!r.host.trim()) {
      ElMessage.error(t('network.hostRequired', { n: i + 1 }))
      return
    }
    for (let j = 0; j < r.paths.length; j++) {
      const p = r.paths[j]
      if (!p.backendService.trim()) {
        ElMessage.error(t('network.backendServiceRequired', { n: i + 1, m: j + 1 }))
        return
      }
      if (!p.backendPort) {
        ElMessage.error(t('network.backendPortRequired', { n: i + 1, m: j + 1 }))
        return
      }
    }
  }

  submitting.value = true
  try {
    if (props.isEdit) {
      await updateIngress({ namespace: form.namespace, name: form.name, yaml: generatedYaml.value })
      ElMessage.success(t('network.ingressUpdated'))
      emit('success')
    } else {
      await createIngress({ namespace: form.namespace, yaml: generatedYaml.value })
      ElMessage.success(t('network.ingressCreated'))
      router.push('/network/ingresses')
    }
  } catch (e: any) {
    ElMessage.error(
      e?.message || (props.isEdit ? t('common.updateFailed') : t('common.createFailed')),
    )
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
              <el-select
                v-model="form.namespace"
                filterable
                placeholder="选择命名空间"
                style="width: 100%"
                :loading="namespaceLoading"
              >
                <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
              </el-select>
            </el-form-item>
            <el-form-item label="Ingress Class Name" prop="ingressClassName">
              <el-select
                v-model="form.ingressClassName"
                filterable
                allow-create
                placeholder="选择或输入 IngressClass"
                style="width: 100%"
                :loading="ingressClassLoading"
              >
                <el-option v-for="ic in ingressClasses" :key="ic" :label="ic" :value="ic" />
              </el-select>
              <div v-if="ingressClasses.length === 0 && !ingressClassLoading" class="form-tip">
                未检测到 IngressClass，可手动输入名称
              </div>
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
            <div style="width: 100%">
              <div v-for="(label, i) in form.labels" :key="i" class="kv-row">
                <el-input v-model="label.key" placeholder="Key" />
                <el-input v-model="label.value" placeholder="Value" />
                <el-button
                  type="danger"
                  text
                  circle
                  :disabled="form.labels.length <= 1"
                  @click="removeLabel(i)"
                >
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
              <el-button text type="primary" size="small" @click="addLabel">
                <el-icon><Plus /></el-icon> 添加标签
              </el-button>
            </div>
          </el-form-item>
        </div>
      </div>

      <!-- Section: Annotations -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">注解</div>
        </div>
        <div class="section-content">
          <el-form-item label="注解">
            <div style="width: 100%">
              <div v-for="(ann, i) in form.annotations" :key="i" class="kv-row">
                <el-input
                  v-model="ann.key"
                  placeholder="Key (如 nginx.ingress.kubernetes.io/rewrite-target)"
                />
                <el-input v-model="ann.value" placeholder="Value" />
                <el-button type="danger" text circle @click="form.annotations.splice(i, 1)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
              <el-button
                text
                type="primary"
                size="small"
                @click="form.annotations.push({ key: '', value: '' })"
              >
                <el-icon><Plus /></el-icon> 添加注解
              </el-button>
              <div class="form-tip">
                常用注解: nginx.ingress.kubernetes.io/rewrite-target,
                nginx.ingress.kubernetes.io/ssl-redirect
              </div>
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
            <div style="width: 100%">
              <div v-for="(rule, ri) in form.rules" :key="ri" class="rule-card">
                <div class="rule-row-top">
                  <el-input
                    v-model="rule.host"
                    placeholder="Host (如 example.com)"
                    style="flex: 2"
                  />
                  <el-button type="danger" text circle @click="removeRule(ri)">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
                <div v-for="(path, pi) in rule.paths" :key="pi" class="path-row">
                  <el-input v-model="path.path" placeholder="Path" style="flex: 1" />
                  <el-select v-model="path.pathType" style="width: 160px">
                    <el-option label="Prefix" value="Prefix" />
                    <el-option label="Exact" value="Exact" />
                    <el-option label="ImplementationSpecific" value="ImplementationSpecific" />
                  </el-select>
                  <span class="backend-label">→</span>
                  <el-input
                    v-model="path.backendService"
                    placeholder="Service 名称"
                    style="flex: 1"
                  />
                  <el-input-number
                    v-model="path.backendPort"
                    :min="1"
                    :max="65535"
                    placeholder="端口"
                    style="width: 140px"
                  />
                  <el-button type="danger" text circle size="small" @click="removePath(ri, pi)">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
                <el-button
                  text
                  type="primary"
                  size="small"
                  style="margin-top: 4px"
                  @click="addPath(ri)"
                >
                  <el-icon><Plus /></el-icon> 添加路径
                </el-button>
              </div>
              <el-button text type="primary" size="small" @click="addRule">
                <el-icon><Plus /></el-icon> 添加规则
              </el-button>
            </div>
          </el-form-item>
        </div>
      </div>

      <!-- Section: Default Backend -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">默认后端</div>
        </div>
        <div class="section-content">
          <el-form-item label="启用默认后端">
            <el-switch v-model="form.defaultBackendEnabled" />
            <div class="form-tip">无匹配规则时的兜底服务</div>
          </el-form-item>
          <template v-if="form.defaultBackendEnabled">
            <div class="fields-grid">
              <el-form-item label="Service 名称">
                <el-input v-model="form.defaultBackendService" placeholder="默认后端 Service" />
              </el-form-item>
              <el-form-item label="端口">
                <el-input-number
                  v-model="form.defaultBackendPort"
                  :min="1"
                  :max="65535"
                  style="width: 100%"
                />
              </el-form-item>
            </div>
          </template>
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
              <div style="width: 100%">
                <div v-for="(t, ti) in form.tls" :key="ti" class="tls-card">
                  <div class="tls-row">
                    <el-input v-model="t.hosts" placeholder="Hosts (逗号分隔)" style="flex: 1" />
                    <el-input
                      v-model="t.secretName"
                      placeholder="Secret 名称"
                      style="width: 200px"
                    />
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
            <el-button type="primary" :loading="submitting" @click="handleSubmit">{{
              isEdit ? '更新' : '创建'
            }}</el-button>
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
  border-radius: var(--gk-radius-md);
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

.path-row {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 8px;
  padding-left: 16px;
}

.backend-label {
  font-size: 13px;
  color: var(--el-text-color-regular);
  white-space: nowrap;
}

.form-tip {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
}

/* TLS cards */
.tls-card {
  border: 1px solid var(--el-border-color-extra-light);
  border-radius: var(--gk-radius-md);
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
