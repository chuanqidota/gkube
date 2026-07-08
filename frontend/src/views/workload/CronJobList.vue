<script setup lang="ts">
import { Plus, Delete, Search } from '@element-plus/icons-vue'
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
  yamlEditing,
  yamlSaving,
  hasMore,
  totalCount,
  fetchResources,
  fetchNextPage,
  handleNamespaceChange,
  handleSelectionChange,
  handleViewYaml,
  handleEditYaml,
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
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <el-input
          :model-value="searchName"
          @input="onSearchInput"
          placeholder="搜索名称"
          style="width: 220px;"
          clearable
        >
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select
          v-model="selectedNamespace"
          placeholder="所有命名空间"
          clearable
          style="width: 180px;"
          @change="handleNamespaceChange"
        >
          <el-option v-for="ns in namespaceList" :key="ns" :label="ns" :value="ns" />
        </el-select>
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
        <el-button type="success" @click="$router.push('/workloads/cronjobs/create')">
          <el-icon><Plus /></el-icon> 创建
        </el-button>
        <el-button type="danger" :disabled="!selectedRows.length" @click="handleBatchDelete">
          <el-icon><Delete /></el-icon> 删除 ({{ selectedRows.length }})
        </el-button>
        <span class="total-count" v-if="totalCount">总计: {{ totalCount }}</span>
      </div>
    </el-card>

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
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
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
      <div style="padding: 6px 12px; display: flex; gap: 8px; border-bottom: 1px solid var(--el-border-color-lighter);">
        <el-button v-if="!yamlEditing" size="small" type="primary" @click="handleEditYaml">编辑</el-button>
        <template v-else>
          <el-button size="small" type="success" :loading="yamlSaving" @click="handleSaveYaml">保存</el-button>
          <el-button size="small" @click="handleCancelYaml">取消</el-button>
        </template>
      </div>
      <div v-loading="yamlLoading" style="height: calc(100vh - 90px);">
        <YamlEditor v-model="yamlContent" height="100%" :read-only="!yamlEditing" auto-format />
      </div>
    </el-drawer>
  </div>
</template>

<style scoped>
.page-container {
  padding: 20px;
}
.filter-card {
  margin-bottom: 16px;
}
.filter-bar {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}
.table-card {
  border-radius: 8px;
}
.total-count {
  color: var(--el-text-color-secondary);
  font-size: 13px;
  margin-left: auto;
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
