package k8s

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sIngress "gkube/pkg/k8s/ingress"
	"gkube/pkg/response"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ingress struct {
}

var Ingress = new(ingress)

// GetIngressList
//
//	@Description: 获取ingress
//	@receiver i
//	@param c
func (i *ingress) GetIngressList(c *gin.Context) {
	var query IngressQueryListParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, err.Error())
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	ingressList, err := k8sIngress.GetIngressList(client, query.Namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("查询ingress失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", ingressList)
}

// GetIngressByName
//
//	@Description: 获取ingress根据名称
//	@receiver i
//	@param c
func (i *ingress) GetIngressByName(c *gin.Context) {
	var query IngressQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, err.Error())
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	ingress, err := k8sIngress.GetIngressByName(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("查询ingress失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", ingress)
}

// GetIngressYaml
//
//	@Description: 获取ingress的yaml
//	@receiver i
//	@param c
func (i *ingress) GetIngressYaml(c *gin.Context) {
	var query IngressQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	ingressYaml, err := k8sIngress.GetIngressYaml(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ingress失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", ingressYaml)
}

// CreateIngress
//
//	@Description: 创建ingress
//	@receiver i
//	@param c
func (i *ingress) CreateIngress(c *gin.Context) {
	var body IngressCreateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sIngress.CreateIngress(client, body.Namespace, body.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("创建ingress失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdateIngress
//
//	@Description: 更新ingress
//	@receiver i
//	@param c
func (i *ingress) UpdateIngress(c *gin.Context) {
	var body IngressUpdateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sIngress.UpdateIngress(client, body.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("更新ingress失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteIngressByName
//
//	@Description: 删除ingress根据名称
//	@receiver i
//	@param c
func (i *ingress) DeleteIngressByName(c *gin.Context) {
	var body IngressDeleteByNameParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)

	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	if err := k8sIngress.DeleteIngressByName(client, body.Namespace, body.Name); err != nil {
		response.Fail(c, fmt.Sprintf("删除ingress失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// GetIngressEvents
//
//	@Description: 获取ingress事件
//	@receiver i
//	@param c
func (i *ingress) GetIngressEvents(c *gin.Context) {
	var query IngressQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	events, err := client.CoreV1().Events(query.Namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Ingress", query.Name),
	})
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ingress事件失败:%s", err.Error()))
		return
	}
	var result []map[string]any
	for _, event := range events.Items {
		lastSeen := ""
		if !event.LastTimestamp.IsZero() {
			lastSeen = event.LastTimestamp.Time.Format("2006-01-02 15:04:05")
		}
		result = append(result, map[string]any{
			"type":      event.Type,
			"reason":    event.Reason,
			"message":   event.Message,
			"last_seen": lastSeen,
		})
	}
	response.Success(c, "执行成功", result)
}

type IngressQueryListParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type IngressQueryByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type IngressCreateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Yaml        string `form:"yaml" json:"yaml" label:"Yaml"`
}

type IngressUpdateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Yaml        string `form:"yaml" json:"yaml" label:"Yaml"`
}

type IngressDeleteByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}
