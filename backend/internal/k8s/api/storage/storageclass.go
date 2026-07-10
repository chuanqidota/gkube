package storage

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"gkube/internal/k8s/params"
	"gkube/pkg/k8s"
	k8sStorageClass "gkube/pkg/k8s/storageclass"
	"gkube/pkg/response"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type storageClass struct {
}

var StorageClass = new(storageClass)

// GetStorageClassList
//
//	@Description: 获取sc列表
//	@receiver s
//	@param c
func (s *storageClass) GetStorageClassList(c *gin.Context) {
	var query params.StorageClassQueryListParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}

	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	storageClasses, err := k8sStorageClass.GetStorageClassList(client)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取StorageClass列表失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", storageClasses)
}

// GetStorageClassByName
//
//	@Description: 获取sc根据名称
//	@receiver s
//	@param c
func (s *storageClass) GetStorageClassByName(c *gin.Context) {
	var query params.StorageClassQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}

	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	storageClass, err := k8sStorageClass.GetStorageClassByName(client, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取StorageClass失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", storageClass)

}

// GetStorageClassYaml
//
//	@Description: 获取sc的yaml文件
//	@receiver s
//	@param c
func (s *storageClass) GetStorageClassYaml(c *gin.Context) {
	var query params.StorageClassQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}

	client, err := k8s.GetK8sClientByName(query.ClusterName)

	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	storageClassYaml, err := k8sStorageClass.GetStorageClassYaml(client, query.Name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取StorageClass失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", storageClassYaml)
}

// CreateStorageClass
//
//	@Description: 创建sc
//	@receiver s
//	@param c
func (s *storageClass) CreateStorageClass(c *gin.Context) {
	var body params.StorageClassCreateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}

	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	if err := k8sStorageClass.CreateStorageClass(client, body.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("创建StorageClass失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdateStorageClass
//
//	@Description: 更新sc
//	@receiver s
//	@param c
func (s *storageClass) UpdateStorageClass(c *gin.Context) {
	var body params.StorageClassUpdateParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}

	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	if err := k8sStorageClass.UpdateStorageClass(client, body.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("更新StorageClass失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteStorageClassByName
//
//	@Description: 删除sc根据名称
//	@receiver s
//	@param c
func (s *storageClass) DeleteStorageClassByName(c *gin.Context) {
	var body params.StorageClassDeleteByNameParams
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}

	client, err := k8s.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	if err := k8sStorageClass.DeleteStorageClassByName(client, body.Name); err != nil {
		response.Fail(c, fmt.Sprintf("删除StorageClass失败:%v", err.Error()))
		return
	}

	response.Success(c, "执行成功", nil)
}

// GetStorageClassEvents
//
//	@Description: 获取sc事件
//	@receiver s
//	@param c
func (s *storageClass) GetStorageClassEvents(c *gin.Context) {
	var query params.StorageClassQueryByNameParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	if query.Name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}

	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}

	events, err := client.CoreV1().Events(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=StorageClass", query.Name),
	})
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取StorageClass事件失败:%v", err.Error()))
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

