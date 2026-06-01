<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getNamespaceList, createNamespace } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'

const clusterStore = useClusterStore()
const loading = ref(false)
const namespaceList = ref<any[]>([])
const createDialogVisible = ref(false)
const newNamespaceName = ref('')
const creating = ref(false)

async function fetchNamespaces() {
  loading.value = true
  try {
    const clusterId = clusterStore.currentCluster?.id
    const res: any = await getNamespaceList({ cluster_id: clusterId })
    namespaceList.value = res.data || []
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load namespaces')
  } finally {
    loading.value = false
  }
}

function statusType(status: string) {
  if (status === 'Active') return 'success'
  if (status === 'Terminating') return 'warning'
  return 'info'
}

async function handleCreate() {
  if (!newNamespaceName.value.trim()) {
    ElMessage.warning('Please enter a namespace name')
    return
  }
  creating.value = true
  try {
    const clusterId = clusterStore.currentCluster?.id
    await createNamespace({ name: newNamespaceName.value.trim(), cluster_id: clusterId })
    ElMessage.success('Namespace created')
    createDialogVisible.value = false
    newNamespaceName.value = ''
    fetchNamespaces()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to create namespace')
  } finally {
    creating.value = false
  }
}

onMounted(fetchNamespaces)
</script>

<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="margin: 0;">Namespaces</h2>
      <div>
        <el-button type="primary" @click="createDialogVisible = true">Create Namespace</el-button>
        <el-button @click="fetchNamespaces" :loading="loading">Refresh</el-button>
      </div>
    </div>

    <el-table :data="namespaceList" v-loading="loading" stripe style="width: 100%">
      <el-table-column prop="name" label="Name" min-width="200" />
      <el-table-column label="Status" width="120">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)" size="small">{{ row.status || 'Unknown' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="age" label="Age" min-width="120" />
    </el-table>

    <el-dialog v-model="createDialogVisible" title="Create Namespace" width="400px">
      <el-form @submit.prevent="handleCreate">
        <el-form-item label="Name">
          <el-input v-model="newNamespaceName" placeholder="Enter namespace name" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">Cancel</el-button>
        <el-button type="primary" @click="handleCreate" :loading="creating">Create</el-button>
      </template>
    </el-dialog>
  </div>
</template>
