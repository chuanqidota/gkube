<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getPvDetail, deletePv } from '@/api/resource'
import YamlDrawer from '@/components/YamlDrawer.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const pv = ref<any>(null)
const yamlDialogVisible = ref(false)

const name = route.params.name as string

const statusTagType = computed(() => {
  const status = (pv.value?.status?.phase || '').toLowerCase()
  if (status === 'bound') return 'success'
  if (status === 'available') return 'primary'
  if (status === 'released') return 'warning'
  if (status === 'failed') return 'danger'
  return 'info'
})

const statusText = computed(() => {
  return pv.value?.status?.phase || 'Unknown'
})

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getPvDetail({ name })
    pv.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load PV detail')
  } finally {
    loading.value = false
  }
}

function handleOpenYaml() {
  yamlDialogVisible.value = true
}

function handleYamlSaved() {
  fetchDetail()
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(`删除持久卷 "${name}"?`, '确认', { type: 'warning' })
    await deletePv({ name })
    ElMessage.success('删除成功')
    router.push('/storage/pvs')
  } catch {
    /* cancelled */
  }
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchDetail, { autoStart: false })

onMounted(fetchDetail)
</script>

<template>
  <div class="detail-page" v-loading="loading">
    <!-- 顶部标题栏 -->
    <div class="page-header">
      <div class="header-left">
        <div class="title-line">
          <h2 class="res-name">{{ name }}</h2>
          <el-tag :type="statusTagType" effect="dark" size="small">{{ statusText }}</el-tag>
          <span class="pv-info" v-if="pv">
            {{ pv.spec?.capacity?.storage || '-' }} / {{ (pv.spec?.accessModes || []).join(', ') }}
          </span>
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
        <el-button type="danger" @click="handleDelete">删除</el-button>
        <el-button @click="router.push('/storage/pvs')">返回列表</el-button>
      </div>
    </div>

    <template v-if="pv">
      <!-- 基本信息 -->
      <div class="detail-section">
        <div class="section-title">基本信息</div>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="名称">{{ pv.metadata?.name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="容量">{{ pv.spec?.capacity?.storage || '-' }}</el-descriptions-item>
          <el-descriptions-item label="访问模式">{{ (pv.spec?.accessModes || []).join(', ') || '-' }}</el-descriptions-item>
          <el-descriptions-item label="存储类">{{ pv.spec?.storageClassName || '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTagType" size="small">{{ statusText }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="回收策略">{{ pv.spec?.persistentVolumeReclaimPolicy || '-' }}</el-descriptions-item>
          <el-descriptions-item label="卷模式">{{ pv.spec?.volumeMode || '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ pv.metadata?.creationTimestamp || '-' }}</el-descriptions-item>
        </el-descriptions>
      </div>

      <!-- 存储源信息 -->
      <div class="detail-section">
        <div class="section-title">存储源</div>
        <el-descriptions :column="2" border>
          <template v-if="pv.spec?.nfs">
            <el-descriptions-item label="类型">NFS</el-descriptions-item>
            <el-descriptions-item label="服务器">{{ pv.spec.nfs.server || '-' }}</el-descriptions-item>
            <el-descriptions-item label="路径">{{ pv.spec.nfs.path || '-' }}</el-descriptions-item>
          </template>
          <template v-else-if="pv.spec?.hostPath">
            <el-descriptions-item label="类型">HostPath</el-descriptions-item>
            <el-descriptions-item label="路径">{{ pv.spec.hostPath.path || '-' }}</el-descriptions-item>
          </template>
          <template v-else-if="pv.spec?.local">
            <el-descriptions-item label="类型">Local</el-descriptions-item>
            <el-descriptions-item label="路径">{{ pv.spec.local.path || '-' }}</el-descriptions-item>
          </template>
          <template v-else>
            <el-descriptions-item label="类型">-</el-descriptions-item>
          </template>
        </el-descriptions>
      </div>

      <!-- 声明引用 -->
      <div class="detail-section" v-if="pv.spec?.claimRef">
        <div class="section-title">声明引用</div>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="命名空间">{{ pv.spec.claimRef.namespace || '-' }}</el-descriptions-item>
          <el-descriptions-item label="名称">{{ pv.spec.claimRef.name || '-' }}</el-descriptions-item>
        </el-descriptions>
      </div>

      <!-- 标签 -->
      <div class="detail-section" v-if="pv.metadata?.labels && Object.keys(pv.metadata.labels).length > 0">
        <div class="section-title">标签</div>
        <div class="labels-container">
          <el-tag
            v-for="(val, key) in pv.metadata.labels"
            :key="key"
            class="label-tag"
          >
            {{ key }}={{ val }}
          </el-tag>
        </div>
      </div>

      <!-- 注解 -->
      <div class="detail-section" v-if="pv.metadata?.annotations && Object.keys(pv.metadata.annotations).length > 0">
        <div class="section-title">注解</div>
        <div class="labels-container">
          <el-tag
            v-for="(val, key) in pv.metadata.annotations"
            :key="key"
            type="info"
            class="label-tag"
          >
            {{ key }}={{ val }}
          </el-tag>
        </div>
      </div>
    </template>

    <!-- YAML Dialog -->
    <YamlDrawer
      v-model="yamlDialogVisible"
      resource-type="pv"
      :name="name"
      @saved="handleYamlSaved"
    />
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

.pv-info {
  font-size: 13px;
  color: var(--el-text-color-regular);
}

.header-actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

/* Detail sections */
.detail-section {
  margin-bottom: 20px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  overflow: hidden;
  background: var(--el-bg-color);
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  padding: 12px 16px;
  background: var(--el-fill-color-lighter);
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.labels-container {
  padding: 12px 16px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.label-tag {
  margin: 0;
}
</style>
