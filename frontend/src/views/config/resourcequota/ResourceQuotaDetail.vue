<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Delete, Edit } from '@element-plus/icons-vue'
import { getResourceQuotaDetail, getResourceQuotaYaml, updateResourceQuota, deleteResourceQuota } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import * as echarts from 'echarts'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const quota = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const activeTab = ref('info')
const chartRef = ref<HTMLElement | null>(null)
const editing = ref(false)
const editYaml = ref('')
const saving = ref(false)

const namespace = route.params.namespace as string
const name = route.params.name as string

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getResourceQuotaDetail({ namespace, name })
    quota.value = res.data
    setTimeout(updateChart, 100)
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load ResourceQuota')
  } finally {
    loading.value = false
  }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getResourceQuotaYaml({ namespace, name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load YAML')
  } finally {
    yamlLoading.value = false
  }
}

function handleTabChange(tab: string) {
  if (tab === 'yaml' && !yamlContent.value) {
    fetchYaml()
  }
  if (tab === 'info') {
    setTimeout(updateChart, 100)
  }
}

function updateChart() {
  if (!chartRef.value || !quota.value) return

  const chart = echarts.init(chartRef.value)
  const hard = quota.value.spec?.hard || {}
  const used = quota.value.status?.used || {}

  const categories = Object.keys(hard)
  if (categories.length === 0) return

  const hardValues = categories.map(k => parseResourceValue(hard[k]))
  const usedValues = categories.map(k => parseResourceValue(used[k] || '0'))

  chart.setOption({
    title: { text: 'Resource Quota Usage', left: 'center' },
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    legend: { data: ['Limit', 'Used'], bottom: 0 },
    xAxis: {
      type: 'category',
      data: categories.map(formatCategory),
      axisLabel: { rotate: 45, fontSize: 10 }
    },
    yAxis: { type: 'value' },
    series: [
      {
        name: 'Limit',
        type: 'bar',
        data: hardValues,
        itemStyle: { color: '#409eff' }
      },
      {
        name: 'Used',
        type: 'bar',
        data: usedValues,
        itemStyle: { color: '#67c23a' }
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
    'requests.cpu': 'CPU Requests',
    'requests.memory': 'Memory Requests',
    'limits.cpu': 'CPU Limits',
    'limits.memory': 'Memory Limits',
    'pods': 'Pods',
    'services': 'Services',
    'secrets': 'Secrets',
    'configmaps': 'ConfigMaps',
    'persistentvolumeclaims': 'PVCs',
    'resourcequotas': 'ResourceQuotas',
    'replicationcontrollers': 'RCs',
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
  if (percent >= 90) return 'var(--gk-color-danger)'
  if (percent >= 70) return 'var(--gk-color-warning)'
  return 'var(--gk-color-primary)'
}

function handleEdit() {
  editYaml.value = yamlContent.value || ''
  editing.value = true
  if (!yamlContent.value) {
    yamlLoading.value = true
    getResourceQuotaYaml({ namespace, name }).then((res: any) => {
      editYaml.value = res.data?.yaml || res.data || ''
      yamlContent.value = editYaml.value
    }).catch((e: any) => {
      ElMessage.error(e?.message || 'Failed to load YAML')
    }).finally(() => {
      yamlLoading.value = false
    })
  }
}

function handleCancelEdit() {
  editing.value = false
}

async function handleSave() {
  saving.value = true
  try {
    await updateResourceQuota({ namespace, yaml: editYaml.value })
    ElMessage.success('ResourceQuota updated successfully')
    editing.value = false
    fetchDetail()
    yamlContent.value = editYaml.value
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to update ResourceQuota')
  } finally {
    saving.value = false
  }
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(`Delete ResourceQuota "${name}" in namespace "${namespace}"?`, 'Confirm', { type: 'warning' })
    await deleteResourceQuota({ namespace, name })
    ElMessage.success('ResourceQuota deleted')
    router.push('/config/resourcequotas')
  } catch { /* cancelled */ }
}

function handleRefresh() {
  fetchDetail()
  if (yamlContent.value) fetchYaml()
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchDetail, { autoStart: false })

onMounted(fetchDetail)
</script>

<template>
  <div class="page-container" v-loading="loading">
    <div class="page-header">
      <h2 style="margin: 0;">ResourceQuota: {{ name }}</h2>
      <div style="display: flex; gap: 8px;">
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
        <el-button @click="handleRefresh"><el-icon><Refresh /></el-icon> Refresh</el-button>
        <el-button type="primary" @click="handleEdit"><el-icon><Edit /></el-icon> Edit</el-button>
        <el-button type="danger" @click="handleDelete"><el-icon><Delete /></el-icon> 删除</el-button>
        <el-button @click="router.push('/config/resourcequotas')">Back to List</el-button>
      </div>
    </div>

    <template v-if="quota">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="Info" name="info">
          <el-row :gutter="16">
            <el-col :span="12">
              <el-card shadow="never">
                <template #header><h4 style="margin: 0;">Basic Info</h4></template>
                <el-descriptions :column="1" border>
                  <el-descriptions-item label="Name">{{ quota.name || quota.metadata?.name }}</el-descriptions-item>
                  <el-descriptions-item label="Namespace">{{ quota.namespace || quota.metadata?.namespace }}</el-descriptions-item>
                  <el-descriptions-item label="Age">{{ quota.age || '-' }}</el-descriptions-item>
                  <el-descriptions-item label="UID">{{ quota.metadata?.uid || '-' }}</el-descriptions-item>
                </el-descriptions>
              </el-card>

              <!-- Labels -->
              <el-card shadow="never" style="margin-top: 16px;">
                <template #header><h4 style="margin: 0;">Labels</h4></template>
                <div v-if="quota.labels && Object.keys(quota.labels).length > 0">
                  <el-tag v-for="(val, key) in quota.labels" :key="key" style="margin: 4px;">
                    {{ key }}={{ val }}
                  </el-tag>
                </div>
                <span v-else style="color: var(--gk-color-text-secondary);">No labels</span>
              </el-card>

              <!-- Annotations -->
              <el-card shadow="never" style="margin-top: 16px;">
                <template #header><h4 style="margin: 0;">Annotations</h4></template>
                <div v-if="quota.annotations && Object.keys(quota.annotations).length > 0">
                  <div v-for="(val, key) in quota.annotations" :key="key" class="annotation-item">
                    <span class="annotation-key">{{ key }}</span>
                    <span class="annotation-value">{{ val }}</span>
                  </div>
                </div>
                <span v-else style="color: var(--gk-color-text-secondary);">No annotations</span>
              </el-card>
            </el-col>

            <el-col :span="12">
              <el-card shadow="never">
                <template #header><h4 style="margin: 0;">Usage</h4></template>
                <div v-if="quota.spec?.hard">
                  <div v-for="(val, key) in quota.spec.hard" :key="key" class="quota-item">
                    <div class="quota-header">
                      <span class="quota-name">{{ formatCategory(String(key)) }}</span>
                      <span class="quota-value">{{ quota.status?.used?.[key] || '0' }} / {{ val }}</span>
                    </div>
                    <el-progress
                      :percentage="usagePercent(val, quota.status?.used?.[key] || '0')"
                      :color="progressColor(usagePercent(val, quota.status?.used?.[key] || '0'))"
                      :stroke-width="16"
                    />
                  </div>
                </div>
                <span v-else style="color: var(--gk-color-text-secondary);">No quota info</span>
              </el-card>
            </el-col>
          </el-row>

          <el-card shadow="never" style="margin-top: 16px;">
            <div ref="chartRef" style="height: 400px;"></div>
          </el-card>
        </el-tab-pane>

        <el-tab-pane label="YAML" name="yaml">
          <el-card shadow="never">
            <div v-if="editing">
              <div style="margin-bottom: 12px; display: flex; gap: 8px;">
                <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
                <el-button @click="handleCancelEdit">取消</el-button>
              </div>
              <YamlEditor v-model="editYaml" height="600px" />
            </div>
            <div v-else v-loading="yamlLoading">
              <div style="margin-bottom: 12px;">
                <el-button type="primary" @click="handleEdit"><el-icon><Edit /></el-icon> Edit YAML</el-button>
              </div>
              <YamlEditor v-model="yamlContent" height="600px" read-only />
            </div>
          </el-card>
        </el-tab-pane>
      </el-tabs>
    </template>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.quota-item { margin-bottom: 16px; }
.quota-header { display: flex; justify-content: space-between; margin-bottom: 8px; }
.quota-name { font-weight: 500; }
.quota-value { color: var(--gk-color-text-secondary); }
.annotation-item { display: flex; gap: 12px; margin-bottom: 8px; padding: 4px 0; border-bottom: 1px solid var(--gk-color-border-light); }
.annotation-key { font-weight: 500; min-width: 200px; word-break: break-all; }
.annotation-value { color: var(--gk-color-text-secondary); word-break: break-all; flex: 1; }
</style>
