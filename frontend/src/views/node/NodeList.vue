<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, Plus } from '@element-plus/icons-vue'
import { getNodeList, getNodeYaml, updateNodeYaml, cordonNode, updateNodeTaints, updateNodeLabels, drainNode, deleteNode, type NodeInfo } from '@/api/resource'
import { usagePercent, progressColor } from '@/utils/helpers'
import YamlEditor from '@/components/YamlEditor.vue'
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
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)
const yamlTarget = ref<any>(null)
const yamlSaving = ref(false)
const taintDialogVisible = ref(false)
const taintTarget = ref<any>(null)
const taints = ref<any[]>([])
const labelsDialogVisible = ref(false)
const labelsTarget = ref<any>(null)
const labels = ref<Record<string, string>>({})
const labelsArray = ref<{ key: string; value: string }[]>([])
const drainDialogVisible = ref(false)
const drainTarget = ref<any>(null)
const drainOptions = ref({
  ignoreDaemonSets: true,
  deleteLocalData: false,
  gracePeriod: -1,
  force: false
})

const filteredList = computed(() => {
  if (!searchName.value) return nodeList.value
  const keyword = searchName.value.toLowerCase()
  return nodeList.value.filter((n) => n.name?.toLowerCase().includes(keyword))
})

async function fetchNodes() {
  loading.value = true
  try {
    const res: any = await getNodeList()
    nodeList.value = res.data || []
  } catch {
    // Silently handle — resource may not exist in cluster
  } finally { loading.value = false }
}

function statusType(node: any) {
  if (node.status === 'Ready') return 'success'
  if (node.status === 'NotReady') return 'danger'
  return 'warning'
}

// 资源数值格式化（CPU 核 / 内存 GiB）
function fmtCpu(n: number) { return (n || 0).toFixed(2) }
function fmtMem(n: number) { return (n || 0).toFixed(1) }

async function handleViewYaml(row: any) {
  yamlTarget.value = row
  yamlDialogVisible.value = true; yamlLoading.value = true; yamlContent.value = ''
  try {
    const res: any = await getNodeYaml({ name: row.name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) { ElMessage.error(e?.message || 'Failed to load YAML'); yamlDialogVisible.value = false }
  finally { yamlLoading.value = false }
}

async function fetchYaml() {
  if (!yamlTarget.value) return
  yamlLoading.value = true
  try {
    const res: any = await getNodeYaml({ name: yamlTarget.value.name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) { ElMessage.error(e?.message || 'Failed to load YAML') }
  finally { yamlLoading.value = false }
}

async function handleSaveYaml() {
  if (!yamlTarget.value) return
  yamlSaving.value = true
  try {
    await updateNodeYaml({ name: yamlTarget.value.name, yaml: yamlContent.value })
    ElMessage.success('YAML 保存成功')
    fetchNodes()
  } catch (e: any) {
    ElMessage.error(e?.message || '保存 YAML 失败')
  } finally { yamlSaving.value = false }
}

function handleDetail(row: any) { router.push(`/nodes/${row.name}`) }

async function handleCordon(row: any) {
  const isCordon = row.unschedulable
  const actionLabel = isCordon ? '解除封锁' : '封锁'
  try {
    await ElMessageBox.confirm(`确定要${actionLabel}节点 "${row.name}" 吗？`, '确认操作', { type: 'warning' })
    await cordonNode({ name: row.name, cordon: !isCordon })
    ElMessage.success(`节点已${actionLabel}`); fetchNodes()
  } catch { /* cancelled */ }
}

function handleTaints(row: any) {
  taintTarget.value = row
  taints.value = (row.taints || []).map((t: any) => ({ ...t }))
  if (taints.value.length === 0) taints.value = [{ key: '', value: '', effect: 'NoSchedule' }]
  taintDialogVisible.value = true
}

function addTaint() { taints.value.push({ key: '', value: '', effect: 'NoSchedule' }) }
function removeTaint(index: number) { taints.value.splice(index, 1) }

async function handleSaveTaints() {
  try {
    await updateNodeTaints({ name: taintTarget.value.name, taints: taints.value.filter(t => t.key) })
    ElMessage.success('污点已更新'); taintDialogVisible.value = false; fetchNodes()
  } catch (e: any) { ElMessage.error(e?.message || 'Failed to update taints') }
}

// Labels
function handleLabels(row: any) {
  labelsTarget.value = row
  labels.value = { ...(row.labels || {}) }
  labelsArray.value = Object.entries(labels.value).map(([key, value]) => ({ key, value }))
  if (labelsArray.value.length === 0) labelsArray.value = [{ key: '', value: '' }]
  labelsDialogVisible.value = true
}

function addLabel() { labelsArray.value.push({ key: '', value: '' }) }
function removeLabel(index: number) { labelsArray.value.splice(index, 1) }

async function handleSaveLabels() {
  try {
    const labelsMap: Record<string, string> = {}
    labelsArray.value.forEach(l => { if (l.key) labelsMap[l.key] = l.value })
    await updateNodeLabels({ name: labelsTarget.value.name, labels: labelsMap })
    ElMessage.success('标签已更新'); labelsDialogVisible.value = false; fetchNodes()
  } catch (e: any) { ElMessage.error(e?.message || 'Failed to update labels') }
}

// Drain
function handleDrain(row: any) {
  drainTarget.value = row
  drainOptions.value = { ignoreDaemonSets: true, deleteLocalData: false, gracePeriod: -1, force: false }
  drainDialogVisible.value = true
}

async function handleConfirmDrain() {
  try {
    await ElMessageBox.confirm(
      `确定要驱逐节点 "${drainTarget.value.name}" 上的所有 Pod 吗？此操作会先封锁节点再驱逐 Pod。`,
      '确认驱逐',
      { type: 'warning', confirmButtonText: '驱逐', cancelButtonText: '取消' }
    )
    const res: any = await drainNode({ name: drainTarget.value.name, ...drainOptions.value })
    const evicted = res.data?.evicted || []
    const skipped = res.data?.skipped || []
    ElMessage.success(`驱逐完成：${evicted.length} 个 Pod 已驱逐，${skipped.length} 个已跳过`)
    drainDialogVisible.value = false
    fetchNodes()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '驱逐失败')
  }
}

// Delete
async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(
      `确定要删除节点 "${row.name}" 吗？此操作不可恢复，节点将从集群中移除。`,
      '确认删除',
      { type: 'error', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await deleteNode({ name: row.name })
    ElMessage.success('节点已删除')
    fetchNodes()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '删除失败')
  }
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchNodes)

onMounted(fetchNodes)
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
        <el-table-column prop="age" label="年龄" width="170" show-overflow-tooltip>
          <template #default="{ row }">{{ row.age || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" min-width="380" fixed="right" align="center">
          <template #default="{ row }">
            <div class="table-actions">
              <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
              <el-button size="small" :type="row.unschedulable ? 'success' : 'warning'" @click="handleCordon(row)">
                {{ row.unschedulable ? '解除封锁' : '封锁' }}
              </el-button>
              <el-button size="small" type="primary" @click="handleTaints(row)">污点</el-button>
              <el-button size="small" type="info" @click="handleLabels(row)">标签</el-button>
              <el-button size="small" type="danger" @click="handleDrain(row)">驱逐</el-button>
              <el-button size="small" type="danger" plain @click="handleDelete(row)">删除</el-button>
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
              {{ node.internal_ip || '-' }}<template v-if="node.version"> · {{ node.version }}</template><template v-if="node.age"> · {{ node.age }}</template>
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
              <el-button size="small" :type="node.unschedulable ? 'success' : 'warning'" @click="handleCordon(node)">
                {{ node.unschedulable ? '解除封锁' : '封锁' }}
              </el-button>
              <el-button size="small" type="primary" @click="handleTaints(node)">污点</el-button>
              <el-button size="small" type="info" @click="handleLabels(node)">标签</el-button>
              <el-button size="small" type="danger" @click="handleDrain(node)">驱逐</el-button>
              <el-button size="small" type="danger" plain @click="handleDelete(node)">删除</el-button>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>

    <!-- YAML Drawer -->
    <el-drawer v-model="yamlDialogVisible" title="节点 YAML" size="85%" direction="rtl" class="yaml-drawer"
      :body-style="{ padding: '0', height: '100%' }">
      <div v-loading="yamlLoading" style="height: calc(100vh - 56px);">
        <YamlEditor v-model="yamlContent" height="100%" auto-format show-toolbar show-save-buttons :saving="yamlSaving" @save="handleSaveYaml" @cancel="fetchYaml" />
      </div>
    </el-drawer>

    <!-- Taints Dialog -->
    <el-dialog v-model="taintDialogVisible" title="管理污点" width="600px">
      <div v-for="(taint, index) in taints" :key="index" style="display: flex; gap: 8px; margin-bottom: 12px; align-items: center;">
        <el-input v-model="taint.key" placeholder="Key" style="flex: 2;" />
        <el-input v-model="taint.value" placeholder="Value" style="flex: 1;" />
        <el-select v-model="taint.effect" style="flex: 1.5;">
          <el-option label="NoSchedule" value="NoSchedule" />
          <el-option label="PreferNoSchedule" value="PreferNoSchedule" />
          <el-option label="NoExecute" value="NoExecute" />
        </el-select>
        <el-button type="danger" circle size="small" @click="removeTaint(index)"><el-icon><Delete /></el-icon></el-button>
      </div>
      <el-button @click="addTaint" style="margin-top: 8px;"><el-icon><Plus /></el-icon> 添加污点</el-button>
      <template #footer>
        <el-button @click="taintDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveTaints">保存</el-button>
      </template>
    </el-dialog>

    <!-- Labels Dialog -->
    <el-dialog v-model="labelsDialogVisible" title="管理标签" width="650px">
      <div v-for="(label, index) in labelsArray" :key="index" style="display: flex; gap: 8px; margin-bottom: 12px; align-items: center;">
        <el-input v-model="label.key" placeholder="Key" style="flex: 2;" />
        <el-input v-model="label.value" placeholder="Value" style="flex: 2;" />
        <el-button type="danger" circle size="small" @click="removeLabel(index)"><el-icon><Delete /></el-icon></el-button>
      </div>
      <el-button @click="addLabel" style="margin-top: 8px;"><el-icon><Plus /></el-icon> 添加标签</el-button>
      <template #footer>
        <el-button @click="labelsDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveLabels">保存</el-button>
      </template>
    </el-dialog>

    <!-- Drain Dialog -->
    <el-dialog v-model="drainDialogVisible" title="驱逐 Pod" width="500px">
      <el-alert type="warning" :closable="false" style="margin-bottom: 16px;">
        <template #title>驱逐操作会先封锁节点，然后驱逐节点上的所有 Pod。请确认以下选项：</template>
      </el-alert>
      <el-form label-width="160px">
        <el-form-item label="忽略 DaemonSet">
          <el-switch v-model="drainOptions.ignoreDaemonSets" />
          <span style="margin-left: 8px; color: #909399; font-size: 12px;">跳过 DaemonSet 管理的 Pod</span>
        </el-form-item>
        <el-form-item label="删除本地数据">
          <el-switch v-model="drainOptions.deleteLocalData" />
          <span style="margin-left: 8px; color: #909399; font-size: 12px;">删除使用 emptyDir 的 Pod</span>
        </el-form-item>
        <el-form-item label="优雅终止时间(秒)">
          <el-input-number v-model="drainOptions.gracePeriod" :min="-1" :max="3600" />
          <span style="margin-left: 8px; color: #909399; font-size: 12px;">-1 使用 Pod 默认值</span>
        </el-form-item>
        <el-form-item label="强制驱逐">
          <el-switch v-model="drainOptions.force" />
          <span style="margin-left: 8px; color: #909399; font-size: 12px;">驱逐 kube-system 下的 Pod</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="drainDialogVisible = false">取消</el-button>
        <el-button type="warning" @click="handleConfirmDrain">确认驱逐</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.table-card { border-radius: 8px; }
.table-actions { display: flex; flex-wrap: nowrap; justify-content: center; align-items: center; gap: 4px; }
.table-actions .el-button { margin-left: 0 !important; }
.node-card { height: 100%; }
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

<style>
.yaml-drawer .el-drawer__header {
  padding: 6px 16px;
  margin-bottom: 0;
  min-height: auto;
}
</style>
