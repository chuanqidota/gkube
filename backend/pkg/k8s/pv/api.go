package pv

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

// GetPVList
//
//	@Description: 获取PV列表
//	@param client
//	@return []corev1.PersistentVolume
//	@return error
func GetPVList(client *kubernetes.Clientset) ([]corev1.PersistentVolume, error) {
	pvList, err := client.CoreV1().PersistentVolumes().List(context.Background(), metav1.ListOptions{ResourceVersion: "0"})
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
func GetPVByField(client *kubernetes.Clientset, fieldMap map[string]string) ([]corev1.PersistentVolume, error) {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	pvList, err := client.CoreV1().PersistentVolumes().List(context.Background(), metav1.ListOptions{
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
func GetPVYaml(client *kubernetes.Clientset, name string) (string, error) {
	pv, err := client.CoreV1().PersistentVolumes().Get(context.Background(), name, metav1.GetOptions{})
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
func CreatePV(client *kubernetes.Clientset, pvYaml string) error {
	var persistentVolume corev1.PersistentVolume
	if err := yaml.Unmarshal([]byte(pvYaml), &persistentVolume); err != nil {
		return fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	_, err := client.CoreV1().PersistentVolumes().Create(context.Background(), &persistentVolume, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("创建pv资源失败:%s", err.Error())
	}
	return nil
}

// UpdatePV
//
//	@Description: 更新PV
//	@param client
//	@param pvYaml
//	@return error
func UpdatePV(client *kubernetes.Clientset, pvYaml string) error {
	var pv corev1.PersistentVolume
	if err := yaml.Unmarshal([]byte(pvYaml), &pv); err != nil {
		return fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := client.CoreV1().PersistentVolumes().Get(context.Background(), pv.Name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("获取PV资源失败:%s", err.Error())
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
		_, err = client.CoreV1().PersistentVolumes().Update(context.Background(), latest, metav1.UpdateOptions{})
		return err
	})
}

// DeletePVByName
//
//	@Description: 删除PV通过名称
//	@param client
//	@param name
//	@return error
func DeletePVByName(client *kubernetes.Clientset, name string) error {
	// 检查 PV 绑定状态,bound 状态下拒绝删除并给出引导性错误
	pv, err := client.CoreV1().PersistentVolumes().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if pv.Status.Phase == corev1.VolumeBound && pv.Spec.ClaimRef != nil {
		return fmt.Errorf("PV %s 正在绑定 PVC %s/%s,请先删除关联的 PVC", name, pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name)
	}
	propagation := metav1.DeletePropagationForeground
	return client.CoreV1().PersistentVolumes().Delete(context.Background(), name, metav1.DeleteOptions{
		PropagationPolicy: &propagation,
	})
}

// DeletePVByLabel
//
//	@Description: 删除PV通过标签
//	@param client
//	@param labelMap
//	@return error
func DeletePVByLabel(client *kubernetes.Clientset, labelMap map[string]string) error {
	labelSelector := labels.SelectorFromSet(labelMap)
	err := client.CoreV1().PersistentVolumes().DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{
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
func DeletePVByField(client *kubernetes.Clientset, fieldMap map[string]string) error {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	err := client.CoreV1().PersistentVolumes().DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return err
	}
	return nil
}
