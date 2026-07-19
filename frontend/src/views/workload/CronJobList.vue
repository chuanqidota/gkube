<script setup lang="ts">
import { Plus, Delete, VideoPause, VideoPlay } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import yaml from 'js-yaml'
import {
  getCronJobList,
  getCronJobYaml,
  updateCronJobYaml,
  deleteCronJob,
  transformCronJobs,
} from '@/api/resource'
import { useResourceList } from '@/composables/useResourceList'
import YamlEditor from '@/components/YamlEditor.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

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

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchResources)

// 暂停 / 恢复 CronJob：后端无专用接口，通过全量 YAML 更新切换 spec.suspend
async function handleToggleSuspend(row: any) {
  const willSuspend = !row.suspend
  const actionLabel = willSuspend ? '暂停' : '恢复'
  try {
    await ElMessageBox.confirm(
      `确定要${actionLabel} CronJob "${row.name}" 吗？`,
      `确认${actionLabel}`,
      { type: 'warning' }
    )
  } catch {
    return // 取消
  }
  try {
    const res: any = await getCronJobYaml({ namespace: row.namespace, name: row.name })
    const raw = res.data?.yaml ?? res.data ?? ''
    const doc: any = yaml.load(raw) || {}
    if (!doc.spec) doc.spec = {}
    doc.spec.suspend = willSuspend
    await updateCronJobYaml({ namespace: row.namespace, name: row.name, yaml: yaml.dump(doc) })
    ElMessage.success(`${row.name} 已${actionLabel}`)
    fetchResources()
  } catch (e: any) {
    ElMessage.error(e?.message || `${actionLabel}失败`)
  }
}
</script>

<template>
  <div class="page-container">
    <ResourceListToolbar
      :search-value="searchName"
      v-model:namespace-value="selectedNamespace"
      :namespace-list="namespaceList"
      :total-count="totalCount"
      :selected-count="selectedRows.length"
      @search-input="onSearchInput"
      @namespace-change="handleNamespaceChange"
    >
      <template #actions>
        <el-button type="success" @click="$router.push('/workloads/cronjobs/create')">
          <el-icon><Plus /></el-icon> 创建
        </el-button>
        <el-button type="danger" :disabled="!selectedRows.length" @click="handleBatchDelete">
          <el-icon><Delete /></el-icon> 删除 ({{ selectedRows.length }})
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
        :data="filteredList"
        v-loading="loading"
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="45" />
        <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <el-button link type="primary" @click="handleDetail(row)">{{ row.name }}</el-button>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="140" />
        <el-table-column prop="schedule" label="调度表达式" width="160" />
        <el-table-column label="暂停" width="90">
          <template #default="{ row }">
            <el-tag :type="row.suspend ? 'warning' : 'success'" size="small">{{ row.suspend ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="active" label="活跃" width="80" />
        <el-table-column prop="age" label="Age" width="120" />
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
            <el-button
              size="small"
              :type="row.suspend ? 'success' : 'warning'"
              :icon="row.suspend ? VideoPlay : VideoPause"
              @click="handleToggleSuspend(row)"
            >{{ row.suspend ? '恢复' : '暂停' }}</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- Load More Button -->
      <div v-if="hasMore" class="load-more">
        <el-button @click="fetchNextPage" :loading="loading" link type="primary">
          加载更多...
        </el-button>
      </div>
    </el-card>

    <!-- YAML Drawer -->
    <el-drawer v-model="yamlDialogVisible" title="CronJob YAML" size="85%" direction="rtl" class="yaml-drawer"
      :body-style="{ padding: '0', height: '100%' }">
      <div v-loading="yamlLoading" style="height: calc(100vh - 52px);">
        <YamlEditor v-model="yamlContent" height="100%" auto-format show-save-buttons :saving="yamlSaving" @save="handleSaveYaml" @cancel="handleCancelYaml" />
      </div>
    </el-drawer>
  </div>
</template>

<style scoped>
.page-container {
  padding: 20px;
}
.table-card {
  border-radius: 8px;
}
.load-more {
  display: flex;
  justify-content: center;
  padding: 12px 0;
  border-top: 1px solid var(--el-border-color-lighter);
}
</style>

<style>
.yaml-drawer .el-drawer__header {
  padding: 6px 16px;
  margin-bottom: 0;
  min-height: auto;
}
</style>
