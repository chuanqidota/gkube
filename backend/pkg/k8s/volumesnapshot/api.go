package volumesnapshot

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
)

var VolumeSnapshotGVR = schema.GroupVersionResource{
	Group:    "snapshot.storage.k8s.io",
	Version:  "v1",
	Resource: "volumesnapshots",
}

// GetVolumeSnapshotList
//
//	@Description: 获取VolumeSnapshot列表
//	@param client
//	@param namespace
//	@return []unstructured.Unstructured
//	@return error
func GetVolumeSnapshotList(client dynamic.Interface, namespace string) ([]unstructured.Unstructured, error) {
	var list *unstructured.UnstructuredList
	var err error
	if namespace != "" {
		list, err = client.Resource(VolumeSnapshotGVR).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	} else {
		list, err = client.Resource(VolumeSnapshotGVR).List(context.TODO(), metav1.ListOptions{})
	}
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetVolumeSnapshotByName
//
//	@Description: 根据名称获取VolumeSnapshot
//	@param client
//	@param namespace
//	@param name
//	@return *unstructured.Unstructured
//	@return error
func GetVolumeSnapshotByName(client dynamic.Interface, namespace, name string) (*unstructured.Unstructured, error) {
	obj, err := client.Resource(VolumeSnapshotGVR).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// GetVolumeSnapshotYaml
//
//	@Description: 获取VolumeSnapshot的YAML
//	@param client
//	@param namespace
//	@param name
//	@return string
//	@return error
func GetVolumeSnapshotYaml(client dynamic.Interface, namespace, name string) (string, error) {
	obj, err := client.Resource(VolumeSnapshotGVR).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	out, err := yaml.Marshal(obj.Object)
	if err != nil {
		return "", fmt.Errorf("failed to marshal VolumeSnapshot to YAML: %w", err)
	}
	return string(out), nil
}

// CreateVolumeSnapshot
//
//	@Description: 创建VolumeSnapshot
//	@param client
//	@param namespace
//	@param yamlContent
//	@return error
func CreateVolumeSnapshot(client dynamic.Interface, namespace, yamlContent string) error {
	obj := make(map[string]any)
	if err := yaml.Unmarshal([]byte(yamlContent), &obj); err != nil {
		return fmt.Errorf("YAML解析错误: %w", err)
	}
	unstructuredObj := &unstructured.Unstructured{Object: obj}
	_, err := client.Resource(VolumeSnapshotGVR).Namespace(namespace).Create(context.TODO(), unstructuredObj, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// UpdateVolumeSnapshot
//
//	@Description: 更新VolumeSnapshot
//	@param client
//	@param namespace
//	@param yamlContent
//	@return error
func UpdateVolumeSnapshot(client dynamic.Interface, namespace, yamlContent string) error {
	obj := make(map[string]any)
	if err := yaml.Unmarshal([]byte(yamlContent), &obj); err != nil {
		return fmt.Errorf("YAML解析错误: %w", err)
	}
	unstructuredObj := &unstructured.Unstructured{Object: obj}
	_, err := client.Resource(VolumeSnapshotGVR).Namespace(namespace).Update(context.TODO(), unstructuredObj, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// DeleteVolumeSnapshotByName
//
//	@Description: 删除VolumeSnapshot
//	@param client
//	@param namespace
//	@param name
//	@return error
func DeleteVolumeSnapshotByName(client dynamic.Interface, namespace, name string) error {
	return client.Resource(VolumeSnapshotGVR).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}
