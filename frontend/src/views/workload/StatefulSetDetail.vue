<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Timer, ArrowLeft, FullScreen, Aim } from '@element-plus/icons-vue'
import {
  getStatefulSetDetail,
  deleteStatefulSet,
  scaleStatefulSet,
  restartStatefulSet,
  getStatefulSetEvents,
  getStatefulSetPods,
  deletePod,
  updateStatefulSetImage,
  rollbackStatefulSet,
  getStatefulSetRollbacks,
} from '@/api/resource'
import YamlDrawer from '@/components/YamlDrawer.vue'
import PodListPanel from '@/components/PodListPanel.vue'
import StatefulSetForm from '@/views/workload/components/StatefulSetForm.vue'
import AutoscalingDrawer from '@/views/workload/components/AutoscalingDrawer.vue'
import ScaleDialog from './components/ScaleDialog.vue'
import UpdateImageDialog from './components/UpdateImageDialog.vue'
import { useDetailPage } from '@/composables/useDetailPage'
import { useResizable } from '@/composables/useResizable'
import { formatAge } from '@/utils/time'
import { buildFullscreenUrl } from '@/utils/pod'

// ---- useDetailPage composable ----
const {
  namespace, name,
  loading, detail: statefulset, events, eventsLoading, yamlDialogVisible,
  isRunning, countdown, currentInterval, availableIntervals,
  toggle, manualRefresh, setIntervalOption,
  fetchDetail, handleDelete, handleOpenYaml,
  router,
  clusterName,
} = useDetailPage({
  resourceName: 'StatefulSet',
  fetchDetail: (p) => getStatefulSetDetail(p),
  fetchEvents: (p) => getStatefulSetEvents(p),
  deleteResource: (p) => deleteStatefulSet(p),
  listRoute: '/workloads/statefulsets',
  buildParams: () => ({ namespace, name }),
  onRefresh: async () => {
    await fetchRevisions()
    await fetchAllPods()
  },
})

// ---- Revisions & Pods ----
const revisions = ref<any[]>([])
const revisionsLoading = ref(false)
const selectedRevision = ref<any>(null)
const allPods = ref<any[]>([])
const rsPods = ref<any[]>([])
const rsPodsLoading = ref(false)

// 左侧视图切换：修订历史 / 基本信息
const leftView = ref<'revisions' | 'info'>('revisions')

// ---- Resize: left-right + top-bottom ----
const { leftWidth, rightTopHeight, resizingH, resizingV, onHResizeStart, onVResizeStart } = useResizable({ initialWidth: 320 })

// ---- 对话框状态 ----
const scaleDialogVisible = ref(false)
const imageDialogVisible = ref(false)
const editDialogVisible = ref(false)
const editFullscreen = ref(false)
const autoscalingDrawerVisible = ref(false)

// ---- Status ----
const statusTagType = computed(() => {
  const ready = statefulset.value?.status?.readyReplicas || 0
  const desired = statefulset.value?.spec?.replicas || 0
  if (ready === desired && desired > 0) return 'success'
  if (ready > 0) return 'warning'
  return 'danger'
})

const statusText = computed(() => {
  const ready = statefulset.value?.status?.readyReplicas || 0
  const desired = statefulset.value?.spec?.replicas || 0
  if (ready === desired && desired > 0) return 'Ready'
  if (ready > 0) return 'Partial'
  return 'Not Ready'
})

const currentReplicas = computed(() => statefulset.value?.spec?.replicas ?? 1)
const readyReplicas = computed(() => statefulset.value?.status?.readyReplicas ?? 0)

const containers = computed(() => {
  return (statefulset.value?.spec?.template?.spec?.containers || []).map((c: any) => ({
    name: c.name,
    image: c.image || '',
  }))
})

// ---- Revision management ----
async function fetchRevisions() {
  revisionsLoading.value = true
  try {
    const res: any = await getStatefulSetRollbacks({ namespace, name })
    revisions.value = res.data || []
    // 自动选中当前 revision
    const currentRev = statefulset.value?.status?.currentRevision
    if (currentRev) {
      const current = revisions.value.find((r: any) => r.name === currentRev)
      if (current) {
        handleRevisionSelect(current)
        return
      }
    }
    // fallback：选中第一个（最新）
    if (revisions.value.length > 0) {
      handleRevisionSelect(revisions.value[0])
    }
  } catch (e) {
    revisions.value = []
  } finally {
    revisionsLoading.value = false
  }
}

async function fetchAllPods() {
  rsPodsLoading.value = true
  try {
    const res: any = await getStatefulSetPods({ namespace, name })
    allPods.value = res.data?.items || res.data || []
    // 若已选中 revision 则过滤
    if (selectedRevision.value) {
      handleRevisionSelect(selectedRevision.value)
    } else {
      rsPods.value = allPods.value
    }
  } catch (e) {
    allPods.value = []
    rsPods.value = []
  } finally {
    rsPodsLoading.value = false
  }
}

function handleRevisionSelect(rev: any) {
  selectedRevision.value = rev
  rsPods.value = allPods.value.filter((pod: any) => {
    const labels = pod.metadata?.labels || {}
    return labels['controller-revision-hash'] === rev.name
  })
}

function revisionPodCount(rev: any): number {
  return allPods.value.filter((pod: any) => {
    const labels = pod.metadata?.labels || {}
    return labels['controller-revision-hash'] === rev.name
  }).length
}

async function handleRevisionRollback(rev: any) {
  try {
    await ElMessageBox.confirm(
      `确定要回滚到 revision ${rev.revision} 吗？`,
      '确认回滚',
      { type: 'warning' }
    )
    await rollbackStatefulSet({ namespace, name, revision: rev.revision })
    ElMessage.success('回滚成功')
    fetchDetail()
    fetchRevisions()
    fetchAllPods()
  } catch (error: any) {
    if (error !== 'cancel' && error !== 'close') {
      ElMessage.error(error?.message || '回滚失败')
    }
  }
}

// ---- Pod operations ----
function handlePodLogs(pod: any) {
  const cluster = clusterName.value
  window.open(buildFullscreenUrl('logs', { namespace: pod.metadata?.namespace || namespace, pod: pod.metadata?.name, cluster }), '_blank')
}

function handlePodExec(pod: any) {
  const cluster = clusterName.value
  window.open(buildFullscreenUrl('terminal', { namespace: pod.metadata?.namespace || namespace, pod: pod.metadata?.name, cluster }), '_blank')
}

async function handlePodDelete(pod: any, force = false) {
  if (force) {
    try {
      await ElMessageBox.confirm(
        `强制删除 Pod "${pod.metadata?.name}" 将跳过优雅终止，控制器管理的 Pod 会被立即重建。确定继续？`,
        '确认强制删除',
        { type: 'warning', confirmButtonText: '强制删除', cancelButtonText: '取消' }
      )
    } catch {
      return
    }
    try {
      await deletePod({ namespace: pod.metadata.namespace || namespace, name: pod.metadata.name, force: true })
      ElMessage.success('Pod 已强制删除')
      if (selectedRevision.value) handleRevisionSelect(selectedRevision.value)
    } catch (e: any) {
      if (e !== 'cancel') ElMessage.error(e?.message || '强制删除失败')
    }
    return
  }
  try {
    await ElMessageBox.confirm(
      `确定要删除 Pod "${pod.metadata?.name}" 吗？`,
      '确认删除',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await deletePod({ namespace: pod.metadata.namespace || namespace, name: pod.metadata.name })
    ElMessage.success('Pod 已删除')
    if (selectedRevision.value) handleRevisionSelect(selectedRevision.value)
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || '删除失败')
    }
  }
}

// ---- YAML ----
function handleYamlSaved() {
  fetchDetail()
  fetchRevisions()
}

// ---- Restart ----
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
    fetchRevisions()
    fetchAllPods()
  } catch {
    // cancelled
  }
}

// ---- Scale ----
function handleScaleSuccess() {
  fetchDetail()
  fetchAllPods()
}

// ---- Update image ----
function handleUpdateImageSuccess() {
  fetchDetail()
  fetchRevisions()
  fetchAllPods()
}

// ---- Edit ----
function handleEdit() {
  editDialogVisible.value = true
}

function handleEditSuccess() {
  editDialogVisible.value = false
  fetchDetail()
  fetchRevisions()
  fetchAllPods()
}

function handleEditCancel() {
  editDialogVisible.value = false
}
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
          <span class="replicas-info" v-if="statefulset">
            {{ statefulset.status?.readyReplicas ?? 0 }}/{{ statefulset.spec?.replicas ?? 0 }} ready
          </span>
        </div>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="scaleDialogVisible = true">扩缩容</el-button>
        <el-button type="warning" @click="handleRestart">重启</el-button>
        <el-button type="success" @click="imageDialogVisible = true">更新镜像</el-button>
        <el-button type="info" @click="handleEdit">编辑</el-button>
        <el-button @click="handleOpenYaml">YAML</el-button>
        <el-button type="warning" @click="autoscalingDrawerVisible = true">弹性伸缩</el-button>
        <el-button type="danger" @click="handleDelete">删除</el-button>
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
          <el-button :icon="ArrowLeft" @click="router.push('/workloads/statefulsets')" />
        </el-tooltip>
      </div>
    </div>

    <template v-if="statefulset">
      <div class="main-layout" :class="{ 'is-resizing': resizingH || resizingV }">

        <!-- 左侧：修订历史 / 基本信息 -->
        <div class="left-panel" :style="{ width: leftWidth + 'px', minWidth: leftWidth + 'px' }">
          <div class="left-tabs">
            <el-segmented
              v-model="leftView"
              :options="[
                { label: '修订历史', value: 'revisions' },
                { label: '基本信息', value: 'info' },
              ]"
              size="small"
              block
            />
          </div>

          <!-- 修订历史 -->
          <div v-show="leftView === 'revisions'" class="rs-list" v-loading="revisionsLoading">
            <div v-if="revisions.length === 0" class="empty-hint">暂无修订历史</div>
            <div
              v-for="rev in revisions"
              :key="rev.revision"
              class="rs-item"
              :class="{ active: selectedRevision?.name === rev.name }"
              @click="handleRevisionSelect(rev)"
            >
              <div class="rs-name">{{ rev.name }}</div>
              <div class="rs-meta">
                <span class="rs-rev">v{{ rev.revision }}</span>
                <span class="rs-replicas">{{ revisionPodCount(rev) }} 个 Pod</span>
                <el-tag
                  v-if="rev.name === statefulset?.status?.currentRevision"
                  type="success" size="small">当前</el-tag>
                <el-tag v-else-if="revisionPodCount(rev) > 0" type="primary" size="small">活跃</el-tag>
              </div>
              <div class="rs-image" v-for="(img, i) in (rev.images || [])" :key="i">{{ img }}</div>
              <div class="rs-age">{{ formatAge(rev.createdAt) }}</div>
              <div class="rs-rollback" v-if="rev.name !== statefulset?.status?.currentRevision">
                <el-button size="small" type="warning" @click.stop="handleRevisionRollback(rev)">回滚</el-button>
              </div>
            </div>
          </div>

          <!-- 基本信息 -->
          <div v-show="leftView === 'info'" class="info-body">
            <el-descriptions :column="1" border size="small">
              <el-descriptions-item label="名称">{{ statefulset?.metadata?.name || '-' }}</el-descriptions-item>
              <el-descriptions-item label="命名空间">{{ statefulset?.metadata?.namespace || '-' }}</el-descriptions-item>
              <el-descriptions-item label="副本">
                {{ statefulset?.spec?.replicas ?? '-' }} 期望 ·
                {{ statefulset?.status?.readyReplicas ?? 0 }} 就绪 ·
                {{ statefulset?.status?.currentReplicas ?? 0 }} 当前 ·
                {{ statefulset?.status?.updatedReplicas ?? 0 }} 更新中
              </el-descriptions-item>
              <el-descriptions-item label="serviceName">{{ statefulset?.spec?.serviceName || '-' }}</el-descriptions-item>
              <el-descriptions-item label="Pod 管理策略">{{ statefulset?.spec?.podManagementPolicy || '-' }}</el-descriptions-item>
              <el-descriptions-item label="更新策略">
                {{ statefulset?.spec?.updateStrategy?.type || '-' }}
                <span v-if="statefulset?.spec?.updateStrategy?.type === 'RollingUpdate'" class="info-sub">
                  (partition {{ statefulset.spec.updateStrategy?.rollingUpdate?.partition ?? 0 }})
                </span>
              </el-descriptions-item>
              <el-descriptions-item label="当前 revision">{{ statefulset?.status?.currentRevision || '-' }}</el-descriptions-item>
              <el-descriptions-item label="更新 revision">{{ statefulset?.status?.updateRevision || '-' }}</el-descriptions-item>
              <el-descriptions-item label="历史上限">{{ statefulset?.spec?.revisionHistoryLimit ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ statefulset?.metadata?.creationTimestamp || '-' }}</el-descriptions-item>
              <el-descriptions-item label="UID">{{ statefulset?.metadata?.uid || '-' }}</el-descriptions-item>
            </el-descriptions>

            <div class="info-section-title">容器镜像</div>
            <div class="vct-list">
              <div v-for="c in (statefulset?.spec?.template?.spec?.containers || [])" :key="c.name" class="vct-item">
                <span class="vct-name">{{ c.name }}</span>
                <span class="vct-meta">{{ c.image || '-' }}</span>
              </div>
              <div v-if="!statefulset?.spec?.template?.spec?.containers?.length" class="info-empty">无</div>
            </div>

            <div class="info-section-title">Conditions</div>
            <div v-if="statefulset?.status?.conditions?.length" class="conditions-list">
              <div v-for="cond in statefulset.status.conditions" :key="cond.type" class="condition-item">
                <div class="condition-head">
                  <span class="condition-type">{{ cond.type }}</span>
                  <el-tag :type="cond.status === 'True' ? 'success' : (cond.status === 'False' ? 'danger' : 'info')" size="small">{{ cond.status }}</el-tag>
                </div>
                <div v-if="cond.reason || cond.message" class="condition-msg">
                  <span v-if="cond.reason" class="condition-reason">{{ cond.reason }}</span>
                  <span v-if="cond.message" class="condition-text">{{ cond.message }}</span>
                </div>
                <div v-if="cond.lastTransitionTime" class="condition-time">{{ cond.lastTransitionTime }}</div>
              </div>
            </div>
            <div v-else class="info-empty">无</div>

            <div class="info-section-title">volumeClaimTemplates</div>
            <div v-if="statefulset?.spec?.volumeClaimTemplates?.length" class="vct-list">
              <div v-for="vct in statefulset.spec.volumeClaimTemplates" :key="vct.name" class="vct-item">
                <span class="vct-name">{{ vct.name }}</span>
                <span class="vct-meta">{{ vct.spec?.resources?.requests?.storage || '-' }} · {{ (vct.spec?.accessModes || []).join(', ') || '-' }}<span v-if="vct.spec?.storageClassName"> · {{ vct.spec.storageClassName }}</span></span>
              </div>
            </div>
            <div v-else class="info-empty">无</div>

            <div class="info-section-title">Selector</div>
            <div class="label-list">
              <el-tag v-for="(v, k) in (statefulset?.spec?.selector?.matchLabels || {})" :key="k" size="small" class="label-tag">{{ k }}={{ v }}</el-tag>
              <span v-if="!statefulset?.spec?.selector?.matchLabels || Object.keys(statefulset.spec.selector.matchLabels).length === 0" class="info-empty">无</span>
            </div>

            <div class="info-section-title">Labels</div>
            <div class="label-list">
              <el-tag v-for="(v, k) in (statefulset?.metadata?.labels || {})" :key="k" size="small" type="info" class="label-tag">{{ k }}={{ v }}</el-tag>
              <span v-if="!statefulset?.metadata?.labels || Object.keys(statefulset.metadata.labels).length === 0" class="info-empty">无</span>
            </div>
          </div>
        </div>

        <!-- 右侧：Pods + Events -->
        <div class="right-panel">

          <!-- Pod 列表 -->
          <div class="right-section" :style="rightTopHeight ? { flex: 'none', height: rightTopHeight + 'px' } : {}">
            <div class="panel-title">
              关联 Pod
              <span class="count-badge">{{ rsPods.length }} 个</span>
              <span class="rs-label" v-if="selectedRevision">{{ selectedRevision.name }}</span>
            </div>
            <PodListPanel
              :pods="rsPods"
              :loading="rsPodsLoading"
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
      resource-type="statefulset"
      :namespace="namespace"
      :name="name"
      @saved="handleYamlSaved"
    />

    <!-- Scale Dialog -->
    <ScaleDialog
      v-model:visible="scaleDialogVisible"
      resource-name="StatefulSet"
      :namespace="namespace"
      :name="name"
      :current-replicas="currentReplicas"
      :ready-replicas="readyReplicas"
      :scale-fn="(p) => scaleStatefulSet(p)"
      @scaled="handleScaleSuccess"
    />

    <!-- Image Update Dialog -->
    <UpdateImageDialog
      v-model:visible="imageDialogVisible"
      resource-name="StatefulSet"
      :namespace="namespace"
      :name="name"
      :containers="containers"
      :update-image-fn="(p) => updateStatefulSetImage(p)"
      @updated="handleUpdateImageSuccess"
    />

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
      <div style="height: calc(100dvh - 52px); overflow-y: auto;">
        <StatefulSetForm
          v-if="editDialogVisible && statefulset"
          :is-edit="true"
          :initial-data="statefulset"
          @success="handleEditSuccess"
          @cancel="handleEditCancel"
        />
      </div>
    </el-drawer>

    <!-- Autoscaling Drawer -->
    <AutoscalingDrawer
      v-model:visible="autoscalingDrawerVisible"
      :namespace="namespace"
      :workload-name="name"
      workload-kind="StatefulSet"
    />
  </div>
</template>

<style scoped>
.detail-page {
  padding: var(--gk-space-4) var(--gk-space-5);
  height: calc(100dvh - var(--gk-header-height));
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
}

/* Header */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--gk-space-3);
  flex-shrink: 0;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: var(--gk-space-1);
}

.res-name {
  margin: 0;
  font-size: var(--gk-font-size-lg);
  font-weight: 600;
  line-height: 1.3;
}

.meta-line {
  display: flex;
  align-items: center;
  gap: var(--gk-space-2);
}

.ns-tag {
  font-size: 11px;
  color: var(--gk-color-text-secondary);
  background: var(--gk-neutral-100);
  padding: 1px 6px;
  border-radius: 4px;
}

.replicas-info {
  font-size: 12px;
  color: var(--gk-color-text-primary);
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
  border-radius: var(--gk-radius-sm) 0 0 var(--gk-radius-sm);
  margin-left: 0;
}

.header-actions .el-button:last-of-type,
.header-actions .el-dropdown:last-of-type {
  border-radius: 0 var(--gk-radius-sm) var(--gk-radius-sm) 0;
}

.action-divider {
  width: 1px;
  height: 20px;
  background: var(--el-border-color-lighter);
  margin: 0 var(--gk-space-1);
}

.auto-refresh-popover {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.popover-title {
  font-size: var(--gk-font-size-sm);
  font-weight: 500;
  color: var(--gk-color-text-primary);
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
  border: 1px solid var(--gk-color-border-light);
  border-radius: var(--gk-radius-md);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--el-bg-color);
}

.panel-title {
  font-size: var(--gk-font-size-sm);
  font-weight: 600;
  padding: var(--gk-space-2) var(--gk-space-4);
  background: var(--gk-neutral-100);
  border-bottom: 1px solid var(--gk-color-border-light);
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.count-badge {
  font-weight: 400;
  font-size: var(--gk-font-size-xs);
  color: var(--gk-color-text-secondary);
}

.rs-label {
  margin-left: auto;
  font-weight: 400;
  font-size: 11px;
  color: var(--gk-color-text-placeholder);
  font-family: var(--gk-font-mono);
  max-width: 140px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rs-list {
  flex: 1;
  overflow-y: auto;
}

/* 左侧视图切换 */
.left-tabs {
  padding: 8px 10px;
  border-bottom: 1px solid var(--el-border-color-extra-light);
  flex-shrink: 0;
}

.info-body {
  flex: 1;
  overflow-y: auto;
  padding: 10px;
}

.info-sub {
  font-size: 11px;
  color: var(--gk-color-text-secondary);
}

.info-section-title {
  font-size: var(--gk-font-size-xs);
  font-weight: 600;
  color: var(--gk-color-text-primary);
  margin: 12px 0 6px;
}

.vct-list {
  display: flex;
  flex-direction: column;
  gap: var(--gk-space-1);
}

.vct-item {
  font-size: 12px;
  display: flex;
  justify-content: space-between;
  gap: 8px;
  padding: 4px 6px;
  background: var(--gk-neutral-100);
  border-radius: 4px;
}

.vct-name {
  font-family: var(--gk-font-mono);
  color: var(--gk-color-text-primary);
}

.vct-meta {
  color: var(--gk-color-text-secondary);
  font-size: 11px;
}

.label-list {
  display: flex;
  flex-wrap: wrap;
  gap: var(--gk-space-1);
}

.label-tag {
  font-family: var(--gk-font-mono);
}

.info-empty {
  font-size: 12px;
  color: var(--gk-color-text-placeholder);
}

.conditions-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.condition-item {
  padding: 6px 8px;
  background: var(--gk-neutral-100);
  border-radius: 4px;
}

.condition-head {
  display: flex;
  align-items: center;
  gap: 6px;
}

.condition-type {
  font-size: 12px;
  font-weight: 600;
  color: var(--gk-color-text-primary);
}

.condition-msg {
  font-size: 11px;
  margin-top: 2px;
  display: flex;
  gap: 6px;
}

.condition-reason {
  color: var(--el-color-warning);
  flex-shrink: 0;
}

.condition-text {
  color: var(--gk-color-text-secondary);
  word-break: break-all;
}

.condition-time {
  font-size: 10px;
  color: var(--gk-color-text-placeholder);
  margin-top: 2px;
}

.rs-item {
  padding: var(--gk-space-2) var(--gk-space-4);
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
  font-size: var(--gk-font-size-sm);
  font-weight: 500;
  font-family: var(--gk-font-mono);
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
  color: var(--gk-color-text-secondary);
}

.rs-image {
  font-size: 11px;
  color: var(--gk-color-text-secondary);
  word-break: break-all;
  margin-bottom: 2px;
}

.rs-age {
  font-size: 11px;
  color: var(--gk-color-text-placeholder);
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
  border: 1px solid var(--gk-color-border-light);
  border-radius: var(--gk-radius-md);
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
  background: var(--gk-color-primary-bg);
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
  background: var(--gk-color-primary-bg);
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
  color: var(--gk-color-text-secondary);
  font-size: var(--gk-font-size-sm);
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
  font-size: var(--gk-font-size-lg);
  font-weight: 600;
}

.fullscreen-btn {
  cursor: pointer;
  font-size: 18px;
  color: var(--gk-color-text-primary);
  transition: color 0.2s;
}

.fullscreen-btn:hover {
  color: var(--el-color-primary);
}
</style>
