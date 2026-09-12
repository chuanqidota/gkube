<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'
import {
  getConfigMapList,
  getConfigMapDetail,
  getConfigMapYaml,
  updateConfigMap,
  deleteConfigMap,
  getNamespaceList,
  extractNamespaceNames,
  transformConfigMaps,
} from '@/api/resource'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'
import { useClusterStore } from '@/stores/cluster'
import type { LabelCondition } from '@/components/LabelFilterPopover.vue'
import YamlDrawer from '@/components/YamlDrawer.vue'
import ConfigDataViewer from '@/components/ConfigDataViewer.vue'

const { t } = useI18n()
const router = useRouter()
const clusterStore = useClusterStore()
const loading = ref(false)
const configMapList = ref<any[]>([])
const namespaceList = ref<string[]>([])
const selectedNamespace = ref('')
const searchName = ref('')
const labelConditions = ref<LabelCondition[]>([])
const selectedRows = ref<any[]>([])
const yamlDialogVisible = ref(false)
const yamlTarget = ref<{ namespace: string; name: string } | null>(null)
const dataDialogVisible = ref(false)
const dataDialogTitle = ref('')
const dataEntries = ref<{ key: string; value: string }[]>([])
const dataLoading = ref(false)

const filteredList = computed(() => {
  if (!searchName.value) return configMapList.value
  const keyword = searchName.value.toLowerCase()
  return configMapList.value.filter((d) => d.name?.toLowerCase().includes(keyword))
})

async function fetchNamespaces() {
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = extractNamespaceNames(res.data)
  } catch {
    /* ignore */
  }
}

async function fetchConfigMaps() {
  loading.value = true
  try {
    const params: any = {}
    if (selectedNamespace.value) params.namespace = selectedNamespace.value
    if (labelConditions.value.length > 0) params.labelFilters = labelConditions.value
    const res: any = await getConfigMapList(params)
    const items = res.data?.items || res.data || []
    configMapList.value = transformConfigMaps(items)
  } catch {
    // Silently handle — resource may not exist in cluster
  } finally {
    loading.value = false
  }
}

function handleNamespaceChange() {
  fetchConfigMaps()
}
function onLabelConditionsChange(conditions: LabelCondition[]) {
  labelConditions.value = conditions
  fetchConfigMaps()
}
function handleSelectionChange(rows: any[]) {
  selectedRows.value = rows
}

function handleViewYaml(row: any) {
  yamlTarget.value = { namespace: row.namespace, name: row.name }
  yamlDialogVisible.value = true
}

async function handleViewData(row: any) {
  dataLoading.value = true
  dataDialogVisible.value = true
  dataDialogTitle.value = t('config.configMapTitle', { name: row.name })
  dataEntries.value = []
  try {
    const res: any = await getConfigMapDetail({ name: row.name, namespace: row.namespace })
    const data = res.data?.data || {}
    const binaryData = res.data?.binaryData || {}
    const merged = [
      ...Object.entries(data).map(([key, value]) => ({ key, value: String(value ?? '') })),
      ...Object.entries(binaryData).map(([key, value]) => ({ key, value: String(value ?? '') })),
    ]
    dataEntries.value = merged
  } catch (e: any) {
    ElMessage.error(e?.message || t('config.loadDataFailed'))
    dataDialogVisible.value = false
  } finally {
    dataLoading.value = false
  }
}

function handleDetail(row: any) {
  router.push(`/config/configmaps/${row.namespace}/${row.name}`)
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(
      t('config.deleteConfigMapConfirm', { name: row.name, namespace: row.namespace }),
      t('common.confirm'),
      { type: 'warning' },
    )
    await deleteConfigMap({ name: row.name, namespace: row.namespace })
    ElMessage.success(t('common.deleteSuccess'))
    fetchConfigMaps()
  } catch {
    /* cancelled */
  }
}

async function handleBatchDelete() {
  if (!selectedRows.value.length) return
  try {
    await ElMessageBox.confirm(
      t('common.batchDeleteConfirm', {
        count: selectedRows.value.length,
        type: t('config.configmap'),
      }),
      t('common.confirm'),
      { type: 'warning' },
    )
    const results = await Promise.allSettled(
      selectedRows.value.map((row: any) =>
        deleteConfigMap({ name: row.name, namespace: row.namespace }),
      ),
    )
    const count = results.filter((r) => r.status === 'fulfilled').length
    const failed = results.length - count
    ElMessage.success(
      t('config.batchDeleteResult', {
        count,
        type: t('config.configmap'),
        failed: failed ? t('common.batchDeletePartialFailed', { success: count, failed }) : '',
      }),
    )
    fetchConfigMaps()
  } catch {
    /* cancelled */
  }
}

const {
  isRunning,
  countdown,
  currentInterval,
  availableIntervals,
  toggle,
  refresh,
  setIntervalOption,
} = useAutoRefresh(fetchConfigMaps)

onMounted(() => {
  fetchNamespaces()
  fetchConfigMaps()
})
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
      resource-type="configmap"
      :label-conditions="labelConditions"
      @search-input="(val: string) => (searchName = val)"
      @namespace-change="handleNamespaceChange"
      @label-selector-change="onLabelConditionsChange"
    >
      <template #actions>
        <el-button type="success" @click="router.push('/config/configmaps/create')">
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
        <el-table-column :label="t('common.labels')" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <template v-if="row.labels && Object.keys(row.labels).length">
              <el-tag
                v-for="(val, key, idx) in row.labels"
                v-show="idx < 3"
                :key="key"
                size="small"
                class="label-tag"
              >
                {{ key }}={{ val }}
              </el-tag>
              <el-tag
                v-if="Object.keys(row.labels).length > 3"
                size="small"
                type="info"
                effect="plain"
              >
                +{{ Object.keys(row.labels).length - 3 }}
              </el-tag>
            </template>
            <span v-else class="no-labels">-</span>
          </template>
        </el-table-column>
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
              <el-button size="small" type="danger" @click="handleDelete(row)">{{
                t('common.delete')
              }}</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <YamlDrawer
      v-model="yamlDialogVisible"
      :get-yaml="getConfigMapYaml"
      :update-yaml="updateConfigMap"
      :namespace="yamlTarget?.namespace || ''"
      :name="yamlTarget?.name || ''"
      title="ConfigMap YAML"
      @saved="fetchConfigMaps"
    />
    <el-drawer
      v-model="dataDialogVisible"
      :title="dataDialogTitle"
      size="85%"
      direction="rtl"
      class="data-drawer"
      :body-style="{ padding: '0', height: '100%' }"
      :destroy-on-close="true"
    >
      <div v-loading="dataLoading" style="height: calc(100dvh - 52px)">
        <ConfigDataViewer :entries="dataEntries" :loading="dataLoading" />
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
.action-buttons {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: var(--gk-space-1);
}
.action-buttons .el-button + .el-button {
  margin-left: 0;
}
.label-tag {
  margin-right: 4px;
  max-width: 160px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.no-labels {
  color: var(--gk-color-text-placeholder);
  font-size: 12px;
}
</style>

<style>
.data-drawer .el-drawer__header {
  padding: 6px 16px;
  margin-bottom: 0;
  min-height: auto;
}
</style>
