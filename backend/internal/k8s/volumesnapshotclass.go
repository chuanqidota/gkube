package k8s

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	k8sVolumeSnapshotClass "gkube/pkg/k8s/volumesnapshotclass"
	"gkube/pkg/logger"
	"gkube/pkg/response"
)

// ---------------------------------------------------------------------------
// 特殊 handler —— 使用 dynamic.Interface（集群级 CRD 资源）
// ---------------------------------------------------------------------------

// GetVolumeSnapshotClassList 列表 —— 非分页 + transform
func GetVolumeSnapshotClassList(c *gin.Context) {
	var p ListParams
	if err := c.ShouldBind(&p); err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := getDynamicClient(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	selector, err := buildLabelSelector(p.LabelFilters)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, err.Error())
		return
	}
	items, err := k8sVolumeSnapshotClass.GetVolumeSnapshotClassList(client, selector)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取VolumeSnapshotClass列表失败:%v", err.Error()))
		return
	}
	var result []map[string]any
	for _, item := range items {
		result = append(result, map[string]any{
			"name":           item.GetName(),
			"age":            item.GetCreationTimestamp().Time.Format("2006-01-02 15:04:05"),
			"labels":         item.GetLabels(),
			"annotations":    item.GetAnnotations(),
			"driver":         item.Object["driver"],
			"deletionPolicy": item.Object["deletionPolicy"],
			"parameters":     item.Object["parameters"],
		})
	}
	response.Success(c, "执行成功", result)
}

// GetVolumeSnapshotClassByName 详情
func GetVolumeSnapshotClassByName(c *gin.Context) {
	var p ClusterScopedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := getDynamicClient(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	obj, err := k8sVolumeSnapshotClass.GetVolumeSnapshotClassByName(client, p.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取VolumeSnapshotClass失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", obj.Object)
}

// GetVolumeSnapshotClassYaml YAML
func GetVolumeSnapshotClassYaml(c *gin.Context) {
	var p ClusterScopedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := getDynamicClient(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	yamlContent, err := k8sVolumeSnapshotClass.GetVolumeSnapshotClassYaml(client, p.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取VolumeSnapshotClass YAML失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yamlContent})
}

// CreateVolumeSnapshotClass 创建 —— 只传 yaml
func CreateVolumeSnapshotClass(c *gin.Context) {
	var p ClusterCreateParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := getDynamicClient(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sVolumeSnapshotClass.CreateVolumeSnapshotClass(client, p.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("创建VolumeSnapshotClass失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdateVolumeSnapshotClass 更新 —— 只传 yaml
func UpdateVolumeSnapshotClass(c *gin.Context) {
	var p ClusterCreateParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := getDynamicClient(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sVolumeSnapshotClass.UpdateVolumeSnapshotClass(client, p.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("更新VolumeSnapshotClass失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteVolumeSnapshotClassByName 删除 —— 原代码用 ShouldBindQuery
func DeleteVolumeSnapshotClassByName(c *gin.Context) {
	var p ClusterScopedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := getDynamicClient(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sVolumeSnapshotClass.DeleteVolumeSnapshotClassByName(client, p.Name); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("删除VolumeSnapshotClass失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}
