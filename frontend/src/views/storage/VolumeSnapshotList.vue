<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'
import { getVolumeSnapshotList, deleteVolumeSnapshot, getNamespaceList, extractNamespaceNames } from '@/api/resource'
import { useI18n } from 'vue-i18n'
import YamlDrawer from '@/components/YamlDrawer.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'

const { t } = useI18n()
const router = useRouter()
const loading = ref(false)
const snapshotList = ref<any[]>([])
const namespaceList = ref<string[]>([])
const selectedNamespace = ref('')
const searchName = ref('')
const selectedRows = ref<any[]>([])
const yamlDialogVisible = ref(false)
const yamlTarget = ref<{ namespace: string; name: string } | null>(null)

const filteredList = computed(() => {
  if (!searchName.value) return snapshotList.value
  const keyword = searchName.value.toLowerCase()
  return snapshotList.value.filter((d) => d.name?.toLowerCase().includes(keyword))
})

function onSearchInput(val: string) { searchName.value = val }

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
  } catch {
    // Silently handle — resource may not exist in cluster
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
  return row.status?.restoreSize || '-'
}

function getSnapshotClass(row: any): string {
  return row.spec?.volumeSnapshotClassName || '-'
}

function getSource(row: any): string {
  const src = row.spec?.source
  if (!src) return '-'
  if (src.persistentVolumeClaimName) return `PVC: ${src.persistentVolumeClaimName}`
  if (src.volumeSnapshotContentName) return `VSC: ${src.volumeSnapshotContentName}`
  return '-'
}

function handleViewYaml(row: any) {
  yamlTarget.value = { namespace: row.namespace, name: row.name }
  yamlDialogVisible.value = true
}

function handleDetail(row: any) { router.push(`/storage/volumesnapshots/${row.namespace}/${row.name}`) }

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(
      t('storage.deleteSnapshotConfirm', { name: row.name, namespace: row.namespace }),
      t('common.confirm'),
      { type: 'warning' }
    )
    await deleteVolumeSnapshot({ name: row.name, namespace: row.namespace })
    ElMessage.success(t('common.delete') + ' ' + t('common.success'))
    fetchSnapshots()
  } catch { /* cancelled */ }
}

async function handleBatchDelete() {
  if (!selectedRows.value.length) return
  try {
    await ElMessageBox.confirm(
      t('storage.deleteSnapshotBatchConfirm', { count: selectedRows.value.length }),
      t('common.confirm'),
      { type: 'warning' }
    )
    let count = 0
    for (const row of selectedRows.value) {
      try { await deleteVolumeSnapshot({ name: row.name, namespace: row.namespace }); count++ } catch { /* continue */ }
    }
    ElMessage.success(t('common.delete') + ` ${count} ` + t('storage.volumeSnapshot'))
    fetchSnapshots()
  } catch { /* cancelled */ }
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchSnapshots)

onMounted(() => { fetchNamespaces(); fetchSnapshots() })
</script>

<template>
  <div class="page-container">
    <ResourceListToolbar
      :search-value="searchName"
      v-model:namespace-value="selectedNamespace"
      :namespace-list="namespaceList"
      :show-total-count="false"
      :selected-count="selectedRows.length"
      @search-input="onSearchInput"
      @namespace-change="handleNamespaceChange"
    >
      <template #actions>
        <el-button type="success" @click="router.push('/storage/volumesnapshots/create')"><el-icon><Plus /></el-icon> {{ t('common.create') }}</el-button>
        <el-button type="danger" :disabled="!selectedRows.length" @click="handleBatchDelete"><el-icon><Delete /></el-icon> {{ t('common.delete') }} ({{ selectedRows.length }})</el-button>
      </template>
      <template #extra>
        <AutoRefreshToolbar
          :is-running="isRunning"
          :countdown="countdown"
          :current-interval="currentInterval"
          :available-intervals="availableIntervals"
          :loading="loading"
          @refresh="manualRefresh()"
          @toggle="toggle()"
          @interval-change="setIntervalOption"
        />
      </template>
    </ResourceListToolbar>
    <el-card shadow="never" class="table-card">
      <el-table :data="filteredList" v-loading="loading" stripe @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="45" />
        <el-table-column prop="name" :label="t('common.name')" min-width="180" show-overflow-tooltip>
          <template #default="{ row }"><el-button link type="primary" @click="handleDetail(row)">{{ row.name }}</el-button></template>
        </el-table-column>
        <el-table-column prop="namespace" :label="t('common.namespace_label')" width="140" />
        <el-table-column :label="t('common.status')" width="120">
          <template #default="{ row }"><el-tag :type="statusType(getStatus(row))" size="small">{{ getStatus(row) }}</el-tag></template>
        </el-table-column>
        <el-table-column :label="t('storage.source')" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">{{ getSource(row) }}</template>
        </el-table-column>
        <el-table-column :label="t('storage.restoreSize')" width="130">
          <template #default="{ row }">{{ getRestoreSize(row) }}</template>
        </el-table-column>
        <el-table-column :label="t('storage.snapshotClass')" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ getSnapshotClass(row) }}</template>
        </el-table-column>
        <el-table-column prop="age" :label="t('common.age')" width="120" />
        <el-table-column :label="t('common.actions')" width="180" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">{{ t('common.yaml') }}</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">{{ t('common.delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <YamlDrawer
      v-model="yamlDialogVisible"
      resource-type="volumesnapshot"
      :namespace="yamlTarget?.namespace || ''"
      :name="yamlTarget?.name || ''"
      @saved="fetchSnapshots"
    />
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.table-card { border-radius: 8px; }
</style>
