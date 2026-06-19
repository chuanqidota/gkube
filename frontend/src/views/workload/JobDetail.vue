<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getJobDetail, getJobYaml, deleteJob } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const job = ref<any>(null)
const yamlContent = ref('')
const yamlLoading = ref(false)
const activeTab = ref('info')

const namespace = route.params.namespace as string
const name = route.params.name as string
const clusterName = (route.query.cluster as string) || ''

async function fetchDetail() {
  loading.value = true
  try {
    const res: any = await getJobDetail({ clusterName, namespace, name })
    job.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load job detail')
  } finally {
    loading.value = false
  }
}

async function fetchYaml() {
  yamlLoading.value = true
  try {
    const res: any = await getJobYaml({ clusterName, namespace, name })
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

async function handleDelete() {
  try {
    await ElMessageBox.confirm(
      `Are you sure to delete Job "${name}" in namespace "${namespace}"?`,
      'Confirm Delete',
      { type: 'warning' }
    )
    await deleteJob({ clusterName, namespace, name })
    ElMessage.success('Job deleted')
    router.push('/workloads/jobs')
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
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="margin: 0;">Job: {{ name }}</h2>
      <div>
        <el-button type="danger" @click="handleDelete">Delete</el-button>
        <el-button @click="router.push('/workloads/jobs')">Back to List</el-button>
      </div>
    </div>

    <template v-if="job">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="Info" name="info">
          <el-descriptions :column="2" border style="margin-top: 8px;">
            <el-descriptions-item label="Name">{{ job.name }}</el-descriptions-item>
            <el-descriptions-item label="Namespace">{{ job.namespace }}</el-descriptions-item>
            <el-descriptions-item label="Completions">{{ job.completions ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="Parallelism">{{ job.parallelism ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="Succeeded">{{ job.succeeded ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="Active">{{ job.active ?? '-' }}</el-descriptions-item>
            <el-descriptions-item label="Start Time">{{ job.startTime || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Completion Time">{{ job.completionTime || '-' }}</el-descriptions-item>
            <el-descriptions-item label="Age">{{ job.age || '-' }}</el-descriptions-item>
          </el-descriptions>

          <!-- Labels -->
          <div v-if="job.labels && Object.keys(job.labels).length > 0" style="margin-top: 16px;">
            <h4>Labels</h4>
            <el-tag
              v-for="(val, key) in job.labels"
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
