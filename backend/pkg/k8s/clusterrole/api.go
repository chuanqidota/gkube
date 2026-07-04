package clusterrole

import (
	"context"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

func GetClusterRoleList(client *kubernetes.Clientset) ([]rbacv1.ClusterRole, error) {
	crList, err := client.RbacV1().ClusterRoles().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return crList.Items, nil
}

func GetClusterRoleYaml(client *kubernetes.Clientset, name string) (string, error) {
	cr, err := client.RbacV1().ClusterRoles().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	cr.TypeMeta = metav1.TypeMeta{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "ClusterRole"}
	out, err := yaml.Marshal(cr)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ClusterRole to YAML: %w", err)
	}
	return string(out), nil
}

func DeleteClusterRole(client *kubernetes.Clientset, name string) error {
	return client.RbacV1().ClusterRoles().Delete(context.TODO(), name, metav1.DeleteOptions{})
}
