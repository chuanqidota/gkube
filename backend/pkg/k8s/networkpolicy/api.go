package networkpolicy

import (
	"context"
	"fmt"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

func GetNetworkPolicyList(client *kubernetes.Clientset, namespace string) ([]networkingv1.NetworkPolicy, error) {
	npList, err := client.NetworkingV1().NetworkPolicies(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return npList.Items, nil
}

func GetNetworkPolicyYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	np, err := client.NetworkingV1().NetworkPolicies(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	np.TypeMeta = metav1.TypeMeta{
		APIVersion: "networking.k8s.io/v1",
		Kind:       "NetworkPolicy",
	}
	out, err := yaml.Marshal(np)
	if err != nil {
		return "", fmt.Errorf("failed to marshal NetworkPolicy to YAML: %w", err)
	}
	return string(out), nil
}

func CreateNetworkPolicy(client *kubernetes.Clientset, namespace, yamlContent string) error {
	var np networkingv1.NetworkPolicy
	if err := yaml.Unmarshal([]byte(yamlContent), &np); err != nil {
		return fmt.Errorf("failed to unmarshal NetworkPolicy YAML: %w", err)
	}
	if np.Namespace == "" {
		np.Namespace = namespace
	}
	_, err := client.NetworkingV1().NetworkPolicies(namespace).Create(context.TODO(), &np, metav1.CreateOptions{})
	return err
}

func UpdateNetworkPolicy(client *kubernetes.Clientset, namespace, yamlContent string) error {
	var np networkingv1.NetworkPolicy
	if err := yaml.Unmarshal([]byte(yamlContent), &np); err != nil {
		return fmt.Errorf("failed to unmarshal NetworkPolicy YAML: %w", err)
	}
	if np.Namespace == "" {
		np.Namespace = namespace
	}
	_, err := client.NetworkingV1().NetworkPolicies(namespace).Update(context.TODO(), &np, metav1.UpdateOptions{})
	return err
}

func DeleteNetworkPolicy(client *kubernetes.Clientset, namespace, name string) error {
	return client.NetworkingV1().NetworkPolicies(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}
