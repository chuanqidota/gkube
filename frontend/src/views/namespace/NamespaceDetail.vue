<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getNamespaceDetail, getNamespaceYaml, updateNamespace, deleteNamespace, getResourceQuotaList, getLimitRangeList } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const loading = ref(false)
const namespace = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const yamlEditing = ref(false)
const yamlSaving = ref(false)
const activeTab = ref('info')
const resourceQuotas = ref<any[]>([])
const limitRanges = ref<any[]>([])

const name = route.params.name as string

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getNamespaceDetail({ name })
    namespace.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load namespace detail')
  } finally { loading.value = false }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getNamespaceYaml({ name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load YAML')
  } finally { yamlLoading.value = false }
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
    await updateNamespace({ yamlContent: yamlContent.value })
    ElMessage.success('YAML saved successfully')
    yamlEditing.value = false
    fetchDetail()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to save YAML')
  } finally { yamlSaving.value = false }
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(`Delete namespace "${name}"? This will delete ALL resources in this namespace!`, 'Confirm', { type: 'error' })
    await deleteNamespace({ name })
    ElMessage.success('Namespace deleted')
    router.push('/namespaces')
  } catch { /* cancelled */ }
}

onMounted(fetchDetail)
</script>

<template>
  <div class="page-container" v-loading="loading">
    <div class="page-header">
      <h2 style="margin: 0;">Namespace: {{ name }}</h2>
      <div style="display: flex; gap: 8px;">
        <el-button type="danger" @click="handleDelete">Delete</el-button>
        <el-button @click="router.push('/namespaces')">Back to List</el-button>
      </div>
    </div>

    <template v-if="namespace">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <!-- Info Tab -->
        <el-tab-pane label="Info" name="info">
          <el-card shadow="never">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="Name">{{ namespace.name }}</el-descriptions-item>
              <el-descriptions-item label="Status">
                <el-tag :type="namespace.status === 'Active' ? 'success' : 'warning'" size="small">{{ namespace.status }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="Age">{{ namespace.age }}</el-descriptions-item>
            </el-descriptions>

            <!-- Labels -->
            <div v-if="namespace.labels && Object.keys(namespace.labels).length > 0" style="margin-top: 16px;">
              <h4>Labels</h4>
              <el-tag v-for="(v, k) in namespace.labels" :key="k" style="margin-right: 8px; margin-bottom: 8px;">{{ k }}={{ v }}</el-tag>
            </div>

            <!-- Annotations -->
            <div v-if="namespace.annotations && Object.keys(namespace.annotations).length > 0" style="margin-top: 16px;">
              <h4>Annotations</h4>
              <div v-for="(v, k) in namespace.annotations" :key="k" style="margin-bottom: 4px;">
                <span style="font-weight: 600;">{{ k }}:</span> {{ v }}
              </div>
            </div>
          </el-card>
        </el-tab-pane>

        <!-- Resource Quotas Tab -->
        <el-tab-pane label="Resource Quotas" name="quotas">
          <el-card shadow="never">
            <el-table :data="resourceQuotas" stripe>
              <el-table-column prop="name" label="Name" min-width="200" />
              <el-table-column label="Hard Limits" min-width="250">
                <template #default="{ row }">
                  <div v-for="(v, k) in (row.hard || {})" :key="k" style="font-size: 12px;">{{ k }}: {{ v }}</div>
                </template>
              </el-table-column>
              <el-table-column label="Used" min-width="250">
                <template #default="{ row }">
                  <div v-for="(v, k) in (row.used || {})" :key="k" style="font-size: 12px;">{{ k }}: {{ v }}</div>
                </template>
              </el-table-column>
              <el-table-column prop="age" label="Age" width="180" />
            </el-table>
            <el-empty v-if="resourceQuotas.length === 0" description="No ResourceQuotas in this namespace" />
          </el-card>
        </el-tab-pane>

        <!-- Limit Ranges Tab -->
        <el-tab-pane label="Limit Ranges" name="limits">
          <el-card shadow="never">
            <el-table :data="limitRanges" stripe>
              <el-table-column prop="name" label="Name" min-width="200" />
              <el-table-column label="Limits" min-width="300">
                <template #default="{ row }">
                  <div v-for="(limit, i) in (row.limits || [])" :key="i" style="font-size: 12px; margin-bottom: 4px;">
                    <el-tag size="small" style="margin-right: 4px;">{{ limit.type }}</el-tag>
                    <span v-for="(v, k) in (limit.max || {})" :key="k">Max {{ k }}: {{ v }} </span>
                    <span v-for="(v, k) in (limit.min || {})" :key="k">Min {{ k }}: {{ v }} </span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column prop="age" label="Age" width="180" />
            </el-table>
            <el-empty v-if="limitRanges.length === 0" description="No LimitRanges in this namespace" />
          </el-card>
        </el-tab-pane>

        <!-- YAML Tab -->
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
