<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Delete, Search } from '@element-plus/icons-vue'
import { getRoleList, getRoleYaml, deleteRole } from '@/api/resource'
import { useResourceList } from '@/composables/useResourceList'
import YamlDrawer from '@/components/YamlDrawer.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const { t } = useI18n()

const {
  loading,
  filteredList,
  selectedNamespace,
  searchName,
  onSearchInput,
  selectedRows,
  namespaceList,
  yamlDialogVisible,
  yamlTarget,
  fetchResources,
  handleNamespaceChange,
  handleSelectionChange,
  handleViewYaml,
  handleDelete,
  handleBatchDelete,
} = useResourceList({
  resourceName: 'Role',
  fetchList: getRoleList,
  getYaml: getRoleYaml,
  deleteResource: deleteRole,
  autoRefreshInterval: 30000,
})

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchResources)

function verbTagType(verb: string): string {
  switch (verb) {
    case 'get':
    case 'list':
    case 'watch':
      return 'success'
    case 'create':
      return ''
    case 'update':
    case 'patch':
      return 'warning'
    case 'delete':
    case 'deletecollection':
      return 'danger'
    case '*':
      return 'danger'
    default:
      return 'info'
  }
}

function formatApiGroups(groups: string[]): string {
  if (!groups || groups.length === 0) return '""'
  if (groups.length === 1 && groups[0] === '') return '"" (core)'
  return groups.join(', ')
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
        <el-table-column label="Rules" min-width="360">
          <template #default="{ row }">
            <div v-if="row.rules && row.rules.length > 0" class="rules-cell">
              <div v-for="(rule, idx) in row.rules.slice(0, 3)" :key="idx" class="rule-line">
                <span class="rule-api">{{ formatApiGroups(rule.apiGroups) }}</span>
                <span style="color: #909399;">/</span>
                <span class="rule-resources">{{ (rule.resources || []).join(', ') }}</span>
                <span style="color: #909399;">:</span>
                <el-tag
                  v-for="verb in (rule.verbs || []).slice(0, 4)"
                  :key="verb"
                  :type="verbTagType(verb)"
                  size="small"
                  style="margin: 1px 2px;"
                >
                  {{ verb }}
                </el-tag>
                <el-tag
                  v-if="(rule.verbs || []).length > 4"
                  size="small"
                  type="info"
                  style="margin: 1px 2px;"
                >
                  +{{ rule.verbs.length - 4 }}
                </el-tag>
              </div>
              <div v-if="row.rules.length > 3" class="rule-line" style="color: #909399;">
                +{{ row.rules.length - 3 }} more rules...
              </div>
            </div>
            <span v-else style="color: #909399;">-</span>
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
      resource-type="role"
      :namespace="yamlTarget?.namespace || ''"
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
.rules-cell {
  font-size: 12px;
  line-height: 1.8;
}
.rule-line {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.rule-api {
  color: #b37feb;
  font-family: monospace;
  font-size: 11px;
}
.rule-resources {
  color: #409eff;
  font-family: monospace;
  font-size: 11px;
}
</style>
