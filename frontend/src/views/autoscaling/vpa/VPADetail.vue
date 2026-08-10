<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Timer, ArrowLeft } from '@element-plus/icons-vue'
import { deleteVpa, getVpaDetail } from '@/api/resource'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import YamlDrawer from '@/components/YamlDrawer.vue'
import VPAForm from './components/VPAForm.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const vpa = ref<any>(null)
const yamlDialogVisible = ref(false)
const editDialogVisible = ref(false)
const editFullscreen = ref(false)

const namespace = route.params.namespace as string
const name = route.params.name as string

const recommendationCondition = computed(() => {
  const conditions = vpa.value?.status?.conditions || []
  return conditions.find((c: any) => c.type === 'RecommendationProvided')
})

const statusTagType = computed(() => {
  if (recommendationCondition.value?.status === 'True') return 'success'
  if (recommendationCondition.value?.status === 'False') return 'warning'
  return 'info'
})

const statusText = computed(() => {
  if (recommendationCondition.value?.status === 'True') return 'Recommended'
  return recommendationCondition.value?.reason || 'Pending'
})

const recommendations = computed(() => vpa.value?.status?.recommendation?.containerRecommendations || [])
const containerPolicies = computed(() => vpa.value?.spec?.resourcePolicy?.containerPolicies || [])
const conditions = computed(() => vpa.value?.status?.conditions || [])

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getVpaDetail({ namespace, name })
    vpa.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || '加载 VPA 详情失败')
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
    await ElMessageBox.confirm(
      `确定要删除 VPA "${name}" 吗？此操作不可恢复。`,
      '确认删除',
      { type: 'error', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await deleteVpa({ namespace, name })
    ElMessage.success('VPA 已删除')
    router.push('/autoscaling/vpa')
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || '删除失败')
    }
  }
}

function handleEdit() {
  editDialogVisible.value = true
}

function handleEditSuccess() {
  editDialogVisible.value = false
  fetchDetail()
}

function handleEditCancel() {
  editDialogVisible.value = false
}

function resourceMapText(value: any) {
  if (!value || typeof value !== 'object') return '-'
  return Object.entries(value).map(([k, v]) => `${k}: ${v}`).join(', ')
}

function listText(value: any) {
  if (!Array.isArray(value) || value.length === 0) return '-'
  return value.join(', ')
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchDetail, { autoStart: false })

onMounted(fetchDetail)
</script>

<template>
  <div class="detail-page" v-loading="loading">
    <div class="page-header">
      <div class="header-left">
        <h2 class="res-name">{{ name }}</h2>
        <div class="meta-line">
          <el-tag :type="statusTagType" effect="dark" size="small">{{ statusText }}</el-tag>
          <span class="ns-tag">ns/{{ namespace }}</span>
          <span class="target-info" v-if="vpa">
            {{ vpa.spec?.targetRef?.kind }}/{{ vpa.spec?.targetRef?.name }}
          </span>
        </div>
      </div>
      <div class="header-actions">
        <el-button type="info" @click="handleEdit">编辑</el-button>
        <el-button @click="handleOpenYaml">YAML</el-button>
        <el-button type="danger" plain @click="handleDelete">删除</el-button>
        <div class="action-divider" />
        <el-popover placement="bottom" :width="200" trigger="click">
          <template #reference>
            <el-button :type="isRunning ? 'success' : 'default'" :icon="Timer" @click="toggle()" />
          </template>
          <div class="auto-refresh-popover">
            <div class="popover-title">{{ isRunning ? `自动刷新中 ${countdown}s` : '自动刷新' }}</div>
            <el-select
              :model-value="currentInterval / 1000"
              @update:model-value="setIntervalOption"
              :teleported="false"
              size="small"
              style="width: 100%;"
            >
              <el-option v-for="sec in availableIntervals" :key="sec" :value="sec" :label="`每 ${sec} 秒刷新`" />
            </el-select>
          </div>
        </el-popover>
        <el-tooltip content="刷新" placement="top">
          <el-button @click="manualRefresh()" :loading="loading" :icon="Refresh" />
        </el-tooltip>
        <el-tooltip content="返回列表" placement="top">
          <el-button :icon="ArrowLeft" @click="router.push('/autoscaling/vpa')" />
        </el-tooltip>
      </div>
    </div>

    <template v-if="vpa">
      <div class="main-layout">
        <div class="left-panel">
          <div class="panel-title">基本信息</div>
          <div class="info-body">
            <el-descriptions :column="1" border size="small">
              <el-descriptions-item label="名称">{{ vpa.metadata?.name || name }}</el-descriptions-item>
              <el-descriptions-item label="命名空间">{{ vpa.metadata?.namespace || namespace }}</el-descriptions-item>
              <el-descriptions-item label="目标资源">
                {{ vpa.spec?.targetRef?.apiVersion || '-' }} {{ vpa.spec?.targetRef?.kind || '-' }}/{{ vpa.spec?.targetRef?.name || '-' }}
              </el-descriptions-item>
              <el-descriptions-item label="更新模式">{{ vpa.spec?.updatePolicy?.updateMode || 'Auto' }}</el-descriptions-item>
            </el-descriptions>

            <el-alert
              v-if="['Auto', 'Recreate'].includes(vpa.spec?.updatePolicy?.updateMode)"
              title="当前 VPA 可能自动驱逐或重建 Pod，请确认业务可接受重启风险。"
              type="warning"
              show-icon
              :closable="false"
              class="detail-warning"
            />

            <div v-if="vpa.metadata?.labels && Object.keys(vpa.metadata.labels).length > 0" class="labels-section">
              <h4>Labels</h4>
              <el-tag v-for="(val, key) in vpa.metadata.labels" :key="key" size="small" class="label-tag">
                {{ key }}={{ val }}
              </el-tag>
            </div>
          </div>
        </div>

        <div class="right-panel">
          <div class="right-section">
            <div class="panel-title">
              资源推荐
              <span class="count-badge">{{ recommendations.length }} 个</span>
            </div>
            <el-table v-if="recommendations.length" :data="recommendations" size="small" stripe>
              <el-table-column prop="containerName" label="容器" width="140" />
              <el-table-column label="Target" min-width="180">
                <template #default="{ row }">{{ resourceMapText(row.target) }}</template>
              </el-table-column>
              <el-table-column label="Lower Bound" min-width="180">
                <template #default="{ row }">{{ resourceMapText(row.lowerBound) }}</template>
              </el-table-column>
              <el-table-column label="Upper Bound" min-width="180">
                <template #default="{ row }">{{ resourceMapText(row.upperBound) }}</template>
              </el-table-column>
              <el-table-column label="Uncapped Target" min-width="180">
                <template #default="{ row }">{{ resourceMapText(row.uncappedTarget) }}</template>
              </el-table-column>
            </el-table>
            <div v-else class="empty-hint">暂无推荐值，VPA Recommender 可能还未生成推荐。</div>
          </div>

          <div class="right-section">
            <div class="panel-title">
              容器策略
              <span class="count-badge">{{ containerPolicies.length }} 个</span>
            </div>
            <el-table v-if="containerPolicies.length" :data="containerPolicies" size="small" stripe>
              <el-table-column prop="containerName" label="容器" width="140" />
              <el-table-column prop="mode" label="模式" width="100" />
              <el-table-column label="控制资源" min-width="150">
                <template #default="{ row }">{{ listText(row.controlledResources) }}</template>
              </el-table-column>
              <el-table-column label="Min Allowed" min-width="180">
                <template #default="{ row }">{{ resourceMapText(row.minAllowed) }}</template>
              </el-table-column>
              <el-table-column label="Max Allowed" min-width="180">
                <template #default="{ row }">{{ resourceMapText(row.maxAllowed) }}</template>
              </el-table-column>
            </el-table>
            <div v-else class="empty-hint">暂无容器策略</div>
          </div>

          <div class="right-section">
            <div class="panel-title">
              Conditions
              <span class="count-badge">{{ conditions.length }} 个</span>
            </div>
            <el-table v-if="conditions.length" :data="conditions" size="small" stripe>
              <el-table-column prop="type" label="类型" width="190" />
              <el-table-column prop="status" label="状态" width="90" />
              <el-table-column prop="reason" label="原因" width="180" show-overflow-tooltip />
              <el-table-column prop="message" label="消息" min-width="260" show-overflow-tooltip />
              <el-table-column prop="lastTransitionTime" label="更新时间" width="180" show-overflow-tooltip />
            </el-table>
            <div v-else class="empty-hint">暂无 Conditions</div>
          </div>
        </div>
      </div>
    </template>

    <YamlDrawer
      v-model="yamlDialogVisible"
      resource-type="vpa"
      :namespace="namespace"
      :name="name"
      @saved="handleYamlSaved"
    />

    <el-drawer
      v-model="editDialogVisible"
      title="编辑 VPA"
      :size="editFullscreen ? '100%' : '70%'"
      direction="rtl"
      class="edit-drawer"
      :body-style="{ padding: '20px 0' }"
    >
      <template #header="{ titleId, titleClass }">
        <div class="drawer-header">
          <h4 :id="titleId" :class="titleClass">编辑 VPA</h4>
          <el-button text @click="editFullscreen = !editFullscreen">
            {{ editFullscreen ? '退出全屏' : '全屏' }}
          </el-button>
        </div>
      </template>
      <VPAForm :is-edit="true" :initial-data="vpa" @success="handleEditSuccess" @cancel="handleEditCancel" />
    </el-drawer>
  </div>
</template>

<style scoped>
.detail-page {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 16px;
}

.res-name {
  margin: 0 0 8px;
  font-size: 24px;
  font-weight: 600;
}

.meta-line {
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--el-text-color-secondary);
}

.ns-tag,
.target-info {
  font-size: 13px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.action-divider {
  width: 1px;
  height: 24px;
  background: var(--el-border-color);
  margin: 0 4px;
}

.main-layout {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}

.left-panel,
.right-section {
  border: 1px solid var(--el-border-color-light);
  border-radius: 8px;
  background: var(--el-bg-color);
}

.left-panel {
  width: 320px;
  flex-shrink: 0;
}

.right-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
}

.panel-title {
  padding: 12px 16px;
  font-size: 14px;
  font-weight: 600;
  border-bottom: 1px solid var(--el-border-color-light);
  display: flex;
  align-items: center;
  gap: 8px;
}

.info-body,
.right-section {
  overflow: hidden;
}

.info-body {
  padding: 16px;
}

.right-section :deep(.el-table) {
  border-radius: 0 0 8px 8px;
}

.count-badge {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  font-weight: 400;
}

.detail-warning {
  margin-top: 12px;
}

.labels-section {
  margin-top: 16px;
}

.labels-section h4 {
  margin: 0 0 8px;
  font-size: 13px;
}

.label-tag {
  margin-right: 8px;
  margin-bottom: 8px;
}

.empty-hint {
  padding: 32px 16px;
  text-align: center;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

@media (max-width: 960px) {
  .main-layout {
    flex-direction: column;
  }

  .left-panel {
    width: 100%;
  }
}
</style>
