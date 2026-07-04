<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getDaemonSetDetail,
  getDaemonSetYaml,
  updateDaemonSetYaml,
  deleteDaemonSet,
  getDaemonSetEvents,
  getDaemonSetPods,
  deletePod,
} from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import { formatAge } from '@/utils/time'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const daemonSet = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const yamlEditorRef = ref<InstanceType<typeof YamlEditor>>()
const events = ref<any[]>([])
const eventsLoading = ref(false)
const pods = ref<any[]>([])
const podsLoading = ref(false)
const activeTab = ref('info')

const namespace = route.params.namespace as string
const name = route.params.name as string

const statusTagType = computed(() => {
  const desired = daemonSet.value?.status?.desiredNumberScheduled || 0
  const ready = daemonSet.value?.status?.numberReady || 0
  if (ready === desired && desired > 0) return 'success'
  if (ready > 0) return 'warning'
  return 'danger'
})

const statusText = computed(() => {
  const desired = daemonSet.value?.status?.desiredNumberScheduled || 0
  const ready = daemonSet.value?.status?.numberReady || 0
  if (ready === desired && desired > 0) return 'Ready'
  if (ready > 0) return 'Partial'
  return 'Not Ready'
})

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getDaemonSetDetail({ namespace, name })
    daemonSet.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load daemonset detail')
  } finally {
    loading.value = false
  }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getDaemonSetYaml({ namespace, name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load YAML')
  } finally {
    yamlLoading.value = false
  }
}

async function fetchEvents() {
  eventsLoading.value = true
  try {
    const res: any = await getDaemonSetEvents({ namespace, name })
    events.value = res.data || []
  } catch (e: any) {
    events.value = []
    ElMessage.error(e?.message || 'Failed to load events')
  } finally {
    eventsLoading.value = false
  }
}

async function fetchPods() {
  podsLoading.value = true
  try {
    const res: any = await getDaemonSetPods({ namespace, name })
    pods.value = res.data?.items || res.data || []
  } catch (e: any) {
    pods.value = []
    ElMessage.error(e?.message || 'Failed to load pods')
  } finally {
    podsLoading.value = false
  }
}

function handleOpenYaml() {
  fetchYaml()
  activeTab.value = 'yaml'
}

async function handleSaveYaml(content: string) {
  try {
    await updateDaemonSetYaml({ namespace, name, yaml: content })
    ElMessage.success('YAML saved successfully')
    fetchDetail()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to save YAML')
    yamlEditorRef.value?.resetSaving()
  }
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(
      `Are you sure to delete DaemonSet "${name}" in namespace "${namespace}"?`,
      'Confirm Delete',
      { type: 'warning' }
    )
    await deleteDaemonSet({ namespace, name })
    ElMessage.success('DaemonSet deleted')
    router.push('/workloads/daemonsets')
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || 'Delete failed')
    }
  }
}

async function handleDeletePod(pod: any) {
  try {
    await ElMessageBox.confirm(
      `Delete pod ${pod.metadata?.name}?`,
      'Confirm Delete',
      { type: 'warning' }
    )
    await deletePod({ namespace, name: pod.metadata.name })
    ElMessage.success('Pod deleted')
    fetchPods()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || 'Delete failed')
    }
  }
}

function handleTabChange(tab: string) {
  if (tab === 'yaml' && !yamlContent.value) fetchYaml()
  if (tab === 'events' && events.value.length === 0) fetchEvents()
  if (tab === 'pods' && pods.value.length === 0) fetchPods()
}

function getPodStatus(pod: any): string {
  return pod.status?.phase || 'Unknown'
}

function getPodStatusType(phase: string): string {
  if (phase === 'Running') return 'success'
  if (phase === 'Succeeded') return 'info'
  if (phase === 'Pending') return 'warning'
  return 'danger'
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchDetail, { autoStart: false })

onMounted(fetchDetail)
</script>

<template>
  <div class="detail-page" v-loading="loading">
    <!-- Header -->
    <div class="page-header">
      <div class="header-left">
        <el-button link type="primary" @click="router.push('/workloads/daemonsets')" class="back-btn">← Back to List</el-button>
        <div class="title-line">
          <h2 class="res-name">{{ name }}</h2>
          <el-tag :type="statusTagType" effect="dark" size="small">{{ statusText }}</el-tag>
          <span class="ns-tag">ns/{{ namespace }}</span>
          <span class="replicas-info" v-if="daemonSet">
            {{ daemonSet.status?.numberReady ?? 0 }}/{{ daemonSet.status?.desiredNumberScheduled ?? 0 }} ready
          </span>
        </div>
      </div>
      <div class="header-actions">
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
        <el-button size="small" @click="handleOpenYaml">YAML</el-button>
        <el-button type="danger" size="small" @click="handleDelete">Delete</el-button>
      </div>
    </div>

    <template v-if="daemonSet">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <!-- Info Tab -->
        <el-tab-pane label="Info" name="info">
          <el-descriptions :column="2" border style="margin-top: 8px;">
            <el-descriptions-item label="Name">{{ daemonSet.metadata?.name }}</el-descriptions-item>
            <el-descriptions-item label="Namespace">{{ daemonSet.metadata?.namespace }}</el-descriptions-item>
            <el-descriptions-item label="Desired Scheduled">{{ daemonSet.status?.desiredNumberScheduled ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="Current Scheduled">{{ daemonSet.status?.currentNumberScheduled ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="Number Ready">{{ daemonSet.status?.numberReady ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="Number Available">{{ daemonSet.status?.numberAvailable ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="Update Strategy">{{ daemonSet.spec?.updateStrategy?.type || 'RollingUpdate' }}</el-descriptions-item>
            <el-descriptions-item label="Creation Time">{{ daemonSet.metadata?.creationTimestamp || '-' }}</el-descriptions-item>
          </el-descriptions>

          <!-- Labels -->
          <div v-if="daemonSet.metadata?.labels && Object.keys(daemonSet.metadata.labels).length > 0" style="margin-top: 16px;">
            <h4>Labels</h4>
            <el-tag
              v-for="(val, key) in daemonSet.metadata.labels"
              :key="key"
              style="margin-right: 8px; margin-bottom: 8px;"
            >
              {{ key }}={{ val }}
            </el-tag>
          </div>

          <!-- Selector -->
          <div v-if="daemonSet.spec?.selector?.matchLabels && Object.keys(daemonSet.spec.selector.matchLabels).length > 0" style="margin-top: 16px;">
            <h4>Selector</h4>
            <el-tag
              v-for="(val, key) in daemonSet.spec.selector.matchLabels"
              :key="key"
              style="margin-right: 8px; margin-bottom: 8px;"
              type="info"
            >
              {{ key }}={{ val }}
            </el-tag>
          </div>
        </el-tab-pane>

        <!-- Pods Tab -->
        <el-tab-pane label="Pods" name="pods">
          <el-table :data="pods" v-loading="podsLoading" border size="small" style="margin-top: 8px;">
            <el-table-column label="Name" min-width="250" show-overflow-tooltip>
              <template #default="{ row }">
                <el-button link type="primary" @click="router.push(`/workloads/pods/${row.metadata?.namespace}/${row.metadata?.name}`)">
                  {{ row.metadata?.name }}
                </el-button>
              </template>
            </el-table-column>
            <el-table-column label="Status" width="120">
              <template #default="{ row }">
                <el-tag :type="getPodStatusType(getPodStatus(row))" size="small">{{ getPodStatus(row) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="Ready" width="80">
              <template #default="{ row }">
                {{ (row.status?.containerStatuses || []).filter((s: any) => s.ready).length }}/{{ row.spec?.containers?.length || 0 }}
              </template>
            </el-table-column>
            <el-table-column label="Node" prop="spec.nodeName" width="150" show-overflow-tooltip />
            <el-table-column label="IP" prop="status.podIP" width="140" />
            <el-table-column label="Restarts" width="90">
              <template #default="{ row }">
                {{ (row.status?.containerStatuses || []).reduce((s: number, c: any) => s + (c.restartCount || 0), 0) }}
              </template>
            </el-table-column>
            <el-table-column label="Age" width="120">
              <template #default="{ row }">{{ formatAge(row.metadata?.creationTimestamp) }}</template>
            </el-table-column>
            <el-table-column label="Actions" width="100" fixed="right">
              <template #default="{ row }">
                <el-button size="small" type="danger" link @click="handleDeletePod(row)">Delete</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- Events Tab -->
        <el-tab-pane label="Events" name="events">
          <el-table :data="events" v-loading="eventsLoading" border size="small" style="margin-top: 8px;">
            <el-table-column label="Type" width="100">
              <template #default="{ row }">
                <el-tag :type="row.type === 'Normal' ? 'info' : 'danger'" size="small">{{ row.type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="Reason" prop="reason" width="160" />
            <el-table-column label="Message" prop="message" min-width="300" show-overflow-tooltip />
            <el-table-column label="Last Seen" prop="last_seen" width="180" />
          </el-table>
        </el-tab-pane>

        <!-- YAML Tab -->
        <el-tab-pane label="YAML" name="yaml">
          <div v-loading="yamlLoading">
            <YamlEditor
              ref="yamlEditorRef"
              v-model="yamlContent"
              height="600px"
              :read-only="false"
              :saveable="true"
              auto-format
              @save="handleSaveYaml"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </template>
  </div>
</template>

<style scoped>
.detail-page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 20px; }
.header-left { display: flex; flex-direction: column; gap: 8px; }
.back-btn { align-self: flex-start; }
.title-line { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.res-name { margin: 0; font-size: 20px; font-weight: 600; }
.ns-tag { color: var(--el-text-color-secondary); font-size: 13px; }
.replicas-info { color: var(--el-text-color-secondary); font-size: 13px; }
.header-actions { display: flex; gap: 8px; }
</style>
