<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Delete, Plus } from '@element-plus/icons-vue'
import { updateNodeLabels } from '@/api/resource'

interface LabelEntry { key: string; value: string }

const emit = defineEmits<{ saved: [] }>()

const visible = ref(false)
const nodeName = ref('')
const labelsArray = ref<LabelEntry[]>([])

function open(name: string, current: Record<string, string>) {
  nodeName.value = name
  labelsArray.value = Object.entries(current || {}).map(([key, value]) => ({ key, value }))
  if (labelsArray.value.length === 0) labelsArray.value = [{ key: '', value: '' }]
  visible.value = true
}

function addLabel() { labelsArray.value.push({ key: '', value: '' }) }
function removeLabel(index: number) { labelsArray.value.splice(index, 1) }

async function handleSave() {
  try {
    const labelsMap: Record<string, string> = {}
    labelsArray.value.forEach(l => { if (l.key) labelsMap[l.key] = l.value })
    await updateNodeLabels({ name: nodeName.value, labels: labelsMap })
    ElMessage.success('标签已更新')
    visible.value = false
    emit('saved')
  } catch (e: any) {
    ElMessage.error(e?.message || '更新标签失败')
  }
}

defineExpose({ open })
</script>

<template>
  <el-dialog v-model="visible" title="管理标签" width="650px">
    <div v-for="(label, index) in labelsArray" :key="index" style="display: flex; gap: 8px; margin-bottom: 12px; align-items: center;">
      <el-input v-model="label.key" placeholder="Key" style="flex: 2;" />
      <el-input v-model="label.value" placeholder="Value" style="flex: 2;" />
      <el-button type="danger" circle size="small" @click="removeLabel(index)"><el-icon><Delete /></el-icon></el-button>
    </div>
    <el-button @click="addLabel" style="margin-top: 8px;"><el-icon><Plus /></el-icon> 添加标签</el-button>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="handleSave">保存</el-button>
    </template>
  </el-dialog>
</template>
