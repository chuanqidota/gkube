<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getDeploymentDetail,
  getDeploymentYaml,
  updateDeploymentYaml,
  restartDeployment,
  rollbackDeployment,
  scaleDeployment,
  getPodList,
  getDeploymentEvents,
  deletePod,
  getDeploymentReplicaSets,
} from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'
import ReplicaSetPanel from '@/components/ReplicaSetPanel.vue'
import PodListPanel from '@/components/PodListPanel.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const deployment = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const yamlDialogVisible = ref(false)
const events = ref<any[]>([])
const eventsLoading = ref(false)

// ReplicaSet & Pod panel state
const replicasets = ref<any[]>([])
const replicasetsLoading = ref(false)
const selectedReplicaset = ref<any>(null)
const rsPods = ref<any[]>([])
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
    // Auto-select the current revision's ReplicaSet
    if (replicasets.value.length > 0) {
      const currentRevision = deployment.value?.metadata?.annotations?.['deployment.kubernetes.io/revision']
      const currentRS = replicasets.value.find(
        (rs: any) => rs.metadata.annotations?.['deployment.kubernetes.io/revision'] === currentRevision
      )
      if (currentRS) {
        handleReplicasetSelect(currentRS)
      }
    }
  } catch (e) {
    console.error('Failed to fetch replicasets:', e)
    ElMessage.error('Failed to load ReplicaSets')
  } finally {
    replicasetsLoading.value = false
  }
}

async function fetchReplicasetPods(rsName: string) {
  rsPodsLoading.value = true
  try {
    // The pod-template-hash is the last segment of the ReplicaSet name
    const hash = rsName.split('-').pop()
    const selector = deployment.value?.spec?.selector?.matchLabels || {}
    const selectorEntries = Object.entries(selector).map(([k, v]) => `${k}=${v}`).join(',')
    const labelSelector = selectorEntries ? `${selectorEntries},pod-template-hash=${hash}` : `pod-template-hash=${hash}`
    const res: any = await getPodList({ namespace, labelSelector })
    rsPods.value = res.data?.items || res.data || []
  } catch (e) {
    console.error('Failed to fetch pods:', e)
    ElMessage.error('Failed to load pods')
  } finally {
    rsPodsLoading.value = false
  }
}

function handleReplicasetSelect(rs: any) {
  selectedReplicaset.value = rs
  fetchReplicasetPods(rs.metadata.name)
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

function handlePodLogs(pod: any) {
  window.open(`/logs?namespace=${pod.metadata.namespace || namespace}&pod=${pod.metadata.name}`, '_blank')
}

function handlePodExec(pod: any) {
  window.open(`/terminal?namespace=${pod.metadata.namespace || namespace}&pod=${pod.metadata.name}`, '_blank')
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
    // Refresh the pod list for the selected ReplicaSet
    if (selectedReplicaset.value?.metadata?.name) {
      fetchReplicasetPods(selectedReplicaset.value.metadata.name)
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
  }
}

async function handleRestart() {
  try {
    await ElMessageBox.confirm(`Restart deployment "${name}"? This will trigger a rolling update.`, 'Confirm Restart', { type: 'warning' })
    await restartDeployment({ namespace, name })
    ElMessage.success('Deployment restarted')
    fetchDetail()
    fetchReplicaSets()
  } catch { /* cancelled */ }
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
  fetchDetail()
  fetchReplicaSets()
  fetchEvents()
})
</script>

<template>
  <div class="page-container" v-loading="loading">
    <div class="page-header">
      <h2 style="margin: 0;">Deployment: {{ name }}</h2>
      <div style="display: flex; gap: 8px;">
        <el-button type="primary" @click="handleScale">Scale</el-button>
        <el-button type="warning" @click="handleRestart">Restart</el-button>
        <el-button type="danger" @click="handleRollback">Rollback</el-button>
        <el-button @click="handleOpenYaml">YAML</el-button>
        <el-button @click="router.push('/workloads/deployments')">Back to List</el-button>
      </div>
    </div>

    <!-- Overview Section -->
    <div class="overview-section" v-if="deployment">
      <el-descriptions :column="{ xs: 1, sm: 2, md: 3, lg: 4 }" border size="small">
        <el-descriptions-item label="Replicas">
          {{ deployment.status?.readyReplicas ?? 0 }}/{{ deployment.spec?.replicas ?? 0 }}
        </el-descriptions-item>
        <el-descriptions-item label="Available">
          {{ deployment.status?.availableReplicas ?? '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="Updated">
          {{ deployment.status?.updatedReplicas ?? '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="Strategy">
          {{ deployment.spec?.strategy?.type || '-' }}
        </el-descriptions-item>
      </el-descriptions>
      <div class="overview-tags" v-if="deployment.metadata?.labels && Object.keys(deployment.metadata.labels).length > 0">
        <span class="tag-label">Labels:</span>
        <el-tag v-for="(val, key) in deployment.metadata.labels" :key="key" size="small">
          {{ key }}={{ val }}
        </el-tag>
      </div>
      <div class="overview-tags" v-if="deployment.spec?.selector?.matchLabels && Object.keys(deployment.spec.selector.matchLabels).length > 0">
        <span class="tag-label">Selector:</span>
        <el-tag v-for="(val, key) in deployment.spec.selector.matchLabels" :key="key" size="small" type="info">
          {{ key }}={{ val }}
        </el-tag>
      </div>
    </div>

    <template v-if="deployment">
      <div class="main-content">
        <!-- Left Panel: ReplicaSet List -->
        <div class="left-panel">
          <ReplicaSetPanel
            :replicasets="replicasets"
            :current-revision="parseInt(deployment?.metadata?.annotations?.['deployment.kubernetes.io/revision'] || '0')"
            :loading="replicasetsLoading"
            :selected-name="selectedReplicaset?.metadata?.name"
            @select="handleReplicasetSelect"
            @rollback="handleReplicasetRollback"
          />
        </div>

        <!-- Right Panel: Events + Pods -->
        <div class="right-panel">
          <!-- Events Section -->
          <div class="events-section">
            <div class="section-header">
              <span class="section-title">Events</span>
            </div>
            <div v-loading="eventsLoading" class="events-content">
              <el-table :data="events" stripe size="small" max-height="200">
                <el-table-column prop="type" label="Type" width="80">
                  <template #default="{ row }"><el-tag :type="row.type === 'Warning' ? 'danger' : 'info'" size="small">{{ row.type }}</el-tag></template>
                </el-table-column>
                <el-table-column prop="reason" label="Reason" width="120" />
                <el-table-column prop="message" label="Message" min-width="200" show-overflow-tooltip />
                <el-table-column prop="last_seen" label="Last Seen" width="150" />
              </el-table>
              <el-empty v-if="!eventsLoading && events.length === 0" description="No events" :image-size="60" />
            </div>
          </div>

          <!-- Pods Section -->
          <div class="pods-section">
            <PodListPanel
              :pods="rsPods"
              :loading="rsPodsLoading"
              :replicaset-name="selectedReplicaset?.metadata?.name || ''"
              @logs="handlePodLogs"
              @exec="handlePodExec"
              @delete="handlePodDelete"
            />
          </div>
        </div>
      </div>
    </template>

    <!-- YAML Dialog -->
    <el-dialog v-model="yamlDialogVisible" title="YAML Editor" width="70%" top="5vh" destroy-on-close>
      <div v-loading="yamlLoading">
        <YamlEditor
          v-model="yamlContent"
          height="600px"
          :read-only="true"
          :saveable="true"
          @save="handleSaveYaml"
        />
      </div>
    </el-dialog>

    <!-- Rollback Dialog -->
    <el-dialog v-model="rollbackDialogVisible" title="Rollback Deployment" width="480px" destroy-on-close>
      <div>
        <p style="margin-bottom: 16px;">Rollback deployment <strong>{{ name }}</strong> in namespace <strong>{{ namespace }}</strong>.</p>
        <el-alert v-if="deployment?.metadata?.annotations?.['deployment.kubernetes.io/revision']" :title="`Current revision: ${deployment.metadata.annotations['deployment.kubernetes.io/revision']}`" type="info" :closable="false" style="margin-bottom: 16px;" />
        <el-form-item label="Target Revision">
          <el-input-number v-model="rollbackRevision" :min="1" style="width: 200px;" />
        </el-form-item>
        <el-alert title="This will roll back the deployment to the specified revision by restoring the Pod template from that revision's ReplicaSet." type="warning" :closable="false" show-icon />
      </div>
      <template #footer>
        <el-button @click="rollbackDialogVisible = false">Cancel</el-button>
        <el-button type="danger" :loading="rollbackLoading" @click="handleRollbackConfirm">Rollback</el-button>
      </template>
    </el-dialog>

    <!-- Scale Dialog -->
    <el-dialog v-model="scaleDialogVisible" title="Scale Deployment" width="480px" destroy-on-close>
      <div>
        <p style="margin-bottom: 16px;">Scale deployment <strong>{{ name }}</strong> in namespace <strong>{{ namespace }}</strong>.</p>
        <el-descriptions :column="1" border style="margin-bottom: 16px;">
          <el-descriptions-item label="Current Replicas">{{ deployment?.spec?.replicas ?? '-' }}</el-descriptions-item>
          <el-descriptions-item label="Ready Replicas">{{ deployment?.status?.readyReplicas ?? '-' }}</el-descriptions-item>
        </el-descriptions>
        <el-form-item label="Target Replicas">
          <el-input-number v-model="scaleReplicas" :min="0" :max="100" style="width: 200px;" />
        </el-form-item>
        <el-alert v-if="scaleReplicas === 0" title="Setting replicas to 0 will stop all pods." type="warning" :closable="false" show-icon style="margin-top: 8px;" />
      </div>
      <template #footer>
        <el-button @click="scaleDialogVisible = false">Cancel</el-button>
        <el-button type="primary" :loading="scaleLoading" @click="handleScaleConfirm">Scale</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }

.main-content {
  display: flex;
  height: calc(100vh - 120px);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 4px;
  overflow: hidden;
}

.left-panel {
  width: 320px;
  min-width: 320px;
  border-right: 1px solid var(--el-border-color-lighter);
  overflow-y: auto;
}

.right-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.events-section {
  border-bottom: 1px solid var(--el-border-color-lighter);
  display: flex;
  flex-direction: column;
}

.section-header {
  padding: 12px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  background-color: var(--el-fill-color-lighter);
}

.section-title {
  font-weight: 500;
  font-size: 14px;
}

.events-content {
  padding: 12px;
  overflow-y: auto;
}

.pods-section {
  flex: 1;
  overflow-y: auto;
}

.overview-section {
  padding: 12px 16px;
  background-color: var(--el-fill-color-lighter);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 4px;
  margin-bottom: 16px;
}

.overview-tags {
  margin-top: 8px;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px;
}

.tag-label {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin-right: 8px;
}

@media (max-width: 1199px) {
  .left-panel {
    width: 280px;
    min-width: 280px;
  }
}

@media (max-width: 768px) {
  .main-content {
    flex-direction: column;
    height: auto;
  }

  .left-panel {
    width: 100%;
    min-width: 100%;
    border-right: none;
    border-bottom: 1px solid var(--el-border-color-lighter);
    max-height: 300px;
  }
}
</style>
