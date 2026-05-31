<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getCluster } from '@/api/cluster'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const cluster = ref<any>(null)

async function fetchDetail() {
  const id = route.params.id as string
  if (!id) return
  loading.value = true
  try {
    const res: any = await getCluster(id)
    cluster.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load cluster detail')
  } finally {
    loading.value = false
  }
}

function statusType(status: string) {
  if (status === 'connected' || status === 'healthy') return 'success'
  if (status === 'disconnected' || status === 'unhealthy') return 'danger'
  return 'info'
}

onMounted(fetchDetail)
</script>

<template>
  <div v-loading="loading">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="margin: 0;">Cluster Detail</h2>
      <el-button @click="router.push('/clusters')">Back to List</el-button>
    </div>

    <template v-if="cluster">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="Cluster Name">{{ cluster.clusterName }}</el-descriptions-item>
        <el-descriptions-item label="Display Name">{{ cluster.displayName }}</el-descriptions-item>
        <el-descriptions-item label="Status">
          <el-tag :type="statusType(cluster.status)" size="small">
            {{ cluster.status || 'unknown' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="Version">{{ cluster.version || '-' }}</el-descriptions-item>
        <el-descriptions-item label="Node Count">{{ cluster.nodeCount ?? '-' }}</el-descriptions-item>
        <el-descriptions-item label="Response Time">
          {{ cluster.responseTimeMs != null ? cluster.responseTimeMs + 'ms' : '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="Description" :span="2">
          {{ cluster.description || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="Created At" :span="2">
          {{ cluster.createdAt || '-' }}
        </el-descriptions-item>
      </el-descriptions>

      <div v-if="cluster.labels && Object.keys(cluster.labels).length > 0" style="margin-top: 16px;">
        <h3>Labels</h3>
        <el-tag
          v-for="(val, key) in cluster.labels"
          :key="key"
          style="margin-right: 8px; margin-bottom: 8px;"
        >
          {{ key }}={{ val }}
        </el-tag>
      </div>
    </template>
  </div>
</template>
