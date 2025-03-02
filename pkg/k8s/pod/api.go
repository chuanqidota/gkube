package pod

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

// GetPodList
//
//	@Description: 获取pod列表
//	@param client
//	@param namespace
//	@return []corev1.Pod
//	@return error
func GetPodList(client *kubernetes.Clientset, namespace string) ([]corev1.Pod, error) {
	podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}

// GetPodByName
//
//	@Description: 获取pod
//	@param client
//	@param namespace
//	@param name
//	@return *corev1.Pod
//	@return error
func GetPodByName(client *kubernetes.Clientset, namespace, name string) (*corev1.Pod, error) {
	pod, err := client.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return pod, nil
}

// GetPodYaml
//
//	@Description: 获取pod yaml
//	@param client
//	@param namespace
//	@param name
//	@return string
//	@return error
func GetPodYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	pod, err := client.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return pod.String(), nil
}

// GetPodByFiled
//
//	@Description: 通过字段查询pod
//	@param client
//	@param namespace
//	@param fieldMap
//	@return []corev1.Pod
//	@return error
func GetPodByFiled(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) ([]corev1.Pod, error) {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}

// GetPodByLabel
//
//	@Description: 通过标签查询pod
//	@param client
//	@param namespace
//	@param labelMap
//	@return []corev1.Pod
//	@return error
func GetPodByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) ([]corev1.Pod, error) {
	labelSelector := labels.SelectorFromSet(labelMap)
	podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}
