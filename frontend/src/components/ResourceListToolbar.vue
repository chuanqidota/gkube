<script setup lang="ts">
import { Search } from '@element-plus/icons-vue'

interface Props {
  searchValue: string
  namespaceValue?: string
  namespaceList?: string[]
  totalCount?: number
  selectedCount?: number
  showCreate?: boolean
  showNamespace?: boolean
  showTotalCount?: boolean
  searchPlaceholder?: string
  namespacePlaceholder?: string
}

const props = withDefaults(defineProps<Props>(), {
  namespaceValue: '',
  namespaceList: () => [],
  selectedCount: 0,
  showCreate: true,
  showNamespace: true,
  showTotalCount: true,
  searchPlaceholder: '搜索名称',
  namespacePlaceholder: '所有命名空间',
})

const emit = defineEmits<{
  'update:searchValue': [value: string]
  'update:namespaceValue': [value: string]
  searchInput: [value: string]
  namespaceChange: [value: string]
  create: []
  batchDelete: []
}>()
</script>

<template>
  <el-card shadow="never" class="filter-card">
    <div class="filter-bar">
      <!-- 左侧筛选区 -->
      <el-input
        :model-value="searchValue"
        @input="emit('searchInput', $event)"
        :placeholder="searchPlaceholder"
        style="width: 220px;"
        clearable
      >
        <template #prefix><el-icon><Search /></el-icon></template>
      </el-input>
      <el-select
        v-if="showNamespace"
        :model-value="namespaceValue"
        @update:model-value="emit('update:namespaceValue', $event)"
        :placeholder="namespacePlaceholder"
        clearable
        style="width: 180px;"
        @change="emit('namespaceChange', $event)"
      >
        <el-option v-for="ns in namespaceList" :key="ns" :label="ns" :value="ns" />
      </el-select>
      <!-- 总计数 -->
      <span class="total-count" v-if="showTotalCount && totalCount">总计: {{ totalCount }}</span>

      <!-- 右侧操作区（推到最右） -->
      <div class="right-actions">
        <!-- 竖线分隔符 -->
        <div class="action-divider" />

        <!-- 主要操作组（连接式按钮组） -->
        <div class="action-group">
          <slot name="actions" />
        </div>

        <!-- 竖线分隔符 -->
        <div class="action-divider" />

        <!-- 辅助工具区 -->
        <slot name="extra" />
      </div>
    </div>
  </el-card>
</template>

<style scoped>
.filter-card {
  margin-bottom: 16px;
}
.filter-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.total-count {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
.right-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: auto;
}
.action-divider {
  width: 1px;
  height: 20px;
  background: var(--el-border-color-lighter);
  margin: 0 4px;
}
.action-group {
  display: inline-flex;
}
.action-group :deep(.el-button) {
  border-radius: 0;
  margin-left: -1px;
}
.action-group :deep(.el-button:first-child) {
  border-radius: 4px 0 0 4px;
  margin-left: 0;
}
.action-group :deep(.el-button:last-child) {
  border-radius: 0 4px 4px 0;
}
</style>
