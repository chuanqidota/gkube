package k8s

import (
	"net/http"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sService "gkube/pkg/k8s/service"
	"gkube/pkg/logger"
	"gkube/pkg/response"
	"k8s.io/client-go/kubernetes"
)

// ---------------------------------------------------------------------------
// 标准 handler
// ---------------------------------------------------------------------------

var GetServicesByName = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sService.GetServicesByName(client, namespace, name)
	},
	"获取成功", "获取Service失败",
)

var GetServicesYaml = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sService.GetServicesYaml(client, namespace, name)
	},
	"获取成功", "获取svc的yaml失败",
)

var CreateService = CreateHandler(
	func(client *kubernetes.Clientset, namespace, yaml string) error {
		return k8sService.CreateService(client, namespace, yaml)
	},
	"执行成功", "创建Service失败",
)

// UpdateService 更新 —— pkg 函数只传 namespace+yaml（无 name）
func UpdateService(c *gin.Context) {
	var body struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace"`
		Yaml        string `json:"yaml" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	if err := k8sService.UpdateService(client, body.Namespace, body.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "更新Service失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

var DeleteService = DeleteHandler(
	func(client *kubernetes.Clientset, namespace, name string) error {
		return k8sService.DeleteService(client, namespace, name)
	},
	"执行成功", "删除Service失败",
)

var GetServiceEvents = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sService.GetServiceEvents(client, namespace, name)
	},
	"获取成功", "获取Service事件失败",
)

var ServicePodList = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sService.ServicePodList(client, namespace, name)
	},
	"执行成功", "获取service关联pod列表失败",
)

var GetServiceEndpoints = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sService.GetServiceEndpoints(client, namespace, name)
	},
	"获取成功", "获取Service endpoints失败",
)

// ---------------------------------------------------------------------------
// 特殊 handler
// ---------------------------------------------------------------------------

// GetServicesList 列表 —— 有 limit>0 分支
func GetServicesList(c *gin.Context) {
	var p ListParams
	if err := c.ShouldBind(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	selector, err := buildLabelSelector(p.LabelFilters)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, err.Error())
		return
	}
	if p.Limit > 0 {
		svcList, err := k8sService.ListServices(client, p.Namespace, p.Limit, p.Continue, selector)
		if err != nil {
			logger.Error(err.Error())
			response.FailWithStatus(c, http.StatusBadGateway, "获取Service列表失败")
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
		services, err := k8sService.GetServicesList(client, p.Namespace, selector)
		if err != nil {
			logger.Error(err.Error())
			response.FailWithStatus(c, http.StatusBadGateway, "获取Service列表失败")
			return
		}
		response.Success(c, "获取成功", services)
	}
}
