<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getSecretDetail, getSecretYaml } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const secret = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const activeTab = ref('info')
const showDecoded = ref(true)

const namespace = route.params.namespace as string
const name = route.params.name as string
const clusterName = (route.query.cluster as string) || ''

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
    const res: any = await getSecretDetail({ clusterName, namespace, name })
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
    const res: any = await getSecretYaml({ clusterName, namespace, name })
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
      <h2 style="margin: 0;">Secret: {{ name }}</h2>
      <el-button @click="router.push('/config/secrets')">Back to List</el-button>
    </div>

    <template v-if="secret">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="Info" name="info">
          <el-descriptions :column="2" border style="margin-top: 8px;">
            <el-descriptions-item label="Name">{{ secret.name }}</el-descriptions-item>
            <el-descriptions-item label="Namespace">{{ secret.namespace }}</el-descriptions-item>
            <el-descriptions-item label="Type">{{ secret.type || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Age">{{ secret.age || '-' }}</el-descriptions-item>
          </el-descriptions>

          <!-- Labels -->
          <div v-if="secret.labels && Object.keys(secret.labels).length > 0" style="margin-top: 16px;">
            <h4>Labels</h4>
            <el-tag
              v-for="(val, key) in secret.labels"
              :key="key"
              style="margin-right: 8px; margin-bottom: 8px;"
            >
              {{ key }}={{ val }}
            </el-tag>
          </div>
        </el-tab-pane>

        <el-tab-pane label="Data" name="data">
          <div style="margin-bottom: 12px;">
            <el-switch v-model="showDecoded" active-text="Decoded (Base64)" inactive-text="Raw (Base64)" />
          </div>
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
