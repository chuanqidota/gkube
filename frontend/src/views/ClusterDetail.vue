<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Delete, Edit } from '@element-plus/icons-vue'
import { getCluster, deleteCluster, checkCluster } from '@/api/cluster'

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
    ElMessage.error(e?.message || '加载集群详情失败')
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
    if (info.connected) {
      ElMessage.success(`连接成功 (v${info.version}, ${info.nodeCount} 节点, ${info.responseTimeMs}ms)`)
      fetchDetail()
    } else {
      ElMessage.warning(info.message || '连接失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '检测失败')
  } finally {
    checking.value = false
  }
}

async function handleDelete() {
  if (!cluster.value) return
  try {
    await ElMessageBox.confirm(
      `确定删除集群 "${cluster.value.displayName || cluster.value.clusterName}" 吗？`,
      '确认删除',
      { type: 'warning' }
    )
    await deleteCluster(cluster.value.id)
    ElMessage.success('删除成功')
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
  if (status === 'online' || status === 'connected') return '在线'
  if (status === 'offline' || status === 'disconnected') return '离线'
  return status || '未知'
}

onMounted(fetchDetail)
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <h3 style="margin: 0;">集群详情</h3>
        <div class="filter-right">
          <el-button @click="fetchDetail"><el-icon><Refresh /></el-icon> 刷新</el-button>
          <el-button type="success" :loading="checking" @click="handleCheck"><el-icon><Refresh /></el-icon> 检测连通性</el-button>
          <el-button type="danger" @click="handleDelete"><el-icon><Delete /></el-icon> 删除</el-button>
          <el-button @click="router.push('/clusters')">返回列表</el-button>
        </div>
      </div>
    </el-card>

    <el-row :gutter="16" v-loading="loading">
      <el-col :span="16">
        <el-card shadow="never">
          <template #header>
            <h4 style="margin: 0;">基本信息</h4>
          </template>
          <el-descriptions :column="2" border v-if="cluster">
            <el-descriptions-item label="集群名称">{{ cluster.clusterName }}</el-descriptions-item>
            <el-descriptions-item label="显示名称">{{ cluster.displayName || '-' }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="statusType(cluster.status)">{{ statusText(cluster.status) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="版本">{{ cluster.clusterVersion || '-' }}</el-descriptions-item>
            <el-descriptions-item label="节点数">{{ cluster.nodeCount ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="最后检查">{{ cluster.lastHealthCheck || '-' }}</el-descriptions-item>
            <el-descriptions-item label="描述" :span="2">{{ cluster.description || '-' }}</el-descriptions-item>
            <el-descriptions-item label="创建时间" :span="2">{{ cluster.createdAt || '-' }}</el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-card shadow="never" style="margin-top: 16px;" v-if="cluster?.labels && Object.keys(cluster.labels).length > 0">
          <template #header>
            <h4 style="margin: 0;">标签</h4>
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
            <h4 style="margin: 0;">快速操作</h4>
          </template>
          <div class="quick-actions">
            <el-button type="primary" style="width: 100%;" @click="router.push('/')">
              进入集群
            </el-button>
            <el-button style="width: 100%;" @click="handleCheck">
              检测连通性
            </el-button>
            <el-button style="width: 100%;" @click="router.push('/clusters')">
              返回列表
            </el-button>
          </div>
        </el-card>

        <el-card shadow="never" style="margin-top: 16px;">
          <template #header>
            <h4 style="margin: 0;">集群信息</h4>
          </template>
          <div class="info-list">
            <div class="info-item">
              <span class="info-label">集群 ID</span>
              <span class="info-value">{{ cluster?.id }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">集群名称</span>
              <span class="info-value">{{ cluster?.clusterName }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">状态</span>
              <span class="info-value">
                <el-tag :type="statusType(cluster?.status)" size="small">{{ statusText(cluster?.status) }}</el-tag>
              </span>
            </div>
            <div class="info-item">
              <span class="info-label">版本</span>
              <span class="info-value">{{ cluster?.clusterVersion || '-' }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">节点数</span>
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
.info-label { color: #909399; }
.info-value { color: #303133; font-weight: 500; }
</style>
