package secret

import (
	"context"
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// GetSecretsList
//
//	@Description: 获取secret列表
//	@param client
//	@param namespace
//	@return []corev1.Secret
//	@return error
func GetSecretsList(client *kubernetes.Clientset, namespace string) ([]corev1.Secret, error) {
	secrets, err := client.CoreV1().Secrets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return secrets.Items, nil
}

// GetSecretByName
//
//	@Description: 获取secret
//	@param client
//	@param namespace
//	@param name
//	@return *corev1.Secret
//	@return error
func GetSecretByName(client *kubernetes.Clientset, namespace, name string) (*corev1.Secret, error) {
	return client.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

// GetSecretYaml
//
//	@Description: 获取secret的yaml
//	@param client
//	@param namespace
//	@param name
//	@return string
//	@return error
func GetSecretYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	secret, err := client.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	secretJSON, err := json.Marshal(secret)
	if err != nil {
		return "", err
	}
	secretYAML, err := yaml.JSONToYAML(secretJSON)
	if err != nil {
		return "", err
	}
	return string(secretYAML), nil
}

// CreateSecret
//
//	@Description: 创建secret
//	@param client
//	@param namespace
//	@param name
//	@param data
//	@return error
func CreateSecret(client *kubernetes.Clientset, namespace, name string, data map[string]string) error {
	// 将字符串数据编码为 base64
	encodedData := make(map[string][]byte)
	for k, v := range data {
		encodedData[k] = []byte(v)
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: encodedData,
		Type: corev1.SecretTypeOpaque,
	}
	_, err := client.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// UpdateSecret
//
//	@Description: 更新secret
//	@param client
//	@param namespace
//	@param name
//	@param data
//	@return bool
//	@return error
func UpdateSecret(client *kubernetes.Clientset, namespace, name string, data map[string]string) error {
	secret, err := client.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	// 更新数据
	encodedData := make(map[string][]byte)
	for k, v := range data {
		encodedData[k] = []byte(v)
	}
	secret.Data = encodedData

	_, err = client.CoreV1().Secrets(namespace).Update(context.TODO(), secret, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// DeleteSecret
//
//	@Description: 删除secret
//	@param client
//	@param namespace
//	@param name
//	@return error
func DeleteSecret(client *kubernetes.Clientset, namespace, name string) error {
	return client.CoreV1().Secrets(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}
