<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getConfigMapList, getConfigMapYaml, getConfigMapDetail, deleteConfigMap } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'
import YamlEditor from '@/components/YamlEditor.vue'

const router = useRouter()

const clusterStore = useClusterStore()
const loading = ref(false)
const configMapList = ref<any[]>([])
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)

// Data dialog
const dataDialogVisible = ref(false)
const dataDialogTitle = ref('')
const dataEntries = ref<{ key: string; value: string }[]>([])
const dataLoading = ref(false)

async function fetchConfigMaps() {
  loading.value = true
  try {
    const clusterId = clusterStore.currentCluster?.id
    const res: any = await getConfigMapList({ cluster_id: clusterId })
    configMapList.value = res.data || []
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load ConfigMaps')
  } finally {
    loading.value = false
  }
}

async function handleViewYaml(row: any) {
  yamlLoading.value = true
  yamlDialogVisible.value = true
  try {
    const clusterName = clusterStore.currentCluster?.clusterName || ''
    const res: any = await getConfigMapYaml({
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

async function handleViewData(row: any) {
  dataLoading.value = true
  dataDialogVisible.value = true
  dataDialogTitle.value = `ConfigMap: ${row.name}`
  dataEntries.value = []
  try {
    const clusterName = clusterStore.currentCluster?.clusterName || ''
    const res: any = await getConfigMapDetail({
      clusterName,
      name: row.name,
      namespace: row.namespace,
    })
    const data = res.data?.data || res.data || {}
    dataEntries.value = Object.entries(data).map(([key, value]) => ({
      key,
      value: String(value ?? ''),
    }))
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load data')
    dataDialogVisible.value = false
  } finally {
    dataLoading.value = false
  }
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(
      `Delete ConfigMap "${row.name}" in namespace "${row.namespace}"?`,
      'Confirm',
      { type: 'warning' }
    )
    const clusterName = clusterStore.currentCluster?.clusterName || ''
    await deleteConfigMap({ clusterName, name: row.name, namespace: row.namespace })
    ElMessage.success('Deleted')
    fetchConfigMaps()
  } catch {
    // cancelled
  }
}

onMounted(fetchConfigMaps)
</script>

<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="margin: 0;">ConfigMaps</h2>
      <div>
        <el-button type="primary" @click="$router.push('/config/configmaps/create')">Create</el-button>
        <el-button @click="fetchConfigMaps" :loading="loading">Refresh</el-button>
      </div>
    </div>

    <el-table :data="configMapList" v-loading="loading" stripe style="width: 100%">
      <el-table-column prop="name" label="Name" min-width="200" show-overflow-tooltip />
      <el-table-column prop="namespace" label="Namespace" width="140" />
      <el-table-column label="Data Keys" width="120">
        <template #default="{ row }">
          <el-tag size="small">{{ row.data_keys_count ?? Object.keys(row.data || {}).length }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="age" label="Age" width="120" />
      <el-table-column label="Actions" width="310" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="router.push(`/config/configmaps/${row.namespace}/${row.name}?cluster=${clusterStore.currentCluster?.clusterName || ''}`)">Detail</el-button>
          <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
          <el-button size="small" type="primary" @click="handleViewData(row)">Data</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">Delete</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- YAML Dialog -->
    <el-dialog v-model="yamlDialogVisible" title="ConfigMap YAML" width="70%" top="5vh">
      <YamlEditor v-if="!yamlLoading" v-model="yamlContent" :read-only="true" height="500px" />
      <div v-else v-loading="true" style="height: 200px;"></div>
    </el-dialog>

    <!-- Data Dialog -->
    <el-dialog v-model="dataDialogVisible" :title="dataDialogTitle" width="60%" top="8vh">
      <div v-loading="dataLoading">
        <el-table :data="dataEntries" stripe style="width: 100%" max-height="400">
          <el-table-column prop="key" label="Key" min-width="200" show-overflow-tooltip />
          <el-table-column prop="value" label="Value" min-width="300">
            <template #default="{ row }">
              <div style="white-space: pre-wrap; word-break: break-all; max-height: 100px; overflow-y: auto;">{{ row.value }}</div>
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-if="!dataLoading && dataEntries.length === 0" description="No data" />
      </div>
    </el-dialog>
  </div>
</template>
