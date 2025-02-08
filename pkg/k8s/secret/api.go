package secret

import (
	"context"
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

func GetSecret(client *kubernetes.Clientset, namespace, name string) (*corev1.Secret, error) {
	return client.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

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


func GetSecretsList(client *kubernetes.Clientset, namespace string) ([]corev1.Secret, error) {
	secrets, err := client.CoreV1().Secrets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return secrets.Items, nil
}

func CreateSecret(client *kubernetes.Clientset, namespace, name string, data map[string]string) (bool, error) {
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
		return false, err
	}
	return true, nil
}

func CreateTLSSecret(client *kubernetes.Clientset, namespace, name, cert, key string) (bool, error) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.crt": []byte(cert),
			"tls.key": []byte(key),
		},
	}
	_, err := client.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		return false, err
	}
	return true, nil
}

func UpdateSecret(client *kubernetes.Clientset, namespace, name string, data map[string]string) (bool, error) {
	secret, err := client.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	// 更新数据
	encodedData := make(map[string][]byte)
	for k, v := range data {
		encodedData[k] = []byte(v)
	}
	secret.Data = encodedData

	_, err = client.CoreV1().Secrets(namespace).Update(context.TODO(), secret, metav1.UpdateOptions{})
	if err != nil {
		return false, err
	}
	return true, nil
}

func DeleteSecret(client *kubernetes.Clientset, namespace, name string) error {
	return client.CoreV1().Secrets(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}
