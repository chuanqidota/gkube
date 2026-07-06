<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Search } from '@element-plus/icons-vue'
import {
  getNamespaceList,
  createNamespace,
  deleteNamespace,
  getNamespaceYaml,
  updateNamespace,
  updateNamespaceLabels,
  transformNamespaces,
  type Namespace,
} from '@/api/resource'
import { useNamespaceStore } from '@/stores/namespace'
import YamlEditor from '@/components/YamlEditor.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'

const router = useRouter()
const namespaceStore = useNamespaceStore()
const loading = ref(false)
const namespaceList = ref<Namespace[]>([])
const searchName = ref('')

// Create dialog
const createDialogVisible = ref(false)
const creating = ref(false)
const createForm = ref({
  name: '',
  labels: [] as Array<{ key: string; value: string }>,
  annotations: [] as Array<{ key: string; value: string }>,
})

// YAML dialog
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)
const yamlTarget = ref<Namespace | null>(null)
const yamlEditorRef = ref<InstanceType<typeof YamlEditor>>()

// Labels dialog
const labelsDialogVisible = ref(false)
const labelsTarget = ref<Namespace | null>(null)
const labelsArray = ref<Array<{ key: string; value: string }>>([])

const filteredList = computed(() => {
  if (!searchName.value) return namespaceList.value
  const keyword = searchName.value.toLowerCase()
  return namespaceList.value.filter((ns) => ns.name?.toLowerCase().includes(keyword))
})

async function fetchNamespaces() {
  loading.value = true
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = transformNamespaces(res.data || [])
  } catch {
    // Silently handle — resource may not exist in cluster
  } finally {
    loading.value = false
  }
}

function statusType(status: string) {
  if (status === 'Active') return 'success'
  if (status === 'Terminating') return 'warning'
  return 'info'
}

// Create
function addLabel() { createForm.value.labels.push({ key: '', value: '' }) }
function removeLabel(i: number) { createForm.value.labels.splice(i, 1) }
function addAnnotation() { createForm.value.annotations.push({ key: '', value: '' }) }
function removeAnnotation(i: number) { createForm.value.annotations.splice(i, 1) }

async function handleCreate() {
  if (!createForm.value.name.trim()) {
    ElMessage.warning('请输入命名空间名称')
    return
  }
  creating.value = true
  try {
    const labels: Record<string, string> = {}
    createForm.value.labels.forEach((l) => {
      if (l.key.trim()) labels[l.key.trim()] = l.value
    })
    const annotations: Record<string, string> = {}
    createForm.value.annotations.forEach((a) => {
      if (a.key.trim()) annotations[a.key.trim()] = a.value
    })
    await createNamespace({
      namespace: createForm.value.name.trim(),
      labels: Object.keys(labels).length > 0 ? labels : undefined,
      annotations: Object.keys(annotations).length > 0 ? annotations : undefined,
    })
    ElMessage.success('命名空间已创建')
    createDialogVisible.value = false
    createForm.value = { name: '', labels: [], annotations: [] }
    namespaceStore.clearCache()
    fetchNamespaces()
  } catch (e: any) {
    ElMessage.error(e?.message || '创建命名空间失败')
  } finally {
    creating.value = false
  }
}

// Delete
async function handleDelete(row: Namespace) {
  try {
    await ElMessageBox.confirm(
      `确定要删除命名空间 "${row.name}" 吗？该命名空间下的所有资源将被删除。`,
      '确认删除',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    await deleteNamespace({ name: row.name })
    ElMessage.success('命名空间已删除')
    namespaceStore.clearCache()
    fetchNamespaces()
  } catch {
    // cancelled
  }
}

// YAML
async function handleViewYaml(row: Namespace) {
  yamlTarget.value = row
  yamlDialogVisible.value = true
  yamlLoading.value = true
  yamlContent.value = ''
  try {
    const res: any = await getNamespaceYaml({ name: row.name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || '加载 YAML 失败')
    yamlDialogVisible.value = false
  } finally {
    yamlLoading.value = false
  }
}

async function handleSaveYaml(content: string) {
  if (!yamlTarget.value) return
  try {
    await updateNamespace({ yaml: content })
    ElMessage.success('YAML 已保存')
    yamlDialogVisible.value = false
    fetchNamespaces()
  } catch (e: any) {
    ElMessage.error(e?.message || '保存 YAML 失败')
    yamlEditorRef.value?.resetSaving()
  }
}

// Labels
function handleLabels(row: Namespace) {
  labelsTarget.value = row
  labelsArray.value = Object.entries(row.labels || {}).map(([key, value]) => ({ key, value }))
  if (labelsArray.value.length === 0) labelsArray.value = [{ key: '', value: '' }]
  labelsDialogVisible.value = true
}

function addEditLabel() { labelsArray.value.push({ key: '', value: '' }) }
function removeEditLabel(i: number) { labelsArray.value.splice(i, 1) }

async function handleSaveLabels() {
  if (!labelsTarget.value) return
  try {
    const labels: Record<string, string> = {}
    labelsArray.value.forEach((l) => {
      if (l.key.trim()) labels[l.key.trim()] = l.value
    })
    await updateNamespaceLabels({ namespace: labelsTarget.value.name, labels })
    ElMessage.success('标签已更新')
    labelsDialogVisible.value = false
    fetchNamespaces()
  } catch (e: any) {
    ElMessage.error(e?.message || '更新标签失败')
  }
}

function handleDetail(row: Namespace) {
  router.push(`/namespaces/${row.name}`)
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchNamespaces)

onMounted(fetchNamespaces)
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <el-input
          v-model="searchName"
          placeholder="搜索命名空间名称"
          style="width: 220px;"
          clearable
        >
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
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
        <el-button type="success" @click="createDialogVisible = true">
          <el-icon><Plus /></el-icon> 创建
        </el-button>
        <span class="total-count" v-if="namespaceList.length">共 {{ namespaceList.length }} 个</span>
      </div>
    </el-card>

    <el-card shadow="never" class="table-card">
      <el-table :data="filteredList" v-loading="loading" stripe>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small" effect="dark">{{ row.status || 'Unknown' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <el-button link type="primary" @click="handleDetail(row)">{{ row.name }}</el-button>
          </template>
        </el-table-column>
        <el-table-column label="标签" min-width="250" show-overflow-tooltip>
          <template #default="{ row }">
            <el-tag
              v-for="(v, k) in (row.labels || {})"
              :key="k"
              size="small"
              style="margin-right: 4px; margin-bottom: 2px;"
            >{{ k }}={{ v }}</el-tag>
            <span v-if="!row.labels || Object.keys(row.labels).length === 0" style="color: var(--el-text-color-secondary);">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="age" label="创建时间" width="180" />
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
            <el-button size="small" type="primary" @click="handleLabels(row)">标签</el-button>
            <el-button size="small" type="danger" plain @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Create Namespace Dialog -->
    <el-dialog v-model="createDialogVisible" title="创建命名空间" width="580px" destroy-on-close>
      <el-form label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="createForm.name" placeholder="my-namespace" />
        </el-form-item>
        <el-form-item label="标签">
          <div style="width: 100%;">
            <div v-for="(label, i) in createForm.labels" :key="i" style="display: flex; gap: 8px; margin-bottom: 8px;">
              <el-input v-model="label.key" placeholder="Key" style="flex: 1;" />
              <el-input v-model="label.value" placeholder="Value" style="flex: 1;" />
              <el-button type="danger" circle size="small" @click="removeLabel(i)"><el-icon><Delete /></el-icon></el-button>
            </div>
            <el-button size="small" @click="addLabel"><el-icon><Plus /></el-icon> 添加标签</el-button>
          </div>
        </el-form-item>
        <el-form-item label="注解">
          <div style="width: 100%;">
            <div v-for="(anno, i) in createForm.annotations" :key="i" style="display: flex; gap: 8px; margin-bottom: 8px;">
              <el-input v-model="anno.key" placeholder="Key" style="flex: 1;" />
              <el-input v-model="anno.value" placeholder="Value" style="flex: 1;" />
              <el-button type="danger" circle size="small" @click="removeAnnotation(i)"><el-icon><Delete /></el-icon></el-button>
            </div>
            <el-button size="small" @click="addAnnotation"><el-icon><Plus /></el-icon> 添加注解</el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreate" :loading="creating">创建</el-button>
      </template>
    </el-dialog>

    <!-- YAML Dialog -->
    <el-dialog v-model="yamlDialogVisible" :title="`命名空间 YAML: ${yamlTarget?.name}`" width="70%" top="5vh" destroy-on-close>
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
    </el-dialog>

    <!-- Labels Dialog -->
    <el-dialog v-model="labelsDialogVisible" :title="`管理标签: ${labelsTarget?.name}`" width="600px" destroy-on-close>
      <div v-for="(label, i) in labelsArray" :key="i" style="display: flex; gap: 8px; margin-bottom: 12px; align-items: center;">
        <el-input v-model="label.key" placeholder="Key" style="flex: 2;" />
        <el-input v-model="label.value" placeholder="Value" style="flex: 2;" />
        <el-button type="danger" circle size="small" @click="removeEditLabel(i)">
          <el-icon><Delete /></el-icon>
        </el-button>
      </div>
      <el-button @click="addEditLabel" style="margin-top: 8px;">
        <el-icon><Plus /></el-icon> 添加标签
      </el-button>
      <template #footer>
        <el-button @click="labelsDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveLabels">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; gap: 12px; align-items: center; flex-wrap: wrap; }
.table-card { border-radius: 8px; }
.total-count { color: var(--el-text-color-secondary); font-size: 13px; margin-left: auto; }
</style>
