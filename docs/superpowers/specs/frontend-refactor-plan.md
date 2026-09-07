# 前端架构重构方案（v4 — 最终版）

> 本文档为可被 AI 执行的精确重构指令。每个步骤包含：目标文件、具体操作、依赖关系。
> 已通过三轮代码 review，包括 15 项技术细节源码验证。

---

## 一、问题总览

| 问题 | 涉及文件 | 重复行数 |
|------|----------|----------|
| Detail 页面 CSS 重复 | 7 个 `*Detail.vue` | ~3500 行 |
| Detail 页面逻辑重复（Pod操作/重启/编辑/删除/YAML） | 7 个 `*Detail.vue` | ~500 行 |
| Detail 页面模板重复（事件表/信息区块/头部/修订列表） | 7 个 `*Detail.vue` | ~600 行 |
| Form 组件类型/工具函数重复 | 3 个 `*Form.vue` | ~800 行 |
| 无 K8s 资源类型定义 | 全局 | 119 个 `any` |
| API 工厂未被使用 | `api/*.ts` | 每个文件手写 CRUD |
| `core.ts` 杂物箱 | `api/core.ts` | 302 行混杂 4 域 |
| `useResizable` 事件泄漏 | `composables/useResizable.ts` | 拖拽中卸载组件会泄漏监听器 |
| 路由守卫安全漏洞 | `router/index.ts:597` | user=null 时放行视为管理员 |

---

## 二、重构步骤

### Step 1: 创建 K8s 资源类型定义

**新建文件**: `src/types/k8s.ts`

**依据**: 读取 `api/workload.ts` 第 6-66 行的 transform 输出接口，以及 Detail 页面中实际访问的 K8s 字段。

**执行指令**: AI 必须先读取 `api/workload.ts` 获取完整的接口定义，再读取各 Detail 页面中实际访问的 K8s 字段（如 `daemonset.value?.status?.desiredNumberScheduled`），综合定义类型。不要用注释占位，每个 interface 必须有完整字段。

**定义以下类型**（使用 `export interface`，非必填字段用 `?`）：

```
基础类型:
- K8sObjectMeta { name, namespace?, uid, resourceVersion, labels?, annotations?, creationTimestamp, deletionTimestamp? }
- K8sLabelSelector { matchLabels?, matchExpressions? }
- K8sCondition { type, status, lastTransitionTime, reason?, message? }
- K8sOwnerReference { apiVersion, kind, name, uid, controller?, blockOwnerDeletion? }

容器相关:
- K8sContainerPort { name?, containerPort, protocol?, hostIP?, hostPort? }
- K8sEnvVar { name, value?, valueFrom? }
- K8sResourceRequirements { requests?: {cpu?, memory?}, limits?: {cpu?, memory?} }
- K8sVolumeMount { name, mountPath, subPath?, readOnly? }
- K8sProbe { httpGet?, tcpSocket?, exec?, initialDelaySeconds?, periodSeconds?, timeoutSeconds?, successThreshold?, failureThreshold? }
- K8sLifecycleHandler { exec?, httpGet?, tcpSocket? }
- K8sLifecycle { postStart?, preStop? }
- K8sSecurityContext { runAsUser?, runAsGroup?, runAsNonRoot?, privileged?, readOnlyRootFilesystem?, allowPrivilegeEscalation?, capabilities? }
- K8sCapabilities { add?: string[], drop?: string[] }
- K8sContainer { name, image, command?, args?, ports?, env?, resources?, volumeMounts?, livenessProbe?, readinessProbe?, startupProbe?, lifecycle?, securityContext?, imagePullPolicy? }

Pod 相关:
- K8sPodSpec { containers, initContainers?, volumes?, nodeSelector?, tolerations?, affinity?, topologySpreadConstraints?, serviceAccountName?, hostNetwork?, dnsPolicy?, restartPolicy?, terminationGracePeriodSeconds?, imagePullSecrets?, priorityClassName?, securityContext? }
- K8sPodTemplateSpec { metadata?: K8sObjectMeta, spec: K8sPodSpec }
- K8sPodStatus { phase?, conditions?: K8sPodCondition[], hostIP?, podIP?, containerStatuses?, initContainerStatuses? }
- K8sPodCondition { type, status, lastTransitionTime, reason?, message? }
- K8sPod { apiVersion, kind, metadata: K8sObjectMeta, spec: K8sPodSpec, status?: K8sPodStatus }

工作负载:
- K8sDeploymentSpec { replicas?, selector: K8sLabelSelector, template: K8sPodTemplateSpec, strategy?: {type?, rollingUpdate?:{maxSurge?,maxUnavailable?}} }
- K8sDeploymentStatus { replicas?, readyReplicas?, availableReplicas?, updatedReplicas?, conditions?: K8sCondition[] }
- K8sDeployment { apiVersion:'apps/v1', kind:'Deployment', metadata: K8sObjectMeta, spec: K8sDeploymentSpec, status?: K8sDeploymentStatus }
- K8sStatefulSet — 类似 Deployment，额外有 serviceName, podManagementPolicy, updateStrategy, volumeClaimTemplates
- K8sDaemonSet — 类似 Deployment，额外有 updateStrategy (type: RollingUpdate/OnDelete)
- K8sReplicaSet — 类似 Deployment
- K8sJob { spec: { completions?, parallelism?, backoffLimit?, activeDeadlineSeconds?, ttlSecondsAfterFinished?, completionMode?, template: K8sPodTemplateSpec }, status: { active?, succeeded?, failed?, conditions? } }
- K8sCronJob { spec: { schedule, concurrencyPolicy?, suspend?, startingDeadlineSeconds?, timeZone?, successfulFailedHistoryLimit?, jobTemplate: { spec: K8sJob } }, status: { active?, lastScheduleTime?, lastSuccessfulTime? } }

网络:
- K8sServicePort { name?, protocol?, port, targetPort?, nodePort?, appProtocol? }
- K8sService { spec: { type?, selector?, ports: K8sServicePort[], clusterIP?, clusterIPs?, externalIPs?, loadBalancerIP?, sessionAffinity?, externalTrafficPolicy? }, status?: { loadBalancer?: { ingress? } } }
- K8sIngressRule { host?, http: { paths: { path, pathType, backend: { service: { name, port: { number } } } }[] } }
- K8sIngressTLS { hosts?: string[], secretName? }
- K8sIngress { spec: { ingressClassName?, rules?: K8sIngressRule[], tls?: K8sIngressTLS[], defaultBackend? }, status?: { loadBalancer? } }

存储:
- K8sPersistentVolume { spec: { capacity?, accessModes?, persistentVolumeReclaimPolicy?, storageClassName?, claimRef?, nfs?, hostPath?, csi? }, status?: { phase? } }
- K8sPersistentVolumeClaim { spec: { accessModes?, resources: { requests: { storage } }, storageClassName?, volumeName? }, status?: { phase?, capacity? } }
- K8sStorageClass { provisioner, parameters?, reclaimPolicy?, volumeBindingMode?, allowVolumeExpansion? }

配置:
- K8sConfigMap { data?: Record<string,string>, binaryData?: Record<string,string> }
- K8sSecret { data?: Record<string,string>, type?, stringData?: Record<string,string> }
- K8sResourceQuota { spec: { hard?: Record<string,string> }, status?: { hard?, used? } }
- K8sLimitRange { spec: { limits: { type, max?, min?, default?, defaultRequest?, maxLimitRequestRatio? }[] } }

其他:
- K8sHPA { spec: { scaleTargetRef, minReplicas?, maxReplicas, metrics?, behavior? }, status?: { currentReplicas, desiredReplicas, conditions? } }
- K8sNamespace { metadata: K8sObjectMeta, spec?: { finalizers? }, status?: { phase? } }
- K8sNode { metadata: K8sObjectMeta, spec: { podCIDR?, taints?, unschedulable? }, status: { conditions?, addresses?: {type,address}[], capacity?, allocatable?, nodeInfo?: { kubeletVersion, osImage, kernelVersion, containerRuntimeVersion } } }
- K8sEvent { involvedObject, reason?, message?, type?, count?, firstTimestamp?, lastTimestamp?, metadata: K8sObjectMeta }
- K8sToleration { key?, operator?, value?, effect?, tolerationSeconds? }
- K8sAffinity { nodeAffinity?, podAffinity?, podAntiAffinity? }
- K8sTopologySpreadConstraint { maxSkew, topologyKey, whenUnsatisfiable, labelSelector?, matchLabelKeys?, minDomains? }
```

---

### Step 2: 修复 `useResizable` 事件泄漏（前置依赖）

**修改文件**: `src/composables/useResizable.ts`

**问题**: `document.addEventListener('mousemove/mouseup')` 注册的监听器在组件卸载时未清理。如果用户拖拽过程中组件被卸载，监听器泄漏。

**修正**:
1. 将 `mousemove` 和 `mouseup` 的 handler 保存为变量引用
2. 在 `onBeforeUnmount` 中调用 `document.removeEventListener('mousemove', handler)` 和 `document.removeEventListener('mouseup', handler)`
3. 在 `mouseup` handler 中也要 `removeEventListener`（拖拽结束后清理 mousemove）

---

### Step 3: 创建 Detail 页面共享 composable

#### 3a. 新建 `src/composables/usePodActions.ts`

**依据**: `DeploymentDetail.vue:170-216` 的 `handlePodLogs`/`handlePodExec`/`handlePodDelete`，与 `DaemonSetDetail.vue:241-285`、`StatefulSetDetail.vue:178-222` 逐字相同。

**接口**:
```ts
import { Ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessageBox, ElMessage } from 'element-plus'
import { deletePod } from '@/api/resource'  // 注意：从 @/api/resource 导入，不是 @/api/workload

export function usePodActions(clusterName: Ref<string>) {
  const router = useRouter()

  const handlePodLogs = (pod: { namespace: string; name: string }) => {
    const route = router.resolve({
      name: 'fullscreen-logs',
      query: { cluster: clusterName.value, namespace: pod.namespace, pod: pod.name }
    })
    window.open(route.href, '_blank')
  }

  const handlePodExec = (pod: { namespace: string; name: string }) => {
    const route = router.resolve({
      name: 'fullscreen-terminal',
      query: { cluster: clusterName.value, namespace: pod.namespace, pod: pod.name }
    })
    window.open(route.href, '_blank')
  }

  // 原始代码签名是 handlePodDelete(pod, force=false)，无 onSuccess 回调
  // 此处增加 onSuccess 回调是改进，调用处需同步修改
  // 注意：没有 forceDeletePod 函数，强制删除通过 deletePod({..., force: true}) 实现
  const handlePodDelete = async (
    pod: { namespace: string; name: string },
    onSuccess?: () => void,
    force = false
  ) => {
    try {
      await ElMessageBox.confirm(
        force
          ? `确定强制删除 Pod "${pod.name}"？这将跳过优雅终止。`
          : `确定删除 Pod "${pod.name}"？`,
        force ? '强制删除 Pod' : '删除 Pod',
        { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
      )
      await deletePod({ namespace: pod.namespace, name: pod.name, force })
      ElMessage.success('删除成功')
      onSuccess?.()
    } catch (e: any) {
      if (e !== 'cancel') {
        ElMessage.error(e?.response?.data?.message || '删除失败')
      }
    }
  }

  return { handlePodLogs, handlePodExec, handlePodDelete }
}
```

#### 3b. 新建 `src/composables/useRestartAction.ts`

**依据**: `DeploymentDetail.vue:225-239` 的 `handleRestart`，与 `DaemonSetDetail.vue:292-307`、`StatefulSetDetail.vue:231-246` 结构相同，只是 API 函数不同。

**接口**:
```ts
import { ElMessageBox, ElMessage } from 'element-plus'

export function useRestartAction(
  kind: 'deployment' | 'statefulset' | 'daemonset',
  restartApi: (params: { namespace: string; name: string }) => Promise<any>
) {
  const kindLabel = { deployment: 'Deployment', statefulset: 'StatefulSet', daemonset: 'DaemonSet' }[kind]

  const handleRestart = async (
    namespace: string,
    name: string,
    onSuccess?: () => void
  ) => {
    try {
      await ElMessageBox.confirm(
        `确定重启 ${kindLabel} "${name}"？这将触发所有 Pod 滚动更新。`,
        '重启确认',
        { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
      )
      await restartApi({ namespace, name })
      ElMessage.success('重启指令已发送')
      onSuccess?.()
    } catch (e: any) {
      if (e !== 'cancel') {
        ElMessage.error(e?.response?.data?.message || '重启失败')
      }
    }
  }

  return { handleRestart }
}
```

#### 3c. 新建 `src/composables/useEditDrawer.ts`

**依据**: 所有 Detail 页面都有相同的编辑抽屉状态管理逻辑。

```ts
import { ref } from 'vue'

export function useEditDrawer(fetchDetail: () => Promise<void>) {
  const editDialogVisible = ref(false)
  const editFullscreen = ref(false)

  const handleEdit = () => { editDialogVisible.value = true }
  const handleEditSuccess = () => { editDialogVisible.value = false; fetchDetail() }
  const handleEditCancel = () => { editDialogVisible.value = false }

  return { editDialogVisible, editFullscreen, handleEdit, handleEditSuccess, handleEditCancel }
}
```

---

### Step 4: 创建 Detail 页面共享组件

#### 4a. 新建 `src/components/DetailPageLayout.vue`

**依据**: 7 个 Detail 页面的 `<style scoped>` 块中 `.detail-page`、`.main-layout`、`.left-panel`、`.right-panel`、`.resize-handle-h`、`.resize-handle-v` 类名完全相同。

**职责**: 提供左右可拖拽分栏布局容器。垂直 resize 为可选功能。

```vue
<template>
  <div class="detail-page">
    <!-- #default slot: 放 DetailPageHeader 等页面头部内容，在 main-layout 之外 -->
    <slot />
    <div class="main-layout">
      <div class="left-panel" :style="{ width: leftWidth + 'px' }">
        <slot name="left" />
      </div>
      <div class="resize-handle-h" @mousedown="onResizeHStart" />
      <div class="right-panel">
        <!-- 如果启用了垂直 resize，分为上下两部分 -->
        <template v-if="resizable">
          <div class="right-top" :style="{ height: topHeight + 'px' }">
            <slot name="right-top" />
          </div>
          <div class="resize-handle-v" @mousedown="onResizeVStart" />
          <div class="right-bottom">
            <slot name="right-bottom" />
          </div>
        </template>
        <!-- 否则整个右侧面板作为一个 slot -->
        <template v-else>
          <slot name="right" />
        </template>
      </div>
    </div>
  </div>
</template>
```

**Props**:
```ts
interface Props {
  resizable?: boolean       // 是否启用垂直 resize，默认 false
  initialLeftWidth?: number // 左侧初始宽度，默认 400
  initialTopHeight?: number // 右上初始高度，默认 300
  minLeftWidth?: number     // 左侧最小宽度，默认 300
  minTopHeight?: number     // 右上最小高度，默认 150
}
```

**script**: 内部使用修复后的 `useResizable`（Step 2）。当 `resizable=false` 时只初始化水平 resize。

**style**: 包含 `.detail-page`、`.main-layout`、`.left-panel`、`.right-panel`、`.resize-handle-h`、`.resize-handle-v` 等布局 CSS，约 80 行。

#### 4b. 新建 `src/components/DetailPageHeader.vue`

**依据**: 所有 Detail 页面的头部结构：左侧（标题 + 状态标签 + 命名空间），右侧（操作按钮 + 自动刷新 + 刷新 + 返回）。

**Props**:
```ts
interface Props {
  title: string
  statusTag?: { text: string; type: 'success' | 'warning' | 'danger' | 'info' }
  namespace?: string
}
```

**Slots**: `actions`（放操作按钮）、`extra`（放自动刷新等额外内容）

**style**: `.page-header`、`.header-left`、`.header-actions` 等，约 30 行。

#### 4c. 新建 `src/components/EventsTable.vue`

**已验证的事件字段结构**:

`useDetailPage.ts:63-64` 直接透传后端返回，不做字段转换。但各 Detail 页面使用的事件时间字段名不统一：
- `DeploymentDetail.vue:505`: `prop="lastTimestamp"`
- `DaemonSetDetail.vue:601`: `prop="last_seen"`
- `ReplicaSetDetail.vue:312`: `prop="last_seen"`

**执行指令**: EventsTable 组件需要兼容两种时间字段名。通过 prop 或 computed 统一：
```ts
interface Props {
  events: any[]
  loading?: boolean
  timeField?: string  // 默认 'lastTimestamp'，DaemonSet/ReplicaSet 传 'last_seen'
}
```
或在组件内用 computed 兼容：`row.lastTimestamp || row.last_seen`

**模板**: 约 15 行，el-table + 4 列（Type / Reason / Message / Age）。Age 列使用 `formatAge` 工具函数（已确认 `src/utils/helpers.ts` 中存在）。

#### 4d. 新建 `src/components/RevisionList.vue`

**已验证的数据结构兼容性 — 不兼容，需要分别处理**:

- **Deployment** 的 revisions 是原生 ReplicaSet 对象（API `/k8s/deployment/replicasets`）：
  - `metadata.name`, `metadata.annotations['deployment.kubernetes.io/revision']`, `spec.template.spec.containers[0].image`, `status.readyReplicas`, `metadata.creationTimestamp`
  - 通过 `pod-template-hash` label 关联 Pod
- **DaemonSet/StatefulSet** 的 revisions 是后端自定义对象（API `/k8s/daemonset/rollbacks`）：
  - `name`, `revision`(number), `isCurrent`(boolean), `images[]`(string array), `createdAt`(string)
  - 通过 `controller-revision-hash` label 关联 Pod

**执行指令**: RevisionList 需要接收统一的 display-model，由各 Detail 页面在调用前转换：

```ts
// 统一的修订显示模型（各 Detail 页面将原始数据转为此格式后再传入）
interface RevisionItem {
  name: string
  revision: string | number
  images: string[]          // 容器镜像列表
  createdAt: string
  podCount: number          // 关联的 Pod 数量
  isCurrent: boolean        // 是否为当前修订
}

interface Props {
  revisions: RevisionItem[]
  selectedRevision: string | null
  loading?: boolean
}
```

各 Detail 页面负责将原始数据 map 为 `RevisionItem[]`：
- DeploymentDetail: `replicasets.map(rs => ({ name: rs.metadata.name, revision: rs.metadata.annotations?.['deployment.kubernetes.io/revision'], images: [...], ... }))`
- DaemonSetDetail/StatefulSetDetail: `rollbacks.map(rb => ({ name: rb.name, revision: rb.revision, images: rb.images, ... }))`

**Events**: `select`(选中修订)、`rollback`(回滚到指定修订)

#### 4e. 新建 `src/components/LabelsBlock.vue`、`ConditionsBlock.vue`、`SelectorBlock.vue`

- `LabelsBlock.vue`: 接收 `labels: Record<string, string>`，渲染 el-tag 列表
- `ConditionsBlock.vue`: 接收 `conditions: K8sCondition[]`，渲染 el-table
- `SelectorBlock.vue`: 接收 `selector: Record<string, string>`，渲染 el-tag 列表

---

### Step 5: 重构 Detail 页面（用新组件替换重复代码）

**对以下 9 个文件执行重构**：

- `src/views/workload/DeploymentDetail.vue`（1026 行 → ~350 行）
- `src/views/workload/DaemonSetDetail.vue`（1245 行 → ~400 行，节点分布是独有的）
- `src/views/workload/StatefulSetDetail.vue`（1018 行 → ~350 行）
- `src/views/workload/PodDetail.vue`（703 行 → ~250 行）
- `src/views/workload/JobDetail.vue`（695 行 → ~250 行）
- `src/views/workload/CronJobDetail.vue`（730 行 → ~280 行）
- `src/views/workload/ReplicaSetDetail.vue`（653 行 → ~250 行）
- `src/views/network/ServiceDetail.vue`（816 行 → ~300 行）
- `src/views/network/IngressDetail.vue`（783 行 → ~280 行）

**重构模式**（以 DeploymentDetail.vue 为例）:

**执行前必须做的事**:
1. 读取 `PodListPanel.vue` 的 `emit` 定义，确认 `delete` 事件的参数签名是 `(pod)` 还是 `(pod, force)`。如果只有 `(pod)`，需扩展 PodListPanel 的 emit 或在调用处适配。
2. 读取 `useDetailPage.ts` 确认返回值中是否包含 `fetchEvents`。如不包含，需扩展其返回值。
3. 用 `grep` 检查当前 Detail 页面的 `<style scoped>` 中哪些 class 是页面特有的（不在 DetailPageLayout 中），特有样式保留。

**操作步骤**:

1. **import 替换**:
   ```ts
   // 删除
   - function handlePodLogs(pod: any) { ... }        // :170-173
   - function handlePodExec(pod: any) { ... }        // :175-178
   - async function handlePodDelete(...) { ... }     // :180-216
   - async function handleRestart() { ... }           // :225-239
   - const editDialogVisible = ref(false)             // 编辑相关
   - const editFullscreen = ref(false)
   - function handleEdit() { ... }
   - function handleEditSuccess() { ... }
   - function handleEditCancel() { ... }

   // 添加
   + import DetailPageLayout from '@/components/DetailPageLayout.vue'
   + import DetailPageHeader from '@/components/DetailPageHeader.vue'
   + import EventsTable from '@/components/EventsTable.vue'
   + import RevisionList from '@/components/RevisionList.vue'
   + import LabelsBlock from '@/components/LabelsBlock.vue'
   + import ConditionsBlock from '@/components/ConditionsBlock.vue'
   + import SelectorBlock from '@/components/SelectorBlock.vue'
   + import { usePodActions } from '@/composables/usePodActions'
   + import { useRestartAction } from '@/composables/useRestartAction'
   + import { useEditDrawer } from '@/composables/useEditDrawer'
   ```

2. **script setup 中**（注意调用顺序，composable 依赖 useDetailPage 的返回值）:
   ```ts
   // 1. 先调用 useDetailPage
   const { namespace, name, loading, detail, events, fetchDetail, fetchEvents, handleDelete } = useDetailPage({ ... })

   // 2. 再调用依赖 fetchDetail 的 composable
   const clusterName = computed(() => clusterStore.clusterName)
   const { handlePodLogs, handlePodExec, handlePodDelete } = usePodActions(clusterName)
   const { handleRestart } = useRestartAction('deployment', restartDeployment)
   const { editDialogVisible, editFullscreen, handleEdit, handleEditSuccess, handleEditCancel } = useEditDrawer(fetchDetail)

   // 3. handleYamlSaved 保留为内联（两行逻辑，不值得单独提取）
   const handleYamlSaved = () => { fetchDetail(); fetchEvents() }
   ```

3. **template 中**:
   ```vue
   <DetailPageLayout :resizable="true">
     <!-- 头部 -->
     <DetailPageHeader :title="detail?.name" :status-tag="statusTag" :namespace="detail?.namespace">
       <template #actions>
         <el-button @click="handleScale">伸缩</el-button>
         <el-button @click="handleRestart(namespace, name, fetchAllPods)">重启</el-button>
         <!-- ... 其他按钮 -->
       </template>
       <template #extra>
         <!-- 自动刷新 popover -->
       </template>
     </DetailPageHeader>

     <template #left>
       <el-segmented v-model="leftView" :options="['ReplicaSet', '基本信息']" />
       <RevisionList v-if="leftView === 'ReplicaSet'" ... />
       <div v-else>
         <el-descriptions>...</el-descriptions>
         <ConditionsBlock :conditions="detail?.conditions" />
         <SelectorBlock :selector="detail?.selector" />
         <LabelsBlock :labels="detail?.labels" />
       </div>
     </template>

     <template #right-top>
       <!-- PodListPanel 的事件签名需先验证 -->
       <PodListPanel :pods="allPods" @logs="handlePodLogs" @exec="handlePodExec" @delete="handlePodDelete" />
     </template>

     <template #right-bottom>
       <EventsTable :events="events" />
     </template>
   </DetailPageLayout>
   ```

4. **style 中**:
   - 删除整个 `<style scoped>` 块
   - 保留页面特有的样式（通过 grep 确认哪些 class 不在 DetailPageLayout 中）

**各页面差异说明**:
- **DaemonSetDetail**: 独有"节点分布"标签页，该 section 的模板和样式保留为内联
- **PodDetail**: 无修订历史、无伸缩/重启/更新镜像，右侧面板是容器表格（有展开行），使用 `resizable=false`（无上下分栏需求）
- **ServiceDetail / IngressDetail**: 不改用 `useDetailPage`，只复用 DetailPageLayout + DetailPageHeader + EventsTable 消除布局/模板/CSS 重复。`handleDelete` 保留为内联。
- **JobDetail / CronJobDetail / ReplicaSetDetail**: 结构与 DeploymentDetail 类似，但无伸缩/自动伸缩

---

### Step 6: 创建 Form 共享模块

#### 6a. 新建 `src/views/workload/components/form-types.ts`

**执行指令**: AI 必须先读取 WorkloadForm.vue:33-75 获取完整的接口定义，完整复制每个 interface 的所有字段，不要用注释占位。

**内容**: 完整定义以下类型（从源码复制）：
- `Label`, `Port`, `EnvVar`, `Resources`, `VolumeMount`, `Probe`, `LifecycleHandler`, `Volume`, `Tolerance`, `Annotation`, `VolumeClaimTemplate`, `AffinityRule`, `TopologySpreadConstraint`, `Container`
- `BaseFormData` — 所有 Form 共有的字段（name, namespace, serviceAccountName, labels, annotations, containers, initContainers, volumes, volumeMounts, tolerations, affinityRules, topologySpreadConstraints, nodeSelector, terminationGracePeriodSeconds, imagePullSecrets, securityContext）
- 工厂函数：`createEmptyEnv()`, `createEmptyContainer()`, `createEmptyLifecycleHandler()`, `createEmptyProbe()`

各 Form 的 `FormData` 继承 `BaseFormData` 并扩展特有字段：
```ts
// WorkloadForm — 特有字段
interface WorkloadFormData extends BaseFormData {
  replicas: number
  strategyType: string
  maxSurge: string
  maxUnavailable: string
  // ... 其他特有字段从源码复制
}

// CronJobForm — 特有字段
interface CronJobFormData extends BaseFormData {
  schedule: string
  concurrencyPolicy: string
  // ... 其他特有字段从源码复制
}

// JobForm — 特有字段
interface JobFormData extends BaseFormData {
  completions: number
  parallelism: number
  // ... 其他特有字段从源码复制
}
```

#### 6b. 新建 `src/views/workload/components/useFormArrays.ts`

**依据**: 3 个 Form 中 ~30 个 add/remove 函数完全相同。

**接口**:
```ts
import { BaseFormData, Container } from './form-types'

// 直接接收 reactive 对象，不用 Reactive<BaseFormData>（避免类型导入问题）
export function useFormArrays(formData: any) {
  const addLabel = () => formData.labels.push({ key: '', value: '' })
  const removeLabel = (index: number) => formData.labels.splice(index, 1)
  const addContainer = () => formData.containers.push(createEmptyContainer())
  const removeContainer = (index: number) => formData.containers.splice(index, 1)
  const addPort = (container: Container) => container.ports.push({ name: '', containerPort: 0, protocol: 'TCP' })
  // ... 其他 24+ 个函数从任一 Form 文件中完整复制
  return { addLabel, removeLabel, addContainer, removeContainer, /* ... */ }
}
```

#### 6c. 新建 `src/views/workload/components/useFormParsing.ts`

**内容**: 提取 K8s 对象 → 表单数据的解析函数。`parseInitialData` 各 Form 不同（CronJob 需要多穿透一层 `spec.jobTemplate.spec.template.spec`），不在此提取。

#### 6d. 新建 `src/views/workload/components/useFormBuild.ts`

**内容**: 提取表单数据 → K8s 资源的构建工具函数。`buildK8sResource` 各 Form 不同（不同 kind），不在此提取。

---

### Step 7: 拆分 Form 模板为子组件

#### 7a. `src/views/workload/components/ContainerConfigForm.vue`

**执行指令**: AI 必须先读取 WorkloadForm.vue template 中容器配置 section（约第 718-822 行），提取完整的模板和 script 逻辑。

**Props**:
```ts
interface Props {
  containers: Container[]
  title?: string   // "容器配置" / "初始化容器"
}
```

**Events**: `update:containers`

**职责**: 容器卡片列表，每个卡片包含：名称、镜像、拉取策略、命令、参数、端口、环境变量、资源限制。

**注意**: 如果组件本身过大（>300 行），考虑进一步拆分为 `ContainerCard.vue`（单个容器卡片）+ `ContainerConfigForm.vue`（列表管理）。

#### 7b. `src/views/workload/components/StorageConfigForm.vue`

**Props**: `volumes: Volume[]`, `volumeMounts: VolumeMount[]`, `volumeClaimTemplates?: VolumeClaimTemplate[]`, `kind?: string`

#### 7c. `src/views/workload/components/HealthCheckForm.vue`

**Props**: `containers: Container[]`

#### 7d. `src/views/workload/components/SecurityContextForm.vue`

**Props**: `securityContext: Record<string, any>`, `containers: Container[]`

#### 7e. `src/views/workload/components/SchedulingConfigForm.vue`

**Props**: `nodeSelector: Record<string, string>`, `tolerations: Tolerance[]`, `affinityRules: AffinityRule[]`, `topologySpreadConstraints: TopologySpreadConstraint[]`

#### 7f. `src/views/workload/components/LabelsAnnotationsForm.vue`

**Props**: `labels: Label[]`, `annotations: Annotation[]`
**Events**: `update:labels`, `update:annotations`

---

### Step 8: 重构 Form 组件

**对以下 3 个文件执行重构**:

- `src/views/workload/components/WorkloadForm.vue`（1642 行 → ~400 行）
- `src/views/workload/components/CronJobForm.vue`（1696 行 → ~400 行）
- `src/views/workload/components/JobForm.vue`（1583 行 → ~400 行）

**重构模式**:

1. 删除所有已提取到 `form-types.ts` 的接口定义
2. 删除所有已提取到 `useFormArrays.ts` 的 add/remove 函数
3. 删除所有已提取到 `useFormParsing.ts` 的 parse 函数
4. 删除所有已提取到 `useFormBuild.ts` 的 build 函数
5. template 中替换为子组件：
   ```vue
   <!-- 标签注解: 替换 -->
   <LabelsAnnotationsForm v-model:labels="formData.labels" v-model:annotations="formData.annotations" />

   <!-- 容器配置: 替换 -->
   <ContainerConfigForm v-model:containers="formData.containers" title="容器配置" />
   <ContainerConfigForm v-model:containers="formData.initContainers" title="初始化容器" />

   <!-- 存储配置: 替换 -->
   <StorageConfigForm v-model:volumes="formData.volumes" v-model:volumeMounts="formData.volumeMounts"
     :volume-claim-templates="formData.volumeClaimTemplates" :kind="kind" />

   <!-- 健康检查: 替换 -->
   <HealthCheckForm :containers="formData.containers" />

   <!-- 安全设置: 替换 -->
   <SecurityContextForm v-model="formData.securityContext" :containers="formData.containers" />

   <!-- 调度配置: 替换 -->
   <SchedulingConfigForm v-model:node-selector="formData.nodeSelector"
     v-model:tolerations="formData.tolerations" v-model:affinity-rules="formData.affinityRules"
     v-model:topology-spread-constraints="formData.topologySpreadConstraints" />
   ```
6. **保留为内联**（不提取为子组件）：
   - 基本信息 section — 各 Form 字段差异大（WorkloadForm 有 replicas/strategy，CronJobForm 有 schedule，JobForm 有 completions）
   - `parseInitialData` 函数 — 各 Form 穿透 K8s 路径不同
   - `buildK8sResource` 函数 — 各 Form 生成不同 kind
   - submit handler — 调用不同的 create/update API
   - kind 特有的高级配置 section — WorkloadForm 有 DNS/priority/hostNetwork

---

### Step 9: 改造 API 工厂并统一使用

#### 9a. 改造 `src/api/factory.ts` 支持泛型

**执行指令**: AI 必须先确认后端 DELETE 接口的参数传递方式（query params vs request body），再决定 `delete` 方法的实现。

**改为**:
```ts
export function createResourceApi<TList = any, TDetail = any>(basePath: string) {
  return {
    list: (data?: any) => request.get<TList>(`${basePath}`, { params: data }),
    detail: (params: { namespace: string; name: string }) =>
      request.get<TDetail>(`${basePath}/${params.namespace}/${params.name}`),
    getYaml: (params: { namespace: string; name: string }) =>
      request.get<string>(`${basePath}/${params.namespace}/${params.name}/yaml`),
    create: (data: any) => request.post(basePath, data),
    updateYaml: (params: { namespace: string; name: string }, data: any) =>
      request.put(`${basePath}/${params.namespace}/${params.name}/yaml`, data),
    // delete 的参数传递方式需根据后端实际接口决定
    delete: (params: { namespace: string; name: string }) =>
      request.delete(`${basePath}/${params.namespace}/${params.name}`, { data: params }),
    events: (params: { namespace: string; name: string }) =>
      request.get(`${basePath}/${params.namespace}/${params.name}/events`),
  }
}
```

#### 9b. 改造 `src/api/workload.ts` 使用工厂（分三步）

**第一步（兼容）**: 用工厂生成 API 对象，保留旧的函数别名
```ts
export const deploymentApi = createResourceApi<any, K8sDeployment>('/v1/k8s/deployment')

/** @deprecated 使用 deploymentApi.list() */
export const getDeploymentList = deploymentApi.list
/** @deprecated 使用 deploymentApi.detail() */
export const getDeploymentDetail = deploymentApi.detail
// ...
```

**第二步（迁移）**: 先用 grep 找出所有 import 位置：
```bash
grep -rn "getDeploymentList\|getDeploymentDetail\|getDeploymentYaml\|createDeployment\|updateDeploymentYaml\|deleteDeployment\|getDeploymentEvents" src/views/ src/api/
grep -rn "getStatefulSetList\|getStatefulSetDetail\|..." src/views/ src/api/
# 对每种资源类型重复
```
然后逐文件修改 import，每改一个文件就 `npm run build` 验证编译通过。

**第三步（清理）**: 所有 import 迁移完成后，删除旧的函数别名。

**同样改造**: `src/api/network.ts`, `src/api/storage.ts`, `src/api/config.ts`

#### 9c. 拆分 `src/api/core.ts`（302 行）

拆分为：
- `src/api/namespace.ts` — namespace 相关 API + `extractNamespaceNames`
- `src/api/node.ts` — node 相关 API + transform
- `src/api/event.ts` — event 相关 API
- `src/api/crd.ts` — CRD 相关 API

保留 `src/api/core.ts` 作为 barrel re-export（兼容现有 import）：
```ts
export * from './namespace'
export * from './node'
export * from './event'
export * from './crd'
```

---

### Step 10: 杂项修复

#### 10a. `src/composables/useResourceList.ts` 复用 `useAutoRefresh`

**当前**: 第 391-406 行自行实现 `setInterval`。
**改为**: import 并使用 `useAutoRefresh` composable。

#### 10b. `src/components/YamlDrawer.vue` 去除硬编码

**当前**: 内部维护 `resourceApis` 映射表（24 种资源类型硬编码）。
**改为**: 通过 prop 传入 API 函数，或接收 resource type 字符串 + `createResourceApi` 动态生成。

**注意**: 此改动是破坏性变更，所有调用处都依赖内部 API 映射。应与 Step 9b 第二步同步执行——在迁移 import 的同时改造 YamlDrawer 的调用方式。

#### 10c. `src/router/index.ts` 修复安全漏洞

**当前**（第 597 行）: `authStore.user` 为 `null` 时条件短路放行。
**改为**:
```ts
if (!authStore.user) {
  return next({ name: 'login', query: { redirect: to.fullPath } })
}
if (!authStore.user.isSuperAdmin && !authStore.user.isAdmin) {
  return next({ name: 'dashboard' })
}
```

#### 10d. `src/stores/cluster.ts` 添加防御

**当前**（第 48 行）: `res.data.items` 假设格式。
**改为**: `clusterList.value = res.data?.items ?? []`

---

## 三、依赖关系和执行顺序

```
Step 1 (K8s 类型) ────────────────────────────┐
Step 2 (修复 useResizable) ───────────────────┤ 无依赖，并行
Step 10c (路由安全修复) ──────────────────────┤
Step 10d (cluster store 防御) ────────────────┘
         │
Step 3 (composable: usePodActions/useRestartAction/useEditDrawer) ← 无依赖
Step 4 (共享组件: DetailPageLayout/Header/EventsTable/RevisionList/LabelsBlock/ConditionsBlock/SelectorBlock) ← 依赖 Step 2
         │
Step 5 (重构 Detail 页面) ← 依赖 Step 3 + 4
         │
Step 6 (form-types/useFormArrays/useFormParsing/useFormBuild) ← 依赖 Step 1
Step 7 (Form 子组件) ← 依赖 Step 6
         │
Step 8 (重构 Form 组件) ← 依赖 Step 6 + 7
         │
Step 9 (API 工厂改造，分三步) ← 依赖 Step 1
         │
Step 10a, 10b (剩余杂项) ← 10b 与 Step 9b 同步执行
```

**推荐执行顺序**:
1. Step 1 + Step 2 + Step 10c + Step 10d（并行，无依赖）
2. Step 3 + Step 4（并行）
3. Step 5（重构 Detail 页面）
4. Step 6 + Step 7（并行）
5. Step 8（重构 Form）
6. Step 9（API 工厂，分三步执行，10b 与 9b 第二步同步）
7. Step 10a

---

## 四、验证方式

每完成一个 Step 后：
1. `cd /data/gkube/frontend && npm run build` — TypeScript 编译通过
2. `grep -r ": any" src/ | wc -l` — 确认 any 数量下降
3. `wc -l src/views/workload/*Detail.vue` — 确认 Detail 行数下降
4. `wc -l src/views/workload/components/*Form.vue` — 确认 Form 行数下降
5. `wc -l src/composables/*.ts` — 确认 composable 行数合理增长
6. `wc -l src/components/*.vue` — 确认共享组件行数合理增长

---

## 五、风险提示

1. **Step 5 中各 Detail 页面差异大**: DaemonSet 有独有的"节点分布"标签页、PodDetail 有容器展开行、ServiceDetail 有 endpoints 标签页，不能完全套用同一模板
2. **Step 5 删除 scoped CSS 前**: 必须用 grep 检查哪些 class 是页面特有的，特有样式保留
3. **Step 5 RevisionList 数据结构不兼容**: Deployment 的 revisions 是原生 ReplicaSet，DaemonSet/StatefulSet 是后端自定义 rollbacks 对象。各 Detail 页面需先转换为统一的 `RevisionItem[]` 再传入组件
4. **Step 5 EventsTable 时间字段不统一**: Deployment 用 `lastTimestamp`，DaemonSet/ReplicaSet 用 `last_seen`。组件内需兼容两种字段名
5. **Step 8 Form 子组件的 v-model 绑定**: Vue 3 的 `v-model:prop` 语法需要子组件 emit `update:prop` 事件，子组件内部需要正确实现
6. **Step 9b 迁移工作量**: 所有 import 从函数别名改为 API 对象，涉及所有 Detail/List 页面，需逐文件验证
7. **Step 9a DELETE API 签名**: 后端 DELETE 接口可能不支持 request body，需先确认
8. **Step 10b YamlDrawer 改造**: 破坏性变更，需与 Step 9b 同步执行
9. **Step 3a usePodActions**: `deletePod` 从 `@/api/resource` 导入（不是 `@/api/workload`），强制删除通过 `deletePod({force:true})` 实现（不存在 `forceDeletePod` 函数）

---

## 六、Phase B: i18n 补全 + 错误处理统一 + List 页面去重

> Phase A（Step 1-10）完成后执行。解决正确性和可维护性问题。

### Step 11: 修复 en.ts 翻译缺失和 typo

**修改文件**: `src/locales/en.ts`

**问题**:
- en.ts 缺少 `search` section（zh-CN.ts:877-893 有定义，en.ts 在 852 行结束）
- en.ts 行 235: `noPodssFound` → 应为 `noPodsFound`（多了一个 s）
- en.ts 行 237: `noPodssOnNode` → 应为 `noPodsOnNode`
- en.ts 行 318: `allPodss` → 应为 `allPods`
- en.ts sidebar section 缺少 `search: 'Resource Search'`

**执行指令**:
1. 读取 zh-CN.ts 的 `search` section（:877-893），翻译为英文写入 en.ts
2. 修正 3 处 typo（235/237/318 行的 `Podss` → `Pods`）
3. 在 en.ts sidebar section 添加 `search: 'Resource Search'`

### Step 12: 统一错误处理策略

**问题**: 三种错误处理模式并存：
- 模式 A: `console.error` — 用户看不到（DashboardView:149, DeploymentDetail:107, ReplicaSetDetail:73, LogView:80）
- 模式 B: `console.warn` — 用户看不到（SecretList:59）
- 模式 C: `ElMessage.error` 硬编码中文 — 英文模式下仍显示中文（334 处中 264 处）

**修正策略**:
1. 所有面向用户的错误统一使用 `ElMessage.error`，不使用 `console.error`/`console.warn`
2. 保留 `console.error` 仅用于开发调试（在 ElMessage 之后追加 console.error）
3. 所有 `ElMessage` 文本走 i18n（使用 `t()` 函数）

**执行指令**:

**Step 12a**: 修复空 catch 和 console.error/warn
```bash
# 找出所有 console.error 位置
grep -rn "console\.error\|console\.warn" src/views/ src/composables/ src/stores/
```

逐个修改：
- `stores/auth.ts:36` — `catch { // ignore }` → 保持（JSON.parse 恢复 localStorage，静默合理）
- `stores/auth.ts:83-84` — `catch { user.value = { ... } }` → 保持（permissions fetch 失败降级合理）
- `stores/cluster.ts:22` — `catch { return null }` → 保持（cluster list fetch 失败降级合理）
- `stores/namespace.ts:22` — `catch { return namespaces.value }` → 保持（缓存降级合理）
- `DashboardView.vue:149,157,163` — 改为 `ElMessage.error(t('dashboard.fetchFailed'))` + 保留 console.error
- `DeploymentDetail.vue:107,133` — 改为 `ElMessage.error(...)` + 保留 console.error
- `ReplicaSetDetail.vue:73` — 同上
- `LogView.vue:80` — 同上
- `SecretList.vue:59` — `console.warn(...)` → `ElMessage.error(t('secret.fetchFailed'))`

**Step 12b**: 将硬编码中文的 ElMessage 走 i18n
```bash
# 找出所有包含中文字符的 ElMessage 调用（更精确，不会误排除已用 i18n 的行）
grep -rn "ElMessage\.\(error\|success\|warning\)" src/views/ src/composables/ | grep "[\x{4e00}-\x{9fff}]"
```

逐文件修改，模式：
```ts
// 改前
ElMessage.error(e?.message || '扩缩容失败')
// 改后
ElMessage.error(e?.message || t('workload.scaleFailed'))
```

需要在 zh-CN.ts 和 en.ts 中新增对应的翻译 key。预计新增约 50-80 个翻译 key。

**Step 12c**: 修复 `request.ts` 中所有硬编码中文的 ElMessage
```bash
# 先找出 request.ts 中所有 ElMessage 调用
grep -n "ElMessage" src/api/request.ts
```

逐个修改，模式：
```ts
// 改前
ElMessage.error(error.response?.data?.msg || '权限不足')
// 改后
ElMessage.error(error.response?.data?.msg || t('common.forbidden'))
```

注意：`request.ts` 中不能直接使用 `useI18n()`（非 Vue 组件上下文），需要导入 i18n 实例：
```ts
import i18n from '@/locales'
const { t } = i18n.global
```

### Step 13: List 页面 CSS 去重

**问题**: 20+ 个 List 页面用 scoped CSS 重复定义 `styles/index.css` 中已有的全局工具类。

**执行指令**:
```bash
# 找出所有重复定义 .page-container 的文件
grep -rn "\.page-container" src/views/ | grep "scoped"
```

对每个 List 页面：
1. 删除 `<style scoped>` 中的 `.page-container`、`.table-card`、`.action-buttons`、`.load-more`、`.pagination-wrapper` 定义
2. 保留页面特有的 scoped 样式
3. 删除非 scoped `<style>` 中的 `.yaml-drawer .el-drawer__header` 重复定义（约 16 个文件），统一放到 `styles/index.css`

**注意**: 删除前确认全局样式优先级足够（scoped 样式优先级高于全局样式，删除后可能需要调整选择器）。

### Step 14: List 页面 scale/restart/image dialog 去重

**问题**: DeploymentList:68-149、StatefulSetList:68-144、DaemonSetList:67-119 中 scale/restart/image dialog 逻辑逐字复制。

**执行指令**:

**Step 14a**: 新建 `src/composables/useListActions.ts`

提取 List 页面中重复的操作逻辑：

**执行指令**: AI 必须先读取 DeploymentList.vue:68-149、StatefulSetList.vue:68-144、DaemonSetList.vue:67-119 获取完整的 dialog 逻辑，再提取 composable。

```ts
import { ref } from 'vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import {
  restartDeployment, restartStatefulSet, restartDaemonSet,
  getDeploymentDetail, getStatefulSetDetail, getDaemonSetDetail
} from '@/api/workload'

export function useListActions(kind: 'deployment' | 'statefulset' | 'daemonset') {
  // DaemonSet 不支持伸缩
  const canScale = kind !== 'daemonset'

  // scale dialog — 仅 deployment/statefulset
  const scaleDialogVisible = ref(false)
  const scaleTarget = ref<any>(null)
  const openScaleDialog = canScale
    ? (row: any) => { scaleTarget.value = row; scaleDialogVisible.value = true }
    : undefined

  // restart — 全部支持
  const restartApiMap = { deployment: restartDeployment, statefulset: restartStatefulSet, daemonset: restartDaemonSet }
  const handleRestart = async (row: any, onSuccess?: () => void) => {
    try {
      await ElMessageBox.confirm(`确定重启 ${row.name}？`, '重启确认', { type: 'warning' })
      await restartApiMap[kind]({ namespace: row.namespace, name: row.name })
      ElMessage.success('重启指令已发送')
      onSuccess?.()
    } catch (e: any) {
      if (e !== 'cancel') ElMessage.error(e?.response?.data?.message || '重启失败')
    }
  }

  // image update — 全部支持
  const imageDialogVisible = ref(false)
  const imageTarget = ref<any>(null)
  const detailApiMap = { deployment: getDeploymentDetail, statefulset: getStatefulSetDetail, daemonset: getDaemonSetDetail }
  const openImageDialog = async (row: any) => {
    const detail = await detailApiMap[kind]({ namespace: row.namespace, name: row.name })
    imageTarget.value = detail
    imageDialogVisible.value = true
  }

  return { canScale, scaleDialogVisible, scaleTarget, openScaleDialog, handleRestart, imageDialogVisible, imageTarget, openImageDialog }
}
```

**Step 14b**: 重构 DeploymentList、StatefulSetList、DaemonSetList 使用 `useListActions`

删除各文件中重复的 dialog 状态和处理函数，替换为 composable 调用。

### Step 15: 统一 status 映射函数

**已验证**: `statusType` 和 `getPodStatusType` **不重叠**，输入域不同：
- `getPodStatusType`（`utils/pod.ts:13`）: 专用于 Pod phase 映射（running→success, succeeded→info, pending→warning, failed/error→danger）
- `statusType`（`utils/helpers.ts:8`）: 通用状态映射（online/connected→success, offline/disconnected→danger 等）

两者只是输出类型相同（Element Plus tag type），实际处理的状态集合完全不同。

**修正**: 不合并这两个函数。`getPodStatusType` 职责清晰（Pod 专用），`statusType` 是通用的。保持现状。

但如果 `utils/pod.ts` 只有 `getPodStatusType` 一个函数且很短（37 行），可以考虑将其移入 `utils/helpers.ts` 减少文件数量，但这不是必须的。

### Step 16: 统一 base64 编解码

**问题**: `utils/helpers.ts:49-69` 有 UTF-8 安全的实现，`SecretList.vue:66` 内联了废弃 API 的实现。

**执行指令**:
1. 读取 `SecretList.vue:66` 附近的 base64 编解码代码
2. 替换为 `import { base64Encode, base64Decode } from '@/utils/helpers'`
3. 删除内联实现

### Step 17: 统一 age 格式化入口

**问题**: `formatAge` 有三个入口：`utils/helpers.ts`（实际实现）、`utils/time.ts`（转发）、`api/core.ts` 的 `calcAge`（委托）。

**执行指令**:
1. 删除 `utils/time.ts`（纯转发文件）
2. 将所有 `import { formatAge } from '@/utils/time'` 改为 `import { formatAge } from '@/utils/helpers'`
3. 将 `api/core.ts` 中的 `calcAge` 改为直接 `export { formatAge }` 或删除并在消费处直接用 `formatAge`
```bash
grep -rn "from '@/utils/time'" src/
grep -rn "calcAge" src/
```

---

## 七、Phase C: 路由模块化 + 构建优化 + List 页面统一

> Phase B 完成后执行。

### Step 18: 路由模块化拆分

**问题**: `router/index.ts` 627 行全内联，90+ 条路由。

**执行指令**:

**Step 18a**: 按 K8s API group 拆分路由文件
```
src/router/
  index.ts          — 主路由 + 全局守卫（保留 ~100 行）
  workload.ts       — Deployment/StatefulSet/DaemonSet/Job/CronJob/Pod/ReplicaSet 路由
  network.ts        — Service/Ingress/NetworkPolicy 路由
  storage.ts        — PV/PVC/StorageClass/VolumeSnapshot 路由
  config.ts         — ConfigMap/Secret/ResourceQuota/LimitRange 路由
  node.ts           — Node 路由
  cluster.ts        — Cluster 管理路由
  event.ts          — Event 路由
```

每个子路由文件导出一个 `RouteRecordRaw[]` 数组，`index.ts` 中合并：
```ts
import { workloadRoutes } from './workload'
import { networkRoutes } from './network'
// ...
const routes: RouteRecordRaw[] = [
  { path: '/login', ... },
  { path: '/fullscreen/...', ... },
  { path: '/', component: AppLayout, children: [
    ...workloadRoutes,
    ...networkRoutes,
    ...storageRoutes,
    ...configRoutes,
    ...nodeRoutes,
    ...clusterRoutes,
    ...eventRoutes,
    { path: ':pathMatch(.*)*', ... }  // 404，保持在 AppLayout 内（渲染侧边栏）
  ]}
]
```

**注意**: 404 路由保持在 AppLayout children 中（当前行为不变）。如需独立 404 页面（不渲染侧边栏），作为后续优化项。

**Step 18b**: 路由 meta.title 改用 i18n key

1. 创建或修改路由 meta 类型扩展（`src/router/typings.ts`）：
```ts
import 'vue-router'

declare module 'vue-router' {
  interface RouteMeta {
    titleKey?: string      // i18n key，替代原 title 硬编码中文
    title?: string          // 保留兼容，逐步迁移到 titleKey
    parent?: string
    requireAdmin?: boolean
  }
}
```

2. 逐文件将 `meta.title` 改为 `meta.titleKey`：
```ts
// 改前
meta: { title: 'Deployment详情' }
// 改后
meta: { titleKey: 'workload.deploymentDetail' }
```

3. 修改 `router/index.ts:622-25` 的 `afterEach` 中 `document.title` 拼接逻辑：
```ts
import i18n from '@/locales'
const { t } = i18n.global

router.afterEach((to) => {
  const title = to.meta.titleKey ? t(to.meta.titleKey) : to.meta.title || ''
  document.title = title ? `${title} - GKube` : 'GKube - Kubernetes 管理平台'
})
```

注意：`afterEach` 中不能直接用 `useI18n()`（非 Vue 组件上下文），需导入 i18n 实例。

### Step 19: 构建优化

**修改文件**: `vite.config.ts`

**Step 19a**: 添加代码分割策略
```ts
export default defineConfig({
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          'element-plus': ['element-plus'],
          'monaco-editor': ['monaco-editor'],
          'vendor': ['vue', 'vue-router', 'pinia', 'axios'],
        }
      }
    }
  }
})
```

**Step 19b**: 添加 Element Plus 预构建优化
```ts
optimizeDeps: {
  include: ['element-plus']
}
```

**Step 19c**: proxy 支持环境变量

注意：Vite 中环境变量通过 `import.meta.env.VITE_*` 访问（浏览器端），但 `vite.config.ts` 是 Node 端执行的，应使用 `process.env`：
```ts
server: {
  proxy: {
    '/api': { target: process.env.VITE_API_URL || 'http://localhost:8080', ... },
    '/v1': { target: process.env.VITE_API_URL || 'http://localhost:8080', ... },
  }
}
```

在项目根目录创建 `.env.local`（git 忽略）设置开发环境变量：
```
VITE_API_URL=http://localhost:8080
```

### Step 20: SecretList 改用 useResourceList

**问题**: `config/secret/SecretList.vue`（388 行）是唯一手动实现列表逻辑的页面。

**执行指令**:
1. 读取 `SecretList.vue` 完整代码
2. 读取其他 List 页面（如 ConfigMapList.vue）作为参考
3. 重构为使用 `useResourceList` composable，删除手写的 `fetchSecrets`、`filteredList`、`selectedRows`、batch delete 逻辑
4. 保留 Secret 特有的功能（base64 编解码显示、type 字段）

### Step 21: 修复 v-for 缺少 :key

**执行指令**:
```bash
# 找出所有 v-for 缺少 :key 的位置
grep -rn 'v-for=' src/views/ src/components/ | grep -v ':key'
```

逐个添加 `:key`，优先使用唯一标识（如 `item.uid`、`item.name`、`index`）。

---

## 八、Phase D: 工程规范 + 性能 + 可访问性

> Phase C 完成后执行。优先级最低，可根据团队需要选择性执行。

### Step 22: 添加 ESLint + Prettier 配置

**执行指令**:
1. 安装依赖（锁定 ESLint 8 以兼容 `.eslintrc.cjs` 格式）：
```bash
cd /data/gkube/frontend
npm install -D eslint@8 @typescript-eslint/parser @typescript-eslint/eslint-plugin eslint-plugin-vue prettier eslint-config-prettier
```

2. 创建 `.eslintrc.cjs`:
```js
module.exports = {
  root: true,
  extends: [
    'eslint:recommended',
    'plugin:@typescript-eslint/recommended',
    'plugin:vue/vue3-recommended',
    'prettier'
  ],
  parser: 'vue-eslint-parser',
  parserOptions: {
    parser: '@typescript-eslint/parser',
    ecmaVersion: 'latest',
    sourceType: 'module'
  },
  rules: {
    '@typescript-eslint/no-explicit-any': 'warn',
    '@typescript-eslint/no-unused-vars': ['warn', { argsIgnorePattern: '^_' }],
    'vue/multi-word-component-names': 'off',
  }
}
```

3. 创建 `.prettierrc`:
```json
{
  "semi": false,
  "singleQuote": true,
  "tabWidth": 2,
  "trailingComma": "es5",
  "printWidth": 120
}
```

4. 在 `package.json` 添加 script：
```json
"lint": "eslint src --ext .ts,.vue --fix",
"format": "prettier --write src/"
```

5. 逐步修复 lint 错误（不要一次性修复所有，按文件逐步推进）

### Step 23: 添加虚拟滚动

**问题**: EventList 等大列表页面无虚拟滚动，大量数据时 DOM 节点过多。

**注意**: Element Plus 的 `el-table` **没有** `virtual-scrolling` prop。虚拟滚动需要使用以下方案之一：

**方案 A（推荐）**: 使用 Element Plus 的 `el-table-v2`（第二代虚拟表格）
```vue
<el-table-v2 :data="events" :columns="columns" :height="600" fixed />
```
`el-table-v2` 原生支持虚拟滚动，但 API 与 `el-table` 不同（columns 配置式），需要重写表格模板。

**方案 B**: 使用 `@tanstack/vue-virtual` 包装原生表格
```bash
npm install @tanstack/vue-virtual
```
用 `useVirtualizer` 包装 `<tbody>`，只渲染可视区域内的行。保留 `el-table` 的外观但需要手动管理行渲染。

**执行指令**:
1. 先确认 EventList.vue 和 PodList.vue 的数据量级（是否真的需要虚拟滚动）
2. 如需要，优先用方案 A（`el-table-v2`），因为它与 Element Plus 集成更好
3. `el-table-v2` 不支持展开行和固定列的某些功能，需测试兼容性

### Step 24: 修复 DashboardView watch 级联更新

**问题**: `DashboardView.vue:292-296` 三个 watch 监听 reactive 对象，批量 fetch 时多次触发图表重绘。

**执行指令**:
1. 读取 DashboardView.vue 的 watch 定义和 updateAllCharts 函数
2. 检查 `package.json` 是否有 `@vueuse/core` 依赖
3. 如有 `@vueuse/core`：
```ts
import { useDebounceFn } from '@vueuse/core'
const debouncedUpdateCharts = useDebounceFn(updateAllCharts, 100)
watch([resources, () => readyCount.value, nsList], () => nextTick(debouncedUpdateCharts))
```
4. 如没有 `@vueuse/core`，自行实现防抖（避免引入新依赖）：
```ts
function debounce<F extends (...args: any[]) => any>(fn: F, delay: number) {
  let timer: ReturnType<typeof setTimeout>
  return (...args: Parameters<F>) => {
    clearTimeout(timer)
    timer = setTimeout(() => fn(...args), delay)
  }
}
const debouncedUpdateCharts = debounce(updateAllCharts, 100)
watch([resources, () => readyCount.value, nsList], () => nextTick(debouncedUpdateCharts))
```

### Step 25: deepClone 改用 structuredClone

**问题**: `utils/helpers.ts:117-119` 使用 `JSON.parse(JSON.stringify())`，另有 2 处内联。

**执行指令**:
```bash
grep -rn "JSON\.parse(JSON\.stringify" src/
```

1. 修改 `utils/helpers.ts` 的 `deepClone`：
```ts
export function deepClone<T>(obj: T): T {
  return structuredClone(obj)
}
```
2. 将 HPAForm.vue:43、NetworkPolicyForm.vue:324 的内联 `JSON.parse(JSON.stringify(...))` 替换为 `deepClone(...)` import
3. 确认目标浏览器支持 `structuredClone`（Chrome 98+, Firefox 94+, Safari 15.4+）

### Step 26: resource.ts 拆分

**问题**: `utils/resource.ts`（122 行）混合了验证逻辑和格式化逻辑。

**执行指令**:
拆分为：
- `utils/k8s-validation.ts` — qualified name 验证、label value 验证、taint effect 验证
- `utils/k8s-format.ts` — CPU/内存格式化

保留 `utils/resource.ts` 作为 barrel re-export（兼容现有 import）。

---

## 九、全局依赖关系

```
Phase A (Step 1-10) ─── 基础架构重构
         │
Phase B (Step 11-17) ← 依赖 Phase A 完成
         │
Phase C (Step 18-21) ← 依赖 Phase B 完成
         │
Phase D (Step 22-26) ← 依赖 Phase C 完成（可选择性执行）
```

Phase A 内部依赖：
```
Step 1 + 2 + 10c + 10d（并行）→ Step 3 + 4（并行）→ Step 5 → Step 6 + 7（并行）→ Step 8 → Step 9 → Step 10a/10b
```

Phase B 内部依赖：
```
Step 11 + 12（并行）→ Step 13 → Step 14（顺序，两者修改同一文件的不同部分）→ Step 16 + 17（并行）
```
注：Step 15 已验证为不重叠，保持现状，不需要执行。

Phase C 内部依赖：
```
Step 18 + 19（并行）→ Step 20 + 21（并行）
```

Phase D 内部依赖：
```
Step 22 → Step 23 + 24 + 25 + 26（并行）
```

---

## 十、全局验证方式

每个 Phase 完成后：
1. `npm run build` — 编译通过
2. `npm run dev` — 手动验证核心功能（列表/详情/创建/删除/YAML 编辑/终端/日志）
3. `grep -r ": any" src/ | wc -l` — any 数量（Phase A 后应从 119 降到 ~30）
4. `wc -l src/views/workload/*Detail.vue` — Detail 行数（Phase A 后应降到 250-400）
5. `wc -l src/views/workload/components/*Form.vue` — Form 行数（Phase A 后应降到 ~400）
6. `wc -l src/router/*.ts` — 路由总行数（Phase C 后应从 627 降到 ~150）
7. `grep -rn "ElMessage\.\(error\|success\|warning\)" src/views/ src/composables/ | grep "[\x{4e00}-\x{9fff}]" | wc -l` — 硬编码中文消息数（Phase B 后应为 0）
