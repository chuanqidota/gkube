<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Search } from '@element-plus/icons-vue'
import {
  getServiceAccountList, getServiceAccountYaml, deleteServiceAccount,
  getClusterRoleList, getClusterRoleYaml, deleteClusterRole,
  getRoleList, getRoleYaml, deleteRole,
  getClusterRoleBindingList, getClusterRoleBindingYaml, deleteClusterRoleBinding,
  getRoleBindingList, getRoleBindingYaml, deleteRoleBinding,
  getNamespaceList,
} from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const activeTab = ref('serviceaccounts')
const loading = ref(false)
const searchName = ref('')
const selectedNamespace = ref('')
const namespaceList = ref<string[]>([])

// Data
const serviceAccounts = ref<any[]>([])
const clusterRoles = ref<any[]>([])
const roles = ref<any[]>([])
const clusterRoleBindings = ref<any[]>([])
const roleBindings = ref<any[]>([])

// YAML dialog
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)

async function fetchNamespaces() {
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = (res.data || []).map((ns: any) => ns.name || ns)
  } catch { /* ignore */ }
}

async function fetchData() {
  loading.value = true
  try {
    const ns = selectedNamespace.value || undefined
    switch (activeTab.value) {
      case 'serviceaccounts': {
        const res: any = await getServiceAccountList({ namespace: ns })
        serviceAccounts.value = res.data || []
        break
      }
      case 'clusterroles': {
        const res: any = await getClusterRoleList()
        clusterRoles.value = res.data || []
        break
      }
      case 'roles': {
        const res: any = await getRoleList({ namespace: ns })
        roles.value = res.data || []
        break
      }
      case 'clusterrolebindings': {
        const res: any = await getClusterRoleBindingList()
        clusterRoleBindings.value = res.data || []
        break
      }
      case 'rolebindings': {
        const res: any = await getRoleBindingList({ namespace: ns })
        roleBindings.value = res.data || []
        break
      }
    }
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load data')
  } finally { loading.value = false }
}

function handleTabChange() {
  searchName.value = ''
  fetchData()
}

function filteredList(list: any[]) {
  if (!searchName.value) return list
  const keyword = searchName.value.toLowerCase()
  return list.filter((item) => item.name?.toLowerCase().includes(keyword))
}

async function handleViewYaml(params: any) {
  yamlDialogVisible.value = true; yamlLoading.value = true; yamlContent.value = ''
  try {
    let res: any
    switch (activeTab.value) {
      case 'serviceaccounts': res = await getServiceAccountYaml(params); break
      case 'clusterroles': res = await getClusterRoleYaml(params); break
      case 'roles': res = await getRoleYaml(params); break
      case 'clusterrolebindings': res = await getClusterRoleBindingYaml(params); break
      case 'rolebindings': res = await getRoleBindingYaml(params); break
    }
    yamlContent.value = res?.data?.yaml || res?.data || ''
  } catch (e: any) { ElMessage.error(e?.message || 'Failed to load YAML'); yamlDialogVisible.value = false }
  finally { yamlLoading.value = false }
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(`Delete ${activeTab.value.slice(0, -1)} "${row.name}"?`, 'Confirm', { type: 'warning' })
    const params = row.namespace ? { namespace: row.namespace, name: row.name } : { name: row.name }
    switch (activeTab.value) {
      case 'serviceaccounts': await deleteServiceAccount(params as any); break
      case 'clusterroles': await deleteClusterRole(params); break
      case 'roles': await deleteRole(params as any); break
      case 'clusterrolebindings': await deleteClusterRoleBinding(params); break
      case 'rolebindings': await deleteRoleBinding(params as any); break
    }
    ElMessage.success('Deleted'); fetchData()
  } catch { /* cancelled */ }
}

onMounted(() => { fetchNamespaces(); fetchData() })
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <el-input v-model="searchName" placeholder="Search by name" style="width: 220px;" clearable>
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select v-if="activeTab === 'serviceaccounts' || activeTab === 'roles' || activeTab === 'rolebindings'" v-model="selectedNamespace" placeholder="All Namespaces" clearable style="width: 180px;" @change="fetchData">
          <el-option v-for="ns in namespaceList" :key="ns" :label="ns" :value="ns" />
        </el-select>
        <el-button type="primary" @click="fetchData"><el-icon><Refresh /></el-icon> Refresh</el-button>
      </div>
    </el-card>

    <el-card shadow="never">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <!-- ServiceAccounts -->
        <el-tab-pane label="ServiceAccounts" name="serviceaccounts">
          <el-table :data="filteredList(serviceAccounts)" v-loading="loading" stripe>
            <el-table-column prop="name" label="Name" min-width="200" show-overflow-tooltip />
            <el-table-column prop="namespace" label="Namespace" width="140" />
            <el-table-column prop="secrets" label="Secrets" width="100" />
            <el-table-column prop="age" label="Age" width="180" />
            <el-table-column label="Actions" width="160" fixed="right">
              <template #default="{ row }">
                <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
                <el-button size="small" type="danger" @click="handleDelete(row)">Delete</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- ClusterRoles -->
        <el-tab-pane label="ClusterRoles" name="clusterroles">
          <el-table :data="filteredList(clusterRoles)" v-loading="loading" stripe>
            <el-table-column prop="name" label="Name" min-width="250" show-overflow-tooltip />
            <el-table-column prop="rules" label="Rules" width="100" />
            <el-table-column label="System" width="100">
              <template #default="{ row }"><el-tag :type="row.is_system ? 'info' : 'success'" size="small">{{ row.is_system ? 'Yes' : 'No' }}</el-tag></template>
            </el-table-column>
            <el-table-column prop="age" label="Age" width="180" />
            <el-table-column label="Actions" width="160" fixed="right">
              <template #default="{ row }">
                <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
                <el-button size="small" type="danger" @click="handleDelete(row)" :disabled="row.is_system">Delete</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- Roles -->
        <el-tab-pane label="Roles" name="roles">
          <el-table :data="filteredList(roles)" v-loading="loading" stripe>
            <el-table-column prop="name" label="Name" min-width="200" show-overflow-tooltip />
            <el-table-column prop="namespace" label="Namespace" width="140" />
            <el-table-column prop="rules" label="Rules" width="100" />
            <el-table-column prop="age" label="Age" width="180" />
            <el-table-column label="Actions" width="160" fixed="right">
              <template #default="{ row }">
                <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
                <el-button size="small" type="danger" @click="handleDelete(row)">Delete</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- ClusterRoleBindings -->
        <el-tab-pane label="ClusterRoleBindings" name="clusterrolebindings">
          <el-table :data="filteredList(clusterRoleBindings)" v-loading="loading" stripe>
            <el-table-column prop="name" label="Name" min-width="250" show-overflow-tooltip />
            <el-table-column prop="role" label="Role" width="200" />
            <el-table-column label="Subjects" min-width="250" show-overflow-tooltip>
              <template #default="{ row }">{{ (row.subjects || []).join(', ') }}</template>
            </el-table-column>
            <el-table-column prop="age" label="Age" width="180" />
            <el-table-column label="Actions" width="160" fixed="right">
              <template #default="{ row }">
                <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
                <el-button size="small" type="danger" @click="handleDelete(row)">Delete</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- RoleBindings -->
        <el-tab-pane label="RoleBindings" name="rolebindings">
          <el-table :data="filteredList(roleBindings)" v-loading="loading" stripe>
            <el-table-column prop="name" label="Name" min-width="200" show-overflow-tooltip />
            <el-table-column prop="namespace" label="Namespace" width="140" />
            <el-table-column prop="role" label="Role" width="180" />
            <el-table-column label="Subjects" min-width="200" show-overflow-tooltip>
              <template #default="{ row }">{{ (row.subjects || []).join(', ') }}</template>
            </el-table-column>
            <el-table-column prop="age" label="Age" width="180" />
            <el-table-column label="Actions" width="160" fixed="right">
              <template #default="{ row }">
                <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
                <el-button size="small" type="danger" @click="handleDelete(row)">Delete</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <el-dialog v-model="yamlDialogVisible" title="YAML" width="70%" top="5vh" destroy-on-close>
      <div v-loading="yamlLoading"><YamlEditor v-model="yamlContent" height="500px" read-only /></div>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; gap: 12px; align-items: center; flex-wrap: wrap; }
</style>
