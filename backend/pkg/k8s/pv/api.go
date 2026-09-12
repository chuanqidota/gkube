package pv

import (
	"context"
	apperr "gkube/pkg/errors"
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

// GetPVList
//
//	@Description: 获取PV列表
//	@param client
//	@return []corev1.PersistentVolume
//	@return error
func GetPVList(ctx context.Context, client *kubernetes.Clientset, labelSelector string) ([]corev1.PersistentVolume, error) {
	listOpts := metav1.ListOptions{ResourceVersion: "0"}
	if labelSelector != "" {
		listOpts.LabelSelector = labelSelector
	}
	result, err := client.CoreV1().PersistentVolumes().List(ctx, listOpts)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

// GetPVByName
//
//	@Description: 根据名称获取PV
//	@param client
//	@param name
//	@return *corev1.PersistentVolume
//	@return error
func GetPVByName(ctx context.Context, client *kubernetes.Clientset, name string) (*corev1.PersistentVolume, error) {
	pv, err := client.CoreV1().PersistentVolumes().Get(ctx, name, metav1.GetOptions{})
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
func GetPVByLabel(ctx context.Context, client *kubernetes.Clientset, labelMap map[string]string) ([]corev1.PersistentVolume, error) {
	labelSelector := labels.SelectorFromSet(labelMap)
	pvList, err := client.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{
		LabelSelector:  labelSelector.String(),
		ResourceVersion: "0",
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
func GetPVByField(ctx context.Context, client *kubernetes.Clientset, fieldMap map[string]string) ([]corev1.PersistentVolume, error) {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	pvList, err := client.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{
		FieldSelector:  fieldSelector.String(),
		ResourceVersion: "0",
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
func GetPVYaml(ctx context.Context, client *kubernetes.Clientset, name string) (string, error) {
	pv, err := client.CoreV1().PersistentVolumes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	yamlStr, err := yamlutil.MarshalWithoutManagedFields(pv)
	if err != nil {
		return "", err
	}
	return yamlStr, nil
}

// CreatePV
//
//	@Description: 创建PV
//	@param client
//	@param pvYaml
//	@return error
func CreatePV(ctx context.Context, client *kubernetes.Clientset, pvYaml string) error {
	var persistentVolume corev1.PersistentVolume
	if err := yaml.Unmarshal([]byte(pvYaml), &persistentVolume); err != nil {
		return apperr.BadRequest("YAML解析失败", err)
	}
	_, err := client.CoreV1().PersistentVolumes().Create(ctx, &persistentVolume, metav1.CreateOptions{})
	if err != nil {
		return apperr.K8sAPIFail("创建pv资源失败", err)
	}
	return nil
}

// UpdatePV
//
//	@Description: 更新PV
//	@param client
//	@param pvYaml
//	@return error
func UpdatePV(ctx context.Context, client *kubernetes.Clientset, pvYaml string) error {
	var pv corev1.PersistentVolume
	if err := yaml.Unmarshal([]byte(pvYaml), &pv); err != nil {
		return apperr.BadRequest("YAML解析失败", err)
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := client.CoreV1().PersistentVolumes().Get(ctx, pv.Name, metav1.GetOptions{})
		if err != nil {
			return apperr.K8sAPIFail("获取PV资源失败", err)
		}
		// 用用户 YAML 的可变字段覆盖最新对象,保留最新 resourceVersion
		latest.Spec.Capacity = pv.Spec.Capacity
		latest.Spec.AccessModes = pv.Spec.AccessModes
		latest.Spec.PersistentVolumeReclaimPolicy = pv.Spec.PersistentVolumeReclaimPolicy
		latest.Spec.StorageClassName = pv.Spec.StorageClassName
		latest.Spec.MountOptions = pv.Spec.MountOptions
		latest.Spec.VolumeMode = pv.Spec.VolumeMode
		latest.Spec.PersistentVolumeSource = pv.Spec.PersistentVolumeSource
		latest.Labels = pv.Labels
		latest.Annotations = pv.Annotations
		_, err = client.CoreV1().PersistentVolumes().Update(ctx, latest, metav1.UpdateOptions{})
		return err
	})
}

// DeletePVByName
//
//	@Description: 删除PV通过名称
//	@param client
//	@param name
//	@return error
func DeletePVByName(ctx context.Context, client *kubernetes.Clientset, name string) error {
	// 检查 PV 绑定状态,bound 状态下拒绝删除并给出引导性错误
	pv, err := client.CoreV1().PersistentVolumes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if pv.Status.Phase == corev1.VolumeBound && pv.Spec.ClaimRef != nil {
		return fmt.Errorf("PV %s 正在绑定 PVC %s/%s,请先删除关联的 PVC", name, pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name)
	}
	propagation := metav1.DeletePropagationForeground
	return client.CoreV1().PersistentVolumes().Delete(ctx, name, metav1.DeleteOptions{
		PropagationPolicy: &propagation,
	})
}

// DeletePVByLabel
//
//	@Description: 删除PV通过标签
//	@param client
//	@param labelMap
//	@return error
func DeletePVByLabel(ctx context.Context, client *kubernetes.Clientset, labelMap map[string]string) error {
	labelSelector := labels.SelectorFromSet(labelMap)
	err := client.CoreV1().PersistentVolumes().DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return err
	}
	return nil
}

// DeletePVByField
//
//	@Description: 删除PV通过字段
//	@param client
//	@param fieldMap
//	@return error
func DeletePVByField(ctx context.Context, client *kubernetes.Clientset, fieldMap map[string]string) error {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	err := client.CoreV1().PersistentVolumes().DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return err
	}
	return nil
}
