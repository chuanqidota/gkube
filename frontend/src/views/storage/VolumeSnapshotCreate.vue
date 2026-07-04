<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import YamlEditor from '@/components/YamlEditor.vue'
import { createVolumeSnapshot, getNamespaceList, extractNamespaceNames } from '@/api/resource'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const router = useRouter()
const submitting = ref(false)
const namespace = ref('default')
const namespaceList = ref<string[]>([])

const defaultYaml = `apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: my-volume-snapshot
  namespace: default
spec:
  volumeSnapshotClassName: csi-hostpath-snapclass
  source:
    persistentVolumeClaimName: my-pvc
`

const yamlContent = ref(defaultYaml)

async function fetchNamespaces() {
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = extractNamespaceNames(res.data)
  } catch { /* ignore */ }
}

function handleNamespaceChange() {
  const doc = yamlContent.value.replace(/namespace:\s*\S+/, `namespace: ${namespace.value}`)
  yamlContent.value = doc
}

async function handleSubmit() {
  submitting.value = true
  try {
    await createVolumeSnapshot({ namespace: namespace.value, yaml: yamlContent.value })
    ElMessage.success(t('common.create') + ' ' + t('common.success'))
    router.push('/storage/volumesnapshots')
  } catch (e: any) {
    ElMessage.error(e?.message || 'Create failed')
  } finally {
    submitting.value = false
  }
}

function handleCancel() {
  router.push('/storage/volumesnapshots')
}

onMounted(fetchNamespaces)
</script>

<template>
  <div class="create-page">
    <div class="form-header">
      <h2>{{ t('common.create') }} {{ t('storage.volumeSnapshot') }}</h2>
    </div>

    <el-form label-width="140px" style="max-width: 700px; margin-bottom: 16px;">
      <el-form-item :label="t('common.namespace_label')" required>
        <el-select v-model="namespace" style="width: 100%;" @change="handleNamespaceChange">
          <el-option v-for="ns in namespaceList" :key="ns" :label="ns" :value="ns" />
        </el-select>
      </el-form-item>
    </el-form>

    <el-alert
      :title="t('storage.createSnapshotYamlHint')"
      type="info"
      :closable="false"
      show-icon
      style="margin-bottom: 16px;"
    />

    <YamlEditor v-model="yamlContent" height="500px" />

    <div class="form-actions">
      <el-button @click="handleCancel">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ t('common.create') }} {{ t('storage.volumeSnapshot') }}</el-button>
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
