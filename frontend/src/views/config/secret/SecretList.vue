<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'
import { secretApi, transformSecrets, getSecretDetail } from '@/api/resource'
import { base64Decode } from '@/utils/helpers'
import { useResourceList } from '@/composables/useResourceList'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'
import { useClusterStore } from '@/stores/cluster'
import YamlDrawer from '@/components/YamlDrawer.vue'

const { t } = useI18n()
const clusterStore = useClusterStore()

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
  handleDetail,
  handleDelete,
  handleBatchDelete,
  labelConditions,
  onLabelConditionsChange,
} = useResourceList({
  resourceName: 'Secret',
  fetchList: secretApi.list,
  transform: transformSecrets,
  getYaml: secretApi.getYaml,
  updateYaml: secretApi.updateYaml,
  deleteResource: secretApi.delete,
  detailRoute: '/config/secrets',
  createRoute: '/config/secrets/create',
})

const {
  isRunning,
  countdown,
  currentInterval,
  availableIntervals,
  toggle,
  refresh,
  setIntervalOption,
} = useAutoRefresh(fetchResources)

// Secret-specific: data viewer with base64 decode
const dataDialogVisible = ref(false)
const dataDialogTitle = ref('')
const dataEntries = ref<{ key: string; rawValue: string; decodedValue: string }[]>([])
const dataLoading = ref(false)
const showDecoded = ref(true)

async function handleViewData(row: any) {
  try {
    await ElMessageBox.confirm(t('config.secretViewConfirm'), t('config.secretViewData'), {
      type: 'warning',
      confirmButtonText: t('config.secretViewConfirmBtn'),
      cancelButtonText: t('common.cancel'),
    })
  } catch {
    return
  }
  dataLoading.value = true
  dataDialogVisible.value = true
  dataDialogTitle.value = t('config.secretTitle', { name: row.name })
  dataEntries.value = []
  try {
    const res: any = await getSecretDetail({ name: row.name, namespace: row.namespace })
    const data = res.data?.data || res.data || {}
    dataEntries.value = Object.entries(data).map(([key, value]) => {
      const rawValue = String(value ?? '')
      return { key, rawValue, decodedValue: base64Decode(rawValue) }
    })
  } catch (e: any) {
    ElMessage.error(e?.message || t('config.loadDataFailed'))
    dataDialogVisible.value = false
  } finally {
    dataLoading.value = false
  }
}
</script>

<template>
  <div class="page-container">
    <ResourceListToolbar
      v-model:namespace-value="selectedNamespace"
      :search-value="searchName"
      :namespace-list="namespaceList"
      :show-total-count="false"
      :selected-count="selectedRows.length"
      :cluster-name="clusterStore.clusterName"
      resource-type="secret"
      :label-conditions="labelConditions"
      @search-input="onSearchInput"
      @namespace-change="handleNamespaceChange"
      @label-selector-change="onLabelConditionsChange"
    >
      <template #actions>
        <el-button type="success" @click="$router.push('/config/secrets/create')">
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
          @refresh="refresh"
          @toggle="toggle"
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
          <template #default="{ row }"
            ><el-button link type="primary" @click="handleDetail(row)">{{
              row.name
            }}</el-button></template
          >
        </el-table-column>
        <el-table-column prop="namespace" :label="t('common.namespace_label')" width="140" />
        <el-table-column
          prop="type"
          :label="t('common.type')"
          min-width="160"
          show-overflow-tooltip
        />
        <el-table-column :label="t('config.dataKeys')" width="120">
          <template #default="{ row }"
            ><el-tag size="small">{{ row.data_keys_count }}</el-tag></template
          >
        </el-table-column>
        <el-table-column prop="age" :label="t('common.creationTime')" width="120" />
        <el-table-column :label="t('common.actions')" width="240" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
              <el-button size="small" type="primary" @click="handleViewData(row)">{{
                t('config.viewData')
              }}</el-button>
              <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <YamlDrawer
      v-model="yamlDialogVisible"
      :get-yaml="secretApi.getYaml"
      :update-yaml="secretApi.updateYaml"
      :namespace="yamlTarget?.namespace || ''"
      :name="yamlTarget?.name || ''"
      title="Secret YAML"
      @saved="fetchResources"
    />
    <el-dialog v-model="dataDialogVisible" :title="dataDialogTitle" width="60%" top="8vh">
      <div style="margin-bottom: var(--gk-space-3)">
        <el-switch
          v-model="showDecoded"
          :active-text="t('config.decoded')"
          :inactive-text="t('config.raw')"
        />
      </div>
      <div v-loading="dataLoading">
        <el-table :data="dataEntries" stripe style="width: 100%" max-height="400">
          <el-table-column
            prop="key"
            :label="t('config.key')"
            min-width="200"
            show-overflow-tooltip
          />
          <el-table-column label="值" min-width="300">
            <template #default="{ row }"
              ><div
                style="
                  white-space: pre-wrap;
                  word-break: break-all;
                  max-height: 100px;
                  overflow-y: auto;
                "
              >
                {{ showDecoded ? row.decodedValue : row.rawValue }}
              </div></template
            >
          </el-table-column>
        </el-table>
        <el-empty
          v-if="!dataLoading && dataEntries.length === 0"
          :description="t('common.noData')"
        />
      </div>
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
</style>
