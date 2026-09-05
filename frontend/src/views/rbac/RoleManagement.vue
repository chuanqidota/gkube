<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getRoles, createRole, updateRole, deleteRole, getResourceDict } from '@/api/rbac'
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'
import { useUIStore } from '@/stores/ui'

interface RoleItem {
  id: number
  name: string
  displayName: string
  scopeType: 'cluster' | 'namespace'
  isSystem: boolean
  description?: string
  permissions: Record<string, string[]>
}
interface ResourceGroupDef {
  group: string
  verbs: string[]
  clusterOnly: boolean
}

const { t } = useI18n()
const uiStore = useUIStore()
const loading = ref(false)
const roles = ref<RoleItem[]>([])
const dict = ref<ResourceGroupDef[]>([])
const searchName = ref('')

// 响应式抽屉宽度：移动端全屏，桌面端撑到左侧菜单栏
const isMobile = ref(window.matchMedia('(max-width: 768px)').matches)
let mobileHandler: ((e: MediaQueryListEvent) => void) | undefined
const drawerSize = computed(() => {
  if (isMobile.value) return '100%'
  return uiStore.sidebarCollapsed
    ? 'calc(100vw - var(--gk-sidebar-collapsed-width))'
    : 'calc(100vw - var(--gk-sidebar-width))'
})

onMounted(() => {
  fetchData()
  mobileHandler = (e: MediaQueryListEvent) => { isMobile.value = e.matches }
  window.matchMedia('(max-width: 768px)').addEventListener('change', mobileHandler)
})
onUnmounted(() => {
  if (mobileHandler) {
    window.matchMedia('(max-width: 768px)').removeEventListener('change', mobileHandler)
  }
})

// 矩阵编辑对话框
const editVisible = ref(false)
const editRole = ref<RoleItem | null>(null)
const editForm = ref({
  name: '',
  displayName: '',
  scopeType: 'namespace' as 'cluster' | 'namespace',
  description: '',
  permissions: {} as Record<string, string[]>,
})
const saving = ref(false)

const filteredRoles = computed(() => {
  if (!searchName.value.trim()) return roles.value
  const q = searchName.value.trim().toLowerCase()
  return roles.value.filter(r =>
    r.name.toLowerCase().includes(q) || r.displayName.toLowerCase().includes(q)
  )
})

// 当前编辑作用域下可配置的资源组（clusterOnly 组仅 cluster 作用域）
const availableGroups = computed(() => {
  if (editForm.value.scopeType === 'cluster') return dict.value
  return dict.value.filter(d => !d.clusterOnly)
})

// 矩阵表头：字典中出现过的全部动词（去重）
const allVerbs = computed(() => {
  const s = new Set<string>()
  for (const d of dict.value) d.verbs.forEach(v => s.add(v))
  return [...s].sort()
})

// 类型归一：后端正常返回对象；若返回 JSON 字符串也能解析
function parsePerms(p: any): Record<string, string[]> {
  if (!p) return {}
  if (typeof p === 'string') {
    try { return JSON.parse(p) } catch { return {} }
  }
  return p
}

async function fetchData() {
  loading.value = true
  try {
    const [rolesRes, dictRes] = await Promise.all([
      getRoles(),
      getResourceDict(),
    ])
    const rData: any = rolesRes?.data ?? rolesRes
    roles.value = (Array.isArray(rData) ? rData : rData?.items || []) as RoleItem[]
    const dData: any = dictRes?.data ?? dictRes
    dict.value = (Array.isArray(dData) ? dData : dData?.items || []) as ResourceGroupDef[]
  } catch (e: any) {
    ElMessage.error(e?.message || t('rbac.loadFailed'))
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editRole.value = null
  editForm.value = { name: '', displayName: '', scopeType: 'namespace', description: '', permissions: {} }
  editVisible.value = true
}

function openEdit(role: RoleItem) {
  if (role.isSystem) {
    // 预置角色：复制为新角色起点
    editRole.value = null
    editForm.value = {
      name: role.name + '-copy',
      displayName: role.displayName + '(副本)',
      scopeType: role.scopeType,
      description: '',
      permissions: parsePerms(role.permissions),
    }
  } else {
    editRole.value = role
    editForm.value = {
      name: role.name,
      displayName: role.displayName,
      scopeType: role.scopeType,
      description: role.description || '',
      permissions: parsePerms(role.permissions),
    }
  }
  editVisible.value = true
}

async function handleDelete(role: RoleItem) {
  try {
    await ElMessageBox.confirm(
      t('rbac.deleteRoleConfirm', { name: role.displayName }),
      t('common.confirm'),
      { type: 'warning' }
    )
  } catch { return }
  try {
    await deleteRole(role.id)
    ElMessage.success(t('common.deleted'))
    fetchData()
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.deleteFailed'))
  }
}

// 矩阵勾选联动
function toggleVerb(group: string, verb: string, checked: boolean) {
  const perms = { ...editForm.value.permissions }
  const cur = perms[group] || []
  if (checked) {
    if (!cur.includes(verb)) cur.push(verb)
  } else {
    const i = cur.indexOf(verb)
    if (i >= 0) cur.splice(i, 1)
  }
  if (cur.length === 0) delete perms[group]
  else perms[group] = cur
  editForm.value.permissions = perms
}

function isVerbOn(group: string, verb: string): boolean {
  return (editForm.value.permissions[group] || []).includes(verb)
}

// 切换作用域时清除 clusterOnly 组的配置
function handleScopeChange() {
  if (editForm.value.scopeType === 'namespace') {
    const clusterOnlyGroups = new Set(dict.value.filter(d => d.clusterOnly).map(d => d.group))
    const perms: Record<string, string[]> = {}
    for (const [g, verbs] of Object.entries(editForm.value.permissions)) {
      if (!clusterOnlyGroups.has(g)) perms[g] = verbs
    }
    editForm.value.permissions = perms
  }
}

async function handleSave() {
  if (!editForm.value.name || !editForm.value.displayName) {
    ElMessage.warning(t('rbac.roleFormRequired'))
    return
  }
  saving.value = true
  try {
    if (editRole.value) {
      await updateRole(editRole.value.id, {
        displayName: editForm.value.displayName,
        description: editForm.value.description,
        permissions: editForm.value.permissions,
      })
    } else {
      await createRole({
        name: editForm.value.name,
        displayName: editForm.value.displayName,
        scopeType: editForm.value.scopeType,
        description: editForm.value.description,
        permissions: editForm.value.permissions,
      })
    }
    ElMessage.success(t('rbac.updateSuccess'))
    editVisible.value = false
    fetchData()
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.failed'))
  } finally {
    saving.value = false
  }
}

</script>

<template>
  <div class="page-container">
    <ResourceListToolbar
      :search-value="searchName"
      :total-count="roles.length"
      :show-namespace="false"
      :search-placeholder="t('rbac.searchRoles')"
      @search-input="searchName = $event"
    >
      <template #actions>
        <el-button type="success" @click="openCreate">
          <el-icon><Plus /></el-icon> {{ t('rbac.createRole') }}
        </el-button>
      </template>
    </ResourceListToolbar>

    <el-card shadow="never" class="table-card">
      <el-table :data="filteredRoles" v-loading="loading" stripe>
        <el-table-column prop="displayName" :label="t('rbac.roleName')" min-width="140" />
        <el-table-column prop="name" :label="t('rbac.roleIdent')" min-width="140" />
        <el-table-column :label="t('rbac.scope')" width="120" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="row.scopeType === 'cluster' ? 'danger' : 'warning'">
              {{ row.scopeType === 'cluster' ? t('rbac.clusterScope') : t('rbac.namespaceScope') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('rbac.source')" width="100" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="row.isSystem ? 'info' : 'success'">
              {{ row.isSystem ? t('rbac.presetRole') : t('rbac.customRole') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" :label="t('rbac.roleDescLabel')" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">{{ row.description || '-' }}</template>
        </el-table-column>
        <el-table-column :label="t('rbac.actions')" width="200" fixed="right" align="center">
          <template #default="{ row }">
            <el-button size="small" type="primary" plain @click="openEdit(row)">
              {{ row.isSystem ? t('rbac.copyRole') : t('common.edit') }}
            </el-button>
            <el-button v-if="!row.isSystem" size="small" type="danger" plain @click="handleDelete(row)">
              {{ t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 矩阵编辑抽屉 -->
    <el-drawer v-model="editVisible" :title="editRole ? t('rbac.editRole') : t('rbac.createRole')" :size="drawerSize" direction="rtl">
      <el-form label-width="80px">
        <el-form-item :label="t('rbac.roleName')">
          <el-input v-model="editForm.displayName" :disabled="!!editRole" style="width: 240px;" />
        </el-form-item>
        <el-form-item v-if="!editRole" :label="t('rbac.roleIdent')">
          <el-input v-model="editForm.name" style="width: 240px;" :placeholder="t('rbac.roleIdentPlaceholder')" />
        </el-form-item>
        <el-form-item :label="t('rbac.scope')">
          <el-radio-group v-model="editForm.scopeType" :disabled="!!editRole" @change="handleScopeChange">
            <el-radio value="namespace">{{ t('rbac.namespaceScope') }}</el-radio>
            <el-radio value="cluster">{{ t('rbac.clusterScope') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="t('rbac.roleDescLabel')">
          <el-input v-model="editForm.description" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>

      <el-divider content-position="left">{{ t('rbac.permissionMatrix') }}</el-divider>
      <el-table :data="availableGroups" size="small" border max-height="360">
        <el-table-column prop="group" :label="t('rbac.resourceGroup')" width="140" />
        <el-table-column v-for="v in allVerbs" :key="v" :label="v" width="90" align="center">
          <template #default="{ row }">
            <el-checkbox
              v-if="row.verbs.includes(v)"
              :model-value="isVerbOn(row.group, v)"
              @change="(val: any) => toggleVerb(row.group, v, !!val)"
            />
            <span v-else class="na">-</span>
          </template>
        </el-table-column>
      </el-table>

      <el-alert type="warning" :closable="false" style="margin-top: 12px;">
        {{ t('rbac.roleCacheWarning') }}
      </el-alert>

      <template #footer>
        <el-button @click="editVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">{{ t('common.save') }}</el-button>
      </template>
    </el-drawer>
  </div>
</template>

<style scoped>
.na { color: var(--gk-color-text-disabled); }
</style>
