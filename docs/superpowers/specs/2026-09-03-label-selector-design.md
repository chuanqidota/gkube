# 列表页 Label Selector 过滤 — 详细设计

## 概述

在所有资源列表页增加 Label Selector 过滤器，支持从集群已有 label 中自动补全，支持 `=`、`!=`、`in`、`notin` 操作符。

## 交互设计

### 设计原则

1. **低频功能用 Popover** — Label 过滤是高级功能，使用频率低于搜索框和命名空间，不常驻占用空间
2. **下拉框选择** — 新手友好，支持自动补全，不要求用户记语法
3. **AND 关系** — 搜索框、命名空间、Label 三个条件联合生效
4. **风格统一** — Label 按钮与命名空间选择器视觉风格一致

### 布局状态

**状态 1：默认收起（无筛选）**
```
┌─────────────────────────────────────────────────────────────┐
│  命名空间 [default ▾]   Label [▾]    🔍 搜索名称... [+ 创建] │
├─────────────────────────────────────────────────────────────┤
│  ☐  名称              命名空间   副本   状态      镜像   操作 │
│  ☐  nginx-deployment  default    3/3   ● 可用   nginx  详情 │
│  ☐  api-server        default    5/5   ● 可用   api:v2 详情 │
└─────────────────────────────────────────────────────────────┘
```

- Label 按钮显示文字 `Label` + 下拉箭头，无图标
- 与命名空间选择器风格一致

**状态 2：有筛选条件（badge 提示）**
```
┌─────────────────────────────────────────────────────────────┐
│  命名空间 [default ▾]  Label [·2 ▾]  🔍 搜索名称... [+ 创建]│
├─────────────────────────────────────────────────────────────┤
│  ☐  名称              命名空间   副本   状态      镜像   操作 │
│  ☐  nginx-deployment  default    3/3   ● 可用   nginx  详情 │
└─────────────────────────────────────────────────────────────┘
```

- `·2` badge 表示有 2 个筛选条件
- 条件已在后台生效，列表已过滤

**状态 3：点击 Label 按钮（Popover 展开）**
```
┌──────────────────────────────────────────────────────────────────┐
│  命名空间 [default ▾]  Label [·2 ▾]  🔍 搜索名称...    [+ 创建] │
│                             ┌──────────────────────────────────┐ │
│                             │ 已选条件:                         │ │
│                             │ [app=nginx ×] [env=prod ×]       │ │
│                             │ ──────────────────────────────── │ │
│                             │ 添加新条件:                       │ │
│                             │ [Key ▾    ] [= ▾] [Value ▾    ] │ │
│                             │                                  │ │
│                             │ [+ 添加]  [清除全部]              │ │
│                             │               [取消] [应用]      │ │
│                             └──────────────────────────────────┘ │
├──────────────────────────────────────────────────────────────────┤
│  表格...                                                         │
└──────────────────────────────────────────────────────────────────┘
```

- Popover 宽度 480px
- Popover 从 Label 按钮下方弹出
- 已选条件显示在顶部，点击 tag 可编辑（移入编辑区），点击 × 可删除
- 下方添加新条件区域，Key/Value 下拉框支持搜索和手写
- 点击 [应用] 后关闭 Popover，列表刷新
- 点击 Popover 外部直接关闭，丢弃未应用的修改（无确认弹窗）

**状态 4：加载中**
```
┌───────────────────────────────────────┐
│ 已选条件:                             │
│ [app=nginx ×]                         │
│ ─────────────────────────────────────│
│ 添加新条件:                           │
│ [Key ▾  ▼ loading] [= ▾] [Value ▾] │
│                                       │
│         [取消] [应用]                 │
└───────────────────────────────────────┘
```

- Key 下拉框显示 loading 状态，但不禁用（允许手动输入）
- 操作符和 Value 下拉框正常可用

**状态 5：加载失败**
```
┌───────────────────────────────────────┐
│ ⚠️ 获取标签失败，您可以手动输入标签键值 │
│ ─────────────────────────────────────│
│ 添加新条件:                           │
│ [Key ▾]      [= ▾] [Value ▾]        │
│                                       │
│         [取消] [应用]                 │
└───────────────────────────────────────┘
```

- 显示警告提示
- Key/Value 下拉框仍可用（allow-create 兜底，允许手动输入）

**状态 6：无可用标签**
```
┌───────────────────────────────────────┐
│ 已选条件:                             │
│ (暂无)                                │
│ ─────────────────────────────────────│
│ 未发现可自动补全的标签                 │
│ 您可以手动输入标签键值                 │
│                                       │
│ [Key ▾]      [= ▾] [Value ▾]        │
│                                       │
│ [+ 添加]  [清除全部]                  │
│         [取消] [应用]                 │
└───────────────────────────────────────┘
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
| 点击「应用」 | 关闭 Popover，触发列表刷新 |
| 点击 Popover 外部 | 直接关闭，丢弃未应用的修改 |
| 点击已选条件 tag | 该条件移入编辑区（从已选区移除），用户修改后重新应用 |
| Tag 点击 × | 移除单个条件 |
| 条件超过 5 个 | 已选条件区域显示滚动条，最多显示 120px 高度 |
| API 加载失败 | 显示警告提示，允许手动输入 |

### 键盘操作

键盘事件监听 Popover 容器内部，不监听 document 级别，避免与 Monaco 编辑器等组件冲突。

| 按键 | 操作 |
|------|------|
| `Tab` | 在 Key / Operator / Value 之间切换 |
| `Enter` | 添加条件（在 Value 输入框中且条件合法时） |
| `Esc` | 关闭 Popover（仅在 Popover 内部元素聚焦时生效） |

### 三个过滤条件的关系

```
最终结果 = 命名空间 AND LabelSelector AND 名称搜索

┌─────────────┐
│  命名空间     │──┐
└─────────────┘  │
                 │    ┌─────────────┐
┌─────────────┐  ├───▶│  K8s API    │───▶ 后端返回结果 ───▶ 前端名称过滤 ───▶ 最终列表
│  Label 过滤  │──┤    │  labelSelector   │
└─────────────┘  │    └─────────────┘
                 │
┌─────────────┐  │
│  名称搜索    │──┘ (前端本地过滤，不传后端)
└─────────────┘
```

- 命名空间 + Label → 后端拼接为 K8s API 参数
- 名称搜索 → 前端本地 filter

## 后端设计

### 错误响应格式

所有接口统一使用以下 JSON 格式：

```json
// 成功
{"code": 200, "data": {...}}

// 失败
{"code": -1, "message": "= 需要恰好1个值"}
```

`response.Fail` 实现（引用现有 `pkg/response` 包）：
```go
func Fail(c *gin.Context, msg string) {
    c.JSON(http.StatusOK, gin.H{
        "code":    -1,
        "message": msg,
    })
}
```

### 1. 新增 API：获取可用 Label

**接口**：`GET /v1/k8s/labels`

**参数**：
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| clusterName | string | 是 | 集群名称 |
| namespace | string | 否 | 命名空间（不传则查所有命名空间，集群级资源忽略此参数） |
| resourceType | string | 是 | 资源类型，见下方说明 |

**resourceType 传参格式**：
- 内置资源：传字符串，如 `"deployment"`、`"pod"`、`"service"`
- CRD 实例：传 `group/version/resource` 格式，如 `"argoproj.io/v1alpha1/applications"`

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

**GVR 映射表**（resourceType → GroupVersionResource）：

| resourceType | Group | Version | Resource | 集群级 |
|---|---|---|---|---|
| `deployment` | apps | v1 | deployments | 否 |
| `statefulset` | apps | v1 | statefulsets | 否 |
| `daemonset` | apps | v1 | daemonsets | 否 |
| `pod` | | v1 | pods | 否 |
| `service` | | v1 | services | 否 |
| `ingress` | networking.k8s.io | v1 | ingresses | 否 |
| `configmap` | | v1 | configmaps | 否 |
| `secret` | | v1 | secrets | 否 |
| `job` | batch | v1 | jobs | 否 |
| `cronjob` | batch | v1 | cronjobs | 否 |
| `replicaset` | apps | v1 | replicasets | 否 |
| `horizontalpodautoscaler` | autoscaling | v2 | horizontalpodautoscalers | 否 |
| `networkpolicy` | networking.k8s.io | v1 | networkpolicies | 否 |
| `persistentvolumeclaim` | | v1 | persistentvolumeclaims | 否 |
| `resourcequota` | | v1 | resourcequotas | 否 |
| `limitrange` | | v1 | limitranges | 否 |
| `namespace` | | v1 | namespaces | 是 |
| `node` | | v1 | nodes | 是 |
| `persistentvolume` | | v1 | persistentvolumes | 是 |
| `storageclass` | storage.k8s.io | v1 | storageclasses | 是 |
| `volumesnapshot` | snapshot.storage.k8s.io | v1 | volumesnapshots | 否 |
| `volumesnapshotclass` | snapshot.storage.k8s.io | v1 | volumesnapshotclasses | 是 |
| `customresourcedefinition` | apiextensions.k8s.io | v1 | customresourcedefinitions | 是 |
| CRD 实例 | 自定义 | 自定义 | 自定义 | 取决于 CRD |

**实现逻辑**：
- 使用 dynamic client 统一处理所有资源类型
- 使用 Limit 限制单次请求量（每页 200 条，最多 5 页，共 1000 条）
- 遍历所有资源的 `.metadata.labels`，收集唯一 key 和每个 key 对应的唯一 values
- 结果缓存 30 秒（sync.Map + TTL），缓存 key 为 `clusterName/namespace/resourceType` 三元组
- 集群级资源忽略 namespace 参数（不报错）
- 每个 key 最多返回50个 value，防止响应过大

**并发安全**：
```go
import "golang.org/x/sync/singleflight"

var labelGroup singleflight.Group

// GetLabels 获取可用 labels，合并并发请求防止缓存击穿
func GetLabels(clusterName, namespace, resourceType string) (*LabelData, error) {
    cacheKey := buildCacheKey(clusterName, namespace, resourceType)

    // 先查缓存
    if cached, ok := cache.Get(cacheKey); ok {
        return cached, nil
    }

    // 合并并发请求
    v, err, _ := labelGroup.Do(cacheKey, func() (interface{}, error) {
        return fetchLabelsFromK8s(clusterName, namespace, resourceType)
    })
    if err != nil {
        return nil, err
    }

    data := v.(*LabelData)
    cache.Set(cacheKey, data, 30*time.Second)
    return data, nil
}

// 缓存 key 明确区分集群级和命名空间级
func buildCacheKey(clusterName, namespace, resourceType string) string {
    if namespace == "" {
        return fmt.Sprintf("%s/_all_/%s", clusterName, resourceType)
    }
    return fmt.Sprintf("%s/%s/%s", clusterName, namespace, resourceType)
}
```

**集群级资源判断**：
```go
var clusterScopedResources = map[schema.GroupVersionResource]bool{
    {Version: "v1", Resource: "namespaces"}:                          true,
    {Version: "v1", Resource: "nodes"}:                               true,
    {Version: "v1", Resource: "persistentvolumes"}:                   true,
    {Group: "storage.k8s.io", Version: "v1", Resource: "storageclasses"}:                      true,
    {Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}:     true,
    {Group: "snapshot.storage.k8s.io", Version: "v1", Resource: "volumesnapshotclasses"}:      true,
}

func isClusterScoped(gvr schema.GroupVersionResource) bool {
    return clusterScopedResources[gvr]
}
```

**GVR 解析函数**：
```go
var builtinGVRs = map[string]schema.GroupVersionResource{
    "deployment":                  {Group: "apps", Version: "v1", Resource: "deployments"},
    "statefulset":                 {Group: "apps", Version: "v1", Resource: "statefulsets"},
    "daemonset":                   {Group: "apps", Version: "v1", Resource: "daemonsets"},
    "pod":                         {Version: "v1", Resource: "pods"},
    "service":                     {Version: "v1", Resource: "services"},
    "ingress":                     {Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"},
    "configmap":                   {Version: "v1", Resource: "configmaps"},
    "secret":                      {Version: "v1", Resource: "secrets"},
    "job":                         {Group: "batch", Version: "v1", Resource: "jobs"},
    "cronjob":                     {Group: "batch", Version: "v1", Resource: "cronjobs"},
    "replicaset":                  {Group: "apps", Version: "v1", Resource: "replicasets"},
    "horizontalpodautoscaler":     {Group: "autoscaling", Version: "v2", Resource: "horizontalpodautoscalers"},
    "networkpolicy":               {Group: "networking.k8s.io", Version: "v1", Resource: "networkpolicies"},
    "persistentvolumeclaim":       {Version: "v1", Resource: "persistentvolumeclaims"},
    "resourcequota":               {Version: "v1", Resource: "resourcequotas"},
    "limitrange":                  {Version: "v1", Resource: "limitranges"},
    "namespace":                   {Version: "v1", Resource: "namespaces"},
    "node":                        {Version: "v1", Resource: "nodes"},
    "persistentvolume":            {Version: "v1", Resource: "persistentvolumes"},
    "storageclass":                {Group: "storage.k8s.io", Version: "v1", Resource: "storageclasses"},
    "volumesnapshot":              {Group: "snapshot.storage.k8s.io", Version: "v1", Resource: "volumesnapshots"},
    "volumesnapshotclass":         {Group: "snapshot.storage.k8s.io", Version: "v1", Resource: "volumesnapshotclasses"},
    "customresourcedefinition":    {Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"},
}

// getGVR 解析 resourceType 为 GroupVersionResource
// 内置资源直接查表，CRD 实例解析 "group/version/resource" 格式
func getGVR(resourceType string) (schema.GroupVersionResource, error) {
    if gvr, ok := builtinGVRs[resourceType]; ok {
        return gvr, nil
    }
    // CRD 实例：解析 "group/version/resource" 格式
    parts := strings.SplitN(resourceType, "/", 3)
    if len(parts) == 3 {
        return schema.GroupVersionResource{
            Group:    parts[0],
            Version:  parts[1],
            Resource: parts[2],
        }, nil
    }
    return schema.GroupVersionResource{}, fmt.Errorf("unsupported resource type: %s", resourceType)
}
```

**分页采集逻辑**：
```go
func fetchLabelsFromK8s(clusterName, namespace, resourceType string) (*LabelData, error) {
    gvr, err := getGVR(resourceType)
    if err != nil {
        return nil, err
    }

    client, err := getDynamicClient(clusterName)
    if err != nil {
        return nil, err
    }

    // 根据是否集群级决定是否传 namespace
    var ri dynamic.ResourceInterface
    if isClusterScoped(gvr) {
        ri = client.Resource(gvr)
    } else {
        ri = client.Resource(gvr).Namespace(namespace)
    }

    keys := make(map[string]bool)
    values := make(map[string]map[string]bool) // key -> set of values
    continueToken := ""
    pageCount := 0
    maxPages := 5
    pageSize := int64(200)

    for pageCount < maxPages {
        listOpts := metav1.ListOptions{Limit: pageSize}
        if continueToken == "" {
            listOpts.ResourceVersion = "0" // 仅首次从缓存读
        } else {
            listOpts.Continue = continueToken
        }

        list, err := ri.List(ctx, listOpts)
        if err != nil {
            return nil, err
        }

        // 从 UnstructuredList 提取 labels
        for _, item := range list.Items {
            labels := item.GetLabels()
            for k, v := range labels {
                keys[k] = true
                if values[k] == nil {
                    values[k] = make(map[string]bool)
                }
                values[k][v] = true
            }
        }

        if list.GetContinue() == "" {
            break
        }
        continueToken = list.GetContinue()
        pageCount++
    }

    // 转换为返回格式
    result := &LabelData{
        Keys:   make([]string, 0, len(keys)),
        Values: make(map[string][]string),
    }
    for k := range keys {
        result.Keys = append(result.Keys, k)
        vSet := values[k]
        vList := make([]string, 0, len(vSet))
        for v := range vSet {
            vList = append(vList, v)
        }
        // 每个 key 最多返回 50 个 value
        if len(vList) > 50 {
            vList = vList[:50]
        }
        result.Values[k] = vList
    }

    return result, nil
}
```

**Handler 实现**：
```go
// backend/internal/k8s/labels.go

type labelsHandler struct{}

var Labels = &labelsHandler{}

func (h *labelsHandler) GetLabels(c *gin.Context) {
    var query struct {
        ClusterName  string `form:"clusterName" binding:"required"`
        Namespace    string `form:"namespace"`
        ResourceType string `form:"resourceType" binding:"required"`
    }
    if err := c.ShouldBindQuery(&query); err != nil {
        response.Fail(c, "参数校验失败")
        return
    }

    data, err := k8sLabels.GetLabels(query.ClusterName, query.Namespace, query.ResourceType)
    if err != nil {
        response.Fail(c, err.Error())
        return
    }

    response.Success(c, data)
}
```

**新增文件**：
- `backend/internal/k8s/labels.go` — handler
- `backend/pkg/k8s/labels/labels.go` — LabelFilter、LabelData 结构体 + BuildLabelSelector + GetLabels + GVR 映射 + 分页采集逻辑

### 2. 修改现有 List 接口：增加结构化 label 过滤参数

**改动范围**：所有资源的 List 接口增加结构化 label 过滤参数（不接收原始 `labelSelector` 字符串，由后端拼接 + 校验）。

**请求方式**：保留 GET 路由兼容性，同时新增 POST 路由支持结构化 body。

```go
// 路由注册：保留 GET，新增 POST
k8s.GET("/deployment/list", k8sHandler.Deployment.GetDeploymentList)    // 保留兼容
k8s.POST("/deployment/list", k8sHandler.Deployment.GetDeploymentList)   // 新增
```

**POST 请求体 JSON 格式**：
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

**GET 请求参数**（保留兼容）：
```
GET /v1/k8s/deployment/list?clusterName=my-cluster&namespace=default&limit=50
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
- 支持 GET 和 POST 两种请求方式
- 用 `c.ShouldBind(&query)` 统一处理，gin 根据 Content-Type 自动选择 JSON 或 query 参数绑定
- POST 时读取结构化 `LabelFilters`
- GET 时不支持 `labelFilters`（保持兼容）
- `LabelFilters` 类型引用 `pkg/k8s/labels.LabelFilter`（同包定义，无循环依赖）
- 调用 `labels.BuildLabelSelector(filters)` 拼接 + 校验
- 将校验后的 selector 字符串传递给 pkg 函数

**Handler 层改动模式**：
```go
import k8sLabels "gkube/pkg/k8s/labels"

// 改后：支持 GET 和 POST，用 ShouldBind 统一绑定
func (dp *deployment) GetDeploymentList(c *gin.Context) {
    var query DeploymentListParams
    if err := c.ShouldBind(&query); err != nil {
        response.Fail(c, "参数校验失败")
        return
    }

    // 构建 label selector（GET 时 LabelFilters 为空，跳过）
    var selector string
    if len(query.LabelFilters) > 0 {
        var err error
        selector, err = k8sLabels.BuildLabelSelector(query.LabelFilters)
        if err != nil {
            response.Fail(c, err.Error())
            return
        }
    }

    // client, limit, continueToken 从现有逻辑获取
    deploymentList, err := k8sDeployment.ListDeployments(client, query.Namespace, limit, continueToken, selector)
    ...
}
```

**Params 结构体增加字段**：
```go
type DeploymentListParams struct {
    ClusterName   string                  `json:"clusterName" form:"clusterName" binding:"required"`
    Namespace     string                  `json:"namespace" form:"namespace"`
    Limit         int64                   `json:"limit" form:"limit"`
    Continue      string                  `json:"continue" form:"continue"`
    LabelFilters  []k8sLabels.LabelFilter `json:"labelFilters"` // 新增，仅 POST 有效
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

**LabelFilter 单元测试**：
```go
// backend/pkg/k8s/labels/labels_test.go

func TestLabelFilterBinding(t *testing.T) {
    cases := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"equal", `{"key":"app","operator":"=","values":["nginx"]}`, false},
        {"not equal", `{"key":"app","operator":"!=","values":["nginx"]}`, false},
        {"in", `{"key":"app","operator":"in","values":["a","b"]}`, false},
        {"notin", `{"key":"app","operator":"notin","values":["a"]}`, false},
        {"double equal invalid", `{"key":"app","operator":"==","values":["a"]}`, true},
        {"like invalid", `{"key":"app","operator":"like","values":["a"]}`, true},
        {"empty values", `{"key":"app","operator":"=","values":[]}`, true},
        {"missing key", `{"operator":"=","values":["a"]}`, true},
    }
    v := validator.New()
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            var f LabelFilter
            if err := json.Unmarshal([]byte(tc.input), &f); err != nil {
                t.Fatalf("unmarshal: %v", err)
            }
            err := v.Struct(f)
            if (err != nil) != tc.wantErr {
                t.Errorf("validate: got err=%v, wantErr=%v", err, tc.wantErr)
            }
        })
    }
}

func TestBuildLabelSelector(t *testing.T) {
    cases := []struct {
        name    string
        filters []LabelFilter
        want    string
        wantErr bool
    }{
        {"empty", nil, "", false},
        {"single equal", []LabelFilter{{Key: "app", Operator: "=", Values: []string{"nginx"}}}, "app=nginx", false},
        {"single notin", []LabelFilter{{Key: "env", Operator: "notin", Values: []string{"dev", "test"}}}, "env notin (dev,test)", false},
        {"multiple", []LabelFilter{
            {Key: "app", Operator: "=", Values: []string{"nginx"}},
            {Key: "env", Operator: "in", Values: []string{"prod", "staging"}},
        }, "app=nginx,env in (prod,staging)", false},
        {"equal needs exactly 1 value", []LabelFilter{{Key: "app", Operator: "=", Values: []string{"a", "b"}}}, "", true},
        {"invalid operator", []LabelFilter{{Key: "app", Operator: "like", Values: []string{"a"}}}, "", true},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            got, err := BuildLabelSelector(tc.filters)
            if (err != nil) != tc.wantErr {
                t.Errorf("BuildLabelSelector: got err=%v, wantErr=%v", err, tc.wantErr)
            }
            if got != tc.want {
                t.Errorf("BuildLabelSelector: got %q, want %q", got, tc.want)
            }
        })
    }
}
```

### 3. 路由注册

**新增路由**（`backend/internal/router/k8s.go`）：
```go
k8s.GET("/labels", k8sHandler.Labels.GetLabels)
```

**路由变更**：所有 list 路由新增 POST 路由（保留 GET 兼容）：
```go
// 改前
k8s.GET("/deployment/list", k8sHandler.Deployment.GetDeploymentList)

// 改后：保留 GET，新增 POST
k8s.GET("/deployment/list", k8sHandler.Deployment.GetDeploymentList)
k8s.POST("/deployment/list", k8sHandler.Deployment.GetDeploymentList)
```

前端 `resource.ts` 的 list 函数改为 `request.post(...)`。

## 前端设计

### 1. 新增组件：`LabelFilterPopover.vue`

**文件**：`frontend/src/components/LabelFilterPopover.vue`

**导出类型**（供其他组件/composable 使用）：
```ts
export interface LabelCondition {
  key: string
  operator: '=' | '!=' | 'in' | 'notin'
  values: string[]
}
```

**内部编辑状态**（统一使用 values 数组）：
```ts
interface EditingCondition {
  key: string
  operator: '=' | '!=' | 'in' | 'notin'
  values: string[]  // 统一用数组，单选时只有一个元素
}
```

**类型转换函数**：
```ts
// LabelCondition[] → EditingCondition[]
function convertToEditing(conditions: LabelCondition[]): EditingCondition[] {
  return conditions.map(c => ({
    key: c.key,
    operator: c.operator,
    values: [...c.values],
  }))
}

// EditingCondition[] → LabelCondition[]（过滤掉未完成的条件）
function convertToLabelConditions(editing: EditingCondition[]): LabelCondition[] {
  return editing
    .filter(c => c.key && c.values.length > 0)
    .map(c => ({
      key: c.key,
      operator: c.operator,
      values: [...c.values],
    }))
}
```

**Props**：
```ts
interface Props {
  clusterName: string
  namespace?: string
  resourceType: string        // 内置资源传字符串，CRD 实例传 "group/version/resource"
  modelValue: LabelCondition[] // v-model
}
```

**Emits**：
```ts
'update:modelValue': [conditions: LabelCondition[]]
```

**内部状态**：
```ts
const popoverRef = ref() // popover 组件引用
const popoverVisible = ref(false)
const editingConditions = ref<EditingCondition[]>([])
const availableKeys = ref<string[]>([])
const availableValues = ref<Record<string, string[]>>({})
const loading = ref(false)
const loadError = ref(false)
```

**组件结构**：
```vue
<template>
  <el-popover
    ref="popoverRef"
    :visible="popoverVisible"
    placement="bottom-start"
    :width="480"
    trigger="click"
    @update:visible="onPopoverVisibleChange"
  >
    <template #reference>
      <el-button :class="{ 'has-filters': modelValue.length > 0 }">
        Label
        <el-badge v-if="modelValue.length" :value="modelValue.length" :max="9" />
        <el-icon class="el-icon--right"><ArrowDown /></el-icon>
      </el-button>
    </template>

    <div class="filter-popover" @keydown.esc="cancel" @keydown.enter="onEnter">
      <!-- 加载失败警告 -->
      <el-alert
        v-if="loadError"
        type="warning"
        :closable="false"
        show-icon
        class="load-error"
      >
        获取标签失败，您可以手动输入标签键值
      </el-alert>

      <!-- 已选条件 -->
      <div v-if="modelValue.length" class="selected-filters">
        <div class="tags-scroll">
          <el-tag
            v-for="(cond, i) in modelValue"
            :key="i"
            closable
            size="small"
            class="condition-tag"
            @click="editCondition(i)"
            @close="removeCondition(i)"
          >
            {{ formatCondition(cond) }}
          </el-tag>
        </div>
        <div v-if="modelValue.length > 5" class="more-hint">
          共 {{ modelValue.length }} 个条件
        </div>
      </div>

      <el-divider v-if="modelValue.length" />

      <!-- 无可用标签提示 -->
      <div v-if="!loading && !loadError && availableKeys.length === 0" class="empty-state">
        <p class="empty-text">未发现可自动补全的标签</p>
        <p class="empty-hint">您可以手动输入标签键值</p>
      </div>

      <!-- 条件编辑行 -->
      <div v-for="(cond, i) in editingConditions" :key="i" class="condition-row">
        <el-select
          v-model="cond.key"
          filterable
          allow-create
          placeholder="Key"
          size="small"
          class="key-select"
          :loading="loading"
          @change="onKeyChange(cond)"
        >
          <el-option v-if="loading" label="加载中..." value="" disabled />
          <el-option v-for="k in availableKeys" :key="k" :label="k" :value="k" />
        </el-select>

        <el-select v-model="cond.operator" size="small" class="op-select">
          <el-option label="=" value="=" />
          <el-option label="!=" value="!=" />
          <el-option label="in" value="in" />
          <el-option label="notin" value="notin" />
        </el-select>

        <!-- = / != 单选：用 computed 中转避免数组绑定问题 -->
        <el-select
          v-if="cond.operator === '=' || cond.operator === '!='"
          :model-value="cond.values[0] || ''"
          @update:model-value="cond.values = $event ? [$event] : []"
          filterable
          allow-create
          placeholder="Value"
          size="small"
          class="value-select"
          :loading="loading"
        >
          <el-option v-if="loading" label="加载中..." value="" disabled />
          <el-option v-for="v in getValuesForKey(cond.key)" :key="v" :label="v" :value="v" />
        </el-select>

        <!-- in / notin 多选 -->
        <el-select
          v-else
          v-model="cond.values"
          filterable
          allow-create
          multiple
          placeholder="Value"
          size="small"
          class="value-select"
          :loading="loading"
        >
          <el-option v-if="loading" label="加载中..." value="" disabled />
          <el-option v-for="v in getValuesForKey(cond.key)" :key="v" :label="v" :value="v" />
        </el-select>

        <el-button text size="small" @click="editingConditions.splice(i, 1)">
          <el-icon><Delete /></el-icon>
        </el-button>
      </div>

      <!-- 操作按钮 -->
      <div class="popover-actions">
        <el-button size="small" text @click="addCondition">+ 添加条件</el-button>
        <el-button size="small" text @click="clearAll" :disabled="!modelValue.length">清除全部</el-button>
        <div class="spacer" />
        <el-button size="small" @click="cancel">取消</el-button>
        <el-button size="small" type="primary" @click="apply">应用</el-button>
      </div>
    </div>
  </el-popover>
</template>
```

**交互流程**：
1. 点击 Label 按钮 → Popover 打开
2. 打开时调用 `GET /v1/k8s/labels` 获取可用 keys 和 values
3. 如果已有条件 → 显示在顶部，可删除或点击编辑
4. 如果无条件 → 自动添加一行空条件编辑行
5. 选择 Key → 自动加载该 Key 的 Values 到下拉框
6. 选择 Value → 点击 [应用]
7. Popover 关闭，按钮显示 badge，触发列表刷新

**去重检查**：点击「应用」时，检查是否已存在相同的 `key + operator + values` 组合，重复则自动合并。

**编辑已有条件**：
```ts
function editCondition(index: number) {
  const cond = props.modelValue[index]
  editingConditions.value.push(convertToEditing([cond])[0])
  removeCondition(index)
}
```

**Popover 关闭处理**（丢弃未应用的修改）：
```ts
function onPopoverVisibleChange(visible: boolean) {
  popoverVisible.value = visible
  if (!visible) {
    // 丢弃未应用的修改，恢复到已应用的状态
    editingConditions.value = convertToEditing(props.modelValue)
  }
}
```

**键盘事件**（仅在 Popover 内部生效）：
```ts
function onEnter(e: KeyboardEvent) {
  // 只在 Value 输入框聚焦时触发添加
  const target = e.target as HTMLElement
  if (!target.closest('.value-select')) return
  if (canAdd.value) {
    addCondition()
  }
}

function cancel() {
  popoverVisible.value = false
  editingConditions.value = convertToEditing(props.modelValue)
}
```

**Element Plus 组件选型**：
- Key/Value 下拉框：`el-select` + `filterable` + `allow-create`（支持搜索和手写）
- 操作符：`el-select`（4 个选项）
- 已选条件 tag：`el-tag` + `closable`（点击可编辑）
- 多选值：`el-select` + `multiple` + `filterable` + `allow-create`
- 弹出层：`el-popover`（宽度 480px）
- 数量 badge：`el-badge`
- 加载失败提示：`el-alert`

**样式**：
```css
.filter-popover {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.selected-filters {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.tags-scroll {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  max-height: 120px;
  overflow-y: auto;
}

.condition-tag {
  cursor: pointer;
}

.condition-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.key-select { width: 140px; flex-shrink: 0; }
.op-select { width: 80px; flex-shrink: 0; }
.value-select { flex: 1; min-width: 120px; }

.popover-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-top: 8px;
  border-top: 1px solid var(--el-border-color-lighter);
}

.spacer { flex: 1; }

.empty-state {
  text-align: center;
  padding: 12px 0;
}
.empty-text { color: var(--el-text-color-regular); margin: 0; }
.empty-hint { color: var(--el-text-color-secondary); font-size: 12px; margin: 4px 0 0; }

.more-hint {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.has-filters {
  color: var(--el-color-primary);
  border-color: var(--el-color-primary);
}
```

### 2. 修改组件：`ResourceListToolbar.vue`

**新增 Props**：
```ts
clusterName?: string      // 集群名称（传给 LabelFilterPopover）
resourceType?: string     // 资源类型（传给 LabelFilterPopover）
labelConditions?: LabelCondition[] // 当前 label 条件
```

**新增 Emits**：
```ts
'labelSelectorChange': [conditions: LabelCondition[]]
```

**布局改动**：在现有 filter-bar 内，命名空间选择器之后、搜索框之前，插入 LabelFilterPopover 组件。

```vue
<template>
  <el-card shadow="never" class="filter-card">
    <div class="filter-bar">
      <!-- 命名空间 -->
      <el-select v-if="showNamespace" ... />

      <!-- Label 过滤（新增） -->
      <LabelFilterPopover
        v-if="clusterName && resourceType"
        :cluster-name="clusterName"
        :namespace="namespaceValue"
        :resource-type="resourceType"
        :model-value="labelConditions || []"
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

**命名空间切换行为**：Toolbar 不需要 watch namespaceValue。命名空间切换时，列表页的 `handleNamespaceChange` 统一清空 label 条件并刷新列表（见 useResourceList 改动）。

**集群级资源**：当列表页不显示命名空间选择器时（如 Node、PV），LabelFilterPopover 的 `namespace` 传空，labels API 忽略 namespace 参数。

**样式**：Label 按钮与命名空间选择器、搜索框在同一行，Popover 弹出层不占用工具栏空间。

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
    if (continueTokens.value.length > 0) {
      body.continue = continueTokens.value[continueTokens.value.length - 1]
    }
  }
  const res = await options.fetchList(body)
  ...
}
```

**修改 `fetchNextPage`**：同样改为 POST body，保留 labelFilters。
```ts
async function fetchNextPage() {
  if (!hasMore.value || continueTokens.value.length === 0) return
  loading.value = true
  try {
    const body: any = {}
    if (selectedNamespace.value) body.namespace = selectedNamespace.value
    if (options.getClusterName) body.clusterName = options.getClusterName()
    if (labelConditions.value.length > 0) {
      body.labelFilters = labelConditions.value
    }
    body.limit = pageSize.value
    body.continue = continueTokens.value[continueTokens.value.length - 1]

    const res: any = await options.fetchList(body)
    // ... 现有分页逻辑不变
  } catch (e: any) {
    ElMessage.error(e?.message || '加载更多失败')
  } finally {
    loading.value = false
  }
}
```

**修改现有方法 `handleNamespaceChange`**：命名空间切换时，同步清空 label 条件，避免用旧条件请求新命名空间。
```ts
// 改前
function handleNamespaceChange() {
  currentPage.value = 1
  continueTokens.value = []
  fetchResources()
}

// 改后
function handleNamespaceChange() {
  labelConditions.value = [] // 新增：清空 label 条件
  currentPage.value = 1
  continueTokens.value = []
  fetchResources()
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
```

**新增 return**：
```ts
return {
  ...
  labelConditions,
  onLabelConditionsChange,
}
```

注：`clearLabelConditions` 不再单独导出，`handleNamespaceChange` 内部直接清空 `labelConditions`。

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

**兼容性说明**：
- 后端保留 GET 路由，前端改为 POST
- 如有外部 API 消费者，仍可使用 GET（不支持 labelFilters）

### 5. URL 同步

**设计**：label selector 同步到 URL query 参数，刷新页面不丢失。

**URL 格式**：JSON 编码后 base64，简洁且支持复杂结构：
```
?ls=W3sia2V5IjoiYXBwIiwib3BlcmF0b3IiOiI9IiwidmFsdWVzIjpbIm5naW54Il19XQ==
```

**实现**（不覆盖已有 query 参数）：
```ts
// 同步到 URL
watch(labelConditions, (val) => {
  const query = { ...router.currentRoute.value.query }
  if (val.length > 0) {
    query.ls = btoa(JSON.stringify(val))
  } else {
    delete query.ls
  }
  router.replace({ query })
}, { deep: true })

// 初始化时从 URL 恢复
onMounted(() => {
  const ls = router.currentRoute.value.query.ls as string
  if (ls) {
    try {
      const conditions = JSON.parse(atob(ls))
      if (Array.isArray(conditions)) {
        labelConditions.value = conditions
      }
    } catch {
      // 解析失败，静默忽略
      const query = { ...router.currentRoute.value.query }
      delete query.ls
      router.replace({ query })
    }
  }
})
```

### 6. 各列表页集成

每个列表页需要传递 `clusterName` 和 `resourceType` 给 `ResourceListToolbar`：

```vue
<ResourceListToolbar
  :cluster-name="clusterStore.clusterName"
  resource-type="deployment"
  :label-conditions="labelConditions"
  @label-selector-change="onLabelConditionsChange"
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

## 实施计划

### Phase 1：后端核心（可独立上线）
- 新增 `backend/pkg/k8s/labels/labels.go`：LabelFilter、LabelData、BuildLabelSelector、GetLabels、GVR 映射、分页采集
- 新增 `backend/pkg/k8s/labels/labels_test.go`：单元测试
- 新增 `backend/internal/k8s/labels.go`：GetLabels handler
- 修改 `backend/internal/router/k8s.go`：注册 GET /labels 路由
- 修改 `backend/pkg/k8s/deployment/api.go`：ListDeployments 增加 labelSelector 参数
- 修改 `backend/internal/k8s/deployment.go`：支持 POST + LabelFilters
- 修改 `backend/internal/router/k8s.go`：deployment/list 新增 POST 路由

### Phase 2：前端核心（依赖 Phase 1）
- 新增 `frontend/src/components/LabelFilterPopover.vue`
- 修改 `frontend/src/components/ResourceListToolbar.vue`：集成 LabelFilterPopover
- 修改 `frontend/src/composables/useResourceList.ts`：labelConditions、fetchResources/fetchNextPage 改 POST、URL 同步
- 修改 `frontend/src/api/resource.ts`：新增 getAvailableLabels、getDeploymentList 改 POST
- 修改 `views/workload/DeploymentList.vue`：传递 clusterName/resourceType/labelConditions

### Phase 3：推广到其他资源（可用脚本批量生成）
- 后端：批量改其他 19 个资源的 List 接口（pkg + handler + router）
- 前端：批量改其他 22 个列表页

## 文件变更清单

### 新增文件
| 文件 | 说明 |
|------|------|
| `frontend/src/components/LabelFilterPopover.vue` | Label Selector Popover 组件 |
| `backend/internal/k8s/labels.go` | Label 查询 handler |
| `backend/pkg/k8s/labels/labels.go` | LabelFilter + LabelData + BuildLabelSelector + GetLabels + GVR 映射 + 分页采集 |
| `backend/pkg/k8s/labels/labels_test.go` | LabelFilter 绑定测试 + BuildLabelSelector 测试 |

### 修改文件
| 文件 | 改动 |
|------|------|
| `frontend/src/components/ResourceListToolbar.vue` | 集成 LabelFilterPopover，新增 props/emits |
| `frontend/src/composables/useResourceList.ts` | labelConditions 状态、fetchResources/fetchNextPage 改 POST、URL 同步 |
| `frontend/src/api/resource.ts` | 新增 getAvailableLabels，所有 list 函数改 POST |
| `backend/internal/router/k8s.go` | 注册 GET /labels，所有 list 路由新增 POST |
| `backend/internal/k8s/*.go`（约 20 个文件） | handler 用 ShouldBind，增加 LabelFilter → labelSelector 拼接 |
| `backend/pkg/k8s/*/api.go`（约 20 个文件） | List 函数增加 labelSelector string 参数 |
| `frontend/src/views/**/*.vue`（约 23 个列表页） | 传递 clusterName/resourceType/labelConditions，list 调用改 POST |
