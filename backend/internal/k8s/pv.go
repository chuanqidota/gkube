package k8s

import (
	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sPv "gkube/pkg/k8s/pv"
	"gkube/pkg/logger"
	"gkube/pkg/response"
)

type pv struct {
}

var Pv = new(pv)

// GetPVList
//
//	@Description: 获取pv列表
//	@receiver p
//	@param c
func (p *pv) GetPVList(c *gin.Context) {
	var query PvListParams
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
	pvList, err := k8sPv.GetPVList(client)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取pv列表失败")
		return
	}
	response.Success(c, "执行成功", pvList)
}

// GetPVByName
//
//	@Description: 根据名称获取pv详情
//	@receiver p
//	@param c
func (p *pv) GetPVByName(c *gin.Context) {
	var query PvQueryByNameParams
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
	pv, err := k8sPv.GetPVByName(client, query.Name)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取pv详情失败")
		return
	}
	response.Success(c, "执行成功", pv)
}

// GetPVYaml
//
//	@Description: 获取pv的yaml
//	@receiver p
//	@param c
func (p *pv) GetPVYaml(c *gin.Context) {
	var query PvQueryByNameParams
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
	yaml, err := k8sPv.GetPVYaml(client, query.Name)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取pv yaml失败")
		return
	}
	response.Success(c, "执行成功", yaml)
}

// CreatePV
//
//	@Description: 创建pv
//	@receiver p
//	@param c
func (p *pv) CreatePV(c *gin.Context) {
	var body PvCreateParams
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
	err = k8sPv.CreatePV(client, body.Yaml)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "创建pv失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdatePV
//
//	@Description: 更新pv
//	@receiver p
//	@param c
func (p *pv) UpdatePV(c *gin.Context) {
	var body PvUpdateParams
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
	err = k8sPv.UpdatePV(client, body.Yaml)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "更新pv失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeletePVByName
//
//	@Description: 根据名称删除pv
//	@receiver p
//	@param c
func (p *pv) DeletePVByName(c *gin.Context) {
	var body PvQueryByNameParams
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
	err = k8sPv.DeletePVByName(client, body.Name)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "删除pv失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

type PvListParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
}

type PvQueryByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
}

type PvCreateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Yaml        string `form:"yaml" json:"yaml" binding:"required" label:"Yaml"`
}

type PvUpdateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Yaml        string `form:"yaml" json:"yaml" binding:"required" label:"Yaml"`
}
