package hpa

import (
	"context"
	apperr "gkube/pkg/errors"
	"strconv"

	k8sEvent "gkube/pkg/k8s/event"
	"gkube/pkg/yamlutil"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/yaml"
)

const (
	annotationPaused        = "gkube.io/paused"
	annotationPausedMin     = "gkube.io/paused-min-replicas"
	annotationPausedMax     = "gkube.io/paused-max-replicas"
)

func GetHPAList(ctx context.Context, client *kubernetes.Clientset, namespace string, labelSelector string) ([]autoscalingv2.HorizontalPodAutoscaler, error) {
	listOpts := metav1.ListOptions{ResourceVersion: "0"}
	if labelSelector != "" {
		listOpts.LabelSelector = labelSelector
	}
	result, err := client.AutoscalingV2().HorizontalPodAutoscalers(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

func GetHPAYaml(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (string, error) {
	hpa, err := client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	hpa.TypeMeta = metav1.TypeMeta{
		APIVersion: "autoscaling/v2",
		Kind:       "HorizontalPodAutoscaler",
	}
	out, err := yamlutil.MarshalWithoutManagedFields(hpa)
	if err != nil {
		return "", apperr.K8sAPIFail("序列化失败", err)
	}
	return string(out), nil
}

func CreateHPA(ctx context.Context, client *kubernetes.Clientset, namespace, yamlContent string) error {
	var hpa autoscalingv2.HorizontalPodAutoscaler
	if err := yaml.Unmarshal([]byte(yamlContent), &hpa); err != nil {
		return apperr.BadRequest("yaml解析失败", err)
	}
	if namespace == "" {
		return apperr.BadRequest("namespace不能为空", nil)
	}
	if hpa.Namespace != "" && hpa.Namespace != namespace {
		return apperr.BadRequest("HPA YAML namespace与请求namespace不一致", nil)
	}
	hpa.Namespace = namespace
	_, err := client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Create(ctx, &hpa, metav1.CreateOptions{})
	return err
}

func UpdateHPA(ctx context.Context, client *kubernetes.Clientset, namespace, yamlContent string) error {
	var hpa autoscalingv2.HorizontalPodAutoscaler
	if err := yaml.Unmarshal([]byte(yamlContent), &hpa); err != nil {
		return apperr.BadRequest("yaml解析失败", err)
	}
	if namespace == "" {
		return apperr.BadRequest("namespace不能为空", nil)
	}
	if hpa.Namespace != "" && hpa.Namespace != namespace {
		return apperr.BadRequest("不允许修改HPA namespace", nil)
	}
	hpa.Namespace = namespace
	if hpa.Name == "" {
		return apperr.BadRequest("metadata.name不能为空", nil)
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		existing, err := client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Get(ctx, hpa.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		existing.Spec = hpa.Spec
		existing.Labels = hpa.Labels
		existing.Annotations = hpa.Annotations
		_, err = client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Update(ctx, existing, metav1.UpdateOptions{})
		return err
	})
}

func DeleteHPA(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error {
	return client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func GetHPADetail(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	return client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Get(ctx, name, metav1.GetOptions{})
}

func GetHPAEvents(ctx context.Context, client *kubernetes.Clientset, namespace, name string) ([]k8sEvent.KubeEvent, error) {
	selector := fields.AndSelectors(
		fields.OneTermEqualSelector("involvedObject.name", name),
		fields.OneTermEqualSelector("involvedObject.kind", "HorizontalPodAutoscaler"),
	).String()
	events, _, _, err := k8sEvent.ListEvents(ctx, client, namespace, selector, 0, "")
	if err != nil {
		return nil, apperr.K8sAPIFail("获取HPA事件失败", err)
	}
	return events, nil
}

// PauseHPA freezes the HPA by setting minReplicas = maxReplicas = currentReplicas
// and saving the original values in annotations.
func PauseHPA(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		hpa, err := client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if hpa.Annotations == nil {
			hpa.Annotations = make(map[string]string)
		}
		// Already paused
		if hpa.Annotations[annotationPaused] == "true" {
			return nil
		}
		// Save original min/max
		var origMin int32 = 1
		if hpa.Spec.MinReplicas != nil {
			origMin = *hpa.Spec.MinReplicas
		}
		hpa.Annotations[annotationPausedMin] = strconv.Itoa(int(origMin))
		hpa.Annotations[annotationPausedMax] = strconv.Itoa(int(hpa.Spec.MaxReplicas))
		hpa.Annotations[annotationPaused] = "true"
		// Freeze to current replicas
		current := hpa.Status.CurrentReplicas
		if current == 0 {
			// K8s requires maxReplicas >= 1; use 1 as minimum valid freeze
			one := int32(1)
			hpa.Spec.MinReplicas = &one
			hpa.Spec.MaxReplicas = 1
		} else {
			hpa.Spec.MinReplicas = &current
			hpa.Spec.MaxReplicas = current
		}
		_, err = client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Update(ctx, hpa, metav1.UpdateOptions{})
		return err
	})
}

// ResumeHPA restores the original min/max from annotations and removes pause markers.
func ResumeHPA(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		hpa, err := client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if hpa.Annotations == nil || hpa.Annotations[annotationPaused] != "true" {
			return nil
		}
		// Restore original min/max
		if minStr, ok := hpa.Annotations[annotationPausedMin]; ok {
			if v, err := strconv.Atoi(minStr); err == nil {
				minVal := int32(v)
				hpa.Spec.MinReplicas = &minVal
			}
		}
		if maxStr, ok := hpa.Annotations[annotationPausedMax]; ok {
			if v, err := strconv.Atoi(maxStr); err == nil {
				hpa.Spec.MaxReplicas = int32(v)
			}
		}
		// Remove pause annotations
		delete(hpa.Annotations, annotationPaused)
		delete(hpa.Annotations, annotationPausedMin)
		delete(hpa.Annotations, annotationPausedMax)
		_, err = client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Update(ctx, hpa, metav1.UpdateOptions{})
		return err
	})
}
