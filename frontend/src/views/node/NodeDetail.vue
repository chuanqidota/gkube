<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Cpu, Coin, Grid, Files, Search, Refresh, Timer, ArrowLeft } from '@element-plus/icons-vue'
import {
  getNodeDetail, getNodePods, getNodeEvents,
  type NodeDetail as NodeDetailType, type K8sPod, type NodeEvent,
} from '@/api/resource'
import { formatAge } from '@/utils/helpers'
import { formatK8sCPU, formatK8sMemory } from '@/utils/resource'
import { formatDateTime } from '@/utils/time'
import YamlDrawer from '@/components/YamlDrawer.vue'
import NodeTaintDialog from '@/components/node/NodeTaintDialog.vue'
import NodeLabelDialog from '@/components/node/NodeLabelDialog.vue'
import NodeDrainDialog from '@/components/node/NodeDrainDialog.vue'
import { useNodeActions } from '@/composables/useNodeActions'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import { useResizable } from '@/composables/useResizable'

interface PodRow {
  name: string
  namespace: string
  status: string
  ip: string
  restarts: number
  age: string
}

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const node = ref<NodeDetailType | null>(null)
const pods = ref<PodRow[]>([])
const podsLoading = ref(false)
const podSearch = ref('')

// 响应式 nodeName：路由参数变化时（同组件复用）自动重载，避免显示旧节点
const nodeName = computed(() => route.params.name as string)

const filteredPods = computed(() => {
  if (!podSearch.value) return pods.value
  const keyword = podSearch.value.toLowerCase()
  return pods.value.filter(pod =>
    pod.name?.toLowerCase().includes(keyword) ||
    pod.namespace?.toLowerCase().includes(keyword) ||
    pod.ip?.toLowerCase().includes(keyword),
  )
})

const events = ref<NodeEvent[]>([])
const eventsLoading = ref(false)
const yamlDialogVisible = ref(false)

const taintDialog = ref<InstanceType<typeof NodeTaintDialog>>()
const labelDialog = ref<InstanceType<typeof NodeLabelDialog>>()
const drainDialog = ref<InstanceType<typeof NodeDrainDialog>>()

const statusTagType = computed(() => {
  if (node.value?.status === 'Ready') return 'success'
  if (node.value?.status === 'NotReady') return 'danger'
  return 'warning'
})

const statusText = computed(() => node.value?.status || 'Unknown')

// silent=true 时不触发页面级遮罩，用于自动刷新，避免每轮整页转圈
async function fetchDetail(silent = false) {
  if (!silent) loading.value = true
  try {
    const res = await getNodeDetail({ name: nodeName.value })
    node.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || '加载节点详情失败')
  } finally {
    loading.value = false
  }
}

async function fetchPods(silent = false) {
  if (!silent) podsLoading.value = true
  try {
    const res = await getNodePods({ name: nodeName.value })
    const rawPods: K8sPod[] = res.data || []
    pods.value = rawPods.map((pod) => {
      const restarts = (pod.status.containerStatuses || []).reduce(
        (sum, cs) => sum + (cs.restartCount || 0), 0,
      )
      const ready = pod.status.conditions?.find(c => c.type === 'Ready')?.status === 'True'
      return {
        name: pod.name,
        namespace: pod.namespace,
        status: pod.status.phase || (ready ? 'Running' : 'Pending'),
        ip: pod.status.podIP || '-',
        restarts,
        age: formatAge(pod.creationTimestamp, false),
      }
    })
  } catch (e: any) {
    ElMessage.error(e?.message || '加载 Pod 列表失败')
  } finally {
    podsLoading.value = false
  }
}

async function fetchEvents(silent = false) {
  if (!silent) eventsLoading.value = true
  try {
    const res = await getNodeEvents({ name: nodeName.value })
    events.value = res.data || []
  } catch (e: any) {
    ElMessage.error(e?.message || '加载事件列表失败')
  } finally {
    eventsLoading.value = false
  }
}

function handleOpenYaml() { yamlDialogVisible.value = true }
function handleYamlSaved() { fetchDetail() }

function handleTaints() { taintDialog.value?.open(nodeName.value, node.value?.taints || []) }
function handleLabels() { labelDialog.value?.open(nodeName.value, node.value?.labels || {}) }
function handleDrain() { drainDialog.value?.open(nodeName.value) }

const { handleCordon, handleDelete } = useNodeActions(() => fetchDetail())

function handlePodDetail(row: PodRow) {
  router.push(`/workloads/pods/${row.namespace}/${row.name}`)
}

function podStatusType(status: string) {
  const s = (status || '').toLowerCase()
  if (s === 'running') return 'success'
  if (s === 'succeeded') return 'info'
  if (s === 'pending') return 'warning'
  if (s === 'failed') return 'danger'
  return 'info'
}

// CPU/内存数量格式化委托到 utils/resource（与全站共享，避免各视图重复实现）
const formatCPU = formatK8sCPU
const formatMemory = formatK8sMemory

function formatCapacity(val: string | undefined): string {
  if (!val) return '-'
  return String(val)
}

// ---- Resize: left-right + top-bottom ----
const { leftWidth, rightTopHeight, resizingH, resizingV, onHResizeStart, onVResizeStart } = useResizable({ initialWidth: 300 })

// 自动刷新走 silent 路径（不触发整页遮罩）；手动刷新显示遮罩；详情/Pods/事件三者独立拉取
const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(
  async () => {
    fetchDetail(true)
    fetchPods(true)
    fetchEvents(true)
  },
  {
    autoStart: false,
    manualFetch: async () => {
      fetchDetail(false)
      fetchPods(false)
      fetchEvents(false)
    },
  },
)

onMounted(() => {
  fetchDetail()
  fetchPods()
  fetchEvents()
})

// 路由参数变化时（同组件复用，如侧栏切到另一个节点）重新拉取，避免显示旧节点
watch(nodeName, () => {
  fetchDetail()
  fetchPods()
  fetchEvents()
})
</script>

<template>
  <div class="detail-page" v-loading="loading">

    <!-- 顶部标题栏 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="res-name">{{ nodeName }}</h2>
        <div class="meta-line">
          <el-tag :type="statusTagType" effect="dark" size="small">{{ statusText }}</el-tag>
          <span v-if="node?.roles" class="role-tag">{{ node.roles }}</span>
          <el-tag v-if="node?.unschedulable" type="warning" size="small" effect="plain">不可调度</el-tag>
          <span v-if="node?.internal_ip" class="info-text">{{ node.internal_ip }}</span>
        </div>
      </div>
      <div class="header-actions">
        <el-button-group>
          <el-button :type="node?.unschedulable ? 'success' : 'warning'" @click="handleCordon(nodeName, node?.unschedulable || false)">
            {{ node?.unschedulable ? '解除封锁' : '封锁' }}
          </el-button>
          <el-button type="primary" @click="handleTaints">污点</el-button>
          <el-button type="info" @click="handleLabels">标签</el-button>
          <el-button @click="handleOpenYaml">YAML</el-button>
          <el-button type="danger" @click="handleDrain">驱逐</el-button>
          <el-tooltip v-if="node?.status === 'Ready'" content="节点在线，删除后会重新注册（需先停止 kubelet）" placement="top">
            <span><el-button type="danger" plain disabled>删除</el-button></span>
          </el-tooltip>
          <el-button v-else type="danger" plain @click="handleDelete(nodeName, node?.status === 'Ready', () => router.push('/nodes'))">删除</el-button>
        </el-button-group>
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
          <el-button :icon="ArrowLeft" @click="router.push('/nodes')" />
        </el-tooltip>
      </div>
    </div>

    <template v-if="node">
      <div class="main-layout" :class="{ 'is-resizing': resizingH || resizingV }">

        <!-- 左侧：基本信息 -->
        <div class="left-panel" :style="{ width: leftWidth + 'px', minWidth: leftWidth + 'px' }">
          <div class="panel-title">基本信息</div>
          <div class="info-body">
            <el-descriptions :column="1" border size="small">
              <el-descriptions-item label="名称">{{ node.name }}</el-descriptions-item>
              <el-descriptions-item label="状态">
                <el-tag :type="statusTagType" size="small">{{ node.status || 'Unknown' }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="角色">{{ node.roles || '-' }}</el-descriptions-item>
              <el-descriptions-item label="Kubelet 版本">{{ node.version || '-' }}</el-descriptions-item>
              <el-descriptions-item label="操作系统">{{ node.os || '-' }}</el-descriptions-item>
              <el-descriptions-item label="内核版本">{{ node.kernel || '-' }}</el-descriptions-item>
              <el-descriptions-item label="容器运行时">{{ node.container_runtime || '-' }}</el-descriptions-item>
              <el-descriptions-item label="内部 IP">{{ node.internal_ip || '-' }}</el-descriptions-item>
              <el-descriptions-item label="外部 IP">{{ node.external_ip || '-' }}</el-descriptions-item>
              <el-descriptions-item label="主机名">{{ node.hostname || '-' }}</el-descriptions-item>
              <el-descriptions-item label="架构">{{ node.architecture || '-' }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ node.creationTimestamp ? formatDateTime(node.creationTimestamp) : '-' }}</el-descriptions-item>
              <el-descriptions-item label="不可调度">
                <el-tag :type="node.unschedulable ? 'danger' : 'success'" size="small">{{ node.unschedulable ? '是' : '否' }}</el-tag>
              </el-descriptions-item>
            </el-descriptions>

            <!-- Labels -->
            <div v-if="node.labels && Object.keys(node.labels).length > 0" style="margin-top: 16px;">
              <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px;">
                <h4 style="margin: 0; font-size: 13px;">Labels</h4>
                <el-button size="small" @click="handleLabels">编辑</el-button>
              </div>
              <el-tag
                v-for="(val, key) in node.labels"
                :key="key"
                style="margin-right: 8px; margin-bottom: 8px;"
                size="small"
              >
                {{ key }}={{ val }}
              </el-tag>
            </div>

            <!-- Taints -->
            <div style="margin-top: 16px;">
              <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px;">
                <h4 style="margin: 0; font-size: 13px;">Taints</h4>
                <el-button size="small" @click="handleTaints">编辑</el-button>
              </div>
              <el-table v-if="node.taints && node.taints.length > 0" :data="node.taints" size="small" border>
                <el-table-column prop="key" label="Key" min-width="150" />
                <el-table-column prop="value" label="Value" min-width="100" />
                <el-table-column prop="effect" label="Effect" min-width="120" />
              </el-table>
              <span v-else style="color: #909399; font-size: 12px;">无污点</span>
            </div>

            <!-- Resource Capacity -->
            <div v-if="node.capacity || node.allocatable" style="margin-top: 16px;">
              <h4 style="margin: 0 0 12px; font-size: 13px;">资源容量</h4>
              <div class="resource-cards">
                <div class="resource-card">
                  <div class="resource-icon cpu-icon"><el-icon><Cpu /></el-icon></div>
                  <div class="resource-info">
                    <div class="resource-label">CPU</div>
                    <div class="resource-values">
                      <div class="value-item">
                        <span class="value-label">总容量</span>
                        <span class="value-number">{{ formatCPU(node.capacity?.cpu) }}</span>
                      </div>
                      <div class="value-item">
                        <span class="value-label">可分配</span>
                        <span class="value-number highlight">{{ formatCPU(node.allocatable?.cpu) }}</span>
                      </div>
                    </div>
                  </div>
                </div>

                <div class="resource-card">
                  <div class="resource-icon memory-icon"><el-icon><Coin /></el-icon></div>
                  <div class="resource-info">
                    <div class="resource-label">内存</div>
                    <div class="resource-values">
                      <div class="value-item">
                        <span class="value-label">总容量</span>
                        <span class="value-number">{{ formatMemory(node.capacity?.memory) }}</span>
                      </div>
                      <div class="value-item">
                        <span class="value-label">可分配</span>
                        <span class="value-number highlight">{{ formatMemory(node.allocatable?.memory) }}</span>
                      </div>
                    </div>
                  </div>
                </div>

                <div class="resource-card">
                  <div class="resource-icon pods-icon"><el-icon><Grid /></el-icon></div>
                  <div class="resource-info">
                    <div class="resource-label">Pod 数量</div>
                    <div class="resource-values">
                      <div class="value-item">
                        <span class="value-label">总容量</span>
                        <span class="value-number">{{ formatCapacity(node.capacity?.pods) }}</span>
                      </div>
                      <div class="value-item">
                        <span class="value-label">可分配</span>
                        <span class="value-number highlight">{{ formatCapacity(node.allocatable?.pods) }}</span>
                      </div>
                    </div>
                  </div>
                </div>

                <div class="resource-card">
                  <div class="resource-icon storage-icon"><el-icon><Files /></el-icon></div>
                  <div class="resource-info">
                    <div class="resource-label">临时存储</div>
                    <div class="resource-values">
                      <div class="value-item">
                        <span class="value-label">总容量</span>
                        <span class="value-number">{{ formatMemory(node.capacity?.['ephemeral-storage']) }}</span>
                      </div>
                      <div class="value-item">
                        <span class="value-label">可分配</span>
                        <span class="value-number highlight">{{ formatMemory(node.allocatable?.['ephemeral-storage']) }}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Conditions -->
            <div v-if="node.conditions && node.conditions.length > 0" style="margin-top: 16px;">
              <h4 style="margin: 0 0 8px; font-size: 13px;">节点状态</h4>
              <el-table :data="node.conditions" size="small" border>
                <el-table-column prop="type" label="类型" width="120" />
                <el-table-column label="状态" width="80">
                  <template #default="{ row }"><el-tag :type="row.status === 'True' ? 'success' : 'danger'" size="small">{{ row.status }}</el-tag></template>
                </el-table-column>
                <el-table-column prop="reason" label="原因" width="150" />
                <el-table-column prop="message" label="消息" min-width="200" show-overflow-tooltip />
                <el-table-column prop="lastTransitionTime" label="最后变更" width="150" />
              </el-table>
            </div>
          </div>
        </div>

        <!-- 右侧：Pods + Events -->
        <div class="right-panel">

          <!-- Pod 列表 -->
          <div class="right-section" :style="rightTopHeight ? { flex: 'none', height: rightTopHeight + 'px' } : {}">
            <div class="panel-title">
              Pods
              <span class="count-badge">{{ pods.length }} 个</span>
              <el-input v-model="podSearch" placeholder="搜索" size="small" style="width: 200px; margin-left: auto;" clearable>
                <template #prefix><el-icon><Search /></el-icon></template>
              </el-input>
            </div>
            <div v-loading="podsLoading" class="pods-body">
              <el-table v-if="filteredPods.length > 0" :data="filteredPods" size="small" stripe>
                <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip>
                  <template #default="{ row }"><el-button link type="primary" @click="handlePodDetail(row)">{{ row.name }}</el-button></template>
                </el-table-column>
                <el-table-column prop="namespace" label="命名空间" width="140" />
                <el-table-column prop="status" label="状态" width="120"><template #default="{ row }"><el-tag :type="podStatusType(row.status)" size="small">{{ row.status }}</el-tag></template></el-table-column>
                <el-table-column prop="ip" label="IP" width="140" />
                <el-table-column prop="restarts" label="重启次数" width="100" />
                <el-table-column prop="age" label="年龄" width="120" />
              </el-table>
              <div v-else class="empty-hint">该节点上暂无 Pod</div>
            </div>
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
      resource-type="node"
      :name="nodeName"
      @saved="handleYamlSaved"
    />

    <!-- 共享对话框 -->
    <NodeTaintDialog ref="taintDialog" @saved="() => fetchDetail()" />
    <NodeLabelDialog ref="labelDialog" @saved="() => fetchDetail()" />
    <NodeDrainDialog ref="drainDialog" @saved="() => fetchDetail()" />
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

.role-tag {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color-lighter);
  padding: 1px 6px;
  border-radius: 4px;
}

.info-text {
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

.pods-body {
  flex: 1;
  overflow-y: auto;
  padding: 0;
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

/* Resource Capacity Cards */
.resource-cards {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.resource-card {
  display: flex;
  align-items: center;
  padding: 14px;
  background: var(--el-fill-color-lighter);
  border-radius: 8px;
  border: 1px solid var(--el-border-color-lighter);
}

.resource-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 12px;
  font-size: 20px;
  color: #fff;
  flex-shrink: 0;
}

.cpu-icon { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
.memory-icon { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); }
.pods-icon { background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); }
.storage-icon { background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%); }

.resource-info {
  flex: 1;
  min-width: 0;
}

.resource-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--el-text-color-regular);
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.resource-values {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.value-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.value-label {
  font-size: 11px;
  color: var(--el-text-color-secondary);
}

.value-number {
  font-size: 14px;
  font-weight: 700;
  color: var(--el-text-color-primary);
  font-variant-numeric: tabular-nums;
}

.value-number.highlight {
  color: var(--el-color-primary);
  font-size: 15px;
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
