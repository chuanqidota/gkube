package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gkube/app/k8s/params"
	"gkube/pkg/k8s"
	k8sStorageClass "gkube/pkg/k8s/storageclass"
	"gkube/pkg/response"
)

type storageClass struct {
}

var StorageClass = new(storageClass)

// GetStorageClassList
//
//	@Description: 获取sc列表
//	@receiver s
//	@param c
func (s *storageClass) GetStorageClassList(c *gin.Context) {
	var query params.StorageClassQueryListParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}

	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	storageClasses, err := k8sStorageClass.GetStorageClassList(client)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取StorageClass列表失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", storageClasses)
}

// GetStorageClassByName
//
//	@Description: 获取sc根据名称
//	@receiver s
//	@param c
func (s *storageClass) GetStorageClassByName(c *gin.Context) {
	var query params.StorageClassQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}

	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	storageClass, err := k8sStorageClass.GetStorageClassByName(client, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取StorageClass失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", storageClass)

}

// GetStorageClassByField
//
//	@Description: 获取sc根据字段
//	@receiver s
//	@param c
func (s *storageClass) GetStorageClassByField(c *gin.Context) {
	var body params.StorageClassQueryByFieldParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)

	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	storageClasses, err := k8sStorageClass.GetStorageClassByField(client, body.FieldMap)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取StorageClass失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", storageClasses)
}

// GetStorageClassByLabel
//
//	@Description: 获取sc根据标签
//	@receiver s
//	@param c
func (s *storageClass) GetStorageClassByLabel(c *gin.Context) {
	var body params.StorageClassQueryByLabelParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	storageClasses, err := k8sStorageClass.GetStorageClassByLabel(client, body.LabelMap)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取StorageClass失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", storageClasses)
}

// GetStorageClassYaml
//
//	@Description: 获取sc的yaml文件
//	@receiver s
//	@param c
func (s *storageClass) GetStorageClassYaml(c *gin.Context) {
	var query params.StorageClassQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}

	client, err := k8s.GetK8sClientByName(query.ClusterName)

	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	storageClassYaml, err := k8sStorageClass.GetStorageClassYaml(client, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取StorageClass失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", storageClassYaml)
}

// CreateStorageClass
//
//	@Description: 创建sc
//	@receiver s
//	@param c
func (s *storageClass) CreateStorageClass(c *gin.Context) {
	var body params.StorageClassCreateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}

	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
	}

	if err := k8sStorageClass.CreateStorageClass(client, body.StorageClassYaml); err != nil {
		response.Fail(c, fmt.Sprintf("创建StorageClass失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdateStorageClass
//
//	@Description: 更新sc
//	@receiver s
//	@param c
func (s *storageClass) UpdateStorageClass(c *gin.Context) {
	var body params.StorageClassUpdateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}

	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
	}

	if err := k8sStorageClass.UpdateStorageClass(client, body.StorageClassYaml); err != nil {
		response.Fail(c, fmt.Sprintf("更新StorageClass失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteStorageClassByName
//
//	@Description: 删除sc根据名称
//	@receiver s
//	@param c
func (s *storageClass) DeleteStorageClassByName(c *gin.Context) {
	var body params.StorageClassDeleteByNameParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}

	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
	}

	if err := k8sStorageClass.DeleteStorageClassByName(client, body.Name); err != nil {
		response.Fail(c, fmt.Sprintf("删除StorageClass失败:%v", err.Error()))
		return
	}

	response.Success(c, "执行成功", nil)
}

// DeleteStorageClassByField
//
//	@Description: 删除sc根据字段
//	@receiver s
//	@param c
func (s *storageClass) DeleteStorageClassByField(c *gin.Context) {
	var body params.StorageClassDeleteByFieldParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sStorageClass.DeleteStorageClassByField(client, body.FieldMap); err != nil {
		response.Fail(c, fmt.Sprintf("删除StorageClass失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteStorageClassByLabel
//
//	@Description: 删除sc根据标签
//	@receiver s
//	@param c
func (s *storageClass) DeleteStorageClassByLabel(c *gin.Context) {
	var body params.StorageClassDeleteByLabelParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sStorageClass.DeleteStorageClassByLabel(client, body.LabelMap); err != nil {
		response.Fail(c, fmt.Sprintf("删除StorageClass失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}
