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
      ports?: Array<{
        containerPort: number
        protocol?: string
      }>
      resources?: {
        limits?: Record<string, string>
        requests?: Record<string, string>
      }
    }>
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
      <span class="title">{{ replicasetName ? `Pods for ${replicasetName}` : 'Pods' }} ({{ pods.length }})</span>
    </div>
    <div v-if="pods.length === 0" class="empty-state">
      No pods found
    </div>
    <el-table v-else :data="pods" style="width: 100%" row-key="metadata.name">
      <el-table-column type="expand">
        <template #default="{ row }">
          <div class="container-details">
            <h4 style="margin: 0 0 12px 0;">Containers</h4>
            <div v-for="container in row.spec.containers" :key="container.name" class="container-item">
              <el-descriptions :column="2" border size="small">
                <el-descriptions-item label="Name">{{ container.name }}</el-descriptions-item>
                <el-descriptions-item label="Image">{{ container.image }}</el-descriptions-item>
                <el-descriptions-item label="Ports" v-if="container.ports && container.ports.length > 0">
                  <el-tag v-for="port in container.ports" :key="port.containerPort" size="small" style="margin-right: 4px;">
                    {{ port.containerPort }}{{ port.protocol ? `/${port.protocol}` : '' }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="Resources" v-if="container.resources">
                  <div v-if="container.resources.limits">
                    <span class="resource-label">Limits:</span>
                    <span v-for="(val, key) in container.resources.limits" :key="key">
                      {{ key }}={{ val }}
                    </span>
                  </div>
                  <div v-if="container.resources.requests">
                    <span class="resource-label">Requests:</span>
                    <span v-for="(val, key) in container.resources.requests" :key="key">
                      {{ key }}={{ val }}
                    </span>
                  </div>
                </el-descriptions-item>
              </el-descriptions>
            </div>
          </div>
        </template>
      </el-table-column>
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
          {{ formatAge(row.metadata.creationTimestamp, false) }}
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

.container-details {
  padding: 16px;
  background-color: var(--el-fill-color-lighter);
}

.container-item {
  margin-bottom: 12px;
}

.container-item:last-child {
  margin-bottom: 0;
}

.resource-label {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-right: 4px;
}
</style>
