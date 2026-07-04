<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Plus, Delete, Search } from '@element-plus/icons-vue'
import { getVolumeSnapshotClassList, getVolumeSnapshotClassYaml, deleteVolumeSnapshotClass } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const router = useRouter()
const loading = ref(false)
const classList = ref<any[]>([])
const searchName = ref('')
const selectedRows = ref<any[]>([])
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)

const filteredList = computed(() => {
  if (!searchName.value) return classList.value
  const keyword = searchName.value.toLowerCase()
  return classList.value.filter((d) => d.name?.toLowerCase().includes(keyword))
})

async function fetchClasses() {
  loading.value = true
  try {
    const res: any = await getVolumeSnapshotClassList()
    classList.value = res.data || []
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load VolumeSnapshotClasses')
  } finally { loading.value = false }
}

function handleSelectionChange(rows: any[]) { selectedRows.value = rows }

async function handleViewYaml(row: any) {
  yamlDialogVisible.value = true; yamlLoading.value = true; yamlContent.value = ''
  try {
    const res: any = await getVolumeSnapshotClassYaml({ name: row.name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) { ElMessage.error(e?.message || 'Failed to load YAML'); yamlDialogVisible.value = false }
  finally { yamlLoading.value = false }
}

function handleDetail(row: any) { router.push(`/storage/volumesnapshotclasses/${row.name}`) }

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(`Delete VolumeSnapshotClass "${row.name}"?`, 'Confirm', { type: 'warning' })
    await deleteVolumeSnapshotClass({ name: row.name })
    ElMessage.success('Deleted'); fetchClasses()
  } catch { /* cancelled */ }
}

async function handleBatchDelete() {
  if (!selectedRows.value.length) return
  try {
    await ElMessageBox.confirm(`Delete ${selectedRows.value.length} selected VolumeSnapshotClass(es)?`, 'Confirm', { type: 'warning' })
    let count = 0
    for (const row of selectedRows.value) {
      try { await deleteVolumeSnapshotClass({ name: row.name }); count++ } catch { /* continue */ }
    }
    ElMessage.success(`Deleted ${count} VolumeSnapshotClass(es)`); fetchClasses()
  } catch { /* cancelled */ }
}

onMounted(fetchClasses)
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <el-input v-model="searchName" placeholder="Search by name" style="width: 220px;" clearable>
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-button type="primary" @click="fetchClasses"><el-icon><Refresh /></el-icon> Refresh</el-button>
        <el-button type="success" @click="router.push('/storage/volumesnapshotclasses/create')"><el-icon><Plus /></el-icon> Create</el-button>
        <el-button type="danger" :disabled="!selectedRows.length" @click="handleBatchDelete"><el-icon><Delete /></el-icon> Delete ({{ selectedRows.length }})</el-button>
      </div>
    </el-card>
    <el-card shadow="never" class="table-card">
      <el-table :data="filteredList" v-loading="loading" stripe @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="45" />
        <el-table-column prop="name" label="Name" min-width="200" show-overflow-tooltip>
          <template #default="{ row }"><el-button link type="primary" @click="handleDetail(row)">{{ row.name }}</el-button></template>
        </el-table-column>
        <el-table-column prop="driver" label="Driver" min-width="250" show-overflow-tooltip />
        <el-table-column prop="deletionPolicy" label="Deletion Policy" width="160" />
        <el-table-column prop="age" label="Age" width="120" />
        <el-table-column label="Actions" width="180" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">Delete</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <el-dialog v-model="yamlDialogVisible" title="VolumeSnapshotClass YAML" width="70%" top="5vh" destroy-on-close>
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
