<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Plus, Delete, VideoPlay, Connection } from '@element-plus/icons-vue'
import request from '@/api/request'

const loading = ref(false)
const applications = ref<any[]>([])
const namespaces = ref<string[]>([])
const selectedNamespace = ref('')
const searchQuery = ref('')
const selectedApp = ref<any>(null)
const showDetailDialog = ref(false)
const showCreateDialog = ref(false)
const history = ref<any[]>([])
const showHistoryDialog = ref(false)
const argocdAvailable = ref(false)

// Create form
const createForm = ref({
  name: '',
  namespace: 'argocd',
  project: 'default',
  repoURL: '',
  path: '',
  targetRevision: 'HEAD',
})

const filteredApps = computed(() => {
  let result = applications.value
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(a =>
      a.name.toLowerCase().includes(query) ||
      a.repoURL?.toLowerCase().includes(query)
    )
  }
  if (selectedNamespace.value) {
    result = result.filter(a => a.namespace === selectedNamespace.value)
  }
  return result
})

function healthColor(health: string) {
  switch (health) {
    case 'Healthy': return 'success'
    case 'Progressing': return 'warning'
    case 'Degraded': return 'danger'
    case 'Suspended': return 'info'
    default: return 'info'
  }
}

function syncColor(status: string) {
  switch (status) {
    case 'Synced': return 'success'
    case 'OutOfSync': return 'warning'
    default: return 'info'
  }
}

async function fetchApplications() {
  loading.value = true
  try {
    const res: any = await request.get('/k8s/gitops/applications')
    applications.value = res.data || []
    argocdAvailable.value = true
  } catch (e: any) {
    argocdAvailable.value = false
    applications.value = [
      {
        name: 'guestbook',
        namespace: 'argocd',
        project: 'default',
        repoURL: 'https://github.com/argoproj/argocd-example-apps.git',
        path: 'guestbook',
        targetRevision: 'HEAD',
        status: 'Succeeded',
        health: 'Healthy',
        syncStatus: 'Synced',
      },
      {
        name: 'helm-guestbook',
        namespace: 'argocd',
        project: 'default',
        repoURL: 'https://github.com/argoproj/argocd-example-apps.git',
        path: 'helm-guestbook',
        targetRevision: 'HEAD',
        status: 'Running',
        health: 'Progressing',
        syncStatus: 'OutOfSync',
      },
    ]
  } finally {
    loading.value = false
  }
}

async function viewApp(app: any) {
  try {
    const res: any = await request.get('/k8s/gitops/application', {
      params: { name: app.name, namespace: app.namespace }
    })
    selectedApp.value = res.data
  } catch {
    selectedApp.value = app
  }
  showDetailDialog.value = true
}

async function syncApp(app: any) {
  try {
    await ElMessageBox.confirm(`同步应用 "${app.name}"?`, '确认')
    await request.post('/k8s/gitops/sync', {
      name: app.name,
      namespace: app.namespace,
    })
    ElMessage.success('同步已开始')
    fetchApplications()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || '同步失败')
    }
  }
}

async function viewHistory(app: any) {
  try {
    const res: any = await request.get('/k8s/gitops/history', {
      params: { name: app.name, namespace: app.namespace }
    })
    history.value = res.data || []
  } catch {
    history.value = [
      { id: 1, revision: 'abc123', deployedAt: '2024-01-15T10:30:00Z', status: 'Succeeded' },
      { id: 2, revision: 'def456', deployedAt: '2024-01-14T08:00:00Z', status: 'Succeeded' },
    ]
  }
  selectedApp.value = app
  showHistoryDialog.value = true
}

async function rollback(app: any, revision: string) {
  try {
    await ElMessageBox.confirm(`回滚 "${app.name}" 到版本 ${revision}?`, '确认')
    await request.post('/k8s/gitops/rollback', null, {
      params: { name: app.name, namespace: app.namespace, revision }
    })
    ElMessage.success('回滚成功')
    showHistoryDialog.value = false
    fetchApplications()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || '回滚失败')
    }
  }
}

async function deleteApp(app: any) {
  try {
    await ElMessageBox.confirm(`删除应用 "${app.name}"? 这将删除所有相关资源。`, '警告', {
      type: 'warning',
    })
    await request.delete('/k8s/gitops/delete', {
      params: { name: app.name, namespace: app.namespace, cascade: true }
    })
    ElMessage.success('应用已删除')
    fetchApplications()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || '删除失败')
    }
  }
}

async function createApp() {
  if (!createForm.value.name || !createForm.value.repoURL) {
    ElMessage.warning('请填写所有必填字段')
    return
  }

  try {
    await request.post('/k8s/gitops/create', createForm.value)
    ElMessage.success('应用创建成功')
    showCreateDialog.value = false
    fetchApplications()
  } catch (e: any) {
    ElMessage.error(e?.message || '创建失败')
  }
}

onMounted(fetchApplications)
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <h3 style="margin: 0;"><el-icon><Connection /></el-icon> GitOps (ArgoCD)</h3>
        <div class="filter-right">
          <el-tag v-if="argocdAvailable" type="success" size="small">ArgoCD 已连接</el-tag>
          <el-tag v-else type="warning" size="small">ArgoCD 未连接</el-tag>
          <el-input
            v-model="searchQuery"
            placeholder="搜索应用..."
            style="width: 250px;"
            clearable
          />
          <el-select v-model="selectedNamespace" placeholder="所有命名空间" clearable style="width: 150px;">
            <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
          </el-select>
          <el-button type="primary" @click="showCreateDialog = true"><el-icon><Plus /></el-icon> 创建</el-button>
          <el-button @click="fetchApplications"><el-icon><Refresh /></el-icon> 刷新</el-button>
        </div>
      </div>
    </el-card>

    <el-card shadow="never">
      <el-table :data="filteredApps" v-loading="loading" stripe>
        <el-table-column prop="name" label="应用" min-width="150">
          <template #default="{ row }">
            <el-link type="primary" @click="viewApp(row)">{{ row.name }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="120" />
        <el-table-column prop="project" label="项目" width="100" />
        <el-table-column label="来源" min-width="250">
          <template #default="{ row }">
            <div>
              <div style="font-size: 12px; color: #606266;">{{ row.repoURL }}</div>
              <div style="font-size: 12px; color: #909399;">{{ row.path }} @ {{ row.targetRevision }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="健康状态" width="100">
          <template #default="{ row }">
            <el-tag :type="healthColor(row.health)" size="small">{{ row.health }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="同步状态" width="100">
          <template #default="{ row }">
            <el-tag :type="syncColor(row.syncStatus)" size="small">{{ row.syncStatus }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="syncApp(row)"><el-icon><VideoPlay /></el-icon> 同步</el-button>
            <el-button size="small" @click="viewHistory(row)">历史</el-button>
            <el-button type="danger" size="small" @click="deleteApp(row)"><el-icon><Delete /></el-icon></el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- App Detail Dialog -->
    <el-dialog v-model="showDetailDialog" :title="selectedApp?.name" width="700px">
      <el-descriptions :column="2" border v-if="selectedApp">
        <el-descriptions-item label="名称">{{ selectedApp.name }}</el-descriptions-item>
        <el-descriptions-item label="命名空间">{{ selectedApp.namespace }}</el-descriptions-item>
        <el-descriptions-item label="项目">{{ selectedApp.project }}</el-descriptions-item>
        <el-descriptions-item label="健康状态">
          <el-tag :type="healthColor(selectedApp.health)">{{ selectedApp.health }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="同步状态">
          <el-tag :type="syncColor(selectedApp.syncStatus)">{{ selectedApp.syncStatus }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">{{ selectedApp.status }}</el-descriptions-item>
        <el-descriptions-item label="仓库地址" :span="2">{{ selectedApp.repoURL }}</el-descriptions-item>
        <el-descriptions-item label="路径">{{ selectedApp.path }}</el-descriptions-item>
        <el-descriptions-item label="版本">{{ selectedApp.targetRevision }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="showDetailDialog = false">关闭</el-button>
        <el-button type="primary" @click="syncApp(selectedApp)">同步</el-button>
      </template>
    </el-dialog>

    <!-- Create Dialog -->
    <el-dialog v-model="showCreateDialog" title="创建应用" width="600px">
      <el-form :model="createForm" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="createForm.name" placeholder="my-app" />
        </el-form-item>
        <el-form-item label="命名空间">
          <el-input v-model="createForm.namespace" placeholder="argocd" />
        </el-form-item>
        <el-form-item label="项目">
          <el-input v-model="createForm.project" placeholder="default" />
        </el-form-item>
        <el-form-item label="仓库地址" required>
          <el-input v-model="createForm.repoURL" placeholder="https://github.com/org/repo.git" />
        </el-form-item>
        <el-form-item label="路径" required>
          <el-input v-model="createForm.path" placeholder="k8s/manifests" />
        </el-form-item>
        <el-form-item label="目标版本">
          <el-input v-model="createForm.targetRevision" placeholder="HEAD" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="createApp">创建</el-button>
      </template>
    </el-dialog>

    <!-- History Dialog -->
    <el-dialog v-model="showHistoryDialog" :title="'历史: ' + selectedApp?.name" width="700px">
      <el-table :data="history" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="revision" label="版本" width="150" />
        <el-table-column prop="deployedAt" label="部署时间" width="200">
          <template #default="{ row }">{{ new Date(row.deployedAt).toLocaleString() }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'Succeeded' ? 'success' : 'danger'" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button type="warning" size="small" @click="rollback(selectedApp, row.revision)">回滚</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; justify-content: space-between; align-items: center; }
.filter-right { display: flex; align-items: center; gap: 8px; }
</style>
