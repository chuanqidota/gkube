<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import YamlEditor from './YamlEditor.vue'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    getYaml: (params: any) => Promise<any>
    updateYaml?: ((data: any) => Promise<any>) | null
    namespace?: string
    name: string
    title?: string
  }>(),
  {
    namespace: '',
    title: '',
    updateYaml: null,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  saved: []
}>()

const { t } = useI18n()

const loading = ref(false)
const saving = ref(false)
const yamlContent = ref('')

// Build params for getYaml
function buildGetYamlParams() {
  if (props.namespace) {
    return { namespace: props.namespace, name: props.name }
  }
  return { name: props.name }
}

// Build data for updateYaml
function buildUpdateData() {
  if (props.namespace) {
    return { namespace: props.namespace, name: props.name, yaml: yamlContent.value }
  }
  return { name: props.name, yaml: yamlContent.value }
}

// Load YAML
async function fetchYaml() {
  if (!props.name) return

  loading.value = true
  yamlContent.value = ''

  try {
    const params = buildGetYamlParams()
    const res = await props.getYaml(params)
    // 兼容两种后端返回格式：直接返回字符串 或 包装在 { yaml: "..." } 中
    const raw = res.data ?? res
    yamlContent.value = typeof raw === 'object' && raw?.yaml ? raw.yaml : raw
  } catch (error: any) {
    ElMessage.error(t('common.yamlLoadFailed') + ': ' + (error.message || t('common.unknown')))
  } finally {
    loading.value = false
  }
}

// Save YAML
async function handleSave() {
  if (!props.updateYaml) {
    ElMessage.warning(t('common.yamlReadOnly'))
    return
  }

  saving.value = true
  try {
    const data = buildUpdateData()
    await props.updateYaml(data)
    ElMessage.success(t('common.yamlSaveSuccess'))
    emit('saved')
    emit('update:modelValue', false)
  } catch (error: any) {
    ElMessage.error(t('common.yamlSaveFailed') + ': ' + (error.message || t('common.unknown')))
  } finally {
    saving.value = false
  }
}

// Cancel and reload
function handleCancel() {
  fetchYaml()
}

// Watch for drawer open
watch(
  () => props.modelValue,
  (visible) => {
    if (visible && props.name) {
      fetchYaml()
    }
  },
)

// Expose for parent to manually refresh
defineExpose({ fetchYaml })
</script>

<template>
  <el-drawer
    :model-value="modelValue"
    :title="title || 'YAML'"
    size="85%"
    direction="rtl"
    class="yaml-drawer"
    :body-style="{ padding: '0', height: '100%' }"
    :destroy-on-close="true"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <div v-loading="loading" style="height: calc(100dvh - 52px)">
      <YamlEditor
        v-if="!loading"
        v-model="yamlContent"
        height="100%"
        auto-format
        show-save-buttons
        :saving="saving"
        @save="handleSave"
        @cancel="handleCancel"
      />
    </div>
  </el-drawer>
</template>

<style>
.yaml-drawer .el-drawer__header {
  padding: 6px 16px;
  margin-bottom: 0;
  min-height: auto;
}
</style>
