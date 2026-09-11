<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Cpu, Coin, Grid, Files, Search, Refresh, Timer, ArrowLeft } from '@element-plus/icons-vue'
import {
  getNodeDetail,
  getNodePods,
  getNodeEvents,
  getNodeYaml,
  updateNodeYaml,
  type NodeDetail as NodeDetailType,
  type K8sPod,
  type NodeEvent,
} from '@/api/resource'
import { formatAge } from '@/utils/helpers'
import { formatK8sCPU, formatK8sMemory } from '@/utils/resource'
import { formatDateTime } from '@/utils/helpers'
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

const { t } = useI18n()
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
  return pods.value.filter(
    (pod) =>
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
    ElMessage.error(e?.message || t('node.loadNodeDetailFailed'))
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
        (sum, cs) => sum + (cs.restartCount || 0),
        0,
      )
      const ready = pod.status.conditions?.find((c) => c.type === 'Ready')?.status === 'True'
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
    ElMessage.error(e?.message || t('node.loadPodListFailed'))
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
    ElMessage.error(e?.message || t('node.loadEventsFailed'))
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

function handleTaints() {
  taintDialog.value?.open(nodeName.value, node.value?.taints || [])
}
function handleLabels() {
  labelDialog.value?.open(nodeName.value, node.value?.labels || {})
}
function handleDrain() {
  drainDialog.value?.open(nodeName.value)
}

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
const { leftWidth, rightTopHeight, resizingH, resizingV, onHResizeStart, onVResizeStart } =
  useResizable({ initialWidth: 300 })

// 自动刷新走 silent 路径（不触发整页遮罩）；手动刷新显示遮罩；详情/Pods/事件三者独立拉取
const {
  isRunning,
  countdown,
  currentInterval,
  availableIntervals,
  toggle,
  refresh: manualRefresh,
  setIntervalOption,
} = useAutoRefresh(
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
  <div v-loading="loading" class="detail-page">
    <!-- 顶部标题栏 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="res-name">{{ nodeName }}</h2>
        <div class="meta-line">
          <el-tag :type="statusTagType" effect="dark" size="small">{{ statusText }}</el-tag>
          <span v-if="node?.roles" class="role-tag">{{ node.roles }}</span>
          <el-tag v-if="node?.unschedulable" type="warning" size="small" effect="plain">{{
            t('node.unschedulable')
          }}</el-tag>
          <span v-if="node?.internal_ip" class="info-text">{{ node.internal_ip }}</span>
        </div>
      </div>
      <div class="header-actions">
        <el-button-group>
          <el-button
            :type="node?.unschedulable ? 'success' : 'warning'"
            @click="handleCordon(nodeName, node?.unschedulable || false)"
          >
            {{ node?.unschedulable ? t('node.uncordonButton') : t('node.cordonButton') }}
          </el-button>
          <el-button type="primary" @click="handleTaints">{{ t('node.taintButton') }}</el-button>
          <el-button type="info" @click="handleLabels">{{ t('node.labelButton') }}</el-button>
          <el-button @click="handleOpenYaml">YAML</el-button>
          <el-button type="danger" @click="handleDrain">{{ t('node.drainButton') }}</el-button>
          <el-tooltip
            v-if="node?.status === 'Ready'"
            :content="t('node.deleteReadyWarning')"
            placement="top"
          >
            <span
              ><el-button type="danger" disabled>{{ t('node.deleteButton') }}</el-button></span
            >
          </el-tooltip>
          <el-button
            v-else
            type="danger"
            @click="handleDelete(nodeName, node?.status === 'Ready', () => router.push('/nodes'))"
            >{{ t('node.deleteButton') }}</el-button
          >
        </el-button-group>
        <div class="action-divider" />
        <el-popover placement="bottom" :width="200" trigger="click">
          <template #reference>
            <el-button :type="isRunning ? 'success' : 'default'" :icon="Timer" @click="toggle()" />
          </template>
          <div class="auto-refresh-popover">
            <div class="popover-title">
              {{ isRunning ? `${t('common.autoRefresh')} ${countdown}s` : t('common.autoRefresh') }}
            </div>
            <el-select
              :model-value="currentInterval / 1000"
              :teleported="false"
              size="small"
              style="width: 100%"
              @update:model-value="setIntervalOption"
            >
              <el-option
                v-for="sec in availableIntervals"
                :key="sec"
                :value="sec"
                :label="`${t('common.refreshInterval')}: ${sec}s`"
              />
            </el-select>
          </div>
        </el-popover>
        <el-tooltip :content="t('common.refresh')" placement="top">
          <el-button :loading="loading" :icon="Refresh" @click="manualRefresh()" />
        </el-tooltip>
        <el-tooltip :content="t('common.backToList')" placement="top">
          <el-button :icon="ArrowLeft" @click="router.push('/nodes')" />
        </el-tooltip>
      </div>
    </div>

    <template v-if="node">
      <div class="main-layout" :class="{ 'is-resizing': resizingH || resizingV }">
        <!-- 左侧：基本信息 -->
        <div class="left-panel" :style="{ width: leftWidth + 'px', minWidth: leftWidth + 'px' }">
          <div class="panel-title">{{ t('node.basicInfo') }}</div>
          <div class="info-body">
            <el-descriptions :column="1" border size="small">
              <el-descriptions-item :label="t('common.name')">{{ node.name }}</el-descriptions-item>
              <el-descriptions-item :label="t('common.status')">
                <el-tag :type="statusTagType" size="small">{{ node.status || 'Unknown' }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item :label="t('node.roles')">{{
                node.roles || '-'
              }}</el-descriptions-item>
              <el-descriptions-item :label="`Kubelet ${t('node.version')}`">{{
                node.version || '-'
              }}</el-descriptions-item>
              <el-descriptions-item :label="t('node.os')">{{
                node.os || '-'
              }}</el-descriptions-item>
              <el-descriptions-item :label="t('node.kernel')">{{
                node.kernel || '-'
              }}</el-descriptions-item>
              <el-descriptions-item :label="t('node.runtime')">{{
                node.container_runtime || '-'
              }}</el-descriptions-item>
              <el-descriptions-item :label="t('node.internalIp')">{{
                node.internal_ip || '-'
              }}</el-descriptions-item>
              <el-descriptions-item :label="t('node.externalIp')">{{
                node.external_ip || '-'
              }}</el-descriptions-item>
              <el-descriptions-item :label="t('node.hostname')">{{
                node.hostname || '-'
              }}</el-descriptions-item>
              <el-descriptions-item :label="t('node.architecture')">{{
                node.architecture || '-'
              }}</el-descriptions-item>
              <el-descriptions-item :label="t('common.age')">{{
                node.creationTimestamp ? formatDateTime(node.creationTimestamp) : '-'
              }}</el-descriptions-item>
              <el-descriptions-item :label="t('node.unschedulable')">
                <el-tag :type="node.unschedulable ? 'danger' : 'success'" size="small">{{
                  node.unschedulable ? t('node.unschedulableYes') : t('node.unschedulableNo')
                }}</el-tag>
              </el-descriptions-item>
            </el-descriptions>

            <!-- Labels -->
            <div
              v-if="node.labels && Object.keys(node.labels).length > 0"
              style="margin-top: var(--gk-space-4)"
            >
              <div
                style="
                  display: flex;
                  justify-content: space-between;
                  align-items: center;
                  margin-bottom: 8px;
                "
              >
                <h4 style="margin: 0; font-size: 13px">{{ t('node.labels') }}</h4>
                <el-button size="small" @click="handleLabels">{{ t('common.edit') }}</el-button>
              </div>
              <el-tag
                v-for="(val, key) in node.labels"
                :key="key"
                style="margin-right: 8px; margin-bottom: 8px"
                size="small"
              >
                {{ key }}={{ val }}
              </el-tag>
            </div>

            <!-- Taints -->
            <div style="margin-top: var(--gk-space-4)">
              <div
                style="
                  display: flex;
                  justify-content: space-between;
                  align-items: center;
                  margin-bottom: 8px;
                "
              >
                <h4 style="margin: 0; font-size: 13px">{{ t('node.taints') }}</h4>
                <el-button size="small" @click="handleTaints">{{ t('common.edit') }}</el-button>
              </div>
              <el-table
                v-if="node.taints && node.taints.length > 0"
                :data="node.taints"
                size="small"
                border
              >
                <el-table-column prop="key" label="Key" min-width="150" />
                <el-table-column prop="value" label="Value" min-width="100" />
                <el-table-column prop="effect" label="Effect" min-width="120" />
              </el-table>
              <span v-else style="color: #909399; font-size: 12px">{{ t('node.noTaints') }}</span>
            </div>

            <!-- Resource Capacity -->
            <div v-if="node.capacity || node.allocatable" style="margin-top: var(--gk-space-4)">
              <h4 style="margin: 0 0 12px; font-size: 13px">{{ t('node.resourceCapacity') }}</h4>
              <div class="resource-cards">
                <div class="resource-card">
                  <div class="resource-icon cpu-icon">
                    <el-icon><Cpu /></el-icon>
                  </div>
                  <div class="resource-info">
                    <div class="resource-label">CPU</div>
                    <div class="resource-values">
                      <div class="value-item">
                        <span class="value-label">{{ t('node.totalCapacity') }}</span>
                        <span class="value-number">{{ formatCPU(node.capacity?.cpu) }}</span>
                      </div>
                      <div class="value-item">
                        <span class="value-label">{{ t('node.allocatable') }}</span>
                        <span class="value-number highlight">{{
                          formatCPU(node.allocatable?.cpu)
                        }}</span>
                      </div>
                    </div>
                  </div>
                </div>

                <div class="resource-card">
                  <div class="resource-icon memory-icon">
                    <el-icon><Coin /></el-icon>
                  </div>
                  <div class="resource-info">
                    <div class="resource-label">{{ t('node.memory') }}</div>
                    <div class="resource-values">
                      <div class="value-item">
                        <span class="value-label">{{ t('node.totalCapacity') }}</span>
                        <span class="value-number">{{ formatMemory(node.capacity?.memory) }}</span>
                      </div>
                      <div class="value-item">
                        <span class="value-label">{{ t('node.allocatable') }}</span>
                        <span class="value-number highlight">{{
                          formatMemory(node.allocatable?.memory)
                        }}</span>
                      </div>
                    </div>
                  </div>
                </div>

                <div class="resource-card">
                  <div class="resource-icon pods-icon">
                    <el-icon><Grid /></el-icon>
                  </div>
                  <div class="resource-info">
                    <div class="resource-label">{{ t('node.podCount') }}</div>
                    <div class="resource-values">
                      <div class="value-item">
                        <span class="value-label">{{ t('node.totalCapacity') }}</span>
                        <span class="value-number">{{ formatCapacity(node.capacity?.pods) }}</span>
                      </div>
                      <div class="value-item">
                        <span class="value-label">{{ t('node.allocatable') }}</span>
                        <span class="value-number highlight">{{
                          formatCapacity(node.allocatable?.pods)
                        }}</span>
                      </div>
                    </div>
                  </div>
                </div>

                <div class="resource-card">
                  <div class="resource-icon storage-icon">
                    <el-icon><Files /></el-icon>
                  </div>
                  <div class="resource-info">
                    <div class="resource-label">{{ t('node.ephemeralStorage') }}</div>
                    <div class="resource-values">
                      <div class="value-item">
                        <span class="value-label">{{ t('node.totalCapacity') }}</span>
                        <span class="value-number">{{
                          formatMemory(node.capacity?.['ephemeral-storage'])
                        }}</span>
                      </div>
                      <div class="value-item">
                        <span class="value-label">{{ t('node.allocatable') }}</span>
                        <span class="value-number highlight">{{
                          formatMemory(node.allocatable?.['ephemeral-storage'])
                        }}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Conditions -->
            <div
              v-if="node.conditions && node.conditions.length > 0"
              style="margin-top: var(--gk-space-4)"
            >
              <h4 style="margin: 0 0 8px; font-size: 13px">{{ t('node.nodeConditions') }}</h4>
              <el-table :data="node.conditions" size="small" border>
                <el-table-column prop="type" :label="t('node.typeLabel')" width="120" />
                <el-table-column :label="t('node.statusLabel')" width="80">
                  <template #default="{ row }"
                    ><el-tag :type="row.status === 'True' ? 'success' : 'danger'" size="small">{{
                      row.status
                    }}</el-tag></template
                  >
                </el-table-column>
                <el-table-column prop="reason" :label="t('node.reasonLabel')" width="150" />
                <el-table-column
                  prop="message"
                  :label="t('node.messageLabel')"
                  min-width="200"
                  show-overflow-tooltip
                />
                <el-table-column
                  prop="lastTransitionTime"
                  :label="t('node.lastTransition')"
                  width="150"
                />
              </el-table>
            </div>
          </div>
        </div>

        <!-- 右侧：Pods + Events -->
        <div class="right-panel">
          <!-- Pod 列表 -->
          <div
            class="right-section"
            :style="rightTopHeight ? { flex: 'none', height: rightTopHeight + 'px' } : {}"
          >
            <div class="panel-title">
              Pods
              <span class="count-badge">{{ pods.length }} {{ t('node.countUnit') }}</span>
              <el-input
                v-model="podSearch"
                :placeholder="t('node.searchPods')"
                size="small"
                style="width: 200px; margin-left: auto"
                clearable
              >
                <template #prefix
                  ><el-icon><Search /></el-icon
                ></template>
              </el-input>
            </div>
            <div v-loading="podsLoading" class="pods-body">
              <el-table v-if="filteredPods.length > 0" :data="filteredPods" size="small" stripe>
                <el-table-column
                  prop="name"
                  :label="t('common.name')"
                  min-width="200"
                  show-overflow-tooltip
                >
                  <template #default="{ row }"
                    ><el-button link type="primary" @click="handlePodDetail(row)">{{
                      row.name
                    }}</el-button></template
                  >
                </el-table-column>
                <el-table-column prop="namespace" :label="t('node.namespaceLabel')" width="140" />
                <el-table-column prop="status" :label="t('common.status')" width="120"
                  ><template #default="{ row }"
                    ><el-tag :type="podStatusType(row.status)" size="small">{{
                      row.status
                    }}</el-tag></template
                  ></el-table-column
                >
                <el-table-column prop="ip" label="IP" width="140" />
                <el-table-column prop="restarts" :label="t('node.restartCount')" width="100" />
                <el-table-column prop="age" :label="t('node.ageLabel')" width="120" />
              </el-table>
              <div v-else class="empty-hint">{{ t('node.noPodsOnNode') }}</div>
            </div>
          </div>

          <!-- 垂直拖拽条 -->
          <div class="resize-handle-v" :class="{ active: resizingV }" @mousedown="onVResizeStart" />

          <!-- Events -->
          <div class="right-section events-section">
            <div class="panel-title">
              {{ t('node.events') }}
              <span class="count-badge">{{ events.length }} {{ t('node.eventUnit') }}</span>
            </div>
            <div v-loading="eventsLoading" class="events-body">
              <el-table
                v-if="events.length > 0"
                :data="events"
                size="small"
                stripe
                max-height="260"
              >
                <el-table-column prop="type" :label="t('node.typeLabel')" width="80">
                  <template #default="{ row }">
                    <el-tag :type="row.type === 'Warning' ? 'danger' : 'info'" size="small">{{
                      row.type
                    }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="reason" :label="t('node.reasonLabel')" width="130" />
                <el-table-column
                  prop="message"
                  :label="t('node.messageLabel')"
                  min-width="200"
                  show-overflow-tooltip
                />
                <el-table-column prop="last_seen" :label="t('event.lastSeen')" width="150" />
              </el-table>
              <div v-else class="empty-hint">{{ t('node.noEvents') }}</div>
            </div>
          </div>
        </div>

        <!-- 水平拖拽条 -->
        <div
          class="resize-handle-h"
          :class="{ active: resizingH }"
          :style="{ left: leftWidth - 3 + 'px' }"
          @mousedown="onHResizeStart"
        />
      </div>
    </template>

    <!-- YAML Drawer -->
    <YamlDrawer
      v-model="yamlDialogVisible"
      :get-yaml="getNodeYaml"
      :update-yaml="updateNodeYaml"
      :name="nodeName"
      title="Node YAML"
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
  border-radius: var(--gk-radius-sm) 0 0 var(--gk-radius-sm);
  margin-left: 0;
}

.header-actions .el-button:last-of-type {
  border-radius: 0 var(--gk-radius-sm) var(--gk-radius-sm) 0;
}

.action-divider {
  width: 1px;
  height: 20px;
  background: var(--gk-color-border-light);
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
  background: var(--el-fill-color-lighter);
  border-bottom: 1px solid var(--gk-color-border-light);
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
  font-size: var(--gk-font-size-sm);
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
  border-radius: var(--gk-radius-md);
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

.cpu-icon {
  background: linear-gradient(
    135deg,
    var(--gk-color-primary-light) 0%,
    var(--gk-color-primary-dark) 100%
  );
}
.memory-icon {
  background: linear-gradient(
    135deg,
    var(--gk-color-primary) 0%,
    var(--gk-color-primary-dark) 100%
  );
}
.pods-icon {
  background: linear-gradient(
    135deg,
    var(--gk-color-primary-light) 0%,
    var(--gk-color-primary) 100%
  );
}
.storage-icon {
  background: linear-gradient(
    135deg,
    var(--gk-color-primary) 0%,
    var(--gk-color-primary-light) 100%
  );
}

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
  gap: var(--gk-space-1);
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
