package crd

import (
	"context"
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"
)

func GetCRDList(client *apiextensionsclientset.Clientset) ([]apiextensionsv1.CustomResourceDefinition, error) {
	crdList, err := client.ApiextensionsV1().CustomResourceDefinitions().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return crdList.Items, nil
}

func GetCRDDetail(client *apiextensionsclientset.Clientset, name string) (*apiextensionsv1.CustomResourceDefinition, error) {
	return client.ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), name, metav1.GetOptions{})
}

func GetCRDYaml(client *apiextensionsclientset.Clientset, name string) (string, error) {
	crd, err := client.ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	crd.TypeMeta = metav1.TypeMeta{APIVersion: "apiextensions.k8s.io/v1", Kind: "CustomResourceDefinition"}
	out, err := yaml.Marshal(crd)
	if err != nil {
		return "", fmt.Errorf("failed to marshal CRD to YAML: %w", err)
	}
	return string(out), nil
}

func GetCustomResourceList(config *rest.Config, gvr schema.GroupVersionResource, namespace string) ([]unstructured.Unstructured, error) {
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	var list *unstructured.UnstructuredList
	if namespace != "" {
		list, err = dynamicClient.Resource(gvr).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	} else {
		list, err = dynamicClient.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
	}
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func GetCustomResourceYaml(config *rest.Config, gvr schema.GroupVersionResource, namespace, name string) (string, error) {
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return "", err
	}
	var obj *unstructured.Unstructured
	if namespace != "" {
		obj, err = dynamicClient.Resource(gvr).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	} else {
		obj, err = dynamicClient.Resource(gvr).Get(context.TODO(), name, metav1.GetOptions{})
	}
	if err != nil {
		return "", err
	}
	out, err := yaml.Marshal(obj.Object)
	if err != nil {
		return "", fmt.Errorf("failed to marshal custom resource to YAML: %w", err)
	}
	return string(out), nil
}

func DeleteCustomResource(config *rest.Config, gvr schema.GroupVersionResource, namespace, name string) error {
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}
	if namespace != "" {
		return dynamicClient.Resource(gvr).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	}
	return dynamicClient.Resource(gvr).Delete(context.TODO(), name, metav1.DeleteOptions{})
}
