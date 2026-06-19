package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gkube/app/k8s/params"
	"gkube/pkg/k8s"
	k8sService "gkube/pkg/k8s/service"
	"gkube/pkg/response"
)

type service struct {
}

var Service = new(service)

// GetServicesList
//
//	@Description: 获取svc列表
//	@receiver s
//	@param c
func (s *service) GetServicesList(c *gin.Context) {
	var query params.ServiceQueryListParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	services, err := k8sService.GetServicesList(client, query.Namespace)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, "获取成功", services)
}

// GetServicesByName
//
//	@Description: 获取svc根据名称
//	@receiver s
//	@param c
func (s *service) GetServicesByName(c *gin.Context) {
	var query params.ServiceQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	service, err := k8sService.GetServicesByName(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, "获取成功", service)
}

// GetServicesByLabel
//
//	@Description: 获取svc根据标签
//	@receiver s
//	@param c
func (s *service) GetServicesByLabel(c *gin.Context) {
	var body params.ServiceQueryByLabelParams
	if err := c.ShouldBindQuery(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	services, err := k8sService.GetServicesByLabel(client, body.Namespace, body.LabelMap)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, "获取成功", services)
}

// GetServicesByField
//
//	@Description: 获取svc根据字段
//	@receiver s
//	@param c
func (s *service) GetServicesByField(c *gin.Context) {
	var body params.ServiceQueryByFieldParams
	if err := c.ShouldBindQuery(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	services, err := k8sService.GetServicesByField(client, body.Namespace, body.FieldMap)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, "获取成功", services)
}

// GetServicesYaml
//
//	@Description: 获取svc的yaml
//	@receiver s
//	@param c
func (s *service) GetServicesYaml(c *gin.Context) {
	var query params.ServiceQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	yaml, err := k8sService.GetServicesYaml(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, "获取成功", yaml)
}

// CreateService
//
//	@Description: 创建svc
//	@receiver s
//	@param c
func (s *service) CreateService(c *gin.Context) {
	var body params.ServiceCreateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sService.CreateService(client, body.Namespace, body.ServiceYaml); err != nil {
		response.Fail(c, fmt.Sprintf("创建Service失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdateService
//
//	@Description: 更新svc
//	@receiver s
//	@param c
func (s *service) UpdateService(c *gin.Context) {
	var body params.ServiceUpdateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	if err := k8sService.UpdateService(client, body.ServiceYaml); err != nil {
		response.Fail(c, fmt.Sprintf("更新Service失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteService
//
//	@Description: 删除svc
//	@receiver s
//	@param c
func (s *service) DeleteService(c *gin.Context) {
	var body params.ServiceDeleteParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sService.DeleteService(client, body.Namespace, body.Name); err != nil {
		response.Fail(c, fmt.Sprintf("删除Service失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}
