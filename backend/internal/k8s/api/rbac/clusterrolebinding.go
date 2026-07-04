package rbac

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"gkube/pkg/k8s"
	k8sClusterRoleBinding "gkube/pkg/k8s/clusterrolebinding"
	"gkube/pkg/response"
)

type clusterrolebinding struct{}

var ClusterRoleBinding = new(clusterrolebinding)

func (crb *clusterrolebinding) GetClusterRoleBindingList(c *gin.Context) {
	clusterName := c.Query("clusterName")
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	crbList, err := k8sClusterRoleBinding.GetClusterRoleBindingList(client)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ClusterRoleBinding列表失败:%s", err.Error()))
		return
	}
	var result []map[string]any
	for _, item := range crbList {
		roleRef := ""
		if item.RoleRef.Kind != "" {
			roleRef = fmt.Sprintf("%s/%s", item.RoleRef.Kind, item.RoleRef.Name)
		}
		var subjects []string
		for _, s := range item.Subjects {
			subjects = append(subjects, fmt.Sprintf("%s/%s", s.Kind, s.Name))
		}
		isSystem := strings.HasPrefix(item.Name, "system:")
		result = append(result, map[string]any{
			"name":      item.Name,
			"roleRef":   roleRef,
			"subjects":  strings.Join(subjects, ", "),
			"isSystem":  isSystem,
			"labels":    item.Labels,
			"age":       item.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
		})
	}
	response.Success(c, "执行成功", result)
}

func (crb *clusterrolebinding) GetClusterRoleBindingYaml(c *gin.Context) {
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
	yamlContent, err := k8sClusterRoleBinding.GetClusterRoleBindingYaml(client, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ClusterRoleBinding YAML失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yamlContent})
}

func (crb *clusterrolebinding) DeleteClusterRoleBinding(c *gin.Context) {
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
	if err := k8sClusterRoleBinding.DeleteClusterRoleBinding(client, name); err != nil {
		response.Fail(c, fmt.Sprintf("删除ClusterRoleBinding失败:%s", err.Error()))
		return
	}
	response.Success(c, "删除ClusterRoleBinding成功", nil)
}
