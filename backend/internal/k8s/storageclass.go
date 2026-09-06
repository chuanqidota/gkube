package k8s

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sStorageClass "gkube/pkg/k8s/storageclass"
	"gkube/pkg/logger"
	"gkube/pkg/response"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

// ---------------------------------------------------------------------------
// 特殊 handler —— StorageClass 是集群级资源，pkg 函数不接受 namespace
// ---------------------------------------------------------------------------

// GetStorageClassList 列表 —— 非分页
func GetStorageClassList(c *gin.Context) {
	var p ListParams
	if err := c.ShouldBind(&p); err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	selector, err := buildLabelSelector(p.LabelFilters)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, err.Error())
		return
	}
	storageClasses, err := k8sStorageClass.GetStorageClassList(client, selector)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取StorageClass列表失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", storageClasses)
}

// GetStorageClassByName 详情
func GetStorageClassByName(c *gin.Context) {
	var p ClusterScopedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	storageClass, err := k8sStorageClass.GetStorageClassByName(client, p.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取StorageClass失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", storageClass)
}

// GetStorageClassYaml YAML
func GetStorageClassYaml(c *gin.Context) {
	var p ClusterScopedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	storageClassYaml, err := k8sStorageClass.GetStorageClassYaml(client, p.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取StorageClass失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", storageClassYaml)
}

// CreateStorageClass 创建 —— 只传 yaml
func CreateStorageClass(c *gin.Context) {
	var p ClusterCreateParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sStorageClass.CreateStorageClass(client, p.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("创建StorageClass失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// UpdateStorageClass 更新 —— 只传 yaml
func UpdateStorageClass(c *gin.Context) {
	var p ClusterCreateParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sStorageClass.UpdateStorageClass(client, p.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("更新StorageClass失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// DeleteStorageClassByName 删除
func DeleteStorageClassByName(c *gin.Context) {
	var p ClusterScopedParams
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	if err := k8sStorageClass.DeleteStorageClassByName(client, p.Name); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("删除StorageClass失败:%v", err.Error()))
		return
	}
	response.Success(c, "执行成功", nil)
}

// GetStorageClassEvents 事件 —— 内联 K8s 调用（集群级资源，需跨命名空间搜索事件）
func GetStorageClassEvents(c *gin.Context) {
	var p ClusterScopedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("参数错误:%v", err.Error()))
		return
	}
	if p.Name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取k8s客户端失败:%v", err.Error()))
		return
	}
	events, err := client.CoreV1().Events(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fields.AndSelectors(
			fields.OneTermEqualSelector("involvedObject.name", p.Name),
			fields.OneTermEqualSelector("involvedObject.kind", "StorageClass"),
		).String(),
	})
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取StorageClass事件失败:%v", err.Error()))
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
