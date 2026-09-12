package k8s

import (
	"context"

	"github.com/gin-gonic/gin"
	apperr "gkube/pkg/errors"
	k8sclient "gkube/pkg/k8s"
	k8sJob "gkube/pkg/k8s/job"
	"gkube/pkg/response"
	"k8s.io/client-go/kubernetes"
)

// ---------------------------------------------------------------------------
// 标准 handler（使用 wrapper）
// ---------------------------------------------------------------------------

var GetJobByName = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sJob.GetJobByName(ctx, client, namespace, name)
	},
	"执行成功",
)

var GetJobYaml = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sJob.GetJobYaml(ctx, client, namespace, name)
	},
	"执行成功",
)

var DeleteJob = DeleteHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error {
		return k8sJob.DeleteJob(ctx, client, namespace, name)
	},
	"执行成功",
)

var GetJobEvents = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		events, err := k8sJob.GetJobEvents(ctx, client, namespace, name)
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

var JobPodList = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sJob.JobPodList(ctx, client, namespace, name)
	},
	"执行成功",
)

// ---------------------------------------------------------------------------
// 特殊 handler
// ---------------------------------------------------------------------------

// GetJobList 列表 —— 有 limit>0 分支
func GetJobList(c *gin.Context) {
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
		jobList, err := k8sJob.ListJobs(c.Request.Context(), client, p.Namespace, p.Limit, p.Continue, selector)
		if err != nil {
			response.FailWithError(c, ensureAppError(err))
			return
		}
		remaining := int64(0)
		if jobList.RemainingItemCount != nil {
			remaining = *jobList.RemainingItemCount
		}
		data := k8sclient.BuildPaginatedData(jobList.Items, jobList.Continue, remaining, p.Limit)
		data.Total = len(jobList.Items)
		response.Success(c, "执行成功", data)
	} else {
		jobs, err := k8sJob.GetJobList(c.Request.Context(), client, p.Namespace, selector)
		if err != nil {
			response.FailWithError(c, ensureAppError(err))
			return
		}
		response.Success(c, "执行成功", jobs)
	}
}

// CreateJob 创建 —— pkg 函数只传 yaml（无 namespace 参数）
func CreateJob(c *gin.Context) {
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
	if err := k8sJob.CreateJob(c.Request.Context(), client, body.Yaml); err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdateJob 更新 —— pkg 函数只传 yaml（无 namespace 参数）
func UpdateJob(c *gin.Context) {
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
	if err := k8sJob.UpdateJob(c.Request.Context(), client, body.Yaml); err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", nil)
}

// RerunJob 重跑
func RerunJob(c *gin.Context) {
	var p NamespacedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	if err := k8sJob.RerunJob(c.Request.Context(), client, p.Namespace, p.Name); err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", nil)
}
