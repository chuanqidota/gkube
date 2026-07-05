<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Delete, Edit } from '@element-plus/icons-vue'
import { getSecretDetail, getSecretYaml, updateSecret, deleteSecret } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const secret = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const activeTab = ref('info')
const showDecoded = ref(true)
const editing = ref(false)
const editYaml = ref('')
const saving = ref(false)

const namespace = route.params.namespace as string
const name = route.params.name as string

const dataEntries = ref<{ key: string; rawValue: string; decodedValue: string }[]>([])

function base64Decode(str: string): string {
  try {
    return atob(str)
  } catch {
    return str
  }
}

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getSecretDetail({ namespace, name })
    secret.value = res.data
    const data = res.data?.data || {}
    dataEntries.value = Object.entries(data).map(([key, value]) => {
      const rawValue = String(value ?? '')
      return {
        key,
        rawValue,
        decodedValue: base64Decode(rawValue),
      }
    })
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load secret detail')
  } finally {
    loading.value = false
  }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getSecretYaml({ namespace, name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load YAML')
  } finally {
    yamlLoading.value = false
  }
}

function handleTabChange(tab: string) {
  if (tab === 'yaml' && !yamlContent.value) {
    fetchYaml()
  }
}

function handleEdit() {
  editYaml.value = yamlContent.value || ''
  editing.value = true
  if (!yamlContent.value) {
    yamlLoading.value = true
    getSecretYaml({ namespace, name }).then((res: any) => {
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
    await updateSecret({ namespace, name, yaml: editYaml.value })
    ElMessage.success('Secret updated successfully')
    editing.value = false
    fetchDetail()
    yamlContent.value = editYaml.value
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to update Secret')
  } finally {
    saving.value = false
  }
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(`Delete Secret "${name}" in namespace "${namespace}"?`, 'Confirm', { type: 'warning' })
    await deleteSecret({ namespace, name })
    ElMessage.success('Secret deleted')
    router.push('/config/secrets')
  } catch { /* cancelled */ }
}

function handleRefresh() {
  fetchDetail()
  if (yamlContent.value) fetchYaml()
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchDetail, { autoStart: false })

onMounted(fetchDetail)
</script>

<template>
  <div class="page-container" v-loading="loading">
    <div class="page-header">
      <h2 style="margin: 0;">Secret: {{ name }}</h2>
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
        <el-button @click="handleRefresh"><el-icon><Refresh /></el-icon> Refresh</el-button>
        <el-button type="primary" @click="handleEdit"><el-icon><Edit /></el-icon> Edit</el-button>
        <el-button type="danger" @click="handleDelete"><el-icon><Delete /></el-icon> 删除</el-button>
        <el-button @click="router.push('/config/secrets')">Back to List</el-button>
      </div>
    </div>

    <template v-if="secret">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="Info" name="info">
          <el-card shadow="never">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="Name">{{ secret.name || secret.metadata?.name }}</el-descriptions-item>
              <el-descriptions-item label="Namespace">{{ secret.namespace || secret.metadata?.namespace }}</el-descriptions-item>
              <el-descriptions-item label="Type">{{ secret.type || '-' }}</el-descriptions-item>
              <el-descriptions-item label="Age">{{ secret.age || '-' }}</el-descriptions-item>
              <el-descriptions-item label="UID">{{ secret.metadata?.uid || '-' }}</el-descriptions-item>
            </el-descriptions>
          </el-card>

          <!-- Labels -->
          <el-card shadow="never" style="margin-top: 16px;">
            <template #header><h4 style="margin: 0;">Labels</h4></template>
            <div v-if="secret.labels && Object.keys(secret.labels).length > 0">
              <el-tag
                v-for="(val, key) in secret.labels"
                :key="key"
                style="margin-right: 8px; margin-bottom: 8px;"
              >
                {{ key }}={{ val }}
              </el-tag>
            </div>
            <span v-else style="color: var(--gk-color-text-secondary);">No labels</span>
          </el-card>

          <!-- Annotations -->
          <el-card shadow="never" style="margin-top: 16px;">
            <template #header><h4 style="margin: 0;">Annotations</h4></template>
            <div v-if="secret.annotations && Object.keys(secret.annotations).length > 0">
              <div v-for="(val, key) in secret.annotations" :key="key" class="annotation-item">
                <span class="annotation-key">{{ key }}</span>
                <span class="annotation-value">{{ val }}</span>
              </div>
            </div>
            <span v-else style="color: var(--gk-color-text-secondary);">No annotations</span>
          </el-card>

          <!-- Data -->
          <el-card shadow="never" style="margin-top: 16px;">
            <template #header>
              <div style="display: flex; justify-content: space-between; align-items: center;">
                <h4 style="margin: 0;">Data</h4>
                <el-switch v-model="showDecoded" active-text="Decoded" inactive-text="Base64" />
              </div>
            </template>
            <el-table :data="dataEntries" border stripe>
              <el-table-column prop="key" label="Key" min-width="200" show-overflow-tooltip />
              <el-table-column label="Value" min-width="400">
                <template #default="{ row }">
                  <div style="white-space: pre-wrap; word-break: break-all; max-height: 150px; overflow-y: auto;">
                    {{ showDecoded ? row.decodedValue : row.rawValue }}
                  </div>
                </template>
              </el-table-column>
            </el-table>
            <el-empty v-if="dataEntries.length === 0" description="No data" />
          </el-card>
        </el-tab-pane>

        <el-tab-pane label="YAML" name="yaml">
          <el-card shadow="never">
            <div v-if="editing">
              <div style="margin-bottom: 12px; display: flex; gap: 8px;">
                <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
                <el-button @click="handleCancelEdit">取消</el-button>
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
