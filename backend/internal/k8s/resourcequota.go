package k8s

import (
	"net/http"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sRq "gkube/pkg/k8s/resourcequota"
	"gkube/pkg/logger"
	"gkube/pkg/response"
	"k8s.io/client-go/kubernetes"
)

// ---------------------------------------------------------------------------
// 标准 handler
// ---------------------------------------------------------------------------

var CreateResourceQuota = CreateHandler(
	func(client *kubernetes.Clientset, namespace, yaml string) error {
		return k8sRq.CreateResourceQuota(client, namespace, yaml)
	},
	"执行成功", "创建ResourceQuota失败",
)

// UpdateResourceQuota 更新 —— pkg 函数只传 namespace+yaml（无 name）
func UpdateResourceQuota(c *gin.Context) {
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
	if err := k8sRq.UpdateResourceQuota(client, body.Namespace, body.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "更新ResourceQuota失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

// ---------------------------------------------------------------------------
// 特殊 handler
// ---------------------------------------------------------------------------

// GetResourceQuotaList 列表 —— 非分页 + transform
func GetResourceQuotaList(c *gin.Context) {
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
	rqList, err := k8sRq.GetResourceQuotaList(client, p.Namespace, selector)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取ResourceQuota列表失败")
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

// GetResourceQuotaDetail 详情 —— 原代码用 c.Query()
func GetResourceQuotaDetail(c *gin.Context) {
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
	detail, err := k8sRq.GetResourceQuotaDetail(client, p.Namespace, p.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取ResourceQuota详情失败")
		return
	}
	response.Success(c, "执行成功", detail)
}

// GetResourceQuotaYaml YAML —— 原代码用 c.Query()
func GetResourceQuotaYaml(c *gin.Context) {
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
	yaml, err := k8sRq.GetResourceQuotaYaml(client, p.Namespace, p.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取ResourceQuota YAML失败")
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yaml})
}

// DeleteResourceQuota 删除 —— 原代码用 c.Query()
func DeleteResourceQuota(c *gin.Context) {
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
	if err := k8sRq.DeleteResourceQuota(client, p.Namespace, p.Name); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "删除ResourceQuota失败")
		return
	}
	response.Success(c, "执行成功", nil)
}
