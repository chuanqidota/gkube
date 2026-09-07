<script setup lang="ts">
import { computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown } from '@element-plus/icons-vue'
import { getPodDetail, deletePod, getPodEvents, calcAge } from '@/api/resource'
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

const {
  namespace, name,
  loading, detail: pod, events, eventsLoading, yamlDialogVisible,
  isRunning, countdown, currentInterval, availableIntervals,
  toggle, manualRefresh, setIntervalOption,
  fetchDetail, handleOpenYaml,
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
    (sum: number, cs: any) => sum + (cs.restartCount || 0), 0
  )
  const specContainers = raw.spec?.containers || []
  const statusContainers = raw.status?.containerStatuses || []
  const containers = specContainers.map((spec: any) => {
    const status = statusContainers.find((s: any) => s.name === spec.name) || {}
    let state = 'Unknown'
    let stateReason = ''
    let exitCode: number | undefined
    if (status.state?.running) state = 'Running'
    else if (status.state?.waiting) { state = 'Waiting'; stateReason = status.state.waiting.reason || '' }
    else if (status.state?.terminated) { state = 'Terminated'; exitCode = status.state.terminated.exitCode }
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
    age: calcAge(raw.metadata?.creationTimestamp),
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
  if (container.state === 'Running') return '运行中'
  if (container.state === 'Waiting') return `等待中 (${container.stateReason || '-'})`
  if (container.state === 'Terminated') return `已终止 (${container.exitCode ?? '-'})`
  return container.state || '-'
}

function handleYamlSaved() {
  fetchDetail()
}

function handleLogs() {
  window.open(buildFullscreenUrl('logs', { namespace, pod: name, cluster: clusterStore.clusterName || undefined }), '_blank')
}

function handleExec() {
  window.open(buildFullscreenUrl('terminal', { namespace, pod: name, cluster: clusterStore.clusterName || undefined }), '_blank')
}

async function handleDelete(force = false) {
  if (force) {
    try {
      await ElMessageBox.confirm(
        `强制删除 Pod "${name}" 将跳过优雅终止，控制器管理的 Pod 会被立即重建。确定继续？`,
        '确认强制删除',
        { type: 'warning', confirmButtonText: '强制删除', cancelButtonText: '取消' }
      )
    } catch {
      return
    }
    try {
      await deletePod({ namespace, name, force: true })
      ElMessage.success('Pod 已强制删除')
      router.push('/workloads/pods')
    } catch (e: any) {
      if (e !== 'cancel') ElMessage.error(e?.message || '强制删除失败')
    }
    return
  }
  try {
    await ElMessageBox.confirm(
      `确定要删除 Pod "${name}"（命名空间：${namespace}）吗？`,
      '确认删除',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await deletePod({ namespace, name })
    ElMessage.success('Pod 已删除')
    router.push('/workloads/pods')
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || '删除失败')
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
        <span class="replicas-info" v-if="pod">
          {{ pod.ip || '-' }}
        </span>
      </template>
      <template #actions>
        <el-button type="primary" @click="handleLogs">日志</el-button>
        <el-button type="success" @click="handleExec">终端</el-button>
        <el-button @click="handleOpenYaml">YAML</el-button>
        <el-dropdown @command="(cmd: string) => handleDelete(cmd === 'force')" trigger="click">
          <el-button type="danger">
            删除 <el-icon><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="normal">删除</el-dropdown-item>
              <el-dropdown-item command="force" divided>强制删除</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </template>
    </DetailPageHeader>

    <!-- 左侧：基本信息 -->
    <template v-if="pod" #left>
      <div class="panel-title">基本信息</div>
      <div class="info-body">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="名称">{{ pod.name }}</el-descriptions-item>
          <el-descriptions-item label="命名空间">{{ pod.namespace }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getPodStatusType(pod.status)" size="small">{{ pod.status }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="Pod IP">{{ pod.ip || '-' }}</el-descriptions-item>
          <el-descriptions-item label="主机 IP">{{ pod.host_ip || '-' }}</el-descriptions-item>
          <el-descriptions-item label="节点">{{ pod.node || '-' }}</el-descriptions-item>
          <el-descriptions-item label="QoS 类别">{{ pod.qos_class || '-' }}</el-descriptions-item>
          <el-descriptions-item label="优先级">{{ pod.priority ?? '-' }}</el-descriptions-item>
          <el-descriptions-item label="服务账号">{{ pod.service_account || '-' }}</el-descriptions-item>
          <el-descriptions-item label="重启次数">{{ pod.restarts ?? '-' }}</el-descriptions-item>
          <el-descriptions-item label="年龄">{{ pod.age || '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ pod.created_at || '-' }}</el-descriptions-item>
        </el-descriptions>

        <!-- Labels -->
        <div v-if="pod.labels && Object.keys(pod.labels).length > 0" style="margin-top: var(--gk-space-4);">
          <h4 style="margin: 0 0 8px; font-size: 13px;">Labels</h4>
          <LabelsBlock :labels="pod.labels" />
        </div>

        <!-- Annotations -->
        <div v-if="pod.annotations && Object.keys(pod.annotations).length > 0" style="margin-top: var(--gk-space-4);">
          <h4 style="margin: 0 0 8px; font-size: 13px;">Annotations</h4>
          <div class="annotation-list">
            <div v-for="(val, key) in pod.annotations" :key="key" class="annotation-item">
              <span class="annotation-key">{{ key }}</span>
              <span class="annotation-value">{{ val }}</span>
            </div>
          </div>
        </div>

        <!-- Pod Conditions -->
        <div v-if="pod.conditions && pod.conditions.length > 0" style="margin-top: var(--gk-space-4);">
          <h4 style="margin: 0 0 8px; font-size: 13px;">Pod 条件</h4>
          <ConditionsBlock :conditions="pod.conditions" />
        </div>
      </div>
    </template>

    <!-- 右上：容器列表 -->
    <template v-if="pod" #right-top>
      <div class="panel-title">
        容器
        <span class="count-badge">{{ pod.containers?.length || 0 }} 个</span>
      </div>
      <div class="table-body">
        <el-table :data="pod.containers || []" size="small" stripe>
          <el-table-column type="expand">
            <template #default="{ row }">
              <div style="padding: 12px 16px;">
                <div v-if="row.ports && row.ports.length > 0" style="margin-bottom: var(--gk-space-4);">
                  <h4 style="margin: 0 0 8px; font-size: 13px;">端口</h4>
                  <el-table :data="row.ports" border size="small">
                    <el-table-column prop="name" label="名称" width="120" />
                    <el-table-column prop="containerPort" label="容器端口" width="130" />
                    <el-table-column prop="protocol" label="协议" width="100" />
                  </el-table>
                </div>
                <div v-if="row.env && row.env.length > 0" style="margin-bottom: var(--gk-space-4);">
                  <h4 style="margin: 0 0 8px; font-size: 13px;">环境变量</h4>
                  <el-table :data="row.env" border size="small">
                    <el-table-column prop="name" label="名称" min-width="180" />
                    <el-table-column label="值" min-width="250">
                      <template #default="{ row: envRow }">
                        <span v-if="envRow.value !== undefined && envRow.value !== ''">{{ envRow.value }}</span>
                        <span v-else-if="envRow.valueFrom" style="color: var(--el-text-color-secondary);">{{ envRow.valueFrom.fieldRef?.fieldPath || envRow.valueFrom.secretKeyRef?.name || envRow.valueFrom.configMapKeyRef?.name || '来自引用' }}</span>
                        <span v-else style="color: var(--el-text-color-secondary);">-</span>
                      </template>
                    </el-table-column>
                  </el-table>
                </div>
                <div v-if="row.volumeMounts && row.volumeMounts.length > 0" style="margin-bottom: var(--gk-space-4);">
                  <h4 style="margin: 0 0 8px; font-size: 13px;">卷挂载</h4>
                  <el-table :data="row.volumeMounts" border size="small">
                    <el-table-column prop="name" label="卷名称" min-width="150" />
                    <el-table-column prop="mountPath" label="挂载路径" min-width="200" />
                    <el-table-column prop="subPath" label="子路径" width="150" />
                    <el-table-column label="只读" width="80">
                      <template #default="{ row: vm }">
                        <el-tag :type="vm.readOnly ? 'warning' : 'success'" size="small">{{ vm.readOnly ? '是' : '否' }}</el-tag>
                      </template>
                    </el-table-column>
                  </el-table>
                </div>
                <div v-if="row.livenessProbe" style="margin-bottom: var(--gk-space-4);">
                  <h4 style="margin: 0 0 8px; font-size: 13px;">存活探针</h4>
                  <el-descriptions :column="2" border size="small">
                    <el-descriptions-item v-if="row.livenessProbe.httpGet" label="类型">HTTP GET</el-descriptions-item>
                    <el-descriptions-item v-if="row.livenessProbe.httpGet" label="路径">{{ row.livenessProbe.httpGet.path || '/' }}</el-descriptions-item>
                    <el-descriptions-item v-if="row.livenessProbe.httpGet" label="端口">{{ row.livenessProbe.httpGet.port }}</el-descriptions-item>
                    <el-descriptions-item v-if="row.livenessProbe.tcpSocket" label="类型">TCP Socket</el-descriptions-item>
                    <el-descriptions-item v-if="row.livenessProbe.exec" label="类型">Exec</el-descriptions-item>
                    <el-descriptions-item v-if="row.livenessProbe.exec" label="命令">{{ (row.livenessProbe.exec.command || []).join(' ') }}</el-descriptions-item>
                    <el-descriptions-item label="初始延迟">{{ row.livenessProbe.initialDelaySeconds ?? '-' }}s</el-descriptions-item>
                    <el-descriptions-item label="检查周期">{{ row.livenessProbe.periodSeconds ?? '-' }}s</el-descriptions-item>
                  </el-descriptions>
                </div>
                <div v-if="row.readinessProbe">
                  <h4 style="margin: 0 0 8px; font-size: 13px;">就绪探针</h4>
                  <el-descriptions :column="2" border size="small">
                    <el-descriptions-item v-if="row.readinessProbe.httpGet" label="类型">HTTP GET</el-descriptions-item>
                    <el-descriptions-item v-if="row.readinessProbe.httpGet" label="路径">{{ row.readinessProbe.httpGet.path || '/' }}</el-descriptions-item>
                    <el-descriptions-item v-if="row.readinessProbe.httpGet" label="端口">{{ row.readinessProbe.httpGet.port }}</el-descriptions-item>
                    <el-descriptions-item label="初始延迟">{{ row.readinessProbe.initialDelaySeconds ?? '-' }}s</el-descriptions-item>
                    <el-descriptions-item label="检查周期">{{ row.readinessProbe.periodSeconds ?? '-' }}s</el-descriptions-item>
                  </el-descriptions>
                </div>
                <el-empty
                  v-if="(!row.ports || row.ports.length === 0) && (!row.env || row.env.length === 0) && (!row.volumeMounts || row.volumeMounts.length === 0) && !row.livenessProbe && !row.readinessProbe"
                  description="无额外容器详情"
                  :image-size="60"
                />
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="name" label="名称" min-width="150" />
          <el-table-column prop="image" label="镜像" min-width="260" show-overflow-tooltip />
          <el-table-column label="就绪" width="70">
            <template #default="{ row }">
              <el-tag :type="row.ready ? 'success' : 'danger'" size="small">{{ row.ready ? '是' : '否' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="restartCount" label="重启" width="70" />
          <el-table-column label="状态" width="160">
            <template #default="{ row }">
              <el-tag :type="containerStateType(row.state)" size="small">{{ getContainerStateLabel(row) }}</el-tag>
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-if="!pod.containers || pod.containers.length === 0" description="无容器" />
      </div>
    </template>

    <!-- 右下：Events -->
    <template v-if="pod" #right-bottom>
      <div class="panel-title">
        事件
        <span class="count-badge">{{ events.length }} 条</span>
      </div>
      <EventsTable :events="events" :loading="eventsLoading" />
    </template>

    <!-- YAML Drawer -->
    <YamlDrawer
      v-model="yamlDialogVisible"
      resource-type="pod"
      :namespace="namespace"
      :name="name"
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
