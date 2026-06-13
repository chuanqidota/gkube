package serviceaccount

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

func GetServiceAccountList(client *kubernetes.Clientset, namespace string) ([]corev1.ServiceAccount, error) {
	saList, err := client.CoreV1().ServiceAccounts(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return saList.Items, nil
}

func GetServiceAccountYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	sa, err := client.CoreV1().ServiceAccounts(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	sa.TypeMeta = metav1.TypeMeta{APIVersion: "v1", Kind: "ServiceAccount"}
	out, err := yaml.Marshal(sa)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ServiceAccount to YAML: %w", err)
	}
	return string(out), nil
}

func DeleteServiceAccount(client *kubernetes.Clientset, namespace, name string) error {
	return client.CoreV1().ServiceAccounts(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}
