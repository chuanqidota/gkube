# gkube 后端架构改造实施文档

> 本文档是给 AI 执行的精确改造指令。每一步都有明确的文件路径、代码内容和验证标准。
> 执行顺序严格按阶段编号，每完成一个阶段需验证编译通过再进入下一阶段。

---

## 目录

- [阶段一：建立结构化错误体系](#阶段一建立结构化错误体系)
- [阶段二：改造 response 包支持 AppError](#阶段二改造-response-包支持-apperror)
- [阶段三：改造 handler 包装器](#阶段三改造-handler-包装器)
- [阶段四：补充集群级资源包装器](#阶段四补充集群级资源包装器)
- [阶段五：改造 pkg 层函数签名（context + AppError）](#阶段五改造-pkg-层函数签名context--apperror)
- [阶段六：用包装器替换手写 handler](#阶段六用包装器替换手写-handler)
- [阶段七：路由组级中间件](#阶段七路由组级中间件)
- [阶段八：修复关键 bug 和安全问题](#阶段八修复关键-bug-和安全问题)

---

## 阶段一：建立结构化错误体系

### 目标
创建 `pkg/errors/errors.go`，定义 `AppError` 类型和错误码常量。

### 操作

**新建文件** `backend/pkg/errors/errors.go`：

```go
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// 标准库 errors 的常用函数透传，方便统一从本包导入
var (
	As = errors.As
	Is = errors.Is
	New = errors.New
)

// ErrorCode 业务错误码，用于前端程序化区分错误类型
type ErrorCode int

const (
	ErrCodeValidation ErrorCode = 1001 // 参数校验失败
	ErrCodeNotFound   ErrorCode = 1002 // 资源不存在
	ErrCodeForbidden  ErrorCode = 1003 // 无权限
	ErrCodeConflict   ErrorCode = 1004 // 资源冲突（409）
	ErrCodeK8sClient  ErrorCode = 1005 // K8s 客户端创建失败
	ErrCodeK8sAPI     ErrorCode = 1006 // K8s API 调用失败
	ErrCodeInternal   ErrorCode = 1007 // 内部错误
)

// AppError 应用级错误，携带错误码、HTTP 状态码、脱敏消息和原始错误
type AppError struct {
	Code       ErrorCode // 业务错误码，写入 JSON body 的 code 字段
	HTTPStatus int       // HTTP 状态码
	Message    string    // 对外脱敏消息，写入 JSON body 的 msg 字段
	Err        error     // 内部原始错误，仅写入日志，不返回给客户端
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// ---------- 构造函数 ----------

// Validation 参数校验失败，HTTP 400
func Validation(msg string, err error) *AppError {
	return &AppError{Code: ErrCodeValidation, HTTPStatus: http.StatusBadRequest, Message: msg, Err: err}
}

// NotFound 资源不存在，HTTP 404
func NotFound(msg string, err error) *AppError {
	return &AppError{Code: ErrCodeNotFound, HTTPStatus: http.StatusNotFound, Message: msg, Err: err}
}

// Forbidden 无权限，HTTP 403
func Forbidden(msg string, err error) *AppError {
	return &AppError{Code: ErrCodeForbidden, HTTPStatus: http.StatusForbidden, Message: msg, Err: err}
}

// Conflict 资源冲突，HTTP 409
func Conflict(msg string, err error) *AppError {
	return &AppError{Code: ErrCodeConflict, HTTPStatus: http.StatusConflict, Message: msg, Err: err}
}

// K8sClientFail K8s 客户端创建失败，HTTP 502
func K8sClientFail(msg string, err error) *AppError {
	return &AppError{Code: ErrCodeK8sClient, HTTPStatus: http.StatusBadGateway, Message: msg, Err: err}
}

// K8sAPIFail K8s API 调用失败，HTTP 502
func K8sAPIFail(msg string, err error) *AppError {
	return &AppError{Code: ErrCodeK8sAPI, HTTPStatus: http.StatusBadGateway, Message: msg, Err: err}
}

// Internal 内部错误，HTTP 500
func Internal(msg string, err error) *AppError {
	return &AppError{Code: ErrCodeInternal, HTTPStatus: http.StatusInternalServerError, Message: msg, Err: err}
}
```

### 验证
```bash
cd /data/gkube/backend && go build ./pkg/errors/...
```

---

## 阶段二：改造 response 包支持 AppError

### 目标
在 `pkg/response/api.go` 中新增 `FailWithError` 方法，自动根据 `AppError` 类型选择 HTTP 状态码和错误码。

### 操作

**修改文件** `backend/pkg/response/api.go`，在现有代码末尾（`File` 函数之后）追加：

```go
// FailWithError 根据 error 类型自动选择响应格式。
//   - *apperr.AppError → 使用其 HTTPStatus 和 Code，msg 脱敏，原始错误写日志
//   - 其他 error → HTTP 500 + code=1007，原始错误写日志
func FailWithError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	var appErr *apperr.AppError
	if errors.As(err, &appErr) {
		if appErr.Err != nil {
			logger.Error(fmt.Sprintf("%s: %s", appErr.Message, appErr.Err.Error()))
		} else {
			logger.Error(appErr.Message)
		}
		c.JSON(appErr.HTTPStatus, gin.H{
			"msg":  appErr.Message,
			"code": int(appErr.Code),
			"data": nil,
		})
		c.Abort()
		return
	}
	// 未知错误，兜底 500
	logger.Error(fmt.Sprintf("未知错误: %s", err.Error()))
	c.JSON(http.StatusInternalServerError, gin.H{
		"msg":  "服务器内部错误",
		"code": int(apperr.ErrCodeInternal),
		"data": nil,
	})
	c.Abort()
}
```

同时修改文件头部 import，增加：

```go
import (
	"errors"
	// ... 其他已有 import ...
	apperr "gkube/pkg/errors"
)
```

### 验证
```bash
cd /data/gkube/backend && go build ./pkg/response/...
```

---

## 阶段三：改造 handler 包装器

### 目标
改造 `internal/k8s/handler.go` 中的 5 个包装器：
1. 增加 `context.Context` 参数传递
2. 使用 `response.FailWithError` 替代手动 `logger.Error` + `FailWithStatus`

### 操作

**替换整个文件** `backend/internal/k8s/handler.go` 为：

```go
package k8s

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	apperr "gkube/pkg/errors"
	k8sclient "gkube/pkg/k8s"
	k8sLabels "gkube/pkg/k8s/labels"
	"gkube/pkg/response"
	"k8s.io/client-go/kubernetes"
)

// ---------------------------------------------------------------------------
// Handler wrapper 函数 — 消除 bind→client→call→respond 样板代码
//
// 每个 wrapper 统一处理：参数绑定、K8s 客户端获取、label selector 构建、
// 错误日志、标准化响应。业务逻辑通过 fn 闭包注入。
//
// fn 的 error 返回值支持两种类型：
//   - *apperr.AppError → 自动使用其 HTTP 状态码和业务错误码
//   - 其他 error → 兜底 HTTP 500
// ---------------------------------------------------------------------------

// ListHandler 分页列表 handler。
// 绑定 ListParams（GET+POST），构建 label selector，将 Limit/Continue 传给 fn。
func ListHandler(
	fn func(ctx context.Context, client *kubernetes.Clientset, namespace, labelSelector string, limit int64, continueToken string) (any, error),
	successMsg, failMsg string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p ListParams
		if err := c.ShouldBind(&p); err != nil {
			response.Fail(c, "参数校验失败")
			return
		}
		client, err := k8sclient.GetK8sClientByName(p.ClusterName)
		if err != nil {
			response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
			return
		}
		selector, err := buildLabelSelector(p.LabelFilters)
		if err != nil {
			response.Fail(c, err.Error())
			return
		}
		data, err := fn(c.Request.Context(), client, p.Namespace, selector, p.Limit, p.Continue)
		if err != nil {
			response.FailWithError(c, err)
			return
		}
		response.Success(c, successMsg, data)
	}
}

// NamespacedHandler 按 namespace + name 获取资源的 handler（GET 请求）。
func NamespacedHandler(
	fn func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error),
	successMsg, failMsg string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p NamespacedParams
		if err := c.ShouldBindQuery(&p); err != nil {
			response.Fail(c, "参数校验失败")
			return
		}
		client, err := k8sclient.GetK8sClientByName(p.ClusterName)
		if err != nil {
			response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
			return
		}
		data, err := fn(c.Request.Context(), client, p.Namespace, p.Name)
		if err != nil {
			response.FailWithError(c, err)
			return
		}
		response.Success(c, successMsg, data)
	}
}

// CreateHandler 从 YAML 创建资源的 handler（POST 请求）。
func CreateHandler(
	fn func(ctx context.Context, client *kubernetes.Clientset, namespace, yaml string) error,
	successMsg, failMsg string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p CreateParams
		if err := c.ShouldBindJSON(&p); err != nil {
			response.Fail(c, "参数校验失败")
			return
		}
		client, err := k8sclient.GetK8sClientByName(p.ClusterName)
		if err != nil {
			response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
			return
		}
		if err := fn(c.Request.Context(), client, p.Namespace, p.Yaml); err != nil {
			response.FailWithError(c, err)
			return
		}
		response.Success(c, successMsg, nil)
	}
}

// UpdateHandler 从 YAML 更新资源的 handler（PUT 请求）。
func UpdateHandler(
	fn func(ctx context.Context, client *kubernetes.Clientset, namespace, name, yaml string) error,
	successMsg, failMsg string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p UpdateParams
		if err := c.ShouldBindJSON(&p); err != nil {
			response.Fail(c, "参数校验失败")
			return
		}
		client, err := k8sclient.GetK8sClientByName(p.ClusterName)
		if err != nil {
			response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
			return
		}
		if err := fn(c.Request.Context(), client, p.Namespace, p.Name, p.Yaml); err != nil {
			response.FailWithError(c, err)
			return
		}
		response.Success(c, successMsg, nil)
	}
}

// DeleteHandler 按 namespace + name 删除资源的 handler（DELETE 请求）。
func DeleteHandler(
	fn func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error,
	successMsg, failMsg string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p NamespacedParams
		if err := c.ShouldBindJSON(&p); err != nil {
			response.Fail(c, "参数校验失败")
			return
		}
		client, err := k8sclient.GetK8sClientByName(p.ClusterName)
		if err != nil {
			response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
			return
		}
		if err := fn(c.Request.Context(), client, p.Namespace, p.Name); err != nil {
			response.FailWithError(c, err)
			return
		}
		response.Success(c, successMsg, nil)
	}
}

// ---------------------------------------------------------------------------
// 内部 helper
// ---------------------------------------------------------------------------

func buildLabelSelector(filters []k8sLabels.LabelFilter) (string, error) {
	if len(filters) == 0 {
		return "", nil
	}
	return k8sLabels.BuildLabelSelector(filters)
}
```

### 验证
```bash
cd /data/gkube/backend && go build ./internal/k8s/...
```

> **注意**：此时所有调用包装器的 handler（如 deployment.go 中的 `var GetDeploymentList = ListHandler(...)`）会因为函数签名不匹配而编译失败。这是预期的，阶段五会修复 pkg 层签名，阶段六会修复 handler 层调用。

---

## 阶段四：补充集群级资源包装器

### 目标
在 `internal/k8s/handler.go` 末尾追加 4 个集群级资源专用包装器，消除 PV/StorageClass/Node/VolumeSnapshotClass 等文件的重复代码。

### 操作

**在 `backend/internal/k8s/handler.go` 的 `buildLabelSelector` 函数之后追加**：

```go
// ---------------------------------------------------------------------------
// 集群级资源 wrapper —— 无 namespace 参数
// ---------------------------------------------------------------------------

// ClusterListHandler 集群级资源的分页列表 handler。
// 绑定 ClusterListParams（GET+POST），构建 label selector。
func ClusterListHandler(
	fn func(ctx context.Context, client *kubernetes.Clientset, labelSelector string, limit int64, continueToken string) (any, error),
	successMsg, failMsg string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p ClusterListParams
		if err := c.ShouldBind(&p); err != nil {
			response.Fail(c, "参数校验失败")
			return
		}
		client, err := k8sclient.GetK8sClientByName(p.ClusterName)
		if err != nil {
			response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
			return
		}
		selector, err := buildLabelSelector(p.LabelFilters)
		if err != nil {
			response.Fail(c, err.Error())
			return
		}
		data, err := fn(c.Request.Context(), client, selector, p.Limit, p.Continue)
		if err != nil {
			response.FailWithError(c, err)
			return
		}
		response.Success(c, successMsg, data)
	}
}

// ClusterGetHandler 集群级资源的详情/YAML 获取 handler（GET 请求）。
// 绑定 ClusterScopedParams via ShouldBindQuery。
func ClusterGetHandler(
	fn func(ctx context.Context, client *kubernetes.Clientset, name string) (any, error),
	successMsg, failMsg string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p ClusterScopedParams
		if err := c.ShouldBindQuery(&p); err != nil {
			response.Fail(c, "参数校验失败")
			return
		}
		client, err := k8sclient.GetK8sClientByName(p.ClusterName)
		if err != nil {
			response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
			return
		}
		data, err := fn(c.Request.Context(), client, p.Name)
		if err != nil {
			response.FailWithError(c, err)
			return
		}
		response.Success(c, successMsg, data)
	}
}

// ClusterCreateHandler 集群级资源的创建 handler（POST 请求）。
// 绑定 ClusterCreateParams via ShouldBindJSON。
func ClusterCreateHandler(
	fn func(ctx context.Context, client *kubernetes.Clientset, yaml string) error,
	successMsg, failMsg string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p ClusterCreateParams
		if err := c.ShouldBindJSON(&p); err != nil {
			response.Fail(c, "参数校验失败")
			return
		}
		client, err := k8sclient.GetK8sClientByName(p.ClusterName)
		if err != nil {
			response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
			return
		}
		if err := fn(c.Request.Context(), client, p.Yaml); err != nil {
			response.FailWithError(c, err)
			return
		}
		response.Success(c, successMsg, nil)
	}
}

// ClusterDeleteHandler 集群级资源的删除 handler（DELETE 请求）。
// 绑定 ClusterScopedParams via ShouldBindJSON。
func ClusterDeleteHandler(
	fn func(ctx context.Context, client *kubernetes.Clientset, name string) error,
	successMsg, failMsg string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p ClusterScopedParams
		if err := c.ShouldBindJSON(&p); err != nil {
			response.Fail(c, "参数校验失败")
			return
		}
		client, err := k8sclient.GetK8sClientByName(p.ClusterName)
		if err != nil {
			response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
			return
		}
		if err := fn(c.Request.Context(), client, p.Name); err != nil {
			response.FailWithError(c, err)
			return
		}
		response.Success(c, successMsg, nil)
	}
}
```

**同时在 `backend/internal/k8s/params.go` 中追加 `ClusterListParams`**（在 `ClusterCreateParams` 之后）：

```go
// ClusterListParams 集群级资源的分页列表参数
// 适用：PV, StorageClass, Node, VolumeSnapshotClass
type ClusterListParams struct {
	ClusterName  string                  `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Limit        int64                   `form:"limit" json:"limit" label:"每页条数"`
	Continue     string                  `form:"continue" json:"continue" label:"分页标记"`
	LabelFilters []k8sLabels.LabelFilter `json:"labelFilters" form:"labelFilters" label:"标签过滤"`
}
```

### 验证
```bash
cd /data/gkube/backend && go build ./internal/k8s/...
```

---

## 阶段五：改造 pkg 层函数签名（context + AppError）

### 目标
给 `pkg/k8s/` 下每个资源包的每个公开函数：
1. 增加 `ctx context.Context` 作为第一个参数
2. 将 `context.TODO()` 替换为 `ctx`
3. 将错误返回值改为 `*apperr.AppError`（K8s notFound → `apperr.NotFound`，其他 → `apperr.K8sAPIFail`）

### 通用改动模式

以 `pkg/k8s/deployment/api.go` 为例，**每个函数的改动模式相同**：

**改动前**：
```go
func GetDeploymentDetail(client *kubernetes.Clientset, namespace, name string) (*appsv1.Deployment, error) {
	deployment, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取deployment详情失败:%s", err.Error())
	}
	return deployment, nil
}
```

**改动后**：
```go
func GetDeploymentDetail(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (*appsv1.Deployment, error) {
	deployment, err := client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, apperr.NotFound("deployment不存在", err)
		}
		return nil, apperr.K8sAPIFail("获取deployment详情失败", err)
	}
	return deployment, nil
}
```

### 每个资源包需要改动的文件清单

改动规则统一：
- 第一个参数增加 `ctx context.Context`
- 所有 `context.TODO()` 和 `context.Background()` 替换为 `ctx`
- 所有 `fmt.Errorf("xxx:%s", err.Error())` 替换为 `apperr.K8sAPIFail("xxx", err)`
- 所有 `Get` 操作增加 `k8serrors.IsNotFound(err)` 判断，返回 `apperr.NotFound`
- import 增加 `apperr "gkube/pkg/errors"` 和 `"k8s.io/apimachinery/pkg/api/errors" as k8serrors`
- import 移除 `"context"` (不再需要 context.TODO)

**需要改动的文件**：

| 文件 | 函数数量 | 备注 |
|------|---------|------|
| `pkg/k8s/deployment/api.go` | 12 | 含 ListDeployments, GetDeploymentYaml, CreateDeployment, UpdateDeployment, DeleteDeployment, ScaleDeployment, RestartDeployment, UpdateDeploymentImage, GetDeploymentDetail, RollbackDeployment, GetDeploymentPods, GetDeploymentReplicaSets, GetDeploymentEvents |
| `pkg/k8s/pod/api.go` | ~10 | |
| `pkg/k8s/service/api.go` | ~8 | CreateService 需要额外修复 namespace 覆盖 |
| `pkg/k8s/configmap/api.go` | ~5 | |
| `pkg/k8s/secret/api.go` | ~5 | |
| `pkg/k8s/statefulset/api.go` | ~10 | |
| `pkg/k8s/daemonset/api.go` | ~8 | |
| `pkg/k8s/job/api.go` | ~6 | |
| `pkg/k8s/cronjob/api.go` | ~8 | |
| `pkg/k8s/replicaset/api.go` | ~5 | |
| `pkg/k8s/hpa/api.go` | ~8 | |
| `pkg/k8s/ingress/api.go` | ~6 | |
| `pkg/k8s/networkpolicy/api.go` | ~6 | |
| `pkg/k8s/pv/api.go` | 8 | GetPVList 签名需增加 limit/continue 参数 |
| `pkg/k8s/pvc/api.go` | ~6 | |
| `pkg/k8s/storageclass/api.go` | ~5 | UpdateStorageClass 缺少 retry.RetryOnConflict，需补上 |
| `pkg/k8s/volumesnapshot/api.go` | ~5 | |
| `pkg/k8s/volumesnapshotclass/api.go` | ~5 | |
| `pkg/k8s/node/api.go` | ~8 | |
| `pkg/k8s/namespace/api.go` | ~6 | |
| `pkg/k8s/event/api.go` | ~3 | |
| `pkg/k8s/crd/api.go` | ~8 | |
| `pkg/k8s/limitrange/api.go` | ~5 | |
| `pkg/k8s/resourcequota/api.go` | ~5 | |
| `pkg/k8s/cluster/api.go` | ~2 | |
| `pkg/k8s/container/api.go` | ~3 | |

### 特殊改动：pv/api.go 的 GetPVList

当前签名不支持分页，需要对齐：

**改动前**：
```go
func GetPVList(client *kubernetes.Clientset, labelSelector string) ([]corev1.PersistentVolume, error) {
```

**改动后**：
```go
func GetPVList(ctx context.Context, client *kubernetes.Clientset, labelSelector string, limit int64, continueToken string) (*corev1.PersistentVolumeList, error) {
	listOpts := metav1.ListOptions{ResourceVersion: "0"}
	if labelSelector != "" {
		listOpts.LabelSelector = labelSelector
	}
	if limit > 0 {
		listOpts.Limit = limit
	}
	if continueToken != "" {
		listOpts.Continue = continueToken
	}
	result, err := client.CoreV1().PersistentVolumes().List(ctx, listOpts)
	if err != nil {
		return nil, apperr.K8sAPIFail("获取PV列表失败", err)
	}
	return result, nil
}
```

> **注意**：返回类型从 `[]corev1.PersistentVolume` 改为 `*corev1.PersistentVolumeList`，handler 层需要相应调整。

### 验证
```bash
cd /data/gkube/backend && go build ./pkg/k8s/...
```

---

## 阶段六：用包装器替换手写 handler

### 目标
将所有手写的 handler 函数替换为包装器调用，同时适配阶段五的新签名。

### 通用改动模式

以 `deployment.go` 为例，**所有闭包增加 `ctx context.Context` 参数**：

**改动前**：
```go
var GetDeploymentList = ListHandler(
	func(client *kubernetes.Clientset, namespace, selector string, limit int64, continueToken string) (any, error) {
		list, err := k8sDeployment.ListDeployments(client, namespace, limit, continueToken, selector)
		// ...
	},
	"获取deployment列表成功", "获取deployment列表失败",
)
```

**改动后**：
```go
var GetDeploymentList = ListHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, selector string, limit int64, continueToken string) (any, error) {
		list, err := k8sDeployment.ListDeployments(ctx, client, namespace, limit, continueToken, selector)
		if err != nil {
			return nil, err
		}
		remaining := int64(0)
		if list.RemainingItemCount != nil {
			remaining = *list.RemainingItemCount
		}
		data := k8sclient.BuildPaginatedData(list.Items, list.Continue, remaining, limit)
		data.Total = len(list.Items) + int(remaining)
		return data, nil
	},
	"获取deployment列表成功", "获取deployment列表失败",
)
```

### 需要改动的文件清单

#### A. 使用标准 wrapper 的文件（增加 ctx 参数）

| 文件 | 包装器类型 |
|------|-----------|
| `deployment.go` | ListHandler, NamespacedHandler, CreateHandler, UpdateHandler, DeleteHandler |
| `statefulset.go` | 同上 |
| `daemonset.go` | 同上 |
| `job.go` | 同上 |
| `cronjob.go` | 同上 |
| `service.go` | 同上 |
| `pod.go` | 同上 |
| `configmap.go` | 同上 |
| `secret.go` | 同上 |
| `pvc.go` | 同上 |
| `replicaset.go` | 同上 |
| `hpa.go` | 同上 |
| `ingress.go` | 同上 |
| `networkpolicy.go` | 同上 |
| `namespace.go` | 同上 |
| `limitrange.go` | 同上 |
| `resourcequota.go` | 同上 |

#### B. 集群级资源文件（替换为集群级 wrapper）

**`pv.go`** — 完整替换为：

```go
package k8s

import (
	"context"

	k8sclient "gkube/pkg/k8s"
	k8sPv "gkube/pkg/k8s/pv"
	"k8s.io/client-go/kubernetes"
)

var GetPVList = ClusterListHandler(
	func(ctx context.Context, client *kubernetes.Clientset, selector string, limit int64, continueToken string) (any, error) {
		list, err := k8sPv.GetPVList(ctx, client, selector, limit, continueToken)
		if err != nil {
			return nil, err
		}
		remaining := int64(0)
		if list.RemainingItemCount != nil {
			remaining = *list.RemainingItemCount
		}
		data := k8sclient.BuildPaginatedData(list.Items, list.Continue, remaining, limit)
		data.Total = len(list.Items) + int(remaining)
		return data, nil
	},
	"获取PV列表成功", "获取PV列表失败",
)

var GetPVByName = ClusterGetHandler(
	func(ctx context.Context, client *kubernetes.Clientset, name string) (any, error) {
		return k8sPv.GetPVByName(ctx, client, name)
	},
	"获取PV详情成功", "获取PV详情失败",
)

var GetPVYaml = ClusterGetHandler(
	func(ctx context.Context, client *kubernetes.Clientset, name string) (any, error) {
		yaml, err := k8sPv.GetPVYaml(ctx, client, name)
		if err != nil {
			return nil, err
		}
		return map[string]string{"yaml": yaml}, nil
	},
	"获取PV YAML成功", "获取PV YAML失败",
)

var CreatePV = ClusterCreateHandler(
	func(ctx context.Context, client *kubernetes.Clientset, yaml string) error {
		return k8sPv.CreatePV(ctx, client, yaml)
	},
	"创建PV成功", "创建PV失败",
)

var UpdatePV = ClusterCreateHandler(
	func(ctx context.Context, client *kubernetes.Clientset, yaml string) error {
		return k8sPv.UpdatePV(ctx, client, yaml)
	},
	"更新PV成功", "更新PV失败",
)

var DeletePVByName = ClusterDeleteHandler(
	func(ctx context.Context, client *kubernetes.Clientset, name string) error {
		return k8sPv.DeletePVByName(ctx, client, name)
	},
	"删除PV成功", "删除PV失败",
)
```

**`storageclass.go`** — 同样替换为 wrapper 调用，但 `GetStorageClassEvents` 需要保留在 handler 层（或移到 pkg 层）。推荐做法：先移到 `pkg/k8s/storageclass/api.go` 中新增 `GetStorageClassEvents` 函数，然后 handler 用 `ClusterGetHandler` 调用。

**`volumesnapshot.go`** — 替换为 wrapper，同时删除本地的 `getDynamicClient` 函数，改用 `k8sclient.GetDynamicClientByName`。

**`volumesnapshotclass.go`** — 同上。

**`node.go`** — 部分操作（CordonNode, DrainNode, UpdateNodeTaints 等）参数超出 wrapper 覆盖范围，保留手写但使用 `response.FailWithError`。

#### C. 特殊 handler（保留手写，改用 FailWithError）

以下 handler 的参数结构超出 wrapper 覆盖范围，保留手写但统一错误处理：

- `deployment.go` 中的 `ScaleDeployment`, `RollbackDeployment`, `UpdateDeploymentImage`
- `statefulset.go` 中的 `ScaleStatefulSet`, `RollbackStatefulSet`, `UpdateStatefulSetImage`, `RestartStatefulSet`
- `daemonset.go` 中的 `RestartDaemonSet`, `RollbackDaemonSet`, `UpdateDaemonSetImage`
- `node.go` 中的所有操作
- `cronjob.go` 中的 `SuspendCronJob`, `ResumeCronJob`, `TriggerCronJob`
- `hpa.go` 中的 `PauseHPA`, `ResumeHPA`

这些 handler 的改动模式统一为：

**改动前**：
```go
func ScaleDeployment(c *gin.Context) {
	// ... bind ...
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	if err := k8sDeployment.ScaleDeployment(client, body.Namespace, body.Name, body.Replicas); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "扩缩容deployment失败")
		return
	}
	// ...
}
```

**改动后**：
```go
func ScaleDeployment(c *gin.Context) {
	// ... bind (不变) ...
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	if err := k8sDeployment.ScaleDeployment(c.Request.Context(), client, body.Namespace, body.Name, body.Replicas); err != nil {
		response.FailWithError(c, err)
		return
	}
	// ...
}
```

统一规则：
1. 删除 `logger.Error(err.Error())`
2. `response.FailWithStatus(c, http.StatusBadGateway, "xxx")` → `response.FailWithError(c, apperr.K8sClientFail("xxx", err))`（客户端错误）
3. pkg 函数返回的 error 直接传给 `response.FailWithError(c, err)`（它已经是 AppError 了）
4. 增加 `c.Request.Context()` 作为 pkg 函数的第一个参数

### 验证
```bash
cd /data/gkube/backend && go build ./...
```

---

## 阶段七：路由组级中间件

### 目标
将 `RequirePermission()` 从每个路由调用移到路由组级别。

### 操作

**修改文件** `backend/internal/router/k8s.go`：

**改动前**：
```go
func registerK8sRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("k8s")
	{
		grp.GET("labels", middleware.RequirePermission(), k8s.Label.GetLabels)
		registerCoreRoutes(grp)
		// ...
	}
}

func registerCoreRoutes(rg *gin.RouterGroup) {
	rg.GET("cluster/version", middleware.RequirePermission(), k8s.Cluster.GetClusterVersion)
	// ... 每个路由都重复 RequirePermission() ...
}
```

**改动后**：
```go
func registerK8sRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("k8s")
	grp.Use(middleware.RequirePermission()) // 整个组统一应用
	{
		grp.GET("labels", k8s.Label.GetLabels)
		registerCoreRoutes(grp)
		registerWorkloadRoutes(grp)
		registerNetworkRoutes(grp)
		registerStorageRoutes(grp)
		registerConfigRoutes(grp)
		registerCrdRoutes(grp)
	}
	// 审计路由单独分组（清除操作需要 Admin）
	auditGroup := rg.Group("k8s/audit")
	auditGroup.Use(middleware.RequirePermission())
	{
		auditGroup.GET("list", k8s.Audit.ListAuditLogs)
		auditGroup.GET("detail", k8s.Audit.GetAuditLog)
		auditGroup.GET("stats", k8s.Audit.GetAuditStats)
	}
	// 审计清除需要更高权限，单独注册
	rg.DELETE("k8s/audit/clear", middleware.RequireAdmin(), k8s.Audit.ClearAuditLogs)
}

func registerCoreRoutes(rg *gin.RouterGroup) {
	// 所有路由去掉 middleware.RequirePermission() 前缀
	rg.GET("cluster/version", k8s.Cluster.GetClusterVersion)
	rg.GET("cluster/nodes", k8s.Cluster.GetClusterNodesInfo)

	rg.GET("node/detail", k8s.Node.GetNodeDetail)
	rg.GET("node/get-yaml", k8s.Node.GetNodeYaml)
	// ... 其余路由同样去掉 middleware.RequirePermission() ...
}
```

**每个 registerXxxRoutes 函数内的所有路由都去掉 `middleware.RequirePermission()` 前缀。**

### 验证
```bash
cd /data/gkube/backend && go build ./...
```

---

## 阶段八：修复关键 bug 和安全问题

### 目标
修复审查中发现的高优先级 bug。

### 8.1 添加 SIGTERM 信号处理

**文件** `backend/cmd/root.go`

**改动前**：
```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, os.Interrupt)
```

**改动后**：
```go
import "syscall"

quit := make(chan os.Signal, 1)
signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
```

### 8.2 修复 OOMKilled 检测

**文件** `backend/internal/dashboard/dashboard.go`

找到 `abnormalWaitingReasons` 定义，将 `OOMKilled` 从中移除：

**改动前**：
```go
abnormalWaitingReasons := map[string]bool{
	"CrashLoopBackOff": true,
	"ImagePullBackOff": true,
	"ErrImagePull":     true,
	"CreateContainerConfigError": true,
	"OOMKilled":        true,
}
```

**改动后**：
```go
abnormalWaitingReasons := map[string]bool{
	"CrashLoopBackOff": true,
	"ImagePullBackOff": true,
	"ErrImagePull":     true,
	"CreateContainerConfigError": true,
}

// OOMKilled 是 Terminated 状态的 reason，不是 Waiting 状态
abnormalTerminatedReasons := map[string]bool{
	"OOMKilled": true,
	"Error":     true,
}
```

然后找到检查 `abnormalWaitingReasons` 的代码段，在其附近增加对 `abnormalTerminatedReasons` 的检查：

```go
// 在现有 Waiting 检查之后增加 Terminated 检查
if cs.State.Terminated != nil && abnormalTerminatedReasons[cs.State.Terminated.Reason] {
    abnormal = append(abnormal, map[string]any{
        "container": cs.Name,
        "reason":    cs.State.Terminated.Reason,
        "message":   cs.State.Terminated.Message,
        "state":     "terminated",
    })
}
```

### 8.3 修复 VolumeSnapshot 动态客户端缓存绕过

**文件** `backend/internal/k8s/volumesnapshot.go`

删除本地的 `getDynamicClient` 函数，改用 `k8sclient.GetDynamicClientByName`。

**改动前**：
```go
func getDynamicClient(clusterName string) (dynamic.Interface, error) {
	config, err := k8sclient.GetRestConfigByName(clusterName)
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(config)
}
```

**改动后**：删除此函数，所有调用处改为 `k8sclient.GetDynamicClientByName(clusterName)`。

### 8.4 统一 DELETE 参数绑定

所有 DELETE handler 统一使用 `ShouldBindJSON`（已在包装器中统一）。对于手写的 DELETE handler（如 `DeleteHPA`），将 `ShouldBindQuery` 改为 `ShouldBindJSON`。

### 8.5 统一日志到 pkg/logger

**文件** `backend/internal/cluster/health.go`

将 `import logrus "github.com/sirupsen/logrus"` 替换为 `import "gkube/pkg/logger"`，将所有 `logrus.Error(...)` 替换为 `logger.Error(...)`。

**文件** `backend/internal/k8s/hpa.go`

将 `import "log"` 替换为 `import "gkube/pkg/logger"`，将 `log.Printf(...)` 替换为 `logger.Error(...)` 或 `logger.Info(...)`。

### 8.6 修复 Service Create 不覆盖 namespace

**文件** `backend/pkg/k8s/service/api.go` 的 `CreateService` 函数

在 `yaml.Unmarshal` 之后增加 `service.Namespace = namespace`。

### 8.7 添加登录限流

**新建文件** `backend/pkg/middleware/ratelimit.go`：

```go
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gkube/pkg/response"
)

// ipLimiter 基于 IP 的简易限流器（生产环境建议用 Redis）
type ipLimiter struct {
	mu       sync.Mutex
	attempts map[string]*attempt
}

type attempt struct {
	count    int
	lastSeen time.Time
}

var loginLimiter = &ipLimiter{
	attempts: make(map[string]*attempt),
}

// RateLimit 登录限流中间件：每个 IP 每分钟最多 10 次尝试
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		loginLimiter.mu.Lock()
		a, exists := loginLimiter.attempts[ip]
		if !exists {
			loginLimiter.attempts[ip] = &attempt{count: 1, lastSeen: time.Now()}
			loginLimiter.mu.Unlock()
			c.Next()
			return
		}
		if time.Since(a.lastSeen) > time.Minute {
			// 窗口重置
			a.count = 1
			a.lastSeen = time.Now()
			loginLimiter.mu.Unlock()
			c.Next()
			return
		}
		a.count++
		a.lastSeen = time.Now()
		if a.count > 10 {
			loginLimiter.mu.Unlock()
			response.FailWithStatus(c, http.StatusTooManyRequests, "请求过于频繁，请稍后再试")
			return
		}
		loginLimiter.mu.Unlock()
		c.Next()
	}
}
```

**修改文件** `backend/internal/router/auth.go`：

```go
// 改动前
rg.POST("auth/login", auth.Auth.Login)

// 改动后
rg.POST("auth/login", middleware.RateLimit(), auth.Auth.Login)
```

### 验证
```bash
cd /data/gkube/backend && go build ./...
```

---

## 最终验证清单

完成所有阶段后，执行以下验证：

```bash
# 1. 编译通过
cd /data/gkube/backend && go build ./...

# 2. 检查无遗留的 context.TODO()
grep -rn "context.TODO()" pkg/k8s/ internal/k8s/ --include="*.go" | grep -v "_test.go"
# 期望：无输出

# 3. 检查无遗留的 502 状态码用于参数错误
grep -rn "StatusBadGateway" internal/k8s/ --include="*.go" | grep "参数"
# 期望：无输出

# 4. 检查所有 pkg 函数都有 ctx 参数
grep -rn "^func [A-Z]" pkg/k8s/ --include="*.go" | grep -v "ctx context.Context" | grep -v "_test.go" | grep -v "func init"
# 期望：无输出（或只有不需要 ctx 的 helper 函数）

# 5. 检查无遗留的直接 logrus 调用
grep -rn "logrus\." internal/ --include="*.go" | grep -v "_test.go"
# 期望：无输出

# 6. 检查 response.FailWithError 被使用
grep -rn "FailWithError" internal/k8s/ --include="*.go" | wc -l
# 期望：> 0

# 7. 检查包装器被使用
grep -rn "ClusterGetHandler\|ClusterListHandler\|ClusterCreateHandler\|ClusterDeleteHandler" internal/k8s/ --include="*.go" | wc -l
# 期望：> 0
```

---

## 文件变更汇总

| 操作 | 文件 |
|------|------|
| **新建** | `pkg/errors/errors.go` |
| **新建** | `pkg/middleware/ratelimit.go` |
| **重写** | `internal/k8s/handler.go` |
| **重写** | `internal/k8s/pv.go` |
| **重写** | `internal/k8s/storageclass.go` |
| **重写** | `internal/k8s/volumesnapshot.go` |
| **重写** | `internal/k8s/volumesnapshotclass.go` |
| **追加** | `internal/k8s/params.go`（增加 ClusterListParams） |
| **追加** | `pkg/response/api.go`（增加 FailWithError） |
| **修改** | `internal/router/k8s.go`（路由组级中间件） |
| **修改** | `internal/router/auth.go`（增加限流） |
| **修改** | `cmd/root.go`（增加 SIGTERM） |
| **修改** | `internal/dashboard/dashboard.go`（OOMKilled） |
| **修改** | `internal/cluster/health.go`（统一 logger） |
| **修改** | `internal/k8s/hpa.go`（统一 logger） |
| **修改** | `pkg/k8s/service/api.go`（namespace 覆盖） |
| **修改** | `pkg/k8s/` 下所有 `*/api.go`（ctx + AppError） |
| **修改** | `internal/k8s/` 下所有 handler 文件（ctx + wrapper） |
