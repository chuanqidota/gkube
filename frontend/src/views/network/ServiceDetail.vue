<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getServiceDetail,
  deleteService,
  getServiceEvents,
  getServicePods,
  getServiceEndpoints,
  deletePod,
} from '@/api/resource'
import { FullScreen, Aim } from '@element-plus/icons-vue'
import YamlDrawer from '@/components/YamlDrawer.vue'
import PodListPanel from '@/components/PodListPanel.vue'
import DetailPageLayout from '@/components/DetailPageLayout.vue'
import DetailPageHeader from '@/components/DetailPageHeader.vue'
import EventsTable from '@/components/EventsTable.vue'
import LabelsBlock from '@/components/LabelsBlock.vue'
import ServiceForm from './components/ServiceForm.vue'
import { useClusterStore } from '@/stores/cluster'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import { useEditDrawer } from '@/composables/useEditDrawer'

const clusterStore = useClusterStore()

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const serviceRaw = ref<any>(null)
const yamlDialogVisible = ref(false)

// Events
const events = ref<any[]>([])
const eventsLoading = ref(false)

// Related Pods
const pods = ref<any[]>([])
const podsLoading = ref(false)

// Endpoints
const endpoints = ref<any[]>([])
const endpointsLoading = ref(false)

// Edit dialog
const { editDialogVisible, editFullscreen, handleEdit, handleEditSuccess: onEditSuccess, handleEditCancel } = useEditDrawer(async () => {
  fetchDetail()
  fetchPods()
  fetchEndpoints()
})

// Left panel tab
const leftTab = ref<'info' | 'endpoint'>('info')

const namespace = route.params.namespace as string
const name = route.params.name as string

// Transformed display data
const service = computed(() => {
  const raw = serviceRaw.value
  if (!raw) return null
  const spec = raw.spec || {}
  const meta = raw.metadata || {}
  const status = raw.status || {}

  const portList = (spec.ports || []).map((p: any) => ({
    name: p.name || '',
    port: p.port,
    targetPort: p.targetPort || p.port,
    protocol: p.protocol || 'TCP',
    nodePort: p.nodePort || null,
  }))

  const ports = portList
    .map((p: any) => `${p.port}${p.nodePort ? ':' + p.nodePort : ''}/${p.protocol}`)
    .join(', ')

  let externalIP = ''
  const lbIngress = status.loadBalancer?.ingress
  if (lbIngress && lbIngress.length > 0) {
    externalIP = lbIngress.map((i: any) => i.ip || i.hostname || '').filter(Boolean).join(', ')
  }

  return {
    name: meta.name || '',
    namespace: meta.namespace || '',
    type: spec.type || 'ClusterIP',
    clusterIP: spec.clusterIP || '',
    externalIP,
    ports,
    portList,
    sessionAffinity: spec.sessionAffinity || 'None',
    selector: spec.selector || {},
    labels: meta.labels || {},
  }
})

const showNodePort = computed(() => {
  const t = service.value?.type
  return t === 'NodePort' || t === 'LoadBalancer'
})

const statusTag = computed(() => {
  return service.value?.type === 'LoadBalancer'
    ? { text: service.value.type, type: 'success' as const }
    : { text: service.value?.type || '-', type: 'info' as const }
})

// Endpoint flat list for the table
const endpointRows = computed(() => {
  const rows: any[] = []
  for (const subset of endpoints.value) {
    for (const addr of subset.addresses || []) {
      rows.push({
        ip: addr.ip,
        port: subset.ports?.map((p: any) => p.port).join(', ') || '-',
        protocol: subset.ports?.map((p: any) => p.protocol).join(', ') || 'TCP',
        podName: addr.pod_name || '-',
        nodeName: addr.node_name || '-',
        ready: true,
      })
    }
    for (const addr of subset.not_ready_addresses || []) {
      rows.push({
        ip: addr.ip,
        port: subset.ports?.map((p: any) => p.port).join(', ') || '-',
        protocol: subset.ports?.map((p: any) => p.protocol).join(', ') || 'TCP',
        podName: addr.pod_name || '-',
        nodeName: addr.node_name || '-',
        ready: false,
      })
    }
  }
  return rows
})

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getServiceDetail({ namespace, name })
    serviceRaw.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || '加载 Service 详情失败')
  } finally {
    loading.value = false
  }
}

async function fetchEvents() {
  eventsLoading.value = true
  try {
    const res: any = await getServiceEvents({ namespace, name })
    events.value = res.data || []
  } catch (e) {
    events.value = []
  } finally {
    eventsLoading.value = false
  }
}

async function fetchPods() {
  podsLoading.value = true
  try {
    const res: any = await getServicePods({ namespace, name })
    pods.value = res.data?.items || res.data || []
  } catch (e) {
    pods.value = []
  } finally {
    podsLoading.value = false
  }
}

async function fetchEndpoints() {
  endpointsLoading.value = true
  try {
    const res: any = await getServiceEndpoints({ namespace, name })
    endpoints.value = res.data || []
  } catch (e) {
    endpoints.value = []
  } finally {
    endpointsLoading.value = false
  }
}

function getClusterName(): string {
  return clusterStore.currentCluster?.clusterName || clusterStore.currentCluster?.cluster_name || clusterStore.currentCluster?.name || ''
}

function handlePodLogs(pod: any) {
  const cluster = getClusterName()
  window.open(`/fullscreen/logs?namespace=${pod.metadata?.namespace || namespace}&pod=${pod.metadata?.name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
}

function handlePodExec(pod: any) {
  const cluster = getClusterName()
  window.open(`/fullscreen/terminal?namespace=${pod.metadata?.namespace || namespace}&pod=${pod.metadata?.name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
}

async function handlePodDelete(pod: any, force = false) {
  if (force) {
    try {
      await ElMessageBox.confirm(
        `强制删除 Pod "${pod.metadata?.name}" 将跳过优雅终止，控制器管理的 Pod 会被立即重建。确定继续？`,
        '确认强制删除',
        { type: 'warning', confirmButtonText: '强制删除', cancelButtonText: '取消' }
      )
    } catch {
      return
    }
    try {
      await deletePod({ namespace, name: pod.metadata.name, force: true })
      ElMessage.success('Pod 已强制删除')
      fetchPods()
    } catch (e: any) {
      if (e !== 'cancel') ElMessage.error(e?.message || '强制删除失败')
    }
    return
  }
  try {
    await ElMessageBox.confirm(
      `确定要删除 Pod "${pod.metadata?.name}" 吗？`,
      '确认删除',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await deletePod({ namespace, name: pod.metadata.name })
    ElMessage.success('Pod 已删除')
    fetchPods()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || '删除失败')
    }
  }
}

function handleOpenYaml() {
  yamlDialogVisible.value = true
}

function handleYamlSaved() {
  fetchDetail()
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(
      `确定要删除 Service "${name}" 吗？此操作不可恢复。`,
      '确认删除',
      { type: 'error', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await deleteService({ namespace, name })
    ElMessage.success('Service 已删除')
    router.push('/network/services')
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || '删除失败')
    }
  }
}

function handleEditSuccess() {
  onEditSuccess()
  fetchPods()
  fetchEndpoints()
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(async () => {
  fetchDetail()
  fetchPods()
  fetchEvents()
  fetchEndpoints()
}, { autoStart: false })

onMounted(() => {
  fetchDetail()
  fetchPods()
  fetchEvents()
  fetchEndpoints()
})
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
      @back="router.push('/network/services')"
    >
      <template #meta>
        <span class="replicas-info" v-if="service?.clusterIP">
          Cluster IP: {{ service.clusterIP }}
        </span>
      </template>
      <template #actions>
        <el-button type="info" @click="handleEdit">编辑</el-button>
        <el-button @click="handleOpenYaml">YAML</el-button>
        <el-button type="danger" @click="handleDelete">删除</el-button>
      </template>
    </DetailPageHeader>

    <!-- Left panel -->
    <template v-if="service" #left>
      <div class="left-tabs">
        <el-segmented
          v-model="leftTab"
          :options="[
            { label: '基本信息', value: 'info' },
            { label: 'Endpoint', value: 'endpoint' },
          ]"
          size="small"
        />
      </div>

      <!-- Info view -->
      <div v-show="leftTab === 'info'" class="left-content">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="名称">{{ service.name }}</el-descriptions-item>
          <el-descriptions-item label="命名空间">{{ service.namespace }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ service.type || '-' }}</el-descriptions-item>
          <el-descriptions-item label="Cluster IP">{{ service.clusterIP || '-' }}</el-descriptions-item>
          <el-descriptions-item label="External IP">{{ service.externalIP || '-' }}</el-descriptions-item>
          <el-descriptions-item label="Session Affinity">{{ service.sessionAffinity || '-' }}</el-descriptions-item>
        </el-descriptions>

        <!-- Port mapping table -->
        <div v-if="service.portList && service.portList.length > 0" style="margin-top: var(--gk-space-4);">
          <h4 style="margin: 0 0 8px; font-size: var(--gk-font-size-sm);">端口映射</h4>
          <el-table :data="service.portList" size="small" border stripe>
            <el-table-column prop="name" label="名称" width="80">
              <template #default="{ row }">{{ row.name || '-' }}</template>
            </el-table-column>
            <el-table-column prop="port" label="Port" width="70" align="center" />
            <el-table-column label="→" width="30" align="center">
              <template #default><span style="color: var(--gk-color-text-placeholder);">→</span></template>
            </el-table-column>
            <el-table-column prop="targetPort" label="TargetPort" width="90" align="center" />
            <el-table-column prop="protocol" label="协议" width="70" align="center" />
            <el-table-column v-if="showNodePort" prop="nodePort" label="NodePort" width="90" align="center">
              <template #default="{ row }">
                <el-tag v-if="row.nodePort" size="small" type="warning">{{ row.nodePort }}</el-tag>
                <span v-else>-</span>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <!-- Selector -->
        <div v-if="service.selector && Object.keys(service.selector).length > 0" style="margin-top: var(--gk-space-4);">
          <h4 style="margin: 0 0 8px; font-size: var(--gk-font-size-sm);">Selector</h4>
          <LabelsBlock :labels="service.selector" />
        </div>

        <!-- Labels -->
        <div v-if="service.labels && Object.keys(service.labels).length > 0" style="margin-top: var(--gk-space-4);">
          <h4 style="margin: 0 0 8px; font-size: var(--gk-font-size-sm);">Labels</h4>
          <LabelsBlock :labels="service.labels" />
        </div>
      </div>

      <!-- Endpoint view -->
      <div v-show="leftTab === 'endpoint'" class="left-content endpoint-tab">
        <div v-loading="endpointsLoading" class="endpoint-table-wrapper">
          <el-table v-if="endpointRows.length > 0" :data="endpointRows" size="small" stripe>
            <el-table-column prop="ip" label="IP" width="130" />
            <el-table-column prop="port" label="Port" width="70" align="center" />
            <el-table-column prop="protocol" label="协议" width="65" align="center" />
            <el-table-column prop="podName" label="Pod" min-width="140" show-overflow-tooltip />
            <el-table-column prop="nodeName" label="Node" min-width="100" show-overflow-tooltip />
            <el-table-column label="状态" width="75" align="center">
              <template #default="{ row }">
                <el-tag :type="row.ready ? 'success' : 'warning'" size="small">
                  {{ row.ready ? 'Ready' : 'NotReady' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
          <div v-else-if="!endpointsLoading" class="empty-hint">暂无 Endpoint</div>
        </div>
      </div>
    </template>

    <!-- Right-top: Pod list -->
    <template v-if="service" #right-top>
      <div class="panel-title">
        关联 Pod
        <span class="count-badge">{{ pods.length }} 个</span>
      </div>
      <PodListPanel
        :pods="pods"
        :loading="podsLoading"
        @logs="handlePodLogs"
        @exec="handlePodExec"
        @delete="handlePodDelete"
      />
    </template>

    <!-- Right-bottom: Events -->
    <template v-if="service" #right-bottom>
      <div class="panel-title">
        事件
        <span class="count-badge">{{ events.length }} 条</span>
      </div>
      <EventsTable :events="events" :loading="eventsLoading" time-field="last_seen" />
    </template>

    <!-- YAML Drawer -->
    <YamlDrawer
      v-model="yamlDialogVisible"
      resource-type="service"
      :namespace="namespace"
      :name="name"
      @saved="handleYamlSaved"
    />

    <!-- Edit Drawer -->
    <el-drawer
      v-model="editDialogVisible"
      title="编辑 Service"
      :size="editFullscreen ? '100%' : '85%'"
      direction="rtl"
      :destroy-on-close="true"
      :body-style="{ padding: '0', height: '100%' }"
    >
      <template #header>
        <div class="drawer-header">
          <span class="drawer-title">编辑 Service</span>
          <el-tooltip :content="editFullscreen ? '退出全屏' : '全屏'" placement="top">
            <el-icon class="fullscreen-btn" @click="editFullscreen = !editFullscreen">
              <FullScreen v-if="!editFullscreen" />
              <Aim v-else />
            </el-icon>
          </el-tooltip>
        </div>
      </template>
      <div style="height: calc(100dvh - 52px); overflow-y: auto;">
        <ServiceForm
          v-if="editDialogVisible && serviceRaw"
          :is-edit="true"
          :initial-data="serviceRaw"
          @success="handleEditSuccess"
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

.left-tabs {
  padding: 8px 14px;
  border-bottom: 1px solid var(--gk-color-border-light);
  flex-shrink: 0;
}

.left-tabs :deep(.el-segmented) {
  width: 100%;
}

.left-content {
  flex: 1;
  overflow-y: auto;
  padding: 14px;
  min-height: 0;
}

.endpoint-tab {
  padding: 0;
  display: flex;
  flex-direction: column;
}

.endpoint-table-wrapper {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
  padding: 14px;
}

.empty-hint {
  padding: 24px;
  text-align: center;
  color: var(--el-text-color-secondary);
  font-size: var(--el-text-color-secondary);
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
