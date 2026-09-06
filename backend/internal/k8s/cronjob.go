package k8s

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	k8sclient "gkube/pkg/k8s"
	k8sCronjob "gkube/pkg/k8s/cronjob"
	"gkube/pkg/logger"
	"gkube/pkg/response"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/client-go/kubernetes"
)

// ---------------------------------------------------------------------------
// 标准 handler（使用 wrapper）
// ---------------------------------------------------------------------------

var GetCronJobYaml = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sCronjob.GetCronJobYaml(client, namespace, name)
	},
	"执行成功", "获取cronjob yaml失败",
)

var CreateCronJob = CreateHandler(
	func(client *kubernetes.Clientset, namespace, yaml string) error {
		return k8sCronjob.CreateCronJob(client, namespace, yaml)
	},
	"执行成功", "创建cronjob失败",
)

// UpdateCronJob 更新 —— pkg 函数只传 namespace+yaml（无 name 参数）
func UpdateCronJob(c *gin.Context) {
	var body struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace"`
		Yaml        string `json:"yaml" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	if err := k8sCronjob.UpdateCronJob(client, body.Namespace, body.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "更新cronjob失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

var DeleteCronJobByName = DeleteHandler(
	func(client *kubernetes.Clientset, namespace, name string) error {
		return k8sCronjob.DeleteCronJobByName(client, namespace, name)
	},
	"执行成功", "删除cronjob失败",
)

var GetCronJobEvents = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		events, err := k8sCronjob.GetCronJobEvents(client, namespace, name)
		if err != nil {
			return nil, err
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
		return result, nil
	},
	"执行成功", "获取cronjob事件失败",
)

var CronJobJobsList = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sCronjob.CronJobJobsList(client, namespace, name)
	},
	"执行成功", "获取cronjob执行历史失败",
)

// ---------------------------------------------------------------------------
// 特殊 handler
// ---------------------------------------------------------------------------

// GetCronJobList 列表 —— 有 limit>0 分支 + augmentCronJob
func GetCronJobList(c *gin.Context) {
	var p ListParams
	if err := c.ShouldBind(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	selector, err := buildLabelSelector(p.LabelFilters)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, err.Error())
		return
	}
	if p.Limit > 0 {
		cjList, err := k8sCronjob.ListCronJobs(client, p.Namespace, p.Limit, p.Continue, selector)
		if err != nil {
			logger.Error(err.Error())
			response.FailWithStatus(c, http.StatusBadGateway, "获取cronjob列表失败")
			return
		}
		augmented := make([]map[string]any, 0, len(cjList.Items))
		for i := range cjList.Items {
			augmented = append(augmented, augmentCronJob(&cjList.Items[i]))
		}
		remaining := int64(0)
		if cjList.RemainingItemCount != nil {
			remaining = *cjList.RemainingItemCount
		}
		data := k8sclient.BuildPaginatedData(augmented, cjList.Continue, remaining, p.Limit)
		data.Total = len(cjList.Items)
		response.Success(c, "执行成功", data)
	} else {
		jobList, err := k8sCronjob.GetCronJobList(client, p.Namespace, selector)
		if err != nil {
			logger.Error(err.Error())
			response.FailWithStatus(c, http.StatusBadGateway, "获取cronjob列表失败")
			return
		}
		augmented := make([]map[string]any, 0, len(jobList))
		for i := range jobList {
			augmented = append(augmented, augmentCronJob(&jobList[i]))
		}
		response.Success(c, "执行成功", augmented)
	}
}

// GetCronJobByName 详情 —— NamespacedHandler + augmentCronJob
var GetCronJobByName = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		cj, err := k8sCronjob.GetCronJobByName(client, namespace, name)
		if err != nil {
			return nil, err
		}
		return augmentCronJob(cj), nil
	},
	"执行成功", "获取cronjob失败",
)

// SuspendCronJob 暂停 —— 原代码用 c.Query()，改为 ShouldBindQuery
func SuspendCronJob(c *gin.Context) {
	var p NamespacedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	if p.Name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	if p.ClusterName == "" {
		response.Fail(c, "clusterName参数不能为空")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	if err := k8sCronjob.SuspendCronJob(client, p.Namespace, p.Name); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("暂停CronJob失败:%s", err.Error()))
		return
	}
	response.Success(c, "暂停CronJob成功", nil)
}

// ResumeCronJob 恢复
func ResumeCronJob(c *gin.Context) {
	var p NamespacedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	if p.Name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	if p.ClusterName == "" {
		response.Fail(c, "clusterName参数不能为空")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	if err := k8sCronjob.ResumeCronJob(client, p.Namespace, p.Name); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("恢复CronJob失败:%s", err.Error()))
		return
	}
	response.Success(c, "恢复CronJob成功", nil)
}

// TriggerCronJob 触发
func TriggerCronJob(c *gin.Context) {
	var p NamespacedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	if p.Name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	if p.ClusterName == "" {
		response.Fail(c, "clusterName参数不能为空")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	job, err := k8sCronjob.TriggerCronJob(client, p.Namespace, p.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("触发CronJob失败:%s", err.Error()))
		return
	}
	response.Success(c, "触发CronJob成功", job)
}

// ---------------------------------------------------------------------------
// 内部 helper
// ---------------------------------------------------------------------------

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
