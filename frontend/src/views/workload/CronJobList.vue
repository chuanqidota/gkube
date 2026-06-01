<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getCronJobList, getCronJobYaml, deleteCronJob, getNamespaceList } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const loading = ref(false)
const cronJobList = ref<any[]>([])
const namespaceList = ref<string[]>([])
const selectedNamespace = ref('')
const clusterName = ref('')

// Pagination
const total = ref(0)
const page = ref(1)
const size = ref(10)

// YAML dialog
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)

async function fetchNamespaces() {
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = (res.data || []).map((ns: any) => ns.name || ns)
  } catch {
    // ignore
  }
}

async function fetchCronJobs() {
  loading.value = true
  try {
    const params: any = { page: page.value, size: size.value }
    if (selectedNamespace.value) params.namespace = selectedNamespace.value
    if (clusterName.value) params.clusterName = clusterName.value
    const res: any = await getCronJobList(params)
    cronJobList.value = res.data?.items || res.data || []
    total.value = res.data?.total || 0
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load cronjobs')
  } finally {
    loading.value = false
  }
}

function handleNamespaceChange() {
  page.value = 1
  fetchCronJobs()
}

function handlePageChange(newPage: number) {
  page.value = newPage
  fetchCronJobs()
}

async function handleViewYaml(row: any) {
  yamlDialogVisible.value = true
  yamlLoading.value = true
  yamlContent.value = ''
  try {
    const res: any = await getCronJobYaml({
      clusterName: clusterName.value,
      namespace: row.namespace,
      name: row.name,
    })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load YAML')
    yamlDialogVisible.value = false
  } finally {
    yamlLoading.value = false
  }
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(
      `Delete cronjob "${row.name}" in namespace "${row.namespace}"?`,
      'Confirm',
      { type: 'warning' }
    )
    await deleteCronJob({ clusterName: clusterName.value, namespace: row.namespace, name: row.name })
    ElMessage.success('CronJob deleted')
    fetchCronJobs()
  } catch {
    // cancelled
  }
}

onMounted(() => {
  fetchNamespaces()
  fetchCronJobs()
})
</script>

<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="margin: 0;">CronJobs</h2>
      <div style="display: flex; gap: 12px; align-items: center;">
        <el-input
          v-model="clusterName"
          placeholder="Cluster Name"
          style="width: 180px;"
          clearable
          @clear="fetchCronJobs"
          @keyup.enter="fetchCronJobs"
        />
        <el-select
          v-model="selectedNamespace"
          placeholder="All Namespaces"
          clearable
          style="width: 180px;"
          @change="handleNamespaceChange"
        >
          <el-option
            v-for="ns in namespaceList"
            :key="ns"
            :label="ns"
            :value="ns"
          />
        </el-select>
        <el-button type="primary" @click="fetchCronJobs">Refresh</el-button>
      </div>
    </div>

    <el-table :data="cronJobList" v-loading="loading" stripe style="width: 100%">
      <el-table-column prop="name" label="Name" min-width="200" show-overflow-tooltip />
      <el-table-column prop="namespace" label="Namespace" width="140" />
      <el-table-column prop="schedule" label="Schedule" width="160" />
      <el-table-column prop="suspend" label="Suspend" width="100" />
      <el-table-column prop="active" label="Active" width="100" />
      <el-table-column prop="age" label="Age" width="120" />
      <el-table-column label="Actions" width="180" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">Delete</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div style="display: flex; justify-content: flex-end; margin-top: 16px;">
      <el-pagination
        v-if="total > size"
        :current-page="page"
        :page-size="size"
        :total="total"
        layout="prev, pager, next"
        @current-change="handlePageChange"
      />
    </div>

    <!-- YAML Dialog -->
    <el-dialog v-model="yamlDialogVisible" title="CronJob YAML" width="70%" top="5vh" destroy-on-close>
      <div v-loading="yamlLoading">
        <YamlEditor v-model="yamlContent" height="500px" read-only />
      </div>
    </el-dialog>
  </div>
</template>
