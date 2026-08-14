<script setup lang="ts">
import { computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { deleteHpa } from '@/api/resource'

const props = defineProps<{
  hpa: any
}>()

const emit = defineEmits<{
  edit: []
  yaml: []
  deleted: []
}>()

// Status
const statusTagType = computed(() => {
  const conditions = props.hpa?.status?.conditions || []
  const scalingActive = conditions.find((c: any) => c.type === 'ScalingActive')
  if (scalingActive?.status === 'True') return 'success'
  return 'danger'
})

const statusText = computed(() => {
  const conditions = props.hpa?.status?.conditions || []
  const scalingActive = conditions.find((c: any) => c.type === 'ScalingActive')
  if (scalingActive?.status === 'True') return '正常'
  return '未激活'
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
    } else if (m.type === 'External') {
      name = m.external?.metric?.name || '-'
      displayName = name
      targetType = m.external?.target?.type || '-'
      targetValue = Number(m.external?.target?.averageValue ?? 0)
    }

    let color = '#67c23a'
    let statusLabel = '当前未触发扩容'
    if (currentValue !== null && targetValue > 0) {
      if (currentValue >= targetValue) {
        color = '#f56c6c'
        statusLabel = '⚠ 已触发扩容'
      } else if (currentValue >= targetValue * 0.8) {
        color = '#e6a23c'
        statusLabel = '接近阈值'
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
      `确定要删除 HPA "${name}" 吗？此操作不可恢复。`,
      '确认删除',
      { type: 'error', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await deleteHpa({ namespace: ns, name })
    ElMessage.success('HPA 已删除')
    emit('deleted')
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || '删除失败')
    }
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
        <el-button size="small" type="info" @click="emit('edit')">编辑</el-button>
        <el-button size="small" @click="emit('yaml')">YAML</el-button>
        <el-button size="small" type="danger" plain @click="handleDelete">删除</el-button>
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
    <div class="section" v-if="metricInfos.length">
      <div class="section-title">指标目标</div>
      <div v-for="(m, idx) in metricInfos" :key="idx" class="metric-item">
        <div class="metric-header">
          <span class="metric-name">{{ m.displayName }}</span>
          <span class="metric-values">
            目标: {{ m.targetValue }}<template v-if="m.targetType === 'Utilization'">%</template>
            &nbsp;当前: <template v-if="m.currentValue !== null">{{ m.currentValue }}<template v-if="m.targetType === 'Utilization'">%</template></template><template v-else>-</template>
          </span>
        </div>
        <el-progress
          :percentage="m.currentValue !== null ? Math.min(m.currentValue, 100) : 0"
          :color="m.color"
          :stroke-width="16"
          :show-text="false"
        />
        <div class="metric-status" v-if="m.currentValue !== null">
          <el-tag :type="m.color === '#f56c6c' ? 'danger' : m.color === '#e6a23c' ? 'warning' : 'success'" size="small" effect="plain">{{ m.statusLabel }}</el-tag>
        </div>
      </div>
    </div>

    <!-- Behavior -->
    <div class="section" v-if="hpa?.spec?.behavior">
      <div class="section-title">扩缩容行为</div>
      <div class="behavior-list">
        <div class="behavior-row" v-if="hpa.spec.behavior.scaleUp">
          <span class="behavior-label">扩容</span>
          <span class="behavior-value">
            稳定窗口 {{ hpa.spec.behavior.scaleUp.stabilizationWindowSeconds ?? 0 }}s
            <template v-if="(hpa.spec.behavior.scaleUp.stabilizationWindowSeconds ?? 0) === 0">(立即)</template>
            · 策略 {{ hpa.spec.behavior.scaleUp.selectPolicy || '-' }}
          </span>
        </div>
        <div class="behavior-row" v-if="hpa.spec.behavior.scaleDown">
          <span class="behavior-label">缩容</span>
          <span class="behavior-value">
            稳定窗口 {{ hpa.spec.behavior.scaleDown.stabilizationWindowSeconds ?? 300 }}s
            ({{ Math.round((hpa.spec.behavior.scaleDown.stabilizationWindowSeconds ?? 300) / 60) }}分钟)
            · 策略 {{ hpa.spec.behavior.scaleDown.selectPolicy || '-' }}
          </span>
        </div>
      </div>
    </div>

    <!-- Conditions -->
    <div class="section" v-if="hpa?.status?.conditions?.length">
      <div class="section-title">状态条件</div>
      <div class="conditions-list">
        <div v-for="(c, idx) in hpa.status.conditions" :key="idx" class="condition-item">
          <el-tag :type="c.status === 'True' ? 'success' : 'danger'" size="small" effect="dark" class="condition-status" />
          <span class="condition-type">{{ c.type }}</span>
          <span class="condition-reason">{{ c.reason }}</span>
        </div>
      </div>
    </div>

    <!-- Last Scale Time -->
    <div class="last-scale" v-if="lastScaleTime">
      Last Scale: {{ lastScaleTime }}
    </div>
  </div>
</template>

<style scoped>
.hpa-status-card {
  padding: 20px;
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
