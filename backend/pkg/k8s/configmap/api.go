package configmap

import (
	"context"

	apperr "gkube/pkg/errors"
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
func GetConfigMapList(ctx context.Context, client *kubernetes.Clientset, namespace string, limit int64, continueToken, labelSelector string) (*corev1.ConfigMapList, error) {
	listOpts := metav1.ListOptions{ResourceVersion: "0"}
	if limit > 0 {
		listOpts.Limit = limit
	}
	if continueToken != "" {
		listOpts.Continue = continueToken
	}
	if labelSelector != "" {
		listOpts.LabelSelector = labelSelector
	}
	return client.CoreV1().ConfigMaps(namespace).List(ctx, listOpts)
}

// GetConfigMapByName
//
//	@Description: 获取ConfigMap
//	@param client
//	@param namespace
//	@param name
//	@return *corev1.ConfigMap
//	@return error
func GetConfigMapByName(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (*corev1.ConfigMap, error) {
	return client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
}

// GetConfigMapYaml
//
//	@Description: 获取ConfigMap的Yaml
//	@param client
//	@param namespace
//	@param name
//	@return string
//	@return error
func GetConfigMapYaml(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (string, error) {
	configmap, err := client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
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
func DeleteConfigMap(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error {
	return client.CoreV1().ConfigMaps(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// UpdateConfigMapFromYaml
//
//	@Description: 通过YAML更新ConfigMap
//	@param client
//	@param namespace
//	@param yamlContent
//	@return error
func UpdateConfigMapFromYaml(ctx context.Context, client *kubernetes.Clientset, namespace, yamlContent string) error {
	if yamlContent == "" {
		return apperr.Validation("YAML内容不能为空", nil)
	}
	var cm corev1.ConfigMap
	if err := yaml.Unmarshal([]byte(yamlContent), &cm); err != nil {
		return apperr.Validation("yaml解析失败", err)
	}
	if cm.Name == "" {
		return apperr.Validation("ConfigMap名称不能为空", nil)
	}
	if cm.Kind != "" && cm.Kind != "ConfigMap" {
		return apperr.Validation("YAML kind必须为ConfigMap", nil)
	}
	if cm.APIVersion != "" && cm.APIVersion != "v1" {
		return apperr.Validation("YAML apiVersion必须为v1", nil)
	}
	cm.Namespace = namespace
	// RetryOnConflict: re-Get latest object to obtain fresh resourceVersion before Update
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := client.CoreV1().ConfigMaps(namespace).Get(ctx, cm.Name, metav1.GetOptions{})
		if err != nil {
			return apperr.K8sAPIFail("获取ConfigMap失败", err)
		}
		latest.Data = cm.Data
		latest.BinaryData = cm.BinaryData
		latest.Labels = cm.Labels
		latest.Annotations = cm.Annotations
		if cm.Immutable != nil {
			latest.Immutable = cm.Immutable
		}
		_, err = client.CoreV1().ConfigMaps(namespace).Update(ctx, latest, metav1.UpdateOptions{})
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
func CreateConfigMapFromYaml(ctx context.Context, client *kubernetes.Clientset, namespace, yamlContent string) error {
	if yamlContent == "" {
		return apperr.Validation("YAML内容不能为空", nil)
	}
	var cm corev1.ConfigMap
	if err := yaml.Unmarshal([]byte(yamlContent), &cm); err != nil {
		return apperr.Validation("yaml解析失败", err)
	}
	if cm.Name == "" {
		return apperr.Validation("ConfigMap名称不能为空", nil)
	}
	if cm.Kind != "" && cm.Kind != "ConfigMap" {
		return apperr.Validation("YAML kind必须为ConfigMap", nil)
	}
	if cm.APIVersion != "" && cm.APIVersion != "v1" {
		return apperr.Validation("YAML apiVersion必须为v1", nil)
	}
	cm.Namespace = namespace
	_, err := client.CoreV1().ConfigMaps(namespace).Create(ctx, &cm, metav1.CreateOptions{})
	return err
}
