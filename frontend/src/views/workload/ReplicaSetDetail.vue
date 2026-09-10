<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getReplicaSetDetail,
  getReplicaSetYaml,
  getReplicaSetEvents,
  deleteReplicaSet,
} from '@/api/resource'
import YamlDrawer from '@/components/YamlDrawer.vue'
import PodListPanel from '@/components/PodListPanel.vue'
import DetailPageLayout from '@/components/DetailPageLayout.vue'
import DetailPageHeader from '@/components/DetailPageHeader.vue'
import EventsTable from '@/components/EventsTable.vue'
import SelectorBlock from '@/components/SelectorBlock.vue'
import { usePodActions } from '@/composables/usePodActions'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import { useClusterStore } from '@/stores/cluster'
import { formatAge } from '@/utils/helpers'

const clusterStore = useClusterStore()
const clusterName = computed(() => clusterStore.clusterName)
const route = useRoute()
const router = useRouter()
const loading = ref(false)
const detail = ref<any>(null)
const yamlDialogVisible = ref(false)
const events = ref<any[]>([])
const eventsLoading = ref(false)

const namespace = route.params.namespace as string
const name = route.params.name as string

const rs = computed(() => detail.value?.rs)

const statusTag = computed(() => {
  const conditions = rs.value?.status?.conditions || []
  const available = conditions.find((c: any) => c.type === 'Available')
  if (available?.status === 'True') return { type: 'success' as const, text: 'Available' }
  const failure = conditions.find((c: any) => c.type === 'ReplicaFailure')
  if (failure?.status === 'True') return { type: 'danger' as const, text: 'Failure' }
  return { type: 'warning' as const, text: 'Progressing' }
})

const controllerOf = computed(() => detail.value?.controllerOf || null)

const containers = computed(() => rs.value?.spec?.template?.spec?.containers || [])

const selectorLabels = computed(() => rs.value?.spec?.selector?.matchLabels || {})

// ---- Shared composables ----
const { handlePodLogs, handlePodExec, handlePodDelete } = usePodActions(clusterName)

function onPodLogs(pod: any) {
  handlePodLogs({ namespace: pod.metadata.namespace || namespace, name: pod.metadata.name })
}

function onPodExec(pod: any) {
  handlePodExec({ namespace: pod.metadata.namespace || namespace, name: pod.metadata.name })
}

function onPodDelete(pod: any, force?: boolean) {
  handlePodDelete(
    { namespace: pod.metadata.namespace || namespace, name: pod.metadata.name },
    () => fetchDetail(),
    force
  )
}

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getReplicaSetDetail({ namespace, name })
    detail.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || '获取详情失败')
  } finally {
    loading.value = false
  }
}

async function fetchEvents() {
  eventsLoading.value = true
  try {
    const res: any = await getReplicaSetEvents({ namespace, name })
    events.value = res.data || []
  } catch (e) {
    console.error('Failed to fetch events:', e)
    ElMessage.error('获取事件失败')
  } finally {
    eventsLoading.value = false
  }
}

function handleOpenYaml() {
  yamlDialogVisible.value = true
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(
      `确定要删除 ReplicaSet "${name}" 吗？此操作不可恢复。`,
      '确认删除',
      { type: 'error', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await deleteReplicaSet({ namespace, name })
    ElMessage.success('ReplicaSet 已删除')
    router.push('/workloads/replicasets')
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || '删除失败')
    }
  }
}

function goController() {
  const c = controllerOf.value
  if (!c) return
  if (c.kind === 'Deployment') {
    router.push(`/workloads/deployments/${c.namespace || namespace}/${c.name}`)
  }
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(async () => {
  fetchDetail()
  fetchEvents()
}, { autoStart: false })

onMounted(() => {
  Promise.all([fetchDetail(), fetchEvents()])
})
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
      @back="router.push('/workloads/replicasets')"
    >
      <template #meta>
        <span class="replicas-info" v-if="rs">
          {{ rs.status?.readyReplicas ?? 0 }}/{{ rs.spec?.replicas ?? 0 }} ready
        </span>
        <el-tag
          v-if="controllerOf"
          type="primary"
          size="small"
          :class="{ clickable: controllerOf.kind === 'Deployment' }"
          @click="goController"
        >
          由 {{ controllerOf.kind }}/{{ controllerOf.name }} 管理
        </el-tag>
      </template>
      <template #actions>
        <el-button @click="handleOpenYaml">YAML</el-button>
        <el-button type="danger" @click="handleDelete">删除</el-button>
      </template>
    </DetailPageHeader>

    <!-- 左侧：基本信息 + 容器模板 + 选择器 -->
    <template v-if="rs" #left>
      <div class="left-scroll">
        <div class="info-block">
          <div class="block-title">基本信息</div>
          <div class="info-row"><span class="info-label">名称</span><span class="info-value mono">{{ rs.metadata?.name }}</span></div>
          <div class="info-row"><span class="info-label">命名空间</span><span class="info-value">{{ rs.metadata?.namespace }}</span></div>
          <div class="info-row"><span class="info-label">期望副本</span><span class="info-value">{{ rs.spec?.replicas ?? 0 }}</span></div>
          <div class="info-row"><span class="info-label">当前副本</span><span class="info-value">{{ rs.status?.replicas ?? 0 }}</span></div>
          <div class="info-row"><span class="info-label">就绪副本</span><span class="info-value">{{ rs.status?.readyReplicas ?? 0 }}</span></div>
          <div class="info-row"><span class="info-label">可用副本</span><span class="info-value">{{ rs.status?.availableReplicas ?? 0 }}</span></div>
          <div class="info-row"><span class="info-label">创建时间</span><span class="info-value">{{ formatAge(rs.metadata?.creationTimestamp) }}</span></div>
          <div class="info-row" v-if="controllerOf">
            <span class="info-label">拥有者</span>
            <span
              class="info-value link"
              :class="{ disabled: controllerOf.kind !== 'Deployment' }"
              @click="goController"
            >{{ controllerOf.kind }}/{{ controllerOf.name }}</span>
          </div>
        </div>

        <div class="info-block">
          <div class="block-title">容器模板</div>
          <div v-if="containers.length === 0" class="empty-hint">暂无容器</div>
          <div v-for="c in containers" :key="c.name" class="container-item">
            <div class="container-name mono">{{ c.name }}</div>
            <div class="container-image mono">{{ c.image || '-' }}</div>
          </div>
        </div>

        <div class="info-block">
          <div class="block-title">选择器</div>
          <SelectorBlock :selector="selectorLabels" />
        </div>
      </div>
    </template>

    <!-- 右上：Pod 列表 -->
    <template v-if="rs" #right-top>
      <div class="panel-title">
        Pod 列表
        <span class="count-badge">{{ (detail?.pods || []).length }} 个</span>
      </div>
      <PodListPanel
        :pods="detail?.pods || []"
        :loading="loading"
        @logs="onPodLogs"
        @exec="onPodExec"
        @delete="onPodDelete"
      />
    </template>

    <!-- 右下：Events -->
    <template v-if="rs" #right-bottom>
      <div class="panel-title">
        事件
        <span class="count-badge">{{ events.length }} 条</span>
      </div>
      <EventsTable :events="events" :loading="eventsLoading" time-field="last_seen" />
    </template>

    <!-- ===== YAML Drawer (只读) ===== -->
    <YamlDrawer
      v-model="yamlDialogVisible"
      :get-yaml="getReplicaSetYaml"
      :update-yaml="null"
      :namespace="namespace"
      :name="name"
      title="ReplicaSet YAML"
    />
  </DetailPageLayout>
</template>

<style scoped>
.replicas-info {
  font-size: 12px;
  color: var(--gk-color-text-primary);
}

.clickable {
  cursor: pointer;
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

.left-scroll {
  flex: 1;
  overflow-y: auto;
}

.info-block {
  padding: var(--gk-space-2) var(--gk-space-4);
  border-bottom: 1px solid var(--el-border-color-extra-light);
}

.info-block:last-child {
  border-bottom: none;
}

.block-title {
  font-size: var(--gk-font-size-sm);
  font-weight: 600;
  margin-bottom: 8px;
  color: var(--gk-color-text-primary);
}

.info-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  font-size: 12px;
  line-height: 1.8;
}

.info-label {
  color: var(--gk-color-text-secondary);
  flex-shrink: 0;
  width: 64px;
}

.info-value {
  color: var(--gk-color-text-primary);
  word-break: break-all;
}

.info-value.link {
  color: var(--el-color-primary);
  cursor: pointer;
}

.info-value.link.disabled {
  cursor: default;
  color: var(--gk-color-text-primary);
}

.container-item {
  padding: 6px 0;
  border-top: 1px dashed var(--el-border-color-extra-light);
}

.container-item:first-of-type {
  border-top: none;
  padding-top: 0;
}

.container-name {
  font-size: 12px;
  font-weight: 500;
  color: var(--gk-color-text-primary);
  margin-bottom: 2px;
}

.container-image {
  font-size: 11px;
  color: var(--gk-color-text-secondary);
  word-break: break-all;
}

.empty-hint {
  padding: 16px 4px;
  text-align: center;
  color: var(--gk-color-text-secondary);
  font-size: 12px;
}

.mono {
  font-family: var(--gk-font-mono);
}
</style>
