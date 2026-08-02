package k8s

import (
	"net/http"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sEvent "gkube/pkg/k8s/event"
	k8sPod "gkube/pkg/k8s/pod"
	"gkube/pkg/logger"
	"gkube/pkg/response"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
)

type pod struct{}

var Pod = new(pod)

// GetPodList 获取pod列表（支持分页）
func (p *pod) GetPodList(c *gin.Context) {
	var query PodQueryListParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}

	limit, continueToken := k8sclient.GetPaginationParams(c)
	if limit > 0 {
		podList, err := k8sPod.ListPods(client, query.Namespace, limit, continueToken)
		if err != nil {
			logger.Error(err.Error())
			response.FailWithStatus(c, http.StatusBadGateway, "获取pod列表失败")
			return
		}
		remaining := int64(0)
		if podList.RemainingItemCount != nil {
			remaining = *podList.RemainingItemCount
		}
		data := k8sclient.BuildPaginatedData(podList.Items, podList.Continue, remaining, limit)
		data.Total = len(podList.Items)
		response.Success(c, "获取pod列表成功", data)
	} else {
		pods, err := k8sPod.GetPodList(client, query.Namespace)
		if err != nil {
			logger.Error(err.Error())
			response.FailWithStatus(c, http.StatusBadGateway, "获取pod列表失败")
			return
		}
		response.Success(c, "获取pod列表成功", pods)
	}
}

// GetPodByName 获取pod根据名称
func (p *pod) GetPodByName(c *gin.Context) {
	var query PodQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	pod, err := k8sPod.GetPodByName(client, query.Namespace, query.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取pod失败")
		return
	}
	response.Success(c, "获取pod成功", pod)
}

// GetPodYaml 获取pod的yaml
func (p *pod) GetPodYaml(c *gin.Context) {
	var query PodQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	podYaml, err := k8sPod.GetPodYaml(client, query.Namespace, query.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取pod失败")
		return
	}
	response.Success(c, "获取pod成功", podYaml)
}

// CreatePod 创建pod
func (p *pod) CreatePod(c *gin.Context) {
	var query PodCreateParams
	if err := c.ShouldBindJSON(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	if err := k8sPod.CreatePod(client, query.PodYaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "创建pod失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdatePod 更新pod
func (p *pod) UpdatePod(c *gin.Context) {
	var query PodUpdateParams
	if err := c.ShouldBindJSON(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	if err := k8sPod.UpdatePod(client, query.PodYaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "更新pod失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeletePodByName 删除pod根据名称
func (p *pod) DeletePodByName(c *gin.Context) {
	var query PodDeleteByNameParams
	if err := c.ShouldBindJSON(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	if err := k8sPod.DeletePodByName(client, query.Namespace, query.Name); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "删除pod失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

// WatchPodEvent streams K8s events for a specific Pod via SSE.
// Deprecated: prefer GET /v1/k8s/event/list or /v1/k8s/event/watch.
func (p *pod) WatchPodEvent(c *gin.Context) {
	var query PodEventQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}

	// 用 fields.Selector 构造,避免 PodName 注入非法字段选择器语法
	selector := fields.OneTermEqualSelector("involvedObject.name", query.PodName)
	watcher, err := k8sEvent.WatchEvents(client, query.Namespace, selector.String())
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "创建watcher失败")
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

type PodQueryListParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type PodQueryByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type PodCreateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	PodYaml     string `form:"podYaml" json:"podYaml" label:"Pod Yaml"`
}

type PodUpdateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	PodYaml     string `form:"podYaml" json:"podYaml" label:"Pod Yaml"`
}

type PodDeleteByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type PodEventQueryParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	PodName     string `form:"podName" json:"podName" label:"Pod名称"`
}
