<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown } from '@element-plus/icons-vue'
import { getPodDetail, getPodYaml, deletePod, getPodEvents } from '@/api/resource'
import { formatAge } from '@/utils/helpers'
import YamlDrawer from '@/components/YamlDrawer.vue'
import DetailPageLayout from '@/components/DetailPageLayout.vue'
import DetailPageHeader from '@/components/DetailPageHeader.vue'
import EventsTable from '@/components/EventsTable.vue'
import LabelsBlock from '@/components/LabelsBlock.vue'
import ConditionsBlock from '@/components/ConditionsBlock.vue'
import { useClusterStore } from '@/stores/cluster'
import { useDetailPage } from '@/composables/useDetailPage'
import { getPodStatusType, buildFullscreenUrl } from '@/utils/pod'

const clusterStore = useClusterStore()
const { t } = useI18n()

const {
  namespace,
  name,
  loading,
  detail: pod,
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
  handleOpenYaml,
  router,
} = useDetailPage({
  resourceName: 'Pod',
  fetchDetail: async (params: any) => {
    const res: any = await getPodDetail(params)
    return { data: transformPodDetail(res.data) }
  },
  fetchEvents: getPodEvents,
  deleteResource: deletePod,
  listRoute: '/workloads/pods',
  buildParams: () => ({ namespace, name }),
})

/**
 * Transform raw K8s Pod object into flat display format.
 */
function transformPodDetail(raw: any): any {
  if (!raw) return null
  const restarts = (raw.status?.containerStatuses || []).reduce(
    (sum: number, cs: any) => sum + (cs.restartCount || 0),
    0,
  )
  const specContainers = raw.spec?.containers || []
  const statusContainers = raw.status?.containerStatuses || []
  const containers = specContainers.map((spec: any) => {
    const status = statusContainers.find((s: any) => s.name === spec.name) || {}
    let state = 'Unknown'
    let stateReason = ''
    let exitCode: number | undefined
    if (status.state?.running) state = 'Running'
    else if (status.state?.waiting) {
      state = 'Waiting'
      stateReason = status.state.waiting.reason || ''
    } else if (status.state?.terminated) {
      state = 'Terminated'
      exitCode = status.state.terminated.exitCode
    }
    return {
      name: spec.name,
      image: spec.image,
      ready: status.ready || false,
      restartCount: status.restartCount || 0,
      state,
      stateReason,
      exitCode,
      ports: spec.ports || [],
      env: spec.env || [],
      volumeMounts: spec.volumeMounts || [],
      livenessProbe: spec.livenessProbe,
      readinessProbe: spec.readinessProbe,
    }
  })
  return {
    name: raw.metadata?.name || '',
    namespace: raw.metadata?.namespace || '',
    status: raw.status?.phase || 'Unknown',
    ip: raw.status?.podIP || '',
    host_ip: raw.status?.hostIP || '',
    node: raw.spec?.nodeName || '',
    restarts,
    qos_class: raw.status?.qosClass || '',
    priority: raw.spec?.priority ?? null,
    age: formatAge(raw.metadata?.creationTimestamp, false),
    created_at: raw.metadata?.creationTimestamp || '',
    service_account: raw.spec?.serviceAccountName || '',
    labels: raw.metadata?.labels || {},
    annotations: raw.metadata?.annotations || {},
    conditions: (raw.status?.conditions || []).map((c: any) => ({
      type: c.type,
      status: c.status,
      reason: c.reason || '',
      message: c.message || '',
      last_transition_time: c.lastTransitionTime || '',
    })),
    containers,
  }
}

const statusTag = computed(() => {
  const phase = pod.value?.status || ''
  const type = getPodStatusType(phase)
  return { text: phase || '-', type: type || 'info' }
})

function containerStateType(state: string) {
  const s = (state || '').toLowerCase()
  if (s === 'running') return 'success'
  if (s === 'waiting') return 'warning'
  if (s === 'terminated') return 'danger'
  return 'info'
}

function getContainerStateLabel(container: any): string {
  if (container.state === 'Running') return t('dashboard.running')
  if (container.state === 'Waiting') return `Waiting (${container.stateReason || '-'})`
  if (container.state === 'Terminated') return `Terminated (${container.exitCode ?? '-'})`
  return container.state || '-'
}

function handleYamlSaved() {
  fetchDetail()
}

function handleLogs() {
  window.open(
    buildFullscreenUrl('logs', {
      namespace,
      pod: name,
      cluster: clusterStore.clusterName || undefined,
    }),
    '_blank',
  )
}

function handleExec() {
  window.open(
    buildFullscreenUrl('terminal', {
      namespace,
      pod: name,
      cluster: clusterStore.clusterName || undefined,
    }),
    '_blank',
  )
}

async function handleDelete(force = false) {
  if (force) {
    try {
      await ElMessageBox.confirm(
        t('common.forceDeleteResourceConfirm', { type: 'Pod', name }),
        t('common.confirmForceDelete'),
        {
          type: 'warning',
          confirmButtonText: t('common.forceDelete'),
          cancelButtonText: t('common.cancel'),
        },
      )
    } catch {
      return
    }
    try {
      await deletePod({ namespace, name, force: true })
      ElMessage.success(t('common.forceDeleteResourceSuccess', { type: 'Pod' }))
      router.push('/workloads/pods')
    } catch (e: any) {
      if (e !== 'cancel')
        ElMessage.error(e?.message || t('common.forceDeleteResourceFailed', { type: 'Pod' }))
    }
    return
  }
  try {
    await ElMessageBox.confirm(
      t('common.deleteResourceConfirmNs', { type: 'Pod', name, ns: namespace }),
      t('common.confirmDelete'),
      {
        type: 'warning',
        confirmButtonText: t('common.delete'),
        cancelButtonText: t('common.cancel'),
      },
    )
    await deletePod({ namespace, name })
    ElMessage.success(t('workload.deleteResourceSuccess', { type: 'Pod' }))
    router.push('/workloads/pods')
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || t('common.deleteFailed'))
    }
  }
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
      @back="router.push('/workloads/pods')"
    >
      <template #meta>
        <span v-if="pod" class="replicas-info">
          {{ pod.ip || '-' }}
        </span>
      </template>
      <template #actions>
        <el-button type="primary" @click="handleLogs">{{ t('log.title') }}</el-button>
        <el-button type="success" @click="handleExec">{{ t('terminal.title') }}</el-button>
        <el-button @click="handleOpenYaml">YAML</el-button>
        <el-dropdown trigger="click" @command="(cmd: string) => handleDelete(cmd === 'force')">
          <el-button type="danger">
            {{ t('common.delete') }} <el-icon><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="normal">{{ t('common.delete') }}</el-dropdown-item>
              <el-dropdown-item command="force" divided>{{
                t('common.forceDelete')
              }}</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </template>
    </DetailPageHeader>

    <!-- 左侧：基本信息 -->
    <template v-if="pod" #left>
      <div class="panel-title">{{ t('config.basicInfo') }}</div>
      <div class="info-body">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item :label="t('common.name')">{{ pod.name }}</el-descriptions-item>
          <el-descriptions-item :label="t('common.namespace_label')">{{
            pod.namespace
          }}</el-descriptions-item>
          <el-descriptions-item :label="t('common.status')">
            <el-tag :type="getPodStatusType(pod.status)" size="small">{{ pod.status }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="Pod IP">{{ pod.ip || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('workload.hostIp')">{{
            pod.host_ip || '-'
          }}</el-descriptions-item>
          <el-descriptions-item label="Node">{{ pod.node || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('workload.qosClass')">{{
            pod.qos_class || '-'
          }}</el-descriptions-item>
          <el-descriptions-item :label="t('workload.priority')">{{
            pod.priority ?? '-'
          }}</el-descriptions-item>
          <el-descriptions-item :label="t('workload.serviceAccount')">{{
            pod.service_account || '-'
          }}</el-descriptions-item>
          <el-descriptions-item :label="t('workload.restarts')">{{
            pod.restarts ?? '-'
          }}</el-descriptions-item>
          <el-descriptions-item label="Age">{{ pod.age || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('workload.created')">{{
            pod.created_at || '-'
          }}</el-descriptions-item>
        </el-descriptions>

        <!-- Labels -->
        <div
          v-if="pod.labels && Object.keys(pod.labels).length > 0"
          style="margin-top: var(--gk-space-4)"
        >
          <h4 style="margin: 0 0 8px; font-size: 13px">Labels</h4>
          <LabelsBlock :labels="pod.labels" />
        </div>

        <!-- Annotations -->
        <div
          v-if="pod.annotations && Object.keys(pod.annotations).length > 0"
          style="margin-top: var(--gk-space-4)"
        >
          <h4 style="margin: 0 0 8px; font-size: 13px">Annotations</h4>
          <div class="annotation-list">
            <div v-for="(val, key) in pod.annotations" :key="key" class="annotation-item">
              <span class="annotation-key">{{ key }}</span>
              <span class="annotation-value">{{ val }}</span>
            </div>
          </div>
        </div>

        <!-- Pod Conditions -->
        <div
          v-if="pod.conditions && pod.conditions.length > 0"
          style="margin-top: var(--gk-space-4)"
        >
          <h4 style="margin: 0 0 8px; font-size: 13px">{{ t('workload.conditions') }}</h4>
          <ConditionsBlock :conditions="pod.conditions" />
        </div>
      </div>
    </template>

    <!-- 右上：容器列表 -->
    <template v-if="pod" #right-top>
      <div class="panel-title">
        {{ t('workload.containers') }}
        <span class="count-badge">{{ pod.containers?.length || 0 }}</span>
      </div>
      <div class="table-body">
        <el-table :data="pod.containers || []" size="small" stripe>
          <el-table-column type="expand">
            <template #default="{ row }">
              <div style="padding: 12px 16px">
                <div
                  v-if="row.ports && row.ports.length > 0"
                  style="margin-bottom: var(--gk-space-4)"
                >
                  <h4 style="margin: 0 0 8px; font-size: 13px">{{ t('workload.ports') }}</h4>
                  <el-table :data="row.ports" border size="small">
                    <el-table-column prop="name" :label="t('common.name')" width="120" />
                    <el-table-column
                      prop="containerPort"
                      :label="t('workload.containerPort')"
                      width="130"
                    />
                    <el-table-column prop="protocol" :label="t('workload.protocol')" width="100" />
                  </el-table>
                </div>
                <div v-if="row.env && row.env.length > 0" style="margin-bottom: var(--gk-space-4)">
                  <h4 style="margin: 0 0 8px; font-size: 13px">
                    {{ t('workload.environmentVariables') }}
                  </h4>
                  <el-table :data="row.env" border size="small">
                    <el-table-column prop="name" :label="t('common.name')" min-width="180" />
                    <el-table-column label="值" min-width="250">
                      <template #default="{ row: envRow }">
                        <span v-if="envRow.value !== undefined && envRow.value !== ''">{{
                          envRow.value
                        }}</span>
                        <span
                          v-else-if="envRow.valueFrom"
                          style="color: var(--el-text-color-secondary)"
                          >{{
                            envRow.valueFrom.fieldRef?.fieldPath ||
                            envRow.valueFrom.secretKeyRef?.name ||
                            envRow.valueFrom.configMapKeyRef?.name ||
                            '来自引用'
                          }}</span
                        >
                        <span v-else style="color: var(--el-text-color-secondary)">-</span>
                      </template>
                    </el-table-column>
                  </el-table>
                </div>
                <div
                  v-if="row.volumeMounts && row.volumeMounts.length > 0"
                  style="margin-bottom: var(--gk-space-4)"
                >
                  <h4 style="margin: 0 0 8px; font-size: 13px">{{ t('workload.volumeMounts') }}</h4>
                  <el-table :data="row.volumeMounts" border size="small">
                    <el-table-column
                      prop="name"
                      :label="t('workload.volumeName')"
                      min-width="150"
                    />
                    <el-table-column
                      prop="mountPath"
                      :label="t('workload.mountPath')"
                      min-width="200"
                    />
                    <el-table-column prop="subPath" :label="t('workload.subPath')" width="150" />
                    <el-table-column :label="t('workload.readOnly')" width="80">
                      <template #default="{ row: vm }">
                        <el-tag :type="vm.readOnly ? 'warning' : 'success'" size="small">{{
                          vm.readOnly ? t('common.yes') : t('common.no')
                        }}</el-tag>
                      </template>
                    </el-table-column>
                  </el-table>
                </div>
                <div v-if="row.livenessProbe" style="margin-bottom: var(--gk-space-4)">
                  <h4 style="margin: 0 0 8px; font-size: 13px">
                    {{ t('workload.livenessProbe') }}
                  </h4>
                  <el-descriptions :column="2" border size="small">
                    <el-descriptions-item v-if="row.livenessProbe.httpGet" label="Type"
                      >HTTP GET</el-descriptions-item
                    >
                    <el-descriptions-item v-if="row.livenessProbe.httpGet" label="Path">{{
                      row.livenessProbe.httpGet.path || '/'
                    }}</el-descriptions-item>
                    <el-descriptions-item v-if="row.livenessProbe.httpGet" label="Port">{{
                      row.livenessProbe.httpGet.port
                    }}</el-descriptions-item>
                    <el-descriptions-item v-if="row.livenessProbe.tcpSocket" label="Type"
                      >TCP Socket</el-descriptions-item
                    >
                    <el-descriptions-item v-if="row.livenessProbe.exec" label="Type"
                      >Exec</el-descriptions-item
                    >
                    <el-descriptions-item
                      v-if="row.livenessProbe.exec"
                      :label="t('workload.command')"
                      >{{ (row.livenessProbe.exec.command || []).join(' ') }}</el-descriptions-item
                    >
                    <el-descriptions-item :label="t('workload.initialDelay')"
                      >{{ row.livenessProbe.initialDelaySeconds ?? '-' }}s</el-descriptions-item
                    >
                    <el-descriptions-item :label="t('workload.period')"
                      >{{ row.livenessProbe.periodSeconds ?? '-' }}s</el-descriptions-item
                    >
                  </el-descriptions>
                </div>
                <div v-if="row.readinessProbe">
                  <h4 style="margin: 0 0 8px; font-size: 13px">
                    {{ t('workload.readinessProbe') }}
                  </h4>
                  <el-descriptions :column="2" border size="small">
                    <el-descriptions-item v-if="row.readinessProbe.httpGet" label="Type"
                      >HTTP GET</el-descriptions-item
                    >
                    <el-descriptions-item v-if="row.readinessProbe.httpGet" label="Path">{{
                      row.readinessProbe.httpGet.path || '/'
                    }}</el-descriptions-item>
                    <el-descriptions-item v-if="row.readinessProbe.httpGet" label="Port">{{
                      row.readinessProbe.httpGet.port
                    }}</el-descriptions-item>
                    <el-descriptions-item :label="t('workload.initialDelay')"
                      >{{ row.readinessProbe.initialDelaySeconds ?? '-' }}s</el-descriptions-item
                    >
                    <el-descriptions-item :label="t('workload.period')"
                      >{{ row.readinessProbe.periodSeconds ?? '-' }}s</el-descriptions-item
                    >
                  </el-descriptions>
                </div>
                <el-empty
                  v-if="
                    (!row.ports || row.ports.length === 0) &&
                    (!row.env || row.env.length === 0) &&
                    (!row.volumeMounts || row.volumeMounts.length === 0) &&
                    !row.livenessProbe &&
                    !row.readinessProbe
                  "
                  :description="t('workload.noAdditionalDetails')"
                  :image-size="60"
                />
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="name" :label="t('common.name')" min-width="150" />
          <el-table-column
            prop="image"
            :label="t('workload.image')"
            min-width="260"
            show-overflow-tooltip
          />
          <el-table-column :label="t('workload.ready')" width="70">
            <template #default="{ row }">
              <el-tag :type="row.ready ? 'success' : 'danger'" size="small">{{
                row.ready ? t('common.yes') : t('common.no')
              }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="restartCount" :label="t('workload.restarts')" width="70" />
          <el-table-column :label="t('common.status')" width="160">
            <template #default="{ row }">
              <el-tag :type="containerStateType(row.state)" size="small">{{
                getContainerStateLabel(row)
              }}</el-tag>
            </template>
          </el-table-column>
        </el-table>
        <el-empty
          v-if="!pod.containers || pod.containers.length === 0"
          :description="t('workload.noContainers')"
        />
      </div>
    </template>

    <!-- 右下：Events -->
    <template v-if="pod" #right-bottom>
      <div class="panel-title">
        {{ t('event.title') }}
        <span class="count-badge">{{ events.length }}</span>
      </div>
      <EventsTable :events="events" :loading="eventsLoading" />
    </template>

    <!-- YAML Drawer -->
    <YamlDrawer
      v-model="yamlDialogVisible"
      :get-yaml="getPodYaml"
      :update-yaml="null"
      :namespace="namespace"
      :name="name"
      title="Pod YAML"
      @saved="handleYamlSaved"
    />
  </DetailPageLayout>
</template>

<style scoped>
.replicas-info {
  font-size: 12px;
  color: var(--el-text-color-regular);
}

.panel-title {
  font-size: var(--gk-font-size-sm);
  font-weight: 600;
  padding: 10px 14px;
  background: var(--el-fill-color-lighter);
  border-bottom: 1px solid var(--gk-color-border-light);
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.count-badge {
  font-weight: 400;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.info-body {
  flex: 1;
  overflow-y: auto;
  padding: 14px;
}

.table-body {
  flex: 1;
  overflow-y: auto;
  padding: 0;
}

/* Annotation list */
.annotation-list {
  max-height: 200px;
  overflow-y: auto;
  border: 1px solid var(--gk-color-border-light);
  border-radius: var(--gk-radius-md);
  padding: 8px;
}

.annotation-item {
  display: flex;
  padding: 4px 6px;
  border-bottom: 1px solid var(--el-border-color-extra-light);
  font-size: 12px;
}

.annotation-item:last-child {
  border-bottom: none;
}

.annotation-key {
  font-weight: 600;
  color: var(--el-text-color-primary);
  min-width: 160px;
  flex-shrink: 0;
  word-break: break-all;
}

.annotation-value {
  color: var(--el-text-color-secondary);
  word-break: break-all;
}
</style>
