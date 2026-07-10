package replicaset

import (
	"gkube/pkg/yamlutil"
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetReplicaSetList(client *kubernetes.Clientset, namespace string) ([]appsv1.ReplicaSet, error) {
	rsList, err := client.AppsV1().ReplicaSets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return rsList.Items, nil
}

func GetReplicaSetYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	rs, err := client.AppsV1().ReplicaSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	rs.TypeMeta = metav1.TypeMeta{APIVersion: "apps/v1", Kind: "ReplicaSet"}
	out, err := yamlutil.MarshalWithoutManagedFields(rs)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ReplicaSet to YAML: %w", err)
	}
	return string(out), nil
}

func DeleteReplicaSet(client *kubernetes.Clientset, namespace, name string) error {
	return client.AppsV1().ReplicaSets(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}
