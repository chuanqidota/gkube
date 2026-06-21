<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getNetworkPolicyDetail, getNetworkPolicyYaml, updateNetworkPolicy, deleteNetworkPolicy } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const np = ref<any>(null)
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
    const res: any = await getNetworkPolicyDetail({ namespace, name })
    np.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load NetworkPolicy detail')
  } finally { loading.value = false }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getNetworkPolicyYaml({ namespace, name })
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
    await updateNetworkPolicy({ namespace, yaml: yamlContent.value })
    ElMessage.success('YAML saved successfully')
    yamlEditing.value = false
    fetchDetail()
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to save YAML')
  } finally { yamlSaving.value = false }
}

async function handleDelete() {
  try {
    await ElMessageBox.confirm(`Delete NetworkPolicy "${name}" in namespace "${namespace}"?`, 'Confirm', { type: 'warning' })
    await deleteNetworkPolicy({ namespace, name })
    ElMessage.success('NetworkPolicy deleted')
    router.push('/network/networkpolicies')
  } catch { /* cancelled */ }
}

onMounted(fetchDetail)
</script>

<template>
  <div class="page-container" v-loading="loading">
    <div class="page-header">
      <h2 style="margin: 0;">NetworkPolicy: {{ name }}</h2>
      <div style="display: flex; gap: 8px;">
        <el-button type="danger" @click="handleDelete">Delete</el-button>
        <el-button @click="router.push('/network/networkpolicies')">Back to List</el-button>
      </div>
    </div>
    <template v-if="np">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="Info" name="info">
          <el-card shadow="never">
            <el-descriptions :column="2" border>
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
            <div v-loading="yamlLoading"><YamlEditor v-model="yamlContent" height="600px" :read-only="!yamlEditing" /></div>
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
