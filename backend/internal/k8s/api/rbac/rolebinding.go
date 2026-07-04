package rbac

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
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
		var subjects []string
		for _, s := range item.Subjects {
			subjects = append(subjects, fmt.Sprintf("%s/%s", s.Kind, s.Name))
		}
		result = append(result, map[string]any{
			"name":      item.Name,
			"namespace": item.Namespace,
			"roleRef":   roleRef,
			"subjects":  strings.Join(subjects, ", "),
			"labels":    item.Labels,
			"age":       item.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
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
