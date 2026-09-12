package k8s

import (
	"context"

	"github.com/gin-gonic/gin"
	apperr "gkube/pkg/errors"
	k8sclient "gkube/pkg/k8s"
	k8sDaemonSet "gkube/pkg/k8s/daemonset"
	"gkube/pkg/response"
	"k8s.io/client-go/kubernetes"
)

// ---------------------------------------------------------------------------
// 标准 handler（使用 wrapper）
// ---------------------------------------------------------------------------

var GetDaemonSetByName = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sDaemonSet.GetDaemonSetByName(ctx, client, namespace, name)
	},
	"执行成功",
)

var GetDaemonSetYaml = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sDaemonSet.GetDaemonSetYaml(ctx, client, namespace, name)
	},
	"执行成功",
)

var CreateDaemonSet = CreateHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, yaml string) error {
		return k8sDaemonSet.CreateDaemonSet(ctx, client, namespace, yaml)
	},
	"执行成功",
)

var UpdateDaemonSet = UpdateHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name, yaml string) error {
		return k8sDaemonSet.UpdateDaemonSet(ctx, client, namespace, name, yaml)
	},
	"执行成功",
)

var DeleteDaemonSetByName = DeleteHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error {
		return k8sDaemonSet.DeleteDaemonSetByName(ctx, client, namespace, name)
	},
	"执行成功",
)

var GetDaemonSetEvents = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		events, err := k8sDaemonSet.GetDaemonSetEvents(ctx, client, namespace, name)
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

var DaemonSetPodList = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sDaemonSet.DaemonSetPodList(ctx, client, namespace, name)
	},
	"执行成功",
)

var GetDaemonSetRollbacks = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sDaemonSet.GetDaemonSetRollbacks(ctx, client, namespace, name)
	},
	"执行成功",
)

// ---------------------------------------------------------------------------
// 特殊 handler
// ---------------------------------------------------------------------------

// GetDaemonSetList 列表 —— 有 limit>0 分支
func GetDaemonSetList(c *gin.Context) {
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
		dsList, err := k8sDaemonSet.ListDaemonSets(c.Request.Context(), client, p.Namespace, p.Limit, p.Continue, selector)
		if err != nil {
			response.FailWithError(c, ensureAppError(err))
			return
		}
		remaining := int64(0)
		if dsList.RemainingItemCount != nil {
			remaining = *dsList.RemainingItemCount
		}
		data := k8sclient.BuildPaginatedData(dsList.Items, dsList.Continue, remaining, p.Limit)
		data.Total = len(dsList.Items) + int(remaining)
		response.Success(c, "执行成功", data)
	} else {
		daemonSets, err := k8sDaemonSet.GetDaemonSetList(c.Request.Context(), client, p.Namespace, selector)
		if err != nil {
			response.FailWithError(c, ensureAppError(err))
			return
		}
		response.Success(c, "执行成功", daemonSets)
	}
}

// RestartDaemonSet 重启 —— 返回 (bool, error)
func RestartDaemonSet(c *gin.Context) {
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
	ok, err := k8sDaemonSet.RestartDaemonSet(c.Request.Context(), client, body.Namespace, body.Name)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	if !ok {
		response.FailWithError(c, apperr.K8sAPIFail("重启DaemonSet失败", err))
		return
	}
	response.Success(c, "执行成功", nil)
}

// RollbackDaemonSet 回滚 —— Revision 字段
func RollbackDaemonSet(c *gin.Context) {
	var body struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace"`
		Name        string `json:"name" binding:"required"`
		Revision    int64  `json:"revision" binding:"required"`
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
	result, err := k8sDaemonSet.RollbackDaemonSet(c.Request.Context(), client, body.Namespace, body.Name, body.Revision)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "回滚成功", result)
}

// UpdateDaemonSetImage 更新容器镜像 —— ContainerName + Image
func UpdateDaemonSetImage(c *gin.Context) {
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
	result, err := k8sDaemonSet.UpdateDaemonSetImage(c.Request.Context(), client, body.Namespace, body.Name, body.ContainerName, body.Image)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "更新镜像成功", result)
}
