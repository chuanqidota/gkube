<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getCronJobDetail,
  getCronJobYaml,
  updateCronJobYaml,
  deleteCronJob,
  getCronJobEvents,
  getCronJobJobs,
} from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'
import { formatAge } from '@/utils/time'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const cronJob = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const yamlEditorRef = ref<InstanceType<typeof YamlEditor>>()
const events = ref<any[]>([])
const eventsLoading = ref(false)
const jobs = ref<any[]>([])
const jobsLoading = ref(false)
const activeTab = ref('info')

const namespace = route.params.namespace as string
const name = route.params.name as string

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getCronJobDetail({ namespace, name })
    cronJob.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load cronjob detail')
  } finally {
    loading.value = false
  }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getCronJobYaml({ namespace, name })
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
    const res: any = await getCronJobEvents({ namespace, name })
    events.value = res.data || []
  } catch (e: any) {
    events.value = []
    ElMessage.error(e?.message || 'Failed to load events')
  } finally {
    eventsLoading.value = false
  }
}

async function fetchJobs() {
  jobsLoading.value = true
  try {
    const res: any = await getCronJobJobs({ namespace, name })
    jobs.value = res.data || []
  } catch (e: any) {
    jobs.value = []
    ElMessage.error(e?.message || 'Failed to load jobs')
  } finally {
    jobsLoading.value = false
  }
}

function handleOpenYaml() {
  fetchYaml()
  activeTab.value = 'yaml'
}

async function handleSaveYaml(content: string) {
  try {
    await updateCronJobYaml({ namespace, name, yaml: content })
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
      `Are you sure to delete CronJob "${name}" in namespace "${namespace}"?`,
      'Confirm Delete',
      { type: 'warning' }
    )
    await deleteCronJob({ namespace, name })
    ElMessage.success('CronJob deleted')
    router.push('/workloads/cronjobs')
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || 'Delete failed')
    }
  }
}

function handleTabChange(tab: string) {
  if (tab === 'yaml' && !yamlContent.value) fetchYaml()
  if (tab === 'events' && events.value.length === 0) fetchEvents()
  if (tab === 'jobs' && jobs.value.length === 0) fetchJobs()
}

function getJobStatus(job: any): string {
  if (job.status?.succeeded > 0) return 'Complete'
  if (job.status?.active > 0) return 'Running'
  if (job.status?.failed > 0) return 'Failed'
  return 'Pending'
}

function getJobStatusType(job: any): string {
  if (job.status?.succeeded > 0) return 'success'
  if (job.status?.active > 0) return 'warning'
  if (job.status?.failed > 0) return 'danger'
  return 'info'
}

onMounted(fetchDetail)
</script>

<template>
  <div class="detail-page" v-loading="loading">
    <!-- Header -->
    <div class="page-header">
      <div class="header-left">
        <el-button link type="primary" @click="router.push('/workloads/cronjobs')" class="back-btn">← Back to List</el-button>
        <div class="title-line">
          <h2 class="res-name">{{ name }}</h2>
          <el-tag v-if="cronJob?.spec?.suspend" type="warning" effect="dark" size="small">Suspended</el-tag>
          <el-tag v-else type="success" effect="dark" size="small">Active</el-tag>
          <span class="ns-tag">ns/{{ namespace }}</span>
          <span class="schedule-info" v-if="cronJob">{{ cronJob.spec?.schedule }}</span>
        </div>
      </div>
      <div class="header-actions">
        <el-button size="small" @click="handleOpenYaml">YAML</el-button>
        <el-button type="danger" size="small" @click="handleDelete">Delete</el-button>
      </div>
    </div>

    <template v-if="cronJob">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <!-- Info Tab -->
        <el-tab-pane label="Info" name="info">
          <el-descriptions :column="2" border style="margin-top: 8px;">
            <el-descriptions-item label="Name">{{ cronJob.metadata?.name }}</el-descriptions-item>
            <el-descriptions-item label="Namespace">{{ cronJob.metadata?.namespace }}</el-descriptions-item>
            <el-descriptions-item label="Schedule">{{ cronJob.spec?.schedule || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Suspend">{{ cronJob.spec?.suspend ?? false }}</el-descriptions-item>
            <el-descriptions-item label="Concurrency Policy">{{ cronJob.spec?.concurrencyPolicy || 'Allow' }}</el-descriptions-item>
            <el-descriptions-item label="Successful History Limit">{{ cronJob.spec?.successfulJobsHistoryLimit ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="Failed History Limit">{{ cronJob.spec?.failedJobsHistoryLimit ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="Last Schedule">{{ cronJob.status?.lastScheduleTime || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Active Jobs">{{ cronJob.status?.active?.length ?? 0 }}</el-descriptions-item>
            <el-descriptions-item label="Creation Time">{{ cronJob.metadata?.creationTimestamp || '-' }}</el-descriptions-item>
          </el-descriptions>

          <!-- Labels -->
          <div v-if="cronJob.metadata?.labels && Object.keys(cronJob.metadata.labels).length > 0" style="margin-top: 16px;">
            <h4>Labels</h4>
            <el-tag
              v-for="(val, key) in cronJob.metadata.labels"
              :key="key"
              style="margin-right: 8px; margin-bottom: 8px;"
            >
              {{ key }}={{ val }}
            </el-tag>
          </div>
        </el-tab-pane>

        <!-- Jobs Tab -->
        <el-tab-pane label="Jobs" name="jobs">
          <el-table :data="jobs" v-loading="jobsLoading" border size="small" style="margin-top: 8px;">
            <el-table-column label="Name" min-width="250" show-overflow-tooltip>
              <template #default="{ row }">
                <el-button link type="primary" @click="router.push(`/workloads/jobs/${row.metadata?.namespace}/${row.metadata?.name}`)">
                  {{ row.metadata?.name }}
                </el-button>
              </template>
            </el-table-column>
            <el-table-column label="Status" width="120">
              <template #default="{ row }">
                <el-tag :type="getJobStatusType(row)" size="small">{{ getJobStatus(row) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="Completions" width="120">
              <template #default="{ row }">
                {{ row.status?.succeeded || 0 }}/{{ row.spec?.completions || 1 }}
              </template>
            </el-table-column>
            <el-table-column label="Age" width="120">
              <template #default="{ row }">{{ formatAge(row.metadata?.creationTimestamp) }}</template>
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
.schedule-info { color: var(--el-text-color-secondary); font-size: 13px; font-family: monospace; }
.header-actions { display: flex; gap: 8px; }
</style>
