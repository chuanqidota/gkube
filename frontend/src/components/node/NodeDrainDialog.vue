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
      `确定要驱逐节点 "${nodeName.value}" 上的所有 Pod 吗？此操作会先封锁节点再提交驱逐请求。`,
      '确认驱逐',
      { type: 'warning', confirmButtonText: '驱逐', cancelButtonText: '取消' },
    )
    const res = await drainNode({ name: nodeName.value, ...drainOptions.value })
    const result = (res.data || {}) as DrainResult
    const evicted = result.evicted || []
    const skipped = result.skipped || []
    const failed = result.failed || []
    // EvictV1 返回 nil 只代表驱逐请求被接受，pod 进入 terminating，并不保证已终止。
    // 文案须诚实：提示"已提交驱逐请求"，引导用户稍后刷新查看实际状态。
    // drain 已先封锁节点，若需恢复调度需手动解除封锁——尤其有失败时必须提示。
    const submitted = evicted.length + failed.length
    const parts = [
      `已提交 ${submitted} 个驱逐请求`,
      skipped.length > 0 ? `${skipped.length} 个已跳过` : '',
      failed.length > 0 ? `${failed.length} 个提交失败` : '',
    ].filter(Boolean)
    const type = failed.length > 0 ? 'warning' : 'success'
    const cordonedHint = failed.length > 0
      ? '节点已封锁且部分驱逐失败，如需恢复调度请先处理失败 Pod 再手动解除封锁。'
      : '节点已封锁，Pod 终止后如需恢复调度请手动解除封锁。'
    ElMessage({
      type,
      message: `驱逐请求已提交：${parts.join('，')}。Pod 正在终止中，请稍后刷新查看实际状态。${cordonedHint}`,
      duration: 8000,
    })
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
      <template #title>驱逐操作会先封锁节点，然后提交驱逐请求。Pod 进入终止需要时间，请稍后刷新查看实际状态。</template>
    </el-alert>
    <el-form label-width="160px">
      <el-form-item label="忽略 DaemonSet">
        <el-switch v-model="drainOptions.ignoreDaemonSets" />
        <span class="hint">跳过 DaemonSet 管理的 Pod</span>
      </el-form-item>
      <el-form-item label="删除本地数据">
        <el-switch v-model="drainOptions.deleteLocalData" />
        <span class="hint">删除使用 emptyDir/hostPath 的 Pod</span>
      </el-form-item>
      <el-form-item label="优雅终止时间(秒)">
        <el-input-number v-model="drainOptions.gracePeriod" :min="-1" :max="3600" />
        <span class="hint">-1 使用 Pod 默认值</span>
      </el-form-item>
      <el-form-item label="强制驱逐">
        <el-switch v-model="drainOptions.force" />
        <span class="hint">驱逐不被控制器管理的 standalone Pod（与 kubectl --force 一致）</span>
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
