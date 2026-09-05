package pvc

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/yaml"

	"gkube/pkg/yamlutil"
)

// GetPVCList
//
//	@Description: 获取PVC列表
//	@param client
//	@param namespace
//	@return []corev1.PersistentVolumeClaim
//	@return error
func GetPVCList(client *kubernetes.Clientset, namespace string, labelSelector string) ([]corev1.PersistentVolumeClaim, error) {
	listOpts := metav1.ListOptions{ResourceVersion: "0"}
	if labelSelector != "" {
		listOpts.LabelSelector = labelSelector
	}
	result, err := client.CoreV1().PersistentVolumeClaims(namespace).List(context.TODO(), listOpts)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

// GetPVCListByStorageClass
//
//	@Description: 根据 storageClassName 获取 PVC 列表（跨命名空间）
//	@param client
//	@param storageClassName
//	@return []corev1.PersistentVolumeClaim
//	@return error
func GetPVCListByStorageClass(client *kubernetes.Clientset, storageClassName string) ([]corev1.PersistentVolumeClaim, error) {
	fieldSelector := fields.SelectorFromSet(map[string]string{
		"spec.storageClassName": storageClassName,
	})
	pvcList, err := client.CoreV1().PersistentVolumeClaims("").List(context.Background(), metav1.ListOptions{
		FieldSelector:  fieldSelector.String(),
		ResourceVersion: "0",
	})
	if err != nil {
		return nil, err
	}
	return pvcList.Items, nil
}

// GetPVCByName
//
//	@Description: 根据名称获取PVC
//	@param client
//	@param namespace
//	@param name
//	@return *corev1.PersistentVolumeClaim
//	@return error
func GetPVCByName(client *kubernetes.Clientset, namespace, name string) (*corev1.PersistentVolumeClaim, error) {
	pvc, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return pvc, nil
}

// GetPVCByLabel
//
//	@Description: 根据标签获取PVC
//	@param client
//	@param namespace
//	@param labelMap
//	@return []corev1.PersistentVolumeClaim
//	@return error
func GetPVCByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) ([]corev1.PersistentVolumeClaim, error) {
	labelSelector := labels.SelectorFromSet(labelMap)
	pvcList, err := client.CoreV1().PersistentVolumeClaims(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector:  labelSelector.String(),
		ResourceVersion: "0",
	})
	if err != nil {
		return nil, err
	}
	return pvcList.Items, nil
}

// GetPVCByField
//
//	@Description: 根据字段获取PVC
//	@param client
//	@param namespace
//	@param fieldMap
//	@return []corev1.PersistentVolumeClaim
//	@return error
func GetPVCByField(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) ([]corev1.PersistentVolumeClaim, error) {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	pvcList, err := client.CoreV1().PersistentVolumeClaims(namespace).List(context.Background(), metav1.ListOptions{
		FieldSelector:  fieldSelector.String(),
		ResourceVersion: "0",
	})
	if err != nil {
		return nil, err
	}
	return pvcList.Items, nil
}

// GetPVCYaml
//
//	@Description: 获取PVC Yaml
//	@param client
//	@param namespace
//	@param name
//	@return string
//	@return error
func GetPVCYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	pvc, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	yamlStr, err := yamlutil.MarshalWithoutManagedFields(pvc)
	if err != nil {
		return "", err
	}
	return yamlStr, nil
}

// CreatePVC
//
//	@Description: 创建PVC
//	@param client
//	@param namespace
//	@param pvcYaml
//	@return error
func CreatePVC(client *kubernetes.Clientset, namespace, pvcYaml string) error {
	var pvc corev1.PersistentVolumeClaim
	if err := yaml.Unmarshal([]byte(pvcYaml), &pvc); err != nil {
		return fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	_, err := client.CoreV1().PersistentVolumeClaims(namespace).Create(context.Background(), &pvc, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("创建pvc资源失败:%s", err.Error())
	}
	return nil
}

// UpdatePVC
//
//	@Description: 更新PVC
//	@param client
//	@param namespace
//	@param pvcYaml
//	@return error
func UpdatePVC(client *kubernetes.Clientset, namespace, pvcYaml string) error {
	var pvc corev1.PersistentVolumeClaim
	if err := yaml.Unmarshal([]byte(pvcYaml), &pvc); err != nil {
		return fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(context.Background(), pvc.Name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("获取PVC资源失败:%s", err.Error())
		}
		// PVC 的 resources/storageClassName/volumeName/selector/volumeMode 不可变,只更新可变字段
		latest.Labels = pvc.Labels
		latest.Annotations = pvc.Annotations
		_, err = client.CoreV1().PersistentVolumeClaims(namespace).Update(context.Background(), latest, metav1.UpdateOptions{})
		return err
	})
}

// DeletePVCByName
//
//	@Description: 删除PVC
//	@param client
//	@param namespace
//	@param name
//	@return error
func DeletePVCByName(client *kubernetes.Clientset, namespace, name string) error {
	err := client.CoreV1().PersistentVolumeClaims(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

// DeletePVCByLabel
//
//	@Description: 删除PVC
//	@param client
//	@param namespace
//	@param labelMap
//	@return error
func DeletePVCByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) error {
	labelSelector := labels.SelectorFromSet(labelMap)
	err := client.CoreV1().PersistentVolumeClaims(namespace).DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return err
	}
	return nil
}

// DeletePVCByField
//
//	@Description: 删除PVC
//	@param client
//	@param namespace
//	@param fieldMap
//	@return error
func DeletePVCByField(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) error {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	err := client.CoreV1().PersistentVolumeClaims(namespace).DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return err
	}
	return nil
}
