<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { CopyDocument } from '@element-plus/icons-vue'
import YamlEditor from '@/components/YamlEditor.vue'
import CloneDialog from '@/components/CloneDialog.vue'
import { useCloneCreate, dumpCloneYaml } from '@/composables/useCloneCreate'
import { createVolumeSnapshotClass, getVolumeSnapshotClassList, getVolumeSnapshotClassYaml } from '@/api/resource'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
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

const {
  cloneMode, cloneName, cloneNameOptions, cloneNameLoading,
  cloneLoading, startClone, cancelClone, handleLoadClone,
} = useCloneCreate({
  api: { list: getVolumeSnapshotClassList, yaml: getVolumeSnapshotClassYaml },
  namespaceScoped: false,
  hasForm: false,
  onCloneToYaml: (parsed) => { yamlContent.value = dumpCloneYaml(parsed) },
})

async function handleSubmit() {
  submitting.value = true
  try {
    await createVolumeSnapshotClass({ yaml: yamlContent.value })
    ElMessage.success(t('common.create') + ' ' + t('common.success'))
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
      <h2>{{ t('common.create') }} {{ t('storage.volumeSnapshotClass') }}</h2>
      <el-button size="small" @click="startClone">
        <el-icon><CopyDocument /></el-icon> 从现有资源克隆
      </el-button>
    </div>

    <CloneDialog
      kind-label="VolumeSnapshotClass"
      :namespace-scoped="false"
      :show-target-choice="false"
      v-model="cloneMode"
      v-model:name-value="cloneName"
      :name-options="cloneNameOptions"
      :name-loading="cloneNameLoading"
      :loading="cloneLoading"
      @load="handleLoadClone"
      @cancel="cancelClone"
    />

    <el-alert
      :title="t('storage.createSnapshotClassYamlHint')"
      type="info"
      :closable="false"
      show-icon
      style="margin-bottom: 16px;"
    />

    <YamlEditor v-model="yamlContent" height="500px" />

    <div class="form-actions">
      <el-button @click="handleCancel">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ t('common.create') }} {{ t('storage.volumeSnapshotClass') }}</el-button>
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
  display: flex;
  justify-content: space-between;
  align-items: center;
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
