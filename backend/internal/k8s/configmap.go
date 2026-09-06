package k8s

import (
	"net/http"

	"k8s.io/client-go/kubernetes"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sConfigMap "gkube/pkg/k8s/configmap"
	"gkube/pkg/logger"
	"gkube/pkg/response"
)

// ---------------------------------------------------------------------------
// 标准 handler
// ---------------------------------------------------------------------------

var GetConfigMapList = ListHandler(
	func(client *kubernetes.Clientset, namespace, selector string, limit int64, continueToken string) (any, error) {
		list, err := k8sConfigMap.GetConfigMapList(client, namespace, limit, continueToken, selector)
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
	"获取ConfigMap列表成功", "获取ConfigMap列表失败",
)

var GetConfigMapByName = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sConfigMap.GetConfigMapByName(client, namespace, name)
	},
	"获取ConfigMap成功", "获取ConfigMap失败",
)

var GetConfigMapYaml = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		yaml, err := k8sConfigMap.GetConfigMapYaml(client, namespace, name)
		if err != nil {
			return nil, err
		}
		return map[string]string{"yaml": yaml}, nil
	},
	"获取ConfigMap YAML成功", "获取ConfigMap YAML失败",
)

var DeleteConfigMapByName = DeleteHandler(
	func(client *kubernetes.Clientset, namespace, name string) error {
		return k8sConfigMap.DeleteConfigMap(client, namespace, name)
	},
	"删除ConfigMap成功", "删除ConfigMap失败",
)

// CreateConfigMapFromYaml 创建 —— namespace required
func CreateConfigMapFromYaml(c *gin.Context) {
	var p struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace" binding:"required"`
		Yaml        string `json:"yaml" binding:"required"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	if err := k8sConfigMap.CreateConfigMapFromYaml(client, p.Namespace, p.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "创建ConfigMap失败")
		return
	}
	response.Success(c, "创建ConfigMap成功", nil)
}

// UpdateConfigMapFromYaml 更新 —— namespace required
func UpdateConfigMapFromYaml(c *gin.Context) {
	var p struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace" binding:"required"`
		Yaml        string `json:"yaml" binding:"required"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	if err := k8sConfigMap.UpdateConfigMapFromYaml(client, p.Namespace, p.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "更新ConfigMap失败")
		return
	}
	response.Success(c, "更新ConfigMap成功", nil)
}
