package config

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gkube/pkg/k8s"
	k8sRq "gkube/pkg/k8s/resourcequota"
	"gkube/pkg/response"
)

type resourceQuota struct{}

var ResourceQuota = new(resourceQuota)

func (rq *resourceQuota) GetResourceQuotaList(c *gin.Context) {
	namespace := c.Query("namespace")
	clusterName := c.Query("clusterName")
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	rqList, err := k8sRq.GetResourceQuotaList(client, namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ResourceQuota列表失败:%s", err.Error()))
		return
	}
	var result []map[string]any
	for _, rq := range rqList {
		var hard, used map[string]string
		hard = make(map[string]string)
		used = make(map[string]string)
		for k, v := range rq.Spec.Hard {
			hard[string(k)] = v.String()
		}
		for k, v := range rq.Status.Used {
			used[string(k)] = v.String()
		}
		result = append(result, map[string]any{
			"name":      rq.Name,
			"namespace": rq.Namespace,
			"hard":      hard,
			"used":      used,
			"age":       rq.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
			"labels":    rq.Labels,
		})
	}
	response.Success(c, "执行成功", result)
}

func (rq *resourceQuota) GetResourceQuotaDetail(c *gin.Context) {
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
	rqList, err := k8sRq.GetResourceQuotaList(client, namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ResourceQuota详情失败:%s", err.Error()))
		return
	}
	for _, r := range rqList {
		if r.Name == name {
			response.Success(c, "执行成功", r)
			return
		}
	}
	response.Fail(c, "ResourceQuota不存在")
}

func (rq *resourceQuota) GetResourceQuotaYaml(c *gin.Context) {
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
	yamlContent, err := k8sRq.GetResourceQuotaYaml(client, namespace, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ResourceQuota YAML失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yamlContent})
}

func (rq *resourceQuota) CreateResourceQuota(c *gin.Context) {
	var req struct {
		ClusterName string `json:"clusterName"`
		Namespace   string `json:"namespace"`
		Yaml string `json:"yaml"`
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
	if err := k8sRq.CreateResourceQuota(client, req.Namespace, req.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("创建ResourceQuota失败:%s", err.Error()))
		return
	}
	response.Success(c, "创建ResourceQuota成功", nil)
}

func (rq *resourceQuota) UpdateResourceQuota(c *gin.Context) {
	var req struct {
		ClusterName string `json:"clusterName"`
		Namespace   string `json:"namespace"`
		Yaml string `json:"yaml"`
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
	if err := k8sRq.UpdateResourceQuota(client, req.Namespace, req.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("更新ResourceQuota失败:%s", err.Error()))
		return
	}
	response.Success(c, "更新ResourceQuota成功", nil)
}

func (rq *resourceQuota) DeleteResourceQuota(c *gin.Context) {
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
	if err := k8sRq.DeleteResourceQuota(client, namespace, name); err != nil {
		response.Fail(c, fmt.Sprintf("删除ResourceQuota失败:%s", err.Error()))
		return
	}
	response.Success(c, "删除ResourceQuota成功", nil)
}
