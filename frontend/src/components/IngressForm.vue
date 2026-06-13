<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import yaml from 'js-yaml'
import type { FormInstance, FormRules } from 'element-plus'
import YamlEditor from '@/components/YamlEditor.vue'
import { getNamespaceList, createIngress } from '@/api/resource'

const router = useRouter()
const currentStep = ref(0)
const submitting = ref(false)
const namespaceLoading = ref(false)
const namespaces = ref<string[]>([])

// ---- Form Data ----

interface Label {
  key: string
  value: string
}

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

const step0FormRef = ref<FormInstance>()

const step0Rules: FormRules = {
  name: [
    { required: true, message: 'Name is required', trigger: 'blur' },
    { pattern: /^[a-z][a-z0-9-]*[a-z0-9]$/, message: 'Lowercase letters, numbers, hyphens only. Must start with letter, end with alphanumeric.', trigger: 'blur' },
    { max: 253, message: 'Max 253 characters', trigger: 'blur' },
  ],
  namespace: [{ required: true, message: 'Namespace is required', trigger: 'change' }],
  ingressClassName: [{ required: true, message: 'Ingress class name is required', trigger: 'blur' }],
}

// ---- Steps ----

const steps = [
  { title: 'Basic Info', icon: 'Document' },
  { title: 'Rules & TLS', icon: 'Connection' },
  { title: 'YAML Preview', icon: 'View' },
]

// ---- Namespace Fetch ----

async function fetchNamespaces() {
  namespaceLoading.value = true
  try {
    const res: any = await getNamespaceList()
    namespaces.value = (res.data || []).map((ns: any) => ns.name || ns)
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

// ---- Rule Management ----

function addRule() {
  form.rules.push({ host: '', path: '/', pathType: 'Prefix', backendService: '', backendPort: 80 })
}

function removeRule(index: number) {
  if (form.rules.length <= 1) {
    ElMessage.warning('At least one rule is required')
    return
  }
  form.rules.splice(index, 1)
}

// ---- TLS Management ----

function addTls() {
  form.tls.push({ hosts: '', secretName: '' })
}

function removeTls(index: number) {
  if (form.tls.length <= 1) {
    ElMessage.warning('At least one TLS entry is required')
    return
  }
  form.tls.splice(index, 1)
}

// ---- Step Navigation ----

async function handleNext() {
  if (currentStep.value === 0) {
    const valid = await step0FormRef.value?.validate().catch(() => false)
    if (!valid) return
  }
  if (currentStep.value === 1) {
    const valid = validateRules()
    if (!valid) return
  }
  if (currentStep.value < 2) {
    currentStep.value++
  }
}

function handlePrev() {
  if (currentStep.value > 0) {
    currentStep.value--
  }
}

function handleStepClick(step: number) {
  if (step <= currentStep.value) {
    currentStep.value = step
  }
}

function validateRules(): boolean {
  for (let i = 0; i < form.rules.length; i++) {
    const r = form.rules[i]
    if (!r.host.trim()) {
      ElMessage.error(`Rule ${i + 1}: host is required`)
      return false
    }
    if (!r.backendService.trim()) {
      ElMessage.error(`Rule ${i + 1}: backend service name is required`)
      return false
    }
    if (!r.backendPort) {
      ElMessage.error(`Rule ${i + 1}: backend port is required`)
      return false
    }
  }
  return true
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
        paths: [
          {
            path: r.path,
            pathType: r.pathType,
            backend: {
              service: {
                name: r.backendService,
                port: {
                  number: r.backendPort,
                },
              },
            },
          },
        ],
      },
    }))

  const resource: Record<string, any> = {
    apiVersion: 'networking.k8s.io/v1',
    kind: 'Ingress',
    metadata: {
      name: form.name,
      namespace: form.namespace,
      labels: { ...labels },
      annotations: {
        'kubernetes.io/ingress.class': form.ingressClassName,
      },
    },
    spec: {
      ingressClassName: form.ingressClassName,
      rules,
    },
  }

  if (form.tlsEnabled) {
    const tls = form.tls
      .filter((t) => t.hosts.trim())
      .map((t) => ({
        hosts: t.hosts.split(',').map((h) => h.trim()).filter(Boolean),
        secretName: t.secretName,
      }))
    if (tls.length > 0) {
      resource.spec.tls = tls
    }
  }

  return resource
}

// ---- Submit ----

async function handleSubmit() {
  submitting.value = true
  try {
    const yamlContent = generatedYaml.value
    await createIngress({
      namespace: form.namespace,
      yamlContent,
    })
    ElMessage.success('Ingress created successfully')
    router.push('/ingresses')
  } catch (e: any) {
    ElMessage.error(e?.message || 'Create failed')
  } finally {
    submitting.value = false
  }
}

function handleCancel() {
  router.push('/ingresses')
}
</script>

<template>
  <div class="ingress-form">
    <!-- Header -->
    <div class="form-header">
      <h2>Create Ingress</h2>
    </div>

    <!-- Steps -->
    <el-steps :active="currentStep" finish-status="success" align-center style="margin-bottom: 32px;">
      <el-step
        v-for="(step, index) in steps"
        :key="index"
        :title="step.title"
        :icon="step.icon"
        @click="handleStepClick(index)"
        style="cursor: pointer;"
      />
    </el-steps>

    <!-- Step Content -->
    <div class="step-content">

      <!-- Step 0: Basic Info -->
      <div v-show="currentStep === 0">
        <el-form
          ref="step0FormRef"
          :model="form"
          :rules="step0Rules"
          label-width="180px"
          style="max-width: 700px;"
        >
          <el-form-item label="Name" prop="name">
            <el-input v-model="form.name" placeholder="e.g. my-ingress" />
          </el-form-item>

          <el-form-item label="Namespace" prop="namespace">
            <el-select
              v-model="form.namespace"
              filterable
              placeholder="Select namespace"
              style="width: 100%;"
              :loading="namespaceLoading"
            >
              <el-option
                v-for="ns in namespaces"
                :key="ns"
                :label="ns"
                :value="ns"
              />
            </el-select>
          </el-form-item>

          <el-form-item label="Ingress Class Name" prop="ingressClassName">
            <el-input v-model="form.ingressClassName" placeholder="e.g. nginx" />
          </el-form-item>

          <el-form-item label="Labels">
            <div style="width: 100%;">
              <div
                v-for="(label, index) in form.labels"
                :key="index"
                style="display: flex; gap: 8px; margin-bottom: 8px;"
              >
                <el-input v-model="label.key" placeholder="Key" style="flex: 1;" />
                <el-input v-model="label.value" placeholder="Value" style="flex: 1;" />
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
                <el-icon><Plus /></el-icon> Add Label
              </el-button>
            </div>
          </el-form-item>
        </el-form>
      </div>

      <!-- Step 1: Rules & TLS -->
      <div v-show="currentStep === 1">
        <el-form label-width="180px" style="max-width: 700px;">
          <!-- Rules -->
          <el-form-item label="Rules" required>
            <div style="width: 100%;">
              <div
                v-for="(rule, ri) in form.rules"
                :key="ri"
                class="rule-card"
              >
                <div style="display: flex; gap: 8px; flex-wrap: wrap; align-items: center; margin-bottom: 8px;">
                  <el-input v-model="rule.host" placeholder="Host (e.g. example.com)" style="flex: 1; min-width: 200px;" />
                  <el-input v-model="rule.path" placeholder="Path" style="width: 150px;" />
                  <el-select v-model="rule.pathType" style="width: 130px;">
                    <el-option label="Prefix" value="Prefix" />
                    <el-option label="Exact" value="Exact" />
                    <el-option label="ImplementationSpecific" value="ImplementationSpecific" />
                  </el-select>
                  <el-button type="danger" circle size="small" @click="removeRule(ri)">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
                <div style="display: flex; gap: 8px; align-items: center;">
                  <span style="font-size: 13px; color: #606266; white-space: nowrap;">Backend:</span>
                  <el-input v-model="rule.backendService" placeholder="Service name" style="flex: 1;" />
                  <el-input-number v-model="rule.backendPort" :min="1" :max="65535" placeholder="Port" style="width: 140px;" />
                </div>
              </div>
              <el-button size="small" @click="addRule">
                <el-icon><Plus /></el-icon> Add Rule
              </el-button>
            </div>
          </el-form-item>

          <!-- TLS -->
          <el-form-item label="Enable TLS">
            <el-switch v-model="form.tlsEnabled" />
          </el-form-item>

          <template v-if="form.tlsEnabled">
            <el-form-item label="TLS Configuration">
              <div style="width: 100%;">
                <div
                  v-for="(t, ti) in form.tls"
                  :key="ti"
                  class="tls-card"
                >
                  <div style="display: flex; gap: 8px; align-items: center;">
                    <el-input v-model="t.hosts" placeholder="Hosts (comma-separated)" style="flex: 1;" />
                    <el-input v-model="t.secretName" placeholder="Secret name" style="width: 200px;" />
                    <el-button type="danger" circle size="small" @click="removeTls(ti)">
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </div>
                </div>
                <el-button size="small" @click="addTls">
                  <el-icon><Plus /></el-icon> Add TLS Entry
                </el-button>
              </div>
            </el-form-item>
          </template>
        </el-form>
      </div>

      <!-- Step 2: YAML Preview -->
      <div v-show="currentStep === 2">
        <el-alert
          title="Generated Ingress YAML"
          description="Review the generated YAML below before creating the resource."
          type="info"
          :closable="false"
          show-icon
          style="margin-bottom: 16px;"
        />
        <YamlEditor
          :model-value="generatedYaml"
          height="500px"
          read-only
        />
      </div>
    </div>

    <!-- Navigation Buttons -->
    <div class="form-actions">
      <el-button @click="handleCancel">Cancel</el-button>
      <el-button v-if="currentStep > 0" @click="handlePrev">Previous</el-button>
      <el-button
        v-if="currentStep < 2"
        type="primary"
        @click="handleNext"
      >
        Next
      </el-button>
      <el-button
        v-if="currentStep === 2"
        type="primary"
        :loading="submitting"
        @click="handleSubmit"
      >
        Create Ingress
      </el-button>
    </div>
  </div>
</template>

<style scoped>
.ingress-form {
  max-width: 900px;
  margin: 0 auto;
  padding: 20px 0;
}

.form-header {
  margin-bottom: 24px;
}

.form-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.step-content {
  min-height: 400px;
  padding: 16px 0;
}

.rule-card {
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  padding: 12px;
  margin-bottom: 8px;
  background: #fafafa;
}

.tls-card {
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  padding: 12px;
  margin-bottom: 8px;
  background: #fafafa;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 24px;
  border-top: 1px solid #e4e7ed;
  margin-top: 24px;
}
</style>
