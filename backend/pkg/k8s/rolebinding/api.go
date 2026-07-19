package rolebinding

import (
	"gkube/pkg/yamlutil"
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
	out, err := yamlutil.MarshalWithoutManagedFields(rb)
	if err != nil {
		return "", fmt.Errorf("failed to marshal RoleBinding to YAML: %w", err)
	}
	return string(out), nil
}

func DeleteRoleBinding(client *kubernetes.Clientset, namespace, name string) error {
	return client.RbacV1().RoleBindings(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func CreateRoleBinding(client *kubernetes.Clientset, namespace string, rbYaml string) (bool, error) {
	var rb *rbacv1.RoleBinding
	if err := yaml.Unmarshal([]byte(rbYaml), &rb); err != nil {
		return false, fmt.Errorf("YAML解析失败:%s", err.Error())
	}
	if namespace != "" {
		rb.Namespace = namespace
	}
	_, err := client.RbacV1().RoleBindings(rb.Namespace).Create(context.TODO(), rb, metav1.CreateOptions{})
	if err != nil {
		return false, fmt.Errorf("创建RoleBinding失败:%s", err.Error())
	}
	return true, nil
}

func GetRoleBindingDetail(client *kubernetes.Clientset, namespace, name string) (*rbacv1.RoleBinding, error) {
	return client.RbacV1().RoleBindings(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func UpdateRoleBinding(client *kubernetes.Clientset, namespace string, rb *rbacv1.RoleBinding) (*rbacv1.RoleBinding, error) {
	return client.RbacV1().RoleBindings(namespace).Update(context.TODO(), rb, metav1.UpdateOptions{})
}
