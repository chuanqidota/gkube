package core

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gkube/internal/k8s/params"
	"gkube/pkg/k8s"
	k8sPod "gkube/pkg/k8s/pod"
	k8sEvent "gkube/pkg/k8s/event"
	"gkube/pkg/response"
	"k8s.io/apimachinery/pkg/watch"
)

type pod struct {
}

var Pod = new(pod)

// GetPodList
//
//	@Description: 获取pod列表（支持分页）
//	@receiver p
//	@param c
func (p *pod) GetPodList(c *gin.Context) {
	var query params.PodQueryListParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}

	limit, continueToken := k8s.GetPaginationParams(c)
	if limit > 0 {
		podList, err := k8sPod.ListPods(client, query.Namespace, limit, continueToken)
		if err != nil {
			response.Fail(c, fmt.Sprintf("获取pod列表失败:%s", err.Error()))
			return
		}
		remaining := int64(0)
		if podList.RemainingItemCount != nil {
			remaining = *podList.RemainingItemCount
		}
		data := k8s.BuildPaginatedData(podList.Items, podList.Continue, remaining, limit)
		data.Total = len(podList.Items)
		response.Success(c, "获取pod列表成功", data)
	} else {
		pods, err := k8sPod.GetPodList(client, query.Namespace)
		if err != nil {
			response.Fail(c, fmt.Sprintf("获取pod列表失败:%s", err.Error()))
			return
		}
		response.Success(c, "获取pod列表成功", pods)
	}
}

// GetPodByName
//
//	@Description: 获取pod根据名称
//	@receiver p
//	@param c
func (p *pod) GetPodByName(c *gin.Context) {
	var query params.PodQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	pod, err := k8sPod.GetPodByName(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取pod失败:%s", err.Error()))
		return
	}
	response.Success(c, "获取pod成功", pod)
}

// GetPodYaml
//
//	@Description: 获取pod的yaml
//	@receiver p
//	@param c
func (p *pod) GetPodYaml(c *gin.Context) {
	var query params.PodQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	podYaml, err := k8sPod.GetPodYaml(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取pod失败:%s", err.Error()))
		return
	}
	response.Success(c, "获取pod成功", podYaml)
}

// CreatePod
//
//	@Description: 创建pod
//	@receiver p
//	@param c
func (p *pod) CreatePod(c *gin.Context) {
	var query params.PodCreateParams
	if err := c.ShouldBindJSON(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}

	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sPod.CreatePod(client, query.PodYaml); err != nil {
		response.Fail(c, fmt.Sprintf("创建pod失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdatePod
//
//	@Description: 更新pod
//	@receiver p
//	@param c
func (p *pod) UpdatePod(c *gin.Context) {
	var query params.PodUpdateParams
	if err := c.ShouldBindJSON(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sPod.UpdatePod(client, query.PodYaml); err != nil {
		response.Fail(c, fmt.Sprintf("更新pod失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeletePodByName
//
//	@Description: 删除pod根据名称
//	@receiver p
//	@param c
func (p *pod) DeletePodByName(c *gin.Context) {
	var query params.PodDeleteByNameParams
	if err := c.ShouldBindJSON(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sPod.DeletePodByName(client, query.Namespace, query.Name); err != nil {
		response.Fail(c, fmt.Sprintf("删除pod失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// WatchPodEvent
//
//	@Description: streams K8s events for a specific Pod via SSE.
//	Deprecated: prefer GET /v1/k8s/event/list or /v1/k8s/event/watch.
//	This endpoint is kept as a backward-compatible alias.
//
//	@receiver p
//	@param c
func (p *pod) WatchPodEvent(c *gin.Context) {
	var query params.PodEventQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}

	fieldSelector := fmt.Sprintf("involvedObject.name=%s", query.PodName)
	watcher, err := k8sEvent.WatchEvents(client, query.Namespace, fieldSelector)
	if err != nil {
		response.Fail(c, fmt.Sprintf("创建watcher失败:%s", err.Error()))
		return
	}
	defer watcher.Stop()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	ctx := c.Request.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return
			}
			if event.Type == watch.Error {
				c.SSEvent("error", gin.H{"message": "watch error occurred"})
				return
			}

			c.SSEvent("message", event.Object)
			c.Writer.Flush()
		}
	}
}
