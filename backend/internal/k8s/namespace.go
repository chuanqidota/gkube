package k8s

import (
	"maps"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	apperr "gkube/pkg/errors"
	k8sclient "gkube/pkg/k8s"
	k8sLabels "gkube/pkg/k8s/labels"
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
	var query struct {
		ClusterName  string                  `form:"clusterName" json:"clusterName" binding:"required"`
		LabelFilters []k8sLabels.LabelFilter `json:"labelFilters" form:"labelFilters"`
	}
	if err := c.ShouldBind(&query); err != nil {
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}
	// 构建 label selector
	var labelSelector string
	if len(query.LabelFilters) > 0 {
		var buildErr error
		labelSelector, buildErr = k8sLabels.BuildLabelSelector(query.LabelFilters)
		if buildErr != nil {
			response.FailWithError(c, apperr.Validation("标签筛选器参数错误", buildErr))
			return
		}
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	namespaces, err := k8sNamespace.GetNamespaceList(c.Request.Context(), client, labelSelector)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
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
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}

	if err := k8sNamespace.CreateNamespace(c.Request.Context(), client, body.Namespace, body.Labels, body.Annotations); err != nil {
		response.FailWithError(c, ensureAppError(err))
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
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}
	if body.Namespace == "" {
		response.FailWithError(c, apperr.Validation("namespace参数不能为空", nil))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	if err := k8sNamespace.UpdateNamespaceLabels(c.Request.Context(), client, body.Namespace, body.Labels); err != nil {
		response.FailWithError(c, ensureAppError(err))
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
		response.FailWithError(c, apperr.Validation("name参数不能为空", nil))
		return
	}
	if clusterName == "" {
		response.FailWithError(c, apperr.Validation("clusterName参数不能为空", nil))
		return
	}
	client, err := k8sclient.GetK8sClientByName(clusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	ns, err := k8sNamespace.GetNamespaceDetail(c.Request.Context(), client, name)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
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
		response.FailWithError(c, apperr.Validation("name参数不能为空", nil))
		return
	}
	if clusterName == "" {
		response.FailWithError(c, apperr.Validation("clusterName参数不能为空", nil))
		return
	}
	client, err := k8sclient.GetK8sClientByName(clusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	yamlContent, err := k8sNamespace.GetNamespaceYaml(c.Request.Context(), client, name)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
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
		ClusterName string `json:"clusterName" binding:"required"`
		Yaml        string `json:"yaml" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithError(c, apperr.Validation("参数错误", err))
		return
	}
	client, err := k8sclient.GetK8sClientByName(req.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	if err := k8sNamespace.UpdateNamespace(c.Request.Context(), client, req.Yaml); err != nil {
		response.FailWithError(c, ensureAppError(err))
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
		response.FailWithError(c, apperr.Validation("name参数不能为空", nil))
		return
	}
	if clusterName == "" {
		response.FailWithError(c, apperr.Validation("clusterName参数不能为空", nil))
		return
	}
	client, err := k8sclient.GetK8sClientByName(clusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	if err := k8sNamespace.DeleteNamespace(c.Request.Context(), client, name); err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "删除命名空间成功", nil)
}

type NamespaceCreateParams struct {
	ClusterName string            `json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string            `json:"namespace" binding:"required" label:"名称"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

type NamespaceLabelsParams struct {
	ClusterName string            `json:"clusterName" label:"集群名称"`
	Namespace   string            `json:"namespace" label:"名称"`
	Labels      map[string]string `json:"labels"`
}
