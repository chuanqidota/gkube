package pv

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

// GetPVList
//
//	@Description: 获取PV列表
//	@param client
//	@return []corev1.PersistentVolume
//	@return error
func GetPVList(client *kubernetes.Clientset) ([]corev1.PersistentVolume, error) {
	pvList, err := client.CoreV1().PersistentVolumes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pvList.Items, nil
}

// GetPVByName
//
//	@Description: 根据名称获取PV
//	@param client
//	@param name
//	@return *corev1.PersistentVolume
//	@return error
func GetPVByName(client *kubernetes.Clientset, name string) (*corev1.PersistentVolume, error) {
	pv, err := client.CoreV1().PersistentVolumes().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return pv, nil
}

// GetPVByLabel
//
//	@Description: 根据标签获取PV
//	@param client
//	@param labelMap
//	@return []corev1.PersistentVolume
//	@return error
func GetPVByLabel(client *kubernetes.Clientset, labelMap map[string]string) ([]corev1.PersistentVolume, error) {
	labelSelector := labels.SelectorFromSet(labelMap)
	pvList, err := client.CoreV1().PersistentVolumes().List(context.Background(), metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return nil, err
	}
	return pvList.Items, nil
}

// GetPVByField
//
//	@Description: 根据字段获取PV
//	@param client
//	@param fieldMap
//	@return []corev1.PersistentVolume
//	@return error
func GetPVByField(client *kubernetes.Clientset, fieldMap map[string]string) ([]corev1.PersistentVolume, error) {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	pvList, err := client.CoreV1().PersistentVolumes().List(context.Background(), metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return nil, err
	}
	return pvList.Items, nil
}

// GetPVYaml
//
//	@Description: 获取PV Yaml
//	@param client
//	@param name
//	@return string
//	@return error
func GetPVYaml(client *kubernetes.Clientset, name string) (string, error) {
	pv, err := client.CoreV1().PersistentVolumes().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return pv.String(), nil
}

// CreatePV
//
//	@Description: 创建PV
//	@param client
//	@param pvYaml
//	@return bool
//	@return error
func CreatePV(client *kubernetes.Clientset, pvYaml string) (bool, error) {

	pv, err := client.CoreV1().PersistentVolumes().Create(context.Background(), &corev1.PersistentVolume{}, metav1.CreateOptions{})
	if err != nil {
		return false, err
	}
	return pv.Name == pvYaml, nil
}

// UpdatePV
//
//	@Description: 更新PV
//	@param client
//	@param pvYaml
//	@return bool
//	@return error
func UpdatePV(client *kubernetes.Clientset, pvYaml string) (bool, error) {
	pv, err := client.CoreV1().PersistentVolumes().Update(context.Background(), &corev1.PersistentVolume{}, metav1.UpdateOptions{})
	if err != nil {
		return false, err
	}
	return pv.Name == pvYaml, nil
}

// DeletePV
//
//	@Description: 删除PV
//	@param client
//	@param name
//	@return bool
//	@return error
func DeletePV(client *kubernetes.Clientset, name string) (bool, error) {
	err := client.CoreV1().PersistentVolumes().Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return false, err
	}
	return true, nil
}

// DeletePVByLabel
//
//	@Description: 删除PV通过标签
//	@param client
//	@param labelMap
//	@return bool
//	@return error
func DeletePVByLabel(client *kubernetes.Clientset, labelMap map[string]string) (bool, error) {
	labelSelector := labels.SelectorFromSet(labelMap)
	err := client.CoreV1().PersistentVolumes().DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

// DeletePVByField
//
//	@Description: 删除PV通过字段
//	@param client
//	@param fieldMap
//	@return bool
//	@return error
func DeletePVByField(client *kubernetes.Clientset, fieldMap map[string]string) (bool, error) {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	err := client.CoreV1().PersistentVolumes().DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

// DeletePVByName
//
//	@Description: 删除PV通过名称
//	@param client
//	@param name
//	@return bool
//	@return error
func DeletePVByName(client *kubernetes.Clientset, name string) (bool, error) {
	err := client.CoreV1().PersistentVolumes().Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return false, err
	}
	return true, nil
}
