package volumesnapshotclass

import (
	"context"
	"gkube/pkg/yamlutil"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/retry"
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
func GetVolumeSnapshotClassList(ctx context.Context, client dynamic.Interface, labelSelector string) ([]unstructured.Unstructured, error) {
	listOpts := metav1.ListOptions{ResourceVersion: "0"}
	if labelSelector != "" {
		listOpts.LabelSelector = labelSelector
	}
	result, err := client.Resource(VolumeSnapshotClassGVR).List(ctx, listOpts)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

// GetVolumeSnapshotClassByName
//
//	@Description: 根据名称获取VolumeSnapshotClass
//	@param client
//	@param name
//	@return *unstructured.Unstructured
//	@return error
func GetVolumeSnapshotClassByName(ctx context.Context, client dynamic.Interface, name string) (*unstructured.Unstructured, error) {
	obj, err := client.Resource(VolumeSnapshotClassGVR).Get(ctx, name, metav1.GetOptions{})
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
func GetVolumeSnapshotClassYaml(ctx context.Context, client dynamic.Interface, name string) (string, error) {
	obj, err := client.Resource(VolumeSnapshotClassGVR).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	out, err := yamlutil.MarshalWithoutManagedFields(obj.Object)
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
func CreateVolumeSnapshotClass(ctx context.Context, client dynamic.Interface, yamlContent string) error {
	obj := make(map[string]any)
	if err := yaml.Unmarshal([]byte(yamlContent), &obj); err != nil {
		return fmt.Errorf("YAML解析错误: %w", err)
	}
	unstructuredObj := &unstructured.Unstructured{Object: obj}
	_, err := client.Resource(VolumeSnapshotClassGVR).Create(ctx, unstructuredObj, metav1.CreateOptions{})
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
func UpdateVolumeSnapshotClass(ctx context.Context, client dynamic.Interface, yamlContent string) error {
	obj := make(map[string]any)
	if err := yaml.Unmarshal([]byte(yamlContent), &obj); err != nil {
		return fmt.Errorf("YAML解析错误: %w", err)
	}
	name, found, err := unstructured.NestedString(obj, "metadata", "name")
	if err != nil || !found {
		return fmt.Errorf("metadata.name is required")
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := client.Resource(VolumeSnapshotClassGVR).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		unstructuredObj := &unstructured.Unstructured{Object: obj}
		unstructuredObj.SetResourceVersion(latest.GetResourceVersion())
		_, err = client.Resource(VolumeSnapshotClassGVR).Update(ctx, unstructuredObj, metav1.UpdateOptions{})
		return err
	})
}

// DeleteVolumeSnapshotClassByName
//
//	@Description: 删除VolumeSnapshotClass
//	@param client
//	@param name
//	@return error
func DeleteVolumeSnapshotClassByName(ctx context.Context, client dynamic.Interface, name string) error {
	return client.Resource(VolumeSnapshotClassGVR).Delete(ctx, name, metav1.DeleteOptions{})
}
