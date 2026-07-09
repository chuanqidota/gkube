<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getServiceDetail,
  getServiceYaml,
  updateService,
  deleteService,
  getServiceEvents,
  getServicePods,
  deletePod,
} from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'
import PodListPanel from '@/components/PodListPanel.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const service = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const yamlDialogVisible = ref(false)
const yamlEditorRef = ref<InstanceType<typeof YamlEditor>>()

// Events
const events = ref<any[]>([])
const eventsLoading = ref(false)

// Related Pods
const pods = ref<any[]>([])
const podsLoading = ref(false)

const namespace = route.params.namespace as string
const name = route.params.name as string

function transformService(raw: any) {
  const spec = raw.spec || {}
  const meta = raw.metadata || {}
  const status = raw.status || {}

  // Format ports
  const ports = (spec.ports || [])
    .map((p: any) => `${p.port}${p.nodePort ? ':' + p.nodePort : ''}/${p.protocol || 'TCP'}`)
    .join(', ')

  // External IP from status
  let externalIP = ''
  const lbIngress = status.loadBalancer?.ingress
  if (lbIngress && lbIngress.length > 0) {
    externalIP = lbIngress.map((i: any) => i.ip || i.hostname || '').filter(Boolean).join(', ')
  }

  // Age from creationTimestamp
  let age = ''
  if (meta.creationTimestamp) {
    const created = new Date(meta.creationTimestamp).getTime()
    const diff = Date.now() - created
    const seconds = Math.floor(diff / 1000)
    if (seconds < 60) age = `${seconds}s`
    else if (seconds < 3600) age = `${Math.floor(seconds / 60)}m`
    else if (seconds < 86400) age = `${Math.floor(seconds / 3600)}h`
    else age = `${Math.floor(seconds / 86400)}d`
  }

  return {
    name: meta.name || '',
    namespace: meta.namespace || '',
    type: spec.type || 'ClusterIP',
    clusterIP: spec.clusterIP || '',
    externalIP,
    ports,
    sessionAffinity: spec.sessionAffinity || 'None',
    age,
    selector: spec.selector || {},
    labels: meta.labels || {},
  }
}

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getServiceDetail({ namespace, name })
    service.value = transformService(res.data)
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load service detail')
  } finally {
    loading.value = false
  }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getServiceYaml({ namespace, name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load YAML')
  } finally {
    yamlLoading.value = false
  }
}

async function fetchEvents() {
  eventsLoading.value = true
  try {
    const res: any = await getServiceEvents({ namespace, name })
    events.value = res.data || []
  } catch (e) {
    console.error('Failed to fetch events:', e)
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
    console.error('Failed to fetch pods:', e)
  } finally {
    podsLoading.value = false
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
      `确认删除 Pod "${pod.metadata.name}"？`,
      '确认删除',
      { type: 'warning' }
    )
    await deletePod({ namespace, name: pod.metadata.name })
    ElMessage.success('Pod 已删除')
    fetchPods()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

function handleOpenYaml() {
  fetchYaml()
  yamlDialogVisible.value = true
}

async function handleSaveYaml(content: string) {
  try {
    await updateService({ namespace, name, yaml: content })
    ElMessage.success('YAML 已保存')
    yamlDialogVisible.value = false
    fetchDetail()
  } catch (e: any) {
    ElMessage.error(e?.message || '保存 YAML 失败')
    yamlEditorRef.value?.resetSaving()
  }
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(
      `确认删除 Service "${name}"（命名空间: ${namespace}）？`,
      '确认删除',
      { type: 'warning' }
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
        <div class="title-line">
          <h2 class="res-name">{{ name }}</h2>
          <el-tag v-if="service?.type" size="small">{{ service.type }}</el-tag>
          <span class="ns-tag">ns/{{ namespace }}</span>
          <span class="cluster-ip" v-if="service?.clusterIP || service?.cluster_ip">
            Cluster IP: {{ service.clusterIP || service.cluster_ip }}
          </span>
        </div>
      </div>
      <div class="header-actions">
        <AutoRefreshToolbar
          :is-running="isRunning"
          :countdown="countdown"
          :current-interval="currentInterval"
          :available-intervals="availableIntervals"
          :loading="loading"
          @refresh="manualRefresh()"
          @toggle="toggle()"
          @interval-change="setIntervalOption"
        />
        <el-button @click="handleOpenYaml">YAML</el-button>
        <el-button type="danger" @click="handleDelete">删除</el-button>
        <el-button @click="router.push('/network/services')">返回列表</el-button>
      </div>
    </div>

    <template v-if="service">
      <div class="main-layout" :class="{ 'is-resizing': resizingH || resizingV }">

        <!-- 左侧：基本信息 -->
        <div class="left-panel" :style="{ width: leftWidth + 'px', minWidth: leftWidth + 'px' }">
          <div class="panel-title">基本信息</div>
          <div class="info-body">
            <el-descriptions :column="1" border size="small">
              <el-descriptions-item label="名称">{{ service.name }}</el-descriptions-item>
              <el-descriptions-item label="命名空间">{{ service.namespace }}</el-descriptions-item>
              <el-descriptions-item label="类型">{{ service.type || '-' }}</el-descriptions-item>
              <el-descriptions-item label="Cluster IP">{{ service.clusterIP || service.cluster_ip || '-' }}</el-descriptions-item>
              <el-descriptions-item label="External IP">{{ service.externalIP || service.external_ip || '-' }}</el-descriptions-item>
              <el-descriptions-item label="端口">{{ service.ports || '-' }}</el-descriptions-item>
              <el-descriptions-item label="Session Affinity">{{ service.sessionAffinity || '-' }}</el-descriptions-item>
              <el-descriptions-item label="Age">{{ service.age || '-' }}</el-descriptions-item>
            </el-descriptions>

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

        <!-- 水平拖拽条（绝对定位，覆盖在左右面板交界） -->
        <div
          class="resize-handle-h"
          :class="{ active: resizingH }"
          :style="{ left: (leftWidth - 3) + 'px' }"
          @mousedown="onHResizeStart"
        />
      </div>
    </template>

    <!-- YAML Edit Drawer -->
    <el-drawer v-model="yamlDialogVisible" title="YAML" size="85%" direction="rtl" class="yaml-drawer"
      :body-style="{ padding: '0', height: '100%' }">
      <div v-loading="yamlLoading" style="height: calc(100vh - 52px);">
        <YamlEditor ref="yamlEditorRef" v-model="yamlContent" height="100%" auto-format show-save-buttons @save="handleSaveYaml" @cancel="fetchYaml" />
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
  gap: 2px;
}

.title-line {
  display: flex;
  align-items: center;
  gap: 10px;
}

.res-name {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.ns-tag {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color-lighter);
  padding: 2px 8px;
  border-radius: 4px;
}

.cluster-ip {
  font-size: 13px;
  color: var(--el-text-color-regular);
  font-family: monospace;
}

.header-actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
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

/* Left Panel - Info */
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
</style>

<style>
.yaml-drawer .el-drawer__header {
  padding: 6px 16px;
  margin-bottom: 0;
  min-height: auto;
}
</style>
