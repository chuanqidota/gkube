<script setup lang="ts">
import { formatAge } from '@/utils/helpers'

interface K8sCondition {
  type: string
  status: string
  lastTransitionTime?: string
  reason?: string
  message?: string
}

interface Props {
  conditions: K8sCondition[]
}

defineProps<Props>()

function conditionStatusType(status: string): '' | 'success' | 'warning' | 'danger' | 'info' {
  if (status === 'True') return 'success'
  if (status === 'False') return 'danger'
  return 'info'
}
</script>

<template>
  <div class="conditions-block">
    <div v-if="!conditions || conditions.length === 0" class="empty-text">无</div>
    <div v-else class="conditions-list">
      <div v-for="c in conditions" :key="c.type" class="condition-item">
        <div class="condition-head">
          <el-tag :type="conditionStatusType(c.status)" size="small">{{ c.status }}</el-tag>
          <span class="condition-type">{{ c.type }}</span>
        </div>
        <div v-if="c.reason || c.message" class="condition-msg">
          <span v-if="c.reason" class="condition-reason">{{ c.reason }}</span>
          <span v-if="c.message" class="condition-text">{{ c.message }}</span>
        </div>
        <div v-if="c.lastTransitionTime" class="condition-time">
          {{ formatAge(c.lastTransitionTime, false) }}
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.conditions-block {
  padding: var(--gk-space-1) 0;
}

.conditions-list {
  display: flex;
  flex-direction: column;
  gap: var(--gk-space-1);
}

.condition-item {
  padding: var(--gk-space-1) var(--gk-space-2);
  background: var(--gk-neutral-100);
  border-radius: var(--gk-radius-sm);
}

.condition-head {
  display: flex;
  align-items: center;
  gap: var(--gk-space-1);
}

.condition-type {
  font-size: var(--gk-font-size-xs);
  font-weight: 600;
  color: var(--gk-color-text-primary);
}

.condition-msg {
  font-size: 11px;
  margin-top: 2px;
  display: flex;
  gap: var(--gk-space-1);
}

.condition-reason {
  color: var(--gk-color-warning);
  flex-shrink: 0;
}

.condition-text {
  color: var(--gk-color-text-secondary);
  word-break: break-all;
}

.condition-time {
  font-size: 10px;
  color: var(--gk-color-text-placeholder);
  margin-top: 2px;
}

.empty-text {
  font-size: var(--gk-font-size-xs);
  color: var(--gk-color-text-placeholder);
}
</style>
