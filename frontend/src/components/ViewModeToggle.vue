<script setup lang="ts">
import { Grid, List } from '@element-plus/icons-vue'

type ViewMode = 'card' | 'table'

const props = defineProps<{
  modelValue: ViewMode
  storageKey?: string
}>()

const emit = defineEmits<{ 'update:modelValue': [mode: ViewMode] }>()

function set(mode: ViewMode) {
  emit('update:modelValue', mode)
  if (props.storageKey) {
    localStorage.setItem(props.storageKey, mode)
  }
}
</script>

<template>
  <el-button-group>
    <el-button
      :type="modelValue === 'card' ? 'primary' : ''"
      size="small"
      @click="set('card')"
    >
      <el-icon><Grid /></el-icon> 卡片
    </el-button>
    <el-button
      :type="modelValue === 'table' ? 'primary' : ''"
      size="small"
      @click="set('table')"
    >
      <el-icon><List /></el-icon> 表格
    </el-button>
  </el-button-group>
</template>
