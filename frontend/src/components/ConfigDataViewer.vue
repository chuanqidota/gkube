<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { Editor as MonacoEditor } from '@guolao/vue-monaco-editor'
import { ElMessage } from 'element-plus'
import { Search, CopyDocument, Download, FullScreen, Aim } from '@element-plus/icons-vue'

interface DataEntry {
  key: string
  value: string
}

const props = withDefaults(defineProps<{
  entries: DataEntry[]
  loading?: boolean
  emptyText?: string
}>(), {
  loading: false,
  emptyText: '暂无数据',
})

const selectedKey = ref('')
const search = ref('')
const isFullscreen = ref(false)

const filteredEntries = computed(() => {
  const kw = search.value.trim().toLowerCase()
  if (!kw) return props.entries
  return props.entries.filter((e) => e.key.toLowerCase().includes(kw))
})

const selectedEntry = computed<DataEntry | undefined>(() =>
  props.entries.find((e) => e.key === selectedKey.value),
)

// 默认选中第一个 key；entries 变化时保持或重置
watch(() => props.entries, (list) => {
  if (!list.length) { selectedKey.value = ''; return }
  if (!list.some((e) => e.key === selectedKey.value)) {
    selectedKey.value = list[0].key
  }
}, { immediate: true })

const EXT_LANG: Record<string, string> = {
  json: 'json',
  yaml: 'yaml', yml: 'yaml',
  js: 'javascript', mjs: 'javascript', cjs: 'javascript',
  ts: 'typescript',
  html: 'html', htm: 'html',
  css: 'css', scss: 'scss', less: 'less',
  sh: 'shell', bash: 'shell', zsh: 'shell',
  py: 'python',
  go: 'go',
  xml: 'xml',
  ini: 'ini', conf: 'ini', cfg: 'ini',
  properties: 'properties',
  md: 'markdown',
  sql: 'sql',
  java: 'java',
}

// 识别语言：先按 key 扩展名，再尝试 JSON 解析，否则纯文本
function detectLanguage(key: string, value: string): string {
  const ext = key.includes('.') ? key.split('.').pop()!.toLowerCase() : ''
  if (ext && EXT_LANG[ext]) return EXT_LANG[ext]
  const trimmed = value.trim()
  if (trimmed.startsWith('{') || trimmed.startsWith('[')) {
    try { JSON.parse(trimmed); return 'json' } catch { /* fallthrough */ }
  }
  return 'plaintext'
}

// 计算右侧展示内容：JSON 美化，其余原样
const LARGE_VALUE_THRESHOLD = 100 * 1024 // 100 KB
const isLargeValue = computed(() => (selectedEntry.value?.value.length || 0) > LARGE_VALUE_THRESHOLD)
const showFull = ref(false)

watch(selectedKey, () => { showFull.value = false })

const displayValue = computed(() => {
  const entry = selectedEntry.value
  if (!entry) return ''
  if (isLargeValue.value && !showFull.value) {
    return entry.value.slice(0, LARGE_VALUE_THRESHOLD)
  }
  const lang = detectLanguage(entry.key, entry.value)
  if (lang === 'json') {
    try { return JSON.stringify(JSON.parse(entry.value.trim()), null, 2) } catch { return entry.value }
  }
  return entry.value
})

const displayLanguage = computed(() =>
  selectedEntry.value ? detectLanguage(selectedEntry.value.key, selectedEntry.value.value) : 'plaintext',
)

const editorOptions = computed(() => ({
  minimap: { enabled: false },
  fontSize: 13,
  lineNumbers: 'on' as const,
  scrollBeyondLastLine: false,
  wordWrap: 'on' as const,
  readOnly: true,
  automaticLayout: true,
  tabSize: 2,
}))

// Monaco 在弹窗动画期间需要重新布局
function handleEditorMount() {
  nextTick(() => {
    setTimeout(() => window.dispatchEvent(new Event('resize')), 300)
  })
}

function handleCopy() {
  if (!selectedEntry.value) return
  navigator.clipboard.writeText(selectedEntry.value.value)
    .then(() => ElMessage.success('已复制到剪贴板'))
    .catch(() => ElMessage.error('复制失败'))
}

function handleDownload() {
  if (!selectedEntry.value) return
  const blob = new Blob([selectedEntry.value.value], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = selectedEntry.value.key
  a.click()
  URL.revokeObjectURL(url)
}

function toggleFullscreen() {
  isFullscreen.value = !isFullscreen.value
  nextTick(() => {
    setTimeout(() => window.dispatchEvent(new Event('resize')), 100)
  })
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && isFullscreen.value) {
    e.preventDefault()
    toggleFullscreen()
  }
}

onMounted(() => document.addEventListener('keydown', handleKeydown))
onUnmounted(() => document.removeEventListener('keydown', handleKeydown))
</script>

<template>
  <div class="config-data-viewer" :class="{ 'is-fullscreen': isFullscreen }" v-loading="loading">
    <el-empty v-if="!loading && !entries.length" :description="emptyText" class="viewer-empty" />
    <div v-else class="viewer-body">
      <!-- 左侧：Key 列表 -->
      <div class="key-panel">
        <el-input
          v-model="search"
          size="small"
          clearable
          placeholder="搜索键"
          :prefix-icon="Search"
          class="key-search"
        />
        <div class="key-list">
          <div
            v-for="entry in filteredEntries"
            :key="entry.key"
            class="key-item"
            :class="{ active: entry.key === selectedKey }"
            :title="entry.key"
            @click="selectedKey = entry.key"
          >
            <span class="key-dot" />
            <span class="key-name">{{ entry.key }}</span>
          </div>
          <div v-if="!filteredEntries.length" class="key-empty">无匹配键</div>
        </div>
        <div class="key-count">共 {{ entries.length }} 项</div>
      </div>

      <!-- 右侧：值视图 -->
      <div class="value-panel">
        <div class="value-toolbar">
          <span class="value-title" :title="selectedEntry?.key">值: {{ selectedEntry?.key || '-' }}</span>
          <el-tag size="small" type="info" effect="plain" class="lang-tag">{{ displayLanguage }}</el-tag>
          <el-tag v-if="isLargeValue" size="small" type="warning" effect="plain">
            {{ (selectedEntry!.value.length / 1024).toFixed(0) }} KB
          </el-tag>
          <div class="toolbar-actions">
            <el-tooltip content="复制" placement="top">
              <el-button size="small" :icon="CopyDocument" :disabled="!selectedEntry" @click="handleCopy" />
            </el-tooltip>
            <el-tooltip content="下载" placement="top">
              <el-button size="small" :icon="Download" :disabled="!selectedEntry" @click="handleDownload" />
            </el-tooltip>
            <el-tooltip :content="isFullscreen ? '退出全屏' : '全屏'" placement="top">
              <el-button size="small" :icon="isFullscreen ? Aim : FullScreen" @click="toggleFullscreen" />
            </el-tooltip>
          </div>
        </div>
        <div v-if="isLargeValue && !showFull" class="large-value-banner">
          <el-alert type="warning" :closable="false" show-icon>
            值较大（{{ (selectedEntry!.value.length / 1024).toFixed(0) }} KB），已截断显示。
            <el-button size="small" text type="primary" @click="showFull = true">加载全部</el-button>
            <el-button size="small" text type="primary" @click="handleDownload">下载</el-button>
          </el-alert>
        </div>
        <div class="editor-wrap">
          <MonacoEditor
            :value="displayValue"
            :language="displayLanguage"
            :options="editorOptions"
            @mount="handleEditorMount"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.config-data-viewer {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
  background: var(--el-bg-color);
}
.config-data-viewer.is-fullscreen {
  position: fixed;
  inset: 0;
  z-index: 3000;
  background: var(--el-bg-color);
  padding: 12px;
  box-sizing: border-box;
}
.viewer-empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.viewer-body {
  display: flex;
  flex: 1;
  min-height: 0;
  gap: 8px;
}

/* 左侧 Key 面板 */
.key-panel {
  width: 220px;
  min-width: 220px;
  display: flex;
  flex-direction: column;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  overflow: hidden;
  background: var(--el-fill-color-blank);
}
.key-search {
  padding: 8px;
  box-sizing: border-box;
}
.key-list {
  flex: 1;
  overflow-y: auto;
  padding: 0 4px 4px;
}
.key-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 10px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  color: var(--el-text-color-regular);
  word-break: break-all;
  transition: background 0.15s;
}
.key-item:hover {
  background: var(--el-fill-color-light);
}
.key-item.active {
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  font-weight: 500;
}
.key-item.active .key-dot {
  background: var(--el-color-primary);
}
.key-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--el-border-color);
  flex-shrink: 0;
}
.key-empty {
  padding: 16px;
  text-align: center;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.key-count {
  padding: 6px 10px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  border-top: 1px solid var(--el-border-color-extra-light);
  flex-shrink: 0;
}

/* 右侧值面板 */
.value-panel {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  overflow: hidden;
  background: var(--el-bg-color);
}
.value-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  background: var(--el-fill-color-lighter);
  flex-shrink: 0;
  flex-wrap: wrap;
}
.large-value-banner {
  padding: 8px;
  flex-shrink: 0;
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.large-value-banner :deep(.el-alert__content) {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.value-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  min-width: 0;
}
.lang-tag {
  flex-shrink: 0;
  text-transform: lowercase;
}
.toolbar-actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}
.editor-wrap {
  flex: 1;
  min-height: 0;
}
.editor-wrap :deep(.monaco-editor) {
  border-radius: 0 0 6px 6px;
}
</style>
