<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getStatefulSetDetail,
  getStatefulSetYaml,
  updateStatefulSetYaml,
  deleteStatefulSet,
  scaleStatefulSet,
  restartStatefulSet,
  getStatefulSetEvents,
  getStatefulSetPods,
  deletePod,
} from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'
import PodListPanel from '@/components/PodListPanel.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const statefulSet = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const yamlDialogVisible = ref(false)
const yamlEditorRef = ref<InstanceType<typeof YamlEditor>>()
const events = ref<any[]>([])
const eventsLoading = ref(false)
const pods = ref<any[]>([])
const podsLoading = ref(false)
const activeTab = ref('info')

// Scale dialog
const scaleDialogVisible = ref(false)
const scaleReplicas = ref<number>(1)
const scaleLoading = ref(false)

const namespace = route.params.namespace as string
const name = route.params.name as string

const statusTagType = computed(() => {
  const ready = statefulSet.value?.status?.readyReplicas || 0
  const desired = statefulSet.value?.spec?.replicas || 0
  if (ready === desired && desired > 0) return 'success'
  if (ready > 0) return 'warning'
  return 'danger'
})

const statusText = computed(() => {
  const ready = statefulSet.value?.status?.readyReplicas || 0
  const desired = statefulSet.value?.spec?.replicas || 0
  if (ready === desired && desired > 0) return 'Ready'
  if (ready > 0) return 'Partial'
  return 'Not Ready'
})

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getStatefulSetDetail({ namespace, name })
    statefulSet.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || '加载 StatefulSet 详情失败')
  } finally {
    loading.value = false
  }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getStatefulSetYaml({ namespace, name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || '加载 YAML 失败')
  } finally {
    yamlLoading.value = false
  }
}

async function fetchEvents() {
  eventsLoading.value = true
  try {
    const res: any = await getStatefulSetEvents({ namespace, name })
    events.value = res.data || []
  } catch (e: any) {
    events.value = []
    ElMessage.error(e?.message || '加载事件失败')
  } finally {
    eventsLoading.value = false
  }
}

async function fetchPods() {
  podsLoading.value = true
  try {
    const res: any = await getStatefulSetPods({ namespace, name })
    pods.value = res.data?.items || res.data || []
  } catch (e: any) {
    pods.value = []
    ElMessage.error(e?.message || '加载 Pod 列表失败')
  } finally {
    podsLoading.value = false
  }
}

function handleOpenYaml() {
  fetchYaml()
  yamlDialogVisible.value = true
}

async function handleSaveYaml(content: string) {
  try {
    await updateStatefulSetYaml({ namespace, name, yaml: content })
    ElMessage.success('YAML 保存成功')
    yamlDialogVisible.value = false
    fetchDetail()
  } catch (e: any) {
    ElMessage.error(e?.message || '保存 YAML 失败')
    yamlEditorRef.value?.resetSaving()
  }
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(
      `确定要删除 StatefulSet "${name}" 吗？此操作不可恢复。`,
      '确认删除',
      { type: 'error', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await deleteStatefulSet({ namespace, name })
    ElMessage.success('StatefulSet 已删除')
    router.push('/workloads/statefulsets')
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || '删除失败')
    }
  }
}

async function handleRestart() {
  try {
    await ElMessageBox.confirm(
      `确定要重启 StatefulSet "${name}" 吗？这将触发滚动更新。`,
      '确认重启',
      { type: 'warning' }
    )
    await restartStatefulSet({ namespace, name })
    ElMessage.success('StatefulSet 已重启')
    fetchDetail()
    fetchPods()
  } catch {
    // cancelled
  }
}

function handleScale() {
  scaleReplicas.value = statefulSet.value?.spec?.replicas ?? 1
  scaleDialogVisible.value = true
}

async function handleScaleConfirm() {
  scaleLoading.value = true
  try {
    await scaleStatefulSet({ namespace, name, replicas: scaleReplicas.value })
    ElMessage.success(`StatefulSet 已扩缩容至 ${scaleReplicas.value} 个副本`)
    scaleDialogVisible.value = false
    fetchDetail()
    // Poll for pod list update (K8s needs time to create/delete pods)
    const expectedPods = scaleReplicas.value
    for (let i = 0; i < 10; i++) {
      await new Promise(r => setTimeout(r, 1000))
      await fetchPods()
      if (pods.value.length === expectedPods) break
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '扩缩容失败')
  } finally {
    scaleLoading.value = false
  }
}

function handleTabChange(tab: string) {
  if (tab === 'yaml' && !yamlContent.value) fetchYaml()
  if (tab === 'events' && events.value.length === 0) fetchEvents()
  if (tab === 'pods' && pods.value.length === 0) fetchPods()
}

function getClusterName(): string {
  try {
    const saved = localStorage.getItem('gkube_cluster')
    if (saved) {
      const c = JSON.parse(saved)
      return c?.clusterName || c?.cluster_name || c?.name || ''
    }
  } catch { /* ignore */ }
  return ''
}

function handlePodLogs(pod: any) {
  const cluster = getClusterName()
  window.open(`/fullscreen/logs?namespace=${pod.metadata?.namespace || namespace}&pod=${pod.metadata?.name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
}

function handlePodExec(pod: any) {
  const cluster = getClusterName()
  window.open(`/fullscreen/terminal?namespace=${pod.metadata?.namespace || namespace}&pod=${pod.metadata?.name}${cluster ? '&cluster=' + cluster : ''}`, '_blank')
}

async function handleDeletePod(pod: any) {
  try {
    await ElMessageBox.confirm(
      `确定要删除 Pod "${pod.metadata?.name}" 吗？`,
      '确认删除',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await deletePod({ namespace, name: pod.metadata.name })
    ElMessage.success('Pod 已删除')
    fetchPods()
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || '删除失败')
    }
  }
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(async () => {
  fetchDetail()
  fetchPods()
  fetchEvents()
}, { autoStart: false })

onMounted(() => {
  fetchDetail().then(() => {
    fetchPods()
  })
  fetchEvents()
})
</script>

<template>
  <div class="detail-page" v-loading="loading">
    <!-- Header -->
    <div class="page-header">
      <div class="header-left">
        <div class="title-line">
          <h2 class="res-name">{{ name }}</h2>
          <el-tag :type="statusTagType" effect="dark" size="small">{{ statusText }}</el-tag>
          <span class="ns-tag">ns/{{ namespace }}</span>
          <span class="replicas-info" v-if="statefulSet">
            {{ statefulSet.status?.readyReplicas ?? 0 }}/{{ statefulSet.spec?.replicas ?? 0 }} ready
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
        <el-button type="primary" @click="handleScale">扩缩容</el-button>
        <el-button type="warning" @click="handleRestart">重启</el-button>
        <el-button @click="handleOpenYaml">YAML</el-button>
        <el-button type="danger" plain @click="handleDelete">删除</el-button>
        <el-button @click="router.push('/workloads/statefulsets')">返回列表</el-button>
      </div>
    </div>

    <template v-if="statefulSet">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <!-- 概览 Tab -->
        <el-tab-pane label="概览" name="info">
          <el-card shadow="never">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="名称">{{ statefulSet.metadata?.name }}</el-descriptions-item>
              <el-descriptions-item label="命名空间">{{ statefulSet.metadata?.namespace }}</el-descriptions-item>
              <el-descriptions-item label="副本数">{{ statefulSet.spec?.replicas ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="就绪副本">{{ statefulSet.status?.readyReplicas ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="已更新副本">{{ statefulSet.status?.updatedReplicas ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="当前副本">{{ statefulSet.status?.currentReplicas ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="服务名称">{{ statefulSet.spec?.serviceName || '-' }}</el-descriptions-item>
              <el-descriptions-item label="更新策略">{{ statefulSet.spec?.updateStrategy?.type || 'RollingUpdate' }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ statefulSet.metadata?.creationTimestamp || '-' }}</el-descriptions-item>
            </el-descriptions>

            <!-- Labels -->
            <div v-if="statefulSet.metadata?.labels && Object.keys(statefulSet.metadata.labels).length > 0" style="margin-top: 16px;">
              <h4>标签</h4>
              <el-tag
                v-for="(val, key) in statefulSet.metadata.labels"
                :key="key"
                style="margin-right: 8px; margin-bottom: 8px;"
              >
                {{ key }}={{ val }}
              </el-tag>
            </div>

            <!-- Selector -->
            <div v-if="statefulSet.spec?.selector?.matchLabels && Object.keys(statefulSet.spec.selector.matchLabels).length > 0" style="margin-top: 16px;">
              <h4>选择器</h4>
              <el-tag
                v-for="(val, key) in statefulSet.spec.selector.matchLabels"
                :key="key"
                style="margin-right: 8px; margin-bottom: 8px;"
                type="info"
              >
                {{ key }}={{ val }}
              </el-tag>
            </div>

            <!-- Volume Claim Templates -->
            <div v-if="statefulSet.spec?.volumeClaimTemplates?.length" style="margin-top: 16px;">
              <h4>持久卷声明模板</h4>
              <el-table :data="statefulSet.spec.volumeClaimTemplates" border size="small">
                <el-table-column label="名称" prop="metadata.name" width="200" />
                <el-table-column label="访问模式">
                  <template #default="{ row }">
                    {{ row.spec?.accessModes?.join(', ') || '-' }}
                  </template>
                </el-table-column>
                <el-table-column label="存储容量">
                  <template #default="{ row }">
                    {{ row.spec?.resources?.requests?.storage || '-' }}
                  </template>
                </el-table-column>
                <el-table-column label="存储类">
                  <template #default="{ row }">
                    {{ row.spec?.storageClassName || '-' }}
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </el-card>
        </el-tab-pane>

        <!-- Pods Tab -->
        <el-tab-pane label="Pods" name="pods">
          <el-card shadow="never">
            <PodListPanel
              :pods="pods"
              :loading="podsLoading"
              @logs="handlePodLogs"
              @exec="handlePodExec"
              @delete="handleDeletePod"
            />
            <el-empty v-if="!podsLoading && pods.length === 0" description="暂无 Pod" />
          </el-card>
        </el-tab-pane>

        <!-- Events Tab -->
        <el-tab-pane label="事件" name="events">
          <el-card shadow="never">
            <el-table :data="events" v-loading="eventsLoading" stripe>
              <el-table-column prop="type" label="类型" width="100">
                <template #default="{ row }">
                  <el-tag :type="row.type === 'Warning' ? 'danger' : 'info'" size="small">{{ row.type }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="reason" label="原因" width="150" />
              <el-table-column prop="message" label="消息" min-width="300" show-overflow-tooltip />
              <el-table-column prop="last_seen" label="最后发生" width="180" />
            </el-table>
            <el-empty v-if="!eventsLoading && events.length === 0" description="暂无事件" />
          </el-card>
        </el-tab-pane>

        <!-- YAML Tab -->
        <el-tab-pane label="YAML" name="yaml">
          <el-card shadow="never">
            <div v-loading="yamlLoading">
              <YamlEditor
                ref="yamlEditorRef"
                v-model="yamlContent"
                height="600px"
                :read-only="false"
                :saveable="true"
                auto-format
                @save="handleSaveYaml"
              />
            </div>
          </el-card>
        </el-tab-pane>
      </el-tabs>
    </template>

    <!-- YAML Dialog -->
    <el-dialog v-model="yamlDialogVisible" title="YAML" width="70%" top="5vh" destroy-on-close>
      <div v-loading="yamlLoading">
        <YamlEditor ref="yamlEditorRef" v-model="yamlContent" height="600px" :read-only="true" :saveable="true" @save="handleSaveYaml" />
      </div>
    </el-dialog>

    <!-- Scale Dialog -->
    <el-dialog v-model="scaleDialogVisible" title="扩缩容" width="480px" destroy-on-close>
      <div>
        <p style="margin-bottom: 16px;">调整 <strong>{{ name }}</strong> 副本数</p>
        <el-descriptions :column="1" border size="small" style="margin-bottom: 16px;">
          <el-descriptions-item label="当前">{{ statefulSet?.spec?.replicas ?? '-' }}</el-descriptions-item>
          <el-descriptions-item label="就绪">{{ statefulSet?.status?.readyReplicas ?? '-' }}</el-descriptions-item>
        </el-descriptions>
        <el-form-item label="目标">
          <el-input-number v-model="scaleReplicas" :min="0" :max="100" style="width: 200px;" />
        </el-form-item>
        <el-alert v-if="scaleReplicas === 0" title="设为 0 将停止所有 Pod。" type="warning" :closable="false" show-icon style="margin-top: 8px;" />
      </div>
      <template #footer>
        <el-button @click="scaleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="scaleLoading" @click="handleScaleConfirm">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.detail-page {
  padding: 16px 20px;
  height: 100vh;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
}

/* Header */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
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

.replicas-info {
  font-size: 13px;
  color: var(--el-text-color-regular);
}

.header-actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}
</style>
