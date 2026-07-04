package api

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gkube/pkg/k8s"
	k8sReplicaSet "gkube/pkg/k8s/replicaset"
	"gkube/pkg/response"
)

type replicaset struct{}

var ReplicaSet = new(replicaset)

func (r *replicaset) GetReplicaSetList(c *gin.Context) {
	namespace := c.Query("namespace")
	clusterName := c.Query("clusterName")
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	rsList, err := k8sReplicaSet.GetReplicaSetList(client, namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ReplicaSet列表失败:%s", err.Error()))
		return
	}
	var result []map[string]any
	for _, rs := range rsList {
		var replicas int32
		if rs.Spec.Replicas != nil {
			replicas = *rs.Spec.Replicas
		}
		result = append(result, map[string]any{
			"name":            rs.Name,
			"namespace":       rs.Namespace,
			"desired":         replicas,
			"current":         rs.Status.Replicas,
			"ready":           rs.Status.ReadyReplicas,
			"available":       rs.Status.AvailableReplicas,
			"fully_labeled":   rs.Status.FullyLabeledReplicas,
			"creation_timestamp": rs.CreationTimestamp.Time.Format(time.RFC3339),
			"labels":          rs.Labels,
			"owner_references": rs.OwnerReferences,
		})
	}
	response.Success(c, "执行成功", result)
}

func (r *replicaset) GetReplicaSetYaml(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	if name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	yamlContent, err := k8sReplicaSet.GetReplicaSetYaml(client, namespace, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ReplicaSet YAML失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yamlContent})
}

func (r *replicaset) DeleteReplicaSet(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	if name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sReplicaSet.DeleteReplicaSet(client, namespace, name); err != nil {
		response.Fail(c, fmt.Sprintf("删除ReplicaSet失败:%s", err.Error()))
		return
	}
	response.Success(c, "删除ReplicaSet成功", nil)
}
