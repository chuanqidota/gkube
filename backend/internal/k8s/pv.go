package k8s

import (
	"fmt"
	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sPv "gkube/pkg/k8s/pv"
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
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	pvList, err := k8sPv.GetPVList(client)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取pv列表失败:%v", err.Error()))
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
	var query PvQueryByName
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	pv, err := k8sPv.GetPVByName(client, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取pv列表失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", pv)
}

// GetPVYaml
//
//	@Description: 获取pv详情
//	@receiver p
//	@param c
func (p *pv) GetPVYaml(c *gin.Context) {
	var query PvQueryByName
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	yaml, err := k8sPv.GetPVYaml(client, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取pv列表失败:%v", err.Error()))
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
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	err = k8sPv.CreatePV(client, body.Yaml)
	if err != nil {
		response.Fail(c, fmt.Sprintf("创建pv失败:%v", err.Error()))
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
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	err = k8sPv.UpdatePV(client, body.Yaml)
	if err != nil {
		response.Fail(c, fmt.Sprintf("更新pv失败:%v", err.Error()))
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
	var body PvQueryByName
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	err = k8sPv.DeletePVByName(client, body.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("删除pv失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

type PvListParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
}

type PvQueryByName struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
}

type PvCreateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Yaml        string `form:"yaml" json:"yaml" label:"Yaml"`
}

type PvUpdateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Yaml        string `form:"yaml" json:"yaml" label:"Yaml"`
}
