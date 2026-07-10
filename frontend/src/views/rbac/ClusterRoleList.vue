<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Delete, Search } from '@element-plus/icons-vue'
import { getClusterRoleList, deleteClusterRole } from '@/api/resource'
import { useResourceList } from '@/composables/useResourceList'
import YamlDrawer from '@/components/YamlDrawer.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const { t } = useI18n()

const {
  loading,
  filteredList,
  searchName,
  onSearchInput,
  selectedRows,
  yamlDialogVisible,
  yamlTarget,
  fetchResources,
  handleSelectionChange,
  handleViewYaml,
  handleDelete,
  handleBatchDelete,
} = useResourceList({
  resourceName: 'ClusterRole',
  fetchList: getClusterRoleList,
  getYaml: getClusterRoleYaml,
  deleteResource: deleteClusterRole,
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
          :placeholder="t('common.searchByName')"
          style="width: 220px;"
          clearable
        >
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
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
        <el-table-column prop="name" label="Name" min-width="250" show-overflow-tooltip>
          <template #default="{ row }">
            <span>{{ row.name }}</span>
            <el-tag v-if="row.isSystem" size="small" type="info" style="margin-left: 8px;">System</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="labels" label="Labels" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <el-tag v-for="(v, k) in (row.labels || {})" :key="k" size="small" style="margin: 2px;">
              {{ k }}={{ v }}
            </el-tag>
            <span v-if="!row.labels || Object.keys(row.labels).length === 0" style="color: #999;">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="age" label="Age" width="120" />
        <el-table-column label="Actions" width="160" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- YAML Dialog -->
    <YamlDrawer
      v-model="yamlDialogVisible"
      resource-type="clusterrole"
      :name="yamlTarget?.name || ''"
      @saved="fetchResources"
    />
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
</style>
