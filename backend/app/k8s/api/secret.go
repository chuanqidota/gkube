package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gkube/app/k8s/params"
	"gkube/pkg/k8s"
	k8sSecret "gkube/pkg/k8s/secret"
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
	var query params.SecretQueryListParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}

	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	secrets, err := k8sSecret.GetSecretsList(client, query.Namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取secret列表失败:%s", err.Error()))
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
	var query params.SecretQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}

	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}

	_secret, err := k8sSecret.GetSecretByName(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取secret失败:%s", err.Error()))
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
	var query params.SecretQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}

	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
	}
	secretYaml, err := k8sSecret.GetSecretYaml(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取secret失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", secretYaml)
}

// CreateSecret
//
//	@Description: 创建secret
//	@receiver s
//	@param c
func (s *secret) CreateSecret(c *gin.Context) {
	var body params.SecretCreateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sSecret.CreateSecret(client, body.Namespace, body.Name, body.Data); err != nil {
		response.Fail(c, fmt.Sprintf("创建secret失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdateSecret
//
//	@Description: 更新secret
//	@receiver s
//	@param c
func (s *secret) UpdateSecret(c *gin.Context) {
	var body params.SecretUpdateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sSecret.UpdateSecret(client, body.Namespace, body.Name, body.Data); err != nil {
		response.Fail(c, fmt.Sprintf("更新secret失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteSecret
//
//	@Description: 删除secret
//	@receiver s
//	@param c
func (s *secret) DeleteSecret(c *gin.Context) {
	var body params.SecretDeleteParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sSecret.DeleteSecret(client, body.Namespace, body.Name); err != nil {
		response.Fail(c, fmt.Sprintf("删除secret失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}
