package k8s

import (
	"net/http"

	"k8s.io/client-go/kubernetes"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sPod "gkube/pkg/k8s/pod"
	"gkube/pkg/logger"
	"gkube/pkg/response"
	"k8s.io/apimachinery/pkg/fields"
	k8sEvent "gkube/pkg/k8s/event"
)

// ---------------------------------------------------------------------------
// 标准 handler
// ---------------------------------------------------------------------------

var GetPodList = ListHandler(
	func(client *kubernetes.Clientset, namespace, selector string, limit int64, continueToken string) (any, error) {
		list, err := k8sPod.ListPods(client, namespace, limit, continueToken, selector)
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
	"获取pod列表成功", "获取pod列表失败",
)

var GetPodByName = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sPod.GetPodByName(client, namespace, name)
	},
	"获取pod成功", "获取pod失败",
)

var GetPodYaml = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		yaml, err := k8sPod.GetPodYaml(client, namespace, name)
		if err != nil {
			return nil, err
		}
		return map[string]string{"yaml": yaml}, nil
	},
	"获取pod成功", "获取pod失败",
)

var CreatePod = CreateHandler(
	func(client *kubernetes.Clientset, namespace, yaml string) error {
		return k8sPod.CreatePod(client, namespace, yaml)
	},
	"执行成功", "创建pod失败",
)

var PatchPodMetadata = UpdateHandler(
	func(client *kubernetes.Clientset, namespace, name, yaml string) error {
		return k8sPod.PatchPodMetadata(client, namespace, name, yaml)
	},
	"执行成功", "更新pod元数据失败",
)

// ---------------------------------------------------------------------------
// 特殊 handler
// ---------------------------------------------------------------------------

// DeletePodByName 删除 —— 有 Force 参数
func DeletePodByName(c *gin.Context) {
	var body struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Namespace   string `json:"namespace"`
		Force       bool   `json:"force"`
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
	if err := k8sPod.DeletePodByName(client, body.Namespace, body.Name, body.Force); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "删除pod失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

// ListPodEvents 获取 Pod 事件 —— 用 fieldSelector 过滤
func ListPodEvents(c *gin.Context) {
	var p NamespacedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	selector := fields.OneTermEqualSelector("involvedObject.name", p.Name).String()
	events, _, _, err := k8sEvent.ListEvents(client, p.Namespace, selector, 0, "")
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取pod事件失败")
		return
	}
	response.Success(c, "获取pod事件成功", events)
}
