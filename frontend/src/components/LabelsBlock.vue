<script setup lang="ts">
interface Props {
  labels: Record<string, string>
}

defineProps<Props>()
</script>

<template>
  <div class="labels-block">
    <div v-if="!labels || Object.keys(labels).length === 0" class="empty-text">无</div>
    <div v-else class="label-list">
      <el-tag v-for="(val, key) in labels" :key="key" size="small" type="info" class="label-tag">
        {{ key }}={{ val }}
      </el-tag>
    </div>
  </div>
</template>

<style scoped>
.labels-block {
  padding: var(--gk-space-1) 0;
}

.label-list {
  display: flex;
  flex-wrap: wrap;
  gap: var(--gk-space-1);
}

.label-tag {
  font-family: var(--gk-font-mono);
  font-size: 11px;
}

.empty-text {
  font-size: var(--gk-font-size-xs);
  color: var(--gk-color-text-placeholder);
}
</style>
