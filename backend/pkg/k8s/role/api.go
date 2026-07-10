package role

import (
	"gkube/pkg/yamlutil"
	"context"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetRoleList(client *kubernetes.Clientset, namespace string) ([]rbacv1.Role, error) {
	roleList, err := client.RbacV1().Roles(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return roleList.Items, nil
}

func GetRoleYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	role, err := client.RbacV1().Roles(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	role.TypeMeta = metav1.TypeMeta{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "Role"}
	out, err := yamlutil.MarshalWithoutManagedFields(role)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Role to YAML: %w", err)
	}
	return string(out), nil
}

func DeleteRole(client *kubernetes.Clientset, namespace, name string) error {
	return client.RbacV1().Roles(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}
