<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { drainNode, type DrainResult } from '@/api/resource'

const emit = defineEmits<{ saved: [] }>()

const { t } = useI18n()

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
  drainOptions.value = {
    ignoreDaemonSets: true,
    deleteLocalData: false,
    gracePeriod: -1,
    force: false,
  }
  visible.value = true
}

async function handleConfirm() {
  try {
    await ElMessageBox.confirm(
      t('node.drainConfirmMsg', { name: nodeName.value }),
      t('node.confirmDrain'),
      {
        type: 'warning',
        confirmButtonText: t('node.confirmDrain'),
        cancelButtonText: t('common.cancel'),
      },
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
      t('node.drainSubmitted', { count: submitted }),
      skipped.length > 0 ? t('node.drainSkipped', { count: skipped.length }) : '',
      failed.length > 0 ? t('node.drainFailed', { count: failed.length }) : '',
    ].filter(Boolean)
    const type = failed.length > 0 ? 'warning' : 'success'
    const cordonedHint =
      failed.length > 0 ? t('node.drainCordonedPartialFailHint') : t('node.drainCordonedHint')
    ElMessage({
      type,
      message: `${t('node.drainResultSubmitted', { parts: parts.join(', ') })}${cordonedHint}`,
      duration: 8000,
    })
    visible.value = false
    emit('saved')
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || t('common.drainFailed'))
  }
}

defineExpose({ open })
</script>

<template>
  <el-dialog v-model="visible" :title="t('node.drainPod')" width="500px">
    <el-alert type="warning" :closable="false" style="margin-bottom: 16px">
      <template #title>{{ t('node.drainWarning') }}</template>
    </el-alert>
    <el-form label-width="160px">
      <el-form-item :label="t('node.ignoreDaemonSets')">
        <el-switch v-model="drainOptions.ignoreDaemonSets" />
        <span class="hint">{{ t('node.ignoreDaemonSetsHint') }}</span>
      </el-form-item>
      <el-form-item :label="t('node.deleteLocalData')">
        <el-switch v-model="drainOptions.deleteLocalData" />
        <span class="hint">{{ t('node.deleteLocalDataHint') }}</span>
      </el-form-item>
      <el-form-item :label="t('node.gracePeriod')">
        <el-input-number v-model="drainOptions.gracePeriod" :min="-1" :max="3600" />
        <span class="hint">{{ t('node.gracePeriodHint') }}</span>
      </el-form-item>
      <el-form-item :label="t('node.forceDrain')">
        <el-switch v-model="drainOptions.force" />
        <span class="hint">{{ t('node.forceDrainHint') }}</span>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">{{ t('common.cancel') }}</el-button>
      <el-button type="warning" @click="handleConfirm">{{ t('node.confirmDrain') }}</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.hint {
  margin-left: 8px;
  color: #909399;
  font-size: 12px;
}
</style>
