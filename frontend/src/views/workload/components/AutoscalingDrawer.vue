<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { FullScreen, Aim } from '@element-plus/icons-vue'
import { getHpaList, getHpaDetail, getHpaYaml, updateHpa } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'
import HPAStatusCard from './HPAStatusCard.vue'
import HPAEmptyState from './HPAEmptyState.vue'
import HPAForm from '@/views/workload/hpa/components/HPAForm.vue'

const props = defineProps<{
  visible: boolean
  namespace: string
  workloadName: string
  workloadKind: string
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  created: []
  deleted: []
}>()

const fullscreen = ref(false)
const loading = ref(false)
const hpaList = ref<any[]>([])
const matchedHpa = ref<any | null>(null)

// Sub-views
const showCreateForm = ref(false)
const showEditForm = ref(false)
const editHpaDetail = ref<any>(null)
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)
const yamlSaving = ref(false)

async function fetchHpaData() {
  if (!props.namespace || !props.workloadName) return
  loading.value = true
  try {
    const res: any = await getHpaList({ namespace: props.namespace })
    const items = res.data || []
    // Filter HPAs targeting this workload
    hpaList.value = items.filter((item: any) => {
      return item.target_kind === props.workloadKind && item.target === props.workloadName
    })
    if (hpaList.value.length > 0) {
      // Fetch full K8s object (list returns flat data, card needs nested metadata/spec/status)
      const detailRes: any = await getHpaDetail({
        namespace: props.namespace,
        name: hpaList.value[0].name,
      })
      matchedHpa.value = detailRes.data
    } else {
      matchedHpa.value = null
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '获取 HPA 信息失败')
    hpaList.value = []
    matchedHpa.value = null
  } finally {
    loading.value = false
  }
}

watch(() => props.visible, (val) => {
  if (val) {
    showCreateForm.value = false
    showEditForm.value = false
    editHpaDetail.value = null
    yamlDialogVisible.value = false
    yamlContent.value = ''
    fetchHpaData()
  }
})

function handleCreate() {
  showCreateForm.value = true
}

function handleCancelCreate() {
  showCreateForm.value = false
}

function handleCreated() {
  showCreateForm.value = false
  fetchHpaData()
  emit('created')
}

function handleEditSuccess() {
  showEditForm.value = false
  editHpaDetail.value = null
  fetchHpaData()
}

function handleEditCancel() {
  showEditForm.value = false
  editHpaDetail.value = null
}

function handleDeleted() {
  fetchHpaData()
  emit('deleted')
}

async function handleEdit() {
  if (!matchedHpa.value) return
  editHpaDetail.value = matchedHpa.value
  showEditForm.value = true
}

async function handleViewYaml() {
  if (!matchedHpa.value) return
  const ns = matchedHpa.value.namespace
  const name = matchedHpa.value.name
  yamlLoading.value = true
  yamlDialogVisible.value = true
  try {
    const res: any = await getHpaYaml({ namespace: ns, name })
    yamlContent.value = res.data?.yaml || ''
  } catch (e: any) {
    ElMessage.error(e?.message || '获取 YAML 失败')
    yamlContent.value = ''
  } finally {
    yamlLoading.value = false
  }
}

async function handleSaveYaml() {
  if (!matchedHpa.value) return
  yamlSaving.value = true
  try {
    await updateHpa({ namespace: matchedHpa.value.namespace, yaml: yamlContent.value })
    ElMessage.success('YAML 已保存')
    yamlDialogVisible.value = false
    fetchHpaData()
  } catch (e: any) {
    ElMessage.error(e?.message || '保存失败')
  } finally {
    yamlSaving.value = false
  }
}

function handleCancelYaml() {
  yamlDialogVisible.value = false
}
</script>

<template>
  <el-drawer
    :model-value="visible"
    @update:model-value="emit('update:visible', $event)"
    title="弹性伸缩"
    :size="fullscreen ? '100%' : '85%'"
    direction="rtl"
    destroy-on-close
    :body-style="{ padding: '0', height: '100%', overflow: 'auto' }"
  >
    <template #header>
      <div class="drawer-header">
        <span class="drawer-title">弹性伸缩</span>
        <el-tooltip :content="fullscreen ? '退出全屏' : '全屏'" placement="top">
          <el-icon class="fullscreen-btn" @click="fullscreen = !fullscreen">
            <FullScreen v-if="!fullscreen" />
            <Aim v-else />
          </el-icon>
        </el-tooltip>
      </div>
    </template>

    <div v-loading="loading" class="drawer-body">
      <!-- HPA Found -->
      <template v-if="matchedHpa && !showCreateForm && !showEditForm">
        <HPAStatusCard
          :hpa="matchedHpa"
          @edit="handleEdit"
          @yaml="handleViewYaml"
          @deleted="handleDeleted"
        />
      </template>

      <!-- Create Form -->
      <template v-else-if="showCreateForm">
        <HPAForm
          :prefill-namespace="namespace"
          :prefill-target-name="workloadName"
          :prefill-target-kind="workloadKind"
          :hide-namespace="true"
          :hide-target="true"
          :auto-name="true"
          @created="handleCreated"
          @cancel="handleCancelCreate"
        />
      </template>

      <!-- Edit Form -->
      <template v-else-if="showEditForm && editHpaDetail">
        <HPAForm
          :is-edit="true"
          :initial-data="editHpaDetail"
          @success="handleEditSuccess"
          @cancel="handleEditCancel"
        />
      </template>

      <!-- Empty State -->
      <template v-else-if="!loading">
        <HPAEmptyState
          :namespace="namespace"
          :workload-name="workloadName"
          :workload-kind="workloadKind"
          @create="handleCreate"
        />
      </template>
    </div>

    <!-- YAML Drawer (nested) -->
    <el-drawer
      v-model="yamlDialogVisible"
      title="HPA YAML"
      size="85%"
      direction="rtl"
      class="yaml-drawer"
      :body-style="{ padding: '0', height: '100%' }"
    >
      <div v-loading="yamlLoading" style="height: calc(100vh - 60px);">
        <YamlEditor
          v-model="yamlContent"
          height="100%"
          auto-format
          show-save-buttons
          :saving="yamlSaving"
          @save="handleSaveYaml"
          @cancel="handleCancelYaml"
        />
      </div>
    </el-drawer>
  </el-drawer>
</template>

<style scoped>
.drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.drawer-title {
  font-size: 16px;
  font-weight: 600;
}

.fullscreen-btn {
  cursor: pointer;
  font-size: 18px;
  color: var(--el-text-color-regular);
  transition: color 0.2s;
}

.fullscreen-btn:hover {
  color: var(--el-color-primary);
}

.drawer-body {
  min-height: 300px;
}
</style>
