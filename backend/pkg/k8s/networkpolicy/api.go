package networkpolicy

import (
	"context"
	apperr "gkube/pkg/errors"
	"gkube/pkg/yamlutil"

	k8sEvent "gkube/pkg/k8s/event"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/yaml"
)

func GetNetworkPolicyList(ctx context.Context, client *kubernetes.Clientset, namespace string, labelSelector string) ([]networkingv1.NetworkPolicy, error) {
	listOpts := metav1.ListOptions{ResourceVersion: "0"}
	if labelSelector != "" {
		listOpts.LabelSelector = labelSelector
	}
	result, err := client.NetworkingV1().NetworkPolicies(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

func GetNetworkPolicyYaml(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (string, error) {
	np, err := client.NetworkingV1().NetworkPolicies(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	np.TypeMeta = metav1.TypeMeta{
		APIVersion: "networking.k8s.io/v1",
		Kind:       "NetworkPolicy",
	}
	out, err := yamlutil.MarshalWithoutManagedFields(np)
	if err != nil {
		return "", apperr.K8sAPIFail("序列化失败", err)
	}
	return string(out), nil
}

func CreateNetworkPolicy(ctx context.Context, client *kubernetes.Clientset, namespace, yamlContent string) error {
	var np networkingv1.NetworkPolicy
	if err := yaml.Unmarshal([]byte(yamlContent), &np); err != nil {
		return apperr.BadRequest("yaml解析失败", err)
	}
	if np.Namespace == "" {
		np.Namespace = namespace
	}
	_, err := client.NetworkingV1().NetworkPolicies(namespace).Create(ctx, &np, metav1.CreateOptions{})
	return err
}

func UpdateNetworkPolicy(ctx context.Context, client *kubernetes.Clientset, namespace, yamlContent string) error {
	var np networkingv1.NetworkPolicy
	if err := yaml.Unmarshal([]byte(yamlContent), &np); err != nil {
		return apperr.BadRequest("yaml解析失败", err)
	}
	if np.Namespace == "" {
		np.Namespace = namespace
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := client.NetworkingV1().NetworkPolicies(namespace).Get(ctx, np.Name, metav1.GetOptions{})
		if err != nil {
			return apperr.K8sAPIFail("获取NetworkPolicy资源失败", err)
		}
		latest.Spec = np.Spec
		latest.Labels = np.Labels
		latest.Annotations = np.Annotations
		_, err = client.NetworkingV1().NetworkPolicies(namespace).Update(ctx, latest, metav1.UpdateOptions{})
		return err
	})
}

func DeleteNetworkPolicy(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error {
	return client.NetworkingV1().NetworkPolicies(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func GetNetworkPolicyDetail(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (*networkingv1.NetworkPolicy, error) {
	return client.NetworkingV1().NetworkPolicies(namespace).Get(ctx, name, metav1.GetOptions{})
}

// GetNetworkPolicyPods returns pods matched by the NetworkPolicy's podSelector.
// Properly handles both MatchLabels and MatchExpressions.
func GetNetworkPolicyPods(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (*corev1.PodList, error) {
	np, err := client.NetworkingV1().NetworkPolicies(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, apperr.K8sAPIFail("获取NetworkPolicy资源失败", err)
	}
	// Use the full label selector (MatchLabels + MatchExpressions)
	selector, err := metav1.LabelSelectorAsSelector(&np.Spec.PodSelector)
	if err != nil {
		return nil, apperr.K8sAPIFail("解析PodSelector失败", err)
	}
	podList, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector:  selector.String(),
		ResourceVersion: "0",
	})
	if err != nil {
		return nil, apperr.K8sAPIFail("获取NetworkPolicy关联pod列表失败", err)
	}
	return podList, nil
}

// GetNetworkPolicyEvents returns the events associated with a NetworkPolicy.
// 用 fields.Selector 防注入。
func GetNetworkPolicyEvents(ctx context.Context, client *kubernetes.Clientset, namespace, name string) ([]k8sEvent.KubeEvent, error) {
	selector := fields.AndSelectors(
		fields.OneTermEqualSelector("involvedObject.name", name),
		fields.OneTermEqualSelector("involvedObject.kind", "NetworkPolicy"),
	).String()
	events, _, _, err := k8sEvent.ListEvents(ctx, client, namespace, selector, 0, "")
	if err != nil {
		return nil, apperr.K8sAPIFail("获取networkpolicy事件失败", err)
	}
	return events, nil
}
