<template>
  <div style="height: 60px; display: flex; align-items: center; padding: 0 20px; background: #fff">
    <el-icon style="cursor: pointer; font-size: 20px" @click="$emit('toggle-collapse')">
      <Fold />
    </el-icon>
  </div>
</template>

<script setup lang="ts">
import { Fold } from '@element-plus/icons-vue'

defineEmits<{
  'toggle-collapse': []
}>()
</script>
