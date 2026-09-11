<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  visible: boolean
  resourceName: string
  namespace: string
  name: string
  containers: Array<{ name: string; image: string }>
  updateImageFn: (data: {
    namespace: string
    name: string
    containerName: string
    image: string
  }) => Promise<any>
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  updated: []
}>()

const { t } = useI18n()

const form = ref({ containerName: '', image: '' })
const loading = ref(false)

watch(
  () => props.visible,
  (val) => {
    if (val && props.containers.length > 0) {
      form.value = {
        containerName: props.containers[0].name,
        image: props.containers[0].image,
      }
    }
  },
)

function onContainerChange() {
  const c = props.containers.find((c) => c.name === form.value.containerName)
  if (c) form.value.image = c.image
}

async function handleConfirm() {
  if (!form.value.containerName || !form.value.image) {
    ElMessage.warning(t('workload.selectContainerAndImage'))
    return
  }
  loading.value = true
  try {
    await props.updateImageFn({
      namespace: props.namespace,
      name: props.name,
      containerName: form.value.containerName,
      image: form.value.image,
    })
    ElMessage.success(t('workload.imageUpdateSuccess'))
    emit('update:visible', false)
    emit('updated')
  } catch (e: any) {
    ElMessage.error(e?.message || t('workload.imageUpdateFailed'))
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <el-dialog
    :model-value="visible"
    :title="`${t('workload.updateImage')} - ${resourceName}`"
    width="520px"
    destroy-on-close
    @update:model-value="emit('update:visible', $event)"
  >
    <el-form label-width="100px">
      <el-form-item :label="t('workload.containers')">
        <el-select v-model="form.containerName" style="width: 100%" @change="onContainerChange">
          <el-option v-for="c in containers" :key="c.name" :label="c.name" :value="c.name" />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('workload.image')">
        <el-input v-model="form.image" placeholder="nginx:latest" />
      </el-form-item>
    </el-form>
    <el-alert
      v-if="form.containerName"
      :title="`${t('workload.currentImage')}: ${containers.find((c) => c.name === form.containerName)?.image || '-'}`"
      type="info"
      :closable="false"
      style="margin-top: 8px"
    />
    <template #footer>
      <el-button @click="emit('update:visible', false)">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" :loading="loading" @click="handleConfirm">{{
        t('common.confirm')
      }}</el-button>
    </template>
  </el-dialog>
</template>
