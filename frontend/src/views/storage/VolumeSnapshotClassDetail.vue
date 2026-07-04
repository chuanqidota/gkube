<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getVolumeSnapshotClassDetail, getVolumeSnapshotClassYaml, updateVolumeSnapshotClass } from '@/api/resource'
import { useI18n } from 'vue-i18n'
import YamlEditor from '@/components/YamlEditor.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const loading = ref(false)
const snapshotClass = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const activeTab = ref('info')
const yamlEditorRef = ref<InstanceType<typeof YamlEditor> | null>(null)

const name = route.params.name as string

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getVolumeSnapshotClassDetail({ name })
    snapshotClass.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load VolumeSnapshotClass detail')
  } finally {
    loading.value = false
  }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getVolumeSnapshotClassYaml({ name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load YAML')
  } finally {
    yamlLoading.value = false
  }
}

function handleTabChange(tab: string) {
  if (tab === 'yaml' && !yamlContent.value) {
    fetchYaml()
  }
}

async function handleSaveYaml(content: string) {
  try {
    await updateVolumeSnapshotClass({ yaml: content })
    ElMessage.success(t('common.save') + ' ' + t('common.success'))
    fetchDetail()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Save failed')
    yamlEditorRef.value?.resetSaving()
  }
}

onMounted(fetchDetail)
</script>

<template>
  <div v-loading="loading">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="margin: 0;">{{ t('storage.volumeSnapshotClassTitle', { name }) }}</h2>
      <el-button @click="router.push('/storage/volumesnapshotclasses')">{{ t('common.back') }}</el-button>
    </div>

    <template v-if="snapshotClass">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane :label="t('common.detail')" name="info">
          <el-descriptions :column="2" border style="margin-top: 8px;">
            <el-descriptions-item :label="t('common.name')">{{ snapshotClass.metadata?.name || name }}</el-descriptions-item>
            <el-descriptions-item :label="t('storage.driver')">{{ snapshotClass.driver || '-' }}</el-descriptions-item>
            <el-descriptions-item :label="t('storage.deletionPolicy')">{{ snapshotClass.deletionPolicy || '-' }}</el-descriptions-item>
            <el-descriptions-item :label="t('common.age')">{{ snapshotClass.metadata?.creationTimestamp || '-' }}</el-descriptions-item>
          </el-descriptions>

          <!-- Parameters -->
          <div v-if="snapshotClass.parameters && Object.keys(snapshotClass.parameters).length > 0" style="margin-top: 16px;">
            <h4>{{ t('storage.parameters') }}</h4>
            <el-tag
              v-for="(val, key) in snapshotClass.parameters"
              :key="key"
              style="margin-right: 8px; margin-bottom: 8px;"
            >
              {{ key }}={{ val }}
            </el-tag>
          </div>

          <!-- Labels -->
          <div v-if="snapshotClass.metadata?.labels && Object.keys(snapshotClass.metadata.labels).length > 0" style="margin-top: 16px;">
            <h4>Labels</h4>
            <el-tag
              v-for="(val, key) in snapshotClass.metadata.labels"
              :key="key"
              style="margin-right: 8px; margin-bottom: 8px;"
            >
              {{ key }}={{ val }}
            </el-tag>
          </div>

          <!-- Annotations -->
          <div v-if="snapshotClass.metadata?.annotations && Object.keys(snapshotClass.metadata.annotations).length > 0" style="margin-top: 16px;">
            <h4>Annotations</h4>
            <div v-for="(val, key) in snapshotClass.metadata.annotations" :key="key" style="margin-bottom: 4px;">
              <el-text size="small" type="info">{{ key }}:</el-text>
              <el-text size="small" style="margin-left: 4px; word-break: break-all;">{{ val }}</el-text>
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane :label="t('common.yaml')" name="yaml">
          <div v-loading="yamlLoading">
            <YamlEditor
              ref="yamlEditorRef"
              v-model="yamlContent"
              height="600px"
              saveable
              @save="handleSaveYaml"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </template>
  </div>
</template>
