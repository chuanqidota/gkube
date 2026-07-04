package api

import (
	"context"
	"fmt"
	"sort"

	"github.com/gin-gonic/gin"
	"gkube/internal/cluster/model"
	"gkube/internal/dashboard/params"
	"gkube/pkg/auth"
	"gkube/pkg/database"
	"gkube/pkg/k8s"
	"gkube/pkg/response"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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

	clusterCount := len(clusters)
	var nodeCount, podCount, namespaceCount int

	for _, cluster := range clusters {
		nodeCount += cluster.NodeCount

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

		// Count pods
		podList, err := client.CoreV1().Pods(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		if err == nil {
			podCount += len(podList.Items)
		}

		// Count namespaces
		nsList, err := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
		if err == nil {
			nsCount := len(nsList.Items)
			if nsCount > namespaceCount {
				namespaceCount = nsCount
			}
		}
	}

	data := map[string]any{
		"cluster_count":   clusterCount,
		"node_count":      nodeCount,
		"pod_count":       podCount,
		"namespace_count": namespaceCount,
	}
	response.Success(c, "获取概览数据成功", data)
}

// Resources
//
//	@Description: 获取集群资源使用情况（CPU/内存/存储）
//	@receiver d
//	@param c
func (d *dashboard) Resources(c *gin.Context) {
	var clusters []model.K8SCluster
	if err := database.DB.Where("status = ?", "online").Find(&clusters).Error; err != nil {
		response.Fail(c, fmt.Sprintf("获取集群列表失败:%s", err.Error()))
		return
	}

	var totalCPUUsed, totalCPUTotal resource.Quantity
	var totalMemUsed, totalMemTotal resource.Quantity
	var totalStorageUsed, totalStorageTotal resource.Quantity

	for _, cluster := range clusters {
		kubeConfig, err := auth.DecryptAES(cluster.KubeConfig)
		if err != nil {
			continue
		}
		client, err := k8s.GetK8sClient(kubeConfig)
		if err != nil {
			continue
		}

		// Sum node capacity and allocatable
		nodeList, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			continue
		}
		for _, node := range nodeList.Items {
			totalCPUTotal.Add(*node.Status.Capacity.Cpu())
			totalMemTotal.Add(*node.Status.Capacity.Memory())
			totalStorageTotal.Add(*node.Status.Capacity.StorageEphemeral())

			totalCPUUsed.Add(*node.Status.Allocatable.Cpu())
			totalMemUsed.Add(*node.Status.Allocatable.Memory())
		}

		// Sum pod resource requests as "used"
		podList, err := client.CoreV1().Pods(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			continue
		}
		for _, pod := range podList.Items {
			if pod.Status.Phase != corev1.PodRunning && pod.Status.Phase != corev1.PodPending {
				continue
			}
			for _, container := range pod.Spec.Containers {
				if req, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
					totalCPUUsed.Add(req)
				}
				if req, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
					totalMemUsed.Add(req)
				}
			}
		}
	}

	// Convert to human-readable units
	// CPU in cores (float64)
	cpuUsed := float64(totalCPUUsed.MilliValue()) / 1000.0
	cpuTotal := float64(totalCPUTotal.MilliValue()) / 1000.0

	// Memory in GiB (float64)
	memUsed := float64(totalMemUsed.Value()) / (1024 * 1024 * 1024)
	memTotal := float64(totalMemTotal.Value()) / (1024 * 1024 * 1024)

	// Storage in GiB (float64)
	storageUsed := float64(totalStorageUsed.Value()) / (1024 * 1024 * 1024)
	storageTotal := float64(totalStorageTotal.Value()) / (1024 * 1024 * 1024)

	data := map[string]any{
		"cpu": map[string]any{
			"used":  cpuUsed,
			"total": cpuTotal,
		},
		"memory": map[string]any{
			"used":  memUsed,
			"total": memTotal,
		},
		"storage": map[string]any{
			"used":  storageUsed,
			"total": storageTotal,
		},
	}
	response.Success(c, "获取资源信息成功", data)
}

// Workloads
//
//	@Description: 获取所有集群工作负载统计
//	@receiver d
//	@param c
func (d *dashboard) Workloads(c *gin.Context) {
	var clusters []model.K8SCluster
	if err := database.DB.Where("status = ?", "online").Find(&clusters).Error; err != nil {
		response.Fail(c, fmt.Sprintf("获取集群列表失败:%s", err.Error()))
		return
	}

	var totalDeployments, totalStatefulSets, totalDaemonSets, totalJobs, totalCronJobs int

	for _, cluster := range clusters {
		kubeConfig, err := auth.DecryptAES(cluster.KubeConfig)
		if err != nil {
			continue
		}
		client, err := k8s.GetK8sClient(kubeConfig)
		if err != nil {
			continue
		}

		deployments, err := client.AppsV1().Deployments(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		if err == nil {
			totalDeployments += len(deployments.Items)
		}

		statefulSets, err := client.AppsV1().StatefulSets(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		if err == nil {
			totalStatefulSets += len(statefulSets.Items)
		}

		daemonSets, err := client.AppsV1().DaemonSets(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		if err == nil {
			totalDaemonSets += len(daemonSets.Items)
		}

		jobs, err := client.BatchV1().Jobs(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		if err == nil {
			totalJobs += len(jobs.Items)
		}

		cronJobs, err := client.BatchV1().CronJobs(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		if err == nil {
			totalCronJobs += len(cronJobs.Items)
		}
	}

	data := map[string]any{
		"deployments":  totalDeployments,
		"statefulsets": totalStatefulSets,
		"daemonsets":   totalDaemonSets,
		"jobs":         totalJobs,
		"cronjobs":     totalCronJobs,
	}
	response.Success(c, "获取工作负载信息成功", data)
}

// Events
//
//	@Description: 获取集群事件列表（支持分页）
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
		query.Limit = 100
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

		// 构建命名空间
		namespace := query.Namespace
		if namespace == "" {
			namespace = corev1.NamespaceAll
		}

		// 构建 ListOptions
		listOpts := metav1.ListOptions{}
		if query.FieldSelector != "" {
			listOpts.FieldSelector = query.FieldSelector
		}

		eventList, err := client.CoreV1().Events(namespace).List(context.TODO(), listOpts)
		if err != nil {
			continue
		}

		for _, event := range eventList.Items {
			// 按类型过滤
			if query.Type != "" && event.Type != query.Type {
				continue
			}

			// 格式化时间
			firstSeen := ""
			if !event.FirstTimestamp.IsZero() {
				firstSeen = event.FirstTimestamp.Time.Format("2006-01-02 15:04:05")
			}
			lastSeen := ""
			if !event.LastTimestamp.IsZero() {
				lastSeen = event.LastTimestamp.Time.Format("2006-01-02 15:04:05")
			} else if !event.EventTime.IsZero() {
				lastSeen = event.EventTime.Time.Format("2006-01-02 15:04:05")
			}

			allEvents = append(allEvents, map[string]any{
				"type":                 event.Type,
				"reason":               event.Reason,
				"message":              event.Message,
				"namespace":            event.Namespace,
				"involved_object":      fmt.Sprintf("%s/%s", event.InvolvedObject.Kind, event.InvolvedObject.Name),
				"involved_object_kind": event.InvolvedObject.Kind,
				"involved_object_name": event.InvolvedObject.Name,
				"first_seen":           firstSeen,
				"last_seen":            lastSeen,
				"count":                event.Count,
				"reporting_component":  event.ReportingController,
				"reporting_instance":   event.ReportingInstance,
				"action":               event.Action,
				"cluster_name":         cluster.ClusterName,
			})
		}
	}

	// 按最后时间倒序排序
	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i]["last_seen"].(string) > allEvents[j]["last_seen"].(string)
	})

	// 计算总数
	total := len(allEvents)

	// 处理分页（基于 offset 的简单分页）
	offset := 0
	if query.Continue != "" {
		fmt.Sscanf(query.Continue, "%d", &offset)
	}

	// 限制返回数量
	end := offset + query.Limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = total
	}

	pagedEvents := allEvents[offset:end]

	// 构建分页响应
	continueToken := ""
	if end < total {
		continueToken = fmt.Sprintf("%d", end)
	}

	data := map[string]any{
		"items":    pagedEvents,
		"total":    total,
		"continue": continueToken,
		"has_more": continueToken != "",
	}

	response.Success(c, "获取事件列表成功", data)
}
