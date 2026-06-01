<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getServiceList, getServiceYaml, deleteService } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'
import YamlEditor from '@/components/YamlEditor.vue'

const router = useRouter()

const clusterStore = useClusterStore()
const loading = ref(false)
const serviceList = ref<any[]>([])
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)

async function fetchServices() {
  loading.value = true
  try {
    const clusterId = clusterStore.currentCluster?.id
    const res: any = await getServiceList({ cluster_id: clusterId })
    serviceList.value = res.data || []
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load services')
  } finally {
    loading.value = false
  }
}

async function handleViewYaml(row: any) {
  yamlLoading.value = true
  yamlDialogVisible.value = true
  try {
    const clusterName = clusterStore.currentCluster?.clusterName || ''
    const res: any = await getServiceYaml({
      clusterName,
      name: row.name,
      namespace: row.namespace,
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
      `Delete Service "${row.name}" in namespace "${row.namespace}"?`,
      'Confirm',
      { type: 'warning' }
    )
    const clusterName = clusterStore.currentCluster?.clusterName || ''
    await deleteService({ clusterName, name: row.name, namespace: row.namespace })
    ElMessage.success('Deleted')
    fetchServices()
  } catch {
    // cancelled
  }
}

onMounted(fetchServices)
</script>

<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="margin: 0;">Services</h2>
      <div>
        <el-button type="primary" @click="$router.push('/services/create')">Create</el-button>
        <el-button @click="fetchServices" :loading="loading">Refresh</el-button>
      </div>
    </div>

    <el-table :data="serviceList" v-loading="loading" stripe style="width: 100%">
      <el-table-column prop="name" label="Name" min-width="160" show-overflow-tooltip />
      <el-table-column prop="namespace" label="Namespace" min-width="120" />
      <el-table-column prop="type" label="Type" min-width="120" />
      <el-table-column prop="cluster_ip" label="Cluster IP" min-width="140" />
      <el-table-column prop="ports" label="Ports" min-width="160" show-overflow-tooltip />
      <el-table-column prop="age" label="Age" min-width="100" />
      <el-table-column label="Actions" width="260" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="router.push(`/services/${row.namespace}/${row.name}?cluster=${clusterStore.currentCluster?.clusterName || ''}`)">Detail</el-button>
          <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">Delete</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="yamlDialogVisible" title="Service YAML" width="70%" top="5vh">
      <YamlEditor v-if="!yamlLoading" v-model="yamlContent" :read-only="true" height="500px" />
      <div v-else v-loading="true" style="height: 200px;"></div>
    </el-dialog>
  </div>
</template>
