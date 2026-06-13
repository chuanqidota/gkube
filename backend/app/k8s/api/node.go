package api

import (
	"context"
	"fmt"
	"gkube/app/k8s/params"
	"gkube/pkg/response"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
	}
	err = k8sNode.EvictsNodeSinglePod(client, body.Namespace, body.PodName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("驱逐节点pod失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// SetTaintNode
//
//	@Description: 给节点设置污点
//	@receiver n
//	@param c
func (n *node) SetTaintNode(c *gin.Context) {
	var body params.TaintNodeParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}

	if err := k8sNode.SetTaintNode(client, body.NodeName, body.Key, body.Value, corev1.TaintEffect(body.Effect)); err != nil {
		response.Fail(c, fmt.Sprintf("设置污点失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// GetNodeDetail
//
//	@Description: 获取节点详情
//	@receiver n
//	@param c
func (n *node) GetNodeDetail(c *gin.Context) {
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	if name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	nodeObj, err := client.CoreV1().Nodes().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取节点详情失败:%s", err.Error()))
		return
	}

	// Get conditions
	var conditions []map[string]any
	for _, cond := range nodeObj.Status.Conditions {
		conditions = append(conditions, map[string]any{
			"type":               string(cond.Type),
			"status":             string(cond.Status),
			"reason":             cond.Reason,
			"message":            cond.Message,
			"lastTransitionTime": cond.LastTransitionTime.Time.Format("2006-01-02 15:04:05"),
		})
	}

	// Get addresses
	var internalIP, externalIP, hostname string
	for _, addr := range nodeObj.Status.Addresses {
		switch addr.Type {
		case corev1.NodeInternalIP:
			internalIP = addr.Address
		case corev1.NodeExternalIP:
			externalIP = addr.Address
		case corev1.NodeHostName:
			hostname = addr.Address
		}
	}

	// Get taints
	var taints []map[string]any
	for _, taint := range nodeObj.Spec.Taints {
		taints = append(taints, map[string]any{
			"key":    taint.Key,
			"value":  taint.Value,
			"effect": string(taint.Effect),
		})
	}

	// Get labels
	labels := make(map[string]string)
	for k, v := range nodeObj.Labels {
		labels[k] = v
	}

	// Determine status
	status := "Unknown"
	for _, cond := range nodeObj.Status.Conditions {
		if cond.Type == corev1.NodeReady {
			if cond.Status == corev1.ConditionTrue {
				status = "Ready"
			} else {
				status = "NotReady"
			}
			break
		}
	}

	// Get roles from labels
	var roles string
	for label := range nodeObj.Labels {
		if len(label) > 24 && label[:24] == "node-role.kubernetes.io/" {
			if roles != "" {
				roles += ", "
			}
			roles += label[24:]
		}
	}

	result := map[string]any{
		"name":               nodeObj.Name,
		"status":             status,
		"roles":              roles,
		"version":            nodeObj.Status.NodeInfo.KubeletVersion,
		"os":                 nodeObj.Status.NodeInfo.OSImage,
		"kernel":             nodeObj.Status.NodeInfo.KernelVersion,
		"container_runtime":  nodeObj.Status.NodeInfo.ContainerRuntimeVersion,
		"internal_ip":        internalIP,
		"external_ip":        externalIP,
		"hostname":           hostname,
		"unschedulable":      nodeObj.Spec.Unschedulable,
		"labels":             labels,
		"taints":             taints,
		"conditions":         conditions,
		"capacity":           nodeObj.Status.Capacity,
		"allocatable":        nodeObj.Status.Allocatable,
		"age":                nodeObj.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
	}
	response.Success(c, "执行成功", result)
}

// GetNodeEvents
//
//	@Description: 获取节点事件
//	@receiver n
//	@param c
func (n *node) GetNodeEvents(c *gin.Context) {
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	if name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	events, err := client.CoreV1().Events(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Node", name),
	})
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取节点事件失败:%s", err.Error()))
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

// UpdateNodeYaml
//
//	@Description: 更新节点（通过YAML）
//	@receiver n
//	@param c
func (n *node) UpdateNodeYaml(c *gin.Context) {
	var body struct {
		ClusterName string `json:"clusterName"`
		Name        string `json:"name"`
		Yaml        string `json:"yaml"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	// Get current node and update labels/taints
	nodeObj, err := client.CoreV1().Nodes().Get(context.TODO(), body.Name, metav1.GetOptions{})
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取节点失败:%s", err.Error()))
		return
	}
	_, err = client.CoreV1().Nodes().Update(context.TODO(), nodeObj, metav1.UpdateOptions{})
	if err != nil {
		response.Fail(c, fmt.Sprintf("更新节点失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}
