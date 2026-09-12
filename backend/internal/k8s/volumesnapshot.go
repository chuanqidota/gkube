package k8s

import (
	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sVolumeSnapshot "gkube/pkg/k8s/volumesnapshot"
	apperr "gkube/pkg/errors"
	"gkube/pkg/response"
)

// ---------------------------------------------------------------------------
// 特殊 handler —— 使用 dynamic.Interface 而非 *kubernetes.Clientset
// ---------------------------------------------------------------------------

// GetVolumeSnapshotList 列表 —— 非分页 + transform
func GetVolumeSnapshotList(c *gin.Context) {
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
	items, err := k8sVolumeSnapshot.GetVolumeSnapshotList(c.Request.Context(), client, p.Namespace, selector)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	var result []map[string]any
	for _, item := range items {
		result = append(result, map[string]any{
			"name":        item.GetName(),
			"namespace":   item.GetNamespace(),
			"age":         item.GetCreationTimestamp().Time.Format("2006-01-02 15:04:05"),
			"labels":      item.GetLabels(),
			"annotations": item.GetAnnotations(),
			"spec":        item.Object["spec"],
			"status":      item.Object["status"],
		})
	}
	response.Success(c, "执行成功", result)
}

// GetVolumeSnapshotByName 详情
func GetVolumeSnapshotByName(c *gin.Context) {
	var p NamespacedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.FailWithError(c, apperr.Validation("参数错误", err))
		return
	}
	client, err := k8sclient.GetDynamicClientByName(p.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	obj, err := k8sVolumeSnapshot.GetVolumeSnapshotByName(c.Request.Context(), client, p.Namespace, p.Name)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", obj.Object)
}

// GetVolumeSnapshotYaml YAML
func GetVolumeSnapshotYaml(c *gin.Context) {
	var p NamespacedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.FailWithError(c, apperr.Validation("参数错误", err))
		return
	}
	client, err := k8sclient.GetDynamicClientByName(p.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	yamlContent, err := k8sVolumeSnapshot.GetVolumeSnapshotYaml(c.Request.Context(), client, p.Namespace, p.Name)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yamlContent})
}

// CreateVolumeSnapshot 创建
func CreateVolumeSnapshot(c *gin.Context) {
	var p CreateParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithError(c, apperr.Validation("参数错误", err))
		return
	}
	client, err := k8sclient.GetDynamicClientByName(p.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	if err := k8sVolumeSnapshot.CreateVolumeSnapshot(c.Request.Context(), client, p.Namespace, p.Yaml); err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdateVolumeSnapshot 更新
func UpdateVolumeSnapshot(c *gin.Context) {
	var p struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace"`
		Yaml        string `json:"yaml" binding:"required"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithError(c, apperr.Validation("参数错误", err))
		return
	}
	client, err := k8sclient.GetDynamicClientByName(p.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	if err := k8sVolumeSnapshot.UpdateVolumeSnapshot(c.Request.Context(), client, p.Namespace, p.Yaml); err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteVolumeSnapshotByName 删除
func DeleteVolumeSnapshotByName(c *gin.Context) {
	var p NamespacedParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithError(c, apperr.Validation("参数错误", err))
		return
	}
	client, err := k8sclient.GetDynamicClientByName(p.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	if err := k8sVolumeSnapshot.DeleteVolumeSnapshotByName(c.Request.Context(), client, p.Namespace, p.Name); err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", nil)
}
