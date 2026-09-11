<script setup lang="ts">
import { Plus, Delete } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import {
  getStatefulSetList,
  getStatefulSetYaml,
  updateStatefulSetYaml,
  deleteStatefulSet,
  transformStatefulSets,
  scaleStatefulSet,
  restartStatefulSet,
  updateStatefulSetImage,
  getStatefulSetDetail,
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
  resourceName: 'StatefulSet',
  fetchList: getStatefulSetList,
  transform: transformStatefulSets,
  getYaml: getStatefulSetYaml,
  updateYaml: updateStatefulSetYaml,
  deleteResource: deleteStatefulSet,
  detailRoute: '/workloads/statefulsets',
  createRoute: '/workloads/statefulsets/create',
  paginated: true,
  pageSize: 50,
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

// ---- Quick Actions ----
const {
  scaleDialogVisible,
  scaleTarget,
  scaleReplicas,
  scaleLoading,
  openScaleDialog,
  handleScaleConfirm,
  handleRestart,
  imageDialogVisible,
  imageTarget,
  imageForm,
  imageContainers,
  imageLoading,
  openImageDialog,
  handleImageConfirm,
} = useListActions({
  kind: 'statefulset',
  scaleApi: scaleStatefulSet,
  restartApi: restartStatefulSet,
  updateImageApi: updateStatefulSetImage,
  detailApi: getStatefulSetDetail,
  onActionSuccess: fetchResources,
})
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
      resource-type="statefulset"
      :label-conditions="labelConditions"
      @search-input="onSearchInput"
      @namespace-change="handleNamespaceChange"
      @label-selector-change="onLabelConditionsChange"
    >
      <template #actions>
        <el-button type="success" @click="$router.push('/workloads/statefulsets/create')">
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
        <el-table-column prop="ready" :label="t('workload.ready')" width="100" />
        <el-table-column
          prop="serviceName"
          :label="t('workload.serviceName')"
          width="160"
          show-overflow-tooltip
        />
        <el-table-column prop="updateStrategy" :label="t('workload.updateStrategy')" width="140" />
        <el-table-column prop="age" label="Age" width="120" />
        <el-table-column :label="t('common.actions')" width="340" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
              <el-button size="small" type="primary" @click="openScaleDialog(row)">{{
                t('workload.scale')
              }}</el-button>
              <el-button size="small" type="warning" @click="handleRestart(row)">{{
                t('workload.restart')
              }}</el-button>
              <el-button size="small" type="primary" @click="openImageDialog(row)">{{
                t('workload.updateImage')
              }}</el-button>
              <el-button size="small" type="danger" @click="handleDelete(row)">{{
                t('common.delete')
              }}</el-button>
            </div>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty :description="t('workload.statefulset')">
            <el-button type="success" @click="$router.push('/workloads/statefulsets/create')">
              <el-icon><Plus /></el-icon> {{ t('common.create') }}
            </el-button>
          </el-empty>
        </template>
      </el-table>

      <!-- Load More Button -->
      <div v-if="hasMore" class="load-more">
        <el-button :loading="loading" link type="primary" @click="fetchNextPage">
          {{ t('workload.loadMore') }}
        </el-button>
      </div>
    </el-card>

    <!-- YAML Drawer -->
    <el-drawer
      v-model="yamlDialogVisible"
      title="StatefulSet YAML"
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

    <!-- Scale Dialog -->
    <el-dialog
      v-model="scaleDialogVisible"
      :title="t('workload.scale')"
      width="420px"
      destroy-on-close
    >
      <div>
        <p style="margin-bottom: var(--gk-space-4)">
          {{ t('workload.targetReplicas') }} <strong>{{ scaleTarget?.name }}</strong>
        </p>
        <el-form-item :label="t('workload.targetReplicas')">
          <el-input-number v-model="scaleReplicas" :min="0" :max="100" style="width: 200px" />
        </el-form-item>
      </div>
      <template #footer>
        <el-button @click="scaleDialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="scaleLoading" @click="handleScaleConfirm">{{
          t('common.confirm')
        }}</el-button>
      </template>
    </el-dialog>

    <!-- Update Image Dialog -->
    <el-dialog
      v-model="imageDialogVisible"
      :title="t('workload.updateImage')"
      width="520px"
      destroy-on-close
    >
      <div>
        <p style="margin-bottom: var(--gk-space-4)">
          {{ t('workload.updateImage') }} <strong>{{ imageTarget?.name }}</strong>
        </p>
        <el-form label-width="80px">
          <el-form-item :label="t('workload.containers')">
            <el-select
              v-model="imageForm.containerName"
              style="width: 100%"
              @change="
                () => {
                  const c = imageContainers.find((c) => c.name === imageForm.containerName)
                  if (c) imageForm.image = c.image
                }
              "
            >
              <el-option
                v-for="c in imageContainers"
                :key="c.name"
                :label="c.name"
                :value="c.name"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('workload.image')">
            <el-input v-model="imageForm.image" placeholder="例如: nginx:1.26" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="imageDialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="imageLoading" @click="handleImageConfirm">{{
          t('common.confirm')
        }}</el-button>
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
.load-more {
  display: flex;
  justify-content: center;
  padding: 12px 0;
  border-top: 1px solid var(--el-border-color-lighter);
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
</style>
