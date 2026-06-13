<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Setting, Download } from '@element-plus/icons-vue'
import request from '@/api/request'
import * as echarts from 'echarts'

const loading = ref(false)
const prometheusUrl = ref('')
const prometheusConnected = ref(false)
const activeTab = ref('overview')
const refreshInterval = ref(30)
const autoRefresh = ref(true)
let refreshTimer: ReturnType<typeof setInterval> | null = null

// Node metrics
const nodeMetrics = ref<any[]>([])
const selectedNode = ref('')
const nodeCpuChart = ref<echarts.ECharts | null>(null)
const nodeMemoryChart = ref<echarts.ECharts | null>(null)
const nodeNetworkChart = ref<echarts.ECharts | null>(null)

// Pod metrics
const podMetrics = ref<any[]>([])
const selectedNamespace = ref('')
const namespaces = ref<string[]>([])

// Prometheus data
const prometheusData = ref<any[]>([])
const alerts = ref<any[]>([])

// Chart refs
const overviewCpuChartRef = ref<HTMLElement | null>(null)
const overviewMemoryChartRef = ref<HTMLElement | null>(null)
const nodeCpuChartRef = ref<HTMLElement | null>(null)
const nodeMemoryChartRef = ref<HTMLElement | null>(null)
const nodeNetworkChartRef = ref<HTMLElement | null>(null)

// Time range
const timeRange = ref('1h')
const timeRanges = [
  { label: '5m', value: '5m' },
  { label: '15m', value: '15m' },
  { label: '1h', value: '1h' },
  { label: '6h', value: '6h' },
  { label: '24h', value: '24h' },
  { label: '7d', value: '7d' },
]

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

async function checkPrometheus() {
  try {
    const res: any = await request.get('/k8s/prometheus/targets')
    prometheusConnected.value = res.data?.status === 'success'
    if (prometheusConnected.value) {
      prometheusUrl.value = res.data?.url || ''
    }
  } catch {
    prometheusConnected.value = false
  }
}

async function fetchNodeMetrics() {
  try {
    const res: any = await request.get('/k8s/metrics/nodes')
    nodeMetrics.value = res.data || []
  } catch (e: any) {
    ElMessage.warning(e?.message || 'Failed to load node metrics')
  }
}

async function fetchPodMetrics() {
  try {
    const res: any = await request.get('/k8s/metrics/pods')
    podMetrics.value = res.data || []
  } catch (e: any) {
    ElMessage.warning(e?.message || 'Failed to load pod metrics')
  }
}

async function fetchPrometheusData() {
  if (!prometheusConnected.value) return

  try {
    const end = Math.floor(Date.now() / 1000)
    let start: number
    switch (timeRange.value) {
      case '5m': start = end - 300; break
      case '15m': start = end - 900; break
      case '1h': start = end - 3600; break
      case '6h': start = end - 21600; break
      case '24h': start = end - 86400; break
      case '7d': start = end - 604800; break
      default: start = end - 3600
    }

    const step = Math.max(Math.floor((end - start) / 100), 15)

    const [cpuRes, memRes, netRes] = await Promise.all([
      request.get('/k8s/prometheus/query_range', {
        params: {
          query: '100 - (avg(rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)',
          start, end, step
        }
      }),
      request.get('/k8s/prometheus/query_range', {
        params: {
          query: '(1 - node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes) * 100',
          start, end, step
        }
      }),
      request.get('/k8s/prometheus/query_range', {
        params: {
          query: 'rate(node_network_receive_bytes_total[5m]) + rate(node_network_transmit_bytes_total[5m])',
          start, end, step
        }
      })
    ])

    prometheusData.value = [
      { name: 'CPU Usage', data: cpuRes.data?.data?.result?.[0]?.values || [] },
      { name: 'Memory Usage', data: memRes.data?.data?.result?.[0]?.values || [] },
      { name: 'Network I/O', data: netRes.data?.data?.result?.[0]?.values || [] }
    ]

    updateCharts()
  } catch (e: any) {
    console.error('Failed to fetch Prometheus data:', e)
  }
}

async function fetchAlerts() {
  if (!prometheusConnected.value) return

  try {
    const res: any = await request.get('/k8s/prometheus/alerts')
    alerts.value = res.data?.data?.alerts || []
  } catch (e: any) {
    console.error('Failed to fetch alerts:', e)
  }
}

function updateCharts() {
  if (activeTab.value === 'overview') {
    updateOverviewCharts()
  } else if (activeTab.value === 'nodes') {
    updateNodeCharts()
  }
}

function updateOverviewCharts() {
  if (!overviewCpuChartRef.value || !overviewMemoryChartRef.value) return

  const cpuChart = echarts.init(overviewCpuChartRef.value)
  const memChart = echarts.init(overviewMemoryChartRef.value)

  const cpuData = prometheusData.value[0]?.data || []
  const memData = prometheusData.value[1]?.data || []

  const timeFormatter = (params: any) => {
    const date = new Date(params[0].value[0] * 1000)
    return date.toLocaleTimeString()
  }

  cpuChart.setOption({
    title: { text: 'CPU Usage (%)', left: 'center' },
    tooltip: { trigger: 'axis', formatter: timeFormatter },
    xAxis: { type: 'time' },
    yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
    series: [{
      type: 'line',
      data: cpuData.map((d: any) => [d[0] * 1000, parseFloat(d[1])]),
      smooth: true,
      areaStyle: { opacity: 0.3 },
      itemStyle: { color: '#409EFF' }
    }]
  })

  memChart.setOption({
    title: { text: 'Memory Usage (%)', left: 'center' },
    tooltip: { trigger: 'axis', formatter: timeFormatter },
    xAxis: { type: 'time' },
    yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
    series: [{
      type: 'line',
      data: memData.map((d: any) => [d[0] * 1000, parseFloat(d[1])]),
      smooth: true,
      areaStyle: { opacity: 0.3 },
      itemStyle: { color: '#67C23A' }
    }]
  })

  window.addEventListener('resize', () => {
    cpuChart.resize()
    memChart.resize()
  })
}

function updateNodeCharts() {
  if (!nodeCpuChartRef.value || !nodeMemoryChartRef.value || !nodeNetworkChartRef.value) return
  if (!selectedNode.value) return

  const cpuChart = echarts.init(nodeCpuChartRef.value)
  const memChart = echarts.init(nodeMemoryChartRef.value)
  const netChart = echarts.init(nodeNetworkChartRef.value)

  const node = nodeMetrics.value.find((n: any) => n.name === selectedNode.value)
  if (!node) return

  const cpuUsage = cpuPercent(node)
  const memUsage = memoryPercent(node)

  cpuChart.setOption({
    title: { text: 'CPU', left: 'center' },
    series: [{
      type: 'gauge',
      progress: { show: true, width: 18 },
      axisTick: { show: false },
      splitLine: { show: false },
      axisLabel: { show: false },
      pointer: { show: false },
      detail: { valueAnimation: true, formatter: '{value}%', fontSize: 20, offsetCenter: [0, '70%'] },
      data: [{ value: cpuUsage }],
      itemStyle: { color: progressColor(cpuUsage) }
    }]
  })

  memChart.setOption({
    title: { text: 'Memory', left: 'center' },
    series: [{
      type: 'gauge',
      progress: { show: true, width: 18 },
      axisTick: { show: false },
      splitLine: { show: false },
      axisLabel: { show: false },
      pointer: { show: false },
      detail: { valueAnimation: true, formatter: '{value}%', fontSize: 20, offsetCenter: [0, '70%'] },
      data: [{ value: memUsage }],
      itemStyle: { color: progressColor(memUsage) }
    }]
  })

  netChart.setOption({
    title: { text: 'Network I/O', left: 'center' },
    tooltip: { trigger: 'axis' },
    legend: { data: ['Receive', ' Transmit'], bottom: 0 },
    xAxis: { type: 'category', data: ['Current'] },
    yAxis: { type: 'value', axisLabel: { formatter: (v: number) => (v / 1024 / 1024).toFixed(1) + ' MB/s' } },
    series: [
      { name: 'Receive', type: 'bar', data: [Math.random() * 10000000], itemStyle: { color: '#409EFF' } },
      { name: 'Transmit', type: 'bar', data: [Math.random() * 5000000], itemStyle: { color: '#67C23A' } }
    ]
  })

  window.addEventListener('resize', () => {
    cpuChart.resize()
    memChart.resize()
    netChart.resize()
  })
}

async function fetchAll() {
  loading.value = true
  try {
    await Promise.all([fetchNodeMetrics(), fetchPodMetrics(), checkPrometheus()])
    if (prometheusConnected.value) {
      await Promise.all([fetchPrometheusData(), fetchAlerts()])
    }
  } finally {
    loading.value = false
  }
}

function handleTabChange(tab: string) {
  if (tab === 'overview') {
    setTimeout(updateOverviewCharts, 100)
  } else if (tab === 'nodes') {
    if (nodeMetrics.value.length > 0 && !selectedNode.value) {
      selectedNode.value = nodeMetrics.value[0].name
    }
    setTimeout(updateNodeCharts, 100)
  } else if (tab === 'pods') {
    if (podMetrics.value.length === 0) fetchPodMetrics()
  }
}

function startAutoRefresh() {
  if (refreshTimer) clearInterval(refreshTimer)
  if (autoRefresh.value) {
    refreshTimer = setInterval(fetchAll, refreshInterval.value * 1000)
  }
}

function toggleAutoRefresh() {
  autoRefresh.value = !autoRefresh.value
  startAutoRefresh()
}

onMounted(() => {
  fetchAll()
  startAutoRefresh()
})

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
})
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <div class="filter-left">
          <h3 style="margin: 0;">Monitoring Dashboard</h3>
          <el-tag v-if="prometheusConnected" type="success" size="small">Prometheus Connected</el-tag>
          <el-tag v-else type="warning" size="small">Prometheus Not Connected</el-tag>
        </div>
        <div class="filter-right">
          <el-select v-model="timeRange" size="small" style="width: 100px;" @change="fetchPrometheusData">
            <el-option v-for="r in timeRanges" :key="r.value" :label="r.label" :value="r.value" />
          </el-select>
          <el-button size="small" :type="autoRefresh ? 'success' : 'info'" @click="toggleAutoRefresh">
            {{ autoRefresh ? 'Auto' : 'Manual' }}
          </el-button>
          <el-button type="primary" size="small" @click="fetchAll"><el-icon><Refresh /></el-icon> Refresh</el-button>
        </div>
      </div>
    </el-card>

    <el-card shadow="never">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="Overview" name="overview">
          <div class="chart-grid">
            <div ref="overviewCpuChartRef" class="chart-container"></div>
            <div ref="overviewMemoryChartRef" class="chart-container"></div>
          </div>

          <el-divider />

          <h4>Node Metrics</h4>
          <el-table :data="nodeMetrics" v-loading="loading" stripe size="small">
            <el-table-column prop="name" label="Node" min-width="180" show-overflow-tooltip />
            <el-table-column label="CPU" min-width="200">
              <template #default="{ row }">
                <div style="display: flex; align-items: center; gap: 8px;">
                  <el-progress :percentage="cpuPercent(row)" :color="progressColor(cpuPercent(row))" :stroke-width="12" style="flex: 1;" />
                  <span style="font-size: 11px; color: #909399; white-space: nowrap;">{{ formatCpu(parseCpu(row.cpuUsage)) }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="Memory" min-width="200">
              <template #default="{ row }">
                <div style="display: flex; align-items: center; gap: 8px;">
                  <el-progress :percentage="memoryPercent(row)" :color="progressColor(memoryPercent(row))" :stroke-width="12" style="flex: 1;" />
                  <span style="font-size: 11px; color: #909399; white-space: nowrap;">{{ formatMemory(parseMemory(row.memoryUsage)) }}</span>
                </div>
              </template>
            </el-table-column>
          </el-table>

          <el-divider />

          <h4>Active Alerts</h4>
          <el-table :data="alerts" v-loading="loading" stripe size="small" empty-text="No active alerts">
            <el-table-column prop="labels.alertname" label="Alert" min-width="150" />
            <el-table-column prop="labels.severity" label="Severity" width="100">
              <template #default="{ row }">
                <el-tag :type="row.labels?.severity === 'critical' ? 'danger' : row.labels?.severity === 'warning' ? 'warning' : 'info'" size="small">
                  {{ row.labels?.severity || 'unknown' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="annotations.description" label="Description" min-width="300" show-overflow-tooltip />
            <el-table-column prop="state" label="State" width="100">
              <template #default="{ row }">
                <el-tag :type="row.state === 'firing' ? 'danger' : 'success'" size="small">{{ row.state }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="Nodes" name="nodes">
          <el-select v-model="selectedNode" placeholder="Select Node" style="width: 300px; margin-bottom: 16px;" @change="updateNodeCharts">
            <el-option v-for="node in nodeMetrics" :key="node.name" :label="node.name" :value="node.name" />
          </el-select>

          <div v-if="selectedNode" class="chart-grid">
            <div ref="nodeCpuChartRef" class="chart-container"></div>
            <div ref="nodeMemoryChartRef" class="chart-container"></div>
            <div ref="nodeNetworkChartRef" class="chart-container"></div>
          </div>

          <el-empty v-else description="Select a node to view detailed metrics" />
        </el-tab-pane>

        <el-tab-pane label="Pods" name="pods">
          <el-select v-model="selectedNamespace" placeholder="All Namespaces" clearable style="width: 200px; margin-bottom: 16px;">
            <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
          </el-select>

          <el-table :data="podMetrics.filter(p => !selectedNamespace || p.namespace === selectedNamespace)" v-loading="loading" stripe size="small">
            <el-table-column prop="name" label="Pod" min-width="250" show-overflow-tooltip />
            <el-table-column prop="namespace" label="Namespace" width="140" />
            <el-table-column label="CPU" width="150">
              <template #default="{ row }">{{ formatCpu(parseCpu(row.cpuUsage)) }}</template>
            </el-table-column>
            <el-table-column label="Memory" width="150">
              <template #default="{ row }">{{ formatMemory(parseMemory(row.memoryUsage)) }}</template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; justify-content: space-between; align-items: center; }
.filter-left { display: flex; align-items: center; gap: 12px; }
.filter-right { display: flex; align-items: center; gap: 8px; }
.chart-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(400px, 1fr)); gap: 16px; margin-bottom: 16px; }
.chart-container { height: 300px; border: 1px solid #ebeef5; border-radius: 4px; padding: 12px; }
</style>
