<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { deleteHpa, pauseHpa, resumeHpa } from '@/api/resource'

const props = defineProps<{
  hpa: any
}>()

const emit = defineEmits<{
  edit: []
  yaml: []
  deleted: []
  refreshed: []
}>()

const { t } = useI18n()

// Status
const isPaused = computed(() => {
  return props.hpa?.metadata?.annotations?.['gkube.io/paused'] === 'true'
})

const statusTagType = computed(() => {
  if (isPaused.value) return 'warning'
  const conditions = props.hpa?.status?.conditions || []
  const scalingActive = conditions.find((c: any) => c.type === 'ScalingActive')
  if (scalingActive?.status === 'True') return 'success'
  return 'danger'
})

const statusText = computed(() => {
  if (isPaused.value) return t('workload.suspend')
  const conditions = props.hpa?.status?.conditions || []
  const scalingActive = conditions.find((c: any) => c.type === 'ScalingActive')
  if (scalingActive?.status === 'True') return t('workload.active')
  return 'Inactive'
})

// Metrics info
interface MetricInfo {
  name: string
  displayName: string
  targetType: string
  targetValue: number
  currentValue: number | null
  color: string
  statusLabel: string
}

const metricInfos = computed<MetricInfo[]>(() => {
  const spec = props.hpa?.spec
  const status = props.hpa?.status
  if (!spec?.metrics) return []

  const currentMap: Record<string, number> = {}
  if (status?.currentMetrics) {
    for (const cm of status.currentMetrics) {
      if (cm.type === 'Resource') {
        const name = cm.resource?.name
        const val = cm.resource?.current?.averageUtilization
        if (name && val !== undefined) currentMap[name] = Number(val)
      } else if (cm.type === 'Pods') {
        const name = cm.pods?.metric?.name
        const val = cm.pods?.current?.averageValue
        if (name && val !== undefined) {
          const numVal = parseFloat(String(val))
          if (!isNaN(numVal)) currentMap[name] = numVal
        }
      } else if (cm.type === 'External') {
        const name = cm.external?.metric?.name
        const val = cm.external?.current?.averageValue
        if (name && val !== undefined) {
          const numVal = parseFloat(String(val))
          if (!isNaN(numVal)) currentMap[name] = numVal
        }
      }
    }
  }

  return spec.metrics.map((m: any) => {
    let name = '-'
    let displayName = '-'
    let targetType = '-'
    let targetValue = 0
    let currentValue: number | null = null

    if (m.type === 'Resource') {
      name = m.resource?.name || '-'
      displayName = name === 'cpu' ? 'CPU 使用率' : name === 'memory' ? 'Memory 使用率' : name
      targetType = m.resource?.target?.type || '-'
      targetValue = Number(m.resource?.target?.averageUtilization ?? 0)
      currentValue = currentMap[name] ?? null
    } else if (m.type === 'Pods') {
      name = m.pods?.metric?.name || '-'
      displayName = name
      targetType = m.pods?.target?.type || '-'
      targetValue = Number(m.pods?.target?.averageValue ?? 0)
    } else if (m.type === 'Object') {
      name = m.object?.metric?.name || '-'
      displayName = name
      targetType = m.object?.target?.type || '-'
      targetValue = Number(m.object?.target?.value ?? 0)
    } else if (m.type === 'External') {
      name = m.external?.metric?.name || '-'
      displayName = name
      targetType = m.external?.target?.type || '-'
      targetValue = Number(m.external?.target?.averageValue ?? 0)
    }

    let color = '#67c23a'
    let statusLabel = t('workload.unscheduledExpand')
    if (currentValue !== null && targetValue > 0) {
      if (currentValue >= targetValue) {
        color = '#f56c6c'
        statusLabel = t('workload.triggeredExpand')
      } else if (currentValue >= targetValue * 0.8) {
        color = '#e6a23c'
        statusLabel = t('workload.nearThreshold')
      }
    }

    return { name, displayName, targetType, targetValue, currentValue, color, statusLabel }
  })
})

// Last scale time
const lastScaleTime = computed(() => {
  return props.hpa?.status?.lastScaleTime || null
})

async function handleDelete() {
  const ns = props.hpa?.metadata?.namespace || ''
  const name = props.hpa?.metadata?.name || ''
  try {
    await ElMessageBox.confirm(
      t('workload.confirmDeleteHpa', { name }),
      t('common.confirmDelete'),
      {
        type: 'error',
        confirmButtonText: t('common.delete'),
        cancelButtonText: t('common.cancel'),
      },
    )
    await deleteHpa({ namespace: ns, name })
    ElMessage.success(t('workload.hpaDeleteSuccess'))
    emit('deleted')
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || t('common.deleteFailed'))
    }
  }
}

async function handlePause() {
  const ns = props.hpa?.metadata?.namespace || ''
  const name = props.hpa?.metadata?.name || ''
  const current = props.hpa?.status?.currentReplicas
  try {
    await ElMessageBox.confirm(
      t('workload.pauseConfirm', { n: current ?? '-' }),
      t('common.confirmAction'),
      {
        type: 'warning',
        confirmButtonText: t('workload.suspend'),
        cancelButtonText: t('common.cancel'),
      },
    )
    await pauseHpa({ namespace: ns, name })
    ElMessage.success(t('workload.hpaPauseSuccess'))
    emit('refreshed')
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || t('workload.pauseFailed'))
    }
  }
}

async function handleResume() {
  const ns = props.hpa?.metadata?.namespace || ''
  const name = props.hpa?.metadata?.name || ''
  try {
    await resumeHpa({ namespace: ns, name })
    ElMessage.success(t('workload.hpaResumeSuccess'))
    emit('refreshed')
  } catch (e: any) {
    ElMessage.error(e?.message || t('workload.resumeFailed'))
  }
}
</script>

<template>
  <div class="hpa-status-card">
    <!-- Header -->
    <div class="card-header">
      <div class="header-info">
        <h3 class="hpa-name">{{ hpa?.metadata?.name }}</h3>
        <el-tag :type="statusTagType" effect="dark" size="small">{{ statusText }}</el-tag>
      </div>
      <div class="header-actions">
        <el-button size="small" type="info" @click="emit('edit')">{{ t('common.edit') }}</el-button>
        <el-button v-if="isPaused" size="small" type="success" @click="handleResume"
          >恢复</el-button
        >
        <el-button v-else size="small" type="warning" @click="handlePause">暂停</el-button>
        <el-button size="small" @click="emit('yaml')">YAML</el-button>
        <el-button size="small" type="danger" @click="handleDelete">{{
          t('common.delete')
        }}</el-button>
      </div>
    </div>

    <!-- Basic Info -->
    <div class="info-bar">
      <span>Min {{ hpa?.spec?.minReplicas ?? '-' }}</span>
      <span class="sep">│</span>
      <span>Max {{ hpa?.spec?.maxReplicas ?? '-' }}</span>
      <span class="sep">│</span>
      <span>Current {{ hpa?.status?.currentReplicas ?? '-' }}</span>
      <span class="sep">│</span>
      <span>Desired {{ hpa?.status?.desiredReplicas ?? '-' }}</span>
    </div>

    <!-- Metrics Progress Bars -->
    <div v-if="metricInfos.length" class="section">
      <div class="section-title">指标目标</div>
      <div v-for="(m, idx) in metricInfos" :key="idx" class="metric-item">
        <div class="metric-header">
          <span class="metric-name">{{ m.displayName }}</span>
          <span class="metric-values">
            目标: {{ m.targetValue
            }}<template v-if="m.targetType === 'Utilization'">%</template> &nbsp;当前:
            <template v-if="m.currentValue !== null"
              >{{ m.currentValue
              }}<template v-if="m.targetType === 'Utilization'">%</template></template
            ><template v-else>-</template>
          </span>
        </div>
        <el-progress
          :percentage="m.currentValue !== null ? Math.min(m.currentValue, 100) : 0"
          :color="m.color"
          :stroke-width="16"
          :show-text="false"
        />
        <div v-if="m.currentValue !== null" class="metric-status">
          <el-tag
            :type="m.color === '#f56c6c' ? 'danger' : m.color === '#e6a23c' ? 'warning' : 'success'"
            size="small"
            effect="plain"
            >{{ m.statusLabel }}</el-tag
          >
        </div>
      </div>
    </div>

    <!-- Behavior -->
    <div v-if="hpa?.spec?.behavior" class="section">
      <div class="section-title">扩缩容行为</div>
      <div class="behavior-list">
        <div v-if="hpa.spec.behavior.scaleUp" class="behavior-row">
          <span class="behavior-label">扩容</span>
          <span class="behavior-value">
            稳定窗口 {{ hpa.spec.behavior.scaleUp.stabilizationWindowSeconds ?? 0 }}s
            <template v-if="(hpa.spec.behavior.scaleUp.stabilizationWindowSeconds ?? 0) === 0"
              >(立即)</template
            >
            · 策略 {{ hpa.spec.behavior.scaleUp.selectPolicy || '-' }}
          </span>
        </div>
        <div v-if="hpa.spec.behavior.scaleDown" class="behavior-row">
          <span class="behavior-label">缩容</span>
          <span class="behavior-value">
            稳定窗口 {{ hpa.spec.behavior.scaleDown.stabilizationWindowSeconds ?? 300 }}s ({{
              Math.round((hpa.spec.behavior.scaleDown.stabilizationWindowSeconds ?? 300) / 60)
            }}分钟) · 策略 {{ hpa.spec.behavior.scaleDown.selectPolicy || '-' }}
          </span>
        </div>
      </div>
    </div>

    <!-- Conditions -->
    <div v-if="hpa?.status?.conditions?.length" class="section">
      <div class="section-title">状态条件</div>
      <div class="conditions-list">
        <div v-for="(c, idx) in hpa.status.conditions" :key="idx" class="condition-item">
          <el-tag
            :type="c.status === 'True' ? 'success' : 'danger'"
            size="small"
            effect="dark"
            class="condition-status"
          />
          <span class="condition-type">{{ c.type }}</span>
          <span class="condition-reason">{{ c.reason }}</span>
        </div>
      </div>
    </div>

    <!-- Last Scale Time -->
    <div v-if="lastScaleTime" class="last-scale">Last Scale: {{ lastScaleTime }}</div>
  </div>
</template>

<style scoped>
.hpa-status-card {
  padding: var(--gk-space-5);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.header-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.hpa-name {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  gap: 4px;
}

.info-bar {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: var(--el-text-color-regular);
  margin-bottom: 16px;
  padding: 8px 12px;
  background: var(--el-fill-color-lighter);
  border-radius: 6px;
}

.sep {
  color: var(--el-border-color);
  margin: 0 4px;
}

.section {
  margin-bottom: 16px;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin-bottom: 8px;
  padding-bottom: 6px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.metric-item {
  padding: 10px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  margin-bottom: 8px;
}

.metric-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.metric-name {
  font-size: 13px;
  font-weight: 600;
}

.metric-values {
  font-size: 12px;
  color: var(--el-text-color-regular);
}

.metric-status {
  margin-top: 4px;
}

.behavior-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.behavior-row {
  display: flex;
  align-items: baseline;
  gap: 8px;
  font-size: 13px;
}

.behavior-label {
  font-weight: 600;
  color: var(--el-text-color-primary);
  flex-shrink: 0;
  width: 36px;
}

.behavior-value {
  color: var(--el-text-color-regular);
}

.conditions-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.condition-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}

.condition-status {
  width: 8px;
  height: 8px;
  padding: 0;
  min-width: 8px;
  border-radius: 50%;
}

.condition-type {
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.condition-reason {
  color: var(--el-text-color-secondary);
}

.last-scale {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 8px;
}
</style>
