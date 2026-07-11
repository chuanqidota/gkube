<script setup lang="ts">
import { Delete } from '@element-plus/icons-vue'
import { getReplicaSetList, getReplicaSetYaml, deleteReplicaSet, calcAge } from '@/api/resource'
import { useResourceList } from '@/composables/useResourceList'
import YamlEditor from '@/components/YamlEditor.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

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
      age: calcAge(rs.creation_timestamp),
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
  fetchResources,
  handleNamespaceChange,
  handleSelectionChange,
  handleViewYaml,
  handleDelete,
  handleBatchDelete,
} = useResourceList({
  resourceName: 'ReplicaSet',
  fetchList: getReplicaSetList,
  transform: transformReplicaSets,
  getYaml: getReplicaSetYaml,
  deleteResource: deleteReplicaSet,
  autoRefreshInterval: 30000,
})

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchResources)
</script>

<template>
  <div class="page-container">
    <ResourceListToolbar
      :search-value="searchName"
      v-model:namespace-value="selectedNamespace"
      :namespace-list="namespaceList"
      :show-create="false"
      :total-count="totalCount"
      :selected-count="selectedRows.length"
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
        <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip />
        <el-table-column prop="namespace" label="命名空间" width="140" />
        <el-table-column prop="desired" label="期望" width="90" align="center" />
        <el-table-column prop="current" label="当前" width="90" align="center" />
        <el-table-column prop="ready" label="就绪" width="90" align="center" />
        <el-table-column prop="available" label="可用" width="100" align="center" />
        <el-table-column prop="owner" label="拥有者" min-width="160" show-overflow-tooltip />
        <el-table-column prop="age" label="Age" width="120" />
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- YAML Drawer -->
    <el-drawer v-model="yamlDialogVisible" title="ReplicaSet YAML" size="85%" direction="rtl" class="yaml-drawer"
      :body-style="{ padding: '0', height: '100%' }">
      <div v-loading="yamlLoading" style="height: 100%;">
        <YamlEditor v-model="yamlContent" height="calc(100vh - 56px)" read-only auto-format />
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
</style>

<style>
.yaml-drawer .el-drawer__header {
  padding: 6px 16px;
  margin-bottom: 0;
  min-height: auto;
}
</style>
