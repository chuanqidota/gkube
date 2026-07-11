<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getStatefulSetDetail,
  deleteStatefulSet,
  scaleStatefulSet,
  restartStatefulSet,
  getStatefulSetEvents,
  getStatefulSetPods,
  deletePod,
} from '@/api/resource'
import { Refresh, Timer, ArrowLeft, FullScreen, Aim } from '@element-plus/icons-vue'
import YamlDrawer from '@/components/YamlDrawer.vue'
import PodListPanel from '@/components/PodListPanel.vue'
import WorkloadForm from '@/views/workload/components/WorkloadForm.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const statefulSet = ref<any>(null)
const yamlDialogVisible = ref(false)

// Events
const events = ref<any[]>([])
const eventsLoading = ref(false)

// Pods
const pods = ref<any[]>([])
const podsLoading = ref(false)

// Scale dialog
const scaleDialogVisible = ref(false)
const scaleReplicas = ref<number>(1)
const scaleLoading = ref(false)

// Edit dialog
const editDialogVisible = ref(false)
const editFullscreen = ref(false)

const namespace = route.params.namespace as string
const name = route.params.name as string

const statusTagType = computed(() => {
  const ready = statefulSet.value?.status?.readyReplicas || 0
  const desired = statefulSet.value?.spec?.replicas || 0
  if (ready === desired && desired > 0) return 'success'
  if (ready > 0) return 'warning'
  return 'danger'
})

const statusText = computed(() => {
  const ready = statefulSet.value?.status?.readyReplicas || 0
  const desired = statefulSet.value?.spec?.replicas || 0
  if (ready === desired && desired > 0) return 'Ready'
  if (ready > 0) return 'Partial'
  return 'Not Ready'
})

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getStatefulSetDetail({ namespace, name })
    statefulSet.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || '加载 StatefulSet 详情失败')
  } finally {
    loading.value = false
  }
}

async function fetchEvents() {
  eventsLoading.value = true
  try {
    const res: any = await getStatefulSetEvents({ namespace, name })
    events.value = res.data || []
  } catch (e: any) {
    events.value = []
  } finally {
    eventsLoading.value = false
  }
}

async function fetchPods() {
  podsLoading.value = true
  try {
    const res: any = await getStatefulSetPods({ namespace, name })
    pods.value = res.data?.items || res.data || []
  } catch (e: any) {
    pods.value = []
  } finally {
    podsLoading.value = false
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
      `确定要删除 StatefulSet "${name}" 吗？此操作不可恢复。`,
      '确认删除',
      { type: 'error', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await deleteStatefulSet({ namespace, name })
    ElMessage.success('StatefulSet 已删除')
    router.push('/workloads/statefulsets')
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || '删除失败')
    }
  }
}

async function handleRestart() {
  try {
    await ElMessageBox.confirm(
      `确定要重启 StatefulSet "${name}" 吗？这将触发滚动更新。`,
      '确认重启',
      { type: 'warning' }
    )
    await restartStatefulSet({ namespace, name })
    ElMessage.success('StatefulSet 已重启')
    fetchDetail()
    fetchPods()
  } catch {
    // cancelled
  }
}

function handleScale() {
  scaleReplicas.value = statefulSet.value?.spec?.replicas ?? 1
  scaleDialogVisible.value = true
}

async function handleScaleConfirm() {
  scaleLoading.value = true
  try {
    await scaleStatefulSet({ namespace, name, replicas: scaleReplicas.value })
    ElMessage.success(`StatefulSet 已扩缩容至 ${scaleReplicas.value} 个副本`)
    scaleDialogVisible.value = false
    fetchDetail()
    // Poll for pod list update
    const expectedPods = scaleReplicas.value
    for (let i = 0; i < 10; i++) {
      await new Promise(r => setTimeout(r, 1000))
      await fetchPods()
      if (pods.value.length === expectedPods) break
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '扩缩容失败')
  } finally {
    scaleLoading.value = false
  }
}

function handleEdit() {
  editDialogVisible.value = true
}

function handleEditSuccess() {
  editDialogVisible.value = false
  fetchDetail()
  fetchPods()
}

function handleEditCancel() {
  editDialogVisible.value = false
}

function getClusterName(): string {
  try {
    const saved = localStorage.getItem('gkube_cluster')
    if (saved) {
      const c = JSON.parse(saved)
      return c?.clusterName || c?.cluster_name || c?.name || ''
    }
  } catch { /* ignore */ }
  return ''
}

function handlePodLogs(pod: any) {
  const cluster = getClusterName()
  window.open(`/fullscreen/logs?namespace=${pod.metadata?.namespace || namespace}&pod=${pod.metadata?.name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
}

function handlePodExec(pod: any) {
  const cluster = getClusterName()
  window.open(`/fullscreen/terminal?namespace=${pod.metadata?.namespace || namespace}&pod=${pod.metadata?.name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
}

async function handleDeletePod(pod: any) {
  try {
    await ElMessageBox.confirm(
      `确定要删除 Pod "${pod.metadata?.name}" 吗？`,
      '确认删除',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await deletePod({ namespace, name: pod.metadata.name })
    ElMessage.success('Pod 已删除')
    fetchPods()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || '删除失败')
    }
  }
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

// ---- Resize: top-bottom (Pods / Events) ----
const rightTopHeight = ref<number | null>(null)
const resizingV = ref(false)
let startY = 0, startH = 0
function onVResizeStart(e: MouseEvent) {
  e.preventDefault()
  resizingV.value = true
  startY = e.clientY
  const rightPanel = (e.target as HTMLElement).closest('.right-panel')
  if (!rightPanel) return
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
  fetchDetail()
  fetchPods()
  fetchEvents()
}, { autoStart: false })

onMounted(() => {
  fetchDetail().then(() => {
    fetchPods()
  })
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
          <span class="replicas-info" v-if="statefulSet">
            {{ statefulSet.status?.readyReplicas ?? 0 }}/{{ statefulSet.spec?.replicas ?? 0 }} ready
          </span>
        </div>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="handleScale">扩缩容</el-button>
        <el-button type="warning" @click="handleRestart">重启</el-button>
        <el-button type="info" @click="handleEdit">编辑</el-button>
        <el-button @click="handleOpenYaml">YAML</el-button>
        <el-button type="danger" plain @click="handleDelete">删除</el-button>
        <div class="action-divider" />
        <el-popover placement="bottom" :width="200" trigger="hover">
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
          <el-button :icon="ArrowLeft" @click="router.push('/workloads/statefulsets')" />
        </el-tooltip>
      </div>
    </div>

    <template v-if="statefulSet">
      <div class="main-layout" :class="{ 'is-resizing': resizingH || resizingV }">

        <!-- 左侧：基本信息 -->
        <div class="left-panel" :style="{ width: leftWidth + 'px', minWidth: leftWidth + 'px' }">
          <div class="panel-title">基本信息</div>
          <div class="info-body">
            <el-descriptions :column="1" border size="small">
              <el-descriptions-item label="名称">{{ statefulSet.metadata?.name }}</el-descriptions-item>
              <el-descriptions-item label="命名空间">{{ statefulSet.metadata?.namespace }}</el-descriptions-item>
              <el-descriptions-item label="副本数">{{ statefulSet.spec?.replicas ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="就绪副本">{{ statefulSet.status?.readyReplicas ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="已更新副本">{{ statefulSet.status?.updatedReplicas ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="当前副本">{{ statefulSet.status?.currentReplicas ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="服务名称">{{ statefulSet.spec?.serviceName || '-' }}</el-descriptions-item>
              <el-descriptions-item label="更新策略">{{ statefulSet.spec?.updateStrategy?.type || 'RollingUpdate' }}</el-descriptions-item>
            </el-descriptions>

            <!-- Labels -->
            <div v-if="statefulSet.metadata?.labels && Object.keys(statefulSet.metadata.labels).length > 0" style="margin-top: 16px;">
              <h4 style="margin: 0 0 8px; font-size: 13px;">Labels</h4>
              <el-tag
                v-for="(val, key) in statefulSet.metadata.labels"
                :key="key"
                style="margin-right: 8px; margin-bottom: 8px;"
                size="small"
              >
                {{ key }}={{ val }}
              </el-tag>
            </div>

            <!-- Selector -->
            <div v-if="statefulSet.spec?.selector?.matchLabels && Object.keys(statefulSet.spec.selector.matchLabels).length > 0" style="margin-top: 16px;">
              <h4 style="margin: 0 0 8px; font-size: 13px;">Selector</h4>
              <el-tag
                v-for="(val, key) in statefulSet.spec.selector.matchLabels"
                :key="key"
                style="margin-right: 8px; margin-bottom: 8px;"
                type="info"
                size="small"
              >
                {{ key }}={{ val }}
              </el-tag>
            </div>

            <!-- Volume Claim Templates -->
            <div v-if="statefulSet.spec?.volumeClaimTemplates?.length" style="margin-top: 16px;">
              <h4 style="margin: 0 0 8px; font-size: 13px;">持久卷声明模板</h4>
              <el-table :data="statefulSet.spec.volumeClaimTemplates" border size="small">
                <el-table-column label="名称" prop="metadata.name" width="150" />
                <el-table-column label="访问模式">
                  <template #default="{ row }">
                    {{ row.spec?.accessModes?.join(', ') || '-' }}
                  </template>
                </el-table-column>
                <el-table-column label="存储容量">
                  <template #default="{ row }">
                    {{ row.spec?.resources?.requests?.storage || '-' }}
                  </template>
                </el-table-column>
                <el-table-column label="存储类">
                  <template #default="{ row }">
                    {{ row.spec?.storageClassName || '-' }}
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </div>
        </div>

        <!-- 右侧：Pods + Events -->
        <div class="right-panel">

          <!-- Pod 列表 -->
          <div class="right-section" :style="rightTopHeight ? { flex: 'none', height: rightTopHeight + 'px' } : {}">
            <div class="panel-title">
              关联 Pod
              <span class="count-badge">{{ pods.length }} 个</span>
            </div>
            <PodListPanel
              :pods="pods"
              :loading="podsLoading"
              @logs="handlePodLogs"
              @exec="handlePodExec"
              @delete="handleDeletePod"
            />
          </div>

          <!-- 垂直拖拽条 -->
          <div class="resize-handle-v" :class="{ active: resizingV }" @mousedown="onVResizeStart" />

          <!-- Events -->
          <div class="right-section events-section">
            <div class="panel-title">
              事件
              <span class="count-badge">{{ events.length }} 条</span>
            </div>
            <div v-loading="eventsLoading" class="events-body">
              <el-table v-if="events.length > 0" :data="events" size="small" stripe max-height="260">
                <el-table-column prop="type" label="类型" width="80">
                  <template #default="{ row }">
                    <el-tag :type="row.type === 'Warning' ? 'danger' : 'info'" size="small">{{ row.type }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="reason" label="原因" width="130" />
                <el-table-column prop="message" label="信息" min-width="200" show-overflow-tooltip />
                <el-table-column prop="last_seen" label="最后发生" width="150" />
              </el-table>
              <div v-else class="empty-hint">暂无事件</div>
            </div>
          </div>

        </div>

        <!-- 水平拖拽条（绝对定位，覆盖在左右面板交界） -->
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
      resource-type="statefulset"
      :namespace="namespace"
      :name="name"
      @saved="handleYamlSaved"
    />

    <!-- Scale Dialog -->
    <el-dialog v-model="scaleDialogVisible" title="扩缩容" width="480px" destroy-on-close>
      <div>
        <p style="margin-bottom: 16px;">调整 <strong>{{ name }}</strong> 副本数</p>
        <el-descriptions :column="1" border size="small" style="margin-bottom: 16px;">
          <el-descriptions-item label="当前">{{ statefulSet?.spec?.replicas ?? '-' }}</el-descriptions-item>
          <el-descriptions-item label="就绪">{{ statefulSet?.status?.readyReplicas ?? '-' }}</el-descriptions-item>
        </el-descriptions>
        <el-form-item label="目标">
          <el-input-number v-model="scaleReplicas" :min="0" :max="100" style="width: 200px;" />
        </el-form-item>
        <el-alert v-if="scaleReplicas === 0" title="设为 0 将停止所有 Pod。" type="warning" :closable="false" show-icon style="margin-top: 8px;" />
      </div>
      <template #footer>
        <el-button @click="scaleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="scaleLoading" @click="handleScaleConfirm">确认</el-button>
      </template>
    </el-dialog>

    <el-drawer
      v-model="editDialogVisible"
      title="编辑 StatefulSet"
      :size="editFullscreen ? '100%' : '85%'"
      direction="rtl"
      :destroy-on-close="true"
      :body-style="{ padding: '0', height: '100%' }"
    >
      <template #header>
        <div class="drawer-header">
          <span class="drawer-title">编辑 StatefulSet</span>
          <el-tooltip :content="editFullscreen ? '退出全屏' : '全屏'" placement="top">
            <el-icon class="fullscreen-btn" @click="editFullscreen = !editFullscreen">
              <FullScreen v-if="!editFullscreen" />
              <Aim v-else />
            </el-icon>
          </el-tooltip>
        </div>
      </template>
      <div style="height: calc(100vh - 52px); overflow-y: auto;">
        <WorkloadForm
          v-if="editDialogVisible && statefulSet"
          kind="StatefulSet"
          :is-edit="true"
          :initial-data="statefulSet"
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

.right-section:first-child {
  flex: 1;
  min-height: 0;
}

.right-section.events-section {
  flex: 1;
  min-height: 0;
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

.events-body {
  flex: 1;
  overflow-y: auto;
  padding: 0;
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
