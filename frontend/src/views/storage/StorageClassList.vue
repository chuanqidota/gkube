<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Refresh, Search } from '@element-plus/icons-vue'
import { getStorageClassList, getStorageClassYaml } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const router = useRouter()
const loading = ref(false)
const storageClassList = ref<any[]>([])
const searchName = ref('')
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)

const filteredList = computed(() => {
  if (!searchName.value) return storageClassList.value
  const keyword = searchName.value.toLowerCase()
  return storageClassList.value.filter((d) => d.name?.toLowerCase().includes(keyword))
})

async function fetchStorageClasses() {
  loading.value = true
  try {
    const res: any = await getStorageClassList()
    storageClassList.value = res.data || []
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load StorageClasses')
  } finally { loading.value = false }
}

async function handleViewYaml(row: any) {
  yamlDialogVisible.value = true; yamlLoading.value = true; yamlContent.value = ''
  try {
    const res: any = await getStorageClassYaml({ name: row.name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) { ElMessage.error(e?.message || 'Failed to load YAML'); yamlDialogVisible.value = false }
  finally { yamlLoading.value = false }
}

function handleDetail(row: any) { router.push(`/storage/storageclasses/${row.name}`) }

onMounted(fetchStorageClasses)
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <el-input v-model="searchName" placeholder="Search by name" style="width: 220px;" clearable>
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-button type="primary" @click="fetchStorageClasses"><el-icon><Refresh /></el-icon> Refresh</el-button>
      </div>
    </el-card>
    <el-card shadow="never" class="table-card">
      <el-table :data="filteredList" v-loading="loading" stripe>
        <el-table-column prop="name" label="Name" min-width="180" show-overflow-tooltip>
          <template #default="{ row }"><el-button link type="primary" @click="handleDetail(row)">{{ row.name }}</el-button></template>
        </el-table-column>
        <el-table-column prop="provisioner" label="Provisioner" min-width="200" show-overflow-tooltip />
        <el-table-column prop="reclaim_policy" label="Reclaim Policy" width="140" />
        <el-table-column prop="volume_binding_mode" label="Volume Binding Mode" min-width="180" show-overflow-tooltip />
        <el-table-column prop="age" label="Age" width="120" />
        <el-table-column label="Actions" width="120" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <el-dialog v-model="yamlDialogVisible" title="StorageClass YAML" width="70%" top="5vh" destroy-on-close>
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
