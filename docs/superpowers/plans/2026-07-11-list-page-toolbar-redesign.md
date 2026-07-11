# 列表页面工具栏按钮重新设计 - 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将所有资源列表页面的工具栏从单行混合布局改为单行左右分区布局，使用连接式按钮组样式，添加竖线分隔符区分筛选区和操作区。

**Architecture:** 创建一个共享的 `ResourceListToolbar` 组件封装新的布局结构（筛选区 + 分隔符 + 操作区 + 分隔符 + 辅助工具区），然后将所有 19 个列表页面从内联 filter-bar 迁移到使用该组件。组件通过 slot 和 props 接收各页面的差异化内容。

**Tech Stack:** Vue 3 + TypeScript + Element Plus

## Global Constraints

- 不修改 `useResourceList` composable、`useAutoRefresh` composable、`AutoRefreshToolbar` 组件
- 不修改表格列定义、YAML 编辑器、路由配置
- 连接式按钮组 CSS 参考 `DeploymentDetail.vue:688-700` 的实现
- 所有页面保持相同的视觉语言，按钮数量可不同

## File Structure

| 文件 | 操作 | 说明 |
|------|------|------|
| `frontend/src/components/ResourceListToolbar.vue` | 创建 | 共享工具栏组件，封装布局结构 |
| `frontend/src/views/workload/DeploymentList.vue` | 修改 | 迁移到新组件（参考实现） |
| `frontend/src/views/workload/StatefulSetList.vue` | 修改 | 迁移到新组件 |
| `frontend/src/views/workload/DaemonSetList.vue` | 修改 | 迁移到新组件 |
| `frontend/src/views/workload/JobList.vue` | 修改 | 迁移到新组件 |
| `frontend/src/views/workload/CronJobList.vue` | 修改 | 迁移到新组件 |
| `frontend/src/views/workload/PodList.vue` | 修改 | 迁移到新组件（无创建按钮） |
| `frontend/src/views/workload/ReplicaSetList.vue` | 修改 | 迁移到新组件（无创建按钮、无总计数） |
| `frontend/src/views/network/ServiceList.vue` | 修改 | 迁移到新组件 |
| `frontend/src/views/network/IngressList.vue` | 修改 | 迁移到新组件 |
| `frontend/src/views/network/networkpolicy/NetworkPolicyList.vue` | 修改 | 迁移到新组件 |
| `frontend/src/views/storage/PVCList.vue` | 修改 | 迁移到新组件 |
| `frontend/src/views/storage/PVList.vue` | 修改 | 迁移到新组件（无命名空间选择） |
| `frontend/src/views/storage/StorageClassList.vue` | 修改 | 迁移到新组件（无命名空间、无总计数） |
| `frontend/src/views/storage/VolumeSnapshotClassList.vue` | 修改 | 迁移到新组件（无命名空间、无总计数） |
| `frontend/src/views/storage/VolumeSnapshotList.vue` | 修改 | 迁移到新组件（无总计数） |
| `frontend/src/views/config/configmap/ConfigMapList.vue` | 修改 | 迁移到新组件 |
| `frontend/src/views/config/secret/SecretList.vue` | 修改 | 迁移到新组件 |
| `frontend/src/views/config/limitrange/LimitRangeList.vue` | 修改 | 迁移到新组件 |
| `frontend/src/views/config/resourcequota/ResourceQuotaList.vue` | 修改 | 迁移到新组件 |

---

### Task 1: 创建 ResourceListToolbar 共享组件

**Files:**
- Create: `frontend/src/components/ResourceListToolbar.vue`

**Interfaces:**
- Consumes: props (searchValue, namespaceValue, namespaceList, totalCount, selectedCount, showCreate, showNamespace, showTotalCount), slots (actions, extra)
- Produces: `ResourceListToolbar` 组件，供所有列表页面使用

- [ ] **Step 1: 创建 ResourceListToolbar 组件**

创建 `frontend/src/components/ResourceListToolbar.vue`，包含以下结构：

```vue
<script setup lang="ts">
import { Search } from '@element-plus/icons-vue'

interface Props {
  searchValue: string
  namespaceValue?: string
  namespaceList?: string[]
  totalCount?: number
  selectedCount?: number
  showCreate?: boolean
  showNamespace?: boolean
  showTotalCount?: boolean
  searchPlaceholder?: string
  namespacePlaceholder?: string
}

const props = withDefaults(defineProps<Props>(), {
  namespaceValue: '',
  namespaceList: () => [],
  selectedCount: 0,
  showCreate: true,
  showNamespace: true,
  showTotalCount: true,
  searchPlaceholder: '搜索名称',
  namespacePlaceholder: '所有命名空间',
})

const emit = defineEmits<{
  'update:searchValue': [value: string]
  'update:namespaceValue': [value: string]
  searchInput: [value: string]
  namespaceChange: [value: string]
  create: []
  batchDelete: []
}>()
</script>

<template>
  <el-card shadow="never" class="filter-card">
    <div class="filter-bar">
      <!-- 左侧筛选区 -->
      <el-input
        :model-value="searchValue"
        @input="emit('searchInput', $event)"
        :placeholder="searchPlaceholder"
        style="width: 220px;"
        clearable
      >
        <template #prefix><el-icon><Search /></el-icon></template>
      </el-input>
      <el-select
        v-if="showNamespace"
        :model-value="namespaceValue"
        @update:model-value="emit('update:namespaceValue', $event)"
        :placeholder="namespacePlaceholder"
        clearable
        style="width: 180px;"
        @change="emit('namespaceChange', $event)"
      >
        <el-option v-for="ns in namespaceList" :key="ns" :label="ns" :value="ns" />
      </el-select>
      <span class="total-count" v-if="showTotalCount && totalCount">总计: {{ totalCount }}</span>

      <!-- 竖线分隔符 -->
      <div class="action-divider" />

      <!-- 主要操作组（连接式按钮组） -->
      <div class="action-group">
        <slot name="actions" />
      </div>

      <!-- 竖线分隔符 -->
      <div class="action-divider" />

      <!-- 辅助工具区 -->
      <slot name="extra" />
    </div>
  </el-card>
</template>

<style scoped>
.filter-card {
  margin-bottom: 16px;
}
.filter-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.total-count {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
.action-divider {
  width: 1px;
  height: 20px;
  background: var(--el-border-color-lighter);
  margin: 0 4px;
}
.action-group {
  display: inline-flex;
}
.action-group :deep(.el-button) {
  border-radius: 0;
  margin-left: -1px;
}
.action-group :deep(.el-button:first-child) {
  border-radius: 4px 0 0 4px;
  margin-left: 0;
}
.action-group :deep(.el-button:last-child) {
  border-radius: 0 4px 4px 0;
}
</style>
```

- [ ] **Step 2: 验证组件编译**

运行: `cd /Users/zqqzqq/05_github/gkube/frontend && npm run build 2>&1 | head -20`
预期: 编译成功，无错误

- [ ] **Step 3: 提交**

```bash
git add frontend/src/components/ResourceListToolbar.vue
git commit -m "feat: add ResourceListToolbar shared component with connected button group layout"
```

---

### Task 2: 迁移 DeploymentList 到新组件（参考实现）

**Files:**
- Modify: `frontend/src/views/workload/DeploymentList.vue`

**Interfaces:**
- Consumes: `ResourceListToolbar` 组件
- Produces: 迁移后的 DeploymentList，作为其他页面的参考模板

- [ ] **Step 1: 修改 DeploymentList.vue 模板和样式**

将 `DeploymentList.vue` 的 `<template>` 部分替换为使用 `ResourceListToolbar`：

```vue
<script setup lang="ts">
import { Plus, Delete } from '@element-plus/icons-vue'
// ... 其他 import 不变 ...
import ResourceListToolbar from '@/components/ResourceListToolbar.vue'
// ... 删除 Search 的 import（已移入组件）
```

模板部分替换 filter-card 为：

```html
<ResourceListToolbar
  :search-value="searchName"
  :namespace-value="selectedNamespace"
  :namespace-list="namespaceList"
  :total-count="totalCount"
  :selected-count="selectedRows.length"
  @search-input="onSearchInput"
  @namespace-change="handleNamespaceChange"
  @create="$router.push('/workloads/deployments/create')"
  @batch-delete="handleBatchDelete"
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
```

样式部分删除已移入组件的 CSS（`.filter-card`, `.filter-bar`, `.total-count`），保留 `.page-container`, `.table-card`, `.load-more` 和全局 `.yaml-drawer` 样式。

- [ ] **Step 2: 验证编译和功能**

运行: `cd /Users/zqqzqq/05_github/gkube/frontend && npm run build 2>&1 | head -20`
预期: 编译成功

- [ ] **Step 3: 提交**

```bash
git add frontend/src/views/workload/DeploymentList.vue
git commit -m "refactor(DeploymentList): migrate to ResourceListToolbar component"
```

---

### Task 3: 迁移 StatefulSetList、DaemonSetList、JobList、CronJobList

**Files:**
- Modify: `frontend/src/views/workload/StatefulSetList.vue`
- Modify: `frontend/src/views/workload/DaemonSetList.vue`
- Modify: `frontend/src/views/workload/JobList.vue`
- Modify: `frontend/src/views/workload/CronJobList.vue`

**Interfaces:**
- Consumes: `ResourceListToolbar` 组件
- Produces: 4 个迁移后的列表页面

这 4 个页面与 DeploymentList 结构完全相同（都有创建按钮、批量删除、命名空间选择、总计数），只需替换各自的路由路径。

- [ ] **Step 1: 修改 StatefulSetList.vue**

替换 filter-card 为 `ResourceListToolbar`，使用 `/workloads/statefulsets/create` 路由。删除 `Search` import。删除已移入组件的 CSS。

- [ ] **Step 2: 修改 DaemonSetList.vue**

同上，路由改为 `/workloads/daemonsets/create`。

- [ ] **Step 3: 修改 JobList.vue**

同上，路由改为 `/workloads/jobs/create`。

- [ ] **Step 4: 修改 CronJobList.vue**

同上，路由改为 `/workloads/cronjobs/create`。

- [ ] **Step 5: 验证编译**

运行: `cd /Users/zqqzqq/05_github/gkube/frontend && npm run build 2>&1 | head -20`
预期: 编译成功

- [ ] **Step 6: 提交**

```bash
git add frontend/src/views/workload/StatefulSetList.vue frontend/src/views/workload/DaemonSetList.vue frontend/src/views/workload/JobList.vue frontend/src/views/workload/CronJobList.vue
git commit -m "refactor(workload lists): migrate StatefulSet, DaemonSet, Job, CronJob to ResourceListToolbar"
```

---

### Task 4: 迁移 PodList、ReplicaSetList（无创建按钮）

**Files:**
- Modify: `frontend/src/views/workload/PodList.vue`
- Modify: `frontend/src/views/workload/ReplicaSetList.vue`

**Interfaces:**
- Consumes: `ResourceListToolbar` 组件
- Produces: 2 个迁移后的列表页面

PodList 无创建按钮，ReplicaSetList 无创建按钮且无总计数。

- [ ] **Step 1: 修改 PodList.vue**

替换 filter-card 为 `ResourceListToolbar`，设置 `:show-create="false"`。actions slot 只包含批量删除按钮。删除 `Search` import。删除已移入组件的 CSS。

- [ ] **Step 2: 修改 ReplicaSetList.vue**

同上，额外设置 `:show-total-count="false"`。

- [ ] **Step 3: 验证编译**

运行: `cd /Users/zqqzqq/05_github/gkube/frontend && npm run build 2>&1 | head -20`
预期: 编译成功

- [ ] **Step 4: 提交**

```bash
git add frontend/src/views/workload/PodList.vue frontend/src/views/workload/ReplicaSetList.vue
git commit -m "refactor(PodList, ReplicaSetList): migrate to ResourceListToolbar"
```

---

### Task 5: 迁移网络列表页面（Service、Ingress、NetworkPolicy）

**Files:**
- Modify: `frontend/src/views/network/ServiceList.vue`
- Modify: `frontend/src/views/network/IngressList.vue`
- Modify: `frontend/src/views/network/networkpolicy/NetworkPolicyList.vue`

**Interfaces:**
- Consumes: `ResourceListToolbar` 组件
- Produces: 3 个迁移后的列表页面

- [ ] **Step 1: 修改 ServiceList.vue**

替换 filter-card 为 `ResourceListToolbar`。删除 `Search` import。删除已移入组件的 CSS。

- [ ] **Step 2: 修改 IngressList.vue**

同上。

- [ ] **Step 3: 修改 NetworkPolicyList.vue**

同上。

- [ ] **Step 4: 验证编译**

运行: `cd /Users/zqqzqq/05_github/gkube/frontend && npm run build 2>&1 | head -20`
预期: 编译成功

- [ ] **Step 5: 提交**

```bash
git add frontend/src/views/network/ServiceList.vue frontend/src/views/network/IngressList.vue frontend/src/views/network/networkpolicy/NetworkPolicyList.vue
git commit -m "refactor(network lists): migrate Service, Ingress, NetworkPolicy to ResourceListToolbar"
```

---

### Task 6: 迁移存储列表页面（PVC、PV、StorageClass、VolumeSnapshotClass、VolumeSnapshot）

**Files:**
- Modify: `frontend/src/views/storage/PVCList.vue`
- Modify: `frontend/src/views/storage/PVList.vue`
- Modify: `frontend/src/views/storage/StorageClassList.vue`
- Modify: `frontend/src/views/storage/VolumeSnapshotClassList.vue`
- Modify: `frontend/src/views/storage/VolumeSnapshotList.vue`

**Interfaces:**
- Consumes: `ResourceListToolbar` 组件
- Produces: 5 个迁移后的列表页面

注意：PV、StorageClass、VolumeSnapshotClass 是集群级别资源，无命名空间选择器。

- [ ] **Step 1: 修改 PVCList.vue**

替换 filter-card 为 `ResourceListToolbar`。

- [ ] **Step 2: 修改 PVList.vue**

替换为 `ResourceListToolbar`，设置 `:show-namespace="false"`。

- [ ] **Step 3: 修改 StorageClassList.vue**

替换为 `ResourceListToolbar`，设置 `:show-namespace="false"` 和 `:show-total-count="false"`。

- [ ] **Step 4: 修改 VolumeSnapshotClassList.vue**

替换为 `ResourceListToolbar`，设置 `:show-namespace="false"` 和 `:show-total-count="false"`。

- [ ] **Step 5: 修改 VolumeSnapshotList.vue**

替换为 `ResourceListToolbar`，设置 `:show-total-count="false"`。

- [ ] **Step 6: 验证编译**

运行: `cd /Users/zqqzqq/05_github/gkube/frontend && npm run build 2>&1 | head -20`
预期: 编译成功

- [ ] **Step 7: 提交**

```bash
git add frontend/src/views/storage/
git commit -m "refactor(storage lists): migrate PVC, PV, StorageClass, VolumeSnapshot pages to ResourceListToolbar"
```

---

### Task 7: 迁移配置列表页面（ConfigMap、Secret、LimitRange、ResourceQuota）

**Files:**
- Modify: `frontend/src/views/config/configmap/ConfigMapList.vue`
- Modify: `frontend/src/views/config/secret/SecretList.vue`
- Modify: `frontend/src/views/config/limitrange/LimitRangeList.vue`
- Modify: `frontend/src/views/config/resourcequota/ResourceQuotaList.vue`

**Interfaces:**
- Consumes: `ResourceListToolbar` 组件
- Produces: 4 个迁移后的列表页面

- [ ] **Step 1: 修改 ConfigMapList.vue**

替换 filter-card 为 `ResourceListToolbar`，设置 `:show-total-count="false"`。

- [ ] **Step 2: 修改 SecretList.vue**

同上。

- [ ] **Step 3: 修改 LimitRangeList.vue**

同上。

- [ ] **Step 4: 修改 ResourceQuotaList.vue**

同上。

- [ ] **Step 5: 验证编译**

运行: `cd /Users/zqqzqq/05_github/gkube/frontend && npm run build 2>&1 | head -20`
预期: 编译成功

- [ ] **Step 6: 提交**

```bash
git add frontend/src/views/config/
git commit -m "refactor(config lists): migrate ConfigMap, Secret, LimitRange, ResourceQuota to ResourceListToolbar"
```

---

### Task 8: 最终验证

**Files:** 无新增/修改

- [ ] **Step 1: 完整构建验证**

运行: `cd /Users/zqqzqq/05_github/gkube/frontend && npm run build`
预期: 构建成功，无 TypeScript 错误

- [ ] **Step 2: 检查所有页面使用情况**

运行: `grep -r "class=\"filter-bar\"" frontend/src/views/ --include="*.vue" -l`
预期: 无结果（所有页面已迁移到 ResourceListToolbar）

- [ ] **Step 3: 提交最终状态**

```bash
git add -A
git commit -m "chore: verify all list pages migrated to ResourceListToolbar"
```
