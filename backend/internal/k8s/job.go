package k8s

import (
	"fmt"
	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sJob "gkube/pkg/k8s/job"
	k8sLabels "gkube/pkg/k8s/labels"
	"gkube/pkg/response"
)

type job struct {
}

var Job = new(job)

// GetJobList
//
//	@Description: 获取job列表
//	@receiver j
//	@param c
func (j *job) GetJobList(c *gin.Context) {
	var query JobListParams
	if err := c.ShouldBind(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%v", err.Error()))
		return
	}

	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	// 构建 label selector
	var selector string
	if len(query.LabelFilters) > 0 {
		selector, err = k8sLabels.BuildLabelSelector(query.LabelFilters)
		if err != nil {
			response.Fail(c, err.Error())
			return
		}
	}


	limit := query.Limit
	continueToken := query.Continue
	if limit > 0 {
		jobList, err := k8sJob.ListJobs(client, query.Namespace, limit, continueToken, selector)
		if err != nil {
			response.Fail(c, fmt.Sprintf("获取job列表失败:%v", err.Error()))
			return
		}
		remaining := int64(0)
		if jobList.RemainingItemCount != nil {
			remaining = *jobList.RemainingItemCount
		}
		data := k8sclient.BuildPaginatedData(jobList.Items, jobList.Continue, remaining, limit)
		data.Total = len(jobList.Items)
		response.Success(c, "执行成功", data)
	} else {
		jobs, err := k8sJob.GetJobList(client, query.Namespace, selector)
		if err != nil {
			response.Fail(c, fmt.Sprintf("获取job列表失败:%v", err.Error()))
			return
		}
		response.Success(c, "执行成功", jobs)
	}
}

// GetJobByName
//
//	@Description: 根据名称查询job
//	@receiver j
//	@param c
func (j *job) GetJobByName(c *gin.Context) {
	var query JobQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	job, err := k8sJob.GetJobByName(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取job失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", job)
}

// GetJobYaml
//
//	@Description: 获取job的yaml
//	@receiver j
//	@param c
func (j *job) GetJobYaml(c *gin.Context) {
	var query JobQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	jobYaml, err := k8sJob.GetJobYaml(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取job失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", jobYaml)
}

// CreateJob
//
//	@Description: 创建job
//	@receiver j
//	@param c
func (j *job) CreateJob(c *gin.Context) {
	var body JobCreateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	if err := k8sJob.CreateJob(client, body.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("创建job失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdateJob
//
//	@Description: 更新job
//	@receiver j
//	@param c
func (j *job) UpdateJob(c *gin.Context) {
	var body JobUpdateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sJob.UpdateJob(client, body.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("更新job失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteJob
//
//	@Description: 删除job
//	@receiver j
//	@param c
func (j *job) DeleteJob(c *gin.Context) {
	var body JobDeleteByNameParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	if err := k8sJob.DeleteJob(client, body.Namespace, body.Name); err != nil {
		response.Fail(c, fmt.Sprintf("删除job失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)

}

// GetJobEvents
//
//	@Description: 获取job事件
//	@receiver j
//	@param c
func (j *job) GetJobEvents(c *gin.Context) {
	var query JobQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	events, err := k8sJob.GetJobEvents(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取job事件失败:%v", err.Error()))
		return
	}
	var result []map[string]any
	for _, event := range events {
		result = append(result, map[string]any{
			"type":      event.Type,
			"reason":    event.Reason,
			"message":   event.Message,
			"last_seen": event.LastTimestamp,
		})
	}
	response.Success(c, "执行成功", result)
}

// JobPodList
//
//	@Description: 获取job关联的pod列表
//	@receiver j
//	@param c
func (j *job) JobPodList(c *gin.Context) {
	var query JobQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	podList, err := k8sJob.JobPodList(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取job pod列表失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", podList)
}

type JobListParams struct {
	ClusterName  string                  `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace    string                  `form:"namespace" json:"namespace" label:"命名空间"`
	Limit        int64                   `form:"limit" json:"limit" label:"每页条数"`
	Continue     string                  `form:"continue" json:"continue" label:"分页标记"`
	LabelFilters []k8sLabels.LabelFilter `json:"labelFilters" form:"labelFilters" label:"标签过滤"`
}

type JobQueryByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type JobCreateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Yaml        string `form:"yaml" json:"yaml" label:"Yaml"`
}

type JobUpdateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Yaml        string `form:"yaml" json:"yaml" label:"Yaml"`
}

type JobDeleteByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

// RerunJob
//
//	@Description: 一键重跑job
//	@receiver j
//	@param c
func (j *job) RerunJob(c *gin.Context) {
	var query JobQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%v", err.Error()))
		return
	}

	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	if err := k8sJob.RerunJob(client, query.Namespace, query.Name); err != nil {
		response.Fail(c, fmt.Sprintf("重跑job失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}
