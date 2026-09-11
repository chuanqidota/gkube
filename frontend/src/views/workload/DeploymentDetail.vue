<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { FullScreen, Aim } from '@element-plus/icons-vue'
import {
  getDeploymentDetail,
  getDeploymentYaml,
  updateDeploymentYaml,
  restartDeployment,
  rollbackDeployment,
  scaleDeployment,
  updateDeploymentImage,
  getDeploymentPodList,
  getDeploymentEvents,
  deleteDeployment,
  getDeploymentReplicaSets,
} from '@/api/resource'
import YamlDrawer from '@/components/YamlDrawer.vue'
import PodListPanel from '@/components/PodListPanel.vue'
import DetailPageLayout from '@/components/DetailPageLayout.vue'
import DetailPageHeader from '@/components/DetailPageHeader.vue'
import EventsTable from '@/components/EventsTable.vue'
import LabelsBlock from '@/components/LabelsBlock.vue'
import ConditionsBlock from '@/components/ConditionsBlock.vue'
import SelectorBlock from '@/components/SelectorBlock.vue'
import DeploymentForm from '@/views/workload/components/DeploymentForm.vue'
import AutoscalingDrawer from '@/views/workload/components/AutoscalingDrawer.vue'
import ScaleDialog from './components/ScaleDialog.vue'
import UpdateImageDialog from './components/UpdateImageDialog.vue'
import { useDetailPage } from '@/composables/useDetailPage'
import { usePodActions } from '@/composables/usePodActions'
import { useRestartAction } from '@/composables/useRestartAction'
import { useEditDrawer } from '@/composables/useEditDrawer'
import { useClusterStore } from '@/stores/cluster'
import { formatAge } from '@/utils/helpers'

const clusterStore = useClusterStore()
const { t } = useI18n()

// ---- useDetailPage composable ----
const {
  namespace, name,
  loading, detail: deployment, events, eventsLoading, yamlDialogVisible,
  isRunning, countdown, currentInterval, availableIntervals,
  toggle, manualRefresh, setIntervalOption,
  fetchDetail, fetchEvents, handleDelete, handleOpenYaml,
  router,
} = useDetailPage({
  resourceName: 'Deployment',
  fetchDetail: (p) => getDeploymentDetail(p),
  fetchEvents: (p) => getDeploymentEvents(p),
  deleteResource: (p) => deleteDeployment(p),
  listRoute: '/workloads/deployments',
  buildParams: () => ({ namespace, name }),
  onRefresh: fetchReplicaSets,
})

// ---- Shared composables ----
const clusterName = computed(() => clusterStore.clusterName)
const { handlePodLogs, handlePodExec, handlePodDelete } = usePodActions(clusterName)
const { handleRestart } = useRestartAction('deployment', restartDeployment)
const { editDialogVisible, editFullscreen, handleEdit, handleEditSuccess, handleEditCancel } = useEditDrawer(fetchDetail)

// ---- ReplicaSet & Pod panel state ----
const replicasets = ref<any[]>([])
const replicasetsLoading = ref(false)
const selectedReplicaset = ref<any>(null)
const rsPods = ref<any[]>([])
const allPods = ref<any[]>([])
const rsPodsLoading = ref(false)

// 左侧视图切换：修订历史 / 基本信息
const leftView = ref<'revisions' | 'info'>('revisions')

// ---- 对话框状态 ----
const scaleDialogVisible = ref(false)
const imageDialogVisible = ref(false)
const autoscalingDrawerVisible = ref(false)

// ---- Status ----
const statusTag = computed(() => {
  const conditions = deployment.value?.status?.conditions || []
  const available = conditions.find((c: any) => c.type === 'Available')
  if (available?.status === 'True') return { text: t('workload.available'), type: 'success' as const }
  const progressing = conditions.find((c: any) => c.type === 'Progressing')
  if (progressing?.status === 'True') return { text: t('workload.upToDate'), type: 'warning' as const }
  return { text: t('common.failed'), type: 'danger' as const }
})

// ---- 副本数 ----
const currentReplicas = computed(() => deployment.value?.spec?.replicas ?? 1)
const readyReplicas = computed(() => deployment.value?.status?.readyReplicas ?? 0)

const containers = computed(() => {
  return (deployment.value?.spec?.template?.spec?.containers || []).map((c: any) => ({
    name: c.name,
    image: c.image || '',
  }))
})

// ---- ReplicaSet 管理 ----
async function fetchReplicaSets() {
  replicasetsLoading.value = true
  try {
    const res: any = await getDeploymentReplicaSets({ namespace, name })
    replicasets.value = res.data?.items || res.data || []
  } catch (e) {
    console.error('Failed to fetch replicasets:', e)
    ElMessage.error(t('common.fetchFailed'))
  } finally {
    replicasetsLoading.value = false
  }

  await fetchAllPods()

  if (replicasets.value.length > 0) {
    const currentRevision = deployment.value?.metadata?.annotations?.['deployment.kubernetes.io/revision']
    const currentRS = replicasets.value.find(
      (rs: any) => rs.metadata.annotations?.['deployment.kubernetes.io/revision'] === currentRevision
    )
    if (currentRS) {
      handleReplicasetSelect(currentRS)
    }
  }
}

async function fetchAllPods() {
  rsPodsLoading.value = true
  try {
    const res: any = await getDeploymentPodList({ namespace, name })
    allPods.value = res.data?.items || res.data || []
    rsPods.value = allPods.value
  } catch (e) {
    console.error("Failed to fetch pods:", e)
    ElMessage.error(t('common.fetchFailed'))
  } finally {
    rsPodsLoading.value = false
  }
}

function handleReplicasetSelect(rs: any) {
  selectedReplicaset.value = rs
  const hash = rs.metadata.name.split('-').pop()
  rsPods.value = allPods.value.filter((pod: any) => {
    const labels = pod.metadata?.labels || {}
    return labels['pod-template-hash'] === hash
  })
}

async function handleReplicasetRollback(rs: any) {
  const revision = rs.metadata.annotations?.['deployment.kubernetes.io/revision']
  if (!revision) return
  try {
    await ElMessageBox.confirm(
      `${t('workload.rollback')} revision ${revision}?`,
      t('common.confirmAction'),
      { type: 'warning' }
    )
    await rollbackDeployment({ namespace, name, revision: parseInt(revision, 10) })
    ElMessage.success(t('workload.rollbackSuccess', { revision }))
    fetchDetail()
    fetchReplicaSets()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(t('common.failed'))
    }
  }
}

// ---- Pod 操作适配 ----
function onPodLogs(pod: any) {
  handlePodLogs({ namespace: pod.metadata.namespace || namespace, name: pod.metadata.name })
}

function onPodExec(pod: any) {
  handlePodExec({ namespace: pod.metadata.namespace || namespace, name: pod.metadata.name })
}

function onPodDelete(pod: any, force?: boolean) {
  handlePodDelete(
    { namespace: pod.metadata.namespace || namespace, name: pod.metadata.name },
    () => { if (selectedReplicaset.value) handleReplicasetSelect(selectedReplicaset.value) },
    force
  )
}

// ---- YAML ----
function handleYamlSaved() {
  fetchDetail()
  fetchEvents()
  fetchReplicaSets()
}

// ---- 重启 ----
function onRestart() {
  handleRestart(namespace, name, () => { fetchDetail(); fetchReplicaSets() })
}

// ---- 扩缩容 ----
function handleScaleSuccess() {
  fetchDetail()
  fetchReplicaSets()
}

// ---- 更新镜像 ----
function handleUpdateImageSuccess() {
  fetchDetail()
  fetchReplicaSets()
}

// ---- 编辑成功 ----
function onEditSuccess() {
  handleEditSuccess()
  fetchReplicaSets()
}
</script>

<template>
  <DetailPageLayout :resizable="true">
    <!-- 头部 -->
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
      @back="router.push('/workloads/deployments')"
    >
      <template #meta>
        <span class="replicas-info" v-if="deployment">
          {{ deployment.status?.readyReplicas ?? 0 }}/{{ deployment.spec?.replicas ?? 0 }} ready
        </span>
      </template>
      <template #actions>
        <el-button type="primary" @click="scaleDialogVisible = true">{{ t('workload.scale') }}</el-button>
        <el-button type="warning" @click="onRestart">{{ t('workload.restart') }}</el-button>
        <el-button type="success" @click="imageDialogVisible = true">{{ t('workload.updateImage') }}</el-button>
        <el-button type="info" @click="handleEdit">{{ t('common.edit') }}</el-button>
        <el-button @click="handleOpenYaml">YAML</el-button>
        <el-button type="warning" @click="autoscalingDrawerVisible = true">{{ t('workload.hpa') }}</el-button>
        <el-button type="danger" @click="handleDelete">{{ t('common.delete') }}</el-button>
      </template>
    </DetailPageHeader>

    <!-- 左侧：ReplicaSet 列表 / 基本信息 -->
    <template v-if="deployment" #left>
        <div class="left-tabs">
          <el-segmented
            v-model="leftView"
            :options="[
              { label: 'ReplicaSet', value: 'revisions' },
              { label: t('config.basicInfo'), value: 'info' },
            ]"
            size="small"
            block
          />
        </div>

        <!-- 修订历史 -->
        <div v-show="leftView === 'revisions'" class="rs-list" v-loading="replicasetsLoading">
          <div v-if="replicasets.length === 0" class="empty-hint">{{ t('common.noData') }}</div>
          <div
            v-for="rs in replicasets"
            :key="rs.metadata.name"
            class="rs-item"
            :class="{ active: selectedReplicaset?.metadata?.name === rs.metadata.name }"
            @click="handleReplicasetSelect(rs)"
          >
            <div class="rs-name">{{ rs.metadata.name }}</div>
            <div class="rs-meta">
              <span class="rs-rev">v{{ rs.metadata.annotations?.['deployment.kubernetes.io/revision'] || '?' }}</span>
              <span class="rs-replicas">{{ rs.status?.readyReplicas ?? 0 }}/{{ rs.spec?.replicas ?? 0 }}</span>
              <el-tag
                v-if="rs.metadata.annotations?.['deployment.kubernetes.io/revision'] === deployment?.metadata?.annotations?.['deployment.kubernetes.io/revision']"
                type="success" size="small">{{ t('workload.current') }}</el-tag>
              <el-tag v-else-if="(rs.status?.readyReplicas || 0) > 0" type="primary" size="small">{{ t('workload.active') }}</el-tag>
            </div>
            <div class="rs-image">{{ rs.spec?.template?.spec?.containers?.[0]?.image || '-' }}</div>
            <div class="rs-age">{{ formatAge(rs.metadata.creationTimestamp, false) }}</div>
            <div class="rs-rollback" v-if="rs.metadata.annotations?.['deployment.kubernetes.io/revision'] !== deployment?.metadata?.annotations?.['deployment.kubernetes.io/revision']">
              <el-button size="small" type="warning" @click.stop="handleReplicasetRollback(rs)">{{ t('workload.rollback') }}</el-button>
            </div>
          </div>
        </div>

        <!-- 基本信息 -->
        <div v-show="leftView === 'info'" class="info-body">
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item :label="t('common.name')">{{ deployment?.metadata?.name || '-' }}</el-descriptions-item>
            <el-descriptions-item :label="t('common.namespace_label')">{{ deployment?.metadata?.namespace || '-' }}</el-descriptions-item>
            <el-descriptions-item :label="t('workload.replicas')">
              {{ deployment?.spec?.replicas ?? '-' }} 期望 ·
              {{ deployment?.status?.readyReplicas ?? 0 }} 就绪 ·
              {{ deployment?.status?.availableReplicas ?? 0 }} 可用 ·
              {{ deployment?.status?.updatedReplicas ?? 0 }} 更新中
            </el-descriptions-item>
            <el-descriptions-item :label="t('workload.strategy')">
              {{ deployment?.spec?.strategy?.type || '-' }}
              <span v-if="deployment?.spec?.strategy?.type === 'RollingUpdate'" class="info-sub">
                (maxSurge {{ deployment.spec.strategy.rollingUpdate?.maxSurge ?? '-' }},
                maxUnavailable {{ deployment.spec.strategy.rollingUpdate?.maxUnavailable ?? '-' }})
              </span>
            </el-descriptions-item>
            <el-descriptions-item label="Revision History Limit">{{ deployment?.spec?.revisionHistoryLimit ?? '-' }}</el-descriptions-item>
            <el-descriptions-item :label="t('workload.suspend')">
              <el-tag :type="deployment?.spec?.paused ? 'warning' : 'info'" size="small">
                {{ deployment?.spec?.paused ? t('workload.suspend') : t('common.no') }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="t('workload.created')">{{ deployment?.metadata?.creationTimestamp || '-' }}</el-descriptions-item>
            <el-descriptions-item label="UID">{{ deployment?.metadata?.uid || '-' }}</el-descriptions-item>
          </el-descriptions>

          <div class="info-section-title">{{ t('workload.containers') }} {{ t('workload.image') }}</div>
          <div class="vct-list">
            <div v-for="c in (deployment?.spec?.template?.spec?.containers || [])" :key="c.name" class="vct-item">
              <span class="vct-name">{{ c.name }}</span>
              <span class="vct-meta">{{ c.image || '-' }}</span>
            </div>
            <div v-if="!deployment?.spec?.template?.spec?.containers?.length" class="info-empty">{{ t('common.noData') }}</div>
          </div>

          <div class="info-section-title">Conditions</div>
          <ConditionsBlock :conditions="deployment?.status?.conditions || []" />

          <div class="info-section-title">Selector</div>
          <SelectorBlock :selector="deployment?.spec?.selector?.matchLabels || {}" />

          <div class="info-section-title">Labels</div>
          <LabelsBlock :labels="deployment?.metadata?.labels || {}" />
        </div>
      </template>

      <!-- 右上：Pod 列表 -->
      <template v-if="deployment" #right-top>
        <div class="panel-title">
          {{ t('workload.pod') }} {{ t('common.list') }}
          <span class="count-badge">{{ rsPods.length }}</span>
          <span class="rs-label" v-if="selectedReplicaset">{{ selectedReplicaset.metadata.name }}</span>
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
      <template v-if="deployment" #right-bottom>
        <div class="panel-title">
          {{ t('event.title') }}
          <span class="count-badge">{{ events.length }}</span>
        </div>
        <EventsTable :events="events" :loading="eventsLoading" />
      </template>

    <!-- ===== Dialogs ===== -->
    <YamlDrawer
      v-model="yamlDialogVisible"
      :get-yaml="getDeploymentYaml"
      :update-yaml="updateDeploymentYaml"
      :namespace="namespace"
      :name="name"
      title="Deployment YAML"
      @saved="handleYamlSaved"
    />

    <ScaleDialog
      v-model:visible="scaleDialogVisible"
      resource-name="Deployment"
      :namespace="namespace"
      :name="name"
      :current-replicas="currentReplicas"
      :ready-replicas="readyReplicas"
      :scale-fn="scaleDeployment"
      @scaled="handleScaleSuccess"
    />

    <UpdateImageDialog
      v-model:visible="imageDialogVisible"
      resource-name="Deployment"
      :namespace="namespace"
      :name="name"
      :containers="containers"
      :update-image-fn="updateDeploymentImage"
      @updated="handleUpdateImageSuccess"
    />

    <el-drawer
      v-model="editDialogVisible"
      :title="t('common.edit') + ' Deployment'"
      :size="editFullscreen ? '100%' : '85%'"
      direction="rtl"
      :destroy-on-close="true"
      :body-style="{ padding: '0', height: '100%' }"
    >
      <template #header>
        <div class="drawer-header">
          <span class="drawer-title">{{ t('common.edit') }} Deployment</span>
          <el-tooltip :content="editFullscreen ? '退出全屏' : '全屏'" placement="top">
            <el-icon class="fullscreen-btn" @click="editFullscreen = !editFullscreen">
              <FullScreen v-if="!editFullscreen" />
              <Aim v-else />
            </el-icon>
          </el-tooltip>
        </div>
      </template>
      <div style="height: calc(100dvh - 52px); overflow-y: auto;">
        <DeploymentForm
          v-if="editDialogVisible && deployment"
          :is-edit="true"
          :initial-data="deployment"
          @success="onEditSuccess"
          @cancel="handleEditCancel"
        />
      </div>
    </el-drawer>

    <!-- Autoscaling Drawer -->
    <AutoscalingDrawer
      v-model:visible="autoscalingDrawerVisible"
      :namespace="namespace"
      :workload-name="name"
      workload-kind="Deployment"
    />
  </DetailPageLayout>
</template>

<style scoped>
.replicas-info {
  font-size: var(--gk-font-size-xs);
  color: var(--gk-color-text-secondary);
}

.panel-title {
  font-size: var(--gk-font-size-sm);
  font-weight: 600;
  padding: var(--gk-space-2) var(--gk-space-4);
  background: var(--el-fill-color-lighter);
  border-bottom: 1px solid var(--gk-color-border-light);
  display: flex;
  align-items: center;
  gap: var(--gk-space-1);
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
  font-size: var(--gk-font-size-xs);
  color: var(--gk-color-text-placeholder);
  font-family: var(--gk-font-mono);
  max-width: 140px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.left-tabs {
  padding: var(--gk-space-2) var(--gk-space-3);
  border-bottom: 1px solid var(--gk-color-border-light);
  flex-shrink: 0;
}

.rs-list {
  flex: 1;
  overflow-y: auto;
}

.info-body {
  flex: 1;
  overflow-y: auto;
  padding: var(--gk-space-3);
}

.info-sub {
  font-size: var(--gk-font-size-xs);
  color: var(--gk-color-text-secondary);
}

.info-section-title {
  font-size: var(--gk-font-size-xs);
  font-weight: 600;
  color: var(--gk-color-text-primary);
  margin: var(--gk-space-3) 0 var(--gk-space-1);
}

.info-empty {
  font-size: var(--gk-font-size-xs);
  color: var(--gk-color-text-placeholder);
}

.vct-list {
  display: flex;
  flex-direction: column;
  gap: var(--gk-space-1);
}

.vct-item {
  display: flex;
  align-items: center;
  gap: var(--gk-space-2);
  padding: var(--gk-space-1) 0;
}

.vct-name {
  font-size: var(--gk-font-size-sm);
  font-weight: 500;
  min-width: 80px;
}

.vct-meta {
  font-size: var(--gk-font-size-xs);
  color: var(--gk-color-text-secondary);
  word-break: break-all;
}

.rs-item {
  padding: var(--gk-space-2) var(--gk-space-4);
  border-bottom: 1px solid var(--gk-color-border-light);
  cursor: pointer;
  transition: background 0.15s;
}

.rs-item:hover {
  background: var(--gk-color-primary-bg);
}

.rs-item.active {
  background: var(--gk-color-primary-bg);
  border-left: 3px solid var(--gk-color-primary);
}

.rs-name {
  font-size: var(--gk-font-size-sm);
  font-weight: 500;
  font-family: var(--gk-font-mono);
  word-break: break-all;
  margin-bottom: var(--gk-space-1);
}

.rs-meta {
  display: flex;
  align-items: center;
  gap: var(--gk-space-2);
  margin-bottom: var(--gk-space-1);
}

.rs-rev {
  font-size: var(--gk-font-size-xs);
  color: var(--gk-color-primary);
  font-weight: 500;
}

.rs-replicas {
  font-size: var(--gk-font-size-xs);
  color: var(--gk-color-text-secondary);
}

.rs-image {
  font-size: var(--gk-font-size-xs);
  color: var(--gk-color-text-secondary);
  word-break: break-all;
  margin-bottom: 2px;
}

.rs-age {
  font-size: var(--gk-font-size-xs);
  color: var(--gk-color-text-placeholder);
}

.rs-rollback {
  margin-top: 6px;
}

.empty-hint {
  padding: 24px;
  text-align: center;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.drawer-title {
  font-size: 16px;
  font-weight: 600;
}

.fullscreen-btn {
  cursor: pointer;
  font-size: 18px;
  color: var(--el-text-color-regular);
  transition: color 0.2s;
}

.fullscreen-btn:hover {
  color: var(--el-color-primary);
}
</style>
