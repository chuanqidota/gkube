<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, User } from '@element-plus/icons-vue'
import { getBindings, deleteBinding, getRoles, getClusterMembers } from '@/api/rbac'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'
import AddBindingDialog from './AddBindingDialog.vue'

const props = defineProps<{
  visible: boolean
  clusterId: number
  clusterName: string
}>()

const emit = defineEmits<{
  'update:visible': [val: boolean]
}>()

const { t } = useI18n()
const authStore = useAuthStore()
const uiStore = useUIStore()
const loading = ref(false)
const bindings = ref<any[]>([])
const roles = ref<any[]>([])
const selectedRows = ref<any[]>([])
const searchQuery = ref('')
const addDialogVisible = ref(false)
const editDialogVisible = ref(false)
const editingBinding = ref<any>(null)

// 响应式抽屉宽度：移动端全屏，桌面端 720px
const isMobile = ref(window.matchMedia('(max-width: 768px)').matches)
let mobileHandler: ((e: MediaQueryListEvent) => void) | undefined
// 抽屉宽度：撑到左侧菜单栏
const drawerSize = computed(() => {
  if (isMobile.value) return '100%'
  return uiStore.sidebarCollapsed
    ? 'calc(100vw - var(--gk-sidebar-collapsed-width))'
    : 'calc(100vw - var(--gk-sidebar-width))'
})

onMounted(() => {
  mobileHandler = (e: MediaQueryListEvent) => { isMobile.value = e.matches }
  window.matchMedia('(max-width: 768px)').addEventListener('change', mobileHandler)
})
onUnmounted(() => {
  if (mobileHandler) {
    window.matchMedia('(max-width: 768px)').removeEventListener('change', mobileHandler)
  }
})

const isAdmin = computed(() => authStore.user?.isAdmin || authStore.user?.isSuperAdmin || false)

watch(() => props.visible, (val) => {
  if (val && props.clusterId) {
    fetchData()
  }
})

// 过滤后的数据（仅搜索）
const filteredData = computed(() => {
  let data = bindings.value

  if (searchQuery.value.trim()) {
    const query = searchQuery.value.trim().toLowerCase()
    data = data.filter(b =>
      (b.username || '').toLowerCase().includes(query) ||
      (b.displayName || '').toLowerCase().includes(query)
    )
  }

  return data
})

// 统计数据
const stats = computed(() => ({
  total: bindings.value.length,
}))

async function fetchData() {
  loading.value = true
  try {
    const [bindingsRes, rolesRes, membersRes] = await Promise.allSettled([
      getBindings({ clusterId: props.clusterId }),
      getRoles(),
      getClusterMembers(props.clusterId),
    ])

    if (bindingsRes.status === 'fulfilled') {
      const bRes: any = bindingsRes.value
      const bData = bRes?.data ?? bRes
      // getBindings 返回 { items: [...], total }，每项含嵌套 user/role 对象
      const rawItems: any[] = bData.items || bData || []
      // 扁平化：把嵌套的 user/role 提取到顶层，保留原始对象用于编辑
      bindings.value = rawItems.map((b: any) => ({
        ...b,
        _raw: b,
        username: b.user?.username || b.username || '',
        displayName: b.user?.display_name || b.displayName || '',
        roleName: b.role?.name || b.roleName || '',
        roleDisplayName: b.role?.displayName || b.roleDisplayName || '',
        isSuperAdmin: b.user?.isSuperAdmin || false,
        userID: b.userId,
      }))
    }

    if (rolesRes.status === 'fulfilled') {
      const rRes: any = rolesRes.value
      const rData = rRes?.data ?? rRes
      // getRoles 返回数组
      roles.value = Array.isArray(rData) ? rData : (rData?.items || [])
    }

    // 用 members 数据补全未绑定的成员
    if (membersRes.status === 'fulfilled') {
      const mRes: any = membersRes.value
      const mData = mRes?.data ?? mRes
      // getClusterMembers 返回 { items: [...], total }，拦截器已解包一层到 response.data
      const members: any[] = mData.items || []
      const existingIds = new Set(bindings.value.map((b: any) => b.userID || b.userId))
      const unboundMembers = members
        .filter((m: any) => !existingIds.has(m.userID || m.userId))
        .map((m: any) => ({
          id: `member-${m.userID || m.userId}`,
          userID: m.userID || m.userId,
          userId: m.userID || m.userId,
          username: m.username,
          displayName: m.displayName,
          roleName: m.roleName,
          roleDisplayName: m.roleDisplayName || m.roleName,
          namespace: m.namespace || '',
          isSuperAdmin: m.isSuperAdmin,
          unbound: true,
        }))
      if (unboundMembers.length > 0) {
        bindings.value = [...bindings.value, ...unboundMembers]
      }
    }
  } catch (e: any) {
    ElMessage.error(e?.message || t('rbac.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function handleDelete(row: any) {
  const name = row.displayName || row.username
  try {
    await ElMessageBox.confirm(
      t('rbac.deleteConfirm', { name: escapeHtml(name) }),
      t('common.confirm'),
      { type: 'warning', dangerouslyUseHTMLString: true }
    )
  } catch {
    return
  }

  try {
    await deleteBinding(row.id)
    ElMessage.success(t('rbac.bindingDeleted'))
    fetchData()
    authStore.fetchPermissions()
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.deleteFailed'))
  }
}

async function handleBatchDelete() {
  if (selectedRows.value.length === 0) return

  const names = selectedRows.value.map(r => escapeHtml(r.displayName || r.username))
  const nameList = names.map(n => `<li>${n}</li>`).join('')
  try {
    await ElMessageBox.confirm(
      `<p>${t('rbac.batchDeleteConfirm', { count: selectedRows.value.length })}</p><ul>${nameList}</ul>`,
      t('common.confirm'),
      { type: 'warning', dangerouslyUseHTMLString: true }
    )
  } catch {
    return
  }

  const deleteResults = await Promise.allSettled(
    selectedRows.value.map(row => deleteBinding(row.id))
  )

  const succeeded = deleteResults.filter(r => r.status === 'fulfilled').length
  const failed = deleteResults.filter(r => r.status === 'rejected').length

  if (succeeded > 0) {
    ElMessage.success(t('rbac.batchDeleteSuccess', { count: succeeded }))
    authStore.fetchPermissions()
  }
  if (failed > 0) {
    ElMessage.error(t('rbac.batchDeleteFailed', { count: failed }))
  }

  selectedRows.value = []
  fetchData()
}

function handleSelectionChange(rows: any[]) {
  selectedRows.value = rows
}

function handleAdd() {
  addDialogVisible.value = true
}

function handleAddSuccess() {
  fetchData()
  authStore.fetchPermissions()
}

function handleEdit(row: any) {
  // 从原始绑定对象中提取编辑所需数据
  const raw = row._raw || row
  editingBinding.value = {
    id: raw.id,
    userId: raw.userId || raw.user?.id,
    username: raw.user?.username || row.username,
    displayName: raw.user?.display_name || row.displayName,
    roleId: raw.roleId || raw.role?.id,
    roleName: raw.role?.name || row.roleName,
    roleDisplayName: raw.role?.displayName || row.roleDisplayName,
    clusterName: raw.clusterName,
    scopeType: raw.scopeType || (raw.namespace ? 'namespace' : 'cluster'),
    namespace: raw.namespace || '',
  }
  editDialogVisible.value = true
}

function handleEditSuccess() {
  fetchData()
  authStore.fetchPermissions()
}

function roleTagType(roleName: string): string {
  if (!roleName) return 'info'
  if (roleName.includes('admin')) return 'danger'
  if (roleName.includes('editor')) return 'primary'
  if (roleName.includes('viewer')) return 'success'
  return 'info'
}

function escapeHtml(str: string): string {
  const map: Record<string, string> = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;',
  }
  return str.replace(/[&<>"']/g, c => map[c])
}
</script>

<template>
  <el-drawer
    :model-value="visible"
    @update:model-value="emit('update:visible', $event)"
    :title="t('rbac.clusterMembers', { cluster: clusterName })"
    :size="drawerSize"
    :close-on-click-modal="!addDialogVisible"
    destroy-on-close
  >
    <template #header>
      <div class="drawer-header">
        <div class="header-title">
          <el-icon :size="20" style="color: var(--gk-color-primary); margin-right: 8px;"><User /></el-icon>
          <span>{{ t('rbac.clusterMembers', { cluster: clusterName }) }}</span>
        </div>
        <div class="header-stats">
          <el-tag size="small" type="info">{{ t('rbac.totalMembers', { count: stats.total }) }}</el-tag>
        </div>
      </div>
    </template>

    <div class="drawer-body">
      <!-- 搜索 + 操作栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <el-input
            v-model="searchQuery"
            :placeholder="t('rbac.searchMembers')"
            clearable
            style="width: 220px;"
            size="default"
          />
        </div>
        <div class="toolbar-right">
          <el-button
            v-if="isAdmin"
            type="danger"
            plain
            size="default"
            :disabled="selectedRows.length === 0"
            @click="handleBatchDelete"
          >
            <el-icon><Delete /></el-icon> {{ t('rbac.batchDelete') }}{{ selectedRows.length > 0 ? ` (${selectedRows.length})` : '' }}
          </el-button>
          <el-button
            v-if="isAdmin"
            type="primary"
            size="default"
            @click="handleAdd"
          >
            <el-icon><Plus /></el-icon> {{ t('rbac.addMember') }}
          </el-button>
        </div>
      </div>

      <!-- 成员列表 -->
      <el-table
        :data="filteredData"
        v-loading="loading"
        stripe
        @selection-change="handleSelectionChange"
        style="width: 100%;"
        :empty-text="t('rbac.noMembers')"
      >
        <el-table-column
          v-if="isAdmin"
          type="selection"
          width="45"
        />
        <el-table-column :label="t('rbac.username')" min-width="120" show-overflow-tooltip>
          <template #default="{ row }">
            <div class="user-info">
              <span class="username">{{ row.username }}</span>
              <span v-if="row.displayName" class="display-name">{{ row.displayName }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column :label="t('rbac.namespace')" width="140" show-overflow-tooltip>
          <template #default="{ row }">
            <el-tag v-if="row.namespace" type="warning" size="small">{{ row.namespace }}</el-tag>
            <span v-else class="no-role">—</span>
          </template>
        </el-table-column>
        <el-table-column :label="t('rbac.roleName')" min-width="130">
          <template #default="{ row }">
            <div class="role-cell">
              <el-tag
                v-if="row.roleDisplayName"
                :type="roleTagType(row.roleName)"
                size="small"
              >
                {{ row.roleDisplayName }}
              </el-tag>
              <span v-else-if="row.isSuperAdmin" class="super-admin-badge">Super Admin</span>
              <span v-else class="no-role">{{ t('rbac.noRole') }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column
          v-if="isAdmin"
          :label="t('rbac.actions')"
          width="160"
          align="center"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              v-if="!row.unbound"
              size="small"
              type="primary"
              plain
              @click="handleEdit(row)"
            >
              {{ t('common.edit') }}
            </el-button>
            <el-button
              v-if="!row.unbound"
              size="small"
              type="danger"
              plain
              @click="handleDelete(row)"
            >
              {{ t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </el-drawer>

  <!-- 添加成员对话框 -->
  <AddBindingDialog
    v-model:visible="addDialogVisible"
    :cluster-id="clusterId"
    :cluster-name="clusterName"
    :roles="roles"
    @success="handleAddSuccess"
  />
  <!-- 编辑成员对话框 -->
  <AddBindingDialog
    v-if="editingBinding"
    v-model:visible="editDialogVisible"
    :cluster-id="clusterId"
    :cluster-name="clusterName"
    :roles="roles"
    :binding="editingBinding"
    @success="handleEditSuccess"
  />
</template>

<style scoped>
.drawer-header {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.header-title {
  display: flex;
  align-items: center;
  font-size: 16px;
  font-weight: 600;
}
.header-stats {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.drawer-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: 100%;
}
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}
.toolbar-left {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}
.toolbar-right {
  display: flex;
  gap: 8px;
  align-items: center;
}
.user-info {
  display: flex;
  flex-direction: column;
}
.username {
  font-weight: 500;
}
.display-name {
  font-size: 12px;
  color: var(--gk-color-text-secondary);
}
.role-cell {
  display: flex;
  align-items: center;
}
.super-admin-badge {
  background: linear-gradient(135deg, #f56c6c 0%, #e6a23c 100%);
  color: white;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}
.no-role {
  color: var(--gk-color-text-disabled);
  font-size: 12px;
}
</style>
