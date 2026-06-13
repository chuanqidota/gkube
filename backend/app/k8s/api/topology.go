package api

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"gkube/pkg/k8s"
	"gkube/pkg/response"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type topology struct{}

var Topology = new(topology)

// GetDeploymentTopology returns the relationship: Deployment -> ReplicaSet -> Pod
func (t *topology) GetDeploymentTopology(c *gin.Context) {
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

	// Get deployment
	deploy, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取Deployment失败:%s", err.Error()))
		return
	}

	// Get ReplicaSets owned by this deployment
	rsList, err := client.AppsV1().ReplicaSets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取ReplicaSet列表失败:%s", err.Error()))
		return
	}

	var replicaSets []map[string]any
	for _, rs := range rsList.Items {
		for _, ownerRef := range rs.OwnerReferences {
			if ownerRef.UID == deploy.UID {
				// Get pods owned by this ReplicaSet
				podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
				if err != nil {
					continue
				}
				var pods []map[string]any
				for _, pod := range podList.Items {
					for _, podOwnerRef := range pod.OwnerReferences {
						if podOwnerRef.UID == rs.UID {
							pods = append(pods, map[string]any{
								"name":   pod.Name,
								"status": string(pod.Status.Phase),
								"ip":     pod.Status.PodIP,
								"age":    pod.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
							})
						}
					}
				}
				replicaSets = append(replicaSets, map[string]any{
					"name":       rs.Name,
					"replicas":   *rs.Spec.Replicas,
					"ready":      rs.Status.ReadyReplicas,
					"age":        rs.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
					"revision":   rs.Annotations["deployment.kubernetes.io/revision"],
					"pods":       pods,
				})
			}
		}
	}

	result := map[string]any{
		"deployment": map[string]any{
			"name":     deploy.Name,
			"replicas": *deploy.Spec.Replicas,
			"ready":    deploy.Status.ReadyReplicas,
			"age":      deploy.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
		},
		"replicaSets": replicaSets,
	}
	response.Success(c, "执行成功", result)
}

// GetStatefulSetTopology returns the relationship: StatefulSet -> Pod
func (t *topology) GetStatefulSetTopology(c *gin.Context) {
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

	sts, err := client.AppsV1().StatefulSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取StatefulSet失败:%s", err.Error()))
		return
	}

	podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取Pod列表失败:%s", err.Error()))
		return
	}

	var pods []map[string]any
	for _, pod := range podList.Items {
		for _, ownerRef := range pod.OwnerReferences {
			if ownerRef.UID == sts.UID {
				pods = append(pods, map[string]any{
					"name":   pod.Name,
					"status": string(pod.Status.Phase),
					"ip":     pod.Status.PodIP,
					"node":   pod.Spec.NodeName,
					"age":    pod.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
				})
			}
		}
	}

	result := map[string]any{
		"statefulSet": map[string]any{
			"name":     sts.Name,
			"replicas": *sts.Spec.Replicas,
			"ready":    sts.Status.ReadyReplicas,
			"age":      sts.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
		},
		"pods": pods,
	}
	response.Success(c, "执行成功", result)
}

// GetDaemonSetTopology returns the relationship: DaemonSet -> Pod
func (t *topology) GetDaemonSetTopology(c *gin.Context) {
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

	ds, err := client.AppsV1().DaemonSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取DaemonSet失败:%s", err.Error()))
		return
	}

	podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取Pod列表失败:%s", err.Error()))
		return
	}

	var pods []map[string]any
	for _, pod := range podList.Items {
		for _, ownerRef := range pod.OwnerReferences {
			if ownerRef.UID == ds.UID {
				pods = append(pods, map[string]any{
					"name":   pod.Name,
					"status": string(pod.Status.Phase),
					"ip":     pod.Status.PodIP,
					"node":   pod.Spec.NodeName,
					"age":    pod.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
				})
			}
		}
	}

	result := map[string]any{
		"daemonSet": map[string]any{
			"name":           ds.Name,
			"desiredNumber":  ds.Status.DesiredNumberScheduled,
			"currentNumber":  ds.Status.CurrentNumberScheduled,
			"readyNumber":    ds.Status.NumberReady,
			"age":            ds.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
		},
		"pods": pods,
	}
	response.Success(c, "执行成功", result)
}
