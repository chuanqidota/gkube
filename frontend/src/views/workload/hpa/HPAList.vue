<script setup lang="ts">
import { ref, computed } from 'vue'
import { Plus, Delete, FullScreen, Aim } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getHpaList,
  getHpaDetail,
  getHpaYaml,
  updateHpa,
  deleteHpa,
  pauseHpa,
  resumeHpa,
} from '@/api/resource'
import { useResourceList } from '@/composables/useResourceList'
import YamlEditor from '@/components/YamlEditor.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import HPAForm from './components/HPAForm.vue'

const {
  loading,
  list,
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
} = useResourceList({
  resourceName: 'HPA',
  fetchList: getHpaList,
  getYaml: getHpaYaml,
  updateYaml: updateHpa,
  deleteResource: deleteHpa,
  detailRoute: '/autoscaling/hpa',
  createRoute: '/autoscaling/hpa/create',
  paginated: true,
  pageSize: 50,
  autoRefreshInterval: 30000,
})

// Status filter
const statusFilter = ref('')

// Edit drawer
const editDrawerVisible = ref(false)
const editFullscreen = ref(false)
const editRow = ref<any>(null)

// Status helpers — now reads from backend-provided fields
function getStatus(row: any): 'paused' | 'active' | 'inactive' | 'orphan' {
  if (row.paused) return 'paused'
  if (row.target_exists === false) return 'orphan'
  const conditions = row.conditions || []
  const scalingActive = conditions.find((c: any) => c.type === 'ScalingActive')
  if (scalingActive?.status === 'True') return 'active'
  return 'inactive'
}

function getStatusText(row: any): string {
  const s = getStatus(row)
  if (s === 'paused') return '已暂停'
  if (s === 'active') return '正常'
  if (s === 'inactive') return '未激活'
  return '孤立'
}

function getStatusType(row: any): string {
  const s = getStatus(row)
  if (s === 'paused') return 'warning'
  if (s === 'active') return 'success'
  if (s === 'inactive') return 'danger'
  return 'warning'
}

// Metrics helpers
function getCpuTarget(row: any): string | null {
  const metrics = row.metrics || []
  for (const m of metrics) {
    if (m.type === 'Resource' && m.resource?.name === 'cpu') {
      return String(m.resource?.target?.averageUtilization ?? '')
    }
  }
  return null
}

function getCpuCurrent(row: any): string | null {
  const currentMetrics = row.current_metrics || []
  for (const m of currentMetrics) {
    if (m.type === 'Resource' && m.resource?.name === 'cpu') {
      return String(m.resource?.current?.averageUtilization ?? '')
    }
  }
  return null
}

// Filtered list with status filter
const displayList = computed(() => {
  if (!statusFilter.value) return filteredList.value
  return filteredList.value.filter((row: any) => getStatus(row) === statusFilter.value)
})

const { isRunning: arRunning, countdown: arCountdown, currentInterval: arInterval, availableIntervals: arIntervals, toggle: arToggle, refresh: arRefresh, setIntervalOption: arSetInterval } = useAutoRefresh(fetchResources, { interval: 30000 })

// Workload route mapping
function getWorkloadRoute(row: any): string | null {
  const kind = row.target_kind
  const ns = row.namespace
  const targetName = row.target
  if (!kind || !targetName || !ns) return null

  if (kind === 'Deployment') return `/workloads/deployments/${ns}/${targetName}`
  if (kind === 'StatefulSet') return `/workloads/statefulsets/${ns}/${targetName}`
  if (kind === 'DaemonSet') return `/workloads/daemonsets/${ns}/${targetName}`
  if (kind === 'ReplicaSet') return `/workloads/replicasets/${ns}/${targetName}`
  return null
}

// Edit drawer — fetch full K8s object (list returns flat data, form needs nested metadata/spec/status)
async function handleEdit(row: any) {
  try {
    const res: any = await getHpaDetail({ namespace: row.namespace, name: row.name })
    editRow.value = res.data
    editDrawerVisible.value = true
  } catch (e: any) {
    ElMessage.error(e?.message || '获取 HPA 详情失败')
  }
}

function handleEditSuccess() {
  editDrawerVisible.value = false
  editRow.value = null
  fetchResources()
}

function handleEditCancel() {
  editDrawerVisible.value = false
  editRow.value = null
}

// Pause / Resume
async function handlePause(row: any) {
  try {
    await ElMessageBox.confirm(
      `暂停后 HPA 将停止自动伸缩，副本数固定在当前值（${row.current_replicas}）。确定暂停吗？`,
      '确认暂停',
      { type: 'warning', confirmButtonText: '暂停', cancelButtonText: '取消' }
    )
    await pauseHpa({ namespace: row.namespace, name: row.name })
    ElMessage.success('HPA 已暂停')
    fetchResources()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || '暂停失败')
    }
  }
}

async function handleResume(row: any) {
  try {
    await resumeHpa({ namespace: row.namespace, name: row.name })
    ElMessage.success('HPA 已恢复')
    fetchResources()
  } catch (e: any) {
    ElMessage.error(e?.message || '恢复失败')
  }
}
</script>

<template>
  <div class="page-container">
    <ResourceListToolbar
      :search-value="searchName"
      v-model:namespace-value="selectedNamespace"
      :namespace-list="namespaceList"
      :total-count="totalCount"
      :selected-count="selectedRows.length"
      @search-input="onSearchInput"
      @namespace-change="handleNamespaceChange"
    >
      <template #actions>
        <el-select
          v-model="statusFilter"
          placeholder="所有状态"
          clearable
          style="width: 130px;"
        >
          <el-option label="正常" value="active" />
          <el-option label="已暂停" value="paused" />
          <el-option label="未激活" value="inactive" />
          <el-option label="孤立" value="orphan" />
        </el-select>
        <el-button type="success" @click="$router.push('/autoscaling/hpa/create')">
          <el-icon><Plus /></el-icon> 创建
        </el-button>
        <el-button type="danger" :disabled="!selectedRows.length" @click="handleBatchDelete">
          <el-icon><Delete /></el-icon> 删除 ({{ selectedRows.length }})
        </el-button>
      </template>
      <template #extra>
        <AutoRefreshToolbar
          :is-running="arRunning"
          :countdown="arCountdown"
          :current-interval="arInterval"
          :available-intervals="arIntervals"
          :loading="loading"
          @refresh="arRefresh()"
          @toggle="arToggle()"
          @interval-change="arSetInterval"
        />
      </template>
    </ResourceListToolbar>

    <el-card shadow="never" class="table-card">
      <el-table
        :data="displayList"
        v-loading="loading"
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="45" />
        <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <el-button link type="primary" @click="handleDetail(row)">{{ row.name }}</el-button>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="140" />
        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row)" effect="dark" size="small">{{ getStatusText(row) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="伸缩目标" min-width="180">
          <template #default="{ row }">
            <el-button
              v-if="getWorkloadRoute(row)"
              link
              type="primary"
              @click="$router.push(getWorkloadRoute(row)!)"
            >{{ row.target_kind }}/{{ row.target }}</el-button>
            <span v-else>{{ row.target_kind }}/{{ row.target }}</span>
          </template>
        </el-table-column>
        <el-table-column label="指标" width="140">
          <template #default="{ row }">
            <span v-if="getCpuTarget(row)">
              CPU {{ getCpuTarget(row) }}%<template v-if="getCpuCurrent(row)">/{{ getCpuCurrent(row) }}%</template>
            </span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="副本数" width="180">
          <template #default="{ row }">{{ row.current_replicas }} ({{ row.min_replicas }}-{{ row.max_replicas }})</template>
        </el-table-column>
        <el-table-column prop="age" label="Age" width="120" />
        <el-table-column label="操作" width="320" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" type="info" @click="handleEdit(row)">编辑</el-button>
              <el-button
                v-if="row.paused"
                size="small"
                type="success"
                plain
                @click="handleResume(row)"
              >恢复</el-button>
              <el-button
                v-else
                size="small"
                type="warning"
                plain
                @click="handlePause(row)"
              >暂停</el-button>
              <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
              <el-button size="small" type="danger" plain @click="handleDelete(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无 HPA 配置">
            <el-button type="success" @click="$router.push('/autoscaling/hpa/create')">
              <el-icon><Plus /></el-icon> 创建 HPA
            </el-button>
          </el-empty>
        </template>
      </el-table>
      <div v-if="hasMore" class="load-more">
        <el-button @click="fetchNextPage" :loading="loading" link type="primary">
          Load More...
        </el-button>
      </div>
    </el-card>

    <!-- YAML Drawer -->
    <el-drawer v-model="yamlDialogVisible" title="HPA YAML" size="85%" direction="rtl" class="yaml-drawer"
      :body-style="{ padding: '0', height: '100%' }">
      <div v-loading="yamlLoading" style="height: calc(100vh - 60px);">
        <YamlEditor v-model="yamlContent" height="100%" auto-format show-save-buttons :saving="yamlSaving" @save="handleSaveYaml" @cancel="handleCancelYaml" />
      </div>
    </el-drawer>

    <!-- Edit Drawer -->
    <el-drawer
      v-model="editDrawerVisible"
      title="编辑 HPA"
      :size="editFullscreen ? '100%' : '85%'"
      direction="rtl"
      destroy-on-close
      :body-style="{ padding: '0', height: '100%' }"
    >
      <template #header>
        <div class="drawer-header">
          <span class="drawer-title">编辑 HPA</span>
          <el-button text @click="editFullscreen = !editFullscreen">
            <el-icon>
              <FullScreen v-if="!editFullscreen" />
              <Aim v-else />
            </el-icon>
          </el-button>
        </div>
      </template>
      <div v-if="editDrawerVisible && editRow" style="height: 100%;">
        <HPAForm
          :is-edit="true"
          :initial-data="editRow"
          @success="handleEditSuccess"
          @cancel="handleEditCancel"
        />
      </div>
    </el-drawer>
  </div>
</template>

<style scoped>
.page-container {
  padding: 20px;
}
.table-card {
  border-radius: 8px;
}
.text-muted {
  color: var(--el-text-color-placeholder);
}
.load-more {
  text-align: center;
  padding: 12px 0;
}
.drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 16px;
  margin-bottom: 0;
  min-height: auto;
}
.drawer-title {
  font-size: 16px;
  font-weight: 600;
}
</style>

<style>
.yaml-drawer .el-drawer__header {
  padding: 6px 16px;
  margin-bottom: 0;
  min-height: auto;
}
</style>
