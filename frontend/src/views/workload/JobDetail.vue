<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { FullScreen, Aim, RefreshRight } from '@element-plus/icons-vue'
import {
  getJobDetail,
  getJobYaml,
  updateJobYaml,
  deleteJob,
  getJobEvents,
  getJobPods,
  rerunJob,
} from '@/api/resource'
import YamlDrawer from '@/components/YamlDrawer.vue'
import PodListPanel from '@/components/PodListPanel.vue'
import DetailPageLayout from '@/components/DetailPageLayout.vue'
import DetailPageHeader from '@/components/DetailPageHeader.vue'
import EventsTable from '@/components/EventsTable.vue'
import LabelsBlock from '@/components/LabelsBlock.vue'
import ConditionsBlock from '@/components/ConditionsBlock.vue'
import SelectorBlock from '@/components/SelectorBlock.vue'
import JobForm from '@/views/workload/components/JobForm.vue'
import { useDetailPage } from '@/composables/useDetailPage'
import { usePodActions } from '@/composables/usePodActions'
import { useEditDrawer } from '@/composables/useEditDrawer'
import { useClusterStore } from '@/stores/cluster'

const clusterStore = useClusterStore()
const { t } = useI18n()
const clusterName = computed(() => clusterStore.clusterName)

const {
  namespace, name,
  loading, detail: job, events, eventsLoading, yamlDialogVisible,
  isRunning, countdown, currentInterval, availableIntervals,
  toggle, manualRefresh, setIntervalOption,
  fetchDetail, fetchEvents, handleDelete, handleOpenYaml,
  router,
} = useDetailPage({
  resourceName: 'Job',
  fetchDetail: getJobDetail,
  fetchEvents: getJobEvents,
  deleteResource: deleteJob,
  listRoute: '/workloads/jobs',
  buildParams: () => ({ namespace, name }),
  onRefresh: async () => { await fetchPods() },
})

const { handlePodLogs, handlePodExec, handlePodDelete } = usePodActions(clusterName)
const { editDialogVisible, editFullscreen, handleEdit, handleEditSuccess, handleEditCancel } = useEditDrawer(fetchDetail)

// ---- Pods ----
const pods = ref<any[]>([])
const podsLoading = ref(false)

// ---- Status ----
const statusTag = computed(() => {
  if (job.value?.status?.succeeded > 0) return { text: 'Complete', type: 'success' as const }
  if (job.value?.status?.active > 0) return { text: 'Running', type: 'warning' as const }
  if (job.value?.status?.failed > 0) return { text: 'Failed', type: 'danger' as const }
  return { text: 'Pending', type: 'info' as const }
})

const ownerCronJob = computed(() => {
  return job.value?.metadata?.ownerReferences?.find((ref: any) => ref.kind === 'CronJob') || null
})

const completionPercent = computed(() => {
  const total = job.value?.spec?.completions ?? 1
  const succeeded = job.value?.status?.succeeded ?? 0
  return Math.min(Math.round((succeeded / total) * 100), 100)
})

const progressStatus = computed(() => {
  if (job.value?.status?.failed > 0 && job.value?.status?.active === 0 && (job.value?.status?.succeeded ?? 0) < (job.value?.spec?.completions ?? 1)) return 'exception'
  if (completionPercent.value >= 100) return 'success'
  return ''
})

const jobConditions = computed(() => {
  return job.value?.status?.conditions || []
})

// ---- Fetches ----
async function fetchPods() {
  podsLoading.value = true
  try {
    const res: any = await getJobPods({ namespace, name })
    pods.value = res.data?.items || res.data || []
  } catch (e: any) {
    pods.value = []
  } finally {
    podsLoading.value = false
  }
}

// ---- Actions ----
function handleYamlSaved() {
  fetchDetail()
  fetchEvents()
}

async function handleRerun() {
  try {
    await ElMessageBox.confirm(
      t('workload.rerunConfirm', { name }),
      t('common.confirmAction'),
      { type: 'warning', confirmButtonText: t('common.confirm'), cancelButtonText: t('common.cancel') }
    )
    await rerunJob({ namespace, name })
    ElMessage.success(t('workload.rerunSuccess'))
    fetchDetail()
  } catch (e: any) {
    if (e === 'cancel' || e?.message === 'cancel') return
    ElMessage.error(e?.message || t('workload.rerunFailed'))
  }
}

function onEditSuccess() {
  handleEditSuccess()
  fetchPods()
}

function onPodDelete(pod: any, force?: boolean) {
  handlePodDelete(
    { namespace: pod.metadata.namespace || namespace, name: pod.metadata.name },
    fetchPods,
    force
  )
}
</script>

<template>
  <DetailPageLayout :resizable="true">
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
      @back="router.push('/workloads/jobs')"
    >
      <template #meta>
        <span class="replicas-info" v-if="job">
          {{ job.status?.succeeded ?? 0 }}/{{ job.spec?.completions ?? 1 }} completed
        </span>
      </template>
      <template #actions>
        <el-button type="success" :disabled="job?.status?.active > 0" @click="handleRerun">
          <el-icon><RefreshRight /></el-icon> 重跑
        </el-button>
        <el-button type="info" @click="handleEdit">{{ t('common.edit') }}</el-button>
        <el-button @click="handleOpenYaml">YAML</el-button>
        <el-button type="danger" @click="handleDelete">{{ t('common.delete') }}</el-button>
      </template>
    </DetailPageHeader>

    <!-- 左侧：基本信息 -->
    <template v-if="job" #left>
      <div class="panel-title">{{ t('config.basicInfo') }}</div>
      <div class="info-body">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item :label="t('common.name')">{{ job.metadata?.name }}</el-descriptions-item>
          <el-descriptions-item :label="t('common.namespace_label')">{{ job.metadata?.namespace }}</el-descriptions-item>
          <el-descriptions-item v-if="ownerCronJob" label="所属 CronJob">
            <el-button link type="primary" @click="router.push(`/workloads/cronjobs/${job.metadata?.namespace || namespace}/${ownerCronJob.name}`)">
              {{ ownerCronJob.name }}
            </el-button>
          </el-descriptions-item>
          <el-descriptions-item :label="t('workload.completions')">{{ job.spec?.completions ?? '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('workload.parallelism')">{{ job.spec?.parallelism ?? '-' }}</el-descriptions-item>
          <el-descriptions-item label="最大失败次数">{{ job.spec?.backoffLimit ?? '-' }}</el-descriptions-item>
          <el-descriptions-item label="已成功">{{ job.status?.succeeded ?? '-' }}</el-descriptions-item>
          <el-descriptions-item label="进行中">{{ job.status?.active ?? '-' }}</el-descriptions-item>
          <el-descriptions-item label="已失败">{{ job.status?.failed ?? '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('workload.startTime')">{{ job.status?.startTime || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('workload.completionTime')">{{ job.status?.completionTime || '-' }}</el-descriptions-item>
        </el-descriptions>

        <!-- Completion Progress -->
        <div style="margin-top: var(--gk-space-4);">
          <h4 style="margin: 0 0 8px; font-size: var(--gk-font-size-sm);">完成进度</h4>
          <el-progress
            :percentage="completionPercent"
            :status="progressStatus"
            :stroke-width="18"
            :text-inside="true"
            style="margin-bottom: 4px;"
          />
          <div style="font-size: 12px; color: var(--gk-color-text-secondary);">
            已成功 {{ job.status?.succeeded ?? 0 }} / {{ job.spec?.completions ?? 1 }} 个 Pod
            <template v-if="job.status?.active > 0"> · 运行中 {{ job.status.active }} 个</template>
            <template v-if="job.status?.failed > 0"> · 失败 {{ job.status.failed }} 个</template>
          </div>
        </div>

        <div style="margin-top: var(--gk-space-4);">
          <h4 style="margin: 0 0 8px; font-size: var(--gk-font-size-sm);">Labels</h4>
          <LabelsBlock :labels="job.metadata?.labels || {}" />
        </div>

        <div style="margin-top: var(--gk-space-4);">
          <h4 style="margin: 0 0 8px; font-size: var(--gk-font-size-sm);">Selector</h4>
          <SelectorBlock :selector="job.spec?.selector?.matchLabels || {}" />
        </div>

        <div v-if="jobConditions.length > 0" style="margin-top: var(--gk-space-4);">
          <h4 style="margin: 0 0 8px; font-size: var(--gk-font-size-sm);">Conditions</h4>
          <ConditionsBlock :conditions="jobConditions" />
        </div>
      </div>
    </template>

    <!-- 右上：执行 Pod 列表 -->
    <template v-if="job" #right-top>
      <div class="panel-title">
        执行 Pod
        <span class="count-badge">{{ pods.length }} 个</span>
      </div>
      <PodListPanel
        :pods="pods"
        :loading="podsLoading"
        @logs="(pod: any) => handlePodLogs({ namespace: pod.metadata.namespace || namespace, name: pod.metadata.name })"
        @exec="(pod: any) => handlePodExec({ namespace: pod.metadata.namespace || namespace, name: pod.metadata.name })"
        @delete="onPodDelete"
      />
    </template>

    <!-- 右下：Events -->
    <template v-if="job" #right-bottom>
      <div class="panel-title">
        {{ t('event.title') }}
        <span class="count-badge">{{ events.length }} 条</span>
      </div>
      <EventsTable :events="events" :loading="eventsLoading" time-field="last_seen" />
    </template>

    <!-- YAML Drawer -->
    <YamlDrawer
      v-model="yamlDialogVisible"
      :get-yaml="getJobYaml"
      :update-yaml="updateJobYaml"
      :namespace="namespace"
      :name="name"
      title="Job YAML"
      @saved="handleYamlSaved"
    />

    <!-- Edit Drawer -->
    <el-drawer
      v-model="editDialogVisible"
      :title="t('common.edit') + ' Job'"
      :size="editFullscreen ? '100%' : '85%'"
      direction="rtl"
      :destroy-on-close="true"
      :body-style="{ padding: '0', height: '100%' }"
    >
      <template #header>
        <div class="drawer-header">
          <span class="drawer-title">{{ t('common.edit') }} Job</span>
          <el-tooltip :content="editFullscreen ? '退出全屏' : '全屏'" placement="top">
            <el-icon class="fullscreen-btn" @click="editFullscreen = !editFullscreen">
              <FullScreen v-if="!editFullscreen" />
              <Aim v-else />
            </el-icon>
          </el-tooltip>
        </div>
      </template>
      <div style="height: calc(100dvh - 52px); overflow-y: auto;">
        <JobForm
          v-if="editDialogVisible && job"
          :is-edit="true"
          :initial-data="job"
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

.info-body {
  flex: 1;
  overflow-y: auto;
  padding: 14px;
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
