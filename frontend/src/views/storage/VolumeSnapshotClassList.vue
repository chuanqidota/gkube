<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'
import {
  getVolumeSnapshotClassList,
  deleteVolumeSnapshotClass,
  getVolumeSnapshotClassYaml,
  updateVolumeSnapshotClass,
} from '@/api/resource'
import { useI18n } from 'vue-i18n'
import YamlDrawer from '@/components/YamlDrawer.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'
import { useClusterStore } from '@/stores/cluster'
import type { LabelCondition } from '@/components/LabelFilterPopover.vue'

const { t } = useI18n()
const router = useRouter()
const clusterStore = useClusterStore()
const loading = ref(false)
const classList = ref<any[]>([])
const searchName = ref('')
const labelConditions = ref<LabelCondition[]>([])
const selectedRows = ref<any[]>([])
const yamlDialogVisible = ref(false)
const yamlTarget = ref<{ name: string } | null>(null)

const filteredList = computed(() => {
  if (!searchName.value) return classList.value
  const keyword = searchName.value.toLowerCase()
  return classList.value.filter((d) => d.name?.toLowerCase().includes(keyword))
})

function onSearchInput(val: string) {
  searchName.value = val
}

function onLabelConditionsChange(conditions: LabelCondition[]) {
  labelConditions.value = conditions
  fetchClasses()
}

async function fetchClasses() {
  loading.value = true
  try {
    const params: any = {}
    if (labelConditions.value.length > 0) params.labelFilters = labelConditions.value
    const res: any = await getVolumeSnapshotClassList(params)
    classList.value = res.data || []
  } catch {
    // Silently handle — resource may not exist in cluster
  } finally {
    loading.value = false
  }
}

function handleSelectionChange(rows: any[]) {
  selectedRows.value = rows
}

function handleViewYaml(row: any) {
  yamlTarget.value = { name: row.name }
  yamlDialogVisible.value = true
}

function handleDetail(row: any) {
  router.push(`/storage/volumesnapshotclasses/${row.name}`)
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(
      t('storage.deleteSnapshotClassConfirm', { name: row.name }),
      t('common.confirm'),
      { type: 'warning' },
    )
    await deleteVolumeSnapshotClass({ name: row.name })
    ElMessage.success(t('common.delete') + ' ' + t('common.success'))
    fetchClasses()
  } catch {
    /* cancelled */
  }
}

async function handleBatchDelete() {
  if (!selectedRows.value.length) return
  try {
    await ElMessageBox.confirm(
      t('storage.deleteSnapshotClassBatchConfirm', { count: selectedRows.value.length }),
      t('common.confirm'),
      { type: 'warning' },
    )
    let count = 0
    for (const row of selectedRows.value) {
      try {
        await deleteVolumeSnapshotClass({ name: row.name })
        count++
      } catch {
        /* continue */
      }
    }
    ElMessage.success(t('common.delete') + ` ${count} ` + t('storage.volumeSnapshotClass'))
    fetchClasses()
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
} = useAutoRefresh(fetchClasses)

onMounted(fetchClasses)
</script>

<template>
  <div class="page-container">
    <ResourceListToolbar
      :search-value="searchName"
      :show-namespace="false"
      :show-total-count="false"
      :selected-count="selectedRows.length"
      :cluster-name="clusterStore.clusterName"
      resource-type="volumesnapshotclass"
      :label-conditions="labelConditions"
      @search-input="onSearchInput"
      @label-selector-change="onLabelConditionsChange"
    >
      <template #actions>
        <el-button type="success" @click="router.push('/storage/volumesnapshotclasses/create')"
          ><el-icon><Plus /></el-icon> {{ t('common.create') }}</el-button
        >
        <el-button type="danger" :disabled="!selectedRows.length" @click="handleBatchDelete"
          ><el-icon><Delete /></el-icon> {{ t('common.delete') }} ({{
            selectedRows.length
          }})</el-button
        >
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
          <template #default="{ row }"
            ><el-button link type="primary" @click="handleDetail(row)">{{
              row.name
            }}</el-button></template
          >
        </el-table-column>
        <el-table-column
          prop="driver"
          :label="t('storage.driver')"
          min-width="250"
          show-overflow-tooltip
        />
        <el-table-column prop="deletionPolicy" :label="t('storage.deletionPolicy')" width="160" />
        <el-table-column prop="age" :label="t('common.age')" width="120" />
        <el-table-column :label="t('common.actions')" width="180" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">{{ t('common.yaml') }}</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">{{
              t('common.delete')
            }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <YamlDrawer
      v-model="yamlDialogVisible"
      :get-yaml="getVolumeSnapshotClassYaml"
      :update-yaml="updateVolumeSnapshotClass"
      :name="yamlTarget?.name || ''"
      title="VolumeSnapshotClass YAML"
      @saved="fetchClasses"
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
</style>
