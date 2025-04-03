package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gkube/app/k8s/params"
	"gkube/pkg/k8s"
	k8sNamespace "gkube/pkg/k8s/namespace"
	"gkube/pkg/response"
)

type namespace struct {
}

var Namespace = new(namespace)

// GetNamespaceList
//
//	@Description: 获取集群命名空间列表
//	@receiver n
//	@param c
func (n *namespace) GetNamespaceList(c *gin.Context) {
	var query params.ClusterQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	namespaces, err := k8sNamespace.GetNamespaceList(client)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取集群命名空间列表失败:%s", err.Error()))
		return
	}
	result := map[string]any{
		"namespaces": namespaces,
	}
	response.Success(c, "执行成功", result)
}

// CreateNamespace
//
//	@Description: 创建namespace
//	@receiver n
//	@param c
func (n *namespace) CreateNamespace(c *gin.Context) {
	var body params.NamespaceCreateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}

	if err := k8sNamespace.CreateNamespace(client, body.Namespace); err != nil {
		response.Fail(c, fmt.Sprintf("创建命名空间失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}
