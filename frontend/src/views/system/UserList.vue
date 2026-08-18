<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'
import request from '@/api/request'
import { useAuthStore } from '@/stores/auth'
import type { FormInstance, FormRules } from 'element-plus'
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

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
const dialogTitle = ref('创建用户')
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
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' },
  ],
}

const form = reactive({
  username: '',
  password: '',
  email: '',
  displayName: '',
})

const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [
    {
      trigger: 'blur',
      validator: (_rule: any, value: string, callback: any) => {
        if (!editingId.value) {
          if (!value) {
            callback(new Error('请输入密码'))
          } else if (value.length < 6) {
            callback(new Error('密码长度不能少于6位'))
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
    ElMessage.error(e?.message || '获取用户列表失败')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = null
  dialogTitle.value = '创建用户'
  form.username = ''
  form.password = ''
  form.email = ''
  form.displayName = ''
  dialogVisible.value = true
}

function openEdit(row: any) {
  editingId.value = row.id
  dialogTitle.value = '编辑用户'
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
      // 编辑的是当前登录用户时，同步更新 Header 显示
      if (authStore.user && editingId.value === authStore.user.id) {
        if (payload.displayName !== undefined) authStore.user.display_name = payload.displayName
        if (payload.email !== undefined) authStore.user.email = payload.email
      }
      ElMessage.success('用户已更新')
    } else {
      await request.post('/users', payload)
      ElMessage.success('用户已创建')
    }
    dialogVisible.value = false
    fetchUsers()
  } catch (e: any) {
    ElMessage.error(e?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(`确定删除用户 "${row.username}" 吗？`, '确认删除', { type: 'warning' })
    await request.delete('/users', { data: { id: row.id } })
    ElMessage.success('已删除')
    fetchUsers()
  } catch {
    // cancelled
  }
}

async function handleBatchDelete() {
  if (!selectedRows.value.length) return
  try {
    await ElMessageBox.confirm(
      `确定删除选中的 ${selectedRows.value.length} 个用户吗？`,
      '确认删除',
      { type: 'warning' }
    )
    const results = await Promise.allSettled(
      selectedRows.value.map((row) => request.delete('/users', { data: { id: row.id } }))
    )
    const successCount = results.filter((r) => r.status === 'fulfilled').length
    const failCount = results.filter((r) => r.status === 'rejected').length
    if (failCount > 0) {
      ElMessage.warning(`已删除 ${successCount} 个，失败 ${failCount} 个`)
    } else {
      ElMessage.success(`已删除 ${successCount} 个用户`)
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
    ElMessage.success('密码重置成功')
    resetDialogVisible.value = false
  } catch (e: any) {
    ElMessage.error(e?.message || '重置密码失败')
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
</script>

<template>
  <div class="page-container">
    <ResourceListToolbar
      :search-value="searchName"
      :total-count="total"
      :selected-count="selectedRows.length"
      :show-namespace="false"
      search-placeholder="搜索用户名、昵称或邮箱"
      @search-input="onSearchInput"
    >
      <template #actions>
        <el-button type="success" @click="openCreate">
          <el-icon><Plus /></el-icon> 创建用户
        </el-button>
        <el-button type="danger" :disabled="!selectedRows.length" @click="handleBatchDelete">
          <el-icon><Delete /></el-icon> 删除 ({{ selectedRows.length }})
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
        <el-table-column prop="username" label="用户名" min-width="140" />
        <el-table-column prop="display_name" label="昵称" min-width="140" show-overflow-tooltip />
        <el-table-column prop="email" label="邮箱" min-width="200" />
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" min-width="180" :formatter="formatDate" />
        <el-table-column label="操作" width="230" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="warning" @click="openResetPassword(row)">重置密码</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
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
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" :disabled="!!editingId" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            :placeholder="editingId ? '留空保持不变' : ''"
          />
        </el-form-item>
        <el-form-item label="昵称" prop="displayName">
          <el-input v-model="form.displayName" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">
          {{ editingId ? '更新' : '创建' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- Reset Password Dialog -->
    <el-dialog v-model="resetDialogVisible" title="重置密码" width="420px" destroy-on-close>
      <el-form ref="resetFormRef" :model="resetForm" :rules="resetRules" label-width="100px">
        <el-form-item label="用户">
          <el-input :model-value="resetTargetUser?.username" disabled />
        </el-form-item>
        <el-form-item label="新密码" prop="newPassword">
          <el-input
            v-model="resetForm.newPassword"
            type="password"
            show-password
            placeholder="请输入新密码（不少于6位）"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="resetSaving" @click="handleResetPassword">确认重置</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container {
  padding: 20px;
}
.table-card {
  border-radius: 8px;
}
.action-buttons {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: 4px;
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
