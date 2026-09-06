package k8s

import (
	"net/http"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sLr "gkube/pkg/k8s/limitrange"
	"gkube/pkg/logger"
	"gkube/pkg/response"
	"k8s.io/client-go/kubernetes"
)

// ---------------------------------------------------------------------------
// 标准 handler
// ---------------------------------------------------------------------------

var CreateLimitRange = CreateHandler(
	func(client *kubernetes.Clientset, namespace, yaml string) error {
		return k8sLr.CreateLimitRange(client, namespace, yaml)
	},
	"执行成功", "创建LimitRange失败",
)

// UpdateLimitRange 更新 —— pkg 函数只传 namespace+yaml（无 name）
func UpdateLimitRange(c *gin.Context) {
	var body struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace"`
		Yaml        string `json:"yaml" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	if err := k8sLr.UpdateLimitRange(client, body.Namespace, body.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "更新LimitRange失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

// ---------------------------------------------------------------------------
// 特殊 handler
// ---------------------------------------------------------------------------

// GetLimitRangeList 列表 —— 非分页 + transform
func GetLimitRangeList(c *gin.Context) {
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
	lrList, err := k8sLr.GetLimitRangeList(client, p.Namespace, selector)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取LimitRange列表失败")
		return
	}
	var result []map[string]any
	for _, lr := range lrList {
		var limits []map[string]any
		for _, l := range lr.Spec.Limits {
			limits = append(limits, map[string]any{
				"type":                    string(l.Type),
				"max":                     l.Max,
				"min":                     l.Min,
				"default":                 l.Default,
				"defaultRequest":          l.DefaultRequest,
				"maxLimitRequestRatio":    l.MaxLimitRequestRatio,
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

// GetLimitRangeDetail 详情 —— 原代码用 c.Query()
func GetLimitRangeDetail(c *gin.Context) {
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
	detail, err := k8sLr.GetLimitRangeDetail(client, p.Namespace, p.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取LimitRange详情失败")
		return
	}
	response.Success(c, "执行成功", detail)
}

// GetLimitRangeYaml YAML —— 原代码用 c.Query()
func GetLimitRangeYaml(c *gin.Context) {
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
	yaml, err := k8sLr.GetLimitRangeYaml(client, p.Namespace, p.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取LimitRange YAML失败")
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yaml})
}

// DeleteLimitRange 删除 —— 原代码用 c.Query()
func DeleteLimitRange(c *gin.Context) {
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
	if err := k8sLr.DeleteLimitRange(client, p.Namespace, p.Name); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "删除LimitRange失败")
		return
	}
	response.Success(c, "执行成功", nil)
}
