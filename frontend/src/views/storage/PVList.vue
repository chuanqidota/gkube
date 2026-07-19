<script setup lang="ts">
import { Plus, Delete } from '@element-plus/icons-vue'
import { getPvList, getPvYaml, updatePvYaml, deletePv, transformPvs } from '@/api/resource'
import { statusType } from '@/utils/helpers'
import { useResourceList } from '@/composables/useResourceList'
import YamlEditor from '@/components/YamlEditor.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const {
  loading,
  filteredList,
  searchName,
  onSearchInput,
  selectedRows,
  yamlDialogVisible,
  yamlContent,
  yamlLoading,
  yamlSaving,
  totalCount,
  fetchResources,
  handleSelectionChange,
  handleViewYaml,
  handleSaveYaml,
  handleCancelYaml,
  handleDetail,
  handleDelete,
  handleBatchDelete,
} = useResourceList({
  resourceName: 'PersistentVolume',
  fetchList: getPvList,
  transform: transformPvs,
  getYaml: getPvYaml,
  updateYaml: updatePvYaml,
  deleteResource: deletePv,
  detailRoute: '/storage/pvs',
  deleteConfirm: (row) => `删除持久卷 "${row.name}"?`,
})

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchResources)
</script>

<template>
  <div class="page-container">
    <ResourceListToolbar
      :search-value="searchName"
      :show-namespace="false"
      :total-count="totalCount"
      :selected-count="selectedRows.length"
      @search-input="onSearchInput"
    >
      <template #actions>
        <el-button type="success" @click="$router.push('/storage/pvs/create')">
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
        <el-table-column prop="capacity" label="容量" width="120" />
        <el-table-column prop="access_modes" label="访问模式" min-width="160" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="claim" label="声明" min-width="180" show-overflow-tooltip />
        <el-table-column prop="storage_class" label="存储类" min-width="140" show-overflow-tooltip />
        <el-table-column prop="reclaim_policy" label="回收策略" width="100" />
        <el-table-column prop="age" label="存活时间" width="120" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- YAML Drawer -->
    <el-drawer
      v-model="yamlDialogVisible"
      title="PersistentVolume YAML"
      size="85%"
      direction="rtl"
      class="yaml-drawer"
      :body-style="{ padding: '0', height: '100%' }"
    >
      <div v-loading="yamlLoading" style="height: calc(100vh - 52px);">
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
