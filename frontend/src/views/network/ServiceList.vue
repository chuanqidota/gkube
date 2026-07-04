<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Plus, Delete, Search } from '@element-plus/icons-vue'
import { getServiceList, getServiceYaml, updateService, deleteService, getNamespaceList, extractNamespaceNames, transformServices } from '@/api/resource'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import YamlEditor from '@/components/YamlEditor.vue'

const router = useRouter()
const loading = ref(false)
const serviceList = ref<any[]>([])
const namespaceList = ref<string[]>([])
const selectedNamespace = ref('')
const searchName = ref('')
const selectedRows = ref<any[]>([])
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)
const yamlEditorRef = ref<InstanceType<typeof YamlEditor>>()
const currentYamlRow = ref<any>(null)

const filteredList = computed(() => {
  if (!searchName.value) return serviceList.value
  const keyword = searchName.value.toLowerCase()
  return serviceList.value.filter((s) => s.name?.toLowerCase().includes(keyword))
})

async function fetchNamespaces() {
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = extractNamespaceNames(res.data)
  } catch {
    // ignore
  }
}

async function fetchServices() {
  loading.value = true
  try {
    const params: any = {}
    if (selectedNamespace.value) params.namespace = selectedNamespace.value
    const res: any = await getServiceList(params)
    const items = res.data?.items || res.data || []
    serviceList.value = transformServices(items)
  } catch {
    // Silently handle — resource may not exist in cluster
  } finally {
    loading.value = false
  }
}

function handleNamespaceChange() {
  fetchServices()
}

function handleSelectionChange(rows: any[]) {
  selectedRows.value = rows
}

async function handleViewYaml(row: any) {
  currentYamlRow.value = row
  yamlLoading.value = true
  yamlDialogVisible.value = true
  try {
    const res: any = await getServiceYaml({
      name: row.name,
      namespace: row.namespace,
    })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load YAML')
    yamlDialogVisible.value = false
  } finally {
    yamlLoading.value = false
  }
}

async function handleSaveYaml(content: string) {
  if (!currentYamlRow.value) return
  try {
    await updateService({ namespace: currentYamlRow.value.namespace, name: currentYamlRow.value.name, yaml: content })
    ElMessage.success('YAML saved successfully')
    yamlDialogVisible.value = false
    fetchServices()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to save YAML')
    yamlEditorRef.value?.resetSaving()
  }
}

function handleDetail(row: any) {
  router.push(`/services/${row.namespace}/${row.name}`)
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(
      `Delete Service "${row.name}" in namespace "${row.namespace}"?`,
      'Confirm',
      { type: 'warning' }
    )
    await deleteService({ name: row.name, namespace: row.namespace })
    ElMessage.success('Deleted')
    fetchServices()
  } catch {
    // cancelled
  }
}

async function handleBatchDelete() {
  if (!selectedRows.value.length) return
  try {
    await ElMessageBox.confirm(
      `Delete ${selectedRows.value.length} selected service(s)?`,
      'Confirm',
      { type: 'warning' }
    )
    let successCount = 0
    for (const row of selectedRows.value) {
      try {
        await deleteService({ name: row.name, namespace: row.namespace })
        successCount++
      } catch {
        // continue
      }
    }
    ElMessage.success(`Deleted ${successCount} service(s)`)
    fetchServices()
  } catch {
    // cancelled
  }
}

const { isRunning, countdown, toggle, refresh } = useAutoRefresh(fetchServices)

onMounted(() => {
  fetchNamespaces()
  fetchServices()
})
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <el-input
          v-model="searchName"
          placeholder="Search by name"
          style="width: 220px;"
          clearable
        >
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select
          v-model="selectedNamespace"
          placeholder="All Namespaces"
          clearable
          style="width: 180px;"
          @change="handleNamespaceChange"
        >
          <el-option v-for="ns in namespaceList" :key="ns" :label="ns" :value="ns" />
        </el-select>
        <el-button type="primary" @click="refresh">
          <el-icon><Refresh /></el-icon> Refresh
        </el-button>
        <el-button :type="isRunning ? 'success' : 'info'" @click="toggle">
          {{ isRunning ? `Auto (${countdown}s)` : 'Manual' }}
        </el-button>
        <el-button type="success" @click="router.push('/services/create')">
          <el-icon><Plus /></el-icon> Create
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
        <el-table-column prop="name" label="Name" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <el-button link type="primary" @click="handleDetail(row)">{{ row.name }}</el-button>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="Namespace" width="140" />
        <el-table-column prop="type" label="Type" width="130" />
        <el-table-column prop="cluster_ip" label="Cluster IP" width="150" />
        <el-table-column prop="ports" label="Ports" min-width="160" show-overflow-tooltip />
        <el-table-column prop="age" label="Age" width="120" />
        <el-table-column label="Actions" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">Delete</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="yamlDialogVisible" title="Service YAML" width="70%" top="5vh" destroy-on-close>
      <div v-loading="yamlLoading">
        <YamlEditor ref="yamlEditorRef" v-model="yamlContent" height="500px" :read-only="true" :saveable="true" auto-format @save="handleSaveYaml" />
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
