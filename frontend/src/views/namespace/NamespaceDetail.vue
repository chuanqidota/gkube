<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, PriceTag, Notebook } from '@element-plus/icons-vue'
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
import YamlEditor from '@/components/YamlEditor.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const route = useRoute()
const router = useRouter()
const namespaceStore = useNamespaceStore()
const loading = ref(false)
const namespace = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const yamlSaving = ref(false)
const activeTab = ref('info')
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

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getNamespaceYaml({ name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || '加载 YAML 失败')
  } finally {
    yamlLoading.value = false
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

function handleTabChange(tab: string) {
  if (tab === 'yaml' && !yamlContent.value) fetchYaml()
  if (tab === 'quotas' && resourceQuotas.value.length === 0) fetchResourceQuotas()
  if (tab === 'limits' && limitRanges.value.length === 0) fetchLimitRanges()
}

async function handleSaveYaml() {
  yamlSaving.value = true
  try {
    await updateNamespace({ yaml: yamlContent.value })
    ElMessage.success('YAML 保存成功')
    fetchDetail()
  } catch (e: any) {
    ElMessage.error(e?.message || '保存 YAML 失败')
  } finally {
    yamlSaving.value = false
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
    ElMessage.info('注解请通过 YAML 标签页编辑')
    annotationsDialogVisible.value = false
  } catch (e: any) {
    ElMessage.error(e?.message || '更新注解失败')
  }
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchDetail, { autoStart: false })

onMounted(fetchDetail)
</script>

<template>
  <div class="page-container" v-loading="loading">
    <div class="page-header">
      <h2 style="margin: 0;">命名空间: {{ name }}</h2>
      <div style="display: flex; gap: 8px;">
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
        <el-button type="primary" @click="handleEditLabels">标签</el-button>
        <el-button type="danger" plain @click="handleDelete">删除</el-button>
        <el-button @click="router.push('/namespaces')">返回列表</el-button>
      </div>
    </div>

    <template v-if="namespace">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <!-- 概览 Tab -->
        <el-tab-pane label="概览" name="info">
          <el-card shadow="never">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="名称">{{ namespace.name }}</el-descriptions-item>
              <el-descriptions-item label="状态">
                <el-tag :type="namespace.status === 'Active' ? 'success' : 'warning'" size="small" effect="dark">{{ namespace.status }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ namespace.age }}</el-descriptions-item>
            </el-descriptions>

            <!-- Labels -->
            <div style="margin-top: 24px;">
              <div class="section-header">
                <el-icon class="section-icon"><PriceTag /></el-icon>
                <span class="section-title">标签</span>
                <el-button size="small" @click="handleEditLabels" style="margin-left: auto;">编辑</el-button>
              </div>
              <div v-if="namespace.labels && Object.keys(namespace.labels).length > 0">
                <el-tag v-for="(v, k) in namespace.labels" :key="k" style="margin-right: 8px; margin-bottom: 8px;">{{ k }}={{ v }}</el-tag>
              </div>
              <span v-else style="color: #909399;">无标签</span>
            </div>

            <!-- Annotations -->
            <div style="margin-top: 24px;">
              <div class="section-header">
                <el-icon class="section-icon"><Notebook /></el-icon>
                <span class="section-title">注解</span>
                <el-button size="small" @click="handleEditAnnotations" style="margin-left: auto;">编辑</el-button>
              </div>
              <div v-if="namespace.annotations && Object.keys(namespace.annotations).length > 0">
                <div v-for="(v, k) in namespace.annotations" :key="k" style="margin-bottom: 4px;">
                  <span style="font-weight: 600;">{{ k }}:</span> {{ v }}
                </div>
              </div>
              <span v-else style="color: #909399;">无注解</span>
            </div>
          </el-card>
        </el-tab-pane>

        <!-- Resource Quotas Tab -->
        <el-tab-pane label="资源配额" name="quotas">
          <el-card shadow="never">
            <template #header>
              <div style="display: flex; justify-content: space-between; align-items: center;">
                <span>资源配额</span>
                <el-button type="primary" size="small" @click="showCreateRqDialog">
                  <el-icon><Plus /></el-icon> 创建资源配额
                </el-button>
              </div>
            </template>
            <el-table :data="resourceQuotas" stripe>
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
            <el-empty v-if="resourceQuotas.length === 0" description="该命名空间下暂无资源配额">
              <el-button type="primary" @click="showCreateRqDialog">创建资源配额</el-button>
            </el-empty>
          </el-card>
        </el-tab-pane>

        <!-- Limit Ranges Tab -->
        <el-tab-pane label="资源限制" name="limits">
          <el-card shadow="never">
            <template #header>
              <div style="display: flex; justify-content: space-between; align-items: center;">
                <span>资源限制</span>
                <el-button type="primary" size="small" @click="showCreateLrDialog">
                  <el-icon><Plus /></el-icon> 创建资源限制
                </el-button>
              </div>
            </template>
            <el-table :data="limitRanges" stripe>
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
            <el-empty v-if="limitRanges.length === 0" description="该命名空间下暂无资源限制">
              <el-button type="primary" @click="showCreateLrDialog">创建资源限制</el-button>
            </el-empty>
          </el-card>
        </el-tab-pane>

        <!-- YAML Tab -->
        <el-tab-pane label="YAML" name="yaml">
          <el-card shadow="never">
            <div v-loading="yamlLoading">
              <YamlEditor v-model="yamlContent" height="600px" show-save-buttons :saving="yamlSaving" @save="handleSaveYaml" @cancel="fetchYaml" />
            </div>
          </el-card>
        </el-tab-pane>
      </el-tabs>
    </template>

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

    <!-- Create ResourceQuota Dialog -->
    <el-dialog v-model="rqDialogVisible" title="创建资源配额" width="600px" destroy-on-close>
      <el-form label-width="160px">
        <el-form-item label="名称" required>
          <el-input v-model="rqForm.name" placeholder="my-resource-quota" />
        </el-form-item>
        <el-divider>资源限制</el-divider>
        <el-form-item label="requests.cpu">
          <el-input v-model="rqForm.requestsCpu" placeholder="例如: 4" />
        </el-form-item>
        <el-form-item label="requests.memory">
          <el-input v-model="rqForm.requestsMemory" placeholder="例如: 8Gi" />
        </el-form-item>
        <el-form-item label="limits.cpu">
          <el-input v-model="rqForm.limitsCpu" placeholder="例如: 8" />
        </el-form-item>
        <el-form-item label="limits.memory">
          <el-input v-model="rqForm.limitsMemory" placeholder="例如: 16Gi" />
        </el-form-item>
        <el-form-item label="pods">
          <el-input v-model="rqForm.pods" placeholder="例如: 20" />
        </el-form-item>
        <el-form-item label="services">
          <el-input v-model="rqForm.services" placeholder="例如: 10" />
        </el-form-item>
        <el-form-item label="persistentvolumeclaims">
          <el-input v-model="rqForm.pvcs" placeholder="例如: 5" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rqDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="rqCreating" @click="handleCreateRq">创建</el-button>
      </template>
    </el-dialog>

    <!-- Create LimitRange Dialog -->
    <el-dialog v-model="lrDialogVisible" title="创建资源限制" width="700px" destroy-on-close>
      <el-form label-width="140px">
        <el-form-item label="名称" required>
          <el-input v-model="lrForm.name" placeholder="my-limit-range" />
        </el-form-item>

        <div v-for="(limit, i) in lrForm.limits" :key="i" style="border: 1px solid var(--el-border-color); border-radius: 8px; padding: 16px; margin-bottom: 16px;">
          <div style="display: flex; justify-content: space-between; margin-bottom: 12px;">
            <el-select v-model="limit.type" style="width: 200px;">
              <el-option label="Container" value="Container" />
              <el-option label="Pod" value="Pod" />
              <el-option label="PersistentVolumeClaim" value="PersistentVolumeClaim" />
            </el-select>
            <el-button v-if="lrForm.limits.length > 1" type="danger" size="small" @click="removeLrLimit(i)">移除</el-button>
          </div>
          <el-form-item label="Max CPU"><el-input v-model="limit.maxCpu" placeholder="例如: 4" /></el-form-item>
          <el-form-item label="Max Memory"><el-input v-model="limit.maxMemory" placeholder="例如: 8Gi" /></el-form-item>
          <el-form-item label="Min CPU"><el-input v-model="limit.minCpu" placeholder="例如: 100m" /></el-form-item>
          <el-form-item label="Min Memory"><el-input v-model="limit.minMemory" placeholder="例如: 128Mi" /></el-form-item>
          <el-form-item label="Default CPU"><el-input v-model="limit.defaultCpu" placeholder="例如: 500m" /></el-form-item>
          <el-form-item label="Default Memory"><el-input v-model="limit.defaultMemory" placeholder="例如: 512Mi" /></el-form-item>
          <el-form-item label="Default Req CPU"><el-input v-model="limit.defaultRequestCpu" placeholder="例如: 100m" /></el-form-item>
          <el-form-item label="Default Req Memory"><el-input v-model="limit.defaultRequestMemory" placeholder="例如: 128Mi" /></el-form-item>
        </div>
        <el-button @click="addLrLimit" style="margin-bottom: 16px;">
          <el-icon><Plus /></el-icon> 添加限制
        </el-button>
      </el-form>
      <template #footer>
        <el-button @click="lrDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="lrCreating" @click="handleCreateLr">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; position: relative; min-height: 100%; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }

.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 2px solid var(--el-color-primary);
}

.section-icon {
  font-size: 20px;
  color: var(--el-color-primary);
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  letter-spacing: 1px;
}
</style>
