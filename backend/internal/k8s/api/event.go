package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gkube/pkg/k8s"
	k8sEvent "gkube/pkg/k8s/event"
	"gkube/pkg/response"
	"k8s.io/apimachinery/pkg/watch"
)

type eventController struct{}

var Event = new(eventController)

func (e *eventController) ListEvents(c *gin.Context) {
	clusterName := c.Query("clusterName")
	namespace := c.Query("namespace")
	fieldSelector := c.Query("fieldSelector")
	limitStr := c.Query("limit")
	continueToken := c.Query("continue")

	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}

	var limit int64
	if limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}

	events, cont, rv, err := k8sEvent.ListEvents(client, namespace, fieldSelector, limit, continueToken)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取事件列表失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", gin.H{
		"items":           events,
		"continue":        cont,
		"resourceVersion": rv,
	})
}

func (e *eventController) WatchEvents(c *gin.Context) {
	clusterName := c.Query("clusterName")
	namespace := c.Query("namespace")
	fieldSelector := c.Query("fieldSelector")

	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}

	watcher, err := k8sEvent.WatchEvents(client, namespace, fieldSelector)
	if err != nil {
		response.Fail(c, fmt.Sprintf("Watch事件失败:%s", err.Error()))
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