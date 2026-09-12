package limitrange

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

func GetLimitRangeList(ctx context.Context, client *kubernetes.Clientset, namespace string, labelSelector string) ([]corev1.LimitRange, error) {
	listOpts := metav1.ListOptions{ResourceVersion: "0"}
	if labelSelector != "" {
		listOpts.LabelSelector = labelSelector
	}
	result, err := client.CoreV1().LimitRanges(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

func GetLimitRangeYaml(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (string, error) {
	lr, err := client.CoreV1().LimitRanges(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	lr.TypeMeta = metav1.TypeMeta{APIVersion: "v1", Kind: "LimitRange"}
	out, err := yamlutil.MarshalWithoutManagedFields(lr)
	if err != nil {
		return "", apperr.K8sAPIFail("序列化失败", err)
	}
	return string(out), nil
}

func CreateLimitRange(ctx context.Context, client *kubernetes.Clientset, namespace, yamlContent string) error {
	var lr corev1.LimitRange
	if err := yaml.Unmarshal([]byte(yamlContent), &lr); err != nil {
		return apperr.BadRequest("yaml解析失败", err)
	}
	if lr.Namespace == "" {
		lr.Namespace = namespace
	}
	_, err := client.CoreV1().LimitRanges(namespace).Create(ctx, &lr, metav1.CreateOptions{})
	return err
}

func UpdateLimitRange(ctx context.Context, client *kubernetes.Clientset, namespace, yamlContent string) error {
	var lr corev1.LimitRange
	if err := yaml.Unmarshal([]byte(yamlContent), &lr); err != nil {
		return apperr.BadRequest("yaml解析失败", err)
	}
	if lr.Namespace == "" {
		lr.Namespace = namespace
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := client.CoreV1().LimitRanges(namespace).Get(ctx, lr.Name, metav1.GetOptions{})
		if err != nil {
			return apperr.K8sAPIFail("获取LimitRange资源失败", err)
		}
		latest.Spec.Limits = lr.Spec.Limits
		latest.Labels = lr.Labels
		latest.Annotations = lr.Annotations
		_, err = client.CoreV1().LimitRanges(namespace).Update(ctx, latest, metav1.UpdateOptions{})
		return err
	})
}

func DeleteLimitRange(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error {
	return client.CoreV1().LimitRanges(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func GetLimitRangeDetail(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (*corev1.LimitRange, error) {
	return client.CoreV1().LimitRanges(namespace).Get(ctx, name, metav1.GetOptions{})
}
