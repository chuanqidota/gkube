package k8s

import (
	"context"

	"github.com/gin-gonic/gin"
	apperr "gkube/pkg/errors"
	k8sclient "gkube/pkg/k8s"
	k8sLr "gkube/pkg/k8s/limitrange"
	"gkube/pkg/response"
	"k8s.io/client-go/kubernetes"
)

// ---------------------------------------------------------------------------
// 标准 handler（使用 wrapper）
// ---------------------------------------------------------------------------

var CreateLimitRange = CreateHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, yaml string) error {
		return k8sLr.CreateLimitRange(ctx, client, namespace, yaml)
	},
	"执行成功",
)

// ---------------------------------------------------------------------------
// 特殊 handler（参数结构超出标准 wrapper 覆盖范围）
// ---------------------------------------------------------------------------

// UpdateLimitRange 更新 —— pkg 函数只传 namespace+yaml（无 name）
func UpdateLimitRange(c *gin.Context) {
	var body struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace"`
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
	if err := k8sLr.UpdateLimitRange(c.Request.Context(), client, body.Namespace, body.Yaml); err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", nil)
}

// GetLimitRangeList 列表 —— 非分页 + transform
func GetLimitRangeList(c *gin.Context) {
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
	lrList, err := k8sLr.GetLimitRangeList(c.Request.Context(), client, p.Namespace, selector)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	var result []map[string]any
	for _, lr := range lrList {
		var limits []map[string]any
		for _, l := range lr.Spec.Limits {
			limits = append(limits, map[string]any{
				"type":                 string(l.Type),
				"max":                  l.Max,
				"min":                  l.Min,
				"default":              l.Default,
				"defaultRequest":       l.DefaultRequest,
				"maxLimitRequestRatio": l.MaxLimitRequestRatio,
			})
		}
		result = append(result, map[string]any{
			"name":      lr.Name,
			"namespace": lr.Namespace,
			"limits":    limits,
			"labels":    lr.Labels,
		})
	}
	response.Success(c, "执行成功", result)
}

// GetLimitRangeDetail 详情
func GetLimitRangeDetail(c *gin.Context) {
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
	detail, err := k8sLr.GetLimitRangeDetail(c.Request.Context(), client, p.Namespace, p.Name)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", detail)
}

// GetLimitRangeYaml YAML
func GetLimitRangeYaml(c *gin.Context) {
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
	yaml, err := k8sLr.GetLimitRangeYaml(c.Request.Context(), client, p.Namespace, p.Name)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yaml})
}

// DeleteLimitRange 删除
func DeleteLimitRange(c *gin.Context) {
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
	if err := k8sLr.DeleteLimitRange(c.Request.Context(), client, p.Namespace, p.Name); err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", nil)
}
