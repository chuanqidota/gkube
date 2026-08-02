package k8s

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sNode "gkube/pkg/k8s/node"
	"gkube/pkg/logger"
	"gkube/pkg/response"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/yaml"
)

type node struct{}

var Node = new(node)

// GetNodeYaml 获取节点yaml
func (n *node) GetNodeYaml(c *gin.Context) {
	var query NodeQueryParams
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
	yamlStr, err := k8sNode.GetNodeYaml(client, query.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取节点yaml失败")
		return
	}
	response.Success(c, "执行成功", map[string]any{"yaml": yamlStr})
}

// GetNodePods 获取节点中的pod
func (n *node) GetNodePods(c *gin.Context) {
	var query NodeQueryParams
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
	pods, err := k8sNode.GetNodePods(client, query.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取节点pod失败")
		return
	}
	response.Success(c, "执行成功", pods)
}

// CordonNode 封锁或解除封锁节点
func (n *node) CordonNode(c *gin.Context) {
	var body CordonNodeParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	isCordon, err := k8sNode.CordonNode(client, body.Name, *body.Cordon)
	if err != nil {
		logger.Error(err.Error())
		action := "封锁失败"
		if !*body.Cordon {
			action = "解除封锁失败"
		}
		response.FailWithStatus(c, http.StatusBadGateway, action)
		return
	}
	response.Success(c, "执行成功", map[string]bool{"isCordon": isCordon})
}

// DrainNode 驱逐节点上的所有 pod（封锁+驱逐）
func (n *node) DrainNode(c *gin.Context) {
	var body DrainNodeParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	opts := k8sNode.DrainOptions{
		IgnoreDaemonSets: body.IgnoreDaemonSets,
		DeleteLocalData:  body.DeleteLocalData,
		GracePeriod:      body.GracePeriod,
		Force:            body.Force,
	}
	evicted, skipped, err := k8sNode.DrainNode(client, body.Name, opts)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "驱逐节点失败")
		return
	}
	response.Success(c, "执行成功", map[string]any{"evicted": evicted, "skipped": skipped})
}

// DeleteNode 删除节点
func (n *node) DeleteNode(c *gin.Context) {
	var body DeleteNodeParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	if err := k8sNode.DeleteNode(client, body.Name); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "删除节点失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdateNodeLabels 更新节点标签（替换式）
func (n *node) UpdateNodeLabels(c *gin.Context) {
	var body UpdateNodeLabelsParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	if err := k8sNode.UpdateNodeLabels(client, body.Name, body.Labels); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "更新标签失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

// SetTaintNode 给节点追加单个污点
func (n *node) SetTaintNode(c *gin.Context) {
	var body TaintNodeParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	if err := k8sNode.SetTaintNode(client, body.Name, body.Key, body.Value, corev1.TaintEffect(body.Effect)); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "设置污点失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdateNodeTaints 替换式更新节点污点
func (n *node) UpdateNodeTaints(c *gin.Context) {
	var body UpdateNodeTaintsParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	var taints []corev1.Taint
	for _, t := range body.Taints {
		if t.Key == "" {
			continue
		}
		taints = append(taints, corev1.Taint{
			Key:    t.Key,
			Value:  t.Value,
			Effect: corev1.TaintEffect(t.Effect),
		})
	}
	if err := k8sNode.UpdateNodeTaints(client, body.Name, taints); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "更新污点失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

// GetNodeDetail 获取节点详情
func (n *node) GetNodeDetail(c *gin.Context) {
	var query NodeQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	if query.Name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	nodeObj, err := client.CoreV1().Nodes().Get(context.TODO(), query.Name, metav1.GetOptions{})
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取节点详情失败")
		return
	}

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

	var taints []map[string]any
	for _, taint := range nodeObj.Spec.Taints {
		taints = append(taints, map[string]any{
			"key":    taint.Key,
			"value":  taint.Value,
			"effect": string(taint.Effect),
		})
	}

	labels := make(map[string]string)
	for k, v := range nodeObj.Labels {
		labels[k] = v
	}

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
		"name":              nodeObj.Name,
		"status":            status,
		"roles":             roles,
		"version":           nodeObj.Status.NodeInfo.KubeletVersion,
		"os":                nodeObj.Status.NodeInfo.OSImage,
		"kernel":            nodeObj.Status.NodeInfo.KernelVersion,
		"container_runtime": nodeObj.Status.NodeInfo.ContainerRuntimeVersion,
		"architecture":      nodeObj.Status.NodeInfo.Architecture,
		"internal_ip":       internalIP,
		"external_ip":       externalIP,
		"hostname":          hostname,
		"unschedulable":     nodeObj.Spec.Unschedulable,
		"labels":            labels,
		"taints":            taints,
		"conditions":        conditions,
		"capacity":          nodeObj.Status.Capacity,
		"allocatable":       nodeObj.Status.Allocatable,
		"age":               nodeObj.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
	}
	response.Success(c, "执行成功", result)
}

// GetNodeEvents 获取节点事件
func (n *node) GetNodeEvents(c *gin.Context) {
	var query NodeQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	if query.Name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	// 用 fields.Selector 构造,避免 Name 注入非法字段选择器语法
	selector := fields.AndSelectors(
		fields.OneTermEqualSelector("involvedObject.name", query.Name),
		fields.OneTermEqualSelector("involvedObject.kind", "Node"),
	)
	events, err := client.CoreV1().Events(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{
		FieldSelector: selector.String(),
	})
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取节点事件失败")
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

// UpdateNodeYaml 更新节点（通过YAML）
func (n *node) UpdateNodeYaml(c *gin.Context) {
	var body struct {
		ClusterName string `json:"clusterName"`
		Name        string `json:"name" binding:"required"`
		Yaml        string `json:"yaml" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	currentNode, err := client.CoreV1().Nodes().Get(context.TODO(), body.Name, metav1.GetOptions{})
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取节点失败")
		return
	}
	var nodeObj corev1.Node
	if err := yaml.Unmarshal([]byte(body.Yaml), &nodeObj); err != nil {
		logger.Error(err.Error())
		response.Fail(c, "YAML解析失败")
		return
	}
	nodeObj.ResourceVersion = currentNode.ResourceVersion
	if _, err := client.CoreV1().Nodes().Update(context.TODO(), &nodeObj, metav1.UpdateOptions{}); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "更新节点失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

type NodeQueryParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	NodeName    string `form:"nodeName" json:"nodeName" label:"节点名称"`
	Name        string `form:"name" json:"name" label:"节点名称"` // 兼容前端 name 参数
}

// CordonNodeParams 封锁/解除封锁节点参数
type CordonNodeParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"节点名称"`
	Cordon      *bool  `json:"cordon" binding:"required" label:"是否封锁"`
}

// TaintNodeParams 单个污点参数（保留兼容）
type TaintNodeParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"节点名称"`
	Key         string `form:"key" json:"key" binding:"required" label:"污点key"`
	Value       string `form:"value" json:"value" label:"污点value"`
	Effect      string `form:"effect" json:"effect" binding:"required" label:"污点effect"`
}

// UpdateNodeTaintsParams 批量更新污点参数（替换式）
type UpdateNodeTaintsParams struct {
	ClusterName string      `json:"clusterName"`
	Name        string      `json:"name" binding:"required" label:"节点名称"`
	Taints      []TaintItem `json:"taints" label:"污点列表"`
}

// TaintItem 污点项
type TaintItem struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Effect string `json:"effect"`
}

// DrainNodeParams 驱逐节点参数
type DrainNodeParams struct {
	ClusterName      string `json:"clusterName"`
	Name             string `json:"name" binding:"required" label:"节点名称"`
	IgnoreDaemonSets bool   `json:"ignoreDaemonSets"` // 是否忽略 DaemonSet
	DeleteLocalData  bool   `json:"deleteLocalData"`  // 是否删除本地数据
	GracePeriod      int    `json:"gracePeriod"`      // 优雅终止秒数，-1=默认
	Force            bool   `json:"force"`            // 是否强制驱逐
}

// DeleteNodeParams 删除节点参数
type DeleteNodeParams struct {
	ClusterName string `json:"clusterName"`
	Name        string `json:"name" binding:"required" label:"节点名称"`
}

// UpdateNodeLabelsParams 更新节点标签参数
type UpdateNodeLabelsParams struct {
	ClusterName string            `json:"clusterName"`
	Name        string            `json:"name" binding:"required" label:"节点名称"`
	Labels      map[string]string `json:"labels" label:"标签"`
}
