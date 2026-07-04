<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import YamlEditor from '@/components/YamlEditor.vue'
import { createVolumeSnapshotClass } from '@/api/resource'

const router = useRouter()
const submitting = ref(false)

const defaultYaml = `apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshotClass
metadata:
  name: my-snapshotclass
driver: hostpath.csi.k8s.io
deletionPolicy: Delete
parameters:
  # Add driver-specific parameters here
`

const yamlContent = ref(defaultYaml)

async function handleSubmit() {
  submitting.value = true
  try {
    await createVolumeSnapshotClass({ yaml: yamlContent.value })
    ElMessage.success('VolumeSnapshotClass created successfully')
    router.push('/storage/volumesnapshotclasses')
  } catch (e: any) {
    ElMessage.error(e?.message || 'Create failed')
  } finally {
    submitting.value = false
  }
}

function handleCancel() {
  router.push('/storage/volumesnapshotclasses')
}
</script>

<template>
  <div class="create-page">
    <div class="form-header">
      <h2>Create VolumeSnapshotClass</h2>
    </div>

    <el-alert
      title="Edit YAML below to configure the VolumeSnapshotClass"
      type="info"
      :closable="false"
      show-icon
      style="margin-bottom: 16px;"
    />

    <YamlEditor v-model="yamlContent" height="500px" />

    <div class="form-actions">
      <el-button @click="handleCancel">Cancel</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">Create VolumeSnapshotClass</el-button>
    </div>
  </div>
</template>

<style scoped>
.create-page {
  max-width: 900px;
  margin: 0 auto;
  padding: 20px 0;
}
.form-header {
  margin-bottom: 24px;
}
.form-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 24px;
  border-top: 1px solid var(--gk-color-border);
  margin-top: 24px;
}
</style>
