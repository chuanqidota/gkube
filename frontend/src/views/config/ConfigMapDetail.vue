<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getConfigMapDetail, getConfigMapYaml } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const configMap = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const activeTab = ref('info')

const namespace = route.params.namespace as string
const name = route.params.name as string
const clusterName = (route.query.cluster as string) || ''

const dataEntries = ref<{ key: string; value: string }[]>([])

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getConfigMapDetail({ clusterName, namespace, name })
    configMap.value = res.data
    const data = res.data?.data || {}
    dataEntries.value = Object.entries(data).map(([key, value]) => ({
      key,
      value: String(value ?? ''),
    }))
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load configmap detail')
  } finally {
    loading.value = false
  }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getConfigMapYaml({ clusterName, namespace, name })
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

onMounted(fetchDetail)
</script>

<template>
  <div v-loading="loading">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="margin: 0;">ConfigMap: {{ name }}</h2>
      <el-button @click="router.push('/config/configmaps')">Back to List</el-button>
    </div>

    <template v-if="configMap">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="Info" name="info">
          <el-descriptions :column="2" border style="margin-top: 8px;">
            <el-descriptions-item label="Name">{{ configMap.name }}</el-descriptions-item>
            <el-descriptions-item label="Namespace">{{ configMap.namespace }}</el-descriptions-item>
            <el-descriptions-item label="Age">{{ configMap.age || '-' }}</el-descriptions-item>
          </el-descriptions>

          <!-- Labels -->
          <div v-if="configMap.labels && Object.keys(configMap.labels).length > 0" style="margin-top: 16px;">
            <h4>Labels</h4>
            <el-tag
              v-for="(val, key) in configMap.labels"
              :key="key"
              style="margin-right: 8px; margin-bottom: 8px;"
            >
              {{ key }}={{ val }}
            </el-tag>
          </div>
        </el-tab-pane>

        <el-tab-pane label="Data" name="data">
          <el-table :data="dataEntries" border stripe style="margin-top: 8px;">
            <el-table-column prop="key" label="Key" min-width="200" show-overflow-tooltip />
            <el-table-column prop="value" label="Value" min-width="400">
              <template #default="{ row }">
                <div style="white-space: pre-wrap; word-break: break-all; max-height: 150px; overflow-y: auto;">{{ row.value }}</div>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="dataEntries.length === 0" description="No data" />
        </el-tab-pane>

        <el-tab-pane label="YAML" name="yaml">
          <div v-loading="yamlLoading">
            <YamlEditor v-model="yamlContent" height="600px" read-only />
          </div>
        </el-tab-pane>
      </el-tabs>
    </template>
  </div>
</template>
