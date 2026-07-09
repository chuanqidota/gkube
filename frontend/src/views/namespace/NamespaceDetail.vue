<script setup lang="ts">
import { ref, onMounted } from 'vue'
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
            <el-empty v-if="resourceQuotas.length === 0" description="该命名空间下暂无资源配额" />
          </el-card>
        </el-tab-pane>

        <!-- Limit Ranges Tab -->
        <el-tab-pane label="资源限制" name="limits">
          <el-card shadow="never">
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
            <el-empty v-if="limitRanges.length === 0" description="该命名空间下暂无资源限制" />
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
