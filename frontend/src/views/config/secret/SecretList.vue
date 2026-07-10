<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Search } from '@element-plus/icons-vue'
import { getSecretList, getSecretDetail, deleteSecret, getNamespaceList, extractNamespaceNames, transformSecrets } from '@/api/resource'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import YamlDrawer from '@/components/YamlDrawer.vue'

const router = useRouter()
const loading = ref(false)
const secretList = ref<any[]>([])
const namespaceList = ref<string[]>([])
const selectedNamespace = ref('')
const searchName = ref('')
const selectedRows = ref<any[]>([])
const yamlDialogVisible = ref(false)
const yamlTarget = ref<{ namespace: string; name: string } | null>(null)
const dataDialogVisible = ref(false)
const dataDialogTitle = ref('')
const dataEntries = ref<{ key: string; rawValue: string; decodedValue: string }[]>([])
const dataLoading = ref(false)
const showDecoded = ref(true)

const filteredList = computed(() => {
  if (!searchName.value) return secretList.value
  const keyword = searchName.value.toLowerCase()
  return secretList.value.filter((d) => d.name?.toLowerCase().includes(keyword))
})

async function fetchNamespaces() {
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = extractNamespaceNames(res.data)
  } catch { /* ignore */ }
}

async function fetchSecrets() {
  loading.value = true
  try {
    const params: any = {}
    if (selectedNamespace.value) params.namespace = selectedNamespace.value
    const res: any = await getSecretList(params)
    const items = res.data?.items || res.data || []
    secretList.value = transformSecrets(items)
  } catch {
    // Silently handle — resource may not exist in cluster
  } finally { loading.value = false }
}

function handleNamespaceChange() { fetchSecrets() }
function handleSelectionChange(rows: any[]) { selectedRows.value = rows }

function base64Decode(str: string): string { try { return atob(str) } catch { return str } }

function handleViewYaml(row: any) {
  yamlTarget.value = { namespace: row.namespace, name: row.name }
  yamlDialogVisible.value = true
}

async function handleViewData(row: any) {
  dataLoading.value = true; dataDialogVisible.value = true; dataDialogTitle.value = `Secret: ${row.name}`; dataEntries.value = []
  try {
    const res: any = await getSecretDetail({ name: row.name, namespace: row.namespace })
    const data = res.data?.data || res.data || {}
    dataEntries.value = Object.entries(data).map(([key, value]) => {
      const rawValue = String(value ?? '')
      return { key, rawValue, decodedValue: base64Decode(rawValue) }
    })
  } catch (e: any) { ElMessage.error(e?.message || 'Failed to load data'); dataDialogVisible.value = false }
  finally { dataLoading.value = false }
}

function handleDetail(row: any) { router.push(`/config/secrets/${row.namespace}/${row.name}`) }

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(`Delete Secret "${row.name}" in namespace "${row.namespace}"?`, 'Confirm', { type: 'warning' })
    await deleteSecret({ name: row.name, namespace: row.namespace })
    ElMessage.success('Deleted'); fetchSecrets()
  } catch { /* cancelled */ }
}

async function handleBatchDelete() {
  if (!selectedRows.value.length) return
  try {
    await ElMessageBox.confirm(`Delete ${selectedRows.value.length} selected Secret(s)?`, 'Confirm', { type: 'warning' })
    let count = 0
    for (const row of selectedRows.value) {
      try { await deleteSecret({ name: row.name, namespace: row.namespace }); count++ } catch { /* continue */ }
    }
    ElMessage.success(`Deleted ${count} Secret(s)`); fetchSecrets()
  } catch { /* cancelled */ }
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh, setIntervalOption } = useAutoRefresh(fetchSecrets)

onMounted(() => { fetchNamespaces(); fetchSecrets() })
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
        <AutoRefreshToolbar
          :is-running="isRunning"
          :countdown="countdown"
          :current-interval="currentInterval"
          :available-intervals="availableIntervals"
          :loading="loading"
          @refresh="refresh"
          @toggle="toggle"
          @interval-change="setIntervalOption"
        />
        <el-button type="success" @click="router.push('/config/secrets/create')"><el-icon><Plus /></el-icon> 创建</el-button>
        <el-button type="danger" :disabled="!selectedRows.length" @click="handleBatchDelete"><el-icon><Delete /></el-icon> 删除 ({{ selectedRows.length }})</el-button>
      </div>
    </el-card>
    <el-card shadow="never" class="table-card">
      <el-table :data="filteredList" v-loading="loading" stripe @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="45" />
        <el-table-column prop="name" label="Name" min-width="200" show-overflow-tooltip>
          <template #default="{ row }"><el-button link type="primary" @click="handleDetail(row)">{{ row.name }}</el-button></template>
        </el-table-column>
        <el-table-column prop="namespace" label="Namespace" width="140" />
        <el-table-column prop="type" label="Type" min-width="160" show-overflow-tooltip />
        <el-table-column label="Data Keys" width="120">
          <template #default="{ row }"><el-tag size="small">{{ row.data_keys_count ?? Object.keys(row.data || {}).length }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="age" label="Age" width="120" />
        <el-table-column label="Actions" width="240" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
            <el-button size="small" type="primary" @click="handleViewData(row)">Data</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <YamlDrawer
      v-model="yamlDialogVisible"
      resource-type="secret"
      :namespace="yamlTarget?.namespace || ''"
      :name="yamlTarget?.name || ''"
      @saved="fetchSecrets"
    />
    <el-dialog v-model="dataDialogVisible" :title="dataDialogTitle" width="60%" top="8vh">
      <div style="margin-bottom: 12px;"><el-switch v-model="showDecoded" active-text="Decoded (Base64)" inactive-text="Raw (Base64)" /></div>
      <div v-loading="dataLoading">
        <el-table :data="dataEntries" stripe style="width: 100%" max-height="400">
          <el-table-column prop="key" label="Key" min-width="200" show-overflow-tooltip />
          <el-table-column label="Value" min-width="300">
            <template #default="{ row }"><div style="white-space: pre-wrap; word-break: break-all; max-height: 100px; overflow-y: auto;">{{ showDecoded ? row.decodedValue : row.rawValue }}</div></template>
          </el-table-column>
        </el-table>
        <el-empty v-if="!dataLoading && dataEntries.length === 0" description="No data" />
      </div>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; gap: 12px; align-items: center; flex-wrap: wrap; }
.table-card { border-radius: 8px; }
</style>
