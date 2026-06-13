package api

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"gkube/pkg/k8s"
	"gkube/pkg/response"
	"k8s.io/client-go/dynamic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type metrics struct{}

var Metrics = new(metrics)

// GetNodeMetrics gets node resource usage from metrics-server
func (m *metrics) GetNodeMetrics(c *gin.Context) {
	clusterName := c.Query("clusterName")
	config, err := k8s.GetRestConfigByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s配置失败:%s", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}

	// Get node metrics from metrics.k8s.io API
	metricsGVR := schema.GroupVersionResource{Group: "metrics.k8s.io", Version: "v1beta1", Resource: "nodes"}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		response.Fail(c, fmt.Sprintf("创建动态客户端失败:%s", err.Error()))
		return
	}
	metricsList, err := dynamicClient.Resource(metricsGVR).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		// Fallback: return node list without metrics
		nodeList, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			response.Fail(c, fmt.Sprintf("获取节点列表失败:%s", err.Error()))
			return
		}
		var result []map[string]any
		for _, node := range nodeList.Items {
			result = append(result, map[string]any{
				"name":     node.Name,
				"cpu":      node.Status.Capacity.Cpu().String(),
				"memory":   node.Status.Capacity.Memory().String(),
				"cpuUsage":    "N/A",
				"memoryUsage": "N/A",
			})
		}
		response.Success(c, "执行成功 (metrics-server not available)", result)
		return
	}

	// Also get node list for capacity info
	nodeList, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取节点列表失败:%s", err.Error()))
		return
	}
	nodeCapacity := make(map[string]map[string]string)
	for _, node := range nodeList.Items {
		nodeCapacity[node.Name] = map[string]string{
			"cpu":    node.Status.Capacity.Cpu().String(),
			"memory": node.Status.Capacity.Memory().String(),
		}
	}

	var result []map[string]any
	for _, item := range metricsList.Items {
		name := item.GetName()
		usage, _, _ := unstructured.NestedMap(item.Object, "usage")
		cap := nodeCapacity[name]
		result = append(result, map[string]any{
			"name":        name,
			"cpu":         cap["cpu"],
			"memory":      cap["memory"],
			"cpuUsage":    usage["cpu"],
			"memoryUsage": usage["memory"],
		})
	}
	response.Success(c, "执行成功", result)
}

// GetPodMetrics gets pod resource usage from metrics-server
func (m *metrics) GetPodMetrics(c *gin.Context) {
	clusterName := c.Query("clusterName")
	namespace := c.Query("namespace")
	config, err := k8s.GetRestConfigByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s配置失败:%s", err.Error()))
		return
	}

	metricsGVR := schema.GroupVersionResource{Group: "metrics.k8s.io", Version: "v1beta1", Resource: "pods"}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		response.Fail(c, fmt.Sprintf("创建动态客户端失败:%s", err.Error()))
		return
	}

	var metricsList *unstructured.UnstructuredList
	if namespace != "" {
		metricsList, err = dynamicClient.Resource(metricsGVR).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	} else {
		metricsList, err = dynamicClient.Resource(metricsGVR).List(context.TODO(), metav1.ListOptions{})
	}
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取Pod指标失败:%s (metrics-server可能未安装)", err.Error()))
		return
	}

	var result []map[string]any
	for _, item := range metricsList.Items {
		name := item.GetName()
		namespace := item.GetNamespace()
		containers, _, _ := unstructured.NestedSlice(item.Object, "containers")
		var totalCPU, totalMemory string
		for _, c := range containers {
			container, ok := c.(map[string]any)
			if !ok {
				continue
			}
			usage, ok := container["usage"].(map[string]any)
			if !ok {
				continue
			}
			totalCPU = fmt.Sprintf("%v", usage["cpu"])
			totalMemory = fmt.Sprintf("%v", usage["memory"])
		}
		result = append(result, map[string]any{
			"name":        name,
			"namespace":   namespace,
			"cpuUsage":    totalCPU,
			"memoryUsage": totalMemory,
		})
	}
	response.Success(c, "执行成功", result)
}
