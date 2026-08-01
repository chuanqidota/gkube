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

// getTargetClusters 按 clusterID 选取目标集群:
//   - clusterID 命中:返回该单个集群(无论状态),与 Events 既有行为一致;
//   - clusterID 为空:返回所有 online 集群(向后兼容,聚合全部)。
//
// 找不到指定集群时返回空切片与 nil error,由调用方自然产出零值结果。
func getTargetClusters(clusterID *uint) ([]model.K8SCluster, error) {
	var clusters []model.K8SCluster
	if clusterID != nil {
		if err := database.DB.Where("id = ?", *clusterID).Find(&clusters).Error; err != nil {
			return nil, err
		}
		return clusters, nil
	}
	if err := database.DB.Where("status = ?", "online").Find(&clusters).Error; err != nil {
		return nil, err
	}
	return clusters, nil
}

// Overview
//
//	@Description: 获取仪表盘概览数据
//	@receiver d
//	@param c
func (d *dashboard) Overview(c *gin.Context) {
	var query params.DashboardQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}

	clusters, err := getTargetClusters(query.ClusterID)
	if err != nil {
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
	var query params.DashboardQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}

	clusters, err := getTargetClusters(query.ClusterID)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取集群列表失败:%s", err.Error()))
		return
	}

	var totalCPUUsed, totalCPUTotal resource.Quantity
	var totalMemUsed, totalMemTotal resource.Quantity
	var totalStorageUsed, totalStorageTotal resource.Quantity

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

		// 总量取 Allocatable(真实可调度,Capacity 减去系统预留),与节点级口径一致
		nodeList, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			continue
		}
		for _, node := range nodeList.Items {
			if allocCPU, ok := node.Status.Allocatable[corev1.ResourceCPU]; ok {
				totalCPUTotal.Add(allocCPU)
			}
			if allocMem, ok := node.Status.Allocatable[corev1.ResourceMemory]; ok {
				totalMemTotal.Add(allocMem)
			}
			// ephemeral-storage 优先 Allocatable,回退 Capacity
			if allocStor := node.Status.Allocatable.StorageEphemeral(); !allocStor.IsZero() {
				totalStorageTotal.Add(*allocStor)
			} else if storageCap := node.Status.Capacity.StorageEphemeral(); !storageCap.IsZero() {
				totalStorageTotal.Add(*storageCap)
			}
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
				if req, ok := container.Resources.Requests[corev1.ResourceEphemeralStorage]; ok {
					totalStorageUsed.Add(req)
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
	var query params.DashboardQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}

	clusters, err := getTargetClusters(query.ClusterID)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取集群列表失败:%s", err.Error()))
		return
	}

	var totalDeployments, totalStatefulSets, totalDaemonSets, totalJobs, totalCronJobs int

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

// Namespaces
//
//	@Description: 获取按命名空间聚合的资源占用(Pod 数 + CPU/内存 requests),及集群总量
//	@receiver d
//	@param c
func (d *dashboard) Namespaces(c *gin.Context) {
	var query params.DashboardQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}

	clusters, err := getTargetClusters(query.ClusterID)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取集群列表失败:%s", err.Error()))
		return
	}

	type nsAgg struct {
		Name        string
		PodCount    int
		RunningPods int
		CPU         resource.Quantity
		Mem         resource.Quantity
	}
	agg := make(map[string]*nsAgg)
	var totalCPU, totalMem resource.Quantity

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

		// 集群总量 = 所有节点 Allocatable 之和(真实可调度)
		nodeList, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err == nil {
			for _, node := range nodeList.Items {
				if allocCPU, ok := node.Status.Allocatable[corev1.ResourceCPU]; ok {
					totalCPU.Add(allocCPU)
				}
				if allocMem, ok := node.Status.Allocatable[corev1.ResourceMemory]; ok {
					totalMem.Add(allocMem)
				}
			}
		}

		// 按命名空间累加 Pod 数与 requests(仅 Running/Pending,与 Resources 口径一致)
		podList, err := client.CoreV1().Pods(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			continue
		}
		for _, pod := range podList.Items {
			ns := pod.Namespace
			if ns == "" {
				continue
			}
			a, ok := agg[ns]
			if !ok {
				a = &nsAgg{Name: ns}
				agg[ns] = a
			}
			// Pod 数统计该 ns 全部 Pod(含 Completed)
			a.PodCount++
			if pod.Status.Phase != corev1.PodRunning && pod.Status.Phase != corev1.PodPending {
				continue
			}
			a.RunningPods++
			for _, container := range pod.Spec.Containers {
				if req, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
					a.CPU.Add(req)
				}
				if req, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
					a.Mem.Add(req)
				}
			}
		}
	}

	// 转为切片并按 CPU 请求降序排序
	items := make([]map[string]any, 0, len(agg))
	for _, a := range agg {
		items = append(items, map[string]any{
			"name":          a.Name,
			"pod_count":     a.PodCount,
			"running_pods":  a.RunningPods,
			"cpu_used":      float64(a.CPU.MilliValue()) / 1000.0,
			"mem_used":      float64(a.Mem.Value()) / (1024 * 1024 * 1024),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i]["cpu_used"].(float64) > items[j]["cpu_used"].(float64)
	})

	data := map[string]any{
		"total_cpu":  float64(totalCPU.MilliValue()) / 1000.0,
		"total_mem":  float64(totalMem.Value()) / (1024 * 1024 * 1024),
		"namespaces": items,
	}
	response.Success(c, "获取命名空间资源占用成功", data)
}

// 重启异常阈值:restartCount 达到此值视为异常
const restartThreshold = 10

// 异常容器 waiting reason 集合(CrashLoop/镜像/创建失败等)
var abnormalWaitingReasons = map[string]bool{
	"CrashLoopBackOff":            true,
	"ImagePullBackOff":            true,
	"ErrImagePull":                true,
	"InvalidImageName":            true,
	"CreateContainerConfigError":  true,
	"CreateContainerError":        true,
	"RunContainerError":           true,
	"ContainerCreating":           true,
	"OOMKilled":                   true,
}

// Health
//
//	@Description: 获取集群健康快照(异常Pod/重启Pod/NotReady节点/压力节点/异常PVC + summary),仅用原生 API
//	@receiver d
//	@param c
func (d *dashboard) Health(c *gin.Context) {
	var query params.DashboardQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, fmt.Sprintf("参数校验失败:%s", err.Error()))
		return
	}

	clusters, err := getTargetClusters(query.ClusterID)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取集群列表失败:%s", err.Error()))
		return
	}

	type abnormalPod struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Phase     string `json:"phase"`
		Reason    string `json:"reason"`
		Node      string `json:"node"`
	}
	type restartingPod struct {
		Name         string `json:"name"`
		Namespace    string `json:"namespace"`
		RestartCount int    `json:"restart_count"`
		Node         string `json:"node"`
	}
	type pressureNode struct {
		Name      string   `json:"name"`
		Pressures []string `json:"pressures"`
	}
	type abnormalPVC struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		Phase     string `json:"phase"`
	}

	var abnormalPods []abnormalPod
	var restartingPods []restartingPod
	var notReadyNodes []string
	var pressureNodes []pressureNode
	var abnormalPVCs []abnormalPVC

	healthyPods, abnormalPodCount := 0, 0
	readyNodes, notReadyNodeCount := 0, 0
	boundPVCs, abnormalPVCCount := 0, 0

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

		// Pods
		podList, err := client.CoreV1().Pods(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		if err == nil {
			for _, pod := range podList.Items {
				phase := string(pod.Status.Phase)
				reason := ""
				abnormal := false

				// Pending / Failed / Unknown 视为异常
				if pod.Status.Phase != corev1.PodRunning && pod.Status.Phase != corev1.PodSucceeded {
					abnormal = true
					reason = phase
				}
				// 容器 waiting reason 命中异常集合
				for _, cs := range pod.Status.ContainerStatuses {
					if cs.State.Waiting != nil && abnormalWaitingReasons[cs.State.Waiting.Reason] {
						abnormal = true
						if reason == "" {
							reason = cs.State.Waiting.Reason
						}
					}
					// 重启异常
					if cs.RestartCount >= restartThreshold {
						restartingPods = append(restartingPods, restartingPod{
							Name:         pod.Name,
							Namespace:    pod.Namespace,
							RestartCount: int(cs.RestartCount),
							Node:         pod.Spec.NodeName,
						})
					}
				}

				if abnormal {
					abnormalPodCount++
					if len(abnormalPods) < 5 {
						abnormalPods = append(abnormalPods, abnormalPod{
							Name: pod.Name, Namespace: pod.Namespace, Phase: phase, Reason: reason, Node: pod.Spec.NodeName,
						})
					}
				} else if pod.Status.Phase == corev1.PodRunning {
					healthyPods++
				}
			}
		}

		// Nodes
		nodeList, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err == nil {
			for _, node := range nodeList.Items {
				ready := false
				var pressures []string
				for _, cond := range node.Status.Conditions {
					if cond.Type == corev1.NodeReady {
						ready = (cond.Status == corev1.ConditionTrue)
					} else if cond.Status == corev1.ConditionTrue {
						switch cond.Type {
						case corev1.NodeMemoryPressure, corev1.NodeDiskPressure, corev1.NodePIDPressure, corev1.NodeNetworkUnavailable:
							pressures = append(pressures, string(cond.Type))
						}
					}
				}
				if ready {
					readyNodes++
				} else {
					notReadyNodeCount++
					notReadyNodes = append(notReadyNodes, node.Name)
				}
				if len(pressures) > 0 {
					pressureNodes = append(pressureNodes, pressureNode{Name: node.Name, Pressures: pressures})
				}
			}
		}

		// PVCs
		pvcList, err := client.CoreV1().PersistentVolumeClaims(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		if err == nil {
			for _, pvc := range pvcList.Items {
				phase := string(pvc.Status.Phase)
				if pvc.Status.Phase == corev1.ClaimBound {
					boundPVCs++
				} else {
					abnormalPVCCount++
					if len(abnormalPVCs) < 5 {
						abnormalPVCs = append(abnormalPVCs, abnormalPVC{Name: pvc.Name, Namespace: pvc.Namespace, Phase: phase})
					}
				}
			}
		}
	}

	// 截断重启 Pod 清单到 5 条
	if len(restartingPods) > 5 {
		restartingPods = restartingPods[:5]
	}

	data := map[string]any{
		"summary": map[string]any{
			"healthy_pods":    healthyPods,
			"abnormal_pods":   abnormalPodCount,
			"ready_nodes":     readyNodes,
			"not_ready_nodes": notReadyNodeCount,
			"bound_pvcs":      boundPVCs,
			"abnormal_pvcs":   abnormalPVCCount,
		},
		"abnormal_pods":    abnormalPods,
		"restarting_pods":  restartingPods,
		"not_ready_nodes":  notReadyNodes,
		"pressure_nodes":   pressureNodes,
		"abnormal_pvcs":    abnormalPVCs,
	}
	response.Success(c, "获取集群健康快照成功", data)
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
