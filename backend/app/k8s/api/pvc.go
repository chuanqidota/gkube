package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gkube/app/k8s/params"
	"gkube/pkg/k8s"
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
	var query params.PvcListParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, err.Error())
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
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
	var query params.PvcQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, err.Error())
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
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

// GetPVCByLabel
//
//	@Description: 根据标签获取pvc
//	@receiver p
//	@param c
func (p *pvc) GetPVCByLabel(c *gin.Context) {
	var body params.PvcQueryByLabelParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	pvcList, err := k8sPvc.GetPVCByLabel(client, body.Namespace, body.LabelMap)
	if err != nil {
		response.Fail(c, fmt.Sprintf("查询pvc失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", pvcList)
}

// GetPVCByField
//
//	@Description: 根据字段获取pvc
//	@receiver p
//	@param c
func (p *pvc) GetPVCByField(c *gin.Context) {
	var body params.PvcQueryByFieldParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	pvcList, err := k8sPvc.GetPVCByField(client, body.Namespace, body.FieldMap)
	if err != nil {
		response.Fail(c, fmt.Sprintf("查询pvc失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", pvcList)
}

// GetPVCYaml
//
//	@Description: 获取pvc的yaml
//	@receiver p
//	@param c
func (p *pvc) GetPVCYaml(c *gin.Context) {
	var query params.PvcQueryYamlParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
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
	var body params.PvcCreateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
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

// DeletePVCByName
//
//	@Description: 根据名称删除pvc
//	@receiver p
//	@param c
func (p *pvc) DeletePVCByName(c *gin.Context) {
	var body params.PvcDeleteByNameParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
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

// DeletePVCByLabel
//
//	@Description: 根据标签删除pvc
//	@receiver p
//	@param c
func (p *pvc) DeletePVCByLabel(c *gin.Context) {
	var body params.PvcDeleteByLabelParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sPvc.DeletePVCByLabel(client, body.Namespace, body.LabelMap); err != nil {
		response.Fail(c, fmt.Sprintf("删除pvc失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeletePVCByField
//
//	@Description: 根据字段删除pvc
//	@receiver p
//	@param c
func (p *pvc) DeletePVCByField(c *gin.Context) {
	var body params.PvcDeleteByFieldParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sPvc.DeletePVCByField(client, body.Namespace, body.FieldMap); err != nil {
		response.Fail(c, fmt.Sprintf("删除pvc失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}
