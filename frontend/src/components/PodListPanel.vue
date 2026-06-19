<script setup lang="ts">
import { ElTag, ElButton } from 'element-plus'

interface Pod {
  metadata: {
    name: string
    namespace: string
    creationTimestamp: string
  }
  status: {
    phase: string
    containerStatuses?: Array<{
      restartCount: number
    }>
  }
  spec: {
    nodeName?: string
  }
}

const props = defineProps<{
  pods: Pod[]
  loading: boolean
  replicasetName: string
}>()

const emit = defineEmits<{
  logs: [pod: Pod]
  exec: [pod: Pod]
  delete: [pod: Pod]
}>()

const getStatusType = (phase: string): 'success' | 'warning' | 'danger' | 'info' => {
  switch (phase) {
    case 'Running':
      return 'success'
    case 'Pending':
      return 'warning'
    case 'Failed':
      return 'danger'
    case 'Succeeded':
      return 'info'
    default:
      return 'info'
  }
}

const getRestarts = (pod: Pod): number => {
  return pod.status.containerStatuses?.reduce((sum, cs) => sum + cs.restartCount, 0) || 0
}

const formatAge = (timestamp: string): string => {
  const now = new Date()
  const created = new Date(timestamp)
  const diffMs = now.getTime() - created.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMins / 60)
  const diffDays = Math.floor(diffHours / 24)

  if (diffDays > 0) return `${diffDays}d`
  if (diffHours > 0) return `${diffHours}h`
  return `${diffMins}m`
}

const handleLogs = (pod: Pod) => {
  emit('logs', pod)
}

const handleExec = (pod: Pod) => {
  emit('exec', pod)
}

const handleDelete = (pod: Pod) => {
  emit('delete', pod)
}
</script>

<template>
  <div class="pod-list-panel" v-loading="loading">
    <div class="panel-header">
      <span class="title">Pods ({{ pods.length }})</span>
    </div>
    <div v-if="pods.length === 0" class="empty-state">
      No pods found
    </div>
    <el-table v-else :data="pods" style="width: 100%">
      <el-table-column label="Name" min-width="200">
        <template #default="{ row }">
          <router-link
            :to="{ name: 'PodDetail', params: { namespace: row.metadata.namespace, name: row.metadata.name } }"
            class="pod-link"
          >
            {{ row.metadata.name }}
          </router-link>
        </template>
      </el-table-column>
      <el-table-column label="Status" width="100">
        <template #default="{ row }">
          <el-tag :type="getStatusType(row.status.phase)" size="small">
            {{ row.status.phase }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="Restarts" width="80">
        <template #default="{ row }">
          <span :class="{ warning: getRestarts(row) > 0 }">
            {{ getRestarts(row) }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="Age" width="80">
        <template #default="{ row }">
          {{ formatAge(row.metadata.creationTimestamp) }}
        </template>
      </el-table-column>
      <el-table-column label="Node" width="120">
        <template #default="{ row }">
          {{ row.spec.nodeName || '-' }}
        </template>
      </el-table-column>
      <el-table-column label="Actions" width="180" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="handleLogs(row)">Logs</el-button>
          <el-button size="small" @click="handleExec(row)">Exec</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">Delete</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<style scoped>
.pod-list-panel {
  height: 100%;
  overflow-y: auto;
}

.panel-header {
  padding: 12px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.title {
  font-weight: 500;
  font-size: 14px;
}

.empty-state {
  padding: 40px 20px;
  text-align: center;
  color: var(--el-text-color-secondary);
}

.pod-link {
  color: var(--el-color-primary);
  text-decoration: none;
}

.pod-link:hover {
  text-decoration: underline;
}

.warning {
  color: var(--el-color-warning);
  font-weight: 500;
}
</style>
