package k8s

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sNp "gkube/pkg/k8s/networkpolicy"
	"gkube/pkg/logger"
	"gkube/pkg/response"
	"k8s.io/client-go/kubernetes"
)

// ---------------------------------------------------------------------------
// 标准 handler
// ---------------------------------------------------------------------------

var GetNetworkPolicyDetail = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sNp.GetNetworkPolicyDetail(client, namespace, name)
	},
	"执行成功", "获取NetworkPolicy失败",
)

var GetNetworkPolicyYaml = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sNp.GetNetworkPolicyYaml(client, namespace, name)
	},
	"执行成功", "获取NetworkPolicy YAML失败",
)

var GetNetworkPolicyPods = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sNp.GetNetworkPolicyPods(client, namespace, name)
	},
	"执行成功", "获取NetworkPolicy关联Pod失败",
)

var GetNetworkPolicyEvents = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		events, err := k8sNp.GetNetworkPolicyEvents(client, namespace, name)
		if err != nil {
			return nil, err
		}
		var result []map[string]any
		for _, event := range events {
			result = append(result, map[string]any{
				"type":      event.Type,
				"reason":    event.Reason,
				"message":   event.Message,
				"last_seen": event.LastTimestamp,
			})
		}
		return result, nil
	},
	"执行成功", "获取NetworkPolicy事件失败",
)

// ---------------------------------------------------------------------------
// 特殊 handler
// ---------------------------------------------------------------------------

// GetNetworkPolicyList 列表 —— 非分页 + transform
func GetNetworkPolicyList(c *gin.Context) {
	var p ListParams
	if err := c.ShouldBind(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	selector, err := buildLabelSelector(p.LabelFilters)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, err.Error())
		return
	}
	npList, err := k8sNp.GetNetworkPolicyList(client, p.Namespace, selector)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取NetworkPolicy列表失败")
		return
	}
	var result []map[string]any
	for _, np := range npList {
		var podSelector string
		if np.Spec.PodSelector.MatchLabels != nil {
			var parts []string
			for k, v := range np.Spec.PodSelector.MatchLabels {
				parts = append(parts, fmt.Sprintf("%s=%s", k, v))
			}
			podSelector = fmt.Sprintf("%v", parts)
		}
		var policyTypes []string
		for _, pt := range np.Spec.PolicyTypes {
			policyTypes = append(policyTypes, string(pt))
		}
		result = append(result, map[string]any{
			"name":         np.Name,
			"namespace":    np.Namespace,
			"pod_selector": podSelector,
			"policy_types": policyTypes,
			"ingress":      np.Spec.Ingress,
			"egress":       np.Spec.Egress,
			"labels":       np.Labels,
		})
	}
	response.Success(c, "执行成功", result)
}

// CreateNetworkPolicy 创建
func CreateNetworkPolicy(c *gin.Context) {
	var body struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace" binding:"required"`
		Yaml        string `json:"yaml" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("参数错误:%s", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sNp.CreateNetworkPolicy(client, body.Namespace, body.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("创建NetworkPolicy失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdateNetworkPolicy 更新
func UpdateNetworkPolicy(c *gin.Context) {
	var body struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace" binding:"required"`
		Yaml        string `json:"yaml" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("参数错误:%s", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sNp.UpdateNetworkPolicy(client, body.Namespace, body.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("更新NetworkPolicy失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteNetworkPolicy 删除 —— 原代码用 c.Query()，改为 ShouldBindQuery
func DeleteNetworkPolicy(c *gin.Context) {
	var p NamespacedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	if p.Name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	if p.ClusterName == "" {
		response.Fail(c, "clusterName参数不能为空")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	if err := k8sNp.DeleteNetworkPolicy(client, p.Namespace, p.Name); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "删除NetworkPolicy失败")
		return
	}
	response.Success(c, "执行成功", nil)
}
