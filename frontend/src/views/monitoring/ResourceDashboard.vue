<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Monitor, Cpu, Coin, FolderOpened } from '@element-plus/icons-vue'
import request from '@/api/request'
import * as echarts from 'echarts'

const loading = ref(false)
const nodeMetrics = ref<any[]>([])
const podMetrics = ref<any[]>([])
const namespaces = ref<string[]>([])
const selectedNamespace = ref('')

// Chart refs
const cpuChartRef = ref<HTMLElement | null>(null)
const memoryChartRef = ref<HTMLElement | null>(null)
const storageChartRef = ref<HTMLElement | null>(null)

// Dashboard stats
const stats = ref({
  totalNodes: 0,
  totalPods: 0,
  totalNamespaces: 0,
  avgCpuUsage: 0,
  avgMemoryUsage: 0,
})

function parseCpu(val: string): number {
  if (!val || val === 'N/A') return 0
  val = String(val)
  if (val.endsWith('n')) return parseInt(val) / 1000000
  if (val.endsWith('u')) return parseInt(val) / 1000
  if (val.endsWith('m')) return parseInt(val)
  return parseFloat(val) * 1000
}

function parseMemory(val: string): number {
  if (!val || val === 'N/A') return 0
  val = String(val)
  if (val.endsWith('Ki')) return parseInt(val) / 1048576
  if (val.endsWith('Mi')) return parseInt(val) / 1024
  if (val.endsWith('Gi')) return parseInt(val)
  if (val.endsWith('Ti')) return parseInt(val) * 1024
  return parseFloat(val) / (1024 * 1024 * 1024)
}

function formatCpu(millicores: number): string {
  if (millicores >= 1000) return (millicores / 1000).toFixed(2) + ' Core'
  return Math.round(millicores) + 'm'
}

function formatMemory(gi: number): string {
  if (gi >= 1) return gi.toFixed(2) + ' Gi'
  return (gi * 1024).toFixed(0) + ' Mi'
}

function cpuPercent(node: any): number {
  const cap = parseCpu(node.cpu)
  const usage = parseCpu(node.cpuUsage)
  if (cap === 0) return 0
  return Math.round((usage / cap) * 100)
}

function memoryPercent(node: any): number {
  const cap = parseMemory(node.memory)
  const usage = parseMemory(node.memoryUsage)
  if (cap === 0) return 0
  return Math.round((usage / cap) * 100)
}

function progressColor(percent: number) {
  if (percent >= 90) return '#F56C6C'
  if (percent >= 70) return '#E6A23C'
  return '#409EFF'
}

async function fetchNodeMetrics() {
  try {
    const res: any = await request.get('/k8s/metrics/nodes')
    nodeMetrics.value = res.data || []
  } catch (e: any) {
    ElMessage.warning('Failed to load node metrics')
  }
}

async function fetchPodMetrics() {
  try {
    const res: any = await request.get('/k8s/metrics/pods')
    podMetrics.value = res.data || []
  } catch (e: any) {
    ElMessage.warning('Failed to load pod metrics')
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

function updateStats() {
  stats.value.totalNodes = nodeMetrics.value.length
  stats.value.totalPods = podMetrics.value.length
  stats.value.totalNamespaces = namespaces.value.length

  if (nodeMetrics.value.length > 0) {
    const totalCpu = nodeMetrics.value.reduce((sum, node) => sum + cpuPercent(node), 0)
    const totalMem = nodeMetrics.value.reduce((sum, node) => sum + memoryPercent(node), 0)
    stats.value.avgCpuUsage = Math.round(totalCpu / nodeMetrics.value.length)
    stats.value.avgMemoryUsage = Math.round(totalMem / nodeMetrics.value.length)
  }
}

function updateCharts() {
  if (!cpuChartRef.value || !memoryChartRef.value || !storageChartRef.value) return

  const cpuChart = echarts.init(cpuChartRef.value)
  const memChart = echarts.init(memoryChartRef.value)
  const storageChart = echarts.init(storageChartRef.value)

  // CPU Chart - Node usage
  cpuChart.setOption({
    title: { text: 'CPU 使用率', left: 'center', textStyle: { fontSize: 14 } },
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    xAxis: {
      type: 'category',
      data: nodeMetrics.value.map(n => n.name),
      axisLabel: { rotate: 45, fontSize: 10 }
    },
    yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
    series: [{
      type: 'bar',
      data: nodeMetrics.value.map(n => ({
        value: cpuPercent(n),
        itemStyle: { color: progressColor(cpuPercent(n)) }
      })),
      label: { show: true, position: 'top', formatter: '{c}%' }
    }]
  })

  // Memory Chart - Node usage
  memChart.setOption({
    title: { text: '内存使用率', left: 'center', textStyle: { fontSize: 14 } },
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    xAxis: {
      type: 'category',
      data: nodeMetrics.value.map(n => n.name),
      axisLabel: { rotate: 45, fontSize: 10 }
    },
    yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
    series: [{
      type: 'bar',
      data: nodeMetrics.value.map(n => ({
        value: memoryPercent(n),
        itemStyle: { color: progressColor(memoryPercent(n)) }
      })),
      label: { show: true, position: 'top', formatter: '{c}%' }
    }]
  })

  // Storage Chart - Pod count by namespace
  const podsByNs: Record<string, number> = {}
  podMetrics.value.forEach(p => {
    podsByNs[p.namespace] = (podsByNs[p.namespace] || 0) + 1
  })
  const nsNames = Object.keys(podsByNs)
  const nsCounts = Object.values(podsByNs)

  storageChart.setOption({
    title: { text: 'Pod 分布（按命名空间）', left: 'center', textStyle: { fontSize: 14 } },
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    series: [{
      type: 'pie',
      radius: '50%',
      data: nsNames.map((name, i) => ({
        name: name,
        value: nsCounts[i]
      })),
      emphasis: {
        itemStyle: {
          shadowBlur: 10,
          shadowOffsetX: 0,
          shadowColor: 'rgba(0, 0, 0, 0.5)'
        }
      }
    }]
  })

  window.addEventListener('resize', () => {
    cpuChart.resize()
    memChart.resize()
    storageChart.resize()
  })
}

const filteredPods = computed(() => {
  if (!selectedNamespace.value) return podMetrics.value
  return podMetrics.value.filter(p => p.namespace === selectedNamespace.value)
})

async function fetchAll() {
  loading.value = true
  try {
    await Promise.all([fetchNodeMetrics(), fetchPodMetrics(), fetchNamespaces()])
    updateStats()
    setTimeout(updateCharts, 100)
  } finally {
    loading.value = false
  }
}

onMounted(fetchAll)
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <h3 style="margin: 0;"><el-icon><Monitor /></el-icon> 资源监控</h3>
        <div class="filter-right">
          <el-button type="primary" @click="fetchAll"><el-icon><Refresh /></el-icon> 刷新</el-button>
        </div>
      </div>
    </el-card>

    <!-- Stats Cards -->
    <el-row :gutter="16" style="margin-bottom: 16px;">
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-icon" style="background: #409EFF;"><el-icon><Cpu /></el-icon></div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.totalNodes }}</div>
            <div class="stat-label">节点数量</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-icon" style="background: #67C23A;"><el-icon><Coin /></el-icon></div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.totalPods }}</div>
            <div class="stat-label">Pod 数量</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-icon" style="background: #E6A23C;"><el-icon><FolderOpened /></el-icon></div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.totalNamespaces }}</div>
            <div class="stat-label">命名空间</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never" class="stat-card">
          <div class="stat-icon" style="background: #F56C6C;"><el-icon><Monitor /></el-icon></div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.avgCpuUsage }}%</div>
            <div class="stat-label">平均 CPU 使用率</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Charts -->
    <el-row :gutter="16" style="margin-bottom: 16px;">
      <el-col :span="8">
        <el-card shadow="never">
          <div ref="cpuChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never">
          <div ref="memoryChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never">
          <div ref="storageChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Node Metrics Table -->
    <el-card shadow="never" style="margin-bottom: 16px;">
      <template #header>
        <h4 style="margin: 0;">节点指标</h4>
      </template>
      <el-table :data="nodeMetrics" v-loading="loading" stripe size="small">
        <el-table-column prop="name" label="节点" min-width="180" show-overflow-tooltip />
        <el-table-column label="CPU" min-width="200">
          <template #default="{ row }">
            <div style="display: flex; align-items: center; gap: 8px;">
              <el-progress :percentage="cpuPercent(row)" :color="progressColor(cpuPercent(row))" :stroke-width="12" style="flex: 1;" />
              <span style="font-size: 11px; color: #909399; white-space: nowrap;">{{ formatCpu(parseCpu(row.cpuUsage)) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="内存" min-width="200">
          <template #default="{ row }">
            <div style="display: flex; align-items: center; gap: 8px;">
              <el-progress :percentage="memoryPercent(row)" :color="progressColor(memoryPercent(row))" :stroke-width="12" style="flex: 1;" />
              <span style="font-size: 11px; color: #909399; white-space: nowrap;">{{ formatMemory(parseMemory(row.memoryUsage)) }}</span>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Pod Metrics Table -->
    <el-card shadow="never">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <h4 style="margin: 0;">Pod 指标</h4>
          <el-select v-model="selectedNamespace" placeholder="所有命名空间" clearable style="width: 200px;" size="small">
            <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
          </el-select>
        </div>
      </template>
      <el-table :data="filteredPods" v-loading="loading" stripe size="small">
        <el-table-column prop="name" label="Pod" min-width="250" show-overflow-tooltip />
        <el-table-column prop="namespace" label="命名空间" width="140" />
        <el-table-column label="CPU" width="150">
          <template #default="{ row }">{{ formatCpu(parseCpu(row.cpuUsage)) }}</template>
        </el-table-column>
        <el-table-column label="内存" width="150">
          <template #default="{ row }">{{ formatMemory(parseMemory(row.memoryUsage)) }}</template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; justify-content: space-between; align-items: center; }
.filter-right { display: flex; align-items: center; gap: 8px; }
.stat-card { display: flex; align-items: center; padding: 16px; }
.stat-icon { width: 48px; height: 48px; border-radius: 8px; display: flex; align-items: center; justify-content: center; color: white; font-size: 24px; margin-right: 16px; }
.stat-info { flex: 1; }
.stat-value { font-size: 24px; font-weight: bold; color: #303133; }
.stat-label { font-size: 14px; color: #909399; }
.chart-container { height: 300px; }
</style>
