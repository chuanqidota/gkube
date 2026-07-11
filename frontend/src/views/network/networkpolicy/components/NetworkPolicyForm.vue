<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'
import yaml from 'js-yaml'
import type { FormInstance, FormRules } from 'element-plus'
import { getNamespaceList, createNetworkPolicy, updateNetworkPolicyYaml, extractNamespaceNames } from '@/api/resource'

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

// ---- Data Model ----

interface Label { key: string; value: string }

interface FromToEntry {
  type: 'podSelector' | 'namespaceSelector' | 'ipBlock'
  labels: Label[]
  cidr: string
  except: string[]
}

interface PortEntry {
  protocol: string
  port: number | null
}

interface RuleItem {
  fromTo: FromToEntry[]
  ports: PortEntry[]
}

const form = reactive({
  name: '',
  namespace: 'default',
  policyTypes: ['Ingress', 'Egress'] as string[],
  podSelectorLabels: [{ key: 'app', value: '' }] as Label[],
  ingressRules: [{
    fromTo: [{ type: 'podSelector', labels: [{ key: 'app', value: '' }], cidr: '', except: [] }],
    ports: [{ protocol: 'TCP', port: 80 }],
  }] as RuleItem[],
  egressRules: [{
    fromTo: [{ type: 'ipBlock', labels: [], cidr: '0.0.0.0/0', except: [] }],
    ports: [{ protocol: 'TCP', port: 443 }],
  }] as RuleItem[],
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

// ---- Pod Selector Management ----

function addPodSelectorLabel() { form.podSelectorLabels.push({ key: '', value: '' }) }
function removePodSelectorLabel(i: number) { form.podSelectorLabels.splice(i, 1) }

// ---- Ingress Rule Management ----

function addIngressRule() {
  form.ingressRules.push({
    fromTo: [{ type: 'podSelector', labels: [{ key: '', value: '' }], cidr: '', except: [] }],
    ports: [{ protocol: 'TCP', port: null }],
  })
}
function removeIngressRule(i: number) { form.ingressRules.splice(i, 1) }

function addIngressFrom(ruleIdx: number) {
  form.ingressRules[ruleIdx].fromTo.push({ type: 'podSelector', labels: [{ key: '', value: '' }], cidr: '', except: [] })
}
function removeIngressFrom(ruleIdx: number, fromIdx: number) {
  form.ingressRules[ruleIdx].fromTo.splice(fromIdx, 1)
}

function addIngressPort(ruleIdx: number) {
  form.ingressRules[ruleIdx].ports.push({ protocol: 'TCP', port: null })
}
function removeIngressPort(ruleIdx: number, portIdx: number) {
  form.ingressRules[ruleIdx].ports.splice(portIdx, 1)
}

// ---- Egress Rule Management ----

function addEgressRule() {
  form.egressRules.push({
    fromTo: [{ type: 'ipBlock', labels: [], cidr: '0.0.0.0/0', except: [] }],
    ports: [{ protocol: 'TCP', port: null }],
  })
}
function removeEgressRule(i: number) { form.egressRules.splice(i, 1) }

function addEgressTo(ruleIdx: number) {
  form.egressRules[ruleIdx].fromTo.push({ type: 'podSelector', labels: [{ key: '', value: '' }], cidr: '', except: [] })
}
function removeEgressTo(ruleIdx: number, toIdx: number) {
  form.egressRules[ruleIdx].fromTo.splice(toIdx, 1)
}

function addEgressPort(ruleIdx: number) {
  form.egressRules[ruleIdx].ports.push({ protocol: 'TCP', port: null })
}
function removeEgressPort(ruleIdx: number, portIdx: number) {
  form.egressRules[ruleIdx].ports.splice(portIdx, 1)
}

// ---- FromTo Label Management ----

function addFromToLabel(rule: RuleItem, fromIdx: number) {
  rule.fromTo[fromIdx].labels.push({ key: '', value: '' })
}
function removeFromToLabel(rule: RuleItem, fromIdx: number, labelIdx: number) {
  rule.fromTo[fromIdx].labels.splice(labelIdx, 1)
}

// ---- Except Management (ipBlock) ----

function addExcept(rule: RuleItem, fromIdx: number) {
  rule.fromTo[fromIdx].except.push('')
}
function removeExcept(rule: RuleItem, fromIdx: number, exceptIdx: number) {
  rule.fromTo[fromIdx].except.splice(exceptIdx, 1)
}

function buildFromTo(entries: FromToEntry[]): Record<string, any>[] {
  return entries
    .filter(e => {
      if (e.type === 'ipBlock') return !!e.cidr.trim()
      return e.labels.some(l => l.key.trim())
    })
    .map(e => {
      if (e.type === 'ipBlock') {
        const ipBlock: Record<string, any> = { cidr: e.cidr.trim() }
        const validExcept = e.except.filter(x => x.trim())
        if (validExcept.length > 0) ipBlock.except = validExcept
        return { ipBlock }
      }
      const labels: Record<string, string> = {}
      e.labels.forEach(l => { if (l.key.trim()) labels[l.key.trim()] = l.value })
      const selector = { matchLabels: labels }
      return e.type === 'podSelector' ? { podSelector: selector } : { namespaceSelector: selector }
    })
}

function buildPorts(entries: PortEntry[]): Record<string, any>[] {
  return entries
    .filter(p => p.port != null && p.port > 0)
    .map(p => ({ protocol: p.protocol, port: p.port }))
}

function buildNetworkPolicy(): Record<string, any> {
  // Pod Selector
  const matchLabels: Record<string, string> = {}
  form.podSelectorLabels.forEach(l => { if (l.key.trim()) matchLabels[l.key.trim()] = l.value })

  const resource: Record<string, any> = {
    apiVersion: 'networking.k8s.io/v1',
    kind: 'NetworkPolicy',
    metadata: { name: form.name, namespace: form.namespace },
    spec: {
      podSelector: { matchLabels },
      policyTypes: form.policyTypes,
    },
  }

  // Ingress
  if (form.policyTypes.includes('Ingress') && form.ingressRules.length > 0) {
    const ingress = form.ingressRules
      .map(rule => {
        const item: Record<string, any> = {}
        const from = buildFromTo(rule.fromTo)
        if (from.length > 0) item.from = from
        const ports = buildPorts(rule.ports)
        if (ports.length > 0) item.ports = ports
        return item
      })
      .filter(r => Object.keys(r).length > 0)
    if (ingress.length > 0) resource.spec.ingress = ingress
  }

  // Egress
  if (form.policyTypes.includes('Egress') && form.egressRules.length > 0) {
    const egress = form.egressRules
      .map(rule => {
        const item: Record<string, any> = {}
        const to = buildFromTo(rule.fromTo)
        if (to.length > 0) item.to = to
        const ports = buildPorts(rule.ports)
        if (ports.length > 0) item.ports = ports
        return item
      })
      .filter(r => Object.keys(r).length > 0)
    if (egress.length > 0) resource.spec.egress = egress
  }

  return resource
}

// ---- Parse Initial Data (Edit Mode) ----

function parseInitialData(data: any) {
  const spec = data.spec || {}
  const meta = data.metadata || {}

  form.name = meta.name || ''
  form.namespace = meta.namespace || 'default'
  form.policyTypes = spec.policyTypes || ['Ingress', 'Egress']

  // Pod Selector
  const matchLabels = spec.podSelector?.matchLabels || {}
  form.podSelectorLabels = Object.keys(matchLabels).length > 0
    ? Object.entries(matchLabels).map(([k, v]) => ({ key: k, value: v as string }))
    : [{ key: 'app', value: '' }]

  // Ingress rules
  const ingress = spec.ingress || []
  form.ingressRules = ingress.length > 0
    ? ingress.map((rule: any) => ({
        fromTo: (rule.from || []).map((f: any) => {
          if (f.ipBlock) return { type: 'ipBlock' as const, labels: [], cidr: f.ipBlock.cidr || '', except: f.ipBlock.except || [] }
          if (f.namespaceSelector) return { type: 'namespaceSelector' as const, labels: Object.entries(f.namespaceSelector.matchLabels || {}).map(([k, v]) => ({ key: k, value: v as string })), cidr: '', except: [] }
          return { type: 'podSelector' as const, labels: Object.entries(f.podSelector?.matchLabels || {}).map(([k, v]) => ({ key: k, value: v as string })), cidr: '', except: [] }
        }),
        ports: (rule.ports || []).map((p: any) => ({ protocol: p.protocol || 'TCP', port: p.port ?? null })),
      }))
    : [{ fromTo: [{ type: 'podSelector' as const, labels: [{ key: 'app', value: '' }], cidr: '', except: [] }], ports: [{ protocol: 'TCP', port: 80 }] }]

  // Egress rules
  const egress = spec.egress || []
  form.egressRules = egress.length > 0
    ? egress.map((rule: any) => ({
        fromTo: (rule.to || []).map((t: any) => {
          if (t.ipBlock) return { type: 'ipBlock' as const, labels: [], cidr: t.ipBlock.cidr || '', except: t.ipBlock.except || [] }
          if (t.namespaceSelector) return { type: 'namespaceSelector' as const, labels: Object.entries(t.namespaceSelector.matchLabels || {}).map(([k, v]) => ({ key: k, value: v as string })), cidr: '', except: [] }
          return { type: 'podSelector' as const, labels: Object.entries(t.podSelector?.matchLabels || {}).map(([k, v]) => ({ key: k, value: v as string })), cidr: '', except: [] }
        }),
        ports: (rule.ports || []).map((p: any) => ({ protocol: p.protocol || 'TCP', port: p.port ?? null })),
      }))
    : [{ fromTo: [{ type: 'ipBlock' as const, labels: [], cidr: '0.0.0.0/0', except: [] }], ports: [{ protocol: 'TCP', port: 443 }] }]
}

// ---- Submit ----

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    const resource = buildNetworkPolicy()
    const yamlContent = yaml.dump(resource, { indent: 2, lineWidth: -1, noRefs: true })
    if (props.isEdit) {
      await updateNetworkPolicyYaml({ namespace: form.namespace, name: form.name, yaml: yamlContent })
      ElMessage.success('NetworkPolicy 更新成功')
      emit('success')
    } else {
      await createNetworkPolicy({ namespace: form.namespace, yaml: yamlContent })
      ElMessage.success('NetworkPolicy 创建成功')
      router.push('/network/networkpolicies')
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
    router.push('/network/networkpolicies')
  }
}
</script>

<template>
  <div class="np-form">
    <el-form ref="formRef" :model="form" :rules="formRules" label-position="top">

      <!-- Section 1: Basic Info -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">基本信息</div>
        </div>
        <div class="section-content">
          <div class="fields-grid">
            <el-form-item label="名称" prop="name">
              <el-input v-model="form.name" placeholder="my-network-policy" />
            </el-form-item>
            <el-form-item label="命名空间" prop="namespace">
              <el-select v-model="form.namespace" filterable placeholder="选择命名空间" style="width: 100%;" :loading="namespaceLoading">
                <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
              </el-select>
            </el-form-item>
          </div>
        </div>
      </div>

      <!-- Section 2: Policy Types -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">策略类型</div>
        </div>
        <div class="section-content">
          <el-form-item label="策略类型">
            <el-checkbox-group v-model="form.policyTypes">
              <el-checkbox value="Ingress" label="Ingress" />
              <el-checkbox value="Egress" label="Egress" />
            </el-checkbox-group>
          </el-form-item>
        </div>
      </div>

      <!-- Section 3: Pod Selector -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">Pod 选择器</div>
        </div>
        <div class="section-content">
          <el-form-item label="Pod Selector">
            <div style="width: 100%;">
              <div v-for="(label, i) in form.podSelectorLabels" :key="i" class="kv-row">
                <el-input v-model="label.key" placeholder="Key" />
                <el-input v-model="label.value" placeholder="Value" />
                <el-button type="danger" text circle :disabled="form.podSelectorLabels.length <= 1" @click="removePodSelectorLabel(i)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
              <el-button text type="primary" @click="addPodSelectorLabel" size="small">
                <el-icon><Plus /></el-icon> 添加标签
              </el-button>
            </div>
          </el-form-item>
        </div>
      </div>

      <!-- Section 4: Ingress Rules -->
      <div v-if="form.policyTypes.includes('Ingress')" class="form-section">
        <div class="section-sidebar">
          <div class="section-title">Ingress 规则</div>
        </div>
        <div class="section-content">
          <div v-for="(rule, ri) in form.ingressRules" :key="ri" class="rule-card">
            <div class="rule-card-header">
              <span class="rule-card-title">规则 {{ ri + 1 }}</span>
              <el-button type="danger" text circle size="small" @click="removeIngressRule(ri)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>

            <!-- From entries -->
            <div class="rule-subsection">
              <div class="rule-subtitle">From 条目</div>
              <div v-for="(entry, fi) in rule.fromTo" :key="fi" class="fromto-entry">
                <div class="fromto-header">
                  <el-select v-model="entry.type" style="width: 180px;" size="small">
                    <el-option label="Pod Selector" value="podSelector" />
                    <el-option label="Namespace Selector" value="namespaceSelector" />
                    <el-option label="IP Block" value="ipBlock" />
                  </el-select>
                  <el-button type="danger" text circle size="small" @click="removeIngressFrom(ri, fi)">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>

                <!-- podSelector / namespaceSelector labels -->
                <div v-if="entry.type === 'podSelector' || entry.type === 'namespaceSelector'" class="fromto-body">
                  <div v-for="(label, li) in entry.labels" :key="li" class="kv-row">
                    <el-input v-model="label.key" placeholder="Key" size="small" />
                    <el-input v-model="label.value" placeholder="Value" size="small" />
                    <el-button type="danger" text circle size="small" @click="removeFromToLabel(rule, fi, li)">
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </div>
                  <el-button text type="primary" size="small" @click="addFromToLabel(rule, fi)">
                    <el-icon><Plus /></el-icon> 添加标签
                  </el-button>
                </div>

                <!-- ipBlock -->
                <div v-else class="fromto-body">
                  <div class="fields-grid">
                    <el-form-item label="CIDR" class="compact-form-item">
                      <el-input v-model="entry.cidr" placeholder="10.0.0.0/8" size="small" />
                    </el-form-item>
                  </div>
                  <div v-if="entry.except.length > 0" style="margin-top: 8px;">
                    <div class="rule-subtitle" style="margin-bottom: 4px;">Except</div>
                    <div v-for="(_ex, ei) in entry.except" :key="ei" class="kv-row">
                      <el-input v-model="entry.except[ei]" placeholder="10.0.1.0/24" size="small" />
                      <el-button type="danger" text circle size="small" @click="removeExcept(rule, fi, ei)">
                        <el-icon><Delete /></el-icon>
                      </el-button>
                    </div>
                  </div>
                  <el-button text type="primary" size="small" @click="addExcept(rule, fi)" style="margin-top: 4px;">
                    <el-icon><Plus /></el-icon> 添加 Except
                  </el-button>
                </div>
              </div>
              <el-button text type="primary" size="small" @click="addIngressFrom(ri)">
                <el-icon><Plus /></el-icon> 添加 From 条目
              </el-button>
            </div>

            <!-- Ports -->
            <div class="rule-subsection">
              <div class="rule-subtitle">Ports</div>
              <div v-for="(port, pi) in rule.ports" :key="pi" class="port-row">
                <el-select v-model="port.protocol" style="width: 100px;" size="small">
                  <el-option label="TCP" value="TCP" />
                  <el-option label="UDP" value="UDP" />
                  <el-option label="SCTP" value="SCTP" />
                </el-select>
                <el-input-number v-model="port.port" :min="1" :max="65535" placeholder="Port" style="flex: 1;" size="small" />
                <el-button type="danger" text circle size="small" @click="removeIngressPort(ri, pi)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
              <el-button text type="primary" size="small" @click="addIngressPort(ri)">
                <el-icon><Plus /></el-icon> 添加端口
              </el-button>
            </div>
          </div>
          <el-button text type="primary" @click="addIngressRule">
            <el-icon><Plus /></el-icon> 添加 Ingress 规则
          </el-button>
        </div>
      </div>

      <!-- Section 5: Egress Rules -->
      <div v-if="form.policyTypes.includes('Egress')" class="form-section">
        <div class="section-sidebar">
          <div class="section-title">Egress 规则</div>
        </div>
        <div class="section-content">
          <div v-for="(rule, ri) in form.egressRules" :key="ri" class="rule-card">
            <div class="rule-card-header">
              <span class="rule-card-title">规则 {{ ri + 1 }}</span>
              <el-button type="danger" text circle size="small" @click="removeEgressRule(ri)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>

            <!-- To entries -->
            <div class="rule-subsection">
              <div class="rule-subtitle">To 条目</div>
              <div v-for="(entry, ti) in rule.fromTo" :key="ti" class="fromto-entry">
                <div class="fromto-header">
                  <el-select v-model="entry.type" style="width: 180px;" size="small">
                    <el-option label="Pod Selector" value="podSelector" />
                    <el-option label="Namespace Selector" value="namespaceSelector" />
                    <el-option label="IP Block" value="ipBlock" />
                  </el-select>
                  <el-button type="danger" text circle size="small" @click="removeEgressTo(ri, ti)">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>

                <div v-if="entry.type === 'podSelector' || entry.type === 'namespaceSelector'" class="fromto-body">
                  <div v-for="(label, li) in entry.labels" :key="li" class="kv-row">
                    <el-input v-model="label.key" placeholder="Key" size="small" />
                    <el-input v-model="label.value" placeholder="Value" size="small" />
                    <el-button type="danger" text circle size="small" @click="removeFromToLabel(rule, ti, li)">
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </div>
                  <el-button text type="primary" size="small" @click="addFromToLabel(rule, ti)">
                    <el-icon><Plus /></el-icon> 添加标签
                  </el-button>
                </div>

                <div v-else class="fromto-body">
                  <div class="fields-grid">
                    <el-form-item label="CIDR" class="compact-form-item">
                      <el-input v-model="entry.cidr" placeholder="10.0.0.0/8" size="small" />
                    </el-form-item>
                  </div>
                  <div v-if="entry.except.length > 0" style="margin-top: 8px;">
                    <div class="rule-subtitle" style="margin-bottom: 4px;">Except</div>
                    <div v-for="(_ex, ei) in entry.except" :key="ei" class="kv-row">
                      <el-input v-model="entry.except[ei]" placeholder="10.0.1.0/24" size="small" />
                      <el-button type="danger" text circle size="small" @click="removeExcept(rule, ti, ei)">
                        <el-icon><Delete /></el-icon>
                      </el-button>
                    </div>
                  </div>
                  <el-button text type="primary" size="small" @click="addExcept(rule, ti)" style="margin-top: 4px;">
                    <el-icon><Plus /></el-icon> 添加 Except
                  </el-button>
                </div>
              </div>
              <el-button text type="primary" size="small" @click="addEgressTo(ri)">
                <el-icon><Plus /></el-icon> 添加 To 条目
              </el-button>
            </div>

            <!-- Ports -->
            <div class="rule-subsection">
              <div class="rule-subtitle">Ports</div>
              <div v-for="(port, pi) in rule.ports" :key="pi" class="port-row">
                <el-select v-model="port.protocol" style="width: 100px;" size="small">
                  <el-option label="TCP" value="TCP" />
                  <el-option label="UDP" value="UDP" />
                  <el-option label="SCTP" value="SCTP" />
                </el-select>
                <el-input-number v-model="port.port" :min="1" :max="65535" placeholder="Port" style="flex: 1;" size="small" />
                <el-button type="danger" text circle size="small" @click="removeEgressPort(ri, pi)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
              <el-button text type="primary" size="small" @click="addEgressPort(ri)">
                <el-icon><Plus /></el-icon> 添加端口
              </el-button>
            </div>
          </div>
          <el-button text type="primary" @click="addEgressRule">
            <el-icon><Plus /></el-icon> 添加 Egress 规则
          </el-button>
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
.np-form {
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

.compact-form-item {
  margin-bottom: 8px !important;
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
  padding: 16px;
  margin-bottom: 12px;
  background: var(--el-fill-color-lighter);
}

.rule-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.rule-card-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.rule-subsection {
  margin-bottom: 12px;
}

.rule-subsection:last-child {
  margin-bottom: 0;
}

.rule-subtitle {
  font-size: 12px;
  font-weight: 600;
  color: var(--el-text-color-secondary);
  margin-bottom: 8px;
}

/* FromTo entries */
.fromto-entry {
  border: 1px solid var(--el-border-color-extra-light);
  border-radius: 6px;
  padding: 12px;
  margin-bottom: 8px;
  background: var(--el-bg-color);
}

.fromto-header {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 8px;
}

.fromto-body {
  padding-left: 8px;
}

/* Port rows */
.port-row {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 8px;
}

</style>
