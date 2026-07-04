<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Refresh, Delete, Search } from '@element-plus/icons-vue'
import { getClusterRoleList, getClusterRoleYaml, deleteClusterRole } from '@/api/resource'
import { useResourceList } from '@/composables/useResourceList'
import YamlEditor from '@/components/YamlEditor.vue'

const { t } = useI18n()

const {
  loading,
  filteredList,
  searchName,
  onSearchInput,
  selectedRows,
  yamlDialogVisible,
  yamlContent,
  yamlLoading,
  autoRefreshEnabled,
  toggleAutoRefresh,
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
        <el-button @click="fetchResources()" :loading="loading">
          <el-icon><Refresh /></el-icon> {{ t('common.refresh') }}
        </el-button>
        <el-button
          :type="autoRefreshEnabled ? 'success' : 'default'"
          @click="toggleAutoRefresh"
        >
          {{ autoRefreshEnabled ? 'Auto' : 'Manual' }}
        </el-button>
        <el-button type="danger" :disabled="!selectedRows.length" @click="handleBatchDelete">
          <el-icon><Delete /></el-icon> Delete ({{ selectedRows.length }})
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
    <el-dialog v-model="yamlDialogVisible" title="ClusterRole YAML" width="70%" top="5vh" destroy-on-close>
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
</style>
