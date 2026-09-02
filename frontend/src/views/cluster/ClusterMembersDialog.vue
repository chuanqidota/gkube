<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, User, ArrowDown, ArrowUp } from '@element-plus/icons-vue'
import { getRoles, getClusterMembers, deleteBinding, removeClusterMember } from '@/api/rbac'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'
import AddBindingDialog from './AddBindingDialog.vue'
import MemberEditDialog from './MemberEditDialog.vue'

interface MemberBinding {
  bindingId: number
  roleId: number
  roleName: string
  roleDisplayName: string
  scopeType: 'cluster' | 'namespace'
  namespace: string
}
interface Member {
  userId: number
  username: string
  displayName: string
  isSuperAdmin: boolean
  bindings: MemberBinding[]
}

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
const members = ref<Member[]>([])
const roles = ref<any[]>([])
const searchQuery = ref('')
const addDialogVisible = ref(false)
const editDialogVisible = ref(false)
const editingMember = ref<Member | null>(null)
const expandedMembers = ref<Set<number>>(new Set())

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
  mobileHandler = (e: MediaQueryListEvent) => { isMobile.value = e.matches }
  window.matchMedia('(max-width: 768px)').addEventListener('change', mobileHandler)
})
onUnmounted(() => {
  if (mobileHandler) {
    window.matchMedia('(max-width: 768px)').removeEventListener('change', mobileHandler)
  }
})

const isAdmin = computed(() => authStore.user?.isAdmin || authStore.user?.isSuperAdmin || false)

const filteredMembers = computed(() => {
  if (!searchQuery.value.trim()) return members.value
  const query = searchQuery.value.trim().toLowerCase()
  return members.value.filter(m =>
    (m.username || '').toLowerCase().includes(query) ||
    (m.displayName || '').toLowerCase().includes(query) ||
    m.bindings.some(b =>
      (b.roleDisplayName || '').toLowerCase().includes(query) ||
      (b.namespace || '').toLowerCase().includes(query)
    )
  )
})

watch(() => props.visible, (val) => {
  if (val && props.clusterId) fetchData()
})

async function fetchData() {
  expandedMembers.value = new Set()
  loading.value = true
  try {
    const [membersRes, rolesRes] = await Promise.all([
      getClusterMembers(props.clusterId),
      getRoles(),
    ])
    const mData: any = membersRes?.data ?? membersRes
    members.value = (mData?.items || []) as Member[]
    const rData: any = rolesRes?.data ?? rolesRes
    roles.value = Array.isArray(rData) ? rData : (rData?.items || [])
  } catch (e: any) {
    ElMessage.error(e?.message || t('rbac.loadFailed'))
  } finally {
    loading.value = false
  }
}

function handleAdd() {
  addDialogVisible.value = true
}

function handleAddSuccess() {
  fetchData()
  authStore.fetchPermissions()
}

function handleEditMember(m: Member) {
  editingMember.value = m
  editDialogVisible.value = true
}

function handleEditSuccess() {
  fetchData()
  authStore.fetchPermissions()
}

function toggleExpand(userId: number) {
  const s = new Set(expandedMembers.value)
  if (s.has(userId)) {
    s.delete(userId)
  } else {
    s.add(userId)
  }
  expandedMembers.value = s
}

function isExpanded(userId: number) {
  return expandedMembers.value.has(userId)
}

// 按组删除：删除该成员在指定角色下的全部绑定
async function handleRemoveRole(m: Member, b: MemberBinding) {
  const label = b.scopeType === 'cluster'
    ? b.roleDisplayName
    : `${b.roleDisplayName} (${b.namespace})`
  try {
    await ElMessageBox.confirm(
      t('rbac.removeRoleConfirm', { name: escapeHtml(m.username), role: escapeHtml(label) }),
      t('common.confirm'),
      { type: 'warning', dangerouslyUseHTMLString: true }
    )
  } catch { return }
  try {
    await deleteBinding(b.bindingId)
    ElMessage.success(t('rbac.removeRoleSuccess'))
    fetchData()
    authStore.fetchPermissions()
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.deleteFailed'))
  }
}

// 移出集群：删除该成员全部绑定
async function handleRemoveMember(m: Member) {
  try {
    await ElMessageBox.confirm(
      t('rbac.removeMemberConfirm', { name: escapeHtml(m.username) }),
      t('common.confirm'),
      { type: 'warning', dangerouslyUseHTMLString: true }
    )
  } catch { return }
  try {
    await removeClusterMember(m.userId, props.clusterId)
    ElMessage.success(t('rbac.removeMemberSuccess'))
    fetchData()
    authStore.fetchPermissions()
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.deleteFailed'))
  }
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
    '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;',
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
    destroy-on-close
  >
    <template #header>
      <div class="drawer-header">
        <div class="header-title">
          <el-icon :size="20" style="color: var(--gk-color-primary); margin-right: 8px;"><User /></el-icon>
          <span>{{ t('rbac.clusterMembers', { cluster: clusterName }) }}</span>
        </div>
        <div class="header-stats">
          <el-tag size="small" type="info">{{ t('rbac.totalMembers', { count: members.length }) }}</el-tag>
        </div>
      </div>
    </template>

    <div class="drawer-body">
      <div class="toolbar">
        <el-input
          v-model="searchQuery"
          :placeholder="t('rbac.searchMembers')"
          clearable
          style="width: 220px;"
          size="default"
        />
        <el-button v-if="isAdmin" type="primary" @click="handleAdd">
          <el-icon><Plus /></el-icon> {{ t('rbac.addMember') }}
        </el-button>
      </div>

      <div v-loading="loading" class="member-cards">
        <el-empty v-if="!loading && filteredMembers.length === 0" :description="t('rbac.noMembers')" />

        <el-card
          v-for="m in filteredMembers"
          :key="m.userId"
          shadow="never"
          class="member-card"
        >
          <div class="member-head">
            <div class="member-info">
              <el-avatar :size="36" class="member-avatar">
                {{ (m.username || '?')[0].toUpperCase() }}
              </el-avatar>
              <div class="member-names">
                <span class="username">{{ m.username }}</span>
                <span v-if="m.displayName" class="display-name">{{ m.displayName }}</span>
              </div>
            </div>
            <div v-if="isAdmin" class="member-actions">
              <el-button size="small" type="primary" plain @click="handleEditMember(m)">
                {{ t('common.edit') }}
              </el-button>
              <el-button size="small" type="danger" plain @click="handleRemoveMember(m)">
                {{ t('rbac.removeMember') }}
              </el-button>
            </div>
          </div>

          <div class="member-bindings">
            <span v-if="m.isSuperAdmin" class="super-admin-badge">Super Admin</span>

            <!-- 收起态 -->
            <div
              v-if="!isExpanded(m.userId) && m.bindings.length > 0"
              class="binding-fold-bar"
              @click="toggleExpand(m.userId)"
            >
              <el-icon><ArrowDown /></el-icon>
              <span>{{ t('rbac.bindingCount', { count: m.bindings.length }) }}</span>
            </div>

            <!-- 展开态 -->
            <template v-if="isExpanded(m.userId)">
              <div v-for="b in m.bindings" :key="b.bindingId" class="binding-row">
                <el-tag :type="roleTagType(b.roleName)" size="small">{{ b.roleDisplayName }}</el-tag>
                <el-tag v-if="b.scopeType === 'cluster'" size="small" type="info">
                  {{ t('rbac.clusterScope') }}
                </el-tag>
                <el-tag v-else type="warning" size="small">{{ b.namespace }}</el-tag>
                <el-button
                  v-if="isAdmin"
                  size="small"
                  type="danger"
                  text
                  @click="handleRemoveRole(m, b)"
                >
                  {{ t('rbac.removeBinding') }}
                </el-button>
              </div>
              <div class="binding-fold-bar" @click="toggleExpand(m.userId)">
                <el-icon><ArrowUp /></el-icon>
                <span>{{ t('rbac.collapseBindings') }}</span>
              </div>
            </template>

            <span v-if="m.bindings.length === 0" class="no-role">{{ t('rbac.noRole') }}</span>
          </div>
        </el-card>
      </div>
    </div>
  </el-drawer>

  <AddBindingDialog
    v-model:visible="addDialogVisible"
    :cluster-id="clusterId"
    :cluster-name="clusterName"
    :roles="roles"
    @success="handleAddSuccess"
  />
  <MemberEditDialog
    v-if="editingMember"
    v-model:visible="editDialogVisible"
    :cluster-id="clusterId"
    :member="editingMember"
    :roles="roles"
    @success="handleEditSuccess"
  />
</template>

<style scoped>
.drawer-header { display: flex; flex-direction: column; gap: 8px; }
.header-title { display: flex; align-items: center; font-size: 16px; font-weight: 600; }
.header-stats { display: flex; gap: 8px; flex-wrap: wrap; }
.drawer-body { display: flex; flex-direction: column; gap: 16px; }
.toolbar { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 12px; }
.member-cards { display: flex; flex-direction: column; gap: 12px; min-height: 120px; }
.member-card { border: 1px solid var(--gk-color-border); }
.member-head { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 8px; }
.member-info { display: flex; align-items: center; gap: 12px; }
.member-avatar { background: var(--gk-color-primary); color: #fff; flex-shrink: 0; }
.member-names { display: flex; flex-direction: column; }
.username { font-weight: 600; }
.display-name { font-size: 12px; color: var(--gk-color-text-secondary); }
.member-actions { display: flex; gap: 8px; }
.member-bindings { margin-top: 12px; display: flex; flex-direction: column; gap: 6px; }
.binding-row { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.no-role { color: var(--gk-color-text-disabled); font-size: 12px; }
.binding-fold-bar {
  display: flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  color: var(--gk-color-primary);
  font-size: 13px;
  padding: 2px 0;
  user-select: none;
}
.binding-fold-bar:hover { opacity: 0.8; }
.super-admin-badge {
  background: linear-gradient(135deg, #f56c6c 0%, #e6a23c 100%);
  color: white; padding: 2px 8px; border-radius: 4px;
  font-size: 12px; font-weight: 600; align-self: flex-start;
}
</style>
