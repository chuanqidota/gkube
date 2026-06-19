package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gkube/app/k8s/params"
	"gkube/pkg/k8s"
	k8sNamespace "gkube/pkg/k8s/namespace"
	"gkube/pkg/response"
	"maps"
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
	// 提取命名空间名称列表，简化前端处理
	var nsNames []string
	for _, ns := range namespaces.Items {
		nsNames = append(nsNames, ns.Name)
	}
	response.Success(c, "执行成功", nsNames)
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

// GetNamespaceDetail
//
//	@Description: 获取命名空间详情
//	@receiver n
//	@param c
func (n *namespace) GetNamespaceDetail(c *gin.Context) {
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	if name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	ns, err := k8sNamespace.GetNamespaceDetail(client, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取命名空间详情失败:%s", err.Error()))
		return
	}

	status := "Unknown"
	if ns.Status.Phase == "Active" {
		status = "Active"
	} else if ns.Status.Phase == "Terminating" {
		status = "Terminating"
	}

	labels := make(map[string]string)
	maps.Copy(labels, ns.Labels)

	annotations := make(map[string]string)
	maps.Copy(annotations, ns.Annotations)

	result := map[string]any{
		"name":        ns.Name,
		"status":      status,
		"labels":      labels,
		"annotations": annotations,
		"age":         ns.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
	}
	response.Success(c, "执行成功", result)
}

// GetNamespaceYaml
//
//	@Description: 获取命名空间YAML
//	@receiver n
//	@param c
func (n *namespace) GetNamespaceYaml(c *gin.Context) {
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	if name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	yamlContent, err := k8sNamespace.GetNamespaceYaml(client, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取命名空间YAML失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yamlContent})
}

// UpdateNamespace
//
//	@Description: 更新命名空间
//	@receiver n
//	@param c
func (n *namespace) UpdateNamespace(c *gin.Context) {
	var req struct {
		ClusterName string `json:"clusterName"`
		YamlContent string `json:"yamlContent"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%s", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(req.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sNamespace.UpdateNamespace(client, req.YamlContent); err != nil {
		response.Fail(c, fmt.Sprintf("更新命名空间失败:%s", err.Error()))
		return
	}
	response.Success(c, "更新命名空间成功", nil)
}

// DeleteNamespace
//
//	@Description: 删除命名空间
//	@receiver n
//	@param c
func (n *namespace) DeleteNamespace(c *gin.Context) {
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	if name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sNamespace.DeleteNamespace(client, name); err != nil {
		response.Fail(c, fmt.Sprintf("删除命名空间失败:%s", err.Error()))
		return
	}
	response.Success(c, "删除命名空间成功", nil)
}
