package vpa

import (
	"context"
	"fmt"

	"gkube/pkg/yamlutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
)

var VPAGVR = schema.GroupVersionResource{
	Group:    "autoscaling.k8s.io",
	Version:  "v1",
	Resource: "verticalpodautoscalers",
}

func GetVPAList(client dynamic.Interface, namespace string) ([]unstructured.Unstructured, error) {
	var list *unstructured.UnstructuredList
	var err error
	if namespace != "" {
		list, err = client.Resource(VPAGVR).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	} else {
		list, err = client.Resource(VPAGVR).List(context.TODO(), metav1.ListOptions{})
	}
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func GetVPADetail(client dynamic.Interface, namespace, name string) (*unstructured.Unstructured, error) {
	return client.Resource(VPAGVR).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func GetVPAYaml(client dynamic.Interface, namespace, name string) (string, error) {
	obj, err := client.Resource(VPAGVR).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	out, err := yamlutil.MarshalWithoutManagedFields(obj.Object)
	if err != nil {
		return "", fmt.Errorf("failed to marshal VPA to YAML: %w", err)
	}
	return string(out), nil
}

func CreateVPA(client dynamic.Interface, namespace, yamlContent string) error {
	obj, ns, _, err := parseVPAYaml(namespace, yamlContent)
	if err != nil {
		return err
	}
	_, err = client.Resource(VPAGVR).Namespace(ns).Create(context.TODO(), obj, metav1.CreateOptions{})
	return err
}

func UpdateVPA(client dynamic.Interface, namespace, yamlContent string) error {
	obj, ns, name, err := parseVPAYaml(namespace, yamlContent)
	if err != nil {
		return err
	}
	if obj.GetResourceVersion() == "" {
		existing, err := client.Resource(VPAGVR).Namespace(ns).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		obj.SetResourceVersion(existing.GetResourceVersion())
	}
	_, err = client.Resource(VPAGVR).Namespace(ns).Update(context.TODO(), obj, metav1.UpdateOptions{})
	return err
}

func DeleteVPA(client dynamic.Interface, namespace, name string) error {
	return client.Resource(VPAGVR).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func parseVPAYaml(namespace, yamlContent string) (*unstructured.Unstructured, string, string, error) {
	obj := make(map[string]any)
	if err := yaml.Unmarshal([]byte(yamlContent), &obj); err != nil {
		return nil, "", "", fmt.Errorf("YAML解析错误: %w", err)
	}
	unstructuredObj := &unstructured.Unstructured{Object: obj}
	unstructured.RemoveNestedField(unstructuredObj.Object, "status")

	if unstructuredObj.GetAPIVersion() == "" {
		unstructuredObj.SetAPIVersion("autoscaling.k8s.io/v1")
	}
	if unstructuredObj.GetKind() == "" {
		unstructuredObj.SetKind("VerticalPodAutoscaler")
	}

	ns := unstructuredObj.GetNamespace()
	if ns == "" {
		ns = namespace
		unstructuredObj.SetNamespace(ns)
	}
	if ns == "" {
		return nil, "", "", fmt.Errorf("namespace不能为空")
	}
	name := unstructuredObj.GetName()
	if name == "" {
		return nil, "", "", fmt.Errorf("metadata.name不能为空")
	}
	return unstructuredObj, ns, name, nil
}
