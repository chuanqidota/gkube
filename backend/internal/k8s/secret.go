package k8s

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sSecret "gkube/pkg/k8s/secret"
	"gkube/pkg/logger"
	"gkube/pkg/response"
	"k8s.io/client-go/kubernetes"
)

// ---------------------------------------------------------------------------
// 标准 handler
// ---------------------------------------------------------------------------

var GetSecretByName = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sSecret.GetSecretByName(client, namespace, name)
	},
	"获取Secret成功", "获取Secret失败",
)

var GetSecretYaml = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sSecret.GetSecretYaml(client, namespace, name)
	},
	"获取Secret YAML成功", "获取Secret YAML失败",
)

var DeleteSecret = DeleteHandler(
	func(client *kubernetes.Clientset, namespace, name string) error {
		return k8sSecret.DeleteSecret(client, namespace, name)
	},
	"删除Secret成功", "删除Secret失败",
)

// ---------------------------------------------------------------------------
// 特殊 handler
// ---------------------------------------------------------------------------

// GetSecretsList 列表 —— 非分页（pkg 函数不接受 limit/continue）
func GetSecretsList(c *gin.Context) {
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
		response.Fail(c, err.Error())
		return
	}
	secrets, err := k8sSecret.GetSecretsList(client, p.Namespace, selector)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取Secret列表失败")
		return
	}
	response.Success(c, "获取Secret列表成功", secrets)
}

// CreateSecretFromYaml 创建 —— 带验证错误区分
func CreateSecretFromYaml(c *gin.Context) {
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
	if err := k8sSecret.CreateSecretFromYaml(client, p.Namespace, p.Yaml); err != nil {
		logger.Error(err.Error())
		if isValidationError(err) {
			response.Fail(c, err.Error())
		} else {
			response.FailWithStatus(c, http.StatusBadGateway, "创建Secret失败")
		}
		return
	}
	response.Success(c, "创建Secret成功", nil)
}

// UpdateSecretFromYaml 更新 —— 带验证错误区分
func UpdateSecretFromYaml(c *gin.Context) {
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
	if err := k8sSecret.UpdateSecretFromYaml(client, p.Namespace, p.Yaml); err != nil {
		logger.Error(err.Error())
		if isValidationError(err) {
			response.Fail(c, err.Error())
		} else {
			response.FailWithStatus(c, http.StatusBadGateway, "更新Secret失败")
		}
		return
	}
	response.Success(c, "更新Secret成功", nil)
}

// isValidationError 检查是否为输入验证错误（可直接展示给用户）
func isValidationError(err error) bool {
	msg := err.Error()
	return strings.HasPrefix(msg, "YAML content cannot be empty") ||
		strings.HasPrefix(msg, "failed to unmarshal") ||
		strings.HasPrefix(msg, "Secret name is required")
}
