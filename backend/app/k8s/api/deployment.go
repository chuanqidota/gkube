package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"gkube/app/k8s/params"
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
//	@Description: 获取deployment列表
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
	deployments, err := k8sDeployment.GetDeploymentList(client, query.Namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取deployment列表失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", deployments)
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

// GetDeploymentByField
//
//	@Description: 根据字段查询deployment列表
//	@receiver dp
//	@param c
func (dp *deployment) GetDeploymentByField(c *gin.Context) {
	var body params.DeploymentQueryByFieldParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	deployments, err := k8sDeployment.GetDeploymentByFiled(client, body.Namespace, body.FieldMap)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取deployment列表失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", deployments)
}

// GetDeploymentByLabel
//
//	@Description: 根据标签查询deployment列表
//	@receiver dp
//	@param c
func (dp *deployment) GetDeploymentByLabel(c *gin.Context) {
	var body params.DeploymentQueryByLabelParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	deployments, err := k8sDeployment.GetDeploymentByLabel(client, body.Namespace, body.LabelMap)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取deployment列表失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", deployments)
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

// DeleteDeploymentByField
//
//	@Description: 根据字段删除deployment
//	@receiver dp
//	@param c
func (dp *deployment) DeleteDeploymentByField(c *gin.Context) {
	var body params.DeploymentDeleteByFieldParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	ok, err := k8sDeployment.DeleteDeploymentByField(client, body.Namespace, body.FieldMap)
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

// DeleteDeploymentByLabel
//
//	@Description: 根据标签删除deployment
//	@receiver dp
//	@param c
func (dp *deployment) DeleteDeploymentByLabel(c *gin.Context) {
	var body params.DeploymentDeleteByLabelParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	ok, err := k8sDeployment.DeleteDeploymentByLabel(client, body.Namespace, body.LabelMap)
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
