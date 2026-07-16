<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Delete, Plus } from '@element-plus/icons-vue'
import yaml from 'js-yaml'
import { createHpa, updateHpa, getNamespaceList, extractNamespaceNames } from '@/api/resource'

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
interface ScalingPolicy { type: string; value: number; periodSeconds: number }

const form = ref({
  name: '',
  namespace: 'default',
  targetKind: 'Deployment',
  targetName: '',
  minReplicas: 1,
  maxReplicas: 10,
  cpuUtilization: 80,
  memoryUtilization: 0,
  customMetrics: [] as Array<{ type: string; name: string; target: number }>,
  labels: [{ key: 'app', value: '' }] as Label[],
  behaviorEnabled: false,
  scaleUpStabilizationSeconds: 0,
  scaleUpSelectPolicy: 'Max',
  scaleUpPolicies: [] as ScalingPolicy[],
  scaleDownStabilizationSeconds: 300,
  scaleDownSelectPolicy: 'Min',
  scaleDownPolicies: [{ type: 'Percent', value: 100, periodSeconds: 15 }] as ScalingPolicy[],
})

function parseInitialData(data: any) {
  form.value.name = data.metadata?.name || ''
  form.value.namespace = data.metadata?.namespace || 'default'
  form.value.targetKind = data.spec?.scaleTargetRef?.kind || 'Deployment'
  form.value.targetName = data.spec?.scaleTargetRef?.name || ''
  form.value.minReplicas = data.spec?.minReplicas ?? 1
  form.value.maxReplicas = data.spec?.maxReplicas ?? 10

  const metrics = data.spec?.metrics || []
  for (const m of metrics) {
    if (m.type === 'Resource' && m.resource?.name === 'cpu') {
      form.value.cpuUtilization = m.resource?.target?.averageUtilization ?? 0
    } else if (m.type === 'Resource' && m.resource?.name === 'memory') {
      form.value.memoryUtilization = m.resource?.target?.averageUtilization ?? 0
    } else if (m.type !== 'Resource') {
      form.value.customMetrics.push({
        type: m.type || 'Resource',
        name: m.resource?.name || m.pods?.metric?.name || m.object?.metric?.name || '',
        target: m.resource?.target?.averageUtilization || m.pods?.target?.averageUtilization || 0,
      })
    }
  }

  // Labels
  const labels = data.metadata?.labels || {}
  const labelEntries = Object.entries(labels).map(([k, v]) => ({ key: k, value: String(v) }))
  if (labelEntries.length > 0) form.value.labels = labelEntries

  // Behavior
  const behavior = data.spec?.behavior
  if (behavior) {
    form.value.behaviorEnabled = true
    if (behavior.scaleUp) {
      form.value.scaleUpStabilizationSeconds = behavior.scaleUp.stabilizationWindowSeconds ?? 0
      form.value.scaleUpSelectPolicy = behavior.scaleUp.selectPolicy || 'Max'
      form.value.scaleUpPolicies = (behavior.scaleUp.policies || []).map((p: any) => ({
        type: p.type || 'Percent', value: p.value ?? 0, periodSeconds: p.periodSeconds ?? 60,
      }))
    }
    if (behavior.scaleDown) {
      form.value.scaleDownStabilizationSeconds = behavior.scaleDown.stabilizationWindowSeconds ?? 300
      form.value.scaleDownSelectPolicy = behavior.scaleDown.selectPolicy || 'Min'
      form.value.scaleDownPolicies = (behavior.scaleDown.policies || []).map((p: any) => ({
        type: p.type || 'Percent', value: p.value ?? 0, periodSeconds: p.periodSeconds ?? 60,
      }))
    }
  }
}

const rules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  namespace: [{ required: true, message: '请选择命名空间', trigger: 'change' }],
  targetName: [{ required: true, message: '请输入目标名称', trigger: 'blur' }],
  minReplicas: [{ required: true, message: '请输入最小副本数', trigger: 'blur' }],
  maxReplicas: [{ required: true, message: '请输入最大副本数', trigger: 'blur' }],
}

const formRef = ref()

function addCustomMetric() {
  form.value.customMetrics.push({ type: 'Resource', name: 'cpu', target: 80 })
}

function removeCustomMetric(index: number) {
  form.value.customMetrics.splice(index, 1)
}

function buildYaml(): string {
  // Labels
  const labels: Record<string, string> = {}
  form.value.labels.forEach(l => { if (l.key.trim()) labels[l.key.trim()] = l.value })

  const metrics: any[] = []

  // CPU metric
  if (form.value.cpuUtilization > 0) {
    metrics.push({
      type: 'Resource',
      resource: {
        name: 'cpu',
        target: {
          type: 'Utilization',
          averageUtilization: form.value.cpuUtilization,
        },
      },
    })
  }

  // Memory metric
  if (form.value.memoryUtilization > 0) {
    metrics.push({
      type: 'Resource',
      resource: {
        name: 'memory',
        target: {
          type: 'Utilization',
          averageUtilization: form.value.memoryUtilization,
        },
      },
    })
  }

  // Custom metrics
  for (const m of form.value.customMetrics) {
    metrics.push({
      type: m.type,
      resource: {
        name: m.name,
        target: {
          type: 'Utilization',
          averageUtilization: m.target,
        },
      },
    })
  }

  const hpa: any = {
    apiVersion: 'autoscaling/v2',
    kind: 'HorizontalPodAutoscaler',
    metadata: {
      name: form.value.name,
      namespace: form.value.namespace,
      ...(Object.keys(labels).length > 0 ? { labels } : {}),
    },
    spec: {
      scaleTargetRef: {
        apiVersion: 'apps/v1',
        kind: form.value.targetKind,
        name: form.value.targetName,
      },
      minReplicas: form.value.minReplicas,
      maxReplicas: form.value.maxReplicas,
      metrics,
    },
  }

  // Behavior
  if (form.value.behaviorEnabled) {
    const behavior: any = {}
    if (form.value.scaleUpPolicies.length > 0 || form.value.scaleUpStabilizationSeconds > 0) {
      behavior.scaleUp = {
        stabilizationWindowSeconds: form.value.scaleUpStabilizationSeconds,
        selectPolicy: form.value.scaleUpSelectPolicy,
        policies: form.value.scaleUpPolicies.map(p => ({ type: p.type, value: p.value, periodSeconds: p.periodSeconds })),
      }
    }
    if (form.value.scaleDownPolicies.length > 0) {
      behavior.scaleDown = {
        stabilizationWindowSeconds: form.value.scaleDownStabilizationSeconds,
        selectPolicy: form.value.scaleDownSelectPolicy,
        policies: form.value.scaleDownPolicies.map(p => ({ type: p.type, value: p.value, periodSeconds: p.periodSeconds })),
      }
    }
    if (Object.keys(behavior).length > 0) hpa.spec.behavior = behavior
  }

  return yaml.dump(hpa, { indent: 2, lineWidth: -1, noRefs: true })
}

async function handleSubmit() {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  loading.value = true
  try {
    const yaml = buildYaml()
    if (props.isEdit) {
      await updateHpa({ namespace: form.value.namespace, yaml })
      ElMessage.success('弹性伸缩更新成功')
    } else {
      await createHpa({ namespace: form.value.namespace, yaml })
      ElMessage.success('弹性伸缩创建成功')
    }
    if (props.isEdit) {
      emit('success')
    } else {
      router.push('/workloads/hpa')
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
    router.push('/workloads/hpa')
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
  <div class="hpa-form">
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
      <!-- Section 1: Basic Info -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">基本信息</div>
        </div>
        <div class="section-content">
          <div class="fields-grid">
            <el-form-item label="名称" prop="name">
              <el-input v-model="form.name" placeholder="请输入HPA名称" />
            </el-form-item>
            <el-form-item label="命名空间" prop="namespace">
              <el-select v-model="form.namespace" filterable placeholder="请选择命名空间" style="width: 100%;">
                <el-option v-for="ns in namespaceList" :key="ns" :label="ns" :value="ns" />
              </el-select>
            </el-form-item>
          </div>
        </div>
      </div>

      <!-- Section: Labels -->
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

      <!-- Section 2: Scale Target -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">伸缩目标</div>
        </div>
        <div class="section-content">
          <div class="fields-grid">
            <el-form-item label="目标类型" prop="targetKind">
              <el-select v-model="form.targetKind" style="width: 100%;">
                <el-option label="Deployment" value="Deployment" />
                <el-option label="StatefulSet" value="StatefulSet" />
                <el-option label="ReplicaSet" value="ReplicaSet" />
              </el-select>
            </el-form-item>
            <el-form-item label="目标名称" prop="targetName">
              <el-input v-model="form.targetName" placeholder="请输入目标工作负载名称" />
            </el-form-item>
          </div>
        </div>
      </div>

      <!-- Section 3: Replica Range -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">副本范围</div>
        </div>
        <div class="section-content">
          <div class="fields-grid">
            <el-form-item label="最小副本数" prop="minReplicas">
              <el-input-number v-model="form.minReplicas" :min="1" :max="form.maxReplicas" style="width: 100%;" />
            </el-form-item>
            <el-form-item label="最大副本数" prop="maxReplicas">
              <el-input-number v-model="form.maxReplicas" :min="form.minReplicas" :max="1000" style="width: 100%;" />
            </el-form-item>
          </div>
        </div>
      </div>

      <!-- Section 4: Metrics -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">指标配置</div>
        </div>
        <div class="section-content">
          <div class="fields-grid">
            <el-form-item label="CPU 目标 (%)">
              <el-input-number v-model="form.cpuUtilization" :min="0" :max="100" :step="5" style="width: 100%;" />
              <div class="form-tip">设为 0 表示不使用 CPU 指标</div>
            </el-form-item>
            <el-form-item label="内存目标 (%)">
              <el-input-number v-model="form.memoryUtilization" :min="0" :max="100" :step="5" style="width: 100%;" />
              <div class="form-tip">设为 0 表示不使用内存指标</div>
            </el-form-item>
          </div>
          <el-form-item label="自定义指标">
            <div style="width: 100%;">
              <div v-for="(metric, index) in form.customMetrics" :key="index" class="custom-metric-row">
                <el-select v-model="metric.type" style="width: 120px;">
                  <el-option label="Resource" value="Resource" />
                  <el-option label="Pods" value="Pods" />
                  <el-option label="Object" value="Object" />
                  <el-option label="External" value="External" />
                </el-select>
                <el-input v-model="metric.name" placeholder="指标名称" style="flex: 1;" />
                <el-input-number v-model="metric.target" :min="1" :max="10000" style="width: 140px;" />
                <el-button type="danger" text circle @click="removeCustomMetric(index)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
              <el-button text type="primary" @click="addCustomMetric" size="small">
                <el-icon><Plus /></el-icon> 添加自定义指标
              </el-button>
            </div>
          </el-form-item>
        </div>
      </div>

      <!-- Section 5: Behavior -->
      <div class="form-section">
        <div class="section-sidebar">
          <div class="section-title">扩缩容行为</div>
        </div>
        <div class="section-content">
          <el-form-item label="启用行为配置">
            <el-switch v-model="form.behaviorEnabled" />
          </el-form-item>

          <template v-if="form.behaviorEnabled">
            <el-divider content-position="left">扩容 (Scale Up)</el-divider>
            <div class="fields-grid">
              <el-form-item label="稳定窗口(秒)">
                <el-input-number v-model="form.scaleUpStabilizationSeconds" :min="0" :max="3600" style="width: 100%;" />
              </el-form-item>
              <el-form-item label="选择策略">
                <el-select v-model="form.scaleUpSelectPolicy" style="width: 100%;">
                  <el-option label="Max (最大值)" value="Max" />
                  <el-option label="Min (最小值)" value="Min" />
                  <el-option label="Disabled (禁用)" value="Disabled" />
                </el-select>
              </el-form-item>
            </div>
            <el-form-item label="扩容策略">
              <div style="width: 100%;">
                <div v-for="(p, i) in form.scaleUpPolicies" :key="i" class="kv-row">
                  <el-select v-model="p.type" style="width: 120px;">
                    <el-option label="Pods" value="Pods" />
                    <el-option label="Percent" value="Percent" />
                  </el-select>
                  <el-input-number v-model="p.value" :min="1" style="flex: 1;" />
                  <el-input-number v-model="p.periodSeconds" :min="1" placeholder="周期(秒)" style="width: 140px;" />
                  <el-button type="danger" text circle @click="form.scaleUpPolicies.splice(i, 1)">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
                <el-button text type="primary" @click="form.scaleUpPolicies.push({ type: 'Pods', value: 4, periodSeconds: 60 })" size="small">
                  <el-icon><Plus /></el-icon> 添加策略
                </el-button>
              </div>
            </el-form-item>

            <el-divider content-position="left">缩容 (Scale Down)</el-divider>
            <div class="fields-grid">
              <el-form-item label="稳定窗口(秒)">
                <el-input-number v-model="form.scaleDownStabilizationSeconds" :min="0" :max="3600" style="width: 100%;" />
              </el-form-item>
              <el-form-item label="选择策略">
                <el-select v-model="form.scaleDownSelectPolicy" style="width: 100%;">
                  <el-option label="Min (最小值)" value="Min" />
                  <el-option label="Max (最大值)" value="Max" />
                  <el-option label="Disabled (禁用)" value="Disabled" />
                </el-select>
              </el-form-item>
            </div>
            <el-form-item label="缩容策略">
              <div style="width: 100%;">
                <div v-for="(p, i) in form.scaleDownPolicies" :key="i" class="kv-row">
                  <el-select v-model="p.type" style="width: 120px;">
                    <el-option label="Pods" value="Pods" />
                    <el-option label="Percent" value="Percent" />
                  </el-select>
                  <el-input-number v-model="p.value" :min="1" style="flex: 1;" />
                  <el-input-number v-model="p.periodSeconds" :min="1" placeholder="周期(秒)" style="width: 140px;" />
                  <el-button type="danger" text circle @click="form.scaleDownPolicies.splice(i, 1)">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
                <el-button text type="primary" @click="form.scaleDownPolicies.push({ type: 'Percent', value: 100, periodSeconds: 60 })" size="small">
                  <el-icon><Plus /></el-icon> 添加策略
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
            <el-button type="primary" :loading="loading" @click="handleSubmit">{{ isEdit ? '更新' : '创建' }}</el-button>
          </div>
        </div>
      </div>
    </el-form>
  </div>
</template>

<style scoped>
.hpa-form {
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

.form-tip {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
}

.custom-metric-row {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 8px;
}

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
