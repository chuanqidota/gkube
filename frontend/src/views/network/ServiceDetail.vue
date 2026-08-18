<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getServiceDetail,
  deleteService,
  getServiceEvents,
  getServicePods,
  getServiceEndpoints,
  deletePod,
} from '@/api/resource'
import { Refresh, Timer, ArrowLeft, FullScreen, Aim } from '@element-plus/icons-vue'
import YamlDrawer from '@/components/YamlDrawer.vue'
import PodListPanel from '@/components/PodListPanel.vue'
import ServiceForm from './components/ServiceForm.vue'
import { useClusterStore } from '@/stores/cluster'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import { useResizable } from '@/composables/useResizable'

const clusterStore = useClusterStore()

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const serviceRaw = ref<any>(null)
const yamlDialogVisible = ref(false)

// Events
const events = ref<any[]>([])
const eventsLoading = ref(false)

// Related Pods
const pods = ref<any[]>([])
const podsLoading = ref(false)

// Endpoints
const endpoints = ref<any[]>([])
const endpointsLoading = ref(false)

// Edit dialog
const editDialogVisible = ref(false)
const editFullscreen = ref(false)

// Left panel tab
const leftTab = ref<'info' | 'endpoint'>('info')

const namespace = route.params.namespace as string
const name = route.params.name as string

// Transformed display data
const service = computed(() => {
  const raw = serviceRaw.value
  if (!raw) return null
  const spec = raw.spec || {}
  const meta = raw.metadata || {}
  const status = raw.status || {}

  // Structured ports for the mapping table
  const portList = (spec.ports || []).map((p: any) => ({
    name: p.name || '',
    port: p.port,
    targetPort: p.targetPort || p.port,
    protocol: p.protocol || 'TCP',
    nodePort: p.nodePort || null,
  }))

  const ports = portList
    .map((p: any) => `${p.port}${p.nodePort ? ':' + p.nodePort : ''}/${p.protocol}`)
    .join(', ')

  let externalIP = ''
  const lbIngress = status.loadBalancer?.ingress
  if (lbIngress && lbIngress.length > 0) {
    externalIP = lbIngress.map((i: any) => i.ip || i.hostname || '').filter(Boolean).join(', ')
  }

  return {
    name: meta.name || '',
    namespace: meta.namespace || '',
    type: spec.type || 'ClusterIP',
    clusterIP: spec.clusterIP || '',
    externalIP,
    ports,
    portList,
    sessionAffinity: spec.sessionAffinity || 'None',
    selector: spec.selector || {},
    labels: meta.labels || {},
  }
})

// Whether to show NodePort column in port mapping table
const showNodePort = computed(() => {
  const t = service.value?.type
  return t === 'NodePort' || t === 'LoadBalancer'
})

const statusTagType = computed(() => {
  return service.value?.type === 'LoadBalancer' ? 'success' : 'info'
})

// Endpoint flat list for the table
const endpointRows = computed(() => {
  const rows: any[] = []
  for (const subset of endpoints.value) {
    for (const addr of subset.addresses || []) {
      rows.push({
        ip: addr.ip,
        port: subset.ports?.map((p: any) => p.port).join(', ') || '-',
        protocol: subset.ports?.map((p: any) => p.protocol).join(', ') || 'TCP',
        podName: addr.pod_name || '-',
        nodeName: addr.node_name || '-',
        ready: true,
      })
    }
    for (const addr of subset.not_ready_addresses || []) {
      rows.push({
        ip: addr.ip,
        port: subset.ports?.map((p: any) => p.port).join(', ') || '-',
        protocol: subset.ports?.map((p: any) => p.protocol).join(', ') || 'TCP',
        podName: addr.pod_name || '-',
        nodeName: addr.node_name || '-',
        ready: false,
      })
    }
  }
  return rows
})

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getServiceDetail({ namespace, name })
    serviceRaw.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || '加载 Service 详情失败')
  } finally {
    loading.value = false
  }
}

async function fetchEvents() {
  eventsLoading.value = true
  try {
    const res: any = await getServiceEvents({ namespace, name })
    events.value = res.data || []
  } catch (e) {
    events.value = []
  } finally {
    eventsLoading.value = false
  }
}

async function fetchPods() {
  podsLoading.value = true
  try {
    const res: any = await getServicePods({ namespace, name })
    pods.value = res.data?.items || res.data || []
  } catch (e) {
    pods.value = []
  } finally {
    podsLoading.value = false
  }
}

async function fetchEndpoints() {
  endpointsLoading.value = true
  try {
    const res: any = await getServiceEndpoints({ namespace, name })
    endpoints.value = res.data || []
  } catch (e) {
    endpoints.value = []
  } finally {
    endpointsLoading.value = false
  }
}

function getClusterName(): string {
  return clusterStore.currentCluster?.clusterName || clusterStore.currentCluster?.cluster_name || clusterStore.currentCluster?.name || ''
}

function handlePodLogs(pod: any) {
  const cluster = getClusterName()
  window.open(`/fullscreen/logs?namespace=${pod.metadata?.namespace || namespace}&pod=${pod.metadata?.name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
}

function handlePodExec(pod: any) {
  const cluster = getClusterName()
  window.open(`/fullscreen/terminal?namespace=${pod.metadata?.namespace || namespace}&pod=${pod.metadata?.name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
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
      await deletePod({ namespace, name: pod.metadata.name, force: true })
      ElMessage.success('Pod 已强制删除')
      fetchPods()
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
    await deletePod({ namespace, name: pod.metadata.name })
    ElMessage.success('Pod 已删除')
    fetchPods()
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
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(
      `确定要删除 Service "${name}" 吗？此操作不可恢复。`,
      '确认删除',
      { type: 'error', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await deleteService({ namespace, name })
    ElMessage.success('Service 已删除')
    router.push('/network/services')
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
  fetchPods()
  fetchEndpoints()
}

function handleEditCancel() {
  editDialogVisible.value = false
}

// ---- Resize: left-right + top-bottom (Pods / Events) ----
const { leftWidth, rightTopHeight, resizingH, resizingV, onHResizeStart, onVResizeStart } = useResizable({ initialWidth: 320 })

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(async () => {
  fetchDetail()
  fetchPods()
  fetchEvents()
  fetchEndpoints()
}, { autoStart: false })

onMounted(() => {
  fetchDetail()
  fetchPods()
  fetchEvents()
  fetchEndpoints()
})
</script>

<template>
  <div class="detail-page" v-loading="loading">

    <!-- 顶部标题栏 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="res-name">{{ name }}</h2>
        <div class="meta-line">
          <el-tag :type="statusTagType" effect="dark" size="small">{{ service?.type || '-' }}</el-tag>
          <span class="ns-tag">ns/{{ namespace }}</span>
          <span class="replicas-info" v-if="service?.clusterIP">
            Cluster IP: {{ service.clusterIP }}
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
          <el-button :icon="ArrowLeft" @click="router.push('/network/services')" />
        </el-tooltip>
      </div>
    </div>

    <template v-if="service">
      <div class="main-layout" :class="{ 'is-resizing': resizingH || resizingV }">

        <!-- 左侧面板 -->
        <div class="left-panel" :style="{ width: leftWidth + 'px', minWidth: leftWidth + 'px' }">
          <!-- 切换标签 -->
          <div class="left-tabs">
            <el-segmented
              v-model="leftTab"
              :options="[
                { label: '基本信息', value: 'info' },
                { label: 'Endpoint', value: 'endpoint' },
              ]"
              size="small"
            />
          </div>

          <!-- 信息视图 -->
          <div v-show="leftTab === 'info'" class="left-content">
            <el-descriptions :column="1" border size="small">
              <el-descriptions-item label="名称">{{ service.name }}</el-descriptions-item>
              <el-descriptions-item label="命名空间">{{ service.namespace }}</el-descriptions-item>
              <el-descriptions-item label="类型">{{ service.type || '-' }}</el-descriptions-item>
              <el-descriptions-item label="Cluster IP">{{ service.clusterIP || '-' }}</el-descriptions-item>
              <el-descriptions-item label="External IP">{{ service.externalIP || '-' }}</el-descriptions-item>
              <el-descriptions-item label="Session Affinity">{{ service.sessionAffinity || '-' }}</el-descriptions-item>
            </el-descriptions>

            <!-- 端口映射表 -->
            <div v-if="service.portList && service.portList.length > 0" style="margin-top: 16px;">
              <h4 style="margin: 0 0 8px; font-size: 13px;">端口映射</h4>
              <el-table :data="service.portList" size="small" border stripe>
                <el-table-column prop="name" label="名称" width="80">
                  <template #default="{ row }">{{ row.name || '-' }}</template>
                </el-table-column>
                <el-table-column prop="port" label="Port" width="70" align="center" />
                <el-table-column label="→" width="30" align="center">
                  <template #default><span style="color: var(--el-text-color-placeholder);">→</span></template>
                </el-table-column>
                <el-table-column prop="targetPort" label="TargetPort" width="90" align="center" />
                <el-table-column prop="protocol" label="协议" width="70" align="center" />
                <el-table-column v-if="showNodePort" prop="nodePort" label="NodePort" width="90" align="center">
                  <template #default="{ row }">
                    <el-tag v-if="row.nodePort" size="small" type="warning">{{ row.nodePort }}</el-tag>
                    <span v-else>-</span>
                  </template>
                </el-table-column>
              </el-table>
            </div>

            <!-- Selector -->
            <div v-if="service.selector && Object.keys(service.selector).length > 0" style="margin-top: 16px;">
              <h4 style="margin: 0 0 8px; font-size: 13px;">Selector</h4>
              <el-tag
                v-for="(val, key) in service.selector"
                :key="key"
                style="margin-right: 8px; margin-bottom: 8px;"
                type="info"
                size="small"
              >
                {{ key }}={{ val }}
              </el-tag>
            </div>

            <!-- Labels -->
            <div v-if="service.labels && Object.keys(service.labels).length > 0" style="margin-top: 16px;">
              <h4 style="margin: 0 0 8px; font-size: 13px;">Labels</h4>
              <el-tag
                v-for="(val, key) in service.labels"
                :key="key"
                style="margin-right: 8px; margin-bottom: 8px;"
                size="small"
              >
                {{ key }}={{ val }}
              </el-tag>
            </div>
          </div>

          <!-- Endpoint 视图 -->
          <div v-show="leftTab === 'endpoint'" class="left-content endpoint-tab">
            <div v-loading="endpointsLoading" class="endpoint-table-wrapper">
              <el-table v-if="endpointRows.length > 0" :data="endpointRows" size="small" stripe>
                <el-table-column prop="ip" label="IP" width="130" />
                <el-table-column prop="port" label="Port" width="70" align="center" />
                <el-table-column prop="protocol" label="协议" width="65" align="center" />
                <el-table-column prop="podName" label="Pod" min-width="140" show-overflow-tooltip />
                <el-table-column prop="nodeName" label="Node" min-width="100" show-overflow-tooltip />
                <el-table-column label="状态" width="75" align="center">
                  <template #default="{ row }">
                    <el-tag :type="row.ready ? 'success' : 'warning'" size="small">
                      {{ row.ready ? 'Ready' : 'NotReady' }}
                    </el-tag>
                  </template>
                </el-table-column>
              </el-table>
              <div v-else-if="!endpointsLoading" class="empty-hint">暂无 Endpoint</div>
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

    <!-- YAML Drawer -->
    <YamlDrawer
      v-model="yamlDialogVisible"
      resource-type="service"
      :namespace="namespace"
      :name="name"
      @saved="handleYamlSaved"
    />

    <!-- Edit Drawer -->
    <el-drawer
      v-model="editDialogVisible"
      title="编辑 Service"
      :size="editFullscreen ? '100%' : '85%'"
      direction="rtl"
      :destroy-on-close="true"
      :body-style="{ padding: '0', height: '100%' }"
    >
      <template #header>
        <div class="drawer-header">
          <span class="drawer-title">编辑 Service</span>
          <el-tooltip :content="editFullscreen ? '退出全屏' : '全屏'" placement="top">
            <el-icon class="fullscreen-btn" @click="editFullscreen = !editFullscreen">
              <FullScreen v-if="!editFullscreen" />
              <Aim v-else />
            </el-icon>
          </el-tooltip>
        </div>
      </template>
      <div style="height: calc(100vh - 52px); overflow-y: auto;">
        <ServiceForm
          v-if="editDialogVisible && serviceRaw"
          :is-edit="true"
          :initial-data="serviceRaw"
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
  font-family: monospace;
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

.left-tabs {
  padding: 8px 14px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  flex-shrink: 0;
}

.left-tabs :deep(.el-segmented) {
  width: 100%;
}

.left-content {
  flex: 1;
  overflow-y: auto;
  padding: 14px;
  min-height: 0;
}

.endpoint-tab {
  padding: 0;
  display: flex;
  flex-direction: column;
}

.endpoint-table-wrapper {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
  padding: 14px;
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
