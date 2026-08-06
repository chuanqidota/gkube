package k8s

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sNode "gkube/pkg/k8s/node"
	"gkube/pkg/logger"
	"gkube/pkg/response"
	corev1 "k8s.io/api/core/v1"
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
	evicted, skipped, failed, err := k8sNode.DrainNode(client, body.Name, opts)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "驱逐节点失败")
		return
	}
	response.Success(c, "执行成功", map[string]any{"evicted": evicted, "skipped": skipped, "failed": failed})
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
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	detail, err := k8sNode.GetNodeDetail(client, query.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取节点详情失败")
		return
	}
	response.Success(c, "执行成功", detail)
}

// GetNodeEvents 获取节点事件
func (n *node) GetNodeEvents(c *gin.Context) {
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
	events, err := k8sNode.GetNodeEvents(client, query.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取节点事件失败")
		return
	}
	response.Success(c, "执行成功", events)
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
	if err := k8sNode.UpdateNodeYaml(client, body.Name, body.Yaml); err != nil {
		logger.Error(err.Error())
		if errors.Is(err, k8sNode.ErrYamlParse) {
			response.Fail(c, "YAML解析失败")
		} else {
			response.FailWithStatus(c, http.StatusBadGateway, "更新节点失败")
		}
		return
	}
	response.Success(c, "执行成功", nil)
}

type NodeQueryParams struct {
	ClusterName string `form:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" binding:"required" label:"节点名称"`
}

// CordonNodeParams 封锁/解除封锁节点参数
type CordonNodeParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"节点名称"`
	Cordon      *bool  `json:"cordon" binding:"required" label:"是否封锁"`
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
	DeleteLocalData  bool   `json:"deleteLocalData"`  // 是否删除本地数据（emptyDir/hostPath）
	GracePeriod      int    `json:"gracePeriod"`      // 优雅终止秒数，-1=默认
	Force            bool   `json:"force"`            // 驱逐不被控制器管理的 standalone Pod（与 kubectl --force 一致）
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
