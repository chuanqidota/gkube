<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'

const props = defineProps<{
  visible: boolean
  resourceName: string
  namespace: string
  name: string
  containers: Array<{ name: string; image: string }>
  updateImageFn: (data: { namespace: string; name: string; containerName: string; image: string }) => Promise<any>
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  'updated': []
}>()

const form = ref({ containerName: '', image: '' })
const loading = ref(false)

watch(() => props.visible, (val) => {
  if (val && props.containers.length > 0) {
    form.value = {
      containerName: props.containers[0].name,
      image: props.containers[0].image,
    }
  }
})

function onContainerChange() {
  const c = props.containers.find(c => c.name === form.value.containerName)
  if (c) form.value.image = c.image
}

async function handleConfirm() {
  if (!form.value.containerName || !form.value.image) {
    ElMessage.warning('请选择容器并填写镜像')
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
    ElMessage.success('镜像更新成功')
    emit('update:visible', false)
    emit('updated')
  } catch (e: any) {
    ElMessage.error(e?.message || '镜像更新失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <el-dialog
    :model-value="visible"
    @update:model-value="emit('update:visible', $event)"
    :title="`更新镜像 - ${resourceName}`"
    width="520px"
    destroy-on-close
  >
    <el-form label-width="100px">
      <el-form-item label="容器">
        <el-select v-model="form.containerName" style="width: 100%" @change="onContainerChange">
          <el-option
            v-for="c in containers"
            :key="c.name"
            :label="c.name"
            :value="c.name"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="镜像">
        <el-input v-model="form.image" placeholder="nginx:latest" />
      </el-form-item>
    </el-form>
    <el-alert
      v-if="form.containerName"
      :title="`当前镜像: ${containers.find(c => c.name === form.containerName)?.image || '-'}`"
      type="info"
      :closable="false"
      style="margin-top: 8px;"
    />
    <template #footer>
      <el-button @click="emit('update:visible', false)">取消</el-button>
      <el-button type="primary" :loading="loading" @click="handleConfirm">确定</el-button>
    </template>
  </el-dialog>
</template>
