<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

const loading = ref(false)
const roleList = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const size = ref(10)

async function fetchRoles() {
  loading.value = true
  try {
    const res: any = await request.get('/roles', { params: { page: page.value, size: size.value } })
    roleList.value = res.data.items || []
    total.value = res.data.total || 0
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load roles')
  } finally {
    loading.value = false
  }
}

function handlePageChange(newPage: number) {
  page.value = newPage
  fetchRoles()
}

onMounted(fetchRoles)
</script>

<template>
  <div>
    <h2 style="margin-bottom: 16px;">Roles</h2>

    <el-table :data="roleList" v-loading="loading" stripe style="width: 100%">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="Role Name" min-width="160" />
      <el-table-column prop="description" label="Description" min-width="300" show-overflow-tooltip />
      <el-table-column prop="createdAt" label="Created At" min-width="180" />
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
  </div>
</template>
