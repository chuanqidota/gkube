<script setup lang="ts">
import { ElTag, ElButton } from 'element-plus'
import { formatAge } from '@/utils/time'

interface Pod {
  metadata: {
    name: string
    namespace: string
    creationTimestamp: string
  }
  status: {
    phase: string
    podIP?: string
    hostIP?: string
    containerStatuses?: Array<{
      name: string
      restartCount: number
      ready: boolean
      image: string
    }>
  }
  spec: {
    nodeName?: string
    containers: Array<{
      name: string
      image: string
    }>
  }
}

const props = defineProps<{
  pods: Pod[]
  loading: boolean
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

const getImage = (pod: Pod): string => {
  return pod.spec.containers?.[0]?.image || '-'
}
</script>

<template>
  <div class="pod-list-panel" v-loading="loading">
    <div v-if="pods.length === 0 && !loading" class="empty-state">
      暂无 Pod
    </div>
    <el-table v-else :data="pods" style="width: 100%" row-key="metadata.name" size="small">
      <el-table-column label="名称" min-width="200">
        <template #default="{ row }">
          <span class="pod-name">{{ row.metadata.name }}</span>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="getStatusType(row.status.phase)" size="small">
            {{ row.status.phase }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="Pod IP" width="125">
        <template #default="{ row }">
          <span class="mono">{{ row.status?.podIP || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="Node" width="110">
        <template #default="{ row }">
          {{ row.spec.nodeName || '-' }}
        </template>
      </el-table-column>
      <el-table-column label="Node IP" width="125">
        <template #default="{ row }">
          <span class="mono">{{ row.status?.hostIP || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="镜像" min-width="160" show-overflow-tooltip>
        <template #default="{ row }">
          {{ getImage(row) }}
        </template>
      </el-table-column>
      <el-table-column label="重启" width="65">
        <template #default="{ row }">
          <span :class="{ warning: getRestarts(row) > 0 }">{{ getRestarts(row) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="Age" width="75">
        <template #default="{ row }">
          {{ formatAge(row.metadata.creationTimestamp, false) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="170" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="emit('logs', row)">日志</el-button>
          <el-button link type="primary" size="small" @click="emit('exec', row)">终端</el-button>
          <el-button link type="danger" size="small" @click="emit('delete', row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<style scoped>
.pod-list-panel {
  width: 100%;
}

.empty-state {
  padding: 32px 20px;
  text-align: center;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.pod-name {
  font-size: 13px;
  font-family: monospace;
}

.mono {
  font-family: monospace;
  font-size: 12px;
}

.warning {
  color: var(--el-color-warning);
  font-weight: 500;
}
</style>
