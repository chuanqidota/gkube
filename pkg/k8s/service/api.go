package service

import (
	"context"

	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

func GetServicesList(client *kubernetes.Clientset, namespace string) ([]corev1.Service, error) {
	services, err := client.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return services.Items, nil
}

func GetServicesByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) ([]corev1.Service, error) {
	labelSelector := labels.SelectorFromSet(labelMap) // 创建标签选择器
	services, err := client.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return nil, err
	}
	return services.Items, nil
}

func GetServicesyFiled(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) ([]corev1.Service, error) {
	fieldSelector := fields.SelectorFromSet(fieldMap) // 创建标签选择器
	services, err := client.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return nil, err
	}
	return services.Items, nil
}

func GetServicesYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	services, err := client.CoreV1().Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", nil
	}
	servicesJSON, err := json.Marshal(services)
	if err != nil {
		return "", err
	}
	servicesYAML, err := yaml.JSONToYAML(servicesJSON)
	if err != nil {
		return "", err
	}
	return string(servicesYAML), nil
}
