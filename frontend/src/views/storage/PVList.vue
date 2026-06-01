<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getPvList, getPvYaml, deletePv } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'
import YamlEditor from '@/components/YamlEditor.vue'

const router = useRouter()

const clusterStore = useClusterStore()
const loading = ref(false)
const pvList = ref<any[]>([])
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)

async function fetchPvs() {
  loading.value = true
  try {
    const clusterId = clusterStore.currentCluster?.id
    const res: any = await getPvList({ cluster_id: clusterId })
    pvList.value = res.data || []
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load PVs')
  } finally {
    loading.value = false
  }
}

async function handleViewYaml(row: any) {
  yamlLoading.value = true
  yamlDialogVisible.value = true
  try {
    const clusterName = clusterStore.currentCluster?.clusterName || ''
    const res: any = await getPvYaml({
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

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(
      `Delete PersistentVolume "${row.name}"?`,
      'Confirm',
      { type: 'warning' }
    )
    const clusterName = clusterStore.currentCluster?.clusterName || ''
    await deletePv({ clusterName, name: row.name })
    ElMessage.success('Deleted')
    fetchPvs()
  } catch {
    // cancelled
  }
}

function statusType(status: string) {
  const s = (status || '').toLowerCase()
  if (s === 'bound') return 'success'
  if (s === 'available') return 'primary'
  if (s === 'released') return 'warning'
  if (s === 'failed') return 'danger'
  return 'info'
}

onMounted(fetchPvs)
</script>

<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="margin: 0;">PersistentVolumes</h2>
      <div>
        <el-button type="primary" @click="router.push('/storage/pvs/create')">Create</el-button>
        <el-button @click="fetchPvs" :loading="loading">Refresh</el-button>
      </div>
    </div>

    <el-table :data="pvList" v-loading="loading" stripe style="width: 100%">
      <el-table-column prop="name" label="Name" min-width="200" show-overflow-tooltip />
      <el-table-column prop="capacity" label="Capacity" width="120" />
      <el-table-column prop="access_modes" label="Access Modes" min-width="160" show-overflow-tooltip />
      <el-table-column prop="status" label="Status" width="120">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)" size="small">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="claim" label="Claim" min-width="180" show-overflow-tooltip />
      <el-table-column prop="storage_class" label="Storage Class" min-width="140" show-overflow-tooltip />
      <el-table-column prop="age" label="Age" width="120" />
      <el-table-column label="Actions" width="240" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="router.push(`/storage/pvs/${row.name}?cluster=${clusterStore.currentCluster?.clusterName || ''}`)">Detail</el-button>
          <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">Delete</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="yamlDialogVisible" title="PersistentVolume YAML" width="70%" top="5vh">
      <YamlEditor v-if="!yamlLoading" v-model="yamlContent" :read-only="true" height="500px" />
      <div v-else v-loading="true" style="height: 200px;"></div>
    </el-dialog>
  </div>
</template>
