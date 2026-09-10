<script setup lang="ts">
import { computed } from 'vue'
import { FullScreen, Aim, VideoPause, VideoPlay } from '@element-plus/icons-vue'
import {
  getCronJobDetail,
  getCronJobYaml,
  updateCronJobYaml,
  deleteCronJob,
  getCronJobEvents,
  getCronJobExecutionHistory,
  triggerCronJob,
  suspendCronJob,
  resumeCronJob,
} from '@/api/resource'
import { ElMessage, ElMessageBox } from 'element-plus'
import YamlDrawer from '@/components/YamlDrawer.vue'
import CronJobForm from '@/views/workload/components/CronJobForm.vue'
import DetailPageLayout from '@/components/DetailPageLayout.vue'
import DetailPageHeader from '@/components/DetailPageHeader.vue'
import EventsTable from '@/components/EventsTable.vue'
import LabelsBlock from '@/components/LabelsBlock.vue'
import { useDetailPage } from '@/composables/useDetailPage'
import { useEditDrawer } from '@/composables/useEditDrawer'
import { formatAge } from '@/utils/helpers'
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'

// ---- useDetailPage composable ----
const {
  namespace, name,
  loading, detail: cronjob, events, eventsLoading, yamlDialogVisible,
  isRunning, countdown, currentInterval, availableIntervals,
  toggle, manualRefresh, setIntervalOption,
  fetchDetail, handleDelete, handleOpenYaml,
  router,
} = useDetailPage({
  resourceName: 'CronJob',
  fetchDetail: (p) => getCronJobDetail(p),
  fetchEvents: (p) => getCronJobEvents(p),
  deleteResource: (p) => deleteCronJob(p),
  listRoute: '/workloads/cronjobs',
  buildParams: () => ({ namespace, name }),
  onRefresh: async () => { await fetchJobs() },
})

// ---- Edit drawer composable ----
const { editDialogVisible, editFullscreen, handleEdit, handleEditSuccess, handleEditCancel } = useEditDrawer(fetchDetail)

// ---- Execution history ----
const jobs = ref<any[]>([])
const jobsLoading = ref(false)

async function fetchJobs() {
  jobsLoading.value = true
  try {
    const res: any = await getCronJobExecutionHistory({ namespace, name })
    jobs.value = res.data || []
  } catch {
    jobs.value = []
  } finally {
    jobsLoading.value = false
  }
}

const { t } = useI18n()

// ---- Status ----
const statusTag = computed(() => {
  if (cronjob.value?.spec?.suspend) return { text: 'Suspended', type: 'warning' as const }
  return { text: 'Active', type: 'success' as const }
})

// ---- Actions ----
function handleYamlSaved() {
  fetchDetail()
}

function onEditSuccess() {
  handleEditSuccess()
  fetchJobs()
}

async function handleTrigger() {
  try {
    await ElMessageBox.confirm(
      t('workload.triggerConfirm', { name }),
      t('common.confirmAction'),
      { type: 'info' }
    )
  } catch {
    return
  }
  try {
    await triggerCronJob({ namespace, name })
    ElMessage.success(t('workload.triggerSuccess', { name: 'CronJob' }))
    fetchJobs()
  } catch (e: any) {
    ElMessage.error(e?.message || t('workload.triggerFailed'))
  }
}

async function handleToggleSuspend() {
  if (!cronjob.value) return
  const willSuspend = !cronjob.value.spec?.suspend
  const actionLabel = willSuspend ? '暂停' : '恢复'
  try {
    await ElMessageBox.confirm(
      `确定要${actionLabel} CronJob "${name}" 吗？`,
      `确认${actionLabel}`,
      { type: 'warning' }
    )
  } catch {
    return
  }
  try {
    if (willSuspend) {
      await suspendCronJob({ namespace, name })
    } else {
      await resumeCronJob({ namespace, name })
    }
    ElMessage.success(`CronJob 已${actionLabel}`)
    fetchDetail()
  } catch (e: any) {
    ElMessage.error(e?.message || `${actionLabel}失败`)
  }
}

// ---- Job helpers ----
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

function getJobStartedAt(job: any): string {
  return job.status?.startTime || job.metadata?.creationTimestamp || ''
}

function getJobFinishedAt(job: any): string {
  return job.status?.completionTime || ''
}

function formatDateTime(value?: string): string {
  if (!value) return '-'
  return value.replace('T', ' ').replace(/\.\d+Z$/, '').replace('Z', '')
}

function getJobDuration(job: any): string {
  const startAt = getJobStartedAt(job)
  const endAt = getJobFinishedAt(job)
  if (!startAt || !endAt) return '-'

  const start = new Date(startAt).getTime()
  const end = new Date(endAt).getTime()
  if (Number.isNaN(start) || Number.isNaN(end) || end < start) return '-'

  const seconds = Math.floor((end - start) / 1000)
  if (seconds < 60) return `${seconds}s`
  const minutes = Math.floor(seconds / 60)
  const remainSeconds = seconds % 60
  if (minutes < 60) return `${minutes}m ${remainSeconds}s`
  const hours = Math.floor(minutes / 60)
  const remainMinutes = minutes % 60
  return `${hours}h ${remainMinutes}m`
}

function getJobImages(job: any): string {
  const containers = job.spec?.template?.spec?.containers || []
  return containers.map((c: any) => c.image).filter(Boolean).join(', ') || '-'
}

function isManualJob(job: any): boolean {
  return job.metadata?.annotations?.['cronjob.kubernetes.io/instantiate'] === 'manual'
}
</script>

<template>
  <DetailPageLayout v-loading="loading" :resizable="true" :initial-left-width="300">

    <!-- Header -->
    <DetailPageHeader
      :title="name"
      :status-tag="statusTag"
      :namespace="namespace"
      :loading="loading"
      :is-running="isRunning"
      :countdown="countdown"
      :current-interval="currentInterval"
      :available-intervals="availableIntervals"
      @refresh="manualRefresh()"
      @toggle="toggle()"
      @set-interval-option="setIntervalOption"
      @back="router.push('/workloads/cronjobs')"
    >
      <template #meta>
        <span class="replicas-info" v-if="cronjob">{{ cronjob.spec?.schedule }}</span>
      </template>
      <template #actions>
        <el-button
          v-if="cronjob"
          :type="cronjob.spec?.suspend ? 'success' : 'warning'"
          :icon="cronjob.spec?.suspend ? VideoPlay : VideoPause"
          @click="handleToggleSuspend"
        >{{ cronjob.spec?.suspend ? '恢复' : '暂停' }}</el-button>
        <el-button type="info" @click="handleEdit">{{ t('common.edit') }}</el-button>
        <el-button @click="handleOpenYaml">YAML</el-button>
        <el-button type="primary" @click="handleTrigger">触发</el-button>
        <el-button type="danger" @click="handleDelete">{{ t('common.delete') }}</el-button>
      </template>
    </DetailPageHeader>

    <!-- Left: Basic info -->
    <template v-if="cronjob" #left>
      <div class="panel-title">{{ t('config.basicInfo') }}</div>
      <div class="info-body">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item :label="t('common.name')">{{ cronjob.metadata?.name }}</el-descriptions-item>
          <el-descriptions-item :label="t('common.namespace_label')">{{ cronjob.metadata?.namespace }}</el-descriptions-item>
          <el-descriptions-item :label="t('workload.schedule')">{{ cronjob.spec?.schedule || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('workload.suspend')">{{ cronjob.spec?.suspend ?? false }}</el-descriptions-item>
          <el-descriptions-item :label="t('workload.concurrencyPolicy')">{{ cronjob.spec?.concurrencyPolicy || 'Allow' }}</el-descriptions-item>
          <el-descriptions-item label="成功历史限制">{{ cronjob.spec?.successfulJobsHistoryLimit ?? '-' }}</el-descriptions-item>
          <el-descriptions-item label="失败历史限制">{{ cronjob.spec?.failedJobsHistoryLimit ?? '-' }}</el-descriptions-item>
          <el-descriptions-item label="最后调度">{{ cronjob.status?.lastScheduleTime || '-' }}</el-descriptions-item>
          <el-descriptions-item label="下次执行时间">
            <span v-if="cronjob.nextScheduleTime">{{ cronjob.nextScheduleTime }}</span>
            <el-tag v-else-if="cronjob.spec?.suspend" type="info" size="small">已暂停</el-tag>
            <span v-else>-</span>
          </el-descriptions-item>
          <el-descriptions-item label="活跃 Job 数">{{ cronjob.status?.active?.length ?? 0 }}</el-descriptions-item>
        </el-descriptions>

        <!-- Labels -->
        <div style="margin-top: var(--gk-space-4);">
          <h4 style="margin: 0 0 8px; font-size: var(--gk-font-size-sm);">Labels</h4>
          <LabelsBlock :labels="cronjob.metadata?.labels || {}" />
        </div>
      </div>
    </template>

    <!-- Right-top: Execution history -->
    <template v-if="cronjob" #right-top>
      <div class="panel-title execution-title">
        <div class="title-main">
          <span>执行历史</span>
          <span class="count-badge">{{ jobs.length }} 条</span>
        </div>
        <span class="title-hint">由 CronJob 创建的 Job，保留数量受成功/失败历史限制控制</span>
      </div>
      <div v-loading="jobsLoading" class="jobs-body">
        <el-table v-if="jobs.length > 0" :data="jobs" size="small" stripe>
          <el-table-column :label="t('common.name')" min-width="240" show-overflow-tooltip>
            <template #default="{ row }">
              <div class="job-name-cell">
                <el-button link type="primary" @click="router.push(`/workloads/jobs/${row.metadata?.namespace}/${row.metadata?.name}`)">
                  {{ row.metadata?.name }}
                </el-button>
                <el-tag v-if="isManualJob(row)" type="info" size="small">手动触发</el-tag>
              </div>
            </template>
          </el-table-column>
          <el-table-column :label="t('common.status')" width="100">
            <template #default="{ row }">
              <el-tag :type="getJobStatusType(row)" size="small">{{ getJobStatus(row) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="完成数" width="90">
            <template #default="{ row }">
              {{ row.status?.succeeded || 0 }}/{{ row.spec?.completions || 1 }}
            </template>
          </el-table-column>
          <el-table-column label="开始时间" width="150">
            <template #default="{ row }">{{ formatDateTime(getJobStartedAt(row)) }}</template>
          </el-table-column>
          <el-table-column label="完成时间" width="150">
            <template #default="{ row }">{{ formatDateTime(getJobFinishedAt(row)) }}</template>
          </el-table-column>
          <el-table-column label="耗时" width="90">
            <template #default="{ row }">{{ getJobDuration(row) }}</template>
          </el-table-column>
          <el-table-column label="Age" width="100">
            <template #default="{ row }">{{ formatAge(row.metadata?.creationTimestamp) }}</template>
          </el-table-column>
          <el-table-column label="镜像" min-width="220" show-overflow-tooltip>
            <template #default="{ row }">{{ getJobImages(row) }}</template>
          </el-table-column>
        </el-table>
        <div v-else class="empty-hint">{{ t('common.noData') }}</div>
      </div>
    </template>

    <!-- Right-bottom: Events -->
    <template v-if="cronjob" #right-bottom>
      <EventsTable :events="events" :loading="eventsLoading" time-field="last_seen" />
    </template>

    <!-- YAML Drawer -->
    <YamlDrawer
      v-model="yamlDialogVisible"
      :get-yaml="getCronJobYaml"
      :update-yaml="updateCronJobYaml"
      :namespace="namespace"
      :name="name"
      title="CronJob YAML"
      @saved="handleYamlSaved"
    />

    <!-- Edit Drawer -->
    <el-drawer
      v-model="editDialogVisible"
      :title="t('common.edit') + ' CronJob'"
      :size="editFullscreen ? '100%' : '85%'"
      direction="rtl"
      :destroy-on-close="true"
      :body-style="{ padding: '0', height: '100%' }"
    >
      <template #header>
        <div class="drawer-header">
          <span class="drawer-title">{{ t('common.edit') }} CronJob</span>
          <el-tooltip :content="editFullscreen ? '退出全屏' : '全屏'" placement="top">
            <el-icon class="fullscreen-btn" @click="editFullscreen = !editFullscreen">
              <FullScreen v-if="!editFullscreen" />
              <Aim v-else />
            </el-icon>
          </el-tooltip>
        </div>
      </template>
      <div style="height: calc(100dvh - 52px); overflow-y: auto;">
        <CronJobForm
          v-if="editDialogVisible && cronjob"
          :is-edit="true"
          :initial-data="cronjob"
          @success="onEditSuccess"
          @cancel="handleEditCancel"
        />
      </div>
    </el-drawer>
  </DetailPageLayout>
</template>

<style scoped>
.replicas-info {
  font-size: 12px;
  color: var(--gk-color-text-primary);
  font-family: var(--gk-font-mono);
}

.panel-title {
  font-size: var(--gk-font-size-sm);
  font-weight: 600;
  padding: var(--gk-space-2) var(--gk-space-4);
  background: var(--gk-neutral-100);
  border-bottom: 1px solid var(--gk-color-border-light);
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.count-badge {
  font-weight: 400;
  font-size: var(--gk-font-size-xs);
  color: var(--gk-color-text-secondary);
}

.execution-title {
  justify-content: space-between;
  align-items: flex-start;
}

.title-main {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.title-hint {
  min-width: 0;
  color: var(--gk-color-text-secondary);
  font-size: 12px;
  font-weight: 400;
  text-align: right;
}

.info-body {
  flex: 1;
  overflow-y: auto;
  padding: 14px;
}

.jobs-body {
  flex: 1;
  overflow-y: auto;
  padding: 0;
}

.job-name-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.job-name-cell .el-button {
  min-width: 0;
}

.empty-hint {
  padding: 24px;
  text-align: center;
  color: var(--el-text-color-secondary);
  font-size: var(--gk-font-size-sm);
}

.drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.drawer-title {
  font-size: var(--gk-font-size-lg);
  font-weight: 600;
}

.fullscreen-btn {
  cursor: pointer;
  font-size: 18px;
  color: var(--gk-color-text-primary);
  transition: color 0.2s;
}

.fullscreen-btn:hover {
  color: var(--el-color-primary);
}
</style>
