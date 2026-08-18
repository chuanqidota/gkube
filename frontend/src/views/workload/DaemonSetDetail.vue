<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Timer, ArrowLeft, FullScreen, Aim } from '@element-plus/icons-vue'
import {
  getDaemonSetDetail,
  deleteDaemonSet,
  restartDaemonSet,
  getDaemonSetEvents,
  getDaemonSetPods,
  deletePod,
  updateDaemonSetImage,
  rollbackDaemonSet,
  getDaemonSetRollbacks,
  getNodeList,
} from '@/api/resource'
import type { NodeInfo } from '@/api/resource'
import YamlDrawer from '@/components/YamlDrawer.vue'
import PodListPanel from '@/components/PodListPanel.vue'
import DaemonSetForm from '@/views/workload/components/DaemonSetForm.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import { useClusterNameRef } from '@/composables/useClusterName'
import { useResizable } from '@/composables/useResizable'
import { formatAge } from '@/utils/time'
import { buildFullscreenUrl } from '@/utils/pod'

const clusterName = useClusterNameRef()

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const daemonSet = ref<any>(null)
const yamlDialogVisible = ref(false)

// Events
const events = ref<any[]>([])
const eventsLoading = ref(false)

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
const imageForm = ref({
  containerName: '',
  image: '',
})
const imageLoading = ref(false)

// Edit dialog
const editDialogVisible = ref(false)
const editFullscreen = ref(false)

const namespace = route.params.namespace as string
const name = route.params.name as string

if (!namespace || !name) {
  ElMessage.error('缺少必需的路由参数')
  router.push('/workloads/daemonsets')
}

// ---- Resize: left-right + top-bottom ----
const { leftWidth, rightTopHeight, resizingH, resizingV, onHResizeStart, onVResizeStart } = useResizable({ initialWidth: 320 })

const statusTagType = computed(() => {
  const desired = daemonSet.value?.status?.desiredNumberScheduled || 0
  const ready = daemonSet.value?.status?.numberReady || 0
  if (ready === desired && desired > 0) return 'success'
  if (ready > 0) return 'warning'
  return 'danger'
})

const statusText = computed(() => {
  const desired = daemonSet.value?.status?.desiredNumberScheduled || 0
  const ready = daemonSet.value?.status?.numberReady || 0
  if (ready === desired && desired > 0) return 'Ready'
  if (ready > 0) return 'Progressing'
  return 'Unavailable'
})

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getDaemonSetDetail({ namespace, name })
    daemonSet.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || '加载 DaemonSet 详情失败')
  } finally {
    loading.value = false
  }
}

async function fetchEvents() {
  eventsLoading.value = true
  try {
    const res: any = await getDaemonSetEvents({ namespace, name })
    events.value = res.data || []
  } catch (e) {
    events.value = []
  } finally {
    eventsLoading.value = false
  }
}

async function fetchRevisions() {
  revisionsLoading.value = true
  try {
    const res: any = await getDaemonSetRollbacks({ namespace, name })
    revisions.value = res.data || []
    // 自动选中当前 revision（DaemonSet status 无 currentRevision，由后端按模板对比标记 isCurrent）
    const current = revisions.value.find((r: any) => r.isCurrent)
    if (current) {
      handleRevisionSelect(current)
      return
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
    const res: any = await getDaemonSetPods({ namespace, name })
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
  // 以集群节点为基准，匹配 DaemonSet Pod
  const nodeMap = new Map<string, NodePodItem>()
  for (const node of nodeList.value) {
    nodeMap.set(node.name, {
      nodeName: node.name,
      ip: node.internal_ip || '-',
      isReady: node.is_ready,
      pods: [],
    })
  }
  // 将 Pod 按 nodeName 归入对应节点
  const displayPods = selectedRevision.value ? rsPods.value : allPods.value
  for (const pod of displayPods) {
    const nodeName = pod.spec?.nodeName || ''
    if (!nodeName) continue
    let entry = nodeMap.get(nodeName)
    if (!entry) {
      // Pod 所在节点不在 nodeList 中（罕见，如节点刚删除）
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
  // 排序：有异常 Pod → 无 Pod → 正常
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

function handleOpenYaml() {
  yamlDialogVisible.value = true
}

function handleYamlSaved() {
  fetchDetail()
  fetchRevisions()
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(
      `确定要删除 DaemonSet "${name}" 吗？此操作不可恢复。`,
      '确认删除',
      { type: 'error', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await deleteDaemonSet({ namespace, name })
    ElMessage.success('DaemonSet 已删除')
    router.push('/workloads/daemonsets')
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || '删除失败')
    }
  }
}

async function handleRestart() {
  try {
    await ElMessageBox.confirm(
      `确定要重启 DaemonSet "${name}" 吗？这将触发滚动更新。`,
      '确认重启',
      { type: 'warning' }
    )
    await restartDaemonSet({ namespace, name })
    ElMessage.success('DaemonSet 已重启')
    fetchDetail()
    fetchRevisions()
    fetchAllPods()
  } catch {
    // cancelled
  }
}

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

// Image update handlers
function handleUpdateImage() {
  const containers = daemonSet.value?.spec?.template?.spec?.containers || []
  if (containers.length > 0) {
    imageForm.value = {
      containerName: containers[0].name,
      image: containers[0].image || '',
    }
  }
  imageDialogVisible.value = true
}

async function handleUpdateImageConfirm() {
  if (!imageForm.value.containerName || !imageForm.value.image) {
    ElMessage.warning('请填写容器名称和镜像')
    return
  }
  imageLoading.value = true
  try {
    await updateDaemonSetImage({
      namespace,
      name,
      containerName: imageForm.value.containerName,
      image: imageForm.value.image,
    })
    ElMessage.success('镜像更新成功')
    imageDialogVisible.value = false
    fetchDetail()
    fetchRevisions()
    fetchAllPods()
  } catch (e: any) {
    ElMessage.error(e?.message || '镜像更新失败')
  } finally {
    imageLoading.value = false
  }
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(async () => {
  fetchDetail()
  fetchRevisions()
  fetchAllPods()
  fetchEvents()
  fetchNodes()
}, { autoStart: false })

onMounted(() => {
  fetchDetail().then(() => {
    fetchRevisions()
    fetchAllPods()
  })
  fetchEvents()
  fetchNodes()
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
          <span class="replicas-info" v-if="daemonSet">
            {{ daemonSet.status?.numberReady ?? 0 }}/{{ daemonSet.status?.desiredNumberScheduled ?? 0 }} ready
          </span>
        </div>
      </div>
      <div class="header-actions">
        <el-button type="warning" @click="handleRestart">重启</el-button>
        <el-button type="success" @click="handleUpdateImage">更新镜像</el-button>
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
          <el-button :icon="ArrowLeft" @click="router.push('/workloads/daemonsets')" />
        </el-tooltip>
      </div>
    </div>

    <template v-if="daemonSet">
      <div class="main-layout" :class="{ 'is-resizing': resizingH || resizingV }">

        <!-- 左侧：修订历史 / 基本信息 -->
        <div class="left-panel" :style="{ width: leftWidth + 'px', minWidth: leftWidth + 'px' }">
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
                <el-tag
                  v-if="rev.isCurrent"
                  type="success" size="small">当前</el-tag>
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
              <el-descriptions-item label="名称">{{ daemonSet?.metadata?.name || '-' }}</el-descriptions-item>
              <el-descriptions-item label="命名空间">{{ daemonSet?.metadata?.namespace || '-' }}</el-descriptions-item>
              <el-descriptions-item label="调度数">
                {{ daemonSet?.status?.desiredNumberScheduled ?? 0 }} 期望 ·
                {{ daemonSet?.status?.currentNumberScheduled ?? 0 }} 当前 ·
                {{ daemonSet?.status?.numberReady ?? 0 }} 就绪 ·
                {{ daemonSet?.status?.updatedNumberScheduled ?? 0 }} 更新中 ·
                {{ daemonSet?.status?.numberAvailable ?? 0 }} 可用 ·
                {{ daemonSet?.status?.numberUnavailable ?? 0 }} 不可用
              </el-descriptions-item>
              <el-descriptions-item label="更新策略">
                {{ daemonSet?.spec?.updateStrategy?.type || 'RollingUpdate' }}
                <span v-if="(daemonSet?.spec?.updateStrategy?.type || 'RollingUpdate') === 'RollingUpdate'" class="info-sub">
                  (maxUnavailable {{ daemonSet?.spec?.updateStrategy?.rollingUpdate?.maxUnavailable ?? '-' }},
                  maxSurge {{ daemonSet?.spec?.updateStrategy?.rollingUpdate?.maxSurge ?? '-' }})
                </span>
              </el-descriptions-item>
              <el-descriptions-item label="当前 revision">{{ revisions.find((r: any) => r.isCurrent)?.name || '-' }}</el-descriptions-item>
              <el-descriptions-item label="历史上限">{{ daemonSet?.spec?.revisionHistoryLimit ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ daemonSet?.metadata?.creationTimestamp || '-' }}</el-descriptions-item>
              <el-descriptions-item label="UID">{{ daemonSet?.metadata?.uid || '-' }}</el-descriptions-item>
            </el-descriptions>

            <div class="info-section-title">容器镜像</div>
            <div class="vct-list">
              <div v-for="c in (daemonSet?.spec?.template?.spec?.containers || [])" :key="c.name" class="vct-item">
                <span class="vct-name">{{ c.name }}</span>
                <span class="vct-meta">{{ c.image || '-' }}</span>
              </div>
              <div v-if="!daemonSet?.spec?.template?.spec?.containers?.length" class="info-empty">无</div>
            </div>

            <div class="info-section-title">Conditions</div>
            <div v-if="daemonSet?.status?.conditions?.length" class="conditions-list">
              <div v-for="cond in daemonSet.status.conditions" :key="cond.type" class="condition-item">
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

            <div class="info-section-title">Selector</div>
            <div class="label-list">
              <el-tag v-for="(v, k) in (daemonSet?.spec?.selector?.matchLabels || {})" :key="k" size="small" class="label-tag">{{ k }}={{ v }}</el-tag>
              <span v-if="!daemonSet?.spec?.selector?.matchLabels || Object.keys(daemonSet.spec.selector.matchLabels).length === 0" class="info-empty">无</span>
            </div>

            <div class="info-section-title">Labels</div>
            <div class="label-list">
              <el-tag v-for="(v, k) in (daemonSet?.metadata?.labels || {})" :key="k" size="small" type="info" class="label-tag">{{ k }}={{ v }}</el-tag>
              <span v-if="!daemonSet?.metadata?.labels || Object.keys(daemonSet.metadata.labels).length === 0" class="info-empty">无</span>
            </div>
          </div>

          <!-- 节点分布 -->
          <div v-show="leftView === 'nodes'" class="node-dist-body" v-loading="nodesLoading">
            <!-- 汇总条 -->
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
                <div
                  class="node-card-bar"
                  :class="nodeBarClass(node)"
                />
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
      resource-type="daemonset"
      :namespace="namespace"
      :name="name"
      @saved="handleYamlSaved"
    />

    <!-- Image Update Dialog -->
    <el-dialog v-model="imageDialogVisible" title="更新镜像" width="520px" destroy-on-close>
      <div>
        <p style="margin-bottom: 16px;">更新 <strong>{{ name }}</strong> 的容器镜像</p>
        <el-form label-width="80px">
          <el-form-item label="容器">
            <el-select v-model="imageForm.containerName" style="width: 100%;">
              <el-option
                v-for="container in daemonSet?.spec?.template?.spec?.containers || []"
                :key="container.name"
                :label="container.name"
                :value="container.name"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="镜像">
            <el-input v-model="imageForm.image" placeholder="例如: nginx:1.25" />
          </el-form-item>
        </el-form>
        <el-alert
          v-if="imageForm.containerName"
          :title="`当前镜像: ${daemonSet?.spec?.template?.spec?.containers?.find((c: any) => c.name === imageForm.containerName)?.image || '-'}`"
          type="info"
          :closable="false"
          style="margin-top: 8px;"
        />
      </div>
      <template #footer>
        <el-button @click="imageDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="imageLoading" @click="handleUpdateImageConfirm">确认更新</el-button>
      </template>
    </el-dialog>

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
      <div style="height: calc(100vh - 52px); overflow-y: auto;">
        <DaemonSetForm
          v-if="editDialogVisible && daemonSet"
          :is-edit="true"
          :initial-data="daemonSet"
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
  gap: 4px;
  flex-shrink: 0;
}

.node-stat {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 6px 4px;
  background: var(--el-fill-color-lighter);
  border-radius: 6px;
}

.node-stat-num {
  font-size: 18px;
  font-weight: 700;
  line-height: 1.2;
  color: var(--el-text-color-primary);
}

.node-stat-label {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  margin-top: 2px;
}

.node-cards {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.node-card {
  display: flex;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
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

.bar-success {
  background: var(--el-color-success);
}

.bar-warning {
  background: var(--el-color-warning);
}

.bar-danger {
  background: var(--el-color-danger);
}

.bar-empty {
  background: var(--el-fill-color);
}

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
  color: var(--el-text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.node-card-ip {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  flex-shrink: 0;
}

.node-card-empty {
  font-size: 11px;
  color: var(--el-text-color-placeholder);
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
  font-family: monospace;
  color: var(--el-text-color-regular);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mono {
  font-family: monospace;
  font-size: 12px;
}

.info-sub {
  font-size: 11px;
  color: var(--el-text-color-secondary);
}

.info-section-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--el-text-color-regular);
  margin: 12px 0 6px;
}

.vct-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.vct-item {
  font-size: 12px;
  display: flex;
  justify-content: space-between;
  gap: 8px;
  padding: 4px 6px;
  background: var(--el-fill-color-lighter);
  border-radius: 4px;
}

.vct-name {
  font-family: monospace;
  color: var(--el-text-color-primary);
}

.vct-meta {
  color: var(--el-text-color-secondary);
  font-size: 11px;
}

.label-list {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.label-tag {
  font-family: monospace;
}

.info-empty {
  font-size: 12px;
  color: var(--el-text-color-placeholder);
}

.conditions-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.condition-item {
  padding: 6px 8px;
  background: var(--el-fill-color-lighter);
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
  color: var(--el-text-color-primary);
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
  color: var(--el-text-color-secondary);
  word-break: break-all;
}

.condition-time {
  font-size: 10px;
  color: var(--el-text-color-placeholder);
  margin-top: 2px;
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
