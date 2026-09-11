<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { FullScreen, Aim } from '@element-plus/icons-vue'
import {
  getStatefulSetDetail,
  getStatefulSetYaml,
  updateStatefulSetYaml,
  deleteStatefulSet,
  scaleStatefulSet,
  restartStatefulSet,
  getStatefulSetEvents,
  getStatefulSetPods,
  updateStatefulSetImage,
  rollbackStatefulSet,
  getStatefulSetRollbacks,
} from '@/api/resource'
import YamlDrawer from '@/components/YamlDrawer.vue'
import PodListPanel from '@/components/PodListPanel.vue'
import DetailPageLayout from '@/components/DetailPageLayout.vue'
import DetailPageHeader from '@/components/DetailPageHeader.vue'
import EventsTable from '@/components/EventsTable.vue'
import LabelsBlock from '@/components/LabelsBlock.vue'
import ConditionsBlock from '@/components/ConditionsBlock.vue'
import SelectorBlock from '@/components/SelectorBlock.vue'
import StatefulSetForm from '@/views/workload/components/StatefulSetForm.vue'
import AutoscalingDrawer from '@/views/workload/components/AutoscalingDrawer.vue'
import ScaleDialog from './components/ScaleDialog.vue'
import UpdateImageDialog from './components/UpdateImageDialog.vue'
import { useDetailPage } from '@/composables/useDetailPage'
import { usePodActions } from '@/composables/usePodActions'
import { useRestartAction } from '@/composables/useRestartAction'
import { useEditDrawer } from '@/composables/useEditDrawer'
import { formatAge } from '@/utils/helpers'
import { useI18n } from 'vue-i18n'

// ---- useDetailPage composable ----
const {
  namespace,
  name,
  loading,
  detail: statefulset,
  events,
  eventsLoading,
  yamlDialogVisible,
  isRunning,
  countdown,
  currentInterval,
  availableIntervals,
  toggle,
  manualRefresh,
  setIntervalOption,
  fetchDetail,
  fetchEvents,
  handleDelete,
  handleOpenYaml,
  router,
  clusterName,
} = useDetailPage({
  resourceName: 'StatefulSet',
  fetchDetail: (p) => getStatefulSetDetail(p),
  fetchEvents: (p) => getStatefulSetEvents(p),
  deleteResource: (p) => deleteStatefulSet(p),
  listRoute: '/workloads/statefulsets',
  buildParams: () => ({ namespace, name }),
  onRefresh: async () => {
    await fetchRevisions()
    await fetchAllPods()
  },
})

// ---- Shared composables ----
const { handlePodLogs, handlePodExec, handlePodDelete } = usePodActions(clusterName)
const { handleRestart } = useRestartAction('statefulset', restartStatefulSet)
const { editDialogVisible, editFullscreen, handleEdit, handleEditSuccess, handleEditCancel } =
  useEditDrawer(fetchDetail)

const { t } = useI18n()

// ---- Revisions & Pods ----
const revisions = ref<any[]>([])
const revisionsLoading = ref(false)
const selectedRevision = ref<any>(null)
const allPods = ref<any[]>([])
const rsPods = ref<any[]>([])
const rsPodsLoading = ref(false)

// 左侧视图切换：修订历史 / 基本信息
const leftView = ref<'revisions' | 'info'>('revisions')

// ---- 对话框状态 ----
const scaleDialogVisible = ref(false)
const imageDialogVisible = ref(false)
const autoscalingDrawerVisible = ref(false)

// ---- Status ----
const statusTag = computed(() => {
  const ready = statefulset.value?.status?.readyReplicas || 0
  const desired = statefulset.value?.spec?.replicas || 0
  if (ready === desired && desired > 0) return { text: 'Ready', type: 'success' as const }
  if (ready > 0) return { text: 'Partial', type: 'warning' as const }
  return { text: 'Not Ready', type: 'danger' as const }
})

const currentReplicas = computed(() => statefulset.value?.spec?.replicas ?? 1)
const readyReplicas = computed(() => statefulset.value?.status?.readyReplicas ?? 0)

const containers = computed(() => {
  return (statefulset.value?.spec?.template?.spec?.containers || []).map((c: any) => ({
    name: c.name,
    image: c.image || '',
  }))
})

// ---- Revision management ----
async function fetchRevisions() {
  revisionsLoading.value = true
  try {
    const res: any = await getStatefulSetRollbacks({ namespace, name })
    revisions.value = res.data || []
    const currentRev = statefulset.value?.status?.currentRevision
    if (currentRev) {
      const current = revisions.value.find((r: any) => r.name === currentRev)
      if (current) {
        handleRevisionSelect(current)
        return
      }
    }
    if (revisions.value.length > 0) {
      handleRevisionSelect(revisions.value[0])
    }
  } catch (_e) {
    revisions.value = []
  } finally {
    revisionsLoading.value = false
  }
}

async function fetchAllPods() {
  rsPodsLoading.value = true
  try {
    const res: any = await getStatefulSetPods({ namespace, name })
    allPods.value = res.data?.items || res.data || []
    if (selectedRevision.value) {
      handleRevisionSelect(selectedRevision.value)
    } else {
      rsPods.value = allPods.value
    }
  } catch (_e) {
    allPods.value = []
    rsPods.value = []
  } finally {
    rsPodsLoading.value = false
  }
}

function handleRevisionSelect(rev: any) {
  selectedRevision.value = rev
  rsPods.value = allPods.value.filter((pod: any) => {
    const labels = pod.metadata?.labels || {}
    return labels['controller-revision-hash'] === rev.name
  })
}

function revisionPodCount(rev: any): number {
  return allPods.value.filter((pod: any) => {
    const labels = pod.metadata?.labels || {}
    return labels['controller-revision-hash'] === rev.name
  }).length
}

async function handleRevisionRollback(rev: any) {
  try {
    await ElMessageBox.confirm(
      `${t('workload.rollback')} revision ${rev.revision}?`,
      t('common.confirmAction'),
      { type: 'warning' },
    )
    await rollbackStatefulSet({ namespace, name, revision: rev.revision })
    ElMessage.success(t('common.success'))
    fetchDetail()
    fetchRevisions()
    fetchAllPods()
  } catch (error: any) {
    if (error !== 'cancel' && error !== 'close') {
      ElMessage.error(error?.message || t('common.failed'))
    }
  }
}

// ---- Pod operations adapter ----
function onPodLogs(pod: any) {
  handlePodLogs({ namespace: pod.metadata?.namespace || namespace, name: pod.metadata?.name })
}

function onPodExec(pod: any) {
  handlePodExec({ namespace: pod.metadata?.namespace || namespace, name: pod.metadata?.name })
}

function onPodDelete(pod: any, force?: boolean) {
  handlePodDelete(
    { namespace: pod.metadata?.namespace || namespace, name: pod.metadata?.name },
    () => {
      if (selectedRevision.value) handleRevisionSelect(selectedRevision.value)
    },
    force,
  )
}

// ---- YAML ----
function handleYamlSaved() {
  fetchDetail()
  fetchEvents()
  fetchRevisions()
}

// ---- Restart ----
function onRestart() {
  handleRestart(namespace, name, () => {
    fetchDetail()
    fetchRevisions()
    fetchAllPods()
  })
}

// ---- Scale ----
function handleScaleSuccess() {
  fetchDetail()
  fetchAllPods()
}

// ---- Update image ----
function handleUpdateImageSuccess() {
  fetchDetail()
  fetchRevisions()
  fetchAllPods()
}

// ---- Edit success ----
function onEditSuccess() {
  handleEditSuccess()
  fetchRevisions()
  fetchAllPods()
}
</script>

<template>
  <DetailPageLayout :resizable="true">
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
      @back="router.push('/workloads/statefulsets')"
    >
      <template #meta>
        <span v-if="statefulset" class="replicas-info">
          {{ statefulset.status?.readyReplicas ?? 0 }}/{{ statefulset.spec?.replicas ?? 0 }} ready
        </span>
      </template>
      <template #actions>
        <el-button type="primary" @click="scaleDialogVisible = true">{{
          t('workload.scale')
        }}</el-button>
        <el-button type="warning" @click="onRestart">{{ t('workload.restart') }}</el-button>
        <el-button type="success" @click="imageDialogVisible = true">{{
          t('workload.updateImage')
        }}</el-button>
        <el-button type="info" @click="handleEdit">{{ t('common.edit') }}</el-button>
        <el-button @click="handleOpenYaml">YAML</el-button>
        <el-button type="warning" @click="autoscalingDrawerVisible = true">{{
          t('workload.hpa')
        }}</el-button>
        <el-button type="danger" @click="handleDelete">{{ t('common.delete') }}</el-button>
      </template>
    </DetailPageHeader>

    <!-- 左侧：修订历史 / 基本信息 -->
    <template v-if="statefulset" #left>
      <div class="left-tabs">
        <el-segmented
          v-model="leftView"
          :options="[
            { label: t('workload.rollback'), value: 'revisions' },
            { label: t('config.basicInfo'), value: 'info' },
          ]"
          size="small"
          block
        />
      </div>

      <!-- 修订历史 -->
      <div v-show="leftView === 'revisions'" v-loading="revisionsLoading" class="rs-list">
        <div v-if="revisions.length === 0" class="empty-hint">{{ t('common.noData') }}</div>
        <div
          v-for="rev in revisions"
          :key="rev.revision"
          class="rs-item"
          :class="{ active: selectedRevision?.name === rev.name }"
          @click="handleRevisionSelect(rev)"
        >
          <div class="rs-name">{{ rev.name }}</div>
          <div class="rs-meta">
            <span class="rs-rev">v{{ rev.revision }}</span>
            <span class="rs-replicas">{{ revisionPodCount(rev) }} 个 Pod</span>
            <el-tag
              v-if="rev.name === statefulset?.status?.currentRevision"
              type="success"
              size="small"
              >{{ t('workload.current') }}</el-tag
            >
            <el-tag v-else-if="revisionPodCount(rev) > 0" type="primary" size="small">{{
              t('workload.active')
            }}</el-tag>
          </div>
          <div v-for="(img, i) in rev.images || []" :key="i" class="rs-image">{{ img }}</div>
          <div class="rs-age">{{ formatAge(rev.createdAt, false) }}</div>
          <div v-if="rev.name !== statefulset?.status?.currentRevision" class="rs-rollback">
            <el-button size="small" type="warning" @click.stop="handleRevisionRollback(rev)">{{
              t('workload.rollback')
            }}</el-button>
          </div>
        </div>
      </div>

      <!-- 基本信息 -->
      <div v-show="leftView === 'info'" class="info-body">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item :label="t('common.name')">{{
            statefulset?.metadata?.name || '-'
          }}</el-descriptions-item>
          <el-descriptions-item :label="t('common.namespace_label')">{{
            statefulset?.metadata?.namespace || '-'
          }}</el-descriptions-item>
          <el-descriptions-item :label="t('workload.replicas')">
            {{ statefulset?.spec?.replicas ?? '-' }} 期望 ·
            {{ statefulset?.status?.readyReplicas ?? 0 }} 就绪 ·
            {{ statefulset?.status?.currentReplicas ?? 0 }} 当前 ·
            {{ statefulset?.status?.updatedReplicas ?? 0 }} 更新中
          </el-descriptions-item>
          <el-descriptions-item label="serviceName">{{
            statefulset?.spec?.serviceName || '-'
          }}</el-descriptions-item>
          <el-descriptions-item label="Pod 管理策略">{{
            statefulset?.spec?.podManagementPolicy || '-'
          }}</el-descriptions-item>
          <el-descriptions-item :label="t('workload.strategy')">
            {{ statefulset?.spec?.updateStrategy?.type || '-' }}
            <span
              v-if="statefulset?.spec?.updateStrategy?.type === 'RollingUpdate'"
              class="info-sub"
            >
              (partition {{ statefulset.spec.updateStrategy?.rollingUpdate?.partition ?? 0 }})
            </span>
          </el-descriptions-item>
          <el-descriptions-item label="当前 revision">{{
            statefulset?.status?.currentRevision || '-'
          }}</el-descriptions-item>
          <el-descriptions-item label="更新 revision">{{
            statefulset?.status?.updateRevision || '-'
          }}</el-descriptions-item>
          <el-descriptions-item label="历史上限">{{
            statefulset?.spec?.revisionHistoryLimit ?? '-'
          }}</el-descriptions-item>
          <el-descriptions-item :label="t('workload.created')">{{
            statefulset?.metadata?.creationTimestamp || '-'
          }}</el-descriptions-item>
          <el-descriptions-item label="UID">{{
            statefulset?.metadata?.uid || '-'
          }}</el-descriptions-item>
        </el-descriptions>

        <div class="info-section-title">
          {{ t('workload.containers') }} {{ t('workload.image') }}
        </div>
        <div class="vct-list">
          <div
            v-for="c in statefulset?.spec?.template?.spec?.containers || []"
            :key="c.name"
            class="vct-item"
          >
            <span class="vct-name">{{ c.name }}</span>
            <span class="vct-meta">{{ c.image || '-' }}</span>
          </div>
          <div v-if="!statefulset?.spec?.template?.spec?.containers?.length" class="info-empty">
            {{ t('common.noData') }}
          </div>
        </div>

        <div class="info-section-title">Conditions</div>
        <ConditionsBlock :conditions="statefulset?.status?.conditions || []" />

        <div class="info-section-title">volumeClaimTemplates</div>
        <div v-if="statefulset?.spec?.volumeClaimTemplates?.length" class="vct-list">
          <div
            v-for="vct in statefulset.spec.volumeClaimTemplates"
            :key="vct.name"
            class="vct-item"
          >
            <span class="vct-name">{{ vct.name }}</span>
            <span class="vct-meta"
              >{{ vct.spec?.resources?.requests?.storage || '-' }} ·
              {{ (vct.spec?.accessModes || []).join(', ') || '-'
              }}<span v-if="vct.spec?.storageClassName">
                · {{ vct.spec.storageClassName }}</span
              ></span
            >
          </div>
        </div>
        <div v-else class="info-empty">{{ t('common.noData') }}</div>

        <div class="info-section-title">Selector</div>
        <SelectorBlock :selector="statefulset?.spec?.selector?.matchLabels || {}" />

        <div class="info-section-title">Labels</div>
        <LabelsBlock :labels="statefulset?.metadata?.labels || {}" />
      </div>
    </template>

    <!-- 右上：Pod 列表 -->
    <template v-if="statefulset" #right-top>
      <div class="panel-title">
        {{ t('workload.pod') }}
        <span class="count-badge">{{ rsPods.length }} 个</span>
        <span v-if="selectedRevision" class="rs-label">{{ selectedRevision.name }}</span>
      </div>
      <PodListPanel
        :pods="rsPods"
        :loading="rsPodsLoading"
        @logs="onPodLogs"
        @exec="onPodExec"
        @delete="onPodDelete"
      />
    </template>

    <!-- 右下：Events -->
    <template v-if="statefulset" #right-bottom>
      <div class="panel-title">
        {{ t('event.title') }}
        <span class="count-badge">{{ events.length }} 条</span>
      </div>
      <EventsTable :events="events" :loading="eventsLoading" time-field="last_seen" />
    </template>

    <!-- ===== Dialogs ===== -->
    <YamlDrawer
      v-model="yamlDialogVisible"
      :get-yaml="getStatefulSetYaml"
      :update-yaml="updateStatefulSetYaml"
      :namespace="namespace"
      :name="name"
      title="StatefulSet YAML"
      @saved="handleYamlSaved"
    />

    <ScaleDialog
      v-model:visible="scaleDialogVisible"
      resource-name="StatefulSet"
      :namespace="namespace"
      :name="name"
      :current-replicas="currentReplicas"
      :ready-replicas="readyReplicas"
      :scale-fn="(p) => scaleStatefulSet(p)"
      @scaled="handleScaleSuccess"
    />

    <UpdateImageDialog
      v-model:visible="imageDialogVisible"
      resource-name="StatefulSet"
      :namespace="namespace"
      :name="name"
      :containers="containers"
      :update-image-fn="(p) => updateStatefulSetImage(p)"
      @updated="handleUpdateImageSuccess"
    />

    <el-drawer
      v-model="editDialogVisible"
      :title="t('common.edit') + ' StatefulSet'"
      :size="editFullscreen ? '100%' : '85%'"
      direction="rtl"
      :destroy-on-close="true"
      :body-style="{ padding: '0', height: '100%' }"
    >
      <template #header>
        <div class="drawer-header">
          <span class="drawer-title">{{ t('common.edit') }} StatefulSet</span>
          <el-tooltip :content="editFullscreen ? '退出全屏' : '全屏'" placement="top">
            <el-icon class="fullscreen-btn" @click="editFullscreen = !editFullscreen">
              <FullScreen v-if="!editFullscreen" />
              <Aim v-else />
            </el-icon>
          </el-tooltip>
        </div>
      </template>
      <div style="height: calc(100dvh - 52px); overflow-y: auto">
        <StatefulSetForm
          v-if="editDialogVisible && statefulset"
          :is-edit="true"
          :initial-data="statefulset"
          @success="onEditSuccess"
          @cancel="handleEditCancel"
        />
      </div>
    </el-drawer>

    <AutoscalingDrawer
      v-model:visible="autoscalingDrawerVisible"
      :namespace="namespace"
      :workload-name="name"
      workload-kind="StatefulSet"
    />
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

.rs-label {
  margin-left: auto;
  font-weight: 400;
  font-size: 11px;
  color: var(--gk-color-text-placeholder);
  font-family: var(--gk-font-mono);
  max-width: 140px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.left-tabs {
  padding: 8px 10px;
  border-bottom: 1px solid var(--el-border-color-extra-light);
  flex-shrink: 0;
}

.rs-list {
  flex: 1;
  overflow-y: auto;
}

.info-body {
  flex: 1;
  overflow-y: auto;
  padding: 10px;
}

.info-sub {
  font-size: 11px;
  color: var(--gk-color-text-secondary);
}

.info-section-title {
  font-size: var(--gk-font-size-xs);
  font-weight: 600;
  color: var(--gk-color-text-primary);
  margin: 12px 0 6px;
}

.vct-list {
  display: flex;
  flex-direction: column;
  gap: var(--gk-space-1);
}

.vct-item {
  font-size: 12px;
  display: flex;
  justify-content: space-between;
  gap: 8px;
  padding: 4px 6px;
  background: var(--gk-neutral-100);
  border-radius: 4px;
}

.vct-name {
  font-family: var(--gk-font-mono);
  color: var(--gk-color-text-primary);
}

.vct-meta {
  color: var(--gk-color-text-secondary);
  font-size: 11px;
}

.info-empty {
  font-size: 12px;
  color: var(--gk-color-text-placeholder);
}

.rs-item {
  padding: var(--gk-space-2) var(--gk-space-4);
  border-bottom: 1px solid var(--el-border-color-extra-light);
  cursor: pointer;
  transition: background 0.15s;
}

.rs-item:hover {
  background: var(--el-fill-color-light);
}

.rs-item.active {
  background: var(--el-color-primary-light-9);
  border-left: 3px solid var(--el-color-primary);
}

.rs-name {
  font-size: var(--gk-font-size-sm);
  font-weight: 500;
  font-family: var(--gk-font-mono);
  word-break: break-all;
  margin-bottom: 4px;
}

.rs-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.rs-rev {
  font-size: 12px;
  color: var(--el-color-primary);
  font-weight: 500;
}

.rs-replicas {
  font-size: 12px;
  color: var(--gk-color-text-secondary);
}

.rs-image {
  font-size: 11px;
  color: var(--gk-color-text-secondary);
  word-break: break-all;
  margin-bottom: 2px;
}

.rs-age {
  font-size: 11px;
  color: var(--gk-color-text-placeholder);
}

.rs-rollback {
  margin-top: 6px;
}

.empty-hint {
  padding: 24px;
  text-align: center;
  color: var(--gk-color-text-secondary);
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
