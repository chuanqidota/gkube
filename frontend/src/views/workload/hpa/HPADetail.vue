<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getHpaDetail, getHpaYaml, updateHpa, deleteHpa } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const hpa = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const yamlEditing = ref(false)
const yamlSaving = ref(false)
const activeTab = ref('info')

const namespace = route.params.namespace as string
const name = route.params.name as string

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getHpaDetail({ namespace, name })
    hpa.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load HPA detail')
  } finally { loading.value = false }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getHpaYaml({ namespace, name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load YAML')
  } finally { yamlLoading.value = false }
}

function handleTabChange(tab: string) {
  if (tab === 'yaml' && !yamlContent.value) fetchYaml()
}

async function handleSaveYaml() {
  yamlSaving.value = true
  try {
    await updateHpa({ namespace, yamlContent: yamlContent.value })
    ElMessage.success('YAML saved successfully')
    yamlEditing.value = false
    fetchDetail()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to save YAML')
  } finally { yamlSaving.value = false }
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(`Delete HPA "${name}" in namespace "${namespace}"?`, 'Confirm', { type: 'warning' })
    await deleteHpa({ namespace, name })
    ElMessage.success('HPA deleted')
    router.push('/workloads/hpa')
  } catch { /* cancelled */ }
}

onMounted(fetchDetail)
</script>

<template>
  <div class="page-container" v-loading="loading">
    <div class="page-header">
      <h2 style="margin: 0;">HPA: {{ name }}</h2>
      <div style="display: flex; gap: 8px;">
        <el-button type="danger" @click="handleDelete">Delete</el-button>
        <el-button @click="router.push('/workloads/hpa')">Back to List</el-button>
      </div>
    </div>

    <template v-if="hpa">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="Info" name="info">
          <el-card shadow="never">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="Name">{{ hpa.metadata?.name || hpa.name }}</el-descriptions-item>
              <el-descriptions-item label="Namespace">{{ hpa.metadata?.namespace || hpa.namespace }}</el-descriptions-item>
              <el-descriptions-item label="Scale Target">{{ hpa.spec?.scaleTargetRef?.kind }}/{{ hpa.spec?.scaleTargetRef?.name }}</el-descriptions-item>
              <el-descriptions-item label="Min Replicas">{{ hpa.spec?.minReplicas ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="Max Replicas">{{ hpa.spec?.maxReplicas ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="Current Replicas">{{ hpa.status?.currentReplicas ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="Desired Replicas">{{ hpa.status?.desiredReplicas ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="Age">{{ hpa.metadata?.creationTimestamp || '-' }}</el-descriptions-item>
            </el-descriptions>

            <!-- Metrics -->
            <div v-if="hpa.spec?.metrics" style="margin-top: 24px;">
              <h4>Metrics</h4>
              <el-table :data="hpa.spec.metrics" border stripe>
                <el-table-column prop="type" label="Type" width="120" />
                <el-table-column label="Resource" width="120">
                  <template #default="{ row }">{{ row.resource?.name || '-' }}</template>
                </el-table-column>
                <el-table-column label="Target Type" width="120">
                  <template #default="{ row }">{{ row.resource?.target?.type || '-' }}</template>
                </el-table-column>
                <el-table-column label="Target Value">
                  <template #default="{ row }">{{ row.resource?.target?.averageUtilization || row.resource?.target?.averageValue || row.resource?.target?.value || '-' }}</template>
                </el-table-column>
              </el-table>
            </div>

            <!-- Conditions -->
            <div v-if="hpa.status?.conditions" style="margin-top: 24px;">
              <h4>Conditions</h4>
              <el-table :data="hpa.status.conditions" border stripe>
                <el-table-column prop="type" label="Type" width="180" />
                <el-table-column label="Status" width="100">
                  <template #default="{ row }"><el-tag :type="row.status === 'True' ? 'success' : 'danger'" size="small">{{ row.status }}</el-tag></template>
                </el-table-column>
                <el-table-column prop="reason" label="Reason" width="180" />
                <el-table-column prop="message" label="Message" min-width="250" show-overflow-tooltip />
              </el-table>
            </div>
          </el-card>
        </el-tab-pane>

        <el-tab-pane label="YAML" name="yaml">
          <el-card shadow="never">
            <div style="margin-bottom: 12px; display: flex; gap: 8px;">
              <el-button v-if="!yamlEditing" type="primary" @click="yamlEditing = true">Edit YAML</el-button>
              <template v-if="yamlEditing">
                <el-button type="success" :loading="yamlSaving" @click="handleSaveYaml">Save</el-button>
                <el-button @click="yamlEditing = false; fetchYaml()">Cancel</el-button>
              </template>
            </div>
            <div v-loading="yamlLoading">
              <YamlEditor v-model="yamlContent" height="600px" :read-only="!yamlEditing" />
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
</style>
