package node

import (
	"context"
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// GetNodeYaml
//
//	@Description: 获取node的yaml
//	@param client
//	@param nodeName
//	@return string
//	@return error
func GetNodeYaml(client *kubernetes.Clientset, nodeName string) (string, error) {
	node, err := client.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	nodeJSON, err := json.Marshal(node)
	if err != nil {
		return "", err
	}
	nodeYAML, err := yaml.JSONToYAML(nodeJSON)
	if err != nil {
		return "", err
	}
	return string(nodeYAML), nil
}

// GetNodePods
//
//	@Description: 获取node的pods
//	@param client
//	@param nodeName
//	@return *corev1.PodList
//	@return error
func GetNodePods(client *kubernetes.Clientset, nodeName string) (*corev1.PodList, error) {
	fieldSelector, err := fields.ParseSelector("spec.nodeName=" + nodeName +
		",status.phase!=" + string(corev1.PodSucceeded) +
		",status.phase!=" + string(corev1.PodFailed))
	if err != nil {
		return nil, err
	}
	return client.CoreV1().Pods(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
}

// UnschedulableNode
//
//	@Description: 禁止调度
//	@param client
//	@param nodeName
//	@return bool
//	@return error
func UnschedulableNode(client *kubernetes.Clientset, nodeName string) (bool, error) {
	node, err := client.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	node.Spec.Unschedulable = true
	_, err = client.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})
	if err != nil {
		return false, err
	}
	return true, nil

}
