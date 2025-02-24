package deployment

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// GetDeploymentList
//
//	@Description: 获取deployment列表
//	@param client
//	@param namespace
//	@return []appsv1.Deployment
//	@return error
func GetDeploymentList(client *kubernetes.Clientset, namespace string) ([]appsv1.Deployment, error) {
	deploymentList, err := client.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return deploymentList.Items, nil
}

// GetDeploymentByFiled
//
//	@Description: 根据字段获取deployment列表
//	@param client
//	@param namespace
//	@param fieldMap
//	@return []appsv1.Deployment
//	@return error
func GetDeploymentByFiled(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) ([]appsv1.Deployment, error) {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	deploymentList, err := client.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return nil, err
	}
	return deploymentList.Items, nil
}

// GetDeploymentByLabel
//
//	@Description: 根据标签获取deployment列表
//	@param client
//	@param namespace
//	@param labelMap
//	@return []appsv1.Deployment
//	@return error
func GetDeploymentByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) ([]appsv1.Deployment, error) {
	labelSelector := fields.SelectorFromSet(labelMap)
	deploymentList, err := client.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return nil, err
	}
	return deploymentList.Items, nil
}

// GetDeploymentYaml
//
//	@Description: 获取deployment yaml
//	@param client
//	@param namespace
//	@param name
//	@return string
//	@return error
func GetDeploymentYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	deployment, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return deployment.String(), nil
}

// CreateDeployment
//
//	@Description: 创建deployment
//	@param client
//	@param namespace
//	@param cronJobYaml
//	@return bool
//	@return error
func CreateDeployment(client *kubernetes.Clientset, namespace string, cronJobYaml string) (bool, error) {
	var deployment *appsv1.Deployment
	if err := yaml.Unmarshal([]byte(cronJobYaml), &deployment); err != nil {
		return false, fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	_, err := client.AppsV1().Deployments(namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		return false, fmt.Errorf("创建deployment资源失败:%s", err.Error())
	}
	return true, nil
}

// UpdateDeployment
//
//	@Description: 更新deployment
//	@param client
//	@param namespace
//	@param cronJobYaml
//	@return bool
//	@return error
func UpdateDeployment(client *kubernetes.Clientset, namespace, cronJobYaml string) (bool, error) {
	var deployment *appsv1.Deployment
	if err := yaml.Unmarshal([]byte(cronJobYaml), &deployment); err != nil {
		return false, fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	_, err := client.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return false, fmt.Errorf("更新deployment资源失败:%s", err.Error())
	}
	return true, nil
}

// DeleteDeployment
//
//	@Description: 删除deployment
//	@param client
//	@param namespace
//	@param name
//	@return bool
//	@return error
func DeleteDeployment(client *kubernetes.Clientset, namespace, name string) (bool, error) {
	err := client.AppsV1().Deployments(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return false, fmt.Errorf("删除deployment资源失败:%s", err.Error())
	}
	return true, nil
}

// DeleteDeploymentByField
//
//	@Description: 根据字段删除deployment
//	@param client
//	@param namespace
//	@param fieldMap
//	@return bool
//	@return error
func DeleteDeploymentByField(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) (bool, error) {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	err := client.AppsV1().Deployments(namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return false, fmt.Errorf("删除deployment资源失败:%s", err.Error())
	}
	return true, nil
}

// DeleteDeploymentByLabel
//
//	@Description: 根据标签删除deployment
//	@param client
//	@param namespace
//	@param labelMap
//	@return bool
//	@return error
func DeleteDeploymentByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) (bool, error) {
	labelSelector := fields.SelectorFromSet(labelMap)
	err := client.AppsV1().Deployments(namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return false, fmt.Errorf("删除deployment资源失败:%s", err.Error())
	}
	return true, nil
}
