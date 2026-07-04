package namespace

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// GetNamespaceList
//
//	@Description: 获取命名空间列表
//	@param client
//	@return *corev1.NamespaceList
//	@return error
func GetNamespaceList(client *kubernetes.Clientset) (*corev1.NamespaceList, error) {
	namespace, err := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return namespace, err
}

// CreateNamespace
//
//	@Description: 创建命名空间
//	@param client
//	@param name
//	@param labels
//	@param annotations
//	@return error
func CreateNamespace(client *kubernetes.Clientset, name string, labels map[string]string, annotations map[string]string) error {
	_, err := client.CoreV1().Namespaces().Create(context.TODO(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      labels,
			Annotations: annotations,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// UpdateNamespaceLabels
//
//	@Description: 更新命名空间标签
//	@param client
//	@param name
//	@param labels
//	@return error
func UpdateNamespaceLabels(client *kubernetes.Clientset, name string, labels map[string]string) error {
	ns, err := client.CoreV1().Namespaces().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	ns.Labels = labels
	_, err = client.CoreV1().Namespaces().Update(context.TODO(), ns, metav1.UpdateOptions{})
	return err
}

// GetNamespaceDetail
//
//	@Description: 获取命名空间详情
//	@param client
//	@param name
//	@return *corev1.Namespace
//	@return error
func GetNamespaceDetail(client *kubernetes.Clientset, name string) (*corev1.Namespace, error) {
	return client.CoreV1().Namespaces().Get(context.TODO(), name, metav1.GetOptions{})
}

// GetNamespaceYaml
//
//	@Description: 获取命名空间YAML
//	@param client
//	@param name
//	@return string
//	@return error
func GetNamespaceYaml(client *kubernetes.Clientset, name string) (string, error) {
	ns, err := client.CoreV1().Namespaces().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	ns.TypeMeta = metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"}
	out, err := yaml.Marshal(ns)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Namespace to YAML: %w", err)
	}
	return string(out), nil
}

// UpdateNamespace
//
//	@Description: 更新命名空间
//	@param client
//	@param yamlContent
//	@return error
func UpdateNamespace(client *kubernetes.Clientset, yamlContent string) error {
	var ns corev1.Namespace
	if err := yaml.Unmarshal([]byte(yamlContent), &ns); err != nil {
		return fmt.Errorf("failed to unmarshal Namespace YAML: %w", err)
	}
	_, err := client.CoreV1().Namespaces().Update(context.TODO(), &ns, metav1.UpdateOptions{})
	return err
}

// DeleteNamespace
//
//	@Description: 删除命名空间
//	@param client
//	@param name
//	@return error
func DeleteNamespace(client *kubernetes.Clientset, name string) error {
	return client.CoreV1().Namespaces().Delete(context.TODO(), name, metav1.DeleteOptions{})
}
