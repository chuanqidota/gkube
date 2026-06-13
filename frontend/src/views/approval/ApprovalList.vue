<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Plus, Check, Close, View, Delete } from '@element-plus/icons-vue'
import request from '@/api/request'

const loading = ref(false)
const approvals = ref<any[]>([])
const selectedStatus = ref('')
const searchQuery = ref('')
const selectedApproval = ref<any>(null)
const showDetailDialog = ref(false)
const showCreateDialog = ref(false)
const stats = ref({ total: 0, pending: 0, approved: 0, rejected: 0 })

// Create form
const createForm = ref({
  type: 'deploy',
  resource: '',
  namespace: 'default',
  cluster: '',
  details: {} as Record<string, string>,
})

const approvalTypes = [
  { value: 'deploy', label: 'Deploy Application' },
  { value: 'scale', label: 'Scale Resources' },
  { value: 'delete', label: 'Delete Resources' },
  { value: 'config', label: 'Configuration Change' },
]

const filteredApprovals = computed(() => {
  let result = approvals.value
  if (selectedStatus.value) {
    result = result.filter(a => a.status === selectedStatus.value)
  }
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(a =>
      a.resource.toLowerCase().includes(query) ||
      a.requestedBy.toLowerCase().includes(query)
    )
  }
  return result
})

function statusColor(status: string) {
  switch (status) {
    case 'pending': return 'warning'
    case 'approved': return 'success'
    case 'rejected': return 'danger'
    default: return 'info'
  }
}

function typeLabel(type: string) {
  const found = approvalTypes.find(t => t.value === type)
  return found ? found.label : type
}

async function fetchApprovals() {
  loading.value = true
  try {
    const res: any = await request.get('/k8s/approval/list')
    approvals.value = res.data || []
  } catch (e: any) {
    ElMessage.warning('Failed to load approvals')
    approvals.value = []
  } finally {
    loading.value = false
  }
}

async function fetchStats() {
  try {
    const res: any = await request.get('/k8s/approval/stats')
    stats.value = res.data || { total: 0, pending: 0, approved: 0, rejected: 0 }
  } catch {
    // ignore
  }
}

function viewApproval(approval: any) {
  selectedApproval.value = approval
  showDetailDialog.value = true
}

async function approveRequest(approval: any) {
  try {
    await ElMessageBox.confirm(`Approve this ${approval.type} request?`, 'Confirm Approval')
    await request.post('/k8s/approval/approve', {
      id: approval.id,
      reviewedBy: 'admin',
      reason: 'Approved by admin',
    })
    ElMessage.success('Request approved')
    fetchApprovals()
    fetchStats()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || 'Failed to approve')
    }
  }
}

async function rejectRequest(approval: any) {
  try {
    const { value: reason } = await ElMessageBox.prompt('Reason for rejection:', 'Reject Request', {
      inputType: 'textarea',
      inputValidator: (v) => !!v || 'Reason is required',
    })
    await request.post('/k8s/approval/reject', {
      id: approval.id,
      reviewedBy: 'admin',
      reason: reason,
    })
    ElMessage.success('Request rejected')
    fetchApprovals()
    fetchStats()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || 'Failed to reject')
    }
  }
}

async function deleteApproval(approval: any) {
  try {
    await ElMessageBox.confirm('Delete this approval request?', 'Confirm')
    await request.delete('/k8s/approval/delete', {
      params: { id: approval.id }
    })
    ElMessage.success('Deleted')
    fetchApprovals()
    fetchStats()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || 'Failed to delete')
    }
  }
}

async function createApproval() {
  if (!createForm.value.resource) {
    ElMessage.warning('Please enter a resource name')
    return
  }

  try {
    await request.post('/k8s/approval/create', createForm.value)
    ElMessage.success('Approval request created')
    showCreateDialog.value = false
    createForm.value = { type: 'deploy', resource: '', namespace: 'default', cluster: '', details: {} }
    fetchApprovals()
    fetchStats()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to create')
  }
}

function formatDate(date: string) {
  if (!date) return '-'
  return new Date(date).toLocaleString()
}

onMounted(() => {
  fetchApprovals()
  fetchStats()
})
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <h3 style="margin: 0;">Approval Workflows</h3>
        <div class="filter-right">
          <el-input v-model="searchQuery" placeholder="Search..." style="width: 200px;" clearable />
          <el-select v-model="selectedStatus" placeholder="All Status" clearable style="width: 120px;">
            <el-option label="Pending" value="pending" />
            <el-option label="Approved" value="approved" />
            <el-option label="Rejected" value="rejected" />
          </el-select>
          <el-button type="primary" @click="showCreateDialog = true"><el-icon><Plus /></el-icon> New Request</el-button>
          <el-button @click="fetchApprovals"><el-icon><Refresh /></el-icon></el-button>
        </div>
      </div>
    </el-card>

    <el-row :gutter="16" style="margin-bottom: 16px;">
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-value">{{ stats.total }}</div>
          <div class="stat-label">Total Requests</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card warning">
          <div class="stat-value">{{ stats.pending }}</div>
          <div class="stat-label">Pending</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card success">
          <div class="stat-value">{{ stats.approved }}</div>
          <div class="stat-label">Approved</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card danger">
          <div class="stat-value">{{ stats.rejected }}</div>
          <div class="stat-label">Rejected</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="never">
      <el-table :data="filteredApprovals" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="150" show-overflow-tooltip />
        <el-table-column label="Type" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ typeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="resource" label="Resource" min-width="150" show-overflow-tooltip />
        <el-table-column prop="namespace" label="Namespace" width="120" />
        <el-table-column prop="requestedBy" label="Requested By" width="120" />
        <el-table-column label="Requested At" width="180">
          <template #default="{ row }">{{ formatDate(row.requestedAt) }}</template>
        </el-table-column>
        <el-table-column label="Status" width="100">
          <template #default="{ row }">
            <el-tag :type="statusColor(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Actions" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="viewApproval(row)"><el-icon><View /></el-icon></el-button>
            <el-button v-if="row.status === 'pending'" type="success" size="small" @click="approveRequest(row)">
              <el-icon><Check /></el-icon>
            </el-button>
            <el-button v-if="row.status === 'pending'" type="danger" size="small" @click="rejectRequest(row)">
              <el-icon><Close /></el-icon>
            </el-button>
            <el-button type="info" size="small" @click="deleteApproval(row)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Detail Dialog -->
    <el-dialog v-model="showDetailDialog" title="Approval Details" width="600px">
      <el-descriptions :column="2" border v-if="selectedApproval">
        <el-descriptions-item label="ID">{{ selectedApproval.id }}</el-descriptions-item>
        <el-descriptions-item label="Type">{{ typeLabel(selectedApproval.type) }}</el-descriptions-item>
        <el-descriptions-item label="Resource">{{ selectedApproval.resource }}</el-descriptions-item>
        <el-descriptions-item label="Namespace">{{ selectedApproval.namespace }}</el-descriptions-item>
        <el-descriptions-item label="Cluster">{{ selectedApproval.cluster || 'default' }}</el-descriptions-item>
        <el-descriptions-item label="Status">
          <el-tag :type="statusColor(selectedApproval.status)">{{ selectedApproval.status }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="Requested By">{{ selectedApproval.requestedBy }}</el-descriptions-item>
        <el-descriptions-item label="Requested At">{{ formatDate(selectedApproval.requestedAt) }}</el-descriptions-item>
        <el-descriptions-item v-if="selectedApproval.reviewedBy" label="Reviewed By">{{ selectedApproval.reviewedBy }}</el-descriptions-item>
        <el-descriptions-item v-if="selectedApproval.reviewedAt" label="Reviewed At">{{ formatDate(selectedApproval.reviewedAt) }}</el-descriptions-item>
        <el-descriptions-item v-if="selectedApproval.reason" label="Reason" :span="2">{{ selectedApproval.reason }}</el-descriptions-item>
      </el-descriptions>

      <template #footer>
        <el-button @click="showDetailDialog = false">Close</el-button>
        <el-button v-if="selectedApproval?.status === 'pending'" type="success" @click="approveRequest(selectedApproval); showDetailDialog = false">Approve</el-button>
        <el-button v-if="selectedApproval?.status === 'pending'" type="danger" @click="rejectRequest(selectedApproval); showDetailDialog = false">Reject</el-button>
      </template>
    </el-dialog>

    <!-- Create Dialog -->
    <el-dialog v-model="showCreateDialog" title="New Approval Request" width="500px">
      <el-form :model="createForm" label-width="120px">
        <el-form-item label="Type" required>
          <el-select v-model="createForm.type" style="width: 100%;">
            <el-option v-for="t in approvalTypes" :key="t.value" :label="t.label" :value="t.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="Resource" required>
          <el-input v-model="createForm.resource" placeholder="deployment/my-app" />
        </el-form-item>
        <el-form-item label="Namespace">
          <el-input v-model="createForm.namespace" placeholder="default" />
        </el-form-item>
        <el-form-item label="Cluster">
          <el-input v-model="createForm.cluster" placeholder="default cluster" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">Cancel</el-button>
        <el-button type="primary" @click="createApproval">Submit</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; justify-content: space-between; align-items: center; }
.filter-right { display: flex; align-items: center; gap: 8px; }
.stat-card { text-align: center; }
.stat-card.warning { border-left: 4px solid #E6A23C; }
.stat-card.success { border-left: 4px solid #67C23A; }
.stat-card.danger { border-left: 4px solid #F56C6C; }
.stat-value { font-size: 32px; font-weight: bold; color: #303133; }
.stat-label { font-size: 14px; color: #909399; margin-top: 4px; }
</style>
