<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Delete, Plus } from '@element-plus/icons-vue'
import yaml from 'js-yaml'
import { createVpa, updateVpa, getNamespaceList, extractNamespaceNames } from '@/api/resource'

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
const loading = ref(false)
const namespaceList = ref<string[]>([])

interface Label { key: string; value: string }
interface ContainerPolicy {
  containerName: string
  mode: string
  controlledResources: string[]
  minCpu: string
  minMemory: string
  maxCpu: string
  maxMemory: string
}

function clone<T>(value: T): T {
  return JSON.parse(JSON.stringify(value || {}))
}

const form = ref({
  name: '',
  namespace: 'default',
  targetAPIVersion: 'apps/v1',
  targetKind: 'Deployment',
  targetName: '',
  updateMode: 'Off',
  labels: [{ key: 'app', value: '' }] as Label[],
  containerPolicies: [{
    containerName: '*',
    mode: 'Auto',
    controlledResources: ['cpu', 'memory'],
    minCpu: '',
    minMemory: '',
    maxCpu: '',
    maxMemory: '',
  }] as ContainerPolicy[],
})

const rules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  namespace: [{ required: true, message: '请选择命名空间', trigger: 'change' }],
  targetAPIVersion: [{ required: true, message: '请输入目标 API 版本', trigger: 'blur' }],
  targetKind: [{ required: true, message: '请选择目标类型', trigger: 'change' }],
  targetName: [{ required: true, message: '请输入目标名称', trigger: 'blur' }],
}

const formRef = ref()

const riskyUpdateMode = computed(() => ['Auto', 'Recreate'].includes(form.value.updateMode))

function parseInitialData(data: any) {
  form.value.name = data.metadata?.name || ''
  form.value.namespace = data.metadata?.namespace || 'default'
  form.value.targetAPIVersion = data.spec?.targetRef?.apiVersion || 'apps/v1'
  form.value.targetKind = data.spec?.targetRef?.kind || 'Deployment'
  form.value.targetName = data.spec?.targetRef?.name || ''
  form.value.updateMode = data.spec?.updatePolicy?.updateMode || ''

  const labels = data.metadata?.labels || {}
  const labelEntries = Object.entries(labels).map(([k, v]) => ({ key: k, value: String(v) }))
  if (labelEntries.length > 0) form.value.labels = labelEntries

  const policies = data.spec?.resourcePolicy?.containerPolicies || []
  if (policies.length > 0) {
    form.value.containerPolicies = policies.map((p: any) => ({
      containerName: p.containerName || '*',
      mode: p.mode || 'Auto',
      controlledResources: p.controlledResources || ['cpu', 'memory'],
      minCpu: p.minAllowed?.cpu || '',
      minMemory: p.minAllowed?.memory || '',
      maxCpu: p.maxAllowed?.cpu || '',
      maxMemory: p.maxAllowed?.memory || '',
    }))
  }
}

function addPolicy() {
  form.value.containerPolicies.push({
    containerName: '*',
    mode: 'Auto',
    controlledResources: ['cpu', 'memory'],
    minCpu: '',
    minMemory: '',
    maxCpu: '',
    maxMemory: '',
  })
}

function buildYaml(): string {
  const labels: Record<string, string> = {}
  form.value.labels.forEach(l => { if (l.key.trim()) labels[l.key.trim()] = l.value })

  const containerPolicies = form.value.containerPolicies.map((p) => {
    const policy: any = {
      containerName: p.containerName || '*',
      mode: p.mode || 'Auto',
      controlledResources: p.controlledResources.length ? p.controlledResources : ['cpu', 'memory'],
    }
    const minAllowed: Record<string, string> = {}
    const maxAllowed: Record<string, string> = {}
    if (p.minCpu.trim()) minAllowed.cpu = p.minCpu.trim()
    if (p.minMemory.trim()) minAllowed.memory = p.minMemory.trim()
    if (p.maxCpu.trim()) maxAllowed.cpu = p.maxCpu.trim()
    if (p.maxMemory.trim()) maxAllowed.memory = p.maxMemory.trim()
    if (Object.keys(minAllowed).length) policy.minAllowed = minAllowed
    if (Object.keys(maxAllowed).length) policy.maxAllowed = maxAllowed
    return policy
  })

  const vpa: any = props.isEdit ? clone(props.initialData) : {
    apiVersion: 'autoscaling.k8s.io/v1',
    kind: 'VerticalPodAutoscaler',
    metadata: {},
    spec: {},
  }
  delete vpa.status
  vpa.apiVersion = vpa.apiVersion || 'autoscaling.k8s.io/v1'
  vpa.kind = vpa.kind || 'VerticalPodAutoscaler'
  vpa.metadata = {
    ...(vpa.metadata || {}),
    name: form.value.name,
    namespace: form.value.namespace,
    ...(Object.keys(labels).length > 0 ? { labels } : { labels: undefined }),
  }
  if (!Object.keys(labels).length) delete vpa.metadata.labels
  vpa.spec = {
    ...(vpa.spec || {}),
    targetRef: {
      ...(vpa.spec?.targetRef || {}),
      apiVersion: form.value.targetAPIVersion,
      kind: form.value.targetKind,
      name: form.value.targetName,
    },
    resourcePolicy: {
      ...(vpa.spec?.resourcePolicy || {}),
      containerPolicies,
    },
  }
  if (form.value.updateMode) {
    vpa.spec.updatePolicy = {
      ...(vpa.spec?.updatePolicy || {}),
      updateMode: form.value.updateMode,
    }
  } else if (vpa.spec?.updatePolicy) {
    delete vpa.spec.updatePolicy.updateMode
    if (Object.keys(vpa.spec.updatePolicy).length === 0) delete vpa.spec.updatePolicy
  }

  return yaml.dump(vpa, { indent: 2, lineWidth: -1, noRefs: true })
}

async function handleSubmit() {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  loading.value = true
  try {
    const yamlContent = buildYaml()
    if (props.isEdit) {
      await updateVpa({ namespace: form.value.namespace, yaml: yamlContent })
      ElMessage.success('VPA 更新成功')
    } else {
      await createVpa({ namespace: form.value.namespace, yaml: yamlContent })
      ElMessage.success('VPA 创建成功')
    }
    if (props.isEdit) {
      emit('success')
    } else {
      router.push('/autoscaling/vpa')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || (props.isEdit ? '更新失败' : '创建失败'))
  } finally {
    loading.value = false
  }
}

function handleCancel() {
  if (props.isEdit) {
    emit('cancel')
  } else {
    router.push('/autoscaling/vpa')
  }
}

async function fetchNamespaces() {
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = extractNamespaceNames(res.data)
  } catch { /* ignore */ }
}

onMounted(() => {
  fetchNamespaces()
  if (props.isEdit && props.initialData) {
    parseInitialData(props.initialData)
  }
})
</script>

<template>
  <div class="vpa-form">
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">基本信息</div>
        </div>
        <div class="section-content">
          <div class="fields-grid">
            <el-form-item label="名称" prop="name">
              <el-input v-model="form.name" placeholder="请输入VPA名称" />
            </el-form-item>
            <el-form-item label="命名空间" prop="namespace">
              <el-select v-model="form.namespace" filterable placeholder="请选择命名空间" style="width: 100%;">
                <el-option v-for="ns in namespaceList" :key="ns" :label="ns" :value="ns" />
              </el-select>
            </el-form-item>
          </div>
        </div>
      </div>

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
                <el-button type="danger" text circle :disabled="form.labels.length <= 1" @click="form.labels.splice(i, 1)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
              <el-button text type="primary" @click="form.labels.push({ key: '', value: '' })" size="small">
                <el-icon><Plus /></el-icon> 添加标签
              </el-button>
            </div>
          </el-form-item>
        </div>
      </div>

      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">目标资源</div>
        </div>
        <div class="section-content">
          <div class="fields-grid">
            <el-form-item label="API 版本" prop="targetAPIVersion">
              <el-input v-model="form.targetAPIVersion" placeholder="apps/v1" />
            </el-form-item>
            <el-form-item label="目标类型" prop="targetKind">
              <el-select v-model="form.targetKind" style="width: 100%;">
                <el-option label="Deployment" value="Deployment" />
                <el-option label="StatefulSet" value="StatefulSet" />
                <el-option label="DaemonSet" value="DaemonSet" />
                <el-option label="ReplicaSet" value="ReplicaSet" />
              </el-select>
            </el-form-item>
            <el-form-item label="目标名称" prop="targetName">
              <el-input v-model="form.targetName" placeholder="请输入目标工作负载名称" />
            </el-form-item>
          </div>
        </div>
      </div>

      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">更新策略</div>
        </div>
        <div class="section-content">
          <el-form-item label="Update Mode">
            <el-select v-model="form.updateMode" style="width: 240px;">
              <el-option label="未指定" value="" />
              <el-option label="Off (仅推荐)" value="Off" />
              <el-option label="Initial" value="Initial" />
              <el-option label="Recreate" value="Recreate" />
              <el-option label="Auto" value="Auto" />
            </el-select>
            <el-alert
              v-if="riskyUpdateMode"
              title="当前模式可能驱逐或重建 Pod，请确认业务可接受重启风险。"
              type="warning"
              show-icon
              :closable="false"
              class="mode-warning"
            />
          </el-form-item>
        </div>
      </div>

      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">容器策略</div>
        </div>
        <div class="section-content">
          <div v-for="(policy, index) in form.containerPolicies" :key="index" class="policy-card">
            <div class="policy-header">
              <span>策略 {{ index + 1 }}</span>
              <el-button type="danger" text circle :disabled="form.containerPolicies.length <= 1" @click="form.containerPolicies.splice(index, 1)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <div class="fields-grid">
              <el-form-item label="容器名称">
                <el-input v-model="policy.containerName" placeholder="*" />
              </el-form-item>
              <el-form-item label="模式">
                <el-select v-model="policy.mode" style="width: 100%;">
                  <el-option label="Auto" value="Auto" />
                  <el-option label="Off" value="Off" />
                </el-select>
              </el-form-item>
              <el-form-item label="控制资源">
                <el-checkbox-group v-model="policy.controlledResources">
                  <el-checkbox label="cpu">CPU</el-checkbox>
                  <el-checkbox label="memory">Memory</el-checkbox>
                </el-checkbox-group>
              </el-form-item>
              <el-form-item label="最小 CPU">
                <el-input v-model="policy.minCpu" placeholder="例如 100m" />
              </el-form-item>
              <el-form-item label="最小 Memory">
                <el-input v-model="policy.minMemory" placeholder="例如 128Mi" />
              </el-form-item>
              <el-form-item label="最大 CPU">
                <el-input v-model="policy.maxCpu" placeholder="例如 2" />
              </el-form-item>
              <el-form-item label="最大 Memory">
                <el-input v-model="policy.maxMemory" placeholder="例如 4Gi" />
              </el-form-item>
            </div>
          </div>
          <el-button text type="primary" @click="addPolicy" size="small">
            <el-icon><Plus /></el-icon> 添加容器策略
          </el-button>
        </div>
      </div>

      <div class="form-section">
        <div class="section-sidebar"></div>
        <div class="section-content">
          <div class="form-actions">
            <el-button @click="handleCancel">取消</el-button>
            <el-button type="primary" :loading="loading" @click="handleSubmit">{{ isEdit ? '更新' : '创建' }}</el-button>
          </div>
        </div>
      </div>
    </el-form>
  </div>
</template>

<style scoped>
.vpa-form {
  padding: 0 40px;
  max-width: 1000px;
  margin: 0 auto;
}

.form-section {
  display: flex;
  gap: 24px;
  margin-bottom: 32px;
  align-items: flex-start;
}

.section-sidebar {
  width: 120px;
  flex-shrink: 0;
  padding-top: 8px;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.section-content {
  flex: 1;
  min-width: 0;
}

.fields-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px 20px;
}

.kv-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: center;
}

.mode-warning {
  margin-top: 12px;
}

.policy-card {
  border: 1px solid var(--el-border-color-light);
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
  background: var(--el-fill-color-lighter);
}

.policy-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
  margin-bottom: 12px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
