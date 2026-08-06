<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Delete, Plus } from '@element-plus/icons-vue'
import { updateNodeTaints } from '@/api/resource'

interface Taint { key: string; value: string; effect: string }

const emit = defineEmits<{ saved: [] }>()

const visible = ref(false)
const nodeName = ref('')
const taints = ref<Taint[]>([])

const EFFECTS = ['NoSchedule', 'PreferNoSchedule', 'NoExecute']

function open(name: string, current: Taint[]) {
  nodeName.value = name
  taints.value = (current || []).map(t => ({ ...t }))
  if (taints.value.length === 0) taints.value = [{ key: '', value: '', effect: 'NoSchedule' }]
  visible.value = true
}

function addTaint() { taints.value.push({ key: '', value: '', effect: 'NoSchedule' }) }
function removeTaint(index: number) { taints.value.splice(index, 1) }

async function handleSave() {
  try {
    await updateNodeTaints({ name: nodeName.value, taints: taints.value.filter(t => t.key) })
    ElMessage.success('污点已更新')
    visible.value = false
    emit('saved')
  } catch (e: any) {
    ElMessage.error(e?.message || '更新污点失败')
  }
}

defineExpose({ open })
</script>

<template>
  <el-dialog v-model="visible" title="管理污点" width="600px">
    <div v-for="(taint, index) in taints" :key="index" style="display: flex; gap: 8px; margin-bottom: 12px; align-items: center;">
      <el-input v-model="taint.key" placeholder="Key" style="flex: 2;" />
      <el-input v-model="taint.value" placeholder="Value" style="flex: 1;" />
      <el-select v-model="taint.effect" style="flex: 1.5;">
        <el-option v-for="eff in EFFECTS" :key="eff" :label="eff" :value="eff" />
      </el-select>
      <el-button type="danger" circle size="small" @click="removeTaint(index)"><el-icon><Delete /></el-icon></el-button>
    </div>
    <el-button @click="addTaint" style="margin-top: 8px;"><el-icon><Plus /></el-icon> 添加污点</el-button>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="handleSave">保存</el-button>
    </template>
  </el-dialog>
</template>
