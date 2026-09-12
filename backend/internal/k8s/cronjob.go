package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	apperr "gkube/pkg/errors"
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
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sCronjob.GetCronJobYaml(ctx, client, namespace, name)
	},
	"执行成功",
)

var CreateCronJob = CreateHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, yaml string) error {
		return k8sCronjob.CreateCronJob(ctx, client, namespace, yaml)
	},
	"执行成功",
)

var DeleteCronJobByName = DeleteHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error {
		return k8sCronjob.DeleteCronJobByName(ctx, client, namespace, name)
	},
	"执行成功",
)

var GetCronJobEvents = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		events, err := k8sCronjob.GetCronJobEvents(ctx, client, namespace, name)
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
	"执行成功",
)

var CronJobJobsList = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sCronjob.CronJobJobsList(ctx, client, namespace, name)
	},
	"执行成功",
)

// ---------------------------------------------------------------------------
// 特殊 handler
// ---------------------------------------------------------------------------

// UpdateCronJob 更新 —— pkg 函数只传 namespace+yaml（无 name 参数）
func UpdateCronJob(c *gin.Context) {
	var body struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace"`
		Yaml        string `json:"yaml" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	if err := k8sCronjob.UpdateCronJob(c.Request.Context(), client, body.Namespace, body.Yaml); err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", nil)
}

// GetCronJobList 列表 —— 有 limit>0 分支 + augmentCronJob
func GetCronJobList(c *gin.Context) {
	var p ListParams
	if err := c.ShouldBind(&p); err != nil {
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	selector, err := buildLabelSelector(p.LabelFilters)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	if p.Limit > 0 {
		cjList, err := k8sCronjob.ListCronJobs(c.Request.Context(), client, p.Namespace, p.Limit, p.Continue, selector)
		if err != nil {
			response.FailWithError(c, ensureAppError(err))
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
		jobList, err := k8sCronjob.GetCronJobList(c.Request.Context(), client, p.Namespace, selector)
		if err != nil {
			response.FailWithError(c, ensureAppError(err))
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
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		cj, err := k8sCronjob.GetCronJobByName(ctx, client, namespace, name)
		if err != nil {
			return nil, err
		}
		return augmentCronJob(cj), nil
	},
	"执行成功",
)

// SuspendCronJob 暂停
func SuspendCronJob(c *gin.Context) {
	var p NamespacedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}
	if p.Name == "" {
		response.FailWithError(c, apperr.BadRequest("name参数不能为空", nil))
		return
	}
	if p.ClusterName == "" {
		response.FailWithError(c, apperr.BadRequest("clusterName参数不能为空", nil))
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	if err := k8sCronjob.SuspendCronJob(c.Request.Context(), client, p.Namespace, p.Name); err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "暂停CronJob成功", nil)
}

// ResumeCronJob 恢复
func ResumeCronJob(c *gin.Context) {
	var p NamespacedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}
	if p.Name == "" {
		response.FailWithError(c, apperr.BadRequest("name参数不能为空", nil))
		return
	}
	if p.ClusterName == "" {
		response.FailWithError(c, apperr.BadRequest("clusterName参数不能为空", nil))
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	if err := k8sCronjob.ResumeCronJob(c.Request.Context(), client, p.Namespace, p.Name); err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "恢复CronJob成功", nil)
}

// TriggerCronJob 触发
func TriggerCronJob(c *gin.Context) {
	var p NamespacedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}
	if p.Name == "" {
		response.FailWithError(c, apperr.BadRequest("name参数不能为空", nil))
		return
	}
	if p.ClusterName == "" {
		response.FailWithError(c, apperr.BadRequest("clusterName参数不能为空", nil))
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	job, err := k8sCronjob.TriggerCronJob(c.Request.Context(), client, p.Namespace, p.Name)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
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
