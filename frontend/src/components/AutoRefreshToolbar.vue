<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Refresh, Timer } from '@element-plus/icons-vue'

const { t } = useI18n()

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
    <el-button @click="emit('refresh')" :loading="loading" :icon="Refresh" size="small">
      {{ t('common.refresh') }}
    </el-button>
    <el-button
      :type="isRunning ? 'success' : 'default'"
      @click="emit('toggle')"
      size="small"
    >
      <el-icon><Timer /></el-icon>
      {{ isRunning ? t('common.autoRefreshOn', { n: countdown }) : t('common.autoRefreshOff') }}
    </el-button>
    <el-select
      :model-value="currentInterval / 1000"
      @update:model-value="handleIntervalChange"
      style="width: 90px;"
      size="small"
    >
      <el-option
        v-for="sec in availableIntervals"
        :key="sec"
        :value="sec"
        :label="`${sec}s`"
      />
    </el-select>
  </div>
</template>

<style scoped>
.auto-refresh-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
