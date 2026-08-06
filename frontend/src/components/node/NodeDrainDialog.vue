<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { drainNode, type DrainResult } from '@/api/resource'

const emit = defineEmits<{ saved: [] }>()

const visible = ref(false)
const nodeName = ref('')
const drainOptions = ref({
  ignoreDaemonSets: true,
  deleteLocalData: false,
  gracePeriod: -1,
  force: false,
})

function open(name: string) {
  nodeName.value = name
  drainOptions.value = { ignoreDaemonSets: true, deleteLocalData: false, gracePeriod: -1, force: false }
  visible.value = true
}

async function handleConfirm() {
  try {
    await ElMessageBox.confirm(
      `确定要驱逐节点 "${nodeName.value}" 上的所有 Pod 吗？此操作会先封锁节点再驱逐 Pod。`,
      '确认驱逐',
      { type: 'warning', confirmButtonText: '驱逐', cancelButtonText: '取消' },
    )
    const res = await drainNode({ name: nodeName.value, ...drainOptions.value })
    const result = (res.data || {}) as DrainResult
    const evicted = result.evicted || []
    const skipped = result.skipped || []
    const failed = result.failed || []
    const parts = [
      `${evicted.length} 个 Pod 已驱逐`,
      `${skipped.length} 个已跳过`,
      `${failed.length} 个失败`,
    ]
    const type = failed.length > 0 ? 'warning' : 'success'
    ElMessage({ type, message: `驱逐完成：${parts.join('，')}` })
    visible.value = false
    emit('saved')
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '驱逐失败')
  }
}

defineExpose({ open })
</script>

<template>
  <el-dialog v-model="visible" title="驱逐 Pod" width="500px">
    <el-alert type="warning" :closable="false" style="margin-bottom: 16px;">
      <template #title>驱逐操作会先封锁节点，然后驱逐节点上的所有 Pod。请确认以下选项：</template>
    </el-alert>
    <el-form label-width="160px">
      <el-form-item label="忽略 DaemonSet">
        <el-switch v-model="drainOptions.ignoreDaemonSets" />
        <span class="hint">跳过 DaemonSet 管理的 Pod</span>
      </el-form-item>
      <el-form-item label="删除本地数据">
        <el-switch v-model="drainOptions.deleteLocalData" />
        <span class="hint">删除使用 emptyDir 的 Pod</span>
      </el-form-item>
      <el-form-item label="优雅终止时间(秒)">
        <el-input-number v-model="drainOptions.gracePeriod" :min="-1" :max="3600" />
        <span class="hint">-1 使用 Pod 默认值</span>
      </el-form-item>
      <el-form-item label="强制驱逐">
        <el-switch v-model="drainOptions.force" />
        <span class="hint">驱逐 kube-system 下的 Pod</span>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="warning" @click="handleConfirm">确认驱逐</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.hint { margin-left: 8px; color: #909399; font-size: 12px; }
</style>
