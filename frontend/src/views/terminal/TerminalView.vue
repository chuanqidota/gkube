<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import '@xterm/xterm/css/xterm.css'
import { getToken } from '@/utils/auth'

const { t } = useI18n()

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

const terminalRef = ref<HTMLDivElement>()
const isConnected = ref(false)

let terminal: Terminal | null = null
let fitAddon: FitAddon | null = null
let ws: WebSocket | null = null

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

function connectTerminal() {
  if (!selectedCluster.value || !selectedNamespace.value || !selectedPod.value || !selectedContainer.value) {
    return
  }

  disconnectTerminal()

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  const token = getToken()
  const params = new URLSearchParams({
    clusterName: selectedCluster.value,
    namespace: selectedNamespace.value,
    podName: selectedPod.value,
    containerName: selectedContainer.value,
    command: '/bin/sh',
    ...(token ? { token } : {}),
  })
  const wsUrl = `${protocol}//${host}/v1/k8s/container/exec?${params.toString()}`

  ws = new WebSocket(wsUrl)

  ws.onopen = () => {
    isConnected.value = true
    terminal?.writeln('\x1b[32m' + t('terminal.connectedToContainer') + '\x1b[0m')
    terminal?.focus()
  }

  ws.onmessage = (event) => {
    if (event.data instanceof Blob) {
      event.data.text().then((text: string) => {
        terminal?.write(text)
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

  if (terminal) {
    terminal.onData((data: string) => {
      if (ws?.readyState === WebSocket.OPEN) {
        ws.send(data)
      }
    })

    terminal.onResize(({ cols, rows }: { cols: number; rows: number }) => {
      if (ws?.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'resize', cols, rows }))
      }
    })
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

onMounted(async () => {
  await fetchClusters()

  await nextTick()

  if (terminalRef.value) {
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

    terminal.writeln('\x1b[36m' + t('terminal.welcome') + '\x1b[0m')
    terminal.writeln(t('terminal.selectInstructions') + '\r\n')

    window.addEventListener('resize', handleResize)
  }
})

onBeforeUnmount(() => {
  disconnectTerminal()
  terminal?.dispose()
  terminal = null
  window.removeEventListener('resize', handleResize)
})

watch(selectedCluster, () => {
  fetchNamespaces()
})

watch(selectedNamespace, () => {
  fetchPods()
})
</script>

<template>
  <div class="terminal-view">
    <el-card shadow="hover">
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

        <el-input
          v-model="selectedContainer"
          :placeholder="t('terminal.containerName')"
          style="width: 180px"
          :disabled="!selectedPod"
        />

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
  </div>
</template>

<style scoped>
.terminal-view {
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
</style>
