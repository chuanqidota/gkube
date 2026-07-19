<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { useClusterStore } from '@/stores/cluster'
import { getPodDetail } from '@/api/resource'
import { ElMessage } from 'element-plus'
import { getToken } from '@/utils/auth'

const { t } = useI18n()
const route = useRoute()

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

const skipWatchers = ref(false)
const logContent = ref('')
const autoScroll = ref(true)
const status = ref<ConnectionStatus>('disconnected')
const logContainerRef = ref<HTMLDivElement>()

// Whether opened from pod context (with query params) — hide selectors
const isEmbedded = ref(false)

let abortController: AbortController | null = null

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
    const token = getToken()
    const res = await fetch('/api/v1/clusters', {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    })
    const json = await res.json()
    const items = json?.data?.items || json?.data || json?.items || []
    clusters.value = items.map((c: any) => ({
      name: c.cluster_name || c.name,
      displayName: c.display_name || c.cluster_name || c.name,
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
    const token = getToken()
    const params = new URLSearchParams({ clusterName: selectedCluster.value })
    const res = await fetch(`/api/v1/k8s/namespace/list?${params}`, {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    })
    const json = await res.json()
    const nsData = json?.data?.namespaces || json?.data || json?.namespaces || []
    namespaces.value = nsData.map((n: any) => typeof n === 'string' ? n : n.name)
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
    const token = getToken()
    const params = new URLSearchParams({
      clusterName: selectedCluster.value,
      namespace: selectedNamespace.value,
    })
    const res = await fetch(`/api/v1/k8s/pod/list?${params}`, {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    })
    const json = await res.json()
    const podData = json?.data || json || []
    pods.value = (Array.isArray(podData) ? podData : []).map((p: any) => typeof p === 'string' ? p : (p.metadata?.name || p.name))
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

async function startLogStream() {
  if (!selectedCluster.value || !selectedNamespace.value || !selectedPod.value || !selectedContainer.value) {
    return
  }

  stopLogStream()
  status.value = 'connecting'

  const token = getToken()
  const params = new URLSearchParams({
    clusterName: selectedCluster.value,
    namespace: selectedNamespace.value,
    podName: selectedPod.value,
    container: selectedContainer.value,
    tailLines: '100',
  })
  const url = `/v1/k8s/log/stream?${params.toString()}`

  abortController = new AbortController()

  try {
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        Accept: 'text/event-stream',
      },
      signal: abortController.signal,
    })

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

    status.value = 'disconnected'
  } catch (err: any) {
    if (err.name === 'AbortError') {
      status.value = 'disconnected'
    } else {
      status.value = 'error'
      appendLog(`[Error] ${err.message}\n`)
    }
  }
}

function stopLogStream() {
  if (abortController) {
    abortController.abort()
    abortController = null
  }
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

  // Set cluster from query param or localStorage
  const clusterStore = useClusterStore()
  if (cluster) {
    selectedCluster.value = cluster as string
    // Ensure localStorage has the cluster set so the API interceptor works
    // The interceptor reads clusterName (camelCase), so we must set that key
    let clusterObj: any = null
    try {
      const saved = localStorage.getItem('gkube_cluster')
      if (saved) clusterObj = JSON.parse(saved)
    } catch { /* ignore */ }
    if (!clusterObj) clusterObj = {}
    clusterObj.clusterName = cluster as string
    clusterObj.cluster_name = cluster as string
    clusterObj.name = cluster as string
    localStorage.setItem('gkube_cluster', JSON.stringify(clusterObj))
    // Also update Pinia store so it stays in sync
    clusterStore.setCurrentCluster(clusterObj)
  } else if (clusterStore.currentCluster) {
    selectedCluster.value = clusterStore.currentCluster.cluster_name || clusterStore.currentCluster.name
  }

  selectedNamespace.value = namespace as string
  selectedPod.value = pod as string

  // Fetch pod detail to get container list
  if (!container) {
    try {
      const res: any = await getPodDetail({ namespace: namespace as string, name: pod as string })
      const containers = res.data?.spec?.containers || []
      if (containers.length > 0) {
        selectedContainer.value = containers[0].name
      } else {
        ElMessage.warning('Pod has no containers')
      }
    } catch (e: any) {
      ElMessage.error('Failed to get pod detail: ' + (e?.message || 'unknown error'))
    }
  } else {
    selectedContainer.value = container as string
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

watch(selectedCluster, () => {
  if (!skipWatchers.value) fetchNamespaces()
})

watch(selectedNamespace, () => {
  if (!skipWatchers.value) fetchPods()
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

        <el-input
          v-model="selectedContainer"
          :placeholder="t('log.containerName')"
          style="width: 180px"
          :disabled="!selectedPod"
        />

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
        <span class="info-text">{{ selectedNamespace }} / {{ selectedPod }} / {{ selectedContainer }}</span>
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
