<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { getDashboardEvents, getNamespaceList, extractNamespaceNames } from '@/api/resource'

const router = useRouter()
const loading = ref(false)
const eventList = ref<any[]>([])
const namespaceList = ref<string[]>([])

// Filters
const clusterName = ref('')
const selectedNamespace = ref('')
const selectedType = ref('')
const reasonSearch = ref('')

// Auto-refresh
const autoRefresh = ref(false)
let refreshTimer: ReturnType<typeof setInterval> | null = null

async function fetchNamespaces() {
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = extractNamespaceNames(res.data)
  } catch {
    // ignore
  }
}

async function fetchEvents() {
  loading.value = true
  try {
    const params: any = { limit: 200 }
    if (selectedType.value) params.type = selectedType.value
    const res: any = await getDashboardEvents(params)
    let events = res.data || []

    // Client-side filters
    if (selectedNamespace.value) {
      events = events.filter((e: any) => e.namespace === selectedNamespace.value)
    }
    if (clusterName.value) {
      events = events.filter((e: any) =>
        (e.clusterName || '').toLowerCase().includes(clusterName.value.toLowerCase())
      )
    }
    if (reasonSearch.value) {
      events = events.filter((e: any) =>
        (e.reason || '').toLowerCase().includes(reasonSearch.value.toLowerCase())
      )
    }

    eventList.value = events
  } catch {
    // Silently handle — resource may not exist in cluster
  } finally {
    loading.value = false
  }
}

function handleFilterChange() {
  fetchEvents()
}

function toggleAutoRefresh() {
  if (autoRefresh.value) {
    refreshTimer = setInterval(fetchEvents, 10000)
  } else {
    if (refreshTimer) {
      clearInterval(refreshTimer)
      refreshTimer = null
    }
  }
}

function handleObjectClick(row: any) {
  const object = row.object || ''
  const parts = object.split('/')
  if (parts.length !== 2) return

  const kind = parts[0].toLowerCase()
  const objName = parts[1]
  const ns = row.namespace
  const cluster = row.clusterName || ''

  const routeMap: Record<string, string> = {
    pod: `/workloads/pods/${ns}/${objName}`,
    deployment: `/workloads/deployments/${ns}/${objName}`,
    statefulset: `/workloads/statefulsets/${ns}/${objName}`,
    daemonset: `/workloads/daemonsets/${ns}/${objName}`,
    job: `/workloads/jobs/${ns}/${objName}`,
    cronjob: `/workloads/cronjobs/${ns}/${objName}`,
    service: `/services/${ns}/${objName}`,
    ingress: `/ingresses/${ns}/${objName}`,
    configmap: `/config/configmaps/${ns}/${objName}`,
    secret: `/config/secrets/${ns}/${objName}`,
    persistentvolumeclaim: `/storage/pvcs/${ns}/${objName}`,
  }

  const path = routeMap[kind]
  if (path) {
    router.push({ path, query: cluster ? { cluster } : undefined })
  }
}

function formatTime(time: string | Date | undefined): string {
  if (!time) return '-'
  const d = new Date(time)
  if (isNaN(d.getTime())) return String(time)
  return d.toLocaleString()
}

onMounted(() => {
  fetchNamespaces()
  fetchEvents()
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="margin: 0;">Events</h2>
      <div style="display: flex; gap: 12px; align-items: center;">
        <el-input
          v-model="clusterName"
          placeholder="Cluster Name"
          style="width: 160px;"
          clearable
          @clear="handleFilterChange"
          @keyup.enter="handleFilterChange"
        />
        <el-select
          v-model="selectedNamespace"
          placeholder="All Namespaces"
          clearable
          style="width: 160px;"
          @change="handleFilterChange"
        >
          <el-option
            v-for="ns in namespaceList"
            :key="ns"
            :label="ns"
            :value="ns"
          />
        </el-select>
        <el-select
          v-model="selectedType"
          placeholder="All Types"
          clearable
          style="width: 140px;"
          @change="handleFilterChange"
        >
          <el-option label="Normal" value="Normal" />
          <el-option label="Warning" value="Warning" />
        </el-select>
        <el-input
          v-model="reasonSearch"
          placeholder="Search reason..."
          style="width: 180px;"
          clearable
          @clear="handleFilterChange"
          @keyup.enter="handleFilterChange"
        />
        <el-button type="primary" @click="fetchEvents">Refresh</el-button>
        <el-tooltip content="Auto-refresh every 10 seconds" placement="top">
          <el-switch
            v-model="autoRefresh"
            active-text="Auto"
            @change="toggleAutoRefresh"
          />
        </el-tooltip>
      </div>
    </div>

    <el-table :data="eventList" v-loading="loading" stripe style="width: 100%" max-height="calc(100vh - 220px)">
      <el-table-column prop="type" label="Type" width="100">
        <template #default="{ row }">
          <el-tag :type="row.type === 'Warning' ? 'danger' : 'success'" size="small">
            {{ row.type }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="reason" label="Reason" width="160" show-overflow-tooltip />
      <el-table-column prop="object" label="Object" min-width="200" show-overflow-tooltip>
        <template #default="{ row }">
          <el-button link type="primary" @click="handleObjectClick(row)">
            {{ row.object }}
          </el-button>
        </template>
      </el-table-column>
      <el-table-column prop="namespace" label="Namespace" width="140" />
      <el-table-column prop="clusterName" label="Cluster" width="140" />
      <el-table-column prop="message" label="Message" min-width="300" show-overflow-tooltip />
      <el-table-column label="Last Seen" width="180">
        <template #default="{ row }">
          {{ formatTime(row.lastTime) }}
        </template>
      </el-table-column>
      <el-table-column prop="count" label="Count" width="80" />
    </el-table>

    <el-empty v-if="!loading && eventList.length === 0" description="No events found" />
  </div>
</template>
