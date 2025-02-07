package cluster

import (
	"context"
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
)

func GetClusterVersion(client *kubernetes.Clientset) (string, error) {
	version, err := client.ServerVersion()
	if err != nil {
		return "", err
	}
	return version.String(), nil
}

func GetClusterNodes(client *kubernetes.Clientset) ([]corev1.Node, error) {
	nodes, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nodes.Items, nil
}

