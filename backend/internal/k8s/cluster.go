package k8s

import (
	"fmt"
	"gkube/pkg/response"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sCluster "gkube/pkg/k8s/cluster"
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
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}

	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}

	version, err := k8sCluster.GetClusterVersion(client)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取集群版本失败:%s", err.Error()))
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
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}

	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	nodes, err := k8sCluster.GetClusterNodesInfo(client)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取集群节点信息失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", nodes)
}

type ClusterQueryParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
}
