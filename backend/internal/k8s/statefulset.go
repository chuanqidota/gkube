package k8s

import (
	"context"

	"github.com/gin-gonic/gin"
	apperr "gkube/pkg/errors"
	k8sclient "gkube/pkg/k8s"
	k8sStatefulSet "gkube/pkg/k8s/statefulset"
	"gkube/pkg/response"
	"k8s.io/client-go/kubernetes"
)

// ---------------------------------------------------------------------------
// 标准 handler（使用 wrapper）
// ---------------------------------------------------------------------------

var GetStatefulSetByName = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sStatefulSet.GetStatefulSetByName(ctx, client, namespace, name)
	},
	"执行成功",
)

var GetStatefulSetYaml = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sStatefulSet.GetStatefulSetYaml(ctx, client, namespace, name)
	},
	"执行成功",
)

var CreateStatefulSet = CreateHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, yaml string) error {
		return k8sStatefulSet.CreateStatefulSet(ctx, client, namespace, yaml)
	},
	"执行成功",
)

var UpdateStatefulSet = UpdateHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name, yaml string) error {
		return k8sStatefulSet.UpdateStatefulSet(ctx, client, namespace, name, yaml)
	},
	"执行成功",
)

var DeleteStatefulSetByName = DeleteHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error {
		return k8sStatefulSet.DeleteStatefulSetByName(ctx, client, namespace, name)
	},
	"执行成功",
)

var GetStatefulSetEvents = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		events, err := k8sStatefulSet.GetStatefulSetEvents(ctx, client, namespace, name)
		if err != nil {
			return nil, err
		}
		var result []map[string]any
		for _, event := range events {
			result = append(result, map[string]any{
				"type":      event.Type,
				"reason":    event.Reason,
				"message":   event.Message,
				"last_seen": event.LastTimestamp,
			})
		}
		return result, nil
	},
	"执行成功",
)

var StatefulSetPodList = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sStatefulSet.StatefulSetPodList(ctx, client, namespace, name)
	},
	"执行成功",
)

var GetStatefulSetRollbacks = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sStatefulSet.GetStatefulSetRollbacks(ctx, client, namespace, name)
	},
	"执行成功",
)

var GetStatefulSetPVCs = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sStatefulSet.GetStatefulSetPVs(ctx, client, namespace, name)
	},
	"执行成功",
)

// ---------------------------------------------------------------------------
// 特殊 handler
// ---------------------------------------------------------------------------

// GetStatefulSetList 列表 —— 有 limit>0 分支 + 始终 augment
func GetStatefulSetList(c *gin.Context) {
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
		ssList, err := k8sStatefulSet.ListStatefulSets(c.Request.Context(), client, p.Namespace, p.Limit, p.Continue, selector)
		if err != nil {
			response.FailWithError(c, ensureAppError(err))
			return
		}
		remaining := int64(0)
		if ssList.RemainingItemCount != nil {
			remaining = *ssList.RemainingItemCount
		}
		data := k8sclient.BuildPaginatedData(ssList.Items, ssList.Continue, remaining, p.Limit)
		data.Total = len(ssList.Items) + int(remaining)
		response.Success(c, "执行成功", data)
	} else {
		statefulSets, err := k8sStatefulSet.GetStatefulSetList(c.Request.Context(), client, p.Namespace, selector)
		if err != nil {
			response.FailWithError(c, ensureAppError(err))
			return
		}
		response.Success(c, "执行成功", statefulSets)
	}
}

// ScaleStatefulSet 扩缩容 —— Replicas 为 int32
func ScaleStatefulSet(c *gin.Context) {
	var body struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Replicas    *int32 `json:"replicas" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}
	if body.Replicas == nil {
		response.FailWithError(c, apperr.BadRequest("replicas参数不能为空", nil))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	ok, err := k8sStatefulSet.ScaleStatefulSet(c.Request.Context(), client, body.Namespace, body.Name, *body.Replicas)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	if !ok {
		response.FailWithError(c, apperr.K8sAPIFail("扩缩容statefulset失败", err))
		return
	}
	response.Success(c, "执行成功", nil)
}

// RestartStatefulSet 重启
func RestartStatefulSet(c *gin.Context) {
	var body struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace"`
		Name        string `json:"name" binding:"required"`
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
	ok, err := k8sStatefulSet.RestartStatefulSet(c.Request.Context(), client, body.Namespace, body.Name)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	if !ok {
		response.FailWithError(c, apperr.K8sAPIFail("重启statefulset失败", err))
		return
	}
	response.Success(c, "执行成功", nil)
}

// RollbackStatefulSet 回滚 —— Revision 字段
func RollbackStatefulSet(c *gin.Context) {
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
	if body.Name == "" {
		response.FailWithError(c, apperr.BadRequest("name参数不能为空", nil))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	result, err := k8sStatefulSet.RollbackStatefulSet(c.Request.Context(), client, body.Namespace, body.Name, body.Revision)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "回滚成功", result)
}

// UpdateStatefulSetImage 更新容器镜像 —— ContainerName + Image
func UpdateStatefulSetImage(c *gin.Context) {
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
	if body.Name == "" || body.ContainerName == "" || body.Image == "" {
		response.FailWithError(c, apperr.BadRequest("name, containerName, image参数不能为空", nil))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	result, err := k8sStatefulSet.UpdateStatefulSetImage(c.Request.Context(), client, body.Namespace, body.Name, body.ContainerName, body.Image)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "更新镜像成功", result)
}
