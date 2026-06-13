<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getPdbDetail, getPdbYaml, deletePdb } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const loading = ref(false)
const pdb = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const activeTab = ref('info')

const namespace = route.params.namespace as string
const name = route.params.name as string

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getPdbDetail({ namespace, name })
    pdb.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load PDB detail')
  } finally { loading.value = false }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getPdbYaml({ namespace, name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load YAML')
  } finally { yamlLoading.value = false }
}

function handleTabChange(tab: string) {
  if (tab === 'yaml' && !yamlContent.value) fetchYaml()
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(`Delete PDB "${name}" in namespace "${namespace}"?`, 'Confirm', { type: 'warning' })
    await deletePdb({ namespace, name })
    ElMessage.success('PDB deleted')
    router.push('/workloads/pdb')
  } catch { /* cancelled */ }
}

onMounted(fetchDetail)
</script>

<template>
  <div class="page-container" v-loading="loading">
    <div class="page-header">
      <h2 style="margin: 0;">PodDisruptionBudget: {{ name }}</h2>
      <div style="display: flex; gap: 8px;">
        <el-button type="danger" @click="handleDelete">Delete</el-button>
        <el-button @click="router.push('/workloads/pdb')">Back to List</el-button>
      </div>
    </div>

    <template v-if="pdb">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="Info" name="info">
          <el-card shadow="never">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="Name">{{ pdb.name || pdb.metadata?.name }}</el-descriptions-item>
              <el-descriptions-item label="Namespace">{{ pdb.namespace || pdb.metadata?.namespace }}</el-descriptions-item>
              <el-descriptions-item label="Min Available">{{ pdb.min_available || pdb.spec?.minAvailable || '-' }}</el-descriptions-item>
              <el-descriptions-item label="Max Unavailable">{{ pdb.max_unavailable || pdb.spec?.maxUnavailable || '-' }}</el-descriptions-item>
              <el-descriptions-item label="Selector">{{ pdb.selector || pdb.spec?.selector?.matchLabels || '-' }}</el-descriptions-item>
              <el-descriptions-item label="Current Healthy">{{ pdb.status?.currentHealthy ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="Desired Healthy">{{ pdb.status?.desiredHealthy ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="Allowed Disruptions">{{ pdb.allowed ?? pdb.status?.disruptionsAllowed ?? '-' }}</el-descriptions-item>
              <el-descriptions-item label="Age">{{ pdb.age || pdb.metadata?.creationTimestamp || '-' }}</el-descriptions-item>
            </el-descriptions>

            <!-- Conditions -->
            <div v-if="pdb.status?.conditions" style="margin-top: 24px;">
              <h4>Conditions</h4>
              <el-table :data="pdb.status.conditions" border stripe>
                <el-table-column prop="type" label="Type" width="180" />
                <el-table-column label="Status" width="100">
                  <template #default="{ row }">
                    <el-tag :type="row.status === 'True' ? 'success' : 'danger'" size="small">{{ row.status }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="reason" label="Reason" width="180" />
                <el-table-column prop="message" label="Message" min-width="250" show-overflow-tooltip />
              </el-table>
            </div>
          </el-card>
        </el-tab-pane>

        <el-tab-pane label="YAML" name="yaml">
          <el-card shadow="never">
            <div v-loading="yamlLoading">
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
</style>
