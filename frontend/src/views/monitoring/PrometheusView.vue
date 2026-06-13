<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Refresh, Setting } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import request from '@/api/request'

const { t } = useI18n()
const loading = ref(false)
const prometheusUrl = ref(localStorage.getItem('gkube_prometheus_url') || 'http://prometheus:9090')
const showSettings = ref(false)
const nodeMetrics = ref<any[]>([])
const podMetrics = ref<any[]>([])
const activeTab = ref('nodes')
const chartRef = ref<HTMLElement>()
const chartInstance = ref<echarts.ECharts | null>(null)

// Prometheus query functions via backend proxy
async function queryPrometheus(query: string): Promise<any[]> {
  try {
    const res = await request.get('/k8s/prometheus/query', { params: { query } })
    if (res.data?.status === 'success' && res.data?.data?.result) {
      return res.data.data.result
    }
    return []
  } catch (e) {
    console.error('Prometheus query failed:', e)
    return []
  }
}

async function queryPrometheusRange(query: string, start: number, end: number, step: string = '60s'): Promise<any[]> {
  try {
    const res = await request.get('/k8s/prometheus/query_range', { params: { query, start, end, step } })
    if (res.data?.status === 'success' && res.data?.data?.result) {
      return res.data.data.result
    }
    return []
  } catch (e) {
    console.error('Prometheus range query failed:', e)
    return []
  }
}

// Fetch node CPU and memory metrics
async function fetchNodeMetrics() {
  loading.value = true
  try {
    // Get node CPU usage
    const cpuResults = await queryPrometheus('instance:node_cpu_utilization:rate5m')
    // Get node memory usage
    const memResults = await queryPrometheus('node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes')
    const memTotalResults = await queryPrometheus('node_memory_MemTotal_bytes')

    const nodeMap = new Map<string, any>()

    // Process CPU results
    cpuResults.forEach((result: any) => {
      const instance = result.metric?.instance || result.metric?.node || 'unknown'
      const node = nodeMap.get(instance) || { name: instance, cpuUsage: 0, memUsage: 0, memTotal: 0 }
      node.cpuUsage = parseFloat(result.value?.[1] || '0') * 100
      nodeMap.set(instance, node)
    })

    // Process memory results
    memResults.forEach((result: any) => {
      const instance = result.metric?.instance || result.metric?.node || 'unknown'
      const node = nodeMap.get(instance) || { name: instance, cpuUsage: 0, memUsage: 0, memTotal: 0 }
      node.memUsage = parseFloat(result.value?.[1] || '0')
      nodeMap.set(instance, node)
    })

    memTotalResults.forEach((result: any) => {
      const instance = result.metric?.instance || result.metric?.node || 'unknown'
      const node = nodeMap.get(instance) || { name: instance, cpuUsage: 0, memUsage: 0, memTotal: 0 }
      node.memTotal = parseFloat(result.value?.[1] || '0')
      nodeMap.set(instance, node)
    })

    nodeMetrics.value = Array.from(nodeMap.values())
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to fetch node metrics')
  } finally {
    loading.value = false
  }
}

// Fetch pod CPU and memory metrics
async function fetchPodMetrics() {
  loading.value = true
  try {
    const cpuResults = await queryPrometheus('sum(rate(container_cpu_usage_seconds_total{container!="",pod!=""}[5m])) by (pod, namespace)')
    const memResults = await queryPrometheus('sum(container_memory_working_set_bytes{container="",pod!=""}) by (pod, namespace)')

    const podMap = new Map<string, any>()

    cpuResults.forEach((result: any) => {
      const key = `${result.metric?.namespace}/${result.metric?.pod}`
      const pod = podMap.get(key) || { name: result.metric?.pod, namespace: result.metric?.namespace, cpuUsage: 0, memUsage: 0 }
      pod.cpuUsage = parseFloat(result.value?.[1] || '0')
      podMap.set(key, pod)
    })

    memResults.forEach((result: any) => {
      const key = `${result.metric?.namespace}/${result.metric?.pod}`
      const pod = podMap.get(key) || { name: result.metric?.pod, namespace: result.metric?.namespace, cpuUsage: 0, memUsage: 0 }
      pod.memUsage = parseFloat(result.value?.[1] || '0')
      podMap.set(key, pod)
    })

    podMetrics.value = Array.from(podMap.values())
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to fetch pod metrics')
  } finally {
    loading.value = false
  }
}

// Fetch time-series data for charts
async function fetchTimeSeriesData() {
  const end = Math.floor(Date.now() / 1000)
  const start = end - 3600 // Last 1 hour

  const cpuData = await queryPrometheusRange('instance:node_cpu_utilization:rate5m', start, end, '60s')
  const memData = await queryPrometheusRange('(node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / node_memory_MemTotal_bytes', start, end, '60s')

  if (cpuData.length > 0 && chartRef.value) {
    if (!chartInstance.value) {
      chartInstance.value = echarts.init(chartRef.value)
    }

    const timestamps = cpuData[0]?.values?.map((v: any) => {
      const date = new Date(v[0] * 1000)
      return date.toLocaleTimeString()
    }) || []

    const series: any[] = []
    cpuData.forEach((result: any, index: number) => {
      const name = result.metric?.instance || result.metric?.node || `Node ${index + 1}`
      series.push({
        name: `${name} CPU`,
        type: 'line',
        smooth: true,
        data: result.values?.map((v: any) => parseFloat(v[1]) * 100) || [],
        showSymbol: false,
      })
    })

    memData.forEach((result: any, index: number) => {
      const name = result.metric?.instance || result.metric?.node || `Node ${index + 1}`
      series.push({
        name: `${name} Memory`,
        type: 'line',
        smooth: true,
        data: result.values?.map((v: any) => parseFloat(v[1]) * 100) || [],
        showSymbol: false,
        lineStyle: { type: 'dashed' },
      })
    })

    chartInstance.value.setOption({
      title: { text: 'Node Resource Usage (Last 1 Hour)', left: 'center' },
      tooltip: {
        trigger: 'axis',
        formatter: (params: any) => {
          let html = `<b>${params[0]?.axisValue}</b><br/>`
          params.forEach((p: any) => {
            html += `${p.seriesName}: ${p.value.toFixed(2)}%<br/>`
          })
          return html
        },
      },
      legend: { bottom: 0, type: 'scroll' },
      grid: { left: '3%', right: '4%', bottom: '15%', containLabel: true },
      xAxis: { type: 'category', data: timestamps, axisLabel: { rotate: 45 } },
      yAxis: { type: 'value', name: 'Usage (%)', max: 100 },
      series,
    })
  }
}

function savePrometheusUrl() {
  localStorage.setItem('gkube_prometheus_url', prometheusUrl.value)
  showSettings.value = false
  ElMessage.success('Prometheus URL saved')
  fetchAll()
}

async function fetchAll() {
  await Promise.all([fetchNodeMetrics(), fetchPodMetrics(), fetchTimeSeriesData()])
}

function handleTabChange(tab: string) {
  if (tab === 'nodes' && nodeMetrics.value.length === 0) fetchNodeMetrics()
  if (tab === 'pods' && podMetrics.value.length === 0) fetchPodMetrics()
}

function formatCpu(cores: number): string {
  if (cores >= 1) return cores.toFixed(2) + ' Core'
  return (cores * 1000).toFixed(0) + 'm'
}

function formatMemory(bytes: number): string {
  if (bytes >= 1073741824) return (bytes / 1073741824).toFixed(2) + ' Gi'
  if (bytes >= 1048576) return (bytes / 1048576).toFixed(0) + ' Mi'
  return bytes.toFixed(0) + ' bytes'
}

onMounted(fetchAll)
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <h3 style="margin: 0;">Prometheus Monitoring</h3>
        <div style="display: flex; gap: 8px;">
          <el-button @click="showSettings = true"><el-icon><Setting /></el-icon> Settings</el-button>
          <el-button type="primary" @click="fetchAll"><el-icon><Refresh /></el-icon> Refresh</el-button>
        </div>
      </div>
    </el-card>

    <!-- Chart Section -->
    <el-card shadow="never" style="margin-bottom: 16px;">
      <div ref="chartRef" style="width: 100%; height: 400px;"></div>
      <el-empty v-if="!loading && nodeMetrics.length === 0" description="No Prometheus data available. Configure Prometheus URL in settings." />
    </el-card>

    <!-- Metrics Tables -->
    <el-card shadow="never">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="Node Metrics" name="nodes">
          <el-table :data="nodeMetrics" v-loading="loading" stripe>
            <el-table-column prop="name" label="Node" min-width="200" show-overflow-tooltip />
            <el-table-column label="CPU Usage" width="150">
              <template #default="{ row }">
                <el-tag :type="row.cpuUsage > 80 ? 'danger' : row.cpuUsage > 60 ? 'warning' : 'success'" size="small">
                  {{ row.cpuUsage.toFixed(1) }}%
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="Memory Usage" width="200">
              <template #default="{ row }">
                <div>{{ formatMemory(row.memUsage) }} / {{ formatMemory(row.memTotal) }}</div>
                <el-progress :percentage="row.memTotal > 0 ? Math.round(row.memUsage / row.memTotal * 100) : 0" :stroke-width="6" :show-text="false" style="margin-top: 4px;" />
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="Pod Metrics" name="pods">
          <el-table :data="podMetrics" v-loading="loading" stripe>
            <el-table-column prop="name" label="Pod" min-width="250" show-overflow-tooltip />
            <el-table-column prop="namespace" label="Namespace" width="140" />
            <el-table-column label="CPU" width="150">
              <template #default="{ row }">{{ formatCpu(row.cpuUsage) }}</template>
            </el-table-column>
            <el-table-column label="Memory" width="150">
              <template #default="{ row }">{{ formatMemory(row.memUsage) }}</template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- Settings Dialog -->
    <el-dialog v-model="showSettings" title="Prometheus Settings" width="500px">
      <el-form label-width="120px">
        <el-form-item label="Prometheus URL">
          <el-input v-model="prometheusUrl" placeholder="http://prometheus:9090" />
        </el-form-item>
        <el-alert title="Enter the URL of your Prometheus server. The server must be accessible from the browser." type="info" :closable="false" show-icon />
      </el-form>
      <template #footer>
        <el-button @click="showSettings = false">Cancel</el-button>
        <el-button type="primary" @click="savePrometheusUrl">Save</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; justify-content: space-between; align-items: center; }
</style>
