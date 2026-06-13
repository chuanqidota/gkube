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
} from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const deployment = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const yamlEditing = ref(false)
const yamlSaving = ref(false)
const activeTab = ref('info')
const pods = ref<any[]>([])
const podsLoading = ref(false)
const events = ref<any[]>([])
const eventsLoading = ref(false)

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

async function fetchPods() {
  podsLoading.value = true
  try {
    const selector = deployment.value?.selector
    const selectorStr = selector ? Object.entries(selector).map(([k, v]) => `${k}=${v}`).join(',') : ''
    const res: any = await getPodList({ namespace, labelSelector: selectorStr })
    pods.value = res.data || []
  } catch {
    // ignore
  } finally {
    podsLoading.value = false
  }
}

async function fetchEvents() {
  eventsLoading.value = true
  try {
    const res: any = await getDeploymentEvents({ namespace, name })
    events.value = res.data || []
  } catch {
    // ignore
  } finally {
    eventsLoading.value = false
  }
}

function handleTabChange(tab: string) {
  if (tab === 'yaml' && !yamlContent.value) fetchYaml()
  if (tab === 'pods' && pods.value.length === 0) fetchPods()
  if (tab === 'events' && events.value.length === 0) fetchEvents()
}

async function handleSaveYaml() {
  yamlSaving.value = true
  try {
    await updateDeploymentYaml({ namespace, name, yaml: yamlContent.value })
    ElMessage.success('YAML saved successfully')
    yamlEditing.value = false
    fetchDetail()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to save YAML')
  } finally {
    yamlSaving.value = false
  }
}

async function handleRestart() {
  try {
    await ElMessageBox.confirm(`Restart deployment "${name}"? This will trigger a rolling update.`, 'Confirm Restart', { type: 'warning' })
    await restartDeployment({ namespace, name })
    ElMessage.success('Deployment restarted')
    fetchDetail()
  } catch { /* cancelled */ }
}

function handleRollback() {
  const annotations = deployment.value?.annotations || {}
  const currentRevision = parseInt(annotations['deployment.kubernetes.io/revision'] || '0', 10)
  rollbackRevision.value = Math.max(1, currentRevision - 1)
  rollbackDialogVisible.value = true
}

function handleScale() {
  scaleReplicas.value = deployment.value?.replicas ?? 1
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
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to rollback deployment')
  } finally {
    rollbackLoading.value = false
  }
}

function handlePodDetail(row: any) {
  router.push(`/workloads/pods/${row.namespace || namespace}/${row.name}`)
}

function handlePodLogs(row: any) {
  router.push({ path: '/logs', query: { namespace: row.namespace || namespace, pod: row.name } })
}

function handlePodExec(row: any) {
  router.push({ path: '/terminal', query: { namespace: row.namespace || namespace, pod: row.name } })
}

function podStatusType(status: string) {
  const s = (status || '').toLowerCase()
  if (s === 'running') return 'success'
  if (s === 'succeeded') return 'info'
  if (s === 'pending') return 'warning'
  if (s === 'failed') return 'danger'
  return 'info'
}

onMounted(fetchDetail)
</script>

<template>
  <div class="page-container" v-loading="loading">
    <div class="page-header">
      <h2 style="margin: 0;">Deployment: {{ name }}</h2>
      <div style="display: flex; gap: 8px;">
        <el-button type="primary" @click="handleScale">Scale</el-button>
        <el-button type="warning" @click="handleRestart">Restart</el-button>
        <el-button type="danger" @click="handleRollback">Rollback</el-button>
        <el-button @click="router.push('/workloads/deployments')">Back to List</el-button>
      </div>
    </div>

    <template v-if="deployment">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <!-- Info Tab -->
        <el-tab-pane label="Info" name="info">
          <el-card shadow="never">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="Name">{{ deployment.name }}</el-descriptions-item>
              <el-descriptions-item label="Namespace">{{ deployment.namespace }}</el-descriptions-item>
              <el-descriptions-item label="Replicas">{{ deployment.replicas ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="Ready">{{ deployment.ready ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="Updated">{{ deployment.updated ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="Available">{{ deployment.available ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="Strategy">{{ deployment.strategy || '-' }}</el-descriptions-item>
              <el-descriptions-item label="Age">{{ deployment.age || '-' }}</el-descriptions-item>
            </el-descriptions>

            <div v-if="deployment.labels && Object.keys(deployment.labels).length > 0" style="margin-top: 16px;">
              <h4>Labels</h4>
              <el-tag v-for="(val, key) in deployment.labels" :key="key" style="margin-right: 8px; margin-bottom: 8px;">{{ key }}={{ val }}</el-tag>
            </div>

            <div v-if="deployment.selector && Object.keys(deployment.selector).length > 0" style="margin-top: 16px;">
              <h4>Selector</h4>
              <el-tag v-for="(val, key) in deployment.selector" :key="key" style="margin-right: 8px; margin-bottom: 8px;" type="info">{{ key }}={{ val }}</el-tag>
            </div>

            <div v-if="deployment.conditions && deployment.conditions.length > 0" style="margin-top: 16px;">
              <h4>Conditions</h4>
              <el-table :data="deployment.conditions" border stripe>
                <el-table-column prop="type" label="Type" width="160" />
                <el-table-column label="Status" width="100">
                  <template #default="{ row }"><el-tag :type="row.status === 'True' ? 'success' : 'danger'" size="small">{{ row.status }}</el-tag></template>
                </el-table-column>
                <el-table-column prop="reason" label="Reason" width="160" />
                <el-table-column prop="message" label="Message" min-width="250" show-overflow-tooltip />
                <el-table-column prop="lastUpdateTime" label="Last Update" width="180" />
              </el-table>
            </div>
          </el-card>
        </el-tab-pane>

        <!-- Pods Tab -->
        <el-tab-pane label="Pods" name="pods">
          <el-card shadow="never">
            <el-table :data="pods" v-loading="podsLoading" stripe>
              <el-table-column prop="name" label="Name" min-width="200" show-overflow-tooltip>
                <template #default="{ row }"><el-button link type="primary" @click="handlePodDetail(row)">{{ row.name }}</el-button></template>
              </el-table-column>
              <el-table-column prop="namespace" label="Namespace" width="140" />
              <el-table-column prop="status" label="Status" width="120">
                <template #default="{ row }"><el-tag :type="podStatusType(row.status)" size="small">{{ row.status }}</el-tag></template>
              </el-table-column>
              <el-table-column prop="restarts" label="Restarts" width="100" />
              <el-table-column prop="age" label="Age" width="120" />
              <el-table-column label="Actions" width="200" fixed="right">
                <template #default="{ row }">
                  <el-button size="small" type="primary" @click="handlePodLogs(row)">Logs</el-button>
                  <el-button size="small" type="success" @click="handlePodExec(row)">Exec</el-button>
                </template>
              </el-table-column>
            </el-table>
            <el-empty v-if="!podsLoading && pods.length === 0" description="No pods found" />
          </el-card>
        </el-tab-pane>

        <!-- Events Tab -->
        <el-tab-pane label="Events" name="events">
          <el-card shadow="never">
            <el-table :data="events" v-loading="eventsLoading" stripe>
              <el-table-column prop="type" label="Type" width="100">
                <template #default="{ row }"><el-tag :type="row.type === 'Warning' ? 'danger' : 'info'" size="small">{{ row.type }}</el-tag></template>
              </el-table-column>
              <el-table-column prop="reason" label="Reason" width="150" />
              <el-table-column prop="message" label="Message" min-width="300" show-overflow-tooltip />
              <el-table-column prop="last_seen" label="Last Seen" width="180" />
            </el-table>
            <el-empty v-if="!eventsLoading && events.length === 0" description="No events" />
          </el-card>
        </el-tab-pane>

        <!-- YAML Tab -->
        <el-tab-pane label="YAML" name="yaml">
          <el-card shadow="never">
            <div style="margin-bottom: 12px; display: flex; gap: 8px;">
              <el-button v-if="!yamlEditing" type="primary" @click="yamlEditing = true">Edit YAML</el-button>
              <template v-if="yamlEditing">
                <el-button type="success" :loading="yamlSaving" @click="handleSaveYaml">Save</el-button>
                <el-button @click="yamlEditing = false; fetchYaml()">Cancel</el-button>
              </template>
            </div>
            <div v-loading="yamlLoading">
              <YamlEditor v-model="yamlContent" height="600px" :read-only="!yamlEditing" />
            </div>
          </el-card>
        </el-tab-pane>
      </el-tabs>
    </template>

    <!-- Rollback Dialog -->
    <el-dialog v-model="rollbackDialogVisible" title="Rollback Deployment" width="480px" destroy-on-close>
      <div>
        <p style="margin-bottom: 16px;">Rollback deployment <strong>{{ name }}</strong> in namespace <strong>{{ namespace }}</strong>.</p>
        <el-alert v-if="deployment?.annotations?.['deployment.kubernetes.io/revision']" :title="`Current revision: ${deployment.annotations['deployment.kubernetes.io/revision']}`" type="info" :closable="false" style="margin-bottom: 16px;" />
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
          <el-descriptions-item label="Current Replicas">{{ deployment?.replicas ?? '-' }}</el-descriptions-item>
          <el-descriptions-item label="Ready Replicas">{{ deployment?.ready ?? '-' }}</el-descriptions-item>
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
</style>
