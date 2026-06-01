<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getSecretList, getSecretYaml, getSecretDetail, deleteSecret } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'
import YamlEditor from '@/components/YamlEditor.vue'

const router = useRouter()

const clusterStore = useClusterStore()
const loading = ref(false)
const secretList = ref<any[]>([])
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)

// Data dialog
const dataDialogVisible = ref(false)
const dataDialogTitle = ref('')
const dataEntries = ref<{ key: string; rawValue: string; decodedValue: string }[]>([])
const dataLoading = ref(false)
const showDecoded = ref(true)

async function fetchSecrets() {
  loading.value = true
  try {
    const clusterId = clusterStore.currentCluster?.id
    const res: any = await getSecretList({ cluster_id: clusterId })
    secretList.value = res.data || []
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load Secrets')
  } finally {
    loading.value = false
  }
}

async function handleViewYaml(row: any) {
  yamlLoading.value = true
  yamlDialogVisible.value = true
  try {
    const clusterName = clusterStore.currentCluster?.clusterName || ''
    const res: any = await getSecretYaml({
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

function base64Decode(str: string): string {
  try {
    return atob(str)
  } catch {
    return str
  }
}

async function handleViewData(row: any) {
  dataLoading.value = true
  dataDialogVisible.value = true
  dataDialogTitle.value = `Secret: ${row.name}`
  dataEntries.value = []
  try {
    const clusterName = clusterStore.currentCluster?.clusterName || ''
    const res: any = await getSecretDetail({
      clusterName,
      name: row.name,
      namespace: row.namespace,
    })
    const data = res.data?.data || res.data || {}
    dataEntries.value = Object.entries(data).map(([key, value]) => {
      const rawValue = String(value ?? '')
      return {
        key,
        rawValue,
        decodedValue: base64Decode(rawValue),
      }
    })
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
      `Delete Secret "${row.name}" in namespace "${row.namespace}"?`,
      'Confirm',
      { type: 'warning' }
    )
    const clusterName = clusterStore.currentCluster?.clusterName || ''
    await deleteSecret({ clusterName, name: row.name, namespace: row.namespace })
    ElMessage.success('Deleted')
    fetchSecrets()
  } catch {
    // cancelled
  }
}

onMounted(fetchSecrets)
</script>

<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="margin: 0;">Secrets</h2>
      <div>
        <el-button type="primary" @click="$router.push('/config/secrets/create')">Create</el-button>
        <el-button @click="fetchSecrets" :loading="loading">Refresh</el-button>
      </div>
    </div>

    <el-table :data="secretList" v-loading="loading" stripe style="width: 100%">
      <el-table-column prop="name" label="Name" min-width="200" show-overflow-tooltip />
      <el-table-column prop="namespace" label="Namespace" width="140" />
      <el-table-column prop="type" label="Type" min-width="160" show-overflow-tooltip />
      <el-table-column label="Data Keys" width="120">
        <template #default="{ row }">
          <el-tag size="small">{{ row.data_keys_count ?? Object.keys(row.data || {}).length }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="age" label="Age" width="120" />
      <el-table-column label="Actions" width="310" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="router.push(`/config/secrets/${row.namespace}/${row.name}?cluster=${clusterStore.currentCluster?.clusterName || ''}`)">Detail</el-button>
          <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
          <el-button size="small" type="primary" @click="handleViewData(row)">Data</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">Delete</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- YAML Dialog -->
    <el-dialog v-model="yamlDialogVisible" title="Secret YAML" width="70%" top="5vh">
      <YamlEditor v-if="!yamlLoading" v-model="yamlContent" :read-only="true" height="500px" />
      <div v-else v-loading="true" style="height: 200px;"></div>
    </el-dialog>

    <!-- Data Dialog -->
    <el-dialog v-model="dataDialogVisible" :title="dataDialogTitle" width="60%" top="8vh">
      <div style="margin-bottom: 12px;">
        <el-switch v-model="showDecoded" active-text="Decoded (Base64)" inactive-text="Raw (Base64)" />
      </div>
      <div v-loading="dataLoading">
        <el-table :data="dataEntries" stripe style="width: 100%" max-height="400">
          <el-table-column prop="key" label="Key" min-width="200" show-overflow-tooltip />
          <el-table-column label="Value" min-width="300">
            <template #default="{ row }">
              <div style="white-space: pre-wrap; word-break: break-all; max-height: 100px; overflow-y: auto;">
                {{ showDecoded ? row.decodedValue : row.rawValue }}
              </div>
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-if="!dataLoading && dataEntries.length === 0" description="No data" />
      </div>
    </el-dialog>
  </div>
</template>
