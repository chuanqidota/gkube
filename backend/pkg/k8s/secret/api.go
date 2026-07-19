package secret

import (
	"context"
	"fmt"

	"gkube/pkg/yamlutil"

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
	secret.TypeMeta = metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"}
	secretYAML, err := yamlutil.MarshalWithoutManagedFields(secret)
	if err != nil {
		return "", err
	}
	return secretYAML, nil
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

// UpdateSecretFromYaml
//
//	@Description: 通过YAML更新Secret
//	@param client
//	@param namespace
//	@param yamlContent
//	@return error
func UpdateSecretFromYaml(client *kubernetes.Clientset, namespace, yamlContent string) error {
	if yamlContent == "" {
		return fmt.Errorf("YAML content cannot be empty")
	}
	var secret corev1.Secret
	if err := yaml.Unmarshal([]byte(yamlContent), &secret); err != nil {
		return fmt.Errorf("failed to unmarshal Secret YAML: %w", err)
	}
	if secret.Name == "" {
		return fmt.Errorf("Secret name is required")
	}
	if secret.Namespace == "" {
		secret.Namespace = namespace
	}
	_, err := client.CoreV1().Secrets(namespace).Update(context.TODO(), &secret, metav1.UpdateOptions{})
	return err
}

// CreateSecretFromYaml
//
//	@Description: 通过YAML创建Secret
//	@param client
//	@param namespace
//	@param yamlContent
//	@return error
func CreateSecretFromYaml(client *kubernetes.Clientset, namespace, yamlContent string) error {
	if yamlContent == "" {
		return fmt.Errorf("YAML content cannot be empty")
	}
	var secret corev1.Secret
	if err := yaml.Unmarshal([]byte(yamlContent), &secret); err != nil {
		return fmt.Errorf("failed to unmarshal Secret YAML: %w", err)
	}
	if secret.Name == "" {
		return fmt.Errorf("Secret name is required")
	}
	if secret.Namespace == "" {
		secret.Namespace = namespace
	}
	_, err := client.CoreV1().Secrets(namespace).Create(context.TODO(), &secret, metav1.CreateOptions{})
	return err
}
