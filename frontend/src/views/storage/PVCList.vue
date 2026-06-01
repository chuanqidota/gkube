<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getPvcList, getPvcYaml, deletePvc } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'
import YamlEditor from '@/components/YamlEditor.vue'

const clusterStore = useClusterStore()
const loading = ref(false)
const pvcList = ref<any[]>([])
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)

async function fetchPvcs() {
  loading.value = true
  try {
    const clusterId = clusterStore.currentCluster?.id
    const res: any = await getPvcList({ cluster_id: clusterId })
    pvcList.value = res.data || []
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load PVCs')
  } finally {
    loading.value = false
  }
}

async function handleViewYaml(row: any) {
  yamlLoading.value = true
  yamlDialogVisible.value = true
  try {
    const clusterId = clusterStore.currentCluster?.id
    const res: any = await getPvcYaml({
      name: row.name,
      namespace: row.namespace,
      cluster_id: clusterId,
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
      `Delete PVC "${row.name}" in namespace "${row.namespace}"?`,
      'Confirm',
      { type: 'warning' }
    )
    const clusterId = clusterStore.currentCluster?.id
    await deletePvc({ name: row.name, namespace: row.namespace, cluster_id: clusterId })
    ElMessage.success('Deleted')
    fetchPvcs()
  } catch {
    // cancelled
  }
}

function statusType(status: string) {
  const s = (status || '').toLowerCase()
  if (s === 'bound') return 'success'
  if (s === 'pending') return 'warning'
  if (s === 'lost') return 'danger'
  return 'info'
}

onMounted(fetchPvcs)
</script>

<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="margin: 0;">PersistentVolumeClaims</h2>
      <el-button @click="fetchPvcs" :loading="loading">Refresh</el-button>
    </div>

    <el-table :data="pvcList" v-loading="loading" stripe style="width: 100%">
      <el-table-column prop="name" label="Name" min-width="180" show-overflow-tooltip />
      <el-table-column prop="namespace" label="Namespace" width="140" />
      <el-table-column prop="status" label="Status" width="120">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)" size="small">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="volume" label="Volume" min-width="160" show-overflow-tooltip />
      <el-table-column prop="capacity" label="Capacity" width="120" />
      <el-table-column prop="access_modes" label="Access Modes" min-width="140" show-overflow-tooltip />
      <el-table-column prop="storage_class" label="Storage Class" min-width="140" show-overflow-tooltip />
      <el-table-column prop="age" label="Age" width="120" />
      <el-table-column label="Actions" width="160" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">Delete</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="yamlDialogVisible" title="PVC YAML" width="70%" top="5vh">
      <YamlEditor v-if="!yamlLoading" v-model="yamlContent" :read-only="true" height="500px" />
      <div v-else v-loading="true" style="height: 200px;"></div>
    </el-dialog>
  </div>
</template>
