<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Search } from '@element-plus/icons-vue'
import { getPvList, getPvYaml, updatePvYaml, deletePv, transformPvs, type Pv } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const router = useRouter()
const loading = ref(false)
const pvList = ref<Pv[]>([])
const searchName = ref('')
const selectedRows = ref<any[]>([])

// YAML drawer state
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)
const yamlSaving = ref(false)
const yamlTarget = ref<any>(null)

const filteredList = computed(() => {
  if (!searchName.value) return pvList.value
  const keyword = searchName.value.toLowerCase()
  return pvList.value.filter((d) => d.name?.toLowerCase().includes(keyword))
})

async function fetchPvs() {
  loading.value = true
  try {
    const res: any = await getPvList()
    const items = res.data?.items || res.data || []
    pvList.value = transformPvs(items)
  } catch {
    // Silently handle — resource may not exist in cluster
  } finally {
    loading.value = false
  }
}

function handleSelectionChange(rows: any[]) {
  selectedRows.value = rows
}

function statusType(status: string) {
  const s = (status || '').toLowerCase()
  if (s === 'bound') return 'success'
  if (s === 'available') return 'primary'
  if (s === 'released') return 'warning'
  if (s === 'failed') return 'danger'
  return 'info'
}

async function handleViewYaml(row: any) {
  yamlTarget.value = row
  yamlDialogVisible.value = true
  yamlLoading.value = true
  yamlContent.value = ''
  try {
    const res: any = await getPvYaml({ name: row.name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load YAML')
    yamlDialogVisible.value = false
  } finally {
    yamlLoading.value = false
  }
}

async function handleSaveYaml() {
  if (!yamlTarget.value) return
  yamlSaving.value = true
  try {
    await updatePvYaml({ name: yamlTarget.value.name, yaml: yamlContent.value })
    ElMessage.success('YAML saved successfully')
    yamlDialogVisible.value = false
    fetchPvs()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to save YAML')
  } finally {
    yamlSaving.value = false
  }
}

function handleCancelYaml() {
  yamlDialogVisible.value = false
}

function handleDetail(row: any) {
  router.push(`/storage/pvs/${row.name}`)
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(`删除持久卷 "${row.name}"?`, '确认', { type: 'warning' })
    await deletePv({ name: row.name })
    ElMessage.success('删除成功')
    fetchPvs()
  } catch {
    /* cancelled */
  }
}

async function handleBatchDelete() {
  if (!selectedRows.value.length) return
  try {
    await ElMessageBox.confirm(`删除选中的 ${selectedRows.value.length} 个持久卷?`, '确认', { type: 'warning' })
    const results = await Promise.allSettled(
      selectedRows.value.map((row) => deletePv({ name: row.name }))
    )
    const successCount = results.filter((r) => r.status === 'fulfilled').length
    const failCount = results.filter((r) => r.status === 'rejected').length
    if (failCount > 0) {
      ElMessage.warning(`删除成功 ${successCount} 个，失败 ${failCount} 个`)
    } else {
      ElMessage.success(`成功删除 ${successCount} 个持久卷`)
    }
    fetchPvs()
  } catch {
    /* cancelled */
  }
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchPvs)

onMounted(fetchPvs)
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <el-input
          v-model="searchName"
          placeholder="搜索名称"
          style="width: 220px;"
          clearable
        >
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
        <el-button type="success" @click="router.push('/storage/pvs/create')">
          <el-icon><Plus /></el-icon> 创建
        </el-button>
        <el-button type="danger" :disabled="!selectedRows.length" @click="handleBatchDelete">
          <el-icon><Delete /></el-icon> 删除 ({{ selectedRows.length }})
        </el-button>
        <span class="total-count" v-if="pvList.length">总计: {{ pvList.length }}</span>
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
        <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <el-button link type="primary" @click="handleDetail(row)">{{ row.name }}</el-button>
          </template>
        </el-table-column>
        <el-table-column prop="capacity" label="容量" width="120" />
        <el-table-column prop="access_modes" label="访问模式" min-width="160" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="claim" label="声明" min-width="180" show-overflow-tooltip />
        <el-table-column prop="storage_class" label="存储类" min-width="140" show-overflow-tooltip />
        <el-table-column prop="reclaim_policy" label="回收策略" width="100" />
        <el-table-column prop="age" label="存活时间" width="120" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- YAML Drawer -->
    <el-drawer
      v-model="yamlDialogVisible"
      title="PersistentVolume YAML"
      size="85%"
      direction="rtl"
      class="yaml-drawer"
      :body-style="{ padding: '0', height: '100%' }"
    >
      <div v-loading="yamlLoading" style="height: calc(100vh - 52px);">
        <YamlEditor
          v-model="yamlContent"
          height="100%"
          auto-format
          show-save-buttons
          :saving="yamlSaving"
          @save="handleSaveYaml"
          @cancel="handleCancelYaml"
        />
      </div>
    </el-drawer>
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
.total-count {
  color: var(--el-text-color-secondary);
  font-size: 13px;
  margin-left: auto;
}
</style>

<style>
.yaml-drawer .el-drawer__header {
  padding: 6px 16px;
  margin-bottom: 0;
  min-height: auto;
}
</style>
