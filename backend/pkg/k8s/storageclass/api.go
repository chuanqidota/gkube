package storageclass

import (
	"context"
	"gkube/pkg/yamlutil"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// GetStorageClassList
//
//	@Description: 获取StorageClass列表
//	@param client
//	@return []storagev1.StorageClass
//	@return error
func GetStorageClassList(ctx context.Context, client *kubernetes.Clientset, labelSelector string) ([]storagev1.StorageClass, error) {
	listOpts := metav1.ListOptions{ResourceVersion: "0"}
	if labelSelector != "" {
		listOpts.LabelSelector = labelSelector
	}
	result, err := client.StorageV1().StorageClasses().List(ctx, listOpts)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

// GetStorageClassByName
//
//	@Description: 获取StorageClass
//	@param client
//	@param name
//	@return *storagev1.StorageClass
//	@return error
func GetStorageClassByName(ctx context.Context, client *kubernetes.Clientset, name string) (*storagev1.StorageClass, error) {
	sc, err := client.StorageV1().StorageClasses().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return sc, nil
}

// GetStorageClassYaml
//
//	@Description: 获取StorageClass的Yaml
//	@param client
//	@param name
//	@return string
//	@return error
func GetStorageClassYaml(ctx context.Context, client *kubernetes.Clientset, name string) (string, error) {
	sc, err := client.StorageV1().StorageClasses().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	yamlStr, err := yamlutil.MarshalWithoutManagedFields(sc)
	if err != nil {
		return "", err
	}
	return yamlStr, nil
}

// GetStorageClassByField
//
//	@Description: 获取StorageClass通过字段查询
//	@param client
//	@param fieldMap
//	@return []storagev1.StorageClass
//	@return error
func GetStorageClassByField(ctx context.Context, client *kubernetes.Clientset, fieldMap map[string]string) ([]storagev1.StorageClass, error) {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	scList, err := client.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{
		FieldSelector:  fieldSelector.String(),
		ResourceVersion: "0",
	})
	if err != nil {
		return nil, err
	}
	return scList.Items, nil
}

// GetStorageClassByLabel
//
//	@Description: 获取StorageClass通过标签查询
//	@param client
//	@param labelMap
//	@return []storagev1.StorageClass
//	@return error
func GetStorageClassByLabel(ctx context.Context, client *kubernetes.Clientset, labelMap map[string]string) ([]storagev1.StorageClass, error) {
	labelSelector := fields.SelectorFromSet(labelMap)
	scList, err := client.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{
		LabelSelector:  labelSelector.String(),
		ResourceVersion: "0",
	})
	if err != nil {
		return nil, err
	}
	return scList.Items, nil
}

// CreateStorageClass
//
//	@Description: 创建StorageClass
//	@param client
//	@param scYaml
//	@return error
func CreateStorageClass(ctx context.Context, client *kubernetes.Clientset, scYaml string) error {
	sc := &storagev1.StorageClass{}
	err := yaml.Unmarshal([]byte(scYaml), sc)
	if err != nil {
		return err
	}
	_, err = client.StorageV1().StorageClasses().Create(ctx, sc, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// UpdateStorageClass
//
//	@Description: 更新StorageClass
//	@param client
//	@param scYaml
//	@return error
func UpdateStorageClass(ctx context.Context, client *kubernetes.Clientset, scYaml string) error {
	sc := &storagev1.StorageClass{}
	err := yaml.Unmarshal([]byte(scYaml), sc)
	if err != nil {
		return err
	}
	_, err = client.StorageV1().StorageClasses().Update(ctx, sc, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// DeleteStorageClassByName
//
//	@Description: 删除StorageClass
//	@param client
//	@param name
//	@return bool
//	@return error
func DeleteStorageClassByName(ctx context.Context, client *kubernetes.Clientset, name string) error {
	err := client.StorageV1().StorageClasses().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

// DeleteStorageClassByField
//
//	@Description: 删除StorageClass通过字段查询
//	@param client
//	@param fieldMap
//	@return bool
//	@return error
func DeleteStorageClassByField(ctx context.Context, client *kubernetes.Clientset, fieldMap map[string]string) error {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	err := client.StorageV1().StorageClasses().DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return err
	}
	return nil
}

// DeleteStorageClassByLabel
//
//	@Description: 删除StorageClass通过标签查询
//	@param client
//	@param labelMap
//	@return error
func DeleteStorageClassByLabel(ctx context.Context, client *kubernetes.Clientset, labelMap map[string]string) error {
	labelSelector := fields.SelectorFromSet(labelMap)
	err := client.StorageV1().StorageClasses().DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return err
	}
	return nil
}
