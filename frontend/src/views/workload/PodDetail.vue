<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import { getPodDetail, getPodYaml, updatePodYaml, deletePod, getPodEvents, getPodLogs } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const pod = ref<any>(null)
const events = ref<any[]>([])
const yamlContent = ref('')
const yamlLoading = ref(false)
const yamlEditing = ref(false)
const yamlSaving = ref(false)
const activeTab = ref('info')

// Logs
const logs = ref('')
const logsLoading = ref(false)
const selectedContainer = ref('')
const tailLines = ref(100)

const namespace = route.params.namespace as string
const name = route.params.name as string

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getPodDetail({ namespace, name })
    pod.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load pod detail')
  } finally {
    loading.value = false
  }
}

async function fetchEvents() {
  try {
    const res: any = await getPodEvents({ namespace, name })
    events.value = res.data || []
  } catch { /* ignore */ }
}

async function fetchLogs() {
  logsLoading.value = true
  try {
    const params: any = { namespace, podName: name, tailLines: tailLines.value }
    if (selectedContainer.value) params.container = selectedContainer.value
    const res: any = await getPodLogs(params)
    logs.value = res.data?.logs || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load logs')
  } finally {
    logsLoading.value = false
  }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getPodYaml({ namespace, name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load YAML')
  } finally {
    yamlLoading.value = false
  }
}

function handleTabChange(tab: string) {
  if (tab === 'yaml' && !yamlContent.value) fetchYaml()
  if (tab === 'events' && events.value.length === 0) fetchEvents()
  if (tab === 'logs' && !logs.value) fetchLogs()
}

async function handleSaveYaml() {
  yamlSaving.value = true
  try {
    await updatePodYaml({ namespace, name, yaml: yamlContent.value })
    ElMessage.success('YAML saved successfully')
    yamlEditing.value = false
    fetchDetail()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to save YAML')
  } finally {
    yamlSaving.value = false
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

function handleLogs() {
  const cluster = getClusterName()
  window.open(`/logs?namespace=${namespace}&pod=${name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
}

function handleExec() {
  const cluster = getClusterName()
  window.open(`/terminal?namespace=${namespace}&pod=${name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
}

function handleFullLogViewer() {
  const cluster = getClusterName()
  window.open(`/logs?namespace=${namespace}&pod=${name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(`Delete pod "${name}" in namespace "${namespace}"?`, 'Confirm', { type: 'warning' })
    await deletePod({ namespace, name })
    ElMessage.success('Pod deleted')
    router.push('/workloads/pods')
  } catch { /* cancelled */ }
}

function statusType(status: string) {
  const s = (status || '').toLowerCase()
  if (s === 'running') return 'success'
  if (s === 'succeeded') return 'info'
  if (s === 'pending') return 'warning'
  if (s === 'failed' || s === 'error') return 'danger'
  return 'info'
}

function conditionStatusType(status: string) {
  if ((status || '').toLowerCase() === 'true') return 'success'
  if ((status || '').toLowerCase() === 'false') return 'danger'
  return 'warning'
}

function containerStateType(state: string) {
  const s = (state || '').toLowerCase()
  if (s === 'running') return 'success'
  if (s === 'waiting') return 'warning'
  if (s === 'terminated') return 'danger'
  return 'info'
}

function getContainerStateLabel(container: any): string {
  if (container.state === 'Running') return 'Running'
  if (container.state === 'Waiting') return `Waiting (${container.stateReason || '-'})`
  if (container.state === 'Terminated') return `Terminated (${container.exitCode ?? '-'})`
  return container.state || '-'
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchDetail, { autoStart: false })

onMounted(fetchDetail)
</script>

<template>
  <div class="page-container" v-loading="loading">
    <div class="page-header">
      <h2 style="margin: 0;">Pod: {{ name }}</h2>
      <div style="display: flex; gap: 8px;">
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
        <el-button type="primary" @click="handleLogs">Logs</el-button>
        <el-button type="success" @click="handleExec">Exec</el-button>
        <el-button type="danger" @click="handleDelete">删除</el-button>
        <el-button @click="router.push('/workloads/pods')">Back to List</el-button>
      </div>
    </div>

    <template v-if="pod">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <!-- Info Tab -->
        <el-tab-pane label="Info" name="info">
          <el-card shadow="never">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="Name">{{ pod.name }}</el-descriptions-item>
              <el-descriptions-item label="Namespace">{{ pod.namespace }}</el-descriptions-item>
              <el-descriptions-item label="Status"><el-tag :type="statusType(pod.status)" size="small">{{ pod.status }}</el-tag></el-descriptions-item>
              <el-descriptions-item label="Pod IP">{{ pod.ip || '-' }}</el-descriptions-item>
              <el-descriptions-item label="Host IP">{{ pod.host_ip || pod.node || '-' }}</el-descriptions-item>
              <el-descriptions-item label="Node">{{ pod.node || '-' }}</el-descriptions-item>
              <el-descriptions-item label="Restarts">{{ pod.restarts ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="QoS Class">{{ pod.qos_class || '-' }}</el-descriptions-item>
              <el-descriptions-item label="Priority">{{ pod.priority ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="Age">{{ pod.age || '-' }}</el-descriptions-item>
              <el-descriptions-item label="Created">{{ pod.created_at || '-' }}</el-descriptions-item>
              <el-descriptions-item label="Service Account">{{ pod.service_account || '-' }}</el-descriptions-item>
            </el-descriptions>

            <div v-if="pod.labels && Object.keys(pod.labels).length > 0" style="margin-top: 16px;">
              <h4>Labels</h4>
              <el-tag v-for="(val, key) in pod.labels" :key="key" style="margin-right: 8px; margin-bottom: 8px;">{{ key }}={{ val }}</el-tag>
            </div>

            <div v-if="pod.conditions && pod.conditions.length > 0" style="margin-top: 16px;">
              <h4>Conditions</h4>
              <el-table :data="pod.conditions" border stripe>
                <el-table-column prop="type" label="Type" min-width="160" />
                <el-table-column label="Status" width="120"><template #default="{ row }"><el-tag :type="conditionStatusType(row.status)" size="small">{{ row.status }}</el-tag></template></el-table-column>
                <el-table-column prop="reason" label="Reason" min-width="160" show-overflow-tooltip />
                <el-table-column prop="message" label="Message" min-width="250" show-overflow-tooltip />
                <el-table-column prop="last_transition_time" label="Last Transition" width="180" />
              </el-table>
            </div>
          </el-card>
        </el-tab-pane>

        <!-- Containers Tab -->
        <el-tab-pane label="Containers" name="containers">
          <el-card shadow="never">
            <el-table :data="pod.containers || []" border stripe row-key="name">
              <el-table-column type="expand">
                <template #default="{ row }">
                  <div style="padding: 12px 16px;">
                    <div v-if="row.ports && row.ports.length > 0" style="margin-bottom: 16px;">
                      <h4 style="margin: 0 0 8px 0;">Ports</h4>
                      <el-table :data="row.ports" border size="small">
                        <el-table-column prop="name" label="Name" width="120" />
                        <el-table-column prop="containerPort" label="Container Port" width="130" />
                        <el-table-column prop="protocol" label="Protocol" width="100" />
                      </el-table>
                    </div>
                    <div v-if="row.env && row.env.length > 0" style="margin-bottom: 16px;">
                      <h4 style="margin: 0 0 8px 0;">Environment Variables</h4>
                      <el-table :data="row.env" border size="small">
                        <el-table-column prop="name" label="Name" min-width="180" />
                        <el-table-column label="Value" min-width="250">
                          <template #default="{ row: envRow }">
                            <span v-if="envRow.value !== undefined && envRow.value !== ''">{{ envRow.value }}</span>
                            <span v-else-if="envRow.valueFrom" style="color: var(--gk-color-text-secondary);">{{ envRow.valueFrom.fieldRef?.fieldPath || envRow.valueFrom.secretKeyRef?.name || envRow.valueFrom.configMapKeyRef?.name || 'From reference' }}</span>
                            <span v-else style="color: var(--gk-color-text-secondary);">-</span>
                          </template>
                        </el-table-column>
                      </el-table>
                    </div>
                    <div v-if="row.volumeMounts && row.volumeMounts.length > 0" style="margin-bottom: 16px;">
                      <h4 style="margin: 0 0 8px 0;">Volume Mounts</h4>
                      <el-table :data="row.volumeMounts" border size="small">
                        <el-table-column prop="name" label="Volume Name" min-width="150" />
                        <el-table-column prop="mountPath" label="Mount Path" min-width="200" />
                        <el-table-column prop="subPath" label="Sub Path" width="150" />
                        <el-table-column label="Read Only" width="100"><template #default="{ row: vm }"><el-tag :type="vm.readOnly ? 'warning' : 'success'" size="small">{{ vm.readOnly ? 'Yes' : 'No' }}</el-tag></template></el-table-column>
                      </el-table>
                    </div>
                    <div v-if="row.livenessProbe" style="margin-bottom: 16px;">
                      <h4 style="margin: 0 0 8px 0;">Liveness Probe</h4>
                      <el-descriptions :column="2" border size="small">
                        <el-descriptions-item v-if="row.livenessProbe.httpGet" label="Type">HTTP GET</el-descriptions-item>
                        <el-descriptions-item v-if="row.livenessProbe.httpGet" label="Path">{{ row.livenessProbe.httpGet.path || '/' }}</el-descriptions-item>
                        <el-descriptions-item v-if="row.livenessProbe.httpGet" label="Port">{{ row.livenessProbe.httpGet.port }}</el-descriptions-item>
                        <el-descriptions-item v-if="row.livenessProbe.tcpSocket" label="Type">TCP Socket</el-descriptions-item>
                        <el-descriptions-item v-if="row.livenessProbe.exec" label="Type">Exec</el-descriptions-item>
                        <el-descriptions-item v-if="row.livenessProbe.exec" label="Command">{{ (row.livenessProbe.exec.command || []).join(' ') }}</el-descriptions-item>
                        <el-descriptions-item label="Initial Delay">{{ row.livenessProbe.initialDelaySeconds ?? '-' }}s</el-descriptions-item>
                        <el-descriptions-item label="Period">{{ row.livenessProbe.periodSeconds ?? '-' }}s</el-descriptions-item>
                      </el-descriptions>
                    </div>
                    <div v-if="row.readinessProbe">
                      <h4 style="margin: 0 0 8px 0;">Readiness Probe</h4>
                      <el-descriptions :column="2" border size="small">
                        <el-descriptions-item v-if="row.readinessProbe.httpGet" label="Type">HTTP GET</el-descriptions-item>
                        <el-descriptions-item v-if="row.readinessProbe.httpGet" label="Path">{{ row.readinessProbe.httpGet.path || '/' }}</el-descriptions-item>
                        <el-descriptions-item v-if="row.readinessProbe.httpGet" label="Port">{{ row.readinessProbe.httpGet.port }}</el-descriptions-item>
                        <el-descriptions-item label="Initial Delay">{{ row.readinessProbe.initialDelaySeconds ?? '-' }}s</el-descriptions-item>
                        <el-descriptions-item label="Period">{{ row.readinessProbe.periodSeconds ?? '-' }}s</el-descriptions-item>
                      </el-descriptions>
                    </div>
                    <el-empty v-if="(!row.ports || row.ports.length === 0) && (!row.env || row.env.length === 0) && (!row.volumeMounts || row.volumeMounts.length === 0) && !row.livenessProbe && !row.readinessProbe" description="No additional container details" :image-size="60" />
                  </div>
                </template>
              </el-table-column>
              <el-table-column prop="name" label="Name" min-width="160" />
              <el-table-column prop="image" label="Image" min-width="280" show-overflow-tooltip />
              <el-table-column label="Ready" width="80"><template #default="{ row }"><el-tag :type="row.ready ? 'success' : 'danger'" size="small">{{ row.ready ? 'Yes' : 'No' }}</el-tag></template></el-table-column>
              <el-table-column prop="restartCount" label="Restarts" width="100" />
              <el-table-column label="State" width="180"><template #default="{ row }"><el-tag :type="containerStateType(row.state)" size="small">{{ getContainerStateLabel(row) }}</el-tag></template></el-table-column>
            </el-table>
            <el-empty v-if="!pod.containers || pod.containers.length === 0" description="No containers" />
          </el-card>
        </el-tab-pane>

        <!-- Events Tab -->
        <el-tab-pane label="Events" name="events">
          <el-card shadow="never">
            <el-table :data="events" border stripe>
              <el-table-column prop="type" label="Type" width="100"><template #default="{ row }"><el-tag :type="row.type === 'Warning' ? 'danger' : 'info'" size="small">{{ row.type }}</el-tag></template></el-table-column>
              <el-table-column prop="reason" label="Reason" width="150" />
              <el-table-column prop="message" label="Message" min-width="300" show-overflow-tooltip />
              <el-table-column prop="last_seen" label="Last Seen" width="180" />
            </el-table>
            <el-empty v-if="events.length === 0" description="No events" />
          </el-card>
        </el-tab-pane>

        <!-- Logs Tab -->
        <el-tab-pane label="Logs" name="logs">
          <el-card shadow="never">
            <div style="margin-bottom: 12px; display: flex; gap: 12px; align-items: center;">
              <el-select v-if="pod.containers && pod.containers.length > 1" v-model="selectedContainer" placeholder="All Containers" clearable style="width: 200px;" @change="fetchLogs">
                <el-option v-for="c in pod.containers" :key="c.name" :label="c.name" :value="c.name" />
              </el-select>
              <el-select v-model="tailLines" style="width: 140px;" @change="fetchLogs">
                <el-option :value="50" label="Last 50 lines" />
                <el-option :value="100" label="Last 100 lines" />
                <el-option :value="200" label="Last 200 lines" />
                <el-option :value="500" label="Last 500 lines" />
                <el-option :value="1000" label="Last 1000 lines" />
              </el-select>
              <el-button type="primary" @click="fetchLogs" :loading="logsLoading">
                <el-icon><Refresh /></el-icon> Refresh
              </el-button>
              <el-button @click="handleFullLogViewer">Full Log Viewer</el-button>
            </div>
            <div v-loading="logsLoading" class="log-container">
              <pre v-if="logs" class="log-content">{{ logs }}</pre>
              <el-empty v-else description="No logs available" />
            </div>
          </el-card>
        </el-tab-pane>

        <!-- YAML Tab -->
        <el-tab-pane label="YAML" name="yaml">
          <el-card shadow="never">
            <div style="margin-bottom: 12px; display: flex; gap: 8px;">
              <el-button v-if="!yamlEditing" type="primary" @click="yamlEditing = true">Edit YAML</el-button>
              <template v-if="yamlEditing">
                <el-button type="success" :loading="yamlSaving" @click="handleSaveYaml">保存</el-button>
                <el-button @click="yamlEditing = false; fetchYaml()">取消</el-button>
              </template>
            </div>
            <div v-loading="yamlLoading">
              <YamlEditor v-model="yamlContent" height="600px" :read-only="!yamlEditing" />
            </div>
          </el-card>
        </el-tab-pane>
      </el-tabs>
    </template>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.log-container {
  background: #1e1e1e;
  border-radius: 4px;
  padding: 16px;
  max-height: 500px;
  overflow-y: auto;
}
.log-content {
  color: #d4d4d4;
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 12px;
  line-height: 1.6;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
