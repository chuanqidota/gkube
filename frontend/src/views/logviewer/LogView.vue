<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { useClusterStore } from '@/stores/cluster'
import { getPodDetail, getPodList, getNamespaceList, extractNamespaceNames } from '@/api/resource'
import { getClusterList } from '@/api/cluster'
import { ElMessage } from 'element-plus'
import { getToken } from '@/utils/auth'

const { t } = useI18n()
const route = useRoute()
const clusterStore = useClusterStore()

interface ClusterOption {
  name: string
  displayName: string
}

type ConnectionStatus = 'disconnected' | 'connecting' | 'connected' | 'error'

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
const logContent = ref('')
const autoScroll = ref(true)
const status = ref<ConnectionStatus>('disconnected')
const logContainerRef = ref<HTMLDivElement>()

// Whether opened from pod context (with query params) — hide selectors
const isEmbedded = ref(false)

let abortController: AbortController | null = null
// 流代际:每次 startLogStream 自增,旧流被 abort 后的异步回调据此跳过,避免覆盖新流状态
let streamGen = 0

// Cap retained log lines to avoid unbounded memory growth on long-running streams
const MAX_LOG_LINES = 5000
function appendLog(text: string) {
  logContent.value += text
  const lines = logContent.value.split('\n')
  if (lines.length > MAX_LOG_LINES) {
    logContent.value = lines.slice(lines.length - MAX_LOG_LINES).join('\n')
  }
}

const statusTextMap: Record<ConnectionStatus, () => string> = {
  disconnected: () => t('log.disconnected'),
  connecting: () => t('log.connecting'),
  connected: () => t('log.connected'),
  error: () => t('log.connectionFailed'),
}

const statusType: Record<ConnectionStatus, string> = {
  disconnected: 'info',
  connecting: 'warning',
  connected: 'success',
  error: 'danger',
}

async function fetchClusters() {
  try {
    const res: any = await getClusterList({ page: 1, size: 100 })
    const items = res.data?.items || []
    clusters.value = items.map((c: any) => ({
      name: c.clusterName || c.cluster_name || c.name,
      displayName: c.displayName || c.display_name || c.clusterName || c.name,
    }))
  } catch (e) {
    console.error('[LogView] Failed to load clusters:', e)
  }
}

async function fetchNamespaces() {
  if (!selectedCluster.value) return
  namespaces.value = []
  selectedNamespace.value = ''
  pods.value = []
  selectedPod.value = ''
  selectedContainer.value = ''
  try {
    const res: any = await getNamespaceList()
    namespaces.value = extractNamespaceNames(res.data)
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.loadFailed'))
  }
}

async function fetchPods() {
  if (!selectedCluster.value || !selectedNamespace.value) return
  pods.value = []
  selectedPod.value = ''
  selectedContainer.value = ''
  try {
    const res: any = await getPodList({ namespace: selectedNamespace.value })
    const items = Array.isArray(res.data) ? res.data : (res.data?.items || [])
    pods.value = items.map((p: any) => p.metadata?.name || p.name).filter(Boolean)
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.loadFailed'))
  }
}

function scrollToBottom() {
  if (autoScroll.value && logContainerRef.value) {
    nextTick(() => {
      if (logContainerRef.value) {
        logContainerRef.value.scrollTop = logContainerRef.value.scrollHeight
      }
    })
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

async function startLogStream() {
  if (!selectedCluster.value || !selectedNamespace.value || !selectedPod.value || !selectedContainer.value) {
    return
  }

  stopLogStream()
  status.value = 'connecting'
  const gen = ++streamGen

  const token = getToken()
  const params = new URLSearchParams({
    clusterName: selectedCluster.value,
    namespace: selectedNamespace.value,
    podName: selectedPod.value,
    container: selectedContainer.value,
    tailLines: '100',
  })
  const url = `/v1/k8s/log/stream?${params.toString()}`

  // 用局部 controller 引用,避免被下一次 startLogStream 覆盖后误 abort
  const myController = new AbortController()
  abortController = myController

  try {
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        Accept: 'text/event-stream',
      },
      signal: myController.signal,
    })

    // 等待 fetch 期间用户切换了容器:旧流放弃,不覆盖新流的 connecting 状态
    if (gen !== streamGen) return

    if (!response.ok) {
      status.value = 'error'
      appendLog(`[Error] HTTP ${response.status}: ${response.statusText}\n`)
      return
    }

    status.value = 'connected'
    appendLog(t('log.connectedToStream') + '\n')

    const reader = response.body?.getReader()
    if (!reader) {
      status.value = 'error'
      return
    }

    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (gen !== streamGen) return
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (line.startsWith('data:')) {
          const data = line.slice(5).trim()
          if (data) {
            appendLog(data + '\n')
          }
        } else if (line.trim() && !line.startsWith(':')) {
          appendLog(line + '\n')
        }
      }

      scrollToBottom()
    }

    if (gen === streamGen) status.value = 'disconnected'
  } catch (err: any) {
    if (err.name === 'AbortError') {
      // 仅当仍是当前流时才置为 disconnected;被新流取代的旧流静默退出
      if (gen === streamGen) status.value = 'disconnected'
    } else {
      if (gen === streamGen) {
        status.value = 'error'
        appendLog(`[Error] ${err.message}\n`)
      }
    }
  }
}

function stopLogStream() {
  if (abortController) {
    abortController.abort()
    abortController = null
  }
  // 同步置 disconnected 覆盖"用户主动停止"场景;若是 startLogStream 内部重连调用,
  // 紧接着的 status='connecting' 会立即覆盖此值。旧流被 abort 后的异步 catch
  // 由代际守卫跳过,不会回写状态,因此不会覆盖新流的 connecting。
  status.value = 'disconnected'
}

function clearLogs() {
  logContent.value = ''
}

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

  // Auto-start log stream
  await nextTick()
  if (selectedContainer.value) {
    startLogStream()
  }
}

onMounted(() => {
  fetchClusters().then(() => initWithQueryParams())
})

onBeforeUnmount(() => {
  stopLogStream()
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
  // embedded 模式下切换容器自动重启日志流;standalone 模式由用户点"开始监听"触发。
  // startLogStream 内部会停掉旧流并用代际守卫,快速连续切换时只有最新容器会真正接管状态。
  if (isEmbedded.value) {
    startLogStream()
  }
})
</script>

<template>
  <div class="log-view">
    <!-- Standalone mode: show full selector card -->
    <el-card v-if="!isEmbedded" shadow="hover">
      <template #header>
        <div class="card-header">
          <span>{{ t('log.title') }}</span>
          <el-tag :type="statusType[status] as any" size="small">
            {{ statusTextMap[status]() }}
          </el-tag>
        </div>
      </template>

      <div class="selector-bar">
        <el-select
          v-model="selectedCluster"
          :placeholder="t('log.selectCluster')"
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
          :placeholder="t('log.selectNamespace')"
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
          :placeholder="t('log.selectPod')"
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
          :placeholder="t('log.containerName')"
          style="width: 180px"
          :disabled="!selectedPod"
          filterable
        >
          <el-option-group v-if="appContainers.length" :label="t('log.container')">
            <el-option
              v-for="c in appContainers"
              :key="c.name"
              :label="c.name"
              :value="c.name"
            />
          </el-option-group>
          <el-option-group v-if="initContainers.length" :label="t('log.initContainer')">
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
          :disabled="!selectedCluster || !selectedNamespace || !selectedPod || !selectedContainer || status === 'connecting'"
          @click="startLogStream"
        >
          {{ t('log.startListening') }}
        </el-button>

        <el-button
          :disabled="status !== 'connected'"
          @click="stopLogStream"
        >
          {{ t('log.stop') }}
        </el-button>

        <el-button @click="clearLogs">
          {{ t('log.clear') }}
        </el-button>

        <el-checkbox v-model="autoScroll">
          {{ t('log.autoScroll') }}
        </el-checkbox>
      </div>

      <div
        ref="logContainerRef"
        class="log-container"
      >
        <pre class="log-content">{{ logContent || t('log.waitingForLogs') }}</pre>
      </div>
    </el-card>

    <!-- Embedded mode: fullscreen log with minimal info bar -->
    <div v-else class="log-fullscreen">
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
            <el-option-group v-if="appContainers.length" :label="t('log.container')">
              <el-option
                v-for="c in appContainers"
                :key="c.name"
                :label="c.name"
                :value="c.name"
              />
            </el-option-group>
            <el-option-group v-if="initContainers.length" :label="t('log.initContainer')">
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
          <el-tag :type="statusType[status] as any" size="small">
            {{ statusTextMap[status]() }}
          </el-tag>
          <el-button size="small" type="danger" :disabled="status !== 'connected'" @click="stopLogStream">
            {{ t('log.stop') }}
          </el-button>
          <el-button size="small" @click="clearLogs">
            {{ t('log.clear') }}
          </el-button>
          <el-checkbox v-model="autoScroll" size="small">
            {{ t('log.autoScroll') }}
          </el-checkbox>
        </div>
      </div>
      <div
        ref="logContainerRef"
        class="log-fullscreen-body"
      >
        <pre class="log-content">{{ logContent || t('log.waitingForLogs') }}</pre>
      </div>
    </div>
  </div>
</template>

<style scoped>
.log-view {
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

.log-container {
  height: calc(100vh - 300px);
  min-height: 400px;
  background: #1e1e1e;
  border-radius: 4px;
  overflow-y: auto;
  padding: 12px;
}

.log-content {
  font-family: Menlo, Monaco, Consolas, 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  color: #d4d4d4;
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
}

/* Fullscreen embedded mode */
.log-fullscreen {
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

.log-fullscreen-body {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
  min-height: 0;
}
</style>
