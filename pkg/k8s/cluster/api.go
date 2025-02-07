package cluster

import (
	"context"
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	modelK8s "gkube/models/k8s"
)

func GetClusterVersion(client *kubernetes.Clientset) (string, error) {
	version, err := client.ServerVersion()
	if err != nil {
		return "", err
	}
	return version.String(), nil
}

func GetClusterNodesInfo(client *kubernetes.Clientset)([]modelK8s.NodeInfo,error) {
	nodes, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil,err
	}
	var nodesInfo []modelK8s.NodeInfo

	// 遍历节点并提取信息
	for _, node := range nodes.Items {
		nodeInfo := modelK8s.NodeInfo{
			Name:            node.Name,
			Labels:          node.Labels,
			Addresses:       node.Status.Addresses,
			OSImage:         node.Status.NodeInfo.OSImage,
			KernelVersion:   node.Status.NodeInfo.KernelVersion,
			ContainerRuntime: node.Status.NodeInfo.ContainerRuntimeVersion,
		}

		// 检查节点是否 Ready
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady {
				nodeInfo.IsReady = condition.Status == corev1.ConditionTrue
				break
			}
		}
		// 提取资源容量和可分配资源
        if cpu, ok := node.Status.Capacity[corev1.ResourceCPU]; ok {
            nodeInfo.CapacityCPU = cpu.String()
        }
        if mem, ok := node.Status.Capacity[corev1.ResourceMemory]; ok {
            nodeInfo.CapacityMemory = formatMemory(mem)
        }
        if cpu, ok := node.Status.Allocatable[corev1.ResourceCPU]; ok {
            nodeInfo.AllocatableCPU = cpu.String()
        }
        if mem, ok := node.Status.Allocatable[corev1.ResourceMemory]; ok {
            nodeInfo.AllocatableMem = formatMemory(mem)
        }
		// 将节点信息添加到数组
		nodesInfo = append(nodesInfo, nodeInfo)
	}

	return nodesInfo, nil
}

