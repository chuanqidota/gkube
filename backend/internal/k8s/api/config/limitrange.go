package config

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gkube/pkg/k8s"
	k8sLr "gkube/pkg/k8s/limitrange"
	"gkube/pkg/response"
)

type limitRange struct{}

var LimitRange = new(limitRange)

func (lr *limitRange) GetLimitRangeList(c *gin.Context) {
	namespace := c.Query("namespace")
	clusterName := c.Query("clusterName")
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	lrList, err := k8sLr.GetLimitRangeList(client, namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取LimitRange列表失败:%s", err.Error()))
		return
	}
	var result []map[string]any
	for _, lr := range lrList {
		var limits []map[string]any
		for _, limit := range lr.Spec.Limits {
			limits = append(limits, map[string]any{
				"type":           string(limit.Type),
				"max":            limit.Max,
				"min":            limit.Min,
				"default":        limit.Default,
				"defaultRequest": limit.DefaultRequest,
				"maxLimitRequestRatio": limit.MaxLimitRequestRatio,
			})
		}
		result = append(result, map[string]any{
			"name":      lr.Name,
			"namespace": lr.Namespace,
			"limits":    limits,
			"age":       lr.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
			"labels":    lr.Labels,
		})
	}
	response.Success(c, "执行成功", result)
}

func (lr *limitRange) GetLimitRangeDetail(c *gin.Context) {
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
	lrDetail, err := k8sLr.GetLimitRangeDetail(client, namespace, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取LimitRange详情失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", lrDetail)
}

func (lr *limitRange) GetLimitRangeYaml(c *gin.Context) {
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
	yamlContent, err := k8sLr.GetLimitRangeYaml(client, namespace, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取LimitRange YAML失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yamlContent})
}

func (lr *limitRange) CreateLimitRange(c *gin.Context) {
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
	if err := k8sLr.CreateLimitRange(client, req.Namespace, req.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("创建LimitRange失败:%s", err.Error()))
		return
	}
	response.Success(c, "创建LimitRange成功", nil)
}

func (lr *limitRange) UpdateLimitRange(c *gin.Context) {
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
	if err := k8sLr.UpdateLimitRange(client, req.Namespace, req.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("更新LimitRange失败:%s", err.Error()))
		return
	}
	response.Success(c, "更新LimitRange成功", nil)
}

func (lr *limitRange) DeleteLimitRange(c *gin.Context) {
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
	if err := k8sLr.DeleteLimitRange(client, namespace, name); err != nil {
		response.Fail(c, fmt.Sprintf("删除LimitRange失败:%s", err.Error()))
		return
	}
	response.Success(c, "删除LimitRange成功", nil)
}
