<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'
import { getConfigMapList, getConfigMapDetail, deleteConfigMap, getNamespaceList, extractNamespaceNames, transformConfigMaps } from '@/api/resource'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'
import YamlDrawer from '@/components/YamlDrawer.vue'
import ConfigDataViewer from '@/components/ConfigDataViewer.vue'

const router = useRouter()
const loading = ref(false)
const configMapList = ref<any[]>([])
const namespaceList = ref<string[]>([])
const selectedNamespace = ref('')
const searchName = ref('')
const selectedRows = ref<any[]>([])
const yamlDialogVisible = ref(false)
const yamlTarget = ref<{ namespace: string; name: string } | null>(null)
const dataDialogVisible = ref(false)
const dataDialogTitle = ref('')
const dataEntries = ref<{ key: string; value: string }[]>([])
const dataLoading = ref(false)

const filteredList = computed(() => {
  if (!searchName.value) return configMapList.value
  const keyword = searchName.value.toLowerCase()
  return configMapList.value.filter((d) => d.name?.toLowerCase().includes(keyword))
})

async function fetchNamespaces() {
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = extractNamespaceNames(res.data)
  } catch { /* ignore */ }
}

async function fetchConfigMaps() {
  loading.value = true
  try {
    const params: any = {}
    if (selectedNamespace.value) params.namespace = selectedNamespace.value
    const res: any = await getConfigMapList(params)
    const items = res.data?.items || res.data || []
    configMapList.value = transformConfigMaps(items)
  } catch {
    // Silently handle — resource may not exist in cluster
  } finally { loading.value = false }
}

function handleNamespaceChange() { fetchConfigMaps() }
function handleSelectionChange(rows: any[]) { selectedRows.value = rows }

function handleViewYaml(row: any) {
  yamlTarget.value = { namespace: row.namespace, name: row.name }
  yamlDialogVisible.value = true
}

async function handleViewData(row: any) {
  dataLoading.value = true; dataDialogVisible.value = true; dataDialogTitle.value = `数据字典: ${row.name}`; dataEntries.value = []
  try {
    const res: any = await getConfigMapDetail({ name: row.name, namespace: row.namespace })
    const data = res.data?.data || {}
    const binaryData = res.data?.binaryData || {}
    const merged = [
      ...Object.entries(data).map(([key, value]) => ({ key, value: String(value ?? '') })),
      ...Object.entries(binaryData).map(([key, value]) => ({ key, value: String(value ?? '') })),
    ]
    dataEntries.value = merged
  } catch (e: any) { ElMessage.error(e?.message || '加载数据失败'); dataDialogVisible.value = false }
  finally { dataLoading.value = false }
}

function handleDetail(row: any) { router.push(`/config/configmaps/${row.namespace}/${row.name}`) }

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(`确定要删除命名空间 "${row.namespace}" 中的数据字典 "${row.name}" 吗？`, '确认', { type: 'warning' })
    await deleteConfigMap({ name: row.name, namespace: row.namespace })
    ElMessage.success('删除成功'); fetchConfigMaps()
  } catch { /* cancelled */ }
}

async function handleBatchDelete() {
  if (!selectedRows.value.length) return
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedRows.value.length} 个数据字典吗？`, '确认', { type: 'warning' })
    const results = await Promise.allSettled(
      selectedRows.value.map((row: any) => deleteConfigMap({ name: row.name, namespace: row.namespace }))
    )
    const count = results.filter((r) => r.status === 'fulfilled').length
    const failed = results.length - count
    ElMessage.success(`已成功删除 ${count} 个数据字典${failed ? `，${failed} 个失败` : ''}`); fetchConfigMaps()
  } catch { /* cancelled */ }
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh, setIntervalOption } = useAutoRefresh(fetchConfigMaps)

onMounted(() => { fetchNamespaces(); fetchConfigMaps() })
</script>

<template>
  <div class="page-container">
    <ResourceListToolbar
      :search-value="searchName"
      v-model:namespace-value="selectedNamespace"
      :namespace-list="namespaceList"
      :show-total-count="false"
      :selected-count="selectedRows.length"
      @search-input="(val: string) => searchName = val"
      @namespace-change="handleNamespaceChange"
    >
      <template #actions>
        <el-button type="success" @click="router.push('/config/configmaps/create')">
          <el-icon><Plus /></el-icon> 创建
        </el-button>
        <el-button type="danger" :disabled="!selectedRows.length" @click="handleBatchDelete">
          <el-icon><Delete /></el-icon> 删除 ({{ selectedRows.length }})
        </el-button>
      </template>
      <template #extra>
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
      </template>
    </ResourceListToolbar>
    <el-card shadow="never" class="table-card">
      <el-table :data="filteredList" v-loading="loading" stripe @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="45" />
        <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip>
          <template #default="{ row }"><el-button link type="primary" @click="handleDetail(row)">{{ row.name }}</el-button></template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="140" />
        <el-table-column label="标签" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <template v-if="row.labels && Object.keys(row.labels).length">
              <el-tag v-for="(val, key, idx) in row.labels" :key="key" size="small" class="label-tag" v-show="idx < 3">
                {{ key }}={{ val }}
              </el-tag>
              <el-tag v-if="Object.keys(row.labels).length > 3" size="small" type="info" effect="plain">
                +{{ Object.keys(row.labels).length - 3 }}
              </el-tag>
            </template>
            <span v-else class="no-labels">-</span>
          </template>
        </el-table-column>
        <el-table-column label="数据键数量" width="120">
          <template #default="{ row }"><el-tag size="small">{{ row.data_keys_count }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="age" label="创建时间" width="120" />
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
            <el-button size="small" type="primary" @click="handleViewData(row)">查看数据</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <YamlDrawer
      v-model="yamlDialogVisible"
      resource-type="configmap"
      :namespace="yamlTarget?.namespace || ''"
      :name="yamlTarget?.name || ''"
      @saved="fetchConfigMaps"
    />
    <el-drawer
      v-model="dataDialogVisible"
      :title="dataDialogTitle"
      size="85%"
      direction="rtl"
      class="data-drawer"
      :body-style="{ padding: '0', height: '100%' }"
      :destroy-on-close="true"
    >
      <div v-loading="dataLoading" style="height: calc(100vh - 52px);">
        <ConfigDataViewer :entries="dataEntries" :loading="dataLoading" />
      </div>
    </el-drawer>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.table-card { border-radius: 8px; }
.action-buttons {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: 4px;
}
.action-buttons .el-button + .el-button {
  margin-left: 0;
}
.label-tag {
  margin-right: 4px;
  max-width: 160px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.no-labels {
  color: var(--el-text-color-placeholder);
  font-size: 12px;
}
</style>

<style>
.data-drawer .el-drawer__header {
  padding: 6px 16px;
  margin-bottom: 0;
  min-height: auto;
}
</style>
