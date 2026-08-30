<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Upload, Delete, Warning } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

export interface FileImportEntry {
  key: string
  value: string
  preEncoded: boolean  // true = value is already base64-encoded (binary file)
}

interface ImportFileItem {
  id: number
  file: File
  key: string         // editable, defaults to cleaned filename
  originalName: string
  isBinary: boolean
  size: number
  status: 'pending' | 'reading' | 'done' | 'error'
  errorMsg?: string
}

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  confirm: [entries: FileImportEntry[]]
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
})

const fileList = ref<ImportFileItem[]>([])
const reading = ref(false)
const dragover = ref(false)
const fileInputRef = ref<HTMLInputElement>()
const binaryDataMap = new Map<number, string>()  // id → base64 content for binary files
let nextId = 1
let dragCounter = 0  // tracks nested dragenter/dragleave to prevent flicker

// ---- Key name cleaning rules ----
// K8s ConfigMap/Secret keys: alphanumeric, '-', '_', '.' (max 253 chars, no '/')
function cleanKey(filename: string): string {
  let key = filename
    .replace(/\//g, '_')           // path separators → _
    .replace(/^\.+/, '')           // leading dots removed
    .replace(/\s+/g, '_')          // whitespace → _
    .replace(/[^\w.\-]/g, '_')     // non-safe chars → _
    .substring(0, 253)             // max length
  return key || 'key'
}

// ---- Binary detection by extension ----
const BINARY_EXTENSIONS = new Set([
  'pem', 'crt', 'cer', 'key', 'p12', 'pfx', 'jks', 'keystore', 'der',
  'png', 'jpg', 'jpeg', 'gif', 'bmp', 'ico', 'webp',
  'bin', 'dat', 'exe', 'dll', 'so', 'dylib',
  'zip', 'gz', 'tar', 'bz2', 'xz', '7z', 'rar',
])

function isBinaryFile(filename: string): boolean {
  const ext = filename.split('.').pop()?.toLowerCase() || ''
  return BINARY_EXTENSIONS.has(ext)
}

// ---- Size estimation (account for base64 4/3 expansion on text values) ----
const METADATA_OVERHEAD = 300
const MAX_TOTAL_SIZE = 1048576 - METADATA_OVERHEAD  // 1 MiB minus metadata

const estimatedTotalSize = computed(() => {
  let total = 0
  for (const item of fileList.value) {
    const keyBytes = new TextEncoder().encode(item.key).length
    // Binary: value is base64 (already expanded). Text: buildYamlStr will base64-encode → ×4/3
    const valueBytes = item.isBinary
      ? Math.ceil(item.size * 4 / 3)  // readAsArrayBuffer → base64
      : Math.ceil(item.size * 4 / 3)  // readAsText → base64 in buildYamlStr
    total += keyBytes + valueBytes
  }
  return total
})

const sizeExceeded = computed(() => estimatedTotalSize.value > MAX_TOTAL_SIZE)

const sizePercentage = computed(() =>
  Math.min(100, Math.round((estimatedTotalSize.value / MAX_TOTAL_SIZE) * 100))
)

function formatBytes(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

// ---- Key validation ----
function isValidKey(key: string): boolean {
  if (!key) return false
  if (key.length > 253) return false
  // K8s key: cannot contain '/', must match [A-Za-z0-9._-]
  return /^[A-Za-z0-9._-]+$/.test(key)
}

const hasInvalidKeys = computed(() =>
  fileList.value.some(item => !isValidKey(item.key))
)

// ---- Duplicate key detection (within batch) ----
function resolveDuplicates(items: ImportFileItem[]) {
  // First pass: resolve actual key suffixes for duplicate file names
  const keyCount = new Map<string, number>()
  for (const item of items) {
    const base = cleanKey(item.file.name)
    const count = keyCount.get(base) || 0
    if (count > 0) {
      item.key = `${base}_${count}`
    } else {
      item.key = base
    }
    keyCount.set(base, count + 1)
  }
  // Second pass: mark source hint for ALL items sharing the same base name
  const baseNameCount = new Map<string, number>()
  for (const item of items) {
    const base = cleanKey(item.file.name)
    baseNameCount.set(base, (baseNameCount.get(base) || 0) + 1)
  }
  const baseNameIndex = new Map<string, number>()
  for (const item of items) {
    const base = cleanKey(item.file.name)
    if ((baseNameCount.get(base) || 0) > 1) {
      const idx = (baseNameIndex.get(base) || 0) + 1
      baseNameIndex.set(base, idx)
      item.originalName = `${item.file.name} (${idx})`
    }
  }
}

// ---- File reading ----
async function processFiles(files: File[]) {
  if (files.length === 0) return

  const newItems: ImportFileItem[] = files.map(file => ({
    id: nextId++,
    file,
    key: cleanKey(file.name),
    originalName: file.name,
    isBinary: isBinaryFile(file.name),
    size: file.size,
    status: 'pending' as const,
  }))

  // Resolve within-batch duplicates
  resolveDuplicates(newItems)

  fileList.value.push(...newItems)

  // Read file contents
  reading.value = true
  for (const item of newItems) {
    item.status = 'reading'
    try {
      if (item.isBinary) {
        // Binary: read as ArrayBuffer → base64
        const buffer = await readFileAsArrayBuffer(item.file)
        binaryDataMap.set(item.id, arrayBufferToBase64(buffer))
      }
      // Text files will be read on confirm
      item.status = 'done'
    } catch (e: any) {
      item.status = 'error'
      item.errorMsg = e?.message || '读取失败'
    }
  }
  reading.value = false
}

function readFileAsText(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = (e) => resolve(e.target?.result as string)
    reader.onerror = () => reject(new Error('无法读取文件'))
    reader.readAsText(file)
  })
}

function readFileAsArrayBuffer(file: File): Promise<ArrayBuffer> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = (e) => resolve(e.target?.result as ArrayBuffer)
    reader.onerror = () => reject(new Error('无法读取文件'))
    reader.readAsArrayBuffer(file)
  })
}

function arrayBufferToBase64(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer)
  let binary = ''
  for (let i = 0; i < bytes.length; i++) {
    binary += String.fromCharCode(bytes[i])
  }
  return btoa(binary)
}

// ---- File input / drag-drop handlers ----
function handleFileInputChange(event: Event) {
  const input = event.target as HTMLInputElement
  if (input.files) {
    processFiles(Array.from(input.files))
  }
  input.value = ''  // reset to allow re-selecting same file
}

function handleDragenter(event: DragEvent) {
  event.preventDefault()
  dragCounter++
  dragover.value = true
}

function handleDragleave(_event: DragEvent) {
  dragCounter--
  if (dragCounter <= 0) {
    dragCounter = 0
    dragover.value = false
  }
}

function handleDrop(event: DragEvent) {
  event.preventDefault()
  dragCounter = 0
  dragover.value = false

  const items = event.dataTransfer?.items
  if (!items) return

  const files: File[] = []
  for (const item of Array.from(items)) {
    // Use webkitGetAsEntry to filter out directories
    const entry = (item as any).webkitGetAsEntry?.()
    if (entry && entry.isDirectory) continue  // skip directories
    const file = item.getAsFile()
    if (file) files.push(file)
  }
  processFiles(files)
}

function removeFile(id: number) {
  fileList.value = fileList.value.filter(item => item.id !== id)
  binaryDataMap.delete(id)
}

// ---- Confirm import ----
async function handleConfirm() {
  if (fileList.value.length === 0) {
    ElMessage.warning('请先选择文件')
    return
  }
  if (hasInvalidKeys.value) {
    ElMessage.warning('存在非法 Key，请修正后再导入')
    return
  }
  if (sizeExceeded.value) {
    ElMessage.warning(`数据总量超出 K8s 1 MiB 限制 (${formatBytes(estimatedTotalSize.value)})`)
    return
  }

  reading.value = true
  const entries: FileImportEntry[] = []

  for (const item of fileList.value) {
    // Skip items that failed during processFiles
    if (item.status === 'error') continue

    try {
      if (item.isBinary) {
        // Retrieve base64 from the map (populated during processFiles)
        const base64 = binaryDataMap.get(item.id)
        if (!base64) {
          ElMessage.error(`文件 ${item.originalName} 的二进制数据未读取`)
          reading.value = false
          return
        }
        entries.push({
          key: item.key,
          value: base64,
          preEncoded: true,
        })
      } else {
        // Read as text now
        const text = await readFileAsText(item.file)
        entries.push({
          key: item.key,
          value: text,
          preEncoded: false,
        })
      }
    } catch (e: any) {
      ElMessage.error(`读取文件 ${item.originalName} 失败: ${e?.message}`)
      reading.value = false
      return
    }
  }

  reading.value = false

  if (entries.length === 0) {
    ElMessage.warning('没有可导入的文件（全部读取失败）')
    return
  }

  emit('confirm', entries)
  ElMessage.success(`已导入 ${entries.length} 个文件`)
  visible.value = false
  fileList.value = []
  binaryDataMap.clear()
}

// ---- Reset on close ----
watch(visible, (v) => {
  if (!v) {
    fileList.value = []
    binaryDataMap.clear()
    reading.value = false
    dragover.value = false
    dragCounter = 0
  }
})
</script>

<template>
  <el-dialog
    v-model="visible"
    title="导入文件到数据"
    width="620px"
    :close-on-click-modal="false"
    @close="fileList = []; binaryDataMap.clear()"
  >
    <!-- Drop zone -->
    <div
      class="drop-zone"
      :class="{ dragover, 'has-files': fileList.length > 0 }"
      @dragenter="handleDragenter"
      @dragleave="handleDragleave"
      @drop="handleDrop"
      @dragover.prevent
      @click="fileInputRef?.click()"
    >
      <input
        ref="fileInputRef"
        type="file"
        multiple
        style="display: none;"
        @change="handleFileInputChange"
      />
      <el-icon :size="36" class="drop-icon"><Upload /></el-icon>
      <div class="drop-text">
        <template v-if="dragover">释放以导入文件</template>
        <template v-else>拖拽文件到此处，或点击选择</template>
      </div>
      <div class="drop-hint">支持多个文件，文件名将作为 Key</div>
    </div>

    <!-- File list -->
    <div v-if="fileList.length > 0" class="file-list">
      <div class="file-list-header">
        <span>已选择 {{ fileList.length }} 个文件</span>
      </div>

      <div class="file-table">
        <div class="file-row file-row-header">
          <span class="col-name">文件</span>
          <span class="col-key">Key（可编辑）</span>
          <span class="col-size">大小</span>
          <span class="col-action"></span>
        </div>
        <div
          v-for="item in fileList"
          :key="item.id"
          class="file-row"
          :class="{ 'file-error': item.status === 'error' }"
        >
          <span class="col-name" :title="item.originalName">
            <el-icon v-if="item.isBinary" :size="14" style="margin-right: 4px; color: var(--el-color-warning);"><Warning /></el-icon>
            {{ item.originalName }}
          </span>
          <span class="col-key">
            <el-input
              v-model="item.key"
              size="small"
              :class="{ 'key-invalid': !isValidKey(item.key) }"
              placeholder="Key"
            />
          </span>
          <span class="col-size">{{ formatBytes(item.size) }}</span>
          <span class="col-action">
            <el-button type="danger" text circle size="small" @click="removeFile(item.id)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </span>
        </div>
      </div>

      <!-- Size estimation -->
      <div class="size-bar" :class="{ 'size-exceeded': sizeExceeded }">
        <div class="size-info">
          <span>预估数据大小: {{ formatBytes(estimatedTotalSize) }} / {{ formatBytes(MAX_TOTAL_SIZE) }}</span>
          <span v-if="sizeExceeded" class="size-warning">⚠️ 超出 K8s 1 MiB 限制</span>
        </div>
        <el-progress
          :percentage="sizePercentage"
          :status="sizeExceeded ? 'exception' : sizePercentage > 80 ? 'warning' : undefined"
          :stroke-width="6"
          :show-text="false"
        />
      </div>
    </div>

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button
        type="primary"
        :loading="reading"
        :disabled="fileList.length === 0 || hasInvalidKeys || sizeExceeded"
        @click="handleConfirm"
      >
        确认导入
      </el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.drop-zone {
  border: 2px dashed var(--el-border-color);
  border-radius: 8px;
  padding: 32px 20px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s;
  background: var(--el-fill-color-lighter);
}

.drop-zone:hover,
.drop-zone.dragover {
  border-color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
}

.drop-zone.dragover {
  border-style: solid;
}

.drop-icon {
  color: var(--el-text-color-placeholder);
  margin-bottom: 8px;
}

.drop-zone.dragover .drop-icon,
.drop-zone:hover .drop-icon {
  color: var(--el-color-primary);
}

.drop-text {
  font-size: 14px;
  color: var(--el-text-color-regular);
  margin-bottom: 4px;
}

.drop-hint {
  font-size: 12px;
  color: var(--el-text-color-placeholder);
}

/* File list */
.file-list {
  margin-top: 16px;
}

.file-list-header {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin-bottom: 8px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.file-table {
  max-height: 240px;
  overflow-y: auto;
}

.file-row {
  display: grid;
  grid-template-columns: 1fr 200px 70px 36px;
  gap: 8px;
  align-items: center;
  padding: 6px 0;
  border-bottom: 1px solid var(--el-border-color-extra-light);
}

.file-row-header {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  font-weight: 500;
  padding-bottom: 4px;
}

.file-row.file-error {
  background: var(--el-color-error-light-9);
}

.col-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
  display: flex;
  align-items: center;
}

.col-key :deep(.el-input__inner) {
  font-family: monospace;
  font-size: 12px;
}

.col-key .key-invalid :deep(.el-input__wrapper) {
  box-shadow: 0 0 0 1px var(--el-color-danger) inset;
}

.col-size {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  text-align: right;
}

.col-action {
  display: flex;
  justify-content: center;
}

/* Size bar */
.size-bar {
  margin-top: 12px;
  padding: 10px 12px;
  background: var(--el-fill-color-lighter);
  border-radius: 6px;
}

.size-bar.size-exceeded {
  background: var(--el-color-error-light-9);
}

.size-info {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-bottom: 6px;
}

.size-warning {
  color: var(--el-color-error);
  font-weight: 600;
}
</style>
