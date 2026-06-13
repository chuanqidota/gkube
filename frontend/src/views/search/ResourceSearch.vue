<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Search, Refresh, View } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'
import request from '@/api/request'

const { t } = useI18n()
const router = useRouter()
const loading = ref(false)
const searchQuery = ref('')
const selectedType = ref('')
const selectedNamespace = ref('')
const searchResults = ref<any[]>([])
const namespaces = ref<string[]>([])

const resourceTypes = [
  { value: '', label: '所有资源' },
  { value: 'pods', label: 'Pods' },
  { value: 'deployments', label: 'Deployments' },
  { value: 'statefulsets', label: 'StatefulSets' },
  { value: 'daemonsets', label: 'DaemonSets' },
  { value: 'services', label: 'Services' },
  { value: 'ingresses', label: 'Ingresses' },
  { value: 'configmaps', label: 'ConfigMaps' },
  { value: 'secrets', label: 'Secrets' },
  { value: 'persistentvolumeclaims', label: 'PVCs' },
  { value: 'persistentvolumes', label: 'PVs' },
  { value: 'storageclasses', label: 'StorageClasses' },
  { value: 'namespaces', label: 'Namespaces' },
  { value: 'nodes', label: 'Nodes' },
  { value: 'serviceaccounts', label: 'ServiceAccounts' },
  { value: 'roles', label: 'Roles' },
  { value: 'clusterroles', label: 'ClusterRoles' },
  { value: 'rolebindings', label: 'RoleBindings' },
  { value: 'clusterrolebindings', label: 'ClusterRoleBindings' },
  { value: 'horizontalpodautoscalers', label: 'HPAs' },
  { value: 'networkpolicies', label: 'NetworkPolicies' },
  { value: 'poddisruptionbudgets', label: 'PDBs' },
  { value: 'resourcequotas', label: 'ResourceQuotas' },
  { value: 'limitranges', label: 'LimitRanges' },
]

async function fetchNamespaces() {
  try {
    const res: any = await request.get('/k8s/namespace/list')
    namespaces.value = res.data?.map((ns: any) => ns.name) || []
  } catch {
    namespaces.value = []
  }
}

async function performSearch() {
  if (!searchQuery.value) {
    ElMessage.warning(t('search.enterKeyword'))
    return
  }

  loading.value = true
  searchResults.value = []

  try {
    const types = selectedType.value ? [selectedType.value] : resourceTypes.map(t => t.value).filter(Boolean)

    const promises = types.map(async (type) => {
      try {
        const params: any = {}
        if (selectedNamespace.value) {
          params.namespace = selectedNamespace.value
        }
        const res: any = await request.get(`/k8s/${type}/list`, { params })
        const items = res.data || []
        return items
          .filter((item: any) => {
            const query = searchQuery.value.toLowerCase()
            return (
              item.name?.toLowerCase().includes(query) ||
              item.metadata?.name?.toLowerCase().includes(query) ||
              item.labels?.app?.toLowerCase().includes(query) ||
              item.annotations?.description?.toLowerCase().includes(query)
            )
          })
          .map((item: any) => ({
            ...item,
            type: type,
            displayName: item.name || item.metadata?.name,
          }))
      } catch {
        return []
      }
    })

    const results = await Promise.all(promises)
    searchResults.value = results.flat()
  } catch (e: any) {
    ElMessage.error(t('search.searchFailed'))
  } finally {
    loading.value = false
  }
}

function getResourceTypeLabel(type: string) {
  const found = resourceTypes.find(t => t.value === type)
  return found ? found.label : type
}

function getResourceStatus(resource: any) {
  return resource.status || resource.phase || '-'
}

function getStatusType(status: string) {
  if (status === 'Running' || status === 'Active' || status === 'Bound') return 'success'
  if (status === 'Pending' || status === 'Terminating') return 'warning'
  if (status === 'Failed' || status === 'Error') return 'danger'
  return 'info'
}

function viewResource(resource: any) {
  const type = resource.type
  const name = resource.displayName
  const namespace = resource.namespace

  if (namespace) {
    router.push(`/workloads/${type}/${namespace}/${name}`)
  } else {
    router.push(`/${type}/${name}`)
  }
}

function formatTime(time: string) {
  if (!time) return '-'
  return new Date(time).toLocaleString()
}

onMounted(fetchNamespaces)
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <h3 style="margin: 0;"><el-icon><Search /></el-icon> {{ t('search.title') }}</h3>
      </div>
    </el-card>

    <el-card shadow="never" style="margin-bottom: 16px;">
      <el-form :inline="true">
        <el-form-item :label="t('search.keyword')">
          <el-input v-model="searchQuery" :placeholder="t('search.searchPlaceholder')" style="width: 300px;" @keyup.enter="performSearch" />
        </el-form-item>
        <el-form-item :label="t('search.resourceType')">
          <el-select v-model="selectedType" :placeholder="t('search.allResources')" clearable style="width: 180px;">
            <el-option v-for="r in resourceTypes" :key="r.value" :label="r.label" :value="r.value" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('search.namespace')">
          <el-select v-model="selectedNamespace" :placeholder="t('search.allNamespaces')" clearable style="width: 150px;">
            <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="performSearch" :loading="loading"><el-icon><Search /></el-icon> {{ t('search.search') }}</el-button>
          <el-button @click="searchResults = []"><el-icon><Refresh /></el-icon> {{ t('common.reset') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" v-if="searchResults.length > 0">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <h4 style="margin: 0;">{{ t('search.searchResults') }}</h4>
          <el-tag>{{ t('search.foundResources', { count: searchResults.length }) }}</el-tag>
        </div>
      </template>
      <el-table :data="searchResults" stripe>
        <el-table-column prop="type" :label="t('common.type')" width="150">
          <template #default="{ row }">
            <el-tag size="small">{{ getResourceTypeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="displayName" :label="t('common.name')" min-width="200" show-overflow-tooltip />
        <el-table-column prop="namespace" :label="t('common.namespace_label')" width="120" />
        <el-table-column :label="t('common.status')" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(getResourceStatus(row))" size="small">
              {{ getResourceStatus(row) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="creationTimestamp" :label="t('search.createdAt')" width="180">
          <template #default="{ row }">{{ formatTime(row.creationTimestamp) }}</template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="100">
          <template #default="{ row }">
            <el-button size="small" @click="viewResource(row)"><el-icon><View /></el-icon> {{ t('common.view') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card shadow="never" v-else-if="searchQuery && !loading">
      <el-empty :description="t('search.noResults')" />
    </el-card>
  </div>
</template>

<style scoped>
.page-container { padding: 20px; }
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; justify-content: space-between; align-items: center; }
</style>
