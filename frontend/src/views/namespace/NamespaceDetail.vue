<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, PriceTag, Notebook, Refresh, Timer, ArrowLeft } from '@element-plus/icons-vue'
import {
  getNamespaceDetail,
  getNamespaceYaml,
  updateNamespace,
  deleteNamespace,
  updateNamespaceLabels,
  getResourceQuotaList,
  getLimitRangeList,
  createResourceQuota,
  createLimitRange,
} from '@/api/resource'
import { useNamespaceStore } from '@/stores/namespace'
import YamlDrawer from '@/components/YamlDrawer.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const route = useRoute()
const router = useRouter()
const namespaceStore = useNamespaceStore()
const loading = ref(false)
const namespace = ref<any>(null)
const yamlDialogVisible = ref(false)
const resourceQuotas = ref<any[]>([])
const limitRanges = ref<any[]>([])

// Labels dialog
const labelsDialogVisible = ref(false)
const labelsArray = ref<Array<{ key: string; value: string }>>([])

// Annotations dialog
const annotationsDialogVisible = ref(false)
const annotationsArray = ref<Array<{ key: string; value: string }>>([])

// Create ResourceQuota dialog
const rqDialogVisible = ref(false)
const rqCreating = ref(false)
const rqForm = reactive({
  name: '',
  requestsCpu: '',
  requestsMemory: '',
  limitsCpu: '',
  limitsMemory: '',
  pods: '',
  services: '',
  pvcs: '',
})

// Create LimitRange dialog
const lrDialogVisible = ref(false)
const lrCreating = ref(false)
const lrForm = reactive({
  name: '',
  limits: [{
    type: 'Container' as string,
    maxCpu: '',
    maxMemory: '',
    minCpu: '',
    minMemory: '',
    defaultCpu: '',
    defaultMemory: '',
    defaultRequestCpu: '',
    defaultRequestMemory: '',
  }],
})

const name = route.params.name as string

const statusTagType = computed(() => {
  return namespace.value?.status === 'Active' ? 'success' : 'warning'
})

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getNamespaceDetail({ name })
    namespace.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || '加载命名空间详情失败')
  } finally {
    loading.value = false
  }
}

async function fetchResourceQuotas() {
  try {
    const res: any = await getResourceQuotaList({ namespace: name })
    resourceQuotas.value = res.data || []
  } catch { /* ignore */ }
}

async function fetchLimitRanges() {
  try {
    const res: any = await getLimitRangeList({ namespace: name })
    limitRanges.value = res.data || []
  } catch { /* ignore */ }
}

function handleOpenYaml() {
  yamlDialogVisible.value = true
}

function handleYamlSaved() {
  fetchDetail()
}

// Create ResourceQuota
function showCreateRqDialog() {
  rqForm.name = ''
  rqForm.requestsCpu = ''
  rqForm.requestsMemory = ''
  rqForm.limitsCpu = ''
  rqForm.limitsMemory = ''
  rqForm.pods = ''
  rqForm.services = ''
  rqForm.pvcs = ''
  rqDialogVisible.value = true
}

async function handleCreateRq() {
  if (!rqForm.name) {
    ElMessage.warning('请输入名称')
    return
  }
  const hard: Record<string, string> = {}
  if (rqForm.requestsCpu) hard['requests.cpu'] = rqForm.requestsCpu
  if (rqForm.requestsMemory) hard['requests.memory'] = rqForm.requestsMemory
  if (rqForm.limitsCpu) hard['limits.cpu'] = rqForm.limitsCpu
  if (rqForm.limitsMemory) hard['limits.memory'] = rqForm.limitsMemory
  if (rqForm.pods) hard['pods'] = rqForm.pods
  if (rqForm.services) hard['services'] = rqForm.services
  if (rqForm.pvcs) hard['persistentvolumeclaims'] = rqForm.pvcs

  if (Object.keys(hard).length === 0) {
    ElMessage.warning('请至少填写一项资源限制')
    return
  }

  rqCreating.value = true
  try {
    const yaml = JSON.stringify({
      apiVersion: 'v1',
      kind: 'ResourceQuota',
      metadata: { name: rqForm.name, namespace: name },
      spec: { hard },
    }, null, 2)
    await createResourceQuota({ namespace: name, yaml })
    ElMessage.success('ResourceQuota 创建成功')
    rqDialogVisible.value = false
    fetchResourceQuotas()
  } catch (e: any) {
    ElMessage.error(e?.message || '创建失败')
  } finally {
    rqCreating.value = false
  }
}

// Create LimitRange
function showCreateLrDialog() {
  lrForm.name = ''
  lrForm.limits = [{
    type: 'Container',
    maxCpu: '', maxMemory: '',
    minCpu: '', minMemory: '',
    defaultCpu: '', defaultMemory: '',
    defaultRequestCpu: '', defaultRequestMemory: '',
  }]
  lrDialogVisible.value = true
}

function addLrLimit() {
  lrForm.limits.push({
    type: 'Container',
    maxCpu: '', maxMemory: '',
    minCpu: '', minMemory: '',
    defaultCpu: '', defaultMemory: '',
    defaultRequestCpu: '', defaultRequestMemory: '',
  })
}

function removeLrLimit(i: number) {
  lrForm.limits.splice(i, 1)
}

async function handleCreateLr() {
  if (!lrForm.name) {
    ElMessage.warning('请输入名称')
    return
  }

  const limits = lrForm.limits.map(l => {
    const limit: any = { type: l.type }
    const max: any = {}; const min: any = {}; const def: any = {}; const defReq: any = {}
    if (l.maxCpu) max.cpu = l.maxCpu
    if (l.maxMemory) max.memory = l.maxMemory
    if (l.minCpu) min.cpu = l.minCpu
    if (l.minMemory) min.memory = l.minMemory
    if (l.defaultCpu) def.cpu = l.defaultCpu
    if (l.defaultMemory) def.memory = l.defaultMemory
    if (l.defaultRequestCpu) defReq.cpu = l.defaultRequestCpu
    if (l.defaultRequestMemory) defReq.memory = l.defaultRequestMemory
    if (Object.keys(max).length) limit.max = max
    if (Object.keys(min).length) limit.min = min
    if (Object.keys(def).length) limit.default = def
    if (Object.keys(defReq).length) limit.defaultRequest = defReq
    return limit
  })

  lrCreating.value = true
  try {
    const yaml = JSON.stringify({
      apiVersion: 'v1',
      kind: 'LimitRange',
      metadata: { name: lrForm.name, namespace: name },
      spec: { limits },
    }, null, 2)
    await createLimitRange({ namespace: name, yaml })
    ElMessage.success('LimitRange 创建成功')
    lrDialogVisible.value = false
    fetchLimitRanges()
  } catch (e: any) {
    ElMessage.error(e?.message || '创建失败')
  } finally {
    lrCreating.value = false
  }
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(
      `确定要删除命名空间 "${name}" 吗？该命名空间下的所有资源将被删除。`,
      '确认删除',
      { type: 'error', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await deleteNamespace({ name })
    ElMessage.success('命名空间已删除')
    namespaceStore.clearCache()
    router.push('/namespaces')
  } catch { /* cancelled */ }
}

// Labels
function handleEditLabels() {
  labelsArray.value = Object.entries(namespace.value?.labels || {}).map(([key, value]) => ({ key, value: value as string }))
  if (labelsArray.value.length === 0) labelsArray.value = [{ key: '', value: '' }]
  labelsDialogVisible.value = true
}

function addLabel() { labelsArray.value.push({ key: '', value: '' }) }
function removeLabel(i: number) { labelsArray.value.splice(i, 1) }

async function handleSaveLabels() {
  try {
    const labels: Record<string, string> = {}
    labelsArray.value.forEach((l) => {
      if (l.key.trim()) labels[l.key.trim()] = l.value
    })
    await updateNamespaceLabels({ namespace: name, labels })
    ElMessage.success('标签已更新')
    labelsDialogVisible.value = false
    fetchDetail()
  } catch (e: any) {
    ElMessage.error(e?.message || '更新标签失败')
  }
}

// Annotations
function handleEditAnnotations() {
  annotationsArray.value = Object.entries(namespace.value?.annotations || {}).map(([key, value]) => ({ key, value: value as string }))
  if (annotationsArray.value.length === 0) annotationsArray.value = [{ key: '', value: '' }]
  annotationsDialogVisible.value = true
}

function addAnnotation() { annotationsArray.value.push({ key: '', value: '' }) }
function removeAnnotation(i: number) { annotationsArray.value.splice(i, 1) }

async function handleSaveAnnotations() {
  try {
    const annotations: Record<string, string> = {}
    annotationsArray.value.forEach((a) => {
      if (a.key.trim()) annotations[a.key.trim()] = a.value
    })
    // Save annotations via YAML update since we don't have a dedicated API
    await getNamespaceYaml({ name })
    ElMessage.info('注解请通过 YAML 编辑')
    annotationsDialogVisible.value = false
  } catch (e: any) {
    ElMessage.error(e?.message || '更新注解失败')
  }
}

// ---- Resize: left-right ----
const leftWidth = ref(300)
const resizingH = ref(false)
let startX = 0, startW = 0
function onHResizeStart(e: MouseEvent) {
  e.preventDefault()
  resizingH.value = true
  startX = e.clientX
  startW = leftWidth.value
  const onMove = (ev: MouseEvent) => {
    leftWidth.value = Math.min(Math.max(startW + ev.clientX - startX, 220), 500)
  }
  const onUp = () => {
    resizingH.value = false
    document.removeEventListener('mousemove', onMove)
    document.removeEventListener('mouseup', onUp)
  }
  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onUp)
}

// ---- Resize: top-bottom (ResourceQuota / LimitRange) ----
const rightTopHeight = ref<number | null>(null)
const resizingV = ref(false)
let startY = 0, startH = 0
function onVResizeStart(e: MouseEvent) {
  e.preventDefault()
  resizingV.value = true
  startY = e.clientY
  const rightPanel = (e.target as HTMLElement).closest('.right-panel')
  if (!rightPanel) return
  startH = rightPanel.getBoundingClientRect().height
  const onMove = (ev: MouseEvent) => {
    const delta = ev.clientY - startY
    rightTopHeight.value = Math.min(Math.max(startH * 0.3 + delta, 120), startH - 120)
  }
  const onUp = () => {
    resizingV.value = false
    document.removeEventListener('mousemove', onMove)
    document.removeEventListener('mouseup', onUp)
  }
  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onUp)
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(async () => {
  fetchDetail()
  fetchResourceQuotas()
  fetchLimitRanges()
}, { autoStart: false })

onMounted(() => {
  fetchDetail()
  fetchResourceQuotas()
  fetchLimitRanges()
})
</script>

<template>
  <div class="detail-page" v-loading="loading">

    <!-- 顶部标题栏 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="res-name">{{ name }}</h2>
        <div class="meta-line">
          <el-tag :type="statusTagType" effect="dark" size="small">{{ namespace?.status || 'Unknown' }}</el-tag>
        </div>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="handleEditLabels">标签</el-button>
        <el-button @click="handleOpenYaml">YAML</el-button>
        <el-button type="danger" plain @click="handleDelete">删除</el-button>
        <div class="action-divider" />
        <el-popover placement="bottom" :width="200" trigger="hover">
          <template #reference>
            <el-button
              :type="isRunning ? 'success' : 'default'"
              :icon="Timer"
              @click="toggle()"
            />
          </template>
          <div class="auto-refresh-popover">
            <div class="popover-title">
              {{ isRunning ? `自动刷新中 ${countdown}s` : '自动刷新' }}
            </div>
            <el-select
              :model-value="currentInterval / 1000"
              @update:model-value="setIntervalOption"
              size="small"
              style="width: 100%;"
            >
              <el-option
                v-for="sec in availableIntervals"
                :key="sec"
                :value="sec"
                :label="`每 ${sec} 秒刷新`"
              />
            </el-select>
          </div>
        </el-popover>
        <el-tooltip content="刷新" placement="top">
          <el-button @click="manualRefresh()" :loading="loading" :icon="Refresh" />
        </el-tooltip>
        <el-tooltip content="返回列表" placement="top">
          <el-button :icon="ArrowLeft" @click="router.push('/namespaces')" />
        </el-tooltip>
      </div>
    </div>

    <template v-if="namespace">
      <div class="main-layout" :class="{ 'is-resizing': resizingH || resizingV }">

        <!-- 左侧：基本信息 -->
        <div class="left-panel" :style="{ width: leftWidth + 'px', minWidth: leftWidth + 'px' }">
          <div class="panel-title">基本信息</div>
          <div class="info-body">
            <el-descriptions :column="1" border size="small">
              <el-descriptions-item label="名称">{{ namespace.name }}</el-descriptions-item>
              <el-descriptions-item label="状态">
                <el-tag :type="statusTagType" size="small" effect="dark">{{ namespace.status }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ namespace.age }}</el-descriptions-item>
            </el-descriptions>

            <!-- Labels -->
            <div style="margin-top: 16px;">
              <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px;">
                <h4 style="margin: 0; font-size: 13px;">Labels</h4>
                <el-button size="small" @click="handleEditLabels">编辑</el-button>
              </div>
              <div v-if="namespace.labels && Object.keys(namespace.labels).length > 0">
                <el-tag
                  v-for="(v, k) in namespace.labels"
                  :key="k"
                  style="margin-right: 8px; margin-bottom: 8px;"
                  size="small"
                >
                  {{ k }}={{ v }}
                </el-tag>
              </div>
              <span v-else style="color: #909399; font-size: 12px;">无标签</span>
            </div>

            <!-- Annotations -->
            <div style="margin-top: 16px;">
              <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px;">
                <h4 style="margin: 0; font-size: 13px;">Annotations</h4>
                <el-button size="small" @click="handleEditAnnotations">编辑</el-button>
              </div>
              <div v-if="namespace.annotations && Object.keys(namespace.annotations).length > 0">
                <div v-for="(v, k) in namespace.annotations" :key="k" style="margin-bottom: 4px; font-size: 12px;">
                  <span style="font-weight: 600;">{{ k }}:</span> {{ v }}
                </div>
              </div>
              <span v-else style="color: #909399; font-size: 12px;">无注解</span>
            </div>
          </div>
        </div>

        <!-- 右侧：ResourceQuota + LimitRange -->
        <div class="right-panel">

          <!-- Resource Quotas -->
          <div class="right-section" :style="rightTopHeight ? { flex: 'none', height: rightTopHeight + 'px' } : {}">
            <div class="panel-title">
              资源配额
              <span class="count-badge">{{ resourceQuotas.length }} 个</span>
              <el-button size="small" type="primary" @click="showCreateRqDialog" style="margin-left: auto;">
                <el-icon><Plus /></el-icon> 创建
              </el-button>
            </div>
            <div class="table-body">
              <el-table v-if="resourceQuotas.length > 0" :data="resourceQuotas" size="small" stripe>
                <el-table-column prop="name" label="名称" min-width="200" />
                <el-table-column label="硬限制" min-width="250">
                  <template #default="{ row }">
                    <div v-for="(v, k) in (row.hard || {})" :key="k" style="font-size: 12px;">{{ k }}: {{ v }}</div>
                  </template>
                </el-table-column>
                <el-table-column label="已使用" min-width="250">
                  <template #default="{ row }">
                    <div v-for="(v, k) in (row.used || {})" :key="k" style="font-size: 12px;">{{ k }}: {{ v }}</div>
                  </template>
                </el-table-column>
                <el-table-column prop="age" label="创建时间" width="180" />
              </el-table>
              <div v-else class="empty-hint">
                暂无资源配额
                <el-button type="primary" size="small" @click="showCreateRqDialog" style="margin-top: 8px;">创建资源配额</el-button>
              </div>
            </div>
          </div>

          <!-- 垂直拖拽条 -->
          <div class="resize-handle-v" :class="{ active: resizingV }" @mousedown="onVResizeStart" />

          <!-- Limit Ranges -->
          <div class="right-section">
            <div class="panel-title">
              资源限制
              <span class="count-badge">{{ limitRanges.length }} 个</span>
              <el-button size="small" type="primary" @click="showCreateLrDialog" style="margin-left: auto;">
                <el-icon><Plus /></el-icon> 创建
              </el-button>
            </div>
            <div class="table-body">
              <el-table v-if="limitRanges.length > 0" :data="limitRanges" size="small" stripe>
                <el-table-column prop="name" label="名称" min-width="200" />
                <el-table-column label="限制" min-width="300">
                  <template #default="{ row }">
                    <div v-for="(limit, i) in (row.limits || [])" :key="i" style="font-size: 12px; margin-bottom: 4px;">
                      <el-tag size="small" style="margin-right: 4px;">{{ limit.type }}</el-tag>
                      <span v-for="(v, k) in (limit.max || {})" :key="k">Max {{ k }}: {{ v }} </span>
                      <span v-for="(v, k) in (limit.min || {})" :key="k">Min {{ k }}: {{ v }} </span>
                    </div>
                  </template>
                </el-table-column>
                <el-table-column prop="age" label="创建时间" width="180" />
              </el-table>
              <div v-else class="empty-hint">
                暂无资源限制
                <el-button type="primary" size="small" @click="showCreateLrDialog" style="margin-top: 8px;">创建资源限制</el-button>
              </div>
            </div>
          </div>

        </div>

        <!-- 水平拖拽条 -->
        <div
          class="resize-handle-h"
          :class="{ active: resizingH }"
          :style="{ left: (leftWidth - 3) + 'px' }"
          @mousedown="onHResizeStart"
        />
      </div>
    </template>

    <!-- YAML Drawer -->
    <YamlDrawer
      v-model="yamlDialogVisible"
      resource-type="namespace"
      :name="name"
      @saved="handleYamlSaved"
    />

    <!-- Labels Dialog -->
    <el-dialog v-model="labelsDialogVisible" title="管理标签" width="600px" destroy-on-close>
      <div v-for="(label, i) in labelsArray" :key="i" style="display: flex; gap: 8px; margin-bottom: 12px; align-items: center;">
        <el-input v-model="label.key" placeholder="Key" style="flex: 2;" />
        <el-input v-model="label.value" placeholder="Value" style="flex: 2;" />
        <el-button type="danger" circle size="small" @click="removeLabel(i)">
          <el-icon><Delete /></el-icon>
        </el-button>
      </div>
      <el-button @click="addLabel" style="margin-top: 8px;">
        <el-icon><Plus /></el-icon> 添加标签
      </el-button>
      <template #footer>
        <el-button @click="labelsDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveLabels">保存</el-button>
      </template>
    </el-dialog>

    <!-- Annotations Dialog -->
    <el-dialog v-model="annotationsDialogVisible" title="管理注解" width="650px" destroy-on-close>
      <div v-for="(anno, i) in annotationsArray" :key="i" style="display: flex; gap: 8px; margin-bottom: 12px; align-items: center;">
        <el-input v-model="anno.key" placeholder="Key" style="flex: 2;" />
        <el-input v-model="anno.value" placeholder="Value" style="flex: 2;" />
        <el-button type="danger" circle size="small" @click="removeAnnotation(i)">
          <el-icon><Delete /></el-icon>
        </el-button>
      </div>
      <el-button @click="addAnnotation" style="margin-top: 8px;">
        <el-icon><Plus /></el-icon> 添加注解
      </el-button>
      <template #footer>
        <el-button @click="annotationsDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveAnnotations">保存</el-button>
      </template>
    </el-dialog>

    <!-- Create ResourceQuota Drawer -->
    <el-drawer v-model="rqDialogVisible" title="创建资源配额" direction="rtl" size="500px" destroy-on-close>
      <el-form label-width="160px">
        <el-form-item label="名称" required>
          <el-input v-model="rqForm.name" placeholder="my-resource-quota" />
        </el-form-item>
        <el-divider>资源限制</el-divider>
        <el-form-item label="CPU 请求上限">
          <el-input v-model="rqForm.requestsCpu" placeholder="例如: 4" />
        </el-form-item>
        <el-form-item label="内存请求上限">
          <el-input v-model="rqForm.requestsMemory" placeholder="例如: 8Gi" />
        </el-form-item>
        <el-form-item label="CPU 限制上限">
          <el-input v-model="rqForm.limitsCpu" placeholder="例如: 8" />
        </el-form-item>
        <el-form-item label="内存限制上限">
          <el-input v-model="rqForm.limitsMemory" placeholder="例如: 16Gi" />
        </el-form-item>
        <el-form-item label="Pod 数量上限">
          <el-input v-model="rqForm.pods" placeholder="例如: 20" />
        </el-form-item>
        <el-form-item label="Service 数量上限">
          <el-input v-model="rqForm.services" placeholder="例如: 10" />
        </el-form-item>
        <el-form-item label="PVC 数量上限">
          <el-input v-model="rqForm.pvcs" placeholder="例如: 5" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rqDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="rqCreating" @click="handleCreateRq">创建</el-button>
      </template>
    </el-drawer>

    <!-- Create LimitRange Drawer -->
    <el-drawer v-model="lrDialogVisible" title="创建资源限制" direction="rtl" size="550px" destroy-on-close>
      <el-form label-width="140px">
        <el-form-item label="名称" required>
          <el-input v-model="lrForm.name" placeholder="my-limit-range" />
        </el-form-item>

        <div v-for="(limit, i) in lrForm.limits" :key="i" style="border: 1px solid var(--el-border-color); border-radius: 8px; padding: 16px; margin-bottom: 16px;">
          <div style="display: flex; justify-content: space-between; margin-bottom: 12px;">
            <el-select v-model="limit.type" style="width: 200px;">
              <el-option label="容器 (Container)" value="Container" />
              <el-option label="Pod" value="Pod" />
              <el-option label="持久卷声明 (PVC)" value="PersistentVolumeClaim" />
            </el-select>
            <el-button v-if="lrForm.limits.length > 1" type="danger" size="small" @click="removeLrLimit(i)">移除</el-button>
          </div>
          <el-form-item label="CPU 最大值"><el-input v-model="limit.maxCpu" placeholder="例如: 4" /></el-form-item>
          <el-form-item label="内存最大值"><el-input v-model="limit.maxMemory" placeholder="例如: 8Gi" /></el-form-item>
          <el-form-item label="CPU 最小值"><el-input v-model="limit.minCpu" placeholder="例如: 100m" /></el-form-item>
          <el-form-item label="内存最小值"><el-input v-model="limit.minMemory" placeholder="例如: 128Mi" /></el-form-item>
          <el-form-item label="默认 CPU 值"><el-input v-model="limit.defaultCpu" placeholder="例如: 500m" /></el-form-item>
          <el-form-item label="默认内存值"><el-input v-model="limit.defaultMemory" placeholder="例如: 512Mi" /></el-form-item>
          <el-form-item label="默认 CPU 请求值"><el-input v-model="limit.defaultRequestCpu" placeholder="例如: 100m" /></el-form-item>
          <el-form-item label="默认内存请求值"><el-input v-model="limit.defaultRequestMemory" placeholder="例如: 128Mi" /></el-form-item>
        </div>
        <el-button @click="addLrLimit" style="margin-bottom: 16px;">
          <el-icon><Plus /></el-icon> 添加限制
        </el-button>
      </el-form>
      <template #footer>
        <el-button @click="lrDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="lrCreating" @click="handleCreateLr">创建</el-button>
      </template>
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
  gap: 4px;
}

.res-name {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  line-height: 1.3;
}

.meta-line {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-actions {
  display: flex;
  flex-shrink: 0;
  align-items: center;
}

.header-actions .el-button {
  border-radius: 0;
  margin-left: -1px;
}

.header-actions .el-button:first-child {
  border-radius: 4px 0 0 4px;
  margin-left: 0;
}

.header-actions .el-button:last-of-type {
  border-radius: 0 4px 4px 0;
}

.action-divider {
  width: 1px;
  height: 20px;
  background: var(--el-border-color-lighter);
  margin: 0 4px;
}

.auto-refresh-popover {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.popover-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

/* Main Layout */
.main-layout {
  display: flex;
  gap: 2px;
  flex: 1;
  min-height: 0;
  overflow: hidden;
  position: relative;
}

/* Left Panel */
.left-panel {
  width: 320px;
  min-width: 320px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--el-bg-color);
}

.panel-title {
  font-size: 13px;
  font-weight: 600;
  padding: 10px 14px;
  background: var(--el-fill-color-lighter);
  border-bottom: 1px solid var(--el-border-color-lighter);
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.count-badge {
  font-weight: 400;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.info-body {
  flex: 1;
  overflow-y: auto;
  padding: 14px;
}

/* Right Panel */
.right-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
  overflow: hidden;
}

.right-section {
  flex: 1;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--el-bg-color);
}

.table-body {
  flex: 1;
  overflow-y: auto;
  padding: 0;
}

.empty-hint {
  padding: 24px;
  text-align: center;
  color: var(--el-text-color-secondary);
  font-size: 13px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

/* Resize handles */
.resize-handle-h {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 8px;
  cursor: col-resize;
  z-index: 10;
}

.resize-handle-h:hover,
.resize-handle-h.active {
  background: var(--el-color-primary-light-7);
}

.resize-handle-v {
  height: 4px;
  cursor: row-resize;
  flex-shrink: 0;
  position: relative;
  z-index: 5;
  margin: -2px 0;
}

.resize-handle-v:hover,
.resize-handle-v.active {
  background: var(--el-color-primary-light-7);
}

.is-resizing {
  user-select: none;
}

.is-resizing * {
  pointer-events: none;
}

/* Responsive */
@media (max-width: 768px) {
  .main-layout {
    flex-direction: column;
    overflow: auto;
  }
  .left-panel {
    width: 100% !important;
    min-width: 100% !important;
    max-height: 300px;
  }
  .resize-handle-h {
    display: none;
  }
}
</style>
