<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getNodeList, type NodeInfo } from '@/api/resource'
import { usagePercent, progressColor, formatAge } from '@/utils/helpers'
import { formatCpuCores, formatMemGiB } from '@/utils/resource'
import YamlDrawer from '@/components/YamlDrawer.vue'
import NodeTaintDialog from '@/components/node/NodeTaintDialog.vue'
import NodeLabelDialog from '@/components/node/NodeLabelDialog.vue'
import NodeDrainDialog from '@/components/node/NodeDrainDialog.vue'
import { useNodeActions } from '@/composables/useNodeActions'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'
import ViewModeToggle from '@/components/ViewModeToggle.vue'

const router = useRouter()
const storedView = localStorage.getItem('gkube.node.viewMode')
const viewMode = ref<'card' | 'table'>(storedView === 'table' || storedView === 'card' ? storedView : 'card')
const loading = ref(false)
const nodeList = ref<NodeInfo[]>([])
const searchName = ref('')

// YAML drawer（复用 YamlDrawer 组件，不再手写 el-drawer + YamlEditor 样板）
const yamlDrawerVisible = ref(false)
const yamlTargetName = ref('')

// 共享对话框
const taintDialog = ref<InstanceType<typeof NodeTaintDialog>>()
const labelDialog = ref<InstanceType<typeof NodeLabelDialog>>()
const drainDialog = ref<InstanceType<typeof NodeDrainDialog>>()

const filteredList = computed(() => {
  if (!searchName.value) return nodeList.value
  const keyword = searchName.value.toLowerCase()
  return nodeList.value.filter((n) => n.name?.toLowerCase().includes(keyword))
})

// silent=true 时不触发页面级 loading 遮罩，用于自动刷新，避免每轮全页转圈
async function fetchNodes(silent = false) {
  if (!silent) loading.value = true
  try {
    const res = await getNodeList()
    nodeList.value = res.data || []
  } catch (e: any) {
    // 不静默吞掉：网络/权限/服务端错误都需要提示，否则用户无法区分"无节点"与"加载失败"
    ElMessage.error(e?.message || '加载节点列表失败')
  } finally {
    loading.value = false
  }
}

function statusType(node: NodeInfo) {
  if (node.status === 'Ready') return 'success'
  if (node.status === 'NotReady') return 'danger'
  return 'warning'
}

// 资源数值格式化委托到 utils/resource（与 Dashboard 等共享）
const fmtCpu = formatCpuCores
const fmtMem = formatMemGiB

function handleViewYaml(row: NodeInfo) {
  yamlTargetName.value = row.name
  yamlDrawerVisible.value = true
}

function handleYamlSaved() { fetchNodes() }

function handleDetail(row: NodeInfo) { router.push(`/nodes/${row.name}`) }

function handleTaints(row: NodeInfo) { taintDialog.value?.open(row.name, row.taints) }
function handleLabels(row: NodeInfo) { labelDialog.value?.open(row.name, row.labels) }
function handleDrain(row: NodeInfo) { drainDialog.value?.open(row.name) }

const { handleCordon, handleDelete } = useNodeActions(() => fetchNodes())

// 自动刷新走 silent 路径（不显示遮罩）；手动刷新走非 silent（显示遮罩）
const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(
  () => fetchNodes(true),
  { manualFetch: () => fetchNodes(false) },
)

onMounted(() => fetchNodes())
</script>

<template>
  <div class="page-container">
    <ResourceListToolbar
      :search-value="searchName"
      :total-count="nodeList.length"
      :show-namespace="false"
      search-placeholder="搜索节点名称"
      @search-input="searchName = $event"
    >
      <template #extra>
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
        <ViewModeToggle v-model="viewMode" storage-key="gkube.node.viewMode" />
      </template>
    </ResourceListToolbar>
    <el-card shadow="never" class="table-card">
      <el-table v-if="viewMode === 'table'" :data="filteredList" v-loading="loading" stripe>
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }"><el-tag :type="statusType(row)" size="small" effect="dark">{{ row.status || 'Unknown' }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip>
          <template #default="{ row }"><el-button link type="primary" @click="handleDetail(row)">{{ row.name }}</el-button></template>
        </el-table-column>
        <el-table-column prop="internal_ip" label="IP 地址" width="150">
          <template #default="{ row }">{{ row.internal_ip || '-' }}</template>
        </el-table-column>
        <el-table-column prop="roles" label="角色" width="110">
          <template #default="{ row }">{{ row.roles || '-' }}</template>
        </el-table-column>
        <el-table-column prop="version" label="版本" width="120" show-overflow-tooltip>
          <template #default="{ row }">{{ row.version || '-' }}</template>
        </el-table-column>
        <el-table-column prop="creationTimestamp" label="年龄" width="110" show-overflow-tooltip>
          <template #default="{ row }">{{ formatAge(row.creationTimestamp, false) }}</template>
        </el-table-column>
        <el-table-column label="操作" min-width="380" fixed="right" align="center">
          <template #default="{ row }">
            <div class="table-actions">
              <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
              <el-button size="small" :type="row.unschedulable ? 'success' : 'warning'" @click="handleCordon(row.name, row.unschedulable)">
                {{ row.unschedulable ? '解除封锁' : '封锁' }}
              </el-button>
              <el-button size="small" type="primary" @click="handleTaints(row)">污点</el-button>
              <el-button size="small" type="info" @click="handleLabels(row)">标签</el-button>
              <el-button size="small" type="danger" @click="handleDrain(row)">驱逐</el-button>
              <el-tooltip v-if="row.status === 'Ready'" content="节点在线，删除后会重新注册（需先停止 kubelet）" placement="top">
                <span><el-button size="small" type="danger" plain disabled>删除</el-button></span>
              </el-tooltip>
              <el-button v-else size="small" type="danger" plain @click="handleDelete(row.name, row.status === 'Ready')">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <el-row v-else :gutter="16">
        <el-col v-for="node in filteredList" :key="node.name" :xs="24" :sm="12" :md="8" style="margin-bottom: 16px;">
          <el-card shadow="hover" class="node-card">
            <template #header>
              <div class="node-header">
                <el-button link type="primary" @click="handleDetail(node)">{{ node.name }}</el-button>
                <div class="node-header-tags">
                  <el-tag :type="statusType(node)" size="small" effect="dark">{{ node.status || 'Unknown' }}</el-tag>
                  <el-tag v-if="node.roles" size="small" effect="plain">{{ node.roles }}</el-tag>
                  <el-tag v-if="node.unschedulable" type="warning" size="small">已封锁</el-tag>
                </div>
              </div>
            </template>

            <div class="node-meta">
              {{ node.internal_ip || '-' }}<template v-if="node.version"> · {{ node.version }}</template><template v-if="node.creationTimestamp"> · {{ formatAge(node.creationTimestamp, false) }}</template>
            </div>

            <div class="node-usage">
              <div class="usage-item">
                <span class="usage-label">CPU</span>
                <el-progress :percentage="usagePercent(node.cpu_used, node.cpu_total)" :color="progressColor(usagePercent(node.cpu_used, node.cpu_total))" :stroke-width="16" :text-inside="true" :format="(p: number) => `${fmtCpu(node.cpu_used)}/${fmtCpu(node.cpu_total)}核 ${p}%`" />
              </div>
              <div class="usage-item">
                <span class="usage-label">内存</span>
                <el-progress :percentage="usagePercent(node.mem_used, node.mem_total)" :color="progressColor(usagePercent(node.mem_used, node.mem_total))" :stroke-width="16" :text-inside="true" :format="(p: number) => `${fmtMem(node.mem_used)}/${fmtMem(node.mem_total)}GiB ${p}%`" />
              </div>
              <div class="usage-item">
                <span class="usage-label">Pods</span>
                <el-progress :percentage="usagePercent(node.pod_count, node.pod_total)" :color="progressColor(usagePercent(node.pod_count, node.pod_total))" :stroke-width="16" :text-inside="true" :format="() => `${node.pod_count || 0}/${node.pod_total || 0}`" />
              </div>
            </div>

            <div class="node-footer">
              <el-button size="small" @click="handleViewYaml(node)">YAML</el-button>
              <el-button size="small" :type="node.unschedulable ? 'success' : 'warning'" @click="handleCordon(node.name, node.unschedulable)">
                {{ node.unschedulable ? '解除封锁' : '封锁' }}
              </el-button>
              <el-button size="small" type="primary" @click="handleTaints(node)">污点</el-button>
              <el-button size="small" type="info" @click="handleLabels(node)">标签</el-button>
              <el-button size="small" type="danger" @click="handleDrain(node)">驱逐</el-button>
              <el-tooltip v-if="node.status === 'Ready'" content="节点在线，删除后会重新注册（需先停止 kubelet）" placement="top">
                <span><el-button size="small" type="danger" plain disabled>删除</el-button></span>
              </el-tooltip>
              <el-button v-else size="small" type="danger" plain @click="handleDelete(node.name, node.status === 'Ready')">删除</el-button>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>

    <!-- YAML Drawer（复用通用组件） -->
    <YamlDrawer
      v-model="yamlDrawerVisible"
      resource-type="node"
      :name="yamlTargetName"
      @saved="handleYamlSaved"
    />

    <!-- 共享对话框 -->
    <NodeTaintDialog ref="taintDialog" @saved="() => fetchNodes()" />
    <NodeLabelDialog ref="labelDialog" @saved="() => fetchNodes()" />
    <NodeDrainDialog ref="drainDialog" @saved="() => fetchNodes()" />
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.table-card { border-radius: 8px; }
.table-actions { display: flex; flex-wrap: nowrap; justify-content: center; align-items: center; gap: 4px; }
.table-actions .el-button { margin-left: 0 !important; }
.node-card {
  height: 100%;
  background: linear-gradient(180deg, var(--gk-color-primary-bg) 0%, var(--gk-color-bg-card) 60%);
  border-color: var(--gk-color-primary-light);
}
.node-header { display: flex; justify-content: space-between; align-items: center; gap: 8px; }
.node-header-tags { display: flex; align-items: center; gap: 6px; }
.node-meta { font-size: 12px; color: var(--gk-color-text-secondary); margin-bottom: 12px; }
.node-usage { margin-bottom: 12px; }
.usage-item { display: flex; align-items: center; margin-bottom: 10px; }
.usage-item:last-child { margin-bottom: 0; }
.usage-label { width: 36px; flex-shrink: 0; font-size: 12px; color: var(--gk-color-text-secondary); }
.usage-item :deep(.el-progress) { flex: 1; }
.usage-item :deep(.el-progress-bar) { padding-right: 0; }
.node-footer { display: flex; flex-wrap: nowrap; align-items: center; gap: 4px; border-top: 1px solid var(--gk-color-border-light); padding-top: 12px; }
.node-footer .el-button { margin-left: 0 !important; padding: 5px 8px; font-size: 12px; height: auto; }
</style>
