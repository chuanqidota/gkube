<script setup lang="ts">
import { ref } from 'vue'
import { Plus, Delete } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getDeploymentList,
  getDeploymentYaml,
  updateDeploymentYaml,
  deleteDeployment,
  transformDeployments,
  scaleDeployment,
  restartDeployment,
  updateDeploymentImage,
} from '@/api/resource'
import { useResourceList } from '@/composables/useResourceList'
import YamlEditor from '@/components/YamlEditor.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const {
  loading,
  filteredList,
  selectedNamespace,
  searchName,
  onSearchInput,
  selectedRows,
  namespaceList,
  yamlDialogVisible,
  yamlContent,
  yamlLoading,
  yamlSaving,
  hasMore,
  totalCount,
  fetchResources,
  fetchNextPage,
  handleNamespaceChange,
  handleSelectionChange,
  handleViewYaml,
  handleSaveYaml,
  handleCancelYaml,
  handleDetail,
  handleDelete,
  handleBatchDelete,
} = useResourceList({
  resourceName: 'Deployment',
  fetchList: getDeploymentList,
  transform: transformDeployments,
  getYaml: getDeploymentYaml,
  updateYaml: updateDeploymentYaml,
  deleteResource: deleteDeployment,
  detailRoute: '/workloads/deployments',
  createRoute: '/workloads/deployments/create',
  paginated: true,
  pageSize: 50,
  autoRefreshInterval: 30000,
})

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchResources)

// ---- Quick Actions ----

// Scale dialog
const scaleDialogVisible = ref(false)
const scaleTarget = ref<{ namespace: string; name: string } | null>(null)
const scaleReplicas = ref<number>(1)
const scaleLoading = ref(false)

function handleQuickScale(row: any) {
  scaleTarget.value = { namespace: row.namespace, name: row.name }
  scaleReplicas.value = row.ready_replicas ?? 1
  scaleDialogVisible.value = true
}

async function handleScaleConfirm() {
  if (!scaleTarget.value) return
  scaleLoading.value = true
  try {
    await scaleDeployment({ ...scaleTarget.value, replicas: scaleReplicas.value })
    ElMessage.success(`已将 ${scaleTarget.value.name} 扩缩容至 ${scaleReplicas.value} 副本`)
    scaleDialogVisible.value = false
    fetchResources()
  } catch (e: any) {
    ElMessage.error(e?.message || '扩缩容失败')
  } finally {
    scaleLoading.value = false
  }
}

// Restart
async function handleQuickRestart(row: any) {
  try {
    await ElMessageBox.confirm(`确定要重启 Deployment "${row.name}" 吗？这将触发滚动更新。`, '确认重启', { type: 'warning' })
    await restartDeployment({ namespace: row.namespace, name: row.name })
    ElMessage.success(`${row.name} 重启成功`)
    fetchResources()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '重启失败')
  }
}

// Update image dialog
const imageDialogVisible = ref(false)
const imageTarget = ref<{ namespace: string; name: string } | null>(null)
const imageForm = ref({ containerName: '', image: '' })
const imageContainers = ref<{ name: string; image: string }[]>([])
const imageLoading = ref(false)

async function handleQuickUpdateImage(row: any) {
  imageTarget.value = { namespace: row.namespace, name: row.name }
  imageForm.value = { containerName: '', image: '' }
  imageContainers.value = []
  imageDialogVisible.value = true
  // Fetch detail to get container list
  try {
    const { getDeploymentDetail } = await import('@/api/resource')
    const res: any = await getDeploymentDetail({ namespace: row.namespace, name: row.name })
    const containers = res.data?.spec?.template?.spec?.containers || []
    imageContainers.value = containers.map((c: any) => ({ name: c.name, image: c.image || '' }))
    if (imageContainers.value.length > 0) {
      imageForm.value.containerName = imageContainers.value[0].name
      imageForm.value.image = imageContainers.value[0].image
    }
  } catch { /* ignore */ }
}

async function handleImageConfirm() {
  if (!imageTarget.value || !imageForm.value.containerName || !imageForm.value.image) {
    ElMessage.warning('请填写容器名称和镜像')
    return
  }
  imageLoading.value = true
  try {
    await updateDeploymentImage({ ...imageTarget.value, ...imageForm.value })
    ElMessage.success('镜像更新成功')
    imageDialogVisible.value = false
    fetchResources()
  } catch (e: any) {
    ElMessage.error(e?.message || '镜像更新失败')
  } finally {
    imageLoading.value = false
  }
}
</script>

<template>
  <div class="page-container">
    <ResourceListToolbar
      :search-value="searchName"
      v-model:namespace-value="selectedNamespace"
      :namespace-list="namespaceList"
      :total-count="totalCount"
      :selected-count="selectedRows.length"
      @search-input="onSearchInput"
      @namespace-change="handleNamespaceChange"
    >
      <template #actions>
        <el-button type="success" @click="$router.push('/workloads/deployments/create')">
          <el-icon><Plus /></el-icon> 创建
        </el-button>
        <el-button type="danger" :disabled="!selectedRows.length" @click="handleBatchDelete">
          <el-icon><Delete /></el-icon> 删除 ({{ selectedRows.length }})
        </el-button>
      </template>
      <template #extra>
        <AutoRefreshToolbar
          :is-running="isRunning"
          :countdown="countdown"
          :current-interval="currentInterval"
          :available-intervals="availableIntervals"
          :loading="loading"
          @refresh="manualRefresh()"
          @toggle="toggle()"
          @interval-change="setIntervalOption"
        />
      </template>
    </ResourceListToolbar>

    <el-card shadow="never" class="table-card">
      <el-table
        :data="filteredList"
        v-loading="loading"
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="45" />
        <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <el-button link type="primary" @click="handleDetail(row)">{{ row.name }}</el-button>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="140" />
        <el-table-column prop="ready" label="就绪" width="100" />
        <el-table-column prop="up_to_date" label="最新" width="110" />
        <el-table-column prop="available" label="可用" width="110" />
        <el-table-column prop="age" label="Age" width="120" />
        <el-table-column label="操作" width="340" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
              <el-button size="small" type="primary" @click="handleQuickScale(row)">扩缩容</el-button>
              <el-button size="small" type="warning" @click="handleQuickRestart(row)">重启</el-button>
              <el-button size="small" type="primary" @click="handleQuickUpdateImage(row)">更新镜像</el-button>
              <el-button size="small" type="danger" plain @click="handleDelete(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- Load More Button -->
      <div v-if="hasMore" class="load-more">
        <el-button @click="fetchNextPage" :loading="loading" link type="primary">
          加载更多...
        </el-button>
      </div>
    </el-card>

    <!-- YAML Drawer -->
    <el-drawer v-model="yamlDialogVisible" title="Deployment YAML" size="85%" direction="rtl" class="yaml-drawer"
      :body-style="{ padding: '0', height: '100%' }">
      <div v-loading="yamlLoading" style="height: calc(100vh - 52px);">
        <YamlEditor v-model="yamlContent" height="100%" auto-format show-save-buttons :saving="yamlSaving" @save="handleSaveYaml" @cancel="handleCancelYaml" />
      </div>
    </el-drawer>

    <!-- Scale Dialog -->
    <el-dialog v-model="scaleDialogVisible" title="扩缩容" width="420px" destroy-on-close>
      <div>
        <p style="margin-bottom: 16px;">调整 <strong>{{ scaleTarget?.name }}</strong> 副本数</p>
        <el-form-item label="目标副本数">
          <el-input-number v-model="scaleReplicas" :min="0" :max="100" style="width: 200px;" />
        </el-form-item>
        <el-alert v-if="scaleReplicas === 0" title="设为 0 将停止所有 Pod。" type="warning" :closable="false" show-icon style="margin-top: 8px;" />
      </div>
      <template #footer>
        <el-button @click="scaleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="scaleLoading" @click="handleScaleConfirm">确认</el-button>
      </template>
    </el-dialog>

    <!-- Update Image Dialog -->
    <el-dialog v-model="imageDialogVisible" title="更新镜像" width="520px" destroy-on-close>
      <div>
        <p style="margin-bottom: 16px;">更新 <strong>{{ imageTarget?.name }}</strong> 的容器镜像</p>
        <el-form label-width="80px">
          <el-form-item label="容器">
            <el-select v-model="imageForm.containerName" style="width: 100%;" @change="() => { const c = imageContainers.find(c => c.name === imageForm.containerName); if (c) imageForm.image = c.image }">
              <el-option v-for="c in imageContainers" :key="c.name" :label="c.name" :value="c.name" />
            </el-select>
          </el-form-item>
          <el-form-item label="镜像">
            <el-input v-model="imageForm.image" placeholder="例如: nginx:1.26" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="imageDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="imageLoading" @click="handleImageConfirm">确认更新</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container {
  padding: 20px;
}
.table-card {
  border-radius: 8px;
}
.load-more {
  display: flex;
  justify-content: center;
  padding: 12px 0;
  border-top: 1px solid var(--el-border-color-lighter);
}
.action-buttons {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: 4px;
}
.action-buttons .el-button + .el-button {
  margin-left: 0;
}
</style>

<style>
.yaml-drawer .el-drawer__header {
  padding: 6px 16px;
  margin-bottom: 0;
  min-height: auto;
}
</style>
