<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Refresh, Search } from '@element-plus/icons-vue'
import request from '@/api/request'
import { getDeploymentList, getStatefulSetList, getDaemonSetList, getNamespaceList } from '@/api/resource'

const { t } = useI18n()
const loading = ref(false)
const selectedNamespace = ref('')
const namespaceList = ref<string[]>([])
const selectedKind = ref('Deployment')
const selectedName = ref('')
const resourceList = ref<any[]>([])
const topologyData = ref<any>(null)

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
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load topology')
    topologyData.value = null
  } finally { loading.value = false }
}

function handleKindChange() {
  selectedName.value = ''
  topologyData.value = null
  fetchResources()
}

function handleNameChange() {
  fetchTopology()
}

function podStatusType(status: string) {
  const s = (status || '').toLowerCase()
  if (s === 'running') return 'success'
  if (s === 'succeeded') return 'info'
  if (s === 'pending') return 'warning'
  if (s === 'failed') return 'danger'
  return 'info'
}

onMounted(() => { fetchNamespaces() })
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

    <div v-loading="loading">
      <!-- Deployment Topology -->
      <template v-if="topologyData && selectedKind === 'Deployment'">
        <el-card shadow="never" class="topology-card">
          <div class="resource-node deployment-node">
            <div class="node-icon">D</div>
            <div class="node-info">
              <div class="node-name">{{ topologyData.deployment?.name }}</div>
              <div class="node-detail">Replicas: {{ topologyData.deployment?.ready }}/{{ topologyData.deployment?.replicas }}</div>
            </div>
          </div>
        </el-card>

        <div v-for="(rs, i) in topologyData.replicaSets" :key="i" style="margin-left: 40px;">
          <div class="connector">└─</div>
          <el-card shadow="never" class="topology-card">
            <div class="resource-node rs-node">
              <div class="node-icon">RS</div>
              <div class="node-info">
                <div class="node-name">{{ rs.name }}</div>
                <div class="node-detail">Revision: {{ rs.revision }} | Ready: {{ rs.ready }}/{{ rs.replicas }}</div>
              </div>
            </div>
          </el-card>

          <div v-for="(pod, j) in rs.pods" :key="j" style="margin-left: 80px;">
            <div class="connector">└─</div>
            <el-card shadow="never" class="topology-card">
              <div class="resource-node pod-node">
                <div class="node-icon">P</div>
                <div class="node-info">
                  <div class="node-name">{{ pod.name }}</div>
                  <div class="node-detail">
                    <el-tag :type="podStatusType(pod.status)" size="small">{{ pod.status }}</el-tag>
                    <span v-if="pod.ip" style="margin-left: 8px;">IP: {{ pod.ip }}</span>
                  </div>
                </div>
              </div>
            </el-card>
          </div>
        </div>

        <el-empty v-if="!topologyData.replicaSets || topologyData.replicaSets.length === 0" description="No ReplicaSets found" />
      </template>

      <!-- StatefulSet/DaemonSet Topology -->
      <template v-if="topologyData && (selectedKind === 'StatefulSet' || selectedKind === 'DaemonSet')">
        <el-card shadow="never" class="topology-card">
          <div class="resource-node sts-node">
            <div class="node-icon">{{ selectedKind === 'StatefulSet' ? 'STS' : 'DS' }}</div>
            <div class="node-info">
              <div class="node-name">{{ topologyData.statefulSet?.name || topologyData.daemonSet?.name }}</div>
              <div class="node-detail">
                Ready: {{ topologyData.statefulSet?.ready || topologyData.daemonSet?.readyNumber }}/{{ topologyData.statefulSet?.replicas || topologyData.daemonSet?.desiredNumber }}
              </div>
            </div>
          </div>
        </el-card>

        <div v-for="(pod, i) in topologyData.pods" :key="i" style="margin-left: 40px;">
          <div class="connector">└─</div>
          <el-card shadow="never" class="topology-card">
            <div class="resource-node pod-node">
              <div class="node-icon">P</div>
              <div class="node-info">
                <div class="node-name">{{ pod.name }}</div>
                <div class="node-detail">
                  <el-tag :type="podStatusType(pod.status)" size="small">{{ pod.status }}</el-tag>
                  <span v-if="pod.ip" style="margin-left: 8px;">IP: {{ pod.ip }}</span>
                  <span v-if="pod.node" style="margin-left: 8px;">Node: {{ pod.node }}</span>
                </div>
              </div>
            </div>
          </el-card>
        </div>

        <el-empty v-if="!topologyData.pods || topologyData.pods.length === 0" description="No Pods found" />
      </template>

      <el-empty v-if="!topologyData && !loading" description="Select a resource to view its topology" />
    </div>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; gap: 12px; align-items: center; flex-wrap: wrap; }
.topology-card { margin-bottom: 8px; }
.topology-card :deep(.el-card__body) { padding: 12px 16px; }
.resource-node { display: flex; align-items: center; gap: 12px; }
.node-icon { width: 36px; height: 36px; border-radius: 8px; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 12px; color: #fff; }
.deployment-node .node-icon { background: #409EFF; }
.rs-node .node-icon { background: #67C23A; }
.sts-node .node-icon { background: #E6A23C; }
.pod-node .node-icon { background: #909399; }
.node-name { font-weight: 600; font-size: 14px; }
.node-detail { font-size: 12px; color: #909399; margin-top: 2px; }
.connector { color: #909399; font-size: 14px; margin-left: 16px; }
</style>
