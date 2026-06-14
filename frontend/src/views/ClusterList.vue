<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Delete, View, CircleCheck } from '@element-plus/icons-vue'
import { getClusterList, deleteCluster, checkCluster } from '@/api/cluster'

const { t } = useI18n()
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
    if (info.status === 'online') {
      ElMessage.success(
        `Connected (v${info.clusterVersion}, ${info.nodeCount} nodes, ${info.responseTimeMs}ms)`
      )
    } else {
      ElMessage.warning(info.message || 'Connection failed')
    }
    fetchClusters()
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
  if (status === 'online' || status === 'connected' || status === 'healthy') return 'success'
  if (status === 'offline' || status === 'disconnected' || status === 'unhealthy') return 'danger'
  return 'info'
}

function statusText(status: string) {
  if (status === 'online' || status === 'connected') return t('cluster.online')
  if (status === 'offline' || status === 'disconnected') return t('cluster.offline')
  return status || t('common.unknown')
}

onMounted(fetchClusters)
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <h3 style="margin: 0;">{{ t('cluster.clusterManagement') }}</h3>
        <div class="filter-right">
          <el-button @click="fetchClusters"><el-icon><Refresh /></el-icon> {{ t('common.refresh') }}</el-button>
          <el-button type="primary" @click="router.push('/clusters/create')"><el-icon><Plus /></el-icon> {{ t('cluster.add') }}</el-button>
        </div>
      </div>
    </el-card>

    <el-row :gutter="16" v-if="clusterList.length > 0">
      <el-col :span="8" v-for="cluster in clusterList" :key="cluster.id" style="margin-bottom: 16px;">
        <el-card shadow="hover" class="cluster-card">
          <template #header>
            <div class="cluster-header">
              <div class="cluster-info">
                <h4 style="margin: 0;">{{ cluster.displayName || cluster.clusterName }}</h4>
                <el-tag :type="statusType(cluster.status)" size="small">{{ statusText(cluster.status) }}</el-tag>
              </div>
              <el-button type="primary" link @click="handleDetail(cluster)"><el-icon><View /></el-icon></el-button>
            </div>
          </template>
          <div class="cluster-body">
            <div class="cluster-detail">
              <span class="label">{{ t('cluster.name') }}:</span>
              <span class="value">{{ cluster.clusterName }}</span>
            </div>
            <div class="cluster-detail">
              <span class="label">{{ t('cluster.version') }}:</span>
              <span class="value">{{ cluster.clusterVersion || '-' }}</span>
            </div>
            <div class="cluster-detail">
              <span class="label">{{ t('cluster.nodes') }}:</span>
              <span class="value">{{ cluster.nodeCount || 0 }}</span>
            </div>
            <div class="cluster-detail" v-if="cluster.description">
              <span class="label">{{ t('cluster.description') }}:</span>
              <span class="value">{{ cluster.description }}</span>
            </div>
          </div>
          <div class="cluster-footer">
            <el-button size="small" @click="handleCheck(cluster)"><el-icon><CircleCheck /></el-icon> {{ t('cluster.checkConnection') }}</el-button>
            <el-button size="small" @click="handleDetail(cluster)"><el-icon><View /></el-icon> {{ t('common.detail') }}</el-button>
            <el-button size="small" type="danger" @click="handleDelete(cluster)"><el-icon><Delete /></el-icon> {{ t('common.delete') }}</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-empty v-if="!loading && clusterList.length === 0" :description="t('cluster.noClusters')">
      <el-button type="primary" @click="router.push('/clusters/create')"><el-icon><Plus /></el-icon> {{ t('cluster.add') }}</el-button>
    </el-empty>

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

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; justify-content: space-between; align-items: center; }
.filter-right { display: flex; align-items: center; gap: 8px; }
.cluster-card { height: 100%; }
.cluster-header { display: flex; justify-content: space-between; align-items: center; }
.cluster-info { display: flex; align-items: center; gap: 8px; }
.cluster-body { margin-bottom: 12px; }
.cluster-detail { display: flex; margin-bottom: 8px; }
.cluster-detail .label { color: #909399; width: 70px; flex-shrink: 0; }
.cluster-detail .value { color: #303133; }
.cluster-footer { display: flex; justify-content: flex-end; gap: 8px; border-top: 1px solid #ebeef5; padding-top: 12px; }
</style>
