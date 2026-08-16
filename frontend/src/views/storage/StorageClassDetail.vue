<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Timer, ArrowLeft, FullScreen, Aim } from '@element-plus/icons-vue'
import { getStorageClassDetail, deleteStorageClass, getPvcListByStorageClass, calcAge } from '@/api/resource'
import YamlDrawer from '@/components/YamlDrawer.vue'
import StorageClassForm from '@/views/storage/components/StorageClassForm.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import { useResizable } from '@/composables/useResizable'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const storageClass = ref<any>(null)
const yamlDialogVisible = ref(false)

// Related PVCs
const pvcs = ref<any[]>([])
const pvcsLoading = ref(false)

// Edit dialog
const editDialogVisible = ref(false)
const editFullscreen = ref(false)

// Right panel vertical resize
const rightTopHeight = ref<number | null>(null)
const resizingV = ref(false)
let vStartY = 0, vStartH = 0

function onVResizeStart(e: MouseEvent) {
  e.preventDefault()
  const rightPanel = (e.target as HTMLElement).closest('.right-panel')
  if (!rightPanel) return
  resizingV.value = true
  vStartY = e.clientY
  vStartH = rightPanel.getBoundingClientRect().height
  const onMove = (ev: MouseEvent) => {
    const delta = ev.clientY - vStartY
    rightTopHeight.value = Math.min(Math.max(vStartH * 0.3 + delta, 120), vStartH - 120)
  }
  const onUp = () => {
    resizingV.value = false
    document.removeEventListener('mousemove', onMove)
    document.removeEventListener('mouseup', onUp)
  }
  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onUp)
}

const name = route.params.name as string

const isDefault = computed(() => {
  const annotations = storageClass.value?.metadata?.annotations || {}
  return annotations['storageclass.kubernetes.io/is-default-class'] === 'true' ||
         annotations['storageclass.beta.kubernetes.io/is-default-class'] === 'true'
})

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getStorageClassDetail({ name })
    storageClass.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || '加载 StorageClass 详情失败')
  } finally {
    loading.value = false
  }
}

async function fetchPvcs() {
  pvcsLoading.value = true
  try {
    const res: any = await getPvcListByStorageClass({ storageClassName: name })
    const allPvcs = res.data || []
    pvcs.value = allPvcs.map((pvc: any) => ({
      namespace: pvc.metadata?.namespace || '-',
      name: pvc.metadata?.name || '-',
      status: pvc.status?.phase || 'Unknown',
      capacity: pvc.spec?.resources?.requests?.storage || '-',
      age: calcAge(pvc.metadata?.creationTimestamp),
    }))
  } catch {
    pvcs.value = []
  } finally {
    pvcsLoading.value = false
  }
}

function handleOpenYaml() {
  yamlDialogVisible.value = true
}

function handleYamlSaved() {
  fetchDetail()
}

function handleEdit() {
  editDialogVisible.value = true
}

function handleEditSuccess() {
  editDialogVisible.value = false
  fetchDetail()
}

function handleEditCancel() {
  editDialogVisible.value = false
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(`确认删除 StorageClass "${name}"？`, '确认删除', { type: 'warning' })
    await deleteStorageClass({ name })
    ElMessage.success('StorageClass 已删除')
    router.push('/storage/storageclasses')
  } catch {
    // cancelled
  }
}

const { leftWidth, resizingH, onHResizeStart } = useResizable({ initialWidth: 320 })

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(async () => {
  fetchDetail()
  fetchPvcs()
}, { autoStart: false })

onMounted(() => {
  fetchDetail()
  fetchPvcs()
})
</script>

<template>
  <div class="detail-page" v-loading="loading">

    <!-- ===== 顶部标题栏 ===== -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="res-name">{{ name }}</h2>
        <div class="meta-line">
          <el-tag v-if="isDefault" type="success" effect="dark" size="small">默认</el-tag>
          <span class="info-text" v-if="storageClass?.provisioner">{{ storageClass.provisioner }}</span>
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
          <el-button :icon="ArrowLeft" @click="router.push('/storage/storageclasses')" />
        </el-tooltip>
      </div>
    </div>

    <template v-if="storageClass">
      <div class="main-layout" :class="{ 'is-resizing': resizingH || resizingV }">

        <!-- 左侧：基本信息 -->
        <div class="left-panel" :style="{ width: leftWidth + 'px', minWidth: leftWidth + 'px' }">
          <div class="panel-title">基本信息</div>
          <div class="info-body">
            <div class="info-row">
              <span class="info-label">名称</span>
              <span class="info-value">{{ storageClass.metadata?.name || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Provisioner</span>
              <span class="info-value">{{ storageClass.provisioner || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">回收策略</span>
              <span class="info-value">{{ storageClass.reclaimPolicy || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">卷绑定模式</span>
              <span class="info-value">{{ storageClass.volumeBindingMode || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">允许卷扩展</span>
              <span class="info-value">
                <el-tag v-if="storageClass.allowVolumeExpansion" type="success" size="small">是</el-tag>
                <span v-else>否</span>
              </span>
            </div>
            <div class="info-row">
              <span class="info-label">默认</span>
              <span class="info-value">
                <el-tag v-if="isDefault" type="success" size="small">是</el-tag>
                <span v-else>否</span>
              </span>
            </div>
            <div class="info-row">
              <span class="info-label">创建时间</span>
              <span class="info-value">{{ storageClass.metadata?.creationTimestamp || '-' }}</span>
            </div>

            <!-- Labels -->
            <template v-if="storageClass.metadata?.labels && Object.keys(storageClass.metadata.labels).length > 0">
              <div class="info-row">
                <span class="info-label">标签</span>
                <span class="info-value">
                  <el-tag v-for="(val, key) in storageClass.metadata.labels" :key="key" size="small" class="label-tag">{{ key }}={{ val }}</el-tag>
                </span>
              </div>
            </template>

            <!-- Mount Options -->
            <template v-if="storageClass.mountOptions && storageClass.mountOptions.length > 0">
              <div class="info-row">
                <span class="info-label">挂载选项</span>
                <span class="info-value">
                  <el-tag v-for="opt in storageClass.mountOptions" :key="opt" size="small" type="info" class="label-tag">{{ opt }}</el-tag>
                </span>
              </div>
            </template>

            <!-- Parameters -->
            <template v-if="storageClass.parameters && Object.keys(storageClass.parameters).length > 0">
              <div class="info-row">
                <span class="info-label">参数</span>
                <span class="info-value">
                  <span v-for="(val, key) in storageClass.parameters" :key="key" class="param-item">
                    <span class="param-key">{{ key }}</span>
                    <span class="param-val">{{ val }}</span>
                  </span>
                </span>
              </div>
            </template>
          </div>
        </div>

        <!-- 水平拖拽条 -->
        <div
          class="resize-handle-h"
          :class="{ active: resizingH }"
          :style="{ left: (leftWidth - 3) + 'px' }"
          @mousedown="onHResizeStart"
        />

        <!-- 右侧：关联 PVC + 注解 -->
        <div class="right-panel">
          <!-- 关联 PVC -->
          <div class="right-section" :style="rightTopHeight ? { flex: 'none', height: rightTopHeight + 'px' } : {}">
            <div class="panel-title">
              关联 PVC
              <span class="count-badge">{{ pvcs.length }} 个</span>
            </div>
            <div v-loading="pvcsLoading" class="events-body">
              <el-table v-if="pvcs.length > 0" :data="pvcs" size="small" stripe>
                <el-table-column prop="namespace" label="命名空间" width="120" />
                <el-table-column prop="name" label="名称" min-width="160" show-overflow-tooltip>
                  <template #default="{ row }">
                    <el-button link type="primary" @click="$router.push(`/storage/pvcs/${row.namespace}/${row.name}`)">{{ row.name }}</el-button>
                  </template>
                </el-table-column>
                <el-table-column prop="status" label="状态" width="100">
                  <template #default="{ row }">
                    <el-tag :type="row.status === 'Bound' ? 'success' : row.status === 'Pending' ? 'warning' : 'info'" size="small">{{ row.status }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="capacity" label="容量" width="100" />
                <el-table-column prop="age" label="Age" width="100" />
              </el-table>
              <div v-else class="empty-hint">暂无关联 PVC</div>
            </div>
          </div>

          <!-- 垂直拖拽条（仅当注解存在时显示） -->
          <template v-if="storageClass.metadata?.annotations && Object.keys(storageClass.metadata.annotations).length > 0">
            <div class="resize-handle-v" :class="{ active: resizingV }" @mousedown="onVResizeStart" />
          </template>

          <!-- 注解 -->
          <div class="right-section" v-if="storageClass.metadata?.annotations && Object.keys(storageClass.metadata.annotations).length > 0">
            <div class="panel-title">注解</div>
            <div class="info-body">
              <div v-for="(val, key) in storageClass.metadata.annotations" :key="key" class="info-row">
                <span class="info-label mono" style="min-width: 120px;">{{ key }}</span>
                <span class="info-value mono" style="word-break: break-all;">{{ val }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- YAML Drawer -->
    <YamlDrawer
      v-model="yamlDialogVisible"
      resource-type="storageclass"
      :name="name"
      @saved="handleYamlSaved"
    />

    <!-- Edit Drawer -->
    <el-drawer
      v-model="editDialogVisible"
      title="编辑 StorageClass"
      :size="editFullscreen ? '100%' : '85%'"
      direction="rtl"
      :destroy-on-close="true"
      :body-style="{ padding: '0', height: '100%' }"
    >
      <template #header>
        <div class="drawer-header">
          <span class="drawer-title">编辑 StorageClass</span>
          <el-tooltip :content="editFullscreen ? '退出全屏' : '全屏'" placement="top">
            <el-icon class="fullscreen-btn" @click="editFullscreen = !editFullscreen">
              <FullScreen v-if="!editFullscreen" />
              <Aim v-else />
            </el-icon>
          </el-tooltip>
        </div>
      </template>
      <div style="height: calc(100vh - 52px); overflow-y: auto;">
        <StorageClassForm
          v-if="editDialogVisible && storageClass"
          :is-edit="true"
          :initial-data="storageClass"
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
  padding: 8px 14px;
  flex: 1;
  overflow-y: auto;
}

.info-row {
  display: flex;
  align-items: flex-start;
  padding: 6px 0;
  font-size: 13px;
  border-bottom: 1px solid var(--el-border-color-extra-light);
}

.info-row:last-child {
  border-bottom: none;
}

.info-label {
  color: var(--el-text-color-secondary);
  min-width: 72px;
  flex-shrink: 0;
}

.info-value {
  color: var(--el-text-color-primary);
  word-break: break-all;
  flex: 1;
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.mono {
  font-family: monospace;
  font-size: 12px;
}

.label-tag {
  margin: 0;
}

/* Parameter display */
.param-item {
  display: inline-flex;
  align-items: center;
  gap: 0;
  border: 1px solid var(--el-border-color-light);
  border-radius: 4px;
  overflow: hidden;
  font-size: 12px;
}

.param-key {
  background: var(--el-fill-color-lighter);
  padding: 2px 8px;
  color: var(--el-text-color-secondary);
  font-weight: 500;
}

.param-val {
  padding: 2px 8px;
  color: var(--el-text-color-primary);
  font-family: monospace;
  font-size: 12px;
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
