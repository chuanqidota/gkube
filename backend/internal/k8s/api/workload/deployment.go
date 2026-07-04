package workload

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"gkube/internal/k8s/params"
	"gkube/pkg/k8s"
	k8sDeployment "gkube/pkg/k8s/deployment"
	"gkube/pkg/response"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type deployment struct {
}

var Deployment = new(deployment)

// GetDeploymentList
//
//	@Description: 获取deployment列表（支持分页）
//	@param c
func (dp *deployment) GetDeploymentList(c *gin.Context) {
	var query params.DeploymentListParams
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
		// Paginated mode
		deploymentList, err := k8sDeployment.ListDeployments(client, query.Namespace, limit, continueToken)
		if err != nil {
			response.Fail(c, fmt.Sprintf("获取deployment列表失败:%s", err.Error()))
			return
		}
		remaining := int64(0)
		if deploymentList.RemainingItemCount != nil {
			remaining = *deploymentList.RemainingItemCount
		}
		data := k8s.BuildPaginatedData(deploymentList.Items, deploymentList.Continue, remaining, limit)
		data.Total = len(deploymentList.Items)
		response.Success(c, "执行成功", data)
	} else {
		// Legacy mode: return all items
		deployments, err := k8sDeployment.GetDeploymentList(client, query.Namespace)
		if err != nil {
			response.Fail(c, fmt.Sprintf("获取deployment列表失败:%s", err.Error()))
			return
		}
		response.Success(c, "执行成功", deployments)
	}
}

// GetDeploymentDetail
//
//	@Description: 获取deployment详情
//	@receiver dp
//	@param c
func (dp *deployment) GetDeploymentDetail(c *gin.Context) {
	var query params.DeploymentDeleteParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	deploymentDetail, err := k8sDeployment.GetDeploymentDetail(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取deployment详情失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", deploymentDetail)
}

// GetDeploymentYaml
//
//	@Description: 获取deployment yaml
//	@receiver dp
//	@param c
func (dp *deployment) GetDeploymentYaml(c *gin.Context) {
	var query params.DeploymentDeleteParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	yamlContent, err := k8sDeployment.GetDeploymentYaml(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取deployment yaml失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yamlContent})
}

// RollbackDeployment
//
//	@Description: 回滚deployment
//	@receiver dp
//	@param c
func (dp *deployment) RollbackDeployment(c *gin.Context) {
	var body params.DeploymentRollbackParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	ok, err := k8sDeployment.RollbackDeployment(client, body.Namespace, body.Name, body.Revision)
	if err != nil {
		response.Fail(c, fmt.Sprintf("回滚deployment失败:%s", err.Error()))
		return
	}
	if !ok {
		response.Fail(c, "回滚deployment失败")
		return
	}
	response.Success(c, "执行成功", nil)
}

// CreateDeployment
//
//	@Description: 创建deployment
//	@receiver dp
//	@param c
func (dp *deployment) CreateDeployment(c *gin.Context) {
	var body params.DeploymentCreateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	ok, err := k8sDeployment.CreateDeployment(client, body.Namespace, body.Yaml)
	if err != nil {
		response.Fail(c, fmt.Sprintf("创建deployment失败:%s", err.Error()))
		return
	}
	if !ok {
		response.Fail(c, fmt.Sprintf("创建deployment失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdateDeployment
//
//	@Description: 更新deployment
//	@receiver dp
//	@param c
func (dp *deployment) UpdateDeployment(c *gin.Context) {
	var body params.DeploymentUpdateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	ok, err := k8sDeployment.UpdateDeployment(client, body.Namespace, body.Yaml)
	if err != nil {
		response.Fail(c, fmt.Sprintf("更新deployment失败:%s", err.Error()))
		return
	}
	if !ok {
		response.Fail(c, fmt.Sprintf("更新deployment失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteDeployment
//
//	@Description: 删除deployment
//	@receiver dp
//	@param c
func (dp *deployment) DeleteDeployment(c *gin.Context) {
	var body params.DeploymentDeleteParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	ok, err := k8sDeployment.DeleteDeployment(client, body.Namespace, body.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("删除deployment失败:%s", err.Error()))
		return
	}
	if !ok {
		response.Fail(c, fmt.Sprintf("删除deployment失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// ScaleDeployment
//
//	@Description: 扩所容deployment
//	@receiver dp
//	@param c
func (dp *deployment) ScaleDeployment(c *gin.Context) {
	var body params.DeploymentScaleParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	ok, err := k8sDeployment.ScaleDeployment(client, body.Namespace, body.Name, body.Replicas)
	if err != nil {
		response.Fail(c, fmt.Sprintf("缩容deployment失败:%s", err.Error()))
		return
	}
	if !ok {
		response.Fail(c, fmt.Sprintf("缩容deployment失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// RestartDeployment
//
//	@Description: 重启deployment
//	@receiver dp
//	@param c
func (dp *deployment) RestartDeployment(c *gin.Context) {
	var body params.DeploymentRestartParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	ok, err := k8sDeployment.RestartDeployment(client, body.Namespace, body.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("重启deployment失败:%s", err.Error()))
		return
	}
	if !ok {
		response.Fail(c, fmt.Sprintf("重启deployment失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeploymentPodList
//
//	@Description: 获取deployment pod列表
//	@receiver dp
//	@param c
func (dp *deployment) DeploymentPodList(c *gin.Context) {
	var query params.DeploymentPodParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	podList, err := k8sDeployment.DpPodList(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取deployment pod列表失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", podList)
}

// GetDeploymentReplicaSets
//
//	@Description: 获取Deployment关联的ReplicaSet列表
//	@receiver dp
//	@param c
func (dp *deployment) GetDeploymentReplicaSets(c *gin.Context) {
	var query params.DeploymentReplicaSetParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	rsList, err := k8sDeployment.GetDeploymentReplicaSets(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ReplicaSet列表失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", rsList)
}

// GetDeploymentEvents
//
//	@Description: 获取deployment事件
//	@receiver dp
//	@param c
func (dp *deployment) GetDeploymentEvents(c *gin.Context) {
	var query params.DeploymentDeleteParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	events, err := client.CoreV1().Events(query.Namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Deployment", query.Name),
	})
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取deployment事件失败:%s", err.Error()))
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
