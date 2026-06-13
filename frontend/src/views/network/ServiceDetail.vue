<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { getServiceDetail, getServiceYaml } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const loading = ref(false)
const service = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const activeTab = ref('info')

const namespace = route.params.namespace as string
const name = route.params.name as string
const clusterName = (route.query.cluster as string) || ''

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getServiceDetail({ clusterName, namespace, name })
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
    const res: any = await getServiceYaml({ clusterName, namespace, name })
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
      <h2 style="margin: 0;">Service: {{ name }}</h2>
      <el-button @click="router.push('/services')">Back to List</el-button>
    </div>

    <template v-if="service">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
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
            <h4>Selector</h4>
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
            <h4>Labels</h4>
            <el-tag
              v-for="(val, key) in service.labels"
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
