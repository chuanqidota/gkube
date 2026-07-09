<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Monitor, Cpu, Grid, Warning, PriceTag, Connection, Document } from '@element-plus/icons-vue'
import { getPodDetail, getPodYaml, updatePodYaml, deletePod, getPodEvents, calcAge } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const pod = ref<any>(null)

/**
 * Transform raw K8s Pod object into flat display format for the detail page.
 */
function transformPodDetail(raw: any): any {
  if (!raw) return null
  const restarts = (raw.status?.containerStatuses || []).reduce(
    (sum: number, cs: any) => sum + (cs.restartCount || 0), 0
  )
  // Merge spec containers with status containerStatuses
  const specContainers = raw.spec?.containers || []
  const statusContainers = raw.status?.containerStatuses || []
  const containers = specContainers.map((spec: any) => {
    const status = statusContainers.find((s: any) => s.name === spec.name) || {}
    let state = 'Unknown'
    let stateReason = ''
    let exitCode: number | undefined
    if (status.state?.running) state = 'Running'
    else if (status.state?.waiting) { state = 'Waiting'; stateReason = status.state.waiting.reason || '' }
    else if (status.state?.terminated) { state = 'Terminated'; exitCode = status.state.terminated.exitCode }
    return {
      name: spec.name,
      image: spec.image,
      ready: status.ready || false,
      restartCount: status.restartCount || 0,
      state,
      stateReason,
      exitCode,
      ports: spec.ports || [],
      env: spec.env || [],
      volumeMounts: spec.volumeMounts || [],
      livenessProbe: spec.livenessProbe,
      readinessProbe: spec.readinessProbe,
    }
  })
  return {
    name: raw.metadata?.name || '',
    namespace: raw.metadata?.namespace || '',
    status: raw.status?.phase || 'Unknown',
    ip: raw.status?.podIP || '',
    host_ip: raw.status?.hostIP || '',
    node: raw.spec?.nodeName || '',
    restarts,
    qos_class: raw.status?.qosClass || '',
    priority: raw.spec?.priority ?? null,
    age: calcAge(raw.metadata?.creationTimestamp),
    created_at: raw.metadata?.creationTimestamp || '',
    service_account: raw.spec?.serviceAccountName || '',
    labels: raw.metadata?.labels || {},
    annotations: raw.metadata?.annotations || {},
    conditions: (raw.status?.conditions || []).map((c: any) => ({
      type: c.type,
      status: c.status,
      reason: c.reason || '',
      message: c.message || '',
      last_transition_time: c.lastTransitionTime || '',
    })),
    containers,
  }
}

const events = ref<any[]>([])
const yamlContent = ref('')
const yamlLoading = ref(false)
const yamlSaving = ref(false)
const activeTab = ref('info')

const namespace = route.params.namespace as string
const name = route.params.name as string

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getPodDetail({ namespace, name })
    pod.value = transformPodDetail(res.data)
  } catch (e: any) {
    ElMessage.error(e?.message || '加载 Pod 详情失败')
  } finally {
    loading.value = false
  }
}

async function fetchEvents() {
  try {
    const res: any = await getPodEvents({ namespace, name })
    events.value = res.data || []
  } catch { /* ignore */ }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getPodYaml({ namespace, name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || '加载 YAML 失败')
  } finally {
    yamlLoading.value = false
  }
}

function handleTabChange(tab: string) {
  if (tab === 'yaml' && !yamlContent.value) fetchYaml()
  if (tab === 'events' && events.value.length === 0) fetchEvents()
}

async function handleSaveYaml() {
  yamlSaving.value = true
  try {
    await updatePodYaml({ namespace, name, yaml: yamlContent.value })
    ElMessage.success('YAML 保存成功')
    fetchDetail()
  } catch (e: any) {
    ElMessage.error(e?.message || '保存 YAML 失败')
  } finally {
    yamlSaving.value = false
  }
}

function getClusterName(): string {
  try {
    const saved = localStorage.getItem('gkube_cluster')
    if (saved) {
      const c = JSON.parse(saved)
      return c?.clusterName || c?.cluster_name || c?.name || ''
    }
  } catch { /* ignore */ }
  return ''
}

function handleLogs() {
  const cluster = getClusterName()
  window.open(`/fullscreen/logs?namespace=${namespace}&pod=${name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
}

function handleExec() {
  const cluster = getClusterName()
  window.open(`/fullscreen/terminal?namespace=${namespace}&pod=${name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(`确定要删除 Pod "${name}"（命名空间：${namespace}）吗？`, '确认删除', { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' })
    await deletePod({ namespace, name })
    ElMessage.success('Pod 已删除')
    router.push('/workloads/pods')
  } catch { /* cancelled */ }
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
  if ((status || '').toLowerCase() === 'true') return 'success'
  if ((status || '').toLowerCase() === 'false') return 'danger'
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
  if (container.state === 'Running') return '运行中'
  if (container.state === 'Waiting') return `等待中 (${container.stateReason || '-'})`
  if (container.state === 'Terminated') return `已终止 (${container.exitCode ?? '-'})`
  return container.state || '-'
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchDetail, { autoStart: false })

onMounted(fetchDetail)
</script>

<template>
  <div class="page-container" v-loading="loading">
    <div class="page-header">
      <h2 style="margin: 0;">Pod: {{ name }}</h2>
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
        <el-button type="primary" @click="handleLogs">日志</el-button>
        <el-button type="success" @click="handleExec">终端</el-button>
        <el-button type="danger" @click="handleDelete">删除</el-button>
        <el-button @click="router.push('/workloads/pods')">返回列表</el-button>
      </div>
    </div>

    <template v-if="pod">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <!-- 概览 -->
        <el-tab-pane label="概览" name="info">
          <el-card shadow="never">
            <!-- 基本信息 -->
            <div class="section-header">
              <el-icon class="section-icon"><Monitor /></el-icon>
              <span class="section-title">基本信息</span>
            </div>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="名称">{{ pod.name }}</el-descriptions-item>
              <el-descriptions-item label="命名空间">{{ pod.namespace }}</el-descriptions-item>
              <el-descriptions-item label="状态"><el-tag :type="statusType(pod.status)" size="small">{{ pod.status }}</el-tag></el-descriptions-item>
              <el-descriptions-item label="Pod IP">{{ pod.ip || '-' }}</el-descriptions-item>
              <el-descriptions-item label="节点">{{ pod.node || '-' }}</el-descriptions-item>
              <el-descriptions-item label="主机 IP">{{ pod.host_ip || '-' }}</el-descriptions-item>
              <el-descriptions-item label="QoS 类别">{{ pod.qos_class || '-' }}</el-descriptions-item>
              <el-descriptions-item label="优先级">{{ pod.priority ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="服务账号">{{ pod.service_account || '-' }}</el-descriptions-item>
              <el-descriptions-item label="重启次数">{{ pod.restarts ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="年龄">{{ pod.age || '-' }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ pod.created_at || '-' }}</el-descriptions-item>
            </el-descriptions>

            <!-- 网络信息 -->
            <div class="resource-section">
              <div class="section-header">
                <el-icon class="section-icon"><Connection /></el-icon>
                <span class="section-title">网络信息</span>
              </div>
              <div class="resource-cards">
                <div class="resource-card">
                  <div class="resource-icon network-icon">
                    <el-icon><Connection /></el-icon>
                  </div>
                  <div class="resource-info">
                    <div class="resource-label">Pod IP</div>
                    <div class="resource-values">
                      <div class="value-item">
                        <span class="value-number highlight">{{ pod.ip || '-' }}</span>
                      </div>
                    </div>
                  </div>
                </div>
                <div class="resource-card">
                  <div class="resource-icon host-icon">
                    <el-icon><Monitor /></el-icon>
                  </div>
                  <div class="resource-info">
                    <div class="resource-label">主机 IP</div>
                    <div class="resource-values">
                      <div class="value-item">
                        <span class="value-number">{{ pod.host_ip || '-' }}</span>
                      </div>
                    </div>
                  </div>
                </div>
                <div class="resource-card">
                  <div class="resource-icon node-icon">
                    <el-icon><Cpu /></el-icon>
                  </div>
                  <div class="resource-info">
                    <div class="resource-label">所在节点</div>
                    <div class="resource-values">
                      <div class="value-item">
                        <span class="value-number">{{ pod.node || '-' }}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 容器状态概览 -->
            <div v-if="pod.containers && pod.containers.length > 0" class="resource-section">
              <div class="section-header">
                <el-icon class="section-icon"><Grid /></el-icon>
                <span class="section-title">容器状态</span>
              </div>
              <div class="resource-cards">
                <div class="resource-card" v-for="c in pod.containers" :key="c.name">
                  <div class="resource-icon" :class="c.ready ? 'ready-icon' : 'not-ready-icon'">
                    <el-icon><Grid /></el-icon>
                  </div>
                  <div class="resource-info">
                    <div class="resource-label">{{ c.name }}</div>
                    <div class="resource-values">
                      <div class="value-item">
                        <span class="value-label">状态</span>
                        <el-tag :type="containerStateType(c.state)" size="small">{{ getContainerStateLabel(c) }}</el-tag>
                      </div>
                      <div class="value-item">
                        <span class="value-label">就绪</span>
                        <el-tag :type="c.ready ? 'success' : 'danger'" size="small">{{ c.ready ? '是' : '否' }}</el-tag>
                      </div>
                      <div class="value-item">
                        <span class="value-label">重启</span>
                        <span class="value-number">{{ c.restartCount }}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 标签 -->
            <div class="resource-section">
              <div class="section-header">
                <el-icon class="section-icon"><PriceTag /></el-icon>
                <span class="section-title">标签</span>
              </div>
              <div v-if="pod.labels && Object.keys(pod.labels).length > 0">
                <el-tag v-for="(val, key) in pod.labels" :key="key" style="margin-right: 8px; margin-bottom: 8px;">{{ key }}={{ val }}</el-tag>
              </div>
              <span v-else style="color: #909399;">无标签</span>
            </div>

            <!-- 注解 -->
            <div v-if="pod.annotations && Object.keys(pod.annotations).length > 0" class="resource-section">
              <div class="section-header">
                <el-icon class="section-icon"><Document /></el-icon>
                <span class="section-title">注解</span>
              </div>
              <div class="annotation-list">
                <div v-for="(val, key) in pod.annotations" :key="key" class="annotation-item">
                  <span class="annotation-key">{{ key }}</span>
                  <span class="annotation-value">{{ val }}</span>
                </div>
              </div>
            </div>

            <!-- Pod 条件 -->
            <div v-if="pod.conditions && pod.conditions.length > 0" class="resource-section">
              <div class="section-header">
                <el-icon class="section-icon"><Warning /></el-icon>
                <span class="section-title">Pod 条件</span>
              </div>
              <el-table :data="pod.conditions" border stripe>
                <el-table-column prop="type" label="类型" min-width="160" />
                <el-table-column label="状态" width="100"><template #default="{ row }"><el-tag :type="conditionStatusType(row.status)" size="small">{{ row.status }}</el-tag></template></el-table-column>
                <el-table-column prop="reason" label="原因" min-width="160" show-overflow-tooltip />
                <el-table-column prop="message" label="消息" min-width="250" show-overflow-tooltip />
                <el-table-column prop="last_transition_time" label="最后变更" width="180" />
              </el-table>
            </div>
          </el-card>
        </el-tab-pane>

        <!-- 容器 -->
        <el-tab-pane label="容器" name="containers">
          <el-card shadow="never">
            <el-table :data="pod.containers || []" border stripe row-key="name">
              <el-table-column type="expand">
                <template #default="{ row }">
                  <div style="padding: 12px 16px;">
                    <div v-if="row.ports && row.ports.length > 0" style="margin-bottom: 16px;">
                      <div class="section-header" style="margin-bottom: 12px;">
                        <span class="section-title" style="font-size: 14px;">端口</span>
                      </div>
                      <el-table :data="row.ports" border size="small">
                        <el-table-column prop="name" label="名称" width="120" />
                        <el-table-column prop="containerPort" label="容器端口" width="130" />
                        <el-table-column prop="protocol" label="协议" width="100" />
                      </el-table>
                    </div>
                    <div v-if="row.env && row.env.length > 0" style="margin-bottom: 16px;">
                      <div class="section-header" style="margin-bottom: 12px;">
                        <span class="section-title" style="font-size: 14px;">环境变量</span>
                      </div>
                      <el-table :data="row.env" border size="small">
                        <el-table-column prop="name" label="名称" min-width="180" />
                        <el-table-column label="值" min-width="250">
                          <template #default="{ row: envRow }">
                            <span v-if="envRow.value !== undefined && envRow.value !== ''">{{ envRow.value }}</span>
                            <span v-else-if="envRow.valueFrom" style="color: var(--gk-color-text-secondary);">{{ envRow.valueFrom.fieldRef?.fieldPath || envRow.valueFrom.secretKeyRef?.name || envRow.valueFrom.configMapKeyRef?.name || '来自引用' }}</span>
                            <span v-else style="color: var(--gk-color-text-secondary);">-</span>
                          </template>
                        </el-table-column>
                      </el-table>
                    </div>
                    <div v-if="row.volumeMounts && row.volumeMounts.length > 0" style="margin-bottom: 16px;">
                      <div class="section-header" style="margin-bottom: 12px;">
                        <span class="section-title" style="font-size: 14px;">卷挂载</span>
                      </div>
                      <el-table :data="row.volumeMounts" border size="small">
                        <el-table-column prop="name" label="卷名称" min-width="150" />
                        <el-table-column prop="mountPath" label="挂载路径" min-width="200" />
                        <el-table-column prop="subPath" label="子路径" width="150" />
                        <el-table-column label="只读" width="100"><template #default="{ row: vm }"><el-tag :type="vm.readOnly ? 'warning' : 'success'" size="small">{{ vm.readOnly ? '是' : '否' }}</el-tag></template></el-table-column>
                      </el-table>
                    </div>
                    <div v-if="row.livenessProbe" style="margin-bottom: 16px;">
                      <div class="section-header" style="margin-bottom: 12px;">
                        <span class="section-title" style="font-size: 14px;">存活探针</span>
                      </div>
                      <el-descriptions :column="2" border size="small">
                        <el-descriptions-item v-if="row.livenessProbe.httpGet" label="类型">HTTP GET</el-descriptions-item>
                        <el-descriptions-item v-if="row.livenessProbe.httpGet" label="路径">{{ row.livenessProbe.httpGet.path || '/' }}</el-descriptions-item>
                        <el-descriptions-item v-if="row.livenessProbe.httpGet" label="端口">{{ row.livenessProbe.httpGet.port }}</el-descriptions-item>
                        <el-descriptions-item v-if="row.livenessProbe.tcpSocket" label="类型">TCP Socket</el-descriptions-item>
                        <el-descriptions-item v-if="row.livenessProbe.exec" label="类型">Exec</el-descriptions-item>
                        <el-descriptions-item v-if="row.livenessProbe.exec" label="命令">{{ (row.livenessProbe.exec.command || []).join(' ') }}</el-descriptions-item>
                        <el-descriptions-item label="初始延迟">{{ row.livenessProbe.initialDelaySeconds ?? '-' }}s</el-descriptions-item>
                        <el-descriptions-item label="检查周期">{{ row.livenessProbe.periodSeconds ?? '-' }}s</el-descriptions-item>
                      </el-descriptions>
                    </div>
                    <div v-if="row.readinessProbe">
                      <div class="section-header" style="margin-bottom: 12px;">
                        <span class="section-title" style="font-size: 14px;">就绪探针</span>
                      </div>
                      <el-descriptions :column="2" border size="small">
                        <el-descriptions-item v-if="row.readinessProbe.httpGet" label="类型">HTTP GET</el-descriptions-item>
                        <el-descriptions-item v-if="row.readinessProbe.httpGet" label="路径">{{ row.readinessProbe.httpGet.path || '/' }}</el-descriptions-item>
                        <el-descriptions-item v-if="row.readinessProbe.httpGet" label="端口">{{ row.readinessProbe.httpGet.port }}</el-descriptions-item>
                        <el-descriptions-item label="初始延迟">{{ row.readinessProbe.initialDelaySeconds ?? '-' }}s</el-descriptions-item>
                        <el-descriptions-item label="检查周期">{{ row.readinessProbe.periodSeconds ?? '-' }}s</el-descriptions-item>
                      </el-descriptions>
                    </div>
                    <el-empty v-if="(!row.ports || row.ports.length === 0) && (!row.env || row.env.length === 0) && (!row.volumeMounts || row.volumeMounts.length === 0) && !row.livenessProbe && !row.readinessProbe" description="无额外容器详情" :image-size="60" />
                  </div>
                </template>
              </el-table-column>
              <el-table-column prop="name" label="名称" min-width="160" />
              <el-table-column prop="image" label="镜像" min-width="280" show-overflow-tooltip />
              <el-table-column label="就绪" width="80"><template #default="{ row }"><el-tag :type="row.ready ? 'success' : 'danger'" size="small">{{ row.ready ? '是' : '否' }}</el-tag></template></el-table-column>
              <el-table-column prop="restartCount" label="重启次数" width="100" />
              <el-table-column label="状态" width="180"><template #default="{ row }"><el-tag :type="containerStateType(row.state)" size="small">{{ getContainerStateLabel(row) }}</el-tag></template></el-table-column>
            </el-table>
            <el-empty v-if="!pod.containers || pod.containers.length === 0" description="无容器" />
          </el-card>
        </el-tab-pane>

        <!-- 事件 -->
        <el-tab-pane label="事件" name="events">
          <el-card shadow="never">
            <el-table :data="events" border stripe>
              <el-table-column prop="type" label="类型" width="100"><template #default="{ row }"><el-tag :type="row.type === 'Warning' ? 'danger' : 'info'" size="small">{{ row.type }}</el-tag></template></el-table-column>
              <el-table-column prop="reason" label="原因" width="150" />
              <el-table-column prop="message" label="消息" min-width="300" show-overflow-tooltip />
              <el-table-column prop="last_seen" label="最后发生" width="180" />
            </el-table>
            <el-empty v-if="events.length === 0" description="暂无事件" />
          </el-card>
        </el-tab-pane>

        <!-- YAML -->
        <el-tab-pane label="YAML" name="yaml">
          <el-card shadow="never">
            <div v-loading="yamlLoading">
              <YamlEditor
                v-model="yamlContent"
                height="600px"
                show-save-buttons
                :saving="yamlSaving"
                @save="handleSaveYaml"
                @cancel="fetchYaml"
              />
            </div>
          </el-card>
        </el-tab-pane>
      </el-tabs>
    </template>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; position: relative; min-height: 100%; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }

/* 区域标题 */
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

/* 资源卡片区域 */
.resource-section {
  margin-top: 24px;
}

.resource-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
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

.network-icon {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.host-icon {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.node-icon {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.ready-icon {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
}

.not-ready-icon {
  background: linear-gradient(135deg, #f5576c 0%, #ff6b6b 100%);
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
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.resource-values {
  display: flex;
  flex-direction: column;
  gap: 6px;
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

/* 注解列表 */
.annotation-list {
  max-height: 200px;
  overflow-y: auto;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  padding: 8px;
}

.annotation-item {
  display: flex;
  padding: 6px 8px;
  border-bottom: 1px solid var(--el-border-color-extra-light);
  font-size: 13px;
}

.annotation-item:last-child {
  border-bottom: none;
}

.annotation-key {
  font-weight: 600;
  color: var(--el-text-color-primary);
  min-width: 200px;
  flex-shrink: 0;
}

.annotation-value {
  color: var(--el-text-color-secondary);
  word-break: break-all;
}
</style>
