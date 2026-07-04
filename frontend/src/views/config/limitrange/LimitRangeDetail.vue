<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Delete, Edit } from '@element-plus/icons-vue'
import { getLimitRangeDetail, getLimitRangeYaml, updateLimitRange, deleteLimitRange } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const limitRange = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const activeTab = ref('info')
const editing = ref(false)
const editYaml = ref('')
const saving = ref(false)

const namespace = route.params.namespace as string
const name = route.params.name as string

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getLimitRangeDetail({ namespace, name })
    limitRange.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load LimitRange detail')
  } finally {
    loading.value = false
  }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getLimitRangeYaml({ namespace, name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load YAML')
  } finally {
    yamlLoading.value = false
  }
}

function handleTabChange(tab: string) {
  if (tab === 'yaml' && !yamlContent.value) fetchYaml()
}

function handleEdit() {
  editYaml.value = yamlContent.value || ''
  editing.value = true
  if (!yamlContent.value) {
    yamlLoading.value = true
    getLimitRangeYaml({ namespace, name }).then((res: any) => {
      editYaml.value = res.data?.yaml || res.data || ''
      yamlContent.value = editYaml.value
    }).catch((e: any) => {
      ElMessage.error(e?.message || 'Failed to load YAML')
    }).finally(() => {
      yamlLoading.value = false
    })
  }
}

function handleCancelEdit() {
  editing.value = false
}

async function handleSave() {
  saving.value = true
  try {
    await updateLimitRange({ namespace, yaml: editYaml.value })
    ElMessage.success('LimitRange updated successfully')
    editing.value = false
    fetchDetail()
    yamlContent.value = editYaml.value
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to update LimitRange')
  } finally {
    saving.value = false
  }
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(`Delete LimitRange "${name}" in namespace "${namespace}"?`, 'Confirm', { type: 'warning' })
    await deleteLimitRange({ namespace, name })
    ElMessage.success('LimitRange deleted')
    router.push('/config/limitranges')
  } catch { /* cancelled */ }
}

function handleRefresh() {
  fetchDetail()
  if (yamlContent.value) fetchYaml()
}

onMounted(fetchDetail)
</script>

<template>
  <div class="page-container" v-loading="loading">
    <div class="page-header">
      <h2 style="margin: 0;">LimitRange: {{ name }}</h2>
      <div style="display: flex; gap: 8px;">
        <el-button @click="handleRefresh"><el-icon><Refresh /></el-icon> Refresh</el-button>
        <el-button type="primary" @click="handleEdit"><el-icon><Edit /></el-icon> Edit</el-button>
        <el-button type="danger" @click="handleDelete"><el-icon><Delete /></el-icon> Delete</el-button>
        <el-button @click="router.push('/config/limitranges')">Back to List</el-button>
      </div>
    </div>

    <template v-if="limitRange">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="Info" name="info">
          <el-card shadow="never">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="Name">{{ limitRange.name || limitRange.metadata?.name }}</el-descriptions-item>
              <el-descriptions-item label="Namespace">{{ limitRange.namespace || limitRange.metadata?.namespace }}</el-descriptions-item>
              <el-descriptions-item label="Age">{{ limitRange.age || limitRange.metadata?.creationTimestamp || '-' }}</el-descriptions-item>
              <el-descriptions-item label="UID">{{ limitRange.metadata?.uid || '-' }}</el-descriptions-item>
            </el-descriptions>
          </el-card>

          <!-- Labels -->
          <el-card shadow="never" style="margin-top: 16px;">
            <template #header><h4 style="margin: 0;">Labels</h4></template>
            <div v-if="limitRange.labels && Object.keys(limitRange.labels).length > 0">
              <el-tag v-for="(val, key) in limitRange.labels" :key="key" style="margin: 4px;">
                {{ key }}={{ val }}
              </el-tag>
            </div>
            <span v-else style="color: var(--gk-color-text-secondary);">No labels</span>
          </el-card>

          <!-- Annotations -->
          <el-card shadow="never" style="margin-top: 16px;">
            <template #header><h4 style="margin: 0;">Annotations</h4></template>
            <div v-if="limitRange.annotations && Object.keys(limitRange.annotations).length > 0">
              <div v-for="(val, key) in limitRange.annotations" :key="key" class="annotation-item">
                <span class="annotation-key">{{ key }}</span>
                <span class="annotation-value">{{ val }}</span>
              </div>
            </div>
            <span v-else style="color: var(--gk-color-text-secondary);">No annotations</span>
          </el-card>

          <!-- Limits -->
          <el-card shadow="never" style="margin-top: 16px;">
            <template #header><h4 style="margin: 0;">Limits</h4></template>
            <el-table :data="limitRange.limits || limitRange.spec?.limits || []" border stripe>
              <el-table-column prop="type" label="Type" width="140" />
              <el-table-column label="Max" min-width="200">
                <template #default="{ row }">
                  <div v-for="(v, k) in (row.max || {})" :key="k" style="font-size: 12px;">{{ k }}: {{ v }}</div>
                  <span v-if="!row.max || Object.keys(row.max).length === 0">-</span>
                </template>
              </el-table-column>
              <el-table-column label="Min" min-width="200">
                <template #default="{ row }">
                  <div v-for="(v, k) in (row.min || {})" :key="k" style="font-size: 12px;">{{ k }}: {{ v }}</div>
                  <span v-if="!row.min || Object.keys(row.min).length === 0">-</span>
                </template>
              </el-table-column>
              <el-table-column label="Default" min-width="200">
                <template #default="{ row }">
                  <div v-for="(v, k) in (row.default || {})" :key="k" style="font-size: 12px;">{{ k }}: {{ v }}</div>
                  <span v-if="!row.default || Object.keys(row.default).length === 0">-</span>
                </template>
              </el-table-column>
              <el-table-column label="Default Request" min-width="200">
                <template #default="{ row }">
                  <div v-for="(v, k) in (row.defaultRequest || {})" :key="k" style="font-size: 12px;">{{ k }}: {{ v }}</div>
                  <span v-if="!row.defaultRequest || Object.keys(row.defaultRequest).length === 0">-</span>
                </template>
              </el-table-column>
              <el-table-column label="Max Limit/Request Ratio" min-width="180">
                <template #default="{ row }">
                  <div v-for="(v, k) in (row.maxLimitRequestRatio || {})" :key="k" style="font-size: 12px;">{{ k }}: {{ v }}</div>
                  <span v-if="!row.maxLimitRequestRatio || Object.keys(row.maxLimitRequestRatio).length === 0">-</span>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-tab-pane>

        <el-tab-pane label="YAML" name="yaml">
          <el-card shadow="never">
            <div v-if="editing">
              <div style="margin-bottom: 12px; display: flex; gap: 8px;">
                <el-button type="primary" :loading="saving" @click="handleSave">Save</el-button>
                <el-button @click="handleCancelEdit">Cancel</el-button>
              </div>
              <YamlEditor v-model="editYaml" height="600px" />
            </div>
            <div v-else v-loading="yamlLoading">
              <div style="margin-bottom: 12px;">
                <el-button type="primary" @click="handleEdit"><el-icon><Edit /></el-icon> Edit YAML</el-button>
              </div>
              <YamlEditor v-model="yamlContent" height="600px" read-only />
            </div>
          </el-card>
        </el-tab-pane>
      </el-tabs>
    </template>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.annotation-item { display: flex; gap: 12px; margin-bottom: 8px; padding: 4px 0; border-bottom: 1px solid var(--gk-color-border-light); }
.annotation-key { font-weight: 500; min-width: 200px; word-break: break-all; }
.annotation-value { color: var(--gk-color-text-secondary); word-break: break-all; flex: 1; }
</style>
