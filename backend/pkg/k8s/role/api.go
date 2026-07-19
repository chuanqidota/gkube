package role

import (
	"gkube/pkg/yamlutil"
	"context"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
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

func CreateRole(client *kubernetes.Clientset, namespace string, roleYaml string) (bool, error) {
	var role *rbacv1.Role
	if err := yaml.Unmarshal([]byte(roleYaml), &role); err != nil {
		return false, fmt.Errorf("YAML解析失败:%s", err.Error())
	}
	if namespace != "" {
		role.Namespace = namespace
	}
	_, err := client.RbacV1().Roles(role.Namespace).Create(context.TODO(), role, metav1.CreateOptions{})
	if err != nil {
		return false, fmt.Errorf("创建Role失败:%s", err.Error())
	}
	return true, nil
}

func GetRoleDetail(client *kubernetes.Clientset, namespace, name string) (*rbacv1.Role, error) {
	return client.RbacV1().Roles(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func UpdateRole(client *kubernetes.Clientset, namespace string, role *rbacv1.Role) (*rbacv1.Role, error) {
	return client.RbacV1().Roles(namespace).Update(context.TODO(), role, metav1.UpdateOptions{})
}
