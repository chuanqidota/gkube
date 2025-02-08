package node

import (
	"context"
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

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
