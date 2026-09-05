package k8s

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	k8sclient "gkube/pkg/k8s"
	k8sCronjob "gkube/pkg/k8s/cronjob"
	k8sLabels "gkube/pkg/k8s/labels"
	"gkube/pkg/logger"
	"gkube/pkg/response"
	batchv1 "k8s.io/api/batch/v1"
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
	var query CronjobListParams
	if err := c.ShouldBind(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
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
		cjList, err := k8sCronjob.ListCronJobs(client, query.Namespace, limit, continueToken, selector)
		if err != nil {
			response.Fail(c, fmt.Sprintf("获取cronjob列表失败:%s", err.Error()))
			return
		}
		remaining := int64(0)
		if cjList.RemainingItemCount != nil {
			remaining = *cjList.RemainingItemCount
		}
		augmented := make([]map[string]any, 0, len(cjList.Items))
		for i := range cjList.Items {
			augmented = append(augmented, augmentCronJob(&cjList.Items[i]))
		}
		data := k8sclient.BuildPaginatedData(augmented, cjList.Continue, remaining, limit)
		data.Total = len(cjList.Items)
		response.Success(c, "执行成功", data)
	} else {
		jobList, err := k8sCronjob.GetCronJobList(client, query.Namespace, selector)
		if err != nil {
			response.Fail(c, fmt.Sprintf("获取cronjob列表失败:%s", err.Error()))
			return
		}
		augmented := make([]map[string]any, 0, len(jobList))
		for i := range jobList {
			augmented = append(augmented, augmentCronJob(&jobList[i]))
		}
		response.Success(c, "执行成功", augmented)
	}
}

// GetCronJobByName
//
//	@Description: 根据名称获取cronjob
//	@receiver cj
//	@param c
func (cj *cronjob) GetCronJobByName(c *gin.Context) {
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
	job, err := k8sCronjob.GetCronJobByName(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取cronjob失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", augmentCronJob(job))
}

// GetCronJobYaml
//
//	@Description: 根据名称获取cronjob的yaml
//	@receiver cj
//	@param c
func (cj *cronjob) GetCronJobYaml(c *gin.Context) {
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
	yaml, err := k8sCronjob.GetCronJobYaml(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取cronjob yaml失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", yaml)
}

// CreateCronJob
//
//	@Description: 创建cronjob
//	@receiver cj
//	@param c
func (cj *cronjob) CreateCronJob(c *gin.Context) {
	var body CronJobCreateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	err = k8sCronjob.CreateCronJob(client, body.Namespace, body.Yaml)
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
	var body CronJobUpdateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	err = k8sCronjob.UpdateCronJob(client, body.Namespace, body.Yaml)
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
	var body CronJobDeleteByNameParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}

	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
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

// GetCronJobEvents
//
//	@Description: 获取cronjob事件
//	@receiver cj
//	@param c
func (cj *cronjob) GetCronJobEvents(c *gin.Context) {
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
	events, err := k8sCronjob.GetCronJobEvents(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取cronjob事件失败:%v", err.Error()))
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

// CronJobJobsList
//
//	@Description: 获取cronjob执行历史Job列表
//	@receiver cj
//	@param c
func (cj *cronjob) CronJobJobsList(c *gin.Context) {
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
	jobList, err := k8sCronjob.CronJobJobsList(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取cronjob执行历史失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", jobList)
}

func (cj *cronjob) SuspendCronJob(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	if name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	if clusterName == "" {
		response.Fail(c, "clusterName参数不能为空")
		return
	}
	client, err := k8sclient.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sCronjob.SuspendCronJob(client, namespace, name); err != nil {
		response.Fail(c, fmt.Sprintf("暂停CronJob失败:%s", err.Error()))
		return
	}
	response.Success(c, "暂停CronJob成功", nil)
}

func (cj *cronjob) ResumeCronJob(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	if name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	if clusterName == "" {
		response.Fail(c, "clusterName参数不能为空")
		return
	}
	client, err := k8sclient.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sCronjob.ResumeCronJob(client, namespace, name); err != nil {
		response.Fail(c, fmt.Sprintf("恢复CronJob失败:%s", err.Error()))
		return
	}
	response.Success(c, "恢复CronJob成功", nil)
}

func (cj *cronjob) TriggerCronJob(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	if name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	if clusterName == "" {
		response.Fail(c, "clusterName参数不能为空")
		return
	}
	client, err := k8sclient.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	job, err := k8sCronjob.TriggerCronJob(client, namespace, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("触发CronJob失败:%s", err.Error()))
		return
	}
	response.Success(c, "触发CronJob成功", job)
}

type CronjobListParams struct {
	ClusterName  string                  `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace    string                  `form:"namespace" json:"namespace" label:"命名空间"`
	Limit        int64                   `form:"limit" json:"limit" label:"每页条数"`
	Continue     string                  `form:"continue" json:"continue" label:"分页标记"`
	LabelFilters []k8sLabels.LabelFilter `json:"labelFilters" form:"labelFilters" label:"标签过滤"`
}

type CronJobCreateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Yaml        string `form:"yaml" json:"yaml" label:"Yaml"`
}

type CronJobUpdateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Yaml        string `form:"yaml" json:"yaml" label:"Yaml"`
}

type CronJobDeleteByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
}

// cronParser is a shared parser for cron expressions (thread-safe)
var cronParser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

// computeNextScheduleTime computes the next execution time for a CronJob
func computeNextScheduleTime(cj *batchv1.CronJob) string {
	if cj.Spec.Suspend != nil && *cj.Spec.Suspend {
		return ""
	}
	schedule, err := cronParser.Parse(cj.Spec.Schedule)
	if err != nil {
		logger.Error(fmt.Sprintf("解析cron表达式失败 [%s/%s]: %s", cj.Namespace, cj.Name, err.Error()))
		return ""
	}
	var base time.Time
	if cj.Status.LastScheduleTime != nil {
		base = cj.Status.LastScheduleTime.Time
	} else {
		base = time.Now()
	}
	next := schedule.Next(base)
	return next.Format("2006-01-02 15:04:05")
}

// augmentCronJob adds computed fields (nextScheduleTime) to a CronJob
func augmentCronJob(cj *batchv1.CronJob) map[string]any {
	return map[string]any{
		"metadata":         cj.ObjectMeta,
		"spec":             cj.Spec,
		"status":           cj.Status,
		"nextScheduleTime": computeNextScheduleTime(cj),
	}
}
