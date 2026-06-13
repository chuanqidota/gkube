<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Plus, Delete, Setting, FolderOpened } from '@element-plus/icons-vue'
import request from '@/api/request'

const { t } = useI18n()
const loading = ref(false)
const namespaces = ref<any[]>([])
const selectedNs = ref<any>(null)
const showDetailDialog = ref(false)
const showCreateDialog = ref(false)
const showQuotaDialog = ref(false)

// Create form
const createForm = ref({
  name: '',
  labels: [] as Array<{ key: string; value: string }>,
})

// Quota form
const quotaForm = ref({
  cpu: '',
  memory: '',
  storage: '',
  pods: 0,
})

function addLabel() {
  createForm.value.labels.push({ key: '', value: '' })
}

function removeLabel(index: number) {
  createForm.value.labels.splice(index, 1)
}

async function fetchNamespaces() {
  loading.value = true
  try {
    const res: any = await request.get('/k8s/namespace/list')
    namespaces.value = res.data || []
  } catch (e: any) {
    ElMessage.warning('Failed to load namespaces')
  } finally {
    loading.value = false
  }
}

async function viewNamespace(ns: any) {
  try {
    const res: any = await request.get('/k8s/namespace/detail', {
      params: { name: ns.name }
    })
    selectedNs.value = res.data
  } catch {
    selectedNs.value = ns
  }
  showDetailDialog.value = true
}

async function createNamespace() {
  if (!createForm.value.name) {
    ElMessage.warning('Please enter a namespace name')
    return
  }

  try {
    const labels: Record<string, string> = {}
    createForm.value.labels.forEach(l => {
      if (l.key.trim()) labels[l.key.trim()] = l.value
    })

    await request.post('/k8s/namespace/create', {
      name: createForm.value.name,
      labels,
    })
    ElMessage.success('Namespace created')
    showCreateDialog.value = false
    createForm.value = { name: '', labels: [] }
    fetchNamespaces()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to create namespace')
  }
}

async function deleteNamespace(ns: any) {
  try {
    await ElMessageBox.confirm(
      `Delete namespace "${ns.name}"? All resources in this namespace will be deleted.`,
      'Warning',
      { type: 'warning' }
    )
    await request.delete('/k8s/namespace/delete', {
      params: { name: ns.name }
    })
    ElMessage.success('Namespace deleted')
    fetchNamespaces()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || 'Failed to delete namespace')
    }
  }
}

async function setQuota(ns: any) {
  selectedNs.value = ns
  // Load existing quota
  try {
    const res: any = await request.get('/k8s/resourcequota/list', {
      params: { namespace: ns.name }
    })
    const quotas = res.data || []
    if (quotas.length > 0) {
      const quota = quotas[0]
      quotaForm.value = {
        cpu: quota.spec?.hard?.['cpu'] || '',
        memory: quota.spec?.hard?.['memory'] || '',
        storage: quota.spec?.hard?.['storage'] || '',
        pods: parseInt(quota.spec?.hard?.['pods']) || 0,
      }
    } else {
      quotaForm.value = { cpu: '', memory: '', storage: '', pods: 0 }
    }
  } catch {
    quotaForm.value = { cpu: '', memory: '', storage: '', pods: 0 }
  }
  showQuotaDialog.value = true
}

async function saveQuota() {
  try {
    await request.post('/k8s/resourcequota/create', {
      name: `${selectedNs.value.name}-quota`,
      namespace: selectedNs.value.name,
      hard: {
        cpu: quotaForm.value.cpu,
        memory: quotaForm.value.memory,
        storage: quotaForm.value.storage,
        pods: quotaForm.value.pods.toString(),
      },
    })
    ElMessage.success('Quota saved')
    showQuotaDialog.value = false
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to save quota')
  }
}

function statusColor(status: string) {
  if (status === 'Active') return 'success'
  if (status === 'Terminating') return 'warning'
  return 'info'
}

onMounted(fetchNamespaces)
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <h3 style="margin: 0;"><el-icon><FolderOpened /></el-icon> {{ t('namespace.title') }}</h3>
        <div class="filter-right">
          <el-button type="primary" @click="showCreateDialog = true"><el-icon><Plus /></el-icon> {{ t('namespace.createNamespace') }}</el-button>
          <el-button @click="fetchNamespaces"><el-icon><Refresh /></el-icon> {{ t('common.refresh') }}</el-button>
        </div>
      </div>
    </el-card>

    <el-card shadow="never">
      <el-table :data="namespaces" v-loading="loading" stripe>
        <el-table-column prop="name" :label="t('common.name')" min-width="150">
          <template #default="{ row }">
            <el-link type="primary" @click="viewNamespace(row)">{{ row.name }}</el-link>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.status')" width="100">
          <template #default="{ row }">
            <el-tag :type="statusColor(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('workload.labels')" min-width="200">
          <template #default="{ row }">
            <el-tag v-for="(val, key) in row.labels" :key="key" size="small" style="margin: 2px;">
              {{ key }}={{ val }}
            </el-tag>
            <span v-if="!row.labels || Object.keys(row.labels).length === 0" style="color: #909399;">{{ t('namespace.noLabels') }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="creationTimestamp" :label="t('config.creationTime')" width="180" />
        <el-table-column :label="t('common.actions')" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="viewNamespace(row)">{{ t('common.detail') }}</el-button>
            <el-button size="small" type="warning" @click="setQuota(row)">{{ t('namespace.quota') }}</el-button>
            <el-button size="small" type="danger" @click="deleteNamespace(row)"><el-icon><Delete /></el-icon></el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Namespace Detail Dialog -->
    <el-dialog v-model="showDetailDialog" :title="selectedNs?.name" width="700px">
      <el-descriptions :column="2" border v-if="selectedNs">
        <el-descriptions-item :label="t('common.name')">{{ selectedNs.name }}</el-descriptions-item>
        <el-descriptions-item :label="t('common.status')">
          <el-tag :type="statusColor(selectedNs.status)">{{ selectedNs.status }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="t('config.creationTime')">{{ selectedNs.creationTimestamp }}</el-descriptions-item>
        <el-descriptions-item :label="t('namespace.uid')">{{ selectedNs.uid }}</el-descriptions-item>
        <el-descriptions-item :label="t('workload.labels')" :span="2">
          <el-tag v-for="(val, key) in selectedNs.labels" :key="key" size="small" style="margin: 2px;">
            {{ key }}={{ val }}
          </el-tag>
          <span v-if="!selectedNs.labels || Object.keys(selectedNs.labels).length === 0">{{ t('common.no') }}</span>
        </el-descriptions-item>
        <el-descriptions-item :label="t('namespace.annotations')" :span="2">
          <div v-for="(val, key) in selectedNs.annotations" :key="key" style="font-size: 12px;">
            {{ key }}: {{ val }}
          </div>
          <span v-if="!selectedNs.annotations || Object.keys(selectedNs.annotations).length === 0">{{ t('common.no') }}</span>
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="showDetailDialog = false">{{ t('common.close') }}</el-button>
        <el-button type="warning" @click="showDetailDialog = false; setQuota(selectedNs)">{{ t('namespace.setQuota') }}</el-button>
      </template>
    </el-dialog>

    <!-- Create Namespace Dialog -->
    <el-dialog v-model="showCreateDialog" :title="t('namespace.createNamespace')" width="500px">
      <el-form :model="createForm" label-width="100px">
        <el-form-item :label="t('common.name')" required>
          <el-input v-model="createForm.name" placeholder="my-namespace" />
        </el-form-item>
        <el-form-item :label="t('workload.labels')">
          <div style="width: 100%;">
            <div v-for="(label, index) in createForm.labels" :key="index" style="display: flex; gap: 8px; margin-bottom: 8px;">
              <el-input v-model="label.key" :placeholder="t('namespace.key')" style="flex: 1;" />
              <el-input v-model="label.value" :placeholder="t('namespace.value')" style="flex: 1;" />
              <el-button type="danger" circle @click="removeLabel(index)">-</el-button>
            </div>
            <el-button @click="addLabel" type="primary" plain>{{ t('namespace.addLabel') }}</el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" @click="createNamespace">{{ t('common.create') }}</el-button>
      </template>
    </el-dialog>

    <!-- Quota Dialog -->
    <el-dialog v-model="showQuotaDialog" :title="t('namespace.resourceQuota') + ': ' + selectedNs?.name" width="500px">
      <el-form :model="quotaForm" label-width="100px">
        <el-form-item label="CPU">
          <el-input v-model="quotaForm.cpu" placeholder="4 cores" />
        </el-form-item>
        <el-form-item :label="t('namespace.memory')">
          <el-input v-model="quotaForm.memory" placeholder="8Gi" />
        </el-form-item>
        <el-form-item :label="t('namespace.storage')">
          <el-input v-model="quotaForm.storage" placeholder="100Gi" />
        </el-form-item>
        <el-form-item :label="t('namespace.podCount')">
          <el-input-number v-model="quotaForm.pods" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showQuotaDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" @click="saveQuota">{{ t('common.save') }}</el-button>
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
