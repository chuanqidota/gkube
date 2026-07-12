package rbac

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"gkube/pkg/k8s"
	k8sClusterRole "gkube/pkg/k8s/clusterrole"
	"gkube/pkg/response"
)

type clusterrole struct{}

var ClusterRole = new(clusterrole)

func (cr *clusterrole) GetClusterRoleList(c *gin.Context) {
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
	for _, item := range crList {
		isSystem := strings.HasPrefix(item.Name, "system:") ||
			item.Labels["kubernetes.io/bootstrapping"] != ""
		var rules []map[string]any
		for _, rule := range item.Rules {
			rules = append(rules, map[string]any{
				"apiGroups": rule.APIGroups,
				"resources": rule.Resources,
				"verbs":     rule.Verbs,
			})
		}
		result = append(result, map[string]any{
			"name":     item.Name,
			"labels":   item.Labels,
			"isSystem": isSystem,
			"rules":    rules,
			"age":      item.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
		})
	}
	response.Success(c, "执行成功", result)
}

func (cr *clusterrole) GetClusterRoleYaml(c *gin.Context) {
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
	yamlContent, err := k8sClusterRole.GetClusterRoleYaml(client, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ClusterRole YAML失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yamlContent})
}

func (cr *clusterrole) DeleteClusterRole(c *gin.Context) {
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
	if err := k8sClusterRole.DeleteClusterRole(client, name); err != nil {
		response.Fail(c, fmt.Sprintf("删除ClusterRole失败:%s", err.Error()))
		return
	}
	response.Success(c, "删除ClusterRole成功", nil)
}

func (cr *clusterrole) CreateClusterRole(c *gin.Context) {
	var req struct {
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
	if _, err := k8sClusterRole.CreateClusterRole(client, req.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("创建ClusterRole失败:%s", err.Error()))
		return
	}
	response.Success(c, "创建ClusterRole成功", nil)
}
