<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, Search, Plus } from '@element-plus/icons-vue'
import { getNodeList, getNodeYaml, updateNodeYaml, cordonNode, updateNodeTaints, updateNodeLabels, drainNode, deleteNode, type NodeInfo } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'

const router = useRouter()
const loading = ref(false)
const nodeList = ref<NodeInfo[]>([])
const searchName = ref('')
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)
const yamlTarget = ref<any>(null)
const yamlEditing = ref(false)
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

async function handleViewYaml(row: any) {
  yamlTarget.value = row
  yamlDialogVisible.value = true; yamlEditing.value = false; yamlLoading.value = true; yamlContent.value = ''
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
    yamlEditing.value = false
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
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <el-input v-model="searchName" placeholder="搜索节点名称" style="width: 220px;" clearable>
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
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
      </div>
    </el-card>
    <el-card shadow="never" class="table-card">
      <el-table :data="filteredList" v-loading="loading" stripe>
        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }"><el-tag :type="statusType(row)" size="small" effect="dark">{{ row.status || 'Unknown' }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="name" label="名称" min-width="240" show-overflow-tooltip>
          <template #default="{ row }"><el-button link type="primary" @click="handleDetail(row)">{{ row.name }}</el-button></template>
        </el-table-column>
        <el-table-column prop="internal_ip" label="IP 地址" min-width="180">
          <template #default="{ row }">{{ row.internal_ip || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" min-width="360" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
            <el-button size="small" :type="row.unschedulable ? 'success' : 'warning'" @click="handleCordon(row)">
              {{ row.unschedulable ? '解除封锁' : '封锁' }}
            </el-button>
            <el-button size="small" type="primary" @click="handleTaints(row)">污点</el-button>
            <el-button size="small" type="info" @click="handleLabels(row)">标签</el-button>
            <el-button size="small" type="danger" @click="handleDrain(row)">驱逐</el-button>
            <el-button size="small" type="danger" plain @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- YAML Drawer -->
    <el-drawer v-model="yamlDialogVisible" title="节点 YAML" size="85%" direction="rtl" class="yaml-drawer"
      :body-style="{ padding: '0', height: '100%' }">
      <div style="padding: 6px 12px; display: flex; gap: 8px; border-bottom: 1px solid var(--el-border-color-lighter);">
        <el-button v-if="!yamlEditing" size="small" type="primary" @click="yamlEditing = true">编辑</el-button>
        <template v-else>
          <el-button size="small" type="success" :loading="yamlSaving" @click="handleSaveYaml">保存</el-button>
          <el-button size="small" @click="yamlEditing = false; fetchYaml()">取消</el-button>
        </template>
      </div>
      <div v-loading="yamlLoading" style="height: calc(100vh - 90px);">
        <YamlEditor v-model="yamlContent" height="100%" :read-only="!yamlEditing" auto-format show-toolbar />
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
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; gap: 12px; align-items: center; flex-wrap: wrap; }
.table-card { border-radius: 8px; }
</style>

<style>
.yaml-drawer .el-drawer__header {
  padding: 6px 16px;
  margin-bottom: 0;
  min-height: auto;
}
</style>
