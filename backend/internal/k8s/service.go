package k8s

import (
	"context"

	"github.com/gin-gonic/gin"
	apperr "gkube/pkg/errors"
	k8sclient "gkube/pkg/k8s"
	k8sService "gkube/pkg/k8s/service"
	"gkube/pkg/response"
	"k8s.io/client-go/kubernetes"
)

// ---------------------------------------------------------------------------
// 标准 handler（使用 wrapper）
// ---------------------------------------------------------------------------

var GetServicesByName = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sService.GetServicesByName(ctx, client, namespace, name)
	},
	"获取成功",
)

var GetServicesYaml = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sService.GetServicesYaml(ctx, client, namespace, name)
	},
	"获取成功",
)

var CreateService = CreateHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, yaml string) error {
		return k8sService.CreateService(ctx, client, namespace, yaml)
	},
	"执行成功",
)

var DeleteService = DeleteHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error {
		return k8sService.DeleteService(ctx, client, namespace, name)
	},
	"执行成功",
)

var GetServiceEvents = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sService.GetServiceEvents(ctx, client, namespace, name)
	},
	"获取成功",
)

var ServicePodList = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sService.ServicePodList(ctx, client, namespace, name)
	},
	"执行成功",
)

var GetServiceEndpoints = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sService.GetServiceEndpoints(ctx, client, namespace, name)
	},
	"获取成功",
)

// ---------------------------------------------------------------------------
// 特殊 handler
// ---------------------------------------------------------------------------

// UpdateService 更新 —— pkg 函数只传 namespace+yaml（无 name）
func UpdateService(c *gin.Context) {
	var body struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace"`
		Yaml        string `json:"yaml" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	if err := k8sService.UpdateService(c.Request.Context(), client, body.Namespace, body.Yaml); err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", nil)
}

// GetServicesList 列表 —— 有 limit>0 分支
func GetServicesList(c *gin.Context) {
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
		response.FailWithError(c, ensureAppError(err))
		return
	}
	if p.Limit > 0 {
		svcList, err := k8sService.ListServices(c.Request.Context(), client, p.Namespace, p.Limit, p.Continue, selector)
		if err != nil {
			response.FailWithError(c, ensureAppError(err))
			return
		}
		remaining := int64(0)
		if svcList.RemainingItemCount != nil {
			remaining = *svcList.RemainingItemCount
		}
		data := k8sclient.BuildPaginatedData(svcList.Items, svcList.Continue, remaining, p.Limit)
		data.Total = len(svcList.Items) + int(remaining)
		response.Success(c, "获取成功", data)
	} else {
		services, err := k8sService.GetServicesList(c.Request.Context(), client, p.Namespace, selector)
		if err != nil {
			response.FailWithError(c, ensureAppError(err))
			return
		}
		response.Success(c, "获取成功", services)
	}
}
