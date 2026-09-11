<script setup lang="ts">
import { ref, computed } from 'vue'
import { Delete, ArrowDown } from '@element-plus/icons-vue'
import { getAvailableLabels } from '@/api/label'

// ============ 导出类型 ============

export interface LabelCondition {
  key: string
  operator: '=' | '!=' | 'in' | 'notin'
  values: string[]
}

// ============ 内部编辑状态 ============

interface EditingCondition {
  key: string
  operator: '=' | '!=' | 'in' | 'notin'
  values: string[]
}

// ============ Props & Emits ============

interface Props {
  clusterName: string
  namespace?: string
  resourceType: string
  modelValue: LabelCondition[]
}

const props = withDefaults(defineProps<Props>(), {
  namespace: '',
})

const emit = defineEmits<{
  'update:modelValue': [conditions: LabelCondition[]]
}>()

// ============ 状态 ============

const popoverVisible = ref(false)
const editingConditions = ref<EditingCondition[]>([])
/** 被编辑的 modelValue 索引（仅在 apply 时替换，不在编辑时移除） */
const editingIndices = ref<Set<number>>(new Set())
const availableKeys = ref<string[]>([])
const availableValues = ref<Record<string, string[]>>({})
const loading = ref(false)
const loadError = ref(false)

// ============ 类型转换 ============

function convertToEditing(conditions: LabelCondition[]): EditingCondition[] {
  return conditions.map((c) => ({
    key: c.key,
    operator: c.operator,
    values: [...c.values],
  }))
}

function convertToLabelConditions(editing: EditingCondition[]): LabelCondition[] {
  return editing
    .filter((c) => c.key && c.values.length > 0 && c.values.some((v) => v !== ''))
    .map((c) => ({
      key: c.key,
      operator: c.operator,
      values: [...c.values],
    }))
}

// ============ 计算属性 ============

const canAdd = computed(() => {
  if (editingConditions.value.length === 0) return false
  const last = editingConditions.value[editingConditions.value.length - 1]
  if (!last.key) return false
  if (last.operator === '=' || last.operator === '!=') {
    return last.values.length > 0 && last.values[0] !== ''
  }
  return last.values.length > 0
})

// ============ 数据加载 ============

async function fetchLabels() {
  loading.value = true
  loadError.value = false
  try {
    const res = await getAvailableLabels({
      clusterName: props.clusterName,
      namespace: props.namespace || undefined,
      resourceType: props.resourceType,
    })
    availableKeys.value = res.data?.keys || []
    availableValues.value = res.data?.values || {}
  } catch {
    loadError.value = true
    availableKeys.value = []
    availableValues.value = {}
  } finally {
    loading.value = false
  }
}

// ============ 交互方法 ============

function getValuesForKey(key: string): string[] {
  return availableValues.value[key] || []
}

function onKeyChange(cond: EditingCondition) {
  // 切换 key 时清空 value
  cond.values = []
}

function addCondition() {
  editingConditions.value.push({
    key: '',
    operator: '=',
    values: [],
  })
}

function editCondition(index: number) {
  const cond = props.modelValue[index]
  editingConditions.value.push(convertToEditing([cond])[0])
  // 记录被编辑的索引，apply 时替换而非重复
  editingIndices.value.add(index)
}

function removeCondition(index: number) {
  const updated = [...props.modelValue]
  updated.splice(index, 1)
  emit('update:modelValue', updated)
}

function clearAll() {
  emit('update:modelValue', [])
  editingConditions.value = [{ key: '', operator: '=', values: [] }]
}

function formatCondition(cond: LabelCondition): string {
  if (!cond.values || cond.values.length === 0) return cond.key
  if (cond.operator === '=' || cond.operator === '!=') {
    return `${cond.key}${cond.operator}${cond.values[0]}`
  }
  return `${cond.key} ${cond.operator} (${cond.values.join(', ')})`
}

function apply() {
  // 保留未被编辑的已应用条件
  const kept = props.modelValue.filter((_, i) => !editingIndices.value.has(i))
  const newConditions = convertToLabelConditions(editingConditions.value)

  // 去重：合并相同 key+operator+values 的条件（用 sorted copy 比较，不修改原数组）
  const merged = [...kept]
  for (const nc of newConditions) {
    const key = `${nc.key}|${nc.operator}|${[...nc.values].sort().join(',')}`
    const isDuplicate = merged.some((m) => {
      const mk = `${m.key}|${m.operator}|${[...m.values].sort().join(',')}`
      return mk === key
    })
    if (!isDuplicate) {
      merged.push(nc)
    }
  }

  emit('update:modelValue', merged)
  editingIndices.value = new Set()
  popoverVisible.value = false
}

function cancel() {
  popoverVisible.value = false
}

function onPopoverVisibleChange(visible: boolean) {
  popoverVisible.value = visible
  if (visible) {
    // 打开时加载数据并初始化编辑状态
    fetchLabels()
    editingConditions.value = convertToEditing(props.modelValue)
    editingIndices.value = new Set()
    if (editingConditions.value.length === 0) {
      addCondition()
    }
  } else {
    // 关闭时丢弃未应用的修改
    editingConditions.value = []
    editingIndices.value = new Set()
  }
}

function onEnter(e: KeyboardEvent) {
  const target = e.target as HTMLElement
  // 只在 Value 输入框聚焦时触发添加
  if (!target.closest('.value-select')) return
  if (canAdd.value) {
    addCondition()
  }
}

function onEsc() {
  popoverVisible.value = false
}
</script>

<template>
  <el-popover
    :visible="popoverVisible"
    placement="bottom-start"
    :width="480"
    trigger="click"
    :show-arrow="false"
    :offset="4"
    @update:visible="onPopoverVisibleChange"
  >
    <template #reference>
      <el-button :class="{ 'has-filters': modelValue.length > 0 }">
        Label
        <el-badge
          v-if="modelValue.length"
          :value="modelValue.length"
          :max="9"
          class="label-badge"
        />
        <el-icon class="el-icon--right"><ArrowDown /></el-icon>
      </el-button>
    </template>

    <div class="filter-popover" @keydown.esc="onEsc" @keydown.enter="onEnter">
      <!-- 加载失败警告 -->
      <el-alert v-if="loadError" type="warning" :closable="false" show-icon class="load-error">
        获取标签失败，您可以手动输入标签键值
      </el-alert>

      <!-- 已选条件 -->
      <div v-if="modelValue.length" class="selected-filters">
        <div class="tags-scroll">
          <el-tag
            v-for="(cond, i) in modelValue"
            :key="i"
            closable
            size="small"
            class="condition-tag"
            @click="editCondition(i)"
            @close="removeCondition(i)"
          >
            {{ formatCondition(cond) }}
          </el-tag>
        </div>
        <div v-if="modelValue.length > 5" class="more-hint">共 {{ modelValue.length }} 个条件</div>
      </div>

      <el-divider v-if="modelValue.length" />

      <!-- 无可用标签提示 -->
      <div v-if="!loading && !loadError && availableKeys.length === 0" class="empty-state">
        <p class="empty-text">未发现可自动补全的标签</p>
        <p class="empty-hint">您可以手动输入标签键值</p>
      </div>

      <!-- 条件编辑行 -->
      <div v-for="(cond, i) in editingConditions" :key="i" class="condition-row">
        <el-select
          v-model="cond.key"
          filterable
          allow-create
          placeholder="Key"
          size="small"
          class="key-select"
          :loading="loading"
          :teleported="false"
          @change="onKeyChange(cond)"
        >
          <el-option v-if="loading" label="加载中..." value="" disabled />
          <el-option v-for="k in availableKeys" :key="k" :label="k" :value="k" />
        </el-select>

        <el-select v-model="cond.operator" size="small" class="op-select" :teleported="false">
          <el-option label="=" value="=" />
          <el-option label="!=" value="!=" />
          <el-option label="in" value="in" />
          <el-option label="notin" value="notin" />
        </el-select>

        <!-- = / != 单选 -->
        <el-select
          v-if="cond.operator === '=' || cond.operator === '!='"
          :model-value="cond.values[0] || ''"
          filterable
          allow-create
          placeholder="Value"
          size="small"
          class="value-select"
          :loading="loading"
          :teleported="false"
          @update:model-value="cond.values = $event ? [$event] : []"
        >
          <el-option v-if="loading" label="加载中..." value="" disabled />
          <el-option v-for="v in getValuesForKey(cond.key)" :key="v" :label="v" :value="v" />
        </el-select>

        <!-- in / notin 多选 -->
        <el-select
          v-else
          v-model="cond.values"
          filterable
          allow-create
          multiple
          placeholder="Value"
          size="small"
          class="value-select"
          :loading="loading"
          :teleported="false"
        >
          <el-option v-if="loading" label="加载中..." value="" disabled />
          <el-option v-for="v in getValuesForKey(cond.key)" :key="v" :label="v" :value="v" />
        </el-select>

        <el-button
          text
          size="small"
          :disabled="editingConditions.length <= 1 && modelValue.length === 0"
          @click="editingConditions.splice(i, 1)"
        >
          <el-icon><Delete /></el-icon>
        </el-button>
      </div>

      <!-- 操作按钮 -->
      <div class="popover-actions">
        <el-button size="small" text @click="addCondition">+ 添加条件</el-button>
        <el-button
          size="small"
          text
          :disabled="modelValue.length === 0 && editingConditions.every((c) => !c.key)"
          @click="clearAll"
          >清除全部</el-button
        >
        <div class="spacer" />
        <el-button size="small" @click="cancel">取消</el-button>
        <el-button size="small" type="primary" @click="apply">应用</el-button>
      </div>
    </div>
  </el-popover>
</template>

<style scoped>
.filter-popover {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.load-error {
  margin-bottom: 4px;
}

.selected-filters {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.tags-scroll {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  max-height: 120px;
  overflow-y: auto;
}

.condition-tag {
  cursor: pointer;
}

.condition-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.key-select {
  width: 140px;
  flex-shrink: 0;
}

.op-select {
  width: 80px;
  flex-shrink: 0;
}

.value-select {
  flex: 1;
  min-width: 120px;
}

.popover-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-top: 8px;
  border-top: 1px solid var(--el-border-color-lighter);
}

.spacer {
  flex: 1;
}

.empty-state {
  text-align: center;
  padding: 12px 0;
}

.empty-text {
  color: var(--el-text-color-regular);
  margin: 0;
}

.empty-hint {
  color: var(--el-text-color-secondary);
  font-size: 12px;
  margin: 4px 0 0;
}

.more-hint {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.has-filters {
  color: var(--el-color-primary);
  border-color: var(--el-color-primary);
}

.label-badge {
  margin-left: 4px;
}

.label-badge :deep(.el-badge__content) {
  transform: translateY(-2px) translateX(4px);
}
</style>
