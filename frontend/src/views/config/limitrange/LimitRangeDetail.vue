<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getLimitRangeDetail, getLimitRangeYaml, deleteLimitRange } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const limitRange = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const activeTab = ref('info')

const namespace = route.params.namespace as string
const name = route.params.name as string

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getLimitRangeDetail({ namespace, name })
    limitRange.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load LimitRange detail')
  } finally { loading.value = false }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getLimitRangeYaml({ namespace, name })
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
    await ElMessageBox.confirm(`Delete LimitRange "${name}" in namespace "${namespace}"?`, 'Confirm', { type: 'warning' })
    await deleteLimitRange({ namespace, name })
    ElMessage.success('LimitRange deleted')
    router.push('/config/limitranges')
  } catch { /* cancelled */ }
}

onMounted(fetchDetail)
</script>

<template>
  <div class="page-container" v-loading="loading">
    <div class="page-header">
      <h2 style="margin: 0;">LimitRange: {{ name }}</h2>
      <div style="display: flex; gap: 8px;">
        <el-button type="danger" @click="handleDelete">Delete</el-button>
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
            </el-descriptions>

            <!-- Limits -->
            <div style="margin-top: 24px;">
              <h4>Limits</h4>
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
