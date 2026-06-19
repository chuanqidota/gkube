<template>
  <div class="yaml-editor">
    <div class="yaml-editor-toolbar">
      <!-- Left: Edit/Save/Cancel -->
      <div class="toolbar-left">
        <template v-if="saveable">
          <el-button v-if="!isEditing" size="small" type="primary" @click="enterEdit">
            <el-icon><Edit /></el-icon> Edit
          </el-button>
          <template v-else>
            <el-button size="small" type="success" :loading="saving" @click="handleSave">
              <el-icon><Check /></el-icon> Save
            </el-button>
            <el-button size="small" @click="handleCancel">Cancel</el-button>
          </template>
        </template>
        <span v-if="title" class="toolbar-title">{{ title }}</span>
      </div>

      <!-- Center: Format/Copy (edit mode only) -->
      <div class="toolbar-center" v-if="isEditing">
        <el-button-group>
          <el-button size="small" @click="handleFormat">Format</el-button>
          <el-button size="small" @click="handleCopy">Copy</el-button>
        </el-button-group>
      </div>

      <!-- Right: Mode indicator -->
      <div class="toolbar-right">
        <el-tag v-if="isEditing" type="success" size="small" effect="plain">Editing</el-tag>
        <el-tag v-else-if="readOnly" type="info" size="small" effect="plain">Read-only</el-tag>
      </div>
    </div>

    <MonacoEditor
      :value="displayValue"
      :options="editorOptions"
      language="yaml"
      :style="{ height: height }"
      @update:value="handleChange"
      @mount="handleEditorMount"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { Editor as MonacoEditor } from '@guolao/vue-monaco-editor'
import { ElMessage } from 'element-plus'
import { Edit, Check } from '@element-plus/icons-vue'
import yaml from 'js-yaml'

const props = withDefaults(defineProps<{
  modelValue: string
  height?: string
  editable?: boolean
  readOnly?: boolean
  autoFormat?: boolean
  title?: string
  saveable?: boolean
}>(), {
  height: '400px',
  editable: false,
  readOnly: false,
  autoFormat: false,
  title: '',
  saveable: false,
})

const emit = defineEmits(['update:modelValue', 'save'])

// Internal editing state
const isEditing = ref(false)
const originalContent = ref('')
const saving = ref(false)
const displayValue = ref('')

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
  }
})

const editorOptions = computed(() => ({
  minimap: { enabled: false },
  fontSize: 13,
  lineNumbers: 'on',
  scrollBeyondLastLine: false,
  wordWrap: 'on',
  readOnly: !isEditing.value && (props.readOnly || !props.editable),
  automaticLayout: true,
  tabSize: 2,
}))

function enterEdit() {
  originalContent.value = props.modelValue
  isEditing.value = true
}

function handleSave() {
  emit('save', props.modelValue)
}

function handleCancel() {
  isEditing.value = false
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

// Keyboard shortcuts
function handleKeydown(e: KeyboardEvent) {
  if (!isEditing.value) return
  if ((e.ctrlKey || e.metaKey) && e.key === 's') {
    e.preventDefault()
    handleSave()
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

// Expose saving state for parent to control
defineExpose({ saving })
</script>

<style scoped>
.yaml-editor {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  overflow: hidden;
}
.yaml-editor-toolbar {
  padding: 6px 12px;
  background: #f5f7fa;
  border-bottom: 1px solid #dcdfe6;
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
  color: #909399;
  margin-left: 4px;
}
</style>
