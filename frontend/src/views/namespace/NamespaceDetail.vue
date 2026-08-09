<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Refresh, Timer, ArrowLeft } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
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
import yaml from 'js-yaml'

const { t } = useI18n()
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
    ElMessage.error(e?.message || t('namespace.loadDetailFailed'))
    namespace.value = null
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
    ElMessage.warning(t('namespace.pleaseEnterName'))
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
    ElMessage.warning(t('namespace.atLeastOneLimit'))
    return
  }

  rqCreating.value = true
  try {
    const yamlContent = JSON.stringify({
      apiVersion: 'v1',
      kind: 'ResourceQuota',
      metadata: { name: rqForm.name, namespace: name },
      spec: { hard },
    }, null, 2)
    await createResourceQuota({ namespace: name, yaml: yamlContent })
    ElMessage.success(t('namespace.rqCreated'))
    rqDialogVisible.value = false
    fetchResourceQuotas()
  } catch (e: any) {
    ElMessage.error(e?.message || t('namespace.createFailed'))
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
    ElMessage.warning(t('namespace.pleaseEnterName'))
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
    const yamlContent = JSON.stringify({
      apiVersion: 'v1',
      kind: 'LimitRange',
      metadata: { name: lrForm.name, namespace: name },
      spec: { limits },
    }, null, 2)
    await createLimitRange({ namespace: name, yaml: yamlContent })
    ElMessage.success(t('namespace.lrCreated'))
    lrDialogVisible.value = false
    fetchLimitRanges()
  } catch (e: any) {
    ElMessage.error(e?.message || t('namespace.createFailed'))
  } finally {
    lrCreating.value = false
  }
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(
      t('namespace.deleteConfirm', { name }),
      t('namespace.confirmDelete'),
      { type: 'error', confirmButtonText: t('namespace.deleteBtn'), cancelButtonText: t('namespace.cancelButton') }
    )
    await deleteNamespace({ name })
    ElMessage.success(t('namespace.deleteSuccess'))
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
    ElMessage.success(t('namespace.labelsUpdated'))
    labelsDialogVisible.value = false
    fetchDetail()
  } catch (e: any) {
    ElMessage.error(e?.message || t('namespace.labelsUpdateFailed'))
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

const annotationsSaving = ref(false)

async function handleSaveAnnotations() {
  const annotations: Record<string, string> = {}
  annotationsArray.value.forEach((a) => {
    if (a.key.trim()) annotations[a.key.trim()] = a.value
  })
  annotationsSaving.value = true
  try {
    // Backend has no dedicated annotation update API, so load YAML → overwrite annotations → submit
    const res: any = await getNamespaceYaml({ name })
    const raw = res.data?.yaml ?? res.data ?? ''
    let doc: any
    try {
      doc = yaml.load(raw)
    } catch {
      ElMessage.error(t('namespace.yamlLoadError'))
      return
    }
    if (!doc?.metadata) doc.metadata = {}
    if (Object.keys(annotations).length > 0) {
      doc.metadata.annotations = annotations
    } else {
      delete doc.metadata.annotations
    }
    await updateNamespace({ yaml: yaml.dump(doc) })
    ElMessage.success(t('namespace.annotationsUpdated'))
    annotationsDialogVisible.value = false
    fetchDetail()
  } catch (e: any) {
    ElMessage.error(e?.message || t('namespace.annotationsUpdateFailed'))
  } finally {
    annotationsSaving.value = false
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
  const rightPanel = (e.target as HTMLElement).closest('.right-panel')
  if (!rightPanel) return
  resizingV.value = true
  startY = e.clientY
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
        <el-button type="primary" @click="handleEditLabels">{{ t('namespace.labelBtn') }}</el-button>
        <el-button @click="handleOpenYaml">{{ t('common.yaml') }}</el-button>
        <el-button type="danger" plain @click="handleDelete">{{ t('namespace.deleteBtn') }}</el-button>
        <div class="action-divider" />
        <el-popover placement="bottom" :width="200" trigger="click">
          <template #reference>
            <el-button
              :type="isRunning ? 'success' : 'default'"
              :icon="Timer"
              @click="toggle()"
            />
          </template>
          <div class="auto-refresh-popover">
            <div class="popover-title">
              {{ isRunning ? t('common.autoRefreshOn', { n: countdown }) : t('common.autoRefresh') }}
            </div>
            <el-select
              :model-value="currentInterval / 1000"
              @update:model-value="setIntervalOption"
              :teleported="false"
              size="small"
              style="width: 100%;"
            >
              <el-option
                v-for="sec in availableIntervals"
                :key="sec"
                :value="sec"
                :label="`${t('namespace.autoRefreshEvery')} ${sec} ${t('namespace.seconds')}`"
              />
            </el-select>
          </div>
        </el-popover>
        <el-tooltip :content="t('common.refresh')" placement="top">
          <el-button @click="manualRefresh()" :loading="loading" :icon="Refresh" />
        </el-tooltip>
        <el-tooltip :content="t('common.back')" placement="top">
          <el-button :icon="ArrowLeft" @click="router.push('/namespaces')" />
        </el-tooltip>
      </div>
    </div>

    <template v-if="namespace">
      <div class="main-layout" :class="{ 'is-resizing': resizingH || resizingV }">

        <!-- 左侧：基本信息 -->
        <div class="left-panel" :style="{ width: leftWidth + 'px', minWidth: leftWidth + 'px' }">
          <div class="panel-title">{{ t('namespace.detail') }}</div>
          <div class="info-body">
            <el-descriptions :column="1" border size="small">
              <el-descriptions-item :label="t('namespace.nameLabel')">{{ namespace.name }}</el-descriptions-item>
              <el-descriptions-item :label="t('common.status')">
                <el-tag :type="statusTagType" size="small" effect="dark">{{ namespace.status }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item :label="t('namespace.ageLabel')">{{ namespace.age }}</el-descriptions-item>
            </el-descriptions>

            <!-- Labels -->
            <div style="margin-top: 16px;">
              <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px;">
                <h4 style="margin: 0; font-size: 13px;">{{ t('namespace.labels') }}</h4>
                <el-button size="small" @click="handleEditLabels">{{ t('common.edit') }}</el-button>
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
              <span v-else style="color: #909399; font-size: 12px;">{{ t('namespace.noLabels') }}</span>
            </div>

            <!-- Annotations -->
            <div style="margin-top: 16px;">
              <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px;">
                <h4 style="margin: 0; font-size: 13px;">{{ t('namespace.annotations') }}</h4>
                <el-button size="small" @click="handleEditAnnotations">{{ t('common.edit') }}</el-button>
              </div>
              <div v-if="namespace.annotations && Object.keys(namespace.annotations).length > 0">
                <div v-for="(v, k) in namespace.annotations" :key="k" style="margin-bottom: 4px; font-size: 12px;">
                  <span style="font-weight: 600;">{{ k }}:</span> {{ v }}
                </div>
              </div>
              <span v-else style="color: #909399; font-size: 12px;">{{ t('namespace.noAnnotations') }}</span>
            </div>
          </div>
        </div>

        <!-- 右侧：ResourceQuota + LimitRange -->
        <div class="right-panel">

          <!-- Resource Quotas -->
          <div class="right-section" :style="rightTopHeight ? { flex: 'none', height: rightTopHeight + 'px' } : {}">
            <div class="panel-title">
              {{ t('namespace.resourceQuotas') }}
              <span class="count-badge">{{ resourceQuotas.length }} {{ t('namespace.count') }}</span>
              <el-button size="small" type="primary" @click="showCreateRqDialog" style="margin-left: auto;">
                <el-icon><Plus /></el-icon> {{ t('namespace.createBtn') }}
              </el-button>
            </div>
            <div class="table-body">
              <el-table v-if="resourceQuotas.length > 0" :data="resourceQuotas" size="small" stripe>
                <el-table-column prop="name" :label="t('namespace.nameLabel')" min-width="200" />
                <el-table-column :label="t('namespace.hardLimit')" min-width="250">
                  <template #default="{ row }">
                    <div v-for="(v, k) in (row.hard || {})" :key="k" style="font-size: 12px;">{{ k }}: {{ v }}</div>
                  </template>
                </el-table-column>
                <el-table-column :label="t('namespace.used')" min-width="250">
                  <template #default="{ row }">
                    <div v-for="(v, k) in (row.used || {})" :key="k" style="font-size: 12px;">{{ k }}: {{ v }}</div>
                  </template>
                </el-table-column>
                <el-table-column prop="age" :label="t('namespace.ageLabel')" width="180" />
              </el-table>
              <div v-else class="empty-hint">
                {{ t('namespace.noResourceQuotas') }}
                <el-button type="primary" size="small" @click="showCreateRqDialog" style="margin-top: 8px;">{{ t('namespace.setQuotaBtn') }}</el-button>
              </div>
            </div>
          </div>

          <!-- 垂直拖拽条 -->
          <div class="resize-handle-v" :class="{ active: resizingV }" @mousedown="onVResizeStart" />

          <!-- Limit Ranges -->
          <div class="right-section">
            <div class="panel-title">
              {{ t('namespace.limitRanges') }}
              <span class="count-badge">{{ limitRanges.length }} {{ t('namespace.count') }}</span>
              <el-button size="small" type="primary" @click="showCreateLrDialog" style="margin-left: auto;">
                <el-icon><Plus /></el-icon> {{ t('namespace.createBtn') }}
              </el-button>
            </div>
            <div class="table-body">
              <el-table v-if="limitRanges.length > 0" :data="limitRanges" size="small" stripe>
                <el-table-column prop="name" :label="t('namespace.nameLabel')" min-width="200" />
                <el-table-column :label="t('namespace.limit')" min-width="300">
                  <template #default="{ row }">
                    <div v-for="(limit, i) in (row.limits || [])" :key="i" style="font-size: 12px; margin-bottom: 4px;">
                      <el-tag size="small" style="margin-right: 4px;">{{ limit.type }}</el-tag>
                      <span v-for="(v, k) in (limit.max || {})" :key="k">Max {{ k }}: {{ v }} </span>
                      <span v-for="(v, k) in (limit.min || {})" :key="k">Min {{ k }}: {{ v }} </span>
                    </div>
                  </template>
                </el-table-column>
                <el-table-column prop="age" :label="t('namespace.ageLabel')" width="180" />
              </el-table>
              <div v-else class="empty-hint">
                {{ t('namespace.noLimitRanges') }}
                <el-button type="primary" size="small" @click="showCreateLrDialog" style="margin-top: 8px;">{{ t('namespace.setLimitBtn') }}</el-button>
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
    <el-dialog v-model="labelsDialogVisible" :title="t('namespace.labelsTitle')" width="600px" destroy-on-close>
      <div v-for="(label, i) in labelsArray" :key="i" style="display: flex; gap: 8px; margin-bottom: 12px; align-items: center;">
        <el-input v-model="label.key" :placeholder="t('namespace.key')" style="flex: 2;" />
        <el-input v-model="label.value" :placeholder="t('namespace.value')" style="flex: 2;" />
        <el-button type="danger" circle size="small" @click="removeLabel(i)">
          <el-icon><Delete /></el-icon>
        </el-button>
      </div>
      <el-button @click="addLabel" style="margin-top: 8px;">
        <el-icon><Plus /></el-icon> {{ t('namespace.addLabel') }}
      </el-button>
      <template #footer>
        <el-button @click="labelsDialogVisible = false">{{ t('namespace.cancelButton') }}</el-button>
        <el-button type="primary" @click="handleSaveLabels">{{ t('namespace.saveBtn') }}</el-button>
      </template>
    </el-dialog>

    <!-- Annotations Dialog -->
    <el-dialog v-model="annotationsDialogVisible" :title="t('namespace.annotationsDialogTitle')" width="650px" destroy-on-close>
      <div v-for="(anno, i) in annotationsArray" :key="i" style="display: flex; gap: 8px; margin-bottom: 12px; align-items: center;">
        <el-input v-model="anno.key" :placeholder="t('namespace.key')" style="flex: 2;" />
        <el-input v-model="anno.value" :placeholder="t('namespace.value')" style="flex: 2;" />
        <el-button type="danger" circle size="small" @click="removeAnnotation(i)">
          <el-icon><Delete /></el-icon>
        </el-button>
      </div>
      <el-button @click="addAnnotation" style="margin-top: 8px;">
        <el-icon><Plus /></el-icon> {{ t('namespace.addAnnotation') }}
      </el-button>
      <template #footer>
        <el-button @click="annotationsDialogVisible = false">{{ t('namespace.cancelButton') }}</el-button>
        <el-button type="primary" :loading="annotationsSaving" @click="handleSaveAnnotations">{{ t('namespace.saveBtn') }}</el-button>
      </template>
    </el-dialog>

    <!-- Create ResourceQuota Drawer -->
    <el-drawer v-model="rqDialogVisible" :title="t('namespace.setQuota')" direction="rtl" size="500px" destroy-on-close>
      <el-form label-width="160px">
        <el-form-item :label="t('common.name')" required>
          <el-input v-model="rqForm.name" :placeholder="t('namespace.createNamePlaceholder')" />
        </el-form-item>
        <el-divider>{{ t('namespace.resourceLimits') }}</el-divider>
        <el-form-item :label="t('namespace.requestsCpu')">
          <el-input v-model="rqForm.requestsCpu" :placeholder="t('namespace.exampleCpu')" />
        </el-form-item>
        <el-form-item :label="t('namespace.requestsMemory')">
          <el-input v-model="rqForm.requestsMemory" :placeholder="t('namespace.exampleMemory')" />
        </el-form-item>
        <el-form-item :label="t('namespace.limitsCpu')">
          <el-input v-model="rqForm.limitsCpu" :placeholder="t('namespace.exampleCpu')" />
        </el-form-item>
        <el-form-item :label="t('namespace.limitsMemory')">
          <el-input v-model="rqForm.limitsMemory" :placeholder="t('namespace.exampleMemory')" />
        </el-form-item>
        <el-form-item :label="t('namespace.podCount')">
          <el-input v-model="rqForm.pods" :placeholder="t('namespace.exampleNumber')" />
        </el-form-item>
        <el-form-item :label="t('namespace.serviceCount')">
          <el-input v-model="rqForm.services" :placeholder="t('namespace.exampleNumber')" />
        </el-form-item>
        <el-form-item :label="t('namespace.pvcCount')">
          <el-input v-model="rqForm.pvcs" :placeholder="t('namespace.exampleNumber')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rqDialogVisible = false">{{ t('namespace.cancelButton') }}</el-button>
        <el-button type="primary" :loading="rqCreating" @click="handleCreateRq">{{ t('namespace.createBtn') }}</el-button>
      </template>
    </el-drawer>

    <!-- Create LimitRange Drawer -->
    <el-drawer v-model="lrDialogVisible" :title="t('namespace.setLimit')" direction="rtl" size="550px" destroy-on-close>
      <el-form label-width="140px">
        <el-form-item :label="t('common.name')" required>
          <el-input v-model="lrForm.name" :placeholder="t('namespace.createNamePlaceholder')" />
        </el-form-item>

        <div v-for="(limit, i) in lrForm.limits" :key="i" style="border: 1px solid var(--el-border-color); border-radius: 8px; padding: 16px; margin-bottom: 16px;">
          <div style="display: flex; justify-content: space-between; margin-bottom: 12px;">
            <el-select v-model="limit.type" style="width: 200px;">
              <el-option :label="t('namespace.containerType')" value="Container" />
              <el-option :label="t('namespace.podType')" value="Pod" />
              <el-option :label="t('namespace.pvcType')" value="PersistentVolumeClaim" />
            </el-select>
            <el-button v-if="lrForm.limits.length > 1" type="danger" size="small" @click="removeLrLimit(i)">{{ t('namespace.removeLimit') }}</el-button>
          </div>
          <el-form-item :label="t('namespace.maxCpu')">
            <el-input v-model="limit.maxCpu" :placeholder="t('namespace.exampleCpu')" />
          </el-form-item>
          <el-form-item :label="t('namespace.maxMemory')">
            <el-input v-model="limit.maxMemory" :placeholder="t('namespace.exampleMemory')" />
          </el-form-item>
          <el-form-item :label="t('namespace.minCpu')">
            <el-input v-model="limit.minCpu" :placeholder="t('namespace.exampleCpu')" />
          </el-form-item>
          <el-form-item :label="t('namespace.minMemory')">
            <el-input v-model="limit.minMemory" :placeholder="t('namespace.exampleMemory')" />
          </el-form-item>
          <el-form-item :label="t('namespace.defaultCpu')">
            <el-input v-model="limit.defaultCpu" :placeholder="t('namespace.exampleCpu')" />
          </el-form-item>
          <el-form-item :label="t('namespace.defaultMemory')">
            <el-input v-model="limit.defaultMemory" :placeholder="t('namespace.exampleMemory')" />
          </el-form-item>
          <el-form-item :label="t('namespace.defaultRequestCpu')">
            <el-input v-model="limit.defaultRequestCpu" :placeholder="t('namespace.exampleCpu')" />
          </el-form-item>
          <el-form-item :label="t('namespace.defaultRequestMemory')">
            <el-input v-model="limit.defaultRequestMemory" :placeholder="t('namespace.exampleMemory')" />
          </el-form-item>
        </div>
        <el-button @click="addLrLimit" style="margin-bottom: 16px;">
          <el-icon><Plus /></el-icon> {{ t('namespace.addLimit') }}
        </el-button>
      </el-form>
      <template #footer>
        <el-button @click="lrDialogVisible = false">{{ t('namespace.cancelButton') }}</el-button>
        <el-button type="primary" :loading="lrCreating" @click="handleCreateLr">{{ t('namespace.createBtn') }}</el-button>
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
