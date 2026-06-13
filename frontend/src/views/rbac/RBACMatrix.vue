<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Download } from '@element-plus/icons-vue'
import request from '@/api/request'

const loading = ref(false)
const roles = ref<any[]>([])
const permissions = ref<any[]>([])
const selectedRole = ref('')
const viewMode = ref('matrix') // 'matrix' or 'list'

const resources = [
  'pods', 'deployments', 'statefulsets', 'daemonsets', 'services', 'ingresses',
  'configmaps', 'secrets', 'persistentvolumeclaims', 'persistentvolumes',
  'namespaces', 'nodes', 'serviceaccounts', 'roles', 'rolebindings',
  'clusterroles', 'clusterrolebindings', 'resourcequotas', 'limitranges',
  'horizontalpodautoscalers', 'networkpolicies', 'poddisruptionbudgets',
]

const verbs = ['get', 'list', 'watch', 'create', 'update', 'patch', 'delete']

const matrixData = ref<Record<string, Record<string, string[]>>>({})

async function fetchRoles() {
  try {
    const res: any = await request.get('/roles')
    roles.value = res.data || []
  } catch {
    roles.value = [
      { id: 1, name: 'super_admin', description: 'Super Administrator' },
      { id: 2, name: 'admin', description: 'Administrator' },
      { id: 3, name: 'developer', description: 'Developer' },
      { id: 4, name: 'viewer', description: 'Read-only User' },
    ]
  }
}

async function fetchPermissions() {
  loading.value = true
  try {
    const res: any = await request.get('/roles')
    const rolesData = res.data || []

    // Build matrix data
    const matrix: Record<string, Record<string, string[]>> = {}
    for (const role of rolesData) {
      matrix[role.name] = {}
      for (const resource of resources) {
        matrix[role.name][resource] = []
      }
    }

    // Simulate permissions for demo
    if (rolesData.length > 0) {
      matrix['super_admin'] = {}
      for (const resource of resources) {
        matrix['super_admin'][resource] = [...verbs]
      }

      matrix['admin'] = {}
      for (const resource of resources) {
        matrix['admin'][resource] = ['get', 'list', 'watch', 'create', 'update', 'delete']
      }

      matrix['developer'] = {
        pods: ['get', 'list', 'watch', 'create', 'update', 'delete'],
        deployments: ['get', 'list', 'watch', 'create', 'update'],
        statefulsets: ['get', 'list', 'watch'],
        services: ['get', 'list', 'watch', 'create', 'update'],
        configmaps: ['get', 'list', 'watch', 'create', 'update'],
        secrets: ['get', 'list', 'watch'],
      }

      matrix['viewer'] = {}
      for (const resource of resources) {
        matrix['viewer'][resource] = ['get', 'list', 'watch']
      }
    }

    matrixData.value = matrix
  } catch (e: any) {
    ElMessage.warning('Failed to load permissions')
  } finally {
    loading.value = false
  }
}

function hasPermission(role: string, resource: string, verb: string): boolean {
  return matrixData.value[role]?.[resource]?.includes(verb) || false
}

function togglePermission(role: string, resource: string, verb: string) {
  if (!matrixData.value[role]) {
    matrixData.value[role] = {}
  }
  if (!matrixData.value[role][resource]) {
    matrixData.value[role][resource] = []
  }

  const index = matrixData.value[role][resource].indexOf(verb)
  if (index >= 0) {
    matrixData.value[role][resource].splice(index, 1)
  } else {
    matrixData.value[role][resource].push(verb)
  }
}

function getResourcePermissions(role: string, resource: string): string[] {
  return matrixData.value[role]?.[resource] || []
}

function exportMatrix() {
  const data = {
    roles: roles.value,
    matrix: matrixData.value,
    exportedAt: new Date().toISOString(),
  }

  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'rbac-matrix.json'
  a.click()
  URL.revokeObjectURL(url)
  ElMessage.success('Matrix exported')
}

const filteredRoles = computed(() => {
  if (selectedRole.value) {
    return roles.value.filter(r => r.name === selectedRole.value)
  }
  return roles.value
})

onMounted(() => {
  fetchRoles()
  fetchPermissions()
})
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <h3 style="margin: 0;">RBAC Permission Matrix</h3>
        <div class="filter-right">
          <el-select v-model="selectedRole" placeholder="All Roles" clearable style="width: 150px;">
            <el-option v-for="role in roles" :key="role.name" :label="role.name" :value="role.name" />
          </el-select>
          <el-radio-group v-model="viewMode" size="small">
            <el-radio-button value="matrix">Matrix</el-radio-button>
            <el-radio-button value="list">List</el-radio-button>
          </el-radio-group>
          <el-button @click="exportMatrix"><el-icon><Download /></el-icon> Export</el-button>
          <el-button @click="fetchPermissions"><el-icon><Refresh /></el-icon></el-button>
        </div>
      </div>
    </el-card>

    <!-- Matrix View -->
    <el-card v-if="viewMode === 'matrix'" shadow="never" v-loading="loading">
      <div class="matrix-container">
        <table class="permission-matrix">
          <thead>
            <tr>
              <th class="resource-header">Resource / Role</th>
              <th v-for="role in filteredRoles" :key="role.name" class="role-header">
                {{ role.name }}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="resource in resources" :key="resource">
              <td class="resource-name">{{ resource }}</td>
              <td v-for="role in filteredRoles" :key="role.name" class="permission-cell">
                <div class="permission-verbs">
                  <el-tag
                    v-for="verb in verbs"
                    :key="verb"
                    :type="hasPermission(role.name, resource, verb) ? 'success' : 'info'"
                    size="small"
                    class="verb-tag"
                    @click="togglePermission(role.name, resource, verb)"
                  >
                    {{ verb.charAt(0).toUpperCase() }}
                  </el-tag>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="legend">
        <span class="legend-item">
          <el-tag type="success" size="small">G</el-tag> = Granted
        </span>
        <span class="legend-item">
          <el-tag type="info" size="small">G</el-tag> = Not Granted
        </span>
        <span class="legend-item">Click to toggle</span>
      </div>
    </el-card>

    <!-- List View -->
    <el-card v-else shadow="never" v-loading="loading">
      <el-collapse>
        <el-collapse-item v-for="role in filteredRoles" :key="role.name" :title="role.name" :name="role.name">
          <el-table :data="resources.map(r => ({ resource: r, permissions: getResourcePermissions(role.name, r) }))" size="small">
            <el-table-column prop="resource" label="Resource" min-width="200" />
            <el-table-column label="Permissions" min-width="400">
              <template #default="{ row }">
                <el-tag
                  v-for="verb in row.permissions"
                  :key="verb"
                  type="success"
                  size="small"
                  style="margin: 2px;"
                >
                  {{ verb }}
                </el-tag>
                <span v-if="row.permissions.length === 0" style="color: #909399;">No permissions</span>
              </template>
            </el-table-column>
          </el-table>
        </el-collapse-item>
      </el-collapse>
    </el-card>

    <!-- Summary -->
    <el-card shadow="never" style="margin-top: 16px;">
      <h4>Permission Summary</h4>
      <el-descriptions :column="3" border size="small">
        <el-descriptions-item v-for="role in filteredRoles" :key="role.name" :label="role.name">
          {{ Object.values(matrixData[role.name] || {}).flat().length }} permissions
        </el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; justify-content: space-between; align-items: center; }
.filter-right { display: flex; align-items: center; gap: 8px; }
.matrix-container { overflow-x: auto; }
.permission-matrix { border-collapse: collapse; width: 100%; }
.permission-matrix th, .permission-matrix td { border: 1px solid #ebeef5; padding: 8px; text-align: center; }
.resource-header { background: #f5f7fa; font-weight: bold; text-align: left; min-width: 180px; }
.role-header { background: #f5f7fa; font-weight: bold; min-width: 120px; }
.resource-name { text-align: left; font-weight: 500; background: #fafafa; }
.permission-cell { padding: 4px !important; }
.permission-verbs { display: flex; gap: 2px; justify-content: center; flex-wrap: wrap; }
.verb-tag { cursor: pointer; min-width: 24px; text-align: center; }
.verb-tag:hover { opacity: 0.8; }
.legend { margin-top: 12px; display: flex; gap: 16px; color: #909399; font-size: 12px; }
</style>
