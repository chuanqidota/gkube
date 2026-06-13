<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Delete, Download, Search } from '@element-plus/icons-vue'
import request from '@/api/request'

const loading = ref(false)
const auditLogs = ref<any[]>([])
const selectedUser = ref('')
const selectedAction = ref('')
const selectedResource = ref('')
const selectedStatus = ref('')
const searchQuery = ref('')
const stats = ref<any>({})

const filteredLogs = computed(() => {
  let result = auditLogs.value
  if (selectedUser.value) {
    result = result.filter(l => l.user === selectedUser.value)
  }
  if (selectedAction.value) {
    result = result.filter(l => l.action === selectedAction.value)
  }
  if (selectedResource.value) {
    result = result.filter(l => l.resource === selectedResource.value)
  }
  if (selectedStatus.value) {
    result = result.filter(l => l.status === selectedStatus.value)
  }
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(l =>
      l.user?.toLowerCase().includes(query) ||
      l.action?.toLowerCase().includes(query) ||
      l.resource?.toLowerCase().includes(query) ||
      l.name?.toLowerCase().includes(query)
    )
  }
  return result
})

const users = computed(() => {
  return [...new Set(auditLogs.value.map(l => l.user).filter(Boolean))]
})

const actions = computed(() => {
  return [...new Set(auditLogs.value.map(l => l.action).filter(Boolean))]
})

const resources = computed(() => {
  return [...new Set(auditLogs.value.map(l => l.resource).filter(Boolean))]
})

async function fetchAuditLogs() {
  loading.value = true
  try {
    const res: any = await request.get('/k8s/audit/list')
    auditLogs.value = res.data || []
  } catch (e: any) {
    ElMessage.warning('Failed to load audit logs')
    auditLogs.value = []
  } finally {
    loading.value = false
  }
}

async function fetchStats() {
  try {
    const res: any = await request.get('/k8s/audit/stats')
    stats.value = res.data || {}
  } catch {
    stats.value = {}
  }
}

async function clearLogs() {
  try {
    await ElMessageBox.confirm('Clear all audit logs?', 'Confirm')
    await request.delete('/k8s/audit/clear')
    ElMessage.success('Audit logs cleared')
    fetchAuditLogs()
    fetchStats()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error('Failed to clear logs')
    }
  }
}

function exportLogs() {
  const data = JSON.stringify(filteredLogs.value, null, 2)
  const blob = new Blob([data], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'audit-logs.json'
  a.click()
  URL.revokeObjectURL(url)
}

function formatTime(time: string) {
  if (!time) return '-'
  return new Date(time).toLocaleString()
}

function actionType(action: string) {
  if (action === 'create' || action === 'update') return 'success'
  if (action === 'delete') return 'danger'
  return 'info'
}

function statusType(status: string) {
  return status === 'success' ? 'success' : 'danger'
}

onMounted(() => {
  fetchAuditLogs()
  fetchStats()
})
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <h3 style="margin: 0;">审计日志</h3>
        <div class="filter-right">
          <el-input v-model="searchQuery" placeholder="搜索..." :prefix-icon="Search" style="width: 200px;" clearable />
          <el-select v-model="selectedUser" placeholder="所有用户" clearable style="width: 120px;">
            <el-option v-for="u in users" :key="u" :label="u" :value="u" />
          </el-select>
          <el-select v-model="selectedAction" placeholder="所有操作" clearable style="width: 120px;">
            <el-option v-for="a in actions" :key="a" :label="a" :value="a" />
          </el-select>
          <el-select v-model="selectedResource" placeholder="所有资源" clearable style="width: 120px;">
            <el-option v-for="r in resources" :key="r" :label="r" :value="r" />
          </el-select>
          <el-select v-model="selectedStatus" placeholder="所有状态" clearable style="width: 100px;">
            <el-option label="成功" value="success" />
            <el-option label="失败" value="failure" />
          </el-select>
          <el-button @click="exportLogs"><el-icon><Download /></el-icon> 导出</el-button>
          <el-button type="danger" @click="clearLogs"><el-icon><Delete /></el-icon> 清除</el-button>
          <el-button type="primary" @click="fetchAuditLogs"><el-icon><Refresh /></el-icon> 刷新</el-button>
        </div>
      </div>
    </el-card>

    <el-row :gutter="16" style="margin-bottom: 16px;">
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-value">{{ stats.total || 0 }}</div>
          <div class="stat-label">总记录数</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card success">
          <div class="stat-value">{{ stats.byStatus?.success || 0 }}</div>
          <div class="stat-label">成功操作</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card danger">
          <div class="stat-value">{{ stats.byStatus?.failure || 0 }}</div>
          <div class="stat-label">失败操作</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-value">{{ Object.keys(stats.byUser || {}).length }}</div>
          <div class="stat-label">活跃用户</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="never">
      <el-table :data="filteredLogs" v-loading="loading" stripe>
        <el-table-column prop="timestamp" label="时间" width="180">
          <template #default="{ row }">{{ formatTime(row.timestamp) }}</template>
        </el-table-column>
        <el-table-column prop="user" label="用户" width="120" />
        <el-table-column prop="action" label="操作" width="100">
          <template #default="{ row }">
            <el-tag :type="actionType(row.action)" size="small">{{ row.action }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="resource" label="资源" width="120" />
        <el-table-column prop="name" label="资源名称" min-width="150" show-overflow-tooltip />
        <el-table-column prop="namespace" label="命名空间" width="120" />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="ip" label="IP 地址" width="120" />
        <el-table-column prop="error" label="错误信息" min-width="150" show-overflow-tooltip />
      </el-table>

      <el-empty v-if="!loading && filteredLogs.length === 0" description="暂无审计日志" />
    </el-card>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; justify-content: space-between; align-items: center; }
.filter-right { display: flex; align-items: center; gap: 8px; }
.stat-card { text-align: center; }
.stat-card.success { border-left: 4px solid #67C23A; }
.stat-card.danger { border-left: 4px solid #F56C6C; }
.stat-value { font-size: 32px; font-weight: bold; color: #303133; }
.stat-label { font-size: 14px; color: #909399; margin-top: 4px; }
</style>
