package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gkube/app/k8s/params"
	k8sVolumeSnapshotClass "gkube/pkg/k8s/volumesnapshotclass"
	"gkube/pkg/response"
)

type volumeSnapshotClass struct{}

var VolumeSnapshotClass = new(volumeSnapshotClass)

// GetVolumeSnapshotClassList
//
//	@Description: 获取VolumeSnapshotClass列表
//	@receiver v
//	@param c
func (v *volumeSnapshotClass) GetVolumeSnapshotClassList(c *gin.Context) {
	var query params.VolumeSnapshotClassQueryListParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := getDynamicClient(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	items, err := k8sVolumeSnapshotClass.GetVolumeSnapshotClassList(client)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取VolumeSnapshotClass列表失败:%v", err.Error()))
		return
	}
	var result []map[string]any
	for _, item := range items {
		result = append(result, map[string]any{
			"name":            item.GetName(),
			"age":             item.GetCreationTimestamp().Time.Format("2006-01-02 15:04:05"),
			"labels":          item.GetLabels(),
			"annotations":     item.GetAnnotations(),
			"driver":          item.Object["driver"],
			"deletionPolicy":  item.Object["deletionPolicy"],
			"parameters":      item.Object["parameters"],
		})
	}
	response.Success(c, "执行成功", result)
}

// GetVolumeSnapshotClassByName
//
//	@Description: 根据名称获取VolumeSnapshotClass详情
//	@receiver v
//	@param c
func (v *volumeSnapshotClass) GetVolumeSnapshotClassByName(c *gin.Context) {
	var query params.VolumeSnapshotClassQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := getDynamicClient(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	obj, err := k8sVolumeSnapshotClass.GetVolumeSnapshotClassByName(client, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取VolumeSnapshotClass失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", obj.Object)
}

// GetVolumeSnapshotClassYaml
//
//	@Description: 获取VolumeSnapshotClass的YAML
//	@receiver v
//	@param c
func (v *volumeSnapshotClass) GetVolumeSnapshotClassYaml(c *gin.Context) {
	var query params.VolumeSnapshotClassQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := getDynamicClient(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	yamlContent, err := k8sVolumeSnapshotClass.GetVolumeSnapshotClassYaml(client, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取VolumeSnapshotClass YAML失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yamlContent})
}

// CreateVolumeSnapshotClass
//
//	@Description: 创建VolumeSnapshotClass
//	@receiver v
//	@param c
func (v *volumeSnapshotClass) CreateVolumeSnapshotClass(c *gin.Context) {
	var body params.VolumeSnapshotClassCreateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := getDynamicClient(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sVolumeSnapshotClass.CreateVolumeSnapshotClass(client, body.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("创建VolumeSnapshotClass失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdateVolumeSnapshotClass
//
//	@Description: 更新VolumeSnapshotClass
//	@receiver v
//	@param c
func (v *volumeSnapshotClass) UpdateVolumeSnapshotClass(c *gin.Context) {
	var body params.VolumeSnapshotClassUpdateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := getDynamicClient(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sVolumeSnapshotClass.UpdateVolumeSnapshotClass(client, body.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("更新VolumeSnapshotClass失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteVolumeSnapshotClassByName
//
//	@Description: 删除VolumeSnapshotClass
//	@receiver v
//	@param c
func (v *volumeSnapshotClass) DeleteVolumeSnapshotClassByName(c *gin.Context) {
	var body params.VolumeSnapshotClassDeleteByNameParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := getDynamicClient(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sVolumeSnapshotClass.DeleteVolumeSnapshotClassByName(client, body.Name); err != nil {
		response.Fail(c, fmt.Sprintf("删除VolumeSnapshotClass失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}
