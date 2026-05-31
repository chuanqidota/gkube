<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getClusterList, deleteCluster, checkCluster } from '@/api/cluster'

const router = useRouter()
const loading = ref(false)
const clusterList = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const size = ref(10)

async function fetchClusters() {
  loading.value = true
  try {
    const res: any = await getClusterList({ page: page.value, size: size.value })
    clusterList.value = res.data.items || []
    total.value = res.data.total || 0
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load clusters')
  } finally {
    loading.value = false
  }
}

function handleDetail(row: any) {
  router.push(`/clusters/${row.id}`)
}

async function handleCheck(row: any) {
  try {
    const res: any = await checkCluster(row.id)
    const info = res.data
    if (info.connected) {
      ElMessage.success(
        `Connected (v${info.version}, ${info.nodeCount} nodes, ${info.responseTimeMs}ms)`
      )
    } else {
      ElMessage.warning(info.message || 'Connection failed')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || 'Check failed')
  }
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(
      `Delete cluster "${row.displayName || row.clusterName}"?`,
      'Confirm',
      { type: 'warning' }
    )
    await deleteCluster(row.id)
    ElMessage.success('Deleted')
    fetchClusters()
  } catch {
    // cancelled
  }
}

function handlePageChange(newPage: number) {
  page.value = newPage
  fetchClusters()
}

function statusType(status: string) {
  if (status === 'connected' || status === 'healthy') return 'success'
  if (status === 'disconnected' || status === 'unhealthy') return 'danger'
  return 'info'
}

onMounted(fetchClusters)
</script>

<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="margin: 0;">Clusters</h2>
      <el-button type="primary" @click="router.push('/clusters/create')">Add Cluster</el-button>
    </div>

    <el-table :data="clusterList" v-loading="loading" stripe style="width: 100%">
      <el-table-column prop="clusterName" label="Cluster Name" min-width="140" />
      <el-table-column prop="displayName" label="Display Name" min-width="140" />
      <el-table-column prop="description" label="Description" min-width="200" show-overflow-tooltip />
      <el-table-column label="Status" width="120">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)" size="small">
            {{ row.status || 'unknown' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="version" label="Version" width="120" />
      <el-table-column prop="nodeCount" label="Nodes" width="80" />
      <el-table-column label="Actions" width="240" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="handleDetail(row)">Detail</el-button>
          <el-button size="small" type="success" @click="handleCheck(row)">Check</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">Delete</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div style="display: flex; justify-content: flex-end; margin-top: 16px;">
      <el-pagination
        v-if="total > size"
        :current-page="page"
        :page-size="size"
        :total="total"
        layout="prev, pager, next"
        @current-change="handlePageChange"
      />
    </div>
  </div>
</template>
