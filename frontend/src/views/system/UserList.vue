<script setup lang="ts">
import { ref, reactive, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'
import request from '@/api/request'
import { useAuthStore } from '@/stores/auth'
import type { FormInstance, FormRules } from 'element-plus'
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const { t } = useI18n()
const loading = ref(false)
const authStore = useAuthStore()
const userList = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const size = ref(20)
const searchName = ref('')
let searchTimer: ReturnType<typeof setTimeout> | null = null
const selectedRows = ref<any[]>([])

const dialogVisible = ref(false)
const dialogTitle = ref(t('user.createUser'))
const formRef = ref<FormInstance>()
const saving = ref(false)
const editingId = ref<number | null>(null)

const resetDialogVisible = ref(false)
const resetTargetUser = ref<any>(null)
const resetSaving = ref(false)
const resetFormRef = ref<FormInstance>()

const resetForm = reactive({ newPassword: '' })

const resetRules: FormRules = {
  newPassword: [
    { required: true, message: t('user.passwordRequired'), trigger: 'blur' },
    { min: 6, message: t('user.passwordMinLength'), trigger: 'blur' },
  ],
}

const form = reactive({
  username: '',
  password: '',
  email: '',
  displayName: '',
})

const rules: FormRules = {
  username: [{ required: true, message: t('user.usernameRequired'), trigger: 'blur' }],
  password: [
    {
      trigger: 'blur',
      validator: (_rule: any, value: string, callback: any) => {
        if (!editingId.value) {
          if (!value) {
            callback(new Error(t('user.passwordRequired')))
          } else if (value.length < 6) {
            callback(new Error(t('user.passwordMinLength')))
          } else {
            callback()
          }
        } else {
          callback()
        }
      },
    },
  ],
}

function formatDate(_row: any, _col: any, cellValue: string) {
  if (!cellValue) return ''
  const d = new Date(cellValue)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

function onSearchInput(value: string) {
  searchName.value = value
  page.value = 1
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => fetchUsers(), 300)
}

function handleSelectionChange(rows: any[]) {
  selectedRows.value = rows
}

async function fetchUsers() {
  loading.value = true
  try {
    const params: any = { page: page.value, size: size.value }
    if (searchName.value) {
      params.keyword = searchName.value
    }
    const res: any = await request.get('/users', { params })
    userList.value = res.data.items || []
    total.value = res.data.total || 0
  } catch (e: any) {
    ElMessage.error(e?.message || t('user.loadUserListFailed'))
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = null
  dialogTitle.value = t('user.createUser')
  form.username = ''
  form.password = ''
  form.email = ''
  form.displayName = ''
  dialogVisible.value = true
}

function openEdit(row: any) {
  editingId.value = row.id
  dialogTitle.value = t('user.editUser')
  form.username = row.username || ''
  form.password = ''
  form.email = row.email || ''
  form.displayName = row.display_name || ''
  dialogVisible.value = true
}

async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  saving.value = true
  try {
    const payload: any = {
      username: form.username,
      email: form.email,
      displayName: form.displayName,
    }
    if (form.password) {
      payload.password = form.password
    }

    if (editingId.value) {
      payload.id = editingId.value
      await request.put('/users', payload)
      // 编辑的是当前登录用户时，同步更新 Header 显示（reassignment 触发 shallow watch）
      if (authStore.user && editingId.value === authStore.user.id) {
        authStore.setUser({
          ...authStore.user,
          ...(payload.displayName !== undefined ? { display_name: payload.displayName } : {}),
          ...(payload.email !== undefined ? { email: payload.email } : {}),
        })
      }
      ElMessage.success(t('user.userUpdated'))
    } else {
      await request.post('/users', payload)
      ElMessage.success(t('user.userCreated'))
    }
    dialogVisible.value = false
    fetchUsers()
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.saveFailed'))
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(t('user.deleteUserConfirm', { name: row.username }), t('common.confirmDelete'), { type: 'warning' })
  } catch {
    return // 用户取消确认框
  }
  try {
    await request.delete('/users', { data: { id: row.id } })
    ElMessage.success(t('common.deleted'))
    fetchUsers()
  } catch (e: any) {
    ElMessage.error(e?.message || t('user.deleteFailed'))
  }
}

async function handleBatchDelete() {
  if (!selectedRows.value.length) return
  try {
    await ElMessageBox.confirm(
      t('common.batchDeleteConfirm', { count: selectedRows.value.length, type: t('user.title') }),
      t('common.confirmDelete'),
      { type: 'warning' }
    )
    const results = await Promise.allSettled(
      selectedRows.value.map((row) => request.delete('/users', { data: { id: row.id } }))
    )
    const successCount = results.filter((r) => r.status === 'fulfilled').length
    const failCount = results.filter((r) => r.status === 'rejected').length
    if (failCount > 0) {
      ElMessage.warning(t('user.batchDeleteResult', { success: successCount, fail: failCount }))
    } else {
      ElMessage.success(t('user.batchDeleteSuccess', { count: successCount }))
    }
    fetchUsers()
  } catch {
    // cancelled
  }
}

function openResetPassword(row: any) {
  resetTargetUser.value = row
  resetForm.newPassword = ''
  resetDialogVisible.value = true
}

async function handleResetPassword() {
  const valid = await resetFormRef.value?.validate().catch(() => false)
  if (!valid) return

  resetSaving.value = true
  try {
    await request.put('/users/reset-password', {
      userId: resetTargetUser.value.id,
      newPassword: resetForm.newPassword,
    })
    ElMessage.success(t('user.passwordResetSuccess'))
    resetDialogVisible.value = false
  } catch (e: any) {
    ElMessage.error(e?.message || t('user.resetPasswordFailed'))
  } finally {
    resetSaving.value = false
  }
}

function handlePageChange(newPage: number) {
  page.value = newPage
  fetchUsers()
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchUsers)

onMounted(fetchUsers)

onBeforeUnmount(() => {
  if (searchTimer) clearTimeout(searchTimer)
})
</script>

<template>
  <div class="page-container">
    <ResourceListToolbar
      :search-value="searchName"
      :total-count="total"
      :selected-count="selectedRows.length"
      :show-namespace="false"
      :search-placeholder="t('user.searchPlaceholder')"
      @search-input="onSearchInput"
    >
      <template #actions>
        <el-button type="success" @click="openCreate">
          <el-icon><Plus /></el-icon> {{ t('common.create') }}
        </el-button>
        <el-button type="danger" :disabled="!selectedRows.length" @click="handleBatchDelete">
          <el-icon><Delete /></el-icon> {{ t('common.delete') }} ({{ selectedRows.length }})
        </el-button>
      </template>
      <template #extra>
        <AutoRefreshToolbar
          :is-running="isRunning"
          :countdown="countdown"
          :current-interval="currentInterval"
          :available-intervals="availableIntervals"
          :loading="loading"
          @refresh="manualRefresh()"
          @toggle="toggle()"
          @interval-change="setIntervalOption"
        />
      </template>
    </ResourceListToolbar>

    <el-card shadow="never" class="table-card">
      <el-table
        :data="userList"
        v-loading="loading"
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="45" />
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" :label="t('user.username')" min-width="140" />
        <el-table-column prop="display_name" :label="t('user.nickname')" min-width="140" show-overflow-tooltip />
        <el-table-column prop="email" :label="t('user.email')" min-width="200" />
        <el-table-column prop="status" :label="t('user.status')" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? t('user.enabled') : t('user.disabled') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" :label="t('user.createdAt')" min-width="180" :formatter="formatDate" />
        <el-table-column :label="t('common.actions')" width="230" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
            <el-button size="small" type="warning" @click="openEdit(row)">{{ t('common.edit') }}</el-button>
            <el-button size="small" type="warning" @click="openResetPassword(row)">{{ t('user.resetPassword') }}</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">{{ t('common.delete') }}</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="total > size" class="pagination">
        <el-pagination
          :current-page="page"
          :page-size="size"
          :total="total"
          layout="prev, pager, next"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>

    <!-- Create / Edit Dialog -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="480px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item :label="t('user.username')" prop="username">
          <el-input v-model="form.username" :disabled="!!editingId" />
        </el-form-item>
        <el-form-item :label="t('user.password')" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            :placeholder="editingId ? t('user.leaveBlankToKeep') : ''"
          />
        </el-form-item>
        <el-form-item :label="t('user.nickname')" prop="displayName">
          <el-input v-model="form.displayName" />
        </el-form-item>
        <el-form-item :label="t('user.email')" prop="email">
          <el-input v-model="form.email" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">
          {{ editingId ? t('common.update') : t('common.create') }}
        </el-button>
      </template>
    </el-dialog>

    <!-- Reset Password Dialog -->
    <el-dialog v-model="resetDialogVisible" :title="t('user.resetPassword')" width="420px" destroy-on-close>
      <el-form ref="resetFormRef" :model="resetForm" :rules="resetRules" label-width="100px">
        <el-form-item :label="t('user.userLabel')">
          <el-input :model-value="resetTargetUser?.username" disabled />
        </el-form-item>
        <el-form-item :label="t('user.newPassword')" prop="newPassword">
          <el-input
            v-model="resetForm.newPassword"
            type="password"
            show-password
            :placeholder="t('user.newPasswordPlaceholder')"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetDialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="resetSaving" @click="handleResetPassword">{{ t('user.confirmReset') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container {
  padding: var(--gk-space-5);
}
.table-card {
  border-radius: var(--gk-radius-md);
}
.action-buttons {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: var(--gk-space-1);
}
.action-buttons .el-button + .el-button {
  margin-left: 0;
}
.pagination {
  display: flex;
  justify-content: flex-end;
  padding: 12px 0;
  border-top: 1px solid var(--el-border-color-lighter);
}
</style>
