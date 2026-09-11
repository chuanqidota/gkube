<script setup lang="ts">
import { computed } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import LabelFilterPopover from './LabelFilterPopover.vue'
import type { LabelCondition } from './LabelFilterPopover.vue'

const props = withDefaults(defineProps<Props>(), {
  namespaceValue: '',
  namespaceList: () => [],
  selectedCount: 0,
  showCreate: true,
  showNamespace: true,
  showTotalCount: true,
  searchPlaceholder: '',
  namespacePlaceholder: '',
  clusterName: '',
  resourceType: '',
  labelConditions: () => [],
})

const emit = defineEmits<{
  'update:searchValue': [value: string]
  'update:namespaceValue': [value: string]
  searchInput: [value: string]
  namespaceChange: [value: string]
  create: []
  batchDelete: []
  labelSelectorChange: [conditions: LabelCondition[]]
}>()

const { t } = useI18n()

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
  /** 集群名称（传给 LabelFilterPopover） */
  clusterName?: string
  /** 资源类型（传给 LabelFilterPopover） */
  resourceType?: string
  /** 当前 label 条件 */
  labelConditions?: LabelCondition[]
}

const searchPlaceholderText = computed(() => props.searchPlaceholder || t('common.searchName'))
const namespacePlaceholderText = computed(
  () => props.namespacePlaceholder || t('common.allNamespaces'),
)
</script>

<template>
  <el-card shadow="never" class="filter-card">
    <div class="filter-bar">
      <!-- 左侧筛选区 -->
      <el-select
        v-if="showNamespace"
        :model-value="namespaceValue"
        :placeholder="namespacePlaceholderText"
        clearable
        style="width: 180px"
        @update:model-value="emit('update:namespaceValue', $event)"
        @change="emit('namespaceChange', $event)"
      >
        <el-option v-for="ns in namespaceList" :key="ns" :label="ns" :value="ns" />
      </el-select>
      <el-input
        :model-value="searchValue"
        :placeholder="searchPlaceholderText"
        style="width: 220px"
        clearable
        @input="emit('searchInput', $event)"
      >
        <template #prefix
          ><el-icon><Search /></el-icon
        ></template>
      </el-input>
      <!-- Label 过滤 -->
      <LabelFilterPopover
        v-if="clusterName && resourceType"
        :cluster-name="clusterName"
        :namespace="namespaceValue"
        :resource-type="resourceType"
        :model-value="labelConditions"
        @update:model-value="emit('labelSelectorChange', $event)"
      />
      <!-- 总计数 -->
      <span v-if="showTotalCount && totalCount" class="total-count"
        >{{ t('common.total') }}: {{ totalCount }}</span
      >

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
  margin-bottom: var(--gk-space-4);
  border: 1px solid var(--gk-color-border-light);
}
.filter-bar {
  display: flex;
  align-items: center;
  gap: var(--gk-space-3);
  flex-wrap: wrap;
}
.total-count {
  color: var(--gk-color-text-secondary);
  font-size: var(--gk-font-size-sm);
}
.right-actions {
  display: flex;
  align-items: center;
  gap: var(--gk-space-1);
  margin-left: auto;
}
.action-divider {
  width: 1px;
  height: 20px;
  background: var(--gk-color-border-light);
  margin: 0 var(--gk-space-1);
}
.action-group {
  display: inline-flex;
}
.action-group :deep(.el-button) {
  border-radius: 0;
  margin-left: -1px;
}
.action-group :deep(.el-button:first-child) {
  border-radius: var(--gk-radius-md) 0 0 var(--gk-radius-md);
  margin-left: 0;
}
.action-group :deep(.el-button:last-child) {
  border-radius: 0 var(--gk-radius-md) var(--gk-radius-md) 0;
}
</style>
