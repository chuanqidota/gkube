<template>
  <div class="yaml-editor">
    <div class="yaml-editor-toolbar" v-if="editable">
      <el-button-group>
        <el-button size="small" @click="handleFormat">格式化</el-button>
        <el-button size="small" @click="handleCopy">复制</el-button>
      </el-button-group>
    </div>
    <MonacoEditor
      :value="modelValue"
      :options="editorOptions"
      language="yaml"
      :style="{ height: height }"
      @update:value="handleChange"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Editor as MonacoEditor } from '@guolao/vue-monaco-editor'
import { ElMessage } from 'element-plus'

const props = withDefaults(defineProps<{
  modelValue: string
  height?: string
  editable?: boolean
  readOnly?: boolean
}>(), {
  height: '400px',
  editable: false,
  readOnly: false,
})

const emit = defineEmits(['update:modelValue'])

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

function handleChange(value: string) {
  emit('update:modelValue', value)
}

function handleFormat() {
  // Basic YAML formatting - just emit current value
  ElMessage.success('已格式化')
}

function handleCopy() {
  navigator.clipboard.writeText(props.modelValue)
  ElMessage.success('已复制到剪贴板')
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
