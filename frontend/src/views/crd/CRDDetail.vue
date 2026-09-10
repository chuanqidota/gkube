<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, ArrowLeft } from '@element-plus/icons-vue'
import { getCrdDetail, getCrdYaml, updateCrd, deleteCrd } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import { useResizable } from '@/composables/useResizable'
import * as jsYaml from 'js-yaml'

const { leftWidth, resizingH, onHResizeStart } = useResizable({ initialWidth: 320 })

const { t } = useI18n()
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
    ElMessage.error(e?.message || t('crd.loadCrdFailed'))
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
    ElMessage.error(e?.message || t('crd.getYamlFailed'))
    yamlDialogVisible.value = false
  } finally {
    yamlLoading.value = false
  }
}

async function handleYamlSave() {
  yamlSaving.value = true
  try {
    await updateCrd({ yaml: yamlContent.value })
    ElMessage.success(t('common.saveSuccess'))
    yamlDialogVisible.value = false
    fetchDetail()
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.saveFailed'))
  } finally {
    yamlSaving.value = false
  }
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(
      t('crd.deleteCrdConfirm', { name }),
      t('crd.confirmDelete'),
      { type: 'error' }
    )
    await deleteCrd({ name })
    ElMessage.success(t('crd.crdDeleted'))
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
            {{ crd?.spec?.scope === 'Namespaced' ? t('crd.namespaced') : t('crd.clusterScope') }}
          </el-tag>
          <el-tag size="small" type="info">{{ crd?.spec?.group }}</el-tag>
          <span class="info-text">{{ crd?.spec?.names?.kind }}</span>
        </div>
      </div>
      <div class="header-actions">
        <el-button type="info" @click="handleOpenYaml">{{ t('crd.yamlEdit') }}</el-button>
        <el-button type="danger" @click="handleDelete">{{ t('common.delete') }}</el-button>
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
              {{ isRunning ? `${t('common.autoRefresh')} ${countdown}s` : t('common.autoRefresh') }}
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
                :label="`${t('common.refreshInterval')}: ${sec}s`"
              />
            </el-select>
          </div>
        </el-popover>
        <el-tooltip :content="t('common.refresh')" placement="top">
          <el-button @click="manualRefresh()" :loading="loading" :icon="Refresh" />
        </el-tooltip>
        <el-tooltip :content="t('common.backToList')" placement="top">
          <el-button :icon="ArrowLeft" @click="router.push('/crd')" />
        </el-tooltip>
      </div>
    </div>

    <template v-if="crd">
      <div class="main-layout" :class="{ 'is-resizing': resizingH }">

        <!-- Left Panel: Basic Info -->
        <div class="left-panel" :style="{ width: leftWidth + 'px', minWidth: leftWidth + 'px' }">
          <div class="panel-title">{{ t('crd.basicInfo') }}</div>
          <div class="info-body">
            <div class="info-row">
              <span class="info-label">{{ t('common.name') }}</span>
              <span class="info-value">{{ crd.metadata?.name || name }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">{{ t('crd.kind') }}</span>
              <span class="info-value">{{ crd.spec?.names?.kind || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">{{ t('crd.plural') }}</span>
              <span class="info-value mono">{{ crd.spec?.names?.plural || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">{{ t('crd.singular') }}</span>
              <span class="info-value mono">{{ crd.spec?.names?.singular || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">{{ t('crd.shortNames') }}</span>
              <span class="info-value mono">{{ shortNames }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">{{ t('crd.apiGroup') }}</span>
              <span class="info-value mono">{{ crd.spec?.group || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">{{ t('crd.scope') }}</span>
              <span class="info-value">
                <el-tag size="small" :type="crd.spec?.scope === 'Namespaced' ? 'info' : 'warning'">
                  {{ crd.spec?.scope === 'Namespaced' ? t('crd.namespaced') : t('crd.clusterScope') }}
                </el-tag>
              </span>
            </div>
            <div class="info-row">
              <span class="info-label">{{ t('crd.versions') }}</span>
              <span class="info-value">
                <el-tag v-for="v in versions" :key="v.name" size="small" :type="v.storage ? 'success' : 'info'" class="label-tag">
                  {{ v.name }}{{ v.storage ? t('crd.storageTag') : '' }}
                </el-tag>
              </span>
            </div>
            <div class="info-row">
              <span class="info-label">{{ t('crd.categories') }}</span>
              <span class="info-value">{{ categories }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">UID</span>
              <span class="info-value mono">{{ crd.metadata?.uid || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">{{ t('crd.createTimestamp') }}</span>
              <span class="info-value">{{ crd.metadata?.creationTimestamp || '-' }}</span>
            </div>

            <!-- Labels -->
            <template v-if="crd.metadata?.labels && Object.keys(crd.metadata.labels).length > 0">
              <div class="info-row">
                <span class="info-label">{{ t('crd.labelTag') }}</span>
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
                  <el-tag v-if="ver.storage" size="small" type="success" style="margin-left: 4px;">{{ t('crd.storageVersion') }}</el-tag>
                  <el-tag v-if="ver.served" size="small" type="info" style="margin-left: 4px;">{{ t('crd.enabled') }}</el-tag>
                </template>

                <div v-if="selectedVersion" class="version-body">
                  <!-- Subresources -->
                  <template v-if="selectedVersion.subresources">
                    <div class="section-block">
                      <div class="section-title">{{ t('crd.subresources') }}</div>
                      <div class="section-content">
                        <el-tag v-if="selectedVersion.subresources.status" size="small" type="warning">status</el-tag>
                        <el-tag v-if="selectedVersion.subresources.scale" size="small" type="warning">scale</el-tag>
                        <span v-if="!selectedVersion.subresources.status && !selectedVersion.subresources.scale" class="empty-text">{{ t('common.no') }}</span>
                      </div>
                    </div>
                  </template>

                  <!-- Additional Printer Columns -->
                  <template v-if="selectedVersion.additionalPrinterColumns?.length > 0">
                    <div class="section-block">
                      <div class="section-title">{{ t('crd.additionalPrinterColumns') }}</div>
                      <div class="section-content">
                        <el-table :data="selectedVersion.additionalPrinterColumns" size="small" border stripe>
                          <el-table-column prop="name" :label="t('common.name')" min-width="120" />
                          <el-table-column prop="type" :label="t('common.type')" width="100" />
                          <el-table-column prop="jsonPath" label="JSON Path" min-width="180" show-overflow-tooltip />
                          <el-table-column prop="description" :label="t('common.description')" min-width="200" show-overflow-tooltip />
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
                      <div v-else class="empty-text" style="padding: 20px;">{{ t('crd.noSchema') }}</div>
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
      <div v-loading="yamlLoading" style="height: calc(100dvh - 52px);">
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
  padding: var(--gk-space-4) var(--gk-space-5);
  height: calc(100dvh - var(--gk-header-height));
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
}

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
  background: var(--gk-color-primary-bg);
}

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
  font-size: var(--gk-font-size-sm);
  border-bottom: 1px solid var(--el-border-color-extra-light);
}

.info-row:last-child {
  border-bottom: none;
}

.info-label {
  color: var(--gk-color-text-secondary);
  min-width: 100px;
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

.version-tabs {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.version-tabs :deep(.el-tabs__header) {
  margin: 0;
  padding: 0 14px;
  background: var(--gk-neutral-100);
  border-bottom: 1px solid var(--gk-color-border-light);
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
  border-radius: var(--gk-radius-md);
  overflow: hidden;
}

.schema-block {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 300px;
}

.section-title {
  font-size: var(--gk-font-size-sm);
  font-weight: 600;
  padding: 8px 12px;
  background: var(--gk-neutral-100);
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
  font-size: var(--gk-font-size-sm);
  color: var(--gk-color-text-placeholder);
}
</style>
