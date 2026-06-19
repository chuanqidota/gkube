<script setup lang="ts">
import { ElTag } from 'element-plus'
import { formatAge } from '@/utils/time'

interface ReplicaSet {
  metadata: {
    name: string
    namespace: string
    creationTimestamp: string
    annotations?: Record<string, string>
    ownerReferences?: Array<{ name: string; kind: string }>
  }
  spec: {
    replicas?: number
    template: {
      spec: {
        containers: Array<{ name: string; image: string }>
      }
    }
  }
  status: {
    readyReplicas?: number
    availableReplicas?: number
  }
}

const props = defineProps<{
  replicasets: ReplicaSet[]
  currentRevision: number
  loading: boolean
  selectedName?: string
}>()

const emit = defineEmits<{
  select: [rs: ReplicaSet]
  rollback: [rs: ReplicaSet]
}>()

const getRevision = (rs: ReplicaSet): number => {
  const revStr = rs.metadata.annotations?.['deployment.kubernetes.io/revision']
  return revStr ? parseInt(revStr, 10) : 0
}

const getImage = (rs: ReplicaSet): string => {
  const image = rs.spec.template.spec.containers[0]?.image || ''
  return image.length > 30 ? image.substring(0, 30) + '...' : image
}

const getReplicas = (rs: ReplicaSet): string => {
  const ready = rs.status.readyReplicas || 0
  const desired = rs.spec.replicas || 0
  return `${ready}/${desired}`
}

const getStatus = (rs: ReplicaSet): { text: string; type: 'success' | 'primary' | 'info' } => {
  const revision = getRevision(rs)
  if (revision === props.currentRevision) {
    return { text: 'Current', type: 'success' }
  }
  const ready = rs.status.readyReplicas || 0
  if (ready > 0) {
    return { text: 'Active', type: 'primary' }
  }
  return { text: 'Inactive', type: 'info' }
}

const handleSelect = (rs: ReplicaSet) => {
  emit('select', rs)
}

const handleRollback = (rs: ReplicaSet) => {
  emit('rollback', rs)
}
</script>

<template>
  <div class="replicaset-panel" v-loading="loading">
    <div v-if="replicasets.length === 0" class="empty-state">
      No ReplicaSets found
    </div>
    <div
      v-for="rs in replicasets"
      :key="rs.metadata.name"
      class="rs-item"
      :class="{ selected: rs.metadata.name === selectedName }"
      @click="handleSelect(rs)"
    >
      <div class="rs-header">
        <span class="rs-name">{{ rs.metadata.name }}</span>
        <el-tag :type="getStatus(rs).type" size="small">
          {{ getStatus(rs).text }}
        </el-tag>
      </div>
      <div class="rs-details">
        <div class="rs-detail">
          <span class="label">Image:</span>
          <span class="value">{{ getImage(rs) }}</span>
        </div>
        <div class="rs-detail">
          <span class="label">Created:</span>
          <span class="value">{{ formatAge(rs.metadata.creationTimestamp) }}</span>
        </div>
        <div class="rs-detail">
          <span class="label">Replicas:</span>
          <span class="value">{{ getReplicas(rs) }}</span>
        </div>
      </div>
      <div class="rs-actions" v-if="getRevision(rs) !== currentRevision">
        <el-button size="small" @click.stop="handleRollback(rs)">Rollback</el-button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.replicaset-panel {
  height: 100%;
  overflow-y: auto;
}

.empty-state {
  padding: 40px 20px;
  text-align: center;
  color: var(--el-text-color-secondary);
}

.rs-item {
  padding: 12px 16px;
  cursor: pointer;
  border-bottom: 1px solid var(--el-border-color-lighter);
  transition: background-color 0.2s;
}

.rs-item:hover {
  background-color: var(--el-fill-color-light);
}

.rs-item.selected {
  background-color: var(--el-color-primary-light-9);
  border-left: 3px solid var(--el-color-primary);
}

.rs-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.rs-name {
  font-weight: 500;
  font-size: 14px;
}

.rs-details {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.rs-detail {
  margin-bottom: 4px;
}

.rs-detail .label {
  margin-right: 4px;
}

.rs-actions {
  margin-top: 8px;
}
</style>
