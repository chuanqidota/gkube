package k8s

import (
	"context"

	"github.com/gin-gonic/gin"
	apperr "gkube/pkg/errors"
	k8sclient "gkube/pkg/k8s"
	k8sDeployment "gkube/pkg/k8s/deployment"
	"gkube/pkg/response"
	"k8s.io/client-go/kubernetes"
)

// ---------------------------------------------------------------------------
// 标准 handler（使用 wrapper）
// ---------------------------------------------------------------------------

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
	"获取deployment列表成功",
)

var GetDeploymentDetail = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sDeployment.GetDeploymentDetail(ctx, client, namespace, name)
	},
	"获取deployment详情成功",
)

var GetDeploymentYaml = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		yaml, err := k8sDeployment.GetDeploymentYaml(ctx, client, namespace, name)
		if err != nil {
			return nil, err
		}
		return map[string]string{"yaml": yaml}, nil
	},
	"获取deployment yaml成功",
)

var CreateDeployment = CreateHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, yaml string) error {
		return k8sDeployment.CreateDeployment(ctx, client, namespace, yaml)
	},
	"创建成功",
)

var UpdateDeployment = UpdateHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name, yaml string) error {
		return k8sDeployment.UpdateDeployment(ctx, client, namespace, name, yaml)
	},
	"更新成功",
)

var DeleteDeployment = DeleteHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error {
		return k8sDeployment.DeleteDeployment(ctx, client, namespace, name)
	},
	"删除成功",
)

var RestartDeployment = DeleteHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error {
		return k8sDeployment.RestartDeployment(ctx, client, namespace, name)
	},
	"重启成功",
)

var GetDeploymentEvents = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sDeployment.GetDeploymentEvents(ctx, client, namespace, name)
	},
	"获取deployment事件成功",
)

var DeploymentPodList = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sDeployment.GetDeploymentPods(ctx, client, namespace, name)
	},
	"获取deployment pod列表成功",
)

var GetDeploymentReplicaSets = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sDeployment.GetDeploymentReplicaSets(ctx, client, namespace, name)
	},
	"获取ReplicaSet列表成功",
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
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}
	if body.Replicas == nil {
		response.FailWithError(c, apperr.BadRequest("副本数不能为空", nil))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	if err := k8sDeployment.ScaleDeployment(c.Request.Context(), client, body.Namespace, body.Name, body.Replicas); err != nil {
		response.FailWithError(c, ensureAppError(err))
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
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	if err := k8sDeployment.RollbackDeployment(c.Request.Context(), client, body.Namespace, body.Name, body.Revision); err != nil {
		response.FailWithError(c, ensureAppError(err))
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
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	if err := k8sDeployment.UpdateDeploymentImage(c.Request.Context(), client, body.Namespace, body.Name, body.ContainerName, body.Image); err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "更新镜像成功", nil)
}
