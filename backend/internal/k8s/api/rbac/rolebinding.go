package rbac

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/yaml"
	"gkube/pkg/k8s"
	k8sRoleBinding "gkube/pkg/k8s/rolebinding"
	"gkube/pkg/response"
)

type rolebinding struct{}

var RoleBinding = new(rolebinding)

func (rb *rolebinding) GetRoleBindingList(c *gin.Context) {
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
	for _, item := range rbList {
		roleRef := ""
		if item.RoleRef.Kind != "" {
			roleRef = fmt.Sprintf("%s/%s", item.RoleRef.Kind, item.RoleRef.Name)
		}
		var subjectStrs []string
		var subjectObjs []map[string]string
		for _, s := range item.Subjects {
			subjectStrs = append(subjectStrs, fmt.Sprintf("%s/%s", s.Kind, s.Name))
			subjectObjs = append(subjectObjs, map[string]string{
				"kind":      string(s.Kind),
				"name":      s.Name,
				"namespace": s.Namespace,
			})
		}
		result = append(result, map[string]any{
			"name":          item.Name,
			"namespace":     item.Namespace,
			"roleRef":       roleRef,
			"roleRefName":   item.RoleRef.Name,
			"roleRefKind":   string(item.RoleRef.Kind),
			"subjects":      strings.Join(subjectStrs, ", "),
			"subjectList":   subjectObjs,
			"labels":        item.Labels,
			"age":           item.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
		})
	}
	response.Success(c, "执行成功", result)
}

func (rb *rolebinding) GetRoleBindingYaml(c *gin.Context) {
	namespace := c.Query("namespace")
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
	yamlContent, err := k8sRoleBinding.GetRoleBindingYaml(client, namespace, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取RoleBinding YAML失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yamlContent})
}

func (rb *rolebinding) DeleteRoleBinding(c *gin.Context) {
	namespace := c.Query("namespace")
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
	if err := k8sRoleBinding.DeleteRoleBinding(client, namespace, name); err != nil {
		response.Fail(c, fmt.Sprintf("删除RoleBinding失败:%s", err.Error()))
		return
	}
	response.Success(c, "删除RoleBinding成功", nil)
}

func (rb *rolebinding) CreateRoleBinding(c *gin.Context) {
	var req struct {
		Namespace   string `json:"namespace"`
		ClusterName string `json:"clusterName"`
		Yaml        string `json:"yaml"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误:"+err.Error())
		return
	}
	if req.Yaml == "" {
		response.Fail(c, "yaml参数不能为空")
		return
	}
	client, err := k8s.GetK8sClientByName(req.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if _, err := k8sRoleBinding.CreateRoleBinding(client, req.Namespace, req.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("创建RoleBinding失败:%s", err.Error()))
		return
	}
	response.Success(c, "创建RoleBinding成功", nil)
}

func (rb *rolebinding) GetRoleBindingDetail(c *gin.Context) {
	namespace := c.Query("namespace")
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
	result, err := k8sRoleBinding.GetRoleBindingDetail(client, namespace, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取RoleBinding详情失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", result)
}

func (rb *rolebinding) UpdateRoleBinding(c *gin.Context) {
	var req struct {
		ClusterName string `json:"clusterName"`
		Yaml        string `json:"yaml"`
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
	var binding rbacv1.RoleBinding
	if err := yaml.Unmarshal([]byte(req.Yaml), &binding); err != nil {
		response.Fail(c, fmt.Sprintf("YAML解析失败:%s", err.Error()))
		return
	}
	result, err := k8sRoleBinding.UpdateRoleBinding(client, binding.Namespace, &binding)
	if err != nil {
		response.Fail(c, fmt.Sprintf("更新RoleBinding失败:%s", err.Error()))
		return
	}
	response.Success(c, "更新RoleBinding成功", result)
}
