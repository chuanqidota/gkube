package k8s

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sVolumeSnapshot "gkube/pkg/k8s/volumesnapshot"
	"gkube/pkg/logger"
	"gkube/pkg/response"
	"k8s.io/client-go/dynamic"
)

// getDynamicClient 从集群名称获取 dynamic client，供 volumesnapshot 和 volumesnapshotclass 共用
func getDynamicClient(clusterName string) (dynamic.Interface, error) {
	config, err := k8sclient.GetRestConfigByName(clusterName)
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(config)
}

// ---------------------------------------------------------------------------
// 特殊 handler —— 使用 dynamic.Interface 而非 *kubernetes.Clientset
// ---------------------------------------------------------------------------

// GetVolumeSnapshotList 列表 —— 非分页 + transform
func GetVolumeSnapshotList(c *gin.Context) {
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
	items, err := k8sVolumeSnapshot.GetVolumeSnapshotList(client, p.Namespace, selector)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取VolumeSnapshot列表失败:%v", err.Error()))
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
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := getDynamicClient(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	obj, err := k8sVolumeSnapshot.GetVolumeSnapshotByName(client, p.Namespace, p.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取VolumeSnapshot失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", obj.Object)
}

// GetVolumeSnapshotYaml YAML
func GetVolumeSnapshotYaml(c *gin.Context) {
	var p NamespacedParams
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
	yamlContent, err := k8sVolumeSnapshot.GetVolumeSnapshotYaml(client, p.Namespace, p.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取VolumeSnapshot YAML失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yamlContent})
}

// CreateVolumeSnapshot 创建
func CreateVolumeSnapshot(c *gin.Context) {
	var p CreateParams
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
	if err := k8sVolumeSnapshot.CreateVolumeSnapshot(client, p.Namespace, p.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("创建VolumeSnapshot失败:%v", err.Error()))
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
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := getDynamicClient(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sVolumeSnapshot.UpdateVolumeSnapshot(client, p.Namespace, p.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("更新VolumeSnapshot失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteVolumeSnapshotByName 删除
func DeleteVolumeSnapshotByName(c *gin.Context) {
	var p NamespacedParams
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
	if err := k8sVolumeSnapshot.DeleteVolumeSnapshotByName(client, p.Namespace, p.Name); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("删除VolumeSnapshot失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}
