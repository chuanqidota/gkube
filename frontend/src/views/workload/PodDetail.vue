<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getPodDetail, getPodYaml, getPodEvents } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const pod = ref<any>(null)
const events = ref<any[]>([])
const yamlContent = ref('')
const yamlLoading = ref(false)
const activeTab = ref('info')

const namespace = route.params.namespace as string
const name = route.params.name as string
const clusterName = (route.query.cluster as string) || ''

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getPodDetail({ clusterName, namespace, name })
    pod.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load pod detail')
  } finally {
    loading.value = false
  }
}

async function fetchEvents() {
  try {
    const res: any = await getPodEvents({ clusterName, namespace, name })
    events.value = res.data || []
  } catch {
    // ignore
  }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getPodYaml({ clusterName, namespace, name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load YAML')
  } finally {
    yamlLoading.value = false
  }
}

function handleTabChange(tab: string) {
  if (tab === 'yaml' && !yamlContent.value) {
    fetchYaml()
  }
  if (tab === 'events' && events.value.length === 0) {
    fetchEvents()
  }
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
  const s = (status || '').toLowerCase()
  if (s === 'true') return 'success'
  if (s === 'false') return 'danger'
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

onMounted(fetchDetail)
</script>

<template>
  <div v-loading="loading">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="margin: 0;">Pod: {{ name }}</h2>
      <el-button @click="router.push('/workloads/pods')">Back to List</el-button>
    </div>

    <template v-if="pod">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <!-- Info Tab -->
        <el-tab-pane label="Info" name="info">
          <el-descriptions :column="2" border style="margin-top: 8px;">
            <el-descriptions-item label="Name">{{ pod.name }}</el-descriptions-item>
            <el-descriptions-item label="Namespace">{{ pod.namespace }}</el-descriptions-item>
            <el-descriptions-item label="Status">
              <el-tag :type="statusType(pod.status)" size="small">{{ pod.status }}</el-tag>
            </el-descriptions-item>
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

          <!-- Labels -->
          <div v-if="pod.labels && Object.keys(pod.labels).length > 0" style="margin-top: 16px;">
            <h4>Labels</h4>
            <el-tag
              v-for="(val, key) in pod.labels"
              :key="key"
              style="margin-right: 8px; margin-bottom: 8px;"
            >
              {{ key }}={{ val }}
            </el-tag>
          </div>

          <!-- Conditions -->
          <div v-if="pod.conditions && pod.conditions.length > 0" style="margin-top: 16px;">
            <h4>Conditions</h4>
            <el-table :data="pod.conditions" border stripe>
              <el-table-column prop="type" label="Type" min-width="160" />
              <el-table-column label="Status" width="120">
                <template #default="{ row }">
                  <el-tag :type="conditionStatusType(row.status)" size="small">{{ row.status }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="reason" label="Reason" min-width="160" show-overflow-tooltip />
              <el-table-column prop="message" label="Message" min-width="250" show-overflow-tooltip />
              <el-table-column prop="last_transition_time" label="Last Transition" width="180" />
            </el-table>
          </div>
        </el-tab-pane>

        <!-- Containers Tab -->
        <el-tab-pane label="Containers" name="containers">
          <el-table
            :data="pod.containers || []"
            border
            stripe
            style="margin-top: 8px;"
            row-key="name"
          >
            <el-table-column type="expand">
              <template #default="{ row }">
                <div style="padding: 12px 16px;">
                  <!-- Ports -->
                  <div v-if="row.ports && row.ports.length > 0" style="margin-bottom: 16px;">
                    <h4 style="margin: 0 0 8px 0;">Ports</h4>
                    <el-table :data="row.ports" border size="small">
                      <el-table-column prop="name" label="Name" width="120" />
                      <el-table-column prop="containerPort" label="Container Port" width="130" />
                      <el-table-column prop="protocol" label="Protocol" width="100" />
                      <el-table-column prop="hostPort" label="Host Port" width="100" />
                    </el-table>
                  </div>

                  <!-- Environment Variables -->
                  <div v-if="row.env && row.env.length > 0" style="margin-bottom: 16px;">
                    <h4 style="margin: 0 0 8px 0;">Environment Variables</h4>
                    <el-table :data="row.env" border size="small">
                      <el-table-column prop="name" label="Name" min-width="180" />
                      <el-table-column label="Value" min-width="250">
                        <template #default="{ row: envRow }">
                          <span v-if="envRow.value !== undefined && envRow.value !== ''">{{ envRow.value }}</span>
                          <span v-else-if="envRow.valueFrom" style="color: #909399;">
                            {{ envRow.valueFrom.fieldRef?.fieldPath || envRow.valueFrom.secretKeyRef?.name || envRow.valueFrom.configMapKeyRef?.name || 'From reference' }}
                          </span>
                          <span v-else style="color: #909399;">-</span>
                        </template>
                      </el-table-column>
                    </el-table>
                  </div>

                  <!-- Volume Mounts -->
                  <div v-if="row.volumeMounts && row.volumeMounts.length > 0" style="margin-bottom: 16px;">
                    <h4 style="margin: 0 0 8px 0;">Volume Mounts</h4>
                    <el-table :data="row.volumeMounts" border size="small">
                      <el-table-column prop="name" label="Volume Name" min-width="150" />
                      <el-table-column prop="mountPath" label="Mount Path" min-width="200" />
                      <el-table-column prop="subPath" label="Sub Path" width="150" />
                      <el-table-column label="Read Only" width="100">
                        <template #default="{ row: vm }">
                          <el-tag :type="vm.readOnly ? 'warning' : 'success'" size="small">
                            {{ vm.readOnly ? 'Yes' : 'No' }}
                          </el-tag>
                        </template>
                      </el-table-column>
                    </el-table>
                  </div>

                  <!-- Liveness Probe -->
                  <div v-if="row.livenessProbe" style="margin-bottom: 16px;">
                    <h4 style="margin: 0 0 8px 0;">Liveness Probe</h4>
                    <el-descriptions :column="2" border size="small">
                      <el-descriptions-item v-if="row.livenessProbe.httpGet" label="Type">HTTP GET</el-descriptions-item>
                      <el-descriptions-item v-if="row.livenessProbe.httpGet" label="Path">{{ row.livenessProbe.httpGet.path || '/' }}</el-descriptions-item>
                      <el-descriptions-item v-if="row.livenessProbe.httpGet" label="Port">{{ row.livenessProbe.httpGet.port }}</el-descriptions-item>
                      <el-descriptions-item v-if="row.livenessProbe.tcpSocket" label="Type">TCP Socket</el-descriptions-item>
                      <el-descriptions-item v-if="row.livenessProbe.tcpSocket" label="Port">{{ row.livenessProbe.tcpSocket.port }}</el-descriptions-item>
                      <el-descriptions-item v-if="row.livenessProbe.exec" label="Type">Exec</el-descriptions-item>
                      <el-descriptions-item v-if="row.livenessProbe.exec" label="Command">{{ (row.livenessProbe.exec.command || []).join(' ') }}</el-descriptions-item>
                      <el-descriptions-item label="Initial Delay">{{ row.livenessProbe.initialDelaySeconds ?? '-' }}s</el-descriptions-item>
                      <el-descriptions-item label="Period">{{ row.livenessProbe.periodSeconds ?? '-' }}s</el-descriptions-item>
                      <el-descriptions-item label="Timeout">{{ row.livenessProbe.timeoutSeconds ?? '-' }}s</el-descriptions-item>
                      <el-descriptions-item label="Failure Threshold">{{ row.livenessProbe.failureThreshold ?? '-' }}</el-descriptions-item>
                    </el-descriptions>
                  </div>

                  <!-- Readiness Probe -->
                  <div v-if="row.readinessProbe">
                    <h4 style="margin: 0 0 8px 0;">Readiness Probe</h4>
                    <el-descriptions :column="2" border size="small">
                      <el-descriptions-item v-if="row.readinessProbe.httpGet" label="Type">HTTP GET</el-descriptions-item>
                      <el-descriptions-item v-if="row.readinessProbe.httpGet" label="Path">{{ row.readinessProbe.httpGet.path || '/' }}</el-descriptions-item>
                      <el-descriptions-item v-if="row.readinessProbe.httpGet" label="Port">{{ row.readinessProbe.httpGet.port }}</el-descriptions-item>
                      <el-descriptions-item v-if="row.readinessProbe.tcpSocket" label="Type">TCP Socket</el-descriptions-item>
                      <el-descriptions-item v-if="row.readinessProbe.tcpSocket" label="Port">{{ row.readinessProbe.tcpSocket.port }}</el-descriptions-item>
                      <el-descriptions-item v-if="row.readinessProbe.exec" label="Type">Exec</el-descriptions-item>
                      <el-descriptions-item v-if="row.readinessProbe.exec" label="Command">{{ (row.readinessProbe.exec.command || []).join(' ') }}</el-descriptions-item>
                      <el-descriptions-item label="Initial Delay">{{ row.readinessProbe.initialDelaySeconds ?? '-' }}s</el-descriptions-item>
                      <el-descriptions-item label="Period">{{ row.readinessProbe.periodSeconds ?? '-' }}s</el-descriptions-item>
                      <el-descriptions-item label="Timeout">{{ row.readinessProbe.timeoutSeconds ?? '-' }}s</el-descriptions-item>
                      <el-descriptions-item label="Failure Threshold">{{ row.readinessProbe.failureThreshold ?? '-' }}</el-descriptions-item>
                    </el-descriptions>
                  </div>

                  <el-empty
                    v-if="(!row.ports || row.ports.length === 0) && (!row.env || row.env.length === 0) && (!row.volumeMounts || row.volumeMounts.length === 0) && !row.livenessProbe && !row.readinessProbe"
                    description="No additional container details available"
                    :image-size="60"
                  />
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="name" label="Name" min-width="160" />
            <el-table-column prop="image" label="Image" min-width="280" show-overflow-tooltip />
            <el-table-column label="Ready" width="80">
              <template #default="{ row }">
                <el-tag :type="row.ready ? 'success' : 'danger'" size="small">
                  {{ row.ready ? 'Yes' : 'No' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="restartCount" label="Restarts" width="100" />
            <el-table-column label="State" width="180">
              <template #default="{ row }">
                <el-tag :type="containerStateType(row.state)" size="small">
                  {{ getContainerStateLabel(row) }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="!pod.containers || pod.containers.length === 0" description="No containers" />
        </el-tab-pane>

        <!-- Events Tab -->
        <el-tab-pane label="Events" name="events">
          <el-table :data="events" border stripe style="margin-top: 8px;">
            <el-table-column prop="type" label="Type" width="100" />
            <el-table-column prop="reason" label="Reason" width="150" />
            <el-table-column prop="message" label="Message" min-width="300" show-overflow-tooltip />
            <el-table-column prop="last_seen" label="Last Seen" width="180" />
          </el-table>
          <el-empty v-if="events.length === 0" description="No events" />
        </el-tab-pane>

        <!-- YAML Tab -->
        <el-tab-pane label="YAML" name="yaml">
          <div v-loading="yamlLoading">
            <YamlEditor v-model="yamlContent" height="600px" read-only />
          </div>
        </el-tab-pane>
      </el-tabs>
    </template>
  </div>
</template>
