<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Delete, Search } from '@element-plus/icons-vue'
import { getNodeList, getNodeYaml, cordonNode, taintNode } from '@/api/resource'
import YamlEditor from '@/components/YamlEditor.vue'

const router = useRouter()
const loading = ref(false)
const nodeList = ref<any[]>([])
const searchName = ref('')
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)
const taintDialogVisible = ref(false)
const taintTarget = ref<any>(null)
const taints = ref<any[]>([])

const filteredList = computed(() => {
  if (!searchName.value) return nodeList.value
  const keyword = searchName.value.toLowerCase()
  return nodeList.value.filter((n) => n.name?.toLowerCase().includes(keyword))
})

async function fetchNodes() {
  loading.value = true
  try {
    const res: any = await getNodeList()
    nodeList.value = res.data || []
  } catch (e: any) {
    ElMessage.error(e?.message || 'Failed to load nodes')
  } finally { loading.value = false }
}

function statusType(node: any) {
  if (node.status === 'Ready') return 'success'
  if (node.status === 'NotReady') return 'danger'
  return 'warning'
}

async function handleViewYaml(row: any) {
  yamlDialogVisible.value = true; yamlLoading.value = true; yamlContent.value = ''
  try {
    const res: any = await getNodeYaml({ name: row.name })
    yamlContent.value = res.data?.yaml || res.data || ''
  } catch (e: any) { ElMessage.error(e?.message || 'Failed to load YAML'); yamlDialogVisible.value = false }
  finally { yamlLoading.value = false }
}

function handleDetail(row: any) { router.push(`/nodes/${row.name}`) }

async function handleCordon(row: any) {
  const isCordon = row.unschedulable || row.cordon
  const action = isCordon ? 'uncordon' : 'cordon'
  try {
    await ElMessageBox.confirm(`${action.charAt(0).toUpperCase() + action.slice(1)} node "${row.name}"?`, 'Confirm', { type: 'warning' })
    await cordonNode({ name: row.name, cordon: !isCordon })
    ElMessage.success(`Node ${action}ed`); fetchNodes()
  } catch { /* cancelled */ }
}

function handleTaints(row: any) {
  taintTarget.value = row
  taints.value = (row.taints || []).map((t: any) => ({ ...t }))
  if (taints.value.length === 0) taints.value = [{ key: '', value: '', effect: 'NoSchedule' }]
  taintDialogVisible.value = true
}

function addTaint() { taints.value.push({ key: '', value: '', effect: 'NoSchedule' }) }
function removeTaint(index: number) { taints.value.splice(index, 1) }

async function handleSaveTaints() {
  try {
    await taintNode({ name: taintTarget.value.name, taints: taints.value.filter(t => t.key) })
    ElMessage.success('Taints updated'); taintDialogVisible.value = false; fetchNodes()
  } catch (e: any) { ElMessage.error(e?.message || 'Failed to update taints') }
}

onMounted(fetchNodes)
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <el-input v-model="searchName" placeholder="Search by name" style="width: 220px;" clearable>
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-button type="primary" @click="fetchNodes"><el-icon><Refresh /></el-icon> Refresh</el-button>
      </div>
    </el-card>
    <el-card shadow="never" class="table-card">
      <el-table :data="filteredList" v-loading="loading" stripe>
        <el-table-column prop="name" label="Name" min-width="200" show-overflow-tooltip>
          <template #default="{ row }"><el-button link type="primary" @click="handleDetail(row)">{{ row.name }}</el-button></template>
        </el-table-column>
        <el-table-column label="Status" width="100">
          <template #default="{ row }"><el-tag :type="statusType(row)" size="small">{{ row.status || 'Unknown' }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="roles" label="Roles" min-width="120" />
        <el-table-column prop="version" label="Version" width="130" />
        <el-table-column prop="internal_ip" label="Internal IP" width="140" />
        <el-table-column prop="age" label="Age" width="120" />
        <el-table-column label="Actions" width="280" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
            <el-button size="small" :type="row.unschedulable || row.cordon ? 'success' : 'warning'" @click="handleCordon(row)">
              {{ row.unschedulable || row.cordon ? 'Uncordon' : 'Cordon' }}
            </el-button>
            <el-button size="small" type="info" @click="handleTaints(row)">Taints</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <el-dialog v-model="yamlDialogVisible" title="Node YAML" width="70%" top="5vh" destroy-on-close>
      <div v-loading="yamlLoading"><YamlEditor v-model="yamlContent" height="500px" read-only auto-format /></div>
    </el-dialog>
    <el-dialog v-model="taintDialogVisible" title="Manage Taints" width="600px">
      <div v-for="(taint, index) in taints" :key="index" style="display: flex; gap: 8px; margin-bottom: 12px; align-items: center;">
        <el-input v-model="taint.key" placeholder="Key" style="flex: 2;" />
        <el-input v-model="taint.value" placeholder="Value" style="flex: 1;" />
        <el-select v-model="taint.effect" style="flex: 1.5;">
          <el-option label="NoSchedule" value="NoSchedule" />
          <el-option label="PreferNoSchedule" value="PreferNoSchedule" />
          <el-option label="NoExecute" value="NoExecute" />
        </el-select>
        <el-button type="danger" circle size="small" @click="removeTaint(index)"><el-icon><Delete /></el-icon></el-button>
      </div>
      <el-button @click="addTaint" style="margin-top: 8px;">Add Taint</el-button>
      <template #footer>
        <el-button @click="taintDialogVisible = false">Cancel</el-button>
        <el-button type="primary" @click="handleSaveTaints">Save</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; gap: 12px; align-items: center; flex-wrap: wrap; }
.table-card { border-radius: 8px; }
</style>
