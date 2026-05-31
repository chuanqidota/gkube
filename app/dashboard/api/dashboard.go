package api

import (
	"context"
	"fmt"
	"sort"

	"github.com/gin-gonic/gin"
	"gkube/app/cluster/model"
	"gkube/app/dashboard/params"
	"gkube/pkg/auth"
	"gkube/pkg/database"
	"gkube/pkg/k8s"
	"gkube/pkg/response"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type dashboard struct{}

var Dashboard = new(dashboard)

// Overview
//
//	@Description: 获取仪表盘概览数据
//	@receiver d
//	@param c
func (d *dashboard) Overview(c *gin.Context) {
	var clusters []model.K8SCluster
	if err := database.DB.Find(&clusters).Error; err != nil {
		response.Fail(c, fmt.Sprintf("获取集群列表失败:%s", err.Error()))
		return
	}

	var totalClusters, onlineClusters, offlineClusters int
	var totalNodes int

	for _, cluster := range clusters {
		totalClusters++
		totalNodes += cluster.NodeCount
		if cluster.Status == "online" {
			onlineClusters++
		} else {
			offlineClusters++
		}
	}

	// 从在线集群获取pod总数
	var totalPods int
	for _, cluster := range clusters {
		if cluster.Status != "online" {
			continue
		}
		kubeConfig, err := auth.DecryptAES(cluster.KubeConfig)
		if err != nil {
			continue
		}
		client, err := k8s.GetK8sClient(kubeConfig)
		if err != nil {
			continue
		}
		podList, err := client.CoreV1().Pods(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			continue
		}
		totalPods += len(podList.Items)
	}

	data := map[string]any{
		"totalClusters":   totalClusters,
		"onlineClusters":  onlineClusters,
		"offlineClusters": offlineClusters,
		"totalNodes":      totalNodes,
		"totalPods":       totalPods,
	}
	response.Success(c, "获取概览数据成功", data)
}

// Resources
//
//	@Description: 获取各集群资源使用情况
//	@receiver d
//	@param c
func (d *dashboard) Resources(c *gin.Context) {
	var clusters []model.K8SCluster
	if err := database.DB.Where("status = ?", "online").Find(&clusters).Error; err != nil {
		response.Fail(c, fmt.Sprintf("获取集群列表失败:%s", err.Error()))
		return
	}

	var resources []map[string]any
	for _, cluster := range clusters {
		kubeConfig, err := auth.DecryptAES(cluster.KubeConfig)
		if err != nil {
			continue
		}
		client, err := k8s.GetK8sClient(kubeConfig)
		if err != nil {
			continue
		}

		// 获取节点数量
		nodeList, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			continue
		}
		nodeCount := len(nodeList.Items)

		// 获取pod数量
		podList, err := client.CoreV1().Pods(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			continue
		}
		podCount := len(podList.Items)

		// 统计运行中的pod数量
		runningPodCount := 0
		for _, pod := range podList.Items {
			if pod.Status.Phase == corev1.PodRunning {
				runningPodCount++
			}
		}

		resources = append(resources, map[string]any{
			"clusterId":     cluster.ID,
			"clusterName":   cluster.ClusterName,
			"displayName":   cluster.DisplayName,
			"nodeCount":     nodeCount,
			"podCount":      podCount,
			"runningPods":   runningPodCount,
		})
	}
	response.Success(c, "获取资源信息成功", resources)
}

// Workloads
//
//	@Description: 获取各集群工作负载统计
//	@receiver d
//	@param c
func (d *dashboard) Workloads(c *gin.Context) {
	var clusters []model.K8SCluster
	if err := database.DB.Where("status = ?", "online").Find(&clusters).Error; err != nil {
		response.Fail(c, fmt.Sprintf("获取集群列表失败:%s", err.Error()))
		return
	}

	var workloads []map[string]any
	for _, cluster := range clusters {
		kubeConfig, err := auth.DecryptAES(cluster.KubeConfig)
		if err != nil {
			continue
		}
		client, err := k8s.GetK8sClient(kubeConfig)
		if err != nil {
			continue
		}

		// 统计Deployment数量
		deploymentList, err := client.AppsV1().Deployments(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			continue
		}
		deploymentCount := len(deploymentList.Items)

		// 统计StatefulSet数量
		statefulSetList, err := client.AppsV1().StatefulSets(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			continue
		}
		statefulSetCount := len(statefulSetList.Items)

		// 统计DaemonSet数量
		daemonSetList, err := client.AppsV1().DaemonSets(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			continue
		}
		daemonSetCount := len(daemonSetList.Items)

		workloads = append(workloads, map[string]any{
			"clusterId":      cluster.ID,
			"clusterName":    cluster.ClusterName,
			"displayName":    cluster.DisplayName,
			"deployments":    deploymentCount,
			"statefulSets":   statefulSetCount,
			"daemonSets":     daemonSetCount,
		})
	}
	response.Success(c, "获取工作负载信息成功", workloads)
}

// Events
//
//	@Description: 获取集群事件列表
//	@receiver d
//	@param c
func (d *dashboard) Events(c *gin.Context) {
	var query params.EventQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}

	// 设置默认限制
	if query.Limit <= 0 {
		query.Limit = 50
	}

	// 获取要查询的集群列表
	var clusters []model.K8SCluster
	if query.ClusterID != nil {
		if err := database.DB.Where("id = ?", *query.ClusterID).Find(&clusters).Error; err != nil {
			response.Fail(c, fmt.Sprintf("获取集群信息失败:%s", err.Error()))
			return
		}
	} else {
		if err := database.DB.Where("status = ?", "online").Find(&clusters).Error; err != nil {
			response.Fail(c, fmt.Sprintf("获取集群列表失败:%s", err.Error()))
			return
		}
	}

	var allEvents []map[string]any
	for _, cluster := range clusters {
		if cluster.Status != "online" {
			continue
		}
		kubeConfig, err := auth.DecryptAES(cluster.KubeConfig)
		if err != nil {
			continue
		}
		client, err := k8s.GetK8sClient(kubeConfig)
		if err != nil {
			continue
		}

		eventList, err := client.CoreV1().Events(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			continue
		}

		for _, event := range eventList.Items {
			// 按类型过滤
			if query.Type != "" && event.Type != query.Type {
				continue
			}
			allEvents = append(allEvents, map[string]any{
				"clusterId":   cluster.ID,
				"clusterName": cluster.ClusterName,
				"namespace":   event.Namespace,
				"type":        event.Type,
				"reason":      event.Reason,
				"message":     event.Message,
				"object":      fmt.Sprintf("%s/%s", event.InvolvedObject.Kind, event.InvolvedObject.Name),
				"count":       event.Count,
				"firstTime":   event.FirstTimestamp.Time,
				"lastTime":    event.LastTimestamp.Time,
			})
		}
	}

	// 按最后时间倒序排序
	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i]["lastTime"].(string) > allEvents[j]["lastTime"].(string)
	})

	// 限制返回数量
	if len(allEvents) > query.Limit {
		allEvents = allEvents[:query.Limit]
	}

	response.Success(c, "获取事件列表成功", allEvents)
}
