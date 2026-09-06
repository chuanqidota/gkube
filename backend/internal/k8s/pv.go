package k8s

import (
	"net/http"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sPv "gkube/pkg/k8s/pv"
	"gkube/pkg/logger"
	"gkube/pkg/response"
)

// ---------------------------------------------------------------------------
// 特殊 handler —— PV 是集群级资源，pkg 函数不接受 namespace
// ---------------------------------------------------------------------------

// GetPVList 列表 —— 非分页
func GetPVList(c *gin.Context) {
	var p ListParams
	if err := c.ShouldBind(&p); err != nil {
		logger.Error(err.Error())
		response.Fail(c, "参数错误")
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
		response.Fail(c, err.Error())
		return
	}
	pvList, err := k8sPv.GetPVList(client, selector)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取pv列表失败")
		return
	}
	response.Success(c, "执行成功", pvList)
}

// GetPVByName 详情
func GetPVByName(c *gin.Context) {
	var p ClusterScopedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		logger.Error(err.Error())
		response.Fail(c, "参数错误")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	pv, err := k8sPv.GetPVByName(client, p.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取pv详情失败")
		return
	}
	response.Success(c, "执行成功", pv)
}

// GetPVYaml YAML
func GetPVYaml(c *gin.Context) {
	var p ClusterScopedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		logger.Error(err.Error())
		response.Fail(c, "参数错误")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	yaml, err := k8sPv.GetPVYaml(client, p.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取pv yaml失败")
		return
	}
	response.Success(c, "执行成功", yaml)
}

// CreatePV 创建 —— 只传 yaml
func CreatePV(c *gin.Context) {
	var p ClusterCreateParams
	if err := c.ShouldBindJSON(&p); err != nil {
		logger.Error(err.Error())
		response.Fail(c, "参数错误")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	if err := k8sPv.CreatePV(client, p.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "创建pv失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdatePV 更新 —— 只传 yaml
func UpdatePV(c *gin.Context) {
	var p ClusterCreateParams
	if err := c.ShouldBindJSON(&p); err != nil {
		logger.Error(err.Error())
		response.Fail(c, "参数错误")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	if err := k8sPv.UpdatePV(client, p.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "更新pv失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeletePVByName 删除
func DeletePVByName(c *gin.Context) {
	var p ClusterScopedParams
	if err := c.ShouldBindJSON(&p); err != nil {
		logger.Error(err.Error())
		response.Fail(c, "参数错误")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	if err := k8sPv.DeletePVByName(client, p.Name); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "删除pv失败")
		return
	}
	response.Success(c, "执行成功", nil)
}
