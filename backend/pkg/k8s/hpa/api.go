package hpa

import (
	"context"
	"fmt"
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

func GetHPAList(client *kubernetes.Clientset, namespace string) ([]autoscalingv2.HorizontalPodAutoscaler, error) {
	hpaList, err := client.AutoscalingV2().HorizontalPodAutoscalers(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return hpaList.Items, nil
}

func GetHPAYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	hpa, err := client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	hpa.TypeMeta = metav1.TypeMeta{
		APIVersion: "autoscaling/v2",
		Kind:       "HorizontalPodAutoscaler",
	}
	out, err := yamlutil.MarshalWithoutManagedFields(hpa)
	if err != nil {
		return "", fmt.Errorf("failed to marshal HPA to YAML: %w", err)
	}
	return string(out), nil
}

func CreateHPA(client *kubernetes.Clientset, namespace, yamlContent string) error {
	var hpa autoscalingv2.HorizontalPodAutoscaler
	if err := yaml.Unmarshal([]byte(yamlContent), &hpa); err != nil {
		return fmt.Errorf("failed to unmarshal HPA YAML: %w", err)
	}
	if namespace == "" {
		return fmt.Errorf("namespace不能为空")
	}
	if hpa.Namespace != "" && hpa.Namespace != namespace {
		return fmt.Errorf("HPA YAML namespace与请求namespace不一致")
	}
	hpa.Namespace = namespace
	_, err := client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Create(context.TODO(), &hpa, metav1.CreateOptions{})
	return err
}

func UpdateHPA(client *kubernetes.Clientset, namespace, yamlContent string) error {
	var hpa autoscalingv2.HorizontalPodAutoscaler
	if err := yaml.Unmarshal([]byte(yamlContent), &hpa); err != nil {
		return fmt.Errorf("failed to unmarshal HPA YAML: %w", err)
	}
	if namespace == "" {
		return fmt.Errorf("namespace不能为空")
	}
	if hpa.Namespace != "" && hpa.Namespace != namespace {
		return fmt.Errorf("不允许修改HPA namespace")
	}
	hpa.Namespace = namespace
	if hpa.Name == "" {
		return fmt.Errorf("metadata.name不能为空")
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		existing, err := client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Get(context.TODO(), hpa.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		existing.Spec = hpa.Spec
		existing.Labels = hpa.Labels
		existing.Annotations = hpa.Annotations
		_, err = client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Update(context.TODO(), existing, metav1.UpdateOptions{})
		return err
	})
}

func DeleteHPA(client *kubernetes.Clientset, namespace, name string) error {
	return client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func GetHPADetail(client *kubernetes.Clientset, namespace, name string) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	return client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func GetHPAEvents(client *kubernetes.Clientset, namespace, name string) ([]k8sEvent.KubeEvent, error) {
	selector := fields.AndSelectors(
		fields.OneTermEqualSelector("involvedObject.name", name),
		fields.OneTermEqualSelector("involvedObject.kind", "HorizontalPodAutoscaler"),
	).String()
	events, _, _, err := k8sEvent.ListEvents(client, namespace, selector, 0, "")
	if err != nil {
		return nil, fmt.Errorf("获取HPA事件失败:%s", err.Error())
	}
	return events, nil
}

// PauseHPA freezes the HPA by setting minReplicas = maxReplicas = currentReplicas
// and saving the original values in annotations.
func PauseHPA(client *kubernetes.Clientset, namespace, name string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		hpa, err := client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Get(context.TODO(), name, metav1.GetOptions{})
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
		_, err = client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Update(context.TODO(), hpa, metav1.UpdateOptions{})
		return err
	})
}

// ResumeHPA restores the original min/max from annotations and removes pause markers.
func ResumeHPA(client *kubernetes.Clientset, namespace, name string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		hpa, err := client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Get(context.TODO(), name, metav1.GetOptions{})
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
		_, err = client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Update(context.TODO(), hpa, metav1.UpdateOptions{})
		return err
	})
}
