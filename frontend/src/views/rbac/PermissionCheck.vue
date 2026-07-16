<script setup lang="ts">
import { ref, computed } from 'vue'
import { Search, Warning } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import {
  getRoleList,
  getClusterRoleList,
  getRoleBindingList,
  getClusterRoleBindingList,
} from '@/api/resource'

// Query form
const subjectType = ref<'User' | 'Group' | 'ServiceAccount'>('User')
const subjectName = ref('')
const namespaceFilter = ref('')
const loading = ref(false)
const hasSearched = ref(false)

// Results
interface PermissionRow {
  resource: string
  verbs: string[]
  scope: 'Cluster' | string // 'Cluster' or namespace name
  roleRef: string
  roleKind: string
}

const permissions = ref<PermissionRow[]>([])
const boundRoles = ref<{ name: string; kind: string; namespace: string; scope: string }[]>([])

const filteredPermissions = computed(() => {
  if (!namespaceFilter.value) return permissions.value
  return permissions.value.filter(
    (p) => p.scope === 'Cluster' || p.scope === namespaceFilter.value
  )
})

// Group permissions by resource for summary
const resourceSummary = computed(() => {
  const map = new Map<string, { verbs: Set<string>; scopes: Set<string> }>()
  for (const p of filteredPermissions.value) {
    const key = p.resource
    if (!map.has(key)) map.set(key, { verbs: new Set(), scopes: new Set() })
    const entry = map.get(key)!
    for (const v of p.verbs) entry.verbs.add(v)
    entry.scopes.add(p.scope)
  }
  return Array.from(map.entries()).map(([resource, { verbs, scopes }]) => ({
    resource,
    verbs: Array.from(verbs).sort(),
    scopes: Array.from(scopes).sort(),
  }))
})

async function handleSearch() {
  if (!subjectName.value.trim()) {
    ElMessage.warning('请输入主体名称')
    return
  }
  loading.value = true
  hasSearched.value = true
  permissions.value = []
  boundRoles.value = []

  try {
    // Fetch all bindings and roles in parallel
    const [roleBindingsRes, clusterRoleBindingsRes, rolesRes, clusterRolesRes] = await Promise.all([
      getRoleBindingList().catch(() => ({ data: [] })),
      getClusterRoleBindingList().catch(() => ({ data: [] })),
      getRoleList().catch(() => ({ data: [] })),
      getClusterRoleList().catch(() => ({ data: [] })),
    ])

    const roleBindings = (roleBindingsRes.data?.items || roleBindingsRes.data || []) as any[]
    const clusterRoleBindings = (clusterRoleBindingsRes.data?.items || clusterRoleBindingsRes.data || []) as any[]
    const roles = (rolesRes.data?.items || rolesRes.data || []) as any[]
    const clusterRoles = (clusterRolesRes.data?.items || clusterRolesRes.data || []) as any[]

    // Find bindings that reference the subject
    const matchedBindings: { roleRef: string; roleKind: string; namespace: string; scope: string }[] = []

    for (const rb of roleBindings) {
      const subjects = rb.subjectList || parseSubjects(rb.subjects)
      if (matchesSubject(subjects)) {
        matchedBindings.push({
          roleRef: rb.roleRefName || rb.roleRef || '',
          roleKind: rb.roleRefKind || 'Role',
          namespace: rb.namespace || '',
          scope: rb.namespace || 'default',
        })
      }
    }

    for (const crb of clusterRoleBindings) {
      const subjects = crb.subjectList || parseSubjects(crb.subjects)
      if (matchesSubject(subjects)) {
        matchedBindings.push({
          roleRef: crb.roleRefName || crb.roleRef || '',
          roleKind: crb.roleRefKind || 'ClusterRole',
          namespace: '',
          scope: 'Cluster',
        })
      }
    }

    // Resolve role rules
    const allPermissions: PermissionRow[] = []

    for (const binding of matchedBindings) {
      let roleData: any = null

      if (binding.roleKind === 'ClusterRole') {
        roleData = clusterRoles.find((cr: any) => cr.name === binding.roleRef)
      } else {
        roleData = roles.find((r: any) => r.name === binding.roleRef && r.namespace === binding.namespace)
        // Also check ClusterRole referenced by RoleBinding
        if (!roleData) {
          roleData = clusterRoles.find((cr: any) => cr.name === binding.roleRef)
        }
      }

      if (roleData && roleData.rules) {
        for (const rule of roleData.rules) {
          const resources = rule.resources || []
          const verbs = rule.verbs || []
          for (const resource of resources) {
            allPermissions.push({
              resource,
              verbs,
              scope: binding.scope,
              roleRef: binding.roleRef,
              roleKind: binding.roleKind,
            })
          }
        }
      }

      boundRoles.value.push({
        name: binding.roleRef,
        kind: binding.roleKind,
        namespace: binding.namespace,
        scope: binding.scope,
      })
    }

    permissions.value = allPermissions

    if (allPermissions.length === 0) {
      ElMessage.info('未找到该主体的权限绑定')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '查询失败')
  } finally {
    loading.value = false
  }
}

function parseSubjects(subjects: any): { kind: string; name: string; namespace?: string }[] {
  if (typeof subjects === 'string') {
    // subjects might be a display string like "User:admin, Group:devs"
    return []
  }
  if (Array.isArray(subjects)) return subjects
  return []
}

function matchesSubject(subjects: { kind: string; name: string }[]): boolean {
  const name = subjectName.value.trim()
  const type = subjectType.value
  return subjects.some((s) => {
    if (type === 'ServiceAccount') {
      return s.kind === 'ServiceAccount' && s.name === name
    }
    return s.kind === type && s.name === name
  })
}

function verbTagType(verb: string): string {
  switch (verb) {
    case 'get':
    case 'list':
    case 'watch':
      return 'success'
    case 'create':
      return ''
    case 'update':
    case 'patch':
      return 'warning'
    case 'delete':
    case 'deletecollection':
      return 'danger'
    case '*':
      return 'danger'
    default:
      return 'info'
  }
}
</script>

<template>
  <div class="page-container">
    <!-- Search Form -->
    <el-card shadow="never" class="filter-card">
      <el-form :inline="true" @submit.prevent="handleSearch">
        <el-form-item label="主体类型">
          <el-select v-model="subjectType" style="width: 160px;">
            <el-option label="User" value="User" />
            <el-option label="Group" value="Group" />
            <el-option label="ServiceAccount" value="ServiceAccount" />
          </el-select>
        </el-form-item>
        <el-form-item label="主体名称">
          <el-input
            v-model="subjectName"
            placeholder="输入用户名、组名或SA名称"
            style="width: 240px;"
            clearable
            @keyup.enter="handleSearch"
          >
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
        </el-form-item>
        <el-form-item label="命名空间过滤">
          <el-input
            v-model="namespaceFilter"
            placeholder="可选，过滤特定命名空间"
            style="width: 200px;"
            clearable
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="handleSearch">
            <el-icon><Search /></el-icon> 查询权限
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- Empty state -->
    <el-card v-if="hasSearched && !loading && permissions.length === 0" shadow="never" class="table-card">
      <el-empty description="未找到该主体的权限绑定">
        <template #image>
          <el-icon :size="64" color="#c0c4cc"><Warning /></el-icon>
        </template>
      </el-empty>
    </el-card>

    <!-- Results -->
    <template v-if="permissions.length > 0">
      <!-- Summary by resource -->
      <el-card shadow="never" class="table-card" style="margin-bottom: 16px;">
        <template #header>
          <div class="card-header">
            <span>权限摘要（按资源类型）</span>
            <el-tag type="info" size="small">共 {{ resourceSummary.length }} 种资源</el-tag>
          </div>
        </template>
        <el-table :data="resourceSummary" stripe size="small">
          <el-table-column prop="resource" label="资源类型" min-width="180" />
          <el-table-column label="允许的操作" min-width="300">
            <template #default="{ row }">
              <el-tag
                v-for="verb in row.verbs"
                :key="verb"
                :type="verbTagType(verb)"
                size="small"
                style="margin: 2px;"
              >
                {{ verb }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="作用域" min-width="200">
            <template #default="{ row }">
              <el-tag
                v-for="scope in row.scopes"
                :key="scope"
                :type="scope === 'Cluster' ? 'danger' : 'info'"
                size="small"
                style="margin: 2px;"
              >
                {{ scope }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- Bound Roles -->
      <el-card shadow="never" class="table-card" style="margin-bottom: 16px;">
        <template #header>
          <div class="card-header">
            <span>绑定的角色</span>
            <el-tag type="info" size="small">共 {{ boundRoles.length }} 个</el-tag>
          </div>
        </template>
        <el-table :data="boundRoles" stripe size="small">
          <el-table-column prop="kind" label="类型" width="140">
            <template #default="{ row }">
              <el-tag :type="row.kind === 'ClusterRole' ? 'danger' : ''" size="small">
                {{ row.kind }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="name" label="角色名称" min-width="200" />
          <el-table-column prop="scope" label="作用域" width="160">
            <template #default="{ row }">
              <el-tag :type="row.scope === 'Cluster' ? 'danger' : 'info'" size="small">
                {{ row.scope }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- Detailed permissions -->
      <el-card shadow="never" class="table-card">
        <template #header>
          <div class="card-header">
            <span>详细权限规则</span>
            <el-tag type="info" size="small">共 {{ filteredPermissions.length }} 条</el-tag>
          </div>
        </template>
        <el-table :data="filteredPermissions" stripe>
          <el-table-column prop="roleRef" label="来源角色" min-width="160" show-overflow-tooltip>
            <template #default="{ row }">
              <el-tag :type="row.roleKind === 'ClusterRole' ? 'danger' : ''" size="small" style="margin-right: 4px;">
                {{ row.roleKind }}
              </el-tag>
              {{ row.roleRef }}
            </template>
          </el-table-column>
          <el-table-column prop="scope" label="作用域" width="140">
            <template #default="{ row }">
              <el-tag :type="row.scope === 'Cluster' ? 'danger' : 'info'" size="small">
                {{ row.scope }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="resource" label="资源" min-width="180" />
          <el-table-column label="操作" min-width="280">
            <template #default="{ row }">
              <el-tag
                v-for="verb in row.verbs"
                :key="verb"
                :type="verbTagType(verb)"
                size="small"
                style="margin: 2px;"
              >
                {{ verb }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </template>
  </div>
</template>

<style scoped>
.page-container {
  padding: 20px;
}
.filter-card {
  margin-bottom: 16px;
}
.table-card {
  border-radius: 8px;
}
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
</style>
