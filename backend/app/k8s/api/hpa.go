package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gkube/pkg/k8s"
	k8sHpa "gkube/pkg/k8s/hpa"
	"gkube/pkg/response"
)

type hpa struct{}

var Hpa = new(hpa)

func (h *hpa) GetHPAList(c *gin.Context) {
	namespace := c.Query("namespace")
	clusterName := c.Query("clusterName")
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	hpaList, err := k8sHpa.GetHPAList(client, namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取HPA列表失败:%s", err.Error()))
		return
	}
	var result []map[string]any
	for _, hpa := range hpaList {
		var minReplicas int32
		if hpa.Spec.MinReplicas != nil {
			minReplicas = *hpa.Spec.MinReplicas
		}
		result = append(result, map[string]any{
			"name":             hpa.Name,
			"namespace":        hpa.Namespace,
			"min_replicas":     minReplicas,
			"max_replicas":     hpa.Spec.MaxReplicas,
			"current_replicas": hpa.Status.CurrentReplicas,
			"desired_replicas": hpa.Status.DesiredReplicas,
			"target":           hpa.Spec.ScaleTargetRef.Name,
			"target_kind":      hpa.Spec.ScaleTargetRef.Kind,
			"age":              hpa.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
			"labels":           hpa.Labels,
		})
	}
	response.Success(c, "执行成功", result)
}

func (h *hpa) GetHPADetail(c *gin.Context) {
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
	hpaList, err := k8sHpa.GetHPAList(client, namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取HPA详情失败:%s", err.Error()))
		return
	}
	for _, h := range hpaList {
		if h.Name == name {
			response.Success(c, "执行成功", h)
			return
		}
	}
	response.Fail(c, "HPA不存在")
}

func (h *hpa) GetHPAYaml(c *gin.Context) {
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
	yamlContent, err := k8sHpa.GetHPAYaml(client, namespace, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取HPA YAML失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yamlContent})
}

func (h *hpa) CreateHPA(c *gin.Context) {
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
	if err := k8sHpa.CreateHPA(client, req.Namespace, req.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("创建HPA失败:%s", err.Error()))
		return
	}
	response.Success(c, "创建HPA成功", nil)
}

func (h *hpa) UpdateHPA(c *gin.Context) {
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
	if err := k8sHpa.UpdateHPA(client, req.Namespace, req.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("更新HPA失败:%s", err.Error()))
		return
	}
	response.Success(c, "更新HPA成功", nil)
}

func (h *hpa) DeleteHPA(c *gin.Context) {
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
	if err := k8sHpa.DeleteHPA(client, namespace, name); err != nil {
		response.Fail(c, fmt.Sprintf("删除HPA失败:%s", err.Error()))
		return
	}
	response.Success(c, "删除HPA成功", nil)
}
