package configmap

import (
	"context"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// GetConfigMapList
//
//	@Description: 获取ConfigMap列表
//	@param client
//	@param namespace
//	@return []corev1.ConfigMap
//	@return error
func GetConfigMapList(client *kubernetes.Clientset, namespace string) ([]corev1.ConfigMap, error) {
	configMaps, err := client.CoreV1().ConfigMaps(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return configMaps.Items, nil

}

// GetConfigMapByName
//
//	@Description: 获取ConfigMap
//	@param client
//	@param namespace
//	@param name
//	@return *corev1.ConfigMap
//	@return error
func GetConfigMapByName(client *kubernetes.Clientset, namespace, name string) (*corev1.ConfigMap, error) {
	return client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

// GetConfigMapYaml
//
//	@Description: 获取ConfigMap的Yaml
//	@param client
//	@param namespace
//	@param name
//	@return string
//	@return error
func GetConfigMapYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	configmap, err := client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	configmapJSON, err := json.Marshal(configmap)
	if err != nil {
		return "", err
	}
	configmapYAML, err := yaml.JSONToYAML(configmapJSON)
	if err != nil {
		return "", err
	}
	return string(configmapYAML), nil
}

// DeleteConfigMap
//
//	@Description: 删除ConfigMap
//	@param client
//	@param namespace
//	@param name
//	@return error
func DeleteConfigMap(client *kubernetes.Clientset, namespace, name string) error {
	return client.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

// UpdateConfigMapFromYaml
//
//	@Description: 通过YAML更新ConfigMap
//	@param client
//	@param namespace
//	@param yamlContent
//	@return error
func UpdateConfigMapFromYaml(client *kubernetes.Clientset, namespace, yamlContent string) error {
	if yamlContent == "" {
		return fmt.Errorf("YAML content cannot be empty")
	}
	var cm corev1.ConfigMap
	if err := yaml.Unmarshal([]byte(yamlContent), &cm); err != nil {
		return fmt.Errorf("failed to unmarshal ConfigMap YAML: %w", err)
	}
	if cm.Name == "" {
		return fmt.Errorf("ConfigMap name is required")
	}
	if cm.Namespace == "" {
		cm.Namespace = namespace
	}
	_, err := client.CoreV1().ConfigMaps(namespace).Update(context.TODO(), &cm, metav1.UpdateOptions{})
	return err
}

// CreateConfigMapFromYaml
//
//	@Description: 通过YAML创建ConfigMap
//	@param client
//	@param namespace
//	@param yamlContent
//	@return error
func CreateConfigMapFromYaml(client *kubernetes.Clientset, namespace, yamlContent string) error {
	if yamlContent == "" {
		return fmt.Errorf("YAML content cannot be empty")
	}
	var cm corev1.ConfigMap
	if err := yaml.Unmarshal([]byte(yamlContent), &cm); err != nil {
		return fmt.Errorf("failed to unmarshal ConfigMap YAML: %w", err)
	}
	if cm.Name == "" {
		return fmt.Errorf("ConfigMap name is required")
	}
	if cm.Namespace == "" {
		cm.Namespace = namespace
	}
	_, err := client.CoreV1().ConfigMaps(namespace).Create(context.TODO(), &cm, metav1.CreateOptions{})
	return err
}
