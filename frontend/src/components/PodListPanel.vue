<script setup lang="ts">
import { ElTag, ElButton } from 'element-plus'
import { useRouter } from 'vue-router'
import { formatAge } from '@/utils/time'

const router = useRouter()

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

</script>

<template>
  <div class="pod-list-panel" v-loading="loading">
    <div v-if="pods.length === 0 && !loading" class="empty-state">
      暂无 Pod
    </div>
    <div v-else class="table-wrapper">
      <el-table :data="pods" style="width: 100%" row-key="metadata.name" size="small" height="100%">
      <el-table-column label="名称" min-width="260">
        <template #default="{ row }">
          <el-button link type="primary" @click="router.push(`/workloads/pods/${row.metadata.namespace}/${row.metadata.name}`)">{{ row.metadata.name }}</el-button>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="getStatusType(row.status.phase)" size="small">
            {{ row.status.phase }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="Pod IP" width="120">
        <template #default="{ row }">
          <span class="mono">{{ row.status?.podIP || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="节点 IP" width="120">
        <template #default="{ row }">
          <span class="mono">{{ row.status?.hostIP || '-' }}</span>
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
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="emit('logs', row)">日志</el-button>
          <el-button size="small" type="success" @click="emit('exec', row)">终端</el-button>
          <el-button size="small" type="danger" plain @click="emit('delete', row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    </div>
  </div>
</template>

<style scoped>
.pod-list-panel {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.table-wrapper {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.empty-state {
  padding: 32px 20px;
  text-align: center;
  color: var(--el-text-color-secondary);
  font-size: 13px;
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
