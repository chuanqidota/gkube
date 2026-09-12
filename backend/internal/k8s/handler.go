package k8s

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	apperr "gkube/pkg/errors"
	k8sclient "gkube/pkg/k8s"
	k8sLabels "gkube/pkg/k8s/labels"
	"gkube/pkg/response"
	"k8s.io/client-go/kubernetes"
)

// ensureAppError 确保 err 为 *apperr.AppError；若为 raw error 则包装为 K8sAPIFail（HTTP 502）。
func ensureAppError(err error) *apperr.AppError {
	var ae *apperr.AppError
	if errors.As(err, &ae) {
		return ae
	}
	return apperr.K8sAPIFail("K8s API 调用失败", err)
}

// ---------------------------------------------------------------------------
// Handler wrapper 函数 — 消除 bind→client→call→respond 样板代码
//
// 每个 wrapper 统一处理：参数绑定、K8s 客户端获取、label selector 构建、
// 错误日志、标准化响应。业务逻辑通过 fn 闭包注入。
//
// fn 的 error 返回值支持两种类型：
//   - *apperr.AppError → 自动使用其 HTTP 状态码和业务错误码
//   - 其他 error → 自动包装为 HTTP 502（K8sAPIFail）
// ---------------------------------------------------------------------------

// ListHandler 分页列表 handler。
// 绑定 ListParams（GET+POST），构建 label selector，将 Limit/Continue 传给 fn。
func ListHandler(
	fn func(ctx context.Context, client *kubernetes.Clientset, namespace, labelSelector string, limit int64, continueToken string) (any, error),
	successMsg string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p ListParams
		if err := c.ShouldBind(&p); err != nil {
			response.FailWithError(c, apperr.Validation("参数校验失败", err))
			return
		}
		client, err := k8sclient.GetK8sClientByName(p.ClusterName)
		if err != nil {
			response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
			return
		}
		selector, err := buildLabelSelector(p.LabelFilters)
		if err != nil {
			response.FailWithError(c, apperr.Validation("标签选择器错误", err))
			return
		}
		data, err := fn(c.Request.Context(), client, p.Namespace, selector, p.Limit, p.Continue)
		if err != nil {
			response.FailWithError(c, ensureAppError(err))
			return
		}
		response.Success(c, successMsg, data)
	}
}

// NamespacedHandler 按 namespace + name 获取资源的 handler（GET 请求）。
func NamespacedHandler(
	fn func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error),
	successMsg string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p NamespacedParams
		if err := c.ShouldBindQuery(&p); err != nil {
			response.FailWithError(c, apperr.Validation("参数校验失败", err))
			return
		}
		client, err := k8sclient.GetK8sClientByName(p.ClusterName)
		if err != nil {
			response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
			return
		}
		data, err := fn(c.Request.Context(), client, p.Namespace, p.Name)
		if err != nil {
			response.FailWithError(c, ensureAppError(err))
			return
		}
		response.Success(c, successMsg, data)
	}
}

// CreateHandler 从 YAML 创建资源的 handler（POST 请求）。
func CreateHandler(
	fn func(ctx context.Context, client *kubernetes.Clientset, namespace, yaml string) error,
	successMsg string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p CreateParams
		if err := c.ShouldBindJSON(&p); err != nil {
			response.FailWithError(c, apperr.Validation("参数校验失败", err))
			return
		}
		client, err := k8sclient.GetK8sClientByName(p.ClusterName)
		if err != nil {
			response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
			return
		}
		if err := fn(c.Request.Context(), client, p.Namespace, p.Yaml); err != nil {
			response.FailWithError(c, ensureAppError(err))
			return
		}
		response.Success(c, successMsg, nil)
	}
}

// UpdateHandler 从 YAML 更新资源的 handler（PUT 请求）。
func UpdateHandler(
	fn func(ctx context.Context, client *kubernetes.Clientset, namespace, name, yaml string) error,
	successMsg string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p UpdateParams
		if err := c.ShouldBindJSON(&p); err != nil {
			response.FailWithError(c, apperr.Validation("参数校验失败", err))
			return
		}
		client, err := k8sclient.GetK8sClientByName(p.ClusterName)
		if err != nil {
			response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
			return
		}
		if err := fn(c.Request.Context(), client, p.Namespace, p.Name, p.Yaml); err != nil {
			response.FailWithError(c, ensureAppError(err))
			return
		}
		response.Success(c, successMsg, nil)
	}
}

// DeleteHandler 按 namespace + name 删除资源的 handler（DELETE 请求）。
func DeleteHandler(
	fn func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error,
	successMsg string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p NamespacedParams
		if err := c.ShouldBindJSON(&p); err != nil {
			response.FailWithError(c, apperr.Validation("参数校验失败", err))
			return
		}
		client, err := k8sclient.GetK8sClientByName(p.ClusterName)
		if err != nil {
			response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
			return
		}
		if err := fn(c.Request.Context(), client, p.Namespace, p.Name); err != nil {
			response.FailWithError(c, ensureAppError(err))
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

// ---------------------------------------------------------------------------
// 集群级资源 wrapper —— 无 namespace 参数
// ---------------------------------------------------------------------------

// ClusterListHandler 集群级资源的分页列表 handler。
// 绑定 ClusterListParams（GET+POST），构建 label selector。
func ClusterListHandler(
	fn func(ctx context.Context, client *kubernetes.Clientset, labelSelector string, limit int64, continueToken string) (any, error),
	successMsg string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p ClusterListParams
		if err := c.ShouldBind(&p); err != nil {
			response.FailWithError(c, apperr.Validation("参数校验失败", err))
			return
		}
		client, err := k8sclient.GetK8sClientByName(p.ClusterName)
		if err != nil {
			response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
			return
		}
		selector, err := buildLabelSelector(p.LabelFilters)
		if err != nil {
			response.FailWithError(c, apperr.Validation("标签选择器错误", err))
			return
		}
		data, err := fn(c.Request.Context(), client, selector, p.Limit, p.Continue)
		if err != nil {
			response.FailWithError(c, ensureAppError(err))
			return
		}
		response.Success(c, successMsg, data)
	}
}

// ClusterGetHandler 集群级资源的详情/YAML 获取 handler（GET 请求）。
// 绑定 ClusterScopedParams via ShouldBindQuery。
func ClusterGetHandler(
	fn func(ctx context.Context, client *kubernetes.Clientset, name string) (any, error),
	successMsg string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p ClusterScopedParams
		if err := c.ShouldBindQuery(&p); err != nil {
			response.FailWithError(c, apperr.Validation("参数校验失败", err))
			return
		}
		client, err := k8sclient.GetK8sClientByName(p.ClusterName)
		if err != nil {
			response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
			return
		}
		data, err := fn(c.Request.Context(), client, p.Name)
		if err != nil {
			response.FailWithError(c, ensureAppError(err))
			return
		}
		response.Success(c, successMsg, data)
	}
}

// ClusterCreateHandler 集群级资源的创建 handler（POST 请求）。
// 绑定 ClusterCreateParams via ShouldBindJSON。
func ClusterCreateHandler(
	fn func(ctx context.Context, client *kubernetes.Clientset, yaml string) error,
	successMsg string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p ClusterCreateParams
		if err := c.ShouldBindJSON(&p); err != nil {
			response.FailWithError(c, apperr.Validation("参数校验失败", err))
			return
		}
		client, err := k8sclient.GetK8sClientByName(p.ClusterName)
		if err != nil {
			response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
			return
		}
		if err := fn(c.Request.Context(), client, p.Yaml); err != nil {
			response.FailWithError(c, ensureAppError(err))
			return
		}
		response.Success(c, successMsg, nil)
	}
}

// ClusterDeleteHandler 集群级资源的删除 handler（DELETE 请求）。
// 绑定 ClusterScopedParams via ShouldBindJSON。
func ClusterDeleteHandler(
	fn func(ctx context.Context, client *kubernetes.Clientset, name string) error,
	successMsg string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p ClusterScopedParams
		if err := c.ShouldBindJSON(&p); err != nil {
			response.FailWithError(c, apperr.Validation("参数校验失败", err))
			return
		}
		client, err := k8sclient.GetK8sClientByName(p.ClusterName)
		if err != nil {
			response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
			return
		}
		if err := fn(c.Request.Context(), client, p.Name); err != nil {
			response.FailWithError(c, ensureAppError(err))
			return
		}
		response.Success(c, successMsg, nil)
	}
}
