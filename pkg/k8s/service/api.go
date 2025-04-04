package service

import (
	"context"
	"fmt"

	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// GetServicesList
//
//	@Description: 获取service列表
//	@param client
//	@param namespace
//	@return []corev1.Service
//	@return error
func GetServicesList(client *kubernetes.Clientset, namespace string) ([]corev1.Service, error) {
	services, err := client.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return services.Items, nil
}

// GetServicesByName
//
//	@Description: 获取svc根据名好吃呢个
//	@param client
//	@param namespace
//	@param name
//	@return *corev1.Service
//	@return error
func GetServicesByName(client *kubernetes.Clientset, namespace, name string) (*corev1.Service, error) {
	service, err := client.CoreV1().Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return &corev1.Service{}, err
	}
	return service, nil
}

// GetServicesByLabel
//
//	@Description: 根据标签获取service列表
//	@param client
//	@param namespace
//	@param labelMap
//	@return []corev1.Service
//	@return error
func GetServicesByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) ([]corev1.Service, error) {
	labelSelector := labels.SelectorFromSet(labelMap) // 创建标签选择器
	services, err := client.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return nil, err
	}
	return services.Items, nil
}

// GetServicesByFiled
//
//	@Description: 根据字段获取service列表
//	@param client
//	@param namespace
//	@param fieldMap
//	@return []corev1.Service
//	@return error
func GetServicesByFiled(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) ([]corev1.Service, error) {
	fieldSelector := fields.SelectorFromSet(fieldMap) // 创建标签选择器
	services, err := client.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return nil, err
	}
	return services.Items, nil
}

// GetServicesYaml
//
//	@Description: 根据名称获取service的yaml
//	@param client
//	@param namespace
//	@param name
//	@return string
//	@return error
func GetServicesYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	services, err := client.CoreV1().Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", nil
	}
	servicesJSON, err := json.Marshal(services)
	if err != nil {
		return "", err
	}
	servicesYAML, err := yaml.JSONToYAML(servicesJSON)
	if err != nil {
		return "", err
	}
	return string(servicesYAML), nil
}

// CreateService
//
//	@Description: 创建service
//	@param client
//	@param namespace
//	@param serviceYAML
//	@return error
func CreateService(client *kubernetes.Clientset, namespace, serviceYAML string) error {
	var service corev1.Service
	if err := yaml.Unmarshal([]byte(serviceYAML), &service); err != nil {
		return fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	_, err := client.CoreV1().Services(namespace).Create(context.TODO(), &service, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("创建service资源失败:%s", err.Error())
	}
	return nil
}

// UpdateService
//
//	@Description: 创建service
//	@param client
//	@param serviceYAML
//	@return bool
//	@return error
func UpdateService(client *kubernetes.Clientset, serviceYAML string) error {
	var service corev1.Service
	if err := yaml.Unmarshal([]byte(serviceYAML), &service); err != nil {
		return fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	_, err := client.CoreV1().Services(service.Namespace).Update(context.TODO(), &service, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("更新service资源失败:%s", err.Error())
	}
	return nil
}

// DeleteService
//
//	@Description: 删除service
//	@param client
//	@param namespace
//	@param name
//	@return bool
//	@return error
func DeleteService(client *kubernetes.Clientset, namespace, name string) error {
	err := client.CoreV1().Services(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("删除service资源失败:%s", err.Error())
	}
	return nil
}
