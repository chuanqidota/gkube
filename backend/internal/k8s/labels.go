package k8s

import (
	apperr "gkube/pkg/errors"
	"fmt"

	"github.com/gin-gonic/gin"

	k8sclient "gkube/pkg/k8s"
	k8sLabels "gkube/pkg/k8s/labels"
	"gkube/pkg/logger"
	"gkube/pkg/response"
)

type labelHandler struct{}

var Label = new(labelHandler)

// GetLabels
//
//	@Description: 获取资源可用标签（keys + values）
//	@receiver l
//	@param c
func (l *labelHandler) GetLabels(c *gin.Context) {
	var query struct {
		ClusterName  string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
		Namespace    string `form:"namespace" json:"namespace" label:"命名空间"`
		ResourceType string `form:"resourceType" json:"resourceType" binding:"required" label:"资源类型"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}

	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithError(c, apperr.K8sClientFail("获取K8s客户端失败", err))
		return
	}

	dynamicClient, err := k8sclient.GetDynamicClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithError(c, apperr.K8sClientFail("获取动态客户端失败", err))
		return
	}

	aeClient, err := k8sclient.GetApiExtensionsClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithError(c, apperr.K8sClientFail("获取API扩展客户端失败", err))
		return
	}

	labelData, err := k8sLabels.GetAvailableLabels(c.Request.Context(), client, dynamicClient, aeClient, query.ClusterName, query.Namespace, query.ResourceType)
	if err != nil {
		logger.Error(fmt.Sprintf("获取标签失败 [cluster=%s, resource=%s]: %s", query.ClusterName, query.ResourceType, err.Error()))
		response.FailWithError(c, apperr.K8sAPIFail("获取标签失败", err))
		return
	}

	response.Success(c, "获取标签成功", gin.H{
		"keys":   labelData.Keys,
		"values": labelData.Values,
	})
}
