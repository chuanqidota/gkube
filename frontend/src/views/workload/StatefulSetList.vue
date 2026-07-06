<script setup lang="ts">
import { ref } from 'vue'
import { Plus, Delete, Search } from '@element-plus/icons-vue'
import {
  getStatefulSetList,
  getStatefulSetYaml,
  updateStatefulSetYaml,
  deleteStatefulSet,
  transformStatefulSets,
} from '@/api/resource'
import { useResourceList } from '@/composables/useResourceList'
import YamlEditor from '@/components/YamlEditor.vue'
import AutoRefreshToolbar from '@/components/AutoRefreshToolbar.vue'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

// @ts-ignore -- used as template ref for YamlEditor component
const yamlEditorRef = ref<InstanceType<typeof YamlEditor>>()

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
  hasMore,
  totalCount,
  fetchResources,
  fetchNextPage,
  handleNamespaceChange,
  handleSelectionChange,
  handleViewYaml,
  handleSaveYaml,
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
</script>

<template>
  <div class="page-container">
    <el-card shadow="never" class="filter-card">
      <div class="filter-bar">
        <el-input
          :model-value="searchName"
          @input="onSearchInput"
          placeholder="搜索名称"
          style="width: 220px;"
          clearable
        >
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-select
          v-model="selectedNamespace"
          placeholder="所有命名空间"
          clearable
          style="width: 180px;"
          @change="handleNamespaceChange"
        >
          <el-option v-for="ns in namespaceList" :key="ns" :label="ns" :value="ns" />
        </el-select>
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
        <el-button type="success" @click="$router.push('/workloads/statefulsets/create')">
          <el-icon><Plus /></el-icon> 创建
        </el-button>
        <el-button type="danger" :disabled="!selectedRows.length" @click="handleBatchDelete">
          <el-icon><Delete /></el-icon> 删除 ({{ selectedRows.length }})
        </el-button>
        <span class="total-count" v-if="totalCount">总计: {{ totalCount }}</span>
      </div>
    </el-card>

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
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewYaml(row)">YAML</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
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

    <!-- YAML Dialog -->
    <el-dialog v-model="yamlDialogVisible" title="StatefulSet YAML" width="70%" top="5vh" destroy-on-close>
      <div v-loading="yamlLoading">
        <YamlEditor
          ref="yamlEditorRef"
          v-model="yamlContent"
          height="600px"
          :read-only="false"
          :saveable="true"
          auto-format
          @save="handleSaveYaml"
        />
      </div>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container {
  padding: 20px;
}
.filter-card {
  margin-bottom: 16px;
}
.filter-bar {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}
.table-card {
  border-radius: 8px;
}
.total-count {
  color: var(--el-text-color-secondary);
  font-size: 13px;
  margin-left: auto;
}
.load-more {
  display: flex;
  justify-content: center;
  padding: 12px 0;
  border-top: 1px solid var(--el-border-color-lighter);
}
</style>
