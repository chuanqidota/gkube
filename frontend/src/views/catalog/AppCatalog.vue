<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Search, Download, Setting } from '@element-plus/icons-vue'
import request from '@/api/request'

const loading = ref(false)
const searchQuery = ref('')
const selectedCategory = ref('')
const charts = ref<any[]>([])
const categories = ref<string[]>([])

// Install dialog
const showInstallDialog = ref(false)
const selectedChart = ref<any>(null)
const installForm = ref({
  name: '',
  namespace: '',
  version: '',
  values: '',
})
const namespaces = ref<string[]>([])

// Chart details
const showChartDetails = ref(false)
const chartDetails = ref<any>(null)
const chartVersions = ref<any[]>([])

const filteredCharts = computed(() => {
  let result = charts.value
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(c =>
      c.name.toLowerCase().includes(query) ||
      c.description?.toLowerCase().includes(query)
    )
  }
  if (selectedCategory.value) {
    result = result.filter(c => c.category === selectedCategory.value)
  }
  return result
})

async function fetchCharts() {
  loading.value = true
  try {
    const res: any = await request.get('/k8s/catalog/charts')
    charts.value = res.data || []
    categories.value = [...new Set(charts.value.map((c: any) => c.category).filter(Boolean))] as string[]
  } catch (e: any) {
    ElMessage.warning('Chart catalog not available')
    // Provide sample charts for demo
    charts.value = [
      { name: 'nginx', description: 'High performance web server', category: 'Web', version: '15.0.0', appVersion: '1.25.0' },
      { name: 'redis', description: 'In-memory data store', category: 'Database', version: '18.0.0', appVersion: '7.2.0' },
      { name: 'mysql', description: 'Relational database', category: 'Database', version: '9.0.0', appVersion: '8.0.0' },
      { name: 'postgresql', description: 'Advanced relational database', category: 'Database', version: '13.0.0', appVersion: '16.0.0' },
      { name: 'mongodb', description: 'NoSQL document database', category: 'Database', version: '14.0.0', appVersion: '7.0.0' },
      { name: 'elasticsearch', description: 'Search and analytics engine', category: 'Search', version: '19.0.0', appVersion: '8.10.0' },
      { name: 'grafana', description: 'Observability platform', category: 'Monitoring', version: '7.0.0', appVersion: '10.0.0' },
      { name: 'prometheus', description: 'Monitoring system', category: 'Monitoring', version: '25.0.0', appVersion: '2.47.0' },
      { name: 'jenkins', description: 'CI/CD automation server', category: 'CI/CD', version: '4.0.0', appVersion: '2.420.0' },
      { name: 'gitlab', description: 'DevOps platform', category: 'CI/CD', version: '7.0.0', appVersion: '16.0.0' },
      { name: 'harbor', description: 'Container registry', category: 'Registry', version: '1.0.0', appVersion: '2.9.0' },
      { name: 'minio', description: 'Object storage', category: 'Storage', version: '5.0.0', appVersion: '2023.09.07' },
    ]
    categories.value = ['Web', 'Database', 'Search', 'Monitoring', 'CI/CD', 'Registry', 'Storage']
  } finally {
    loading.value = false
  }
}

async function fetchNamespaces() {
  try {
    const res: any = await request.get('/k8s/namespace/list')
    namespaces.value = res.data?.map((ns: any) => ns.name) || []
  } catch {
    namespaces.value = ['default']
  }
}

function viewChartDetails(chart: any) {
  selectedChart.value = chart
  chartDetails.value = chart
  showChartDetails.value = true
  // Fetch versions
  chartVersions.value = [
    { version: chart.version, appVersion: chart.appVersion, created: '2024-01-01' },
    { version: '1.0.0', appVersion: '1.0.0', created: '2023-01-01' },
  ]
}

function openInstallDialog(chart: any) {
  selectedChart.value = chart
  installForm.value = {
    name: chart.name,
    namespace: 'default',
    version: chart.version,
    values: '',
  }
  showInstallDialog.value = true
}

async function installChart() {
  if (!installForm.value.name || !installForm.value.namespace) {
    ElMessage.warning('Please fill in all required fields')
    return
  }

  try {
    await request.post('/k8s/catalog/install', {
      chart: selectedChart.value.name,
      name: installForm.value.name,
      namespace: installForm.value.namespace,
      version: installForm.value.version,
      values: installForm.value.values,
    })
    ElMessage.success('Application installed successfully')
    showInstallDialog.value = false
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to install application')
  }
}

function getCategoryIcon(category: string) {
  const icons: Record<string, string> = {
    'Web': '🌐',
    'Database': '🗄️',
    'Search': '🔍',
    'Monitoring': '📊',
    'CI/CD': '🔄',
    'Registry': '📦',
    'Storage': '💾',
  }
  return icons[category] || '📦'
}

onMounted(() => {
  fetchCharts()
  fetchNamespaces()
})
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <h3 style="margin: 0;">Application Catalog</h3>
        <div class="filter-right">
          <el-input
            v-model="searchQuery"
            placeholder="Search charts..."
            :prefix-icon="Search"
            style="width: 250px;"
            clearable
          />
          <el-select v-model="selectedCategory" placeholder="All Categories" clearable style="width: 150px;">
            <el-option v-for="cat in categories" :key="cat" :label="cat" :value="cat" />
          </el-select>
          <el-button type="primary" @click="fetchCharts"><el-icon><Refresh /></el-icon> Refresh</el-button>
        </div>
      </div>
    </el-card>

    <div class="chart-grid" v-loading="loading">
      <el-card
        v-for="chart in filteredCharts"
        :key="chart.name"
        class="chart-card"
        shadow="hover"
        @click="viewChartDetails(chart)"
      >
        <div class="chart-header">
          <div class="chart-icon">{{ getCategoryIcon(chart.category) }}</div>
          <div class="chart-info">
            <h4 class="chart-name">{{ chart.name }}</h4>
            <el-tag size="small" type="info">{{ chart.category || 'Other' }}</el-tag>
          </div>
        </div>
        <p class="chart-description">{{ chart.description || 'No description' }}</p>
        <div class="chart-footer">
          <span class="chart-version">v{{ chart.version }}</span>
          <el-button type="primary" size="small" @click.stop="openInstallDialog(chart)">
            <el-icon><Download /></el-icon> Install
          </el-button>
        </div>
      </el-card>
    </div>

    <el-empty v-if="!loading && filteredCharts.length === 0" description="No charts found" />

    <!-- Chart Details Dialog -->
    <el-dialog v-model="showChartDetails" :title="chartDetails?.name" width="600px">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="Name">{{ chartDetails?.name }}</el-descriptions-item>
        <el-descriptions-item label="Category">{{ chartDetails?.category || 'Other' }}</el-descriptions-item>
        <el-descriptions-item label="Version">{{ chartDetails?.version }}</el-descriptions-item>
        <el-descriptions-item label="App Version">{{ chartDetails?.appVersion }}</el-descriptions-item>
        <el-descriptions-item label="Description" :span="2">{{ chartDetails?.description }}</el-descriptions-item>
      </el-descriptions>

      <h4 style="margin-top: 20px;">Available Versions</h4>
      <el-table :data="chartVersions" size="small">
        <el-table-column prop="version" label="Version" />
        <el-table-column prop="appVersion" label="App Version" />
        <el-table-column prop="created" label="Created" />
        <el-table-column label="Action" width="100">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="installForm.version = row.version; openInstallDialog(chartDetails)">Install</el-button>
          </template>
        </el-table-column>
      </el-table>

      <template #footer>
        <el-button @click="showChartDetails = false">Close</el-button>
        <el-button type="primary" @click="openInstallDialog(chartDetails)">Install</el-button>
      </template>
    </el-dialog>

    <!-- Install Dialog -->
    <el-dialog v-model="showInstallDialog" :title="'Install ' + selectedChart?.name" width="500px">
      <el-form :model="installForm" label-width="120px">
        <el-form-item label="Release Name" required>
          <el-input v-model="installForm.name" placeholder="my-release" />
        </el-form-item>
        <el-form-item label="Namespace" required>
          <el-select v-model="installForm.namespace" style="width: 100%;">
            <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
          </el-select>
        </el-form-item>
        <el-form-item label="Version">
          <el-input v-model="installForm.version" />
        </el-form-item>
        <el-form-item label="Custom Values">
          <el-input
            v-model="installForm.values"
            type="textarea"
            :rows="6"
            placeholder="key: value"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showInstallDialog = false">Cancel</el-button>
        <el-button type="primary" @click="installChart">Install</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; justify-content: space-between; align-items: center; }
.filter-right { display: flex; align-items: center; gap: 8px; }
.chart-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 16px; }
.chart-card { cursor: pointer; transition: all 0.3s; }
.chart-card:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1); }
.chart-header { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.chart-icon { font-size: 32px; }
.chart-info { flex: 1; }
.chart-name { margin: 0 0 4px; font-size: 16px; }
.chart-description { color: #606266; font-size: 14px; margin: 0 0 12px; min-height: 40px; }
.chart-footer { display: flex; justify-content: space-between; align-items: center; }
.chart-version { color: #909399; font-size: 12px; }
</style>
