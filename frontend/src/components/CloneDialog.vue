<script setup lang="ts">
/**
 * Reusable clone dialog for any K8s resource create page.
 *
 * Props (one-way, never mutated by this component):
 *   kindLabel         — display name for the resource
 *   modelValue        — dialog open state (v-model)
 *   namespaceScoped   — show namespace selector (default true; false for PV/StorageClass/...)
 *   showTargetChoice  — show form/yaml target radio (default true; false for YAML-only pages)
 *   nsValue / nameValue — current namespace/name (one-way)
 *   nsOptions / nameOptions — loaded name lists
 *   nsLoading / nameLoading / loading — loading states
 *   target            — clone target 'form' | 'yaml' (v-model:target)
 *
 * Emits:
 *   update:modelValue — dialog open/close
 *   update:nsValue    — namespace changed
 *   update:nameValue  — resource name changed
 *   update:target     — target changed
 *   load              — user clicked "加载"
 *   cancel            — user clicked "取消"
 */
const props = withDefaults(defineProps<{
  kindLabel: string
  modelValue: boolean
  namespaceScoped?: boolean
  showTargetChoice?: boolean
  nsValue?: string
  nsOptions?: string[]
  nsLoading?: boolean
  nameValue: string
  nameOptions: string[]
  nameLoading: boolean
  loading: boolean
  target?: 'form' | 'yaml'
}>(), {
  namespaceScoped: true,
  showTargetChoice: true,
  nsValue: '',
  nsOptions: () => [],
  nsLoading: false,
  target: 'form',
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'update:nsValue': [value: string]
  'update:nameValue': [value: string]
  'update:target': [value: 'form' | 'yaml']
  load: []
  cancel: []
}>()

function onNsChange(v: string) {
  emit('update:nsValue', v)
  emit('update:nameValue', '') // 命名空间切换时清空名称
}
</script>

<template>
  <el-dialog
    :model-value="modelValue"
    @update:model-value="emit('update:modelValue', $event)"
    :title="`从现有${kindLabel}克隆`"
    width="480px"
    destroy-on-close
  >
    <div>
      <p style="margin-bottom: 16px; color: var(--el-text-color-secondary); font-size: 13px;">
        选择已有的 {{ kindLabel }} 作为模板，自动填充创建表单
      </p>
      <el-form label-width="100px">
        <el-form-item v-if="namespaceScoped" label="命名空间">
          <el-select
            :model-value="nsValue"
            @update:model-value="(v: any) => onNsChange(v)"
            filterable
            :loading="nsLoading"
            style="width: 100%;"
          >
            <el-option v-for="ns in nsOptions" :key="ns" :label="ns" :value="ns" />
          </el-select>
        </el-form-item>
        <el-form-item label="资源名称">
          <el-select
            :model-value="nameValue"
            @update:model-value="(v: any) => emit('update:nameValue', v)"
            filterable
            :loading="nameLoading"
            style="width: 100%;"
            placeholder="选择或搜索..."
          >
            <el-option v-for="n in nameOptions" :key="n" :label="n" :value="n" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="showTargetChoice" label="填充到">
          <el-radio-group :model-value="target" @update:model-value="(v: any) => emit('update:target', v)">
            <el-radio value="form">表单</el-radio>
            <el-radio value="yaml">YAML 编辑器</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
    </div>
    <template #footer>
      <el-button @click="emit('cancel')">取消</el-button>
      <el-button type="primary" :loading="loading" @click="emit('load')">加载</el-button>
    </template>
  </el-dialog>
</template>