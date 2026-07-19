package rbac

import (
	"fmt"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
	"gkube/pkg/k8s"
	k8sServiceAccount "gkube/pkg/k8s/serviceaccount"
	"gkube/pkg/response"
)

type serviceaccount struct{}

var ServiceAccount = new(serviceaccount)

func (s *serviceaccount) GetServiceAccountList(c *gin.Context) {
	namespace := c.Query("namespace")
	clusterName := c.Query("clusterName")
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	saList, err := k8sServiceAccount.GetServiceAccountList(client, namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ServiceAccount列表失败:%s", err.Error()))
		return
	}
	var result []map[string]any
	for _, sa := range saList {
		result = append(result, map[string]any{
			"name":      sa.Name,
			"namespace": sa.Namespace,
			"labels":    sa.Labels,
			"age":       sa.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
		})
	}
	response.Success(c, "执行成功", result)
}

func (s *serviceaccount) GetServiceAccountYaml(c *gin.Context) {
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
	yamlContent, err := k8sServiceAccount.GetServiceAccountYaml(client, namespace, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ServiceAccount YAML失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yamlContent})
}

func (s *serviceaccount) DeleteServiceAccount(c *gin.Context) {
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
	if err := k8sServiceAccount.DeleteServiceAccount(client, namespace, name); err != nil {
		response.Fail(c, fmt.Sprintf("删除ServiceAccount失败:%s", err.Error()))
		return
	}
	response.Success(c, "删除ServiceAccount成功", nil)
}

func (s *serviceaccount) GetServiceAccountDetail(c *gin.Context) {
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
	sa, err := k8sServiceAccount.GetServiceAccountDetail(client, namespace, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ServiceAccount详情失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", sa)
}

func (s *serviceaccount) CreateServiceAccount(c *gin.Context) {
	var req struct {
		ClusterName string `json:"clusterName"`
		Namespace   string `json:"namespace"`
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
	var sa corev1.ServiceAccount
	if err := yaml.Unmarshal([]byte(req.Yaml), &sa); err != nil {
		response.Fail(c, fmt.Sprintf("YAML解析失败:%s", err.Error()))
		return
	}
	result, err := k8sServiceAccount.CreateServiceAccount(client, req.Namespace, &sa)
	if err != nil {
		response.Fail(c, fmt.Sprintf("创建ServiceAccount失败:%s", err.Error()))
		return
	}
	response.Success(c, "创建ServiceAccount成功", result)
}

func (s *serviceaccount) UpdateServiceAccount(c *gin.Context) {
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
	var sa corev1.ServiceAccount
	if err := yaml.Unmarshal([]byte(req.Yaml), &sa); err != nil {
		response.Fail(c, fmt.Sprintf("YAML解析失败:%s", err.Error()))
		return
	}
	result, err := k8sServiceAccount.UpdateServiceAccount(client, sa.Namespace, &sa)
	if err != nil {
		response.Fail(c, fmt.Sprintf("更新ServiceAccount失败:%s", err.Error()))
		return
	}
	response.Success(c, "更新ServiceAccount成功", result)
}
