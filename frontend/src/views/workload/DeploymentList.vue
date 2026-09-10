<script setup lang="ts">
import { ref } from 'vue'
import { Plus, Delete } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  getDeploymentList,
  getDeploymentYaml,
  updateDeploymentYaml,
  deleteDeployment,
  transformDeployments,
  scaleDeployment,
  restartDeployment,
  updateDeploymentImage,
  getDeploymentDetail,
} from '@/api/resource'
import { useResourceList } from '@/composables/useResourceList'
import { useClusterStore } from '@/stores/cluster'
import YamlEditor from '@/components/YamlEditor.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const clusterStore = useClusterStore()
const { t } = useI18n()

const {
  loading,
  filteredList,
  selectedNamespace,
  searchName,
  onSearchInput,
  selectedRows,
  labelConditions,
  onLabelConditionsChange,
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
  scaleReplicas.value = row.replicas ?? row.ready_replicas ?? 1
  scaleDialogVisible.value = true
}

async function handleScaleConfirm() {
  if (!scaleTarget.value) return
  scaleLoading.value = true
  try {
    await scaleDeployment({ ...scaleTarget.value, replicas: scaleReplicas.value })
    ElMessage.success(t('workload.scaledToReplicas', { name: scaleTarget.value.name, n: scaleReplicas.value }))
    scaleDialogVisible.value = false
    fetchResources()
  } catch (e: any) {
    ElMessage.error(e?.message || t('common.scaleFailed'))
  } finally {
    scaleLoading.value = false
  }
}

// Restart
async function handleQuickRestart(row: any) {
  try {
    await ElMessageBox.confirm(t('workload.restartConfirm', { type: 'Deployment', name: row.name }), t('common.confirmAction'), { type: 'warning' })
    await restartDeployment({ namespace: row.namespace, name: row.name })
    ElMessage.success(t('workload.restartSuccessMsg', { name: row.name }))
    fetchResources()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || t('workload.restartFailedMsg'))
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
    const res: any = await getDeploymentDetail({ namespace: row.namespace, name: row.name })
    const containers = res.data?.spec?.template?.spec?.containers || []
    imageContainers.value = containers.map((c: any) => ({ name: c.name, image: c.image || '' }))
    if (imageContainers.value.length > 0) {
      imageForm.value.containerName = imageContainers.value[0].name
      imageForm.value.image = imageContainers.value[0].image
    }
  } catch (e: any) {
    ElMessage.error(t('workload.fetchContainerFailed') + ': ' + (e?.message || 'unknown error'))
  }
}

async function handleImageConfirm() {
  if (!imageTarget.value || !imageForm.value.containerName || !imageForm.value.image) {
    ElMessage.warning(t('workload.selectContainerAndImage'))
    return
  }
  imageLoading.value = true
  try {
    await updateDeploymentImage({ ...imageTarget.value, ...imageForm.value })
    ElMessage.success(t('workload.imageUpdateSuccess'))
    imageDialogVisible.value = false
    fetchResources()
  } catch (e: any) {
    ElMessage.error(e?.message || t('workload.imageUpdateFailed'))
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
      :cluster-name="clusterStore.clusterName"
      resource-type="deployment"
      :label-conditions="labelConditions"
      @search-input="onSearchInput"
      @namespace-change="handleNamespaceChange"
      @label-selector-change="onLabelConditionsChange"
    >
      <template #actions>
        <el-button type="success" @click="$router.push('/workloads/deployments/create')">
          <el-icon><Plus /></el-icon> {{ t('common.create') }}
        </el-button>
        <el-button type="danger" :disabled="!selectedRows.length" @click="handleBatchDelete">
          <el-icon><Delete /></el-icon> {{ t('common.delete') }} ({{ selectedRows.length }})
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
        <el-table-column prop="name" :label="t('common.name')" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <el-button link type="primary" @click="handleDetail(row)">{{ row.name }}</el-button>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" :label="t('common.namespace_label')" width="140" />
        <el-table-column prop="ready" :label="t('workload.ready')" width="100" />
        <el-table-column prop="up_to_date" :label="t('workload.upToDate')" width="110" />
        <el-table-column prop="available" :label="t('workload.available')" width="110" />
        <el-table-column prop="age" label="Age" width="120" />
        <el-table-column :label="t('common.actions')" width="340" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
              <el-button size="small" type="primary" @click="handleQuickScale(row)">{{ t('workload.scale') }}</el-button>
              <el-button size="small" type="warning" @click="handleQuickRestart(row)">{{ t('workload.restart') }}</el-button>
              <el-button size="small" type="primary" @click="handleQuickUpdateImage(row)">{{ t('workload.updateImage') }}</el-button>
              <el-button size="small" type="danger" @click="handleDelete(row)">{{ t('common.delete') }}</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- Load More Button -->
      <div v-if="hasMore" class="load-more">
        <el-button @click="fetchNextPage" :loading="loading" link type="primary">
          {{ t('workload.loadMore') }}
        </el-button>
      </div>
    </el-card>

    <!-- YAML Drawer -->
    <el-drawer v-model="yamlDialogVisible" title="Deployment YAML" size="85%" direction="rtl" class="yaml-drawer"
      :body-style="{ padding: '0', height: '100%' }">
      <div v-loading="yamlLoading" style="height: calc(100dvh - 52px);">
        <YamlEditor v-model="yamlContent" height="100%" auto-format show-save-buttons :saving="yamlSaving" @save="handleSaveYaml" @cancel="handleCancelYaml" />
      </div>
    </el-drawer>

    <!-- Scale Dialog -->
    <el-dialog v-model="scaleDialogVisible" :title="t('workload.scale')" width="420px" destroy-on-close>
      <div>
        <p style="margin-bottom: var(--gk-space-4);">{{ t('workload.targetReplicas') }} <strong>{{ scaleTarget?.name }}</strong></p>
        <el-form-item :label="t('workload.targetReplicas')">
          <el-input-number v-model="scaleReplicas" :min="0" :max="10000" style="width: 200px;" />
        </el-form-item>
        <el-alert v-if="scaleReplicas === 0" :title="t('workload.zeroReplicasWarning')" type="warning" :closable="false" show-icon style="margin-top: 8px;" />
      </div>
      <template #footer>
        <el-button @click="scaleDialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="scaleLoading" @click="handleScaleConfirm">{{ t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- Update Image Dialog -->
    <el-dialog v-model="imageDialogVisible" :title="t('workload.updateImage')" width="520px" destroy-on-close>
      <div>
        <p style="margin-bottom: var(--gk-space-4);">{{ t('workload.updateImage') }} <strong>{{ imageTarget?.name }}</strong></p>
        <el-form label-width="80px">
          <el-form-item :label="t('workload.containers')">
            <el-select v-model="imageForm.containerName" style="width: 100%;" @change="() => { const c = imageContainers.find(c => c.name === imageForm.containerName); if (c) imageForm.image = c.image }">
              <el-option v-for="c in imageContainers" :key="c.name" :label="c.name" :value="c.name" />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('workload.image')">
            <el-input v-model="imageForm.image" placeholder="例如: nginx:1.26" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="imageDialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="imageLoading" @click="handleImageConfirm">{{ t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container {
  padding: var(--gk-space-5);
}
.table-card {
  border-radius: var(--gk-radius-md);
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
  gap: var(--gk-space-1);
}
.action-buttons .el-button + .el-button {
  margin-left: 0;
}
</style>
