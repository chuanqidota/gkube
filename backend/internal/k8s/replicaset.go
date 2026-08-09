package k8s

import (
	"fmt"
	"net/http"
	"time"

	"context"
	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sReplicaSet "gkube/pkg/k8s/replicaset"
	"gkube/pkg/logger"
	"gkube/pkg/response"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type replicaset struct{}

var ReplicaSet = new(replicaset)

// ReplicaSetListParams 列表查询参数
type ReplicaSetListParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

// ReplicaSetNamespacedParams 按命名空间+名称定位资源，供 detail/yaml/pods/events/delete 复用
type ReplicaSetNamespacedParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" binding:"required" label:"命名空间"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
}

func (r *replicaset) GetReplicaSetList(c *gin.Context) {
	var query ReplicaSetListParams
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
	rsList, err := k8sReplicaSet.GetReplicaSetList(client, query.Namespace)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取ReplicaSet列表失败")
		return
	}
	var result []map[string]any
	for _, rs := range rsList {
		var replicas int32
		if rs.Spec.Replicas != nil {
			replicas = *rs.Spec.Replicas
		}
		result = append(result, map[string]any{
			"name":               rs.Name,
			"namespace":          rs.Namespace,
			"desired":            replicas,
			"current":            rs.Status.Replicas,
			"ready":              rs.Status.ReadyReplicas,
			"available":          rs.Status.AvailableReplicas,
			"fully_labeled":      rs.Status.FullyLabeledReplicas,
			"creation_timestamp": rs.CreationTimestamp.Time.Format(time.RFC3339),
			"labels":             rs.Labels,
			"owner_references":   rs.OwnerReferences,
		})
	}
	response.Success(c, "获取ReplicaSet列表成功", result)
}

func (r *replicaset) GetReplicaSetYaml(c *gin.Context) {
	var query ReplicaSetNamespacedParams
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
	yamlContent, err := k8sReplicaSet.GetReplicaSetYaml(client, query.Namespace, query.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取ReplicaSet YAML失败")
		return
	}
	response.Success(c, "获取ReplicaSet YAML成功", map[string]string{"yaml": yamlContent})
}

func (r *replicaset) DeleteReplicaSet(c *gin.Context) {
	var query ReplicaSetNamespacedParams
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
	if err := k8sReplicaSet.DeleteReplicaSet(client, query.Namespace, query.Name); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "删除ReplicaSet失败")
		return
	}
	response.Success(c, "删除ReplicaSet成功", nil)
}

func (r *replicaset) GetReplicaSetDetail(c *gin.Context) {
	var query ReplicaSetNamespacedParams
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
	detail, err := k8sReplicaSet.GetReplicaSetDetail(client, query.Namespace, query.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取ReplicaSet详情失败")
		return
	}
	response.Success(c, "获取ReplicaSet详情成功", detail)
}

func (r *replicaset) GetReplicaSetPodList(c *gin.Context) {
	var query ReplicaSetNamespacedParams
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
	podList, err := k8sReplicaSet.GetReplicaSetPodList(client, query.Namespace, query.Name)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取ReplicaSet关联Pod失败")
		return
	}
	response.Success(c, "获取ReplicaSet关联Pod成功", podList)
}

func (r *replicaset) GetReplicaSetEvents(c *gin.Context) {
	var query ReplicaSetNamespacedParams
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
	events, err := client.CoreV1().Events(query.Namespace).List(context.Background(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=ReplicaSet", query.Name),
	})
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取ReplicaSet事件失败")
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
	response.Success(c, "获取ReplicaSet事件成功", result)
}
