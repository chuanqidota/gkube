<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Bell, Warning, CircleCheck, InfoFilled } from '@element-plus/icons-vue'
import request from '@/api/request'

const loading = ref(false)
const events = ref<any[]>([])
const namespaces = ref<string[]>([])
const selectedNamespace = ref('')
const selectedType = ref('')
const searchQuery = ref('')
const autoRefresh = ref(true)
let refreshTimer: ReturnType<typeof setInterval> | null = null

const eventTypes = [
  { value: 'Normal', label: '正常', type: 'info' },
  { value: 'Warning', label: '警告', type: 'warning' },
]

const filteredEvents = computed(() => {
  let result = events.value
  if (selectedNamespace.value) {
    result = result.filter(e => e.namespace === selectedNamespace.value)
  }
  if (selectedType.value) {
    result = result.filter(e => e.type === selectedType.value)
  }
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(e =>
      e.reason?.toLowerCase().includes(query) ||
      e.message?.toLowerCase().includes(query) ||
      e.involvedObject?.name?.toLowerCase().includes(query)
    )
  }
  return result
})

function eventType(type: string) {
  return type === 'Warning' ? 'warning' : 'info'
}

function eventIcon(type: string) {
  return type === 'Warning' ? Warning : InfoFilled
}

function formatTime(time: string) {
  if (!time) return '-'
  const date = new Date(time)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes} 分钟前`
  if (hours < 24) return `${hours} 小时前`
  return `${days} 天前`
}

async function fetchEvents() {
  loading.value = true
  try {
    const res: any = await request.get('/k8s/event/list')
    events.value = res.data || []
  } catch (e: any) {
    ElMessage.warning('Failed to load events')
    events.value = []
  } finally {
    loading.value = false
  }
}

async function fetchNamespaces() {
  try {
    const res: any = await request.get('/k8s/namespace/list')
    namespaces.value = res.data?.map((ns: any) => ns.name) || []
  } catch {
    namespaces.value = []
  }
}

function startAutoRefresh() {
  if (refreshTimer) clearInterval(refreshTimer)
  if (autoRefresh.value) {
    refreshTimer = setInterval(fetchEvents, 30000)
  }
}

function toggleAutoRefresh() {
  autoRefresh.value = !autoRefresh.value
  startAutoRefresh()
}

onMounted(() => {
  fetchEvents()
  fetchNamespaces()
  startAutoRefresh()
})
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <h3 style="margin: 0;"><el-icon><Bell /></el-icon> 事件查看器</h3>
        <div class="filter-right">
          <el-input v-model="searchQuery" placeholder="搜索事件..." style="width: 250px;" clearable />
          <el-select v-model="selectedNamespace" placeholder="所有命名空间" clearable style="width: 150px;">
            <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
          </el-select>
          <el-select v-model="selectedType" placeholder="所有类型" clearable style="width: 120px;">
            <el-option v-for="t in eventTypes" :key="t.value" :label="t.label" :value="t.value" />
          </el-select>
          <el-button :type="autoRefresh ? 'success' : 'info'" @click="toggleAutoRefresh">
            {{ autoRefresh ? '自动刷新' : '手动刷新' }}
          </el-button>
          <el-button type="primary" @click="fetchEvents"><el-icon><Refresh /></el-icon> 刷新</el-button>
        </div>
      </div>
    </el-card>

    <el-card shadow="never">
      <el-table :data="filteredEvents" v-loading="loading" stripe>
        <el-table-column label="类型" width="80">
          <template #default="{ row }">
            <el-icon :style="{ color: row.type === 'Warning' ? '#E6A23C' : '#409EFF' }">
              <component :is="eventIcon(row.type)" />
            </el-icon>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="120" />
        <el-table-column label="资源" min-width="150">
          <template #default="{ row }">
            <div>
              <div style="font-weight: 500;">{{ row.involvedObject?.name }}</div>
              <div style="font-size: 12px; color: #909399;">{{ row.involvedObject?.kind }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="reason" label="原因" width="150" />
        <el-table-column prop="message" label="消息" min-width="300" show-overflow-tooltip />
        <el-table-column label="时间" width="150">
          <template #default="{ row }">{{ formatTime(row.lastTimestamp || row.eventTime) }}</template>
        </el-table-column>
        <el-table-column prop="count" label="次数" width="80" />
      </el-table>

      <el-empty v-if="!loading && filteredEvents.length === 0" description="暂无事件" />
    </el-card>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; justify-content: space-between; align-items: center; }
.filter-right { display: flex; align-items: center; gap: 8px; }
</style>
