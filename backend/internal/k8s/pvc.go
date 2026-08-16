package k8s

import (
	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sPvc "gkube/pkg/k8s/pvc"
	"gkube/pkg/logger"
	"gkube/pkg/response"
)

type pvc struct {
}

var Pvc = new(pvc)

// GetPVCList
//
//	@Description: 获取pvc列表
//	@receiver p
//	@param c
func (p *pvc) GetPVCList(c *gin.Context) {
	var query PvcListParams
	if err := c.ShouldBindQuery(&query); err != nil {
		logger.Error(err.Error())
		response.Fail(c, "参数错误")
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	pvcList, err := k8sPvc.GetPVCList(client, query.Namespace)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "查询pvc列表失败")
		return
	}
	response.Success(c, "执行成功", pvcList)
}

// GetPVCListByStorageClass
//
//	@Description: 根据 storageClassName 获取 PVC 列表（跨命名空间）
//	@receiver p
//	@param c
func (p *pvc) GetPVCListByStorageClass(c *gin.Context) {
	var query PvcListByStorageClassParams
	if err := c.ShouldBindQuery(&query); err != nil {
		logger.Error(err.Error())
		response.Fail(c, "参数错误")
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	pvcList, err := k8sPvc.GetPVCListByStorageClass(client, query.StorageClassName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "查询pvc列表失败")
		return
	}
	response.Success(c, "执行成功", pvcList)
}

// GetPVCByName
//
//	@Description: 根据名称获取pvc
//	@receiver p
//	@param c
func (p *pvc) GetPVCByName(c *gin.Context) {
	var query PvcQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		logger.Error(err.Error())
		response.Fail(c, "参数错误")
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	pvc, err := k8sPvc.GetPVCByName(client, query.Namespace, query.Name)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "查询pvc详情失败")
		return
	}
	response.Success(c, "执行成功", pvc)
}

// GetPVCYaml
//
//	@Description: 获取pvc的yaml
//	@receiver p
//	@param c
func (p *pvc) GetPVCYaml(c *gin.Context) {
	var query PvcQueryYamlParams
	if err := c.ShouldBindQuery(&query); err != nil {
		logger.Error(err.Error())
		response.Fail(c, "参数错误")
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	pvcYaml, err := k8sPvc.GetPVCYaml(client, query.Namespace, query.Name)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "查询pvc yaml失败")
		return
	}
	response.Success(c, "执行成功", pvcYaml)
}

// CreatePVC
//
//	@Description: 创建pvc
//	@receiver p
//	@param c
func (p *pvc) CreatePVC(c *gin.Context) {
	var body PvcCreateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		logger.Error(err.Error())
		response.Fail(c, "参数错误")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	if err := k8sPvc.CreatePVC(client, body.Namespace, body.Yaml); err != nil {
		logger.Error(err.Error())
		response.Fail(c, "创建pvc失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdatePVC
//
//	@Description: 更新pvc
//	@receiver p
//	@param c
func (p *pvc) UpdatePVC(c *gin.Context) {
	var body PvcUpdateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		logger.Error(err.Error())
		response.Fail(c, "参数错误")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	if err := k8sPvc.UpdatePVC(client, body.Namespace, body.Yaml); err != nil {
		logger.Error(err.Error())
		response.Fail(c, "更新pvc失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeletePVCByName
//
//	@Description: 根据名称删除pvc
//	@receiver p
//	@param c
func (p *pvc) DeletePVCByName(c *gin.Context) {
	var body PvcDeleteByNameParams
	if err := c.ShouldBindJSON(&body); err != nil {
		logger.Error(err.Error())
		response.Fail(c, "参数错误")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	if err := k8sPvc.DeletePVCByName(client, body.Namespace, body.Name); err != nil {
		logger.Error(err.Error())
		response.Fail(c, "删除pvc失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

type PvcListParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type PvcListByStorageClassParams struct {
	ClusterName      string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	StorageClassName string `form:"storageClassName" json:"storageClassName" binding:"required" label:"存储类名称"`
}

type PvcQueryByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" binding:"required" label:"命名空间"`
}

type PvcQueryYamlParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" binding:"required" label:"命名空间"`
}

type PvcCreateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" binding:"required" label:"命名空间"`
	Yaml        string `form:"yaml" json:"yaml" binding:"required" label:"Yaml"`
}

type PvcUpdateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" binding:"required" label:"命名空间"`
	Yaml        string `form:"yaml" json:"yaml" binding:"required" label:"Yaml"`
}

type PvcDeleteByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" binding:"required" label:"命名空间"`
}
