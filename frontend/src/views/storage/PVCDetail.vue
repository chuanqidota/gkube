<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Timer, ArrowLeft, FullScreen, Aim } from '@element-plus/icons-vue'
import { getPvcDetail, deletePvc } from '@/api/resource'
import YamlDrawer from '@/components/YamlDrawer.vue'
import PVCForm from '@/views/storage/components/PVCForm.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const pvc = ref<any>(null)
const yamlDialogVisible = ref(false)

const namespace = route.params.namespace as string
const name = route.params.name as string

// Edit dialog
const editDialogVisible = ref(false)
const editFullscreen = ref(false)

const statusTagType = computed(() => {
  const s = (pvc.value?.status?.phase || '').toLowerCase()
  if (s === 'bound') return 'success'
  if (s === 'pending') return 'warning'
  if (s === 'lost') return 'danger'
  return 'info'
})

const pvcStatus = computed(() => pvc.value?.status?.phase || '-')

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getPvcDetail({ namespace, name })
    pvc.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || '加载PVC详情失败')
  } finally {
    loading.value = false
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
    await ElMessageBox.confirm(`删除持久卷声明 "${name}"?`, '确认', { type: 'warning' })
    await deletePvc({ namespace, name })
    ElMessage.success('删除成功')
    router.push('/storage/pvcs')
  } catch {
    /* cancelled */
  }
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchDetail, { autoStart: false })

onMounted(fetchDetail)
</script>

<template>
  <div class="detail-page" v-loading="loading">

    <!-- ===== 顶部标题栏 ===== -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="res-name">{{ name }}</h2>
        <div class="meta-line">
          <el-tag v-if="pvcStatus !== '-'" :type="statusTagType" effect="dark" size="small">{{ pvcStatus }}</el-tag>
          <span class="ns-tag">ns/{{ namespace }}</span>
        </div>
      </div>
      <div class="header-actions">
        <el-button type="info" @click="handleEdit">编辑</el-button>
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
          <el-button :icon="ArrowLeft" @click="router.push('/storage/pvcs')" />
        </el-tooltip>
      </div>
    </div>

    <template v-if="pvc">
      <div class="main-layout">

        <!-- 左侧：基本信息 -->
        <div class="left-panel">
          <div class="panel-title">基本信息</div>
          <div class="info-body">
            <div class="info-row">
              <span class="info-label">名称</span>
              <span class="info-value">{{ pvc.metadata?.name || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">命名空间</span>
              <span class="info-value">{{ pvc.metadata?.namespace || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">状态</span>
              <span class="info-value">
                <el-tag :type="statusTagType" size="small">{{ pvcStatus }}</el-tag>
              </span>
            </div>
            <div class="info-row">
              <span class="info-label">卷名</span>
              <span class="info-value">{{ pvc.spec?.volumeName || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">容量</span>
              <span class="info-value">{{ pvc.status?.capacity?.storage || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">访问模式</span>
              <span class="info-value">{{ (pvc.spec?.accessModes || []).join(', ') || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">存储类名</span>
              <span class="info-value">{{ pvc.spec?.storageClassName || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">创建时间</span>
              <span class="info-value">{{ pvc.metadata?.creationTimestamp || '-' }}</span>
            </div>

            <!-- Labels -->
            <template v-if="pvc.metadata?.labels && Object.keys(pvc.metadata.labels).length > 0">
              <div class="info-row">
                <span class="info-label">标签</span>
                <span class="info-value">
                  <el-tag v-for="(val, key) in pvc.metadata.labels" :key="key" size="small" class="label-tag">{{ key }}={{ val }}</el-tag>
                </span>
              </div>
            </template>
          </div>
        </div>

        <!-- 右侧：注解 -->
        <div class="right-panel">
          <div class="right-section" v-if="pvc.metadata?.annotations && Object.keys(pvc.metadata.annotations).length > 0">
            <div class="panel-title">注解</div>
            <div class="info-body">
              <div v-for="(val, key) in pvc.metadata.annotations" :key="key" class="info-row">
                <span class="info-label mono" style="min-width: 120px;">{{ key }}</span>
                <span class="info-value mono" style="word-break: break-all;">{{ val }}</span>
              </div>
            </div>
          </div>
          <div v-else class="right-section">
            <div class="panel-title">详情</div>
            <div class="empty-hint">暂无额外信息</div>
          </div>
        </div>
      </div>
    </template>

    <!-- YAML Drawer -->
    <YamlDrawer
      v-model="yamlDialogVisible"
      resource-type="pvc"
      :namespace="namespace"
      :name="name"
      @saved="handleYamlSaved"
    />

    <!-- Edit Drawer -->
    <el-drawer
      v-model="editDialogVisible"
      title="编辑 PersistentVolumeClaim"
      :size="editFullscreen ? '100%' : '85%'"
      direction="rtl"
      :destroy-on-close="true"
      :body-style="{ padding: '0', height: '100%' }"
    >
      <template #header>
        <div class="drawer-header">
          <span class="drawer-title">编辑 PersistentVolumeClaim</span>
          <el-tooltip :content="editFullscreen ? '退出全屏' : '全屏'" placement="top">
            <el-icon class="fullscreen-btn" @click="editFullscreen = !editFullscreen">
              <FullScreen v-if="!editFullscreen" />
              <Aim v-else />
            </el-icon>
          </el-tooltip>
        </div>
      </template>
      <div style="height: calc(100vh - 52px); overflow-y: auto;">
        <PVCForm
          v-if="editDialogVisible && pvc"
          :is-edit="true"
          :initial-data="pvc"
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
  width: 320px;
  min-width: 320px;
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
