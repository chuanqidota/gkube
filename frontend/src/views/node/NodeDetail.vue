<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, Plus, Monitor, Cpu, Coin, Grid, Files, Warning, PriceTag, CircleClose, Search } from '@element-plus/icons-vue'
import { getNodeDetail, getNodeYaml, getNodePods, getNodeEvents, cordonNode, updateNodeTaints, updateNodeLabels, updateNodeYaml, drainNode, deleteNode, type NodeDetail as NodeDetailType } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const node = ref<NodeDetailType | null>(null)
const pods = ref<any[]>([])
const podsLoading = ref(false)
const podSearch = ref('')

// 过滤后的 Pods 列表
const filteredPods = computed(() => {
  if (!podSearch.value) return pods.value
  const keyword = podSearch.value.toLowerCase()
  return pods.value.filter(pod =>
    pod.name?.toLowerCase().includes(keyword) ||
    pod.namespace?.toLowerCase().includes(keyword) ||
    pod.ip?.toLowerCase().includes(keyword)
  )
})
const events = ref<any[]>([])
const eventsLoading = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)
const yamlEditing = ref(false)
const yamlSaving = ref(false)
const activeTab = ref('info')
const taintDialogVisible = ref(false)
const taints = ref<any[]>([])
const labelsDialogVisible = ref(false)
const labelsArray = ref<{ key: string; value: string }[]>([])
const drainDialogVisible = ref(false)
const drainOptions = ref({
  ignoreDaemonSets: true,
  deleteLocalData: false,
  gracePeriod: -1,
  force: false
})

const nodeName = route.params.name as string

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getNodeDetail({ name: nodeName })
    node.value = res.data
    if (node.value) fetchPods()
  } catch (e: any) {
    ElMessage.error(e?.message || '加载节点详情失败')
  } finally {
    loading.value = false
  }
}

async function fetchPods() {
  podsLoading.value = true
  try {
    const res: any = await getNodePods({ name: nodeName })
    const rawPods = res.data || []
    pods.value = rawPods.map((pod: any) => {
      const restarts = (pod.status?.containerStatuses || []).reduce(
        (sum: number, cs: any) => sum + (cs.restartCount || 0), 0
      )
      return {
        name: pod.metadata?.name,
        namespace: pod.metadata?.namespace,
        status: pod.status?.phase || (pod.status?.conditions?.find((c: any) => c.type === 'Ready')?.status === 'True' ? 'Running' : 'Pending'),
        ip: pod.status?.podIP || '-',
        restarts,
        age: pod.metadata?.creationTimestamp,
      }
    })
  } catch (e: any) {
    ElMessage.error(e?.message || '加载 Pod 列表失败')
  }
  finally { podsLoading.value = false }
}

async function fetchEvents() {
  eventsLoading.value = true
  try {
    const res: any = await getNodeEvents({ name: nodeName })
    events.value = res.data || []
  } catch (e: any) {
    ElMessage.error(e?.message || '加载事件列表失败')
  }
  finally { eventsLoading.value = false }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getNodeYaml({ name: nodeName })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || '加载 YAML 失败')
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
    ElMessage.success('YAML 保存成功')
    yamlEditing.value = false
    fetchDetail()
  } catch (e: any) {
    ElMessage.error(e?.message || '保存 YAML 失败')
  } finally { yamlSaving.value = false }
}

function statusType(status: string) {
  if (status === 'Ready') return 'success'
  if (status === 'NotReady') return 'danger'
  return 'warning'
}

async function handleCordon() {
  const isCordon = node.value?.unschedulable
  const actionLabel = isCordon ? '解除封锁' : '封锁'
  try {
    await ElMessageBox.confirm(`确定要${actionLabel}节点 "${nodeName}" 吗？`, '确认操作', { type: 'warning' })
    await cordonNode({ name: nodeName, cordon: !isCordon })
    ElMessage.success(`节点已${actionLabel}`)
    fetchDetail()
  } catch { /* cancelled */ }
}

// Taints
function handleTaints() {
  taints.value = (node.value?.taints || []).map((t: any) => ({ ...t }))
  if (taints.value.length === 0) taints.value = [{ key: '', value: '', effect: 'NoSchedule' }]
  taintDialogVisible.value = true
}

function addTaint() { taints.value.push({ key: '', value: '', effect: 'NoSchedule' }) }
function removeTaint(index: number) { taints.value.splice(index, 1) }

async function handleSaveTaints() {
  try {
    await updateNodeTaints({ name: nodeName, taints: taints.value.filter(t => t.key) })
    ElMessage.success('污点已更新')
    taintDialogVisible.value = false
    fetchDetail()
  } catch (e: any) { ElMessage.error(e?.message || '更新污点失败') }
}

// Labels
function handleLabels() {
  labelsArray.value = Object.entries(node.value?.labels || {}).map(([key, value]) => ({ key, value }))
  if (labelsArray.value.length === 0) labelsArray.value = [{ key: '', value: '' }]
  labelsDialogVisible.value = true
}

function addLabel() { labelsArray.value.push({ key: '', value: '' }) }
function removeLabel(index: number) { labelsArray.value.splice(index, 1) }

async function handleSaveLabels() {
  try {
    const labelsMap: Record<string, string> = {}
    labelsArray.value.forEach(l => { if (l.key) labelsMap[l.key] = l.value })
    await updateNodeLabels({ name: nodeName, labels: labelsMap })
    ElMessage.success('标签已更新')
    labelsDialogVisible.value = false
    fetchDetail()
  } catch (e: any) { ElMessage.error(e?.message || '更新标签失败') }
}

// Drain
function handleDrain() {
  drainOptions.value = { ignoreDaemonSets: true, deleteLocalData: false, gracePeriod: -1, force: false }
  drainDialogVisible.value = true
}

async function handleConfirmDrain() {
  try {
    await ElMessageBox.confirm(
      `确定要驱逐节点 "${nodeName}" 上的所有 Pod 吗？此操作会先封锁节点再驱逐 Pod。`,
      '确认驱逐',
      { type: 'warning', confirmButtonText: '驱逐', cancelButtonText: '取消' }
    )
    const res: any = await drainNode({ name: nodeName, ...drainOptions.value })
    const evicted = res.data?.evicted || []
    const skipped = res.data?.skipped || []
    ElMessage.success(`驱逐完成：${evicted.length} 个 Pod 已驱逐，${skipped.length} 个已跳过`)
    drainDialogVisible.value = false
    fetchDetail()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '驱逐失败')
  }
}

// Delete
async function handleDelete() {
  try {
    await ElMessageBox.confirm(
      `确定要删除节点 "${nodeName}" 吗？此操作不可恢复，节点将从集群中移除。`,
      '确认删除',
      { type: 'error', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await deleteNode({ name: nodeName })
    ElMessage.success('节点已删除')
    router.push('/nodes')
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '删除失败')
  }
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

// 格式化 CPU 值
function formatCPU(val: any): string {
  if (!val) return '-'
  const s = String(val)
  // 如果是毫核格式 (如 "500m")
  if (s.endsWith('m')) {
    return s
  }
  // 纯数字表示核数
  const num = parseFloat(s)
  if (!isNaN(num)) {
    return `${num} Core`
  }
  return s
}

// 格式化内存/存储值 (Ki -> Gi/Mi)
function formatMemory(val: any): string {
  if (!val) return '-'
  const s = String(val)
  // 处理 Ki 单位
  if (s.endsWith('Ki')) {
    const ki = parseInt(s)
    if (ki >= 1048576) { // >= 1Gi
      return `${(ki / 1048576).toFixed(1)} Gi`
    } else if (ki >= 1024) { // >= 1Mi
      return `${(ki / 1024).toFixed(0)} Mi`
    }
    return `${ki} KiB`
  }
  // 处理 Mi 单位
  if (s.endsWith('Mi')) {
    const mi = parseInt(s)
    if (mi >= 1024) {
      return `${(mi / 1024).toFixed(1)} Gi`
    }
    return `${mi} Mi`
  }
  // 处理 Gi 单位
  if (s.endsWith('Gi')) {
    return s
  }
  // 处理 Ti 单位
  if (s.endsWith('Ti')) {
    return s
  }
  // 处理纯数字（字节）
  const num = parseInt(s)
  if (!isNaN(num)) {
    if (num >= 1073741824) { // >= 1Gi
      return `${(num / 1073741824).toFixed(1)} Gi`
    } else if (num >= 1048576) { // >= 1Mi
      return `${(num / 1048576).toFixed(0)} Mi`
    } else if (num >= 1024) { // >= 1Ki
      return `${(num / 1024).toFixed(0)} KiB`
    }
    return `${num} B`
  }
  return s
}

// 通用格式化
function formatCapacity(val: any): string {
  if (!val) return '-'
  return String(val)
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchDetail, { autoStart: false })

onMounted(fetchDetail)
</script>

<template>
  <div class="page-container" v-loading="loading">
    <div class="page-header">
      <h2 style="margin: 0;">节点: {{ nodeName }}</h2>
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
        <el-button :type="node?.unschedulable ? 'success' : 'warning'" @click="handleCordon">
          {{ node?.unschedulable ? '解除封锁' : '封锁' }}
        </el-button>
        <el-button type="primary" @click="handleTaints">污点</el-button>
        <el-button type="info" @click="handleLabels">标签</el-button>
        <el-button type="danger" @click="handleDrain">驱逐</el-button>
        <el-button type="danger" plain @click="handleDelete">删除</el-button>
        <el-button @click="router.push('/nodes')">返回列表</el-button>
      </div>
    </div>

    <template v-if="node">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <!-- Info Tab -->
        <el-tab-pane label="概览" name="info">
          <el-card shadow="never">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="名称">{{ node.name }}</el-descriptions-item>
              <el-descriptions-item label="状态"><el-tag :type="statusType(node.status)" size="small">{{ node.status || 'Unknown' }}</el-tag></el-descriptions-item>
              <el-descriptions-item label="角色">{{ node.roles || '-' }}</el-descriptions-item>
              <el-descriptions-item label="Kubelet 版本">{{ node.version || '-' }}</el-descriptions-item>
              <el-descriptions-item label="操作系统">{{ node.os || '-' }}</el-descriptions-item>
              <el-descriptions-item label="内核版本">{{ node.kernel || '-' }}</el-descriptions-item>
              <el-descriptions-item label="容器运行时">{{ node.container_runtime || '-' }}</el-descriptions-item>
              <el-descriptions-item label="内部 IP">{{ node.internal_ip || '-' }}</el-descriptions-item>
              <el-descriptions-item label="外部 IP">{{ node.external_ip || '-' }}</el-descriptions-item>
              <el-descriptions-item label="主机名">{{ node.hostname || '-' }}</el-descriptions-item>
              <el-descriptions-item label="架构">{{ node.architecture || '-' }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ node.age || '-' }}</el-descriptions-item>
              <el-descriptions-item label="不可调度">
                <el-tag :type="node.unschedulable ? 'danger' : 'success'" size="small">{{ node.unschedulable ? '是' : '否' }}</el-tag>
              </el-descriptions-item>
            </el-descriptions>

            <!-- Resource Capacity -->
            <div v-if="node.capacity || node.allocatable" class="resource-section">
              <div class="section-header">
                <el-icon class="section-icon"><Monitor /></el-icon>
                <span class="section-title">资源容量</span>
              </div>
              <div class="resource-cards">
                <div class="resource-card">
                  <div class="resource-icon cpu-icon">
                    <el-icon><Cpu /></el-icon>
                  </div>
                  <div class="resource-info">
                    <div class="resource-label">CPU</div>
                    <div class="resource-values">
                      <div class="value-item">
                        <span class="value-label">总容量</span>
                        <span class="value-number">{{ formatCPU(node.capacity?.cpu) }}</span>
                      </div>
                      <div class="value-item">
                        <span class="value-label">可分配</span>
                        <span class="value-number highlight">{{ formatCPU(node.allocatable?.cpu) }}</span>
                      </div>
                    </div>
                  </div>
                </div>

                <div class="resource-card">
                  <div class="resource-icon memory-icon">
                    <el-icon><Coin /></el-icon>
                  </div>
                  <div class="resource-info">
                    <div class="resource-label">内存</div>
                    <div class="resource-values">
                      <div class="value-item">
                        <span class="value-label">总容量</span>
                        <span class="value-number">{{ formatMemory(node.capacity?.memory) }}</span>
                      </div>
                      <div class="value-item">
                        <span class="value-label">可分配</span>
                        <span class="value-number highlight">{{ formatMemory(node.allocatable?.memory) }}</span>
                      </div>
                    </div>
                  </div>
                </div>

                <div class="resource-card">
                  <div class="resource-icon pods-icon">
                    <el-icon><Grid /></el-icon>
                  </div>
                  <div class="resource-info">
                    <div class="resource-label">Pod 数量</div>
                    <div class="resource-values">
                      <div class="value-item">
                        <span class="value-label">总容量</span>
                        <span class="value-number">{{ formatCapacity(node.capacity?.pods) }}</span>
                      </div>
                      <div class="value-item">
                        <span class="value-label">可分配</span>
                        <span class="value-number highlight">{{ formatCapacity(node.allocatable?.pods) }}</span>
                      </div>
                    </div>
                  </div>
                </div>

                <div class="resource-card">
                  <div class="resource-icon storage-icon">
                    <el-icon><Files /></el-icon>
                  </div>
                  <div class="resource-info">
                    <div class="resource-label">临时存储</div>
                    <div class="resource-values">
                      <div class="value-item">
                        <span class="value-label">总容量</span>
                        <span class="value-number">{{ formatMemory(node.capacity?.['ephemeral-storage']) }}</span>
                      </div>
                      <div class="value-item">
                        <span class="value-label">可分配</span>
                        <span class="value-number highlight">{{ formatMemory(node.allocatable?.['ephemeral-storage']) }}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Conditions -->
            <div v-if="node.conditions && node.conditions.length > 0" style="margin-top: 24px;">
              <div class="section-header">
                <el-icon class="section-icon"><Warning /></el-icon>
                <span class="section-title">节点状态</span>
              </div>
              <el-table :data="node.conditions" border stripe>
                <el-table-column prop="type" label="类型" width="180" />
                <el-table-column label="状态" width="100">
                  <template #default="{ row }"><el-tag :type="row.status === 'True' ? 'success' : 'danger'" size="small">{{ row.status }}</el-tag></template>
                </el-table-column>
                <el-table-column prop="reason" label="原因" width="180" />
                <el-table-column prop="message" label="消息" min-width="250" show-overflow-tooltip />
                <el-table-column prop="lastTransitionTime" label="最后变更" width="180" />
              </el-table>
            </div>

            <!-- Labels -->
            <div style="margin-top: 24px;">
              <div class="section-header">
                <el-icon class="section-icon"><PriceTag /></el-icon>
                <span class="section-title">标签</span>
                <el-button size="small" @click="handleLabels" style="margin-left: auto;">编辑</el-button>
              </div>
              <div v-if="node.labels && Object.keys(node.labels).length > 0">
                <el-tag v-for="(val, key) in node.labels" :key="key" style="margin-right: 8px; margin-bottom: 8px;">{{ key }}={{ val }}</el-tag>
              </div>
              <span v-else style="color: #909399;">无标签</span>
            </div>

            <!-- Taints -->
            <div style="margin-top: 24px;">
              <div class="section-header">
                <el-icon class="section-icon"><CircleClose /></el-icon>
                <span class="section-title">污点</span>
                <el-button size="small" @click="handleTaints" style="margin-left: auto;">编辑</el-button>
              </div>
              <el-table v-if="node.taints && node.taints.length > 0" :data="node.taints" stripe border>
                <el-table-column prop="key" label="Key" min-width="200" />
                <el-table-column prop="value" label="Value" min-width="120" />
                <el-table-column prop="effect" label="Effect" min-width="150" />
              </el-table>
              <span v-else style="color: #909399;">无污点</span>
            </div>
          </el-card>
        </el-tab-pane>

        <!-- Pods Tab -->
        <el-tab-pane label="Pods" name="pods">
          <el-card shadow="never">
            <div style="margin-bottom: 16px;">
              <el-input v-model="podSearch" placeholder="搜索 Pod 名称、命名空间或 IP" style="width: 300px;" clearable>
                <template #prefix><el-icon><Search /></el-icon></template>
              </el-input>
              <span style="margin-left: 12px; color: #909399; font-size: 13px;">
                共 {{ pods.length }} 个 Pod{{ podSearch ? `，筛选出 ${filteredPods.length} 个` : '' }}
              </span>
            </div>
            <el-table :data="filteredPods" v-loading="podsLoading" stripe>
              <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip>
                <template #default="{ row }"><el-button link type="primary" @click="handlePodDetail(row)">{{ row.name }}</el-button></template>
              </el-table-column>
              <el-table-column prop="namespace" label="命名空间" width="140" />
              <el-table-column prop="status" label="状态" width="120"><template #default="{ row }"><el-tag :type="podStatusType(row.status)" size="small">{{ row.status }}</el-tag></template></el-table-column>
              <el-table-column prop="ip" label="IP" width="140" />
              <el-table-column prop="restarts" label="重启次数" width="100" />
              <el-table-column prop="age" label="年龄" width="120" />
            </el-table>
            <el-empty v-if="!podsLoading && pods.length === 0" description="该节点上暂无 Pod" />
          </el-card>
        </el-tab-pane>

        <!-- Events Tab -->
        <el-tab-pane label="事件" name="events">
          <el-card shadow="never">
            <el-table :data="events" v-loading="eventsLoading" stripe>
              <el-table-column prop="type" label="类型" width="100"><template #default="{ row }"><el-tag :type="row.type === 'Warning' ? 'danger' : 'info'" size="small">{{ row.type }}</el-tag></template></el-table-column>
              <el-table-column prop="reason" label="原因" width="150" />
              <el-table-column prop="message" label="消息" min-width="300" show-overflow-tooltip />
              <el-table-column prop="last_seen" label="最后发生" width="180" />
            </el-table>
            <el-empty v-if="!eventsLoading && events.length === 0" description="暂无事件" />
          </el-card>
        </el-tab-pane>

        <!-- YAML Tab -->
        <el-tab-pane label="YAML" name="yaml">
          <el-card shadow="never">
            <div style="margin-bottom: 12px; display: flex; gap: 8px;">
              <el-button v-if="!yamlEditing" type="primary" @click="yamlEditing = true">编辑 YAML</el-button>
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
.page-container { padding: 20px; position: relative; min-height: 100%; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }

/* 资源容量区域样式 */
.resource-section {
  margin-top: 24px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 2px solid var(--el-color-primary);
}

.section-icon {
  font-size: 20px;
  color: var(--el-color-primary);
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  letter-spacing: 1px;
}

/* 资源卡片布局 */
.resource-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 16px;
}

.resource-card {
  display: flex;
  align-items: center;
  padding: 20px;
  background: linear-gradient(135deg, var(--el-fill-color-lighter) 0%, var(--el-fill-color-light) 100%);
  border-radius: 12px;
  border: 1px solid var(--el-border-color-lighter);
  transition: all 0.3s ease;
}

.resource-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--el-box-shadow-light);
  border-color: var(--el-color-primary-light-5);
}

/* 资源图标 */
.resource-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16px;
  font-size: 24px;
  color: #fff;
  flex-shrink: 0;
}

.cpu-icon {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.memory-icon {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.pods-icon {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.storage-icon {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
}

/* 资源信息 */
.resource-info {
  flex: 1;
  min-width: 0;
}

.resource-label {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-regular);
  margin-bottom: 12px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.resource-values {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.value-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.value-label {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.value-number {
  font-size: 16px;
  font-weight: 700;
  color: var(--el-text-color-primary);
  font-variant-numeric: tabular-nums;
}

.value-number.highlight {
  color: var(--el-color-primary);
  font-size: 18px;
}
</style>
