<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Monitor, Cpu, Coin, FolderOpened, Bell, Connection } from '@element-plus/icons-vue'
import request from '@/api/request'
import * as echarts from 'echarts'

const loading = ref(false)
const refreshTimer = ref<ReturnType<typeof setInterval> | null>(null)

// Stats
const stats = ref({
  clusters: 0,
  nodes: 0,
  pods: 0,
  namespaces: 0,
  deployments: 0,
  services: 0,
})

// Resource usage
const resourceUsage = ref({
  cpu: { used: 0, total: 0 },
  memory: { used: 0, total: 0 },
  storage: { used: 0, total: 0 },
})

// Recent events
const recentEvents = ref<any[]>([])

// Charts
const cpuChartRef = ref<HTMLElement | null>(null)
const memoryChartRef = ref<HTMLElement | null>(null)
const podChartRef = ref<HTMLElement | null>(null)

async function fetchStats() {
  try {
    const [clusterRes, nodeRes, podRes, nsRes, deployRes, svcRes] = await Promise.all([
      request.get('/clusters'),
      request.get('/k8s/metrics/nodes'),
      request.get('/k8s/metrics/pods'),
      request.get('/k8s/namespace/list'),
      request.get('/k8s/deployment/list'),
      request.get('/k8s/service/list'),
    ])

    stats.value = {
      clusters: clusterRes.data?.total || 0,
      nodes: nodeRes.data?.length || 0,
      pods: podRes.data?.length || 0,
      namespaces: nsRes.data?.length || 0,
      deployments: deployRes.data?.length || 0,
      services: svcRes.data?.length || 0,
    }
  } catch (e: any) {
    console.error('Failed to fetch stats:', e)
  }
}

async function fetchResourceUsage() {
  try {
    const res: any = await request.get('/dashboard/resources')
    if (res.data) {
      resourceUsage.value = res.data
    }
  } catch {
    // Use default values
  }
}

async function fetchRecentEvents() {
  try {
    const res: any = await request.get('/k8s/event/list')
    recentEvents.value = (res.data || []).slice(0, 10)
  } catch {
    recentEvents.value = []
  }
}

function updateCharts() {
  if (!cpuChartRef.value || !memoryChartRef.value || !podChartRef.value) return

  const cpuChart = echarts.init(cpuChartRef.value)
  const memChart = echarts.init(memoryChartRef.value)
  const podChart = echarts.init(podChartRef.value)

  // CPU Chart
  const cpuPercent = resourceUsage.value.cpu.total > 0
    ? Math.round((resourceUsage.value.cpu.used / resourceUsage.value.cpu.total) * 100)
    : 0

  cpuChart.setOption({
    title: { text: 'CPU 使用率', left: 'center', textStyle: { fontSize: 14 } },
    series: [{
      type: 'gauge',
      progress: { show: true, width: 18 },
      axisTick: { show: false },
      splitLine: { show: false },
      axisLabel: { show: false },
      pointer: { show: false },
      detail: { valueAnimation: true, formatter: '{value}%', fontSize: 24, offsetCenter: [0, '70%'] },
      data: [{ value: cpuPercent }],
      itemStyle: { color: cpuPercent >= 80 ? '#F56C6C' : cpuPercent >= 60 ? '#E6A23C' : '#409EFF' }
    }]
  })

  // Memory Chart
  const memPercent = resourceUsage.value.memory.total > 0
    ? Math.round((resourceUsage.value.memory.used / resourceUsage.value.memory.total) * 100)
    : 0

  memChart.setOption({
    title: { text: '内存使用率', left: 'center', textStyle: { fontSize: 14 } },
    series: [{
      type: 'gauge',
      progress: { show: true, width: 18 },
      axisTick: { show: false },
      splitLine: { show: false },
      axisLabel: { show: false },
      pointer: { show: false },
      detail: { valueAnimation: true, formatter: '{value}%', fontSize: 24, offsetCenter: [0, '70%'] },
      data: [{ value: memPercent }],
      itemStyle: { color: memPercent >= 80 ? '#F56C6C' : memPercent >= 60 ? '#E6A23C' : '#409EFF' }
    }]
  })

  // Pod Distribution Chart
  podChart.setOption({
    title: { text: 'Pod 分布', left: 'center', textStyle: { fontSize: 14 } },
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    series: [{
      type: 'pie',
      radius: ['40%', '70%'],
      avoidLabelOverlap: false,
      itemStyle: { borderRadius: 10, borderColor: '#fff', borderWidth: 2 },
      label: { show: false },
      emphasis: { label: { show: true, fontSize: 14, fontWeight: 'bold' } },
      labelLine: { show: false },
      data: [
        { value: stats.value.deployments, name: 'Deployments' },
        { value: stats.value.services, name: 'Services' },
        { value: stats.value.pods, name: 'Pods' },
      ]
    }]
  })

  window.addEventListener('resize', () => {
    cpuChart.resize()
    memChart.resize()
    podChart.resize()
  })
}

function eventType(type: string) {
  return type === 'Warning' ? 'warning' : 'info'
}

function formatTime(time: string) {
  if (!time) return '-'
  const date = new Date(time)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)

  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes} 分钟前`
  return `${hours} 小时前`
}

async function fetchAll() {
  loading.value = true
  try {
    await Promise.all([fetchStats(), fetchResourceUsage(), fetchRecentEvents()])
    setTimeout(updateCharts, 100)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchAll()
  refreshTimer.value = setInterval(fetchAll, 60000)
})

onUnmounted(() => {
  if (refreshTimer.value) {
    clearInterval(refreshTimer.value)
  }
})
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <h3 style="margin: 0;"><el-icon><Monitor /></el-icon> 系统概览</h3>
        <el-button type="primary" @click="fetchAll"><el-icon><Refresh /></el-icon> 刷新</el-button>
      </div>
    </el-card>

    <!-- Stats Cards -->
    <el-row :gutter="16" style="margin-bottom: 16px;">
      <el-col :span="4">
        <el-card shadow="never" class="stat-card">
          <div class="stat-icon" style="background: #409EFF;"><el-icon><Connection /></el-icon></div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.clusters }}</div>
            <div class="stat-label">集群</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card shadow="never" class="stat-card">
          <div class="stat-icon" style="background: #67C23A;"><el-icon><Cpu /></el-icon></div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.nodes }}</div>
            <div class="stat-label">节点</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card shadow="never" class="stat-card">
          <div class="stat-icon" style="background: #E6A23C;"><el-icon><Coin /></el-icon></div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.pods }}</div>
            <div class="stat-label">Pod</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card shadow="never" class="stat-card">
          <div class="stat-icon" style="background: #F56C6C;"><el-icon><FolderOpened /></el-icon></div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.namespaces }}</div>
            <div class="stat-label">命名空间</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card shadow="never" class="stat-card">
          <div class="stat-icon" style="background: #909399;"><el-icon><Monitor /></el-icon></div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.deployments }}</div>
            <div class="stat-label">Deployment</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card shadow="never" class="stat-card">
          <div class="stat-icon" style="background: #606266;"><el-icon><Connection /></el-icon></div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.services }}</div>
            <div class="stat-label">Service</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" style="margin-bottom: 16px;">
      <el-col :span="8">
        <el-card shadow="never">
          <div ref="cpuChartRef" style="height: 250px;"></div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never">
          <div ref="memoryChartRef" style="height: 250px;"></div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never">
          <div ref="podChartRef" style="height: 250px;"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16">
      <el-col :span="16">
        <el-card shadow="never">
          <template #header>
            <h4 style="margin: 0;"><el-icon><Bell /></el-icon> 最近事件</h4>
          </template>
          <el-table :data="recentEvents" stripe size="small" empty-text="暂无事件">
            <el-table-column label="类型" width="60">
              <template #default="{ row }">
                <el-tag :type="eventType(row.type)" size="small">{{ row.type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="reason" label="原因" width="120" />
            <el-table-column prop="message" label="消息" min-width="250" show-overflow-tooltip />
            <el-table-column label="时间" width="120">
              <template #default="{ row }">{{ formatTime(row.lastTimestamp) }}</template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card shadow="never">
          <template #header>
            <h4 style="margin: 0;">资源使用</h4>
          </template>
          <div class="resource-usage">
            <div class="usage-item">
              <div class="usage-header">
                <span>CPU</span>
                <span>{{ resourceUsage.cpu.used }} / {{ resourceUsage.cpu.total }} 核</span>
              </div>
              <el-progress
                :percentage="resourceUsage.cpu.total > 0 ? Math.round((resourceUsage.cpu.used / resourceUsage.cpu.total) * 100) : 0"
                :color="resourceUsage.cpu.total > 0 && (resourceUsage.cpu.used / resourceUsage.cpu.total) >= 0.8 ? '#F56C6C' : '#409EFF'"
              />
            </div>
            <div class="usage-item">
              <div class="usage-header">
                <span>内存</span>
                <span>{{ resourceUsage.memory.used }} / {{ resourceUsage.memory.total }} Gi</span>
              </div>
              <el-progress
                :percentage="resourceUsage.memory.total > 0 ? Math.round((resourceUsage.memory.used / resourceUsage.memory.total) * 100) : 0"
                :color="resourceUsage.memory.total > 0 && (resourceUsage.memory.used / resourceUsage.memory.total) >= 0.8 ? '#F56C6C' : '#67C23A'"
              />
            </div>
            <div class="usage-item">
              <div class="usage-header">
                <span>存储</span>
                <span>{{ resourceUsage.storage.used }} / {{ resourceUsage.storage.total }} Gi</span>
              </div>
              <el-progress
                :percentage="resourceUsage.storage.total > 0 ? Math.round((resourceUsage.storage.used / resourceUsage.storage.total) * 100) : 0"
                :color="resourceUsage.storage.total > 0 && (resourceUsage.storage.used / resourceUsage.storage.total) >= 0.8 ? '#F56C6C' : '#E6A23C'"
              />
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; justify-content: space-between; align-items: center; }
.stat-card { display: flex; align-items: center; padding: 16px; }
.stat-icon { width: 48px; height: 48px; border-radius: 8px; display: flex; align-items: center; justify-content: center; color: white; font-size: 24px; margin-right: 16px; }
.stat-info { flex: 1; }
.stat-value { font-size: 24px; font-weight: bold; color: #303133; }
.stat-label { font-size: 14px; color: #909399; }
.resource-usage { display: flex; flex-direction: column; gap: 20px; }
.usage-item { }
.usage-header { display: flex; justify-content: space-between; margin-bottom: 8px; }
</style>
