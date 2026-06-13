<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Delete, VideoPlay, VideoPause, Setting } from '@element-plus/icons-vue'
import request from '@/api/request'

const loading = ref(false)
const selectedResource = ref('pods')
const selectedNamespace = ref('')
const namespaces = ref<string[]>([])
const resources = ref<any[]>([])
const selectedResources = ref<string[]>([])
const showDeleteDialog = ref(false)
const showScaleDialog = ref(false)
const scaleReplicas = ref(1)

const resourceTypes = [
  { value: 'pods', label: 'Pods' },
  { value: 'deployments', label: 'Deployments' },
  { value: 'statefulsets', label: 'StatefulSets' },
  { value: 'daemonsets', label: 'DaemonSets' },
  { value: 'services', label: 'Services' },
  { value: 'configmaps', label: 'ConfigMaps' },
  { value: 'secrets', label: 'Secrets' },
]

async function fetchNamespaces() {
  try {
    const res: any = await request.get('/k8s/namespace/list')
    namespaces.value = res.data?.map((ns: any) => ns.name) || []
  } catch {
    namespaces.value = []
  }
}

async function fetchResources() {
  loading.value = true
  try {
    const params: any = {}
    if (selectedNamespace.value) {
      params.namespace = selectedNamespace.value
    }
    const res: any = await request.get(`/k8s/${selectedResource.value}/list`, { params })
    resources.value = res.data || []
    selectedResources.value = []
  } catch (e: any) {
    ElMessage.warning('Failed to load resources')
  } finally {
    loading.value = false
  }
}

function handleSelectionChange(selection: any[]) {
  selectedResources.value = selection.map(s => s.name)
}

async function batchDelete() {
  if (selectedResources.value.length === 0) {
    ElMessage.warning('Please select resources to delete')
    return
  }

  try {
    await ElMessageBox.confirm(
      `Delete ${selectedResources.value.length} selected ${selectedResource.value}?`,
      'Confirm Batch Delete',
      { type: 'warning' }
    )

    loading.value = true
    let successCount = 0
    let failCount = 0

    for (const name of selectedResources.value) {
      try {
        await request.delete(`/k8s/${selectedResource.value}/delete`, {
          params: {
            name,
            namespace: selectedNamespace.value || 'default',
          }
        })
        successCount++
      } catch {
        failCount++
      }
    }

    ElMessage.success(`Deleted: ${successCount}, Failed: ${failCount}`)
    fetchResources()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error('Batch delete failed')
    }
  } finally {
    loading.value = false
    showDeleteDialog.value = false
  }
}

async function batchScale() {
  if (selectedResources.value.length === 0) {
    ElMessage.warning('Please select deployments to scale')
    return
  }

  if (selectedResource.value !== 'deployments') {
    ElMessage.warning('Scale operation only supports deployments')
    return
  }

  try {
    loading.value = true
    let successCount = 0
    let failCount = 0

    for (const name of selectedResources.value) {
      try {
        await request.put(`/k8s/deployment/scale`, {
          name,
          namespace: selectedNamespace.value || 'default',
          replicas: scaleReplicas.value,
        })
        successCount++
      } catch {
        failCount++
      }
    }

    ElMessage.success(`Scaled: ${successCount}, Failed: ${failCount}`)
    fetchResources()
  } catch (e: any) {
    ElMessage.error('Batch scale failed')
  } finally {
    loading.value = false
    showScaleDialog.value = false
  }
}

function formatDate(date: string) {
  if (!date) return '-'
  return new Date(date).toLocaleString()
}

onMounted(() => {
  fetchNamespaces()
  fetchResources()
})
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <h3 style="margin: 0;"><el-icon><Setting /></el-icon> 批量操作</h3>
        <div class="filter-right">
          <el-select v-model="selectedResource" style="width: 150px;" @change="fetchResources">
            <el-option v-for="r in resourceTypes" :key="r.value" :label="r.label" :value="r.value" />
          </el-select>
          <el-select v-model="selectedNamespace" placeholder="所有命名空间" clearable style="width: 150px;" @change="fetchResources">
            <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
          </el-select>
          <el-button type="primary" @click="fetchResources"><el-icon><Refresh /></el-icon> 刷新</el-button>
        </div>
      </div>
    </el-card>

    <el-card shadow="never">
      <div style="margin-bottom: 16px; display: flex; gap: 8px;">
        <el-button type="danger" :disabled="selectedResources.length === 0" @click="showDeleteDialog = true">
          <el-icon><Delete /></el-icon> 批量删除 ({{ selectedResources.length }})
        </el-button>
        <el-button type="warning" :disabled="selectedResources.length === 0 || selectedResource !== 'deployments'" @click="showScaleDialog = true">
          <el-icon><VideoPlay }}</el-icon> 批量伸缩 ({{ selectedResources.length }})
        </el-button>
      </div>

      <el-table
        :data="resources"
        v-loading="loading"
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip />
        <el-table-column prop="namespace" label="命名空间" width="120" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'Running' || row.status === 'Active' ? 'success' : 'warning'" size="small">
              {{ row.status || '-' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="creationTimestamp" label="创建时间" width="180">
          <template #default="{ row }">{{ formatDate(row.creationTimestamp) }}</template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Delete Confirmation Dialog -->
    <el-dialog v-model="showDeleteDialog" title="批量删除" width="400px">
      <p>确定删除选中的 {{ selectedResources.length }} 个资源吗？</p>
      <p style="color: #F56C6C;">此操作不可恢复！</p>
      <template #footer>
        <el-button @click="showDeleteDialog = false">取消</el-button>
        <el-button type="danger" @click="batchDelete">确定删除</el-button>
      </template>
    </el-dialog>

    <!-- Scale Dialog -->
    <el-dialog v-model="showScaleDialog" title="批量伸缩" width="400px">
      <el-form label-width="100px">
        <el-form-item label="副本数">
          <el-input-number v-model="scaleReplicas" :min="0" :max="100" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showScaleDialog = false">取消</el-button>
        <el-button type="primary" @click="batchScale">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; justify-content: space-between; align-items: center; }
.filter-right { display: flex; align-items: center; gap: 8px; }
</style>
