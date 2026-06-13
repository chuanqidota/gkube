<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Plus, Delete, Search } from '@element-plus/icons-vue'
import { getResourceQuotaList, getResourceQuotaYaml, deleteResourceQuota, getNamespaceList } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const loading = ref(false)
const rqList = ref<any[]>([])
const namespaceList = ref<string[]>([])
const selectedNamespace = ref('')
const searchName = ref('')
const selectedRows = ref<any[]>([])
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)

const filteredList = computed(() => {
  if (!searchName.value) return rqList.value
  const keyword = searchName.value.toLowerCase()
  return rqList.value.filter((d) => d.name?.toLowerCase().includes(keyword))
})

async function fetchNamespaces() {
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = (res.data || []).map((ns: any) => ns.name || ns)
  } catch { /* ignore */ }
}

async function fetchResourceQuotas() {
  loading.value = true
  try {
    const params: any = {}
    if (selectedNamespace.value) params.namespace = selectedNamespace.value
    const res: any = await getResourceQuotaList(params)
    rqList.value = res.data || []
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load ResourceQuotas')
  } finally { loading.value = false }
}

function handleNamespaceChange() { fetchResourceQuotas() }
function handleSelectionChange(rows: any[]) { selectedRows.value = rows }

async function handleViewYaml(row: any) {
  yamlDialogVisible.value = true; yamlLoading.value = true; yamlContent.value = ''
  try {
    const res: any = await getResourceQuotaYaml({ namespace: row.namespace, name: row.name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) { ElMessage.error(e?.message || 'Failed to load YAML'); yamlDialogVisible.value = false }
  finally { yamlLoading.value = false }
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(`Delete ResourceQuota "${row.name}" in namespace "${row.namespace}"?`, 'Confirm', { type: 'warning' })
    await deleteResourceQuota({ namespace: row.namespace, name: row.name })
    ElMessage.success('Deleted'); fetchResourceQuotas()
  } catch { /* cancelled */ }
}

async function handleBatchDelete() {
  if (!selectedRows.value.length) return
  try {
    await ElMessageBox.confirm(`Delete ${selectedRows.value.length} selected ResourceQuota(s)?`, 'Confirm', { type: 'warning' })
    let count = 0
    for (const row of selectedRows.value) {
      try { await deleteResourceQuota({ namespace: row.namespace, name: row.name }); count++ } catch { /* continue */ }
    }
    ElMessage.success(`Deleted ${count} ResourceQuota(s)`); fetchResourceQuotas()
  } catch { /* cancelled */ }
}

onMounted(() => { fetchNamespaces(); fetchResourceQuotas() })
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
        <el-button type="primary" @click="fetchResourceQuotas"><el-icon><Refresh /></el-icon> Refresh</el-button>
        <el-button type="danger" :disabled="!selectedRows.length" @click="handleBatchDelete"><el-icon><Delete /></el-icon> Delete ({{ selectedRows.length }})</el-button>
      </div>
    </el-card>
    <el-card shadow="never" class="table-card">
      <el-table :data="filteredList" v-loading="loading" stripe @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="45" />
        <el-table-column prop="name" label="Name" min-width="200" show-overflow-tooltip />
        <el-table-column prop="namespace" label="Namespace" width="140" />
        <el-table-column label="Hard Limits" min-width="250">
          <template #default="{ row }">
            <div v-for="(v, k) in (row.hard || {})" :key="k" style="font-size: 12px;">{{ k }}: {{ v }}</div>
          </template>
        </el-table-column>
        <el-table-column label="Used" min-width="250">
          <template #default="{ row }">
            <div v-for="(v, k) in (row.used || {})" :key="k" style="font-size: 12px;">{{ k }}: {{ v }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="age" label="Age" width="180" />
        <el-table-column label="Actions" width="160" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">Delete</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <el-dialog v-model="yamlDialogVisible" title="ResourceQuota YAML" width="70%" top="5vh" destroy-on-close>
      <div v-loading="yamlLoading"><YamlEditor v-model="yamlContent" height="500px" read-only /></div>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; gap: 12px; align-items: center; flex-wrap: wrap; }
.table-card { border-radius: 8px; }
</style>
