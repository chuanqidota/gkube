<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { getOverview, getWorkloads, getEvents, getResources } from '@/api/dashboard'
import type { Overview, WorkloadSummary, K8sEvent, ResourceMetrics } from '@/api/dashboard'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import { ElMessage } from 'element-plus'
import {
  Monitor,
  CircleCheck,
  Cpu,
  Box,
  DataLine,
  Bell,
  Refresh,
} from '@element-plus/icons-vue'

const router = useRouter()
const { t } = useI18n()

const overviewLoading = ref(false)
const workloadsLoading = ref(false)
const eventsLoading = ref(false)
const resourcesLoading = ref(false)

const overview = ref<Overview>({
  cluster_count: 0,
  node_count: 0,
  pod_count: 0,
  namespace_count: 0,
})

const resources = ref<ResourceMetrics>({
  cpu: { used: 0, total: 0 },
  memory: { used: 0, total: 0 },
  storage: { used: 0, total: 0 },
})

const workloads = ref<WorkloadSummary>({
  deployments: 0,
  statefulsets: 0,
  daemonsets: 0,
  jobs: 0,
  cronjobs: 0,
})

const events = ref<K8sEvent[]>([])

const cpuPercent = computed(() => {
  if (!resources.value.cpu.total) return 0
  return Math.round((resources.value.cpu.used / resources.value.cpu.total) * 100)
})

const memoryPercent = computed(() => {
  if (!resources.value.memory.total) return 0
  return Math.round((resources.value.memory.used / resources.value.memory.total) * 100)
})

const storagePercent = computed(() => {
  if (!resources.value.storage.total) return 0
  return Math.round((resources.value.storage.used / resources.value.storage.total) * 100)
})

function progressColor(percent: number) {
  if (percent >= 90) return 'var(--gk-color-danger)'
  if (percent >= 70) return 'var(--gk-color-warning)'
  return 'var(--gk-color-primary)'
}

const workloadItems = computed(() => [
  { label: t('workload.deployment'), value: workloads.value.deployments },
  { label: t('workload.statefulset'), value: workloads.value.statefulsets },
  { label: t('workload.daemonset'), value: workloads.value.daemonsets },
  { label: t('workload.job'), value: workloads.value.jobs },
  { label: t('workload.cronjob'), value: workloads.value.cronjobs },
])

async function fetchOverview() {
  overviewLoading.value = true
  try {
    const res = await getOverview()
    overview.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || t('dashboard.loadFailed'))
  } finally {
    overviewLoading.value = false
  }
}

async function fetchResources() {
  resourcesLoading.value = true
  try {
    const res = await getResources()
    resources.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || t('dashboard.loadFailed'))
  } finally {
    resourcesLoading.value = false
  }
}

async function fetchWorkloads() {
  workloadsLoading.value = true
  try {
    const res = await getWorkloads()
    workloads.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || t('dashboard.loadFailed'))
  } finally {
    workloadsLoading.value = false
  }
}

async function fetchEvents() {
  eventsLoading.value = true
  try {
    const res = await getEvents()
    events.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || t('dashboard.loadFailed'))
  } finally {
    eventsLoading.value = false
  }
}

async function fetchAll() {
  await Promise.all([fetchOverview(), fetchResources(), fetchWorkloads(), fetchEvents()])
}

const { isRunning, countdown, toggle, refresh: autoRefresh } = useAutoRefresh(fetchAll, 15000)

onMounted(() => {
  fetchAll()
})
</script>

<template>
  <div class="dashboard">
    <!-- Auto-refresh toolbar -->
    <div class="refresh-toolbar">
      <el-button @click="autoRefresh()" :icon="Refresh" size="small">
        {{ t('common.refresh') }} ({{ countdown }}s)
      </el-button>
      <el-button @click="toggle()" :type="isRunning ? 'warning' : 'success'" size="small">
        {{ isRunning ? t('common.paused') : t('common.resume') }}
      </el-button>
    </div>
    <!-- Stat Cards -->
    <el-row :gutter="16" class="stat-row">
      <el-col :xs="12" :sm="12" :md="6">
        <el-card v-loading="overviewLoading" shadow="hover" class="stat-card stat-card-blue" @click="router.push('/clusters')">
          <div class="stat-content">
            <div class="stat-icon">
              <el-icon :size="40"><Monitor /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ overview.cluster_count }}</div>
              <div class="stat-label">{{ t('dashboard.clusterCount') }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="12" :md="6">
        <el-card v-loading="overviewLoading" shadow="hover" class="stat-card stat-card-green" @click="router.push('/nodes')">
          <div class="stat-content">
            <div class="stat-icon">
              <el-icon :size="40"><CircleCheck /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ overview.node_count }}</div>
              <div class="stat-label">{{ t('dashboard.nodeCount') }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="12" :md="6">
        <el-card v-loading="overviewLoading" shadow="hover" class="stat-card stat-card-orange" @click="router.push('/workloads/pods')">
          <div class="stat-content">
            <div class="stat-icon">
              <el-icon :size="40"><Cpu /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ overview.pod_count }}</div>
              <div class="stat-label">{{ t('dashboard.podCount') }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="12" :md="6">
        <el-card v-loading="overviewLoading" shadow="hover" class="stat-card stat-card-purple" @click="router.push('/namespaces')">
          <div class="stat-content">
            <div class="stat-icon">
              <el-icon :size="40"><Box /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ overview.namespace_count }}</div>
              <div class="stat-label">{{ t('dashboard.namespaceCount') }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Resource Usage -->
    <el-card v-loading="resourcesLoading" shadow="hover" class="section-card">
      <template #header>
        <div class="card-header">
          <span><el-icon><Cpu /></el-icon> {{ t('dashboard.resourceUsage') }}</span>
        </div>
      </template>
      <el-row :gutter="24">
        <el-col :xs="24" :sm="24" :md="8">
          <div class="resource-item">
            <div class="resource-header">
              <span class="resource-title">{{ t('dashboard.cpu') }}</span>
              <span class="resource-detail">{{ resources.cpu.used }} / {{ resources.cpu.total }} Core</span>
            </div>
            <el-progress
              :percentage="cpuPercent"
              :color="progressColor(cpuPercent)"
              :stroke-width="18"
              :text-inside="true"
            />
          </div>
        </el-col>
        <el-col :xs="24" :sm="24" :md="8">
          <div class="resource-item">
            <div class="resource-header">
              <span class="resource-title">{{ t('dashboard.memory') }}</span>
              <span class="resource-detail">{{ resources.memory.used }} / {{ resources.memory.total }} Gi</span>
            </div>
            <el-progress
              :percentage="memoryPercent"
              :color="progressColor(memoryPercent)"
              :stroke-width="18"
              :text-inside="true"
            />
          </div>
        </el-col>
        <el-col :xs="24" :sm="24" :md="8">
          <div class="resource-item">
            <div class="resource-header">
              <span class="resource-title">{{ t('dashboard.storage') }}</span>
              <span class="resource-detail">{{ resources.storage.used }} / {{ resources.storage.total }} Gi</span>
            </div>
            <el-progress
              :percentage="storagePercent"
              :color="progressColor(storagePercent)"
              :stroke-width="18"
              :text-inside="true"
            />
          </div>
        </el-col>
      </el-row>
    </el-card>

    <!-- Workloads and Events -->
    <el-row :gutter="16">
      <el-col :xs="24" :sm="24" :md="12">
        <el-card v-loading="workloadsLoading" shadow="hover" class="section-card">
          <template #header>
            <div class="card-header">
              <span><el-icon><DataLine /></el-icon> {{ t('dashboard.workloadStats') }}</span>
            </div>
          </template>
          <div class="workload-grid">
            <div v-for="item in workloadItems" :key="item.label" class="workload-item">
              <div class="workload-count">{{ item.value }}</div>
              <div class="workload-label">{{ item.label }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="24" :md="12">
        <el-card v-loading="eventsLoading" shadow="hover" class="section-card">
          <template #header>
            <div class="card-header">
              <span><el-icon><Bell /></el-icon> {{ t('dashboard.recentEvents') }}</span>
              <el-button text size="small" @click="router.push('/events')">{{ t('dashboard.viewAll') }}</el-button>
            </div>
          </template>
          <el-table :data="events" style="width: 100%" max-height="300" size="small">
            <el-table-column prop="type" :label="t('event.type')" width="80">
              <template #default="{ row }">
                <el-tag
                  :type="row.type === 'Warning' ? 'danger' : 'info'"
                  size="small"
                  effect="plain"
                >
                  {{ row.type }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="reason" :label="t('event.reason')" width="120" />
            <el-table-column prop="involved_object" :label="t('event.involvedObject')" show-overflow-tooltip />
            <el-table-column prop="last_seen" :label="t('event.lastSeen')" width="160" />
            <template #empty>
              <el-empty :description="t('common.noData')" :image-size="60" />
            </template>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.dashboard {
  padding: 20px;
  background: var(--gk-neutral-100);
  min-height: calc(100vh - 84px);
}

.refresh-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
}

.stat-row {
  margin-bottom: 16px;
}

.stat-card {
  border-radius: 8px;
  overflow: hidden;
  border: none;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}

.stat-card:hover {
  transform: translateY(-2px);
}

.stat-card :deep(.el-card__body) {
  padding: 20px;
}

.stat-card-blue {
  background: linear-gradient(135deg, var(--gk-color-primary) 0%, #66b1ff 100%);
  color: #fff;
}

.stat-card-green {
  background: linear-gradient(135deg, var(--gk-color-success) 0%, #85ce61 100%);
  color: #fff;
}

.stat-card-orange {
  background: linear-gradient(135deg, var(--gk-color-warning) 0%, #ebb563 100%);
  color: #fff;
}

.stat-card-purple {
  background: linear-gradient(135deg, #8B5CF6 0%, #a78bfa 100%);
  color: #fff;
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-icon {
  opacity: 0.85;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 32px;
  font-weight: 700;
  line-height: 1.2;
}

.stat-label {
  font-size: 14px;
  opacity: 0.9;
  margin-top: 4px;
}

.section-card {
  margin-bottom: 16px;
  border-radius: 8px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 600;
  font-size: 15px;
}

.card-header .el-icon {
  margin-right: 6px;
  vertical-align: middle;
}

/* Resource usage */
.resource-item {
  padding: 8px 0;
}

.resource-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.resource-title {
  font-weight: 600;
  font-size: 14px;
  color: var(--gk-color-text-primary);
}

.resource-detail {
  font-size: 13px;
  color: var(--gk-color-text-secondary);
}

/* Workload grid */
.workload-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

@media (max-width: 768px) {
  .workload-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 480px) {
  .workload-grid {
    grid-template-columns: 1fr;
  }
}

.workload-item {
  text-align: center;
  padding: 16px 8px;
  background: var(--gk-neutral-100);
  border-radius: 8px;
  transition: all 0.2s;
}

.workload-item:hover {
  background: var(--gk-color-primary-bg);
  transform: translateY(-2px);
}

.workload-count {
  font-size: 28px;
  font-weight: 700;
  color: var(--gk-color-primary);
  line-height: 1.2;
}

.workload-label {
  font-size: 13px;
  color: var(--gk-color-text-secondary);
  margin-top: 6px;
}
</style>
