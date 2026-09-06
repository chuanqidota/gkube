package k8s

import (
	"net/http"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sDaemonSet "gkube/pkg/k8s/daemonset"
	"gkube/pkg/logger"
	"gkube/pkg/response"
	"k8s.io/client-go/kubernetes"
)

// ---------------------------------------------------------------------------
// 标准 handler（使用 wrapper）
// ---------------------------------------------------------------------------

var GetDaemonSetByName = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sDaemonSet.GetDaemonSetByName(client, namespace, name)
	},
	"执行成功", "获取DaemonSet失败",
)

var GetDaemonSetYaml = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sDaemonSet.GetDaemonSetYaml(client, namespace, name)
	},
	"执行成功", "获取DaemonSet YAML失败",
)

var CreateDaemonSet = CreateHandler(
	func(client *kubernetes.Clientset, namespace, yaml string) error {
		return k8sDaemonSet.CreateDaemonSet(client, namespace, yaml)
	},
	"执行成功", "创建DaemonSet失败",
)

var UpdateDaemonSet = UpdateHandler(
	func(client *kubernetes.Clientset, namespace, name, yaml string) error {
		return k8sDaemonSet.UpdateDaemonSet(client, namespace, name, yaml)
	},
	"执行成功", "更新DaemonSet失败",
)

var DeleteDaemonSetByName = DeleteHandler(
	func(client *kubernetes.Clientset, namespace, name string) error {
		return k8sDaemonSet.DeleteDaemonSetByName(client, namespace, name)
	},
	"执行成功", "删除DaemonSet失败",
)

var GetDaemonSetEvents = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		events, err := k8sDaemonSet.GetDaemonSetEvents(client, namespace, name)
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
	"执行成功", "获取DaemonSet事件失败",
)

var DaemonSetPodList = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sDaemonSet.DaemonSetPodList(client, namespace, name)
	},
	"执行成功", "获取DaemonSet Pod列表失败",
)

var GetDaemonSetRollbacks = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sDaemonSet.GetDaemonSetRollbacks(client, namespace, name)
	},
	"执行成功", "获取DaemonSet回滚列表失败",
)

// ---------------------------------------------------------------------------
// 特殊 handler
// ---------------------------------------------------------------------------

// GetDaemonSetList 列表 —— 有 limit>0 分支
func GetDaemonSetList(c *gin.Context) {
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
		dsList, err := k8sDaemonSet.ListDaemonSets(client, p.Namespace, p.Limit, p.Continue, selector)
		if err != nil {
			logger.Error(err.Error())
			response.FailWithStatus(c, http.StatusBadGateway, "获取DaemonSet列表失败")
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
		daemonSets, err := k8sDaemonSet.GetDaemonSetList(client, p.Namespace, selector)
		if err != nil {
			logger.Error(err.Error())
			response.FailWithStatus(c, http.StatusBadGateway, "获取DaemonSet列表失败")
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
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	ok, err := k8sDaemonSet.RestartDaemonSet(client, body.Namespace, body.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "重启DaemonSet失败")
		return
	}
	if !ok {
		response.FailWithStatus(c, http.StatusBadGateway, "重启DaemonSet失败")
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
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	result, err := k8sDaemonSet.RollbackDaemonSet(client, body.Namespace, body.Name, body.Revision)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "回滚DaemonSet失败")
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
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	result, err := k8sDaemonSet.UpdateDaemonSetImage(client, body.Namespace, body.Name, body.ContainerName, body.Image)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "更新DaemonSet镜像失败")
		return
	}
	response.Success(c, "更新镜像成功", result)
}
