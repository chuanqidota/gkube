package k8s

import (
	"net/http"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sStatefulSet "gkube/pkg/k8s/statefulset"
	"gkube/pkg/logger"
	"gkube/pkg/response"
	"k8s.io/client-go/kubernetes"
)

// ---------------------------------------------------------------------------
// 标准 handler（使用 wrapper）
// ---------------------------------------------------------------------------

var GetStatefulSetByName = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sStatefulSet.GetStatefulSetByName(client, namespace, name)
	},
	"执行成功", "获取statefulset失败",
)

var GetStatefulSetYaml = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sStatefulSet.GetStatefulSetYaml(client, namespace, name)
	},
	"执行成功", "获取statefulset yaml失败",
)

var CreateStatefulSet = CreateHandler(
	func(client *kubernetes.Clientset, namespace, yaml string) error {
		return k8sStatefulSet.CreateStatefulSet(client, namespace, yaml)
	},
	"执行成功", "创建statefulset失败",
)

var UpdateStatefulSet = UpdateHandler(
	func(client *kubernetes.Clientset, namespace, name, yaml string) error {
		return k8sStatefulSet.UpdateStatefulSet(client, namespace, name, yaml)
	},
	"执行成功", "更新statefulset失败",
)

var DeleteStatefulSetByName = DeleteHandler(
	func(client *kubernetes.Clientset, namespace, name string) error {
		return k8sStatefulSet.DeleteStatefulSetByName(client, namespace, name)
	},
	"执行成功", "删除statefulset失败",
)

var GetStatefulSetEvents = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		events, err := k8sStatefulSet.GetStatefulSetEvents(client, namespace, name)
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
	"执行成功", "获取statefulset事件失败",
)

var StatefulSetPodList = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sStatefulSet.StatefulSetPodList(client, namespace, name)
	},
	"执行成功", "获取statefulset pod列表失败",
)

var GetStatefulSetRollbacks = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sStatefulSet.GetStatefulSetRollbacks(client, namespace, name)
	},
	"执行成功", "获取回滚列表失败",
)

var GetStatefulSetPVCs = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sStatefulSet.GetStatefulSetPVs(client, namespace, name)
	},
	"执行成功", "获取PVC列表失败",
)

// ---------------------------------------------------------------------------
// 特殊 handler
// ---------------------------------------------------------------------------

// GetStatefulSetList 列表 —— 有 limit>0 分支 + 始终 augment
func GetStatefulSetList(c *gin.Context) {
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
		ssList, err := k8sStatefulSet.ListStatefulSets(client, p.Namespace, p.Limit, p.Continue, selector)
		if err != nil {
			logger.Error(err.Error())
			response.FailWithStatus(c, http.StatusBadGateway, "获取statefulset列表失败")
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
		statefulSets, err := k8sStatefulSet.GetStatefulSetList(client, p.Namespace, selector)
		if err != nil {
			logger.Error(err.Error())
			response.FailWithStatus(c, http.StatusBadGateway, "获取statefulset列表失败")
			return
		}
		response.Success(c, "执行成功", statefulSets)
	}
}

// ScaleStatefulSet 扩缩容 —— Replicas 为 int32，返回 (bool, error)
func ScaleStatefulSet(c *gin.Context) {
	var body struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Replicas    *int32 `json:"replicas" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	if body.Replicas == nil {
		response.Fail(c, "replicas参数不能为空")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	ok, err := k8sStatefulSet.ScaleStatefulSet(client, body.Namespace, body.Name, *body.Replicas)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "扩缩容statefulset失败")
		return
	}
	if !ok {
		response.FailWithStatus(c, http.StatusBadGateway, "扩缩容statefulset失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

// RestartStatefulSet 重启 —— 返回 (bool, error)
func RestartStatefulSet(c *gin.Context) {
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
	ok, err := k8sStatefulSet.RestartStatefulSet(client, body.Namespace, body.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "重启statefulset失败")
		return
	}
	if !ok {
		response.FailWithStatus(c, http.StatusBadGateway, "重启statefulset失败")
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
		response.Fail(c, "参数校验失败")
		return
	}
	if body.Name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	result, err := k8sStatefulSet.RollbackStatefulSet(client, body.Namespace, body.Name, body.Revision)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "回滚StatefulSet失败")
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
		response.Fail(c, "参数校验失败")
		return
	}
	if body.Name == "" || body.ContainerName == "" || body.Image == "" {
		response.Fail(c, "name, containerName, image参数不能为空")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	result, err := k8sStatefulSet.UpdateStatefulSetImage(client, body.Namespace, body.Name, body.ContainerName, body.Image)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "更新StatefulSet镜像失败")
		return
	}
	response.Success(c, "更新镜像成功", result)
}
