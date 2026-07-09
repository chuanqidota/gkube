<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import yaml from 'js-yaml'
import type { FormInstance, FormRules } from 'element-plus'
import { getNamespaceList, createService, extractNamespaceNames } from '@/api/resource'

const router = useRouter()
const submitting = ref(false)
const namespaceLoading = ref(false)
const namespaces = ref<string[]>([])

// ---- Form Data ----

interface Label { key: string; value: string }
interface ServicePort { name: string; port: number | null; targetPort: number | null; protocol: string; nodePort: number | null }
interface Selector { key: string; value: string }

interface FormData {
  name: string
  namespace: string
  type: string
  labels: Label[]
  ports: ServicePort[]
  selectors: Selector[]
  sessionAffinity: string
  externalTrafficPolicy: string
}

const form = reactive<FormData>({
  name: '',
  namespace: 'default',
  type: 'ClusterIP',
  labels: [{ key: 'app', value: '' }],
  ports: [{ name: 'http', port: 80, targetPort: 80, protocol: 'TCP', nodePort: null }],
  selectors: [{ key: 'app', value: '' }],
  sessionAffinity: 'None',
  externalTrafficPolicy: 'Cluster',
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
  type: [{ required: true, message: '请选择 Service 类型', trigger: 'change' }],
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

function addLabel() { form.labels.push({ key: '', value: '' }) }
function removeLabel(i: number) { form.labels.splice(i, 1) }

// ---- Selector Management ----

function addSelector() { form.selectors.push({ key: '', value: '' }) }
function removeSelector(i: number) { form.selectors.splice(i, 1) }

// ---- Port Management ----

function addPort() { form.ports.push({ name: '', port: null, targetPort: null, protocol: 'TCP', nodePort: null }) }
function removePort(i: number) {
  if (form.ports.length <= 1) { ElMessage.warning('至少需要一个端口'); return }
  form.ports.splice(i, 1)
}

// ---- YAML Generation ----

const generatedYaml = computed(() => {
  const resource = buildK8sService()
  return yaml.dump(resource, { indent: 2, lineWidth: -1, noRefs: true })
})

function buildK8sService(): Record<string, any> {
  const labels: Record<string, string> = {}
  form.labels.forEach(l => { if (l.key.trim()) labels[l.key.trim()] = l.value })

  const selector: Record<string, string> = {}
  form.selectors.forEach(s => { if (s.key.trim()) selector[s.key.trim()] = s.value })

  const ports = form.ports
    .filter(p => p.port && p.targetPort)
    .map(p => {
      const port: Record<string, any> = { port: p.port, targetPort: p.targetPort, protocol: p.protocol }
      if (p.name) port.name = p.name
      if (form.type === 'NodePort' && p.nodePort) port.nodePort = p.nodePort
      return port
    })

  const resource: Record<string, any> = {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: { name: form.name, namespace: form.namespace, labels: { ...labels } },
    spec: { type: form.type, selector, ports },
  }

  if (form.sessionAffinity !== 'None') resource.spec.sessionAffinity = form.sessionAffinity
  if (form.type === 'LoadBalancer' || form.type === 'NodePort') {
    resource.spec.externalTrafficPolicy = form.externalTrafficPolicy
  }

  return resource
}

// ---- Submit ----

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  for (let i = 0; i < form.ports.length; i++) {
    if (!form.ports[i].port) { ElMessage.error(`端口 ${i + 1}: 端口号不能为空`); return }
    if (!form.ports[i].targetPort) { ElMessage.error(`端口 ${i + 1}: 目标端口不能为空`); return }
  }

  submitting.value = true
  try {
    await createService({ namespace: form.namespace, yaml: generatedYaml.value })
    ElMessage.success('Service 创建成功')
    router.push('/network/services')
  } catch (e: any) {
    ElMessage.error(e?.message || '创建失败')
  } finally {
    submitting.value = false
  }
}

function handleCancel() { router.push('/network/services') }
</script>

<template>
  <div class="service-form">
    <el-form ref="formRef" :model="form" :rules="formRules" label-position="top">

      <!-- Section 1: Basic Info -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">基本信息</div>
        </div>
        <div class="section-content">
          <div class="fields-grid">
            <el-form-item label="名称" prop="name">
              <el-input v-model="form.name" placeholder="my-service" />
            </el-form-item>
            <el-form-item label="命名空间" prop="namespace">
              <el-select v-model="form.namespace" filterable placeholder="选择命名空间" style="width: 100%;" :loading="namespaceLoading">
                <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
              </el-select>
            </el-form-item>
            <el-form-item label="类型" prop="type">
              <el-select v-model="form.type" style="width: 100%;">
                <el-option label="ClusterIP" value="ClusterIP" />
                <el-option label="NodePort" value="NodePort" />
                <el-option label="LoadBalancer" value="LoadBalancer" />
              </el-select>
            </el-form-item>
            <el-form-item label="Session Affinity">
              <el-select v-model="form.sessionAffinity" style="width: 100%;">
                <el-option label="None" value="None" />
                <el-option label="ClientIP" value="ClientIP" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="form.type === 'LoadBalancer' || form.type === 'NodePort'" label="外部流量策略">
              <el-select v-model="form.externalTrafficPolicy" style="width: 100%;">
                <el-option label="Cluster" value="Cluster" />
                <el-option label="Local" value="Local" />
              </el-select>
            </el-form-item>
          </div>
        </div>
      </div>

      <!-- Section 2: Labels & Selector -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">标签与选择器</div>
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
          <el-form-item label="Selector">
            <div style="width: 100%;">
              <div v-for="(sel, i) in form.selectors" :key="i" class="kv-row">
                <el-input v-model="sel.key" placeholder="Key" />
                <el-input v-model="sel.value" placeholder="Value" />
                <el-button type="danger" text circle :disabled="form.selectors.length <= 1" @click="removeSelector(i)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
              <el-button text type="primary" @click="addSelector" size="small">
                <el-icon><Plus /></el-icon> 添加选择器
              </el-button>
            </div>
          </el-form-item>
        </div>
      </div>

      <!-- Section 3: Ports -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">端口配置</div>
        </div>
        <div class="section-content">
          <el-form-item label="端口" required>
            <div style="width: 100%;">
              <div v-for="(port, pi) in form.ports" :key="pi" class="port-card">
                <div class="port-row">
                  <el-input v-model="port.name" placeholder="名称 (可选)" style="width: 120px;" />
                  <el-input-number v-model="port.port" :min="1" :max="65535" placeholder="Port" style="flex: 1;" />
                  <el-input-number v-model="port.targetPort" :min="1" :max="65535" placeholder="Target Port" style="flex: 1;" />
                  <el-select v-model="port.protocol" style="width: 100px;">
                    <el-option label="TCP" value="TCP" />
                    <el-option label="UDP" value="UDP" />
                    <el-option label="SCTP" value="SCTP" />
                  </el-select>
                  <el-input-number
                    v-if="form.type === 'NodePort'"
                    v-model="port.nodePort"
                    :min="30000"
                    :max="32767"
                    placeholder="NodePort"
                    style="width: 140px;"
                  />
                  <el-button type="danger" text circle @click="removePort(pi)">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
              </div>
              <el-button text type="primary" size="small" @click="addPort">
                <el-icon><Plus /></el-icon> 添加端口
              </el-button>
            </div>
          </el-form-item>
        </div>
      </div>

      <!-- Submit Button -->
      <div class="form-section">
        <div class="section-sidebar"></div>
        <div class="section-content">
          <div class="form-actions">
            <el-button @click="handleCancel">取消</el-button>
            <el-button type="primary" :loading="submitting" @click="handleSubmit">创建</el-button>
          </div>
        </div>
      </div>
    </el-form>
  </div>
</template>

<style scoped>
.service-form {
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

/* Port cards */
.port-card {
  border: 1px solid var(--el-border-color-extra-light);
  border-radius: 8px;
  padding: 14px;
  margin-bottom: 8px;
  background: var(--el-fill-color-lighter);
}

.port-row {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}
</style>
