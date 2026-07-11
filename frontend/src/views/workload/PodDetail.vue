<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Timer, ArrowLeft } from '@element-plus/icons-vue'
import { getPodDetail, deletePod, getPodEvents, calcAge } from '@/api/resource'
import YamlDrawer from '@/components/YamlDrawer.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const pod = ref<any>(null)
const yamlDialogVisible = ref(false)

// Events
const events = ref<any[]>([])
const eventsLoading = ref(false)

const namespace = route.params.namespace as string
const name = route.params.name as string

/**
 * Transform raw K8s Pod object into flat display format.
 */
function transformPodDetail(raw: any): any {
  if (!raw) return null
  const restarts = (raw.status?.containerStatuses || []).reduce(
    (sum: number, cs: any) => sum + (cs.restartCount || 0), 0
  )
  const specContainers = raw.spec?.containers || []
  const statusContainers = raw.status?.containerStatuses || []
  const containers = specContainers.map((spec: any) => {
    const status = statusContainers.find((s: any) => s.name === spec.name) || {}
    let state = 'Unknown'
    let stateReason = ''
    let exitCode: number | undefined
    if (status.state?.running) state = 'Running'
    else if (status.state?.waiting) { state = 'Waiting'; stateReason = status.state.waiting.reason || '' }
    else if (status.state?.terminated) { state = 'Terminated'; exitCode = status.state.terminated.exitCode }
    return {
      name: spec.name,
      image: spec.image,
      ready: status.ready || false,
      restartCount: status.restartCount || 0,
      state,
      stateReason,
      exitCode,
      ports: spec.ports || [],
      env: spec.env || [],
      volumeMounts: spec.volumeMounts || [],
      livenessProbe: spec.livenessProbe,
      readinessProbe: spec.readinessProbe,
    }
  })
  return {
    name: raw.metadata?.name || '',
    namespace: raw.metadata?.namespace || '',
    status: raw.status?.phase || 'Unknown',
    ip: raw.status?.podIP || '',
    host_ip: raw.status?.hostIP || '',
    node: raw.spec?.nodeName || '',
    restarts,
    qos_class: raw.status?.qosClass || '',
    priority: raw.spec?.priority ?? null,
    age: calcAge(raw.metadata?.creationTimestamp),
    created_at: raw.metadata?.creationTimestamp || '',
    service_account: raw.spec?.serviceAccountName || '',
    labels: raw.metadata?.labels || {},
    annotations: raw.metadata?.annotations || {},
    conditions: (raw.status?.conditions || []).map((c: any) => ({
      type: c.type,
      status: c.status,
      reason: c.reason || '',
      message: c.message || '',
      last_transition_time: c.lastTransitionTime || '',
    })),
    containers,
  }
}

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

// ---- Resize: top-bottom (Containers / Events) ----
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

const statusTagType = computed(() => {
  const s = (pod.value?.status || '').toLowerCase()
  if (s === 'running') return 'success'
  if (s === 'succeeded') return 'info'
  if (s === 'pending') return 'warning'
  if (s === 'failed' || s === 'error') return 'danger'
  return 'info'
})

function statusType(status: string) {
  const s = (status || '').toLowerCase()
  if (s === 'running') return 'success'
  if (s === 'succeeded') return 'info'
  if (s === 'pending') return 'warning'
  if (s === 'failed' || s === 'error') return 'danger'
  return 'info'
}

function containerStateType(state: string) {
  const s = (state || '').toLowerCase()
  if (s === 'running') return 'success'
  if (s === 'waiting') return 'warning'
  if (s === 'terminated') return 'danger'
  return 'info'
}

function getContainerStateLabel(container: any): string {
  if (container.state === 'Running') return '运行中'
  if (container.state === 'Waiting') return `等待中 (${container.stateReason || '-'})`
  if (container.state === 'Terminated') return `已终止 (${container.exitCode ?? '-'})`
  return container.state || '-'
}

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getPodDetail({ namespace, name })
    pod.value = transformPodDetail(res.data)
  } catch (e: any) {
    ElMessage.error(e?.message || '加载 Pod 详情失败')
  } finally {
    loading.value = false
  }
}

async function fetchEvents() {
  eventsLoading.value = true
  try {
    const res: any = await getPodEvents({ namespace, name })
    events.value = res.data || []
  } catch {
    events.value = []
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

function handleLogs() {
  const cluster = getClusterName()
  window.open(`/fullscreen/logs?namespace=${namespace}&pod=${name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
}

function handleExec() {
  const cluster = getClusterName()
  window.open(`/fullscreen/terminal?namespace=${namespace}&pod=${name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(
      `确定要删除 Pod "${name}"（命名空间：${namespace}）吗？`,
      '确认删除',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await deletePod({ namespace, name })
    ElMessage.success('Pod 已删除')
    router.push('/workloads/pods')
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || '删除失败')
    }
  }
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(async () => {
  fetchDetail()
  fetchEvents()
}, { autoStart: false })

onMounted(() => {
  fetchDetail()
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
          <el-tag :type="statusTagType" effect="dark" size="small">{{ pod?.status || '-' }}</el-tag>
          <span class="ns-tag">ns/{{ namespace }}</span>
          <span class="replicas-info" v-if="pod">
            {{ pod.ip || '-' }}
          </span>
        </div>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="handleLogs">日志</el-button>
        <el-button type="success" @click="handleExec">终端</el-button>
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
          <el-button :icon="ArrowLeft" @click="router.push('/workloads/pods')" />
        </el-tooltip>
      </div>
    </div>

    <template v-if="pod">
      <div class="main-layout" :class="{ 'is-resizing': resizingH || resizingV }">

        <!-- 左侧：基本信息 -->
        <div class="left-panel" :style="{ width: leftWidth + 'px', minWidth: leftWidth + 'px' }">
          <div class="panel-title">基本信息</div>
          <div class="info-body">
            <el-descriptions :column="1" border size="small">
              <el-descriptions-item label="名称">{{ pod.name }}</el-descriptions-item>
              <el-descriptions-item label="命名空间">{{ pod.namespace }}</el-descriptions-item>
              <el-descriptions-item label="状态">
                <el-tag :type="statusType(pod.status)" size="small">{{ pod.status }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="Pod IP">{{ pod.ip || '-' }}</el-descriptions-item>
              <el-descriptions-item label="主机 IP">{{ pod.host_ip || '-' }}</el-descriptions-item>
              <el-descriptions-item label="节点">{{ pod.node || '-' }}</el-descriptions-item>
              <el-descriptions-item label="QoS 类别">{{ pod.qos_class || '-' }}</el-descriptions-item>
              <el-descriptions-item label="优先级">{{ pod.priority ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="服务账号">{{ pod.service_account || '-' }}</el-descriptions-item>
              <el-descriptions-item label="重启次数">{{ pod.restarts ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="年龄">{{ pod.age || '-' }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ pod.created_at || '-' }}</el-descriptions-item>
            </el-descriptions>

            <!-- Labels -->
            <div v-if="pod.labels && Object.keys(pod.labels).length > 0" style="margin-top: 16px;">
              <h4 style="margin: 0 0 8px; font-size: 13px;">Labels</h4>
              <el-tag
                v-for="(val, key) in pod.labels"
                :key="key"
                style="margin-right: 8px; margin-bottom: 8px;"
                size="small"
              >
                {{ key }}={{ val }}
              </el-tag>
            </div>

            <!-- Annotations -->
            <div v-if="pod.annotations && Object.keys(pod.annotations).length > 0" style="margin-top: 16px;">
              <h4 style="margin: 0 0 8px; font-size: 13px;">Annotations</h4>
              <div class="annotation-list">
                <div v-for="(val, key) in pod.annotations" :key="key" class="annotation-item">
                  <span class="annotation-key">{{ key }}</span>
                  <span class="annotation-value">{{ val }}</span>
                </div>
              </div>
            </div>

            <!-- Pod Conditions -->
            <div v-if="pod.conditions && pod.conditions.length > 0" style="margin-top: 16px;">
              <h4 style="margin: 0 0 8px; font-size: 13px;">Pod 条件</h4>
              <el-table :data="pod.conditions" border size="small" stripe>
                <el-table-column prop="type" label="类型" width="140" />
                <el-table-column label="状态" width="80">
                  <template #default="{ row }">
                    <el-tag :type="(row.status || '').toLowerCase() === 'true' ? 'success' : 'danger'" size="small">{{ row.status }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="reason" label="原因" min-width="120" show-overflow-tooltip />
              </el-table>
            </div>
          </div>
        </div>

        <!-- 右侧：容器 + 事件 -->
        <div class="right-panel">

          <!-- 容器列表 -->
          <div class="right-section" :style="rightTopHeight ? { flex: 'none', height: rightTopHeight + 'px' } : {}">
            <div class="panel-title">
              容器
              <span class="count-badge">{{ pod.containers?.length || 0 }} 个</span>
            </div>
            <div class="table-body">
              <el-table :data="pod.containers || []" size="small" stripe>
                <el-table-column type="expand">
                  <template #default="{ row }">
                    <div style="padding: 12px 16px;">
                      <div v-if="row.ports && row.ports.length > 0" style="margin-bottom: 16px;">
                        <h4 style="margin: 0 0 8px; font-size: 13px;">端口</h4>
                        <el-table :data="row.ports" border size="small">
                          <el-table-column prop="name" label="名称" width="120" />
                          <el-table-column prop="containerPort" label="容器端口" width="130" />
                          <el-table-column prop="protocol" label="协议" width="100" />
                        </el-table>
                      </div>
                      <div v-if="row.env && row.env.length > 0" style="margin-bottom: 16px;">
                        <h4 style="margin: 0 0 8px; font-size: 13px;">环境变量</h4>
                        <el-table :data="row.env" border size="small">
                          <el-table-column prop="name" label="名称" min-width="180" />
                          <el-table-column label="值" min-width="250">
                            <template #default="{ row: envRow }">
                              <span v-if="envRow.value !== undefined && envRow.value !== ''">{{ envRow.value }}</span>
                              <span v-else-if="envRow.valueFrom" style="color: var(--el-text-color-secondary);">{{ envRow.valueFrom.fieldRef?.fieldPath || envRow.valueFrom.secretKeyRef?.name || envRow.valueFrom.configMapKeyRef?.name || '来自引用' }}</span>
                              <span v-else style="color: var(--el-text-color-secondary);">-</span>
                            </template>
                          </el-table-column>
                        </el-table>
                      </div>
                      <div v-if="row.volumeMounts && row.volumeMounts.length > 0" style="margin-bottom: 16px;">
                        <h4 style="margin: 0 0 8px; font-size: 13px;">卷挂载</h4>
                        <el-table :data="row.volumeMounts" border size="small">
                          <el-table-column prop="name" label="卷名称" min-width="150" />
                          <el-table-column prop="mountPath" label="挂载路径" min-width="200" />
                          <el-table-column prop="subPath" label="子路径" width="150" />
                          <el-table-column label="只读" width="80">
                            <template #default="{ row: vm }">
                              <el-tag :type="vm.readOnly ? 'warning' : 'success'" size="small">{{ vm.readOnly ? '是' : '否' }}</el-tag>
                            </template>
                          </el-table-column>
                        </el-table>
                      </div>
                      <div v-if="row.livenessProbe" style="margin-bottom: 16px;">
                        <h4 style="margin: 0 0 8px; font-size: 13px;">存活探针</h4>
                        <el-descriptions :column="2" border size="small">
                          <el-descriptions-item v-if="row.livenessProbe.httpGet" label="类型">HTTP GET</el-descriptions-item>
                          <el-descriptions-item v-if="row.livenessProbe.httpGet" label="路径">{{ row.livenessProbe.httpGet.path || '/' }}</el-descriptions-item>
                          <el-descriptions-item v-if="row.livenessProbe.httpGet" label="端口">{{ row.livenessProbe.httpGet.port }}</el-descriptions-item>
                          <el-descriptions-item v-if="row.livenessProbe.tcpSocket" label="类型">TCP Socket</el-descriptions-item>
                          <el-descriptions-item v-if="row.livenessProbe.exec" label="类型">Exec</el-descriptions-item>
                          <el-descriptions-item v-if="row.livenessProbe.exec" label="命令">{{ (row.livenessProbe.exec.command || []).join(' ') }}</el-descriptions-item>
                          <el-descriptions-item label="初始延迟">{{ row.livenessProbe.initialDelaySeconds ?? '-' }}s</el-descriptions-item>
                          <el-descriptions-item label="检查周期">{{ row.livenessProbe.periodSeconds ?? '-' }}s</el-descriptions-item>
                        </el-descriptions>
                      </div>
                      <div v-if="row.readinessProbe">
                        <h4 style="margin: 0 0 8px; font-size: 13px;">就绪探针</h4>
                        <el-descriptions :column="2" border size="small">
                          <el-descriptions-item v-if="row.readinessProbe.httpGet" label="类型">HTTP GET</el-descriptions-item>
                          <el-descriptions-item v-if="row.readinessProbe.httpGet" label="路径">{{ row.readinessProbe.httpGet.path || '/' }}</el-descriptions-item>
                          <el-descriptions-item v-if="row.readinessProbe.httpGet" label="端口">{{ row.readinessProbe.httpGet.port }}</el-descriptions-item>
                          <el-descriptions-item label="初始延迟">{{ row.readinessProbe.initialDelaySeconds ?? '-' }}s</el-descriptions-item>
                          <el-descriptions-item label="检查周期">{{ row.readinessProbe.periodSeconds ?? '-' }}s</el-descriptions-item>
                        </el-descriptions>
                      </div>
                      <el-empty
                        v-if="(!row.ports || row.ports.length === 0) && (!row.env || row.env.length === 0) && (!row.volumeMounts || row.volumeMounts.length === 0) && !row.livenessProbe && !row.readinessProbe"
                        description="无额外容器详情"
                        :image-size="60"
                      />
                    </div>
                  </template>
                </el-table-column>
                <el-table-column prop="name" label="名称" min-width="150" />
                <el-table-column prop="image" label="镜像" min-width="260" show-overflow-tooltip />
                <el-table-column label="就绪" width="70">
                  <template #default="{ row }">
                    <el-tag :type="row.ready ? 'success' : 'danger'" size="small">{{ row.ready ? '是' : '否' }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="restartCount" label="重启" width="70" />
                <el-table-column label="状态" width="160">
                  <template #default="{ row }">
                    <el-tag :type="containerStateType(row.state)" size="small">{{ getContainerStateLabel(row) }}</el-tag>
                  </template>
                </el-table-column>
              </el-table>
              <el-empty v-if="!pod.containers || pod.containers.length === 0" description="无容器" />
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
      resource-type="pod"
      :namespace="namespace"
      :name="name"
      @saved="handleYamlSaved"
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

.table-body {
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

/* Annotation list */
.annotation-list {
  max-height: 200px;
  overflow-y: auto;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  padding: 8px;
}

.annotation-item {
  display: flex;
  padding: 4px 6px;
  border-bottom: 1px solid var(--el-border-color-extra-light);
  font-size: 12px;
}

.annotation-item:last-child {
  border-bottom: none;
}

.annotation-key {
  font-weight: 600;
  color: var(--el-text-color-primary);
  min-width: 160px;
  flex-shrink: 0;
  word-break: break-all;
}

.annotation-value {
  color: var(--el-text-color-secondary);
  word-break: break-all;
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
