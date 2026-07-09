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

// Track a local formatted version for display
watch(() => props.modelValue, (val) => {
  if (props.autoFormat && val) {
    try {
      const parsed = yaml.load(val)
      const formatted = yaml.dump(parsed, {
        indent: 2,
        lineWidth: 120,
        noRefs: true,
        sortKeys: false,
      })
      displayValue.value = formatted
      if (formatted !== val) {
        emit('update:modelValue', formatted)
      }
    } catch {
      displayValue.value = val
    }
  } else {
    displayValue.value = val || ''
  }
}, { immediate: true })

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
  if (originalContent.value !== props.modelValue) {
    emit('update:modelValue', originalContent.value)
  }
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
  emit('update:modelValue', value)
}

function handleFormat() {
  try {
    const parsed = yaml.load(props.modelValue)
    const formatted = yaml.dump(parsed, {
      indent: 2,
      lineWidth: 120,
      noRefs: true,
      sortKeys: false,
    })
    emit('update:modelValue', formatted)
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
