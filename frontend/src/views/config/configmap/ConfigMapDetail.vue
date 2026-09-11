<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Timer, ArrowLeft, FullScreen, Aim } from '@element-plus/icons-vue'
import { getConfigMapDetail, deleteConfigMap, configMapApi } from '@/api/resource'
import YamlDrawer from '@/components/YamlDrawer.vue'
import ConfigDataViewer from '@/components/ConfigDataViewer.vue'
import ConfigMapForm from '@/views/config/components/ConfigMapForm.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import { useResizable } from '@/composables/useResizable'

const { leftWidth, resizingH, onHResizeStart } = useResizable({ initialWidth: 320 })

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const loading = ref(false)
const configMap = ref<any>(null)
const yamlDialogVisible = ref(false)

const namespace = route.params.namespace as string
const name = route.params.name as string

// Edit dialog
const editDialogVisible = ref(false)
const editFullscreen = ref(false)
const activeDataTab = ref('data')

const dataEntries = computed(() => {
  const data = configMap.value?.data || {}
  return Object.entries(data).map(([key, value]) => ({
    key,
    value: String(value ?? ''),
  }))
})

const binaryDataEntries = computed(() => {
  const binaryData = configMap.value?.binaryData || {}
  return Object.entries(binaryData).map(([key, value]) => ({
    key,
    value: String(value ?? ''),
  }))
})

const dataCount = computed(() => dataEntries.value.length)
const binaryDataCount = computed(() => binaryDataEntries.value.length)

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getConfigMapDetail({ namespace, name })
    configMap.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || t('config.loadDetailFailed'))
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
    await ElMessageBox.confirm(`删除配置字典 "${name}"?`, '确认', { type: 'warning' })
    await deleteConfigMap({ namespace, name })
    ElMessage.success(t('common.deleteSuccess'))
    router.push('/config/configmaps')
  } catch {
    /* cancelled */
  }
}

const {
  isRunning,
  countdown,
  currentInterval,
  availableIntervals,
  toggle,
  refresh: manualRefresh,
  setIntervalOption,
} = useAutoRefresh(fetchDetail, { autoStart: false })

onMounted(fetchDetail)
</script>

<template>
  <div v-loading="loading" class="detail-page">
    <!-- ===== 顶部标题栏 ===== -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="res-name">{{ name }}</h2>
        <div class="meta-line">
          <span class="ns-tag">ns/{{ namespace }}</span>
          <span class="info-text">{{ dataCount + binaryDataCount }} 个数据项</span>
        </div>
      </div>
      <div class="header-actions">
        <el-button type="info" @click="handleEdit">编辑</el-button>
        <el-button @click="handleOpenYaml">YAML</el-button>
        <el-button type="danger" @click="handleDelete">删除</el-button>
        <div class="action-divider" />
        <el-popover placement="bottom" :width="200" trigger="click">
          <template #reference>
            <el-button :type="isRunning ? 'success' : 'default'" :icon="Timer" @click="toggle()" />
          </template>
          <div class="auto-refresh-popover">
            <div class="popover-title">
              {{ isRunning ? `自动刷新中 ${countdown}s` : '自动刷新' }}
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
                :label="`每 ${sec} 秒刷新`"
              />
            </el-select>
          </div>
        </el-popover>
        <el-tooltip content="刷新" placement="top">
          <el-button :loading="loading" :icon="Refresh" @click="manualRefresh()" />
        </el-tooltip>
        <el-tooltip content="返回列表" placement="top">
          <el-button :icon="ArrowLeft" @click="router.push('/config/configmaps')" />
        </el-tooltip>
      </div>
    </div>

    <template v-if="configMap">
      <div class="main-layout" :class="{ 'is-resizing': resizingH }">
        <!-- 左侧：基本信息 -->
        <div class="left-panel" :style="{ width: leftWidth + 'px', minWidth: leftWidth + 'px' }">
          <div class="panel-title">基本信息</div>
          <div class="info-body">
            <div class="info-row">
              <span class="info-label">名称</span>
              <span class="info-value">{{
                configMap.metadata?.name || configMap.name || '-'
              }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">命名空间</span>
              <span class="info-value">{{
                configMap.metadata?.namespace || configMap.namespace || '-'
              }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">UID</span>
              <span class="info-value mono">{{ configMap.metadata?.uid || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">创建时间</span>
              <span class="info-value">{{ configMap.metadata?.creationTimestamp || '-' }}</span>
            </div>

            <!-- Labels -->
            <template
              v-if="configMap.metadata?.labels && Object.keys(configMap.metadata.labels).length > 0"
            >
              <div class="info-row">
                <span class="info-label">标签</span>
                <span class="info-value">
                  <el-tag
                    v-for="(val, key) in configMap.metadata.labels"
                    :key="key"
                    size="small"
                    class="label-tag"
                    >{{ key }}={{ val }}</el-tag
                  >
                </span>
              </div>
            </template>

            <!-- Annotations -->
            <template
              v-if="
                configMap.metadata?.annotations &&
                Object.keys(configMap.metadata.annotations).length > 0
              "
            >
              <div class="info-row" style="flex-direction: column">
                <span class="info-label" style="margin-bottom: 4px">注解</span>
                <div
                  v-for="(val, key) in configMap.metadata.annotations"
                  :key="key"
                  class="annotation-row"
                >
                  <span class="annotation-key mono">{{ key }}</span>
                  <span class="annotation-value mono">{{ val }}</span>
                </div>
              </div>
            </template>
          </div>
        </div>

        <!-- 拖拽分隔条 -->
        <div
          class="resize-handle-h"
          :class="{ active: resizingH }"
          :style="{ left: leftWidth - 3 + 'px' }"
          @mousedown="onHResizeStart"
        />

        <!-- 右侧：Data / BinaryData -->
        <div class="right-panel">
          <div class="right-section">
            <el-tabs v-model="activeDataTab" class="data-tabs">
              <el-tab-pane name="data">
                <template #label>
                  数据 <span class="count-badge">{{ dataCount }}</span>
                </template>
                <div class="data-body">
                  <ConfigDataViewer :entries="dataEntries" />
                </div>
              </el-tab-pane>
              <el-tab-pane name="binaryData">
                <template #label>
                  二进制数据 <span class="count-badge">{{ binaryDataCount }}</span>
                </template>
                <div class="data-body">
                  <ConfigDataViewer :entries="binaryDataEntries" />
                </div>
              </el-tab-pane>
            </el-tabs>
          </div>
        </div>
      </div>
    </template>

    <!-- YAML Drawer -->
    <YamlDrawer
      v-model="yamlDialogVisible"
      :get-yaml="configMapApi.getYaml"
      :update-yaml="configMapApi.updateYaml"
      :namespace="namespace"
      :name="name"
      title="ConfigMap YAML"
      @saved="handleYamlSaved"
    />

    <!-- Edit Drawer -->
    <el-drawer
      v-model="editDialogVisible"
      title="编辑 ConfigMap"
      :size="editFullscreen ? '100%' : '85%'"
      direction="rtl"
      :destroy-on-close="true"
      :body-style="{ padding: '0', height: '100%' }"
    >
      <template #header>
        <div class="drawer-header">
          <span class="drawer-title">编辑 ConfigMap</span>
          <el-tooltip :content="editFullscreen ? '退出全屏' : '全屏'" placement="top">
            <el-icon class="fullscreen-btn" @click="editFullscreen = !editFullscreen">
              <FullScreen v-if="!editFullscreen" />
              <Aim v-else />
            </el-icon>
          </el-tooltip>
        </div>
      </template>
      <div style="height: calc(100dvh - 52px); overflow-y: auto">
        <ConfigMapForm
          v-if="editDialogVisible && configMap"
          :is-edit="true"
          :initial-data="configMap"
          @success="handleEditSuccess"
          @cancel="handleEditCancel"
        />
      </div>
    </el-drawer>
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

.info-text {
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

.header-actions .el-button:last-of-type {
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

.main-layout.is-resizing {
  user-select: none;
  pointer-events: none;
}

/* 拖拽分隔条 */
.resize-handle-h {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 6px;
  cursor: col-resize;
  z-index: 10;
  transition: background 0.15s;
}

.resize-handle-h:hover,
.resize-handle-h.active {
  background: var(--gk-color-primary-bg);
}

/* Left Panel */
.left-panel {
  width: 320px;
  min-width: 320px;
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

.info-body {
  padding: 8px 14px;
  flex: 1;
  overflow-y: auto;
}

.info-row {
  display: flex;
  align-items: flex-start;
  padding: 6px 0;
  font-size: var(--gk-font-size-sm);
  border-bottom: 1px solid var(--el-border-color-extra-light);
}

.info-row:last-child {
  border-bottom: none;
}

.info-label {
  color: var(--gk-color-text-secondary);
  min-width: 72px;
  flex-shrink: 0;
}

.info-value {
  color: var(--gk-color-text-primary);
  word-break: break-all;
  flex: 1;
  display: flex;
  flex-wrap: wrap;
  gap: var(--gk-space-1);
}

.mono {
  font-family: var(--gk-font-mono);
  font-size: 12px;
}

.label-tag {
  margin: 0;
}

.annotation-row {
  display: flex;
  gap: 8px;
  padding: 2px 0;
  font-size: 12px;
  border-bottom: 1px solid var(--el-border-color-extra-light);
  width: 100%;
}

.annotation-row:last-child {
  border-bottom: none;
}

.annotation-key {
  color: var(--gk-color-text-secondary);
  min-width: 80px;
  word-break: break-all;
}

.annotation-value {
  color: var(--gk-color-text-primary);
  word-break: break-all;
  flex: 1;
}

/* Right Panel */
.right-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
}

.right-section {
  flex: 1;
  border: 1px solid var(--gk-color-border-light);
  border-radius: var(--gk-radius-md);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--el-bg-color);
}

.data-tabs {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.data-tabs :deep(.el-tabs__header) {
  margin: 0;
  padding: 0 14px;
  background: var(--gk-neutral-100);
  border-bottom: 1px solid var(--gk-color-border-light);
  flex-shrink: 0;
}

.data-tabs :deep(.el-tabs__content) {
  flex: 1;
  min-height: 0;
}

.data-tabs :deep(.el-tab-pane) {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.data-body {
  flex: 1;
  overflow: hidden;
  padding: 8px;
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
