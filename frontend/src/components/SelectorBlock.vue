<script setup lang="ts">
interface Props {
  selector: Record<string, string>
}

defineProps<Props>()
</script>

<template>
  <div class="selector-block">
    <div v-if="!selector || Object.keys(selector).length === 0" class="empty-text">无</div>
    <div v-else class="label-list">
      <el-tag
        v-for="(val, key) in selector"
        :key="key"
        size="small"
        type="warning"
        class="label-tag"
      >
        {{ key }}={{ val }}
      </el-tag>
    </div>
  </div>
</template>

<style scoped>
.selector-block {
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
