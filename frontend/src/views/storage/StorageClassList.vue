<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getStorageClassList, getStorageClassYaml } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'
import YamlEditor from '@/components/YamlEditor.vue'

const router = useRouter()

const clusterStore = useClusterStore()
const loading = ref(false)
const storageClassList = ref<any[]>([])
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)

async function fetchStorageClasses() {
  loading.value = true
  try {
    const clusterId = clusterStore.currentCluster?.id
    const res: any = await getStorageClassList({ cluster_id: clusterId })
    storageClassList.value = res.data || []
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load StorageClasses')
  } finally {
    loading.value = false
  }
}

async function handleViewYaml(row: any) {
  yamlLoading.value = true
  yamlDialogVisible.value = true
  try {
    const clusterName = clusterStore.currentCluster?.clusterName || ''
    const res: any = await getStorageClassYaml({
      clusterName,
      name: row.name,
    })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load YAML')
    yamlDialogVisible.value = false
  } finally {
    yamlLoading.value = false
  }
}

onMounted(fetchStorageClasses)
</script>

<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="margin: 0;">StorageClasses</h2>
      <el-button @click="fetchStorageClasses" :loading="loading">Refresh</el-button>
    </div>

    <el-table :data="storageClassList" v-loading="loading" stripe style="width: 100%">
      <el-table-column prop="name" label="Name" min-width="180" show-overflow-tooltip />
      <el-table-column prop="provisioner" label="Provisioner" min-width="200" show-overflow-tooltip />
      <el-table-column prop="reclaim_policy" label="Reclaim Policy" width="140" />
      <el-table-column prop="volume_binding_mode" label="Volume Binding Mode" min-width="180" show-overflow-tooltip />
      <el-table-column prop="age" label="Age" width="120" />
      <el-table-column label="Actions" width="180" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="router.push(`/storage/storageclasses/${row.name}?cluster=${clusterStore.currentCluster?.clusterName || ''}`)">Detail</el-button>
          <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="yamlDialogVisible" title="StorageClass YAML" width="70%" top="5vh">
      <YamlEditor v-if="!yamlLoading" v-model="yamlContent" :read-only="true" height="500px" />
      <div v-else v-loading="true" style="height: 200px;"></div>
    </el-dialog>
  </div>
</template>
