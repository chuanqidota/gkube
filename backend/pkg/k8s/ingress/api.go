package ingress

import (
	"context"
	apperr "gkube/pkg/errors"

	k8sEvent "gkube/pkg/k8s/event"
	"gkube/pkg/yamlutil"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"

	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/yaml"
)

// GetIngressList
//
//	@Description: 获取ingress列表
//	@param client
//	@param namespace
//	@return []netv1.Ingress
//	@return error
func GetIngressList(ctx context.Context, client *kubernetes.Clientset, namespace string, labelSelector string) ([]netv1.Ingress, error) {
	listOpts := metav1.ListOptions{ResourceVersion: "0"}
	if labelSelector != "" {
		listOpts.LabelSelector = labelSelector
	}
	result, err := client.NetworkingV1().Ingresses(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

// GetIngressByName
//
//	@Description:
//	@param client
//	@param namespace
//	@param name
//	@return netv1.Ingress
//	@return error
func GetIngressByName(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (*netv1.Ingress, error) {
	ingress, err := client.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return &netv1.Ingress{}, err
	}
	return ingress, nil
}

// GetIngressByLabel
//
//	@Description: 根据label获取ingress
//	@param client
//	@param namespace
//	@param labelMap
//	@return []netv1.Ingress
//	@return error
func GetIngressByLabel(ctx context.Context, client *kubernetes.Clientset, namespace string, labelMap map[string]string) ([]netv1.Ingress, error) {
	labelSelector := labels.SelectorFromSet(labelMap) // 创建标签选择器
	ingress, err := client.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{
		LabelSelector:  labelSelector.String(),
		ResourceVersion: "0",
	})
	if err != nil {
		return nil, err
	}
	return ingress.Items, nil
}

// GetIngressByFiled
//
//	@Description: 根据字段获取ingress
//	@param client
//	@param namespace
//	@param fieldMap
//	@return []netv1.Ingress
//	@return error
func GetIngressByFiled(ctx context.Context, client *kubernetes.Clientset, namespace string, fieldMap map[string]string) ([]netv1.Ingress, error) {
	fieldSelector := fields.SelectorFromSet(fieldMap) // 创建标签选择器
	ingress, err := client.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{
		FieldSelector:  fieldSelector.String(),
		ResourceVersion: "0",
	})
	if err != nil {
		return nil, err
	}
	return ingress.Items, nil
}

// GetIngressYaml
//
//	@Description: 获取ingress yaml
//	@param client
//	@param namespace
//	@param name
//	@return string
//	@return error
func GetIngressYaml(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (string, error) {
	ingress, err := client.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	ingress.TypeMeta = metav1.TypeMeta{APIVersion: "networking.k8s.io/v1", Kind: "Ingress"}
	ingressYAML, err := yamlutil.MarshalWithoutManagedFields(ingress)
	if err != nil {
		return "", err
	}
	return ingressYAML, nil
}

// CreateIngress
//
//	@Description: 创建ingress
//	@param client
//	@param namespace
//	@param ingressYAML
//	@return error
func CreateIngress(ctx context.Context, client *kubernetes.Clientset, namespace, ingressYAML string) error {
	var ingress netv1.Ingress
	if err := yaml.Unmarshal([]byte(ingressYAML), &ingress); err != nil {
		return apperr.K8sAPIFail("yaml文件错误", err)
	}
	_, err := client.NetworkingV1().Ingresses(namespace).Create(ctx, &ingress, metav1.CreateOptions{})
	if err != nil {
		return apperr.K8sAPIFail("创建ingress资源失败", err)
	}
	return nil
}

func UpdateIngress(ctx context.Context, client *kubernetes.Clientset, namespace, ingressYaml string) error {
	var ingress netv1.Ingress
	if err := yaml.Unmarshal([]byte(ingressYaml), &ingress); err != nil {
		return apperr.K8sAPIFail("yaml文件错误", err)
	}
	// 以请求参数的 namespace 为准,防止 YAML 内嵌 namespace 与请求不一致
	if namespace != "" {
		ingress.Namespace = namespace
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := client.NetworkingV1().Ingresses(ingress.Namespace).Get(ctx, ingress.Name, metav1.GetOptions{})
		if err != nil {
			return apperr.K8sAPIFail("获取ingress资源失败", err)
		}
		latest.Spec = ingress.Spec
		latest.Labels = ingress.Labels
		latest.Annotations = ingress.Annotations
		_, err = client.NetworkingV1().Ingresses(ingress.Namespace).Update(ctx, latest, metav1.UpdateOptions{})
		if err != nil {
			return apperr.K8sAPIFail("更新ingress资源失败", err)
		}
		return nil
	})
}

// DeleteIngressByName
//
//	@Description: 删除ingress通过名称
//	@param client
//	@param namespace
//	@param name
//	@return error
func DeleteIngressByName(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error {
	err := client.NetworkingV1().Ingresses(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return apperr.K8sAPIFail("删除ingress资源失败", err)
	}
	return nil
}

// DeleteIngressByLabel
//
//	@Description: 删除ingress通过标签
//	@param client
//	@param namespace
//	@param labelMap
//	@return bool
//	@return error
func DeleteIngressByLabel(ctx context.Context, client *kubernetes.Clientset, namespace string, labelMap map[string]string) error {
	labelSelector := labels.SelectorFromSet(labelMap) // 创建标签选择器
	err := client.NetworkingV1().Ingresses(namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return apperr.K8sAPIFail("删除ingress资源失败", err)
	}
	return nil
}

// DeleteIngressByField
//
//	@Description: 删除ingress通过字段
//	@param client
//	@param namespace
//	@param fieldMap
//	@return error
func DeleteIngressByField(ctx context.Context, client *kubernetes.Clientset, namespace string, fieldMap map[string]string) error {
	fieldSelector := fields.SelectorFromSet(fieldMap) // 创建标签选择器
	err := client.NetworkingV1().Ingresses(namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return apperr.K8sAPIFail("删除ingress资源失败", err)
	}
	return nil
}

// ListIngressClasses 返回集群中所有 IngressClass 的名称列表。
// 集群级资源，无需 namespace。
func ListIngressClasses(ctx context.Context, client *kubernetes.Clientset) ([]string, error) {
	list, err := client.NetworkingV1().IngressClasses().List(ctx, metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		return nil, apperr.K8sAPIFail("获取IngressClass列表失败", err)
	}
	names := make([]string, 0, len(list.Items))
	for _, ic := range list.Items {
		names = append(names, ic.Name)
	}
	return names, nil
}

// GetIngressEvents returns the events associated with an Ingress.
// 用 fields.Selector 防注入。
func GetIngressEvents(ctx context.Context, client *kubernetes.Clientset, namespace, name string) ([]k8sEvent.KubeEvent, error) {
	selector := fields.AndSelectors(
		fields.OneTermEqualSelector("involvedObject.name", name),
		fields.OneTermEqualSelector("involvedObject.kind", "Ingress"),
	).String()
	events, _, _, err := k8sEvent.ListEvents(ctx, client, namespace, selector, 0, "")
	if err != nil {
		return nil, apperr.K8sAPIFail("获取ingress事件失败", err)
	}
	return events, nil
}
