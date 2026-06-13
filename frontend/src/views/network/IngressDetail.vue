<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { getIngressDetail, getIngressYaml } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const loading = ref(false)
const ingress = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const activeTab = ref('info')

const namespace = route.params.namespace as string
const name = route.params.name as string
const clusterName = (route.query.cluster as string) || ''

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getIngressDetail({ clusterName, namespace, name })
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
    const res: any = await getIngressYaml({ clusterName, namespace, name })
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
      <h2 style="margin: 0;">Ingress: {{ name }}</h2>
      <el-button @click="router.push('/ingresses')">Back to List</el-button>
    </div>

    <template v-if="ingress">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="Info" name="info">
          <el-descriptions :column="2" border style="margin-top: 8px;">
            <el-descriptions-item label="Name">{{ ingress.name }}</el-descriptions-item>
            <el-descriptions-item label="Namespace">{{ ingress.namespace }}</el-descriptions-item>
            <el-descriptions-item label="Ingress Class Name">{{ ingress.ingressClassName || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Age">{{ ingress.age || '-' }}</el-descriptions-item>
          </el-descriptions>

          <!-- Rules -->
          <div v-if="ingress.rules && ingress.rules.length > 0" style="margin-top: 16px;">
            <h4>Rules</h4>
            <el-table :data="ingress.rules" border stripe>
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
            <h4>TLS</h4>
            <el-table :data="ingress.tls" border stripe>
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
            <h4>Labels</h4>
            <el-tag
              v-for="(val, key) in ingress.labels"
              :key="key"
              style="margin-right: 8px; margin-bottom: 8px;"
            >
              {{ key }}={{ val }}
            </el-tag>
          </div>
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
