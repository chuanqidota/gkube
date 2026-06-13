<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Delete, Search } from '@element-plus/icons-vue'
import { getCustomResourceList, getCustomResourceYaml, deleteCustomResource, getNamespaceList } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const resourceList = ref<any[]>([])
const namespaceList = ref<string[]>([])
const selectedNamespace = ref('')
const searchName = ref('')
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)

const group = route.query.group as string
const version = route.query.version as string
const resource = route.query.resource as string
const scope = route.query.scope as string

const filteredList = computed(() => {
  if (!searchName.value) return resourceList.value
  const keyword = searchName.value.toLowerCase()
  return resourceList.value.filter((d) => d.name?.toLowerCase().includes(keyword))
})

async function fetchNamespaces() {
  try {
    const res: any = await getNamespaceList()
    namespaceList.value = (res.data || []).map((ns: any) => ns.name || ns)
  } catch { /* ignore */ }
}

async function fetchResources() {
  loading.value = true
  try {
    const params: any = { group, version, resource }
    if (selectedNamespace.value) params.namespace = selectedNamespace.value
    const res: any = await getCustomResourceList(params)
    resourceList.value = res.data || []
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load custom resources')
  } finally { loading.value = false }
}

function handleNamespaceChange() { fetchResources() }

async function handleViewYaml(row: any) {
  yamlDialogVisible.value = true; yamlLoading.value = true; yamlContent.value = ''
  try {
    const params: any = { group, version, resource, name: row.name }
    if (row.namespace) params.namespace = row.namespace
    const res: any = await getCustomResourceYaml(params)
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) { ElMessage.error(e?.message || 'Failed to load YAML'); yamlDialogVisible.value = false }
  finally { yamlLoading.value = false }
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(`Delete ${resource} "${row.name}"?`, 'Confirm', { type: 'warning' })
    const params: any = { group, version, resource, name: row.name }
    if (row.namespace) params.namespace = row.namespace
    await deleteCustomResource(params)
    ElMessage.success('Deleted'); fetchResources()
  } catch { /* cancelled */ }
}

onMounted(() => { fetchNamespaces(); fetchResources() })
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h2 style="margin: 0;">{{ resource }} <span style="color: #909399; font-size: 14px;">({{ group }}/{{ version }})</span></h2>
      <el-button @click="router.push('/crd')">Back to CRDs</el-button>
    </div>
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <el-input v-model="searchName" placeholder="Search by name" style="width: 220px;" clearable>
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select v-if="scope === 'Namespaced'" v-model="selectedNamespace" placeholder="All Namespaces" clearable style="width: 180px;" @change="handleNamespaceChange">
          <el-option v-for="ns in namespaceList" :key="ns" :label="ns" :value="ns" />
        </el-select>
        <el-button type="primary" @click="fetchResources"><el-icon><Refresh /></el-icon> Refresh</el-button>
      </div>
    </el-card>
    <el-card shadow="never" class="table-card">
      <el-table :data="filteredList" v-loading="loading" stripe>
        <el-table-column prop="name" label="Name" min-width="250" show-overflow-tooltip />
        <el-table-column v-if="scope === 'Namespaced'" prop="namespace" label="Namespace" width="140" />
        <el-table-column prop="age" label="Age" width="180" />
        <el-table-column label="Actions" width="160" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">Delete</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <el-dialog v-model="yamlDialogVisible" title="Custom Resource YAML" width="70%" top="5vh" destroy-on-close>
      <div v-loading="yamlLoading"><YamlEditor v-model="yamlContent" height="500px" read-only /></div>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; gap: 12px; align-items: center; flex-wrap: wrap; }
.table-card { border-radius: 8px; }
</style>
