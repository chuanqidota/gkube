package daemonset

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
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

// GetDaemonSet
//
//	@Description: 获取daemonSet
//	@param client
//	@param namespace
//	@param name
//	@return *appsv1.DaemonSet
//	@return error
func GetDaemonSet(client *kubernetes.Clientset, namespace, name string) (*appsv1.DaemonSet, error) {
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

	return daemonSet.String(), nil
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
//	@return bool
//	@return error
func CreateDaemonSet(client *kubernetes.Clientset, namespace, daemonSetYaml string) (bool, error) {
	daemonSet := &appsv1.DaemonSet{}
	err := yaml.Unmarshal([]byte(daemonSetYaml), daemonSet)
	if err != nil {
		return false, err
	}
	_, err = client.AppsV1().DaemonSets(namespace).Create(context.Background(), daemonSet, metav1.CreateOptions{})
	if err != nil {
		return false, err
	}
	return true, nil
}

// UpdateDaemonSet
//
//	@Description: 更新daemonSet
//	@param client
//	@param namespace
//	@param daemonSetYaml
//	@return bool
//	@return error
func UpdateDaemonSet(client *kubernetes.Clientset, namespace, daemonSetYaml string) (bool, error) {
	daemonSet := &appsv1.DaemonSet{}
	err := yaml.Unmarshal([]byte(daemonSetYaml), daemonSet)
	if err != nil {
		return false, err
	}
	_, err = client.AppsV1().DaemonSets(namespace).Update(context.Background(), daemonSet, metav1.UpdateOptions{})
	if err != nil {
		return false, fmt.Errorf("更新daemonSet资源失败:%s", err.Error())
	}
	return true, nil
}

// DeleteDaemonSet
//
//	@Description: 删除daemonSet
//	@param client
//	@param namespace
//	@param name
//	@return bool
//	@return error
func DeleteDaemonSet(client *kubernetes.Clientset, namespace, name string) (bool, error) {
	err := client.AppsV1().DaemonSets(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return false, fmt.Errorf("删除daemonSet资源失败:%s", err.Error())
	}
	return true, nil
}

// DeleteDaemonSetByField
//
//	@Description: 根据字段删除daemonSet
//	@param client
//	@param namespace
//	@param fieldMap
//	@return bool
//	@return error
func DeleteDaemonSetByField(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) (bool, error) {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	err := client.AppsV1().DaemonSets(namespace).DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return false, fmt.Errorf("删除daemonSet资源失败:%s", err.Error())
	}
	return true, nil
}

// DeleteDaemonSetByLabel
//
//	@Description: 根据标签删除daemonSet
//	@param client
//	@param namespace
//	@param labelMap
//	@return bool
//	@return error
func DeleteDaemonSetByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) (bool, error) {
	labelSelector := fields.SelectorFromSet(labelMap)
	err := client.AppsV1().DaemonSets(namespace).DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return false, fmt.Errorf("删除daemonSet资源失败:%s", err.Error())
	}
	return true, nil
}
