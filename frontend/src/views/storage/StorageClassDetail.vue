<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getStorageClassDetail,
  deleteStorageClass,
  getStorageClassEvents,
} from '@/api/resource'
import YamlDrawer from '@/components/YamlDrawer.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const storageClass = ref<any>(null)
const yamlDialogVisible = ref(false)
const events = ref<any[]>([])
const eventsLoading = ref(false)
const activeTab = ref('info')

const name = route.params.name as string

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getStorageClassDetail({ name })
    storageClass.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || '加载 StorageClass 详情失败')
  } finally {
    loading.value = false
  }
}

async function fetchEvents() {
  eventsLoading.value = true
  try {
    const res: any = await getStorageClassEvents({ name })
    events.value = res.data || []
  } catch {
    // Events API may not exist yet — silently ignore
    events.value = []
  } finally {
    eventsLoading.value = false
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
    await ElMessageBox.confirm(
      `确认删除 StorageClass "${name}"？此操作不可恢复。`,
      '确认删除',
      { type: 'warning' }
    )
    await deleteStorageClass({ name })
    ElMessage.success('StorageClass 已删除')
    router.push('/storage/storageclasses')
  } catch {
    // cancelled
  }
}

function handleTabChange(tab: string) {
  if (tab === 'yaml' && !yamlContent.value) {
    fetchYaml()
  }
  if (tab === 'events' && events.value.length === 0) {
    fetchEvents()
  }
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(async () => {
  fetchDetail()
  fetchEvents()
}, { autoStart: false })

onMounted(() => {
  fetchDetail()
  fetchEvents()
})
</script>

<template>
  <div class="detail-page" v-loading="loading">

    <!-- 顶部标题栏 -->
    <div class="page-header">
      <div class="header-left">
        <div class="title-line">
          <h2 class="res-name">{{ name }}</h2>
          <el-tag v-if="storageClass?.default" type="success" effect="dark" size="small">默认</el-tag>
          <span class="sc-provisioner" v-if="storageClass?.provisioner">{{ storageClass.provisioner }}</span>
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
        <el-button @click="router.push('/storage/storageclasses')">返回列表</el-button>
      </div>
    </div>

    <template v-if="storageClass">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">

        <!-- 基本信息 Tab -->
        <el-tab-pane label="基本信息" name="info">
          <div class="info-content">
            <el-descriptions :column="2" border style="margin-top: 8px;">
              <el-descriptions-item label="名称">{{ storageClass.name }}</el-descriptions-item>
              <el-descriptions-item label="Provisioner">{{ storageClass.provisioner || '-' }}</el-descriptions-item>
              <el-descriptions-item label="回收策略">{{ storageClass.reclaimPolicy || storageClass.reclaim_policy || '-' }}</el-descriptions-item>
              <el-descriptions-item label="卷绑定模式">{{ storageClass.volumeBindingMode || storageClass.volume_binding_mode || '-' }}</el-descriptions-item>
              <el-descriptions-item label="默认 StorageClass">
                <el-tag v-if="storageClass.default" type="success" size="small">是</el-tag>
                <span v-else>否</span>
              </el-descriptions-item>
              <el-descriptions-item label="Age">{{ storageClass.age || '-' }}</el-descriptions-item>
              <el-descriptions-item label="创建时间" :span="2">{{ storageClass.creationTimestamp || storageClass.creation_timestamp || '-' }}</el-descriptions-item>
            </el-descriptions>

            <!-- Parameters -->
            <div v-if="storageClass.parameters && Object.keys(storageClass.parameters).length > 0" class="section-block">
              <h4 class="section-title">参数</h4>
              <el-table :data="Object.entries(storageClass.parameters).map(([k, v]) => ({ key: k, value: v }))" size="small" border stripe>
                <el-table-column prop="key" label="键" width="250" />
                <el-table-column prop="value" label="值" min-width="200" show-overflow-tooltip />
              </el-table>
            </div>

            <!-- Labels -->
            <div v-if="storageClass.labels && Object.keys(storageClass.labels).length > 0" class="section-block">
              <h4 class="section-title">标签</h4>
              <div class="tags-wrap">
                <el-tag
                  v-for="(val, key) in storageClass.labels"
                  :key="key"
                  style="margin-right: 8px; margin-bottom: 8px;"
                >
                  {{ key }}={{ val }}
                </el-tag>
              </div>
            </div>

            <!-- Annotations -->
            <div v-if="storageClass.annotations && Object.keys(storageClass.annotations).length > 0" class="section-block">
              <h4 class="section-title">注解</h4>
              <div class="tags-wrap">
                <el-tag
                  v-for="(val, key) in storageClass.annotations"
                  :key="key"
                  type="info"
                  style="margin-right: 8px; margin-bottom: 8px;"
                >
                  {{ key }}={{ val }}
                </el-tag>
              </div>
            </div>
          </div>
        </el-tab-pane>

        <!-- 事件 Tab -->
        <el-tab-pane label="事件" name="events">
          <div v-loading="eventsLoading" class="events-body">
            <el-table v-if="events.length > 0" :data="events" size="small" stripe>
              <el-table-column prop="type" label="类型" width="100">
                <template #default="{ row }">
                  <el-tag :type="row.type === 'Warning' ? 'danger' : 'info'" size="small">{{ row.type }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="reason" label="原因" width="150" />
              <el-table-column prop="message" label="信息" min-width="300" show-overflow-tooltip />
              <el-table-column prop="last_seen" label="最后发生" width="180" />
            </el-table>
            <div v-else class="empty-hint">暂无事件</div>
          </div>
        </el-tab-pane>

        <!-- YAML Tab -->
        <el-tab-pane label="YAML" name="yaml">
          <div v-loading="yamlLoading">
            <YamlEditor ref="yamlEditorRef" v-model="yamlContent" height="600px" :read-only="true" :saveable="true" @save="handleSaveYaml" />
          </div>
        </el-tab-pane>
      </el-tabs>
    </template>

    <!-- YAML Dialog -->
    <YamlDrawer
      v-model="yamlDialogVisible"
      resource-type="storageclass"
      :name="name"
      @saved="handleYamlSaved"
    />
  </div>
</template>

<style scoped>
.detail-page {
  padding: 16px 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
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

.sc-provisioner {
  font-size: 13px;
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

.info-content {
  padding: 8px 0;
}

.section-block {
  margin-top: 20px;
}

.section-title {
  margin: 0 0 12px 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.tags-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.events-body {
  min-height: 200px;
}

.empty-hint {
  padding: 40px;
  text-align: center;
  color: var(--el-text-color-secondary);
  font-size: 14px;
}
</style>
