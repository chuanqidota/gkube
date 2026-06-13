<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Plus, Delete, VideoPlay, VideoPause, View } from '@element-plus/icons-vue'
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
  } catch (e: any) {
    ElMessage.warning('ArgoCD not available, showing sample data')
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
    await ElMessageBox.confirm(`Sync application "${app.name}"?`, 'Confirm')
    await request.post('/k8s/gitops/sync', {
      name: app.name,
      namespace: app.namespace,
    })
    ElMessage.success('Sync started')
    fetchApplications()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || 'Sync failed')
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
    await ElMessageBox.confirm(`Rollback "${app.name}" to revision ${revision}?`, 'Confirm')
    await request.post('/k8s/gitops/rollback', null, {
      params: { name: app.name, namespace: app.namespace, revision }
    })
    ElMessage.success('Rollback successful')
    showHistoryDialog.value = false
    fetchApplications()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || 'Rollback failed')
    }
  }
}

async function deleteApp(app: any) {
  try {
    await ElMessageBox.confirm(`Delete application "${app.name}"? This will remove all resources.`, 'Warning', {
      type: 'warning',
    })
    await request.delete('/k8s/gitops/delete', {
      params: { name: app.name, namespace: app.namespace, cascade: true }
    })
    ElMessage.success('Application deleted')
    fetchApplications()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || 'Delete failed')
    }
  }
}

async function createApp() {
  if (!createForm.value.name || !createForm.value.repoURL) {
    ElMessage.warning('Please fill in all required fields')
    return
  }

  try {
    await request.post('/k8s/gitops/create', createForm.value)
    ElMessage.success('Application created')
    showCreateDialog.value = false
    fetchApplications()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Create failed')
  }
}

onMounted(fetchApplications)
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <h3 style="margin: 0;">GitOps (ArgoCD)</h3>
        <div class="filter-right">
          <el-input
            v-model="searchQuery"
            placeholder="Search applications..."
            style="width: 250px;"
            clearable
          />
          <el-select v-model="selectedNamespace" placeholder="All Namespaces" clearable style="width: 150px;">
            <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
          </el-select>
          <el-button type="primary" @click="showCreateDialog = true"><el-icon><Plus /></el-icon> Create</el-button>
          <el-button @click="fetchApplications"><el-icon><Refresh /></el-icon> Refresh</el-button>
        </div>
      </div>
    </el-card>

    <el-card shadow="never">
      <el-table :data="filteredApps" v-loading="loading" stripe>
        <el-table-column prop="name" label="Application" min-width="150">
          <template #default="{ row }">
            <el-link type="primary" @click="viewApp(row)">{{ row.name }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="Namespace" width="120" />
        <el-table-column prop="project" label="Project" width="100" />
        <el-table-column label="Source" min-width="250">
          <template #default="{ row }">
            <div>
              <div style="font-size: 12px; color: #606266;">{{ row.repoURL }}</div>
              <div style="font-size: 12px; color: #909399;">{{ row.path }} @ {{ row.targetRevision }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="Health" width="100">
          <template #default="{ row }">
            <el-tag :type="healthColor(row.health)" size="small">{{ row.health }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Sync" width="100">
          <template #default="{ row }">
            <el-tag :type="syncColor(row.syncStatus)" size="small">{{ row.syncStatus }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Actions" width="250" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="syncApp(row)"><el-icon><VideoPlay /></el-icon> Sync</el-button>
            <el-button size="small" @click="viewHistory(row)">History</el-button>
            <el-button type="danger" size="small" @click="deleteApp(row)"><el-icon><Delete /></el-icon></el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- App Detail Dialog -->
    <el-dialog v-model="showDetailDialog" :title="selectedApp?.name" width="700px">
      <el-descriptions :column="2" border v-if="selectedApp">
        <el-descriptions-item label="Name">{{ selectedApp.name }}</el-descriptions-item>
        <el-descriptions-item label="Namespace">{{ selectedApp.namespace }}</el-descriptions-item>
        <el-descriptions-item label="Project">{{ selectedApp.project }}</el-descriptions-item>
        <el-descriptions-item label="Health">
          <el-tag :type="healthColor(selectedApp.health)">{{ selectedApp.health }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="Sync Status">
          <el-tag :type="syncColor(selectedApp.syncStatus)">{{ selectedApp.syncStatus }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="Status">{{ selectedApp.status }}</el-descriptions-item>
        <el-descriptions-item label="Repository" :span="2">{{ selectedApp.repoURL }}</el-descriptions-item>
        <el-descriptions-item label="Path">{{ selectedApp.path }}</el-descriptions-item>
        <el-descriptions-item label="Revision">{{ selectedApp.targetRevision }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="showDetailDialog = false">Close</el-button>
        <el-button type="primary" @click="syncApp(selectedApp)">Sync</el-button>
      </template>
    </el-dialog>

    <!-- Create Dialog -->
    <el-dialog v-model="showCreateDialog" title="Create Application" width="600px">
      <el-form :model="createForm" label-width="140px">
        <el-form-item label="Name" required>
          <el-input v-model="createForm.name" placeholder="my-app" />
        </el-form-item>
        <el-form-item label="Namespace">
          <el-input v-model="createForm.namespace" placeholder="argocd" />
        </el-form-item>
        <el-form-item label="Project">
          <el-input v-model="createForm.project" placeholder="default" />
        </el-form-item>
        <el-form-item label="Repository URL" required>
          <el-input v-model="createForm.repoURL" placeholder="https://github.com/org/repo.git" />
        </el-form-item>
        <el-form-item label="Path" required>
          <el-input v-model="createForm.path" placeholder="k8s/manifests" />
        </el-form-item>
        <el-form-item label="Target Revision">
          <el-input v-model="createForm.targetRevision" placeholder="HEAD" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">Cancel</el-button>
        <el-button type="primary" @click="createApp">Create</el-button>
      </template>
    </el-dialog>

    <!-- History Dialog -->
    <el-dialog v-model="showHistoryDialog" :title="'History: ' + selectedApp?.name" width="700px">
      <el-table :data="history" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="revision" label="Revision" width="150" />
        <el-table-column prop="deployedAt" label="Deployed At" width="200">
          <template #default="{ row }">{{ new Date(row.deployedAt).toLocaleString() }}</template>
        </el-table-column>
        <el-table-column prop="status" label="Status" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'Succeeded' ? 'success' : 'danger'" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Action" width="120">
          <template #default="{ row }">
            <el-button type="warning" size="small" @click="rollback(selectedApp, row.revision)">Rollback</el-button>
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
