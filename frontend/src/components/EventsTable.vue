<script setup lang="ts">
import { formatAge } from '@/utils/helpers'

interface Props {
  events: any[]
  loading?: boolean
  timeField?: string
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  timeField: 'lastTimestamp',
})

function getAge(row: any): string {
  const ts = row[props.timeField] || row.lastTimestamp || row.last_seen || ''
  return ts ? formatAge(ts, false) : '-'
}
</script>

<template>
  <div class="events-table" v-loading="loading">
    <el-table :data="events" style="width: 100%" size="small" height="100%">
      <el-table-column label="Type" width="80">
        <template #default="{ row }">
          <el-tag :type="row.type === 'Normal' ? 'success' : 'warning'" size="small">
            {{ row.type || '-' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="Reason" width="160">
        <template #default="{ row }">
          <span style="font-size: 12px; font-weight: 500;">{{ row.reason || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="Message" min-width="300" show-overflow-tooltip>
        <template #default="{ row }">
          <span style="font-size: 12px; color: var(--el-text-color-secondary);">{{ row.message || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="Age" width="100">
        <template #default="{ row }">
          <span style="font-size: 12px; color: var(--el-text-color-placeholder);">{{ getAge(row) }}</span>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<style scoped>
.events-table {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
</style>
