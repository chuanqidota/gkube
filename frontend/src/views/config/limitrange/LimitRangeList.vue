<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'
import { getLimitRangeList, deleteLimitRange, getNamespaceList, extractNamespaceNames } from '@/api/resource'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'
import YamlDrawer from '@/components/YamlDrawer.vue'

const router = useRouter()
const loading = ref(false)
const lrList = ref<any[]>([])
const namespaceList = ref<string[]>([])
const selectedNamespace = ref('')
const searchName = ref('')
const selectedRows = ref<any[]>([])
const yamlDialogVisible = ref(false)
const yamlTarget = ref<{ namespace: string; name: string } | null>(null)

const filteredList = computed(() => {
  if (!searchName.value) return lrList.value
  const keyword = searchName.value.toLowerCase()
  return lrList.value.filter((d) => d.name?.toLowerCase().includes(keyword))
})

async function fetchNamespaces() {
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = extractNamespaceNames(res.data)
  } catch { /* ignore */ }
}

async function fetchLimitRanges() {
  loading.value = true
  try {
    const params: any = {}
    if (selectedNamespace.value) params.namespace = selectedNamespace.value
    const res: any = await getLimitRangeList(params)
    lrList.value = res.data || []
  } catch {
    // Silently handle — resource may not exist in cluster
  } finally { loading.value = false }
}

function handleNamespaceChange() { fetchLimitRanges() }
function handleSelectionChange(rows: any[]) { selectedRows.value = rows }
function handleDetail(row: any) { router.push(`/config/limitranges/${row.namespace}/${row.name}`) }

function handleViewYaml(row: any) {
  yamlTarget.value = { namespace: row.namespace, name: row.name }
  yamlDialogVisible.value = true
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(`Delete LimitRange "${row.name}" in namespace "${row.namespace}"?`, 'Confirm', { type: 'warning' })
    await deleteLimitRange({ namespace: row.namespace, name: row.name })
    ElMessage.success('Deleted'); fetchLimitRanges()
  } catch { /* cancelled */ }
}

async function handleBatchDelete() {
  if (!selectedRows.value.length) return
  try {
    await ElMessageBox.confirm(`Delete ${selectedRows.value.length} selected LimitRange(s)?`, 'Confirm', { type: 'warning' })
    let count = 0
    for (const row of selectedRows.value) {
      try { await deleteLimitRange({ namespace: row.namespace, name: row.name }); count++ } catch { /* continue */ }
    }
    ElMessage.success(`Deleted ${count} LimitRange(s)`); fetchLimitRanges()
  } catch { /* cancelled */ }
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh, setIntervalOption } = useAutoRefresh(fetchLimitRanges)

onMounted(() => { fetchNamespaces(); fetchLimitRanges() })
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
        <el-button type="success" @click="router.push('/config/limitranges/create')">
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
        <el-table-column prop="name" label="Name" min-width="200" show-overflow-tooltip>
          <template #default="{ row }"><el-button link type="primary" @click="handleDetail(row)">{{ row.name }}</el-button></template>
        </el-table-column>
        <el-table-column prop="namespace" label="Namespace" width="140" />
        <el-table-column label="Limits" min-width="300">
          <template #default="{ row }">
            <div v-for="(limit, i) in (row.limits || [])" :key="i" style="font-size: 12px; margin-bottom: 4px;">
              <el-tag size="small" style="margin-right: 4px;">{{ limit.type }}</el-tag>
              <span v-for="(v, k) in (limit.max || {})" :key="k">Max {{ k }}: {{ v }} </span>
              <span v-for="(v, k) in (limit.min || {})" :key="k">Min {{ k }}: {{ v }} </span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="age" label="Age" width="180" />
        <el-table-column label="Actions" width="160" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <YamlDrawer
      v-model="yamlDialogVisible"
      resource-type="limitrange"
      :namespace="yamlTarget?.namespace || ''"
      :name="yamlTarget?.name || ''"
      @saved="fetchLimitRanges"
    />
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.table-card { border-radius: 8px; }
</style>
