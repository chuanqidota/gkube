package workload

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"gkube/internal/k8s/params"
	"gkube/pkg/k8s"
	k8sJob "gkube/pkg/k8s/job"
	"gkube/pkg/response"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	var query params.JobListParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%v", err.Error()))
		return
	}

	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	limit, continueToken := k8s.GetPaginationParams(c)
	if limit > 0 {
		jobList, err := k8sJob.ListJobs(client, query.Namespace, limit, continueToken)
		if err != nil {
			response.Fail(c, fmt.Sprintf("获取job列表失败:%v", err.Error()))
			return
		}
		remaining := int64(0)
		if jobList.RemainingItemCount != nil {
			remaining = *jobList.RemainingItemCount
		}
		data := k8s.BuildPaginatedData(jobList.Items, jobList.Continue, remaining, limit)
		data.Total = len(jobList.Items)
		response.Success(c, "执行成功", data)
	} else {
		jobs, err := k8sJob.GetJobList(client, query.Namespace)
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
	var query params.JobQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
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

// GetJobByFiled
//
//	@Description: 根据字段查询job
//	@receiver j
//	@param c
func (j *job) GetJobByFiled(c *gin.Context) {
	var body params.JobQueryByFiledParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	jobs, err := k8sJob.GetJobByFiled(client, body.Namespace, body.FieldMap)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取job失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", jobs)
}

// GetJobByLabel
//
//	@Description: 根据标签查询job
//	@receiver j
//	@param c
func (j *job) GetJobByLabel(c *gin.Context) {
	var body params.JobQueryByLabelParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	jobs, err := k8sJob.GetJobByLabel(client, body.Namespace, body.LabelMap)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取job失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", jobs)
}

// GetJobYaml
//
//	@Description: 获取job的yaml
//	@receiver j
//	@param c
func (j *job) GetJobYaml(c *gin.Context) {
	var query params.JobQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
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
	var body params.JobCreateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
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
	var body params.JobUpdateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
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
	var body params.JobDeleteByNameParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
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

// DeleteJobByField
//
//	@Description: 根据字段删除job
//	@receiver j
//	@param c
func (j *job) DeleteJobByField(c *gin.Context) {
	var body params.JobDeleteByFieldParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sJob.DeleteJobByField(client, body.Namespace, body.FieldMap); err != nil {
		response.Fail(c, fmt.Sprintf("删除job失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteJobByLabel
//
//	@Description: 根据标签删除job
//	@receiver j
//	@param c
func (j *job) DeleteJobByLabel(c *gin.Context) {
	var body params.JobDeleteByLabelParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sJob.DeleteJobByLabel(client, body.Namespace, body.LabelMap); err != nil {
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
	events, err := client.CoreV1().Events(query.Namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Job", query.Name),
	})
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取job事件失败:%v", err.Error()))
		return
	}
	var result []map[string]any
	for _, event := range events.Items {
		lastSeen := ""
		if !event.LastTimestamp.IsZero() {
			lastSeen = event.LastTimestamp.Time.Format("2006-01-02 15:04:05")
		}
		result = append(result, map[string]any{
			"type":      event.Type,
			"reason":    event.Reason,
			"message":   event.Message,
			"last_seen": lastSeen,
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
	podList, err := k8sJob.JobPodList(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取job pod列表失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", podList)
}
