package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gkube/app/k8s/params"
	"gkube/pkg/k8s"
	k8sCronjob "gkube/pkg/k8s/cronjob"
	"gkube/pkg/response"
)

type cronjob struct {
}

var Cronjob = new(cronjob)

// GetCronJobList
//
//	@Description: 获取cronjob列表
//	@receiver cj
//	@param c
func (cj *cronjob) GetCronJobList(c *gin.Context) {
	var query params.CronjobListParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}

	jobList, err := k8sCronjob.GetCronJobList(client, query.Namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取cronjob列表失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", jobList)
}

// GetCronJobByName
//
//	@Description: 根据名称获取cronjob
//	@receiver cj
//	@param c
func (cj *cronjob) GetCronJobByName(c *gin.Context) {
	var query params.JobQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	job, err := k8sCronjob.GetCronJobByName(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取cronjob失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", job)
}

// GetCronJobYaml
//
//	@Description: 根据名称获取cronjob的yaml
//	@receiver cj
//	@param c
func (cj *cronjob) GetCronJobYaml(c *gin.Context) {
	var query params.JobQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}

	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	yaml, err := k8sCronjob.GetCronJobYaml(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取cronjob yaml失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", yaml)
}

// GetCronJobByLabel
//
//	@Description: 根据标签获取cronjob
//	@receiver cj
//	@param c
func (cj *cronjob) GetCronJobByLabel(c *gin.Context) {
	var body params.CronJobQueryByLabelParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	jobList, err := k8sCronjob.GetCronJobByLabel(client, body.Namespace, body.LabelMap)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取cronjob列表失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", jobList)
}

// GetCronJobByField
//
//	@Description: 根据字段获取cronjob
//	@receiver cj
//	@param c
func (cj *cronjob) GetCronJobByField(c *gin.Context) {
	var body params.CronJobQueryByFieldParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	jobList, err := k8sCronjob.GetCronJobByField(client, body.Namespace, body.FieldMap)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取cronjob列表失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", jobList)
}

// CreateCronJob
//
//	@Description: 创建cronjob
//	@receiver cj
//	@param c
func (cj *cronjob) CreateCronJob(c *gin.Context) {
	var body params.CronJobCreateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	err = k8sCronjob.CreateCronJob(client, body.Namespace, body.CronJobYaml)
	if err != nil {
		response.Fail(c, fmt.Sprintf("创建cronjob失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdateCronJob
//
//	@Description: 更新cronjob
//	@receiver cj
//	@param c
func (cj *cronjob) UpdateCronJob(c *gin.Context) {
	var body params.CronJobUpdateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	err = k8sCronjob.UpdateCronJob(client, body.Namespace, body.CronJobYaml)
	if err != nil {
		response.Fail(c, fmt.Sprintf("更新cronjob失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteCronJobByName
//
//	@Description: 删除cronjob根据名称
//	@receiver cj
//	@param c
func (cj *cronjob) DeleteCronJobByName(c *gin.Context) {
	var body params.CronJobDeleteByNameParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}

	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	err = k8sCronjob.DeleteCronJobByName(client, body.Namespace, body.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("删除cronjob失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteCronJobByLabel
//
//	@Description: 删除cronjob根据标签
//	@receiver cj
//	@param c
func (cj *cronjob) DeleteCronJobByLabel(c *gin.Context) {
	var body params.CronJobDeleteByLabelParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	err = k8sCronjob.DeleteCronJobByLabel(client, body.Namespace, body.LabelMap)
	if err != nil {
		response.Fail(c, fmt.Sprintf("删除cronjob失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteCronJobByField
//
//	@Description: 删除cronjob根据字段
//	@receiver cj
//	@param c
func (cj *cronjob) DeleteCronJobByField(c *gin.Context) {
	var body params.CronJobDeleteByFieldParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
	}

	err = k8sCronjob.DeleteCronJobByField(client, body.Namespace, body.FieldMap)
	if err != nil {
		response.Fail(c, fmt.Sprintf("删除cronjob失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}
