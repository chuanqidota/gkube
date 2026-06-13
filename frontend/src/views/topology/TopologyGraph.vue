<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import request from '@/api/request'
import { getDeploymentList, getStatefulSetList, getDaemonSetList, getNamespaceList } from '@/api/resource'
import { Network } from 'vis-network'
import { DataSet } from 'vis-data'

const loading = ref(false)
const selectedNamespace = ref('')
const namespaceList = ref<string[]>([])
const selectedKind = ref('Deployment')
const selectedName = ref('')
const resourceList = ref<any[]>([])
const topologyData = ref<any>(null)
const graphContainer = ref<HTMLElement>()
let network: Network | null = null

async function fetchNamespaces() {
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = (res.data || []).map((ns: any) => ns.name || ns)
  } catch { /* ignore */ }
}

async function fetchResources() {
  try {
    const params: any = {}
    if (selectedNamespace.value) params.namespace = selectedNamespace.value
    let res: any
    switch (selectedKind.value) {
      case 'Deployment': res = await getDeploymentList(params); break
      case 'StatefulSet': res = await getStatefulSetList(params); break
      case 'DaemonSet': res = await getDaemonSetList(params); break
    }
    resourceList.value = res?.data || []
  } catch { resourceList.value = [] }
}

async function fetchTopology() {
  if (!selectedName.value) return
  loading.value = true
  try {
    const params: any = { name: selectedName.value }
    if (selectedNamespace.value) params.namespace = selectedNamespace.value
    let endpoint = ''
    switch (selectedKind.value) {
      case 'Deployment': endpoint = '/k8s/topology/deployment'; break
      case 'StatefulSet': endpoint = '/k8s/topology/statefulset'; break
      case 'DaemonSet': endpoint = '/k8s/topology/daemonset'; break
    }
    const res: any = await request.get(endpoint, { params })
    topologyData.value = res.data
    await nextTick()
    renderGraph()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load topology')
    topologyData.value = null
  } finally { loading.value = false }
}

function renderGraph() {
  if (!graphContainer.value || !topologyData.value) return

  const nodes = new DataSet<any>([])
  const edges = new DataSet<any>([])

  if (selectedKind.value === 'Deployment' && topologyData.value.deployment) {
    const deploy = topologyData.value.deployment
    nodes.add({
      id: `deploy-${deploy.name}`,
      label: `Deployment\n${deploy.name}`,
      shape: 'box',
      color: { background: '#409EFF', border: '#337ecc' },
      font: { color: '#fff', size: 14 },
      margin: 10,
    })

    topologyData.value.replicaSets?.forEach((rs: any, _i: number) => {
      const rsId = `rs-${rs.name}`
      nodes.add({
        id: rsId,
        label: `ReplicaSet\n${rs.name}\nRev: ${rs.revision || '-'}`,
        shape: 'box',
        color: { background: '#67C23A', border: '#529b2e' },
        font: { color: '#fff', size: 12 },
        margin: 10,
      })
      edges.add({ from: `deploy-${deploy.name}`, to: rsId, arrows: 'to', color: { color: '#909399' } })

      rs.pods?.forEach((pod: any) => {
        const podId = `pod-${pod.name}`
        const statusColor = pod.status === 'Running' ? '#67C23A' : pod.status === 'Pending' ? '#E6A23C' : '#F56C6C'
        nodes.add({
          id: podId,
          label: `Pod\n${pod.name}\n${pod.status}`,
          shape: 'ellipse',
          color: { background: statusColor, border: '#606266' },
          font: { color: '#fff', size: 11 },
          margin: 8,
        })
        edges.add({ from: rsId, to: podId, arrows: 'to', color: { color: '#C0C4CC' } })
      })
    })
  } else if ((selectedKind.value === 'StatefulSet' || selectedKind.value === 'DaemonSet') && (topologyData.value.statefulSet || topologyData.value.daemonSet)) {
    const resource = topologyData.value.statefulSet || topologyData.value.daemonSet
    const kind = selectedKind.value
    const color = kind === 'StatefulSet' ? '#E6A23C' : '#909399'
    nodes.add({
      id: `resource-${resource.name}`,
      label: `${kind}\n${resource.name}`,
      shape: 'box',
      color: { background: color, border: '#606266' },
      font: { color: '#fff', size: 14 },
      margin: 10,
    })

    topologyData.value.pods?.forEach((pod: any) => {
      const podId = `pod-${pod.name}`
      const statusColor = pod.status === 'Running' ? '#67C23A' : pod.status === 'Pending' ? '#E6A23C' : '#F56C6C'
      nodes.add({
        id: podId,
        label: `Pod\n${pod.name}\n${pod.status}\nNode: ${pod.node || '-'}`,
        shape: 'ellipse',
        color: { background: statusColor, border: '#606266' },
        font: { color: '#fff', size: 11 },
        margin: 8,
      })
      edges.add({ from: `resource-${resource.name}`, to: podId, arrows: 'to', color: { color: '#C0C4CC' } })
    })
  }

  const options = {
    layout: {
      hierarchical: {
        direction: 'UD',
        sortMethod: 'directed',
        levelSeparation: 150,
        nodeSpacing: 200,
      },
    },
    edges: {
      smooth: { type: 'cubicBezier' },
      width: 2,
    },
    physics: { enabled: false },
    interaction: { hover: true, zoomView: true, dragView: true },
  }

  if (network) {
    network.destroy()
  }
  network = new Network(graphContainer.value, { nodes, edges }, options)
}

function handleKindChange() {
  selectedName.value = ''
  topologyData.value = null
  fetchResources()
}

function handleNameChange() {
  fetchTopology()
}

onMounted(() => { fetchNamespaces() })
onUnmounted(() => { if (network) network.destroy() })
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <h3 style="margin: 0;">Resource Topology</h3>
        <el-select v-model="selectedKind" style="width: 160px;" @change="handleKindChange">
          <el-option label="Deployment" value="Deployment" />
          <el-option label="StatefulSet" value="StatefulSet" />
          <el-option label="DaemonSet" value="DaemonSet" />
        </el-select>
        <el-select v-model="selectedNamespace" placeholder="Namespace" clearable style="width: 160px;" @change="handleKindChange">
          <el-option v-for="ns in namespaceList" :key="ns" :label="ns" :value="ns" />
        </el-select>
        <el-select v-model="selectedName" placeholder="Select resource" style="width: 250px;" filterable @change="handleNameChange">
          <el-option v-for="r in resourceList" :key="r.name" :label="r.name" :value="r.name" />
        </el-select>
        <el-button type="primary" @click="fetchTopology" :disabled="!selectedName"><el-icon><Refresh /></el-icon> Refresh</el-button>
      </div>
    </el-card>

    <el-card shadow="never" v-loading="loading">
      <div ref="graphContainer" style="width: 100%; height: 600px; border: 1px solid #e4e7ed; border-radius: 4px;"></div>
      <el-empty v-if="!topologyData && !loading" description="Select a resource to view its topology graph" />
    </el-card>

    <!-- Legend -->
    <el-card shadow="never" style="margin-top: 16px;">
      <div style="display: flex; gap: 24px; align-items: center;">
        <span style="font-weight: 600;">Legend:</span>
        <div style="display: flex; align-items: center; gap: 8px;">
          <div style="width: 16px; height: 16px; background: #409EFF; border-radius: 3px;"></div>
          <span>Deployment</span>
        </div>
        <div style="display: flex; align-items: center; gap: 8px;">
          <div style="width: 16px; height: 16px; background: #67C23A; border-radius: 3px;"></div>
          <span>ReplicaSet / Running Pod</span>
        </div>
        <div style="display: flex; align-items: center; gap: 8px;">
          <div style="width: 16px; height: 16px; background: #E6A23C; border-radius: 3px;"></div>
          <span>StatefulSet / Pending Pod</span>
        </div>
        <div style="display: flex; align-items: center; gap: 8px;">
          <div style="width: 16px; height: 16px; background: #F56C6C; border-radius: 3px;"></div>
          <span>Failed Pod</span>
        </div>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; gap: 12px; align-items: center; flex-wrap: wrap; }
</style>
