package core

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gkube/internal/k8s/params"
	"gkube/pkg/k8s"
	k8sPod "gkube/pkg/k8s/pod"
	"gkube/pkg/response"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"time"
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

// GetPodByLabel
//
//	@Description: 获取pod根据标签
//	@receiver p
//	@param c
func (p *pod) GetPodByLabel(c *gin.Context) {
	var query params.PodQueryByLabelParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	pods, err := k8sPod.GetPodByLabel(client, query.Namespace, query.LabelMap)

	if err != nil {
		response.Fail(c, fmt.Sprintf("获取pod失败:%s", err.Error()))
		return
	}
	response.Success(c, "获取pod成功", pods)
}

// GetPodByField
//
//	@Description: 获取pod根据字段
//	@receiver p
//	@param c
func (p *pod) GetPodByField(c *gin.Context) {
	var query params.PodQueryByFiledParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}

	pods, err := k8sPod.GetPodByField(client, query.Namespace, query.FieldMap)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取pod失败:%s", err.Error()))
		return
	}
	response.Success(c, "获取pod成功", pods)
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

// DeletePodByLabel
//
//	@Description: 删除pod根据标签
//	@receiver p
//	@param c
func (p *pod) DeletePodByLabel(c *gin.Context) {
	var query params.PodDeleteByLabelParams
	if err := c.ShouldBindJSON(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sPod.DeletePodByLabel(client, query.Namespace, query.LabelMap); err != nil {
		response.Fail(c, fmt.Sprintf("删除pod失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeletePodByField
//
//	@Description: 删除pod根据字段
//	@receiver p
//	@param c
func (p *pod) DeletePodByField(c *gin.Context) {
	var query params.PodDeleteByFieldParams
	if err := c.ShouldBindJSON(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sPod.DeletePodByField(client, query.Namespace, query.FieldMap); err != nil {
		response.Fail(c, fmt.Sprintf("删除pod失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// WatchPodEvent
//
//	@Description: 监听pod的event事件
//	@receiver p
//	@param c
func (p *pod) WatchPodEvent(c *gin.Context) {
	var query params.PodEventQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}
	// 1. 初始化客户端
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	// 2. 创建Watcher
	watcher, err := client.CoreV1().Pods(query.Namespace).Watch(c.Request.Context(), metav1.ListOptions{
		FieldSelector: "metadata.name=" + query.PodName,
	})
	if err != nil {
		response.Fail(c, fmt.Sprintf("创建watcher失败:%s", err.Error()))
		return
	}
	defer watcher.Stop()
	// 3. 设置SSE响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// 4. 实时推送事件
	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return
			}
			if event.Type == watch.Error {
				c.SSEvent("error", gin.H{"message": "发生错误"})
				return
			} else {
				pod, ok := event.Object.(*corev1.Pod)
				if !ok {
					c.SSEvent("error", gin.H{"message": "unexpected event object type"})
					continue
				}
				c.SSEvent("message", gin.H{
					"type":      event.Type,
					"name":      pod.Name,
					"namespace": pod.Namespace,
					"status":    pod.Status.Phase,
					"message":   pod.Status.Message,
					"reason":    pod.Status.Reason,
					"time":      time.Now().Format(time.DateTime),
				})
				c.Writer.Flush()
			}
		case <-c.Request.Context().Done():
			return
		}
	}
}
