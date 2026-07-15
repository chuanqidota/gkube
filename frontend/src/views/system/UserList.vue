<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'
import request from '@/api/request'
import type { FormInstance, FormRules } from 'element-plus'
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const loading = ref(false)
const userList = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const size = ref(20)
const searchName = ref('')
const selectedRows = ref<any[]>([])

const dialogVisible = ref(false)
const dialogTitle = ref('创建用户')
const formRef = ref<FormInstance>()
const saving = ref(false)
const editingId = ref<number | null>(null)

const form = reactive({
  username: '',
  password: '',
  email: '',
  nickname: '',
})

const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [
    {
      required: true,
      message: '请输入密码',
      trigger: 'blur',
      validator: (_rule: any, value: string, callback: any) => {
        if (!editingId.value && !value) {
          callback(new Error('请输入密码'))
        } else {
          callback()
        }
      },
    },
  ],
}

const filteredList = computed(() => {
  if (!searchName.value) return userList.value
  const keyword = searchName.value.toLowerCase()
  return userList.value.filter(
    (u) =>
      u.username?.toLowerCase().includes(keyword) ||
      u.nickname?.toLowerCase().includes(keyword) ||
      u.email?.toLowerCase().includes(keyword)
  )
})

function onSearchInput(value: string) {
  searchName.value = value
}

function handleSelectionChange(rows: any[]) {
  selectedRows.value = rows
}

async function fetchUsers() {
  loading.value = true
  try {
    const res: any = await request.get('/users', { params: { page: page.value, size: size.value } })
    userList.value = res.data.items || []
    total.value = res.data.total || 0
  } catch {
    // Silently handle
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
  form.nickname = ''
  dialogVisible.value = true
}

function openEdit(row: any) {
  editingId.value = row.id
  dialogTitle.value = '编辑用户'
  form.username = row.username || ''
  form.password = ''
  form.email = row.email || ''
  form.nickname = row.nickname || ''
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
      nickname: form.nickname,
    }
    if (form.password) {
      payload.password = form.password
    }

    if (editingId.value) {
      await request.put(`/users/${editingId.value}`, payload)
      ElMessage.success('用户已更新')
    } else {
      payload.password = form.password
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
        :data="filteredList"
        v-loading="loading"
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="45" />
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" min-width="140" />
        <el-table-column prop="nickname" label="昵称" min-width="140" />
        <el-table-column prop="email" label="邮箱" min-width="200" />
        <el-table-column prop="createdAt" label="创建时间" min-width="180" />
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
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
        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="form.nickname" />
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
  </div>
</template>

<style scoped>
.page-container {
  padding: 20px;
}
.table-card {
  border-radius: 8px;
}
.pagination {
  display: flex;
  justify-content: flex-end;
  padding: 12px 0;
  border-top: 1px solid var(--el-border-color-lighter);
}
</style>
