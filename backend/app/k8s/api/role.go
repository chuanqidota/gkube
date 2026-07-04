package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gkube/pkg/k8s"
	k8sRole "gkube/pkg/k8s/role"
	"gkube/pkg/response"
)

type role struct{}

var Role = new(role)

func (r *role) GetRoleList(c *gin.Context) {
	namespace := c.Query("namespace")
	clusterName := c.Query("clusterName")
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	roleList, err := k8sRole.GetRoleList(client, namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取Role列表失败:%s", err.Error()))
		return
	}
	var result []map[string]any
	for _, item := range roleList {
		result = append(result, map[string]any{
			"name":      item.Name,
			"namespace": item.Namespace,
			"labels":    item.Labels,
			"age":       item.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
		})
	}
	response.Success(c, "执行成功", result)
}

func (r *role) GetRoleYaml(c *gin.Context) {
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
	yamlContent, err := k8sRole.GetRoleYaml(client, namespace, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取Role YAML失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yamlContent})
}

func (r *role) DeleteRole(c *gin.Context) {
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
	if err := k8sRole.DeleteRole(client, namespace, name); err != nil {
		response.Fail(c, fmt.Sprintf("删除Role失败:%s", err.Error()))
		return
	}
	response.Success(c, "删除Role成功", nil)
}
