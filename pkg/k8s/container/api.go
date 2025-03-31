package container

import (
	"context"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// GetPodContainerLog 获取指定命名空间、Pod 和容器的日志。
// 参数:
// - client: Kubernetes 客户端集。
// - namespace: Pod 所在的命名空间。
// - podName: Pod 的名称。
// - containerName: 容器的名称。
// - tailLines: 返回日志的最后 N 行。
// 返回值:
// - 日志内容字符串。
// - 如果发生错误，返回错误信息。
func GetPodContainerLog(client *kubernetes.Clientset, namespace, podName, containerName string, tailLines int64) (string, error) {
	// 创建一个获取 Pod 日志的请求
	req := client.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container:  containerName, // 指定容器名称
		Follow:     false,         // 是否实时跟踪（类似 -f）
		TailLines:  &tailLines,    // 获取最后 N 行日志
		Previous:   false,         // 获取已终止容器的日志
		Timestamps: true,          // 包含时间戳
	})

	// 执行请求并获取日志流
	stream, err := req.Stream(context.Background())
	if err != nil {
		return "", fmt.Errorf("创建日志流失败: %v", err.Error())
	}
	defer stream.Close() // 确保在函数结束时关闭流

	// 读取日志内容
	logContent, err := io.ReadAll(stream)
	if err != nil {
		return "", fmt.Errorf("读取日志失败: %v", err.Error())
	}

	// 返回日志内容字符串
	return string(logContent), nil
}
