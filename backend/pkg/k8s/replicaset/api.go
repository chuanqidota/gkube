package replicaset

import (
	"context"
	apperr "gkube/pkg/errors"

	k8sEvent "gkube/pkg/k8s/event"
	"gkube/pkg/yamlutil"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
)

func GetReplicaSetList(ctx context.Context, client *kubernetes.Clientset, namespace string, labelSelector string) ([]appsv1.ReplicaSet, error) {
	listOpts := metav1.ListOptions{ResourceVersion: "0"}
	if labelSelector != "" {
		listOpts.LabelSelector = labelSelector
	}
	rsList, err := client.AppsV1().ReplicaSets(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, err
	}
	return rsList.Items, nil
}

func GetReplicaSetYaml(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (string, error) {
	rs, err := client.AppsV1().ReplicaSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	rs.TypeMeta = metav1.TypeMeta{APIVersion: "apps/v1", Kind: "ReplicaSet"}
	out, err := yamlutil.MarshalWithoutManagedFields(rs)
	if err != nil {
		return "", apperr.K8sAPIFail("序列化失败", err)
	}
	return string(out), nil
}

func DeleteReplicaSet(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error {
	return client.AppsV1().ReplicaSets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// GetReplicaSetPodList returns the full Pod list controlled by the ReplicaSet (matched by its selector).
func GetReplicaSetPodList(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (*corev1.PodList, error) {
	rs, err := client.AppsV1().ReplicaSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, apperr.K8sAPIFail("获取ReplicaSet资源失败", err)
	}
	selector := metav1.FormatLabelSelector(rs.Spec.Selector)
	podList, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector:  selector,
		ResourceVersion: "0",
	})
	if err != nil {
		return nil, apperr.K8sAPIFail("获取Pod列表失败", err)
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

func GetReplicaSetDetail(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (*ReplicaSetDetailDTO, error) {
	rs, err := client.AppsV1().ReplicaSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	selector := metav1.FormatLabelSelector(rs.Spec.Selector)
	podList, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector:  selector,
		ResourceVersion: "0",
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

// GetReplicaSetEvents returns the events associated with a ReplicaSet.
// 用 fields.Selector 防注入。
func GetReplicaSetEvents(ctx context.Context, client *kubernetes.Clientset, namespace, name string) ([]k8sEvent.KubeEvent, error) {
	selector := fields.AndSelectors(
		fields.OneTermEqualSelector("involvedObject.name", name),
		fields.OneTermEqualSelector("involvedObject.kind", "ReplicaSet"),
	).String()
	events, _, _, err := k8sEvent.ListEvents(ctx, client, namespace, selector, 200, "")
	if err != nil {
		return nil, apperr.K8sAPIFail("获取replicaset事件失败", err)
	}
	return events, nil
}
