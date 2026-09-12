<script setup lang="ts">
import { Plus, Delete } from '@element-plus/icons-vue'
import {
  getNetworkPolicyList,
  getNetworkPolicyYaml,
  updateNetworkPolicyYaml,
  deleteNetworkPolicy,
} from '@/api/resource'
import { useResourceList } from '@/composables/useResourceList'
import YamlEditor from '@/components/YamlEditor.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import { useClusterStore } from '@/stores/cluster'
import { useI18n } from 'vue-i18n'

const clusterStore = useClusterStore()
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
  labelConditions,
  onLabelConditionsChange,
} = useResourceList({
  resourceName: 'NetworkPolicy',
  fetchList: getNetworkPolicyList,
  getYaml: getNetworkPolicyYaml,
  updateYaml: updateNetworkPolicyYaml,
  deleteResource: deleteNetworkPolicy,
  detailRoute: '/network/networkpolicies',
  createRoute: '/network/networkpolicies/create',
  autoRefreshInterval: 30000,
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
      :total-count="totalCount"
      :selected-count="selectedRows.length"
      :cluster-name="clusterStore.clusterName"
      resource-type="networkpolicy"
      :label-conditions="labelConditions"
      @search-input="onSearchInput"
      @namespace-change="handleNamespaceChange"
      @label-selector-change="onLabelConditionsChange"
    >
      <template #actions>
        <el-button type="success" @click="$router.push('/network/networkpolicies/create')">
          <el-icon><Plus /></el-icon> {{ t('common.create') }}
        </el-button>
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
        <el-table-column
          prop="pod_selector"
          :label="t('network.podSelector')"
          min-width="200"
          show-overflow-tooltip
        />
        <el-table-column :label="t('network.policyTypes')" width="160">
          <template #default="{ row }">
            <el-tag
              v-for="pt in row.policy_types || []"
              :key="pt"
              size="small"
              style="margin-right: 4px"
              >{{ pt }}</el-tag
            >
          </template>
        </el-table-column>
        <el-table-column :label="t('network.rules')" min-width="220">
          <template #default="{ row }">
            <div
              style="display: flex; flex-wrap: wrap; gap: var(--gk-space-1); align-items: center"
            >
              <span style="font-size: 12px; color: var(--gk-color-text-primary)"
                >Ingress: {{ row.ingress_rules }}, Egress: {{ row.egress_rules }}</span
              >
              <el-tag
                v-if="row.policy_types?.includes('Ingress') && row.ingress_rules === 0"
                type="danger"
                size="small"
                effect="dark"
                >Deny All Ingress</el-tag
              >
              <el-tag
                v-if="row.policy_types?.includes('Egress') && row.egress_rules === 0"
                type="danger"
                size="small"
                effect="dark"
                >Deny All Egress</el-tag
              >
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="age" :label="t('common.age')" width="120" />
        <el-table-column :label="t('common.actions')" width="240" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button
                size="small"
                @click="
                  $router.push(
                    `/network/networkpolicies/create?clone=${row.name}&namespace=${row.namespace}`,
                  )
                "
                >{{ t('common.clone') }}</el-button
              >
              <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
              <el-button size="small" type="danger" @click="handleDelete(row)">{{
                t('common.delete')
              }}</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- YAML Drawer -->
    <el-drawer
      v-model="yamlDialogVisible"
      title="NetworkPolicy YAML"
      size="85%"
      direction="rtl"
      class="yaml-drawer"
      :body-style="{ padding: '0', height: '100%' }"
    >
      <div v-loading="yamlLoading" style="height: calc(100dvh - 52px)">
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
  padding: var(--gk-space-5);
}
.table-card {
  border-radius: var(--gk-radius-md);
}
</style>
