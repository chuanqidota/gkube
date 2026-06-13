<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Refresh, Search } from '@element-plus/icons-vue'
import { getCrdList, getCrdYaml } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const router = useRouter()
const loading = ref(false)
const crdList = ref<any[]>([])
const searchName = ref('')
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)

const filteredList = computed(() => {
  if (!searchName.value) return crdList.value
  const keyword = searchName.value.toLowerCase()
  return crdList.value.filter((d) => d.name?.toLowerCase().includes(keyword) || d.kind?.toLowerCase().includes(keyword))
})

async function fetchCrds() {
  loading.value = true
  try {
    const res: any = await getCrdList()
    crdList.value = res.data || []
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load CRDs')
  } finally { loading.value = false }
}

async function handleViewYaml(row: any) {
  yamlDialogVisible.value = true; yamlLoading.value = true; yamlContent.value = ''
  try {
    const res: any = await getCrdYaml({ name: row.name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) { ElMessage.error(e?.message || 'Failed to load YAML'); yamlDialogVisible.value = false }
  finally { yamlLoading.value = false }
}

function handleBrowse(row: any) {
  const group = row.group
  const version = row.versions?.[0] || 'v1'
  const resource = row.plural
  router.push(`/crd/resources?group=${group}&version=${version}&resource=${resource}&scope=${row.scope}`)
}

onMounted(fetchCrds)
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <el-input v-model="searchName" placeholder="Search by name or kind" style="width: 280px;" clearable>
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-button type="primary" @click="fetchCrds"><el-icon><Refresh /></el-icon> Refresh</el-button>
      </div>
    </el-card>
    <el-card shadow="never" class="table-card">
      <el-table :data="filteredList" v-loading="loading" stripe>
        <el-table-column prop="kind" label="Kind" min-width="180" show-overflow-tooltip>
          <template #default="{ row }"><el-button link type="primary" @click="handleBrowse(row)">{{ row.kind }}</el-button></template>
        </el-table-column>
        <el-table-column prop="name" label="Name" min-width="280" show-overflow-tooltip />
        <el-table-column prop="group" label="Group" min-width="180" show-overflow-tooltip />
        <el-table-column label="Versions" width="140">
          <template #default="{ row }"><el-tag v-for="v in (row.versions || [])" :key="v" size="small" style="margin-right: 4px;">{{ v }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="scope" label="Scope" width="120">
          <template #default="{ row }"><el-tag :type="row.scope === 'Namespaced' ? 'info' : 'warning'" size="small">{{ row.scope }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="age" label="Age" width="180" />
        <el-table-column label="Actions" width="120" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <el-dialog v-model="yamlDialogVisible" title="CRD YAML" width="70%" top="5vh" destroy-on-close>
      <div v-loading="yamlLoading"><YamlEditor v-model="yamlContent" height="500px" read-only /></div>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; gap: 12px; align-items: center; flex-wrap: wrap; }
.table-card { border-radius: 8px; }
</style>
