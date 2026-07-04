<script setup lang="ts">
import { ref, onMounted } from 'vue'
import request from '@/api/request'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'

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
  } catch {
    // Silently handle — resource may not exist in cluster
  } finally {
    loading.value = false
  }
}

function handlePageChange(newPage: number) {
  page.value = newPage
  fetchRoles()
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchRoles)

onMounted(fetchRoles)
</script>

<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="margin: 0;">角色管理</h2>
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
    </div>

    <el-table :data="roleList" v-loading="loading" stripe style="width: 100%">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="角色名称" min-width="160" />
      <el-table-column prop="description" label="描述" min-width="300" show-overflow-tooltip />
      <el-table-column prop="createdAt" label="创建时间" min-width="180" />
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
