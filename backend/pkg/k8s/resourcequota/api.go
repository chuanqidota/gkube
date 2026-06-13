package resourcequota

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

func GetResourceQuotaList(client *kubernetes.Clientset, namespace string) ([]corev1.ResourceQuota, error) {
	rqList, err := client.CoreV1().ResourceQuotas(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return rqList.Items, nil
}

func GetResourceQuotaYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	rq, err := client.CoreV1().ResourceQuotas(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	rq.TypeMeta = metav1.TypeMeta{APIVersion: "v1", Kind: "ResourceQuota"}
	out, err := yaml.Marshal(rq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ResourceQuota to YAML: %w", err)
	}
	return string(out), nil
}

func CreateResourceQuota(client *kubernetes.Clientset, namespace, yamlContent string) error {
	var rq corev1.ResourceQuota
	if err := yaml.Unmarshal([]byte(yamlContent), &rq); err != nil {
		return fmt.Errorf("failed to unmarshal ResourceQuota YAML: %w", err)
	}
	if rq.Namespace == "" {
		rq.Namespace = namespace
	}
	_, err := client.CoreV1().ResourceQuotas(namespace).Create(context.TODO(), &rq, metav1.CreateOptions{})
	return err
}

func UpdateResourceQuota(client *kubernetes.Clientset, namespace, yamlContent string) error {
	var rq corev1.ResourceQuota
	if err := yaml.Unmarshal([]byte(yamlContent), &rq); err != nil {
		return fmt.Errorf("failed to unmarshal ResourceQuota YAML: %w", err)
	}
	if rq.Namespace == "" {
		rq.Namespace = namespace
	}
	_, err := client.CoreV1().ResourceQuotas(namespace).Update(context.TODO(), &rq, metav1.UpdateOptions{})
	return err
}

func DeleteResourceQuota(client *kubernetes.Clientset, namespace, name string) error {
	return client.CoreV1().ResourceQuotas(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}
