<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Edit, CircleCheck } from '@element-plus/icons-vue'
import { getClusterList, deleteCluster, checkCluster, updateCluster } from '@/api/cluster'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
const { t } = useI18n()
const router = useRouter()
const loading = ref(false)
const clusterList = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const size = ref(10)

// 编辑对话框相关
const editVisible = ref(false)
const editLoading = ref(false)
const editClusterId = ref(0)
const editForm = reactive({
  displayName: '',
  description: '',
  labels: [] as Array<{ key: string; value: string }>,
})

async function fetchClusters() {
  loading.value = true
  try {
    const res: any = await getClusterList({ page: page.value, size: size.value })
    clusterList.value = res.data.items || []
    total.value = res.data.total || 0
  } catch (e: any) {
    ElMessage.error(e?.message || t('cluster.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function handleCheck(row: any) {
  try {
    const res: any = await checkCluster(row.id)
    const info = res.data
    if (info.status === 'online') {
      ElMessage.success(
        t('cluster.connectedSuccess', {
          version: info.clusterVersion,
          nodeCount: info.nodeCount,
          responseTimeMs: info.responseTimeMs,
        })
      )
    } else {
      ElMessage.warning(info.message || t('cluster.connectionFailed'))
    }
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
  editLoading.value = true
  try {
    const labels: Record<string, string> = {}
    editForm.labels.forEach((l) => {
      if (l.key.trim()) labels[l.key.trim()] = l.value
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
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <h3 style="margin: 0;">{{ t('cluster.clusterManagement') }}</h3>
        <div class="filter-right">
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
          <el-button type="primary" @click="router.push('/clusters/create')"><el-icon><Plus /></el-icon> {{ t('cluster.add') }}</el-button>
        </div>
      </div>
    </el-card>

    <el-row :gutter="16" v-if="clusterList.length > 0">
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
            <el-button size="small" @click="handleCheck(cluster)"><el-icon><CircleCheck /></el-icon> {{ t('cluster.checkConnection') }}</el-button>
            <el-button size="small" @click="handleEdit(cluster)"><el-icon><Edit /></el-icon> {{ t('common.edit') }}</el-button>
            <el-button size="small" type="danger" @click="handleDelete(cluster)"><el-icon><Delete /></el-icon> {{ t('common.delete') }}</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-empty v-if="!loading && clusterList.length === 0" :description="t('cluster.noClusters')">
      <el-button type="primary" @click="router.push('/clusters/create')"><el-icon><Plus /></el-icon> {{ t('cluster.add') }}</el-button>
    </el-empty>

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
      <el-form :model="editForm" label-width="100px">
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
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; justify-content: space-between; align-items: center; }
.filter-right { display: flex; align-items: center; gap: 8px; }
.cluster-card { height: 100%; }
.cluster-header { display: flex; justify-content: space-between; align-items: center; }
.cluster-info { display: flex; align-items: center; gap: 8px; }
.cluster-body { margin-bottom: 12px; }
.cluster-detail { display: flex; margin-bottom: 8px; }
.cluster-detail .label { color: var(--gk-color-text-secondary); width: 70px; flex-shrink: 0; }
.cluster-detail .value { color: var(--gk-color-text-primary); }
.cluster-footer { display: flex; justify-content: flex-end; gap: 8px; border-top: 1px solid var(--gk-color-border-light); padding-top: 12px; }
</style>
