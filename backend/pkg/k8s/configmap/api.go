package configmap

import (
	"context"
	"fmt"

	"gkube/pkg/yamlutil"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/yaml"
)

// GetConfigMapList
//
//	@Description: 获取ConfigMap列表
//	@param client
//	@param namespace
//	@param limit
//	@param continueToken
//	@param labelSelector
//	@return *corev1.ConfigMapList
//	@return error
func GetConfigMapList(client *kubernetes.Clientset, namespace string, limit int64, continueToken, labelSelector string) (*corev1.ConfigMapList, error) {
	listOpts := metav1.ListOptions{}
	if limit > 0 {
		listOpts.Limit = limit
	}
	if continueToken != "" {
		listOpts.Continue = continueToken
	}
	if labelSelector != "" {
		listOpts.LabelSelector = labelSelector
	}
	return client.CoreV1().ConfigMaps(namespace).List(context.TODO(), listOpts)
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
	configmap.TypeMeta = metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"}
	configmapYAML, err := yamlutil.MarshalWithoutManagedFields(configmap)
	if err != nil {
		return "", err
	}
	return configmapYAML, nil
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
	if cm.Kind != "" && cm.Kind != "ConfigMap" {
		return fmt.Errorf("YAML kind is %q, expected ConfigMap", cm.Kind)
	}
	if cm.APIVersion != "" && cm.APIVersion != "v1" {
		return fmt.Errorf("YAML apiVersion is %q, expected v1", cm.APIVersion)
	}
	cm.Namespace = namespace
	// RetryOnConflict: re-Get latest object to obtain fresh resourceVersion before Update
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), cm.Name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get ConfigMap: %w", err)
		}
		latest.Data = cm.Data
		latest.BinaryData = cm.BinaryData
		latest.Labels = cm.Labels
		latest.Annotations = cm.Annotations
		if cm.Immutable != nil {
			latest.Immutable = cm.Immutable
		}
		_, err = client.CoreV1().ConfigMaps(namespace).Update(context.TODO(), latest, metav1.UpdateOptions{})
		return err
	})
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
	if cm.Kind != "" && cm.Kind != "ConfigMap" {
		return fmt.Errorf("YAML kind is %q, expected ConfigMap", cm.Kind)
	}
	if cm.APIVersion != "" && cm.APIVersion != "v1" {
		return fmt.Errorf("YAML apiVersion is %q, expected v1", cm.APIVersion)
	}
	cm.Namespace = namespace
	_, err := client.CoreV1().ConfigMaps(namespace).Create(context.TODO(), &cm, metav1.CreateOptions{})
	return err
}
