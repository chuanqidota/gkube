<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getStorageClassDetail, getStorageClassYaml } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const storageClass = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const activeTab = ref('info')

const name = route.params.name as string
const clusterName = (route.query.cluster as string) || ''

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getStorageClassDetail({ clusterName, name })
    storageClass.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load StorageClass detail')
  } finally {
    loading.value = false
  }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getStorageClassYaml({ clusterName, name })
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

onMounted(fetchDetail)
</script>

<template>
  <div v-loading="loading">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="margin: 0;">StorageClass: {{ name }}</h2>
      <el-button @click="router.push('/storage/storageclasses')">Back to List</el-button>
    </div>

    <template v-if="storageClass">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="Info" name="info">
          <el-descriptions :column="2" border style="margin-top: 8px;">
            <el-descriptions-item label="Name">{{ storageClass.name }}</el-descriptions-item>
            <el-descriptions-item label="Provisioner">{{ storageClass.provisioner || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Reclaim Policy">{{ storageClass.reclaimPolicy || storageClass.reclaim_policy || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Volume Binding Mode">{{ storageClass.volumeBindingMode || storageClass.volume_binding_mode || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Age">{{ storageClass.age || '-' }}</el-descriptions-item>
          </el-descriptions>

          <!-- Parameters -->
          <div v-if="storageClass.parameters && Object.keys(storageClass.parameters).length > 0" style="margin-top: 16px;">
            <h4>Parameters</h4>
            <el-tag
              v-for="(val, key) in storageClass.parameters"
              :key="key"
              style="margin-right: 8px; margin-bottom: 8px;"
            >
              {{ key }}={{ val }}
            </el-tag>
          </div>

          <!-- Labels -->
          <div v-if="storageClass.labels && Object.keys(storageClass.labels).length > 0" style="margin-top: 16px;">
            <h4>Labels</h4>
            <el-tag
              v-for="(val, key) in storageClass.labels"
              :key="key"
              style="margin-right: 8px; margin-bottom: 8px;"
            >
              {{ key }}={{ val }}
            </el-tag>
          </div>
        </el-tab-pane>

        <el-tab-pane label="YAML" name="yaml">
          <div v-loading="yamlLoading">
            <YamlEditor v-model="yamlContent" height="600px" read-only />
          </div>
        </el-tab-pane>
      </el-tabs>
    </template>
  </div>
</template>
