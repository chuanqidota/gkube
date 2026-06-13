<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Search, Download, Setting, View, Delete } from '@element-plus/icons-vue'
import request from '@/api/request'

const { t } = useI18n()
const loading = ref(false)
const searchQuery = ref('')
const selectedCategory = ref('')
const charts = ref<any[]>([])
const categories = ref<string[]>([])
const releases = ref<any[]>([])
const showReleases = ref(false)

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
    charts.value = getSampleCharts()
    categories.value = ['Web', 'Database', 'Search', 'Monitoring', 'CI/CD', 'Registry', 'Storage', 'Security']
  } finally {
    loading.value = false
  }
}

async function fetchReleases() {
  try {
    const res: any = await request.get('/k8s/catalog/releases')
    releases.value = res.data || []
  } catch {
    releases.value = []
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
    fetchReleases()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to install application')
  }
}

async function uninstallRelease(release: any) {
  try {
    await ElMessageBox.confirm(`Uninstall "${release.name}"?`, 'Confirm')
    await request.delete('/k8s/catalog/release', {
      params: { name: release.name, namespace: release.namespace }
    })
    ElMessage.success('Uninstalled')
    fetchReleases()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || 'Failed to uninstall')
    }
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
    'Security': '🔒',
    'Messaging': '📨',
  }
  return icons[category] || '📦'
}

function getSampleCharts() {
  return [
    { name: 'nginx', version: '15.0.0', appVersion: '1.25.0', description: 'High performance web server', category: 'Web', repository: 'bitnami' },
    { name: 'redis', version: '18.0.0', appVersion: '7.2.0', description: 'In-memory data store', category: 'Database', repository: 'bitnami' },
    { name: 'mysql', version: '9.0.0', appVersion: '8.0.0', description: 'Relational database', category: 'Database', repository: 'bitnami' },
    { name: 'postgresql', version: '13.0.0', appVersion: '16.0.0', description: 'Advanced relational database', category: 'Database', repository: 'bitnami' },
    { name: 'mongodb', version: '14.0.0', appVersion: '7.0.0', description: 'NoSQL document database', category: 'Database', repository: 'bitnami' },
    { name: 'elasticsearch', version: '19.0.0', appVersion: '8.10.0', description: 'Search and analytics engine', category: 'Search', repository: 'elastic' },
    { name: 'grafana', version: '7.0.0', appVersion: '10.0.0', description: 'Observability platform', category: 'Monitoring', repository: 'grafana' },
    { name: 'prometheus', version: '25.0.0', appVersion: '2.47.0', description: 'Monitoring system', category: 'Monitoring', repository: 'prometheus' },
    { name: 'jenkins', version: '4.0.0', appVersion: '2.420.0', description: 'CI/CD automation server', category: 'CI/CD', repository: 'jenkins' },
    { name: 'harbor', version: '1.0.0', appVersion: '2.9.0', description: 'Container registry', category: 'Registry', repository: 'harbor' },
    { name: 'minio', version: '5.0.0', appVersion: '2023.09.07', description: 'Object storage', category: 'Storage', repository: 'bitnami' },
    { name: 'keycloak', version: '16.0.0', appVersion: '22.0.0', description: 'Identity and access management', category: 'Security', repository: 'bitnami' },
  ]
}

onMounted(() => {
  fetchCharts()
  fetchReleases()
  fetchNamespaces()
})
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <h3 style="margin: 0;"><el-icon><Download /></el-icon> 应用目录</h3>
        <div class="filter-right">
          <el-input
            v-model="searchQuery"
            placeholder="搜索应用..."
            :prefix-icon="Search"
            style="width: 250px;"
            clearable
          />
          <el-select v-model="selectedCategory" placeholder="所有分类" clearable style="width: 150px;">
            <el-option v-for="cat in categories" :key="cat" :label="cat" :value="cat" />
          </el-select>
          <el-button :type="showReleases ? 'success' : 'default'" @click="showReleases = !showReleases">
            已安装 ({{ releases.length }})
          </el-button>
          <el-button type="primary" @click="fetchCharts"><el-icon><Refresh /></el-icon> 刷新</el-button>
        </div>
      </div>
    </el-card>

    <!-- Installed Releases -->
    <el-card v-if="showReleases" shadow="never" style="margin-bottom: 16px;">
      <template #header>
        <h4 style="margin: 0;">已安装应用</h4>
      </template>
      <el-table :data="releases" stripe size="small" empty-text="暂无已安装应用">
        <el-table-column prop="name" label="名称" min-width="150" />
        <el-table-column prop="namespace" label="命名空间" width="120" />
        <el-table-column prop="chart" label="Chart" width="150" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'deployed' ? 'success' : 'warning'" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="revision" label="版本" width="80" />
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button type="danger" size="small" @click="uninstallRelease(row)">卸载</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Chart Grid -->
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
            <el-icon><Download /></el-icon> 安装
          </el-button>
        </div>
      </el-card>
    </div>

    <el-empty v-if="!loading && filteredCharts.length === 0" description="未找到应用" />

    <!-- Chart Details Dialog -->
    <el-dialog v-model="showChartDetails" :title="chartDetails?.name" width="600px">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="名称">{{ chartDetails?.name }}</el-descriptions-item>
        <el-descriptions-item label="分类">{{ chartDetails?.category || 'Other' }}</el-descriptions-item>
        <el-descriptions-item label="版本">{{ chartDetails?.version }}</el-descriptions-item>
        <el-descriptions-item label="应用版本">{{ chartDetails?.appVersion }}</el-descriptions-item>
        <el-descriptions-item label="仓库">{{ chartDetails?.repository || '-' }}</el-descriptions-item>
        <el-descriptions-item label="描述" :span="2">{{ chartDetails?.description }}</el-descriptions-item>
      </el-descriptions>

      <h4 style="margin-top: 20px;">可用版本</h4>
      <el-table :data="chartVersions" size="small">
        <el-table-column prop="version" label="版本" />
        <el-table-column prop="appVersion" label="应用版本" />
        <el-table-column prop="created" label="创建时间" />
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="installForm.version = row.version; openInstallDialog(chartDetails)">安装</el-button>
          </template>
        </el-table-column>
      </el-table>

      <template #footer>
        <el-button @click="showChartDetails = false">关闭</el-button>
        <el-button type="primary" @click="openInstallDialog(chartDetails)">安装</el-button>
      </template>
    </el-dialog>

    <!-- Install Dialog -->
    <el-dialog v-model="showInstallDialog" :title="'安装 ' + selectedChart?.name" width="500px">
      <el-form :model="installForm" label-width="100px">
        <el-form-item label="发布名称" required>
          <el-input v-model="installForm.name" placeholder="my-release" />
        </el-form-item>
        <el-form-item label="命名空间" required>
          <el-select v-model="installForm.namespace" style="width: 100%;">
            <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
          </el-select>
        </el-form-item>
        <el-form-item label="版本">
          <el-input v-model="installForm.version" />
        </el-form-item>
        <el-form-item label="自定义配置">
          <el-input
            v-model="installForm.values"
            type="textarea"
            :rows="6"
            placeholder="key: value"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showInstallDialog = false">取消</el-button>
        <el-button type="primary" @click="installChart">安装</el-button>
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
