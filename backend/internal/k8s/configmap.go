package k8s

import (
	"context"

	"github.com/gin-gonic/gin"
	apperr "gkube/pkg/errors"
	k8sclient "gkube/pkg/k8s"
	k8sConfigMap "gkube/pkg/k8s/configmap"
	"gkube/pkg/response"
	"k8s.io/client-go/kubernetes"
)

// ---------------------------------------------------------------------------
// 标准 handler（使用 wrapper）
// ---------------------------------------------------------------------------

var GetConfigMapList = ListHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, selector string, limit int64, continueToken string) (any, error) {
		list, err := k8sConfigMap.GetConfigMapList(ctx, client, namespace, limit, continueToken, selector)
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
	"获取ConfigMap列表成功",
)

var GetConfigMapByName = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sConfigMap.GetConfigMapByName(ctx, client, namespace, name)
	},
	"获取ConfigMap成功",
)

var GetConfigMapYaml = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		yaml, err := k8sConfigMap.GetConfigMapYaml(ctx, client, namespace, name)
		if err != nil {
			return nil, err
		}
		return map[string]string{"yaml": yaml}, nil
	},
	"获取ConfigMap YAML成功",
)

var DeleteConfigMapByName = DeleteHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error {
		return k8sConfigMap.DeleteConfigMap(ctx, client, namespace, name)
	},
	"删除ConfigMap成功",
)

// ---------------------------------------------------------------------------
// 特殊 handler（namespace required）
// ---------------------------------------------------------------------------

// CreateConfigMapFromYaml 创建 —— namespace required
func CreateConfigMapFromYaml(c *gin.Context) {
	var p struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace" binding:"required"`
		Yaml        string `json:"yaml" binding:"required"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	if err := k8sConfigMap.CreateConfigMapFromYaml(c.Request.Context(), client, p.Namespace, p.Yaml); err != nil {
		response.FailWithError(c, ensureAppError(err))
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
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	if err := k8sConfigMap.UpdateConfigMapFromYaml(c.Request.Context(), client, p.Namespace, p.Yaml); err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "更新ConfigMap成功", nil)
}
