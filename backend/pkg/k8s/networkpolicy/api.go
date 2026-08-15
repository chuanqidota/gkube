package networkpolicy

import (
	"gkube/pkg/yamlutil"
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
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
	out, err := yamlutil.MarshalWithoutManagedFields(np)
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
	// Fetch the existing object to get resourceVersion (required by K8s for optimistic concurrency)
	existing, err := client.NetworkingV1().NetworkPolicies(namespace).Get(context.TODO(), np.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get existing NetworkPolicy: %w", err)
	}
	np.ResourceVersion = existing.ResourceVersion
	_, err = client.NetworkingV1().NetworkPolicies(namespace).Update(context.TODO(), &np, metav1.UpdateOptions{})
	return err
}

func DeleteNetworkPolicy(client *kubernetes.Clientset, namespace, name string) error {
	return client.NetworkingV1().NetworkPolicies(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func GetNetworkPolicyDetail(client *kubernetes.Clientset, namespace, name string) (*networkingv1.NetworkPolicy, error) {
	return client.NetworkingV1().NetworkPolicies(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

// GetNetworkPolicyPods returns pods matched by the NetworkPolicy's podSelector.
// Properly handles both MatchLabels and MatchExpressions.
func GetNetworkPolicyPods(client *kubernetes.Clientset, namespace, name string) (*corev1.PodList, error) {
	np, err := client.NetworkingV1().NetworkPolicies(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取NetworkPolicy资源失败:%s", err.Error())
	}
	// Use the full label selector (MatchLabels + MatchExpressions)
	selector, err := metav1.LabelSelectorAsSelector(&np.Spec.PodSelector)
	if err != nil {
		return nil, fmt.Errorf("解析PodSelector失败:%s", err.Error())
	}
	podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("获取NetworkPolicy关联pod列表失败:%s", err.Error())
	}
	return podList, nil
}
