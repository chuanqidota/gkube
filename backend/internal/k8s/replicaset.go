package k8s

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	k8sclient "gkube/pkg/k8s"
	k8sReplicaSet "gkube/pkg/k8s/replicaset"
	"gkube/pkg/logger"
	"gkube/pkg/response"
	"k8s.io/client-go/kubernetes"
)

// ---------------------------------------------------------------------------
// 标准 handler
// ---------------------------------------------------------------------------

var GetReplicaSetDetail = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sReplicaSet.GetReplicaSetDetail(client, namespace, name)
	},
	"获取ReplicaSet详情成功", "获取ReplicaSet详情失败",
)

var GetReplicaSetYaml = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		yaml, err := k8sReplicaSet.GetReplicaSetYaml(client, namespace, name)
		if err != nil {
			return nil, err
		}
		return map[string]string{"yaml": yaml}, nil
	},
	"获取ReplicaSet YAML成功", "获取ReplicaSet YAML失败",
)

var GetReplicaSetPodList = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sReplicaSet.GetReplicaSetPodList(client, namespace, name)
	},
	"获取ReplicaSet关联Pod成功", "获取ReplicaSet关联Pod失败",
)

var GetReplicaSetEvents = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		events, err := k8sReplicaSet.GetReplicaSetEvents(client, namespace, name)
		if err != nil {
			return nil, err
		}
		var result []map[string]any
		for _, event := range events {
			result = append(result, map[string]any{
				"type":      event.Type,
				"reason":    event.Reason,
				"message":   event.Message,
				"last_seen": event.LastTimestamp,
			})
		}
		return result, nil
	},
	"获取ReplicaSet事件成功", "获取ReplicaSet事件失败",
)

// ---------------------------------------------------------------------------
// 特殊 handler
// ---------------------------------------------------------------------------

// GetReplicaSetList 列表 —— 非分页 + transform 到 map[string]any
func GetReplicaSetList(c *gin.Context) {
	var p ListParams
	if err := c.ShouldBind(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	selector, err := buildLabelSelector(p.LabelFilters)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, err.Error())
		return
	}
	rsList, err := k8sReplicaSet.GetReplicaSetList(client, p.Namespace, selector)
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
		var ownerRefs []map[string]string
		for _, ref := range rs.OwnerReferences {
			ownerRefs = append(ownerRefs, map[string]string{"kind": ref.Kind, "name": ref.Name})
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
			"owner_references":   ownerRefs,
		})
	}
	response.Success(c, "获取ReplicaSet列表成功", result)
}

// DeleteReplicaSet 删除 —— 原代码用 ShouldBindQuery（GET 请求传参）
func DeleteReplicaSet(c *gin.Context) {
	var p NamespacedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	if err := k8sReplicaSet.DeleteReplicaSet(client, p.Namespace, p.Name); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "删除ReplicaSet失败")
		return
	}
	response.Success(c, "删除ReplicaSet成功", nil)
}
