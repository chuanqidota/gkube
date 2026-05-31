package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gkube/app/k8s/params"
	"gkube/pkg/k8s"
	k8sConfigMap "gkube/pkg/k8s/configmap"
	"gkube/pkg/response"
)

type configmap struct {
}

var ConfigMap = new(configmap)

// GetConfigMapList
//
//	@Description: 获取cm列表
//	@receiver cm
//	@param c
func (cm *configmap) GetConfigMapList(c *gin.Context) {
	var query params.ConfigMapQueryListParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)

	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	configMaps, err := k8sConfigMap.GetConfigMapList(client, query.Namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ConfigMap列表失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", configMaps)
}

// GetConfigMapByName
//
//	@Description: 获取cm根据名称
//	@receiver cm
//	@param c
func (cm *configmap) GetConfigMapByName(c *gin.Context) {
	var query params.ConfigMapQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}

	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	configMap, err := k8sConfigMap.GetConfigMapByName(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ConfigMap失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", configMap)
}

// GetConfigMapYaml
//
//	@Description: 获取cm的yaml
//	@receiver cm
//	@param c
func (cm *configmap) GetConfigMapYaml(c *gin.Context) {
	var query params.ConfigMapQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	configMapYaml, err := k8sConfigMap.GetConfigMapYaml(client, query.Namespace, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ConfigMap Yaml失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", configMapYaml)
}

// CreateConfigMap
//
//	@Description: 创建cm
//	@receiver cm
//	@param c
func (cm *configmap) CreateConfigMap(c *gin.Context) {
	var body params.ConfigMapCreateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sConfigMap.CreateConfigMap(client, body.Namespace, body.Name, body.Data); err != nil {
		response.Fail(c, fmt.Sprintf("创建ConfigMap失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdateConfigMap
//
//	@Description: 更新cm
//	@receiver cm
//	@param c
func (cm *configmap) UpdateConfigMap(c *gin.Context) {
	var body params.ConfigMapUpdateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
	}

	if err := k8sConfigMap.UpdateConfigMap(client, body.Namespace, body.Name, body.Data); err != nil {
		response.Fail(c, fmt.Sprintf("更新ConfigMap失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteConfigMapByName
//
//	@Description: 删除cm根据名称
//	@receiver cm
//	@param c
func (cm *configmap) DeleteConfigMapByName(c *gin.Context) {
	var body params.ConfigMapQueryByNameParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sConfigMap.DeleteConfigMap(client, body.Namespace, body.Name); err != nil {
		response.Fail(c, fmt.Sprintf("删除ConfigMap失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}
