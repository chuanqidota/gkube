<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'

const props = withDefaults(defineProps<{
  visible: boolean
  resourceName: string
  namespace: string
  name: string
  currentReplicas: number
  readyReplicas?: number
  scaleFn: (data: { namespace: string; name: string; replicas: number }) => Promise<any>
}>(), {
  readyReplicas: undefined,
})

const emit = defineEmits<{
  'update:visible': [value: boolean]
  'scaled': []
}>()

const replicas = ref(1)
const loading = ref(false)

watch(() => props.visible, (val) => {
  if (val) replicas.value = props.currentReplicas
})

async function handleConfirm() {
  loading.value = true
  try {
    await props.scaleFn({ namespace: props.namespace, name: props.name, replicas: replicas.value })
    ElMessage.success(`已扩缩容至 ${replicas.value} 副本`)
    emit('update:visible', false)
    emit('scaled')
  } catch (e: any) {
    ElMessage.error(e?.message || '扩缩容失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <el-dialog
    :model-value="visible"
    @update:model-value="emit('update:visible', $event)"
    :title="`扩缩容 ${resourceName}`"
    width="480px"
    destroy-on-close
  >
    <div>
      <p style="margin-bottom: var(--gk-space-4);">调整 <strong>{{ name }}</strong> 副本数</p>
      <el-descriptions :column="1" border size="small" style="margin-bottom: var(--gk-space-4);">
        <el-descriptions-item label="当前">{{ currentReplicas }}</el-descriptions-item>
        <el-descriptions-item v-if="readyReplicas !== undefined" label="就绪">{{ readyReplicas }}</el-descriptions-item>
      </el-descriptions>
      <el-form-item label="目标">
        <el-input-number v-model="replicas" :min="0" :max="10000" style="width: 200px;" />
      </el-form-item>
      <el-alert v-if="replicas === 0" title="设为 0 将停止所有 Pod。" type="warning" :closable="false" show-icon style="margin-top: 8px;" />
    </div>
    <template #footer>
      <el-button @click="emit('update:visible', false)">取消</el-button>
      <el-button type="primary" :loading="loading" @click="handleConfirm">确认</el-button>
    </template>
  </el-dialog>
</template>
