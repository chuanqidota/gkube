package k8s

import (
	"fmt"
	"maps"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	k8sclient "gkube/pkg/k8s"
	k8sNamespace "gkube/pkg/k8s/namespace"
	"gkube/pkg/response"
)

type namespace struct {
}

var Namespace = new(namespace)

// namespaceStatus returns a user-friendly status string from a K8s NamespacePhase.
func namespaceStatus(phase corev1.NamespacePhase) string {
	switch phase {
	case corev1.NamespaceActive:
		return "Active"
	case corev1.NamespaceTerminating:
		return "Terminating"
	default:
		return "Unknown"
	}
}

// GetNamespaceList
//
//	@Description: 获取集群命名空间列表
//	@receiver n
//	@param c
func (n *namespace) GetNamespaceList(c *gin.Context) {
	var query ClusterQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	namespaces, err := k8sNamespace.GetNamespaceList(client)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取集群命名空间列表失败:%s", err.Error()))
		return
	}
	// 返回完整的命名空间对象列表
	var result []map[string]any
	for _, ns := range namespaces.Items {
		labels := make(map[string]string)
		maps.Copy(labels, ns.Labels)
		annotations := make(map[string]string)
		maps.Copy(annotations, ns.Annotations)
		result = append(result, map[string]any{
			"name":        ns.Name,
			"status":      namespaceStatus(ns.Status.Phase),
			"labels":      labels,
			"annotations": annotations,
			"age":         ns.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
		})
	}
	response.Success(c, "执行成功", result)
}

// CreateNamespace
//
//	@Description: 创建namespace
//	@receiver n
//	@param c
func (n *namespace) CreateNamespace(c *gin.Context) {
	var body NamespaceCreateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}

	if err := k8sNamespace.CreateNamespace(client, body.Namespace, body.Labels, body.Annotations); err != nil {
		response.Fail(c, fmt.Sprintf("创建命名空间失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdateNamespaceLabels
//
//	@Description: 更新命名空间标签
//	@receiver n
//	@param c
func (n *namespace) UpdateNamespaceLabels(c *gin.Context) {
	var body NamespaceLabelsParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}
	if body.Namespace == "" {
		response.Fail(c, "namespace参数不能为空")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sNamespace.UpdateNamespaceLabels(client, body.Namespace, body.Labels); err != nil {
		response.Fail(c, fmt.Sprintf("更新命名空间标签失败:%s", err.Error()))
		return
	}
	response.Success(c, "更新命名空间标签成功", nil)
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
	if clusterName == "" {
		response.Fail(c, "clusterName参数不能为空")
		return
	}
	client, err := k8sclient.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	ns, err := k8sNamespace.GetNamespaceDetail(client, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取命名空间详情失败:%s", err.Error()))
		return
	}
	labels := make(map[string]string)
	maps.Copy(labels, ns.Labels)
	annotations := make(map[string]string)
	maps.Copy(annotations, ns.Annotations)
	result := map[string]any{
		"name":        ns.Name,
		"status":      namespaceStatus(ns.Status.Phase),
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
	if clusterName == "" {
		response.Fail(c, "clusterName参数不能为空")
		return
	}
	client, err := k8sclient.GetK8sClientByName(clusterName)
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
		Yaml        string `json:"yaml"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%s", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(req.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sNamespace.UpdateNamespace(client, req.Yaml); err != nil {
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
	if clusterName == "" {
		response.Fail(c, "clusterName参数不能为空")
		return
	}
	client, err := k8sclient.GetK8sClientByName(clusterName)
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

type NamespaceCreateParams struct {
	ClusterName string            `json:"clusterName" label:"集群名称"`
	Namespace   string            `json:"namespace" label:"名称"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

type NamespaceLabelsParams struct {
	ClusterName string            `json:"clusterName" label:"集群名称"`
	Namespace   string            `json:"namespace" label:"名称"`
	Labels      map[string]string `json:"labels"`
}
