<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getNetworkPolicyDetail, getNetworkPolicyYaml, updateNetworkPolicy, deleteNetworkPolicy, getNetworkPolicyEvents } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const np = ref<any>(null)
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
    const res: any = await getNetworkPolicyDetail({ namespace, name })
    np.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load NetworkPolicy detail')
  } finally {
    loading.value = false
  }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getNetworkPolicyYaml({ namespace, name })
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
    const res: any = await getNetworkPolicyEvents({ namespace, name })
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
    await updateNetworkPolicy({ namespace, yaml: content })
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
    await ElMessageBox.confirm(`Delete NetworkPolicy "${name}" in namespace "${namespace}"?`, 'Confirm', { type: 'warning' })
    await deleteNetworkPolicy({ namespace, name })
    ElMessage.success('NetworkPolicy deleted')
    router.push('/network/networkpolicies')
  } catch {
    // cancelled
  }
}

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchDetail, { autoStart: false })

onMounted(fetchDetail)
</script>

<template>
  <div v-loading="loading">
    <!-- Header -->
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <div style="display: flex; align-items: center; gap: 12px;">
        <el-button link @click="router.push('/network/networkpolicies')">
          <el-icon><ArrowLeft /></el-icon>
        </el-button>
        <h2 style="margin: 0;">{{ name }}</h2>
        <el-tag v-if="np?.namespace" type="info" size="small">{{ np.namespace }}</el-tag>
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
        <el-button type="danger" @click="handleDelete">删除</el-button>
      </div>
    </div>

    <template v-if="np">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <!-- Info Tab -->
        <el-tab-pane label="Info" name="info">
          <el-descriptions :column="2" border style="margin-top: 8px;">
            <el-descriptions-item label="Name">{{ np.metadata?.name || np.name }}</el-descriptions-item>
            <el-descriptions-item label="Namespace">{{ np.metadata?.namespace || np.namespace }}</el-descriptions-item>
            <el-descriptions-item label="Pod Selector" :span="2">
              <el-tag v-for="(v, k) in (np.spec?.podSelector?.matchLabels || {})" :key="k" style="margin-right: 4px;">{{ k }}={{ v }}</el-tag>
              <span v-if="!np.spec?.podSelector?.matchLabels || Object.keys(np.spec.podSelector.matchLabels).length === 0">All pods</span>
            </el-descriptions-item>
            <el-descriptions-item label="Policy Types">
              <el-tag v-for="pt in (np.spec?.policyTypes || [])" :key="pt" style="margin-right: 4px;">{{ pt }}</el-tag>
            </el-descriptions-item>
          </el-descriptions>

          <!-- Labels -->
          <div v-if="np.metadata?.labels && Object.keys(np.metadata.labels).length > 0" style="margin-top: 16px;">
            <h4 style="margin: 0 0 8px;">Labels</h4>
            <el-tag
              v-for="(val, key) in np.metadata.labels"
              :key="key"
              style="margin-right: 8px; margin-bottom: 8px;"
            >
              {{ key }}={{ val }}
            </el-tag>
          </div>

          <!-- Ingress Rules -->
          <div v-if="np.spec?.ingress && np.spec.ingress.length > 0" style="margin-top: 24px;">
            <h4>Ingress Rules</h4>
            <div v-for="(rule, i) in np.spec.ingress" :key="i" style="margin-bottom: 16px;">
              <el-card shadow="never">
                <div v-if="rule.from && rule.from.length > 0">
                  <h5>From:</h5>
                  <div v-for="(from, j) in rule.from" :key="j" style="margin-bottom: 8px;">
                    <el-tag v-if="from.podSelector" type="info" size="small">Pod: {{ Object.entries(from.podSelector.matchLabels || {}).map(([k,v]) => k+'='+v).join(', ') }}</el-tag>
                    <el-tag v-if="from.namespaceSelector" type="warning" size="small">NS: {{ Object.entries(from.namespaceSelector.matchLabels || {}).map(([k,v]) => k+'='+v).join(', ') }}</el-tag>
                    <el-tag v-if="from.ipBlock" type="success" size="small">IP: {{ from.ipBlock.cidr }}</el-tag>
                  </div>
                </div>
                <div v-if="rule.ports && rule.ports.length > 0" style="margin-top: 8px;">
                  <h5>Ports:</h5>
                  <el-tag v-for="(port, j) in rule.ports" :key="j" size="small" style="margin-right: 4px;">{{ port.protocol }}/{{ port.port }}</el-tag>
                </div>
              </el-card>
            </div>
          </div>

          <!-- Egress Rules -->
          <div v-if="np.spec?.egress && np.spec.egress.length > 0" style="margin-top: 24px;">
            <h4>Egress Rules</h4>
            <div v-for="(rule, i) in np.spec.egress" :key="i" style="margin-bottom: 16px;">
              <el-card shadow="never">
                <div v-if="rule.to && rule.to.length > 0">
                  <h5>To:</h5>
                  <div v-for="(to, j) in rule.to" :key="j" style="margin-bottom: 8px;">
                    <el-tag v-if="to.podSelector" type="info" size="small">Pod: {{ Object.entries(to.podSelector.matchLabels || {}).map(([k,v]) => k+'='+v).join(', ') }}</el-tag>
                    <el-tag v-if="to.namespaceSelector" type="warning" size="small">NS: {{ Object.entries(to.namespaceSelector.matchLabels || {}).map(([k,v]) => k+'='+v).join(', ') }}</el-tag>
                    <el-tag v-if="to.ipBlock" type="success" size="small">IP: {{ to.ipBlock.cidr }}</el-tag>
                  </div>
                </div>
                <div v-if="rule.ports && rule.ports.length > 0" style="margin-top: 8px;">
                  <h5>Ports:</h5>
                  <el-tag v-for="(port, j) in rule.ports" :key="j" size="small" style="margin-right: 4px;">{{ port.protocol }}/{{ port.port }}</el-tag>
                </div>
              </el-card>
            </div>
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
    <el-drawer v-model="yamlDialogVisible" title="Edit YAML" size="85%" direction="rtl" class="yaml-drawer"
      :body-style="{ padding: '0', height: '100%' }">
      <div v-loading="yamlLoading" style="height: calc(100vh - 52px);">
        <YamlEditor ref="yamlEditorRef" v-model="yamlContent" height="100%" auto-format show-save-buttons @save="handleSaveYaml" @cancel="fetchYaml" />
      </div>
    </el-drawer>
  </div>
</template>

<style>
.yaml-drawer .el-drawer__header {
  padding: 6px 16px;
  margin-bottom: 0;
  min-height: auto;
}
</style>
