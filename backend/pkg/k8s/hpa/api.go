package hpa

import (
	"gkube/pkg/yamlutil"
	"context"
	"fmt"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

func GetHPAList(client *kubernetes.Clientset, namespace string) ([]autoscalingv2.HorizontalPodAutoscaler, error) {
	hpaList, err := client.AutoscalingV2().HorizontalPodAutoscalers(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return hpaList.Items, nil
}

func GetHPAYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	hpa, err := client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	hpa.TypeMeta = metav1.TypeMeta{
		APIVersion: "autoscaling/v2",
		Kind:       "HorizontalPodAutoscaler",
	}
	out, err := yamlutil.MarshalWithoutManagedFields(hpa)
	if err != nil {
		return "", fmt.Errorf("failed to marshal HPA to YAML: %w", err)
	}
	return string(out), nil
}

func CreateHPA(client *kubernetes.Clientset, namespace, yamlContent string) error {
	var hpa autoscalingv2.HorizontalPodAutoscaler
	if err := yaml.Unmarshal([]byte(yamlContent), &hpa); err != nil {
		return fmt.Errorf("failed to unmarshal HPA YAML: %w", err)
	}
	if hpa.Namespace == "" {
		hpa.Namespace = namespace
	}
	_, err := client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Create(context.TODO(), &hpa, metav1.CreateOptions{})
	return err
}

func UpdateHPA(client *kubernetes.Clientset, namespace, yamlContent string) error {
	var hpa autoscalingv2.HorizontalPodAutoscaler
	if err := yaml.Unmarshal([]byte(yamlContent), &hpa); err != nil {
		return fmt.Errorf("failed to unmarshal HPA YAML: %w", err)
	}
	if hpa.Namespace == "" {
		hpa.Namespace = namespace
	}
	_, err := client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Update(context.TODO(), &hpa, metav1.UpdateOptions{})
	return err
}

func DeleteHPA(client *kubernetes.Clientset, namespace, name string) error {
	return client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}
