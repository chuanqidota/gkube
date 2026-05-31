<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getOverview, getWorkloads, getEvents } from '@/api/dashboard'
import type { Overview, WorkloadSummary, K8sEvent } from '@/api/dashboard'

const overviewLoading = ref(false)
const workloadsLoading = ref(false)
const eventsLoading = ref(false)

const overview = ref<Overview>({
  cluster_count: 0,
  node_count: 0,
  pod_count: 0,
  namespace_count: 0,
})

const workloads = ref<WorkloadSummary>({
  deployments: 0,
  statefulsets: 0,
  daemonsets: 0,
  jobs: 0,
  cronjobs: 0,
})

const events = ref<K8sEvent[]>([])

async function fetchOverview() {
  overviewLoading.value = true
  try {
    const res = await getOverview()
    overview.value = res.data
  } catch {
    // silently fail
  } finally {
    overviewLoading.value = false
  }
}

async function fetchWorkloads() {
  workloadsLoading.value = true
  try {
    const res = await getWorkloads()
    workloads.value = res.data
  } catch {
    // silently fail
  } finally {
    workloadsLoading.value = false
  }
}

async function fetchEvents() {
  eventsLoading.value = true
  try {
    const res = await getEvents()
    events.value = res.data
  } catch {
    // silently fail
  } finally {
    eventsLoading.value = false
  }
}

onMounted(() => {
  fetchOverview()
  fetchWorkloads()
  fetchEvents()
})
</script>

<template>
  <div class="dashboard">
    <h2>Dashboard</h2>

    <!-- Stat Cards -->
    <el-row :gutter="20" class="stat-row">
      <el-col :span="6">
        <el-card v-loading="overviewLoading" shadow="hover" class="stat-card">
          <div class="stat-label">Total Clusters</div>
          <div class="stat-value">{{ overview.cluster_count }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card v-loading="overviewLoading" shadow="hover" class="stat-card">
          <div class="stat-label">Nodes</div>
          <div class="stat-value">{{ overview.node_count }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card v-loading="overviewLoading" shadow="hover" class="stat-card">
          <div class="stat-label">Pods</div>
          <div class="stat-value">{{ overview.pod_count }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card v-loading="overviewLoading" shadow="hover" class="stat-card">
          <div class="stat-label">Namespaces</div>
          <div class="stat-value">{{ overview.namespace_count }}</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Workloads Section -->
    <el-card v-loading="workloadsLoading" shadow="hover" class="section-card">
      <template #header>
        <span>Workloads</span>
      </template>
      <el-descriptions :column="3" border>
        <el-descriptions-item label="Deployments">
          {{ workloads.deployments }}
        </el-descriptions-item>
        <el-descriptions-item label="StatefulSets">
          {{ workloads.statefulsets }}
        </el-descriptions-item>
        <el-descriptions-item label="DaemonSets">
          {{ workloads.daemonsets }}
        </el-descriptions-item>
        <el-descriptions-item label="Jobs">
          {{ workloads.jobs }}
        </el-descriptions-item>
        <el-descriptions-item label="CronJobs">
          {{ workloads.cronjobs }}
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <!-- Events Section -->
    <el-card v-loading="eventsLoading" shadow="hover" class="section-card">
      <template #header>
        <span>Recent Events</span>
      </template>
      <el-table :data="events" stripe style="width: 100%">
        <el-table-column prop="type" label="Type" width="80" />
        <el-table-column prop="reason" label="Reason" width="160" />
        <el-table-column prop="namespace" label="Namespace" width="140" />
        <el-table-column prop="involved_object" label="Object" width="180" />
        <el-table-column prop="message" label="Message" show-overflow-tooltip />
        <el-table-column prop="last_seen" label="Last Seen" width="180" />
      </el-table>
    </el-card>
  </div>
</template>

<style scoped>
.dashboard {
  padding: 24px;
}

.stat-row {
  margin-bottom: 20px;
}

.stat-card {
  text-align: center;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-bottom: 8px;
}

.stat-value {
  font-size: 32px;
  font-weight: 600;
  color: #303133;
}

.section-card {
  margin-bottom: 20px;
}
</style>
