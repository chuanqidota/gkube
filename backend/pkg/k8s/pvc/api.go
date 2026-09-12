package pvc

import (
	"context"
	apperr "gkube/pkg/errors"

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
func GetPVCList(ctx context.Context, client *kubernetes.Clientset, namespace string, labelSelector string) ([]corev1.PersistentVolumeClaim, error) {
	listOpts := metav1.ListOptions{ResourceVersion: "0"}
	if labelSelector != "" {
		listOpts.LabelSelector = labelSelector
	}
	result, err := client.CoreV1().PersistentVolumeClaims(namespace).List(ctx, listOpts)
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
func GetPVCListByStorageClass(ctx context.Context, client *kubernetes.Clientset, storageClassName string) ([]corev1.PersistentVolumeClaim, error) {
	fieldSelector := fields.SelectorFromSet(map[string]string{
		"spec.storageClassName": storageClassName,
	})
	pvcList, err := client.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{
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
func GetPVCByName(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (*corev1.PersistentVolumeClaim, error) {
	pvc, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
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
func GetPVCByLabel(ctx context.Context, client *kubernetes.Clientset, namespace string, labelMap map[string]string) ([]corev1.PersistentVolumeClaim, error) {
	labelSelector := labels.SelectorFromSet(labelMap)
	pvcList, err := client.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{
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
func GetPVCByField(ctx context.Context, client *kubernetes.Clientset, namespace string, fieldMap map[string]string) ([]corev1.PersistentVolumeClaim, error) {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	pvcList, err := client.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{
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
func GetPVCYaml(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (string, error) {
	pvc, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
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
func CreatePVC(ctx context.Context, client *kubernetes.Clientset, namespace, pvcYaml string) error {
	var pvc corev1.PersistentVolumeClaim
	if err := yaml.Unmarshal([]byte(pvcYaml), &pvc); err != nil {
		return apperr.K8sAPIFail("yaml文件错误", err)
	}
	_, err := client.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, &pvc, metav1.CreateOptions{})
	if err != nil {
		return apperr.K8sAPIFail("创建pvc资源失败", err)
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
func UpdatePVC(ctx context.Context, client *kubernetes.Clientset, namespace, pvcYaml string) error {
	var pvc corev1.PersistentVolumeClaim
	if err := yaml.Unmarshal([]byte(pvcYaml), &pvc); err != nil {
		return apperr.K8sAPIFail("yaml文件错误", err)
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, pvc.Name, metav1.GetOptions{})
		if err != nil {
			return apperr.K8sAPIFail("获取PVC资源失败", err)
		}
		// PVC 的 resources/storageClassName/volumeName/selector/volumeMode 不可变,只更新可变字段
		latest.Labels = pvc.Labels
		latest.Annotations = pvc.Annotations
		_, err = client.CoreV1().PersistentVolumeClaims(namespace).Update(ctx, latest, metav1.UpdateOptions{})
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
func DeletePVCByName(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error {
	err := client.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, name, metav1.DeleteOptions{})
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
func DeletePVCByLabel(ctx context.Context, client *kubernetes.Clientset, namespace string, labelMap map[string]string) error {
	labelSelector := labels.SelectorFromSet(labelMap)
	err := client.CoreV1().PersistentVolumeClaims(namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{
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
func DeletePVCByField(ctx context.Context, client *kubernetes.Clientset, namespace string, fieldMap map[string]string) error {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	err := client.CoreV1().PersistentVolumeClaims(namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return err
	}
	return nil
}
