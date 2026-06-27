<script setup lang="ts">
import { Refresh, Plus, Delete, Search } from '@element-plus/icons-vue'
import { getDaemonSetList, getDaemonSetYaml, deleteDaemonSet } from '@/api/resource'
import { useResourceList } from '@/composables/useResourceList'
import YamlEditor from '@/components/YamlEditor.vue'

const {
  loading,
  filteredList,
  selectedNamespace,
  searchName,
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
  resourceName: 'DaemonSet',
  fetchList: getDaemonSetList,
  getYaml: getDaemonSetYaml,
  deleteResource: deleteDaemonSet,
  detailRoute: '/workloads/daemonsets',
  createRoute: '/workloads/daemonsets/create',
  paginated: true,
  pageSize: 50,
})
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <el-input v-model="searchName" placeholder="Search by name" style="width: 220px;" clearable>
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select v-model="selectedNamespace" placeholder="All Namespaces" clearable style="width: 180px;" @change="handleNamespaceChange">
          <el-option v-for="ns in namespaceList" :key="ns" :label="ns" :value="ns" />
        </el-select>
        <el-button @click="fetchResources()" :loading="loading"><el-icon><Refresh /></el-icon> Refresh</el-button>
        <el-button type="success" @click="$router.push('/workloads/daemonsets/create')"><el-icon><Plus /></el-icon> Create</el-button>
        <el-button type="danger" :disabled="!selectedRows.length" @click="handleBatchDelete"><el-icon><Delete /></el-icon> Delete ({{ selectedRows.length }})</el-button>
        <span class="total-count" v-if="totalCount">Total: {{ totalCount }}</span>
      </div>
    </el-card>
    <el-card shadow="never" class="table-card">
      <el-table :data="filteredList" v-loading="loading" stripe @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="45" />
        <el-table-column prop="name" label="Name" min-width="200" show-overflow-tooltip>
          <template #default="{ row }"><el-button link type="primary" @click="handleDetail(row)">{{ row.name }}</el-button></template>
        </el-table-column>
        <el-table-column prop="namespace" label="Namespace" width="140" />
        <el-table-column prop="desired" label="Desired" width="100" />
        <el-table-column prop="current" label="Current" width="100" />
        <el-table-column prop="ready" label="Ready" width="100" />
        <el-table-column prop="age" label="Age" width="120" />
        <el-table-column label="Actions" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">Delete</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="No DaemonSets found">
            <el-button type="primary" @click="$router.push('/workloads/daemonsets/create')">Create DaemonSet</el-button>
          </el-empty>
        </template>
      </el-table>
      <div v-if="hasMore" class="load-more">
        <el-button @click="fetchNextPage" :loading="loading" link type="primary">Load More...</el-button>
      </div>
    </el-card>
    <el-dialog v-model="yamlDialogVisible" title="DaemonSet YAML" width="70%" top="5vh" destroy-on-close>
      <div v-loading="yamlLoading"><YamlEditor v-model="yamlContent" height="500px" read-only auto-format /></div>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; gap: 12px; align-items: center; flex-wrap: wrap; }
.table-card { border-radius: 8px; }
.total-count { color: var(--el-text-color-secondary); font-size: 13px; margin-left: auto; }
.load-more { display: flex; justify-content: center; padding: 12px 0; border-top: 1px solid var(--el-border-color-lighter); }
</style>
