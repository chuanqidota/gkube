<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { useClusterStore } from '@/stores/cluster'
import { getPodDetail, getPodList } from '@/api/resource'
import { getClusterList } from '@/api/cluster'
import { extractNamespaceNames, getNamespaceList } from '@/api/resource'
import { ElMessage } from 'element-plus'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import '@xterm/xterm/css/xterm.css'
import { getWsTicket } from '@/api/auth'

const { t } = useI18n()
const route = useRoute()
const clusterStore = useClusterStore()

interface ClusterOption {
  name: string
  displayName: string
}

const clusters = ref<ClusterOption[]>([])
const namespaces = ref<string[]>([])
const pods = ref<string[]>([])

const selectedCluster = ref('')
const selectedNamespace = ref('')
const selectedPod = ref('')
const selectedContainer = ref('')
const containers = ref<{ name: string; isInit: boolean }[]>([])
const appContainers = computed(() => containers.value.filter((c) => !c.isInit))
const initContainers = computed(() => containers.value.filter((c) => c.isInit))
const skipWatchers = ref(false)

const terminalRef = ref<HTMLDivElement>()
const isConnected = ref(false)

// Whether opened from pod context (with query params) — hide selectors
const isEmbedded = ref(false)

async function initWithQueryParams() {
  const { namespace, pod, container, cluster } = route.query
  if (!namespace || !pod) return
  isEmbedded.value = true
  skipWatchers.value = true

  // Set cluster from query param or store
  if (cluster) {
    const name = cluster as string
    selectedCluster.value = name
    clusterStore.setCurrentCluster({
      clusterName: name,
      cluster_name: name,
      name,
    })
  } else if (clusterStore.clusterName) {
    selectedCluster.value = clusterStore.clusterName
  }

  selectedNamespace.value = namespace as string
  selectedPod.value = pod as string

  // Fetch container list (app + init) for the pod, then pick the target container
  const ok = await fetchContainers()
  if (!ok) {
    // 获取失败已报错,选择器保持空,不误报"Pod 无容器"
  } else if (container) {
    selectedContainer.value = container as string
  } else if (appContainers.value.length > 0) {
    selectedContainer.value = appContainers.value[0].name
  } else if (initContainers.value.length > 0) {
    selectedContainer.value = initContainers.value[0].name
  } else {
    ElMessage.warning('Pod has no containers')
  }

  skipWatchers.value = false
}

let terminal: Terminal | null = null
let fitAddon: FitAddon | null = null
let ws: WebSocket | null = null
// 连接代际:每次 connectTerminal 自增,异步等待 ticket 期间若被更新的连接取代则放弃
let connectGen = 0

async function fetchClusters() {
  try {
    const res: any = await getClusterList({ page: 1, size: 100 })
    const items = res.data?.items || []
    clusters.value = items.map((c: any) => ({
      name: c.clusterName || c.cluster_name || c.name,
      displayName: c.displayName || c.display_name || c.clusterName || c.name,
    }))
  } catch {
    // silently fail
  }
}

async function fetchNamespaces() {
  if (!selectedCluster.value) return
  namespaces.value = []
  selectedNamespace.value = ''
  pods.value = []
  selectedPod.value = ''
  try {
    const res: any = await getNamespaceList()
    namespaces.value = extractNamespaceNames(res.data)
  } catch {
    // silently fail
  }
}

async function fetchPods() {
  if (!selectedCluster.value || !selectedNamespace.value) return
  pods.value = []
  selectedPod.value = ''
  try {
    const res: any = await getPodList({ namespace: selectedNamespace.value })
    const items = Array.isArray(res.data) ? res.data : (res.data?.items || [])
    pods.value = items.map((p: any) => p.metadata?.name || p.name).filter(Boolean)
  } catch {
    // silently fail
  }
}

async function fetchContainers(): Promise<boolean> {
  containers.value = []
  if (!selectedPod.value) return false
  try {
    const res: any = await getPodDetail({ namespace: selectedNamespace.value, name: selectedPod.value })
    const spec = res.data?.spec || {}
    const app = (spec.containers || []).map((c: any) => ({ name: c.name as string, isInit: false }))
    const init = (spec.initContainers || []).map((c: any) => ({ name: c.name as string, isInit: true }))
    containers.value = [...app, ...init]
    return true
  } catch (e: any) {
    // 取容器列表失败时显式报错,避免误报"Pod 无容器"并把选择器卡死
    ElMessage.error('获取容器列表失败: ' + (e?.message || 'unknown error'))
    return false
  }
}

async function connectTerminal() {
  if (!selectedCluster.value || !selectedNamespace.value || !selectedPod.value || !selectedContainer.value) {
    return
  }

  disconnectTerminal()
  const gen = ++connectGen

  // WebSocket 无法设置自定义 header，改用一次性短期 ticket 鉴权（?ticket=），
  // 避免长效 access token 进入 URL 被网关/浏览器历史记录。
  let ticket = ''
  try {
    const res: any = await getWsTicket()
    ticket = res.data?.ticket || ''
  } catch (e: any) {
    if (gen === connectGen) ElMessage.error('获取终端鉴权票据失败：' + (e?.message || 'unknown error'))
    return
  }
  if (!ticket) {
    if (gen === connectGen) ElMessage.error('获取终端鉴权票据失败')
    return
  }
  // 等待 ticket 期间用户又切换了容器:让更新的连接接管,本次放弃
  if (gen !== connectGen) return

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  const params = new URLSearchParams({
    clusterName: selectedCluster.value,
    namespace: selectedNamespace.value,
    podName: selectedPod.value,
    container: selectedContainer.value,
    command: '/bin/sh',
    ticket,
  })
  const wsUrl = `${protocol}//${host}/v1/k8s/container/exec?${params.toString()}`

  ws = new WebSocket(wsUrl)

  ws.onopen = () => {
    isConnected.value = true
    // 后端要求第一条消息必须是窗口大小
    if (terminal) {
      ws?.send(JSON.stringify({ resize: [terminal.cols, terminal.rows] }))
    }
    terminal?.writeln('\x1b[32m' + t('terminal.connectedToContainer') + '\x1b[0m')
    terminal?.focus()
  }

  ws.onmessage = (event) => {
    if (event.data instanceof Blob) {
      event.data.text().then((text: string) => {
        terminal?.write(text)
      }).catch(() => {
        // 忽略 Blob 读取失败
      })
    } else {
      terminal?.write(event.data)
    }
  }

  ws.onclose = () => {
    isConnected.value = false
    terminal?.writeln('\r\n\x1b[31m' + t('terminal.connectionClosed') + '\x1b[0m')
  }

  ws.onerror = () => {
    isConnected.value = false
    terminal?.writeln('\r\n\x1b[31m' + t('terminal.connectionError') + '\x1b[0m')
  }
}

function disconnectTerminal() {
  if (ws) {
    ws.close()
    ws = null
  }
  isConnected.value = false
}

function handleResize() {
  fitAddon?.fit()
}

function initTerminal() {
  if (!terminalRef.value) return

  // 如果已有terminal实例（模式切换时），先销毁
  if (terminal) {
    terminal.dispose()
    terminal = null
  }

  terminal = new Terminal({
    cursorBlink: true,
    fontSize: 14,
    fontFamily: 'Menlo, Monaco, Consolas, "Courier New", monospace',
    theme: {
      background: '#1e1e1e',
      foreground: '#d4d4d4',
      cursor: '#d4d4d4',
    },
  })

  fitAddon = new FitAddon()
  terminal.loadAddon(fitAddon)
  terminal.loadAddon(new WebLinksAddon())
  terminal.open(terminalRef.value)
  fitAddon.fit()

  // 注册 onData / onResize 一次，始终引用当前 ws
  terminal.onData((data: string) => {
    if (ws?.readyState === WebSocket.OPEN) {
      ws.send(data)
    }
  })

  terminal.onResize(({ cols, rows }: { cols: number; rows: number }) => {
    if (ws?.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ resize: [cols, rows] }))
    }
  })
}

onMounted(async () => {
  await fetchClusters()

  const isQueryMode = !!(route.query.namespace && route.query.pod)

  if (isQueryMode) {
    // embedded 模式：先设置状态，再等Vue渲染正确的DOM，最后初始化terminal并连接
    await initWithQueryParams()
    await nextTick()
    initTerminal()
    // terminal就绪后自动连接
    if (selectedContainer.value) {
      connectTerminal()
    }
  } else {
    // standalone 模式：直接初始化terminal
    await nextTick()
    initTerminal()
    terminal?.writeln('\x1b[36m' + t('terminal.welcome') + '\x1b[0m')
    terminal?.writeln(t('terminal.selectInstructions') + '\r\n')
  }

  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  disconnectTerminal()
  terminal?.dispose()
  terminal = null
  window.removeEventListener('resize', handleResize)
})

watch(selectedCluster, (val) => {
  if (skipWatchers.value) return
  // 同步 store,使共享 axios 拦截器为后续 namespace/pod 请求注入正确的 clusterName
  if (val) {
    clusterStore.setCurrentCluster({ clusterName: val, cluster_name: val, name: val })
  }
  fetchNamespaces()
})

watch(selectedNamespace, () => {
  if (!skipWatchers.value) fetchPods()
})

watch(selectedPod, (val) => {
  if (skipWatchers.value) return
  selectedContainer.value = ''
  containers.value = []
  if (val) fetchContainers()
})

watch(selectedContainer, (val) => {
  if (skipWatchers.value) return
  if (!val) return
  // embedded 模式下切换容器自动重连;standalone 模式由用户点"连接"按钮触发。
  // connectTerminal 内部用代际守卫,快速连续切换时只有最新容器会真正连上。
  if (isEmbedded.value) {
    connectTerminal()
  }
})
</script>

<template>
  <div class="terminal-view">
    <!-- Standalone mode: show full selector card -->
    <el-card v-if="!isEmbedded" shadow="hover">
      <template #header>
        <div class="card-header">
          <span>{{ t('terminal.title') }}</span>
          <el-tag :type="isConnected ? 'success' : 'info'" size="small">
            {{ isConnected ? t('terminal.connected') : t('terminal.notConnected') }}
          </el-tag>
        </div>
      </template>

      <div class="selector-bar">
        <el-select
          v-model="selectedCluster"
          :placeholder="t('terminal.selectCluster')"
          style="width: 200px"
          filterable
        >
          <el-option
            v-for="c in clusters"
            :key="c.name"
            :label="c.displayName"
            :value="c.name"
          />
        </el-select>

        <el-select
          v-model="selectedNamespace"
          :placeholder="t('terminal.selectNamespace')"
          style="width: 200px"
          filterable
          :disabled="!selectedCluster"
        >
          <el-option
            v-for="ns in namespaces"
            :key="ns"
            :label="ns"
            :value="ns"
          />
        </el-select>

        <el-select
          v-model="selectedPod"
          :placeholder="t('terminal.selectPod')"
          style="width: 240px"
          filterable
          :disabled="!selectedNamespace"
        >
          <el-option
            v-for="p in pods"
            :key="p"
            :label="p"
            :value="p"
          />
        </el-select>

        <el-select
          v-model="selectedContainer"
          :placeholder="t('terminal.containerName')"
          style="width: 180px"
          :disabled="!selectedPod"
          filterable
        >
          <el-option-group v-if="appContainers.length" :label="t('terminal.container')">
            <el-option
              v-for="c in appContainers"
              :key="c.name"
              :label="c.name"
              :value="c.name"
            />
          </el-option-group>
          <el-option-group v-if="initContainers.length" :label="t('terminal.initContainer')">
            <el-option
              v-for="c in initContainers"
              :key="c.name"
              :label="c.name"
              :value="c.name"
            />
          </el-option-group>
        </el-select>

        <el-button
          type="primary"
          :disabled="!selectedCluster || !selectedNamespace || !selectedPod || !selectedContainer"
          @click="connectTerminal"
        >
          {{ t('terminal.connect') }}
        </el-button>

        <el-button
          :disabled="!isConnected"
          @click="disconnectTerminal"
        >
          {{ t('terminal.disconnect') }}
        </el-button>
      </div>

      <div ref="terminalRef" class="terminal-container" />
    </el-card>

    <!-- Embedded mode: fullscreen terminal with minimal info bar -->
    <div v-else class="terminal-fullscreen">
      <div class="info-bar">
        <div style="display: flex; gap: 8px; align-items: center;">
          <span class="info-text">{{ selectedNamespace }} / {{ selectedPod }}</span>
          <el-select
            v-model="selectedContainer"
            size="small"
            style="width: 180px"
            :disabled="!containers.length"
            filterable
          >
            <el-option-group v-if="appContainers.length" :label="t('terminal.container')">
              <el-option
                v-for="c in appContainers"
                :key="c.name"
                :label="c.name"
                :value="c.name"
              />
            </el-option-group>
            <el-option-group v-if="initContainers.length" :label="t('terminal.initContainer')">
              <el-option
                v-for="c in initContainers"
                :key="c.name"
                :label="c.name"
                :value="c.name"
              />
            </el-option-group>
          </el-select>
        </div>
        <div style="display: flex; gap: 8px; align-items: center;">
          <el-tag :type="isConnected ? 'success' : 'danger'" size="small">
            {{ isConnected ? t('terminal.connected') : t('terminal.notConnected') }}
          </el-tag>
          <el-button size="small" type="danger" :disabled="!isConnected" @click="disconnectTerminal">
            {{ t('terminal.disconnect') }}
          </el-button>
        </div>
      </div>
      <div ref="terminalRef" class="terminal-fullscreen-body" />
    </div>
  </div>
</template>

<style scoped>
.terminal-view {
  height: 100%;
}

.terminal-view > .el-card {
  padding: 24px;
  height: 100%;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.selector-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
  align-items: center;
}

.terminal-container {
  height: calc(100vh - 300px);
  min-height: 400px;
  background: #1e1e1e;
  border-radius: 4px;
  padding: 4px;
}

/* Fullscreen embedded mode */
.terminal-fullscreen {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #1e1e1e;
}

.info-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 16px;
  background: #252526;
  color: #cccccc;
  font-size: 13px;
  flex-shrink: 0;
}

.info-text {
  font-family: Menlo, Monaco, Consolas, 'Courier New', monospace;
}

.terminal-fullscreen-body {
  flex: 1;
  padding: 4px;
  min-height: 0;
}
</style>
