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
		isSystem := strings.HasPrefix(item.Name, "system:")
		result = append(result, map[string]any{
			"name":          item.Name,
			"roleRef":       roleRef,
			"roleRefName":   item.RoleRef.Name,
			"roleRefKind":   string(item.RoleRef.Kind),
			"subjects":      strings.Join(subjectStrs, ", "),
			"subjectList":   subjectObjs,
			"isSystem":      isSystem,
			"labels":        item.Labels,
			"age":           item.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
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

func (crb *clusterrolebinding) CreateClusterRoleBinding(c *gin.Context) {
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
	if _, err := k8sClusterRoleBinding.CreateClusterRoleBinding(client, req.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("创建ClusterRoleBinding失败:%s", err.Error()))
		return
	}
	response.Success(c, "创建ClusterRoleBinding成功", nil)
}
