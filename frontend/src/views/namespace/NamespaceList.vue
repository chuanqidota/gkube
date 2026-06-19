<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Plus, Search } from '@element-plus/icons-vue'
import { getNamespaceList, createNamespace, extractNamespaceNames } from '@/api/resource'

const loading = ref(false)
const namespaceList = ref<any[]>([])
const searchName = ref('')
const createDialogVisible = ref(false)
const newNamespaceName = ref('')
const newNamespaceLabels = ref([{ key: '', value: '' }])
const creating = ref(false)

const filteredList = computed(() => {
  if (!searchName.value) return namespaceList.value
  const keyword = searchName.value.toLowerCase()
  return namespaceList.value.filter((ns) => ns.name?.toLowerCase().includes(keyword))
})

import { computed } from 'vue'

async function fetchNamespaces() {
  loading.value = true
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = extractNamespaceNames(res.data)
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load namespaces')
  } finally { loading.value = false }
}

function statusType(status: string) {
  if (status === 'Active') return 'success'
  if (status === 'Terminating') return 'warning'
  return 'info'
}

function addLabel() { newNamespaceLabels.value.push({ key: '', value: '' }) }
function removeLabel(i: number) { newNamespaceLabels.value.splice(i, 1) }

async function handleCreate() {
  if (!newNamespaceName.value.trim()) {
    ElMessage.warning('Please enter a namespace name')
    return
  }
  creating.value = true
  try {
    await createNamespace({ name: newNamespaceName.value.trim() })
    ElMessage.success('Namespace created')
    createDialogVisible.value = false
    newNamespaceName.value = ''
    newNamespaceLabels.value = [{ key: '', value: '' }]
    fetchNamespaces()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to create namespace')
  } finally { creating.value = false }
}

onMounted(fetchNamespaces)
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <el-input v-model="searchName" placeholder="Search by name" style="width: 220px;" clearable>
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-button type="primary" @click="fetchNamespaces"><el-icon><Refresh /></el-icon> Refresh</el-button>
        <el-button type="success" @click="createDialogVisible = true"><el-icon><Plus /></el-icon> Create</el-button>
      </div>
    </el-card>

    <el-card shadow="never" class="table-card">
      <el-table :data="filteredList" v-loading="loading" stripe>
        <el-table-column prop="name" label="Name" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <el-button link type="primary" @click="$router.push(`/namespaces/${row.name}`)">{{ row.name }}</el-button>
          </template>
        </el-table-column>
        <el-table-column label="Status" width="120">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ row.status || 'Unknown' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Labels" min-width="250" show-overflow-tooltip>
          <template #default="{ row }">
            <el-tag v-for="(v, k) in (row.labels || {})" :key="k" size="small" style="margin-right: 4px; margin-bottom: 2px;">{{ k }}={{ v }}</el-tag>
            <span v-if="!row.labels || Object.keys(row.labels).length === 0" style="color: #909399;">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="age" label="Age" width="180" />
      </el-table>
    </el-card>

    <el-dialog v-model="createDialogVisible" title="Create Namespace" width="500px">
      <el-form @submit.prevent="handleCreate" label-width="100px">
        <el-form-item label="Name" required>
          <el-input v-model="newNamespaceName" placeholder="Enter namespace name" />
        </el-form-item>
        <el-form-item label="Labels">
          <div style="width: 100%;">
            <div v-for="(label, i) in newNamespaceLabels" :key="i" style="display: flex; gap: 8px; margin-bottom: 8px;">
              <el-input v-model="label.key" placeholder="Key" style="flex: 1;" />
              <el-input v-model="label.value" placeholder="Value" style="flex: 1;" />
              <el-button type="danger" circle size="small" @click="removeLabel(i)">X</el-button>
            </div>
            <el-button size="small" @click="addLabel">+ Add Label</el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">Cancel</el-button>
        <el-button type="primary" @click="handleCreate" :loading="creating">Create</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; gap: 12px; align-items: center; flex-wrap: wrap; }
.table-card { border-radius: 8px; }
</style>
