package k8s

import (
	"net/http"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sDeployment "gkube/pkg/k8s/deployment"
	"gkube/pkg/logger"
	"gkube/pkg/response"
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
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}

	limit, continueToken := k8sclient.GetPaginationParams(c)
	deploymentList, err := k8sDeployment.ListDeployments(client, query.Namespace, limit, continueToken)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取deployment列表失败")
		return
	}
	remaining := int64(0)
	if deploymentList.RemainingItemCount != nil {
		remaining = *deploymentList.RemainingItemCount
	}
	data := k8sclient.BuildPaginatedData(deploymentList.Items, deploymentList.Continue, remaining, limit)
	// Total = 当前页条数 + 剩余条数,接近集群内真实总数
	data.Total = len(deploymentList.Items) + int(remaining)
	response.Success(c, "获取deployment列表成功", data)
}

// GetDeploymentDetail
//
//	@Description: 获取deployment详情
//	@receiver dp
//	@param c
func (dp *deployment) GetDeploymentDetail(c *gin.Context) {
	var query DeploymentNamespacedParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	deploymentDetail, err := k8sDeployment.GetDeploymentDetail(client, query.Namespace, query.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取deployment详情失败")
		return
	}
	response.Success(c, "获取deployment详情成功", deploymentDetail)
}

// GetDeploymentYaml
//
//	@Description: 获取deployment yaml
//	@receiver dp
//	@param c
func (dp *deployment) GetDeploymentYaml(c *gin.Context) {
	var query DeploymentNamespacedParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	yamlContent, err := k8sDeployment.GetDeploymentYaml(client, query.Namespace, query.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取deployment yaml失败")
		return
	}
	response.Success(c, "获取deployment yaml成功", map[string]string{"yaml": yamlContent})
}

// RollbackDeployment
//
//	@Description: 回滚deployment
//	@receiver dp
//	@param c
func (dp *deployment) RollbackDeployment(c *gin.Context) {
	var body DeploymentRollbackParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	if err := k8sDeployment.RollbackDeployment(client, body.Namespace, body.Name, body.Revision); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "回滚deployment失败")
		return
	}
	response.Success(c, "回滚成功", nil)
}

// CreateDeployment
//
//	@Description: 创建deployment
//	@receiver dp
//	@param c
func (dp *deployment) CreateDeployment(c *gin.Context) {
	var body DeploymentCreateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}

	if err := k8sDeployment.CreateDeployment(client, body.Namespace, body.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "创建deployment失败")
		return
	}
	response.Success(c, "创建成功", nil)
}

// UpdateDeployment
//
//	@Description: 更新deployment
//	@receiver dp
//	@param c
func (dp *deployment) UpdateDeployment(c *gin.Context) {
	var body DeploymentUpdateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	if err := k8sDeployment.UpdateDeployment(client, body.Namespace, body.Name, body.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "更新deployment失败")
		return
	}
	response.Success(c, "更新成功", nil)
}

// DeleteDeployment
//
//	@Description: 删除deployment
//	@receiver dp
//	@param c
func (dp *deployment) DeleteDeployment(c *gin.Context) {
	var body DeploymentNamespacedParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	if err := k8sDeployment.DeleteDeployment(client, body.Namespace, body.Name); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "删除deployment失败")
		return
	}
	response.Success(c, "删除成功", nil)
}

// ScaleDeployment
//
//	@Description: 扩所容deployment
//	@receiver dp
//	@param c
func (dp *deployment) ScaleDeployment(c *gin.Context) {
	var body DeploymentScaleParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	// Replicas 为 *int32 且 required,防止漏传/字段名错误时零值缩容到 0
	if body.Replicas == nil {
		response.Fail(c, "副本数不能为空")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	if err := k8sDeployment.ScaleDeployment(client, body.Namespace, body.Name, body.Replicas); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "扩缩容deployment失败")
		return
	}
	response.Success(c, "扩缩容成功", nil)
}

// RestartDeployment
//
//	@Description: 重启deployment
//	@receiver dp
//	@param c
func (dp *deployment) RestartDeployment(c *gin.Context) {
	var body DeploymentRestartParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	if err := k8sDeployment.RestartDeployment(client, body.Namespace, body.Name); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "重启deployment失败")
		return
	}
	response.Success(c, "重启成功", nil)
}

// UpdateDeploymentImage
//
//	@Description: 更新deployment容器镜像
//	@receiver dp
//	@param c
func (dp *deployment) UpdateDeploymentImage(c *gin.Context) {
	var body DeploymentImageUpdateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	if err := k8sDeployment.UpdateDeploymentImage(client, body.Namespace, body.Name, body.ContainerName, body.Image); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "更新deployment镜像失败")
		return
	}
	response.Success(c, "更新镜像成功", nil)
}

// DeploymentPodList
//
//	@Description: 获取deployment pod列表
//	@receiver dp
//	@param c
func (dp *deployment) DeploymentPodList(c *gin.Context) {
	var query DeploymentPodParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	podList, err := k8sDeployment.GetDeploymentPods(client, query.Namespace, query.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取deployment pod列表失败")
		return
	}
	response.Success(c, "获取deployment pod列表成功", podList)
}

// GetDeploymentReplicaSets
//
//	@Description: 获取Deployment关联的ReplicaSet列表
//	@receiver dp
//	@param c
func (dp *deployment) GetDeploymentReplicaSets(c *gin.Context) {
	var query DeploymentReplicaSetParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	rsList, err := k8sDeployment.GetDeploymentReplicaSets(client, query.Namespace, query.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取ReplicaSet列表失败")
		return
	}
	response.Success(c, "获取ReplicaSet列表成功", rsList)
}

// GetDeploymentEvents
//
//	@Description: 获取deployment事件
//	@receiver dp
//	@param c
func (dp *deployment) GetDeploymentEvents(c *gin.Context) {
	var query DeploymentNamespacedParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	result, err := k8sDeployment.GetDeploymentEvents(client, query.Namespace, query.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取deployment事件失败")
		return
	}
	response.Success(c, "获取deployment事件成功", result)
}

type DeploymentListParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type DeploymentCreateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Yaml        string `form:"yaml" json:"yaml" binding:"required" label:"yaml"`
}

type DeploymentUpdateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Yaml        string `form:"yaml" json:"yaml" binding:"required" label:"yaml"`
}

// DeploymentNamespacedParams 按 命名空间+名称 定位 Deployment，供 detail/yaml/events/delete 等多个 handler 复用
type DeploymentNamespacedParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
}

type DeploymentScaleParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Replicas    *int32 `form:"replicas" json:"replicas" binding:"required" label:"副本数"`
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
	Name        string `form:"name" json:"name" binding:"required" label:"Deployment名称"`
}

// DeploymentImageUpdateParams 更新 Deployment 容器镜像
type DeploymentImageUpdateParams struct {
	ClusterName   string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace     string `form:"namespace" json:"namespace" label:"命名空间"`
	Name          string `form:"name" json:"name" binding:"required" label:"Deployment名称"`
	ContainerName string `form:"containerName" json:"containerName" binding:"required" label:"容器名称"`
	Image         string `form:"image" json:"image" binding:"required" label:"镜像"`
}
