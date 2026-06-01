<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getPodList, getPodYaml, deletePod, getNamespaceList } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const router = useRouter()
const loading = ref(false)
const podList = ref<any[]>([])
const namespaceList = ref<string[]>([])
const selectedNamespace = ref('')
const clusterName = ref('')

// YAML dialog
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)

async function fetchNamespaces() {
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = (res.data || []).map((ns: any) => ns.name || ns)
  } catch {
    // ignore
  }
}

async function fetchPods() {
  loading.value = true
  try {
    const params: any = {}
    if (selectedNamespace.value) params.namespace = selectedNamespace.value
    if (clusterName.value) params.clusterName = clusterName.value
    const res: any = await getPodList(params)
    podList.value = res.data || []
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load pods')
  } finally {
    loading.value = false
  }
}

function handleNamespaceChange() {
  fetchPods()
}

async function handleViewYaml(row: any) {
  yamlDialogVisible.value = true
  yamlLoading.value = true
  yamlContent.value = ''
  try {
    const res: any = await getPodYaml({
      clusterName: clusterName.value,
      namespace: row.namespace,
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

function handleViewLogs(row: any) {
  router.push({
    path: '/logs',
    query: { namespace: row.namespace, pod: row.name, cluster: clusterName.value },
  })
}

function handleDetail(row: any) {
  router.push(`/workloads/pods/${row.namespace}/${row.name}`)
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(
      `Delete pod "${row.name}" in namespace "${row.namespace}"?`,
      'Confirm',
      { type: 'warning' }
    )
    await deletePod({ clusterName: clusterName.value, namespace: row.namespace, name: row.name })
    ElMessage.success('Pod deleted')
    fetchPods()
  } catch {
    // cancelled
  }
}

function statusType(status: string) {
  const s = (status || '').toLowerCase()
  if (s === 'running') return 'success'
  if (s === 'succeeded') return 'info'
  if (s === 'pending') return 'warning'
  if (s === 'failed' || s === 'error') return 'danger'
  return 'info'
}

onMounted(() => {
  fetchNamespaces()
  fetchPods()
})
</script>

<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="margin: 0;">Pods</h2>
      <div style="display: flex; gap: 12px; align-items: center;">
        <el-input
          v-model="clusterName"
          placeholder="Cluster Name"
          style="width: 180px;"
          clearable
          @clear="fetchPods"
          @keyup.enter="fetchPods"
        />
        <el-select
          v-model="selectedNamespace"
          placeholder="All Namespaces"
          clearable
          style="width: 180px;"
          @change="handleNamespaceChange"
        >
          <el-option
            v-for="ns in namespaceList"
            :key="ns"
            :label="ns"
            :value="ns"
          />
        </el-select>
        <el-button type="primary" @click="fetchPods">Refresh</el-button>
      </div>
    </div>

    <el-table :data="podList" v-loading="loading" stripe style="width: 100%">
      <el-table-column prop="name" label="Name" min-width="200" show-overflow-tooltip>
        <template #default="{ row }">
          <el-button link type="primary" @click="handleDetail(row)">{{ row.name }}</el-button>
        </template>
      </el-table-column>
      <el-table-column prop="namespace" label="Namespace" width="140" />
      <el-table-column prop="status" label="Status" width="120">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)" size="small">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="restarts" label="Restarts" width="100" />
      <el-table-column prop="age" label="Age" width="120" />
      <el-table-column prop="node" label="Node" width="160" show-overflow-tooltip />
      <el-table-column label="Actions" width="260" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
          <el-button size="small" type="primary" @click="handleViewLogs(row)">Logs</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">Delete</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- YAML Dialog -->
    <el-dialog v-model="yamlDialogVisible" title="Pod YAML" width="70%" top="5vh" destroy-on-close>
      <div v-loading="yamlLoading">
        <YamlEditor v-model="yamlContent" height="500px" read-only />
      </div>
    </el-dialog>
  </div>
</template>
