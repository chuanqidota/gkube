package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gkube/pkg/k8s"
	k8sClusterRole "gkube/pkg/k8s/clusterrole"
	k8sRoleBinding "gkube/pkg/k8s/rolebinding"
	k8sSA "gkube/pkg/k8s/serviceaccount"
	"gkube/pkg/response"
)

type rbac struct{}

var Rbac = new(rbac)

// ServiceAccount
func (r *rbac) GetServiceAccountList(c *gin.Context) {
	namespace := c.Query("namespace")
	clusterName := c.Query("clusterName")
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	saList, err := k8sSA.GetServiceAccountList(client, namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ServiceAccount列表失败:%s", err.Error()))
		return
	}
	var result []map[string]any
	for _, sa := range saList {
		result = append(result, map[string]any{
			"name":      sa.Name,
			"namespace": sa.Namespace,
			"secrets":   len(sa.Secrets),
			"age":       sa.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
			"labels":    sa.Labels,
		})
	}
	response.Success(c, "执行成功", result)
}

func (r *rbac) GetServiceAccountYaml(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	yamlContent, err := k8sSA.GetServiceAccountYaml(client, namespace, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ServiceAccount YAML失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yamlContent})
}

func (r *rbac) DeleteServiceAccount(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sSA.DeleteServiceAccount(client, namespace, name); err != nil {
		response.Fail(c, fmt.Sprintf("删除ServiceAccount失败:%s", err.Error()))
		return
	}
	response.Success(c, "删除ServiceAccount成功", nil)
}

// ClusterRole
func (r *rbac) GetClusterRoleList(c *gin.Context) {
	clusterName := c.Query("clusterName")
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	crList, err := k8sClusterRole.GetClusterRoleList(client)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ClusterRole列表失败:%s", err.Error()))
		return
	}
	var result []map[string]any
	for _, cr := range crList {
		result = append(result, map[string]any{
			"name":           cr.Name,
			"rules":          len(cr.Rules),
			"age":            cr.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
			"labels":         cr.Labels,
			"is_system":      cr.Labels["kubernetes.io/bootstrapping"] == "rbac-defaults",
		})
	}
	response.Success(c, "执行成功", result)
}

func (r *rbac) GetClusterRoleYaml(c *gin.Context) {
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	yamlContent, err := k8sClusterRole.GetClusterRoleYaml(client, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ClusterRole YAML失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yamlContent})
}

func (r *rbac) DeleteClusterRole(c *gin.Context) {
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sClusterRole.DeleteClusterRole(client, name); err != nil {
		response.Fail(c, fmt.Sprintf("删除ClusterRole失败:%s", err.Error()))
		return
	}
	response.Success(c, "删除ClusterRole成功", nil)
}

// Role
func (r *rbac) GetRoleList(c *gin.Context) {
	namespace := c.Query("namespace")
	clusterName := c.Query("clusterName")
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	roleList, err := k8sClusterRole.GetRoleList(client, namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取Role列表失败:%s", err.Error()))
		return
	}
	var result []map[string]any
	for _, role := range roleList {
		result = append(result, map[string]any{
			"name":      role.Name,
			"namespace": role.Namespace,
			"rules":     len(role.Rules),
			"age":       role.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
			"labels":    role.Labels,
		})
	}
	response.Success(c, "执行成功", result)
}

func (r *rbac) GetRoleYaml(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	yamlContent, err := k8sClusterRole.GetRoleYaml(client, namespace, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取Role YAML失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yamlContent})
}

func (r *rbac) DeleteRole(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sClusterRole.DeleteRole(client, namespace, name); err != nil {
		response.Fail(c, fmt.Sprintf("删除Role失败:%s", err.Error()))
		return
	}
	response.Success(c, "删除Role成功", nil)
}

// ClusterRoleBinding
func (r *rbac) GetClusterRoleBindingList(c *gin.Context) {
	clusterName := c.Query("clusterName")
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	crbList, err := k8sRoleBinding.GetClusterRoleBindingList(client)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ClusterRoleBinding列表失败:%s", err.Error()))
		return
	}
	var result []map[string]any
	for _, crb := range crbList {
		var subjects []string
		for _, s := range crb.Subjects {
			subjects = append(subjects, fmt.Sprintf("%s/%s", s.Kind, s.Name))
		}
		result = append(result, map[string]any{
			"name":        crb.Name,
			"role":        crb.RoleRef.Name,
			"subjects":    subjects,
			"age":         crb.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
			"labels":      crb.Labels,
		})
	}
	response.Success(c, "执行成功", result)
}

func (r *rbac) GetClusterRoleBindingYaml(c *gin.Context) {
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	yamlContent, err := k8sRoleBinding.GetClusterRoleBindingYaml(client, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ClusterRoleBinding YAML失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yamlContent})
}

func (r *rbac) DeleteClusterRoleBinding(c *gin.Context) {
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sRoleBinding.DeleteClusterRoleBinding(client, name); err != nil {
		response.Fail(c, fmt.Sprintf("删除ClusterRoleBinding失败:%s", err.Error()))
		return
	}
	response.Success(c, "删除ClusterRoleBinding成功", nil)
}

// RoleBinding
func (r *rbac) GetRoleBindingList(c *gin.Context) {
	namespace := c.Query("namespace")
	clusterName := c.Query("clusterName")
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	rbList, err := k8sRoleBinding.GetRoleBindingList(client, namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取RoleBinding列表失败:%s", err.Error()))
		return
	}
	var result []map[string]any
	for _, rb := range rbList {
		var subjects []string
		for _, s := range rb.Subjects {
			subjects = append(subjects, fmt.Sprintf("%s/%s", s.Kind, s.Name))
		}
		result = append(result, map[string]any{
			"name":      rb.Name,
			"namespace": rb.Namespace,
			"role":      rb.RoleRef.Name,
			"subjects":  subjects,
			"age":       rb.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
			"labels":    rb.Labels,
		})
	}
	response.Success(c, "执行成功", result)
}

func (r *rbac) GetRoleBindingYaml(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	yamlContent, err := k8sRoleBinding.GetRoleBindingYaml(client, namespace, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取RoleBinding YAML失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yamlContent})
}

func (r *rbac) DeleteRoleBinding(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sRoleBinding.DeleteRoleBinding(client, namespace, name); err != nil {
		response.Fail(c, fmt.Sprintf("删除RoleBinding失败:%s", err.Error()))
		return
	}
	response.Success(c, "删除RoleBinding成功", nil)
}
