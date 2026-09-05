package k8s

import (
	"net/http"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sConfigMap "gkube/pkg/k8s/configmap"
	k8sLabels "gkube/pkg/k8s/labels"
	"gkube/pkg/logger"
	"gkube/pkg/response"
)

type configmap struct {
}

var ConfigMap = new(configmap)

// GetConfigMapList
//
//	@Description: 获取cm列表（支持分页+labelSelector）
//	@receiver cm
//	@param c
func (cm *configmap) GetConfigMapList(c *gin.Context) {
	var query ConfigMapQueryListParams
	if err := c.ShouldBind(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	// 构建 label selector
	var selector string
	if len(query.LabelFilters) > 0 {
		selector, err = k8sLabels.BuildLabelSelector(query.LabelFilters)
		if err != nil {
			response.Fail(c, err.Error())
			return
		}
	}


	limit := query.Limit
	continueToken := query.Continue
	configMapList, err := k8sConfigMap.GetConfigMapList(client, query.Namespace, limit, continueToken, selector)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取ConfigMap列表失败")
		return
	}
	remaining := int64(0)
	if configMapList.RemainingItemCount != nil {
		remaining = *configMapList.RemainingItemCount
	}
	data := k8sclient.BuildPaginatedData(configMapList.Items, configMapList.Continue, remaining, limit)
	data.Total = len(configMapList.Items) + int(remaining)
	response.Success(c, "获取ConfigMap列表成功", data)
}

// GetConfigMapByName
//
//	@Description: 获取cm根据名称
//	@receiver cm
//	@param c
func (cm *configmap) GetConfigMapByName(c *gin.Context) {
	var query ConfigMapQueryByNameParams
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

	configMap, err := k8sConfigMap.GetConfigMapByName(client, query.Namespace, query.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取ConfigMap失败")
		return
	}
	response.Success(c, "获取ConfigMap成功", configMap)
}

// GetConfigMapYaml
//
//	@Description: 获取cm的yaml
//	@receiver cm
//	@param c
func (cm *configmap) GetConfigMapYaml(c *gin.Context) {
	var query ConfigMapQueryByNameParams
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
	configMapYaml, err := k8sConfigMap.GetConfigMapYaml(client, query.Namespace, query.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取ConfigMap YAML失败")
		return
	}
	response.Success(c, "获取ConfigMap YAML成功", map[string]string{"yaml": configMapYaml})
}

// DeleteConfigMapByName
//
//	@Description: 删除cm根据名称
//	@receiver cm
//	@param c
func (cm *configmap) DeleteConfigMapByName(c *gin.Context) {
	var body ConfigMapQueryByNameParams
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
	if err := k8sConfigMap.DeleteConfigMap(client, body.Namespace, body.Name); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "删除ConfigMap失败")
		return
	}
	response.Success(c, "删除ConfigMap成功", nil)
}

type ConfigMapYamlParams struct {
	ClusterName string `json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `json:"namespace" binding:"required" label:"命名空间"`
	Yaml        string `json:"yaml" binding:"required" label:"YAML内容"`
}

// UpdateConfigMapFromYaml
//
//	@Description: 通过YAML更新cm
//	@receiver cm
//	@param c
func (cm *configmap) UpdateConfigMapFromYaml(c *gin.Context) {
	var req ConfigMapYamlParams
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(req.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	if err := k8sConfigMap.UpdateConfigMapFromYaml(client, req.Namespace, req.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "更新ConfigMap失败")
		return
	}
	response.Success(c, "更新ConfigMap成功", nil)
}

// CreateConfigMapFromYaml
//
//	@Description: 通过YAML创建cm
//	@receiver cm
//	@param c
func (cm *configmap) CreateConfigMapFromYaml(c *gin.Context) {
	var req ConfigMapYamlParams
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(req.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	if err := k8sConfigMap.CreateConfigMapFromYaml(client, req.Namespace, req.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "创建ConfigMap失败")
		return
	}
	response.Success(c, "创建ConfigMap成功", nil)
}

type ConfigMapQueryListParams struct {
	ClusterName  string                  `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace    string                  `form:"namespace" json:"namespace" label:"命名空间"`
	Limit        int64                   `form:"limit" json:"limit" label:"每页条数"`
	Continue     string                  `form:"continue" json:"continue" label:"分页标记"`
	LabelFilters []k8sLabels.LabelFilter `json:"labelFilters" form:"labelFilters" label:"标签过滤"`
}

type ConfigMapQueryByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}
