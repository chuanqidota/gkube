package limitrange

import (
	"gkube/pkg/yamlutil"
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/yaml"
)

func GetLimitRangeList(client *kubernetes.Clientset, namespace string, labelSelector string) ([]corev1.LimitRange, error) {
	listOpts := metav1.ListOptions{ResourceVersion: "0"}
	if labelSelector != "" {
		listOpts.LabelSelector = labelSelector
	}
	result, err := client.CoreV1().LimitRanges(namespace).List(context.TODO(), listOpts)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

func GetLimitRangeYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	lr, err := client.CoreV1().LimitRanges(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	lr.TypeMeta = metav1.TypeMeta{APIVersion: "v1", Kind: "LimitRange"}
	out, err := yamlutil.MarshalWithoutManagedFields(lr)
	if err != nil {
		return "", fmt.Errorf("failed to marshal LimitRange to YAML: %w", err)
	}
	return string(out), nil
}

func CreateLimitRange(client *kubernetes.Clientset, namespace, yamlContent string) error {
	var lr corev1.LimitRange
	if err := yaml.Unmarshal([]byte(yamlContent), &lr); err != nil {
		return fmt.Errorf("failed to unmarshal LimitRange YAML: %w", err)
	}
	if lr.Namespace == "" {
		lr.Namespace = namespace
	}
	_, err := client.CoreV1().LimitRanges(namespace).Create(context.TODO(), &lr, metav1.CreateOptions{})
	return err
}

func UpdateLimitRange(client *kubernetes.Clientset, namespace, yamlContent string) error {
	var lr corev1.LimitRange
	if err := yaml.Unmarshal([]byte(yamlContent), &lr); err != nil {
		return fmt.Errorf("failed to unmarshal LimitRange YAML: %w", err)
	}
	if lr.Namespace == "" {
		lr.Namespace = namespace
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := client.CoreV1().LimitRanges(namespace).Get(context.TODO(), lr.Name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get LimitRange: %w", err)
		}
		latest.Spec.Limits = lr.Spec.Limits
		latest.Labels = lr.Labels
		latest.Annotations = lr.Annotations
		_, err = client.CoreV1().LimitRanges(namespace).Update(context.TODO(), latest, metav1.UpdateOptions{})
		return err
	})
}

func DeleteLimitRange(client *kubernetes.Clientset, namespace, name string) error {
	return client.CoreV1().LimitRanges(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func GetLimitRangeDetail(client *kubernetes.Clientset, namespace, name string) (*corev1.LimitRange, error) {
	return client.CoreV1().LimitRanges(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}
