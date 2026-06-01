<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getNodeList, getNodeYaml, cordonNode, taintNode } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'
import YamlEditor from '@/components/YamlEditor.vue'

const router = useRouter()
const clusterStore = useClusterStore()
const loading = ref(false)
const nodeList = ref<any[]>([])
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)
const taintDialogVisible = ref(false)
const taintTarget = ref<any>(null)
const taints = ref<any[]>([])

async function fetchNodes() {
  loading.value = true
  try {
    const clusterId = clusterStore.currentCluster?.id
    const res: any = await getNodeList({ cluster_id: clusterId })
    nodeList.value = res.data || []
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load nodes')
  } finally {
    loading.value = false
  }
}

function statusType(node: any) {
  if (node.status === 'Ready') return 'success'
  if (node.status === 'NotReady') return 'danger'
  return 'warning'
}

async function handleViewYaml(row: any) {
  yamlLoading.value = true
  yamlDialogVisible.value = true
  try {
    const clusterId = clusterStore.currentCluster?.id
    const res: any = await getNodeYaml({ name: row.name, cluster_id: clusterId })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load YAML')
    yamlDialogVisible.value = false
  } finally {
    yamlLoading.value = false
  }
}

async function handleCordon(row: any) {
  const isCordon = row.unschedulable || row.cordon
  const action = isCordon ? 'uncordon' : 'cordon'
  try {
    await ElMessageBox.confirm(
      `${action.charAt(0).toUpperCase() + action.slice(1)} node "${row.name}"?`,
      'Confirm',
      { type: 'warning' }
    )
    const clusterId = clusterStore.currentCluster?.id
    await cordonNode({ name: row.name, cordon: !isCordon, cluster_id: clusterId })
    ElMessage.success(`Node ${action}ed`)
    fetchNodes()
  } catch {
    // cancelled
  }
}

function handleTaints(row: any) {
  taintTarget.value = row
  taints.value = (row.taints || []).map((t: any) => ({ ...t }))
  if (taints.value.length === 0) {
    taints.value = [{ key: '', value: '', effect: 'NoSchedule' }]
  }
  taintDialogVisible.value = true
}

function addTaint() {
  taints.value.push({ key: '', value: '', effect: 'NoSchedule' })
}

function removeTaint(index: number) {
  taints.value.splice(index, 1)
}

async function handleSaveTaints() {
  try {
    const clusterId = clusterStore.currentCluster?.id
    await taintNode({
      name: taintTarget.value.name,
      taints: taints.value.filter(t => t.key),
      cluster_id: clusterId,
    })
    ElMessage.success('Taints updated')
    taintDialogVisible.value = false
    fetchNodes()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to update taints')
  }
}

onMounted(fetchNodes)
</script>

<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="margin: 0;">Nodes</h2>
      <el-button @click="fetchNodes" :loading="loading">Refresh</el-button>
    </div>

    <el-table :data="nodeList" v-loading="loading" stripe style="width: 100%">
      <el-table-column prop="name" label="Name" min-width="180" show-overflow-tooltip />
      <el-table-column label="Status" width="100">
        <template #default="{ row }">
          <el-tag :type="statusType(row)" size="small">{{ row.status || 'Unknown' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="roles" label="Roles" min-width="120" />
      <el-table-column prop="age" label="Age" min-width="100" />
      <el-table-column prop="version" label="Version" min-width="120" />
      <el-table-column prop="internal_ip" label="Internal IP" min-width="140" />
      <el-table-column label="Actions" width="260" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="router.push(`/nodes/${row.name}`)">Detail</el-button>
          <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
          <el-button
            size="small"
            :type="row.unschedulable || row.cordon ? 'success' : 'warning'"
            @click="handleCordon(row)"
          >
            {{ row.unschedulable || row.cordon ? 'Uncordon' : 'Cordon' }}
          </el-button>
          <el-button size="small" type="info" @click="handleTaints(row)">Taints</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="yamlDialogVisible" title="Node YAML" width="70%" top="5vh">
      <YamlEditor v-if="!yamlLoading" v-model="yamlContent" :read-only="true" height="500px" />
      <div v-else v-loading="true" style="height: 200px;"></div>
    </el-dialog>

    <el-dialog v-model="taintDialogVisible" title="Manage Taints" width="600px">
      <div v-for="(taint, index) in taints" :key="index" style="display: flex; gap: 8px; margin-bottom: 12px; align-items: center;">
        <el-input v-model="taint.key" placeholder="Key" style="flex: 2;" />
        <el-input v-model="taint.value" placeholder="Value" style="flex: 1;" />
        <el-select v-model="taint.effect" style="flex: 1.5;">
          <el-option label="NoSchedule" value="NoSchedule" />
          <el-option label="PreferNoSchedule" value="PreferNoSchedule" />
          <el-option label="NoExecute" value="NoExecute" />
        </el-select>
        <el-button type="danger" :icon="'Delete'" circle size="small" @click="removeTaint(index)" />
      </div>
      <el-button @click="addTaint" style="margin-top: 8px;">Add Taint</el-button>
      <template #footer>
        <el-button @click="taintDialogVisible = false">Cancel</el-button>
        <el-button type="primary" @click="handleSaveTaints">Save</el-button>
      </template>
    </el-dialog>
  </div>
</template>
