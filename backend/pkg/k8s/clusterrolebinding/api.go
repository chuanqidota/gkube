package clusterrolebinding

import (
	"context"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

func GetClusterRoleBindingList(client *kubernetes.Clientset) ([]rbacv1.ClusterRoleBinding, error) {
	crbList, err := client.RbacV1().ClusterRoleBindings().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return crbList.Items, nil
}

func GetClusterRoleBindingYaml(client *kubernetes.Clientset, name string) (string, error) {
	crb, err := client.RbacV1().ClusterRoleBindings().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	crb.TypeMeta = metav1.TypeMeta{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "ClusterRoleBinding"}
	out, err := yaml.Marshal(crb)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ClusterRoleBinding to YAML: %w", err)
	}
	return string(out), nil
}

func DeleteClusterRoleBinding(client *kubernetes.Clientset, name string) error {
	return client.RbacV1().ClusterRoleBindings().Delete(context.TODO(), name, metav1.DeleteOptions{})
}
