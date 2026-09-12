package k8s

import (
	"github.com/gin-gonic/gin"
	apperr "gkube/pkg/errors"
	k8sclient "gkube/pkg/k8s"
	k8sCluster "gkube/pkg/k8s/cluster"
	"gkube/pkg/response"
)

type cluster struct {
}

var Cluster = new(cluster)

// GetClusterVersion
//
//	@Description: 获取集群版本
//	@receiver cl
//	@param c
func (cl *cluster) GetClusterVersion(c *gin.Context) {
	var query ClusterQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}

	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}

	version, err := k8sCluster.GetClusterVersion(c.Request.Context(), client)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	result := map[string]string{
		"version": version,
	}
	response.Success(c, "执行成功", result)
}

// GetClusterNodesInfo
//
//	@Description: 获取集群节点信息
//	@receiver cl
//	@param c
func (cl *cluster) GetClusterNodesInfo(c *gin.Context) {
	var query ClusterQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}

	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	nodes, err := k8sCluster.GetClusterNodesInfo(c.Request.Context(), client)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", nodes)
}

type ClusterQueryParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
}
