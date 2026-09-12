package k8s

import (
	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sVolumeSnapshotClass "gkube/pkg/k8s/volumesnapshotclass"
	apperr "gkube/pkg/errors"
	"gkube/pkg/response"
)

// ---------------------------------------------------------------------------
// 特殊 handler —— 使用 dynamic.Interface（集群级 CRD 资源）
// ---------------------------------------------------------------------------

// GetVolumeSnapshotClassList 列表 —— 非分页 + transform
func GetVolumeSnapshotClassList(c *gin.Context) {
	var p ListParams
	if err := c.ShouldBind(&p); err != nil {
		response.FailWithError(c, apperr.Validation("参数错误", err))
		return
	}
	client, err := k8sclient.GetDynamicClientByName(p.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	selector, err := buildLabelSelector(p.LabelFilters)
	if err != nil {
		response.FailWithError(c, apperr.Validation("标签选择器错误", err))
		return
	}
	items, err := k8sVolumeSnapshotClass.GetVolumeSnapshotClassList(c.Request.Context(), client, selector)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
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
		response.FailWithError(c, apperr.Validation("参数错误", err))
		return
	}
	client, err := k8sclient.GetDynamicClientByName(p.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	obj, err := k8sVolumeSnapshotClass.GetVolumeSnapshotClassByName(c.Request.Context(), client, p.Name)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", obj.Object)
}

// GetVolumeSnapshotClassYaml YAML
func GetVolumeSnapshotClassYaml(c *gin.Context) {
	var p ClusterScopedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.FailWithError(c, apperr.Validation("参数错误", err))
		return
	}
	client, err := k8sclient.GetDynamicClientByName(p.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	yamlContent, err := k8sVolumeSnapshotClass.GetVolumeSnapshotClassYaml(c.Request.Context(), client, p.Name)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yamlContent})
}

// CreateVolumeSnapshotClass 创建 —— 只传 yaml
func CreateVolumeSnapshotClass(c *gin.Context) {
	var p ClusterCreateParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithError(c, apperr.Validation("参数错误", err))
		return
	}
	client, err := k8sclient.GetDynamicClientByName(p.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	if err := k8sVolumeSnapshotClass.CreateVolumeSnapshotClass(c.Request.Context(), client, p.Yaml); err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdateVolumeSnapshotClass 更新 —— 只传 yaml
func UpdateVolumeSnapshotClass(c *gin.Context) {
	var p ClusterCreateParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithError(c, apperr.Validation("参数错误", err))
		return
	}
	client, err := k8sclient.GetDynamicClientByName(p.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	if err := k8sVolumeSnapshotClass.UpdateVolumeSnapshotClass(c.Request.Context(), client, p.Yaml); err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteVolumeSnapshotClassByName 删除 —— 原代码用 ShouldBindQuery
func DeleteVolumeSnapshotClassByName(c *gin.Context) {
	var p ClusterScopedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.FailWithError(c, apperr.Validation("参数错误", err))
		return
	}
	client, err := k8sclient.GetDynamicClientByName(p.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	if err := k8sVolumeSnapshotClass.DeleteVolumeSnapshotClassByName(c.Request.Context(), client, p.Name); err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", nil)
}
