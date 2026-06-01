<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getDeploymentDetail, getDeploymentYaml } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const deployment = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const activeTab = ref('info')

const namespace = route.params.namespace as string
const name = route.params.name as string
const clusterName = (route.query.cluster as string) || ''

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getDeploymentDetail({ clusterName, namespace, name })
    deployment.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load deployment detail')
  } finally {
    loading.value = false
  }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getDeploymentYaml({ clusterName, namespace, name })
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
      <h2 style="margin: 0;">Deployment: {{ name }}</h2>
      <el-button @click="router.push('/workloads/deployments')">Back to List</el-button>
    </div>

    <template v-if="deployment">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="Info" name="info">
          <el-descriptions :column="2" border style="margin-top: 8px;">
            <el-descriptions-item label="Name">{{ deployment.name }}</el-descriptions-item>
            <el-descriptions-item label="Namespace">{{ deployment.namespace }}</el-descriptions-item>
            <el-descriptions-item label="Replicas">{{ deployment.replicas ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="Ready">{{ deployment.ready ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="Updated">{{ deployment.updated ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="Available">{{ deployment.available ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="Strategy">{{ deployment.strategy || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Age">{{ deployment.age || '-' }}</el-descriptions-item>
          </el-descriptions>

          <!-- Labels -->
          <div v-if="deployment.labels && Object.keys(deployment.labels).length > 0" style="margin-top: 16px;">
            <h4>Labels</h4>
            <el-tag
              v-for="(val, key) in deployment.labels"
              :key="key"
              style="margin-right: 8px; margin-bottom: 8px;"
            >
              {{ key }}={{ val }}
            </el-tag>
          </div>

          <!-- Selector -->
          <div v-if="deployment.selector && Object.keys(deployment.selector).length > 0" style="margin-top: 16px;">
            <h4>Selector</h4>
            <el-tag
              v-for="(val, key) in deployment.selector"
              :key="key"
              style="margin-right: 8px; margin-bottom: 8px;"
              type="info"
            >
              {{ key }}={{ val }}
            </el-tag>
          </div>

          <!-- Conditions -->
          <div v-if="deployment.conditions && deployment.conditions.length > 0" style="margin-top: 16px;">
            <h4>Conditions</h4>
            <el-table :data="deployment.conditions" border stripe>
              <el-table-column prop="type" label="Type" width="160" />
              <el-table-column label="Status" width="100">
                <template #default="{ row }">
                  <el-tag :type="row.status === 'True' ? 'success' : 'danger'" size="small">
                    {{ row.status }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="reason" label="Reason" width="160" />
              <el-table-column prop="message" label="Message" min-width="250" show-overflow-tooltip />
              <el-table-column prop="lastUpdateTime" label="Last Update" width="180" />
            </el-table>
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
