<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getDeploymentList,
  getDeploymentYaml,
  scaleDeployment,
  restartDeployment,
  deleteDeployment,
  getNamespaceList,
} from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const router = useRouter()
const loading = ref(false)
const deploymentList = ref<any[]>([])
const namespaceList = ref<string[]>([])
const selectedNamespace = ref('')
const clusterName = ref('')

// YAML dialog
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)

// Scale dialog
const scaleDialogVisible = ref(false)
const scaleTarget = ref<any>(null)
const scaleReplicas = ref(1)
const scaleLoading = ref(false)

async function fetchNamespaces() {
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = (res.data || []).map((ns: any) => ns.name || ns)
  } catch {
    // ignore
  }
}

async function fetchDeployments() {
  loading.value = true
  try {
    const params: any = {}
    if (selectedNamespace.value) params.namespace = selectedNamespace.value
    if (clusterName.value) params.clusterName = clusterName.value
    const res: any = await getDeploymentList(params)
    deploymentList.value = res.data || []
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load deployments')
  } finally {
    loading.value = false
  }
}

function handleNamespaceChange() {
  fetchDeployments()
}

async function handleViewYaml(row: any) {
  yamlDialogVisible.value = true
  yamlLoading.value = true
  yamlContent.value = ''
  try {
    const res: any = await getDeploymentYaml({
      clusterName: clusterName.value,
      namespace: row.namespace,
      name: row.name,
    })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load YAML')
    yamlDialogVisible.value = false
  } finally {
    yamlLoading.value = false
  }
}

function handleScale(row: any) {
  scaleTarget.value = row
  // Parse ready replicas (e.g., "2/3" -> 3)
  const readyStr = row.ready || '0'
  const parts = readyStr.split('/')
  scaleReplicas.value = parseInt(parts[1] || parts[0]) || 1
  scaleDialogVisible.value = true
}

async function handleScaleConfirm() {
  if (!scaleTarget.value) return
  scaleLoading.value = true
  try {
    await scaleDeployment({
      clusterName: clusterName.value,
      namespace: scaleTarget.value.namespace,
      name: scaleTarget.value.name,
      replicas: scaleReplicas.value,
    })
    ElMessage.success(`Scaled to ${scaleReplicas.value} replicas`)
    scaleDialogVisible.value = false
    fetchDeployments()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to scale')
  } finally {
    scaleLoading.value = false
  }
}

async function handleRestart(row: any) {
  try {
    await ElMessageBox.confirm(
      `Restart deployment "${row.name}"?`,
      'Confirm',
      { type: 'warning' }
    )
    await restartDeployment({
      clusterName: clusterName.value,
      namespace: row.namespace,
      name: row.name,
    })
    ElMessage.success('Deployment restarted')
    fetchDeployments()
  } catch {
    // cancelled
  }
}

function handleDetail(row: any) {
  router.push(`/workloads/deployments/${row.namespace}/${row.name}`)
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(
      `Delete deployment "${row.name}" in namespace "${row.namespace}"?`,
      'Confirm',
      { type: 'warning' }
    )
    await deleteDeployment({
      clusterName: clusterName.value,
      namespace: row.namespace,
      name: row.name,
    })
    ElMessage.success('Deployment deleted')
    fetchDeployments()
  } catch {
    // cancelled
  }
}

onMounted(() => {
  fetchNamespaces()
  fetchDeployments()
})
</script>

<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="margin: 0;">Deployments</h2>
      <div style="display: flex; gap: 12px; align-items: center;">
        <el-input
          v-model="clusterName"
          placeholder="Cluster Name"
          style="width: 180px;"
          clearable
          @clear="fetchDeployments"
          @keyup.enter="fetchDeployments"
        />
        <el-select
          v-model="selectedNamespace"
          placeholder="All Namespaces"
          clearable
          style="width: 180px;"
          @change="handleNamespaceChange"
        >
          <el-option
            v-for="ns in namespaceList"
            :key="ns"
            :label="ns"
            :value="ns"
          />
        </el-select>
        <el-button type="primary" @click="fetchDeployments">Refresh</el-button>
        <el-button type="success" @click="router.push('/workloads/deployments/create')">Create</el-button>
      </div>
    </div>

    <el-table :data="deploymentList" v-loading="loading" stripe style="width: 100%">
      <el-table-column prop="name" label="Name" min-width="200" show-overflow-tooltip>
        <template #default="{ row }">
          <el-button link type="primary" @click="handleDetail(row)">{{ row.name }}</el-button>
        </template>
      </el-table-column>
      <el-table-column prop="namespace" label="Namespace" width="140" />
      <el-table-column prop="ready" label="Ready" width="100" />
      <el-table-column prop="up_to_date" label="Up-to-date" width="110" />
      <el-table-column prop="available" label="Available" width="110" />
      <el-table-column prop="age" label="Age" width="120" />
      <el-table-column label="Actions" width="320" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
          <el-button size="small" type="warning" @click="handleScale(row)">Scale</el-button>
          <el-button size="small" type="success" @click="handleRestart(row)">Restart</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">Delete</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- YAML Dialog -->
    <el-dialog v-model="yamlDialogVisible" title="Deployment YAML" width="70%" top="5vh" destroy-on-close>
      <div v-loading="yamlLoading">
        <YamlEditor v-model="yamlContent" height="500px" read-only />
      </div>
    </el-dialog>

    <!-- Scale Dialog -->
    <el-dialog v-model="scaleDialogVisible" title="Scale Deployment" width="400px" destroy-on-close>
      <div v-if="scaleTarget">
        <p style="margin-bottom: 12px;">
          Deployment: <strong>{{ scaleTarget.name }}</strong>
        </p>
        <el-form-item label="Replicas">
          <el-input-number v-model="scaleReplicas" :min="0" :max="100" />
        </el-form-item>
      </div>
      <template #footer>
        <el-button @click="scaleDialogVisible = false">Cancel</el-button>
        <el-button type="primary" :loading="scaleLoading" @click="handleScaleConfirm">Scale</el-button>
      </template>
    </el-dialog>
  </div>
</template>
