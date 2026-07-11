<script setup lang="ts">
import { Plus, Delete } from '@element-plus/icons-vue'
import {
  getServiceList,
  getServiceYaml,
  updateServiceYaml,
  deleteService,
  transformServices,
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
  totalCount,
  fetchResources,
  handleNamespaceChange,
  handleSelectionChange,
  handleViewYaml,
  handleSaveYaml,
  handleCancelYaml,
  handleDetail,
  handleDelete,
  handleBatchDelete,
} = useResourceList({
  resourceName: 'Service',
  fetchList: getServiceList,
  transform: transformServices,
  getYaml: getServiceYaml,
  updateYaml: updateServiceYaml,
  deleteResource: deleteService,
  detailRoute: '/network/services',
  createRoute: '/network/services/create',
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
      :total-count="totalCount"
      :selected-count="selectedRows.length"
      @search-input="onSearchInput"
      @namespace-change="handleNamespaceChange"
    >
      <template #actions>
        <el-button type="success" @click="$router.push('/network/services/create')">
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
        <el-table-column prop="name" label="名称" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <el-button link type="primary" @click="handleDetail(row)">{{ row.name }}</el-button>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="140" />
        <el-table-column prop="type" label="类型" width="130" />
        <el-table-column prop="cluster_ip" label="Cluster IP" width="150" />
        <el-table-column prop="ports" label="端口" min-width="160" show-overflow-tooltip />
        <el-table-column prop="age" label="Age" width="120" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- YAML Drawer -->
    <el-drawer v-model="yamlDialogVisible" title="Service YAML" size="85%" direction="rtl" class="yaml-drawer"
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
</style>

<style>
.yaml-drawer .el-drawer__header {
  padding: 6px 16px;
  margin-bottom: 0;
  min-height: auto;
}
</style>
