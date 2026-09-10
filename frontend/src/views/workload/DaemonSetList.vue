<script setup lang="ts">
import { Plus, Delete } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import {
  getDaemonSetList,
  getDaemonSetYaml,
  updateDaemonSetYaml,
  deleteDaemonSet,
  transformDaemonSets,
  restartDaemonSet,
  updateDaemonSetImage,
  getDaemonSetDetail,
} from '@/api/resource'
import { useResourceList } from '@/composables/useResourceList'
import { useListActions } from '@/composables/useListActions'
import YamlEditor from '@/components/YamlEditor.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import { useClusterStore } from '@/stores/cluster'

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
  hasMore,
  totalCount,
  fetchResources,
  fetchNextPage,
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
  resourceName: 'DaemonSet',
  fetchList: getDaemonSetList,
  transform: transformDaemonSets,
  getYaml: getDaemonSetYaml,
  updateYaml: updateDaemonSetYaml,
  deleteResource: deleteDaemonSet,
  detailRoute: '/workloads/daemonsets',
  createRoute: '/workloads/daemonsets/create',
  paginated: true,
  pageSize: 50,
  autoRefreshInterval: 30000,
})

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchResources)

// ---- Quick Actions ----
const {
  handleRestart,
  imageDialogVisible, imageTarget, imageForm, imageContainers, imageLoading,
  openImageDialog, handleImageConfirm,
} = useListActions({
  kind: 'daemonset',
  restartApi: restartDaemonSet,
  updateImageApi: updateDaemonSetImage,
  detailApi: getDaemonSetDetail,
  onActionSuccess: fetchResources,
})
</script>

<template>
  <div class="page-container">
    <ResourceListToolbar
      :search-value="searchName"
      v-model:namespace-value="selectedNamespace"
      :namespace-list="namespaceList"
      :total-count="totalCount"
      :selected-count="selectedRows.length"
      :cluster-name="clusterStore.clusterName"
      resource-type="daemonset"
      :label-conditions="labelConditions"
      @search-input="onSearchInput"
      @namespace-change="handleNamespaceChange"
      @label-selector-change="onLabelConditionsChange"
    >
      <template #actions>
        <el-button type="success" @click="$router.push('/workloads/daemonsets/create')">
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
        :data="filteredList"
        v-loading="loading"
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="45" />
        <el-table-column prop="name" :label="t('common.name')" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <el-button link type="primary" @click="handleDetail(row)">{{ row.name }}</el-button>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" :label="t('common.namespace_label')" width="140" />
        <el-table-column prop="desired" :label="t('workload.desired')" width="90" />
        <el-table-column prop="current" :label="t('workload.current')" width="90" />
        <el-table-column prop="ready" :label="t('workload.ready')" width="90" />
        <el-table-column prop="updateStrategy" :label="t('workload.updateStrategy')" width="120" />
        <el-table-column prop="age" label="Age" width="120" />
        <el-table-column :label="t('common.actions')" width="300" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
            <el-button size="small" type="warning" @click="handleRestart(row)">{{ t('workload.restart') }}</el-button>
            <el-button size="small" type="primary" @click="openImageDialog(row)">{{ t('workload.updateImage') }}</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">{{ t('common.delete') }}</el-button>
            </div>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty :description="t('workload.daemonset')">
            <el-button type="success" @click="$router.push('/workloads/daemonsets/create')">
              <el-icon><Plus /></el-icon> {{ t('common.create') }}
            </el-button>
          </el-empty>
        </template>
      </el-table>

      <!-- Load More Button -->
      <div v-if="hasMore" class="load-more">
        <el-button @click="fetchNextPage" :loading="loading" link type="primary">
          {{ t('workload.loadMore') }}
        </el-button>
      </div>
    </el-card>

    <!-- YAML Drawer -->
    <el-drawer v-model="yamlDialogVisible" title="DaemonSet YAML" size="85%" direction="rtl" class="yaml-drawer"
      :body-style="{ padding: '0', height: '100%' }">
      <div v-loading="yamlLoading" style="height: calc(100dvh - 52px);">
        <YamlEditor v-model="yamlContent" height="100%" auto-format show-save-buttons :saving="yamlSaving" @save="handleSaveYaml" @cancel="handleCancelYaml" />
      </div>
    </el-drawer>
    <!-- Update Image Dialog -->
    <el-dialog v-model="imageDialogVisible" :title="t('workload.updateImage')" width="520px" destroy-on-close>
      <div>
        <p style="margin-bottom: var(--gk-space-4);">{{ t('workload.updateImage') }} <strong>{{ imageTarget?.name }}</strong></p>
        <el-form label-width="80px">
          <el-form-item :label="t('workload.containers')">
            <el-select v-model="imageForm.containerName" style="width: 100%;" @change="() => { const c = imageContainers.find((c: any) => c.name === imageForm.containerName); if (c) imageForm.image = c.image }">
              <el-option v-for="c in imageContainers" :key="c.name" :label="c.name" :value="c.name" />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('workload.image')">
            <el-input v-model="imageForm.image" placeholder="例如: nginx:1.26" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="imageDialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="imageLoading" @click="handleImageConfirm">{{ t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container {
  padding: var(--gk-space-5);
}
.table-card {
  border-radius: var(--gk-radius-md);
}
.action-buttons {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: var(--gk-space-1);
}
.action-buttons .el-button + .el-button {
  margin-left: 0;
}
.load-more {
  display: flex;
  justify-content: center;
  padding: 12px 0;
  border-top: 1px solid var(--el-border-color-lighter);
}
</style>
