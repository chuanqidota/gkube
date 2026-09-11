<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'
import { getCrdList, deleteCrd, getCrdYaml, updateCrd } from '@/api/resource'
import YamlDrawer from '@/components/YamlDrawer.vue'
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import { useClusterStore } from '@/stores/cluster'
import type { LabelCondition } from '@/components/LabelFilterPopover.vue'

const router = useRouter()
const { t } = useI18n()
const clusterStore = useClusterStore()
const loading = ref(false)
const crdList = ref<any[]>([])
const searchName = ref('')
const labelConditions = ref<LabelCondition[]>([])
const selectedRows = ref<any[]>([])
const yamlDialogVisible = ref(false)
const yamlTarget = ref<{ name: string } | null>(null)

const filteredList = computed(() => {
  if (!searchName.value) return crdList.value
  const keyword = searchName.value.toLowerCase()
  return crdList.value.filter(
    (d) => d.name?.toLowerCase().includes(keyword) || d.kind?.toLowerCase().includes(keyword),
  )
})

function onSearchInput(value: string) {
  searchName.value = value
}

function onLabelConditionsChange(conditions: LabelCondition[]) {
  labelConditions.value = conditions
  fetchCrds()
}

function handleSelectionChange(rows: any[]) {
  selectedRows.value = rows
}

async function fetchCrds() {
  loading.value = true
  try {
    const params: any = {}
    if (labelConditions.value.length > 0) params.labelFilters = labelConditions.value
    const res: any = await getCrdList(params)
    crdList.value = res.data || []
  } catch {
    // Silently handle — resource may not exist in cluster
  } finally {
    loading.value = false
  }
}

function handleViewYaml(row: any) {
  yamlTarget.value = { name: row.name }
  yamlDialogVisible.value = true
}

function handleBrowse(row: any) {
  const group = row.group
  const version = row.versions?.[0] || 'v1'
  const resource = row.plural
  router.push(
    `/crd/resources?group=${group}&version=${version}&resource=${resource}&scope=${row.scope}&kind=${row.kind}`,
  )
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(
      `删除 CRD "${row.name}"？这将同时删除该类型的所有自定义资源！`,
      '确认删除',
      { type: 'error' },
    )
    await deleteCrd({ name: row.name })
    ElMessage.success(t('crd.crdDeleted'))
    fetchCrds()
  } catch {
    /* cancelled */
  }
}

async function handleBatchDelete() {
  if (!selectedRows.value.length) return
  try {
    await ElMessageBox.confirm(
      `删除选中的 ${selectedRows.value.length} 个 CRD？这将同时删除对应类型的所有自定义资源！`,
      '确认删除',
      { type: 'error' },
    )
    const results = await Promise.allSettled(
      selectedRows.value.map((row) => deleteCrd({ name: row.name })),
    )
    const successCount = results.filter((r) => r.status === 'fulfilled').length
    const failCount = results.filter((r) => r.status === 'rejected').length
    if (failCount > 0) {
      ElMessage.warning(t('crd.batchDeleteResult', { success: successCount, failed: failCount }))
    } else {
      ElMessage.success(t('crd.batchDeleteSuccess', { count: successCount }))
    }
    fetchCrds()
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
  refresh: manualRefresh,
  setIntervalOption,
} = useAutoRefresh(fetchCrds)

onMounted(fetchCrds)
</script>

<template>
  <div class="page-container">
    <ResourceListToolbar
      :search-value="searchName"
      :total-count="crdList.length"
      :selected-count="selectedRows.length"
      :show-namespace="false"
      search-placeholder="搜索名称或 Kind"
      :cluster-name="clusterStore.clusterName"
      resource-type="customresourcedefinition"
      :label-conditions="labelConditions"
      @search-input="onSearchInput"
      @label-selector-change="onLabelConditionsChange"
    >
      <template #actions>
        <el-button type="success" @click="router.push('/crd/create')">
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
        v-loading="loading"
        :data="filteredList"
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="45" />
        <el-table-column prop="kind" label="Kind" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <el-button link type="primary" @click="handleBrowse(row)">{{ row.kind }}</el-button>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="名称" min-width="280" show-overflow-tooltip>
          <template #default="{ row }">
            <el-button link type="primary" @click="router.push(`/crd/detail?name=${row.name}`)">{{
              row.name
            }}</el-button>
          </template>
        </el-table-column>
        <el-table-column prop="group" label="API 组" min-width="180" show-overflow-tooltip />
        <el-table-column label="版本" width="140">
          <template #default="{ row }">
            <el-tag
              v-for="v in row.versions || []"
              :key="v"
              size="small"
              style="margin-right: 4px"
              >{{ v }}</el-tag
            >
          </template>
        </el-table-column>
        <el-table-column prop="scope" label="作用域" width="120">
          <template #default="{ row }">
            <el-tag :type="row.scope === 'Namespaced' ? 'info' : 'warning'" size="small">
              {{ row.scope === 'Namespaced' ? '命名空间' : '集群' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="age" label="Age" width="180" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
              <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <YamlDrawer
      v-model="yamlDialogVisible"
      :get-yaml="getCrdYaml"
      :update-yaml="updateCrd"
      :name="yamlTarget?.name || ''"
      title="CRD YAML"
      @saved="fetchCrds"
    />
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
