<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getDeploymentDetail,
  getDeploymentYaml,
  updateDeploymentYaml,
  restartDeployment,
  rollbackDeployment,
  scaleDeployment,
  getDeploymentPodList,
  getDeploymentEvents,
  deletePod,
  getDeploymentReplicaSets,
} from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'
import PodListPanel from '@/components/PodListPanel.vue'
import { formatAge } from '@/utils/time'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const deployment = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const yamlDialogVisible = ref(false)
const yamlEditorRef = ref<InstanceType<typeof YamlEditor>>()
const events = ref<any[]>([])
const eventsLoading = ref(false)

// ReplicaSet & Pod panel state
const replicasets = ref<any[]>([])
const replicasetsLoading = ref(false)
const selectedReplicaset = ref<any>(null)
const rsPods = ref<any[]>([])
const allPods = ref<any[]>([])
const rsPodsLoading = ref(false)

// Rollback dialog
const rollbackDialogVisible = ref(false)
const rollbackRevision = ref<number>(1)
const rollbackLoading = ref(false)

// Scale dialog
const scaleDialogVisible = ref(false)
const scaleReplicas = ref<number>(1)
const scaleLoading = ref(false)

const namespace = route.params.namespace as string
const name = route.params.name as string

const statusTagType = computed(() => {
  const conditions = deployment.value?.status?.conditions || []
  const available = conditions.find((c: any) => c.type === 'Available')
  if (available?.status === 'True') return 'success'
  const progressing = conditions.find((c: any) => c.type === 'Progressing')
  if (progressing?.status === 'True') return 'warning'
  return 'danger'
})

const statusText = computed(() => {
  const conditions = deployment.value?.status?.conditions || []
  const available = conditions.find((c: any) => c.type === 'Available')
  if (available?.status === 'True') return 'Available'
  const progressing = conditions.find((c: any) => c.type === 'Progressing')
  if (progressing?.status === 'True') return 'Progressing'
  return 'Unavailable'
})

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getDeploymentDetail({ namespace, name })
    deployment.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load deployment detail')
  } finally {
    loading.value = false
  }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getDeploymentYaml({ namespace, name })
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
    const res: any = await getDeploymentEvents({ namespace, name })
    events.value = res.data || []
  } catch (e) {
    console.error('Failed to fetch events:', e)
    ElMessage.error('Failed to load events')
  } finally {
    eventsLoading.value = false
  }
}

async function fetchReplicaSets() {
  replicasetsLoading.value = true
  try {
    const res: any = await getDeploymentReplicaSets({ namespace, name })
    replicasets.value = res.data?.items || res.data || []
    if (replicasets.value.length > 0) {
      const currentRevision = deployment.value?.metadata?.annotations?.['deployment.kubernetes.io/revision']
      const currentRS = replicasets.value.find(
        (rs: any) => rs.metadata.annotations?.['deployment.kubernetes.io/revision'] === currentRevision
      )
      await fetchAllPods()
      if (currentRS) {
        handleReplicasetSelect(currentRS)
      } else {
        rsPods.value = allPods.value
      }
    }
  } catch (e) {
    console.error('Failed to fetch replicasets:', e)
    ElMessage.error('Failed to load ReplicaSets')
  } finally {
    replicasetsLoading.value = false
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
    ElMessage.error("Failed to load pods")
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
      `Are you sure you want to rollback to revision ${revision}?`,
      'Confirm Rollback',
      { type: 'warning' }
    )
    await rollbackDeployment({ namespace, name, revision: parseInt(revision, 10) })
    ElMessage.success('Rollback successful')
    fetchDetail()
    fetchReplicaSets()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('Rollback failed')
    }
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
  window.open(`/logs?namespace=${pod.metadata.namespace || namespace}&pod=${pod.metadata.name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
}

function handlePodExec(pod: any) {
  const cluster = getClusterName()
  window.open(`/terminal?namespace=${pod.metadata.namespace || namespace}&pod=${pod.metadata.name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
}

async function handlePodDelete(pod: any) {
  try {
    await ElMessageBox.confirm(
      `Are you sure you want to delete pod ${pod.metadata.name}?`,
      'Confirm Delete',
      { type: 'warning' }
    )
    await deletePod({ namespace, name: pod.metadata.name })
    ElMessage.success('Pod deleted')
    if (selectedReplicaset.value) {
      handleReplicasetSelect(selectedReplicaset.value)
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('Delete failed')
    }
  }
}

function handleOpenYaml() {
  fetchYaml()
  yamlDialogVisible.value = true
}

async function handleSaveYaml(content: string) {
  try {
    await updateDeploymentYaml({ namespace, name, yaml: content })
    ElMessage.success('YAML saved successfully')
    yamlDialogVisible.value = false
    fetchDetail()
    fetchReplicaSets()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to save YAML')
    yamlEditorRef.value?.resetSaving()
  }
}

async function handleRestart() {
  try {
    await ElMessageBox.confirm(
      `Restart deployment "${name}"? This will trigger a rolling update.`,
      'Confirm Restart',
      { type: 'warning' }
    )
    await restartDeployment({ namespace, name })
    ElMessage.success('Deployment restarted')
    fetchDetail()
    fetchReplicaSets()
  } catch {
    // cancelled
  }
}

function handleRollback() {
  const annotations = deployment.value?.metadata?.annotations || {}
  const currentRevision = parseInt(annotations['deployment.kubernetes.io/revision'] || '0', 10)
  rollbackRevision.value = Math.max(1, currentRevision - 1)
  rollbackDialogVisible.value = true
}

function handleScale() {
  scaleReplicas.value = deployment.value?.spec?.replicas ?? 1
  scaleDialogVisible.value = true
}

async function handleScaleConfirm() {
  scaleLoading.value = true
  try {
    await scaleDeployment({ namespace, name, replicas: scaleReplicas.value })
    ElMessage.success(`Deployment scaled to ${scaleReplicas.value} replicas`)
    scaleDialogVisible.value = false
    fetchDetail()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to scale deployment')
  } finally {
    scaleLoading.value = false
  }
}

async function handleRollbackConfirm() {
  if (!rollbackRevision.value || rollbackRevision.value < 1) {
    ElMessage.warning('Please enter a valid revision number')
    return
  }
  rollbackLoading.value = true
  try {
    await rollbackDeployment({ namespace, name, revision: rollbackRevision.value })
    ElMessage.success(`Deployment rolled back to revision ${rollbackRevision.value}`)
    rollbackDialogVisible.value = false
    fetchDetail()
    fetchReplicaSets()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to rollback deployment')
  } finally {
    rollbackLoading.value = false
  }
}

onMounted(() => {
  fetchDetail().then(() => {
    fetchReplicaSets()
  })
  fetchEvents()
})
</script>

<template>
  <div class="detail-page" v-loading="loading">

    <!-- ===== 顶部标题栏 ===== -->
    <div class="page-header">
      <div class="header-left">
        <el-button link type="primary" @click="router.push('/workloads/deployments')" class="back-btn">← 返回列表</el-button>
        <div class="title-line">
          <h2 class="res-name">{{ name }}</h2>
          <el-tag :type="statusTagType" effect="dark" size="small">{{ statusText }}</el-tag>
          <span class="ns-tag">ns/{{ namespace }}</span>
          <span class="replicas-info" v-if="deployment">
            {{ deployment.status?.readyReplicas ?? 0 }}/{{ deployment.spec?.replicas ?? 0 }} ready
          </span>
        </div>
      </div>
      <div class="header-actions">
        <el-button type="primary" size="small" @click="handleScale">扩缩容</el-button>
        <el-button type="warning" size="small" @click="handleRestart">重启</el-button>
        <el-button type="danger" size="small" @click="handleRollback">回滚</el-button>
        <el-button size="small" @click="handleOpenYaml">YAML</el-button>
      </div>
    </div>

    <template v-if="deployment">
      <!-- ===== 主体：左侧 RS + 右侧 Pods / Events ===== -->
      <div class="main-layout">

        <!-- 左侧：ReplicaSet 列表 -->
        <div class="left-panel">
          <div class="panel-title">ReplicaSet ({{ replicasets.length }})</div>
          <div class="rs-list" v-loading="replicasetsLoading">
            <div v-if="replicasets.length === 0" class="empty-hint">暂无 ReplicaSet</div>
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
                  type="success" size="small">当前</el-tag>
                <el-tag v-else-if="(rs.status?.readyReplicas || 0) > 0" type="primary" size="small">活跃</el-tag>
              </div>
              <div class="rs-image">{{ rs.spec?.template?.spec?.containers?.[0]?.image || '-' }}</div>
              <div class="rs-age">{{ formatAge(rs.metadata.creationTimestamp) }}</div>
              <div class="rs-rollback" v-if="rs.metadata.annotations?.['deployment.kubernetes.io/revision'] !== deployment?.metadata?.annotations?.['deployment.kubernetes.io/revision']">
                <el-button size="small" type="warning" @click.stop="handleReplicasetRollback(rs)">回滚</el-button>
              </div>
            </div>
          </div>
        </div>

        <!-- 右侧：Pods + Events -->
        <div class="right-panel">

          <!-- Pod 列表 -->
          <div class="right-section">
            <div class="panel-title">
              Pod 列表
              <span class="count-badge">{{ rsPods.length }} 个</span>
              <span class="rs-label" v-if="selectedReplicaset">{{ selectedReplicaset.metadata.name }}</span>
            </div>
            <PodListPanel
              :pods="rsPods"
              :loading="rsPodsLoading"
              :replicaset-name="selectedReplicaset?.metadata?.name || ''"
              @logs="handlePodLogs"
              @exec="handlePodExec"
              @delete="handlePodDelete"
            />
          </div>

          <!-- Events -->
          <div class="right-section">
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
        <YamlEditor ref="yamlEditorRef" v-model="yamlContent" height="600px" :read-only="true" :saveable="true" @save="handleSaveYaml" />
      </div>
    </el-dialog>

    <el-dialog v-model="rollbackDialogVisible" title="回滚" width="480px" destroy-on-close>
      <div>
        <p style="margin-bottom: 16px;">回滚 <strong>{{ name }}</strong>（{{ namespace }}）</p>
        <el-alert v-if="deployment?.metadata?.annotations?.['deployment.kubernetes.io/revision']" :title="`当前版本: ${deployment.metadata.annotations['deployment.kubernetes.io/revision']}`" type="info" :closable="false" style="margin-bottom: 16px;" />
        <el-form-item label="目标版本">
          <el-input-number v-model="rollbackRevision" :min="1" style="width: 200px;" />
        </el-form-item>
        <el-alert title="回滚将用指定版本的 Pod 模板替换当前模板。" type="warning" :closable="false" show-icon />
      </div>
      <template #footer>
        <el-button @click="rollbackDialogVisible = false">取消</el-button>
        <el-button type="danger" :loading="rollbackLoading" @click="handleRollbackConfirm">确认回滚</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="scaleDialogVisible" title="扩缩容" width="480px" destroy-on-close>
      <div>
        <p style="margin-bottom: 16px;">调整 <strong>{{ name }}</strong> 副本数</p>
        <el-descriptions :column="1" border size="small" style="margin-bottom: 16px;">
          <el-descriptions-item label="当前">{{ deployment?.spec?.replicas ?? '-' }}</el-descriptions-item>
          <el-descriptions-item label="就绪">{{ deployment?.status?.readyReplicas ?? '-' }}</el-descriptions-item>
        </el-descriptions>
        <el-form-item label="目标">
          <el-input-number v-model="scaleReplicas" :min="0" :max="100" style="width: 200px;" />
        </el-form-item>
        <el-alert v-if="scaleReplicas === 0" title="设为 0 将停止所有 Pod。" type="warning" :closable="false" show-icon style="margin-top: 8px;" />
      </div>
      <template #footer>
        <el-button @click="scaleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="scaleLoading" @click="handleScaleConfirm">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.detail-page {
  padding: 16px 20px;
  height: 100vh;
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

.back-btn {
  align-self: flex-start;
  margin-bottom: 2px;
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

/* Main Layout: left RS + right Pods/Events */
.main-layout {
  display: flex;
  gap: 12px;
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

/* Left Panel - RS List */
.left-panel {
  width: 320px;
  min-width: 320px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--el-bg-color);
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

.rs-label {
  margin-left: auto;
  font-weight: 400;
  font-size: 11px;
  color: var(--el-text-color-placeholder);
  font-family: monospace;
  max-width: 140px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rs-list {
  flex: 1;
  overflow-y: auto;
}

.rs-item {
  padding: 10px 14px;
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
  font-size: 13px;
  font-weight: 500;
  font-family: monospace;
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
  color: var(--el-text-color-secondary);
}

.rs-image {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  word-break: break-all;
  margin-bottom: 2px;
}

.rs-age {
  font-size: 11px;
  color: var(--el-text-color-placeholder);
}

.rs-rollback {
  margin-top: 6px;
}

/* Right Panel */
.right-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
  overflow: hidden;
}

.right-section {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--el-bg-color);
}

.right-section:first-child {
  flex: 1;
  min-height: 0;
}

.right-section:last-child {
  max-height: 300px;
  flex-shrink: 0;
}

.events-body {
  flex: 1;
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
@media (max-width: 1199px) {
  .left-panel {
    width: 260px;
    min-width: 260px;
  }
}

@media (max-width: 768px) {
  .main-layout {
    flex-direction: column;
    overflow: auto;
  }
  .left-panel {
    width: 100%;
    min-width: 100%;
    max-height: 260px;
  }
}
</style>
