package crd

import (
	"context"
	"gkube/pkg/yamlutil"
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/yaml"
)

func GetCRDList(ctx context.Context, client *apiextensionsclientset.Clientset, labelSelector string) ([]apiextensionsv1.CustomResourceDefinition, error) {
	listOpts := metav1.ListOptions{ResourceVersion: "0"}
	if labelSelector != "" {
		listOpts.LabelSelector = labelSelector
	}
	crdList, err := client.ApiextensionsV1().CustomResourceDefinitions().List(ctx, listOpts)
	if err != nil {
		return nil, err
	}
	return crdList.Items, nil
}

func GetCRDDetail(ctx context.Context, client *apiextensionsclientset.Clientset, name string) (*apiextensionsv1.CustomResourceDefinition, error) {
	return client.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, name, metav1.GetOptions{})
}

func GetCRDYaml(ctx context.Context, client *apiextensionsclientset.Clientset, name string) (string, error) {
	crd, err := client.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	crd.TypeMeta = metav1.TypeMeta{APIVersion: "apiextensions.k8s.io/v1", Kind: "CustomResourceDefinition"}
	out, err := yamlutil.MarshalWithoutManagedFields(crd)
	if err != nil {
		return "", fmt.Errorf("failed to marshal CRD to YAML: %w", err)
	}
	return string(out), nil
}

func GetCustomResourceList(ctx context.Context, config *rest.Config, gvr schema.GroupVersionResource, namespace string, labelSelector string) ([]unstructured.Unstructured, error) {
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	listOpts := metav1.ListOptions{ResourceVersion: "0"}
	if labelSelector != "" {
		listOpts.LabelSelector = labelSelector
	}
	var list *unstructured.UnstructuredList
	if namespace != "" {
		list, err = dynamicClient.Resource(gvr).Namespace(namespace).List(ctx, listOpts)
	} else {
		list, err = dynamicClient.Resource(gvr).List(ctx, listOpts)
	}
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func GetCustomResourceYaml(ctx context.Context, config *rest.Config, gvr schema.GroupVersionResource, namespace, name string) (string, error) {
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return "", err
	}
	var obj *unstructured.Unstructured
	if namespace != "" {
		obj, err = dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	} else {
		obj, err = dynamicClient.Resource(gvr).Get(ctx, name, metav1.GetOptions{})
	}
	if err != nil {
		return "", err
	}
	out, err := yamlutil.MarshalWithoutManagedFields(obj.Object)
	if err != nil {
		return "", fmt.Errorf("failed to marshal custom resource to YAML: %w", err)
	}
	return string(out), nil
}

func GetCustomResourceDetail(ctx context.Context, config *rest.Config, gvr schema.GroupVersionResource, namespace, name string) (map[string]any, error) {
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	var obj *unstructured.Unstructured
	if namespace != "" {
		obj, err = dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	} else {
		obj, err = dynamicClient.Resource(gvr).Get(ctx, name, metav1.GetOptions{})
	}
	if err != nil {
		return nil, err
	}
	return obj.Object, nil
}

func DeleteCustomResource(ctx context.Context, config *rest.Config, gvr schema.GroupVersionResource, namespace, name string) error {
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}
	if namespace != "" {
		return dynamicClient.Resource(gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	}
	return dynamicClient.Resource(gvr).Delete(ctx, name, metav1.DeleteOptions{})
}

func CreateCRD(ctx context.Context, client *apiextensionsclientset.Clientset, yamlContent string) error {
	var crd apiextensionsv1.CustomResourceDefinition
	if err := yaml.Unmarshal([]byte(yamlContent), &crd); err != nil {
		return fmt.Errorf("failed to unmarshal CRD YAML: %w", err)
	}
	_, err := client.ApiextensionsV1().CustomResourceDefinitions().Create(ctx, &crd, metav1.CreateOptions{})
	return err
}

func UpdateCRD(ctx context.Context, client *apiextensionsclientset.Clientset, yamlContent string) error {
	var crd apiextensionsv1.CustomResourceDefinition
	if err := yaml.Unmarshal([]byte(yamlContent), &crd); err != nil {
		return fmt.Errorf("failed to unmarshal CRD YAML: %w", err)
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := client.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crd.Name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get CRD: %w", err)
		}
		latest.Spec = crd.Spec
		latest.Labels = crd.Labels
		latest.Annotations = crd.Annotations
		_, err = client.ApiextensionsV1().CustomResourceDefinitions().Update(ctx, latest, metav1.UpdateOptions{})
		return err
	})
}

func DeleteCRD(ctx context.Context, client *apiextensionsclientset.Clientset, name string) error {
	return client.ApiextensionsV1().CustomResourceDefinitions().Delete(ctx, name, metav1.DeleteOptions{})
}

func CreateCustomResource(ctx context.Context, config *rest.Config, gvr schema.GroupVersionResource, namespace, yamlContent string) error {
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}
	obj := make(map[string]any)
	if err := yaml.Unmarshal([]byte(yamlContent), &obj); err != nil {
		return fmt.Errorf("failed to unmarshal custom resource YAML: %w", err)
	}
	unstructuredObj := &unstructured.Unstructured{Object: obj}
	if namespace != "" {
		_, err = dynamicClient.Resource(gvr).Namespace(namespace).Create(ctx, unstructuredObj, metav1.CreateOptions{})
	} else {
		_, err = dynamicClient.Resource(gvr).Create(ctx, unstructuredObj, metav1.CreateOptions{})
	}
	return err
}

func UpdateDynamicResource(ctx context.Context, client dynamic.Interface, gvr schema.GroupVersionResource, namespace, yamlContent string) (*unstructured.Unstructured, error) {
	obj := make(map[string]any)
	if err := yaml.Unmarshal([]byte(yamlContent), &obj); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom resource YAML: %w", err)
	}
	name, found, err := unstructured.NestedString(obj, "metadata", "name")
	if err != nil || !found {
		return nil, fmt.Errorf("metadata.name is required")
	}

	var result *unstructured.Unstructured
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var latest *unstructured.Unstructured
		var getErr error
		if namespace != "" {
			latest, getErr = client.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
		} else {
			latest, getErr = client.Resource(gvr).Get(ctx, name, metav1.GetOptions{})
		}
		if getErr != nil {
			return getErr
		}
		unstructuredObj := &unstructured.Unstructured{Object: obj}
		unstructured.RemoveNestedField(unstructuredObj.Object, "status")
		unstructuredObj.SetResourceVersion(latest.GetResourceVersion())

		if namespace != "" {
			result, getErr = client.Resource(gvr).Namespace(namespace).Update(ctx, unstructuredObj, metav1.UpdateOptions{})
		} else {
			result, getErr = client.Resource(gvr).Update(ctx, unstructuredObj, metav1.UpdateOptions{})
		}
		return getErr
	})
	return result, err
}

func PatchDynamicResource(ctx context.Context, client dynamic.Interface, gvr schema.GroupVersionResource, namespace, name, patchData string, patchType string) (*unstructured.Unstructured, error) {
	pt := types.StrategicMergePatchType
	switch patchType {
	case "merge":
		pt = types.MergePatchType
	case "json":
		pt = types.JSONPatchType
	case "strategic":
		pt = types.StrategicMergePatchType
	default:
		return nil, fmt.Errorf("unsupported patch type: %s (supported: strategic, merge, json)", patchType)
	}

	var result *unstructured.Unstructured
	var err error
	if namespace != "" {
		result, err = client.Resource(gvr).Namespace(namespace).Patch(ctx, name, pt, []byte(patchData), metav1.PatchOptions{})
	} else {
		result, err = client.Resource(gvr).Patch(ctx, name, pt, []byte(patchData), metav1.PatchOptions{})
	}
	return result, err
}
