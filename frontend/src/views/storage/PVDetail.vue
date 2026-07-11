<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Timer, ArrowLeft, FullScreen, Aim } from '@element-plus/icons-vue'
import { getPvDetail, deletePv } from '@/api/resource'
import YamlDrawer from '@/components/YamlDrawer.vue'
import PVForm from '@/views/storage/components/PVForm.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const pv = ref<any>(null)
const yamlDialogVisible = ref(false)

const name = route.params.name as string

// Edit dialog
const editDialogVisible = ref(false)
const editFullscreen = ref(false)

const statusTagType = computed(() => {
  const status = (pv.value?.status?.phase || '').toLowerCase()
  if (status === 'bound') return 'success'
  if (status === 'available') return 'primary'
  if (status === 'released') return 'warning'
  if (status === 'failed') return 'danger'
  return 'info'
})

const statusText = computed(() => pv.value?.status?.phase || 'Unknown')

const storageType = computed(() => {
  if (pv.value?.spec?.nfs) return 'NFS'
  if (pv.value?.spec?.hostPath) return 'HostPath'
  if (pv.value?.spec?.local) return 'Local'
  return '-'
})

const storageSource = computed(() => {
  if (pv.value?.spec?.nfs) return `${pv.value.spec.nfs.server}:${pv.value.spec.nfs.path}`
  if (pv.value?.spec?.hostPath) return pv.value.spec.hostPath.path
  if (pv.value?.spec?.local) return pv.value.spec.local.path
  return '-'
})

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getPvDetail({ name })
    pv.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load PV detail')
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
    await ElMessageBox.confirm(`删除持久卷 "${name}"?`, '确认', { type: 'warning' })
    await deletePv({ name })
    ElMessage.success('删除成功')
    router.push('/storage/pvs')
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
          <el-tag :type="statusTagType" effect="dark" size="small">{{ statusText }}</el-tag>
          <span class="info-text" v-if="pv">{{ pv.spec?.capacity?.storage || '-' }} / {{ (pv.spec?.accessModes || []).join(', ') }}</span>
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
          <el-button :icon="ArrowLeft" @click="router.push('/storage/pvs')" />
        </el-tooltip>
      </div>
    </div>

    <template v-if="pv">
      <div class="main-layout">

        <!-- 左侧：基本信息 -->
        <div class="left-panel">
          <div class="panel-title">基本信息</div>
          <div class="info-body">
            <div class="info-row">
              <span class="info-label">名称</span>
              <span class="info-value">{{ pv.metadata?.name || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">容量</span>
              <span class="info-value">{{ pv.spec?.capacity?.storage || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">访问模式</span>
              <span class="info-value">{{ (pv.spec?.accessModes || []).join(', ') || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">存储类</span>
              <span class="info-value">{{ pv.spec?.storageClassName || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">状态</span>
              <span class="info-value">
                <el-tag :type="statusTagType" size="small">{{ statusText }}</el-tag>
              </span>
            </div>
            <div class="info-row">
              <span class="info-label">回收策略</span>
              <span class="info-value">{{ pv.spec?.persistentVolumeReclaimPolicy || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">卷模式</span>
              <span class="info-value">{{ pv.spec?.volumeMode || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">创建时间</span>
              <span class="info-value">{{ pv.metadata?.creationTimestamp || '-' }}</span>
            </div>

            <!-- Labels -->
            <template v-if="pv.metadata?.labels && Object.keys(pv.metadata.labels).length > 0">
              <div class="info-row">
                <span class="info-label">标签</span>
                <span class="info-value">
                  <el-tag v-for="(val, key) in pv.metadata.labels" :key="key" size="small" class="label-tag">{{ key }}={{ val }}</el-tag>
                </span>
              </div>
            </template>
          </div>
        </div>

        <!-- 右侧：存储源 + 声明引用 + 注解 -->
        <div class="right-panel">
          <!-- 存储源 -->
          <div class="right-section" style="flex: none;">
            <div class="panel-title">存储源</div>
            <div class="info-body">
              <div class="info-row">
                <span class="info-label">类型</span>
                <span class="info-value">{{ storageType }}</span>
              </div>
              <div class="info-row" v-if="pv.spec?.nfs">
                <span class="info-label">服务器</span>
                <span class="info-value mono">{{ pv.spec.nfs.server || '-' }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">路径</span>
                <span class="info-value mono">{{ storageSource }}</span>
              </div>
            </div>
          </div>

          <!-- 声明引用 -->
          <div class="right-section" v-if="pv.spec?.claimRef" style="flex: none;">
            <div class="panel-title">声明引用</div>
            <div class="info-body">
              <div class="info-row">
                <span class="info-label">命名空间</span>
                <span class="info-value">{{ pv.spec.claimRef.namespace || '-' }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">名称</span>
                <span class="info-value">{{ pv.spec.claimRef.name || '-' }}</span>
              </div>
            </div>
          </div>

          <!-- 注解 -->
          <div class="right-section" v-if="pv.metadata?.annotations && Object.keys(pv.metadata.annotations).length > 0">
            <div class="panel-title">注解</div>
            <div class="info-body">
              <div v-for="(val, key) in pv.metadata.annotations" :key="key" class="info-row">
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
      resource-type="pv"
      :name="name"
      @saved="handleYamlSaved"
    />

    <!-- Edit Drawer -->
    <el-drawer
      v-model="editDialogVisible"
      title="编辑 PersistentVolume"
      :size="editFullscreen ? '100%' : '85%'"
      direction="rtl"
      :destroy-on-close="true"
      :body-style="{ padding: '0', height: '100%' }"
    >
      <template #header>
        <div class="drawer-header">
          <span class="drawer-title">编辑 PersistentVolume</span>
          <el-tooltip :content="editFullscreen ? '退出全屏' : '全屏'" placement="top">
            <el-icon class="fullscreen-btn" @click="editFullscreen = !editFullscreen">
              <FullScreen v-if="!editFullscreen" />
              <Aim v-else />
            </el-icon>
          </el-tooltip>
        </div>
      </template>
      <div style="height: calc(100vh - 52px); overflow-y: auto;">
        <PVForm
          v-if="editDialogVisible && pv"
          :is-edit="true"
          :initial-data="pv"
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
