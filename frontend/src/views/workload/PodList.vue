<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Delete, Search } from '@element-plus/icons-vue'
import { getPodList, getPodYaml, deletePod, transformPods } from '@/api/resource'
import { useResourceList } from '@/composables/useResourceList'
import { useClusterStore } from '@/stores/cluster'
import YamlEditor from '@/components/YamlEditor.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const { t } = useI18n()
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
  detailRoute: '/workloads/pods',
  paginated: true,
  pageSize: 50,
  autoRefreshInterval: 15000,
})

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchResources)

function getClusterName(): string {
  return clusterStore.currentCluster?.clusterName || clusterStore.currentCluster?.name || ''
}

function handleViewLogs(row: any) {
  const cluster = getClusterName()
  window.open(`/fullscreen/logs?namespace=${row.namespace}&pod=${row.name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
}

function handleExec(row: any) {
  const cluster = getClusterName()
  window.open(`/fullscreen/terminal?namespace=${row.namespace}&pod=${row.name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
}

function statusType(status: string) {
  const s = (status || '').toLowerCase()
  if (s === 'running') return 'success'
  if (s === 'succeeded') return 'info'
  if (s === 'pending') return 'warning'
  if (s === 'failed' || s === 'error') return 'danger'
  return 'info'
}
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <el-input
          :model-value="searchName"
          @input="onSearchInput"
          :placeholder="t('common.searchByName')"
          style="width: 220px;"
          clearable
        >
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select
          v-model="selectedNamespace"
          :placeholder="t('common.allNamespaces')"
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
        <el-button type="danger" :disabled="!selectedRows.length" @click="handleBatchDelete">
          <el-icon><Delete /></el-icon> 删除 ({{ selectedRows.length }})
        </el-button>
        <span class="total-count" v-if="totalCount">Total: {{ totalCount }}</span>
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
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="ip" label="Pod IP" width="140" />
        <el-table-column prop="hostIP" label="节点 IP" width="140" />
        <el-table-column prop="restarts" label="重启" width="100" />
        <el-table-column prop="age" label="Age" width="120" />
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
            <el-button size="small" type="primary" @click="handleViewLogs(row)">日志</el-button>
            <el-button size="small" type="success" @click="handleExec(row)">终端</el-button>
            <el-button size="small" type="danger" plain @click="handleDelete(row)">删除</el-button>
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

    <!-- YAML Dialog -->
    <el-dialog v-model="yamlDialogVisible" title="Pod YAML" width="70%" top="5vh" destroy-on-close>
      <div v-loading="yamlLoading">
        <YamlEditor v-model="yamlContent" height="500px" read-only auto-format />
      </div>
    </el-dialog>
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
