package namespace

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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
//	@param namespace
//	@return error
func CreateNamespace(client *kubernetes.Clientset, namespace string) error {
	_, err := client.CoreV1().Namespaces().Create(context.TODO(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}
