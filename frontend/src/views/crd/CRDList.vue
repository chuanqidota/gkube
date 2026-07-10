<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'

import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Plus } from '@element-plus/icons-vue'
import { getCrdList, deleteCrd } from '@/api/resource'
import YamlDrawer from '@/components/YamlDrawer.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'

const router = useRouter()
const loading = ref(false)
const crdList = ref<any[]>([])
const searchName = ref('')
const yamlDialogVisible = ref(false)
const yamlTarget = ref<{ name: string } | null>(null)

const filteredList = computed(() => {
  if (!searchName.value) return crdList.value
  const keyword = searchName.value.toLowerCase()
  return crdList.value.filter((d) => d.name?.toLowerCase().includes(keyword) || d.kind?.toLowerCase().includes(keyword))
})

async function fetchCrds() {
  loading.value = true
  try {
    const res: any = await getCrdList()
    crdList.value = res.data || []
  } catch {
    // Silently handle — resource may not exist in cluster
  } finally { loading.value = false }
}

function handleViewYaml(row: any) {
  yamlTarget.value = { name: row.name }
  yamlDialogVisible.value = true
}

function handleBrowse(row: any) {
  const group = row.group
  const version = row.versions?.[0] || 'v1'
  const resource = row.plural
  router.push(`/crd/resources?group=${group}&version=${version}&resource=${resource}&scope=${row.scope}`)
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(`Delete CRD "${row.name}"? This will delete ALL custom resources of this type!`, 'Confirm', { type: 'error' })
    await deleteCrd({ name: row.name })
    ElMessage.success('CRD deleted')
    fetchCrds()
  } catch { /* cancelled */ }
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchCrds)

onMounted(fetchCrds)
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <el-input v-model="searchName" placeholder="Search by name or kind" style="width: 280px;" clearable>
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
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
        <el-button type="success" @click="router.push('/crd/create')"><el-icon><Plus /></el-icon> 创建 CRD</el-button>
      </div>
    </el-card>
    <el-card shadow="never" class="table-card">
      <el-table :data="filteredList" v-loading="loading" stripe>
        <el-table-column prop="kind" label="Kind" min-width="180" show-overflow-tooltip>
          <template #default="{ row }"><el-button link type="primary" @click="handleBrowse(row)">{{ row.kind }}</el-button></template>
        </el-table-column>
        <el-table-column prop="name" label="Name" min-width="280" show-overflow-tooltip />
        <el-table-column prop="group" label="Group" min-width="180" show-overflow-tooltip />
        <el-table-column label="Versions" width="140">
          <template #default="{ row }"><el-tag v-for="v in (row.versions || [])" :key="v" size="small" style="margin-right: 4px;">{{ v }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="scope" label="Scope" width="120">
          <template #default="{ row }"><el-tag :type="row.scope === 'Namespaced' ? 'info' : 'warning'" size="small">{{ row.scope }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="age" label="Age" width="180" />
        <el-table-column label="Actions" width="180" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <YamlDrawer
      v-model="yamlDialogVisible"
      resource-type="crd"
      :name="yamlTarget?.name || ''"
      @saved="fetchCrds"
    />
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; gap: 12px; align-items: center; flex-wrap: wrap; }
.table-card { border-radius: 8px; }
</style>
