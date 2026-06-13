<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getNodeDetail, getNodeYaml, getNodePods, getNodeEvents, cordonNode, taintNode, updateNodeYaml } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const node = ref<any>(null)
const pods = ref<any[]>([])
const podsLoading = ref(false)
const events = ref<any[]>([])
const eventsLoading = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)
const yamlEditing = ref(false)
const yamlSaving = ref(false)
const activeTab = ref('info')
const taintDialogVisible = ref(false)
const taints = ref<any[]>([])

const nodeName = route.params.name as string

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getNodeDetail({ name: nodeName })
    node.value = res.data
    if (node.value) fetchPods()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load node detail')
  } finally {
    loading.value = false
  }
}

async function fetchPods() {
  podsLoading.value = true
  try {
    const res: any = await getNodePods({ name: nodeName })
    pods.value = res.data || []
  } catch { /* ignore */ }
  finally { podsLoading.value = false }
}

async function fetchEvents() {
  eventsLoading.value = true
  try {
    const res: any = await getNodeEvents({ name: nodeName })
    events.value = res.data || []
  } catch { /* ignore */ }
  finally { eventsLoading.value = false }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getNodeYaml({ name: nodeName })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load YAML')
  } finally { yamlLoading.value = false }
}

function handleTabChange(tab: string) {
  if (tab === 'yaml' && !yamlContent.value) fetchYaml()
  if (tab === 'events' && events.value.length === 0) fetchEvents()
}

async function handleSaveYaml() {
  yamlSaving.value = true
  try {
    await updateNodeYaml({ name: nodeName, yaml: yamlContent.value })
    ElMessage.success('YAML saved successfully')
    yamlEditing.value = false
    fetchDetail()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to save YAML')
  } finally { yamlSaving.value = false }
}

function statusType(status: string) {
  if (status === 'Ready') return 'success'
  if (status === 'NotReady') return 'danger'
  return 'warning'
}

async function handleCordon() {
  const isCordon = node.value?.unschedulable || node.value?.cordon
  const action = isCordon ? 'uncordon' : 'cordon'
  try {
    await ElMessageBox.confirm(`${action.charAt(0).toUpperCase() + action.slice(1)} node "${nodeName}"?`, 'Confirm', { type: 'warning' })
    await cordonNode({ name: nodeName, cordon: !isCordon })
    ElMessage.success(`Node ${action}ed`)
    fetchDetail()
  } catch { /* cancelled */ }
}

function handleTaints() {
  taints.value = (node.value?.taints || []).map((t: any) => ({ ...t }))
  if (taints.value.length === 0) taints.value = [{ key: '', value: '', effect: 'NoSchedule' }]
  taintDialogVisible.value = true
}

function addTaint() { taints.value.push({ key: '', value: '', effect: 'NoSchedule' }) }
function removeTaint(index: number) { taints.value.splice(index, 1) }

async function handleSaveTaints() {
  try {
    await taintNode({ name: nodeName, taints: taints.value.filter(t => t.key) })
    ElMessage.success('Taints updated')
    taintDialogVisible.value = false
    fetchDetail()
  } catch (e: any) { ElMessage.error(e?.message || 'Failed to update taints') }
}

function handlePodDetail(row: any) {
  router.push(`/workloads/pods/${row.namespace}/${row.name}`)
}

function podStatusType(status: string) {
  const s = (status || '').toLowerCase()
  if (s === 'running') return 'success'
  if (s === 'succeeded') return 'info'
  if (s === 'pending') return 'warning'
  if (s === 'failed') return 'danger'
  return 'info'
}

function formatCapacity(val: any): string {
  if (!val) return '-'
  return String(val)
}

onMounted(fetchDetail)
</script>

<template>
  <div class="page-container" v-loading="loading">
    <div class="page-header">
      <h2 style="margin: 0;">Node: {{ nodeName }}</h2>
      <div style="display: flex; gap: 8px;">
        <el-button :type="node?.unschedulable || node?.cordon ? 'success' : 'warning'" @click="handleCordon">
          {{ node?.unschedulable || node?.cordon ? 'Uncordon' : 'Cordon' }}
        </el-button>
        <el-button type="info" @click="handleTaints">Taints</el-button>
        <el-button @click="router.push('/nodes')">Back to List</el-button>
      </div>
    </div>

    <template v-if="node">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <!-- Info Tab -->
        <el-tab-pane label="Info" name="info">
          <el-card shadow="never">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="Name">{{ node.name }}</el-descriptions-item>
              <el-descriptions-item label="Status"><el-tag :type="statusType(node.status)" size="small">{{ node.status || 'Unknown' }}</el-tag></el-descriptions-item>
              <el-descriptions-item label="Roles">{{ node.roles || '-' }}</el-descriptions-item>
              <el-descriptions-item label="Version">{{ node.version || '-' }}</el-descriptions-item>
              <el-descriptions-item label="OS">{{ node.os || '-' }}</el-descriptions-item>
              <el-descriptions-item label="Kernel">{{ node.kernel || '-' }}</el-descriptions-item>
              <el-descriptions-item label="Container Runtime">{{ node.container_runtime || '-' }}</el-descriptions-item>
              <el-descriptions-item label="Internal IP">{{ node.internal_ip || '-' }}</el-descriptions-item>
              <el-descriptions-item label="Age">{{ node.age || '-' }}</el-descriptions-item>
              <el-descriptions-item label="Unschedulable">
                <el-tag :type="node.unschedulable || node.cordon ? 'danger' : 'success'" size="small">{{ node.unschedulable || node.cordon ? 'Yes' : 'No' }}</el-tag>
              </el-descriptions-item>
            </el-descriptions>

            <!-- Resource Capacity -->
            <div v-if="node.capacity || node.allocatable" style="margin-top: 24px;">
              <h4>Resources</h4>
              <el-table :data="[
                { resource: 'CPU', capacity: formatCapacity(node.capacity?.cpu), allocatable: formatCapacity(node.allocatable?.cpu) },
                { resource: 'Memory', capacity: formatCapacity(node.capacity?.memory), allocatable: formatCapacity(node.allocatable?.memory) },
                { resource: 'Pods', capacity: formatCapacity(node.capacity?.pods), allocatable: formatCapacity(node.allocatable?.pods) },
                { resource: 'Ephemeral Storage', capacity: formatCapacity(node.capacity?.['ephemeral-storage']), allocatable: formatCapacity(node.allocatable?.['ephemeral-storage']) },
              ]" border stripe>
                <el-table-column prop="resource" label="Resource" width="160" />
                <el-table-column prop="capacity" label="Capacity" min-width="150" />
                <el-table-column prop="allocatable" label="Allocatable" min-width="150" />
              </el-table>
            </div>

            <!-- Conditions -->
            <div v-if="node.conditions && node.conditions.length > 0" style="margin-top: 24px;">
              <h4>Conditions</h4>
              <el-table :data="node.conditions" border stripe>
                <el-table-column prop="type" label="Type" width="180" />
                <el-table-column label="Status" width="100">
                  <template #default="{ row }"><el-tag :type="row.status === 'True' ? 'success' : 'danger'" size="small">{{ row.status }}</el-tag></template>
                </el-table-column>
                <el-table-column prop="reason" label="Reason" width="180" />
                <el-table-column prop="message" label="Message" min-width="250" show-overflow-tooltip />
                <el-table-column prop="lastTransitionTime" label="Last Transition" width="180" />
              </el-table>
            </div>

            <!-- Labels -->
            <div v-if="node.labels && Object.keys(node.labels).length > 0" style="margin-top: 24px;">
              <h4>Labels</h4>
              <el-tag v-for="(val, key) in node.labels" :key="key" style="margin-right: 8px; margin-bottom: 8px;">{{ key }}={{ val }}</el-tag>
            </div>

            <!-- Taints -->
            <div v-if="node.taints && node.taints.length > 0" style="margin-top: 24px;">
              <h4>Taints</h4>
              <el-table :data="node.taints" stripe border>
                <el-table-column prop="key" label="Key" min-width="200" />
                <el-table-column prop="value" label="Value" min-width="120" />
                <el-table-column prop="effect" label="Effect" min-width="150" />
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
              <el-table-column prop="status" label="Status" width="120"><template #default="{ row }"><el-tag :type="podStatusType(row.status)" size="small">{{ row.status }}</el-tag></template></el-table-column>
              <el-table-column prop="ip" label="IP" width="140" />
              <el-table-column prop="restarts" label="Restarts" width="100" />
              <el-table-column prop="age" label="Age" width="120" />
            </el-table>
            <el-empty v-if="!podsLoading && pods.length === 0" description="No pods on this node" />
          </el-card>
        </el-tab-pane>

        <!-- Events Tab -->
        <el-tab-pane label="Events" name="events">
          <el-card shadow="never">
            <el-table :data="events" v-loading="eventsLoading" stripe>
              <el-table-column prop="type" label="Type" width="100"><template #default="{ row }"><el-tag :type="row.type === 'Warning' ? 'danger' : 'info'" size="small">{{ row.type }}</el-tag></template></el-table-column>
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

    <!-- Taints Dialog -->
    <el-dialog v-model="taintDialogVisible" title="Manage Taints" width="600px">
      <div v-for="(taint, index) in taints" :key="index" style="display: flex; gap: 8px; margin-bottom: 12px; align-items: center;">
        <el-input v-model="taint.key" placeholder="Key" style="flex: 2;" />
        <el-input v-model="taint.value" placeholder="Value" style="flex: 1;" />
        <el-select v-model="taint.effect" style="flex: 1.5;">
          <el-option label="NoSchedule" value="NoSchedule" />
          <el-option label="PreferNoSchedule" value="PreferNoSchedule" />
          <el-option label="NoExecute" value="NoExecute" />
        </el-select>
        <el-button type="danger" circle size="small" @click="removeTaint(index)">X</el-button>
      </div>
      <el-button @click="addTaint" style="margin-top: 8px;">Add Taint</el-button>
      <template #footer>
        <el-button @click="taintDialogVisible = false">Cancel</el-button>
        <el-button type="primary" @click="handleSaveTaints">Save</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
</style>
