package k8s

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sSecret "gkube/pkg/k8s/secret"
	"gkube/pkg/logger"
	"gkube/pkg/response"
)

type secret struct {
}

var Secret = new(secret)

// GetSecretsList
//
//	@Description: 查询secret列表
//	@receiver s
//	@param c
func (s *secret) GetSecretsList(c *gin.Context) {
	var query SecretQueryListParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get k8s client: %v", err))
		response.Fail(c, "获取 Kubernetes 客户端失败")
		return
	}
	secrets, err := k8sSecret.GetSecretsList(client, query.Namespace)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get secret list: %v", err))
		response.Fail(c, "获取 Secret 列表失败")
		return
	}
	response.Success(c, "执行成功", secrets)
}

// GetSecretByName
//
//	@Description: 查询secret根据名称
//	@receiver s
//	@param c
func (s *secret) GetSecretByName(c *gin.Context) {
	var query SecretQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get k8s client: %v", err))
		response.Fail(c, "获取 Kubernetes 客户端失败")
		return
	}

	_secret, err := k8sSecret.GetSecretByName(client, query.Namespace, query.Name)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get secret %s/%s: %v", query.Namespace, query.Name, err))
		response.Fail(c, "获取 Secret 失败")
		return
	}
	response.Success(c, "执行成功", _secret)
}

// GetSecretYaml
//
//	@Description: 获取secret的yaml
//	@receiver s
//	@param c
func (s *secret) GetSecretYaml(c *gin.Context) {
	var query SecretQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get k8s client: %v", err))
		response.Fail(c, "获取 Kubernetes 客户端失败")
		return
	}
	secretYaml, err := k8sSecret.GetSecretYaml(client, query.Namespace, query.Name)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get secret yaml %s/%s: %v", query.Namespace, query.Name, err))
		response.Fail(c, "获取 Secret YAML 失败")
		return
	}
	response.Success(c, "执行成功", secretYaml)
}

// DeleteSecret
//
//	@Description: 删除secret
//	@receiver s
//	@param c
func (s *secret) DeleteSecret(c *gin.Context) {
	var body SecretDeleteParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get k8s client: %v", err))
		response.Fail(c, "获取 Kubernetes 客户端失败")
		return
	}
	if err := k8sSecret.DeleteSecret(client, body.Namespace, body.Name); err != nil {
		logger.Error(fmt.Sprintf("Failed to delete secret %s/%s: %v", body.Namespace, body.Name, err))
		response.Fail(c, "删除 Secret 失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdateSecretFromYaml
//
//	@Description: 通过YAML更新secret
//	@receiver s
//	@param c
func (s *secret) UpdateSecretFromYaml(c *gin.Context) {
	var req SecretYamlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	client, err := k8sclient.GetK8sClientByName(req.ClusterName)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get k8s client: %v", err))
		response.Fail(c, "获取 Kubernetes 客户端失败")
		return
	}
	if err := k8sSecret.UpdateSecretFromYaml(client, req.Namespace, req.Yaml); err != nil {
		logger.Error(fmt.Sprintf("Failed to update secret: %v", err))
		if isValidationError(err) {
			response.Fail(c, err.Error())
		} else {
			response.Fail(c, "更新 Secret 失败")
		}
		return
	}
	response.Success(c, "更新 Secret 成功", nil)
}

// CreateSecretFromYaml
//
//	@Description: 通过YAML创建secret
//	@receiver s
//	@param c
func (s *secret) CreateSecretFromYaml(c *gin.Context) {
	var req SecretYamlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	client, err := k8sclient.GetK8sClientByName(req.ClusterName)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get k8s client: %v", err))
		response.Fail(c, "获取 Kubernetes 客户端失败")
		return
	}
	if err := k8sSecret.CreateSecretFromYaml(client, req.Namespace, req.Yaml); err != nil {
		logger.Error(fmt.Sprintf("Failed to create secret: %v", err))
		if isValidationError(err) {
			response.Fail(c, err.Error())
		} else {
			response.Fail(c, "创建 Secret 失败")
		}
		return
	}
	response.Success(c, "创建 Secret 成功", nil)
}

// isValidationError 区分校验错误与 K8s API 内部错误，前者可安全展示给用户
func isValidationError(err error) bool {
	msg := err.Error()
	return strings.HasPrefix(msg, "YAML content cannot be empty") ||
		strings.HasPrefix(msg, "failed to unmarshal") ||
		strings.HasPrefix(msg, "Secret name is required")
}

type SecretQueryListParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type SecretQueryByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type SecretDeleteParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
}

type SecretYamlRequest struct {
	ClusterName string `json:"clusterName" binding:"required"`
	Namespace   string `json:"namespace"`
	Yaml        string `json:"yaml" binding:"required"`
}
