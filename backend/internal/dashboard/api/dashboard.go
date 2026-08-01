package api

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gkube/internal/cluster/model"
	"gkube/internal/dashboard/params"
	"gkube/pkg/database"
	"gkube/pkg/k8s"
	"gkube/pkg/logger"
	"gkube/pkg/response"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type dashboard struct{}

var Dashboard = new(dashboard)

const (
	dashboardTimeout      = 30 * time.Second
	maxConcurrentClusters = 5
)

// dashboardCtx 派生请求级带超时的 context。
func dashboardCtx(c *gin.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Request.Context(), dashboardTimeout)
}

// forEachCluster 并发(信号量限流)遍历集群执行 fn,fn 内对共享状态的写操作需自行加锁(mu)。
// 单个集群出错被忽略,与既有行为一致(返回可得的聚合结果)。
func forEachCluster(c *gin.Context, clusters []model.K8SCluster, fn func(ctx context.Context, cluster model.K8SCluster, mu *sync.Mutex)) {
	ctx, cancel := dashboardCtx(c)
	defer cancel()
	sem := make(chan struct{}, maxConcurrentClusters)
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, cl := range clusters {
		cl := cl
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			fn(ctx, cl, &mu)
		}()
	}
	wg.Wait()
}

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

// Overview 获取仪表盘概览数据
func (d *dashboard) Overview(c *gin.Context) {
	var query params.DashboardQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	clusters, err := getTargetClusters(query.ClusterID)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusInternalServerError, "获取集群列表失败")
		return
	}

	clusterCount := len(clusters)
	var nodeCount, podCount, namespaceCount int

	forEachCluster(c, clusters, func(ctx context.Context, cluster model.K8SCluster, mu *sync.Mutex) {
		mu.Lock()
		nodeCount += cluster.NodeCount
		mu.Unlock()
		if cluster.Status != "online" {
			return
		}
		// 复用缓存客户端
		client, err := k8s.GetK8sClientByName(cluster.ClusterName)
		if err != nil {
			return
		}

		var localPod, localNS int
		// Count pods
		if podList, err := client.CoreV1().Pods(corev1.NamespaceAll).List(ctx, metav1.ListOptions{}); err == nil {
			localPod = len(podList.Items)
		}
		// Count namespaces
		if nsList, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{}); err == nil {
			localNS = len(nsList.Items)
		}
		mu.Lock()
		podCount += localPod
		if localNS > namespaceCount {
			namespaceCount = localNS
		}
		mu.Unlock()
	})

	data := map[string]any{
		"cluster_count":   clusterCount,
		"node_count":      nodeCount,
		"pod_count":       podCount,
		"namespace_count": namespaceCount,
	}
	response.Success(c, "获取概览数据成功", data)
}

// Resources 获取集群资源使用情况（CPU/内存/存储）
func (d *dashboard) Resources(c *gin.Context) {
	var query params.DashboardQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	clusters, err := getTargetClusters(query.ClusterID)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusInternalServerError, "获取集群列表失败")
		return
	}

	var mu sync.Mutex
	var totalCPUUsed, totalCPUTotal resource.Quantity
	var totalMemUsed, totalMemTotal resource.Quantity
	var totalStorageUsed, totalStorageTotal resource.Quantity

	forEachCluster(c, clusters, func(ctx context.Context, cluster model.K8SCluster, _ *sync.Mutex) {
		if cluster.Status != "online" {
			return
		}
		client, err := k8s.GetK8sClientByName(cluster.ClusterName)
		if err != nil {
			return
		}

		// 总量取 Allocatable(真实可调度)
		nodeList, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			return
		}
		var cpuTotal, memTotal, storTotal resource.Quantity
		var cpuUsed, memUsed, storUsed resource.Quantity
		for _, node := range nodeList.Items {
			if allocCPU, ok := node.Status.Allocatable[corev1.ResourceCPU]; ok {
				cpuTotal.Add(allocCPU)
			}
			if allocMem, ok := node.Status.Allocatable[corev1.ResourceMemory]; ok {
				memTotal.Add(allocMem)
			}
			if allocStor := node.Status.Allocatable.StorageEphemeral(); !allocStor.IsZero() {
				storTotal.Add(*allocStor)
			} else if storageCap := node.Status.Capacity.StorageEphemeral(); !storageCap.IsZero() {
				storTotal.Add(*storageCap)
			}
		}

		// Sum pod resource requests as "used"
		podList, err := client.CoreV1().Pods(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
		if err != nil {
			return
		}
		for _, pod := range podList.Items {
			if pod.Status.Phase != corev1.PodRunning && pod.Status.Phase != corev1.PodPending {
				continue
			}
			for _, container := range pod.Spec.Containers {
				if req, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
					cpuUsed.Add(req)
				}
				if req, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
					memUsed.Add(req)
				}
				if req, ok := container.Resources.Requests[corev1.ResourceEphemeralStorage]; ok {
					storUsed.Add(req)
				}
			}
		}
		mu.Lock()
		totalCPUUsed.Add(cpuUsed)
		totalCPUTotal.Add(cpuTotal)
		totalMemUsed.Add(memUsed)
		totalMemTotal.Add(memTotal)
		totalStorageUsed.Add(storUsed)
		totalStorageTotal.Add(storTotal)
		mu.Unlock()
	})

	data := map[string]any{
		"cpu": map[string]any{
			"used":  float64(totalCPUUsed.MilliValue()) / 1000.0,
			"total": float64(totalCPUTotal.MilliValue()) / 1000.0,
		},
		"memory": map[string]any{
			"used":  float64(totalMemUsed.Value()) / (1024 * 1024 * 1024),
			"total": float64(totalMemTotal.Value()) / (1024 * 1024 * 1024),
		},
		"storage": map[string]any{
			"used":  float64(totalStorageUsed.Value()) / (1024 * 1024 * 1024),
			"total": float64(totalStorageTotal.Value()) / (1024 * 1024 * 1024),
		},
	}
	response.Success(c, "获取资源信息成功", data)
}

// Workloads 获取所有集群工作负载统计
func (d *dashboard) Workloads(c *gin.Context) {
	var query params.DashboardQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	clusters, err := getTargetClusters(query.ClusterID)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusInternalServerError, "获取集群列表失败")
		return
	}

	var mu sync.Mutex
	var totalDeployments, totalStatefulSets, totalDaemonSets, totalJobs, totalCronJobs, totalIngresses int

	forEachCluster(c, clusters, func(ctx context.Context, cluster model.K8SCluster, _ *sync.Mutex) {
		if cluster.Status != "online" {
			return
		}
		client, err := k8s.GetK8sClientByName(cluster.ClusterName)
		if err != nil {
			return
		}

		var d, ss, ds, j, cj, ing int
		if deployments, err := client.AppsV1().Deployments(corev1.NamespaceAll).List(ctx, metav1.ListOptions{}); err == nil {
			d = len(deployments.Items)
		}
		if statefulSets, err := client.AppsV1().StatefulSets(corev1.NamespaceAll).List(ctx, metav1.ListOptions{}); err == nil {
			ss = len(statefulSets.Items)
		}
		if daemonSets, err := client.AppsV1().DaemonSets(corev1.NamespaceAll).List(ctx, metav1.ListOptions{}); err == nil {
			ds = len(daemonSets.Items)
		}
		if jobs, err := client.BatchV1().Jobs(corev1.NamespaceAll).List(ctx, metav1.ListOptions{}); err == nil {
			j = len(jobs.Items)
		}
		if cronJobs, err := client.BatchV1().CronJobs(corev1.NamespaceAll).List(ctx, metav1.ListOptions{}); err == nil {
			cj = len(cronJobs.Items)
		}
		if ingresses, err := client.NetworkingV1().Ingresses(corev1.NamespaceAll).List(ctx, metav1.ListOptions{}); err == nil {
			ing = len(ingresses.Items)
		}
		mu.Lock()
		totalDeployments += d
		totalStatefulSets += ss
		totalDaemonSets += ds
		totalJobs += j
		totalCronJobs += cj
		totalIngresses += ing
		mu.Unlock()
	})

	data := map[string]any{
		"deployments":  totalDeployments,
		"statefulsets": totalStatefulSets,
		"daemonsets":   totalDaemonSets,
		"jobs":         totalJobs,
		"cronjobs":     totalCronJobs,
		"ingresses":    totalIngresses,
	}
	response.Success(c, "获取工作负载信息成功", data)
}

// Namespaces 获取按命名空间聚合的资源占用
func (d *dashboard) Namespaces(c *gin.Context) {
	var query params.DashboardQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	clusters, err := getTargetClusters(query.ClusterID)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusInternalServerError, "获取集群列表失败")
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
	var mu sync.Mutex
	var totalCPU, totalMem resource.Quantity

	forEachCluster(c, clusters, func(ctx context.Context, cluster model.K8SCluster, _ *sync.Mutex) {
		if cluster.Status != "online" {
			return
		}
		client, err := k8s.GetK8sClientByName(cluster.ClusterName)
		if err != nil {
			return
		}

		// 集群总量 = 所有节点 Allocatable 之和
		var localCPU, localMem resource.Quantity
		if nodeList, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{}); err == nil {
			for _, node := range nodeList.Items {
				if allocCPU, ok := node.Status.Allocatable[corev1.ResourceCPU]; ok {
					localCPU.Add(allocCPU)
				}
				if allocMem, ok := node.Status.Allocatable[corev1.ResourceMemory]; ok {
					localMem.Add(allocMem)
				}
			}
		}

		// 按命名空间累加
		localAgg := make(map[string]*nsAgg)
		podList, err := client.CoreV1().Pods(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, pod := range podList.Items {
				ns := pod.Namespace
				if ns == "" {
					continue
				}
				a, ok := localAgg[ns]
				if !ok {
					a = &nsAgg{Name: ns}
					localAgg[ns] = a
				}
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

		// 补齐没有 Pod 的命名空间
		if nsList, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{}); err == nil {
			for _, ns := range nsList.Items {
				if ns.Name == "" {
					continue
				}
				if _, ok := localAgg[ns.Name]; !ok {
					localAgg[ns.Name] = &nsAgg{Name: ns.Name}
				}
			}
		}

		// 合并到全局 agg
		mu.Lock()
		totalCPU.Add(localCPU)
		totalMem.Add(localMem)
		for name, a := range localAgg {
			g, ok := agg[name]
			if !ok {
				g = &nsAgg{Name: name}
				agg[name] = g
			}
			g.PodCount += a.PodCount
			g.RunningPods += a.RunningPods
			g.CPU.Add(a.CPU)
			g.Mem.Add(a.Mem)
		}
		mu.Unlock()
	})

	// 转为切片并按 CPU 请求降序排序
	items := make([]map[string]any, 0, len(agg))
	for _, a := range agg {
		items = append(items, map[string]any{
			"name":         a.Name,
			"pod_count":    a.PodCount,
			"running_pods": a.RunningPods,
			"cpu_used":     float64(a.CPU.MilliValue()) / 1000.0,
			"mem_used":     float64(a.Mem.Value()) / (1024 * 1024 * 1024),
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

// 异常容器 waiting reason 集合
var abnormalWaitingReasons = map[string]bool{
	"CrashLoopBackOff":           true,
	"ImagePullBackOff":           true,
	"ErrImagePull":               true,
	"InvalidImageName":           true,
	"CreateContainerConfigError": true,
	"CreateContainerError":       true,
	"RunContainerError":          true,
	"ContainerCreating":          true,
	"OOMKilled":                  true,
}

// Health 获取集群健康快照
func (d *dashboard) Health(c *gin.Context) {
	var query params.DashboardQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	clusters, err := getTargetClusters(query.ClusterID)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusInternalServerError, "获取集群列表失败")
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

	var mu sync.Mutex
	var abnormalPods []abnormalPod
	var restartingPods []restartingPod
	var notReadyNodes []string
	var pressureNodes []pressureNode
	var abnormalPVCs []abnormalPVC

	healthyPods, abnormalPodCount := 0, 0
	readyNodes, notReadyNodeCount := 0, 0
	boundPVCs, abnormalPVCCount := 0, 0

	forEachCluster(c, clusters, func(ctx context.Context, cluster model.K8SCluster, _ *sync.Mutex) {
		if cluster.Status != "online" {
			return
		}
		client, err := k8s.GetK8sClientByName(cluster.ClusterName)
		if err != nil {
			return
		}

		var lAbnPods []abnormalPod
		var lRestartPods []restartingPod
		var lNotReadyNodes []string
		var lPressureNodes []pressureNode
		var lAbnPVCs []abnormalPVC
		lHealthy, lAbnPod := 0, 0
		lReady, lNotReady := 0, 0
		lBound, lAbnPVC := 0, 0

		// Pods
		if podList, err := client.CoreV1().Pods(corev1.NamespaceAll).List(ctx, metav1.ListOptions{}); err == nil {
			for _, pod := range podList.Items {
				phase := string(pod.Status.Phase)
				reason := ""
				abnormal := false

				if pod.Status.Phase != corev1.PodRunning && pod.Status.Phase != corev1.PodSucceeded {
					abnormal = true
					reason = phase
				}
				for _, cs := range pod.Status.ContainerStatuses {
					if cs.State.Waiting != nil && abnormalWaitingReasons[cs.State.Waiting.Reason] {
						abnormal = true
						if reason == "" {
							reason = cs.State.Waiting.Reason
						}
					}
					if cs.RestartCount >= restartThreshold {
						lRestartPods = append(lRestartPods, restartingPod{
							Name:         pod.Name,
							Namespace:    pod.Namespace,
							RestartCount: int(cs.RestartCount),
							Node:         pod.Spec.NodeName,
						})
					}
				}

				if abnormal {
					lAbnPod++
					if len(lAbnPods) < 5 {
						lAbnPods = append(lAbnPods, abnormalPod{
							Name: pod.Name, Namespace: pod.Namespace, Phase: phase, Reason: reason, Node: pod.Spec.NodeName,
						})
					}
				} else if pod.Status.Phase == corev1.PodRunning {
					lHealthy++
				}
			}
		}

		// Nodes
		if nodeList, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{}); err == nil {
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
					lReady++
				} else {
					lNotReady++
					lNotReadyNodes = append(lNotReadyNodes, node.Name)
				}
				if len(pressures) > 0 {
					lPressureNodes = append(lPressureNodes, pressureNode{Name: node.Name, Pressures: pressures})
				}
			}
		}

		// PVCs
		if pvcList, err := client.CoreV1().PersistentVolumeClaims(corev1.NamespaceAll).List(ctx, metav1.ListOptions{}); err == nil {
			for _, pvc := range pvcList.Items {
				phase := string(pvc.Status.Phase)
				if pvc.Status.Phase == corev1.ClaimBound {
					lBound++
				} else {
					lAbnPVC++
					if len(lAbnPVCs) < 5 {
						lAbnPVCs = append(lAbnPVCs, abnormalPVC{Name: pvc.Name, Namespace: pvc.Namespace, Phase: phase})
					}
				}
			}
		}

		mu.Lock()
		abnormalPods = append(abnormalPods, lAbnPods...)
		restartingPods = append(restartingPods, lRestartPods...)
		notReadyNodes = append(notReadyNodes, lNotReadyNodes...)
		pressureNodes = append(pressureNodes, lPressureNodes...)
		abnormalPVCs = append(abnormalPVCs, lAbnPVCs...)
		healthyPods += lHealthy
		abnormalPodCount += lAbnPod
		readyNodes += lReady
		notReadyNodeCount += lNotReady
		boundPVCs += lBound
		abnormalPVCCount += lAbnPVC
		mu.Unlock()
	})

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
		"abnormal_pods":   abnormalPods,
		"restarting_pods": restartingPods,
		"not_ready_nodes": notReadyNodes,
		"pressure_nodes":  pressureNodes,
		"abnormal_pvcs":   abnormalPVCs,
	}
	response.Success(c, "获取集群健康快照成功", data)
}

// Events 获取集群事件列表（支持分页）
func (d *dashboard) Events(c *gin.Context) {
	var query params.EventQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}

	// 设置默认限制并加上限
	if query.Limit <= 0 {
		query.Limit = 100
	}
	if query.Limit > 500 {
		query.Limit = 500
	}

	// 获取要查询的集群列表
	var clusters []model.K8SCluster
	if query.ClusterID != nil {
		if err := database.DB.Where("id = ?", *query.ClusterID).Find(&clusters).Error; err != nil {
			logger.Error(err.Error())
			response.FailWithStatus(c, http.StatusInternalServerError, "获取集群信息失败")
			return
		}
	} else {
		if err := database.DB.Where("status = ?", "online").Find(&clusters).Error; err != nil {
			logger.Error(err.Error())
			response.FailWithStatus(c, http.StatusInternalServerError, "获取集群列表失败")
			return
		}
	}

	var mu sync.Mutex
	var allEvents []map[string]any

	forEachCluster(c, clusters, func(ctx context.Context, cluster model.K8SCluster, _ *sync.Mutex) {
		if cluster.Status != "online" {
			return
		}
		client, err := k8s.GetK8sClientByName(cluster.ClusterName)
		if err != nil {
			return
		}

		namespace := query.Namespace
		if namespace == "" {
			namespace = corev1.NamespaceAll
		}

		listOpts := metav1.ListOptions{}
		if query.FieldSelector != "" {
			listOpts.FieldSelector = query.FieldSelector
		}

		eventList, err := client.CoreV1().Events(namespace).List(ctx, listOpts)
		if err != nil {
			return
		}

		var local []map[string]any
		for _, event := range eventList.Items {
			if query.Type != "" && event.Type != query.Type {
				continue
			}

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

			local = append(local, map[string]any{
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
		mu.Lock()
		allEvents = append(allEvents, local...)
		mu.Unlock()
	})

	// 按最后时间倒序排序
	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i]["last_seen"].(string) > allEvents[j]["last_seen"].(string)
	})

	total := len(allEvents)

	offset := 0
	if query.Continue != "" {
		fmt.Sscanf(query.Continue, "%d", &offset)
	}

	end := offset + query.Limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = total
	}

	pagedEvents := allEvents[offset:end]

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
