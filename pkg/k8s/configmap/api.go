package configmap

import (
	"context"
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// GetConfigMap
//
//	@Description: 获取ConfigMap
//	@param client
//	@param namespace
//	@param name
//	@return *corev1.ConfigMap
//	@return error
func GetConfigMap(client *kubernetes.Clientset, namespace, name string) (*corev1.ConfigMap, error) {
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

// CreateConfigMap
//
//	@Description: 创建ConfigMap
//	@param client
//	@param namespace
//	@param name
//	@param data
//	@return bool
//	@return error
func CreateConfigMap(client *kubernetes.Clientset, namespace, name string, data map[string]string) (bool, error) {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: data,
	}
	_, err := client.CoreV1().ConfigMaps(namespace).Create(context.TODO(), cm, metav1.CreateOptions{})
	if err != nil {
		return false, err
	}
	return true, nil
}

// UpdateConfigMap
//
//	@Description: 更新ConfigMap
//	@param client
//	@param namespace
//	@param name
//	@param data
//	@return bool
//	@return error
func UpdateConfigMap(client *kubernetes.Clientset, namespace, name string, data map[string]string) (bool, error) {
	cm, err := client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	for key, value := range data {
		cm.Data[key] = value
	}
	_, err = client.CoreV1().ConfigMaps(namespace).Update(context.TODO(), cm, metav1.UpdateOptions{})
	if err != nil {
		return false, err
	}
	return true, nil
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
