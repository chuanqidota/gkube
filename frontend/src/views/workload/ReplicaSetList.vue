<script setup lang="ts">
import { Delete } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import { getReplicaSetList, getReplicaSetYaml, deleteReplicaSet } from '@/api/resource'
import { formatAge } from '@/utils/helpers'
import { useResourceList } from '@/composables/useResourceList'
import YamlEditor from '@/components/YamlEditor.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import { useClusterStore } from '@/stores/cluster'

const clusterStore = useClusterStore()
const { t } = useI18n()
void t // used in template

function transformReplicaSets(items: any[]) {
  if (!Array.isArray(items)) return []
  return items.map((rs: any) => {
    const ownerRefs = rs.owner_references || []
    const owner = ownerRefs.find((ref: any) => ref.kind === 'Deployment')
    return {
      name: rs.name || '',
      namespace: rs.namespace || '',
      desired: rs.desired || 0,
      current: rs.current || 0,
      ready: rs.ready || 0,
      available: rs.available || 0,
      owner: owner ? `Deployment/${owner.name}` : '-',
      age: formatAge(rs.creation_timestamp, false),
    }
  })
}

const {
  loading,
  filteredList,
  selectedNamespace,
  searchName,
  onSearchInput,
  selectedRows,
  namespaceList,
  yamlDialogVisible,
  totalCount,
  yamlContent,
  yamlLoading,
  hasMore,
  fetchResources,
  fetchNextPage,
  handleNamespaceChange,
  handleSelectionChange,
  handleViewYaml,
  handleDelete,
  handleBatchDelete,
  handleDetail,
  labelConditions,
  onLabelConditionsChange,
} = useResourceList({
  resourceName: 'ReplicaSet',
  fetchList: getReplicaSetList,
  transform: transformReplicaSets,
  getYaml: getReplicaSetYaml,
  deleteResource: deleteReplicaSet,
  detailRoute: '/workloads/replicasets',
  autoRefreshInterval: 30000,
  paginated: true,
  pageSize: 50,
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
</script>

<template>
  <div class="page-container">
    <ResourceListToolbar
      v-model:namespace-value="selectedNamespace"
      :search-value="searchName"
      :namespace-list="namespaceList"
      :show-create="false"
      :total-count="totalCount"
      :selected-count="selectedRows.length"
      :cluster-name="clusterStore.clusterName"
      resource-type="replicaset"
      :label-conditions="labelConditions"
      @search-input="onSearchInput"
      @namespace-change="handleNamespaceChange"
      @label-selector-change="onLabelConditionsChange"
    >
      <template #actions>
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
        <el-table-column prop="desired" :label="t('workload.desired')" width="90" align="center" />
        <el-table-column prop="current" :label="t('workload.current')" width="90" align="center" />
        <el-table-column prop="ready" :label="t('workload.ready')" width="90" align="center" />
        <el-table-column
          prop="available"
          :label="t('workload.available')"
          width="100"
          align="center"
        />
        <el-table-column
          prop="owner"
          :label="t('workload.selector')"
          min-width="160"
          show-overflow-tooltip
        />
        <el-table-column prop="age" label="Age" width="120" />
        <el-table-column :label="t('common.actions')" width="140" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
              <el-button size="small" type="danger" @click="handleDelete(row)">{{
                t('common.delete')
              }}</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <div v-if="hasMore" class="load-more">
        <el-button :loading="loading" link type="primary" @click="fetchNextPage">
          {{ t('workload.loadMore') }}
        </el-button>
      </div>
    </el-card>

    <!-- YAML Drawer -->
    <el-drawer
      v-model="yamlDialogVisible"
      title="ReplicaSet YAML"
      size="85%"
      direction="rtl"
      class="yaml-drawer"
      :body-style="{ padding: '0', height: '100%' }"
    >
      <div v-loading="yamlLoading" style="height: 100%">
        <YamlEditor v-model="yamlContent" height="calc(100dvh - 56px)" read-only auto-format />
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
  text-align: center;
  padding: 12px 0;
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
