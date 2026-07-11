<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Timer, Back } from '@element-plus/icons-vue'
import {
  getDeploymentDetail,
  restartDeployment,
  rollbackDeployment,
  scaleDeployment,
  getDeploymentPodList,
  getDeploymentEvents,
  deletePod,
  deleteDeployment,
  getDeploymentReplicaSets,
} from '@/api/resource'
import YamlDrawer from '@/components/YamlDrawer.vue'
import PodListPanel from '@/components/PodListPanel.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import { formatAge } from '@/utils/time'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const deployment = ref<any>(null)
const yamlDialogVisible = ref(false)
const events = ref<any[]>([])
const eventsLoading = ref(false)

// ReplicaSet & Pod panel state
const replicasets = ref<any[]>([])
const replicasetsLoading = ref(false)
const selectedReplicaset = ref<any>(null)
const rsPods = ref<any[]>([])
const allPods = ref<any[]>([])
const rsPodsLoading = ref(false)

// ---- Resize: left-right ----
const leftWidth = ref(320)
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

// Scale dialog
const scaleDialogVisible = ref(false)
const scaleReplicas = ref<number>(1)
const scaleLoading = ref(false)

const namespace = route.params.namespace as string
const name = route.params.name as string

const statusTagType = computed(() => {
  const conditions = deployment.value?.status?.conditions || []
  const available = conditions.find((c: any) => c.type === 'Available')
  if (available?.status === 'True') return 'success'
  const progressing = conditions.find((c: any) => c.type === 'Progressing')
  if (progressing?.status === 'True') return 'warning'
  return 'danger'
})

const statusText = computed(() => {
  const conditions = deployment.value?.status?.conditions || []
  const available = conditions.find((c: any) => c.type === 'Available')
  if (available?.status === 'True') return 'Available'
  const progressing = conditions.find((c: any) => c.type === 'Progressing')
  if (progressing?.status === 'True') return 'Progressing'
  return 'Unavailable'
})

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getDeploymentDetail({ namespace, name })
    deployment.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load deployment detail')
  } finally {
    loading.value = false
  }
}

async function fetchEvents() {
  eventsLoading.value = true
  try {
    const res: any = await getDeploymentEvents({ namespace, name })
    events.value = res.data || []
  } catch (e) {
    console.error('Failed to fetch events:', e)
    ElMessage.error('Failed to load events')
  } finally {
    eventsLoading.value = false
  }
}

async function fetchReplicaSets() {
  replicasetsLoading.value = true
  try {
    const res: any = await getDeploymentReplicaSets({ namespace, name })
    replicasets.value = res.data?.items || res.data || []
  } catch (e) {
    console.error('Failed to fetch replicasets:', e)
    ElMessage.error('Failed to load ReplicaSets')
  } finally {
    replicasetsLoading.value = false
  }

  // Always fetch pods, regardless of replicasets count
  await fetchAllPods()

  // If replicasets exist, try to select the current one
  if (replicasets.value.length > 0) {
    const currentRevision = deployment.value?.metadata?.annotations?.['deployment.kubernetes.io/revision']
    const currentRS = replicasets.value.find(
      (rs: any) => rs.metadata.annotations?.['deployment.kubernetes.io/revision'] === currentRevision
    )
    if (currentRS) {
      handleReplicasetSelect(currentRS)
    }
  }
}

async function fetchAllPods() {
  rsPodsLoading.value = true
  try {
    const res: any = await getDeploymentPodList({ namespace, name })
    allPods.value = res.data?.items || res.data || []
    rsPods.value = allPods.value
  } catch (e) {
    console.error("Failed to fetch pods:", e)
    ElMessage.error("Failed to load pods")
  } finally {
    rsPodsLoading.value = false
  }
}

function handleReplicasetSelect(rs: any) {
  selectedReplicaset.value = rs
  const hash = rs.metadata.name.split('-').pop()
  rsPods.value = allPods.value.filter((pod: any) => {
    const labels = pod.metadata?.labels || {}
    return labels['pod-template-hash'] === hash
  })
}

async function handleReplicasetRollback(rs: any) {
  const revision = rs.metadata.annotations?.['deployment.kubernetes.io/revision']
  if (!revision) return
  try {
    await ElMessageBox.confirm(
      `Are you sure you want to rollback to revision ${revision}?`,
      'Confirm Rollback',
      { type: 'warning' }
    )
    await rollbackDeployment({ namespace, name, revision: parseInt(revision, 10) })
    ElMessage.success('Rollback successful')
    fetchDetail()
    fetchReplicaSets()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('Rollback failed')
    }
  }
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
  window.open(`/fullscreen/logs?namespace=${pod.metadata.namespace || namespace}&pod=${pod.metadata.name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
}

function handlePodExec(pod: any) {
  const cluster = getClusterName()
  window.open(`/fullscreen/terminal?namespace=${pod.metadata.namespace || namespace}&pod=${pod.metadata.name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
}

async function handlePodDelete(pod: any) {
  try {
    await ElMessageBox.confirm(
      `Are you sure you want to delete pod ${pod.metadata.name}?`,
      'Confirm Delete',
      { type: 'warning' }
    )
    await deletePod({ namespace, name: pod.metadata.name })
    ElMessage.success('Pod deleted')
    if (selectedReplicaset.value) {
      handleReplicasetSelect(selectedReplicaset.value)
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('Delete failed')
    }
  }
}

function handleOpenYaml() {
  yamlDialogVisible.value = true
}

function handleYamlSaved() {
  fetchDetail()
  fetchReplicaSets()
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(
      `确定要删除 Deployment "${name}" 吗？此操作不可恢复。`,
      '确认删除',
      { type: 'error', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await deleteDeployment({ namespace, name })
    ElMessage.success('Deployment 已删除')
    router.push('/workloads/deployments')
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || '删除失败')
    }
  }
}

async function handleRestart() {
  try {
    await ElMessageBox.confirm(
      `Restart deployment "${name}"? This will trigger a rolling update.`,
      'Confirm Restart',
      { type: 'warning' }
    )
    await restartDeployment({ namespace, name })
    ElMessage.success('Deployment restarted')
    fetchDetail()
    fetchReplicaSets()
  } catch {
    // cancelled
  }
}

function handleScale() {
  scaleReplicas.value = deployment.value?.spec?.replicas ?? 1
  scaleDialogVisible.value = true
}

async function handleScaleConfirm() {
  scaleLoading.value = true
  try {
    await scaleDeployment({ namespace, name, replicas: scaleReplicas.value })
    ElMessage.success(`Deployment scaled to ${scaleReplicas.value} replicas`)
    scaleDialogVisible.value = false
    await fetchDetail()
    // Poll for pod list update (K8s needs time to create/delete pods)
    const expectedPods = scaleReplicas.value
    for (let i = 0; i < 10; i++) {
      await new Promise(r => setTimeout(r, 1000))
      await fetchReplicaSets()
      if (allPods.value.length === expectedPods) break
    }
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to scale deployment')
  } finally {
    scaleLoading.value = false
  }
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(async () => {
  fetchDetail()
  fetchReplicaSets()
  fetchEvents()
}, { autoStart: false })

onMounted(() => {
  fetchDetail().then(() => {
    fetchReplicaSets()
  })
  fetchEvents()
})
</script>

<template>
  <div class="detail-page" v-loading="loading">

    <!-- ===== 顶部标题栏 ===== -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="res-name">{{ name }}</h2>
        <div class="meta-line">
          <el-tag :type="statusTagType" effect="dark" size="small">{{ statusText }}</el-tag>
          <span class="ns-tag">ns/{{ namespace }}</span>
          <span class="replicas-info" v-if="deployment">
            {{ deployment.status?.readyReplicas ?? 0 }}/{{ deployment.spec?.replicas ?? 0 }} ready
          </span>
        </div>
      </div>
      <div class="header-actions">
        <el-button @click="manualRefresh()" :loading="loading" :icon="Refresh">刷新</el-button>
        <el-button
          :type="isRunning ? 'success' : 'default'"
          @click="toggle()"
        >
          <el-icon><Timer /></el-icon>
          {{ isRunning ? `自动刷新 ${countdown}s` : '自动刷新' }}
        </el-button>
        <el-select
          :model-value="currentInterval / 1000"
          @update:model-value="setIntervalOption"
          style="width: 90px;"
        >
          <el-option
            v-for="sec in availableIntervals"
            :key="sec"
            :value="sec"
            :label="`${sec}s`"
          />
        </el-select>
        <el-button type="primary" @click="handleScale">扩缩容</el-button>
        <el-button type="warning" @click="handleRestart">重启</el-button>
        <el-button @click="handleOpenYaml">YAML</el-button>
        <el-button type="danger" plain @click="handleDelete">删除</el-button>
        <el-tooltip content="返回列表" placement="top">
          <el-button :icon="Back" @click="router.push('/workloads/deployments')" />
        </el-tooltip>
      </div>
    </div>

    <template v-if="deployment">
      <div class="main-layout" :class="{ 'is-resizing': resizingH || resizingV }">

        <!-- 左侧：ReplicaSet 列表 -->
        <div class="left-panel" :style="{ width: leftWidth + 'px', minWidth: leftWidth + 'px' }">
          <div class="panel-title">
            ReplicaSet
            <span class="count-badge">{{ replicasets.length }} 个</span>
          </div>
          <div class="rs-list" v-loading="replicasetsLoading">
            <div v-if="replicasets.length === 0" class="empty-hint">暂无 ReplicaSet</div>
            <div
              v-for="rs in replicasets"
              :key="rs.metadata.name"
              class="rs-item"
              :class="{ active: selectedReplicaset?.metadata?.name === rs.metadata.name }"
              @click="handleReplicasetSelect(rs)"
            >
              <div class="rs-name">{{ rs.metadata.name }}</div>
              <div class="rs-meta">
                <span class="rs-rev">v{{ rs.metadata.annotations?.['deployment.kubernetes.io/revision'] || '?' }}</span>
                <span class="rs-replicas">{{ rs.status?.readyReplicas ?? 0 }}/{{ rs.spec?.replicas ?? 0 }}</span>
                <el-tag
                  v-if="rs.metadata.annotations?.['deployment.kubernetes.io/revision'] === deployment?.metadata?.annotations?.['deployment.kubernetes.io/revision']"
                  type="success" size="small">当前</el-tag>
                <el-tag v-else-if="(rs.status?.readyReplicas || 0) > 0" type="primary" size="small">活跃</el-tag>
              </div>
              <div class="rs-image">{{ rs.spec?.template?.spec?.containers?.[0]?.image || '-' }}</div>
              <div class="rs-age">{{ formatAge(rs.metadata.creationTimestamp) }}</div>
              <div class="rs-rollback" v-if="rs.metadata.annotations?.['deployment.kubernetes.io/revision'] !== deployment?.metadata?.annotations?.['deployment.kubernetes.io/revision']">
                <el-button size="small" type="warning" @click.stop="handleReplicasetRollback(rs)">回滚</el-button>
              </div>
            </div>
          </div>
        </div>

        <!-- 右侧：Pods + Events -->
        <div class="right-panel">

          <!-- Pod 列表 -->
          <div class="right-section" :style="rightTopHeight ? { flex: 'none', height: rightTopHeight + 'px' } : {}">
            <div class="panel-title">
              Pod 列表
              <span class="count-badge">{{ rsPods.length }} 个</span>
              <span class="rs-label" v-if="selectedReplicaset">{{ selectedReplicaset.metadata.name }}</span>
            </div>
            <PodListPanel
              :pods="rsPods"
              :loading="rsPodsLoading"
              :replicaset-name="selectedReplicaset?.metadata?.name || ''"
              @logs="handlePodLogs"
              @exec="handlePodExec"
              @delete="handlePodDelete"
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

        <!-- 水平拖拽条 -->
        <div
          class="resize-handle-h"
          :class="{ active: resizingH }"
          :style="{ left: (leftWidth - 3) + 'px' }"
          @mousedown="onHResizeStart"
        />
      </div>
    </template>

    <!-- ===== Dialogs ===== -->
    <YamlDrawer
      v-model="yamlDialogVisible"
      resource-type="deployment"
      :namespace="namespace"
      :name="name"
      @saved="handleYamlSaved"
    />

    <el-dialog v-model="scaleDialogVisible" title="扩缩容" width="480px" destroy-on-close>
      <div>
        <p style="margin-bottom: 16px;">调整 <strong>{{ name }}</strong> 副本数</p>
        <el-descriptions :column="1" border size="small" style="margin-bottom: 16px;">
          <el-descriptions-item label="当前">{{ deployment?.spec?.replicas ?? '-' }}</el-descriptions-item>
          <el-descriptions-item label="就绪">{{ deployment?.status?.readyReplicas ?? '-' }}</el-descriptions-item>
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
  gap: 6px;
  flex-shrink: 0;
  align-items: center;
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

.rs-label {
  margin-left: auto;
  font-weight: 400;
  font-size: 11px;
  color: var(--el-text-color-placeholder);
  font-family: monospace;
  max-width: 140px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rs-list {
  flex: 1;
  overflow-y: auto;
}

.rs-item {
  padding: 10px 14px;
  border-bottom: 1px solid var(--el-border-color-extra-light);
  cursor: pointer;
  transition: background 0.15s;
}

.rs-item:hover {
  background: var(--el-fill-color-light);
}

.rs-item.active {
  background: var(--el-color-primary-light-9);
  border-left: 3px solid var(--el-color-primary);
}

.rs-name {
  font-size: 13px;
  font-weight: 500;
  font-family: monospace;
  word-break: break-all;
  margin-bottom: 4px;
}

.rs-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.rs-rev {
  font-size: 12px;
  color: var(--el-color-primary);
  font-weight: 500;
}

.rs-replicas {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.rs-image {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  word-break: break-all;
  margin-bottom: 2px;
}

.rs-age {
  font-size: 11px;
  color: var(--el-text-color-placeholder);
}

.rs-rollback {
  margin-top: 6px;
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
</style>
