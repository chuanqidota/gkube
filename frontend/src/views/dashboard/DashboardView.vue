<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'
import { getOverview, getResources, getWorkloads, getNamespaceResources, getHealth } from '@/api/dashboard'
import type { Overview, ResourceMetrics, WorkloadSummary, NamespaceUsage, ClusterHealth } from '@/api/dashboard'
import { getNodeList, type NodeInfo } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import { usagePercent } from '@/utils/helpers'

const router = useRouter()
const { t } = useI18n()
const clusterStore = useClusterStore()

const clusterId = computed(() => clusterStore.currentCluster?.id)
const clusterName = computed(() => clusterStore.currentCluster?.displayName || clusterStore.currentCluster?.clusterName || '-')

// ---- loading ----
const resourcesLoading = ref(false)
const overviewLoading = ref(false)
const workloadsLoading = ref(false)
const nodesLoading = ref(false)
const nsLoading = ref(false)
const healthLoading = ref(false)
const loading = computed(() => resourcesLoading.value || nodesLoading.value || nsLoading.value || healthLoading.value)

// ---- state ----
const overview = ref<Overview>({ cluster_count: 0, node_count: 0, pod_count: 0, namespace_count: 0 })
const resources = ref<ResourceMetrics>({ cpu: { used: 0, total: 0 }, memory: { used: 0, total: 0 }, storage: { used: 0, total: 0 } })
const workloads = ref<WorkloadSummary>({ deployments: 0, statefulsets: 0, daemonsets: 0, jobs: 0, cronjobs: 0, ingresses: 0 })
const nodeList = ref<NodeInfo[]>([])
const nsList = ref<NamespaceUsage[]>([])
const nsTotal = ref({ cpu: 0, mem: 0 })
const sortKey = ref<'cpu' | 'mem'>('cpu')
const health = ref<ClusterHealth | null>(null)

function fmtCpu(n: number) { return (n || 0).toFixed(2) }
function fmtMem(n: number) { return (n || 0).toFixed(1) }

// 读取 CSS token(ECharts 需具体色值,运行时读取保证主题一致)
function tk(name: string) {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim()
}
function threshColor(pct: number) {
  if (pct >= 90) return tk('--gk-color-danger')
  if (pct >= 70) return tk('--gk-color-warning')
  return tk('--gk-color-primary')
}
function readyColor(pct: number) {
  if (pct >= 90) return tk('--gk-color-success')
  if (pct >= 70) return tk('--gk-color-warning')
  return tk('--gk-color-danger')
}

const sortedNamespaces = computed(() => {
  const list = [...nsList.value]
  list.sort((a, b) => (sortKey.value === 'cpu' ? b.cpu_used - a.cpu_used : b.mem_used - a.mem_used))
  return list
})

// 命名空间柱状图:默认只展示 Top N(按当前排序),超量提供「查看全部」入口。
// 避免几十个 ns 挤成细线;Top N 已能反映资源消耗分布。
const BAR_TOP_N = 8
const displayNamespaces = computed(() => sortedNamespaces.value.slice(0, BAR_TOP_N))
const hasMoreNamespaces = computed(() => sortedNamespaces.value.length > BAR_TOP_N)

// 柱状图高度:按展示条数驱动,有上下限。
const BAR_ROW = 28
const BAR_PAD = 44
const barHeight = computed(() => {
  const n = displayNamespaces.value.length
  if (n === 0) return 120
  return n * BAR_ROW + BAR_PAD
})

function clusterStatusType(status: string) {
  if (status === 'online' || status === 'connected' || status === 'healthy') return 'success'
  if (status === 'offline' || status === 'disconnected' || status === 'unhealthy') return 'danger'
  return 'info'
}
function clusterStatusText(status: string) {
  if (status === 'online' || status === 'connected') return t('cluster.online')
  if (status === 'offline' || status === 'disconnected') return t('cluster.offline')
  return status || t('common.unknown')
}

const readyCount = computed(() => nodeList.value.filter((n) => n.is_ready).length)

// stat 胶囊(命令栏内联)
const stats = computed(() => [
  { label: t('dashboard.nodeCount'), value: overview.value.node_count, route: '/nodes' },
  { label: t('dashboard.podCount'), value: overview.value.pod_count, route: '/workloads/pods' },
  { label: t('dashboard.namespaceCount'), value: overview.value.namespace_count, route: '/namespaces' },
  { label: t('workload.deployment'), value: workloads.value.deployments, route: '/workloads/deployments' },
  { label: t('network.ingress'), value: workloads.value.ingresses, route: '/network/ingresses' },
])

// 容量环(口径:已分配 / Allocatable,语义为分配率)
const rings = computed(() => [
  { key: 'cpu', label: t('dashboard.cpuAllocated'), used: resources.value.cpu.used, total: resources.value.cpu.total, unit: t('dashboard.cores'), fmt: fmtCpu, color: threshColor },
  { key: 'mem', label: t('dashboard.memAllocated'), used: resources.value.memory.used, total: resources.value.memory.total, unit: 'GiB', fmt: fmtMem, color: threshColor },
  { key: 'storage', label: t('dashboard.storageAllocated'), used: resources.value.storage.used, total: resources.value.storage.total, unit: 'GiB', fmt: fmtMem, color: threshColor },
  { key: 'ready', label: t('dashboard.nodeReadiness'), used: readyCount.value, total: nodeList.value.length, unit: '', fmt: (n: number) => String(n || 0), color: readyColor },
])

// 健康派生
const abnormalPods = computed(() => health.value?.abnormal_pods || [])
const restartingPods = computed(() => health.value?.restarting_pods || [])
const issueCount = computed(() => (health.value?.summary.abnormal_pods || 0) + (health.value?.summary.not_ready_nodes || 0) + (health.value?.summary.abnormal_pvcs || 0) + restartingPods.value.length)
const allHealthy = computed(() => issueCount.value === 0)

// 健康数字组:abnormal 标异常(ok 用其反面)
const healthStats = computed(() => {
  const s = health.value?.summary
  return [
    { label: t('dashboard.healthyPods'), value: s?.healthy_pods || 0, abnormal: false, ok: true, route: '/workloads/pods' },
    { label: t('dashboard.abnormalPods'), value: s?.abnormal_pods || 0, abnormal: (s?.abnormal_pods || 0) > 0, ok: false, route: '/workloads/pods' },
    { label: t('dashboard.readyNodes'), value: s?.ready_nodes || 0, abnormal: false, ok: true, route: '/nodes' },
    { label: t('dashboard.notReadyNodes'), value: s?.not_ready_nodes || 0, abnormal: (s?.not_ready_nodes || 0) > 0, ok: false, route: '/nodes' },
    { label: t('dashboard.boundPvcs'), value: s?.bound_pvcs || 0, abnormal: false, ok: true, route: '/storage/pvcs' },
    { label: t('dashboard.abnormalPvcs'), value: s?.abnormal_pvcs || 0, abnormal: (s?.abnormal_pvcs || 0) > 0, ok: false, route: '/storage/pvcs' },
  ]
})

// ---- fetch ----
async function fetchOverview() {
  overviewLoading.value = true
  try { const res = await getOverview({ clusterId: clusterId.value }); overview.value = res.data }
  catch (e: any) { ElMessage.error(e?.message || t('dashboard.loadFailed')) }
  finally { overviewLoading.value = false }
}
async function fetchResources() {
  resourcesLoading.value = true
  try { const res = await getResources({ clusterId: clusterId.value }); resources.value = res.data }
  catch (e: any) { ElMessage.error(e?.message || t('dashboard.loadFailed')) }
  finally { resourcesLoading.value = false }
}
async function fetchWorkloads() {
  workloadsLoading.value = true
  try { const res = await getWorkloads({ clusterId: clusterId.value }); workloads.value = res.data }
  catch (e: any) { ElMessage.error(e?.message || t('dashboard.loadFailed')) }
  finally { workloadsLoading.value = false }
}
async function fetchNodes() {
  nodesLoading.value = true
  try { const res: any = await getNodeList(); nodeList.value = res.data || [] }
  catch { /* 节点不可达时静默处理 */ }
  finally { nodesLoading.value = false }
}
async function fetchNamespaces() {
  nsLoading.value = true
  try {
    const res = await getNamespaceResources({ clusterId: clusterId.value })
    nsList.value = res.data.namespaces || []
    nsTotal.value = { cpu: res.data.total_cpu, mem: res.data.total_mem }
  } catch { nsList.value = []; nsTotal.value = { cpu: 0, mem: 0 } }
  finally { nsLoading.value = false }
}
async function fetchHealth() {
  healthLoading.value = true
  try { const res = await getHealth({ clusterId: clusterId.value }); health.value = res.data }
  catch { health.value = null }
  finally { healthLoading.value = false }
}

async function fetchAll() {
  if (!clusterId.value) return
  await Promise.all([fetchOverview(), fetchResources(), fetchWorkloads(), fetchNodes(), fetchNamespaces(), fetchHealth()])
  nextTick(updateAllCharts)
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchAll, { interval: 15000, autoStart: false })

watch(clusterId, (val) => { if (val) fetchAll() })

// ===== ECharts =====
const ringRefs = ref<Record<string, HTMLElement | null>>({})
const barRef = ref<HTMLElement | null>(null)
const charts: Record<string, echarts.ECharts | null> = {}
let themeObserver: MutationObserver | null = null

function setRingRef(key: string) {
  return (el: any) => { ringRefs.value[key] = el }
}

function initRing(el: HTMLElement): echarts.ECharts {
  const chart = echarts.init(el)
  chart.setOption({
    series: [{
      type: 'gauge',
      startAngle: 90,
      endAngle: -270,
      radius: '90%',
      pointer: { show: false },
      progress: { show: true, overlap: false, roundCap: true, clip: false, width: 10 },
      axisLine: { lineStyle: { width: 10, color: [[1, tk('--gk-color-border-light')]] } },
      splitLine: { show: false },
      axisTick: { show: false },
      axisLabel: { show: false },
      data: [{ value: 0 }],
      detail: { valueAnimation: true, formatter: '{value}%', fontSize: 22, fontWeight: 700, offsetCenter: [0, '5%'], color: tk('--gk-color-primary'), fontFamily: 'JetBrains Mono, monospace' },
    }],
  })
  return chart
}

function updateRing(key: string, pct: number, colorFn: (p: number) => string) {
  const chart = charts[key]
  if (!chart) return
  const color = colorFn(pct)
  chart.setOption({
    series: [{
      progress: { itemStyle: { color, shadowBlur: 10, shadowColor: color } },
      axisLine: { lineStyle: { color: [[1, tk('--gk-color-border-light')]] } },
      data: [{ value: pct }],
      detail: { color },
    }],
  })
}

function updateBar() {
  const chart = charts.bar
  if (!chart) return
  const list = displayNamespaces.value
  const names = list.map((n) => n.name)
  const values = list.map((n) => (sortKey.value === 'cpu' ? n.cpu_used : n.mem_used))
  // max 基于全量(非仅 Top N),使条长比例反映真实差距
  const fullMax = Math.max(...sortedNamespaces.value.map((n) => (sortKey.value === 'cpu' ? n.cpu_used : n.mem_used)), 0.001)
  const primary = tk('--gk-color-primary')
  const primaryLight = tk('--gk-color-primary-light')
  const secondary = tk('--gk-color-text-secondary')
  chart.setOption({
    grid: { left: 8, right: 58, top: 8, bottom: 8, containLabel: true },
    xAxis: { type: 'value', show: false, max: fullMax * 1.2 },
    yAxis: {
      type: 'category', inverse: true, data: names,
      axisLine: { show: false }, axisTick: { show: false },
      axisLabel: { color: secondary, fontSize: 12, margin: 12 },
    },
    series: [{
      type: 'bar', data: values, barWidth: 14,
      itemStyle: {
        borderRadius: 7,
        color: new echarts.graphic.LinearGradient(0, 0, 1, 0, [
          { offset: 0, color: primaryLight }, { offset: 1, color: primary },
        ]),
        shadowBlur: 6, shadowColor: primary + '33',
      },
      label: {
        show: true, position: 'right',
        formatter: (p: any) => sortKey.value === 'cpu' ? fmtCpu(p.value) + ' 核' : fmtMem(p.value) + 'G',
        color: secondary, fontSize: 11, fontFamily: 'JetBrains Mono, monospace',
      },
    }],
  }, true)
}

function updateAllCharts() {
  updateRing('cpu', usagePercent(resources.value.cpu.used, resources.value.cpu.total), threshColor)
  updateRing('mem', usagePercent(resources.value.memory.used, resources.value.memory.total), threshColor)
  updateRing('storage', usagePercent(resources.value.storage.used, resources.value.storage.total), threshColor)
  updateRing('ready', usagePercent(readyCount.value, nodeList.value.length), readyColor)
  updateBar()
}

function initCharts() {
  for (const key of ['cpu', 'mem', 'storage', 'ready']) {
    const el = ringRefs.value[key]
    if (el) charts[key] = initRing(el)
  }
  if (barRef.value) charts.bar = echarts.init(barRef.value)
  updateAllCharts()
}

function handleResize() { Object.values(charts).forEach((c) => c?.resize()) }

onMounted(() => {
  if (clusterId.value) fetchAll()
  nextTick(initCharts)
  window.addEventListener('resize', handleResize)
  themeObserver = new MutationObserver(() => nextTick(updateAllCharts))
  themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] })
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  themeObserver?.disconnect()
  Object.values(charts).forEach((c) => c?.dispose())
})

watch([resources, () => readyCount.value, nsList, sortKey], () => nextTick(updateAllCharts))

// 柱状图高度随 ns 数量变化时,resize ECharts
watch(barHeight, () => nextTick(() => charts.bar?.resize()))

function goto(route: string) { router.push(route) }
function nodePipClass(n: NodeInfo) {
  if (n.status === 'Ready') return 'success'
  if (n.status === 'NotReady') return 'danger'
  return 'warning'
}
</script>

<template>
  <div class="dash">
    <div class="atmosphere" />

    <!-- 1. 命令栏 + 内联 stat -->
    <header class="cmd-bar">
      <div class="cmd-left">
        <span class="cmd-accent" />
        <h1 class="cmd-name">{{ clusterName }}</h1>
        <span class="pip" :class="`pip-${clusterStatusType(clusterStore.currentCluster?.status || '')}`">
          <i class="pip-dot" />{{ clusterStatusText(clusterStore.currentCluster?.status || '') }}
        </span>
        <span v-if="clusterStore.currentCluster?.clusterVersion" class="chip-mono">v{{ clusterStore.currentCluster.clusterVersion }}</span>
      </div>
      <div class="cmd-stats">
        <button v-for="s in stats" :key="s.label" class="stat-pill" @click="goto(s.route)">
          <span class="stat-value mono">{{ s.value }}</span>
          <span class="stat-label">{{ s.label }}</span>
        </button>
      </div>
      <div class="cmd-actions">
        <AutoRefreshToolbar
          :is-running="isRunning" :countdown="countdown" :current-interval="currentInterval"
          :available-intervals="availableIntervals" :loading="loading"
          @refresh="manualRefresh()" @toggle="toggle()" @interval-change="setIntervalOption"
        />
      </div>
    </header>

    <!-- 2. grip 主区(填满剩余视口) -->
    <div class="grip">
      <!-- 左:健康汇总面板 -->
      <section class="cell cell-health" v-loading="healthLoading">
        <div class="cell-head">
          <span class="cell-title">{{ t('dashboard.health') }}</span>
          <span v-if="allHealthy" class="health-badge health-badge-ok">{{ t('dashboard.clusterHealthy') }}</span>
          <span v-else class="health-badge health-badge-warn">{{ t('dashboard.hasIssues', { n: issueCount }) }}</span>
        </div>
        <div class="health-grid">
          <button v-for="hs in healthStats" :key="hs.label" class="health-cell" :class="{ 'health-cell-abnormal': hs.abnormal }" @click="goto(hs.route)">
            <span class="health-num mono" :class="hs.abnormal ? 'num-danger' : (hs.ok ? 'num-ok' : '')">{{ hs.value }}</span>
            <span class="health-label">{{ hs.label }}</span>
          </button>
        </div>
        <div class="issue-list">
          <div v-if="abnormalPods.length" class="issue-group">
            <div class="issue-group-title">{{ t('dashboard.abnormalPodList') }} ({{ health?.summary.abnormal_pods || 0 }})</div>
            <div v-for="(p, i) in abnormalPods" :key="'ab'+i" class="issue-row" @click="goto('/workloads/pods')">
              <i class="issue-dot issue-dot-danger" />
              <span class="issue-name name-link">{{ p.namespace }}/{{ p.name }}</span>
              <span class="issue-reason mono">{{ p.reason || p.phase }}</span>
            </div>
          </div>
          <div v-if="restartingPods.length" class="issue-group">
            <div class="issue-group-title">{{ t('dashboard.restartingPodList') }}</div>
            <div v-for="(p, i) in restartingPods" :key="'rs'+i" class="issue-row" @click="goto('/workloads/pods')">
              <i class="issue-dot issue-dot-warn" />
              <span class="issue-name name-link">{{ p.namespace }}/{{ p.name }}</span>
              <span class="issue-reason mono">{{ t('dashboard.restarts') }} {{ p.restart_count }}</span>
            </div>
          </div>
          <div v-if="!abnormalPods.length && !restartingPods.length" class="issue-empty">
            <i class="issue-ok-dot" />
            <span>{{ t('dashboard.noAbnormal') }}</span>
          </div>
        </div>
      </section>

      <!-- 中:容量环 2x2(分配率) -->
      <section class="cell cell-rings" v-loading="resourcesLoading || nodesLoading">
        <div class="cell-title">{{ t('dashboard.allocationRate') }}</div>
        <div class="rings-grid">
          <div v-for="ring in rings" :key="ring.key" class="ring-cell">
            <div class="ring-chart" :ref="setRingRef(ring.key)"></div>
            <div class="ring-foot">
              <span class="ring-label">{{ ring.label }}</span>
              <span class="ring-nums mono">{{ ring.fmt(ring.used) }}<span class="sep">/</span>{{ ring.fmt(ring.total) }}<span v-if="ring.unit" class="unit"> {{ ring.unit }}</span></span>
            </div>
          </div>
        </div>
      </section>

      <!-- 右:命名空间柱状图(上) + 节点 mini(下) -->
      <section class="cell cell-right">
        <div class="right-top">
          <div class="cell-head">
            <span class="cell-title">{{ t('dashboard.nsDistribution') }}</span>
            <div class="ns-sort">
              <button class="sort-pill" :class="{ 'sort-pill-active': sortKey === 'cpu' }" @click="sortKey = 'cpu'">CPU</button>
              <button class="sort-pill" :class="{ 'sort-pill-active': sortKey === 'mem' }" @click="sortKey = 'mem'">{{ t('dashboard.memAllocated') }}</button>
            </div>
          </div>
          <div class="bar-wrap" v-loading="nsLoading">
            <div class="bar-chart" ref="barRef" :style="{ height: barHeight + 'px' }"></div>
            <div v-if="!sortedNamespaces.length && !nsLoading" class="bar-empty">
              <el-empty :description="t('common.noData')" :image-size="44" />
            </div>
          </div>
          <div v-if="hasMoreNamespaces" class="bar-more" @click="goto('/namespaces')">
            {{ t('dashboard.viewTopN', { shown: displayNamespaces.length, total: sortedNamespaces.length }) }} →
          </div>
        </div>

        <div class="right-bottom">
          <div class="cell-head">
            <span class="cell-title">{{ t('dashboard.nodeCapacity') }}</span>
            <span class="cell-hint">{{ nodeList.length }}</span>
          </div>
          <div class="node-list" v-loading="nodesLoading">
            <div v-if="nodeList.length" class="node-rows">
              <div v-for="node in nodeList" :key="node.name" class="node-row" @click="goto('/nodes')">
                <div class="node-top">
                  <span class="node-dot" :class="nodePipClass(node)" />
                  <span class="node-name name-link">{{ node.name }}</span>
                  <span class="node-pods mono">{{ node.pod_count || 0 }}/{{ node.pod_total || 0 }}</span>
                </div>
                <div class="node-mini-bar">
                  <div class="mini-track">
                    <div class="mini-fill" :style="{ width: usagePercent(node.cpu_used, node.cpu_total) + '%', background: threshColor(usagePercent(node.cpu_used, node.cpu_total)) }" />
                  </div>
                  <span class="mini-num mono">{{ fmtCpu(node.cpu_used) }}<span class="mini-sep">/</span>{{ fmtCpu(node.cpu_total) }}</span>
                </div>
                <div class="node-mini-bar">
                  <div class="mini-track">
                    <div class="mini-fill" :style="{ width: usagePercent(node.mem_used, node.mem_total) + '%', background: threshColor(usagePercent(node.mem_used, node.mem_total)) }" />
                  </div>
                  <span class="mini-num mono">{{ fmtMem(node.mem_used) }}<span class="mini-sep">/</span>{{ fmtMem(node.mem_total) }}G</span>
                </div>
              </div>
            </div>
            <div v-else class="node-empty">
              <el-empty :description="t('common.noData')" :image-size="44" />
            </div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.dash {
  height: calc(100vh - var(--gk-header-height));
  padding: var(--gk-space-4);
  display: flex;
  flex-direction: column;
  gap: var(--gk-space-4);
  position: relative;
  overflow: hidden;
}

.mono {
  font-family: var(--gk-font-mono);
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.01em;
}

/* 大气层:蓝色径向辉光(品牌主色,非 AI 紫) */
.atmosphere {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 0;
  background:
    radial-gradient(circle at 8% -8%, var(--gk-color-primary-bg) 0%, transparent 35%),
    radial-gradient(circle at 98% 6%, var(--gk-color-info-bg) 0%, transparent 30%);
}

.cmd-bar, .grip { position: relative; z-index: 1; }

/* ===== 1. 命令栏 ===== */
.cmd-bar {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: var(--gk-space-4);
  padding: var(--gk-space-3) var(--gk-space-5);
  background: linear-gradient(135deg, var(--gk-color-primary-bg) 0%, var(--gk-color-bg-card) 50%);
  border: 1px solid var(--gk-color-primary-light);
  border-radius: var(--gk-radius-lg);
  box-shadow: var(--gk-shadow-card);
}

.cmd-left {
  display: flex;
  align-items: center;
  gap: var(--gk-space-3);
  min-width: 0;
  flex-shrink: 1;
}

.cmd-accent {
  width: 4px;
  height: 22px;
  border-radius: var(--gk-radius-full);
  background: linear-gradient(180deg, var(--gk-color-primary-light), var(--gk-color-primary));
  box-shadow: 0 0 10px var(--gk-color-primary);
  flex-shrink: 0;
}

.cmd-name {
  font-size: var(--gk-font-size-xl);
  font-weight: 700;
  color: var(--gk-color-text-primary);
  margin: 0;
  line-height: 1.1;
  letter-spacing: -0.02em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.pip {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 2px 10px;
  border-radius: var(--gk-radius-full);
  font-size: var(--gk-font-size-xs);
  font-weight: 600;
  flex-shrink: 0;
}
.pip-dot { width: 6px; height: 6px; border-radius: 50%; display: inline-block; }
.pip-success { color: var(--gk-color-success); background: var(--gk-color-success-bg); }
.pip-success .pip-dot { background: var(--gk-color-success); box-shadow: 0 0 0 3px rgba(34,197,94,0.18); }
.pip-danger { color: var(--gk-color-danger); background: var(--gk-color-danger-bg); }
.pip-danger .pip-dot { background: var(--gk-color-danger); box-shadow: 0 0 0 3px rgba(239,68,68,0.18); }
.pip-info { color: var(--gk-color-text-secondary); background: var(--gk-color-border-light); }
.pip-info .pip-dot { background: var(--gk-color-text-secondary); }

.chip-mono {
  font-family: var(--gk-font-mono);
  font-size: var(--gk-font-size-xs);
  color: var(--gk-color-text-secondary);
  background: var(--gk-color-bg-card);
  border: 1px solid var(--gk-color-border);
  padding: 2px 8px;
  border-radius: var(--gk-radius-sm);
  flex-shrink: 0;
}

.cmd-stats {
  display: flex;
  gap: var(--gk-space-2);
  flex: 1;
  justify-content: center;
  min-width: 0;
}

.stat-pill {
  display: flex;
  flex-direction: column;
  gap: 0;
  padding: 4px 12px;
  background: var(--gk-color-bg-card);
  border: 1px solid var(--gk-color-border);
  border-radius: var(--gk-radius-md);
  box-shadow: var(--gk-shadow-sm);
  cursor: pointer;
  transition: all var(--gk-transition-fast);
  text-align: center;
  min-width: 76px;
}
.stat-pill:hover { border-color: var(--gk-color-primary-light); box-shadow: 0 0 0 3px var(--gk-color-primary-bg); }
.stat-value { font-size: var(--gk-font-size-lg); font-weight: 700; color: var(--gk-color-text-primary); line-height: 1.2; }
.stat-label { font-size: 11px; color: var(--gk-color-text-secondary); }

.cmd-actions { flex-shrink: 0; }

/* ===== 2. grip 主区 ===== */
.grip {
  flex: 1;
  min-height: 0;
  display: grid;
  grid-template-columns: 1.1fr 1.5fr 1.3fr;
  gap: var(--gk-space-4);
}

.cell {
  background: var(--gk-color-bg-card);
  border: 1px solid var(--gk-color-border);
  border-radius: var(--gk-radius-lg);
  box-shadow: var(--gk-shadow-card);
  padding: var(--gk-space-4);
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.cell-title { font-size: var(--gk-font-size-sm); font-weight: 600; color: var(--gk-color-text-primary); letter-spacing: 0.02em; flex-shrink: 0; }
.cell-head { display: flex; align-items: center; justify-content: space-between; gap: var(--gk-space-2); flex-shrink: 0; flex-wrap: nowrap; margin-bottom: var(--gk-space-3); }
.cell-title-row { display: flex; align-items: baseline; gap: var(--gk-space-2); flex-shrink: 1; min-width: 0; }
.cell-hint { font-size: var(--gk-font-size-xs); color: var(--gk-color-text-secondary); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

/* 健康面板 */
.cell-health { background: linear-gradient(180deg, var(--gk-color-primary-bg) 0%, var(--gk-color-bg-card) 60%); border-color: var(--gk-color-primary-light); }

.health-badge {
  padding: 2px 10px;
  border-radius: var(--gk-radius-full);
  font-size: 11px;
  font-weight: 600;
  flex-shrink: 0;
}
.health-badge-ok { color: var(--gk-color-success); background: var(--gk-color-success-bg); }
.health-badge-warn { color: var(--gk-color-danger); background: var(--gk-color-danger-bg); }

.health-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--gk-space-2);
  flex-shrink: 0;
  margin-bottom: var(--gk-space-3);
}

.health-cell {
  display: flex;
  flex-direction: column;
  gap: 1px;
  padding: var(--gk-space-2) var(--gk-space-3);
  background: var(--gk-color-bg-card);
  border: 1px solid var(--gk-color-border);
  border-radius: var(--gk-radius-md);
  cursor: pointer;
  transition: all var(--gk-transition-fast);
  text-align: left;
}
.health-cell:hover { border-color: var(--gk-color-primary-light); }
.health-cell-abnormal { border-color: var(--gk-color-danger-light); background: var(--gk-color-danger-bg); }

.health-num { font-size: var(--gk-font-size-lg); font-weight: 700; line-height: 1.2; color: var(--gk-color-text-primary); }
.num-ok { color: var(--gk-color-success); }
.num-danger { color: var(--gk-color-danger); }
.health-label { font-size: 11px; color: var(--gk-color-text-secondary); }

/* 异常清单 */
.issue-list { flex: 1; min-height: 0; overflow-y: auto; display: flex; flex-direction: column; gap: var(--gk-space-3); }
.issue-group { display: flex; flex-direction: column; gap: 4px; }
.issue-group-title { font-size: 11px; color: var(--gk-color-text-secondary); font-weight: 600; }
.issue-row {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 6px;
  border-radius: var(--gk-radius-sm);
  cursor: pointer;
  transition: background var(--gk-transition-fast);
  font-size: 11px;
}
.issue-row:hover { background: var(--gk-color-primary-bg); }
.issue-dot { width: 6px; height: 6px; border-radius: 50%; flex-shrink: 0; }
.issue-dot-danger { background: var(--gk-color-danger); }
.issue-dot-warn { background: var(--gk-color-warning); }
.issue-name { color: var(--gk-color-text-primary); flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.issue-reason { color: var(--gk-color-text-secondary); flex-shrink: 0; }

.issue-empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  color: var(--gk-color-success);
  font-size: var(--gk-font-size-xs);
}
.issue-ok-dot { width: 7px; height: 7px; border-radius: 50%; background: var(--gk-color-success); display: inline-block; }

/* 容量环 2x2 */
.cell-rings { background: linear-gradient(180deg, var(--gk-color-primary-bg) 0%, var(--gk-color-bg-card) 60%); border-color: var(--gk-color-primary-light); }
.rings-grid {
  flex: 1;
  display: grid;
  grid-template-columns: 1fr 1fr;
  grid-template-rows: 1fr 1fr;
  gap: var(--gk-space-2);
  min-height: 0;
}
.ring-cell {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  background: var(--gk-color-bg-card);
  border: 1px solid var(--gk-color-border);
  border-radius: var(--gk-radius-md);
  padding: var(--gk-space-2);
  min-height: 0;
}
.ring-chart { width: 100%; flex: 1; min-height: 0; }
.ring-foot { display: flex; flex-direction: column; align-items: center; gap: 1px; }
.ring-label { font-size: 11px; color: var(--gk-color-text-secondary); }
.ring-nums { font-size: var(--gk-font-size-xs); color: var(--gk-color-text-primary); }
.ring-nums .sep { color: var(--gk-color-text-placeholder); margin: 0 2px; }
.ring-nums .unit { color: var(--gk-color-text-placeholder); }

/* 右列:上下分栏,柱状图区内容驱动、节点区吃剩余空间 */
.cell-right {
  padding: 0;
  background: linear-gradient(180deg, var(--gk-color-primary-bg) 0%, var(--gk-color-bg-card) 60%);
  border-color: var(--gk-color-primary-light);
}
.right-top { flex: 0 1 auto; min-height: 0; display: flex; flex-direction: column; padding: var(--gk-space-4); border-bottom: 1px solid var(--gk-color-border-light); }
.right-bottom { flex: 1; min-height: 0; display: flex; flex-direction: column; padding: var(--gk-space-4); }

.bar-wrap { position: relative; min-height: 0; }
.bar-chart { width: 100%; }
.bar-empty { position: absolute; inset: 0; display: flex; align-items: center; justify-content: center; }
.bar-more {
  flex-shrink: 0;
  margin-top: var(--gk-space-2);
  font-size: 11px;
  color: var(--gk-color-primary);
  cursor: pointer;
  text-align: right;
  transition: opacity var(--gk-transition-fast);
}
.bar-more:hover { opacity: 0.75; }

.ns-sort { display: flex; gap: 4px; flex-shrink: 0; }
.sort-pill {
  background: transparent;
  border: 1px solid var(--gk-color-border);
  color: var(--gk-color-text-secondary);
  padding: 3px 10px;
  border-radius: var(--gk-radius-full);
  font-size: 11px;
  cursor: pointer;
  transition: all var(--gk-transition-fast);
}
.sort-pill:hover { border-color: var(--gk-color-primary-light); color: var(--gk-color-primary); }
.sort-pill-active { background: var(--gk-color-primary-bg); border-color: var(--gk-color-primary); color: var(--gk-color-primary); font-weight: 600; }

/* 节点迷你条列表 */
.node-list { flex: 1; min-height: 0; overflow-y: auto; display: flex; flex-direction: column; gap: var(--gk-space-2); }
.node-rows { display: flex; flex-direction: column; gap: var(--gk-space-2); }
.node-row {
  padding: var(--gk-space-2) var(--gk-space-3);
  background: var(--gk-color-bg-page);
  border: 1px solid var(--gk-color-border-light);
  border-radius: var(--gk-radius-md);
  cursor: pointer;
  transition: all var(--gk-transition-fast);
  display: flex;
  flex-direction: column;
  gap: 5px;
}
.node-row:hover { border-color: var(--gk-color-primary-light); background: var(--gk-color-primary-bg); }

.node-top { display: flex; align-items: center; gap: 6px; }
.node-dot { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
.node-dot.success { background: var(--gk-color-success); box-shadow: 0 0 0 2px rgba(34,197,94,0.18); }
.node-dot.danger { background: var(--gk-color-danger); box-shadow: 0 0 0 2px rgba(239,68,68,0.18); }
.node-dot.warning { background: var(--gk-color-warning); box-shadow: 0 0 0 2px rgba(245,158,11,0.18); }
.node-name { font-size: var(--gk-font-size-xs); font-weight: 500; color: var(--gk-color-text-primary); flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.node-pods { font-size: 11px; color: var(--gk-color-text-secondary); }

.node-mini-bar { display: flex; align-items: center; gap: 6px; }
.mini-track { flex: 1; height: 4px; background: var(--gk-color-border-light); border-radius: var(--gk-radius-full); overflow: hidden; }
.mini-fill { height: 100%; border-radius: var(--gk-radius-full); transition: width var(--gk-transition-slow); }
.mini-num { font-size: 10px; color: var(--gk-color-text-secondary); flex-shrink: 0; }
.mini-num .mini-sep { color: var(--gk-color-text-placeholder); margin: 0 1px; }

.name-link { color: var(--gk-color-primary); }
.node-empty { flex: 1; display: flex; align-items: center; justify-content: center; }

/* ===== 响应:窄屏退化为单列滚动 ===== */
@media (max-width: 1200px) {
  .dash { height: auto; min-height: calc(100vh - var(--gk-header-height)); overflow: visible; }
  .grip { grid-template-columns: 1fr; }
  .cmd-stats { display: none; }
  .ring-chart { min-height: 130px; }
  .bar-chart { min-height: 180px; }
}
</style>
