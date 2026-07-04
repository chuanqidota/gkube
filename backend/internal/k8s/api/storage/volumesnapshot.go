package storage

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gkube/internal/k8s/params"
	"gkube/pkg/k8s"
	k8sVolumeSnapshot "gkube/pkg/k8s/volumesnapshot"
	"gkube/pkg/response"
	"k8s.io/client-go/dynamic"
)

type volumeSnapshot struct{}

var VolumeSnapshot = new(volumeSnapshot)

func getDynamicClient(clusterName string) (dynamic.Interface, error) {
	config, err := k8s.GetRestConfigByName(clusterName)
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(config)
}

// GetVolumeSnapshotList
//
//	@Description: 获取VolumeSnapshot列表
//	@receiver v
//	@param c
func (v *volumeSnapshot) GetVolumeSnapshotList(c *gin.Context) {
	var query params.VolumeSnapshotQueryListParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := getDynamicClient(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	items, err := k8sVolumeSnapshot.GetVolumeSnapshotList(client, query.Namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取VolumeSnapshot列表失败:%v", err.Error()))
		return
	}
	var result []map[string]any
	for _, item := range items {
		result = append(result, map[string]any{
			"name":              item.GetName(),
			"namespace":         item.GetNamespace(),
			"age":               item.GetCreationTimestamp().Time.Format("2006-01-02 15:04:05"),
			"labels":            item.GetLabels(),
			"annotations":       item.GetAnnotations(),
			"spec":              item.Object["spec"],
			"status":            item.Object["status"],
		})
	}
	response.Success(c, "执行成功", result)
}

// GetVolumeSnapshotByName
//
//	@Description: 根据名称获取VolumeSnapshot详情
//	@receiver v
//	@param c
func (v *volumeSnapshot) GetVolumeSnapshotByName(c *gin.Context) {
	var query params.VolumeSnapshotQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := getDynamicClient(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	obj, err := k8sVolumeSnapshot.GetVolumeSnapshotByName(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取VolumeSnapshot失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", obj.Object)
}

// GetVolumeSnapshotYaml
//
//	@Description: 获取VolumeSnapshot的YAML
//	@receiver v
//	@param c
func (v *volumeSnapshot) GetVolumeSnapshotYaml(c *gin.Context) {
	var query params.VolumeSnapshotQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := getDynamicClient(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	yamlContent, err := k8sVolumeSnapshot.GetVolumeSnapshotYaml(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取VolumeSnapshot YAML失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yamlContent})
}

// CreateVolumeSnapshot
//
//	@Description: 创建VolumeSnapshot
//	@receiver v
//	@param c
func (v *volumeSnapshot) CreateVolumeSnapshot(c *gin.Context) {
	var body params.VolumeSnapshotCreateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := getDynamicClient(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sVolumeSnapshot.CreateVolumeSnapshot(client, body.Namespace, body.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("创建VolumeSnapshot失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdateVolumeSnapshot
//
//	@Description: 更新VolumeSnapshot
//	@receiver v
//	@param c
func (v *volumeSnapshot) UpdateVolumeSnapshot(c *gin.Context) {
	var body params.VolumeSnapshotUpdateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := getDynamicClient(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sVolumeSnapshot.UpdateVolumeSnapshot(client, body.Namespace, body.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("更新VolumeSnapshot失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteVolumeSnapshotByName
//
//	@Description: 删除VolumeSnapshot
//	@receiver v
//	@param c
func (v *volumeSnapshot) DeleteVolumeSnapshotByName(c *gin.Context) {
	var body params.VolumeSnapshotDeleteByNameParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := getDynamicClient(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sVolumeSnapshot.DeleteVolumeSnapshotByName(client, body.Namespace, body.Name); err != nil {
		response.Fail(c, fmt.Sprintf("删除VolumeSnapshot失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}
