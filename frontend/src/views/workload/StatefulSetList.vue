<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Plus, Delete, Search } from '@element-plus/icons-vue'
import { getStatefulSetList, getStatefulSetYaml, deleteStatefulSet, getNamespaceList, extractNamespaceNames } from '@/api/resource'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import YamlEditor from '@/components/YamlEditor.vue'

const router = useRouter()
const loading = ref(false)
const statefulSetList = ref<any[]>([])
const namespaceList = ref<string[]>([])
const selectedNamespace = ref('')
const searchName = ref('')
const selectedRows = ref<any[]>([])
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)

const filteredList = computed(() => {
  if (!searchName.value) return statefulSetList.value
  const keyword = searchName.value.toLowerCase()
  return statefulSetList.value.filter((d) => d.name?.toLowerCase().includes(keyword))
})

async function fetchNamespaces() {
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = extractNamespaceNames(res.data)
  } catch { /* ignore */ }
}

async function fetchStatefulSets() {
  loading.value = true
  try {
    const params: any = {}
    if (selectedNamespace.value) params.namespace = selectedNamespace.value
    const res: any = await getStatefulSetList(params)
    statefulSetList.value = res.data?.items || res.data || []
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load StatefulSets')
  } finally {
    loading.value = false
  }
}

function handleNamespaceChange() { fetchStatefulSets() }
function handleSelectionChange(rows: any[]) { selectedRows.value = rows }

async function handleViewYaml(row: any) {
  yamlDialogVisible.value = true
  yamlLoading.value = true
  yamlContent.value = ''
  try {
    const res: any = await getStatefulSetYaml({ namespace: row.namespace, name: row.name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load YAML')
    yamlDialogVisible.value = false
  } finally {
    yamlLoading.value = false
  }
}

function handleDetail(row: any) {
  router.push(`/workloads/statefulsets/${row.namespace}/${row.name}`)
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(`Delete StatefulSet "${row.name}" in namespace "${row.namespace}"?`, 'Confirm', { type: 'warning' })
    await deleteStatefulSet({ namespace: row.namespace, name: row.name })
    ElMessage.success('StatefulSet deleted')
    fetchStatefulSets()
  } catch { /* cancelled */ }
}

async function handleBatchDelete() {
  if (!selectedRows.value.length) return
  try {
    await ElMessageBox.confirm(`Delete ${selectedRows.value.length} selected StatefulSet(s)?`, 'Confirm', { type: 'warning' })
    let count = 0
    for (const row of selectedRows.value) {
      try { await deleteStatefulSet({ namespace: row.namespace, name: row.name }); count++ } catch { /* continue */ }
    }
    ElMessage.success(`Deleted ${count} StatefulSet(s)`)
    fetchStatefulSets()
  } catch { /* cancelled */ }
}

const { isRunning, countdown, toggle, refresh } = useAutoRefresh(fetchStatefulSets)

onMounted(() => { fetchNamespaces(); fetchStatefulSets() })
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <el-input v-model="searchName" placeholder="Search by name" style="width: 220px;" clearable>
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select v-model="selectedNamespace" placeholder="All Namespaces" clearable style="width: 180px;" @change="handleNamespaceChange">
          <el-option v-for="ns in namespaceList" :key="ns" :label="ns" :value="ns" />
        </el-select>
        <el-button type="primary" @click="refresh"><el-icon><Refresh /></el-icon> Refresh</el-button>
        <el-button :type="isRunning ? 'success' : 'info'" @click="toggle">
          {{ isRunning ? `Auto (${countdown}s)` : 'Manual' }}
        </el-button>
        <el-button type="success" @click="router.push('/workloads/statefulsets/create')"><el-icon><Plus /></el-icon> Create</el-button>
        <el-button type="danger" :disabled="!selectedRows.length" @click="handleBatchDelete"><el-icon><Delete /></el-icon> Delete ({{ selectedRows.length }})</el-button>
      </div>
    </el-card>
    <el-card shadow="never" class="table-card">
      <el-table :data="filteredList" v-loading="loading" stripe @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="45" />
        <el-table-column prop="name" label="Name" min-width="200" show-overflow-tooltip>
          <template #default="{ row }"><el-button link type="primary" @click="handleDetail(row)">{{ row.name }}</el-button></template>
        </el-table-column>
        <el-table-column prop="namespace" label="Namespace" width="140" />
        <el-table-column label="Ready" width="100"><template #default="{ row }">{{ row.readyReplicas || 0 }}/{{ row.replicas || 0 }}</template></el-table-column>
        <el-table-column prop="age" label="Age" width="120" />
        <el-table-column label="Actions" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">Delete</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="No StatefulSets found">
            <el-button type="primary" @click="router.push('/workloads/statefulsets/create')">Create StatefulSet</el-button>
          </el-empty>
        </template>
      </el-table>
    </el-card>
    <el-dialog v-model="yamlDialogVisible" title="StatefulSet YAML" width="70%" top="5vh" destroy-on-close>
      <div v-loading="yamlLoading"><YamlEditor v-model="yamlContent" height="500px" read-only auto-format /></div>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; gap: 12px; align-items: center; flex-wrap: wrap; }
.table-card { border-radius: 8px; }
</style>
