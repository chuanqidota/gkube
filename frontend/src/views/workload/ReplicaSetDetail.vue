<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Timer, ArrowLeft } from '@element-plus/icons-vue'
import {
  getReplicaSetDetail,
  getReplicaSetPodList,
  getReplicaSetEvents,
  deleteReplicaSet,
  deletePod,
} from '@/api/resource'
import YamlDrawer from '@/components/YamlDrawer.vue'
import PodListPanel from '@/components/PodListPanel.vue'
import { useClusterStore } from '@/stores/cluster'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const clusterStore = useClusterStore()
import { useResizable } from '@/composables/useResizable'
import { formatAge } from '@/utils/time'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const detail = ref<any>(null)
const yamlDialogVisible = ref(false)
const pods = ref<any[]>([])
const podsLoading = ref(false)
const events = ref<any[]>([])
const eventsLoading = ref(false)

const namespace = route.params.namespace as string
const name = route.params.name as string

// ---- Resize: left-right + top-bottom ----
const { leftWidth, rightTopHeight, resizingH, resizingV, onHResizeStart, onVResizeStart } = useResizable({ initialWidth: 320 })

const rs = computed(() => detail.value?.rs)

const statusTagType = computed(() => {
  const conditions = rs.value?.status?.conditions || []
  const available = conditions.find((c: any) => c.type === 'Available')
  if (available?.status === 'True') return 'success'
  const failure = conditions.find((c: any) => c.type === 'ReplicaFailure')
  if (failure?.status === 'True') return 'danger'
  return 'warning'
})

const statusText = computed(() => {
  const conditions = rs.value?.status?.conditions || []
  const available = conditions.find((c: any) => c.type === 'Available')
  if (available?.status === 'True') return 'Available'
  const failure = conditions.find((c: any) => c.type === 'ReplicaFailure')
  if (failure?.status === 'True') return 'Failure'
  return 'Progressing'
})

const controllerOf = computed(() => detail.value?.controllerOf || null)

const containers = computed(() => rs.value?.spec?.template?.spec?.containers || [])

const selectorLabels = computed(() => {
  const matchLabels = rs.value?.spec?.selector?.matchLabels || {}
  return Object.entries(matchLabels).map(([k, v]) => `${k}=${v}`)
})

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getReplicaSetDetail({ namespace, name })
    detail.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || '获取详情失败')
  } finally {
    loading.value = false
  }
}

async function fetchPods() {
  podsLoading.value = true
  try {
    const res: any = await getReplicaSetPodList({ namespace, name })
    pods.value = res.data?.items || res.data || []
  } catch (e: any) {
    console.error('Failed to fetch pods:', e)
    ElMessage.error('获取 Pod 列表失败')
  } finally {
    podsLoading.value = false
  }
}

async function fetchEvents() {
  eventsLoading.value = true
  try {
    const res: any = await getReplicaSetEvents({ namespace, name })
    events.value = res.data || []
  } catch (e) {
    console.error('Failed to fetch events:', e)
    ElMessage.error('获取事件失败')
  } finally {
    eventsLoading.value = false
  }
}

function getClusterName(): string {
  return clusterStore.currentCluster?.clusterName || clusterStore.currentCluster?.cluster_name || clusterStore.currentCluster?.name || ''
}

function handlePodLogs(pod: any) {
  const cluster = getClusterName()
  window.open(`/fullscreen/logs?namespace=${pod.metadata.namespace || namespace}&pod=${pod.metadata.name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
}

function handlePodExec(pod: any) {
  const cluster = getClusterName()
  window.open(`/fullscreen/terminal?namespace=${pod.metadata.namespace || namespace}&pod=${pod.metadata.name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
}

async function handlePodDelete(pod: any, force = false) {
  if (force) {
    try {
      await ElMessageBox.confirm(
        `强制删除 Pod "${pod.metadata.name}" 将跳过优雅终止，控制器管理的 Pod 会被立即重建。确定继续？`,
        '确认强制删除',
        { type: 'warning', confirmButtonText: '强制删除', cancelButtonText: '取消' }
      )
    } catch {
      return
    }
    try {
      await deletePod({ namespace, name: pod.metadata.name, force: true })
      ElMessage.success('Pod 已强制删除')
      await fetchPods()
    } catch (e: any) {
      if (e !== 'cancel') ElMessage.error(e?.message || '强制删除失败')
    }
    return
  }
  try {
    await ElMessageBox.confirm(
      `确定要删除 Pod ${pod.metadata.name} 吗？`,
      '确认删除',
      { type: 'warning' }
    )
    await deletePod({ namespace, name: pod.metadata.name })
    ElMessage.success('Pod 已删除')
    await fetchPods()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

function handleOpenYaml() {
  yamlDialogVisible.value = true
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(
      `确定要删除 ReplicaSet "${name}" 吗？此操作不可恢复。`,
      '确认删除',
      { type: 'error', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await deleteReplicaSet({ namespace, name })
    ElMessage.success('ReplicaSet 已删除')
    router.push('/workloads/replicasets')
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || '删除失败')
    }
  }
}

function goController() {
  const c = controllerOf.value
  if (!c) return
  if (c.kind === 'Deployment') {
    router.push(`/workloads/deployments/${c.namespace || namespace}/${c.name}`)
  }
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

    <!-- ===== 顶部标题栏 ===== -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="res-name">{{ name }}</h2>
        <div class="meta-line">
          <el-tag :type="statusTagType" effect="dark" size="small">{{ statusText }}</el-tag>
          <span class="ns-tag">ns/{{ namespace }}</span>
          <span class="replicas-info" v-if="rs">
            {{ rs.status?.readyReplicas ?? 0 }}/{{ rs.spec?.replicas ?? 0 }} ready
          </span>
          <el-tag
            v-if="controllerOf"
            type="primary"
            size="small"
            :class="{ clickable: controllerOf.kind === 'Deployment' }"
            @click="goController"
          >
            由 {{ controllerOf.kind }}/{{ controllerOf.name }} 管理
          </el-tag>
        </div>
      </div>
      <div class="header-actions">
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
          <el-button :icon="ArrowLeft" @click="router.push('/workloads/replicasets')" />
        </el-tooltip>
      </div>
    </div>

    <template v-if="rs">
      <div class="main-layout" :class="{ 'is-resizing': resizingH || resizingV }">

        <!-- 左侧：基本信息 + 容器模板 + 选择器 -->
        <div class="left-panel" :style="{ width: leftWidth + 'px', minWidth: leftWidth + 'px' }">
          <div class="left-scroll">

            <div class="info-block">
              <div class="block-title">基本信息</div>
              <div class="info-row"><span class="info-label">名称</span><span class="info-value mono">{{ rs.metadata?.name }}</span></div>
              <div class="info-row"><span class="info-label">命名空间</span><span class="info-value">{{ rs.metadata?.namespace }}</span></div>
              <div class="info-row"><span class="info-label">期望副本</span><span class="info-value">{{ rs.spec?.replicas ?? 0 }}</span></div>
              <div class="info-row"><span class="info-label">当前副本</span><span class="info-value">{{ rs.status?.replicas ?? 0 }}</span></div>
              <div class="info-row"><span class="info-label">就绪副本</span><span class="info-value">{{ rs.status?.readyReplicas ?? 0 }}</span></div>
              <div class="info-row"><span class="info-label">可用副本</span><span class="info-value">{{ rs.status?.availableReplicas ?? 0 }}</span></div>
              <div class="info-row"><span class="info-label">创建时间</span><span class="info-value">{{ formatAge(rs.metadata?.creationTimestamp) }}</span></div>
              <div class="info-row" v-if="controllerOf">
                <span class="info-label">拥有者</span>
                <span
                  class="info-value link"
                  :class="{ disabled: controllerOf.kind !== 'Deployment' }"
                  @click="goController"
                >{{ controllerOf.kind }}/{{ controllerOf.name }}</span>
              </div>
            </div>

            <div class="info-block">
              <div class="block-title">容器模板</div>
              <div v-if="containers.length === 0" class="empty-hint">暂无容器</div>
              <div v-for="c in containers" :key="c.name" class="container-item">
                <div class="container-name mono">{{ c.name }}</div>
                <div class="container-image mono">{{ c.image || '-' }}</div>
              </div>
            </div>

            <div class="info-block">
              <div class="block-title">选择器</div>
              <div v-if="selectorLabels.length === 0" class="empty-hint">无</div>
              <div class="label-list">
                <el-tag v-for="l in selectorLabels" :key="l" size="small" class="label-tag">{{ l }}</el-tag>
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
              <span class="count-badge">{{ pods.length }} 个</span>
            </div>
            <PodListPanel
              :pods="pods"
              :loading="podsLoading"
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

    <!-- ===== YAML Drawer (只读) ===== -->
    <YamlDrawer
      v-model="yamlDialogVisible"
      resource-type="replicaset"
      :namespace="namespace"
      :name="name"
    />
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

.clickable {
  cursor: pointer;
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

.header-actions .el-button:last-of-type,
.header-actions .el-dropdown:last-of-type {
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

.left-scroll {
  flex: 1;
  overflow-y: auto;
}

.info-block {
  padding: 10px 14px;
  border-bottom: 1px solid var(--el-border-color-extra-light);
}

.info-block:last-child {
  border-bottom: none;
}

.block-title {
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 8px;
  color: var(--el-text-color-primary);
}

.info-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  font-size: 12px;
  line-height: 1.8;
}

.info-label {
  color: var(--el-text-color-secondary);
  flex-shrink: 0;
  width: 64px;
}

.info-value {
  color: var(--el-text-color-regular);
  word-break: break-all;
}

.info-value.link {
  color: var(--el-color-primary);
  cursor: pointer;
}

.info-value.link.disabled {
  cursor: default;
  color: var(--el-text-color-regular);
}

.container-item {
  padding: 6px 0;
  border-top: 1px dashed var(--el-border-color-extra-light);
}

.container-item:first-of-type {
  border-top: none;
  padding-top: 0;
}

.container-name {
  font-size: 12px;
  font-weight: 500;
  color: var(--el-text-color-primary);
  margin-bottom: 2px;
}

.container-image {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  word-break: break-all;
}

.label-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.label-tag {
  font-family: monospace;
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
  padding: 16px 4px;
  text-align: center;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.mono {
  font-family: monospace;
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
