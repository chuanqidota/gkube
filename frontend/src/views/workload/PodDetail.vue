<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getPodDetail, getPodYaml, getPodEvents } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const pod = ref<any>(null)
const events = ref<any[]>([])
const yamlContent = ref('')
const yamlLoading = ref(false)
const activeTab = ref('info')

const namespace = route.params.namespace as string
const name = route.params.name as string
const clusterName = (route.query.cluster as string) || ''

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getPodDetail({ clusterName, namespace, name })
    pod.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load pod detail')
  } finally {
    loading.value = false
  }
}

async function fetchEvents() {
  try {
    const res: any = await getPodEvents({ clusterName, namespace, name })
    events.value = res.data || []
  } catch {
    // ignore
  }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getPodYaml({ clusterName, namespace, name })
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
  if (tab === 'events' && events.value.length === 0) {
    fetchEvents()
  }
}

function statusType(status: string) {
  const s = (status || '').toLowerCase()
  if (s === 'running') return 'success'
  if (s === 'succeeded') return 'info'
  if (s === 'pending') return 'warning'
  if (s === 'failed' || s === 'error') return 'danger'
  return 'info'
}

onMounted(fetchDetail)
</script>

<template>
  <div v-loading="loading">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="margin: 0;">Pod: {{ name }}</h2>
      <el-button @click="router.push('/workloads/pods')">Back to List</el-button>
    </div>

    <template v-if="pod">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="Info" name="info">
          <el-descriptions :column="2" border style="margin-top: 8px;">
            <el-descriptions-item label="Name">{{ pod.name }}</el-descriptions-item>
            <el-descriptions-item label="Namespace">{{ pod.namespace }}</el-descriptions-item>
            <el-descriptions-item label="Status">
              <el-tag :type="statusType(pod.status)" size="small">{{ pod.status }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="IP">{{ pod.ip || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Node">{{ pod.node || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Restarts">{{ pod.restarts ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="Age">{{ pod.age || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Created">{{ pod.created_at || '-' }}</el-descriptions-item>
          </el-descriptions>

          <!-- Labels -->
          <div v-if="pod.labels && Object.keys(pod.labels).length > 0" style="margin-top: 16px;">
            <h4>Labels</h4>
            <el-tag
              v-for="(val, key) in pod.labels"
              :key="key"
              style="margin-right: 8px; margin-bottom: 8px;"
            >
              {{ key }}={{ val }}
            </el-tag>
          </div>

          <!-- Containers -->
          <div v-if="pod.containers && pod.containers.length > 0" style="margin-top: 16px;">
            <h4>Containers</h4>
            <el-table :data="pod.containers" border stripe>
              <el-table-column prop="name" label="Name" min-width="150" />
              <el-table-column prop="image" label="Image" min-width="250" show-overflow-tooltip />
              <el-table-column label="Ready" width="80">
                <template #default="{ row }">
                  <el-tag :type="row.ready ? 'success' : 'danger'" size="small">
                    {{ row.ready ? 'Yes' : 'No' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="restartCount" label="Restarts" width="100" />
              <el-table-column label="Status" width="120">
                <template #default="{ row }">
                  <el-tag :type="statusType(row.state)" size="small">{{ row.state || '-' }}</el-tag>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>

        <el-tab-pane label="Events" name="events">
          <el-table :data="events" border stripe style="margin-top: 8px;">
            <el-table-column prop="type" label="Type" width="100" />
            <el-table-column prop="reason" label="Reason" width="150" />
            <el-table-column prop="message" label="Message" min-width="300" show-overflow-tooltip />
            <el-table-column prop="last_seen" label="Last Seen" width="180" />
          </el-table>
          <el-empty v-if="events.length === 0" description="No events" />
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
