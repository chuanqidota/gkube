package k8s

import (
	"context"

	"github.com/gin-gonic/gin"
	apperr "gkube/pkg/errors"
	k8sclient "gkube/pkg/k8s"
	k8sRq "gkube/pkg/k8s/resourcequota"
	"gkube/pkg/response"
	"k8s.io/client-go/kubernetes"
)

// ---------------------------------------------------------------------------
// 标准 handler（使用 wrapper）
// ---------------------------------------------------------------------------

var CreateResourceQuota = CreateHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, yaml string) error {
		return k8sRq.CreateResourceQuota(ctx, client, namespace, yaml)
	},
	"执行成功",
)

// ---------------------------------------------------------------------------
// 特殊 handler（参数结构超出标准 wrapper 覆盖范围）
// ---------------------------------------------------------------------------

// UpdateResourceQuota 更新 —— pkg 函数只传 namespace+yaml（无 name）
func UpdateResourceQuota(c *gin.Context) {
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
	if err := k8sRq.UpdateResourceQuota(c.Request.Context(), client, body.Namespace, body.Yaml); err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", nil)
}

// GetResourceQuotaList 列表 —— 非分页 + transform
func GetResourceQuotaList(c *gin.Context) {
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
	rqList, err := k8sRq.GetResourceQuotaList(c.Request.Context(), client, p.Namespace, selector)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	var result []map[string]any
	for _, rq := range rqList {
		hard := make(map[string]string)
		for k, v := range rq.Spec.Hard {
			hard[string(k)] = v.String()
		}
		used := make(map[string]string)
		for k, v := range rq.Status.Used {
			used[string(k)] = v.String()
		}
		result = append(result, map[string]any{
			"name":      rq.Name,
			"namespace": rq.Namespace,
			"hard":      hard,
			"used":      used,
			"labels":    rq.Labels,
		})
	}
	response.Success(c, "执行成功", result)
}

// GetResourceQuotaDetail 详情
func GetResourceQuotaDetail(c *gin.Context) {
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
	detail, err := k8sRq.GetResourceQuotaDetail(c.Request.Context(), client, p.Namespace, p.Name)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", detail)
}

// GetResourceQuotaYaml YAML
func GetResourceQuotaYaml(c *gin.Context) {
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
	yaml, err := k8sRq.GetResourceQuotaYaml(c.Request.Context(), client, p.Namespace, p.Name)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yaml})
}

// DeleteResourceQuota 删除
func DeleteResourceQuota(c *gin.Context) {
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
	if err := k8sRq.DeleteResourceQuota(c.Request.Context(), client, p.Namespace, p.Name); err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", nil)
}
