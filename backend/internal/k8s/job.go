package k8s

import (
	"net/http"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sJob "gkube/pkg/k8s/job"
	"gkube/pkg/logger"
	"gkube/pkg/response"
	"k8s.io/client-go/kubernetes"
)

// ---------------------------------------------------------------------------
// 标准 handler（使用 wrapper）
// ---------------------------------------------------------------------------

var GetJobByName = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sJob.GetJobByName(client, namespace, name)
	},
	"执行成功", "获取job失败",
)

var GetJobYaml = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sJob.GetJobYaml(client, namespace, name)
	},
	"执行成功", "获取job失败",
)

var DeleteJob = DeleteHandler(
	func(client *kubernetes.Clientset, namespace, name string) error {
		return k8sJob.DeleteJob(client, namespace, name)
	},
	"执行成功", "删除job失败",
)

var GetJobEvents = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		events, err := k8sJob.GetJobEvents(client, namespace, name)
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
	"执行成功", "获取job事件失败",
)

var JobPodList = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sJob.JobPodList(client, namespace, name)
	},
	"执行成功", "获取job pod列表失败",
)

// ---------------------------------------------------------------------------
// 特殊 handler
// ---------------------------------------------------------------------------

// GetJobList 列表 —— 有 limit>0 分支
func GetJobList(c *gin.Context) {
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
		jobList, err := k8sJob.ListJobs(client, p.Namespace, p.Limit, p.Continue, selector)
		if err != nil {
			logger.Error(err.Error())
			response.FailWithStatus(c, http.StatusBadGateway, "获取job列表失败")
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
		jobs, err := k8sJob.GetJobList(client, p.Namespace, selector)
		if err != nil {
			logger.Error(err.Error())
			response.FailWithStatus(c, http.StatusBadGateway, "获取job列表失败")
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
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	if err := k8sJob.CreateJob(client, body.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "创建job失败")
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
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	if err := k8sJob.UpdateJob(client, body.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "更新job失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

// RerunJob 重跑 —— 特殊操作
func RerunJob(c *gin.Context) {
	var p NamespacedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	if err := k8sJob.RerunJob(client, p.Namespace, p.Name); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "重跑job失败")
		return
	}
	response.Success(c, "执行成功", nil)
}
