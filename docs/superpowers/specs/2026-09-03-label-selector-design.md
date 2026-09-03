# 列表页 Label Selector 过滤 — 详细设计

## 概述

在所有资源列表页增加可折叠的 Label Selector 过滤器，支持从集群已有 label 中自动补全，支持 `=`、`!=`、`in`、`notin` 操作符。

## 交互设计

### 布局状态

**状态 1：无筛选（默认）**
```
┌──────────────────────────────────────────────────────────────────┐
│  命名空间 [default ▾]  [Label 过滤]  🔍 搜索名称...   总计: 8 [+ 创建] │
├──────────────────────────────────────────────────────────────────┤
│  ☐  名称              命名空间   副本   状态      镜像       操作  │
│  ☐  nginx-deployment  default    3/3   ● 可用   nginx:1.25 详情  │
│  ☐  api-server        default    5/5   ● 可用   api:v2.1   详情  │
└──────────────────────────────────────────────────────────────────┘
```

**状态 2：有筛选但收起**
```
┌──────────────────────────────────────────────────────────────────┐
│  命名空间 [default ▾]  [Label 过滤 ·2]  🔍 搜索名称...  总计: 3 [+ 创建]│
├──────────────────────────────────────────────────────────────────┤
│  ☐  名称              命名空间   副本   状态      镜像       操作  │
│  ☐  nginx-deployment  default    3/3   ● 可用   nginx:1.25 详情  │
└──────────────────────────────────────────────────────────────────┘
```
`·2` badge 表示有 2 个筛选条件，标签和选择器全部隐藏。

**状态 3：展开**
```
┌──────────────────────────────────────────────────────────────────┐
│  命名空间 [default ▾]  [Label 过滤 ·2 ▲]  🔍 搜索名称... 总计: 3 [+ 创建]│
│                                                                  │
│  Key [app ▾]       操作 [= ▾]       Value [nginx ▾]         [添加]│
│  [app = nginx ×]  [env in (prod,staging) ×]          [清除全部]  │
├──────────────────────────────────────────────────────────────────┤
│  ☐  名称              命名空间   副本   状态      镜像       操作  │
│  ☐  nginx-deployment  default    3/3   ● 可用   nginx:1.25 详情  │
└──────────────────────────────────────────────────────────────────┘
```

### 操作符对应的 Value 输入方式

| 操作符 | Value 输入方式 | 示例 |
|--------|--------------|------|
| `=` | 单选下拉（filterable + allow-create） | `app = nginx` |
| `!=` | 单选下拉（filterable + allow-create） | `app != nginx` |
| `in` | 多选下拉（filterable + allow-create） | `env in (prod, staging)` |
| `notin` | 多选下拉（filterable + allow-create） | `env notin (dev)` |

Key 和 Value 下拉框均支持 `filterable`（搜索过滤）+ `allow-create`（允许手写输入不在列表中的值）。

### 交互规则

| 场景 | 行为 |
|------|------|
| 命名空间切换 | 自动清除所有 label 条件，badge 消失 |
| 点击「添加」 | 清空 Value，保留 Key 和操作符（方便连续添加同 key 不同 value） |
| Tag 文本过长（>30字符） | `in`/`notin` 只显示前 2 个值 + `+N`，hover 显示完整列表 |
| 展开/收起 | 点击 `[Label 过滤]` 按钮切换，按钮样式为 el-button text + 🏷 图标 |

## 后端设计

### 1. 新增 API：获取可用 Label

**接口**：`GET /v1/k8s/labels`

**参数**：
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| clusterName | string | 是 | 集群名称 |
| namespace | string | 否 | 命名空间（不传则查所有命名空间） |
| resourceType | string | 是 | 资源类型（deployment/pod/service 等） |

**返回**：
```json
{
  "code": 200,
  "data": {
    "keys": ["app", "env", "tier", "version"],
    "values": {
      "app": ["nginx", "api-server", "frontend"],
      "env": ["prod", "staging", "dev"],
      "tier": ["frontend", "backend"],
      "version": ["v1", "v2"]
    }
  }
}
```

**resourceType 映射表**（前端传值 → 后端 client-go 调用）：

| resourceType | client-go 方法 | 是否集群级 |
|---|---|---|
| `deployment` | `client.AppsV1().Deployments(ns).List()` | 否 |
| `statefulset` | `client.AppsV1().StatefulSets(ns).List()` | 否 |
| `daemonset` | `client.AppsV1().DaemonSets(ns).List()` | 否 |
| `pod` | `client.CoreV1().Pods(ns).List()` | 否 |
| `service` | `client.CoreV1().Services(ns).List()` | 否 |
| `ingress` | `client.NetworkingV1().Ingresses(ns).List()` | 否 |
| `configmap` | `client.CoreV1().ConfigMaps(ns).List()` | 否 |
| `secret` | `client.CoreV1().Secrets(ns).List()` | 否 |
| `job` | `client.BatchV1().Jobs(ns).List()` | 否 |
| `cronjob` | `client.BatchV1().CronJobs(ns).List()` | 否 |
| `replicaset` | `client.AppsV1().ReplicaSets(ns).List()` | 否 |
| `horizontalpodautoscaler` | `client.AutoscalingV2().HPAs(ns).List()` | 否 |
| `networkpolicy` | `client.NetworkingV1().NetworkPolicies(ns).List()` | 否 |
| `persistentvolumeclaim` | `client.CoreV1().PersistentVolumeClaims(ns).List()` | 否 |
| `resourcequota` | `client.CoreV1().ResourceQuotas(ns).List()` | 否 |
| `limitrange` | `client.CoreV1().LimitRanges(ns).List()` | 否 |
| `namespace` | `client.CoreV1().Namespaces().List()` | 是 |
| `node` | `client.CoreV1().Nodes().List()` | 是 |
| `persistentvolume` | `client.CoreV1().PersistentVolumes().List()` | 是 |
| `storageclass` | `client.StorageV1().StorageClasses().List()` | 是 |
| `volumesnapshot` | `dynamic.Resource(volumesnapshotGVR).Namespace(ns).List()` | 否 |
| `volumesnapshotclass` | `dynamic.Resource(volumesnapshotClassGVR).List()` | 是 |
| `customresourcedefinition` | `aeClient.ApiextensionsV1().CustomResourceDefinitions().List()` | 是 |

**实现逻辑**：
- 根据 `resourceType` 调用对应的 client-go List 方法（取前 200 条，避免大集群全量扫描）
- 遍历所有资源的 `.metadata.labels`，收集唯一 key 和每个 key 对应的唯一 values
- 结果缓存 30 秒（sync.Map + TTL），缓存 key 为 `clusterName/namespace/resourceType` 三元组
- 集群级资源忽略 namespace 参数
- 每个 key 最多返回50个 value，防止响应过大

**新增文件**：
- `backend/internal/k8s/labels.go` — handler
- `backend/pkg/k8s/labels/labels.go` — 逻辑（遍历资源提取 label）

### 2. 修改现有 List 接口：增加结构化 label 过滤参数

**改动范围**：所有资源的 List 接口增加结构化 label 过滤参数（不接收原始 `labelSelector` 字符串，由后端拼接 + 校验）。

**请求方式**：将现有 GET list 接口改为 **POST**（GET 不支持嵌套数组参数）。请求体 JSON 格式：
```json
{
  "clusterName": "my-cluster",
  "namespace": "default",
  "limit": 50,
  "labelFilters": [
    {"key": "app", "operator": "=", "values": ["nginx"]},
    {"key": "env", "operator": "in", "values": ["prod", "staging"]}
  ]
}
```

**集群级资源**（node/pv/storageclass）：`namespace` 传空或不传，`labelFilters` 正常生效。

**后端拼接逻辑**（`backend/pkg/k8s/labels/labels.go`）：
```go
// BuildLabelSelector 将结构化 LabelFilter 切片拼接为 K8s label selector 字符串。
// filters 为空或 nil 时返回空字符串（不过滤）。
func BuildLabelSelector(filters []LabelFilter) (string, error) {
    if len(filters) == 0 {
        return "", nil
    }
    var parts []string
    for _, f := range filters {
        switch f.Operator {
        case "=":
            if len(f.Values) != 1 { return "", fmt.Errorf("= 需要恰好1个值") }
            parts = append(parts, fmt.Sprintf("%s=%s", f.Key, f.Values[0]))
        case "!=":
            if len(f.Values) != 1 { return "", fmt.Errorf("!= 需要恰好1个值") }
            parts = append(parts, fmt.Sprintf("%s!=%s", f.Key, f.Values[0]))
        case "in":
            parts = append(parts, fmt.Sprintf("%s in (%s)", f.Key, strings.Join(f.Values, ",")))
        case "notin":
            parts = append(parts, fmt.Sprintf("%s notin (%s)", f.Key, strings.Join(f.Values, ",")))
        default:
            return "", fmt.Errorf("不支持的操作符 %s", f.Operator)
        }
    }
    selector := strings.Join(parts, ",")
    // 用 K8s labels.Parse 校验合法性
    if _, err := labels.Parse(selector); err != nil {
        return "", fmt.Errorf("非法的 label selector: %w", err)
    }
    return selector, nil
}
```

**改动文件**（`backend/pkg/k8s/*/api.go`）：
- `deployment/api.go` — `ListDeployments`
- `pod/api.go` — `ListPods`
- `service/api.go` — `ListServices`
- `statefulset/api.go` — `ListStatefulSets`
- `daemonset/api.go` — `ListDaemonSets`
- `job/api.go` — `ListJobs`
- `cronjob/api.go` — `ListCronJobs`
- `ingress/api.go` — `ListIngresses`
- `configmap/api.go` — `ListConfigMaps`
- `secret/api.go` — `ListSecrets`
- `pvc/api.go` — `ListPVCs`
- `pv/api.go` — `ListPVs`
- `storageclass/api.go` — `ListStorageClasses`
- `networkpolicy/api.go` — `ListNetworkPolicies`
- `hpa/api.go` — `ListHPAs`
- `replicaset/api.go` — `ListReplicaSets`
- `namespace/api.go` — `ListNamespaces`
- `node/api.go` — `ListNodes`
- `crd/api.go` — 相关函数

**统一改动模式**：pkg 层接收已校验的 `labelSelector string`（handler 层负责拼接 + 校验）。
```go
// 改前
func ListDeployments(client *kubernetes.Clientset, namespace string, limit int64, continueToken string) (*appsv1.DeploymentList, error) {
    listOpts := metav1.ListOptions{ResourceVersion: "0"}
    ...
}

// 改后
func ListDeployments(client *kubernetes.Clientset, namespace string, limit int64, continueToken string, labelSelector string) (*appsv1.DeploymentList, error) {
    listOpts := metav1.ListOptions{ResourceVersion: "0"}
    if labelSelector != "" {
        listOpts.LabelSelector = labelSelector
    }
    ...
}
```

**Handler 层改动**（`backend/internal/k8s/*.go`）：
- 改用 `ShouldBindJSON`（POST body）读取结构化 `LabelFilters`
- `LabelFilters` 类型引用 `pkg/k8s/labels.LabelFilter`（同包定义，无循环依赖）
- 调用 `labels.BuildLabelSelector(filters)` 拼接 + 校验
- 将校验后的 selector 字符串传递给 pkg 函数

**Handler 层改动模式**：
```go
import k8sLabels "gkube/pkg/k8s/labels"

// 改后
func (dp *deployment) GetDeploymentList(c *gin.Context) {
    var query DeploymentListParams
    if err := c.ShouldBindJSON(&query); err != nil {
        response.Fail(c, "参数校验失败")
        return
    }
    // 构建 label selector（LabelFilters 为空时返回空字符串，跳过校验）
    selector, err := k8sLabels.BuildLabelSelector(query.LabelFilters)
    if err != nil {
        response.Fail(c, err.Error())
        return
    }
    deploymentList, err := k8sDeployment.ListDeployments(client, query.Namespace, limit, continueToken, selector)
}
```

**Params 结构体增加字段**：
```go
type DeploymentListParams struct {
    ClusterName   string                `json:"clusterName" binding:"required"`
    Namespace     string                `json:"namespace"`
    Limit         int64                 `json:"limit"`
    Continue      string                `json:"continue"`
    LabelFilters  []k8sLabels.LabelFilter `json:"labelFilters"` // 新增，引用 pkg 类型
}
```

**LabelFilter 结构体**（`backend/pkg/k8s/labels/labels.go`，与 `BuildLabelSelector` 同包）：
```go
type LabelFilter struct {
    Key      string   `json:"key" binding:"required"`
    Operator string   `json:"operator" binding:"required,oneof== != in notin"`
    Values   []string `json:"values" binding:"required,min=1"`
}
```
注意：`oneof` 标签中 `=` 后无空格，四个合法值为 `=`、`!=`、`in`、`notin`（不含 `==`）。

### 3. 路由注册

**新增路由**（`backend/internal/router/k8s.go`）：
```go
k8s.GET("/labels", k8sHandler.GetLabels)
```

**路由变更**：所有 list 路由从 GET 改为 POST（支持结构化 body）：
```go
// 改前
k8s.GET("/deployment/list", k8sHandler.Deployment.GetDeploymentList)
// 改后
k8s.POST("/deployment/list", k8sHandler.Deployment.GetDeploymentList)
```

前端 `resource.ts` 的 list 函数对应改为 `request.post(...)`。

## 前端设计

### 1. 新增组件：`LabelSelector.vue`

**文件**：`frontend/src/components/LabelSelector.vue`

**导出类型**（供其他组件/composable 使用）：
```ts
export interface LabelCondition {
  key: string
  operator: '=' | '!=' | 'in' | 'notin'
  values: string[]
}
```

**Props**：
```ts
interface Props {
  clusterName: string
  namespace?: string
  resourceType: string        // 'deployment' | 'pod' | 'service' 等
  modelValue: LabelCondition[] // v-model
}
```

**Emits**：
```ts
'update:modelValue': [conditions: LabelCondition[]]
```

**内部状态**：
```ts
const expanded = ref(false)  // 展开/收起
const loading = ref(false)   // labels API 加载中
const availableKeys = ref<string[]>([])
const availableValues = ref<Record<string, string[]>>({})
const selectedKey = ref('')
const selectedOperator = ref<'=' | '!=' | 'in' | 'notin'>('=')
const selectedValues = ref<string[]>([])
```

**去重检查**：点击「添加」时，检查是否已存在相同的 `key + operator + values` 组合，重复则提示"条件已存在"。

**交互流程**：
1. 点击 `[Label 过滤]` 按钮切换展开/收起
2. 展开时调用 `GET /v1/k8s/labels` 获取可用 keys 和 values（首次展开时加载，之后缓存）
3. 加载中：Key/Value 下拉框显示 `loading` 状态，禁用操作
4. 加载失败：显示警告提示"获取标签失败"，Key/Value 下拉框仍可用（allow-create 兜底，允许手动输入）
5. Key 下拉框展示 `availableKeys`，支持 `filterable`（搜索过滤）+ `allow-create`（允许手写）
6. 选定 Key 后，Value 下拉框展示该 key 对应的 `availableValues[key]`，同样支持 `filterable` + `allow-create`
7. 操作符为 `in`/`notin` 时，Value 下拉框变为多选
8. 点击「添加」按钮前，检查是否已存在相同 `key + operator + values` 组合，重复则提示"条件已存在"
9. 添加后**清空 Value，保留 Key 和操作符**（方便连续添加同 key 不同 value）
10. 已选条件以 tag 形式展示在选择器下方，点击 × 移除
11. 添加/移除条件后自动触发列表刷新
12. Tag 文本超过30字符时，`in`/`notin` 只显示前2个值 + `+N`，hover 显示完整列表

**Element Plus 组件选型**：
- Key/Value 下拉框：`el-select` + `filterable` + `allow-create`（支持搜索和手写）
- 操作符：`el-select`（4 个选项）
- 已选条件 tag：`el-tag` + `closable`，超长文本用 `el-tooltip` 展示完整内容
- 多选值：`el-select` + `multiple` + `filterable` + `allow-create`
- 展开/收起按钮：`el-button` + `text` 类型，带 🏷 图标
- 条件数量 badge：`el-badge`

### 2. 修改组件：`ResourceListToolbar.vue`

**新增 Props**：
```ts
clusterName?: string      // 集群名称（传给 LabelSelector）
resourceType?: string     // 资源类型（传给 LabelSelector）
```

**新增 Emits**：
```ts
'labelSelectorChange': [conditions: LabelCondition[]]
```

**布局改动**：在现有 filter-bar 内，命名空间选择器之后、搜索框之前，插入 Label 过滤按钮和 LabelSelector 组件。

```vue
<template>
  <el-card shadow="never" class="filter-card">
    <div class="filter-bar">
      <!-- 命名空间 -->
      <el-select ... />

      <!-- Label 过滤按钮（新增） -->
      <LabelSelector
        v-if="clusterName && resourceType"
        :cluster-name="clusterName"
        :namespace="namespaceValue"
        :resource-type="resourceType"
        :model-value="labelConditions"
        @update:model-value="emit('labelSelectorChange', $event)"
      />

      <!-- 搜索框 -->
      <el-input ... />

      <!-- 总计 + 右侧操作 -->
      ...
    </div>
  </el-card>
</template>
```

**命名空间切换行为**：watch `namespaceValue`，变化时 emit `labelSelectorChange([])`（清空 label 条件）。

**集群级资源**：当列表页不显示命名空间选择器时（如 Node、PV），LabelSelector 的 `namespace` 传空，labels API 忽略 namespace 参数。

**样式**：Label 过滤按钮与命名空间选择器、搜索框在同一行，不换行。展开时选择器和 tag 行在工具栏下方。

### 3. 修改 Composable：`useResourceList.ts`

**新增选项**：
```ts
export interface ResourceListOptions {
  ...existing options...
  /** 获取当前集群名称（用于 label selector） */
  getClusterName?: () => string
}
```

**新增状态**：
```ts
const labelConditions = ref<LabelCondition[]>([])
```

**修改 `fetchResources`**：改为 POST，传结构化 body。
```ts
async function fetchResources() {
  const body: any = {}
  if (selectedNamespace.value) body.namespace = selectedNamespace.value
  if (options.getClusterName) body.clusterName = options.getClusterName()
  if (labelConditions.value.length > 0) {
    body.labelFilters = labelConditions.value
  }
  if (options.paginated) {
    body.limit = pageSize.value
    // 非首页时传 continue token
    if (continueTokens.value.length > 0) {
      body.continue = continueTokens.value[continueTokens.value.length - 1]
    }
  }
  const res = await options.fetchList(body)
  ...
}
```

**新增方法**：
```ts
function onLabelConditionsChange(conditions: LabelCondition[]) {
  labelConditions.value = conditions
  currentPage.value = 1
  continueTokens.value = []
  fetchResources()
}

/** 清空 label 条件（命名空间切换时调用） */
function clearLabelConditions() {
  labelConditions.value = []
}
```

**新增 return**：
```ts
return {
  ...
  labelConditions,
  onLabelConditionsChange,
  clearLabelConditions,
}
```

### 4. API 层：`resource.ts`

**类型定义**（在 `resource.ts` 顶部或独立 `types/label.ts`）：
```ts
export interface LabelFilter {
  key: string
  operator: '=' | '!=' | 'in' | 'notin'
  values: string[]
}
```

**新增函数**（GET，参数简单）：
```ts
export function getAvailableLabels(params: {
  clusterName: string
  namespace?: string
  resourceType: string
}) {
  return request.get('/k8s/labels', { params })
}
```

**修改现有函数**：所有 list 函数改为 **POST**，传结构化 body：
```ts
export function getDeploymentList(data: {
  namespace?: string
  clusterName: string
  limit?: number
  continue?: string
  labelFilters?: LabelFilter[]
}) {
  return request.post('/k8s/deployment/list', data)
}
```

**调用方改动**：列表页调用从 `getDeploymentList({ namespace })` 改为 `getDeploymentList({ clusterName, namespace, labelFilters })`。

**迁移注意**：GET 改 POST 是破坏性变更，需同步修改所有前端调用方。如有外部 API 消费者，需提供迁移期或保留 GET 兼容路由。

### 5. URL 同步

**设计**：label selector 同步到 URL query 参数，刷新页面不丢失。

**URL 格式**：JSON 编码后 base64，简洁且支持复杂结构：
```
?ls=W3sia2V5IjoiYXBwIiwib3BlcmF0b3IiOiI9IiwidmFsdWVzIjpbIm5naW54Il19XQ==
```

**实现**：
- `useResourceList` 中 watch `labelConditions`，序列化为 base64 后同步到 `router.replace({ query: { ..., ls: base64 } })`
- 初始化时从 URL query 解析 `ls`，还原为 `labelConditions`
- 解析失败时静默忽略（清除无效参数）

### 6. 各列表页集成

每个列表页需要传递 `clusterName` 和 `resourceType` 给 `ResourceListToolbar`：

```vue
<ResourceListToolbar
  :cluster-name="clusterStore.clusterName"
  resource-type="deployment"
  ...
/>
```

**涉及的列表页**：
- `views/workload/DeploymentList.vue` — resourceType: `'deployment'`
- `views/workload/PodList.vue` — `'pod'`
- `views/workload/StatefulSetList.vue` — `'statefulset'`
- `views/workload/DaemonSetList.vue` — `'daemonset'`
- `views/workload/JobList.vue` — `'job'`
- `views/workload/CronJobList.vue` — `'cronjob'`
- `views/workload/ReplicaSetList.vue` — `'replicaset'`
- `views/workload/hpa/HPAList.vue` — `'horizontalpodautoscaler'`
- `views/network/ServiceList.vue` — `'service'`
- `views/network/IngressList.vue` — `'ingress'`
- `views/network/networkpolicy/NetworkPolicyList.vue` — `'networkpolicy'`
- `views/config/configmap/ConfigMapList.vue` — `'configmap'`
- `views/config/secret/SecretList.vue` — `'secret'`
- `views/config/resourcequota/ResourceQuotaList.vue` — `'resourcequota'`
- `views/config/limitrange/LimitRangeList.vue` — `'limitrange'`
- `views/storage/PVList.vue` — `'persistentvolume'`
- `views/storage/PVCList.vue` — `'persistentvolumeclaim'`
- `views/storage/StorageClassList.vue` — `'storageclass'`
- `views/storage/VolumeSnapshotList.vue` — `'volumesnapshot'`
- `views/storage/VolumeSnapshotClassList.vue` — `'volumesnapshotclass'`
- `views/namespace/NamespaceList.vue` — `'namespace'`
- `views/node/NodeList.vue` — `'node'`
- `views/crd/CRDList.vue` — `'customresourcedefinition'`

## 文件变更清单

### 新增文件
| 文件 | 说明 |
|------|------|
| `frontend/src/components/LabelSelector.vue` | Label Selector 组件（折叠式，filterable + allow-create，export LabelCondition 类型） |
| `backend/internal/k8s/labels.go` | Label 查询 handler |
| `backend/pkg/k8s/labels/labels.go` | LabelFilter 结构体定义 + BuildLabelSelector 工具函数 + Label 查询逻辑 |

### 修改文件
| 文件 | 改动 |
|------|------|
| `frontend/src/components/ResourceListToolbar.vue` | 集成 LabelSelector，新增 props/emits，命名空间切换时清空条件 |
| `frontend/src/composables/useResourceList.ts` | 增加 labelConditions 状态、onLabelConditionsChange/clearLabelConditions 方法、fetchResources 改 POST、URL 同步 |
| `frontend/src/api/resource.ts` | 新增 getAvailableLabels（GET），所有 list 函数改为 POST + 结构化 body |
| `backend/internal/router/k8s.go` | 注册 GET /labels 路由 |
| `backend/internal/k8s/*.go`（约 20 个文件） | handler 改用 ShouldBindJSON，增加 LabelFilter → labelSelector 拼接 |
| `backend/pkg/k8s/*/api.go`（约 20 个文件） | List 函数增加 labelSelector string 参数 |
| `frontend/src/views/**/*.vue`（约 23 个列表页） | 传递 clusterName/resourceType 给 Toolbar，list 调用改为 POST |
