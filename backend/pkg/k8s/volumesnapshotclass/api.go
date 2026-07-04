package volumesnapshotclass

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
)

var VolumeSnapshotClassGVR = schema.GroupVersionResource{
	Group:    "snapshot.storage.k8s.io",
	Version:  "v1",
	Resource: "volumesnapshotclasses",
}

// GetVolumeSnapshotClassList
//
//	@Description: 获取VolumeSnapshotClass列表
//	@param client
//	@return []unstructured.Unstructured
//	@return error
func GetVolumeSnapshotClassList(client dynamic.Interface) ([]unstructured.Unstructured, error) {
	list, err := client.Resource(VolumeSnapshotClassGVR).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetVolumeSnapshotClassByName
//
//	@Description: 根据名称获取VolumeSnapshotClass
//	@param client
//	@param name
//	@return *unstructured.Unstructured
//	@return error
func GetVolumeSnapshotClassByName(client dynamic.Interface, name string) (*unstructured.Unstructured, error) {
	obj, err := client.Resource(VolumeSnapshotClassGVR).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// GetVolumeSnapshotClassYaml
//
//	@Description: 获取VolumeSnapshotClass的YAML
//	@param client
//	@param name
//	@return string
//	@return error
func GetVolumeSnapshotClassYaml(client dynamic.Interface, name string) (string, error) {
	obj, err := client.Resource(VolumeSnapshotClassGVR).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	out, err := yaml.Marshal(obj.Object)
	if err != nil {
		return "", fmt.Errorf("failed to marshal VolumeSnapshotClass to YAML: %w", err)
	}
	return string(out), nil
}

// CreateVolumeSnapshotClass
//
//	@Description: 创建VolumeSnapshotClass
//	@param client
//	@param yamlContent
//	@return error
func CreateVolumeSnapshotClass(client dynamic.Interface, yamlContent string) error {
	obj := make(map[string]any)
	if err := yaml.Unmarshal([]byte(yamlContent), &obj); err != nil {
		return fmt.Errorf("YAML解析错误: %w", err)
	}
	unstructuredObj := &unstructured.Unstructured{Object: obj}
	_, err := client.Resource(VolumeSnapshotClassGVR).Create(context.TODO(), unstructuredObj, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// UpdateVolumeSnapshotClass
//
//	@Description: 更新VolumeSnapshotClass
//	@param client
//	@param yamlContent
//	@return error
func UpdateVolumeSnapshotClass(client dynamic.Interface, yamlContent string) error {
	obj := make(map[string]any)
	if err := yaml.Unmarshal([]byte(yamlContent), &obj); err != nil {
		return fmt.Errorf("YAML解析错误: %w", err)
	}
	unstructuredObj := &unstructured.Unstructured{Object: obj}
	_, err := client.Resource(VolumeSnapshotClassGVR).Update(context.TODO(), unstructuredObj, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// DeleteVolumeSnapshotClassByName
//
//	@Description: 删除VolumeSnapshotClass
//	@param client
//	@param name
//	@return error
func DeleteVolumeSnapshotClassByName(client dynamic.Interface, name string) error {
	return client.Resource(VolumeSnapshotClassGVR).Delete(context.TODO(), name, metav1.DeleteOptions{})
}
