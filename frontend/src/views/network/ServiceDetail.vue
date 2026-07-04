<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getServiceDetail,
  getServiceYaml,
  updateService,
  deleteService,
  getServiceEvents,
  getServicePods,
} from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const service = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const yamlDialogVisible = ref(false)
const yamlEditorRef = ref<InstanceType<typeof YamlEditor>>()
const activeTab = ref('info')

// Events
const events = ref<any[]>([])
const eventsLoading = ref(false)

// Related Pods
const pods = ref<any[]>([])
const podsLoading = ref(false)

const namespace = route.params.namespace as string
const name = route.params.name as string

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getServiceDetail({ namespace, name })
    service.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load service detail')
  } finally {
    loading.value = false
  }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getServiceYaml({ namespace, name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load YAML')
  } finally {
    yamlLoading.value = false
  }
}

async function fetchEvents() {
  eventsLoading.value = true
  try {
    const res: any = await getServiceEvents({ namespace, name })
    events.value = res.data || []
  } catch (e) {
    console.error('Failed to fetch events:', e)
  } finally {
    eventsLoading.value = false
  }
}

async function fetchPods() {
  podsLoading.value = true
  try {
    const res: any = await getServicePods({ namespace, name })
    pods.value = res.data?.items || res.data || []
  } catch (e) {
    console.error('Failed to fetch pods:', e)
  } finally {
    podsLoading.value = false
  }
}

function handleTabChange(tab: string | number) {
  if (tab === 'yaml' && !yamlContent.value) {
    fetchYaml()
  } else if (tab === 'events' && events.value.length === 0) {
    fetchEvents()
  }
}

function handleOpenYaml() {
  fetchYaml()
  yamlDialogVisible.value = true
}

async function handleSaveYaml(content: string) {
  try {
    await updateService({ namespace, name, yaml: content })
    ElMessage.success('YAML saved successfully')
    yamlDialogVisible.value = false
    fetchDetail()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to save YAML')
    yamlEditorRef.value?.resetSaving()
  }
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(
      `Are you sure to delete Service "${name}" in namespace "${namespace}"?`,
      'Confirm Delete',
      { type: 'warning' }
    )
    await deleteService({ namespace, name })
    ElMessage.success('Service deleted')
    router.push('/services')
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || 'Delete failed')
    }
  }
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(async () => {
  fetchDetail()
  fetchPods()
}, { autoStart: false })

onMounted(() => {
  fetchDetail()
  fetchPods()
})
</script>

<template>
  <div v-loading="loading">
    <!-- Header -->
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <div style="display: flex; align-items: center; gap: 12px;">
        <el-button link @click="router.push('/services')">
          <el-icon><ArrowLeft /></el-icon>
        </el-button>
        <h2 style="margin: 0;">{{ name }}</h2>
        <el-tag v-if="service?.type" size="small">{{ service.type }}</el-tag>
        <el-tag v-if="service?.namespace" type="info" size="small">{{ service.namespace }}</el-tag>
      </div>
      <div>
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
        <el-button @click="handleOpenYaml">Edit YAML</el-button>
        <el-button type="danger" @click="handleDelete">Delete</el-button>
      </div>
    </div>

    <template v-if="service">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <!-- Info Tab -->
        <el-tab-pane label="Info" name="info">
          <el-descriptions :column="2" border style="margin-top: 8px;">
            <el-descriptions-item label="Name">{{ service.name }}</el-descriptions-item>
            <el-descriptions-item label="Namespace">{{ service.namespace }}</el-descriptions-item>
            <el-descriptions-item label="Type">{{ service.type || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Cluster IP">{{ service.clusterIP || service.cluster_ip || '-' }}</el-descriptions-item>
            <el-descriptions-item label="External IP">{{ service.externalIP || service.external_ip || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Ports">{{ service.ports || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Session Affinity">{{ service.sessionAffinity || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Age">{{ service.age || '-' }}</el-descriptions-item>
          </el-descriptions>

          <!-- Selector -->
          <div v-if="service.selector && Object.keys(service.selector).length > 0" style="margin-top: 16px;">
            <h4 style="margin: 0 0 8px;">Selector</h4>
            <el-tag
              v-for="(val, key) in service.selector"
              :key="key"
              style="margin-right: 8px; margin-bottom: 8px;"
              type="info"
            >
              {{ key }}={{ val }}
            </el-tag>
          </div>

          <!-- Labels -->
          <div v-if="service.labels && Object.keys(service.labels).length > 0" style="margin-top: 16px;">
            <h4 style="margin: 0 0 8px;">Labels</h4>
            <el-tag
              v-for="(val, key) in service.labels"
              :key="key"
              style="margin-right: 8px; margin-bottom: 8px;"
            >
              {{ key }}={{ val }}
            </el-tag>
          </div>

          <!-- Related Pods -->
          <div style="margin-top: 20px;">
            <h4 style="margin: 0 0 8px;">
              Related Pods
              <el-tag size="small" type="info" style="margin-left: 8px;">{{ pods.length }}</el-tag>
            </h4>
            <el-table :data="pods" size="small" stripe v-loading="podsLoading" max-height="300">
              <el-table-column label="Name" min-width="200">
                <template #default="{ row }">
                  <router-link
                    :to="`/pods/${row.metadata?.namespace}/${row.metadata?.name}`"
                    class="resource-link"
                  >
                    {{ row.metadata?.name }}
                  </router-link>
                </template>
              </el-table-column>
              <el-table-column label="Status" width="120">
                <template #default="{ row }">
                  <el-tag :type="row.status?.phase === 'Running' ? 'success' : row.status?.phase === 'Succeeded' ? 'info' : 'danger'" size="small">
                    {{ row.status?.phase }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="Ready" width="100">
                <template #default="{ row }">
                  {{ (row.status?.containerStatuses || []).filter((c: any) => c.ready).length }} / {{ (row.status?.containerStatuses || []).length }}
                </template>
              </el-table-column>
              <el-table-column label="Node" prop="spec.nodeName" width="150" show-overflow-tooltip />
              <el-table-column label="IP" prop="status.podIP" width="140" />
            </el-table>
          </div>
        </el-tab-pane>

        <!-- YAML Tab -->
        <el-tab-pane label="YAML" name="yaml">
          <div v-loading="yamlLoading">
            <YamlEditor v-model="yamlContent" height="600px" read-only />
          </div>
        </el-tab-pane>

        <!-- Events Tab -->
        <el-tab-pane label="Events" name="events">
          <div v-loading="eventsLoading">
            <el-table v-if="events.length > 0" :data="events" size="small" stripe>
              <el-table-column prop="type" label="Type" width="100">
                <template #default="{ row }">
                  <el-tag :type="row.type === 'Warning' ? 'danger' : 'info'" size="small">{{ row.type }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="reason" label="Reason" width="150" />
              <el-table-column prop="message" label="Message" min-width="300" show-overflow-tooltip />
              <el-table-column prop="last_seen" label="Last Seen" width="180" />
            </el-table>
            <el-empty v-else description="No events" />
          </div>
        </el-tab-pane>
      </el-tabs>
    </template>

    <!-- YAML Edit Dialog -->
    <el-dialog v-model="yamlDialogVisible" title="Edit YAML" width="70%" top="5vh" destroy-on-close>
      <div v-loading="yamlLoading">
        <YamlEditor ref="yamlEditorRef" v-model="yamlContent" height="600px" :read-only="true" :saveable="true" @save="handleSaveYaml" />
      </div>
    </el-dialog>
  </div>
</template>

<style scoped>
.resource-link {
  color: var(--el-color-primary);
  text-decoration: none;
}
.resource-link:hover {
  text-decoration: underline;
}
</style>
