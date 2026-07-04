<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { Bell, Warning, InfoFilled, Search } from '@element-plus/icons-vue'
import { getDashboardEvents, getNamespaceList, extractNamespaceNames } from '@/api/resource'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'

const { t } = useI18n()
const router = useRouter()
const loading = ref(false)
const eventList = ref<any[]>([])
const namespaceList = ref<string[]>([])

// Pagination
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(50)
const continueToken = ref('')
const hasMore = ref(false)

// Filters
const selectedNamespace = ref('')
const selectedType = ref('')
const reasonSearch = ref('')

// Auto-refresh
const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchEvents)

// Event detail drawer
const drawerVisible = ref(false)
const selectedEvent = ref<any>(null)

// Sort
const sortProp = ref('last_seen')
const sortOrder = ref<'ascending' | 'descending'>('descending')

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
    const params: any = {
      limit: pageSize.value,
    }
    if (selectedType.value) params.type = selectedType.value
    if (selectedNamespace.value) params.namespace = selectedNamespace.value
    if (continueToken.value) params.continue = continueToken.value

    const res: any = await getDashboardEvents(params)
    const data = res.data || {}

    eventList.value = data.items || []
    total.value = data.total || 0
    hasMore.value = data.has_more || false
    continueToken.value = data.continue || ''
  } catch {
    // Silently handle — resource may not exist in cluster
  } finally {
    loading.value = false
  }
}

function handleFilterChange() {
  continueToken.value = ''
  currentPage.value = 1
  fetchEvents()
}

function handlePageChange(page: number) {
  currentPage.value = page
  // Calculate continue token based on page
  continueToken.value = String((page - 1) * pageSize.value)
  fetchEvents()
}

function handleSizeChange(size: number) {
  pageSize.value = size
  continueToken.value = ''
  currentPage.value = 1
  fetchEvents()
}

function handleObjectClick(row: any) {
  const kind = (row.involved_object_kind || '').toLowerCase()
  const objName = row.involved_object_name
  const ns = row.namespace

  if (!kind || !objName) return

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
    router.push({ path })
  }
}

function showEventDetail(row: any) {
  selectedEvent.value = row
  drawerVisible.value = true
}

function formatRelativeTime(time: string | Date | undefined): string {
  if (!time) return '-'
  const date = new Date(time)
  if (isNaN(date.getTime())) return String(time)

  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const seconds = Math.floor(diff / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  if (seconds < 60) return t('event.justNow')
  if (minutes < 60) return t('event.minutesAgo', { n: minutes })
  if (hours < 24) return t('event.hoursAgo', { n: hours })
  return t('event.daysAgo', { n: days })
}

function formatFullTime(time: string | Date | undefined): string {
  if (!time) return '-'
  const d = new Date(time)
  if (isNaN(d.getTime())) return String(time)
  return d.toLocaleString()
}

function eventTypeStyle(type: string) {
  return type === 'Warning'
    ? { background: 'var(--el-color-warning-light-9)', color: 'var(--el-color-warning)' }
    : { background: 'var(--el-color-success-light-9)', color: 'var(--el-color-success)' }
}

function eventTypeIcon(type: string) {
  return type === 'Warning' ? Warning : InfoFilled
}

function handleSortChange({ prop, order }: { prop: string; order: 'ascending' | 'descending' | null }) {
  if (prop) {
    sortProp.value = prop
    sortOrder.value = order || 'descending'
  }
}

const sortedEvents = computed(() => {
  const events = [...eventList.value]
  if (sortProp.value === 'last_seen') {
    events.sort((a, b) => {
      const timeA = new Date(a.last_seen || 0).getTime()
      const timeB = new Date(b.last_seen || 0).getTime()
      return sortOrder.value === 'ascending' ? timeA - timeB : timeB - timeA
    })
  } else if (sortProp.value === 'count') {
    events.sort((a, b) => {
      return sortOrder.value === 'ascending' ? a.count - b.count : b.count - a.count
    })
  }
  return events
})

const filteredEvents = computed(() => {
  let result = sortedEvents.value
  if (reasonSearch.value) {
    const query = reasonSearch.value.toLowerCase()
    result = result.filter(e =>
      (e.reason || '').toLowerCase().includes(query) ||
      (e.message || '').toLowerCase().includes(query)
    )
  }
  return result
})

onMounted(() => {
  fetchNamespaces()
})
</script>

<template>
  <div class="event-page">
    <!-- Header -->
    <div class="page-header">
      <div class="header-left">
        <el-icon :size="20"><Bell /></el-icon>
        <h2>{{ t('event.title') }}</h2>
      </div>
      <div class="header-right">
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
      </div>
    </div>

    <!-- Filters -->
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <el-select
          v-model="selectedNamespace"
          :placeholder="t('event.allNamespaces')"
          clearable
          style="width: 180px;"
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
          :placeholder="t('event.allTypes')"
          clearable
          style="width: 140px;"
          @change="handleFilterChange"
        >
          <el-option :label="t('event.normal')" value="Normal" />
          <el-option :label="t('event.warning')" value="Warning" />
        </el-select>
        <el-input
          v-model="reasonSearch"
          :placeholder="t('event.searchEvents')"
          style="width: 250px;"
          clearable
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>
    </el-card>

    <!-- Event Table -->
    <el-card shadow="never" class="table-card">
      <el-table
        :data="filteredEvents"
        v-loading="loading"
        stripe
        style="width: 100%"
        max-height="calc(100vh - 320px)"
        :default-sort="{ prop: 'last_seen', order: 'descending' }"
        @sort-change="handleSortChange"
        @row-click="showEventDetail"
      >
        <el-table-column
          prop="type"
          :label="t('event.type')"
          width="80"
          sortable="custom"
        >
          <template #default="{ row }">
            <div
              class="event-type-badge"
              :style="eventTypeStyle(row.type)"
            >
              <el-icon :size="14">
                <component :is="eventTypeIcon(row.type)" />
              </el-icon>
            </div>
          </template>
        </el-table-column>
        <el-table-column
          prop="reason"
          :label="t('event.reason')"
          width="150"
          show-overflow-tooltip
          sortable="custom"
        />
        <el-table-column
          prop="involved_object"
          :label="t('event.object')"
          min-width="200"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            <el-button link type="primary" @click.stop="handleObjectClick(row)">
              {{ row.involved_object }}
            </el-button>
          </template>
        </el-table-column>
        <el-table-column
          prop="namespace"
          :label="t('event.namespace')"
          width="130"
          show-overflow-tooltip
          sortable="custom"
        />
        <el-table-column
          prop="message"
          :label="t('event.message')"
          min-width="300"
          show-overflow-tooltip
        />
        <el-table-column
          prop="last_seen"
          :label="t('event.lastSeen')"
          width="130"
          sortable="custom"
        >
          <template #default="{ row }">
            <el-tooltip :content="formatFullTime(row.last_seen)" placement="top">
              <span>{{ formatRelativeTime(row.last_seen) }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column
          prop="count"
          :label="t('event.count')"
          width="80"
          sortable="custom"
          align="center"
        />
      </el-table>

      <!-- Pagination -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[20, 50, 100, 200]"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="handlePageChange"
          @size-change="handleSizeChange"
        />
      </div>

      <el-empty v-if="!loading && filteredEvents.length === 0" :description="t('event.noEvents')" />
    </el-card>

    <!-- Event Detail Drawer -->
    <el-drawer
      v-model="drawerVisible"
      :title="t('event.eventDetails')"
      size="500px"
      direction="rtl"
    >
      <template v-if="selectedEvent">
        <div class="event-detail">
          <div class="detail-section">
            <h4>{{ t('event.basicInfo') }}</h4>
            <el-descriptions :column="1" border>
              <el-descriptions-item :label="t('event.type')">
                <div
                  class="event-type-badge"
                  :style="eventTypeStyle(selectedEvent.type)"
                >
                  <el-icon :size="14">
                    <component :is="eventTypeIcon(selectedEvent.type)" />
                  </el-icon>
                  {{ selectedEvent.type }}
                </div>
              </el-descriptions-item>
              <el-descriptions-item :label="t('event.reason')">
                {{ selectedEvent.reason || '-' }}
              </el-descriptions-item>
              <el-descriptions-item :label="t('event.object')">
                <el-button link type="primary" @click="handleObjectClick(selectedEvent)">
                  {{ selectedEvent.involved_object }}
                </el-button>
              </el-descriptions-item>
              <el-descriptions-item :label="t('event.namespace')">
                {{ selectedEvent.namespace || '-' }}
              </el-descriptions-item>
              <el-descriptions-item :label="t('event.cluster')">
                {{ selectedEvent.cluster_name || '-' }}
              </el-descriptions-item>
            </el-descriptions>
          </div>

          <div class="detail-section">
            <h4>{{ t('event.timeInfo') }}</h4>
            <el-descriptions :column="1" border>
              <el-descriptions-item :label="t('event.firstSeen')">
                {{ formatFullTime(selectedEvent.first_seen) }}
              </el-descriptions-item>
              <el-descriptions-item :label="t('event.lastSeen')">
                {{ formatFullTime(selectedEvent.last_seen) }}
              </el-descriptions-item>
              <el-descriptions-item :label="t('event.count')">
                {{ selectedEvent.count || 0 }}
              </el-descriptions-item>
            </el-descriptions>
          </div>

          <div class="detail-section">
            <h4>{{ t('event.sourceInfo') }}</h4>
            <el-descriptions :column="1" border>
              <el-descriptions-item :label="t('event.reportingComponent')">
                {{ selectedEvent.reporting_component || '-' }}
              </el-descriptions-item>
              <el-descriptions-item :label="t('event.reportingInstance')">
                {{ selectedEvent.reporting_instance || '-' }}
              </el-descriptions-item>
              <el-descriptions-item :label="t('event.action')">
                {{ selectedEvent.action || '-' }}
              </el-descriptions-item>
            </el-descriptions>
          </div>

          <div class="detail-section">
            <h4>{{ t('event.message') }}</h4>
            <div class="message-content">
              {{ selectedEvent.message || '-' }}
            </div>
          </div>
        </div>
      </template>
    </el-drawer>
  </div>
</template>

<script lang="ts">
export default {
  name: 'EventList',
}
</script>

<style scoped>
.event-page {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-left h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.filter-card {
  margin-bottom: 16px;
}

.filter-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.table-card {
  margin-bottom: 16px;
}

.event-type-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
  padding: 8px 0;
}

.event-detail {
  padding: 0 16px;
}

.detail-section {
  margin-bottom: 24px;
}

.detail-section h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  border-bottom: 1px solid var(--el-border-color-lighter);
  padding-bottom: 8px;
}

.message-content {
  background: var(--el-fill-color-light);
  border-radius: 4px;
  padding: 12px;
  font-family: monospace;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 300px;
  overflow-y: auto;
}

:deep(.el-table) {
  cursor: pointer;
}

:deep(.el-table .el-table__row:hover) {
  background-color: var(--el-fill-color-light);
}
</style>
