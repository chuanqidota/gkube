package daemonset

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// GetDaemonSetList
//
//	@Description: 获取daemonSet列表
//	@param client
//	@param namespace
//	@return []appsv1.DaemonSet
//	@return error
func GetDaemonSetList(client *kubernetes.Clientset, namespace string) ([]appsv1.DaemonSet, error) {
	daemonSetList, err := client.AppsV1().DaemonSets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return daemonSetList.Items, nil
}

// ListDaemonSets returns a paginated daemonset list with metadata
func ListDaemonSets(client *kubernetes.Clientset, namespace string, limit int64, continueToken string) (*appsv1.DaemonSetList, error) {
	listOpts := metav1.ListOptions{}
	if limit > 0 {
		listOpts.Limit = limit
	}
	if continueToken != "" {
		listOpts.Continue = continueToken
	}
	return client.AppsV1().DaemonSets(namespace).List(context.Background(), listOpts)
}

// GetDaemonSetByName
//
//	@Description: 获取daemonSet
//	@param client
//	@param namespace
//	@param name
//	@return *appsv1.DaemonSet
//	@return error
func GetDaemonSetByName(client *kubernetes.Clientset, namespace, name string) (*appsv1.DaemonSet, error) {
	daemonSet, err := client.AppsV1().DaemonSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return daemonSet, nil
}

// GetDaemonSetYaml
//
//	@Description: 获取daemonSetYaml
//	@param client
//	@param namespace
//	@param name
//	@return string
//	@return error
func GetDaemonSetYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	daemonSet, err := client.AppsV1().DaemonSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	yamlBytes, err := yaml.Marshal(daemonSet)
	if err != nil {
		return "", err
	}
	return string(yamlBytes), nil
}

// GetDaemonSetByField
//
//	@Description: 根据字段查询daemonSet
//	@param client
//	@param namespace
//	@param fieldMap
//	@return []appsv1.DaemonSet
//	@return error
func GetDaemonSetByField(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) ([]appsv1.DaemonSet, error) {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	daemonSetList, err := client.AppsV1().DaemonSets(namespace).List(context.Background(), metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return nil, err
	}
	return daemonSetList.Items, nil
}

// GetDaemonSetByLabel
//
//	@Description: 根据标签查询daemonSet
//	@param client
//	@param namespace
//	@param labelMap
//	@return []appsv1.DaemonSet
//	@return error
func GetDaemonSetByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) ([]appsv1.DaemonSet, error) {
	labelSelector := fields.SelectorFromSet(labelMap)
	daemonSetList, err := client.AppsV1().DaemonSets(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return nil, err
	}
	return daemonSetList.Items, nil
}

// CreateDaemonSet
//
//	@Description: 创建daemonSet
//	@param client
//	@param namespace
//	@param daemonSetYaml
//	@return error
func CreateDaemonSet(client *kubernetes.Clientset, namespace, daemonSetYaml string) error {
	daemonSet := &appsv1.DaemonSet{}
	err := yaml.Unmarshal([]byte(daemonSetYaml), daemonSet)
	if err != nil {
		return err
	}
	_, err = client.AppsV1().DaemonSets(namespace).Create(context.Background(), daemonSet, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// UpdateDaemonSet
//
//	@Description: 更新daemonSet
//	@param client
//	@param namespace
//	@param daemonSetYaml
//	@return error
func UpdateDaemonSet(client *kubernetes.Clientset, namespace, daemonSetYaml string) error {
	daemonSet := &appsv1.DaemonSet{}
	err := yaml.Unmarshal([]byte(daemonSetYaml), daemonSet)
	if err != nil {
		return err
	}
	_, err = client.AppsV1().DaemonSets(namespace).Update(context.Background(), daemonSet, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("更新daemonSet资源失败:%s", err.Error())
	}
	return nil
}

// DeleteDaemonSetByName
//
//	@Description: 删除daemonSet
//	@param client
//	@param namespace
//	@param name
//	@return error
func DeleteDaemonSetByName(client *kubernetes.Clientset, namespace, name string) error {
	err := client.AppsV1().DaemonSets(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("删除daemonSet资源失败:%s", err.Error())
	}
	return nil
}

// DeleteDaemonSetByField
//
//	@Description: 根据字段删除daemonSet
//	@param client
//	@param namespace
//	@param fieldMap
//	@return error
func DeleteDaemonSetByField(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) error {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	err := client.AppsV1().DaemonSets(namespace).DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return fmt.Errorf("删除daemonSet资源失败:%s", err.Error())
	}
	return nil
}

// DeleteDaemonSetByLabel
//
//	@Description: 根据标签删除daemonSet
//	@param client
//	@param namespace
//	@param labelMap
//	@return error
func DeleteDaemonSetByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) error {
	labelSelector := fields.SelectorFromSet(labelMap)
	err := client.AppsV1().DaemonSets(namespace).DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return fmt.Errorf("删除daemonSet资源失败:%s", err.Error())
	}
	return nil
}

// DaemonSetPodList
//
//	@Description: 获取daemonSet关联的pod列表
//	@param client
//	@param namespace
//	@param name
//	@return *corev1.PodList
//	@return error
func DaemonSetPodList(client *kubernetes.Clientset, namespace, name string) (*corev1.PodList, error) {
	daemonSet, err := client.AppsV1().DaemonSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取daemonSet资源失败:%s", err.Error())
	}
	selector := labels.Set(daemonSet.Spec.Selector.MatchLabels).AsSelectorPreValidated()
	podList, err := client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("获取pod资源失败:%s", err.Error())
	}
	return podList, nil
}
