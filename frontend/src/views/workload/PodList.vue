<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Delete, Search, Monitor } from '@element-plus/icons-vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import { getPodList, getPodYaml, deletePod, getNamespaceList, extractNamespaceNames, transformPods } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const { t } = useI18n()
const router = useRouter()
const loading = ref(false)
const podList = ref<any[]>([])
const namespaceList = ref<string[]>([])
const selectedNamespace = ref('')
const searchName = ref('')
const selectedRows = ref<any[]>([])

// YAML dialog
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)

const filteredList = computed(() => {
  if (!searchName.value) return podList.value
  const keyword = searchName.value.toLowerCase()
  return podList.value.filter((p) => p.name?.toLowerCase().includes(keyword))
})

async function fetchNamespaces() {
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = extractNamespaceNames(res.data)
  } catch {
    // ignore
  }
}

async function fetchPods() {
  loading.value = true
  try {
    const params: any = {}
    if (selectedNamespace.value) params.namespace = selectedNamespace.value
    const res: any = await getPodList(params)
    // API returns raw K8s Pod objects; transform to simplified display format
    const items = res.data?.items || res.data || []
    podList.value = transformPods(items)
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load pods')
  } finally {
    loading.value = false
  }
}

function handleNamespaceChange() {
  fetchPods()
}

function handleSelectionChange(rows: any[]) {
  selectedRows.value = rows
}

async function handleViewYaml(row: any) {
  yamlDialogVisible.value = true
  yamlLoading.value = true
  yamlContent.value = ''
  try {
    const res: any = await getPodYaml({
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

function getClusterName(): string {
  try {
    const saved = localStorage.getItem('gkube_cluster')
    if (saved) {
      const c = JSON.parse(saved)
      return c?.clusterName || c?.cluster_name || c?.name || ''
    }
  } catch { /* ignore */ }
  return ''
}

function handleViewLogs(row: any) {
  const cluster = getClusterName()
  window.open(`/logs?namespace=${row.namespace}&pod=${row.name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
}

function handleExec(row: any) {
  const cluster = getClusterName()
  window.open(`/terminal?namespace=${row.namespace}&pod=${row.name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
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
    await deletePod({ namespace: row.namespace, name: row.name })
    ElMessage.success('Pod deleted')
    fetchPods()
  } catch {
    // cancelled
  }
}

async function handleBatchDelete() {
  if (!selectedRows.value.length) return
  try {
    await ElMessageBox.confirm(
      `Delete ${selectedRows.value.length} selected pod(s)?`,
      'Confirm',
      { type: 'warning' }
    )
    let successCount = 0
    for (const row of selectedRows.value) {
      try {
        await deletePod({ namespace: row.namespace, name: row.name })
        successCount++
      } catch {
        // continue
      }
    }
    ElMessage.success(`Deleted ${successCount} pod(s)`)
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

const { isRunning, countdown, toggle, refresh: autoRefresh } = useAutoRefresh(fetchPods, 15000)

onMounted(() => {
  fetchNamespaces()
  fetchPods()
})
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <el-input
          v-model="searchName"
          :placeholder="t('common.searchByName')"
          style="width: 220px;"
          clearable
        >
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select
          v-model="selectedNamespace"
          :placeholder="t('common.allNamespaces')"
          clearable
          style="width: 180px;"
          @change="handleNamespaceChange"
        >
          <el-option v-for="ns in namespaceList" :key="ns" :label="ns" :value="ns" />
        </el-select>
        <el-button type="primary" @click="autoRefresh()">
          <el-icon><Refresh /></el-icon> {{ t('common.refresh') }} ({{ countdown }}s)
        </el-button>
        <el-button @click="toggle()" :type="isRunning ? 'warning' : 'success'" size="default">
          {{ isRunning ? t('common.paused') : t('common.resume') }}
        </el-button>
        <el-button type="danger" :disabled="!selectedRows.length" @click="handleBatchDelete">
          <el-icon><Delete /></el-icon> Delete ({{ selectedRows.length }})
        </el-button>
      </div>
    </el-card>

    <el-card shadow="never" class="table-card">
      <el-table
        :data="filteredList"
        v-loading="loading"
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="45" />
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
        <el-table-column label="Actions" width="280" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
            <el-button size="small" type="primary" @click="handleViewLogs(row)">
              <el-icon><Monitor /></el-icon> Logs
            </el-button>
            <el-button size="small" type="success" @click="handleExec(row)">Exec</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- YAML Dialog -->
    <el-dialog v-model="yamlDialogVisible" title="Pod YAML" width="70%" top="5vh" destroy-on-close>
      <div v-loading="yamlLoading">
        <YamlEditor v-model="yamlContent" height="500px" read-only auto-format />
      </div>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container {
  padding: 20px;
}
.filter-card {
  margin-bottom: 16px;
}
.filter-bar {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}
.table-card {
  border-radius: 8px;
}
</style>
