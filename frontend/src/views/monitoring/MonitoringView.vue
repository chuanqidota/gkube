<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import request from '@/api/request'

const { t } = useI18n()
const loading = ref(false)
const nodeMetrics = ref<any[]>([])
const podMetrics = ref<any[]>([])
const activeTab = ref('nodes')

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
    ElMessage.warning(e?.message || 'Failed to load node metrics (metrics-server may not be installed)')
  }
}

async function fetchPodMetrics() {
  try {
    const res: any = await request.get('/k8s/metrics/pods')
    podMetrics.value = res.data || []
  } catch (e: any) {
    ElMessage.warning(e?.message || 'Failed to load pod metrics (metrics-server may not be installed)')
  }
}

async function fetchAll() {
  loading.value = true
  try {
    await Promise.all([fetchNodeMetrics(), fetchPodMetrics()])
  } finally {
    loading.value = false
  }
}

function handleTabChange(tab: string) {
  if (tab === 'nodes' && nodeMetrics.value.length === 0) fetchNodeMetrics()
  if (tab === 'pods' && podMetrics.value.length === 0) fetchPodMetrics()
}

onMounted(fetchAll)
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <h3 style="margin: 0;">Monitoring</h3>
        <el-button type="primary" @click="fetchAll"><el-icon><Refresh /></el-icon> Refresh</el-button>
      </div>
    </el-card>

    <el-card shadow="never">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="Node Metrics" name="nodes">
          <el-table :data="nodeMetrics" v-loading="loading" stripe>
            <el-table-column prop="name" label="Node" min-width="180" show-overflow-tooltip />
            <el-table-column label="CPU" min-width="250">
              <template #default="{ row }">
                <div style="display: flex; align-items: center; gap: 12px;">
                  <el-progress :percentage="cpuPercent(row)" :color="progressColor(cpuPercent(row))" :stroke-width="16" style="flex: 1;" />
                  <span style="font-size: 12px; color: #909399; white-space: nowrap;">{{ formatCpu(parseCpu(row.cpuUsage)) }} / {{ formatCpu(parseCpu(row.cpu)) }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="Memory" min-width="250">
              <template #default="{ row }">
                <div style="display: flex; align-items: center; gap: 12px;">
                  <el-progress :percentage="memoryPercent(row)" :color="progressColor(memoryPercent(row))" :stroke-width="16" style="flex: 1;" />
                  <span style="font-size: 12px; color: #909399; white-space: nowrap;">{{ formatMemory(parseMemory(row.memoryUsage)) }} / {{ formatMemory(parseMemory(row.memory)) }}</span>
                </div>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="!loading && nodeMetrics.length === 0" description="No node metrics available. metrics-server may not be installed." />
        </el-tab-pane>

        <el-tab-pane label="Pod Metrics" name="pods">
          <el-table :data="podMetrics" v-loading="loading" stripe>
            <el-table-column prop="name" label="Pod" min-width="250" show-overflow-tooltip />
            <el-table-column prop="namespace" label="Namespace" width="140" />
            <el-table-column label="CPU" width="150">
              <template #default="{ row }">{{ formatCpu(parseCpu(row.cpuUsage)) }}</template>
            </el-table-column>
            <el-table-column label="Memory" width="150">
              <template #default="{ row }">{{ formatMemory(parseMemory(row.memoryUsage)) }}</template>
            </el-table-column>
          </el-table>
          <el-empty v-if="!loading && podMetrics.length === 0" description="No pod metrics available. metrics-server may not be installed." />
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; justify-content: space-between; align-items: center; }
</style>
