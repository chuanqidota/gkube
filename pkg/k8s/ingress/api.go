package ingress

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"

	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// GetIngressList
//
//	@Description: 获取ingress列表
//	@param client
//	@param namespace
//	@return []netv1.Ingress
//	@return error
func GetIngressList(client *kubernetes.Clientset, namespace string) ([]netv1.Ingress, error) {
	ingress, err := client.NetworkingV1().Ingresses(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return ingress.Items, nil
}

// GetIngressByName
//
//	@Description:
//	@param client
//	@param namespace
//	@param name
//	@return netv1.Ingress
//	@return error
func GetIngressByName(client *kubernetes.Clientset, namespace, name string) (*netv1.Ingress, error) {
	ingress, err := client.NetworkingV1().Ingresses(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return &netv1.Ingress{}, err
	}
	return ingress, nil
}

// GetIngressByLabel
//
//	@Description: 根据label获取ingress
//	@param client
//	@param namespace
//	@param labelMap
//	@return []netv1.Ingress
//	@return error
func GetIngressByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) ([]netv1.Ingress, error) {
	labelSelector := labels.SelectorFromSet(labelMap) // 创建标签选择器
	ingress, err := client.NetworkingV1().Ingresses(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return nil, err
	}
	return ingress.Items, nil
}

// GetIngressByFiled
//
//	@Description: 根据字段获取ingress
//	@param client
//	@param namespace
//	@param fieldMap
//	@return []netv1.Ingress
//	@return error
func GetIngressByFiled(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) ([]netv1.Ingress, error) {
	fieldSelector := fields.SelectorFromSet(fieldMap) // 创建标签选择器
	ingress, err := client.NetworkingV1().Ingresses(namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return nil, err
	}
	return ingress.Items, nil
}

// GetIngressYaml
//
//	@Description: 获取ingress yaml
//	@param client
//	@param namespace
//	@param name
//	@return string
//	@return error
func GetIngressYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	ingress, err := client.NetworkingV1().Ingresses(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", nil
	}
	ingressJSON, err := json.Marshal(ingress)
	if err != nil {
		return "", err
	}
	ingressYAML, err := yaml.JSONToYAML(ingressJSON)
	if err != nil {
		return "", err
	}
	return string(ingressYAML), nil
}

// CreateIngress
//
//	@Description: 创建ingress
//	@param client
//	@param namespace
//	@param ingressYAML
//	@return error
func CreateIngress(client *kubernetes.Clientset, namespace, ingressYAML string) error {
	var ingress netv1.Ingress
	if err := yaml.Unmarshal([]byte(ingressYAML), &ingress); err != nil {
		return fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	_, err := client.NetworkingV1().Ingresses(namespace).Create(context.TODO(), &ingress, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("创建ingress资源失败:%s", err.Error())
	}
	return nil
}

func UpdateIngress(client *kubernetes.Clientset, ingressYaml string) error {
	var ingress netv1.Ingress
	if err := yaml.Unmarshal([]byte(ingressYaml), &ingress); err != nil {
		return fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	_, err := client.NetworkingV1().Ingresses(ingress.Namespace).Update(context.TODO(), &ingress, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("更新ingress资源失败:%s", err.Error())
	}
	return nil
}

// DeleteIngressByName
//
//	@Description: 删除ingress通过名称
//	@param client
//	@param namespace
//	@param name
//	@return error
func DeleteIngressByName(client *kubernetes.Clientset, namespace, name string) error {
	err := client.NetworkingV1().Ingresses(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("删除ingress资源失败:%s", err.Error())
	}
	return nil
}

// DeleteIngressByLabel
//
//	@Description: 删除ingress通过标签
//	@param client
//	@param namespace
//	@param labelMap
//	@return bool
//	@return error
func DeleteIngressByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) error {
	labelSelector := labels.SelectorFromSet(labelMap) // 创建标签选择器
	err := client.NetworkingV1().Ingresses(namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return fmt.Errorf("删除ingress资源失败:%s", err.Error())
	}
	return nil
}

// DeleteIngressByField
//
//	@Description: 删除ingress通过字段
//	@param client
//	@param namespace
//	@param fieldMap
//	@return error
func DeleteIngressByField(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) error {
	fieldSelector := fields.SelectorFromSet(fieldMap) // 创建标签选择器
	err := client.NetworkingV1().Ingresses(namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return fmt.Errorf("删除ingress资源失败:%s", err.Error())
	}
	return nil
}
