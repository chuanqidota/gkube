<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '@/api/request'
import type { FormInstance, FormRules } from 'element-plus'

const loading = ref(false)
const userList = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const size = ref(10)

const dialogVisible = ref(false)
const dialogTitle = ref('Create User')
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
  username: [{ required: true, message: 'Username is required', trigger: 'blur' }],
  password: [
    {
      required: true,
      message: 'Password is required',
      trigger: 'blur',
      // Only required when creating (not editing)
      validator: (_rule: any, value: string, callback: any) => {
        if (!editingId.value && !value) {
          callback(new Error('Password is required'))
        } else {
          callback()
        }
      },
    },
  ],
}

async function fetchUsers() {
  loading.value = true
  try {
    const res: any = await request.get('/users', { params: { page: page.value, size: size.value } })
    userList.value = res.data.items || []
    total.value = res.data.total || 0
  } catch {
    // Silently handle — resource may not exist in cluster
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = null
  dialogTitle.value = 'Create User'
  form.username = ''
  form.password = ''
  form.email = ''
  form.nickname = ''
  dialogVisible.value = true
}

function openEdit(row: any) {
  editingId.value = row.id
  dialogTitle.value = 'Edit User'
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
      ElMessage.success('User updated')
    } else {
      payload.password = form.password
      await request.post('/users', payload)
      ElMessage.success('User created')
    }
    dialogVisible.value = false
    fetchUsers()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Save failed')
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(`Delete user "${row.username}"?`, 'Confirm', { type: 'warning' })
    await request.delete('/users', { data: { id: row.id } })
    ElMessage.success('Deleted')
    fetchUsers()
  } catch {
    // cancelled
  }
}

function handlePageChange(newPage: number) {
  page.value = newPage
  fetchUsers()
}

onMounted(fetchUsers)
</script>

<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="margin: 0;">Users</h2>
      <el-button type="primary" @click="openCreate">Create User</el-button>
    </div>

    <el-table :data="userList" v-loading="loading" stripe style="width: 100%">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="username" label="Username" min-width="140" />
      <el-table-column prop="nickname" label="Nickname" min-width="140" />
      <el-table-column prop="email" label="Email" min-width="200" />
      <el-table-column prop="createdAt" label="Created At" min-width="180" />
      <el-table-column label="Actions" width="160" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="openEdit(row)">Edit</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">Delete</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div style="display: flex; justify-content: flex-end; margin-top: 16px;">
      <el-pagination
        v-if="total > size"
        :current-page="page"
        :page-size="size"
        :total="total"
        layout="prev, pager, next"
        @current-change="handlePageChange"
      />
    </div>

    <!-- Create / Edit Dialog -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="480px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="Username" prop="username">
          <el-input v-model="form.username" :disabled="!!editingId" />
        </el-form-item>
        <el-form-item label="Password" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            :placeholder="editingId ? 'Leave blank to keep current' : ''"
          />
        </el-form-item>
        <el-form-item label="Nickname" prop="nickname">
          <el-input v-model="form.nickname" />
        </el-form-item>
        <el-form-item label="Email" prop="email">
          <el-input v-model="form.email" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">Cancel</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">
          {{ editingId ? 'Update' : 'Create' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>
