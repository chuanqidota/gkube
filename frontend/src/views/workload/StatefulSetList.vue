<script setup lang="ts">
import { Plus, Delete } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ref } from 'vue'
import {
  getStatefulSetList,
  getStatefulSetYaml,
  updateStatefulSetYaml,
  deleteStatefulSet,
  transformStatefulSets,
  scaleStatefulSet,
  restartStatefulSet,
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
  resourceName: 'StatefulSet',
  fetchList: getStatefulSetList,
  transform: transformStatefulSets,
  getYaml: getStatefulSetYaml,
  updateYaml: updateStatefulSetYaml,
  deleteResource: deleteStatefulSet,
  detailRoute: '/workloads/statefulsets',
  createRoute: '/workloads/statefulsets/create',
  paginated: true,
  pageSize: 50,
  autoRefreshInterval: 30000,
})

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchResources)

// ---- Quick Actions ----
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
    await scaleStatefulSet({ ...scaleTarget.value, replicas: scaleReplicas.value })
    ElMessage.success(`已将 ${scaleTarget.value.name} 扩缩容至 ${scaleReplicas.value} 副本`)
    scaleDialogVisible.value = false
    fetchResources()
  } catch (e: any) {
    ElMessage.error(e?.message || '扩缩容失败')
  } finally {
    scaleLoading.value = false
  }
}

async function handleQuickRestart(row: any) {
  try {
    await ElMessageBox.confirm(`确定要重启 StatefulSet "${row.name}" 吗？`, '确认重启', { type: 'warning' })
    await restartStatefulSet({ namespace: row.namespace, name: row.name })
    ElMessage.success(`${row.name} 重启成功`)
    fetchResources()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '重启失败')
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
        <el-button type="success" @click="$router.push('/workloads/statefulsets/create')">
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
        <el-table-column prop="serviceName" label="服务" width="160" show-overflow-tooltip />
        <el-table-column prop="updateStrategy" label="更新策略" width="140" />
        <el-table-column prop="age" label="Age" width="120" />
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
              <el-button size="small" type="primary" @click="handleQuickScale(row)">扩缩容</el-button>
              <el-button size="small" type="warning" @click="handleQuickRestart(row)">重启</el-button>
              <el-button size="small" type="danger" plain @click="handleDelete(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无有状态负载">
            <el-button type="primary" @click="$router.push('/workloads/statefulsets/create')">创建 StatefulSet</el-button>
          </el-empty>
        </template>
      </el-table>

      <!-- Load More Button -->
      <div v-if="hasMore" class="load-more">
        <el-button @click="fetchNextPage" :loading="loading" link type="primary">
          加载更多...
        </el-button>
      </div>
    </el-card>

    <!-- YAML Drawer -->
    <el-drawer v-model="yamlDialogVisible" title="StatefulSet YAML" size="85%" direction="rtl" class="yaml-drawer"
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
      </div>
      <template #footer>
        <el-button @click="scaleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="scaleLoading" @click="handleScaleConfirm">确认</el-button>
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
