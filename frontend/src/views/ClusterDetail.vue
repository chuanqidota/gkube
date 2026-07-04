<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Delete } from '@element-plus/icons-vue'
import { getCluster, deleteCluster, checkCluster } from '@/api/cluster'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const loading = ref(false)
const checking = ref(false)
const cluster = ref<any>(null)

async function fetchDetail() {
  const id = route.params.id as string
  if (!id) return
  loading.value = true
  try {
    const res: any = await getCluster(id)
    cluster.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function handleCheck() {
  if (!cluster.value) return
  checking.value = true
  try {
    const res: any = await checkCluster(cluster.value.id)
    const info = res.data
    if (info.status === 'online') {
      ElMessage.success(t('cluster.connectedSuccess', { version: info.clusterVersion, nodeCount: info.nodeCount, responseTimeMs: info.responseTimeMs }))
      fetchDetail()
    } else {
      ElMessage.warning(info.message || t('cluster.connectionFailed'))
    }
  } catch (e: any) {
    ElMessage.error(e?.message || t('cluster.checkFailed'))
  } finally {
    checking.value = false
  }
}

async function handleDelete() {
  if (!cluster.value) return
  try {
    await ElMessageBox.confirm(
      t('cluster.deleteClusterConfirm', { name: cluster.value.displayName || cluster.value.clusterName }),
      t('common.confirmDelete'),
      { type: 'warning' }
    )
    await deleteCluster(cluster.value.id)
    ElMessage.success(t('cluster.deleted'))
    router.push('/clusters')
  } catch {
    // cancelled
  }
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

onMounted(fetchDetail)
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <h3 style="margin: 0;">{{ t('cluster.clusterDetail') }}</h3>
        <div class="filter-right">
          <el-button @click="fetchDetail"><el-icon><Refresh /></el-icon> {{ t('common.refresh') }}</el-button>
          <el-button type="success" :loading="checking" @click="handleCheck"><el-icon><Refresh /></el-icon> {{ t('cluster.check') }}</el-button>
          <el-button type="danger" @click="handleDelete"><el-icon><Delete /></el-icon> {{ t('common.delete') }}</el-button>
          <el-button @click="router.push('/clusters')">{{ t('common.backToList') }}</el-button>
        </div>
      </div>
    </el-card>

    <el-row :gutter="16" v-loading="loading">
      <el-col :span="16">
        <el-card shadow="never">
          <template #header>
            <h4 style="margin: 0;">{{ t('cluster.basicInfo') }}</h4>
          </template>
          <el-descriptions :column="2" border v-if="cluster">
            <el-descriptions-item :label="t('cluster.name')">{{ cluster.clusterName }}</el-descriptions-item>
            <el-descriptions-item :label="t('cluster.displayName')">{{ cluster.displayName || '-' }}</el-descriptions-item>
            <el-descriptions-item :label="t('cluster.status')">
              <el-tag :type="statusType(cluster.status)">{{ statusText(cluster.status) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="t('cluster.version')">{{ cluster.clusterVersion || '-' }}</el-descriptions-item>
            <el-descriptions-item :label="t('cluster.nodes')">{{ cluster.nodeCount ?? '-' }}</el-descriptions-item>
            <el-descriptions-item :label="t('cluster.lastCheck')">{{ cluster.lastHealthCheck || '-' }}</el-descriptions-item>
            <el-descriptions-item :label="t('cluster.description')" :span="2">{{ cluster.description || '-' }}</el-descriptions-item>
            <el-descriptions-item :label="t('config.creationTime')" :span="2">{{ cluster.createdAt || '-' }}</el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-card shadow="never" style="margin-top: 16px;" v-if="cluster?.labels && Object.keys(cluster.labels).length > 0">
          <template #header>
            <h4 style="margin: 0;">{{ t('cluster.labels') }}</h4>
          </template>
          <el-tag
            v-for="(val, key) in cluster.labels"
            :key="key"
            style="margin-right: 8px; margin-bottom: 8px;"
          >
            {{ key }}={{ val }}
          </el-tag>
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card shadow="never">
          <template #header>
            <h4 style="margin: 0;">{{ t('cluster.quickActions') }}</h4>
          </template>
          <div class="quick-actions">
            <el-button type="primary" style="width: 100%;" @click="router.push('/')">
              {{ t('cluster.enterCluster') }}
            </el-button>
            <el-button style="width: 100%;" @click="handleCheck">
              {{ t('cluster.check') }}
            </el-button>
            <el-button style="width: 100%;" @click="router.push('/clusters')">
              {{ t('common.backToList') }}
            </el-button>
          </div>
        </el-card>

        <el-card shadow="never" style="margin-top: 16px;">
          <template #header>
            <h4 style="margin: 0;">{{ t('cluster.clusterInfo') }}</h4>
          </template>
          <div class="info-list">
            <div class="info-item">
              <span class="info-label">{{ t('cluster.clusterId') }}</span>
              <span class="info-value">{{ cluster?.id }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">{{ t('cluster.name') }}</span>
              <span class="info-value">{{ cluster?.clusterName }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">{{ t('cluster.status') }}</span>
              <span class="info-value">
                <el-tag :type="statusType(cluster?.status)" size="small">{{ statusText(cluster?.status) }}</el-tag>
              </span>
            </div>
            <div class="info-item">
              <span class="info-label">{{ t('cluster.version') }}</span>
              <span class="info-value">{{ cluster?.clusterVersion || '-' }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">{{ t('cluster.nodes') }}</span>
              <span class="info-value">{{ cluster?.nodeCount ?? '-' }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; justify-content: space-between; align-items: center; }
.filter-right { display: flex; align-items: center; gap: 8px; }
.quick-actions { display: flex; flex-direction: column; gap: 8px; }
.info-list { display: flex; flex-direction: column; gap: 12px; }
.info-item { display: flex; justify-content: space-between; align-items: center; }
.info-label { color: var(--gk-color-text-secondary); }
.info-value { color: var(--gk-color-text-primary); font-weight: 500; }
</style>
