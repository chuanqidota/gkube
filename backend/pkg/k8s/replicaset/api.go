package replicaset

import (
	"fmt"

	"gkube/pkg/yamlutil"
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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

// GetReplicaSetPodList returns the full Pod list controlled by the ReplicaSet (matched by its selector).
func GetReplicaSetPodList(client *kubernetes.Clientset, namespace, name string) (*corev1.PodList, error) {
	rs, err := client.AppsV1().ReplicaSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get ReplicaSet: %w", err)
	}
	selector := metav1.FormatLabelSelector(rs.Spec.Selector)
	podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list Pods: %w", err)
	}
	return podList, nil
}

// OwnerRef represents the controller owner of a resource
type OwnerRef struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// ReplicaSetDetailDTO contains the ReplicaSet, its controller reference, and related Pods
type ReplicaSetDetailDTO struct {
	RS           appsv1.ReplicaSet `json:"rs"`
	Pods         []corev1.Pod      `json:"pods"`
	ControllerOf *OwnerRef         `json:"controllerOf"`
}

func GetReplicaSetDetail(client *kubernetes.Clientset, namespace, name string) (*ReplicaSetDetailDTO, error) {
	rs, err := client.AppsV1().ReplicaSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	selector := metav1.FormatLabelSelector(rs.Spec.Selector)
	podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, err
	}

	pods := podList.Items

	var controllerOf *OwnerRef
	for _, ref := range rs.OwnerReferences {
		if ref.Controller != nil && *ref.Controller {
			controllerOf = &OwnerRef{
				Kind:      ref.Kind,
				Name:      ref.Name,
				Namespace: rs.Namespace,
			}
			break
		}
	}

	return &ReplicaSetDetailDTO{
		RS:           *rs,
		Pods:         pods,
		ControllerOf: controllerOf,
	}, nil
}
