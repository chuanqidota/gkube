package k8s

import (
	"net/http"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sPvc "gkube/pkg/k8s/pvc"
	"gkube/pkg/logger"
	"gkube/pkg/response"
)

// ---------------------------------------------------------------------------
// 特殊 handler —— PVC 的 pkg 函数 namespace 为 required
// ---------------------------------------------------------------------------

// GetPVCList 列表 —— 非分页
func GetPVCList(c *gin.Context) {
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
	pvcList, err := k8sPvc.GetPVCList(client, p.Namespace, selector)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取pvc列表失败")
		return
	}
	response.Success(c, "执行成功", pvcList)
}

// GetPVCListByStorageClass 按 StorageClass 查询 PVC
func GetPVCListByStorageClass(c *gin.Context) {
	var p struct {
		ClusterName      string `form:"clusterName" binding:"required"`
		StorageClassName string `form:"storageClassName" binding:"required"`
	}
	if err := c.ShouldBindQuery(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	pvcList, err := k8sPvc.GetPVCListByStorageClass(client, p.StorageClassName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取pvc列表失败")
		return
	}
	response.Success(c, "执行成功", pvcList)
}

// GetPVCByName 详情
func GetPVCByName(c *gin.Context) {
	var p NamespacedRequiredParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	pvc, err := k8sPvc.GetPVCByName(client, p.Namespace, p.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取pvc详情失败")
		return
	}
	response.Success(c, "执行成功", pvc)
}

// GetPVCYaml YAML
func GetPVCYaml(c *gin.Context) {
	var p NamespacedRequiredParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	yaml, err := k8sPvc.GetPVCYaml(client, p.Namespace, p.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取pvc yaml失败")
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yaml})
}

// CreatePVC 创建 —— namespace required
func CreatePVC(c *gin.Context) {
	var p struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace" binding:"required"`
		Yaml        string `json:"yaml" binding:"required"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	if err := k8sPvc.CreatePVC(client, p.Namespace, p.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "创建pvc失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdatePVC 更新 —— namespace required
func UpdatePVC(c *gin.Context) {
	var p struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace" binding:"required"`
		Yaml        string `json:"yaml" binding:"required"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	if err := k8sPvc.UpdatePVC(client, p.Namespace, p.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "更新pvc失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeletePVCByName 删除 —— namespace required
func DeletePVCByName(c *gin.Context) {
	var p NamespacedRequiredParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	if err := k8sPvc.DeletePVCByName(client, p.Namespace, p.Name); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "删除pvc失败")
		return
	}
	response.Success(c, "执行成功", nil)
}
