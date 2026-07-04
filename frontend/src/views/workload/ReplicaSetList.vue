<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Refresh, Delete, Search } from '@element-plus/icons-vue'
import { getReplicaSetList, getReplicaSetYaml, deleteReplicaSet, calcAge } from '@/api/resource'
import { useResourceList } from '@/composables/useResourceList'
import YamlEditor from '@/components/YamlEditor.vue'

const { t } = useI18n()

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
  yamlContent,
  yamlLoading,
  autoRefreshEnabled,
  toggleAutoRefresh,
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
        <el-table-column prop="name" label="Name" min-width="200" show-overflow-tooltip />
        <el-table-column prop="namespace" label="Namespace" width="140" />
        <el-table-column prop="desired" label="Desired" width="90" align="center" />
        <el-table-column prop="current" label="Current" width="90" align="center" />
        <el-table-column prop="ready" label="Ready" width="90" align="center" />
        <el-table-column prop="available" label="Available" width="100" align="center" />
        <el-table-column prop="owner" label="Owner" min-width="160" show-overflow-tooltip />
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
    <el-dialog v-model="yamlDialogVisible" title="ReplicaSet YAML" width="70%" top="5vh" destroy-on-close>
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
