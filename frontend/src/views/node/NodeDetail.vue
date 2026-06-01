<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getNodeList, getNodePods } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'

const route = useRoute()
const router = useRouter()
const clusterStore = useClusterStore()
const loading = ref(false)
const node = ref<any>(null)
const pods = ref<any[]>([])
const podsLoading = ref(false)

async function fetchDetail() {
  const name = route.params.name as string
  if (!name) return
  loading.value = true
  try {
    const clusterId = clusterStore.currentCluster?.id
    const res: any = await getNodeList({ cluster_id: clusterId })
    const nodes = res.data || []
    node.value = nodes.find((n: any) => n.name === name) || null
    if (node.value) {
      fetchPods(name)
    }
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load node detail')
  } finally {
    loading.value = false
  }
}

async function fetchPods(name: string) {
  podsLoading.value = true
  try {
    const clusterId = clusterStore.currentCluster?.id
    const res: any = await getNodePods({ name, cluster_id: clusterId })
    pods.value = res.data || []
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load pods')
  } finally {
    podsLoading.value = false
  }
}

function statusType(status: string) {
  if (status === 'Ready') return 'success'
  if (status === 'NotReady') return 'danger'
  return 'warning'
}

function podStatusType(status: string) {
  if (status === 'Running') return 'success'
  if (status === 'Succeeded') return 'info'
  if (status === 'Failed' || status === 'Error') return 'danger'
  return 'warning'
}

onMounted(fetchDetail)
</script>

<template>
  <div v-loading="loading">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="margin: 0;">Node Detail</h2>
      <el-button @click="router.push('/nodes')">Back to List</el-button>
    </div>

    <template v-if="node">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="Name">{{ node.name }}</el-descriptions-item>
        <el-descriptions-item label="Status">
          <el-tag :type="statusType(node.status)" size="small">{{ node.status || 'Unknown' }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="Roles">{{ node.roles || '-' }}</el-descriptions-item>
        <el-descriptions-item label="Version">{{ node.version || '-' }}</el-descriptions-item>
        <el-descriptions-item label="OS">{{ node.os || '-' }}</el-descriptions-item>
        <el-descriptions-item label="Kernel">{{ node.kernel || '-' }}</el-descriptions-item>
        <el-descriptions-item label="Container Runtime">{{ node.container_runtime || '-' }}</el-descriptions-item>
        <el-descriptions-item label="Internal IP">{{ node.internal_ip || '-' }}</el-descriptions-item>
        <el-descriptions-item label="Age">{{ node.age || '-' }}</el-descriptions-item>
        <el-descriptions-item label="Unschedulable">
          <el-tag :type="node.unschedulable || node.cordon ? 'danger' : 'success'" size="small">
            {{ node.unschedulable || node.cordon ? 'Yes' : 'No' }}
          </el-tag>
        </el-descriptions-item>
      </el-descriptions>

      <div v-if="node.labels && Object.keys(node.labels).length > 0" style="margin-top: 24px;">
        <h3 style="margin-bottom: 12px;">Labels</h3>
        <el-tag
          v-for="(val, key) in node.labels"
          :key="key"
          style="margin-right: 8px; margin-bottom: 8px;"
        >
          {{ key }}={{ val }}
        </el-tag>
      </div>

      <div v-if="node.taints && node.taints.length > 0" style="margin-top: 24px;">
        <h3 style="margin-bottom: 12px;">Taints</h3>
        <el-table :data="node.taints" stripe style="width: 100%;">
          <el-table-column prop="key" label="Key" min-width="200" />
          <el-table-column prop="value" label="Value" min-width="120" />
          <el-table-column prop="effect" label="Effect" min-width="150" />
        </el-table>
      </div>

      <div style="margin-top: 24px;">
        <h3 style="margin-bottom: 12px;">Pods on this Node</h3>
        <el-table :data="pods" v-loading="podsLoading" stripe style="width: 100%;">
          <el-table-column prop="name" label="Name" min-width="200" show-overflow-tooltip />
          <el-table-column prop="namespace" label="Namespace" min-width="120" />
          <el-table-column label="Status" width="100">
            <template #default="{ row }">
              <el-tag :type="podStatusType(row.status)" size="small">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="ip" label="IP" min-width="130" />
          <el-table-column prop="restarts" label="Restarts" width="90" />
          <el-table-column prop="age" label="Age" min-width="100" />
        </el-table>
      </div>
    </template>
  </div>
</template>
