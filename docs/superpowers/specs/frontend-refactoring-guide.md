# 前端巨型文件拆解方案

> 日期: 2026-09-02 | 状态: 待实施
> 目标读者: 实施 AI（Claude Code / 其他编码代理）
> 原则: 每步独立可验证，拆解不改变任何运行时行为

---

## 总览

| # | 文件 | 现状 | 拆解目标 |
|---|------|------|---------|
| 1 | `api/resource.ts` | 1284 行，225 个 export，所有资源 API 混在一起 | 按域拆分为 6 个文件 + 1 个泛型工厂 |
| 2 | 详情页（10+ 个） | 每个 700-1350 行，布局/事件/操作逻辑大量复制 | 提取 `useDetailPage` composable，详情页缩减 60%+ |
| 3 | `WorkloadForm.vue` | 91KB，容器/卷/探针/环境变量全部内联 | 拆为 5 个子组件 |

---

## 一、`api/resource.ts` 拆解

### 1.1 现状分析

当前文件按出现顺序排列，没有分组。每种资源重复相同模式：

```typescript
// 每种资源 5-10 个函数，模式完全相同
export function getXxxList(params?) { return request.get('/k8s/xxx/list', { params }) }
export function getXxxDetail(params) { return request.get('/k8s/xxx/detail', { params }) }
export function getXxxYaml(params) { return request.get('/k8s/xxx/get-yaml', { params }) }
export function createXxx(data) { return request.post('/k8s/xxx/create', data) }
export function updateXxx(data) { return request.put('/k8s/xxx/update-yaml', data) }
export function deleteXxx(data) { return request.delete('/k8s/xxx/delete', { data }) }
export function getXxxEvents(params) { return request.get('/k8s/xxx/events', { params }) }
```

### 1.2 拆解步骤

#### 第一步：新建 `api/factory.ts` — 泛型资源 API 工厂

```typescript
// frontend/src/api/factory.ts
import request from './request'

/**
 * 为一种 K8s 资源生成标准 CRUD API 函数。
 * 资源特有的操作（scale/restart/rollback）在各域文件中单独定义。
 *
 * @param basePath API 路径前缀，如 '/k8s/deployment'
 */
export function createResourceApi(basePath: string) {
  return {
    list:       (params?: any) => request.get(`${basePath}/list`, { params }),
    detail:     (params: any) => request.get(`${basePath}/detail`, { params }),
    getYaml:    (params: any) => request.get(`${basePath}/get-yaml`, { params }),
    create:     (data: any) => request.post(`${basePath}/create`, data),
    updateYaml: (data: any) => request.put(`${basePath}/update-yaml`, data),
    delete:     (data: any) => request.delete(`${basePath}/delete`, { data }),
    events:     (params: any) => request.get(`${basePath}/events`, { params }),
  }
}
```

#### 第二步：按域拆分为 6 个文件

```
frontend/src/api/
├── factory.ts          ← 泛型工厂（新建）
├── workload.ts         ← Pod/Deployment/StatefulSet/DaemonSet/Job/CronJob/ReplicaSet/HPA
├── network.ts          ← Service/Ingress/NetworkPolicy
├── storage.ts          ← PV/PVC/StorageClass/VolumeSnapshot/VolumeSnapshotClass
├── config.ts           ← ConfigMap/Secret/ResourceQuota/LimitRange
├── core.ts             ← Node/Namespace/Event/CRD/CustomResource/Dashboard
├── resource.ts         ← 旧文件，改为 re-export 所有域的 API（保持向后兼容）
├── auth.ts             ← 不变
├── cluster.ts          ← 不变
├── rbac.ts             ← 不变
└── request.ts          ← 不变
```

#### 第三步：各域文件内容规范

以 `api/workload.ts` 为例，展示完整结构：

```typescript
// frontend/src/api/workload.ts
import request from './request'
import { createResourceApi } from './factory'

// ============ 类型定义 ============

export interface Pod {
  namespace: string
  name: string
  status: string
  restarts: number
  node: string
  ip: string
  age: string
  raw: any
}

export interface Deployment {
  namespace: string
  name: string
  ready: string
  upToDate: number
  available: number
  age: string
  images: string
  raw: any
}

export interface StatefulSet { /* ... */ }
export interface DaemonSet { /* ... */ }
export interface Job { /* ... */ }
export interface CronJob { /* ... */ }

// ============ Transform 函数 ============

export function transformPods(items: any[]): Pod[] { /* ... */ }
export function transformDeployments(items: any[]): Deployment[] { /* ... */ }
export function transformStatefulSets(items: any[]): StatefulSet[] { /* ... */ }
export function transformDaemonSets(items: any[]): DaemonSet[] { /* ... */ }
export function transformJobs(items: any[]): Job[] { /* ... */ }
export function transformCronJobs(items: any[]): CronJob[] { /* ... */ }

// ============ 标准 CRUD（工厂生成） ============

export const podApi = createResourceApi('/k8s/pod')
export const deploymentApi = createResourceApi('/k8s/deployment')
export const statefulSetApi = createResourceApi('/k8s/statefulset')
export const daemonSetApi = createResourceApi('/k8s/daemonset')
export const jobApi = createResourceApi('/k8s/job')
export const cronJobApi = createResourceApi('/k8s/cronjob')
export const replicaSetApi = createResourceApi('/k8s/replicaset')
export const hpaApi = createResourceApi('/k8s/hpa')

// ============ 资源特有操作（工厂无法覆盖的部分） ============

// Pod
export const deletePod = (data: any) => request.delete('/k8s/pod/delete', { data })

// Deployment
export const scaleDeployment = (data: any) => request.put('/k8s/deployment/scale', data)
export const restartDeployment = (data: any) => request.post('/k8s/deployment/restart', data)
export const rollbackDeployment = (data: any) => request.post('/k8s/deployment/rollback', data)
export const updateDeploymentImage = (data: any) => request.put('/k8s/deployment/update-image', data)
export const getDeploymentReplicaSets = (params: any) => request.get('/k8s/deployment/replicasets', { params })
export const getDeploymentPodList = (params: any) => request.get('/k8s/deployment/pods', { params })

// StatefulSet
export const scaleStatefulSet = (data: any) => request.put('/k8s/statefulset/scale', data)
export const restartStatefulSet = (data: any) => request.post('/k8s/statefulset/restart', data)
export const rollbackStatefulSet = (data: any) => request.post('/k8s/statefulset/rollback', data)
export const updateStatefulSetImage = (data: any) => request.put('/k8s/statefulset/update-image', data)
export const getStatefulSetRollbacks = (params: any) => request.get('/k8s/statefulset/rollbacks', { params })
export const getStatefulSetPVCs = (params: any) => request.get('/k8s/statefulset/pvcs', { params })

// DaemonSet
export const restartDaemonSet = (data: any) => request.post('/k8s/daemonset/restart', data)
export const rollbackDaemonSet = (data: any) => request.post('/k8s/daemonset/rollback', data)
export const updateDaemonSetImage = (data: any) => request.put('/k8s/daemonset/update-image', data)
export const getDaemonSetRollbacks = (params: any) => request.get('/k8s/daemonset/rollbacks', { params })

// Job
export const rerunJob = (params: any) => request.post('/k8s/job/rerun', params)

// CronJob
export const suspendCronJob = (params: any) => request.put('/k8s/cronjob/suspend', params)
export const resumeCronJob = (params: any) => request.put('/k8s/cronjob/resume', params)
export const triggerCronJob = (params: any) => request.post('/k8s/cronjob/trigger', params)

// HPA
export const pauseHpa = (params: any) => request.post('/k8s/hpa/pause', params)
export const resumeHpa = (params: any) => request.post('/k8s/hpa/resume', params)
```

其他域文件（`network.ts`、`storage.ts`、`config.ts`、`core.ts`）遵循相同结构。

#### 第四步：旧 `resource.ts` 改为 re-export 门面

```typescript
// frontend/src/api/resource.ts — 保持向后兼容
// 所有 import 从这里来的页面不需要改任何代码

// 重新导出所有域的 API
export * from './factory'
export * from './workload'
export * from './network'
export * from './storage'
export * from './config'
export * from './core'

// 重新导出通用工具函数
export { extractNamespaceNames, calcAge } from './core'
```

**关键：** 第四步保证现有页面 `import { getDeploymentList } from '@/api/resource'` 不会报错。迁移完成后可以逐步将 import 改为直接从域文件导入。

### 1.3 验收

- [ ] `npm run build` 零错误
- [ ] 现有页面所有 import 不变（通过 re-export 门面兼容）
- [ ] 新建 `factory.ts` + 5 个域文件 + 1 个门面文件
- [ ] 每个域文件包含：类型定义 + transform + 标准 CRUD（工厂） + 特有操作

---

## 二、详情页拆解

### 2.1 现状分析

14 个详情页，每个 650-1350 行。以下是代码重复度对比：

| 逻辑 | Deployment | StatefulSet | DaemonSet | Job | CronJob | Pod |
|------|:---:|:---:|:---:|:---:|:---:|:---:|
| fetchDetail + fetchEvents | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| useAutoRefresh | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| 左右可调布局 (useResizable) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| YamlDrawer | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| handleDelete | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| handleScale 弹窗 | ✅ | ✅ | — | — | — | — |
| handleRestart | ✅ | ✅ | ✅ | — | — | — |
| handleUpdateImage 弹窗 | ✅ | ✅ | ✅ | — | — | — |
| 事件表格 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| header-actions 按钮栏 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| CSS (header/panel/actions) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

### 2.2 拆解步骤

#### 第一步：新建 `composables/useDetailPage.ts`

```typescript
// frontend/src/composables/useDetailPage.ts

import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAutoRefresh } from './useAutoRefresh'
import { useClusterStore } from '@/stores/cluster'

export interface DetailPageOptions {
  /** 资源显示名，如 'Deployment' */
  resourceName: string
  /** 获取详情 */
  fetchDetail: (params: any) => Promise<any>
  /** 获取事件（可选） */
  fetchEvents?: (params: any) => Promise<any>
  /** 获取 YAML */
  getYaml: (params: any) => Promise<any>
  /** 更新 YAML（可选） */
  updateYaml?: (data: any) => Promise<any>
  /** 删除资源 */
  deleteResource: (data: any) => Promise<any>
  /** 强制删除（可选，Pod 用） */
  forceDeleteResource?: (data: any) => Promise<any>
  /** 列表页路由（返回按钮用） */
  listRoute: string
  /** 从 route params 构造 API 请求参数 */
  buildParams: () => Record<string, any>
  /** 自定义删除确认消息（可选） */
  deleteConfirm?: (detail: any) => string
}

export function useDetailPage(options: DetailPageOptions) {
  const route = useRoute()
  const router = useRouter()
  const clusterStore = useClusterStore()

  const loading = ref(false)
  const detail = ref<any>(null)
  const events = ref<any[]>([])
  const yamlVisible = ref(false)
  const deleteLoading = ref(false)

  // ---- 数据获取 ----

  async function fetchDetail() {
    loading.value = true
    try {
      const res = await options.fetchDetail(options.buildParams())
      detail.value = res?.data ?? res
    } catch (e: any) {
      ElMessage.error(e?.message || `获取${options.resourceName}详情失败`)
    } finally {
      loading.value = false
    }
  }

  async function fetchEvents() {
    if (!options.fetchEvents) return
    try {
      const res = await options.fetchEvents(options.buildParams())
      events.value = res?.data ?? res ?? []
    } catch {
      // 事件加载失败不阻塞页面
    }
  }

  // ---- 操作 ----

  async function handleDelete(force = false) {
    const msg = options.deleteConfirm?.(detail.value)
      || `确定删除 ${options.resourceName} "${detail.value?.metadata?.name}" 吗？`
    try {
      await ElMessageBox.confirm(msg, '确认删除', { type: 'warning' })
    } catch { return }

    deleteLoading.value = true
    try {
      const params = options.buildParams()
      if (force && options.forceDeleteResource) {
        await options.forceDeleteResource({ ...params, force: true })
      } else {
        await options.deleteResource(params)
      }
      ElMessage.success('已删除')
      router.push(options.listRoute)
    } catch (e: any) {
      ElMessage.error(e?.message || '删除失败')
    } finally {
      deleteLoading.value = false
    }
  }

  function handleOpenYaml() {
    yamlVisible.value = true
  }

  // ---- 自动刷新 ----

  const { isRunning, countdown, currentInterval, availableIntervals,
          toggle, refresh: manualRefresh, setIntervalOption } = useAutoRefresh(async () => {
    await fetchDetail()
    await fetchEvents()
  })

  // ---- 初始化 ----

  onMounted(async () => {
    await fetchDetail()
    fetchEvents()
  })

  return {
    // 状态
    loading, detail, events, yamlVisible, deleteLoading,
    // 自动刷新
    isRunning, countdown, currentInterval, availableIntervals,
    toggle, manualRefresh, setIntervalOption,
    // 操作
    fetchDetail, fetchEvents, handleDelete, handleOpenYaml,
    // 工具
    router,
  }
}
```

#### 第二步：新建 `components/DetailPageLayout.vue` — 共享布局组件

```vue
<!-- frontend/src/components/DetailPageLayout.vue -->
<!-- 提供：header 区 + 左右可调面板 + 事件表格的统一布局 -->

<script setup lang="ts">
defineProps<{
  title: string
  loading?: boolean
}>()
</script>

<template>
  <div class="detail-page">
    <!-- 顶部操作栏 -->
    <div class="detail-header">
      <div class="header-left">
        <slot name="header-left" />
      </div>
      <div class="header-actions">
        <slot name="actions" />
      </div>
    </div>

    <!-- 刷新 + 返回 -->
    <div class="detail-toolbar">
      <slot name="toolbar" />
    </div>

    <!-- 左右面板 -->
    <div class="detail-body">
      <div class="panel-left">
        <slot name="left" />
      </div>
      <div class="panel-right">
        <slot name="right" />
      </div>
    </div>

    <!-- 事件表格（通用） -->
    <slot name="events" />
  </div>
</template>

<style scoped>
.detail-page { padding: 16px; display: flex; flex-direction: column; height: 100%; }
.detail-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.header-actions { display: flex; gap: 8px; flex-wrap: wrap; }
.detail-toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.detail-body { display: flex; gap: 16px; flex: 1; min-height: 0; }
.panel-left { flex: 1; overflow: auto; }
.panel-right { flex: 2; overflow: auto; }
</style>
```

#### 第三步：逐个改造详情页

以 `DeploymentDetail.vue` 为例，改造前后对比：

**改造前（1096 行）：**
- 100+ 行 script（fetchDetail/fetchEvents/handleDelete/handleScale/handleRestart/handleUpdateImage/...）
- 300+ 行 template（header/panels/events/scale-dialog/restart-dialog/image-dialog/yaml-drawer）
- 200+ 行 CSS

**改造后（约 400 行）：**
```vue
<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useDetailPage } from '@/composables/useDetailPage'
import DetailPageLayout from '@/components/DetailPageLayout.vue'
import YamlDrawer from '@/components/YamlDrawer.vue'
import ScaleDialog from './components/ScaleDialog.vue'
import RestartDialog from './components/RestartDialog.vue'
import UpdateImageDialog from './components/UpdateImageDialog.vue'
import {
  deploymentApi, getDeploymentYaml, updateDeploymentYaml,
  deleteDeployment, scaleDeployment, restartDeployment,
  rollbackDeployment, updateDeploymentImage,
  getDeploymentReplicaSets, getDeploymentPodList,
  transformDeployments,
} from '@/api/workload'

const route = useRoute()
const {
  loading, detail, events, yamlVisible,
  isRunning, countdown, currentInterval, availableIntervals,
  toggle, manualRefresh, setIntervalOption,
  fetchDetail, fetchEvents, handleDelete, handleOpenYaml, router,
} = useDetailPage({
  resourceName: 'Deployment',
  fetchDetail: (p) => deploymentApi.detail(p),
  fetchEvents: (p) => deploymentApi.events(p),
  getYaml: getDeploymentYaml,
  updateYaml: updateDeploymentYaml,
  deleteResource: deleteDeployment,
  listRoute: '/workloads/deployments',
  buildParams: () => ({
    namespace: route.query.namespace as string,
    name: route.params.name as string,
  }),
})

// ---- Deployment 特有逻辑 ----
const scaleDialogVisible = ref(false)
const restartDialogVisible = ref(false)
const imageDialogVisible = ref(false)
const replicasets = ref<any[]>([])
// ... 特有操作函数（每个约 10 行）
</script>

<template>
  <DetailPageLayout :loading="loading">
    <template #actions>
      <el-button type="primary" @click="scaleDialogVisible = true">扩缩容</el-button>
      <el-button type="warning" @click="restartDialogVisible = true">重启</el-button>
      <el-button type="success" @click="imageDialogVisible = true">更新镜像</el-button>
      <el-button @click="handleOpenYaml">YAML</el-button>
      <el-button type="danger" plain @click="handleDelete">删除</el-button>
    </template>

    <template #toolbar>
      <!-- 返回按钮 + 刷新按钮 + 自动刷新控件 -->
    </template>

    <template #left>
      <!-- ReplicaSet 历史 + 基本信息 -->
    </template>

    <template #right>
      <!-- Pod 列表 -->
    </template>

    <template #events>
      <!-- 事件表格（可复用 EventsTable 组件） -->
    </template>
  </DetailPageLayout>

  <YamlDrawer ... />
  <ScaleDialog v-model:visible="scaleDialogVisible" ... />
  <RestartDialog v-model:visible="restartDialogVisible" ... />
  <UpdateImageDialog v-model:visible="imageDialogVisible" ... />
</template>
```

#### 第四步：提取可复用的操作弹窗组件

```
frontend/src/views/workload/components/
├── ScaleDialog.vue          ← 从 DeploymentDetail/StatefulSetDetail 提取
├── RestartDialog.vue        ← 从 DeploymentDetail/StatefulSetDetail/DaemonSetDetail 提取
├── UpdateImageDialog.vue    ← 从 DeploymentDetail/StatefulSetDetail/DaemonSetDetail 提取
└── EventsTable.vue          ← 从所有详情页提取（通用事件表格）
```

### 2.3 验收

- [ ] `npm run build` 零错误
- [ ] `useDetailPage` composable 被至少 3 个详情页使用
- [ ] `DetailPageLayout` 组件被至少 3 个详情页使用
- [ ] `ScaleDialog`/`RestartDialog`/`UpdateImageDialog` 各自被至少 2 个详情页复用
- [ ] 每个改造后的详情页行数减少 50% 以上
- [ ] 所有详情页的功能（刷新/删除/YAML/事件/操作弹窗）行为不变

---

## 三、`WorkloadForm.vue` 拆解

### 3.1 现状分析

91KB / ~2500 行，包含：
- 容器列表管理（添加/删除/编辑容器）
- 端口配置
- 环境变量配置（key-value / secret 引用 / configmap 引用）
- 卷挂载配置（emptyDir/hostPath/configMap/secret/pvc/nfs）
- 存活/就绪/启动探针配置
- 标签/注解配置
- 资源限制配置
- 调度策略（nodeSelector/affinity/tolerations）
- 更新策略
- ServiceAccount / HostNetwork / DNS 配置

### 3.2 拆解步骤

#### 子组件结构

```
frontend/src/views/workload/components/
├── WorkloadForm.vue              ← 主表单（瘦身后 ~800 行），负责表单状态 + 提交
├── form/
│   ├── ContainerForm.vue         ← 容器编辑弹窗（镜像/端口/命令/工作目录/安全上下文）
│   ├── EnvVarForm.vue            ← 环境变量编辑（plain/secretRef/configMapRef 三种模式）
│   ├── VolumeMountForm.vue       ← 卷挂载编辑（6 种卷类型 + 挂载路径/子路径/只读）
│   ├── ProbeForm.vue             ← 探针编辑（httpGet/exec/tcpSocket/grpc + 参数）
│   └── SchedulingForm.vue        ← 调度策略（nodeSelector/affinity/tolerations）
```

#### 拆解方式

**原则：每个子组件管理自己的局部状态，通过 v-model 与父组件通信。**

以 `ContainerForm.vue` 为例：

```vue
<!-- ContainerForm.vue — 容器编辑弹窗 -->
<script setup lang="ts">
// Props: container 对象（编辑模式）或 null（新建模式）
// Emits: submit(container) — 用户确认后把容器数据传回父组件
// 内部状态: 镜像、端口列表、命令、参数、工作目录、安全上下文
// 不直接操作父组件的 containers 数组
</script>

<template>
  <el-dialog title="编辑容器" width="720px">
    <el-form ...>
      <el-form-item label="镜像"><el-input v-model="form.image" /></el-form-item>
      <el-form-item label="端口">
        <!-- 端口列表，可增删 -->
      </el-form-item>
      <!-- 其他字段 -->
    </el-form>
    <template #footer>
      <el-button @click="emit('cancel')">取消</el-button>
      <el-button type="primary" @click="emit('submit', form)">确定</el-button>
    </template>
  </el-dialog>
</template>
```

**WorkloadForm 瘦身后：**

```vue
<script setup lang="ts">
import ContainerForm from './form/ContainerForm.vue'
import EnvVarForm from './form/EnvVarForm.vue'
import VolumeMountForm from './form/VolumeMountForm.vue'
import ProbeForm from './form/ProbeForm.vue'
import SchedulingForm from './form/SchedulingForm.vue'

// 主表单状态
const form = ref({
  name: '',
  replicas: 1,
  containers: [],      // 容器列表（每个容器由 ContainerForm 编辑）
  volumes: [],
  labels: [],
  annotations: {},
  // ...
})

// 子组件弹窗状态
const containerDialogVisible = ref(false)
const editingContainerIndex = ref(-1)

// 子组件回调
function onContainerSubmit(container: any) {
  if (editingContainerIndex.value >= 0) {
    form.value.containers[editingContainerIndex.value] = container
  } else {
    form.value.containers.push(container)
  }
  containerDialogVisible.value = false
}
</script>

<template>
  <el-form>
    <!-- 基本信息 -->
    <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
    <el-form-item label="副本数"><el-input-number v-model="form.replicas" /></el-form-item>

    <!-- 容器列表（展示卡片 + 添加按钮） -->
    <el-form-item label="容器">
      <div v-for="(c, i) in form.containers" :key="i" class="container-card">
        {{ c.image }}
        <el-button @click="editContainer(i)">编辑</el-button>
        <el-button @click="form.containers.splice(i, 1)">删除</el-button>
      </div>
      <el-button @click="addContainer">+ 添加容器</el-button>
    </el-form-item>

    <!-- 卷、标签等其他区块 -->
  </el-form>

  <ContainerForm v-model:visible="containerDialogVisible" :container="editingContainer" @submit="onContainerSubmit" />
</template>
```

### 3.3 验收

- [ ] `npm run build` 零错误
- [ ] `WorkloadForm.vue` 行数从 ~2500 减少到 ~800
- [ ] 新增 5 个子组件文件
- [ ] Deployment/StatefulSet/DaemonSet 创建/编辑表单功能不变
- [ ] 子组件弹窗的表单校验行为不变

---

## 四、实施顺序

```
阶段 1：API 拆解（独立，不影响页面）
├── 1.1 新建 api/factory.ts
├── 1.2 新建 api/workload.ts（从 resource.ts 提取）
├── 1.3 新建 api/network.ts
├── 1.4 新建 api/storage.ts
├── 1.5 新建 api/config.ts
├── 1.6 新建 api/core.ts
├── 1.7 改 resource.ts 为 re-export 门面
└── 1.8 验收：npm run build + 所有页面 import 不变

阶段 2：详情页拆解（逐个改造，每改一个验证一个）
├── 2.1 新建 composables/useDetailPage.ts
├── 2.2 新建 components/DetailPageLayout.vue
├── 2.3 新建 workload/components/EventsTable.vue
├── 2.4 新建 workload/components/ScaleDialog.vue
├── 2.5 新建 workload/components/RestartDialog.vue
├── 2.6 新建 workload/components/UpdateImageDialog.vue
├── 2.7 改造 DeploymentDetail.vue
├── 2.8 改造 StatefulSetDetail.vue
├── 2.9 改造 DaemonSetDetail.vue
├── 2.10 改造 PodDetail.vue
├── 2.11 改造 JobDetail.vue / CronJobDetail.vue
└── 2.12 改造其余详情页

阶段 3：表单拆解（独立，不影响其他页面）
├── 3.1 新建 form/ContainerForm.vue
├── 3.2 新建 form/EnvVarForm.vue
├── 3.3 新建 form/VolumeMountForm.vue
├── 3.4 新建 form/ProbeForm.vue
├── 3.5 新建 form/SchedulingForm.vue
├── 3.6 改造 WorkloadForm.vue 使用子组件
└── 3.7 改造 CronJobForm.vue / JobForm.vue
```

每个阶段独立可交付。阶段之间无依赖，可并行。

---

## 五、全局禁令

1. **不改任何运行时行为。** 拆解是纯结构重组，不修改业务逻辑。
2. **每步完成后 `npm run build` 必须通过。** vue-tsc 捕获类型错误。
3. **不改 import 路径的消费者。** `resource.ts` 的 re-export 门面保证现有 import 不报错。
4. **不做"顺手改进"。** 不改命名、不改类型注解、不加新功能。
5. **子组件通过 props/emits 通信，不直接操作父组件状态。**
