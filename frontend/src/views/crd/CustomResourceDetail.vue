<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, ArrowLeft } from '@element-plus/icons-vue'
import { getCustomResourceDetail, getCustomResourceYaml, updateCustomResource, deleteCustomResource } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import { useResizable } from '@/composables/useResizable'
import * as jsYaml from 'js-yaml'

const { leftWidth, resizingH, onHResizeStart } = useResizable({ initialWidth: 320 })

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const resource = ref<any>(null)

const group = route.query.group as string
const version = route.query.version as string
const resourceName = route.query.resource as string
const namespace = route.query.namespace as string || ''
const name = route.query.name as string
const scope = route.query.scope as string

// YAML drawer state
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)
const yamlSaving = ref(false)

const backRoute = computed(() =>
  `/crd/resources?group=${group}&version=${version}&resource=${resourceName}&scope=${scope}`
)

const labels = computed(() => {
  const l = resource.value?.metadata?.labels
  return l && Object.keys(l).length > 0 ? l : null
})

const annotations = computed(() => {
  const a = resource.value?.metadata?.annotations
  if (!a) return null
  // Filter out large system annotations
  const filtered: Record<string, string> = {}
  for (const [k, v] of Object.entries(a)) {
    if (!k.startsWith('kubectl.kubernetes.io/')) {
      filtered[k] = v as string
    }
  }
  return Object.keys(filtered).length > 0 ? filtered : null
})

const specYaml = computed(() => {
  if (!resource.value) return ''
  // Show everything except metadata and status
  const obj = { ...resource.value }
  delete obj.metadata
  delete obj.status
  delete obj.apiVersion
  delete obj.kind
  try {
    return jsYaml.dump(obj, { indent: 2, lineWidth: -1 })
  } catch {
    return JSON.stringify(obj, null, 2)
  }
})

async function fetchDetail() {
  loading.value = true
  try {
    const params: any = { group, version, resource: resourceName, name }
    if (namespace) params.namespace = namespace
    const res: any = await getCustomResourceDetail(params)
    resource.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || '加载详情失败')
  } finally {
    loading.value = false
  }
}

async function handleOpenYaml() {
  yamlDialogVisible.value = true
  yamlLoading.value = true
  yamlContent.value = ''
  try {
    const params: any = { group, version, resource: resourceName, name }
    if (namespace) params.namespace = namespace
    const res: any = await getCustomResourceYaml(params)
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
    const data: any = { group, version, resource: resourceName, yaml: yamlContent.value }
    if (namespace) data.namespace = namespace
    await updateCustomResource(data)
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
    await ElMessageBox.confirm(`删除自定义资源 "${name}"?`, '确认', { type: 'warning' })
    const params: any = { group, version, resource: resourceName, name }
    if (namespace) params.namespace = namespace
    await deleteCustomResource(params)
    ElMessage.success('删除成功')
    router.push(backRoute.value)
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
          <span v-if="namespace" class="ns-tag">ns/{{ namespace }}</span>
          <el-tag size="small" type="info">{{ group }}/{{ version }}</el-tag>
          <span class="info-text">{{ resourceName }}</span>
        </div>
      </div>
      <div class="header-actions">
        <el-button type="info" @click="handleOpenYaml">YAML 编辑</el-button>
        <el-button type="danger" @click="handleDelete">删除</el-button>
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
          <el-button :icon="ArrowLeft" @click="router.push(backRoute)" />
        </el-tooltip>
      </div>
    </div>

    <template v-if="resource">
      <div class="main-layout" :class="{ 'is-resizing': resizingH }">

        <!-- Left Panel: Metadata -->
        <div class="left-panel" :style="{ width: leftWidth + 'px', minWidth: leftWidth + 'px' }">
          <div class="panel-title">基本信息</div>
          <div class="info-body">
            <div class="info-row">
              <span class="info-label">名称</span>
              <span class="info-value">{{ resource.metadata?.name || name }}</span>
            </div>
            <div v-if="namespace" class="info-row">
              <span class="info-label">命名空间</span>
              <span class="info-value">{{ resource.metadata?.namespace || namespace }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">UID</span>
              <span class="info-value mono">{{ resource.metadata?.uid || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">创建时间</span>
              <span class="info-value">{{ resource.metadata?.creationTimestamp || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">ResourceVersion</span>
              <span class="info-value mono">{{ resource.metadata?.resourceVersion || '-' }}</span>
            </div>

            <!-- Labels -->
            <template v-if="labels">
              <div class="info-row">
                <span class="info-label">标签</span>
                <span class="info-value">
                  <el-tag v-for="(val, key) in labels" :key="key" size="small" class="label-tag">{{ key }}={{ val }}</el-tag>
                </span>
              </div>
            </template>

            <!-- Annotations -->
            <template v-if="annotations">
              <div class="info-row" style="flex-direction: column;">
                <span class="info-label" style="margin-bottom: 4px;">注解</span>
                <div v-for="(val, key) in annotations" :key="key" class="annotation-row">
                  <span class="annotation-key mono">{{ key }}</span>
                  <span class="annotation-value mono">{{ val }}</span>
                </div>
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

        <!-- Right Panel: Spec YAML -->
        <div class="right-panel">
          <div class="right-section">
            <div class="panel-title">资源内容</div>
            <div class="yaml-body">
              <YamlEditor
                v-if="specYaml"
                :model-value="specYaml"
                height="100%"
                read-only
                auto-format
                :show-toolbar="false"
              />
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- YAML Edit Drawer -->
    <el-drawer
      v-model="yamlDialogVisible"
      :title="`YAML: ${name}`"
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

.yaml-body {
  flex: 1;
  overflow: hidden;
}
</style>
