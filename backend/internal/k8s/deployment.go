package k8s

import (
	"net/http"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sDeployment "gkube/pkg/k8s/deployment"
	"gkube/pkg/logger"
	"gkube/pkg/response"
	"k8s.io/client-go/kubernetes"
)

// ---------------------------------------------------------------------------
// 标准 handler（使用 wrapper）
// ---------------------------------------------------------------------------

var GetDeploymentList = ListHandler(
	func(client *kubernetes.Clientset, namespace, selector string, limit int64, continueToken string) (any, error) {
		list, err := k8sDeployment.ListDeployments(client, namespace, limit, continueToken, selector)
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

var GetDeploymentDetail = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sDeployment.GetDeploymentDetail(client, namespace, name)
	},
	"获取deployment详情成功", "获取deployment详情失败",
)

var GetDeploymentYaml = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		yaml, err := k8sDeployment.GetDeploymentYaml(client, namespace, name)
		if err != nil {
			return nil, err
		}
		return map[string]string{"yaml": yaml}, nil
	},
	"获取deployment yaml成功", "获取deployment yaml失败",
)

var CreateDeployment = CreateHandler(
	func(client *kubernetes.Clientset, namespace, yaml string) error {
		return k8sDeployment.CreateDeployment(client, namespace, yaml)
	},
	"创建成功", "创建deployment失败",
)

var UpdateDeployment = UpdateHandler(
	func(client *kubernetes.Clientset, namespace, name, yaml string) error {
		return k8sDeployment.UpdateDeployment(client, namespace, name, yaml)
	},
	"更新成功", "更新deployment失败",
)

var DeleteDeployment = DeleteHandler(
	func(client *kubernetes.Clientset, namespace, name string) error {
		return k8sDeployment.DeleteDeployment(client, namespace, name)
	},
	"删除成功", "删除deployment失败",
)

var RestartDeployment = DeleteHandler(
	func(client *kubernetes.Clientset, namespace, name string) error {
		return k8sDeployment.RestartDeployment(client, namespace, name)
	},
	"重启成功", "重启deployment失败",
)

var GetDeploymentEvents = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sDeployment.GetDeploymentEvents(client, namespace, name)
	},
	"获取deployment事件成功", "获取deployment事件失败",
)

var DeploymentPodList = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sDeployment.GetDeploymentPods(client, namespace, name)
	},
	"获取deployment pod列表成功", "获取deployment pod列表失败",
)

var GetDeploymentReplicaSets = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sDeployment.GetDeploymentReplicaSets(client, namespace, name)
	},
	"获取ReplicaSet列表成功", "获取ReplicaSet列表失败",
)

// ---------------------------------------------------------------------------
// 特殊 handler（参数结构超出标准 wrapper 覆盖范围）
// ---------------------------------------------------------------------------

// ScaleDeployment 扩缩容 —— Replicas 为 *int32，需要 nil 检查防误缩到 0
func ScaleDeployment(c *gin.Context) {
	var body struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace"`
		Name        string `json:"name" binding:"required"`
		Replicas    *int32 `json:"replicas" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	if body.Replicas == nil {
		response.Fail(c, "副本数不能为空")
		return
	}
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
	response.Success(c, "扩缩容成功", nil)
}

// RollbackDeployment 回滚 —— 额外需要 Revision 字段
func RollbackDeployment(c *gin.Context) {
	var body struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace"`
		Name        string `json:"name" binding:"required"`
		Revision    int64  `json:"revision"`
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
	if err := k8sDeployment.RollbackDeployment(client, body.Namespace, body.Name, body.Revision); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "回滚deployment失败")
		return
	}
	response.Success(c, "回滚成功", nil)
}

// UpdateDeploymentImage 更新容器镜像 —— 额外需要 ContainerName + Image
func UpdateDeploymentImage(c *gin.Context) {
	var body struct {
		ClusterName   string `json:"clusterName" binding:"required"`
		Namespace     string `json:"namespace"`
		Name          string `json:"name" binding:"required"`
		ContainerName string `json:"containerName" binding:"required"`
		Image         string `json:"image" binding:"required"`
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
	if err := k8sDeployment.UpdateDeploymentImage(client, body.Namespace, body.Name, body.ContainerName, body.Image); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "更新deployment镜像失败")
		return
	}
	response.Success(c, "更新镜像成功", nil)
}
