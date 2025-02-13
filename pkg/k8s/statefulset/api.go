package statefulset

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// GetStatefulSetList
//
//	@Description: 获取statefulSet列表
//	@param client
//	@param namespace
//	@return []appsv1.StatefulSet
//	@return error
func GetStatefulSetList(client *kubernetes.Clientset, namespace string) ([]appsv1.StatefulSet, error) {
	statefulSetList, err := client.AppsV1().StatefulSets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return statefulSetList.Items, nil
}

// GetStatefulSet
//
//	@Description: 获取statefulSet
//	@param client
//	@param namespace
//	@param name
//	@return *appsv1.StatefulSet
//	@return error
func GetStatefulSet(client *kubernetes.Clientset, namespace, name string) (*appsv1.StatefulSet, error) {
	statefulSet, err := client.AppsV1().StatefulSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return statefulSet, nil
}

// GetStatefulSetYaml
//
//	@Description: 获取statefulSet的yaml
//	@param client
//	@param namespace
//	@param name
//	@return string
//	@return error
func GetStatefulSetYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	statefulSet, err := client.AppsV1().StatefulSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return statefulSet.String(), nil
}

// GetStatefulSetByField
//
//	@Description: 根据字段查询statefulSet
//	@param client
//	@param namespace
//	@param fieldMap
//	@return []appsv1.StatefulSet
//	@return error
func GetStatefulSetByField(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) ([]appsv1.StatefulSet, error) {
	statefulSetList, err := client.AppsV1().StatefulSets(namespace).List(context.Background(), metav1.ListOptions{
		FieldSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{
			MatchLabels: fieldMap,
		}),
	})
	if err != nil {
		return nil, err
	}
	return statefulSetList.Items, nil
}

// GetStatefulSetByLabel
//
//	@Description: 根据标签查询statefulSet
//	@param client
//	@param namespace
//	@param labelMap
//	@return []appsv1.StatefulSet
//	@return error
func GetStatefulSetByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) ([]appsv1.StatefulSet, error) {
	statefulSetList, err := client.AppsV1().StatefulSets(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{
			MatchLabels: labelMap,
		}),
	})
	if err != nil {
		return nil, err
	}
	return statefulSetList.Items, nil
}

// CreateStatefulSet
//
//	@Description: 创建statefulSet
//	@param client
//	@param namespace
//	@param statefulSetYaml
//	@return bool
//	@return error
func CreateStatefulSet(client *kubernetes.Clientset, namespace, statefulSetYaml string) (bool, error) {
	var statefulSet appsv1.StatefulSet
	if err := yaml.Unmarshal([]byte(statefulSetYaml), &statefulSet); err != nil {
		return false, fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	_, err := client.AppsV1().StatefulSets(namespace).Create(context.Background(), &statefulSet, metav1.CreateOptions{})
	if err != nil {
		return false, fmt.Errorf("创建statefulSet资源失败:%s", err.Error())
	}
	return true, nil
}

// UpdateStatefulSet
//
//	@Description: 更新statefulSet
//	@param client
//	@param namespace
//	@param statefulSetYaml
//	@return bool
//	@return error
func UpdateStatefulSet(client *kubernetes.Clientset, namespace, statefulSetYaml string) (bool, error) {
	var statefulSet appsv1.StatefulSet
	if err := yaml.Unmarshal([]byte(statefulSetYaml), &statefulSet); err != nil {
		return false, fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	_, err := client.AppsV1().StatefulSets(namespace).Update(context.Background(), &statefulSet, metav1.UpdateOptions{})
	if err != nil {
		return false, fmt.Errorf("更新statefulSet资源失败:%s", err.Error())
	}
	return true, nil
}

// DeleteStatefulSet
//
//	@Description: 删除statefulSet
//	@param client
//	@param namespace
//	@param name
//	@return bool
//	@return error
func DeleteStatefulSet(client *kubernetes.Clientset, namespace, name string) (bool, error) {
	err := client.AppsV1().StatefulSets(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return false, fmt.Errorf("删除statefulSet资源失败:%s", err.Error())
	}
	return true, nil
}

// DeleteStatefulSetByLabel
//
//	@Description: 根据标签删除statefulSet
//	@param client
//	@param namespace
//	@param labelMap
//	@return bool
//	@return error
func DeleteStatefulSetByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) (bool, error) {
	err := client.AppsV1().StatefulSets(namespace).DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{
			MatchLabels: labelMap,
		}),
	})
	if err != nil {
		return false, fmt.Errorf("删除statefulSet资源失败:%s", err.Error())
	}
	return true, nil
}

// DeleteStatefulSetByField
//
//	@Description: 根据字段删除statefulSet
//	@param client
//	@param namespace
//	@param fieldMap
//	@return bool
//	@return error
func DeleteStatefulSetByField(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) (bool, error) {
	err := client.AppsV1().StatefulSets(namespace).DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{
		FieldSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{
			MatchLabels: fieldMap,
		}),
	})
	if err != nil {
		return false, fmt.Errorf("删除statefulSet资源失败:%s", err.Error())
	}
	return true, nil
}
