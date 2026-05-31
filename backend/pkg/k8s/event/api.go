package event

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetEventList
//
//	@Description: 获取事件列表
//	@param client
//	@param namespace
//	@return []corev1.Event
//	@return error
func GetEventList(client *kubernetes.Clientset, namespace string) ([]corev1.Event, error) {
	eventList, err := client.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return eventList.Items, nil
}

// GetPodsEventList
//
//	@Description: 根据pod名称获取事件列表
//	@param client
//	@param namespace
//	@param podName
//	@return []corev1.Event
//	@return error
func GetPodsEventList(client *kubernetes.Clientset, namespace, podName string) ([]corev1.Event, error) {
	eventList, err := client.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: "involvedObject.name=" + podName,
	})
	if err != nil {
		return nil, err
	}
	return eventList.Items, nil
}
