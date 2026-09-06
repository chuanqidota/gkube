package k8s

import (
	"net/http"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sLabels "gkube/pkg/k8s/labels"
	"gkube/pkg/logger"
	"gkube/pkg/response"
	"k8s.io/client-go/kubernetes"
)

// ---------------------------------------------------------------------------
// Handler wrapper 函数 — 消除 bind→client→call→respond 样板代码
//
// 每个 wrapper 统一处理：参数绑定、K8s 客户端获取、label selector 构建、
// 错误日志、标准化响应。业务逻辑通过 fn 闭包注入。
// ---------------------------------------------------------------------------

// ListHandler 分页列表 handler。
// 绑定 ListParams（GET+POST），构建 label selector，将 Limit/Continue 传给 fn。
// 适用：Deployment, Pod, ConfigMap, Secret, PVC, ReplicaSet 等始终分页的资源。
func ListHandler(
	fn func(client *kubernetes.Clientset, namespace, labelSelector string, limit int64, continueToken string) (any, error),
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
			logger.Error(err.Error())
			response.Fail(c, "获取k8s客户端失败")
			return
		}
		selector, err := buildLabelSelector(p.LabelFilters)
		if err != nil {
			response.Fail(c, err.Error())
			return
		}
		data, err := fn(client, p.Namespace, selector, p.Limit, p.Continue)
		if err != nil {
			logger.Error(err.Error())
			response.FailWithStatus(c, http.StatusBadGateway, failMsg)
			return
		}
		response.Success(c, successMsg, data)
	}
}

// NamespacedHandler 按 namespace + name 获取资源的 handler（GET 请求）。
// 绑定 NamespacedParams via ShouldBindQuery。
// 适用：detail, yaml, events, pods, rollbacks, pvcs 等所有按名称查询的接口。
func NamespacedHandler(
	fn func(client *kubernetes.Clientset, namespace, name string) (any, error),
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
			logger.Error(err.Error())
			response.Fail(c, "获取k8s客户端失败")
			return
		}
		data, err := fn(client, p.Namespace, p.Name)
		if err != nil {
			logger.Error(err.Error())
			response.FailWithStatus(c, http.StatusBadGateway, failMsg)
			return
		}
		response.Success(c, successMsg, data)
	}
}

// CreateHandler 从 YAML 创建资源的 handler（POST 请求）。
// 绑定 CreateParams via ShouldBindJSON。
func CreateHandler(
	fn func(client *kubernetes.Clientset, namespace, yaml string) error,
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
			logger.Error(err.Error())
			response.Fail(c, "获取k8s客户端失败")
			return
		}
		if err := fn(client, p.Namespace, p.Yaml); err != nil {
			logger.Error(err.Error())
			response.FailWithStatus(c, http.StatusBadGateway, failMsg)
			return
		}
		response.Success(c, successMsg, nil)
	}
}

// UpdateHandler 从 YAML 更新资源的 handler（PUT 请求）。
// 绑定 UpdateParams via ShouldBindJSON。
func UpdateHandler(
	fn func(client *kubernetes.Clientset, namespace, name, yaml string) error,
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
			logger.Error(err.Error())
			response.Fail(c, "获取k8s客户端失败")
			return
		}
		if err := fn(client, p.Namespace, p.Name, p.Yaml); err != nil {
			logger.Error(err.Error())
			response.FailWithStatus(c, http.StatusBadGateway, failMsg)
			return
		}
		response.Success(c, successMsg, nil)
	}
}

// DeleteHandler 按 namespace + name 删除资源的 handler（DELETE 请求）。
// 绑定 NamespacedParams via ShouldBindJSON。
func DeleteHandler(
	fn func(client *kubernetes.Clientset, namespace, name string) error,
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
			logger.Error(err.Error())
			response.Fail(c, "获取k8s客户端失败")
			return
		}
		if err := fn(client, p.Namespace, p.Name); err != nil {
			logger.Error(err.Error())
			response.FailWithStatus(c, http.StatusBadGateway, failMsg)
			return
		}
		response.Success(c, successMsg, nil)
	}
}

// ---------------------------------------------------------------------------
// 内部 helper
// ---------------------------------------------------------------------------

// buildLabelSelector 从 LabelFilters 构建 K8s label selector 字符串。
// filters 为空时返回 ""（不过滤）。
func buildLabelSelector(filters []k8sLabels.LabelFilter) (string, error) {
	if len(filters) == 0 {
		return "", nil
	}
	return k8sLabels.BuildLabelSelector(filters)
}
