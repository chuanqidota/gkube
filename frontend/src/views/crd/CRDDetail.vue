<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, ArrowLeft } from '@element-plus/icons-vue'
import { getCrdDetail, getCrdYaml, updateCrd, deleteCrd } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import { useResizable } from '@/composables/useResizable'
import * as jsYaml from 'js-yaml'

const { leftWidth, resizingH, onHResizeStart } = useResizable({ initialWidth: 320 })

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const crd = ref<any>(null)

const name = route.query.name as string

// YAML drawer state
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)
const yamlSaving = ref(false)

const activeVersion = ref('')

const versions = computed(() => {
  if (!crd.value?.spec?.versions) return []
  return crd.value.spec.versions.map((v: any) => ({
    name: v.name,
    served: v.served,
    storage: v.storage,
    schema: v.schema?.openAPIV3Schema ? jsYaml.dump(v.schema.openAPIV3Schema, { indent: 2, lineWidth: -1 }) : '',
    subresources: v.subresources || null,
    additionalPrinterColumns: v.additionalPrinterColumns || [],
  }))
})

const shortNames = computed(() => {
  return crd.value?.spec?.names?.shortNames?.join(', ') || '-'
})

const categories = computed(() => {
  return crd.value?.spec?.names?.categories?.join(', ') || '-'
})

const selectedVersion = computed(() => {
  return versions.value.find((v: any) => v.name === activeVersion.value) || versions.value[0]
})

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getCrdDetail({ name })
    crd.value = res.data
    if (versions.value.length > 0 && !activeVersion.value) {
      // Default to storage version
      const storageVersion = versions.value.find((v: any) => v.storage)
      activeVersion.value = storageVersion?.name || versions.value[0].name
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '加载 CRD 详情失败')
  } finally {
    loading.value = false
  }
}

async function handleOpenYaml() {
  yamlDialogVisible.value = true
  yamlLoading.value = true
  yamlContent.value = ''
  try {
    const res: any = await getCrdYaml({ name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || '获取 YAML 失败')
    yamlDialogVisible.value = false
  } finally {
    yamlLoading.value = false
  }
}

async function handleYamlSave() {
  yamlSaving.value = true
  try {
    await updateCrd({ yaml: yamlContent.value })
    ElMessage.success('保存成功')
    yamlDialogVisible.value = false
    fetchDetail()
  } catch (e: any) {
    ElMessage.error(e?.message || '保存失败')
  } finally {
    yamlSaving.value = false
  }
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(
      `删除 CRD "${name}"？这将同时删除该类型的所有自定义资源！`,
      '确认删除',
      { type: 'error' }
    )
    await deleteCrd({ name })
    ElMessage.success('CRD 已删除')
    router.push('/crd')
  } catch {
    /* cancelled */
  }
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchDetail, { autoStart: false })

onMounted(fetchDetail)
</script>

<template>
  <div class="detail-page" v-loading="loading">

    <!-- Header -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="res-name">{{ name }}</h2>
        <div class="meta-line">
          <el-tag size="small" :type="crd?.spec?.scope === 'Namespaced' ? 'info' : 'warning'">
            {{ crd?.spec?.scope === 'Namespaced' ? '命名空间' : '集群' }}
          </el-tag>
          <el-tag size="small" type="info">{{ crd?.spec?.group }}</el-tag>
          <span class="info-text">{{ crd?.spec?.names?.kind }}</span>
        </div>
      </div>
      <div class="header-actions">
        <el-button type="info" @click="handleOpenYaml">YAML 编辑</el-button>
        <el-button type="danger" plain @click="handleDelete">删除</el-button>
        <div class="action-divider" />
        <el-popover placement="bottom" :width="200" trigger="click">
          <template #reference>
            <el-button
              :type="isRunning ? 'success' : 'default'"
              :icon="Refresh"
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
          <el-button :icon="ArrowLeft" @click="router.push('/crd')" />
        </el-tooltip>
      </div>
    </div>

    <template v-if="crd">
      <div class="main-layout" :class="{ 'is-resizing': resizingH }">

        <!-- Left Panel: Basic Info -->
        <div class="left-panel" :style="{ width: leftWidth + 'px', minWidth: leftWidth + 'px' }">
          <div class="panel-title">基本信息</div>
          <div class="info-body">
            <div class="info-row">
              <span class="info-label">名称</span>
              <span class="info-value">{{ crd.metadata?.name || name }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Kind</span>
              <span class="info-value">{{ crd.spec?.names?.kind || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Plural</span>
              <span class="info-value mono">{{ crd.spec?.names?.plural || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Singular</span>
              <span class="info-value mono">{{ crd.spec?.names?.singular || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Short Names</span>
              <span class="info-value mono">{{ shortNames }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">API 组</span>
              <span class="info-value mono">{{ crd.spec?.group || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">作用域</span>
              <span class="info-value">
                <el-tag size="small" :type="crd.spec?.scope === 'Namespaced' ? 'info' : 'warning'">
                  {{ crd.spec?.scope === 'Namespaced' ? '命名空间' : '集群' }}
                </el-tag>
              </span>
            </div>
            <div class="info-row">
              <span class="info-label">版本</span>
              <span class="info-value">
                <el-tag v-for="v in versions" :key="v.name" size="small" :type="v.storage ? 'success' : 'info'" class="label-tag">
                  {{ v.name }}{{ v.storage ? ' (存储)' : '' }}
                </el-tag>
              </span>
            </div>
            <div class="info-row">
              <span class="info-label">分类</span>
              <span class="info-value">{{ categories }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">UID</span>
              <span class="info-value mono">{{ crd.metadata?.uid || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">创建时间</span>
              <span class="info-value">{{ crd.metadata?.creationTimestamp || '-' }}</span>
            </div>

            <!-- Labels -->
            <template v-if="crd.metadata?.labels && Object.keys(crd.metadata.labels).length > 0">
              <div class="info-row">
                <span class="info-label">标签</span>
                <span class="info-value">
                  <el-tag v-for="(val, key) in crd.metadata.labels" :key="key" size="small" class="label-tag">{{ key }}={{ val }}</el-tag>
                </span>
              </div>
            </template>
          </div>
        </div>

        <!-- Drag handle -->
        <div
          class="resize-handle-h"
          :class="{ active: resizingH }"
          :style="{ left: (leftWidth - 3) + 'px' }"
          @mousedown="onHResizeStart"
        />

        <!-- Right Panel: Version Details -->
        <div class="right-panel">
          <div class="right-section">
            <el-tabs v-model="activeVersion" class="version-tabs">
              <el-tab-pane
                v-for="ver in versions"
                :key="ver.name"
                :name="ver.name"
              >
                <template #label>
                  <span>{{ ver.name }}</span>
                  <el-tag v-if="ver.storage" size="small" type="success" style="margin-left: 4px;">存储</el-tag>
                  <el-tag v-if="ver.served" size="small" type="info" style="margin-left: 4px;">启用</el-tag>
                </template>

                <div v-if="selectedVersion" class="version-body">
                  <!-- Subresources -->
                  <template v-if="selectedVersion.subresources">
                    <div class="section-block">
                      <div class="section-title">子资源 (Subresources)</div>
                      <div class="section-content">
                        <el-tag v-if="selectedVersion.subresources.status" size="small" type="warning">status</el-tag>
                        <el-tag v-if="selectedVersion.subresources.scale" size="small" type="warning">scale</el-tag>
                        <span v-if="!selectedVersion.subresources.status && !selectedVersion.subresources.scale" class="empty-text">无</span>
                      </div>
                    </div>
                  </template>

                  <!-- Additional Printer Columns -->
                  <template v-if="selectedVersion.additionalPrinterColumns?.length > 0">
                    <div class="section-block">
                      <div class="section-title">额外打印列 (Additional Printer Columns)</div>
                      <div class="section-content">
                        <el-table :data="selectedVersion.additionalPrinterColumns" size="small" border stripe>
                          <el-table-column prop="name" label="名称" min-width="120" />
                          <el-table-column prop="type" label="类型" width="100" />
                          <el-table-column prop="jsonPath" label="JSON Path" min-width="180" show-overflow-tooltip />
                          <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
                        </el-table>
                      </div>
                    </div>
                  </template>

                  <!-- Schema -->
                  <div class="section-block schema-block">
                    <div class="section-title">OpenAPI v3 Schema</div>
                    <div class="schema-editor">
                      <YamlEditor
                        v-if="selectedVersion.schema"
                        :model-value="selectedVersion.schema"
                        height="100%"
                        read-only
                        auto-format
                        :show-toolbar="false"
                      />
                      <div v-else class="empty-text" style="padding: 20px;">无 Schema 定义</div>
                    </div>
                  </div>
                </div>
              </el-tab-pane>
            </el-tabs>
          </div>
        </div>
      </div>
    </template>

    <!-- YAML Edit Drawer -->
    <el-drawer
      v-model="yamlDialogVisible"
      :title="`CRD YAML: ${name}`"
      size="85%"
      direction="rtl"
      :body-style="{ padding: '0', height: '100%' }"
      :destroy-on-close="true"
    >
      <div v-loading="yamlLoading" style="height: calc(100vh - 52px);">
        <YamlEditor
          v-if="!yamlLoading"
          v-model="yamlContent"
          height="100%"
          auto-format
          show-save-buttons
          :saving="yamlSaving"
          @save="handleYamlSave"
          @cancel="handleOpenYaml"
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
  background: var(--el-color-primary-light-7);
}

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
  min-width: 100px;
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

.version-tabs {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.version-tabs :deep(.el-tabs__header) {
  margin: 0;
  padding: 0 14px;
  background: var(--el-fill-color-lighter);
  border-bottom: 1px solid var(--el-border-color-lighter);
  flex-shrink: 0;
}

.version-tabs :deep(.el-tabs__content) {
  flex: 1;
  min-height: 0;
}

.version-tabs :deep(.el-tab-pane) {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.version-body {
  flex: 1;
  overflow-y: auto;
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.section-block {
  border: 1px solid var(--el-border-color-extra-light);
  border-radius: 6px;
  overflow: hidden;
}

.schema-block {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 300px;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  padding: 8px 12px;
  background: var(--el-fill-color-lighter);
  border-bottom: 1px solid var(--el-border-color-extra-light);
}

.section-content {
  padding: 10px 12px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.schema-editor {
  flex: 1;
  overflow: hidden;
  min-height: 200px;
}

.empty-text {
  font-size: 13px;
  color: var(--el-text-color-placeholder);
}
</style>
