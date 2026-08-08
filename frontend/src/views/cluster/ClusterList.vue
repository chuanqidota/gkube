<script setup lang="ts">
import { ref, onMounted, onUnmounted, reactive, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Edit, CircleCheck } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'
import { getClusterList, deleteCluster, checkCluster, updateCluster } from '@/api/cluster'
import { useClusterStore } from '@/stores/cluster'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'
import ViewModeToggle from '@/components/ViewModeToggle.vue'
const { t } = useI18n()
const router = useRouter()
const clusterStore = useClusterStore()
const loading = ref(false)
const clusterList = ref<any[]>([])
const searchName = ref('')
const total = ref(0)
const page = ref(1)
const size = ref(10)

const storedView = localStorage.getItem('gkube.cluster.viewMode')
const viewMode = ref<'card' | 'table'>(storedView === 'table' || storedView === 'card' ? storedView : 'card')

// 编辑对话框相关
const editVisible = ref(false)
const editLoading = ref(false)
const editClusterId = ref(0)
const editFormRef = ref<FormInstance>()
const editForm = reactive({
  displayName: '',
  description: '',
  labels: [] as Array<{ key: string; value: string }>,
})

const editRules: FormRules = {
  displayName: [{ max: 200, message: t('cluster.displayNameMax'), trigger: 'blur' }],
  description: [{ max: 500, message: t('cluster.descriptionMax'), trigger: 'blur' }],
}

let searchDebounce: ReturnType<typeof setTimeout> | undefined

async function fetchClusters() {
  loading.value = true
  try {
    const params: Record<string, unknown> = { page: page.value, size: size.value }
    if (searchName.value) params.keyword = searchName.value
    const res: any = await getClusterList(params)
    clusterList.value = res.data.items || []
    total.value = res.data.total || 0
  } catch (e: any) {
    ElMessage.error(e?.message || t('cluster.loadFailed'))
  } finally {
    loading.value = false
  }
}

// Search triggers server-side fetch with debounce
watch(searchName, () => {
  page.value = 1
  clearTimeout(searchDebounce)
  searchDebounce = setTimeout(fetchClusters, 300)
})

async function handleCheck(row: any) {
  try {
    const res: any = await checkCluster(row.id)
    const info = res.data
    ElMessage.success(
      t('cluster.connectedSuccess', {
        version: info.clusterVersion,
        nodeCount: info.nodeCount,
        responseTimeMs: info.responseTimeMs,
      })
    )
    fetchClusters()
  } catch (e: any) {
    ElMessage.error(e?.message || t('cluster.checkFailed'))
  }
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(
      t('cluster.deleteClusterConfirm', { name: row.displayName || row.clusterName }),
      t('common.confirm'),
      { type: 'warning' }
    )
  } catch {
    return // user cancelled
  }
  try {
    await deleteCluster(row.id)
    ElMessage.success(t('cluster.deleted'))
    fetchClusters()
    clusterStore.fetchClusters()
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.deleteFailed'))
  }
}

function handleEdit(row: any) {
  editClusterId.value = row.id
  editForm.displayName = row.displayName || ''
  editForm.description = row.description || ''
  // 解析 labels JSON 字符串为 key-value 数组
  editForm.labels = []
  if (row.labels) {
    try {
      const parsed = typeof row.labels === 'string' ? JSON.parse(row.labels) : row.labels
      Object.entries(parsed).forEach(([key, value]) => {
        editForm.labels.push({ key, value: value as string })
      })
    } catch {
      // ignore parse errors
    }
  }
  editVisible.value = true
}

function addEditLabel() {
  editForm.labels.push({ key: '', value: '' })
}

function removeEditLabel(index: number) {
  editForm.labels.splice(index, 1)
}

async function handleEditSubmit() {
  // Validate form
  const valid = await editFormRef.value?.validate().catch(() => false)
  if (!valid) return

  // Validate labels: key is required when a label row is added
  for (const l of editForm.labels) {
    if (l.key.trim() === '' && l.value.trim() !== '') {
      ElMessage.warning(t('cluster.labelKeyRequired'))
      return
    }
  }

  editLoading.value = true
  try {
    const labels: Record<string, string> = {}
    editForm.labels.forEach((l) => {
      if (l.key.trim()) labels[l.key.trim()] = l.value.trim()
    })

    await updateCluster(editClusterId.value, {
      displayName: editForm.displayName,
      description: editForm.description,
      labels,
    })
    ElMessage.success(t('common.saveSuccess'))
    editVisible.value = false
    fetchClusters()
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.saveFailed'))
  } finally {
    editLoading.value = false
  }
}

function handlePageChange(newPage: number) {
  page.value = newPage
  fetchClusters()
}

function statusType(status: string) {
  if (status === 'online' || status === 'connected' || status === 'healthy') return 'success'
  if (status === 'offline' || status === 'disconnected' || status === 'unhealthy') return 'danger'
  return 'info'
}

function statusText(status: string) {
  if (status === 'online' || status === 'connected') return t('cluster.online')
  if (status === 'offline' || status === 'disconnected') return t('cluster.offline')
  return status || t('common.unknown')
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchClusters)

onMounted(fetchClusters)
onUnmounted(() => clearTimeout(searchDebounce))
</script>

<template>
  <div class="page-container">
    <ResourceListToolbar
      :search-value="searchName"
      :total-count="total"
      :show-namespace="false"
      search-placeholder="搜索集群名称"
      @search-input="searchName = $event"
    >
      <template #actions>
        <el-button type="success" @click="router.push('/clusters/create')">
          <el-icon><Plus /></el-icon> {{ t('cluster.add') }}
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
        <ViewModeToggle v-model="viewMode" storage-key="gkube.cluster.viewMode" />
      </template>
    </ResourceListToolbar>

    <el-card shadow="never" class="table-card">
    <el-table v-if="viewMode === 'table'" :data="clusterList" v-loading="loading" stripe>
      <el-table-column label="状态" width="100" align="center">
        <template #default="{ row }"><el-tag :type="statusType(row.status)" size="small">{{ statusText(row.status) }}</el-tag></template>
      </el-table-column>
      <el-table-column label="显示名称" min-width="160" show-overflow-tooltip>
        <template #default="{ row }">{{ row.displayName || row.clusterName }}</template>
      </el-table-column>
      <el-table-column prop="clusterName" label="集群名称" min-width="160" show-overflow-tooltip>
        <template #default="{ row }">{{ row.clusterName || '-' }}</template>
      </el-table-column>
      <el-table-column label="版本" width="130" show-overflow-tooltip>
        <template #default="{ row }">{{ row.clusterVersion || '-' }}</template>
      </el-table-column>
      <el-table-column label="节点数" width="90" align="center">
        <template #default="{ row }">{{ row.nodeCount || 0 }}</template>
      </el-table-column>
      <el-table-column label="描述" min-width="180" show-overflow-tooltip>
        <template #default="{ row }">{{ row.description || '-' }}</template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right" align="center">
        <template #default="{ row }">
          <div class="action-buttons">
            <el-button size="small" @click="handleCheck(row)">{{ t('cluster.checkConnection') }}</el-button>
            <el-button size="small" @click="handleEdit(row)">{{ t('common.edit') }}</el-button>
            <el-button size="small" type="danger" plain @click="handleDelete(row)">{{ t('common.delete') }}</el-button>
          </div>
        </template>
      </el-table-column>
    </el-table>

    <el-row :gutter="16" v-else-if="viewMode === 'card' && clusterList.length > 0">
      <el-col :span="8" v-for="cluster in clusterList" :key="cluster.id" style="margin-bottom: 16px;">
        <el-card shadow="hover" class="cluster-card">
          <template #header>
            <div class="cluster-header">
              <div class="cluster-info">
                <h4 style="margin: 0;">{{ cluster.displayName || cluster.clusterName }}</h4>
                <el-tag :type="statusType(cluster.status)" size="small">{{ statusText(cluster.status) }}</el-tag>
              </div>
            </div>
          </template>
          <div class="cluster-body">
            <div class="cluster-detail">
              <span class="label">{{ t('cluster.name') }}:</span>
              <span class="value">{{ cluster.clusterName }}</span>
            </div>
            <div class="cluster-detail">
              <span class="label">{{ t('cluster.version') }}:</span>
              <span class="value">{{ cluster.clusterVersion || '-' }}</span>
            </div>
            <div class="cluster-detail">
              <span class="label">{{ t('cluster.nodes') }}:</span>
              <span class="value">{{ cluster.nodeCount || 0 }}</span>
            </div>
            <div class="cluster-detail" v-if="cluster.description">
              <span class="label">{{ t('cluster.description') }}:</span>
              <span class="value">{{ cluster.description }}</span>
            </div>
          </div>
          <div class="cluster-footer">
            <el-button size="small" @click="handleCheck(cluster)">{{ t('cluster.checkConnection') }}</el-button>
            <el-button size="small" @click="handleEdit(cluster)">{{ t('common.edit') }}</el-button>
            <el-button size="small" type="danger" plain @click="handleDelete(cluster)">{{ t('common.delete') }}</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-empty v-if="!loading && clusterList.length === 0" :description="searchName ? t('cluster.noSearchResults') : t('cluster.noClusters')">
      <el-button type="primary" @click="router.push('/clusters/create')"><el-icon><Plus /></el-icon> {{ t('cluster.add') }}</el-button>
    </el-empty>
    </el-card>

    <div style="display: flex; justify-content: flex-end; margin-top: 16px;">
      <el-pagination
        v-if="total > size"
        :current-page="page"
        :page-size="size"
        :total="total"
        layout="prev, pager, next"
        @current-change="handlePageChange"
      />
    </div>

    <!-- 编辑集群对话框 -->
    <el-dialog v-model="editVisible" :title="t('cluster.edit')" width="560px" destroy-on-close>
      <el-form ref="editFormRef" :model="editForm" :rules="editRules" label-width="100px">
        <el-form-item :label="t('cluster.displayName')">
          <el-input v-model="editForm.displayName" :placeholder="t('cluster.displayNamePlaceholder')" />
        </el-form-item>
        <el-form-item :label="t('cluster.description')">
          <el-input v-model="editForm.description" type="textarea" :rows="3" :placeholder="t('cluster.descriptionPlaceholder')" />
        </el-form-item>
        <el-form-item :label="t('cluster.labels')">
          <div style="width: 100%;">
            <div
              v-for="(label, index) in editForm.labels"
              :key="index"
              style="display: flex; gap: 8px; margin-bottom: 8px;"
            >
              <el-input v-model="label.key" :placeholder="t('cluster.keyPlaceholder')" style="flex: 1;" />
              <el-input v-model="label.value" :placeholder="t('cluster.valuePlaceholder')" style="flex: 1;" />
              <el-button type="danger" circle @click="removeEditLabel(index)">-</el-button>
            </div>
            <el-button @click="addEditLabel" type="primary" plain>{{ t('cluster.addLabel') }}</el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="editLoading" @click="handleEditSubmit">{{ t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.table-card { border-radius: 8px; }
.cluster-card {
  height: 100%;
  background: linear-gradient(180deg, var(--gk-color-primary-bg) 0%, var(--gk-color-bg-card) 60%);
  border-color: var(--gk-color-primary-light);
}
.cluster-header { display: flex; justify-content: space-between; align-items: center; }
.cluster-info { display: flex; align-items: center; gap: 8px; }
.cluster-body { margin-bottom: 12px; }
.cluster-detail { display: flex; margin-bottom: 8px; }
.cluster-detail .label { color: var(--gk-color-text-secondary); width: 70px; flex-shrink: 0; }
.cluster-detail .value { color: var(--gk-color-text-primary); }
.cluster-footer { display: flex; flex-wrap: nowrap; align-items: center; gap: 4px; border-top: 1px solid var(--gk-color-border-light); padding-top: 12px; }
.cluster-footer .el-button { margin-left: 0 !important; padding: 5px 8px; font-size: 12px; height: auto; }
</style>
