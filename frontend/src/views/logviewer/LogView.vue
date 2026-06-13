<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { getToken } from '@/utils/auth'

const { t } = useI18n()

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

const logContent = ref('')
const autoScroll = ref(true)
const status = ref<ConnectionStatus>('disconnected')
const logContainerRef = ref<HTMLDivElement>()

let abortController: AbortController | null = null

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
  } catch {
    // silently fail
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
    pods.value = (Array.isArray(podData) ? podData : []).map((p: any) => typeof p === 'string' ? p : p.name)
  } catch {
    // silently fail
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
    containerName: selectedContainer.value,
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
      logContent.value += `[Error] HTTP ${response.status}: ${response.statusText}\n`
      return
    }

    status.value = 'connected'
    logContent.value += t('log.connectedToStream') + '\n'

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
            logContent.value += data + '\n'
          }
        } else if (line.trim() && !line.startsWith(':')) {
          logContent.value += line + '\n'
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
      logContent.value += `[Error] ${err.message}\n`
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

onMounted(() => {
  fetchClusters()
})

onBeforeUnmount(() => {
  stopLogStream()
})

watch(selectedCluster, () => {
  fetchNamespaces()
})

watch(selectedNamespace, () => {
  fetchPods()
})
</script>

<template>
  <div class="log-view">
    <el-card shadow="hover">
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
  </div>
</template>

<style scoped>
.log-view {
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
</style>
