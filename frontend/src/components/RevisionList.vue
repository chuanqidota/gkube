<script setup lang="ts">
import { formatAge } from '@/utils/helpers'

export interface RevisionItem {
  name: string
  revision: string | number
  images: string[]
  createdAt: string
  podCount: number
  isCurrent: boolean
}

interface Props {
  revisions: RevisionItem[]
  selectedRevision: string | null
  loading?: boolean
}

defineProps<Props>()

defineEmits<{
  select: [revision: RevisionItem]
  rollback: [revision: RevisionItem]
}>()
</script>

<template>
  <div class="revision-list" v-loading="loading">
    <div v-if="revisions.length === 0 && !loading" class="empty-state">
      暂无修订历史
    </div>
    <div v-else class="rs-list">
      <div
        v-for="rs in revisions"
        :key="rs.name"
        class="rs-item"
        :class="{ active: selectedRevision === rs.name }"
        @click="$emit('select', rs)"
      >
        <div class="rs-name">{{ rs.name }}</div>
        <div class="rs-meta">
          <span class="rs-rev">Rev {{ rs.revision }}</span>
          <span class="rs-replicas">{{ rs.podCount }} Pod{{ rs.podCount !== 1 ? 's' : '' }}</span>
          <el-tag v-if="rs.isCurrent" type="success" size="small">当前</el-tag>
        </div>
        <div v-for="img in rs.images" :key="img" class="rs-image">{{ img }}</div>
        <div class="rs-age">{{ formatAge(rs.createdAt, false) }}</div>
        <div v-if="!rs.isCurrent" class="rs-rollback">
          <el-button size="small" type="warning" @click.stop="$emit('rollback', rs)">
            回滚
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.revision-list {
  flex: 1;
  overflow-y: auto;
}

.empty-state {
  padding: 32px 20px;
  text-align: center;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.rs-list {
  flex: 1;
  overflow-y: auto;
}

.rs-item {
  padding: var(--gk-space-2) var(--gk-space-4);
  border-bottom: 1px solid var(--gk-color-border-light);
  cursor: pointer;
  transition: background 0.15s;
}

.rs-item:hover {
  background: var(--gk-color-primary-bg);
}

.rs-item.active {
  background: var(--gk-color-primary-bg);
  border-left: 3px solid var(--gk-color-primary);
}

.rs-name {
  font-size: var(--gk-font-size-sm);
  font-weight: 500;
  font-family: var(--gk-font-mono);
  word-break: break-all;
  margin-bottom: var(--gk-space-1);
}

.rs-meta {
  display: flex;
  align-items: center;
  gap: var(--gk-space-2);
  margin-bottom: var(--gk-space-1);
}

.rs-rev {
  font-size: var(--gk-font-size-xs);
  color: var(--gk-color-primary);
  font-weight: 500;
}

.rs-replicas {
  font-size: var(--gk-font-size-xs);
  color: var(--gk-color-text-secondary);
}

.rs-image {
  font-size: var(--gk-font-size-xs);
  color: var(--gk-color-text-secondary);
  word-break: break-all;
  margin-bottom: 2px;
}

.rs-age {
  font-size: var(--gk-font-size-xs);
  color: var(--gk-color-text-placeholder);
}

.rs-rollback {
  margin-top: 6px;
}
</style>
