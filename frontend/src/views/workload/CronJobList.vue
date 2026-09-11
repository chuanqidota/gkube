<script setup lang="ts">
import { Plus, Delete, VideoPause, VideoPlay } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  getCronJobList,
  getCronJobYaml,
  updateCronJobYaml,
  deleteCronJob,
  transformCronJobs,
  suspendCronJob,
  resumeCronJob,
  triggerCronJob,
} from '@/api/resource'
import { useResourceList } from '@/composables/useResourceList'
import YamlEditor from '@/components/YamlEditor.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import { useClusterStore } from '@/stores/cluster'

const clusterStore = useClusterStore()
const { t } = useI18n()

const {
  loading,
  filteredList,
  selectedNamespace,
  searchName,
  onSearchInput,
  selectedRows,
  namespaceList,
  yamlDialogVisible,
  yamlContent,
  yamlLoading,
  yamlSaving,
  hasMore,
  totalCount,
  fetchResources,
  fetchNextPage,
  handleNamespaceChange,
  handleSelectionChange,
  handleViewYaml,
  handleSaveYaml,
  handleCancelYaml,
  handleDetail,
  handleDelete,
  handleBatchDelete,
  labelConditions,
  onLabelConditionsChange,
} = useResourceList({
  resourceName: 'CronJob',
  fetchList: getCronJobList,
  transform: transformCronJobs,
  getYaml: getCronJobYaml,
  updateYaml: updateCronJobYaml,
  deleteResource: deleteCronJob,
  detailRoute: '/workloads/cronjobs',
  createRoute: '/workloads/cronjobs/create',
  paginated: true,
  pageSize: 50,
  autoRefreshInterval: 30000,
})

const {
  isRunning,
  countdown,
  currentInterval,
  availableIntervals,
  toggle,
  refresh: manualRefresh,
  setIntervalOption,
} = useAutoRefresh(fetchResources)

// 暂停 / 恢复 CronJob：使用专用 API 端点
async function handleToggleSuspend(row: any) {
  const willSuspend = !row.suspend
  const actionLabel = willSuspend
    ? t('workload.suspend')
    : t('workload.hpaResumeSuccess').split(' ')[1] || 'Resume'
  try {
    await ElMessageBox.confirm(
      t('workload.suspendConfirm', { action: actionLabel, type: 'CronJob', name: row.name }),
      t('common.confirmAction'),
      { type: 'warning' },
    )
  } catch {
    return // 取消
  }
  try {
    if (willSuspend) {
      await suspendCronJob({ namespace: row.namespace, name: row.name })
    } else {
      await resumeCronJob({ namespace: row.namespace, name: row.name })
    }
    ElMessage.success(t('workload.suspendSuccess', { name: row.name, action: actionLabel }))
    fetchResources()
  } catch (e: any) {
    ElMessage.error(e?.message || t('workload.suspendFailed', { action: actionLabel }))
  }
}

// 手动触发 CronJob
async function handleTrigger(row: any) {
  try {
    await ElMessageBox.confirm(
      t('workload.triggerConfirm', { name: row.name }),
      t('common.confirmAction'),
      { type: 'info' },
    )
  } catch {
    return
  }
  try {
    await triggerCronJob({ namespace: row.namespace, name: row.name })
    ElMessage.success(t('workload.triggerSuccess', { name: row.name }))
    fetchResources()
  } catch (e: any) {
    ElMessage.error(e?.message || t('workload.triggerFailed'))
  }
}
</script>

<template>
  <div class="page-container">
    <ResourceListToolbar
      v-model:namespace-value="selectedNamespace"
      :search-value="searchName"
      :namespace-list="namespaceList"
      :total-count="totalCount"
      :selected-count="selectedRows.length"
      :cluster-name="clusterStore.clusterName"
      resource-type="cronjob"
      :label-conditions="labelConditions"
      @search-input="onSearchInput"
      @namespace-change="handleNamespaceChange"
      @label-selector-change="onLabelConditionsChange"
    >
      <template #actions>
        <el-button type="success" @click="$router.push('/workloads/cronjobs/create')">
          <el-icon><Plus /></el-icon> {{ t('common.create') }}
        </el-button>
        <el-button type="danger" :disabled="!selectedRows.length" @click="handleBatchDelete">
          <el-icon><Delete /></el-icon> {{ t('common.delete') }} ({{ selectedRows.length }})
        </el-button>
      </template>
      <template #extra>
        <AutoRefreshToolbar
          :is-running="isRunning"
          :countdown="countdown"
          :current-interval="currentInterval"
          :available-intervals="availableIntervals"
          :loading="loading"
          @refresh="manualRefresh()"
          @toggle="toggle()"
          @interval-change="setIntervalOption"
        />
      </template>
    </ResourceListToolbar>

    <el-card shadow="never" class="table-card">
      <el-table
        v-loading="loading"
        :data="filteredList"
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="45" />
        <el-table-column
          prop="name"
          :label="t('common.name')"
          min-width="200"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            <el-button link type="primary" @click="handleDetail(row)">{{ row.name }}</el-button>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" :label="t('common.namespace_label')" width="140" />
        <el-table-column prop="schedule" :label="t('workload.schedule')" width="160" />
        <el-table-column :label="t('workload.suspend')" width="90">
          <template #default="{ row }">
            <el-tag :type="row.suspend ? 'warning' : 'success'" size="small">{{
              row.suspend ? '是' : '否'
            }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="active" :label="t('workload.active')" width="80" />
        <el-table-column
          prop="nextScheduleTime"
          :label="t('workload.lastScheduleTime')"
          width="170"
        >
          <template #default="{ row }">
            <span v-if="row.nextScheduleTime">{{ row.nextScheduleTime }}</span>
            <el-tag v-else-if="row.suspend" type="info" size="small">{{
              t('workload.suspended')
            }}</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="age" :label="t('common.age')" width="120" />
        <el-table-column :label="t('common.actions')" width="350" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
              <el-button
                size="small"
                :type="row.suspend ? 'success' : 'warning'"
                :icon="row.suspend ? VideoPlay : VideoPause"
                @click="handleToggleSuspend(row)"
                >{{
                  row.suspend
                    ? t('workload.hpaResumeSuccess').split(' ')[1] || 'Resume'
                    : t('workload.suspend')
                }}</el-button
              >
              <el-button size="small" type="primary" @click="handleTrigger(row)">{{
                t('workload.triggerSuccess').split(' ')[0] || 'Trigger'
              }}</el-button>
              <el-button size="small" type="danger" @click="handleDelete(row)">{{
                t('common.delete')
              }}</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- Load More Button -->
      <div v-if="hasMore" class="load-more">
        <el-button :loading="loading" link type="primary" @click="fetchNextPage">
          {{ t('workload.loadMore') }}
        </el-button>
      </div>
    </el-card>

    <!-- YAML Drawer -->
    <el-drawer
      v-model="yamlDialogVisible"
      title="CronJob YAML"
      size="85%"
      direction="rtl"
      class="yaml-drawer"
      :body-style="{ padding: '0', height: '100%' }"
    >
      <div v-loading="yamlLoading" style="height: calc(100dvh - 52px)">
        <YamlEditor
          v-model="yamlContent"
          height="100%"
          auto-format
          show-save-buttons
          :saving="yamlSaving"
          @save="handleSaveYaml"
          @cancel="handleCancelYaml"
        />
      </div>
    </el-drawer>
  </div>
</template>

<style scoped>
.page-container {
  padding: var(--gk-space-5);
}
.table-card {
  border-radius: var(--gk-radius-md);
}
.load-more {
  display: flex;
  justify-content: center;
  padding: 12px 0;
  border-top: 1px solid var(--el-border-color-lighter);
}
</style>

<style>
.action-buttons {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: var(--gk-space-1);
}
.action-buttons .el-button + .el-button {
  margin-left: 0;
}
</style>
