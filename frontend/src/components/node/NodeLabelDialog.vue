<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Delete, Plus } from '@element-plus/icons-vue'
import { updateNodeLabels } from '@/api/resource'
import { validateQualifiedName, validateLabelValue } from '@/utils/resource'

const emit = defineEmits<{ saved: [] }>()

const { t } = useI18n()

interface LabelEntry {
  key: string
  value: string
}

const visible = ref(false)
const nodeName = ref('')
const labelsArray = ref<LabelEntry[]>([])

function open(name: string, current: Record<string, string>) {
  nodeName.value = name
  labelsArray.value = Object.entries(current || {}).map(([key, value]) => ({ key, value }))
  if (labelsArray.value.length === 0) labelsArray.value = [{ key: '', value: '' }]
  visible.value = true
}

function addLabel() {
  labelsArray.value.push({ key: '', value: '' })
}
function removeLabel(index: number) {
  labelsArray.value.splice(index, 1)
}

// 校验所有 label：key 格式 + value 格式 + key 唯一。返回首个错误提示。
function validate(): string {
  const seen = new Set<string>()
  for (const l of labelsArray.value) {
    if (!l.key) continue // 空 key 行视为待删除，保存时过滤
    const keyErr = validateQualifiedName(l.key)
    if (keyErr) return t('node.labelKeyError', { key: l.key, error: keyErr })
    const valErr = validateLabelValue(l.value)
    if (valErr) return t('node.labelKeyError', { key: l.key, error: valErr })
    if (seen.has(l.key)) return t('node.labelDuplicate', { key: l.key })
    seen.add(l.key)
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
    const labelsMap: Record<string, string> = {}
    labelsArray.value.forEach((l) => {
      if (l.key) labelsMap[l.key] = l.value
    })
    await updateNodeLabels({ name: nodeName.value, labels: labelsMap })
    ElMessage.success(t('common.labelUpdateSuccess'))
    visible.value = false
    emit('saved')
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.labelUpdateFailed'))
  }
}

defineExpose({ open })
</script>

<template>
  <el-dialog v-model="visible" :title="t('node.manageLabels')" width="650px">
    <el-alert type="warning" :closable="false" show-icon style="margin-bottom: 16px">
      <template #title>{{ t('node.labelSaveWarning') }}</template>
    </el-alert>
    <div
      v-for="(label, index) in labelsArray"
      :key="index"
      style="display: flex; gap: 8px; margin-bottom: 12px; align-items: center"
    >
      <el-input v-model="label.key" :placeholder="t('node.labelKeyPlaceholder')" style="flex: 2" />
      <el-input
        v-model="label.value"
        :placeholder="t('node.labelValuePlaceholder')"
        style="flex: 2"
      />
      <el-button type="danger" circle size="small" @click="removeLabel(index)"
        ><el-icon><Delete /></el-icon
      ></el-button>
    </div>
    <el-button style="margin-top: 8px" @click="addLabel"
      ><el-icon><Plus /></el-icon> {{ t('node.addLabel') }}</el-button
    >
    <template #footer>
      <el-button @click="visible = false">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" @click="handleSave">{{ t('common.save') }}</el-button>
    </template>
  </el-dialog>
</template>
