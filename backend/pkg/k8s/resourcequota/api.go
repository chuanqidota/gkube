package resourcequota

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

func GetResourceQuotaList(ctx context.Context, client *kubernetes.Clientset, namespace string, labelSelector string) ([]corev1.ResourceQuota, error) {
	listOpts := metav1.ListOptions{ResourceVersion: "0"}
	if labelSelector != "" {
		listOpts.LabelSelector = labelSelector
	}
	result, err := client.CoreV1().ResourceQuotas(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

func GetResourceQuotaYaml(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (string, error) {
	rq, err := client.CoreV1().ResourceQuotas(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	rq.TypeMeta = metav1.TypeMeta{APIVersion: "v1", Kind: "ResourceQuota"}
	out, err := yamlutil.MarshalWithoutManagedFields(rq)
	if err != nil {
		return "", apperr.K8sAPIFail("序列化失败", err)
	}
	return string(out), nil
}

func CreateResourceQuota(ctx context.Context, client *kubernetes.Clientset, namespace, yamlContent string) error {
	var rq corev1.ResourceQuota
	if err := yaml.Unmarshal([]byte(yamlContent), &rq); err != nil {
		return apperr.BadRequest("yaml解析失败", err)
	}
	if rq.Namespace == "" {
		rq.Namespace = namespace
	}
	_, err := client.CoreV1().ResourceQuotas(namespace).Create(ctx, &rq, metav1.CreateOptions{})
	return err
}

func UpdateResourceQuota(ctx context.Context, client *kubernetes.Clientset, namespace, yamlContent string) error {
	var rq corev1.ResourceQuota
	if err := yaml.Unmarshal([]byte(yamlContent), &rq); err != nil {
		return apperr.BadRequest("yaml解析失败", err)
	}
	if rq.Namespace == "" {
		rq.Namespace = namespace
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := client.CoreV1().ResourceQuotas(namespace).Get(ctx, rq.Name, metav1.GetOptions{})
		if err != nil {
			return apperr.K8sAPIFail("获取ResourceQuota资源失败", err)
		}
		latest.Spec.Hard = rq.Spec.Hard
		latest.Labels = rq.Labels
		latest.Annotations = rq.Annotations
		_, err = client.CoreV1().ResourceQuotas(namespace).Update(ctx, latest, metav1.UpdateOptions{})
		return err
	})
}

func DeleteResourceQuota(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error {
	return client.CoreV1().ResourceQuotas(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func GetResourceQuotaDetail(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (*corev1.ResourceQuota, error) {
	return client.CoreV1().ResourceQuotas(namespace).Get(ctx, name, metav1.GetOptions{})
}
