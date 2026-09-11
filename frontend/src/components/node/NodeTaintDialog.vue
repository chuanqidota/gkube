<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Delete, Plus } from '@element-plus/icons-vue'
import { updateNodeTaints } from '@/api/resource'
import { validateQualifiedName, validateTaintEffect } from '@/utils/resource'

const emit = defineEmits<{ saved: [] }>()

const { t } = useI18n()

interface Taint {
  key: string
  value: string
  effect: string
}

const visible = ref(false)
const nodeName = ref('')
const taints = ref<Taint[]>([])

const EFFECTS = ['NoSchedule', 'PreferNoSchedule', 'NoExecute']

function open(name: string, current: Taint[]) {
  nodeName.value = name
  taints.value = (current || []).map((t) => ({ ...t }))
  if (taints.value.length === 0) taints.value = [{ key: '', value: '', effect: 'NoSchedule' }]
  visible.value = true
}

function addTaint() {
  taints.value.push({ key: '', value: '', effect: 'NoSchedule' })
}
function removeTaint(index: number) {
  taints.value.splice(index, 1)
}

// 校验所有 taint：key 格式 + effect 枚举 + key+effect 唯一。返回首个错误提示。
function validate(): string {
  const seen = new Set<string>()
  for (const taint of taints.value) {
    if (!taint.key) continue // 空 key 行视为待删除，保存时过滤
    const keyErr = validateQualifiedName(taint.key)
    if (keyErr) return t('node.taintKeyError', { key: taint.key, error: keyErr })
    const effErr = validateTaintEffect(taint.effect)
    if (effErr) return t('node.taintKeyError', { key: taint.key, error: effErr })
    const dedupKey = `${taint.key}/${taint.effect}`
    if (seen.has(dedupKey))
      return t('node.taintDuplicate', { key: taint.key, effect: taint.effect })
    seen.add(dedupKey)
  }
  return ''
}

async function handleSave() {
  const err = validate()
  if (err) {
    ElMessage.error(err)
    return
  }
  try {
    await updateNodeTaints({ name: nodeName.value, taints: taints.value.filter((t) => t.key) })
    ElMessage.success(t('common.taintUpdateSuccess'))
    visible.value = false
    emit('saved')
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.taintUpdateFailed'))
  }
}

defineExpose({ open })
</script>

<template>
  <el-dialog v-model="visible" :title="t('node.manageTaints')" width="600px">
    <div
      v-for="(taint, index) in taints"
      :key="index"
      style="display: flex; gap: 8px; margin-bottom: 12px; align-items: center"
    >
      <el-input v-model="taint.key" :placeholder="t('node.taintKeyPlaceholder')" style="flex: 2" />
      <el-input v-model="taint.value" :placeholder="t('common.value')" style="flex: 1" />
      <el-select v-model="taint.effect" style="flex: 1.5">
        <el-option v-for="eff in EFFECTS" :key="eff" :label="eff" :value="eff" />
      </el-select>
      <el-button type="danger" circle size="small" @click="removeTaint(index)"
        ><el-icon><Delete /></el-icon
      ></el-button>
    </div>
    <el-button style="margin-top: 8px" @click="addTaint"
      ><el-icon><Plus /></el-icon> {{ t('node.addTaint') }}</el-button
    >
    <template #footer>
      <el-button @click="visible = false">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" @click="handleSave">{{ t('common.save') }}</el-button>
    </template>
  </el-dialog>
</template>
