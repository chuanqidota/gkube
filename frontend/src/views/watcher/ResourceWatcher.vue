<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, VideoPlay, VideoPause, Bell } from '@element-plus/icons-vue'
import request from '@/api/request'

const loading = ref(false)
const watching = ref(false)
const selectedResource = ref('pods')
const selectedNamespace = ref('')
const namespaces = ref<string[]>([])
const resources = ref<any[]>([])
const changes = ref<any[]>([])
let watchTimer: ReturnType<typeof setInterval> | null = null

const resourceTypes = [
  { value: 'pods', label: 'Pods' },
  { value: 'deployments', label: 'Deployments' },
  { value: 'services', label: 'Services' },
  { value: 'configmaps', label: 'ConfigMaps' },
  { value: 'secrets', label: 'Secrets' },
  { value: 'nodes', label: 'Nodes' },
]

async function fetchNamespaces() {
  try {
    const res: any = await request.get('/k8s/namespace/list')
    namespaces.value = res.data?.map((ns: any) => ns.name) || []
  } catch {
    namespaces.value = []
  }
}

async function fetchResources() {
  loading.value = true
  try {
    const params: any = {}
    if (selectedNamespace.value) {
      params.namespace = selectedNamespace.value
    }
    const res: any = await request.get(`/k8s/${selectedResource.value}/list`, { params })
    const newResources = res.data || []

    // Detect changes
    const oldNames = new Set(resources.value.map(r => r.name))
    const newNames = new Set(newResources.map((r: any) => r.name))

    // Added
    newResources.forEach((r: any) => {
      if (!oldNames.has(r.name)) {
        changes.value.unshift({
          type: 'added',
          resource: selectedResource.value,
          name: r.name,
          namespace: r.namespace,
          time: new Date(),
        })
      }
    })

    // Removed
    resources.value.forEach(r => {
      if (!newNames.has(r.name)) {
        changes.value.unshift({
          type: 'removed',
          resource: selectedResource.value,
          name: r.name,
          namespace: r.namespace,
          time: new Date(),
        })
      }
    })

    resources.value = newResources
  } catch (e: any) {
    ElMessage.warning('Failed to load resources')
  } finally {
    loading.value = false
  }
}

function startWatching() {
  watching.value = true
  fetchResources()
  watchTimer = setInterval(fetchResources, 5000)
}

function stopWatching() {
  watching.value = false
  if (watchTimer) {
    clearInterval(watchTimer)
    watchTimer = null
  }
}

function toggleWatching() {
  if (watching.value) {
    stopWatching()
  } else {
    startWatching()
  }
}

function changeType(type: string) {
  return type === 'added' ? 'success' : type === 'removed' ? 'danger' : 'info'
}

function changeIcon(type: string) {
  return type === 'added' ? '+' : type === 'removed' ? '-' : '~'
}

function formatTime(time: Date) {
  return time.toLocaleTimeString()
}

onMounted(() => {
  fetchNamespaces()
  fetchResources()
})

onUnmounted(() => {
  stopWatching()
})
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <h3 style="margin: 0;"><el-icon><VideoPlay /></el-icon> 资源监视器</h3>
        <div class="filter-right">
          <el-select v-model="selectedResource" style="width: 150px;" @change="fetchResources">
            <el-option v-for="r in resourceTypes" :key="r.value" :label="r.label" :value="r.value" />
          </el-select>
          <el-select v-model="selectedNamespace" placeholder="所有命名空间" clearable style="width: 150px;" @change="fetchResources">
            <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
          </el-select>
          <el-button :type="watching ? 'danger' : 'success'" @click="toggleWatching">
            <el-icon><component :is="watching ? VideoPause : VideoPlay" /></el-icon>
            {{ watching ? '停止监视' : '开始监视' }}
          </el-button>
          <el-button type="primary" @click="fetchResources"><el-icon><Refresh /></el-icon> 刷新</el-button>
        </div>
      </div>
    </el-card>

    <el-row :gutter="16">
      <el-col :span="16">
        <el-card shadow="never">
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center;">
              <h4 style="margin: 0;">资源列表</h4>
              <el-tag v-if="watching" type="success" size="small">监视中...</el-tag>
            </div>
          </template>
          <el-table :data="resources" v-loading="loading" stripe size="small">
            <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip />
            <el-table-column prop="namespace" label="命名空间" width="120" />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'Running' || row.status === 'Active' ? 'success' : 'warning'" size="small">
                  {{ row.status || '-' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="age" label="存在时间" width="120" />
          </el-table>
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card shadow="never">
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center;">
              <h4 style="margin: 0;"><el-icon><Bell /></el-icon> 变更记录</h4>
              <el-tag size="small">{{ changes.length }}</el-tag>
            </div>
          </template>
          <div class="changes-list">
            <div v-for="(change, index) in changes.slice(0, 50)" :key="index" class="change-item">
              <div class="change-header">
                <el-tag :type="changeType(change.type)" size="small">{{ changeIcon(change.type) }}</el-tag>
                <span class="change-name">{{ change.name }}</span>
                <span class="change-time">{{ formatTime(change.time) }}</span>
              </div>
              <div class="change-detail">
                {{ change.resource }} - {{ change.namespace || 'cluster-scoped' }}
              </div>
            </div>
            <el-empty v-if="changes.length === 0" description="暂无变更" :image-size="60" />
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; justify-content: space-between; align-items: center; }
.filter-right { display: flex; align-items: center; gap: 8px; }
.changes-list { max-height: 500px; overflow-y: auto; }
.change-item { padding: 8px; border-bottom: 1px solid #ebeef5; }
.change-header { display: flex; align-items: center; gap: 8px; margin-bottom: 4px; }
.change-name { font-weight: 500; flex: 1; }
.change-time { font-size: 12px; color: #909399; }
.change-detail { font-size: 12px; color: #606266; }
</style>
