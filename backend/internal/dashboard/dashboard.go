package dashboard

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gkube/internal/cluster/model"
	"gkube/pkg/database"
	"gkube/pkg/k8s"
	"gkube/pkg/logger"
	"gkube/pkg/response"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DashboardQueryParams 仪表盘聚合查询参数(overview/resources/workloads 共用)。
// ClusterID 命中时仅查询该集群;为空时聚合所有 online 集群(向后兼容)。
type DashboardQueryParams struct {
	ClusterID *uint `form:"clusterId" json:"clusterId" label:"集群ID"`
}

type EventQueryParams struct {
	ClusterID     *uint  `form:"clusterId" json:"clusterId" label:"集群ID"`
	Namespace     string `form:"namespace" json:"namespace" label:"命名空间"`
	Type          string `form:"type" json:"type" label:"事件类型"`
	FieldSelector string `form:"fieldSelector" json:"fieldSelector" label:"字段选择器"`
	Limit         int    `form:"limit" json:"limit" label:"数量限制"`
	Continue      string `form:"continue" json:"continue" label:"分页标记"`
}

type dashboard struct{}

var Dashboard = new(dashboard)

const (
	dashboardTimeout      = 30 * time.Second
	maxConcurrentClusters = 5
	// maxHealthEntities 每集群最多返回的异常实体条目数(abnormalPods/restartingPods/abnormalPVCs)
	maxHealthEntities = 5
)

// dashboardCtx 派生请求级带超时的 context。
func dashboardCtx(c *gin.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Request.Context(), dashboardTimeout)
}

// forEachCluster 并发(信号量限流)遍历集群执行 fn。
// fn 内对共享状态的写操作需自行加锁——调用方负责声明并传递自己的 sync.Mutex。
// 单个集群出错被忽略,与既有行为一致(返回可得的聚合结果)。
func forEachCluster(c *gin.Context, clusters []model.K8SCluster, fn func(ctx context.Context, cluster model.K8SCluster)) {
	ctx, cancel := dashboardCtx(c)
	defer cancel()
	sem := make(chan struct{}, maxConcurrentClusters)
	var wg sync.WaitGroup
	for _, cl := range clusters {
		cl := cl
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			fn(ctx, cl)
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

// sumNodeAllocatable 累加节点列表的 Allocatable CPU 和内存总量。
func sumNodeAllocatable(nodes []corev1.Node) (cpu, mem resource.Quantity) {
	for _, node := range nodes {
		if allocCPU, ok := node.Status.Allocatable[corev1.ResourceCPU]; ok {
			cpu.Add(allocCPU)
		}
		if allocMem, ok := node.Status.Allocatable[corev1.ResourceMemory]; ok {
			mem.Add(allocMem)
		}
	}
	return
}

// sumNodeStorage 累加节点列表的存储总量(Allocatable 优先,回退到 Capacity)。
func sumNodeStorage(nodes []corev1.Node) resource.Quantity {
	var total resource.Quantity
	for _, node := range nodes {
		if stor, ok := node.Status.Allocatable[corev1.ResourceEphemeralStorage]; ok {
			total.Add(stor)
		} else if stor, ok := node.Status.Capacity[corev1.ResourceEphemeralStorage]; ok {
			total.Add(stor)
		}
	}
	return total
}

// Overview 获取仪表盘概览数据
func (d *dashboard) Overview(c *gin.Context) {
	var query DashboardQueryParams
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
	var mu sync.Mutex

	forEachCluster(c, clusters, func(ctx context.Context, cluster model.K8SCluster) {
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
		namespaceCount += localNS
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
	var query DashboardQueryParams
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

	forEachCluster(c, clusters, func(ctx context.Context, cluster model.K8SCluster) {
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
		cpuTotal, memTotal := sumNodeAllocatable(nodeList.Items)
		storTotal := sumNodeStorage(nodeList.Items)

		// 先登记集群容量(即使后续 podList 失败也不丢失总量)
		mu.Lock()
		totalCPUTotal.Add(cpuTotal)
		totalMemTotal.Add(memTotal)
		totalStorageTotal.Add(storTotal)
		mu.Unlock()

		var cpuUsed, memUsed, storUsed resource.Quantity

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
		totalMemUsed.Add(memUsed)
		totalStorageUsed.Add(storUsed)
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
	var query DashboardQueryParams
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

	forEachCluster(c, clusters, func(ctx context.Context, cluster model.K8SCluster) {
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
	var query DashboardQueryParams
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

	forEachCluster(c, clusters, func(ctx context.Context, cluster model.K8SCluster) {
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
			localCPU, localMem = sumNodeAllocatable(nodeList.Items)
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

// 异常容器 waiting reason 集合（不包含过渡态 ContainerCreating，避免假阳性）。
var abnormalWaitingReasons = map[string]bool{
	"CrashLoopBackOff":           true,
	"ImagePullBackOff":           true,
	"ErrImagePull":               true,
	"InvalidImageName":           true,
	"CreateContainerConfigError": true,
	"CreateContainerError":       true,
	"RunContainerError":          true,
	"OOMKilled":                  true,
}

// Health 获取集群健康快照
func (d *dashboard) Health(c *gin.Context) {
	var query DashboardQueryParams
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

	forEachCluster(c, clusters, func(ctx context.Context, cluster model.K8SCluster) {
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
						if len(lRestartPods) < maxHealthEntities {
							lRestartPods = append(lRestartPods, restartingPod{
								Name:         pod.Name,
								Namespace:    pod.Namespace,
								RestartCount: int(cs.RestartCount),
								Node:         pod.Spec.NodeName,
							})
						}
					}
				}

				if abnormal {
					lAbnPod++
					if len(lAbnPods) < maxHealthEntities {
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
					if len(lAbnPVCs) < maxHealthEntities {
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

// eventInfo 从 K8s Event 提取的展示字段
type eventInfo struct {
	Type               string `json:"type"`
	Reason             string `json:"reason"`
	Message            string `json:"message"`
	Namespace          string `json:"namespace"`
	InvolvedObject     string `json:"involved_object"`
	InvolvedObjectKind string `json:"involved_object_kind"`
	InvolvedObjectName string `json:"involved_object_name"`
	FirstSeen          string `json:"first_seen"`
	LastSeen           string `json:"last_seen"`
	Count              int32  `json:"count"`
	ReportingComponent string `json:"reporting_component"`
	ReportingInstance  string `json:"reporting_instance"`
	Action             string `json:"action"`
	ClusterName        string `json:"cluster_name"`
	eventName          string // 内部排序键,不导出到 JSON
}

// toEventInfo 将 K8s Event 转为展示结构
func toEventInfo(event corev1.Event, clusterName string) eventInfo {
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
	return eventInfo{
		Type:               event.Type,
		Reason:             event.Reason,
		Message:            event.Message,
		Namespace:          event.Namespace,
		InvolvedObject:     fmt.Sprintf("%s/%s", event.InvolvedObject.Kind, event.InvolvedObject.Name),
		InvolvedObjectKind: event.InvolvedObject.Kind,
		InvolvedObjectName: event.InvolvedObject.Name,
		FirstSeen:          firstSeen,
		LastSeen:           lastSeen,
		Count:              event.Count,
		ReportingComponent: event.ReportingController,
		ReportingInstance:  event.ReportingInstance,
		Action:             event.Action,
		ClusterName:        clusterName,
		eventName:          event.Name,
	}
}

// Events 获取集群事件列表（支持分页，使用 K8s 原生 continue token 避免全量拉取）。
// 单集群模式:支持 type 过滤（下推到 K8s fieldSelector）和 continue 分页。
// 多集群模式:各集群取最近 limit 条合并返回,不支持翻页（K8s continue token 是 per-cluster 的）。
func (d *dashboard) Events(c *gin.Context) {
	var query EventQueryParams
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
	clusters, err := getTargetClusters(query.ClusterID)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusInternalServerError, "获取集群列表失败")
		return
	}
	if len(clusters) == 0 {
		response.Success(c, "获取事件列表成功", map[string]any{
			"items":    []eventInfo{},
			"total":    0,
			"continue": "",
			"has_more": false,
		})
		return
	}

	namespace := query.Namespace
	if namespace == "" {
		namespace = corev1.NamespaceAll
	}

	// 构建 fieldSelector:将 type 过滤下推到 K8s API,避免分页不准
	fs := query.FieldSelector
	if query.Type != "" {
		typeSelector := fmt.Sprintf("type=%s", query.Type)
		if fs != "" {
			fs = fs + "," + typeSelector
		} else {
			fs = typeSelector
		}
	}

	// 单集群模式:直接查询该集群,支持 K8s 原生 continue 分页。
	if query.ClusterID != nil {
		cluster := clusters[0]
		if cluster.Status != "online" {
			response.Success(c, "获取事件列表成功", map[string]any{
				"items": []eventInfo{}, "total": 0, "continue": "", "has_more": false,
			})
			return
		}
		client, err := k8s.GetK8sClientByName(cluster.ClusterName)
		if err != nil {
			logger.Error(err.Error())
			response.FailWithStatus(c, http.StatusInternalServerError, "获取事件列表失败")
			return
		}

		listOpts := metav1.ListOptions{Limit: int64(query.Limit)}
		if fs != "" {
			listOpts.FieldSelector = fs
		}
		if query.Continue != "" {
			listOpts.Continue = query.Continue
		}

		ctx, cancel := dashboardCtx(c)
		defer cancel()
		eventList, err := client.CoreV1().Events(namespace).List(ctx, listOpts)
		if err != nil {
			response.FailWithStatus(c, http.StatusInternalServerError, "获取事件列表失败")
			return
		}

		items := make([]eventInfo, 0, len(eventList.Items))
		for _, event := range eventList.Items {
			items = append(items, toEventInfo(event, cluster.ClusterName))
		}

		// 稳定排序(时间相同时按事件名做第二键)
		sort.SliceStable(items, func(i, j int) bool {
			if items[i].LastSeen != items[j].LastSeen {
				return items[i].LastSeen > items[j].LastSeen
			}
			return items[i].eventName > items[j].eventName
		})

		k8sContinue := eventList.ListMeta.Continue
		response.Success(c, "获取事件列表成功", map[string]any{
			"items":    items,
			"total":    len(items),
			"continue": k8sContinue,
			"has_more": k8sContinue != "",
		})
		return
	}

	// 多集群模式:各集群取最近 limit 条,合并后排序。
	// 注意:多集群不支持翻页（K8s continue token 是 per-cluster 的,无法跨集群统一）。
	var mu sync.Mutex
	var allEvents []eventInfo

	forEachCluster(c, clusters, func(ctx context.Context, cluster model.K8SCluster) {
		if cluster.Status != "online" {
			return
		}
		client, err := k8s.GetK8sClientByName(cluster.ClusterName)
		if err != nil {
			return
		}

		listOpts := metav1.ListOptions{Limit: int64(query.Limit)}
		if fs != "" {
			listOpts.FieldSelector = fs
		}

		eventList, err := client.CoreV1().Events(namespace).List(ctx, listOpts)
		if err != nil {
			return
		}

		local := make([]eventInfo, 0, len(eventList.Items))
		for _, event := range eventList.Items {
			local = append(local, toEventInfo(event, cluster.ClusterName))
		}
		mu.Lock()
		allEvents = append(allEvents, local...)
		mu.Unlock()
	})

	sort.SliceStable(allEvents, func(i, j int) bool {
		if allEvents[i].LastSeen != allEvents[j].LastSeen {
			return allEvents[i].LastSeen > allEvents[j].LastSeen
		}
		return allEvents[i].eventName > allEvents[j].eventName
	})

	response.Success(c, "获取事件列表成功", map[string]any{
		"items":    allEvents,
		"total":    len(allEvents),
		"continue": "",
		"has_more": false,
	})
}