package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gkube/app/k8s/params"
	"gkube/pkg/k8s"
	k8sIngress "gkube/pkg/k8s/ingress"
	"gkube/pkg/response"
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
	var query params.IngressQueryListParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, err.Error())
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
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
	var query params.IngressQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, err.Error())
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
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

// GetIngressByLabel
//
//	@Description: 获取ingress根据标签
//	@receiver i
//	@param c
func (i *ingress) GetIngressByLabel(c *gin.Context) {
	var body params.IngressQueryByLabelParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, err.Error())
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	ingressList, err := k8sIngress.GetIngressByLabel(client, body.Namespace, body.LabelMap)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ingress失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", ingressList)
}

// GetIngressByField
//
//	@Description: 获取ingress根据字段
//	@receiver i
//	@param c
func (i *ingress) GetIngressByField(c *gin.Context) {
	var body params.IngressQueryByFieldParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, err.Error())
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	ingressList, err := k8sIngress.GetIngressByFiled(client, body.Namespace, body.FieldMap)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ingress失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", ingressList)
}

// GetIngressYaml
//
//	@Description: 获取ingress的yaml
//	@receiver i
//	@param c
func (i *ingress) GetIngressYaml(c *gin.Context) {
	var query params.IngressQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
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
	var body params.IngressCreateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
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
	var body params.IngressUpdateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
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
	var query params.IngressDeleteByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)

	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	if err := k8sIngress.DeleteIngressByName(client, query.Namespace, query.Name); err != nil {
		response.Fail(c, fmt.Sprintf("删除ingress失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteIngressByLabel
//
//	@Description: 删除ingress根据标签
//	@receiver i
//	@param c
func (i *ingress) DeleteIngressByLabel(c *gin.Context) {
	var body params.IngressDeleteByLabelParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sIngress.DeleteIngressByLabel(client, body.Namespace, body.LabelMap); err != nil {
		response.Fail(c, fmt.Sprintf("删除ingress失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteIngressByField
//
//	@Description: 删除ingress根据字段
//	@receiver i
//	@param c
func (i *ingress) DeleteIngressByField(c *gin.Context) {
	var body params.IngressDeleteByFieldParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sIngress.DeleteIngressByField(client, body.Namespace, body.FieldMap); err != nil {
		response.Fail(c, fmt.Sprintf("删除ingress失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}
