<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getDaemonSetDetail,
  getDaemonSetYaml,
  updateDaemonSetYaml,
  deleteDaemonSet,
  restartDaemonSet,
  getDaemonSetEvents,
  getDaemonSetPods,
  deletePod,
} from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'
import PodListPanel from '@/components/PodListPanel.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const daemonSet = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const yamlDialogVisible = ref(false)
const yamlEditorRef = ref<InstanceType<typeof YamlEditor>>()
const events = ref<any[]>([])
const eventsLoading = ref(false)
const pods = ref<any[]>([])
const podsLoading = ref(false)

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
  if (ready === desired && desired > 0) return '就绪'
  if (ready > 0) return '部分就绪'
  return '未就绪'
})

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getDaemonSetDetail({ namespace, name })
    daemonSet.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || '加载 DaemonSet 详情失败')
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
    ElMessage.error(e?.message || '加载 YAML 失败')
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
    ElMessage.error(e?.message || '加载事件失败')
  } finally {
    eventsLoading.value = false
  }
}

async function fetchPods() {
  podsLoading.value = true
  try {
    const res: any = await getDaemonSetPods({ namespace, name })
    // res.data is already unwrapped by the response interceptor
    // If it's a PodList object, it has an 'items' field
    // If it's directly an array, use it as-is
    const data = res.data
    if (Array.isArray(data)) {
      pods.value = data
    } else if (data?.items && Array.isArray(data.items)) {
      pods.value = data.items
    } else {
      pods.value = []
    }
    console.log('[DaemonSet] pods loaded:', pods.value.length, 'namespace:', namespace, 'name:', name)
  } catch (e: any) {
    pods.value = []
    console.error('[DaemonSet] fetchPods error:', e)
    ElMessage.error(e?.message || '加载 Pod 列表失败')
  } finally {
    podsLoading.value = false
  }
}

function handleOpenYaml() {
  fetchYaml()
  yamlDialogVisible.value = true
}

async function handleSaveYaml(content: string) {
  try {
    await updateDaemonSetYaml({ namespace, name, yaml: content })
    ElMessage.success('YAML 保存成功')
    yamlDialogVisible.value = false
    fetchDetail()
  } catch (e: any) {
    ElMessage.error(e?.message || '保存 YAML 失败')
    yamlEditorRef.value?.resetSaving()
  }
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(
      `确定删除命名空间 "${namespace}" 中的 DaemonSet "${name}" 吗？`,
      '确认删除',
      { type: 'warning' }
    )
    await deleteDaemonSet({ namespace, name })
    ElMessage.success('DaemonSet 已删除')
    router.push('/workloads/daemonsets')
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || '删除失败')
    }
  }
}

async function handleRestart() {
  try {
    await ElMessageBox.confirm(
      `确定重启 DaemonSet "${name}" 吗？这将触发滚动更新。`,
      '确认重启',
      { type: 'warning' }
    )
    await restartDaemonSet({ namespace, name })
    ElMessage.success('DaemonSet 已重启')
    fetchDetail()
    fetchPods()
  } catch {
    // cancelled
  }
}

function getClusterName(): string {
  try {
    const saved = localStorage.getItem('gkube_cluster')
    if (saved) {
      const c = JSON.parse(saved)
      return c?.clusterName || c?.cluster_name || c?.name || ''
    }
  } catch { /* ignore */ }
  return ''
}

function handlePodLogs(pod: any) {
  const cluster = getClusterName()
  window.open(`/fullscreen/logs?namespace=${pod.metadata.namespace || namespace}&pod=${pod.metadata.name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
}

function handlePodExec(pod: any) {
  const cluster = getClusterName()
  window.open(`/fullscreen/terminal?namespace=${pod.metadata.namespace || namespace}&pod=${pod.metadata.name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
}

async function handlePodDelete(pod: any) {
  try {
    await ElMessageBox.confirm(
      `确定删除 Pod ${pod.metadata.name} 吗？`,
      '确认删除',
      { type: 'warning' }
    )
    await deletePod({ namespace, name: pod.metadata.name })
    ElMessage.success('Pod 已删除')
    fetchPods()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(async () => {
  fetchDetail()
  fetchPods()
  fetchEvents()
}, { autoStart: false })

onMounted(() => {
  fetchDetail()
  fetchPods()
  fetchEvents()
})
</script>

<template>
  <div class="detail-page" v-loading="loading">

    <!-- ===== 顶部标题栏 ===== -->
    <div class="page-header">
      <div class="header-left">
        <div class="title-line">
          <h2 class="res-name">{{ name }}</h2>
          <el-tag :type="statusTagType" effect="dark" size="small">{{ statusText }}</el-tag>
          <span class="ns-tag">ns/{{ namespace }}</span>
          <span class="replicas-info" v-if="daemonSet">
            {{ daemonSet.status?.numberReady ?? 0 }}/{{ daemonSet.status?.desiredNumberScheduled ?? 0 }} 就绪
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
        <el-button type="warning" @click="handleRestart">重启</el-button>
        <el-button @click="handleOpenYaml">YAML</el-button>
        <el-button type="danger" @click="handleDelete">删除</el-button>
        <el-button @click="router.push('/workloads/daemonsets')">返回列表</el-button>
      </div>
    </div>

    <template v-if="daemonSet">
      <!-- ===== 主体：Pods + Events ===== -->
      <div class="main-layout">

        <!-- 右侧：Pods + Events -->
        <div class="right-panel">

          <!-- Pod 列表 -->
          <div class="right-section">
            <div class="panel-title">
              Pod 列表
              <span class="count-badge">{{ pods.length }} 个</span>
            </div>
            <PodListPanel
              :pods="pods"
              :loading="podsLoading"
              @logs="handlePodLogs"
              @exec="handlePodExec"
              @delete="handlePodDelete"
            />
          </div>

          <!-- Info -->
          <div class="right-section">
            <div class="panel-title">基本信息</div>
            <div class="info-body">
              <el-descriptions :column="2" border size="small">
                <el-descriptions-item label="名称">{{ daemonSet.metadata?.name }}</el-descriptions-item>
                <el-descriptions-item label="命名空间">{{ daemonSet.metadata?.namespace }}</el-descriptions-item>
                <el-descriptions-item label="预期调度">{{ daemonSet.status?.desiredNumberScheduled ?? '-' }}</el-descriptions-item>
                <el-descriptions-item label="当前调度">{{ daemonSet.status?.currentNumberScheduled ?? '-' }}</el-descriptions-item>
                <el-descriptions-item label="就绪数量">{{ daemonSet.status?.numberReady ?? '-' }}</el-descriptions-item>
                <el-descriptions-item label="可用数量">{{ daemonSet.status?.numberAvailable ?? '-' }}</el-descriptions-item>
                <el-descriptions-item label="更新策略">{{ daemonSet.spec?.updateStrategy?.type || 'RollingUpdate' }}</el-descriptions-item>
                <el-descriptions-item label="创建时间">{{ daemonSet.metadata?.creationTimestamp || '-' }}</el-descriptions-item>
              </el-descriptions>

              <!-- Labels -->
              <div v-if="daemonSet.metadata?.labels && Object.keys(daemonSet.metadata.labels).length > 0" style="margin-top: 12px;">
                <span class="sub-title">标签</span>
                <div style="margin-top: 6px;">
                  <el-tag
                    v-for="(val, key) in daemonSet.metadata.labels"
                    :key="key"
                    style="margin-right: 8px; margin-bottom: 8px;"
                  >
                    {{ key }}={{ val }}
                  </el-tag>
                </div>
              </div>

              <!-- Selector -->
              <div v-if="daemonSet.spec?.selector?.matchLabels && Object.keys(daemonSet.spec.selector.matchLabels).length > 0" style="margin-top: 12px;">
                <span class="sub-title">选择器</span>
                <div style="margin-top: 6px;">
                  <el-tag
                    v-for="(val, key) in daemonSet.spec.selector.matchLabels"
                    :key="key"
                    style="margin-right: 8px; margin-bottom: 8px;"
                    type="info"
                  >
                    {{ key }}={{ val }}
                  </el-tag>
                </div>
              </div>
            </div>
          </div>

          <!-- Events -->
          <div class="right-section events-section">
            <div class="panel-title">
              事件
              <span class="count-badge">{{ events.length }} 条</span>
            </div>
            <div v-loading="eventsLoading" class="events-body">
              <el-table v-if="events.length > 0" :data="events" size="small" stripe max-height="260">
                <el-table-column prop="type" label="类型" width="80">
                  <template #default="{ row }">
                    <el-tag :type="row.type === 'Warning' ? 'danger' : 'info'" size="small">{{ row.type }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="reason" label="原因" width="130" />
                <el-table-column prop="message" label="信息" min-width="200" show-overflow-tooltip />
                <el-table-column prop="last_seen" label="最后发生" width="150" />
              </el-table>
              <div v-else class="empty-hint">暂无事件</div>
            </div>
          </div>

        </div>
      </div>
    </template>

    <!-- ===== Dialogs ===== -->
    <el-dialog v-model="yamlDialogVisible" title="YAML" width="70%" top="5vh" destroy-on-close>
      <div v-loading="yamlLoading">
        <YamlEditor ref="yamlEditorRef" v-model="yamlContent" height="600px" :read-only="false" :saveable="true" @save="handleSaveYaml" />
      </div>
    </el-dialog>
  </div>
</template>

<style scoped>
.detail-page {
  padding: 16px 20px;
  height: 100%;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
}

/* Header */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  flex-shrink: 0;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.title-line {
  display: flex;
  align-items: center;
  gap: 10px;
}

.res-name {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.ns-tag {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color-lighter);
  padding: 2px 8px;
  border-radius: 4px;
}

.replicas-info {
  font-size: 13px;
  color: var(--el-text-color-regular);
}

.header-actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

/* Main Layout */
.main-layout {
  flex: 1;
  min-height: 0;
  overflow: auto;
}

/* Right Panel */
.right-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 0;
}

.right-section {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  display: flex;
  flex-direction: column;
  background: var(--el-bg-color);
}

.right-section:first-child {
  min-height: 200px;
}

.right-section:last-child {
  max-height: 300px;
  flex-shrink: 0;
}

.panel-title {
  font-size: 13px;
  font-weight: 600;
  padding: 10px 14px;
  background: var(--el-fill-color-lighter);
  border-bottom: 1px solid var(--el-border-color-lighter);
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

.sub-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-regular);
}

.info-body {
  padding: 12px 14px;
  overflow-y: auto;
}

.events-body {
  overflow-y: auto;
  padding: 0;
}

/* Empty hints */
.empty-hint {
  padding: 24px;
  text-align: center;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

/* Responsive */
@media (max-width: 768px) {
  .main-layout {
    overflow: auto;
  }
}
</style>
