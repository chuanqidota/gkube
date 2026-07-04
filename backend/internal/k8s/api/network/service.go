package network

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"gkube/internal/k8s/params"
	"gkube/pkg/k8s"
	k8sService "gkube/pkg/k8s/service"
	"gkube/pkg/response"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	limit, continueToken := k8s.GetPaginationParams(c)
	if limit > 0 {
		svcList, err := k8sService.ListServices(client, query.Namespace, limit, continueToken)
		if err != nil {
			response.Fail(c, err.Error())
			return
		}
		remaining := int64(0)
		if svcList.RemainingItemCount != nil {
			remaining = *svcList.RemainingItemCount
		}
		data := k8s.BuildPaginatedData(svcList.Items, svcList.Continue, remaining, limit)
		data.Total = len(svcList.Items)
		response.Success(c, "获取成功", data)
	} else {
		services, err := k8sService.GetServicesList(client, query.Namespace)
		if err != nil {
			response.Fail(c, err.Error())
			return
		}
		response.Success(c, "获取成功", services)
	}
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
	if err := k8sService.CreateService(client, body.Namespace, body.Yaml); err != nil {
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

	if err := k8sService.UpdateService(client, body.Yaml); err != nil {
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

// GetServiceEvents
//
//	@Description: 获取service事件
//	@receiver s
//	@param c
func (s *service) GetServiceEvents(c *gin.Context) {
	var query params.ServiceQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	events, err := client.CoreV1().Events(query.Namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Service", query.Name),
	})
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取service事件失败:%s", err.Error()))
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

// ServicePodList
//
//	@Description: 获取service关联的pod列表
//	@receiver s
//	@param c
func (s *service) ServicePodList(c *gin.Context) {
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
	podList, err := k8sService.ServicePodList(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取service关联pod列表失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", podList)
}
