package k8s

import (
	"fmt"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sVpa "gkube/pkg/k8s/vpa"
	"gkube/pkg/response"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type vpa struct{}

var Vpa = new(vpa)

func (v *vpa) GetVPAList(c *gin.Context) {
	namespace := c.Query("namespace")
	clusterName := c.Query("clusterName")
	client, err := k8sclient.GetDynamicClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	vpaList, err := k8sVpa.GetVPAList(client, namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取VPA列表失败:当前集群未安装 VPA CRD 或不支持 autoscaling.k8s.io/v1 VerticalPodAutoscaler:%s", err.Error()))
		return
	}
	var result []map[string]any
	for _, item := range vpaList {
		targetName, _, _ := unstructured.NestedString(item.Object, "spec", "targetRef", "name")
		targetKind, _, _ := unstructured.NestedString(item.Object, "spec", "targetRef", "kind")
		targetAPIVersion, _, _ := unstructured.NestedString(item.Object, "spec", "targetRef", "apiVersion")
		updateMode, _, _ := unstructured.NestedString(item.Object, "spec", "updatePolicy", "updateMode")
		if updateMode == "" {
			updateMode = "Auto"
		}
		conditions, _, _ := unstructured.NestedSlice(item.Object, "status", "conditions")
		recommendations, _, _ := unstructured.NestedSlice(item.Object, "status", "recommendation", "containerRecommendations")
		result = append(result, map[string]any{
			"name":                 item.GetName(),
			"namespace":            item.GetNamespace(),
			"target":               targetName,
			"target_kind":          targetKind,
			"target_api_version":   targetAPIVersion,
			"update_mode":          updateMode,
			"recommendation_count": len(recommendations),
			"conditions":           conditions,
			"age":                  item.GetCreationTimestamp().Time.Format("2006-01-02 15:04:05"),
			"labels":               item.GetLabels(),
			"spec":                 item.Object["spec"],
			"status":               item.Object["status"],
		})
	}
	response.Success(c, "执行成功", result)
}

func (v *vpa) GetVPADetail(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	if name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	client, err := k8sclient.GetDynamicClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	vpa, err := k8sVpa.GetVPADetail(client, namespace, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取VPA详情失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", vpa.Object)
}

func (v *vpa) GetVPAYaml(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	if name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	client, err := k8sclient.GetDynamicClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	yamlContent, err := k8sVpa.GetVPAYaml(client, namespace, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取VPA YAML失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yamlContent})
}

func (v *vpa) CreateVPA(c *gin.Context) {
	var req struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace" binding:"required"`
		Yaml        string `json:"yaml" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%s", err.Error()))
		return
	}
	client, err := k8sclient.GetDynamicClientByName(req.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sVpa.CreateVPA(client, req.Namespace, req.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("创建VPA失败:%s", err.Error()))
		return
	}
	response.Success(c, "创建VPA成功", nil)
}

func (v *vpa) UpdateVPA(c *gin.Context) {
	var req struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace" binding:"required"`
		Yaml        string `json:"yaml" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%s", err.Error()))
		return
	}
	client, err := k8sclient.GetDynamicClientByName(req.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sVpa.UpdateVPA(client, req.Namespace, req.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("更新VPA失败:%s", err.Error()))
		return
	}
	response.Success(c, "更新VPA成功", nil)
}

func (v *vpa) DeleteVPA(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	if name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	client, err := k8sclient.GetDynamicClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sVpa.DeleteVPA(client, namespace, name); err != nil {
		response.Fail(c, fmt.Sprintf("删除VPA失败:%s", err.Error()))
		return
	}
	response.Success(c, "删除VPA成功", nil)
}
