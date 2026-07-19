<template>
  <div class="yaml-editor" :class="{ 'is-fullscreen': isFullscreen }" :style="isFullscreen ? {} : { height: height }">
    <!-- Fullscreen toolbar (shown when fullscreen, even if showToolbar is false) -->
    <div class="yaml-editor-toolbar" v-if="isFullscreen">
      <div class="toolbar-left">
        <!-- saveable mode: Edit / Save+Cancel -->
        <template v-if="saveable">
          <el-button v-if="!isEditing" size="small" type="primary" @click="enterEdit">
            <el-icon><Edit /></el-icon> Edit
          </el-button>
          <template v-else>
            <el-button size="small" type="success" :loading="saving" @click="handleSave">
              <el-icon><Check /></el-icon> Save
            </el-button>
            <el-button size="small" @click="handleCancel">取消</el-button>
          </template>
        </template>
        <!-- showSaveButtons mode: Save + Cancel -->
        <template v-if="showSaveButtons && !saveable">
          <el-button size="small" type="success" :loading="saving" @click="emit('save')">保存</el-button>
          <el-button size="small" @click="emit('cancel')">取消</el-button>
        </template>
        <!-- fullscreen-actions slot (for create pages) -->
        <slot name="fullscreen-actions"></slot>
        <span v-if="title" class="toolbar-title">{{ title }}</span>
      </div>
      <div class="toolbar-center">
        <el-button-group>
          <el-button size="small" @click="handleFormat">Format</el-button>
          <el-button size="small" @click="handleCopy">复制</el-button>
        </el-button-group>
      </div>
      <div class="toolbar-right">
        <el-tooltip content="精简视图隐藏 status、resourceVersion 等系统字段" placement="top">
          <el-switch
            v-model="showSystemFields"
            size="small"
            inline-prompt
            active-text="完整"
            inactive-text="精简"
            :disabled="isDirty"
            class="fields-switch"
          />
        </el-tooltip>
        <el-tag v-if="isEditing" type="success" size="small" effect="plain">Editing</el-tag>
        <el-tag v-else-if="readOnly" type="info" size="small" effect="plain">Read-only</el-tag>
        <el-tooltip content="还原" placement="top">
          <el-icon class="toolbar-action" @click="toggleFullscreen">
            <ScaleToOriginal />
          </el-icon>
        </el-tooltip>
      </div>
    </div>

    <!-- Normal toolbar -->
    <div class="yaml-editor-toolbar" v-else-if="showToolbar">
      <!-- Left: Edit/Save/Cancel / SaveButtons -->
      <div class="toolbar-left">
        <template v-if="saveable">
          <el-button v-if="!isEditing" size="small" type="primary" @click="enterEdit">
            <el-icon><Edit /></el-icon> Edit
          </el-button>
          <template v-else>
            <el-button size="small" type="success" :loading="saving" @click="handleSave">
              <el-icon><Check /></el-icon> Save
            </el-button>
            <el-button size="small" @click="handleCancel">取消</el-button>
          </template>
        </template>
        <template v-if="showSaveButtons && !saveable">
          <el-button size="small" type="success" :loading="saving" @click="emit('save')">保存</el-button>
          <el-button size="small" @click="emit('cancel')">取消</el-button>
        </template>
        <span v-if="title" class="toolbar-title">{{ title }}</span>
      </div>

      <!-- Center: Format/Copy (when content is editable) -->
      <div class="toolbar-center" v-if="!readOnly || isEditing">
        <el-button-group>
          <el-button size="small" @click="handleFormat">Format</el-button>
          <el-button size="small" @click="handleCopy">复制</el-button>
        </el-button-group>
      </div>

      <!-- Right: Mode indicator + Fullscreen toggle -->
      <div class="toolbar-right">
        <el-tooltip content="精简视图隐藏 status、resourceVersion 等系统字段" placement="top">
          <el-switch
            v-model="showSystemFields"
            size="small"
            inline-prompt
            active-text="完整"
            inactive-text="精简"
            :disabled="isDirty"
            class="fields-switch"
          />
        </el-tooltip>
        <el-tag v-if="isEditing" type="success" size="small" effect="plain">Editing</el-tag>
        <el-tag v-else-if="readOnly" type="info" size="small" effect="plain">Read-only</el-tag>
        <el-tooltip :content="isFullscreen ? '还原' : '最大化'" placement="top">
          <el-icon class="toolbar-action" @click="toggleFullscreen">
            <ScaleToOriginal v-if="isFullscreen" />
            <FullScreen v-else />
          </el-icon>
        </el-tooltip>
      </div>
    </div>

    <MonacoEditor
      :value="displayValue"
      :options="editorOptions"
      language="yaml"
      style="flex: 1; min-height: 0;"
      @update:value="handleChange"
      @mount="handleEditorMount"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { Editor as MonacoEditor } from '@guolao/vue-monaco-editor'
import { ElMessage } from 'element-plus'
import { Edit, Check, FullScreen, ScaleToOriginal } from '@element-plus/icons-vue'
import yaml from 'js-yaml'

const props = withDefaults(defineProps<{
  modelValue: string
  height?: string
  editable?: boolean
  readOnly?: boolean
  autoFormat?: boolean
  title?: string
  saveable?: boolean
  showToolbar?: boolean
  showSaveButtons?: boolean
  saving?: boolean
}>(), {
  height: '400px',
  editable: false,
  readOnly: false,
  autoFormat: false,
  title: '',
  saveable: false,
  showToolbar: true,
  showSaveButtons: false,
  saving: false,
})

const emit = defineEmits(['update:modelValue', 'save', 'cancel'])

// Internal editing state
const isEditing = ref(!props.readOnly && props.saveable)
const originalContent = ref('')
const saving = ref(false)
const displayValue = ref('')
const isFullscreen = ref(false)
// 精简/完整视图切换：默认精简（隐藏 status 及系统元数据）
const showSystemFields = ref(false)
// 用户是否已修改编辑器内容（脏标记）。为脏时禁用视图切换，避免丢失编辑。
const isDirty = ref(false)
// 记录本组件自身最近一次回写的值，用于在 watch 中识别“自己触发的更新”并跳过重渲染
let lastEmitted: string | null = null

const dumpOptions = { indent: 2, lineWidth: 120, noRefs: true, sortKeys: false } as const

function emitModel(v: string) {
  lastEmitted = v
  emit('update:modelValue', v)
}

// 删除仅供服务端展示、编辑时无意义的系统字段
function stripSystemFields(obj: any) {
  if (!obj || typeof obj !== 'object') return
  delete obj.status
  const md = obj.metadata
  if (md && typeof md === 'object') {
    for (const k of ['resourceVersion', 'uid', 'creationTimestamp', 'generation', 'selfLink']) {
      delete md[k]
    }
  }
}

// 根据传入内容 + 当前视图模式，渲染 Monaco 中显示的文本
function renderDisplay(source?: string) {
  const val = source ?? props.modelValue ?? ''
  if (!val) {
    displayValue.value = ''
    return
  }
  // 完整视图且不要求自动格式化：原样显示
  if (showSystemFields.value && !props.autoFormat) {
    displayValue.value = val
    return
  }
  try {
    const parsed = yaml.load(val)
    if (!showSystemFields.value) {
      stripSystemFields(parsed)
    }
    displayValue.value = yaml.dump(parsed, dumpOptions)
  } catch {
    // 解析失败则回退显示原文
    displayValue.value = val
  }
}

// modelValue 由父组件更新时（新数据/取消还原）重新渲染并清除脏标记；
// 若是本组件自身回写触发的，则跳过，避免打断用户输入。
watch(() => props.modelValue, (val) => {
  if (val === lastEmitted) {
    lastEmitted = null
    return
  }
  isDirty.value = false
  renderDisplay()
}, { immediate: true })

// 切换精简/完整视图（脏状态下开关禁用，故此处一定是未编辑状态）
watch(showSystemFields, () => {
  if (!isDirty.value) renderDisplay()
})

// Sync isEditing when readOnly prop changes (e.g., parent resets after save)
watch(() => props.readOnly, (val) => {
  if (val) {
    isEditing.value = false
    saving.value = false
  }
})

const editorOptions = computed(() => ({
  minimap: { enabled: false },
  fontSize: 13,
  lineNumbers: 'on',
  scrollBeyondLastLine: false,
  wordWrap: 'on',
  readOnly: props.saveable ? (!isEditing.value) : props.readOnly,
  automaticLayout: true,
  tabSize: 2,
}))

function enterEdit() {
  originalContent.value = props.modelValue
  isEditing.value = true
}

function handleSave() {
  saving.value = true
  emit('save', props.modelValue)
}

function handleCancel() {
  isEditing.value = false
  saving.value = false
  isDirty.value = false
  if (originalContent.value !== props.modelValue) {
    emitModel(originalContent.value)
  }
  renderDisplay(originalContent.value)
}

// Force Monaco to re-layout after dialog open animation
function handleEditorMount() {
  nextTick(() => {
    setTimeout(() => {
      window.dispatchEvent(new Event('resize'))
    }, 300)
  })
}

function handleChange(value: string) {
  isDirty.value = true
  emitModel(value)
}

function handleFormat() {
  try {
    const parsed = yaml.load(props.modelValue)
    const formatted = yaml.dump(parsed, dumpOptions)
    emitModel(formatted)
    displayValue.value = formatted
    ElMessage.success('Formatted')
  } catch (e: any) {
    ElMessage.error('Invalid YAML: ' + (e.message || 'Format failed'))
  }
}

function handleCopy() {
  navigator.clipboard.writeText(props.modelValue)
  ElMessage.success('Copied to clipboard')
}

function toggleFullscreen() {
  isFullscreen.value = !isFullscreen.value
  nextTick(() => {
    setTimeout(() => {
      window.dispatchEvent(new Event('resize'))
    }, 100)
  })
}

// Keyboard shortcuts
function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && isFullscreen.value) {
    e.preventDefault()
    toggleFullscreen()
    return
  }
  if (!isEditing.value && !props.showSaveButtons) return
  if ((e.ctrlKey || e.metaKey) && e.key === 's') {
    e.preventDefault()
    if (props.showSaveButtons && !props.saveable) {
      emit('save')
    } else if (props.saveable) {
      handleSave()
    }
  }
  if (e.key === 'Escape') {
    e.preventDefault()
    handleCancel()
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})

function resetSaving() {
  saving.value = false
}

// Expose saving state and utility functions for parent to control
defineExpose({ saving, resetSaving, handleFormat, handleCopy, toggleFullscreen })
</script>

<style scoped>
.yaml-editor {
  border: 1px solid var(--gk-color-border);
  border-radius: 4px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
.yaml-editor.is-fullscreen {
  position: fixed;
  inset: 0;
  z-index: 3000;
  border-radius: 0;
  border: none;
  background: #fff;
}
.toolbar-action {
  cursor: pointer;
  font-size: 16px;
  color: var(--gk-color-text-secondary);
  margin-left: 8px;
  transition: color 0.2s;
}
.toolbar-action:hover {
  color: var(--el-color-primary);
}
.yaml-editor-toolbar {
  padding: 6px 12px;
  background: var(--gk-neutral-100);
  border-bottom: 1px solid var(--gk-color-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.toolbar-left {
  display: flex;
  align-items: center;
  gap: 8px;
}
.toolbar-center {
  display: flex;
  align-items: center;
}
.toolbar-right {
  display: flex;
  align-items: center;
}
.toolbar-title {
  font-size: 13px;
  color: var(--gk-color-text-secondary);
  margin-left: 4px;
}
</style>
