package k8s

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	apperr "gkube/pkg/errors"
	k8sclient "gkube/pkg/k8s"
	k8sNp "gkube/pkg/k8s/networkpolicy"
	"gkube/pkg/response"
	"k8s.io/client-go/kubernetes"
)

// ---------------------------------------------------------------------------
// 标准 handler（使用 wrapper）
// ---------------------------------------------------------------------------

var GetNetworkPolicyDetail = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sNp.GetNetworkPolicyDetail(ctx, client, namespace, name)
	},
	"执行成功",
)

var GetNetworkPolicyYaml = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sNp.GetNetworkPolicyYaml(ctx, client, namespace, name)
	},
	"执行成功",
)

var GetNetworkPolicyPods = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sNp.GetNetworkPolicyPods(ctx, client, namespace, name)
	},
	"执行成功",
)

var GetNetworkPolicyEvents = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		events, err := k8sNp.GetNetworkPolicyEvents(ctx, client, namespace, name)
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
	"执行成功",
)

// ---------------------------------------------------------------------------
// 特殊 handler
// ---------------------------------------------------------------------------

// GetNetworkPolicyList 列表 —— 非分页 + transform
func GetNetworkPolicyList(c *gin.Context) {
	var p ListParams
	if err := c.ShouldBind(&p); err != nil {
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	selector, err := buildLabelSelector(p.LabelFilters)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	npList, err := k8sNp.GetNetworkPolicyList(c.Request.Context(), client, p.Namespace, selector)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
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
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	if err := k8sNp.CreateNetworkPolicy(c.Request.Context(), client, body.Namespace, body.Yaml); err != nil {
		response.FailWithError(c, ensureAppError(err))
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
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	if err := k8sNp.UpdateNetworkPolicy(c.Request.Context(), client, body.Namespace, body.Yaml); err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteNetworkPolicy 删除
func DeleteNetworkPolicy(c *gin.Context) {
	var p NamespacedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}
	if p.Name == "" {
		response.FailWithError(c, apperr.BadRequest("name参数不能为空", nil))
		return
	}
	if p.ClusterName == "" {
		response.FailWithError(c, apperr.BadRequest("clusterName参数不能为空", nil))
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	if err := k8sNp.DeleteNetworkPolicy(c.Request.Context(), client, p.Namespace, p.Name); err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", nil)
}
