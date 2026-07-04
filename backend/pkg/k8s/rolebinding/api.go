package rolebinding

import (
	"context"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

func GetRoleBindingList(client *kubernetes.Clientset, namespace string) ([]rbacv1.RoleBinding, error) {
	rbList, err := client.RbacV1().RoleBindings(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return rbList.Items, nil
}

func GetRoleBindingYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	rb, err := client.RbacV1().RoleBindings(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	rb.TypeMeta = metav1.TypeMeta{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "RoleBinding"}
	out, err := yaml.Marshal(rb)
	if err != nil {
		return "", fmt.Errorf("failed to marshal RoleBinding to YAML: %w", err)
	}
	return string(out), nil
}

func DeleteRoleBinding(client *kubernetes.Clientset, namespace, name string) error {
	return client.RbacV1().RoleBindings(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}
