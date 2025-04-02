package api

import (
	"fmt"
	"gkube/app/k8s/params"
	"gkube/pkg/response"

	"gkube/pkg/k8s"
	k8sNode "gkube/pkg/k8s/node"

	"github.com/gin-gonic/gin"
)

type node struct {
}

var Node = new(node)

// GetNodeYaml
//
//	@Description: 获取节点yaml
//	@receiver n
//	@param c
func (n *node) GetNodeYaml(c *gin.Context) {
	var query params.NodeQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return

	}
	yaml, err := k8sNode.GetNodeYaml(client, query.NodeName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取节点yaml失败:%s", err.Error()))
		return
	}
	result := map[string]any{
		"yaml": yaml,
	}
	response.Success(c, "执行成功", result)
}

// GetNodePods
//
//	@Description: 获取节点中的pod
//	@receiver n
//	@param c
func (n *node) GetNodePods(c *gin.Context) {
	var query params.NodeQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	pods, err := k8sNode.GetNodePods(client, query.NodeName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取节点pod失败:%s", err.Error()))
		return
	}
	result := map[string]any{
		"pods": pods,
	}
	response.Success(c, "执行成功", result)
}

// UnscheduledNode
//
//	@Description: 禁止调度
//	@receiver n
//	@param c
func (n *node) UnscheduledNode(c *gin.Context) {
	var body params.NodeQueryParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	isCordon, err := k8sNode.UnscheduledNode(client, body.NodeName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("禁止调度失败:%s", err.Error()))
		return
	}
	result := map[string]bool{
		"isCordon": isCordon,
	}
	response.Success(c, "执行成功", result)
}

// EvictsNodeAllPods
//
//	@Description: 驱逐节点中的所有pod
//	@receiver n
//	@param c
func (n *node) EvictsNodeAllPods(c *gin.Context) {
	var body params.NodeQueryParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	isEvict, err := k8sNode.EvictsNodeAllPods(client, body.NodeName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("驱逐节点pod失败:%s", err.Error()))
		return
	}
	result := map[string]bool{
		"isEvict": isEvict,
	}
	response.Success(c, "执行成功", result)
}

// EvictsNodeSinglePod
//
//	@Description: 驱逐节点中的指定pod
//	@receiver n
//	@param c
func (n *node) EvictsNodeSinglePod(c *gin.Context) {
	var body params.NodeEvictParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("驱逐节点pod失败:%s", err.Error()))
	}
	err = k8sNode.EvictsNodeSinglePod(client, body.Namespace, body.PodName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("驱逐节点pod失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}
