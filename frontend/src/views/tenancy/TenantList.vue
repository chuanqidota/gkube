<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Plus, Delete, Setting, FolderOpened } from '@element-plus/icons-vue'
import request from '@/api/request'

const { t } = useI18n()
const loading = ref(false)
const tenants = ref<any[]>([])
const namespaces = ref<string[]>([])
const selectedTenant = ref<any>(null)
const showDetailDialog = ref(false)
const showCreateDialog = ref(false)

// Create form
const createForm = ref({
  name: '',
  users: [] as string[],
  quotas: {
    cpu: '',
    memory: '',
    storage: '',
    pods: 0,
  },
  labels: {} as Record<string, string>,
  annotations: {} as Record<string, string>,
})

const newUser = ref('')

async function fetchTenants() {
  loading.value = true
  try {
    const res: any = await request.get('/k8s/tenancy/tenants')
    tenants.value = res.data || []
  } catch (e: any) {
    ElMessage.warning('Failed to load tenants')
    tenants.value = []
  } finally {
    loading.value = false
  }
}

async function fetchNamespaces() {
  try {
    const res: any = await request.get('/k8s/namespace/list')
    namespaces.value = res.data?.map((ns: any) => ns.name) || []
  } catch {
    namespaces.value = []
  }
}

async function viewTenant(tenant: any) {
  try {
    const res: any = await request.get('/k8s/tenancy/tenant', {
      params: { name: tenant.name }
    })
    selectedTenant.value = res.data
  } catch {
    selectedTenant.value = tenant
  }
  showDetailDialog.value = true
}

async function deleteTenant(tenant: any) {
  try {
    await ElMessageBox.confirm(
      `Delete tenant "${tenant.name}" and all its namespaces? This action cannot be undone.`,
      'Warning',
      { type: 'warning' }
    )
    await request.delete('/k8s/tenancy/delete', {
      params: { name: tenant.name }
    })
    ElMessage.success('Tenant deleted')
    fetchTenants()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || 'Delete failed')
    }
  }
}

function addUser() {
  if (newUser.value && !createForm.value.users.includes(newUser.value)) {
    createForm.value.users.push(newUser.value)
    newUser.value = ''
  }
}

function removeUser(index: number) {
  createForm.value.users.splice(index, 1)
}

async function createTenant() {
  if (!createForm.value.name) {
    ElMessage.warning('Please enter a tenant name')
    return
  }

  try {
    await request.post('/k8s/tenancy/create', createForm.value)
    ElMessage.success('Tenant created')
    showCreateDialog.value = false
    resetCreateForm()
    fetchTenants()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Create failed')
  }
}

async function addNamespaceToTenant(tenant: any, namespace: string) {
  try {
    await request.post('/k8s/tenancy/namespace/add', null, {
      params: { tenant: tenant.name, namespace }
    })
    ElMessage.success('Namespace added to tenant')
    viewTenant(tenant)
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to add namespace')
  }
}

async function removeNamespaceFromTenant(namespace: string) {
  try {
    await request.post('/k8s/tenancy/namespace/remove', null, {
      params: { namespace }
    })
    ElMessage.success('Namespace removed from tenant')
    if (selectedTenant.value) {
      viewTenant(selectedTenant.value)
    }
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to remove namespace')
  }
}

function resetCreateForm() {
  createForm.value = {
    name: '',
    users: [],
    quotas: { cpu: '', memory: '', storage: '', pods: 0 },
    labels: {},
    annotations: {},
  }
}

onMounted(() => {
  fetchTenants()
  fetchNamespaces()
})
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <h3 style="margin: 0;">Multi-Tenancy</h3>
        <div class="filter-right">
          <el-button type="primary" @click="showCreateDialog = true"><el-icon><Plus /></el-icon> Create Tenant</el-button>
          <el-button @click="fetchTenants"><el-icon><Refresh /></el-icon> Refresh</el-button>
        </div>
      </div>
    </el-card>

    <el-card shadow="never">
      <el-table :data="tenants" v-loading="loading" stripe>
        <el-table-column prop="name" label="Tenant Name" min-width="150">
          <template #default="{ row }">
            <el-link type="primary" @click="viewTenant(row)">{{ row.name }}</el-link>
          </template>
        </el-table-column>
        <el-table-column label="Namespaces" min-width="200">
          <template #default="{ row }">
            <el-tag v-for="ns in row.namespaces" :key="ns" size="small" style="margin: 2px;">{{ ns }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Users" min-width="150">
          <template #default="{ row }">
            <el-tag v-for="user in row.users" :key="user" type="info" size="small" style="margin: 2px;">{{ user }}</el-tag>
            <span v-if="!row.users?.length" style="color: #909399;">No users</span>
          </template>
        </el-table-column>
        <el-table-column label="Quotas" min-width="150">
          <template #default="{ row }">
            <div v-if="row.quotas">
              <div>CPU: {{ row.quotas.cpu || 'N/A' }}</div>
              <div>Memory: {{ row.quotas.memory || 'N/A' }}</div>
            </div>
            <span v-else style="color: #909399;">No quotas</span>
          </template>
        </el-table-column>
        <el-table-column label="Actions" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="danger" size="small" @click="deleteTenant(row)"><el-icon><Delete /></el-icon></el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Tenant Detail Dialog -->
    <el-dialog v-model="showDetailDialog" :title="'Tenant: ' + selectedTenant?.name" width="700px">
      <el-descriptions :column="2" border v-if="selectedTenant">
        <el-descriptions-item label="Name">{{ selectedTenant.name }}</el-descriptions-item>
        <el-descriptions-item label="Namespaces">
          <el-tag v-for="ns in selectedTenant.namespaces" :key="ns" size="small" style="margin: 2px;">{{ ns }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="Users" :span="2">
          <el-tag v-for="user in selectedTenant.users" :key="user" type="info" size="small" style="margin: 2px;">{{ user }}</el-tag>
          <span v-if="!selectedTenant.users?.length" style="color: #909399;">No users assigned</span>
        </el-descriptions-item>
        <el-descriptions-item label="CPU Quota">{{ selectedTenant.quotas?.cpu || 'N/A' }}</el-descriptions-item>
        <el-descriptions-item label="Memory Quota">{{ selectedTenant.quotas?.memory || 'N/A' }}</el-descriptions-item>
        <el-descriptions-item label="Storage Quota">{{ selectedTenant.quotas?.storage || 'N/A' }}</el-descriptions-item>
        <el-descriptions-item label="Pods Quota">{{ selectedTenant.quotas?.pods || 'N/A' }}</el-descriptions-item>
      </el-descriptions>

      <h4 style="margin-top: 20px;">Manage Namespaces</h4>
      <div style="display: flex; gap: 8px; margin-bottom: 12px;">
        <el-select v-model="selectedTenant.newNamespace" placeholder="Add namespace" style="flex: 1;">
          <el-option
            v-for="ns in namespaces.filter(n => !selectedTenant?.namespaces?.includes(n))"
            :key="ns"
            :label="ns"
            :value="ns"
          />
        </el-select>
        <el-button type="primary" @click="addNamespaceToTenant(selectedTenant, selectedTenant.newNamespace)">Add</el-button>
      </div>

      <el-table :data="selectedTenant?.namespaces?.map((ns: string) => ({ name: ns })) || []" size="small">
        <el-table-column prop="name" label="Namespace" />
        <el-table-column label="Action" width="100">
          <template #default="{ row }">
            <el-button type="danger" size="small" @click="removeNamespaceFromTenant(row.name)">Remove</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- Create Tenant Dialog -->
    <el-dialog v-model="showCreateDialog" title="Create Tenant" width="600px">
      <el-form :model="createForm" label-width="120px">
        <el-form-item label="Tenant Name" required>
          <el-input v-model="createForm.name" placeholder="my-tenant" />
        </el-form-item>

        <el-form-item label="Users">
          <div style="display: flex; gap: 8px; width: 100%;">
            <el-input v-model="newUser" placeholder="Add user" @keyup.enter="addUser" style="flex: 1;" />
            <el-button @click="addUser">Add</el-button>
          </div>
          <div style="margin-top: 8px;">
            <el-tag
              v-for="(user, index) in createForm.users"
              :key="user"
              closable
              @close="removeUser(index)"
              style="margin: 2px;"
            >
              {{ user }}
            </el-tag>
          </div>
        </el-form-item>

        <el-divider>Resource Quotas</el-divider>

        <el-form-item label="CPU">
          <el-input v-model="createForm.quotas.cpu" placeholder="4 cores" />
        </el-form-item>
        <el-form-item label="Memory">
          <el-input v-model="createForm.quotas.memory" placeholder="8Gi" />
        </el-form-item>
        <el-form-item label="Storage">
          <el-input v-model="createForm.quotas.storage" placeholder="100Gi" />
        </el-form-item>
        <el-form-item label="Pods">
          <el-input-number v-model="createForm.quotas.pods" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">Cancel</el-button>
        <el-button type="primary" @click="createTenant">Create</el-button>
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
