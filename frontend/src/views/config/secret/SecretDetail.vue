<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Timer, ArrowLeft, FullScreen, Aim } from '@element-plus/icons-vue'
import { getSecretDetail, deleteSecret } from '@/api/resource'
import YamlDrawer from '@/components/YamlDrawer.vue'
import SecretForm from '@/views/config/components/SecretForm.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const secret = ref<any>(null)
const yamlDialogVisible = ref(false)

const namespace = route.params.namespace as string
const name = route.params.name as string

// Edit dialog
const editDialogVisible = ref(false)
const editFullscreen = ref(false)

// Decode toggle
const showDecoded = ref(true)

function base64Decode(str: string): string {
  try {
    return atob(str)
  } catch {
    return str
  }
}

const dataEntries = computed(() => {
  const data = secret.value?.data || {}
  return Object.entries(data).map(([key, value]) => {
    const rawValue = String(value ?? '')
    return {
      key,
      rawValue,
      decodedValue: base64Decode(rawValue),
    }
  })
})

const dataCount = computed(() => dataEntries.value.length)

const secretType = computed(() => secret.value?.type || '-')

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getSecretDetail({ namespace, name })
    secret.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || '加载Secret详情失败')
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
    await ElMessageBox.confirm(`删除保密字典 "${name}"?`, '确认', { type: 'warning' })
    await deleteSecret({ namespace, name })
    ElMessage.success('删除成功')
    router.push('/config/secrets')
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
          <span class="ns-tag">ns/{{ namespace }}</span>
          <el-tag size="small" type="info">{{ secretType }}</el-tag>
          <span class="info-text">{{ dataCount }} 个数据项</span>
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
          <el-button :icon="ArrowLeft" @click="router.push('/config/secrets')" />
        </el-tooltip>
      </div>
    </div>

    <template v-if="secret">
      <div class="main-layout">

        <!-- 左侧：基本信息 -->
        <div class="left-panel">
          <div class="panel-title">基本信息</div>
          <div class="info-body">
            <div class="info-row">
              <span class="info-label">名称</span>
              <span class="info-value">{{ secret.metadata?.name || secret.name || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">命名空间</span>
              <span class="info-value">{{ secret.metadata?.namespace || secret.namespace || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">类型</span>
              <span class="info-value">{{ secret.type || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">UID</span>
              <span class="info-value mono">{{ secret.metadata?.uid || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">创建时间</span>
              <span class="info-value">{{ secret.metadata?.creationTimestamp || '-' }}</span>
            </div>

            <!-- Labels -->
            <template v-if="secret.metadata?.labels && Object.keys(secret.metadata.labels).length > 0">
              <div class="info-row">
                <span class="info-label">标签</span>
                <span class="info-value">
                  <el-tag v-for="(val, key) in secret.metadata.labels" :key="key" size="small" class="label-tag">{{ key }}={{ val }}</el-tag>
                </span>
              </div>
            </template>

            <!-- Annotations -->
            <template v-if="secret.metadata?.annotations && Object.keys(secret.metadata.annotations).length > 0">
              <div class="info-row" style="flex-direction: column;">
                <span class="info-label" style="margin-bottom: 4px;">注解</span>
                <div v-for="(val, key) in secret.metadata.annotations" :key="key" class="annotation-row">
                  <span class="annotation-key mono">{{ key }}</span>
                  <span class="annotation-value mono">{{ val }}</span>
                </div>
              </div>
            </template>
          </div>
        </div>

        <!-- 右侧：Data -->
        <div class="right-panel">
          <div class="right-section">
            <div class="panel-title">
              数据
              <span class="count-badge">{{ dataCount }} 项</span>
              <span class="toggle-wrap">
                <el-switch v-model="showDecoded" active-text="解码" inactive-text="Base64" size="small" />
              </span>
            </div>
            <div class="data-body">
              <el-table v-if="dataEntries.length > 0" :data="dataEntries" size="small" border stripe>
                <el-table-column prop="key" label="Key" width="220" show-overflow-tooltip />
                <el-table-column label="Value" min-width="300">
                  <template #default="{ row }">
                    <div class="value-cell">{{ showDecoded ? row.decodedValue : row.rawValue }}</div>
                  </template>
                </el-table-column>
              </el-table>
              <div v-else class="empty-hint">暂无数据</div>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- YAML Drawer -->
    <YamlDrawer
      v-model="yamlDialogVisible"
      resource-type="secret"
      :namespace="namespace"
      :name="name"
      @saved="handleYamlSaved"
    />

    <!-- Edit Drawer -->
    <el-drawer
      v-model="editDialogVisible"
      title="编辑 Secret"
      :size="editFullscreen ? '100%' : '85%'"
      direction="rtl"
      :destroy-on-close="true"
      :body-style="{ padding: '0', height: '100%' }"
    >
      <template #header>
        <div class="drawer-header">
          <span class="drawer-title">编辑 Secret</span>
          <el-tooltip :content="editFullscreen ? '退出全屏' : '全屏'" placement="top">
            <el-icon class="fullscreen-btn" @click="editFullscreen = !editFullscreen">
              <FullScreen v-if="!editFullscreen" />
              <Aim v-else />
            </el-icon>
          </el-tooltip>
        </div>
      </template>
      <div style="height: calc(100vh - 52px); overflow-y: auto;">
        <SecretForm
          v-if="editDialogVisible && secret"
          :is-edit="true"
          :initial-data="secret"
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

.count-badge {
  font-weight: 400;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.toggle-wrap {
  margin-left: auto;
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
  color: var(--el-text-color-secondary);
  min-width: 80px;
  word-break: break-all;
}

.annotation-value {
  color: var(--el-text-color-primary);
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
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--el-bg-color);
}

.data-body {
  flex: 1;
  overflow-y: auto;
  padding: 0;
}

.value-cell {
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 120px;
  overflow-y: auto;
  font-family: monospace;
  font-size: 12px;
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
