<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getIngressDetail, getIngressYaml, updateIngress, deleteIngress, getIngressEvents } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const ingress = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const yamlDialogVisible = ref(false)
const yamlEditorRef = ref<InstanceType<typeof YamlEditor>>()
const activeTab = ref('info')

// Events
const events = ref<any[]>([])
const eventsLoading = ref(false)

const namespace = route.params.namespace as string
const name = route.params.name as string

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getIngressDetail({ namespace, name })
    ingress.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load ingress detail')
  } finally {
    loading.value = false
  }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getIngressYaml({ namespace, name })
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
    const res: any = await getIngressEvents({ namespace, name })
    events.value = res.data || []
  } catch (e) {
    console.error('Failed to fetch events:', e)
  } finally {
    eventsLoading.value = false
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
    await updateIngress({ namespace, name, yaml: content })
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
      `Are you sure to delete Ingress "${name}" in namespace "${namespace}"?`,
      'Confirm Delete',
      { type: 'warning' }
    )
    await deleteIngress({ namespace, name })
    ElMessage.success('Ingress deleted')
    router.push('/ingresses')
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e?.message || 'Delete failed')
    }
  }
}

onMounted(fetchDetail)
</script>

<template>
  <div v-loading="loading">
    <!-- Header -->
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <div style="display: flex; align-items: center; gap: 12px;">
        <el-button link @click="router.push('/ingresses')">
          <el-icon><ArrowLeft /></el-icon>
        </el-button>
        <h2 style="margin: 0;">{{ name }}</h2>
        <el-tag v-if="ingress?.ingressClassName" size="small">{{ ingress.ingressClassName }}</el-tag>
        <el-tag v-if="ingress?.namespace" type="info" size="small">{{ ingress.namespace }}</el-tag>
      </div>
      <div>
        <el-button @click="handleOpenYaml">Edit YAML</el-button>
        <el-button type="danger" @click="handleDelete">Delete</el-button>
      </div>
    </div>

    <template v-if="ingress">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <!-- Info Tab -->
        <el-tab-pane label="Info" name="info">
          <el-descriptions :column="2" border style="margin-top: 8px;">
            <el-descriptions-item label="Name">{{ ingress.name }}</el-descriptions-item>
            <el-descriptions-item label="Namespace">{{ ingress.namespace }}</el-descriptions-item>
            <el-descriptions-item label="Ingress Class Name">{{ ingress.ingressClassName || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Age">{{ ingress.age || '-' }}</el-descriptions-item>
          </el-descriptions>

          <!-- Rules -->
          <div v-if="ingress.rules && ingress.rules.length > 0" style="margin-top: 16px;">
            <h4 style="margin: 0 0 8px;">Rules</h4>
            <el-table :data="ingress.rules" border stripe size="small">
              <el-table-column prop="host" label="Host" min-width="200" show-overflow-tooltip />
              <el-table-column label="Paths" min-width="300">
                <template #default="{ row }">
                  <div v-if="row.paths && row.paths.length > 0">
                    <div v-for="(p, idx) in row.paths" :key="idx" style="margin-bottom: 4px;">
                      <el-tag size="small" type="info">{{ p.pathType || 'ImplementationSpecific' }}</el-tag>
                      {{ p.path || '/' }} -> {{ p.backend?.serviceName || p.backend?.service?.name || '-' }}:{{ p.backend?.servicePort || p.backend?.service?.port?.number || '-' }}
                    </div>
                  </div>
                  <span v-else>-</span>
                </template>
              </el-table-column>
            </el-table>
          </div>

          <!-- TLS -->
          <div v-if="ingress.tls && ingress.tls.length > 0" style="margin-top: 16px;">
            <h4 style="margin: 0 0 8px;">TLS</h4>
            <el-table :data="ingress.tls" border stripe size="small">
              <el-table-column label="Hosts" min-width="200">
                <template #default="{ row }">
                  <el-tag v-for="h in (row.hosts || [])" :key="h" size="small" style="margin-right: 4px;">{{ h }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="secretName" label="Secret Name" min-width="200" />
            </el-table>
          </div>

          <!-- Labels -->
          <div v-if="ingress.labels && Object.keys(ingress.labels).length > 0" style="margin-top: 16px;">
            <h4 style="margin: 0 0 8px;">Labels</h4>
            <el-tag
              v-for="(val, key) in ingress.labels"
              :key="key"
              style="margin-right: 8px; margin-bottom: 8px;"
            >
              {{ key }}={{ val }}
            </el-tag>
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
