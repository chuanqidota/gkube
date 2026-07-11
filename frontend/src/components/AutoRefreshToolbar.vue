<script setup lang="ts">
import { Refresh, Timer } from '@element-plus/icons-vue'

interface Props {
  isRunning: boolean
  countdown: number
  currentInterval: number
  availableIntervals?: number[]
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  availableIntervals: () => [5, 10, 15, 30, 60],
  loading: false,
})

const emit = defineEmits<{
  refresh: []
  toggle: []
  'interval-change': [seconds: number]
}>()

function handleIntervalChange(seconds: number) {
  emit('interval-change', seconds)
}
</script>

<template>
  <div class="auto-refresh-toolbar">
    <!-- 自动刷新按钮（图标 + popover） -->
    <el-popover placement="bottom" :width="200" trigger="hover">
      <template #reference>
        <el-button
          :type="isRunning ? 'success' : 'default'"
          :icon="Timer"
          @click="emit('toggle')"
        />
      </template>
      <div class="auto-refresh-popover">
        <div class="popover-title">
          {{ isRunning ? `自动刷新中 ${countdown}s` : '自动刷新' }}
        </div>
        <el-select
          :model-value="currentInterval / 1000"
          @update:model-value="handleIntervalChange"
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
    <!-- 手动刷新按钮（图标 + tooltip） -->
    <el-tooltip content="刷新" placement="top">
      <el-button @click="emit('refresh')" :loading="loading" :icon="Refresh" />
    </el-tooltip>
  </div>
</template>

<style scoped>
.auto-refresh-toolbar {
  display: flex;
  align-items: center;
  gap: 4px;
}
.auto-refresh-popover {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.popover-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}
</style>
