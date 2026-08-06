package cluster

import (
	"context"
	"fmt"
	"time"

	clusterModel "gkube/internal/cluster/model"
	k8sNode "gkube/pkg/k8s/node"
	"gkube/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetClusterVersion
//
//	@Description: 获取集群版本信息
//	@param client
//	@return string
//	@return error
func GetClusterVersion(client *kubernetes.Clientset) (string, error) {
	version, err := client.ServerVersion()
	if err != nil {
		return "", err
	}
	return version.String(), nil
}

// GetClusterNodesInfo
//
//	@Description: 获取集群节点信息
//	@param client
//	@return []clusterModel.NodeInfo
//	@return error
func GetClusterNodesInfo(client *kubernetes.Clientset) ([]clusterModel.NodeInfo, error) {
	nodes, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// 获取每个节点的 Pod 数量，并按节点累加 Pod 资源请求（仅 Running/Pending）
	podCounts := make(map[string]int)
	type nodeRequests struct {
		cpu resource.Quantity
		mem resource.Quantity
	}
	nodeReqs := make(map[string]nodeRequests)
	pods, podListErr := client.CoreV1().Pods(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if podListErr != nil {
		// 不静默吞掉：失败时记日志，下游 PodCount/CPUUsed/MemUsed 保持 0，
		// 调用方/前端至少能从日志查到根因，而非误以为节点真的空闲。
		logger.Error(fmt.Sprintf("获取集群 Pod 列表失败，节点使用量统计将为 0: %s", podListErr.Error()))
	} else {
		for _, pod := range pods.Items {
			if pod.Spec.NodeName == "" {
				continue
			}
			if pod.Status.Phase != corev1.PodSucceeded && pod.Status.Phase != corev1.PodFailed {
				podCounts[pod.Spec.NodeName]++
			}
			if pod.Status.Phase != corev1.PodRunning && pod.Status.Phase != corev1.PodPending {
				continue
			}
			reqs := nodeReqs[pod.Spec.NodeName]
			for _, container := range pod.Spec.Containers {
				if req, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
					reqs.cpu.Add(req)
				}
				if req, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
					reqs.mem.Add(req)
				}
			}
			nodeReqs[pod.Spec.NodeName] = reqs
		}
	}

	var nodesInfo []clusterModel.NodeInfo

	for _, node := range nodes.Items {
		status, isReady := k8sNode.NodeStatus(node.Status.Conditions)
		roles := k8sNode.NodeRoles(node.Labels)

		// Get addresses
		var internalIP, externalIP string
		for _, addr := range node.Status.Addresses {
			switch addr.Type {
			case corev1.NodeInternalIP:
				internalIP = addr.Address
			case corev1.NodeExternalIP:
				externalIP = addr.Address
			}
		}

		nodeInfo := clusterModel.NodeInfo{
			Name:             node.Name,
			Status:           status,
			Roles:            roles,
			Version:          node.Status.NodeInfo.KubeletVersion,
			InternalIP:       internalIP,
			ExternalIP:       externalIP,
			Architecture:     node.Status.NodeInfo.Architecture,
			Unschedulable:    node.Spec.Unschedulable,
			PodCount:         podCounts[node.Name],
			Labels:           node.Labels,
			Taints:           node.Spec.Taints,
			IsReady:          isReady,
			Addresses:        node.Status.Addresses,
			OSImage:          node.Status.NodeInfo.OSImage,
			KernelVersion:    node.Status.NodeInfo.KernelVersion,
			ContainerRuntime: node.Status.NodeInfo.ContainerRuntimeVersion,
			CreationTimestamp: node.CreationTimestamp.Format(time.RFC3339),
		}

		// 进度条用数字字段：已请求 / 可分配（口径为调度压力，非真实负载）
		if allocCPU, ok := node.Status.Allocatable[corev1.ResourceCPU]; ok {
			nodeInfo.CPUTotal = float64(allocCPU.MilliValue()) / 1000.0
		}
		if allocMem, ok := node.Status.Allocatable[corev1.ResourceMemory]; ok {
			nodeInfo.MemTotal = float64(allocMem.Value()) / (1024 * 1024 * 1024)
		}
		if podsCap, ok := node.Status.Capacity[corev1.ResourcePods]; ok {
			nodeInfo.PodTotal = int(podsCap.Value())
		}
		if reqs, ok := nodeReqs[node.Name]; ok {
			nodeInfo.CPUUsed = float64(reqs.cpu.MilliValue()) / 1000.0
			nodeInfo.MemUsed = float64(reqs.mem.Value()) / (1024 * 1024 * 1024)
		}

		nodesInfo = append(nodesInfo, nodeInfo)
	}

	return nodesInfo, nil
}
