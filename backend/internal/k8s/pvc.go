package k8s

import (
	"fmt"
	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sPvc "gkube/pkg/k8s/pvc"
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
		response.Fail(c, err.Error())
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	pvcList, err := k8sPvc.GetPVCList(client, query.Namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("查询pvc失败:%v", err.Error()))
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
		response.Fail(c, err.Error())
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	pvc, err := k8sPvc.GetPVCByName(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("查询pvc失败:%v", err.Error()))
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
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	pvcYaml, err := k8sPvc.GetPVCYaml(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("查询pvc失败:%v", err.Error()))
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
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sPvc.CreatePVC(client, body.Namespace, body.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("创建pvc失败:%v", err.Error()))
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
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sPvc.UpdatePVC(client, body.Namespace, body.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("更新pvc失败:%v", err.Error()))
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
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sPvc.DeletePVCByName(client, body.Namespace, body.Name); err != nil {
		response.Fail(c, fmt.Sprintf("删除pvc失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

type PvcListParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type PvcQueryByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type PvcQueryYamlParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type PvcCreateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Yaml        string `form:"yaml" json:"yaml" label:"Yaml"`
}

type PvcUpdateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Yaml        string `form:"yaml" json:"yaml" label:"Yaml"`
}

type PvcDeleteByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}
