<script setup lang="ts">
import { computed, ref } from 'vue'
import { Plus, Delete } from '@element-plus/icons-vue'
import {
  getVpaList,
  getVpaYaml,
  updateVpa,
  deleteVpa,
} from '@/api/resource'
import { useResourceList } from '@/composables/useResourceList'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'
import YamlEditor from '@/components/YamlEditor.vue'

const loadError = ref('')

async function fetchVpaList(params?: { namespace?: string }) {
  try {
    const res = await getVpaList(params)
    loadError.value = ''
    return res
  } catch (e: any) {
    loadError.value = e?.message || '当前集群未安装 VPA CRD 或不支持 autoscaling.k8s.io/v1 VerticalPodAutoscaler'
    throw e
  }
}

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
  totalCount,
  fetchResources,
  handleNamespaceChange,
  handleSelectionChange,
  handleViewYaml,
  handleSaveYaml,
  handleCancelYaml,
  handleDetail,
  handleDelete,
  handleBatchDelete,
} = useResourceList({
  resourceName: 'VPA',
  fetchList: fetchVpaList,
  getYaml: getVpaYaml,
  updateYaml: updateVpa,
  deleteResource: deleteVpa,
  detailRoute: '/autoscaling/vpa',
  createRoute: '/autoscaling/vpa/create',
  autoRefreshInterval: 30000,
})

const { isRunning, countdown, currentInterval, availableIntervals, toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(fetchResources)

const warningTitle = computed(() => (
  loadError.value || '当前集群未安装 VPA CRD 或不支持 autoscaling.k8s.io/v1 VerticalPodAutoscaler'
))

function conditionStatus(row: any) {
  const conditions = row.conditions || []
  const recommendation = conditions.find((c: any) => c.type === 'RecommendationProvided')
  if (!recommendation) return { text: 'Pending', type: 'info' }
  if (recommendation.status === 'True') return { text: 'Recommended', type: 'success' }
  return { text: recommendation.reason || 'Warning', type: 'warning' }
}
</script>

<template>
  <div class="page-container">
    <el-alert
      v-if="loadError"
      :title="warningTitle"
      type="warning"
      show-icon
      :closable="false"
      class="vpa-warning"
    />

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
        <el-button type="success" @click="$router.push('/autoscaling/vpa/create')">
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
        <el-table-column label="伸缩目标" min-width="180">
          <template #default="{ row }">
            {{ row.target_kind || '-' }}/{{ row.target || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="更新模式" width="120">
          <template #default="{ row }">
            <el-tag size="small" :type="row.update_mode === 'Off' ? 'info' : 'warning'">
              {{ row.update_mode || 'Auto' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="推荐状态" width="150">
          <template #default="{ row }">
            <el-tag size="small" :type="conditionStatus(row).type">
              {{ conditionStatus(row).text }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="推荐容器" width="110">
          <template #default="{ row }">{{ row.recommendation_count || 0 }}</template>
        </el-table-column>
        <el-table-column prop="age" label="Age" width="150" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
              <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-drawer v-model="yamlDialogVisible" title="VPA YAML" size="85%" direction="rtl" class="yaml-drawer"
      :body-style="{ padding: '0', height: '100%' }">
      <div v-loading="yamlLoading" style="height: calc(100vh - 60px);">
        <YamlEditor v-model="yamlContent" height="100%" auto-format show-save-buttons :saving="yamlSaving" @save="handleSaveYaml" @cancel="handleCancelYaml" />
      </div>
    </el-drawer>
  </div>
</template>

<style scoped>
.page-container {
  padding: 20px;
}

.vpa-warning {
  margin-bottom: 16px;
}

.table-card {
  border-radius: 8px;
}

.action-buttons {
  display: flex;
  gap: 8px;
}
</style>
