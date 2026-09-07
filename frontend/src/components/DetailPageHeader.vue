<script setup lang="ts">
import { ArrowLeft, Refresh, Timer } from '@element-plus/icons-vue'

interface StatusTag {
  text: string
  type: 'success' | 'warning' | 'danger' | 'info'
}

interface Props {
  title: string
  statusTag?: StatusTag
  namespace?: string
  loading?: boolean
  isRunning?: boolean
  countdown?: number
  currentInterval?: number
  availableIntervals?: number[]
}

defineProps<Props>()

defineEmits<{
  refresh: []
  toggle: []
  setIntervalOption: [interval: number]
  back: []
}>()
</script>

<template>
  <div class="page-header">
    <div class="header-left">
      <div class="title-row">
        <el-button :icon="ArrowLeft" link @click="$emit('back')" class="back-btn" />
        <h2 class="res-name">{{ title }}</h2>
        <el-tag v-if="statusTag" :type="statusTag.type" size="small" class="status-tag">
          {{ statusTag.text }}
        </el-tag>
      </div>
      <div class="meta-line">
        <span v-if="namespace" class="ns-tag">{{ namespace }}</span>
        <slot name="meta" />
      </div>
    </div>
    <div class="header-actions">
      <slot name="actions" />
      <div class="action-divider" />
      <slot name="extra" />
      <el-popover trigger="click" width="200">
        <template #reference>
          <el-button :icon="Timer" :type="isRunning ? 'success' : 'default'" size="small">
            {{ isRunning ? `${countdown}s` : '自动刷新' }}
          </el-button>
        </template>
        <div class="auto-refresh-popover">
          <span class="popover-title">刷新间隔</span>
          <el-radio-group
            :model-value="currentInterval"
            @update:model-value="(v: number) => $emit('setIntervalOption', v)"
            size="small"
          >
            <el-radio-button
              v-for="interval in availableIntervals"
              :key="interval"
              :value="interval"
            >
              {{ interval }}s
            </el-radio-button>
          </el-radio-group>
          <el-button
            size="small"
            :type="isRunning ? 'danger' : 'primary'"
            @click="$emit('toggle')"
          >
            {{ isRunning ? '停止' : '开始' }}
          </el-button>
        </div>
      </el-popover>
      <el-button :icon="Refresh" @click="$emit('refresh')" :loading="loading" size="small" />
    </div>
  </div>
</template>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--gk-space-3);
  flex-shrink: 0;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: var(--gk-space-1);
}

.title-row {
  display: flex;
  align-items: center;
  gap: var(--gk-space-2);
}

.back-btn {
  font-size: 18px;
}

.res-name {
  margin: 0;
  font-size: var(--gk-font-size-lg);
  font-weight: 600;
  line-height: 1.3;
}

.status-tag {
  font-size: var(--gk-font-size-xs);
}

.meta-line {
  display: flex;
  align-items: center;
  gap: var(--gk-space-2);
}

.ns-tag {
  font-size: var(--gk-font-size-xs);
  color: var(--gk-color-text-secondary);
  background: var(--gk-neutral-100);
  padding: 1px 6px;
  border-radius: var(--gk-radius-sm);
}

.header-actions {
  display: flex;
  flex-shrink: 0;
  align-items: center;
}

.header-actions :deep(.el-button) {
  border-radius: 0;
  margin-left: -1px;
}

.header-actions :deep(.el-button:first-child) {
  border-radius: var(--gk-radius-sm) 0 0 var(--gk-radius-sm);
  margin-left: 0;
}

.header-actions :deep(.el-button:last-of-type),
.header-actions :deep(.el-dropdown:last-of-type) {
  border-radius: 0 var(--gk-radius-sm) var(--gk-radius-sm) 0;
}

.action-divider {
  width: 1px;
  height: 20px;
  background: var(--gk-color-border-light);
  margin: 0 var(--gk-space-1);
}

.auto-refresh-popover {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.popover-title {
  font-size: var(--gk-font-size-sm);
  font-weight: 500;
  color: var(--gk-color-text-primary);
}
</style>
