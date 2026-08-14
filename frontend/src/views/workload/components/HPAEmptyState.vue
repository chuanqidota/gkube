<script setup lang="ts">
defineProps<{
  namespace: string
  workloadName: string
  workloadKind: string
}>()

const emit = defineEmits<{
  create: []
}>()
</script>

<template>
  <div class="hpa-empty-state">
    <div class="empty-icon">
      <svg width="64" height="64" viewBox="0 0 64 64" fill="none" xmlns="http://www.w3.org/2000/svg">
        <rect x="8" y="24" width="48" height="20" rx="4" stroke="currentColor" stroke-width="2" fill="none" opacity="0.3"/>
        <path d="M20 20 L20 12 L32 12 L32 20" stroke="currentColor" stroke-width="2" fill="none"/>
        <path d="M32 44 L32 52 L44 52 L44 44" stroke="currentColor" stroke-width="2" fill="none"/>
        <path d="M18 30 L22 34 L18 38" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round"/>
        <path d="M46 30 L42 34 L46 38" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round"/>
      </svg>
    </div>
    <h3 class="empty-title">此工作负载未配置弹性伸缩</h3>
    <p class="empty-desc">
      根据 CPU/Memory 等指标自动调整 Pod 副本数，适用于应对流量波动，保证服务可用性
    </p>
    <el-button type="primary" @click="emit('create')">
      + 创建弹性伸缩
    </el-button>
  </div>
</template>

<style scoped>
.hpa-empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
}

.empty-icon {
  color: var(--el-text-color-placeholder);
  margin-bottom: 16px;
}

.empty-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin: 0 0 8px;
}

.empty-desc {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin: 0 0 20px;
  max-width: 400px;
  line-height: 1.6;
}
</style>
