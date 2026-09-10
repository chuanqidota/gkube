<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { FullScreen, Aim } from '@element-plus/icons-vue'
import {
  getDaemonSetDetail,
  getDaemonSetYaml,
  updateDaemonSetYaml,
  getDaemonSetEvents,
  deleteDaemonSet,
  restartDaemonSet,
  getDaemonSetPods,
  updateDaemonSetImage,
  rollbackDaemonSet,
  getDaemonSetRollbacks,
  getNodeList,
} from '@/api/resource'
import type { NodeInfo } from '@/api/resource'
import YamlDrawer from '@/components/YamlDrawer.vue'
import PodListPanel from '@/components/PodListPanel.vue'
import DetailPageLayout from '@/components/DetailPageLayout.vue'
import DetailPageHeader from '@/components/DetailPageHeader.vue'
import EventsTable from '@/components/EventsTable.vue'
import LabelsBlock from '@/components/LabelsBlock.vue'
import ConditionsBlock from '@/components/ConditionsBlock.vue'
import SelectorBlock from '@/components/SelectorBlock.vue'
import DaemonSetForm from '@/views/workload/components/DaemonSetForm.vue'
import UpdateImageDialog from '@/views/workload/components/UpdateImageDialog.vue'
import { useDetailPage } from '@/composables/useDetailPage'
import { usePodActions } from '@/composables/usePodActions'
import { useRestartAction } from '@/composables/useRestartAction'
import { useEditDrawer } from '@/composables/useEditDrawer'
import { formatAge } from '@/utils/helpers'

const {
  namespace, name,
  loading, detail: daemonset, events, eventsLoading, yamlDialogVisible,
  isRunning, countdown, currentInterval, availableIntervals,
  toggle, manualRefresh, setIntervalOption,
  fetchDetail, fetchEvents, handleDelete, handleOpenYaml,
  router, clusterName,
} = useDetailPage({
  resourceName: 'DaemonSet',
  fetchDetail: getDaemonSetDetail,
  fetchEvents: getDaemonSetEvents,
  deleteResource: deleteDaemonSet,
  listRoute: '/workloads/daemonsets',
  buildParams: () => ({ namespace, name }),
  onRefresh: async () => {
    await fetchRevisions()
    await fetchAllPods()
    await fetchNodes()
  },
})

// Shared composables
const { handlePodLogs, handlePodExec, handlePodDelete } = usePodActions(clusterName)
const { handleRestart } = useRestartAction('daemonset', restartDaemonSet)
const { editDialogVisible, editFullscreen, handleEdit, handleEditSuccess, handleEditCancel } = useEditDrawer(fetchDetail)

// Revisions & Pods
const revisions = ref<any[]>([])
const revisionsLoading = ref(false)
const selectedRevision = ref<any>(null)
const allPods = ref<any[]>([])
const rsPods = ref<any[]>([])
const rsPodsLoading = ref(false)

// 节点列表
const nodeList = ref<NodeInfo[]>([])
const nodesLoading = ref(false)

// 左侧视图切换：修订历史 / 基本信息 / 节点分布
const leftView = ref<'revisions' | 'info' | 'nodes'>('revisions')

// Image update dialog
const imageDialogVisible = ref(false)

const statusTag = computed(() => {
  const desired = daemonset.value?.status?.desiredNumberScheduled || 0
  const ready = daemonset.value?.status?.numberReady || 0
  if (ready === desired && desired > 0) return { text: 'Ready', type: 'success' as const }
  if (ready > 0) return { text: 'Progressing', type: 'warning' as const }
  return { text: 'Unavailable', type: 'danger' as const }
})

const imageContainers = computed(() => {
  const containers = daemonset.value?.spec?.template?.spec?.containers || []
  return containers.map((c: any) => ({ name: c.name, image: c.image || '' }))
})

async function fetchRevisions() {
  revisionsLoading.value = true
  try {
    const res: any = await getDaemonSetRollbacks({ namespace, name })
    revisions.value = res.data || []
    const current = revisions.value.find((r: any) => r.isCurrent)
    if (current) {
      handleRevisionSelect(current)
      return
    }
    if (revisions.value.length > 0) {
      handleRevisionSelect(revisions.value[0])
    }
  } catch {
    revisions.value = []
  } finally {
    revisionsLoading.value = false
  }
}

async function fetchAllPods() {
  rsPodsLoading.value = true
  try {
    const res: any = await getDaemonSetPods({ namespace, name })
    allPods.value = res.data?.items || res.data || []
    if (selectedRevision.value) {
      handleRevisionSelect(selectedRevision.value)
    } else {
      rsPods.value = allPods.value
    }
  } catch {
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

async function fetchNodes() {
  nodesLoading.value = true
  try {
    const res: any = await getNodeList()
    nodeList.value = res.data || []
  } catch {
    nodeList.value = []
  } finally {
    nodesLoading.value = false
  }
}

// 节点-Pod 分布矩阵
interface NodePodItem {
  nodeName: string
  ip: string
  isReady: boolean
  pods: { name: string; phase: string; ready: boolean }[]
}

function nodeBarClass(node: NodePodItem): string {
  if (node.pods.length === 0) return 'bar-empty'
  if (!node.isReady) return 'bar-danger'
  if (node.pods.some(p => p.phase !== 'Running' || !p.ready)) return 'bar-warning'
  return 'bar-success'
}

const nodeDistribution = computed<NodePodItem[]>(() => {
  const nodeMap = new Map<string, NodePodItem>()
  for (const node of nodeList.value) {
    nodeMap.set(node.name, {
      nodeName: node.name,
      ip: node.internal_ip || '-',
      isReady: node.is_ready,
      pods: [],
    })
  }
  const displayPods = selectedRevision.value ? rsPods.value : allPods.value
  for (const pod of displayPods) {
    const nodeName = pod.spec?.nodeName || ''
    if (!nodeName) continue
    let entry = nodeMap.get(nodeName)
    if (!entry) {
      entry = { nodeName, ip: pod.status?.hostIP || '-', isReady: false, pods: [] }
      nodeMap.set(nodeName, entry)
    }
    const allReady = (pod.status?.containerStatuses || []).every((cs: any) => cs.ready)
    entry.pods.push({
      name: pod.metadata?.name || '',
      phase: pod.status?.phase || 'Unknown',
      ready: allReady,
    })
  }
  const items = Array.from(nodeMap.values())
  const score = (item: NodePodItem) => item.pods.some(p => p.phase !== 'Running' || !p.ready) ? 0 : item.pods.length === 0 ? 1 : 2
  return items.sort((a, b) => score(a) - score(b))
})

const nodeDistStats = computed(() => {
  const total = nodeDistribution.value.length
  const withPod = nodeDistribution.value.filter(n => n.pods.length > 0).length
  const abnormal = nodeDistribution.value.filter(n => n.pods.some(p => p.phase !== 'Running' || !p.ready)).length
  const missing = total - withPod
  return { total, withPod, missing, abnormal }
})

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
    await rollbackDaemonSet({ namespace, name, revision: rev.revision })
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

// Pod action wrappers
function onPodLogs(pod: any) {
  handlePodLogs({ namespace: pod.metadata?.namespace || namespace, name: pod.metadata?.name })
}

function onPodExec(pod: any) {
  handlePodExec({ namespace: pod.metadata?.namespace || namespace, name: pod.metadata?.name })
}

function onPodDelete(pod: any, force?: boolean) {
  handlePodDelete(
    { namespace: pod.metadata?.namespace || namespace, name: pod.metadata?.name },
    () => { if (selectedRevision.value) handleRevisionSelect(selectedRevision.value) },
    force
  )
}

function handleYamlSaved() {
  fetchDetail()
  fetchEvents()
  fetchRevisions()
}

function onRestart() {
  handleRestart(namespace, name, () => {
    fetchDetail()
    fetchRevisions()
    fetchAllPods()
  })
}

function onEditSuccess() {
  handleEditSuccess()
  fetchRevisions()
  fetchAllPods()
}

// Image update handlers
async function handleImageUpdateFn(data: { namespace: string; name: string; containerName: string; image: string }) {
  return updateDaemonSetImage(data)
}

function handleImageUpdated() {
  imageDialogVisible.value = false
  fetchDetail()
  fetchRevisions()
  fetchAllPods()
}
</script>

<template>
  <DetailPageLayout :resizable="true">
    <DetailPageHeader
      :title="name"
      :status-tag="statusTag"
      :namespace="namespace"
      :loading="loading"
      :is-running="isRunning"
      :countdown="countdown"
      :current-interval="currentInterval"
      :available-intervals="availableIntervals"
      @refresh="manualRefresh()"
      @toggle="toggle()"
      @set-interval-option="setIntervalOption"
      @back="router.push('/workloads/daemonsets')"
    >
      <template #meta>
        <span class="replicas-info" v-if="daemonset">
          {{ daemonset.status?.numberReady ?? 0 }}/{{ daemonset.status?.desiredNumberScheduled ?? 0 }} ready
        </span>
      </template>
      <template #actions>
        <el-button type="warning" @click="onRestart">重启</el-button>
        <el-button type="success" @click="imageDialogVisible = true">更新镜像</el-button>
        <el-button type="info" @click="handleEdit">编辑</el-button>
        <el-button @click="handleOpenYaml">YAML</el-button>
        <el-button type="danger" @click="handleDelete">删除</el-button>
      </template>
    </DetailPageHeader>

    <template v-if="daemonset" #left>
      <div class="left-tabs">
        <el-segmented
          v-model="leftView"
          :options="[
            { label: '修订历史', value: 'revisions' },
            { label: '基本信息', value: 'info' },
            { label: '节点分布', value: 'nodes' },
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
            <el-tag v-if="rev.isCurrent" type="success" size="small">当前</el-tag>
            <el-tag v-else-if="revisionPodCount(rev) > 0" type="primary" size="small">活跃</el-tag>
          </div>
          <div class="rs-image" v-for="(img, i) in (rev.images || [])" :key="i">{{ img }}</div>
          <div class="rs-age">{{ formatAge(rev.createdAt) }}</div>
          <div class="rs-rollback" v-if="!rev.isCurrent">
            <el-button size="small" type="warning" @click.stop="handleRevisionRollback(rev)">回滚</el-button>
          </div>
        </div>
      </div>

      <!-- 基本信息 -->
      <div v-show="leftView === 'info'" class="info-body">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="名称">{{ daemonset?.metadata?.name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="命名空间">{{ daemonset?.metadata?.namespace || '-' }}</el-descriptions-item>
          <el-descriptions-item label="调度数">
            {{ daemonset?.status?.desiredNumberScheduled ?? 0 }} 期望 ·
            {{ daemonset?.status?.currentNumberScheduled ?? 0 }} 当前 ·
            {{ daemonset?.status?.numberReady ?? 0 }} 就绪 ·
            {{ daemonset?.status?.updatedNumberScheduled ?? 0 }} 更新中 ·
            {{ daemonset?.status?.numberAvailable ?? 0 }} 可用 ·
            {{ daemonset?.status?.numberUnavailable ?? 0 }} 不可用
          </el-descriptions-item>
          <el-descriptions-item label="更新策略">
            {{ daemonset?.spec?.updateStrategy?.type || 'RollingUpdate' }}
            <span v-if="(daemonset?.spec?.updateStrategy?.type || 'RollingUpdate') === 'RollingUpdate'" class="info-sub">
              (maxUnavailable {{ daemonset?.spec?.updateStrategy?.rollingUpdate?.maxUnavailable ?? '-' }},
              maxSurge {{ daemonset?.spec?.updateStrategy?.rollingUpdate?.maxSurge ?? '-' }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item label="当前 revision">{{ revisions.find((r: any) => r.isCurrent)?.name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="历史上限">{{ daemonset?.spec?.revisionHistoryLimit ?? '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ daemonset?.metadata?.creationTimestamp || '-' }}</el-descriptions-item>
          <el-descriptions-item label="UID">{{ daemonset?.metadata?.uid || '-' }}</el-descriptions-item>
        </el-descriptions>

        <div class="info-section-title">容器镜像</div>
        <div class="vct-list">
          <div v-for="c in (daemonset?.spec?.template?.spec?.containers || [])" :key="c.name" class="vct-item">
            <span class="vct-name">{{ c.name }}</span>
            <span class="vct-meta">{{ c.image || '-' }}</span>
          </div>
          <div v-if="!daemonset?.spec?.template?.spec?.containers?.length" class="info-empty">无</div>
        </div>

        <div class="info-section-title">Conditions</div>
        <ConditionsBlock :conditions="daemonset?.status?.conditions || []" />

        <div class="info-section-title">Selector</div>
        <SelectorBlock :selector="daemonset?.spec?.selector?.matchLabels || {}" />

        <div class="info-section-title">Labels</div>
        <LabelsBlock :labels="daemonset?.metadata?.labels || {}" />
      </div>

      <!-- 节点分布 -->
      <div v-show="leftView === 'nodes'" class="node-dist-body" v-loading="nodesLoading">
        <div class="node-stats">
          <div class="node-stat">
            <span class="node-stat-num">{{ nodeDistStats.total }}</span>
            <span class="node-stat-label">节点</span>
          </div>
          <div class="node-stat">
            <span class="node-stat-num" style="color: var(--el-color-success)">{{ nodeDistStats.withPod }}</span>
            <span class="node-stat-label">有 Pod</span>
          </div>
          <div class="node-stat">
            <span class="node-stat-num" style="color: var(--el-color-warning)">{{ nodeDistStats.missing }}</span>
            <span class="node-stat-label">缺失</span>
          </div>
          <div class="node-stat">
            <span class="node-stat-num" style="color: var(--el-color-danger)">{{ nodeDistStats.abnormal }}</span>
            <span class="node-stat-label">异常</span>
          </div>
        </div>

        <div v-if="nodeDistribution.length === 0 && !nodesLoading" class="empty-hint">暂无节点信息</div>

        <div class="node-cards">
          <div
            v-for="node in nodeDistribution"
            :key="node.nodeName"
            class="node-card"
          >
            <div class="node-card-bar" :class="nodeBarClass(node)" />
            <div class="node-card-body">
              <div class="node-card-head">
                <span class="node-card-name" :title="node.nodeName">{{ node.nodeName }}</span>
                <span class="node-card-ip mono">{{ node.ip }}</span>
              </div>
              <div v-if="node.pods.length === 0" class="node-card-empty">未调度</div>
              <div v-else class="node-card-pods">
                <div v-for="pod in node.pods" :key="pod.name" class="node-pod-row">
                  <span class="node-pod-name" :title="pod.name">{{ pod.name }}</span>
                  <el-tag
                    :type="pod.phase === 'Running' && pod.ready ? 'success' : pod.phase === 'Pending' ? 'warning' : 'danger'"
                    size="small"
                  >{{ pod.phase }}</el-tag>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>

    <template v-if="daemonset" #right-top>
      <div class="panel-title">
        关联 Pod
        <span class="count-badge">{{ rsPods.length }} 个</span>
        <span class="rs-label" v-if="selectedRevision">{{ selectedRevision.name }}</span>
      </div>
      <PodListPanel
        :pods="rsPods"
        :loading="rsPodsLoading"
        @logs="onPodLogs"
        @exec="onPodExec"
        @delete="onPodDelete"
      />
    </template>

    <template v-if="daemonset" #right-bottom>
      <div class="panel-title">
        事件
        <span class="count-badge">{{ events.length }} 条</span>
      </div>
      <EventsTable :events="events" :loading="eventsLoading" time-field="last_seen" />
    </template>

    <!-- ===== Dialogs ===== -->
    <YamlDrawer
      v-model="yamlDialogVisible"
      :get-yaml="getDaemonSetYaml"
      :update-yaml="updateDaemonSetYaml"
      :namespace="namespace"
      :name="name"
      title="DaemonSet YAML"
      @saved="handleYamlSaved"
    />

    <UpdateImageDialog
      v-model:visible="imageDialogVisible"
      resource-name="DaemonSet"
      :namespace="namespace"
      :name="name"
      :containers="imageContainers"
      :update-image-fn="handleImageUpdateFn"
      @updated="handleImageUpdated"
    />

    <el-drawer
      v-model="editDialogVisible"
      title="编辑 DaemonSet"
      :size="editFullscreen ? '100%' : '85%'"
      direction="rtl"
      :destroy-on-close="true"
      :body-style="{ padding: '0', height: '100%' }"
    >
      <template #header>
        <div class="drawer-header">
          <span class="drawer-title">编辑 DaemonSet</span>
          <el-tooltip :content="editFullscreen ? '退出全屏' : '全屏'" placement="top">
            <el-icon class="fullscreen-btn" @click="editFullscreen = !editFullscreen">
              <FullScreen v-if="!editFullscreen" />
              <Aim v-else />
            </el-icon>
          </el-tooltip>
        </div>
      </template>
      <div style="height: calc(100dvh - 52px); overflow-y: auto;">
        <DaemonSetForm
          v-if="editDialogVisible && daemonset"
          :is-edit="true"
          :initial-data="daemonset"
          @success="onEditSuccess"
          @cancel="handleEditCancel"
        />
      </div>
    </el-drawer>
  </DetailPageLayout>
</template>

<style scoped>
.replicas-info {
  font-size: 12px;
  color: var(--gk-color-text-primary);
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

.rs-list {
  flex: 1;
  overflow-y: auto;
}

/* 节点分布 */
.node-dist-body {
  flex: 1;
  overflow-y: auto;
  padding: 10px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.node-stats {
  display: flex;
  gap: var(--gk-space-1);
  flex-shrink: 0;
}

.node-stat {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 6px 4px;
  background: var(--gk-neutral-100);
  border-radius: var(--gk-radius-md);
}

.node-stat-num {
  font-size: 18px;
  font-weight: 700;
  line-height: 1.2;
  color: var(--gk-color-text-primary);
}

.node-stat-label {
  font-size: 11px;
  color: var(--gk-color-text-secondary);
  margin-top: 2px;
}

.node-cards {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.node-card {
  display: flex;
  border: 1px solid var(--gk-color-border-light);
  border-radius: var(--gk-radius-md);
  overflow: hidden;
  background: var(--el-bg-color);
  transition: border-color 0.15s;
}

.node-card:hover {
  border-color: var(--el-border-color);
}

.node-card-bar {
  width: 4px;
  flex-shrink: 0;
}

.bar-success { background: var(--el-color-success); }
.bar-warning { background: var(--el-color-warning); }
.bar-danger { background: var(--el-color-danger); }
.bar-empty { background: var(--el-fill-color); }

.node-card-body {
  flex: 1;
  min-width: 0;
  padding: 8px 10px;
}

.node-card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
  margin-bottom: 4px;
}

.node-card-name {
  font-size: 12px;
  font-weight: 600;
  color: var(--gk-color-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.node-card-ip {
  font-size: 11px;
  color: var(--gk-color-text-secondary);
  flex-shrink: 0;
}

.node-card-empty {
  font-size: 11px;
  color: var(--gk-color-text-placeholder);
  font-style: italic;
}

.node-card-pods {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.node-pod-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
}

.node-pod-name {
  font-size: 11px;
  font-family: var(--gk-font-mono);
  color: var(--gk-color-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mono {
  font-family: var(--gk-font-mono);
  font-size: 12px;
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

.info-empty {
  font-size: 12px;
  color: var(--gk-color-text-placeholder);
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

.empty-hint {
  padding: 24px;
  text-align: center;
  color: var(--el-text-color-secondary);
  font-size: var(--gk-font-size-sm);
}

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
