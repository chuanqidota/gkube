<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getHpaDetail, deleteHpa, getHpaEvents } from '@/api/resource'
import { Refresh, Timer, ArrowLeft, FullScreen, Aim } from '@element-plus/icons-vue'
import YamlDrawer from '@/components/YamlDrawer.vue'
import HPAForm from './components/HPAForm.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const hpa = ref<any>(null)
const yamlDialogVisible = ref(false)

// Edit dialog
const editDialogVisible = ref(false)
const editFullscreen = ref(false)

// Events
const events = ref<any[]>([])
const eventsLoading = ref(false)

const namespace = route.params.namespace as string
const name = route.params.name as string

const statusTagType = computed(() => {
  const conditions = hpa.value?.status?.conditions || []
  const scalingReady = conditions.find((c: any) => c.type === 'ScalingActive')
  if (scalingReady?.status === 'True') return 'success'
  return 'danger'
})

const statusText = computed(() => {
  const conditions = hpa.value?.status?.conditions || []
  const scalingReady = conditions.find((c: any) => c.type === 'ScalingActive')
  if (scalingReady?.status === 'True') return '正常'
  return '未激活'
})

// Scale target link
const targetRoute = computed(() => {
  const kind = hpa.value?.spec?.scaleTargetRef?.kind
  const targetName = hpa.value?.spec?.scaleTargetRef?.name
  const ns = hpa.value?.metadata?.namespace || namespace
  if (!kind || !targetName || !ns) return null
  if (kind === 'Deployment') return `/workloads/deployments/${ns}/${targetName}`
  if (kind === 'StatefulSet') return `/workloads/statefulsets/${ns}/${targetName}`
  if (kind === 'DaemonSet') return `/workloads/daemonsets/${ns}/${targetName}`
  if (kind === 'ReplicaSet') return `/workloads/replicasets/${ns}/${targetName}`
  return null
})

// Metrics helpers
interface MetricInfo {
  name: string
  targetType: string
  targetValue: number
  currentValue: number | null
  color: 'success' | 'warning' | 'exception'
  statusText: string
}

const metricInfos = computed<MetricInfo[]>(() => {
  const spec = hpa.value?.spec
  const status = hpa.value?.status
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
    let targetType = '-'
    let targetValue = 0
    let currentValue: number | null = null

    if (m.type === 'Resource') {
      name = m.resource?.name || '-'
      targetType = m.resource?.target?.type || '-'
      targetValue = Number(m.resource?.target?.averageUtilization ?? 0)
      currentValue = currentMap[name] ?? null
    } else if (m.type === 'Pods') {
      name = m.pods?.metric?.name || '-'
      targetType = m.pods?.target?.type || '-'
      targetValue = Number(m.pods?.target?.averageValue ?? 0)
      currentValue = currentMap[name] ?? null
    } else if (m.type === 'Object') {
      name = m.object?.metric?.name || '-'
      targetType = m.object?.target?.type || '-'
      targetValue = Number(m.object?.target?.value ?? 0)
    } else if (m.type === 'External') {
      name = m.external?.metric?.name || '-'
      targetType = m.external?.target?.type || '-'
      targetValue = Number(m.external?.target?.averageValue ?? 0)
      currentValue = currentMap[name] ?? null
    }

    let color: 'success' | 'warning' | 'exception' = 'success'
    let statusLabel = '当前未触发扩容'
    if (currentValue !== null && targetValue > 0) {
      if (currentValue >= targetValue) {
        color = 'exception'
        statusLabel = '⚠ 已触发扩容'
      } else if (currentValue >= targetValue * 0.8) {
        color = 'warning'
        statusLabel = '接近阈值'
      }
    }

    return { name, targetType, targetValue, currentValue, color, statusText: statusLabel }
  })
})

// Events filtered for scaling events
const scalingEvents = computed(() => {
  return (events.value || []).slice(0, 10)
})

// Behavior rows for the table
const behaviorRows = computed(() => {
  const behavior = hpa.value?.spec?.behavior
  if (!behavior) return []
  const rows: any[] = []
  if (behavior.scaleUp) {
    const w = behavior.scaleUp.stabilizationWindowSeconds ?? 0
    rows.push({
      direction: '扩容',
      window: w,
      windowLabel: w === 0 ? '立即' : `${Math.round(w / 60)}分钟`,
      selectPolicy: behavior.scaleUp.selectPolicy || '-',
      policies: behavior.scaleUp.policies || [],
    })
  }
  if (behavior.scaleDown) {
    const w = behavior.scaleDown.stabilizationWindowSeconds ?? 300
    rows.push({
      direction: '缩容',
      window: w,
      windowLabel: w === 0 ? '立即' : `${Math.round(w / 60)}分钟`,
      selectPolicy: behavior.scaleDown.selectPolicy || '-',
      policies: behavior.scaleDown.policies || [],
    })
  }
  return rows
})

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getHpaDetail({ namespace, name })
    hpa.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || '加载 HPA 详情失败')
  } finally {
    loading.value = false
  }
}

async function fetchEvents() {
  eventsLoading.value = true
  try {
    const res: any = await getHpaEvents({ namespace, name })
    events.value = res.data || []
  } catch {
    // Events are optional — don't block the page
  } finally {
    eventsLoading.value = false
  }
}

function handleOpenYaml() {
  yamlDialogVisible.value = true
}

function handleYamlSaved() {
  fetchDetail()
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(
      `确定要删除 HPA "${name}" 吗？此操作不可恢复。`,
      '确认删除',
      { type: 'error', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await deleteHpa({ namespace, name })
    ElMessage.success('HPA 已删除')
    router.push('/autoscaling/hpa')
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || '删除失败')
    }
  }
}

function handleEdit() {
  editDialogVisible.value = true
}

function handleEditSuccess() {
  editDialogVisible.value = false
  fetchDetail()
}

function handleEditCancel() {
  editDialogVisible.value = false
}

// ---- Resize: left-right ----
const leftWidth = ref(300)
const resizingH = ref(false)
let startX = 0, startW = 0
function onHResizeStart(e: MouseEvent) {
  e.preventDefault()
  resizingH.value = true
  startX = e.clientX
  startW = leftWidth.value
  const onMove = (ev: MouseEvent) => {
    leftWidth.value = Math.min(Math.max(startW + ev.clientX - startX, 220), 500)
  }
  const onUp = () => {
    resizingH.value = false
    document.removeEventListener('mousemove', onMove)
    document.removeEventListener('mouseup', onUp)
  }
  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onUp)
}

// ---- Resize: top-bottom (Metrics / Conditions) ----
const rightTopHeight = ref<number | null>(null)
const resizingV = ref(false)
let startY = 0, startH = 0
function onVResizeStart(e: MouseEvent) {
  e.preventDefault()
  const rightPanel = (e.target as HTMLElement).closest('.right-panel')
  if (!rightPanel) return
  resizingV.value = true
  startY = e.clientY
  startH = rightPanel.getBoundingClientRect().height
  const onMove = (ev: MouseEvent) => {
    const delta = ev.clientY - startY
    rightTopHeight.value = Math.min(Math.max(startH * 0.3 + delta, 120), startH - 120)
  }
  const onUp = () => {
    resizingV.value = false
    document.removeEventListener('mousemove', onMove)
    document.removeEventListener('mouseup', onUp)
  }
  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onUp)
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(async () => {
  await fetchDetail()
  await fetchEvents()
}, { autoStart: false })

onMounted(() => {
  fetchDetail()
  fetchEvents()
})
</script>

<template>
  <div class="detail-page" v-loading="loading">

    <!-- 顶部标题栏 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="res-name">{{ name }}</h2>
        <div class="meta-line">
          <el-tag :type="statusTagType" effect="dark" size="small">{{ statusText }}</el-tag>
          <span class="ns-tag">ns/{{ namespace }}</span>
          <el-button
            v-if="hpa?.spec?.scaleTargetRef && targetRoute"
            link
            type="primary"
            class="target-link"
            @click="$router.push(targetRoute!)"
          >{{ hpa.spec.scaleTargetRef.kind }}/{{ hpa.spec.scaleTargetRef.name }}</el-button>
          <span v-else-if="hpa?.spec?.scaleTargetRef" class="replicas-info">
            {{ hpa.spec.scaleTargetRef.kind }}/{{ hpa.spec.scaleTargetRef.name }}
          </span>
        </div>
      </div>
      <div class="header-actions">
        <el-button type="info" @click="handleEdit">编辑</el-button>
        <el-button @click="handleOpenYaml">YAML</el-button>
        <el-button type="danger" plain @click="handleDelete">删除</el-button>
        <div class="action-divider" />
        <el-popover placement="bottom" :width="200" trigger="click">
          <template #reference>
            <el-button
              :type="isRunning ? 'success' : 'default'"
              :icon="Timer"
              @click="toggle()"
            />
          </template>
          <div class="auto-refresh-popover">
            <div class="popover-title">
              {{ isRunning ? `自动刷新中 ${countdown}s` : '自动刷新' }}
            </div>
            <el-select
              :model-value="currentInterval / 1000"
              @update:model-value="setIntervalOption"
              :teleported="false"
              size="small"
              style="width: 100%;"
            >
              <el-option
                v-for="sec in availableIntervals"
                :key="sec"
                :value="sec"
                :label="`每 ${sec} 秒刷新`"
              />
            </el-select>
          </div>
        </el-popover>
        <el-tooltip content="刷新" placement="top">
          <el-button @click="manualRefresh()" :loading="loading" :icon="Refresh" />
        </el-tooltip>
        <el-tooltip content="返回列表" placement="top">
          <el-button :icon="ArrowLeft" @click="router.push('/autoscaling/hpa')" />
        </el-tooltip>
      </div>
    </div>

    <template v-if="hpa">
      <div class="main-layout" :class="{ 'is-resizing': resizingH || resizingV }">

        <!-- 左侧：基本信息 -->
        <div class="left-panel" :style="{ width: leftWidth + 'px', minWidth: leftWidth + 'px' }">
          <div class="panel-title">基本信息</div>
          <div class="info-body">
            <el-descriptions :column="1" border size="small">
              <el-descriptions-item label="名称">{{ hpa.metadata?.name || hpa.name }}</el-descriptions-item>
              <el-descriptions-item label="命名空间">{{ hpa.metadata?.namespace || hpa.namespace }}</el-descriptions-item>
              <el-descriptions-item label="伸缩目标">
                <el-button
                  v-if="targetRoute"
                  link
                  type="primary"
                  @click="$router.push(targetRoute!)"
                >{{ hpa.spec?.scaleTargetRef?.kind }}/{{ hpa.spec?.scaleTargetRef?.name }}</el-button>
                <span v-else>{{ hpa.spec?.scaleTargetRef?.kind }}/{{ hpa.spec?.scaleTargetRef?.name }}</span>
              </el-descriptions-item>
              <el-descriptions-item label="最小副本数">{{ hpa.spec?.minReplicas ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="最大副本数">{{ hpa.spec?.maxReplicas ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="当前副本数">{{ hpa.status?.currentReplicas ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="期望副本数">{{ hpa.status?.desiredReplicas ?? '-' }}</el-descriptions-item>
            </el-descriptions>

            <!-- Labels -->
            <div v-if="hpa.metadata?.labels && Object.keys(hpa.metadata.labels).length > 0" style="margin-top: 16px;">
              <h4 style="margin: 0 0 8px; font-size: 13px;">Labels</h4>
              <el-tag
                v-for="(val, key) in hpa.metadata.labels"
                :key="key"
                style="margin-right: 8px; margin-bottom: 8px;"
                size="small"
              >
                {{ key }}={{ val }}
              </el-tag>
            </div>
          </div>
        </div>

        <!-- 右侧：Metrics + Behavior + Conditions + Events -->
        <div class="right-panel">

          <!-- Metrics Progress Bars -->
          <div class="right-section metrics-section" :style="rightTopHeight ? { flex: 'none', height: rightTopHeight + 'px' } : {}">
            <div class="panel-title">
              指标目标
              <span class="count-badge">{{ metricInfos.length }} 个</span>
            </div>
            <div class="metrics-body">
              <div v-if="metricInfos.length" class="metrics-grid">
                <div v-for="(m, idx) in metricInfos" :key="idx" class="metric-card">
                  <div class="metric-header">
                    <span class="metric-name">{{ m.name === 'cpu' ? 'CPU 使用率' : m.name === 'memory' ? 'Memory 使用率' : m.name }}</span>
                    <span class="metric-values">
                      目标: {{ m.targetValue }}<template v-if="m.targetType === 'Utilization'">%</template>
                      &nbsp;&nbsp;
                      当前: <template v-if="m.currentValue !== null">{{ m.currentValue }}<template v-if="m.targetType === 'Utilization'">%</template></template><template v-else>-</template>
                    </span>
                  </div>
                  <el-progress
                    :percentage="m.currentValue !== null ? Math.min(m.currentValue, 100) : 0"
                    :color="m.color === 'success' ? '#67c23a' : m.color === 'warning' ? '#e6a23c' : '#f56c6c'"
                    :stroke-width="18"
                    :show-text="false"
                    :status="m.currentValue === null ? undefined : undefined"
                  />
                  <div class="metric-markers">
                    <span class="marker marker-target" :style="{ left: Math.min(m.targetValue, 100) + '%' }">↑目标{{ m.targetValue }}%</span>
                    <span v-if="m.currentValue !== null" class="marker marker-current" :style="{ left: Math.min(m.currentValue, 100) + '%' }">↑当前{{ m.currentValue }}%</span>
                  </div>
                  <div class="metric-status" v-if="m.currentValue !== null">
                    <el-tag :type="m.color" size="small" effect="plain">{{ m.statusText }}</el-tag>
                  </div>
                  <div class="metric-status" v-else>
                    <el-tag type="info" size="small" effect="plain">无当前数据</el-tag>
                  </div>
                </div>
              </div>
              <div v-else class="empty-hint">暂无指标配置</div>
            </div>
          </div>

          <!-- 垂直拖拽条 -->
          <div class="resize-handle-v" :class="{ active: resizingV }" @mousedown="onVResizeStart" />

          <!-- Behavior -->
          <div class="right-section behavior-section" v-if="hpa.spec?.behavior">
            <div class="panel-title">扩缩容行为</div>
            <div class="behavior-body">
              <el-table :data="behaviorRows" size="small" stripe>
                <el-table-column prop="direction" label="方向" width="80" />
                <el-table-column label="稳定窗口" width="200">
                  <template #default="{ row }">
                    {{ row.window }}s
                    <span v-if="row.windowLabel" class="text-hint"> ({{ row.windowLabel }})</span>
                  </template>
                </el-table-column>
                <el-table-column prop="selectPolicy" label="选择策略" width="140" />
                <el-table-column label="策略" min-width="250">
                  <template #default="{ row }">
                    <span v-if="row.policies.length === 0" class="text-hint">默认</span>
                    <span v-else>{{ row.policies.map((p: any) => `${p.type} ${p.value}/${p.periodSeconds}s`).join(', ') }}</span>
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </div>

          <!-- Conditions -->
          <div class="right-section conditions-section">
            <div class="panel-title">
              状态条件
              <span class="count-badge">{{ hpa.status?.conditions?.length || 0 }} 条</span>
            </div>
            <div class="conditions-body">
              <el-table v-if="hpa.status?.conditions?.length" :data="hpa.status.conditions" size="small" stripe max-height="260">
                <el-table-column prop="type" label="类型" width="180" />
                <el-table-column label="状态" width="100">
                  <template #default="{ row }">
                    <el-tag :type="row.status === 'True' ? 'success' : 'danger'" size="small">{{ row.status }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="reason" label="原因" width="180" />
                <el-table-column prop="message" label="信息" min-width="250" show-overflow-tooltip />
              </el-table>
              <div v-else class="empty-hint">暂无状态条件</div>
            </div>
          </div>

          <!-- Events Timeline -->
          <div class="right-section events-section">
            <div class="panel-title">
              伸缩事件
              <span class="count-badge">{{ scalingEvents.length }} 条</span>
            </div>
            <div class="events-body" v-loading="eventsLoading">
              <el-table v-if="scalingEvents.length" :data="scalingEvents" size="small" stripe max-height="260">
                <el-table-column label="时间" width="180">
                  <template #default="{ row }">
                    {{ row.lastTimestamp || row.firstTimestamp || '-' }}
                  </template>
                </el-table-column>
                <el-table-column label="类型" width="100">
                  <template #default="{ row }">
                    <el-tag :type="row.type === 'Normal' ? 'success' : 'warning'" size="small">{{ row.type || '-' }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="reason" label="原因" width="180" />
                <el-table-column prop="message" label="消息" min-width="300" show-overflow-tooltip />
              </el-table>
              <div v-else class="empty-hint">{{ eventsLoading ? '加载中...' : '暂无伸缩事件' }}</div>
            </div>
          </div>

        </div>

        <!-- 水平拖拽条 -->
        <div
          class="resize-handle-h"
          :class="{ active: resizingH }"
          :style="{ left: (leftWidth - 3) + 'px' }"
          @mousedown="onHResizeStart"
        />
      </div>
    </template>

    <!-- YAML Drawer -->
    <YamlDrawer
      v-model="yamlDialogVisible"
      resource-type="hpa"
      :namespace="namespace"
      :name="name"
      @saved="handleYamlSaved"
    />

    <!-- Edit Drawer -->
    <el-drawer
      v-model="editDialogVisible"
      title="编辑 HPA"
      :size="editFullscreen ? '100%' : '85%'"
      direction="rtl"
      :destroy-on-close="true"
      :body-style="{ padding: '0', height: '100%' }"
    >
      <template #header>
        <div class="drawer-header">
          <span class="drawer-title">编辑 HPA</span>
          <el-tooltip :content="editFullscreen ? '退出全屏' : '全屏'" placement="top">
            <el-icon class="fullscreen-btn" @click="editFullscreen = !editFullscreen">
              <FullScreen v-if="!editFullscreen" />
              <Aim v-else />
            </el-icon>
          </el-tooltip>
        </div>
      </template>
      <div style="height: calc(100vh - 52px); overflow-y: auto;">
        <HPAForm
          v-if="editDialogVisible && hpa"
          :is-edit="true"
          :initial-data="hpa"
          @success="handleEditSuccess"
          @cancel="handleEditCancel"
        />
      </div>
    </el-drawer>
  </div>
</template>

<style scoped>
.detail-page {
  padding: 16px 20px;
  height: 100vh;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
}

/* Header */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  flex-shrink: 0;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.res-name {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  line-height: 1.3;
}

.meta-line {
  display: flex;
  align-items: center;
  gap: 8px;
}

.ns-tag {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color-lighter);
  padding: 1px 6px;
  border-radius: 4px;
}

.replicas-info {
  font-size: 12px;
  color: var(--el-text-color-regular);
}

.target-link {
  font-size: 12px;
}

.header-actions {
  display: flex;
  flex-shrink: 0;
  align-items: center;
}

.header-actions .el-button {
  border-radius: 0;
  margin-left: -1px;
}

.header-actions .el-button:first-child {
  border-radius: 4px 0 0 4px;
  margin-left: 0;
}

.header-actions .el-button:last-of-type {
  border-radius: 0 4px 4px 0;
}

.action-divider {
  width: 1px;
  height: 20px;
  background: var(--el-border-color-lighter);
  margin: 0 4px;
}

.auto-refresh-popover {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.popover-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

/* Main Layout */
.main-layout {
  display: flex;
  gap: 2px;
  flex: 1;
  min-height: 0;
  overflow: hidden;
  position: relative;
}

/* Left Panel */
.left-panel {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--el-bg-color);
}

.panel-title {
  font-size: 13px;
  font-weight: 600;
  padding: 10px 14px;
  background: var(--el-fill-color-lighter);
  border-bottom: 1px solid var(--el-border-color-lighter);
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.count-badge {
  font-weight: 400;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.info-body {
  flex: 1;
  overflow-y: auto;
  padding: 14px;
}

/* Right Panel */
.right-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
  overflow: hidden;
}

.right-section {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--el-bg-color);
}

.metrics-section {
  flex: 1;
  min-height: 0;
}

.right-section.behavior-section {
  flex: 0 0 auto;
  max-height: 300px;
}

.right-section.conditions-section {
  flex: 0 0 auto;
  max-height: 300px;
}

.right-section.events-section {
  flex: 1;
  min-height: 0;
}

/* Metrics */
.metrics-body {
  flex: 1;
  overflow-y: auto;
  padding: 14px;
}

.metrics-grid {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.metric-card {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  padding: 12px 14px;
}

.metric-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.metric-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.metric-values {
  font-size: 12px;
  color: var(--el-text-color-regular);
}

.metric-markers {
  position: relative;
  height: 28px;
  margin-top: 2px;
}

.marker {
  position: absolute;
  font-size: 10px;
  transform: translateX(-50%);
  white-space: nowrap;
}

.marker-target {
  top: 0;
  color: var(--el-text-color-secondary);
}

.marker-current {
  top: 14px;
  color: var(--el-color-primary);
}

.metric-status {
  margin-top: 4px;
}

/* Behavior */
.behavior-body {
  flex: 1;
  overflow-y: auto;
  padding: 0;
}

.text-hint {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

/* Conditions */
.conditions-body {
  flex: 1;
  overflow-y: auto;
  padding: 0;
}

/* Events */
.events-body {
  flex: 1;
  overflow-y: auto;
  padding: 0;
}

/* Resize handles */
.resize-handle-h {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 8px;
  cursor: col-resize;
  z-index: 10;
}

.resize-handle-h:hover,
.resize-handle-h.active {
  background: var(--el-color-primary-light-7);
}

.resize-handle-v {
  height: 4px;
  cursor: row-resize;
  flex-shrink: 0;
  position: relative;
  z-index: 5;
  margin: -2px 0;
}

.resize-handle-v:hover,
.resize-handle-v.active {
  background: var(--el-color-primary-light-7);
}

.is-resizing {
  user-select: none;
}

.is-resizing * {
  pointer-events: none;
}

.empty-hint {
  padding: 24px;
  text-align: center;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

/* Responsive */
@media (max-width: 768px) {
  .main-layout {
    flex-direction: column;
    overflow: auto;
  }
  .left-panel {
    width: 100% !important;
    min-width: 100% !important;
    max-height: 300px;
  }
  .resize-handle-h {
    display: none;
  }
}

/* Edit Drawer */
.drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.drawer-title {
  font-size: 16px;
  font-weight: 600;
}

.fullscreen-btn {
  cursor: pointer;
  font-size: 18px;
  color: var(--el-text-color-regular);
  transition: color 0.2s;
}

.fullscreen-btn:hover {
  color: var(--el-color-primary);
}
</style>
