<script setup lang="ts">
import { ArrowDown } from '@element-plus/icons-vue'
import { getPodList, getPodYaml, deletePod, transformPods } from '@/api/resource'
import { useResourceList } from '@/composables/useResourceList'
import { useClusterStore } from '@/stores/cluster'
import { getPodStatusType, buildFullscreenUrl } from '@/utils/pod'
import YamlEditor from '@/components/YamlEditor.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const clusterStore = useClusterStore()

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
  hasMore,
  totalCount,
  fetchResources,
  fetchNextPage,
  handleNamespaceChange,
  handleSelectionChange,
  handleViewYaml,
  handleDetail,
  handleDelete,
  handleBatchDelete,
} = useResourceList({
  resourceName: 'Pod',
  fetchList: getPodList,
  transform: transformPods,
  getYaml: getPodYaml,
  deleteResource: deletePod,
  forceDeleteResource: (data: any) => deletePod({ ...data, force: true }),
  detailRoute: '/workloads/pods',
  paginated: true,
  pageSize: 50,
})

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchResources)

function handleViewLogs(row: any) {
  window.open(buildFullscreenUrl('logs', { namespace: row.namespace, pod: row.name, cluster: clusterStore.clusterName || undefined }), '_blank')
}

function handleExec(row: any) {
  window.open(buildFullscreenUrl('terminal', { namespace: row.namespace, pod: row.name, cluster: clusterStore.clusterName || undefined }), '_blank')
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
      :show-create="false"
      @search-input="onSearchInput"
      @namespace-change="handleNamespaceChange"
    >
      <template #actions>
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
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getPodStatusType(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="ip" label="Pod IP" width="140" />
        <el-table-column prop="hostIP" label="节点 IP" width="140" />
        <el-table-column prop="restarts" label="重启" width="100" />
        <el-table-column prop="age" label="Age" width="120" />
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
              <el-button size="small" type="primary" @click="handleViewLogs(row)">日志</el-button>
              <el-button size="small" type="success" @click="handleExec(row)">终端</el-button>
              <el-dropdown @command="(cmd: string) => handleDelete(row, cmd === 'force')" trigger="click">
                <el-button size="small" type="danger" plain>
                  删除 <el-icon><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="normal">删除</el-dropdown-item>
                    <el-dropdown-item command="force" divided>强制删除</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- Load More Button -->
      <div v-if="hasMore" class="load-more">
        <el-button @click="fetchNextPage" :loading="loading" link type="primary">
          Load More...
        </el-button>
      </div>
    </el-card>

    <!-- YAML Drawer -->
    <el-drawer v-model="yamlDialogVisible" title="Pod YAML" size="85%" direction="rtl" class="yaml-drawer"
      :body-style="{ padding: '0', height: '100%' }">
      <div v-loading="yamlLoading" style="height: 100%;">
        <YamlEditor v-model="yamlContent" height="calc(100vh - 56px)" read-only auto-format />
      </div>
    </el-drawer>
  </div>
</template>

<style scoped>
.action-buttons {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: 4px;
}
.action-buttons .el-button + .el-button,
.action-buttons .el-dropdown + .el-button {
  margin-left: 0;
}

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
