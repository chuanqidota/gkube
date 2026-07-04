<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Plus, Delete, Search } from '@element-plus/icons-vue'
import { getVolumeSnapshotList, getVolumeSnapshotYaml, deleteVolumeSnapshot, getNamespaceList, extractNamespaceNames } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const router = useRouter()
const loading = ref(false)
const snapshotList = ref<any[]>([])
const namespaceList = ref<string[]>([])
const selectedNamespace = ref('')
const searchName = ref('')
const selectedRows = ref<any[]>([])
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)

const filteredList = computed(() => {
  if (!searchName.value) return snapshotList.value
  const keyword = searchName.value.toLowerCase()
  return snapshotList.value.filter((d) => d.name?.toLowerCase().includes(keyword))
})

async function fetchNamespaces() {
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = extractNamespaceNames(res.data)
  } catch { /* ignore */ }
}

async function fetchSnapshots() {
  loading.value = true
  try {
    const params: any = {}
    if (selectedNamespace.value) params.namespace = selectedNamespace.value
    const res: any = await getVolumeSnapshotList(params)
    snapshotList.value = res.data || []
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load VolumeSnapshots')
  } finally { loading.value = false }
}

function handleNamespaceChange() { fetchSnapshots() }
function handleSelectionChange(rows: any[]) { selectedRows.value = rows }

function statusType(status: string) {
  const s = (status || '').toLowerCase()
  if (s === 'ready' || s === 'bound') return 'success'
  if (s === 'pending') return 'warning'
  if (s === 'error' || s === 'failed') return 'danger'
  return 'info'
}

function getStatus(row: any): string {
  if (row.status?.readyToUse) return 'Ready'
  if (row.status?.error?.message) return 'Error'
  if (row.status?.boundVolumeSnapshotContentName) return 'Bound'
  return 'Pending'
}

function getRestoreSize(row: any): string {
  return row.status?.restoreSize || row.spec?.volumeSnapshotClassName || '-'
}

function getSnapshotClass(row: any): string {
  return row.spec?.volumeSnapshotClassName || '-'
}

function getSource(row: any): string {
  const src = row.spec?.source
  if (!src) return '-'
  if (src.persistentVolumeClaimName) return `PVC:${src.persistentVolumeClaimName}`
  if (src.volumeSnapshotContentName) return `VSC:${src.volumeSnapshotContentName}`
  return '-'
}

async function handleViewYaml(row: any) {
  yamlDialogVisible.value = true; yamlLoading.value = true; yamlContent.value = ''
  try {
    const res: any = await getVolumeSnapshotYaml({ name: row.name, namespace: row.namespace })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) { ElMessage.error(e?.message || 'Failed to load YAML'); yamlDialogVisible.value = false }
  finally { yamlLoading.value = false }
}

function handleDetail(row: any) { router.push(`/storage/volumesnapshots/${row.namespace}/${row.name}`) }

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(`Delete VolumeSnapshot "${row.name}" in namespace "${row.namespace}"?`, 'Confirm', { type: 'warning' })
    await deleteVolumeSnapshot({ name: row.name, namespace: row.namespace })
    ElMessage.success('Deleted'); fetchSnapshots()
  } catch { /* cancelled */ }
}

async function handleBatchDelete() {
  if (!selectedRows.value.length) return
  try {
    await ElMessageBox.confirm(`Delete ${selectedRows.value.length} selected VolumeSnapshot(s)?`, 'Confirm', { type: 'warning' })
    let count = 0
    for (const row of selectedRows.value) {
      try { await deleteVolumeSnapshot({ name: row.name, namespace: row.namespace }); count++ } catch { /* continue */ }
    }
    ElMessage.success(`Deleted ${count} VolumeSnapshot(s)`); fetchSnapshots()
  } catch { /* cancelled */ }
}

onMounted(() => { fetchNamespaces(); fetchSnapshots() })
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
        <el-button type="primary" @click="fetchSnapshots"><el-icon><Refresh /></el-icon> Refresh</el-button>
        <el-button type="success" @click="router.push('/storage/volumesnapshots/create')"><el-icon><Plus /></el-icon> Create</el-button>
        <el-button type="danger" :disabled="!selectedRows.length" @click="handleBatchDelete"><el-icon><Delete /></el-icon> Delete ({{ selectedRows.length }})</el-button>
      </div>
    </el-card>
    <el-card shadow="never" class="table-card">
      <el-table :data="filteredList" v-loading="loading" stripe @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="45" />
        <el-table-column prop="name" label="Name" min-width="180" show-overflow-tooltip>
          <template #default="{ row }"><el-button link type="primary" @click="handleDetail(row)">{{ row.name }}</el-button></template>
        </el-table-column>
        <el-table-column prop="namespace" label="Namespace" width="140" />
        <el-table-column label="Status" width="120">
          <template #default="{ row }"><el-tag :type="statusType(getStatus(row))" size="small">{{ getStatus(row) }}</el-tag></template>
        </el-table-column>
        <el-table-column label="Source" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ getSource(row) }}</template>
        </el-table-column>
        <el-table-column label="Restore Size" width="130">
          <template #default="{ row }">{{ getRestoreSize(row) }}</template>
        </el-table-column>
        <el-table-column label="Snapshot Class" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ getSnapshotClass(row) }}</template>
        </el-table-column>
        <el-table-column prop="age" label="Age" width="120" />
        <el-table-column label="Actions" width="180" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">Delete</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <el-dialog v-model="yamlDialogVisible" title="VolumeSnapshot YAML" width="70%" top="5vh" destroy-on-close>
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
