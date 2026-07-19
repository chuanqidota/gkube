package workload

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"gkube/internal/k8s/params"
	"gkube/pkg/k8s"
	k8sCronjob "gkube/pkg/k8s/cronjob"
	"gkube/pkg/response"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	limit, continueToken := k8s.GetPaginationParams(c)
	if limit > 0 {
		cjList, err := k8sCronjob.ListCronJobs(client, query.Namespace, limit, continueToken)
		if err != nil {
			response.Fail(c, fmt.Sprintf("获取cronjob列表失败:%s", err.Error()))
			return
		}
		remaining := int64(0)
		if cjList.RemainingItemCount != nil {
			remaining = *cjList.RemainingItemCount
		}
		data := k8s.BuildPaginatedData(cjList.Items, cjList.Continue, remaining, limit)
		data.Total = len(cjList.Items)
		response.Success(c, "执行成功", data)
	} else {
		jobList, err := k8sCronjob.GetCronJobList(client, query.Namespace)
		if err != nil {
			response.Fail(c, fmt.Sprintf("获取cronjob列表失败:%s", err.Error()))
			return
		}
		response.Success(c, "执行成功", jobList)
	}
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

// GetCronJobEvents
//
//	@Description: 获取cronjob事件
//	@receiver cj
//	@param c
func (cj *cronjob) GetCronJobEvents(c *gin.Context) {
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
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=CronJob", query.Name),
	})
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取cronjob事件失败:%v", err.Error()))
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

// CronJobJobsList
//
//	@Description: 获取cronjob关联的job列表
//	@receiver cj
//	@param c
func (cj *cronjob) CronJobJobsList(c *gin.Context) {
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
	jobList, err := k8sCronjob.CronJobJobsList(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取cronjob关联job列表失败:%v", err.Error()))
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
	client, err := k8s.GetK8sClientByName(clusterName)
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
	client, err := k8s.GetK8sClientByName(clusterName)
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
	client, err := k8s.GetK8sClientByName(clusterName)
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
