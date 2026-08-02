package k8s

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
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
	var query DeploymentListParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}

	limit, continueToken := k8sclient.GetPaginationParams(c)
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
		data := k8sclient.BuildPaginatedData(deploymentList.Items, deploymentList.Continue, remaining, limit)
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
	var query DeploymentDeleteParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
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
	var query DeploymentDeleteParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
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
	var body DeploymentRollbackParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
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
	var body DeploymentCreateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
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
	var body DeploymentUpdateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
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
	var body DeploymentDeleteParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
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
	var body DeploymentScaleParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
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
	var body DeploymentRestartParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
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

// UpdateDeploymentImage
//
//	@Description: 更新deployment容器镜像
//	@receiver dp
//	@param c
func (dp *deployment) UpdateDeploymentImage(c *gin.Context) {
	var body DeploymentImageUpdateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	ok, err := k8sDeployment.UpdateDeploymentImage(client, body.Namespace, body.Name, body.ContainerName, body.Image)
	if err != nil {
		response.Fail(c, fmt.Sprintf("更新deployment镜像失败:%s", err.Error()))
		return
	}
	if !ok {
		response.Fail(c, fmt.Sprintf("更新deployment镜像失败:%s", err.Error()))
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
	var query DeploymentPodParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
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
	var query DeploymentReplicaSetParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
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
	var query DeploymentDeleteParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
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

type DeploymentListParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type DeploymentCreateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Yaml        string `form:"yaml" json:"yaml" label:"yaml"`
}

type DeploymentUpdateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Yaml        string `form:"yaml" json:"yaml" label:"yaml"`
}

type DeploymentDeleteParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
}

type DeploymentScaleParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Replicas    int32  `form:"replicas" json:"replicas" label:"副本数"`
}

type DeploymentRestartParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
}

type DeploymentPodParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
}

type DeploymentRollbackParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Revision    int64  `form:"revision" json:"revision" label:"回滚版本"`
}

// DeploymentReplicaSetParams 获取 Deployment 关联的 ReplicaSet 列表
type DeploymentReplicaSetParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" label:"Deployment名称"`
}

// DeploymentImageUpdateParams 更新 Deployment 容器镜像
type DeploymentImageUpdateParams struct {
	ClusterName   string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace     string `form:"namespace" json:"namespace" label:"命名空间"`
	Name          string `form:"name" json:"name" label:"Deployment名称"`
	ContainerName string `form:"containerName" json:"containerName" label:"容器名称"`
	Image         string `form:"image" json:"image" label:"镜像"`
}
