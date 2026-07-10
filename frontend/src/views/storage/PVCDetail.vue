<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getPvcDetail, getPvcYaml } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const pvc = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const yamlDialogVisible = ref(false)

const namespace = route.params.namespace as string
const name = route.params.name as string

const statusTagType = computed(() => {
  const s = (pvc.value?.status?.phase || '').toLowerCase()
  if (s === 'bound') return 'success'
  if (s === 'pending') return 'warning'
  if (s === 'lost') return 'danger'
  return 'info'
})

const pvcName = computed(() => pvc.value?.metadata?.name || name)
const pvcNamespace = computed(() => pvc.value?.metadata?.namespace || namespace)
const pvcStatus = computed(() => pvc.value?.status?.phase || '-')
const pvcVolumeName = computed(() => pvc.value?.spec?.volumeName || '-')
const pvcCapacity = computed(() => pvc.value?.status?.capacity?.storage || '-')
const pvcAccessModes = computed(() => (pvc.value?.spec?.accessModes || []).join(', ') || '-')
const pvcStorageClassName = computed(() => pvc.value?.spec?.storageClassName || '-')
const pvcLabels = computed(() => pvc.value?.metadata?.labels || {})
const pvcAge = computed(() => {
  const ts = pvc.value?.metadata?.creationTimestamp
  if (!ts) return '-'
  const created = new Date(ts).getTime()
  const now = Date.now()
  const diff = Math.floor((now - created) / 1000)
  if (diff < 60) return `${diff}s`
  if (diff < 3600) return `${Math.floor(diff / 60)}m`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h`
  const days = Math.floor(diff / 86400)
  if (days < 365) return `${days}d`
  return `${Math.floor(days / 365)}y`
})

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getPvcDetail({ namespace, name })
    pvc.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || '加载PVC详情失败')
  } finally {
    loading.value = false
  }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getPvcYaml({ namespace, name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || '加载YAML失败')
  } finally {
    yamlLoading.value = false
  }
}

function handleOpenYaml() {
  fetchYaml()
  yamlDialogVisible.value = true
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchDetail, { autoStart: false })

onMounted(fetchDetail)
</script>

<template>
  <div class="detail-page" v-loading="loading">

    <!-- ===== 顶部标题栏 ===== -->
    <div class="page-header">
      <div class="header-left">
        <div class="title-line">
          <h2 class="res-name">{{ pvcName }}</h2>
          <el-tag v-if="pvcStatus !== '-'" :type="statusTagType" effect="dark" size="small">{{ pvcStatus }}</el-tag>
          <span class="ns-tag">ns/{{ pvcNamespace }}</span>
        </div>
      </div>
      <div class="header-actions">
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
        <el-button @click="handleOpenYaml">YAML</el-button>
        <el-button @click="router.push('/storage/pvcs')">返回列表</el-button>
      </div>
    </div>

    <template v-if="pvc">
      <!-- ===== 基本信息 ===== -->
      <el-card shadow="never" class="detail-card">
        <template #header>
          <span class="card-title">基本信息</span>
        </template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="名称">{{ pvcName }}</el-descriptions-item>
          <el-descriptions-item label="命名空间">{{ pvcNamespace }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTagType" size="small">{{ pvcStatus }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="卷名">{{ pvcVolumeName }}</el-descriptions-item>
          <el-descriptions-item label="容量">{{ pvcCapacity }}</el-descriptions-item>
          <el-descriptions-item label="访问模式">{{ pvcAccessModes }}</el-descriptions-item>
          <el-descriptions-item label="存储类名">{{ pvcStorageClassName }}</el-descriptions-item>
          <el-descriptions-item label="Age">{{ pvcAge }}</el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- ===== 标签 ===== -->
      <el-card v-if="Object.keys(pvcLabels).length > 0" shadow="never" class="detail-card">
        <template #header>
          <span class="card-title">标签</span>
        </template>
        <div class="labels-container">
          <el-tag
            v-for="(val, key) in pvcLabels"
            :key="key"
            class="label-tag"
          >
            {{ key }}={{ val }}
          </el-tag>
        </div>
      </el-card>
    </template>

    <!-- YAML Drawer -->
    <el-drawer v-model="yamlDialogVisible" title="PVC YAML" size="85%" direction="rtl" class="yaml-drawer"
      :body-style="{ padding: '0', height: '100%' }">
      <div v-loading="yamlLoading" style="height: calc(100vh - 52px);">
        <YamlEditor v-model="yamlContent" height="100%" auto-format read-only />
      </div>
    </el-drawer>
  </div>
</template>

<style scoped>
.detail-page {
  padding: 16px 20px;
  height: 100vh;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  overflow-y: auto;
}

/* Header */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  flex-shrink: 0;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.title-line {
  display: flex;
  align-items: center;
  gap: 10px;
}

.res-name {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.ns-tag {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color-lighter);
  padding: 2px 8px;
  border-radius: 4px;
}

.header-actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

/* Cards */
.detail-card {
  margin-bottom: 16px;
}

.card-title {
  font-size: 14px;
  font-weight: 600;
}

/* Labels */
.labels-container {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.label-tag {
  font-size: 12px;
}
</style>

<style>
.yaml-drawer .el-drawer__header {
  padding: 6px 16px;
  margin-bottom: 0;
  min-height: auto;
}
</style>
