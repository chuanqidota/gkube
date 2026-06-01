<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getCronJobDetail, getCronJobYaml } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const cronJob = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const activeTab = ref('info')

const namespace = route.params.namespace as string
const name = route.params.name as string
const clusterName = (route.query.cluster as string) || ''

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getCronJobDetail({ clusterName, namespace, name })
    cronJob.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load cronjob detail')
  } finally {
    loading.value = false
  }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getCronJobYaml({ clusterName, namespace, name })
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
      <h2 style="margin: 0;">CronJob: {{ name }}</h2>
      <el-button @click="router.push('/workloads/cronjobs')">Back to List</el-button>
    </div>

    <template v-if="cronJob">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="Info" name="info">
          <el-descriptions :column="2" border style="margin-top: 8px;">
            <el-descriptions-item label="Name">{{ cronJob.name }}</el-descriptions-item>
            <el-descriptions-item label="Namespace">{{ cronJob.namespace }}</el-descriptions-item>
            <el-descriptions-item label="Schedule">{{ cronJob.schedule || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Suspend">{{ cronJob.suspend ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="Active Jobs">{{ cronJob.active ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="Last Schedule Time">{{ cronJob.lastScheduleTime || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Concurrency Policy">{{ cronJob.concurrencyPolicy || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Age">{{ cronJob.age || '-' }}</el-descriptions-item>
          </el-descriptions>

          <!-- Labels -->
          <div v-if="cronJob.labels && Object.keys(cronJob.labels).length > 0" style="margin-top: 16px;">
            <h4>Labels</h4>
            <el-tag
              v-for="(val, key) in cronJob.labels"
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
