<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Delete, Edit } from '@element-plus/icons-vue'
import request from '@/api/request'
import * as echarts from 'echarts'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const quota = ref<any>(null)
const chartRef = ref<HTMLElement | null>(null)

async function fetchQuota() {
  const namespace = route.params.namespace as string
  const name = route.params.name as string

  if (!namespace || !name) return

  loading.value = true
  try {
    const res: any = await request.get('/k8s/resourcequota/detail', {
      params: { namespace, name }
    })
    quota.value = res.data
    setTimeout(updateChart, 100)
  } catch (e: any) {
    ElMessage.error('Failed to load resource quota')
  } finally {
    loading.value = false
  }
}

function updateChart() {
  if (!chartRef.value || !quota.value) return

  const chart = echarts.init(chartRef.value)
  const hard = quota.value.spec?.hard || {}
  const used = quota.value.status?.used || {}

  const categories = Object.keys(hard)
  const hardValues = categories.map(k => parseResourceValue(hard[k]))
  const usedValues = categories.map(k => parseResourceValue(used[k] || '0'))

  chart.setOption({
    title: { text: '资源配额使用情况', left: 'center' },
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    legend: { data: ['配额', '已使用'], bottom: 0 },
    xAxis: {
      type: 'category',
      data: categories.map(formatCategory),
      axisLabel: { rotate: 45, fontSize: 10 }
    },
    yAxis: { type: 'value' },
    series: [
      {
        name: '配额',
        type: 'bar',
        data: hardValues,
        itemStyle: { color: '#409EFF' }
      },
      {
        name: '已使用',
        type: 'bar',
        data: usedValues,
        itemStyle: { color: '#67C23A' }
      }
    ]
  })

  window.addEventListener('resize', () => chart.resize())
}

function parseResourceValue(val: string): number {
  if (!val) return 0
  val = String(val)
  if (val.endsWith('Gi')) return parseFloat(val) * 1024
  if (val.endsWith('Mi')) return parseFloat(val)
  if (val.endsWith('Ki')) return parseFloat(val) / 1024
  if (val.endsWith('m')) return parseFloat(val) / 1000
  return parseFloat(val)
}

function formatCategory(cat: string): string {
  const map: Record<string, string> = {
    'requests.cpu': 'CPU 请求',
    'requests.memory': '内存请求',
    'limits.cpu': 'CPU 限制',
    'limits.memory': '内存限制',
    'pods': 'Pod 数量',
    'services': '服务数量',
    'secrets': 'Secret 数量',
    'configmaps': 'ConfigMap 数量',
    'persistentvolumeclaims': 'PVC 数量',
    'resourcequotas': '配额数量',
    'replicationcontrollers': 'RC 数量',
  }
  return map[cat] || cat
}

function usagePercent(hard: string, used: string): number {
  const h = parseResourceValue(hard)
  const u = parseResourceValue(used)
  if (h === 0) return 0
  return Math.round((u / h) * 100)
}

function progressColor(percent: number) {
  if (percent >= 90) return '#F56C6C'
  if (percent >= 70) return '#E6A23C'
  return '#409EFF'
}

async function deleteQuota() {
  try {
    await ElMessageBox.confirm('Delete this resource quota?', 'Confirm')
    await request.delete('/k8s/resourcequota/delete', {
      params: {
        namespace: quota.value.metadata?.namespace,
        name: quota.value.metadata?.name,
      }
    })
    ElMessage.success('Deleted')
    router.push('/config/resourcequotas')
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error('Failed to delete')
    }
  }
}

onMounted(fetchQuota)
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <h3 style="margin: 0;">资源配额详情</h3>
        <div class="filter-right">
          <el-button @click="fetchQuota"><el-icon><Refresh /></el-icon> 刷新</el-button>
          <el-button type="danger" @click="deleteQuota"><el-icon><Delete /></el-icon> 删除</el-button>
          <el-button @click="router.push('/config/resourcequotas')">返回列表</el-button>
        </div>
      </div>
    </el-card>

    <el-row :gutter="16" v-loading="loading">
      <el-col :span="12">
        <el-card shadow="never">
          <template #header>
            <h4 style="margin: 0;">基本信息</h4>
          </template>
          <el-descriptions :column="1" border v-if="quota">
            <el-descriptions-item label="名称">{{ quota.metadata?.name }}</el-descriptions-item>
            <el-descriptions-item label="命名空间">{{ quota.metadata?.namespace }}</el-descriptions-item>
            <el-descriptions-item label="创建时间">{{ quota.metadata?.creationTimestamp }}</el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-card shadow="never" style="margin-top: 16px;">
          <template #header>
            <h4 style="margin: 0;">标签</h4>
          </template>
          <div v-if="quota?.metadata?.labels">
            <el-tag v-for="(val, key) in quota.metadata.labels" :key="key" style="margin: 4px;">
              {{ key }}={{ val }}
            </el-tag>
          </div>
          <span v-else style="color: #909399;">无标签</span>
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card shadow="never">
          <template #header>
            <h4 style="margin: 0;">使用情况</h4>
          </template>
          <div v-if="quota?.spec?.hard">
            <div v-for="(val, key) in quota.spec.hard" :key="key" class="quota-item">
              <div class="quota-header">
                <span class="quota-name">{{ formatCategory(key) }}</span>
                <span class="quota-value">{{ quota.status?.used?.[key] || '0' }} / {{ val }}</span>
              </div>
              <el-progress
                :percentage="usagePercent(val, quota.status?.used?.[key] || '0')"
                :color="progressColor(usagePercent(val, quota.status?.used?.[key] || '0'))"
                :stroke-width="16"
              />
            </div>
          </div>
          <span v-else style="color: #909399;">无配额信息</span>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="never" style="margin-top: 16px;">
      <div ref="chartRef" style="height: 400px;"></div>
    </el-card>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; justify-content: space-between; align-items: center; }
.filter-right { display: flex; align-items: center; gap: 8px; }
.quota-item { margin-bottom: 16px; }
.quota-header { display: flex; justify-content: space-between; margin-bottom: 8px; }
.quota-name { font-weight: 500; }
.quota-value { color: #909399; }
</style>
