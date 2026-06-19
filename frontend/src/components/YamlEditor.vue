<template>
  <div class="yaml-editor">
    <div class="yaml-editor-toolbar" v-if="editable">
      <el-button-group>
        <el-button size="small" @click="handleFormat">Format</el-button>
        <el-button size="small" @click="handleCopy">Copy</el-button>
      </el-button-group>
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
import { computed, ref, watch, nextTick } from 'vue'
import { Editor as MonacoEditor } from '@guolao/vue-monaco-editor'
import { ElMessage } from 'element-plus'
import yaml from 'js-yaml'

const props = withDefaults(defineProps<{
  modelValue: string
  height?: string
  editable?: boolean
  readOnly?: boolean
  autoFormat?: boolean
}>(), {
  height: '400px',
  editable: false,
  readOnly: false,
  autoFormat: false,
})

const emit = defineEmits(['update:modelValue'])

// Track a local formatted version for display
const displayValue = ref('')

// When modelValue changes, auto-format if enabled
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
      // Emit formatted value back to parent
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

const editorOptions = computed(() => ({
  minimap: { enabled: false },
  fontSize: 13,
  lineNumbers: 'on',
  scrollBeyondLastLine: false,
  wordWrap: 'on',
  readOnly: props.readOnly || !props.editable,
  automaticLayout: true,
  tabSize: 2,
}))

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
</script>

<style scoped>
.yaml-editor {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  overflow: hidden;
}
.yaml-editor-toolbar {
  padding: 4px 8px;
  background: #f5f7fa;
  border-bottom: 1px solid #dcdfe6;
}
</style>
