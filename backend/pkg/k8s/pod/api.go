package pod

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/yaml"

	"gkube/pkg/yamlutil"
)

// ListPods returns a paginated pod list with metadata.
// 传 limit<=0 且 continueToken="" 时等价于全量列举(不分页)。
func ListPods(client *kubernetes.Clientset, namespace string, limit int64, continueToken string) (*corev1.PodList, error) {
	listOpts := metav1.ListOptions{}
	if limit > 0 {
		listOpts.Limit = limit
	}
	if continueToken != "" {
		listOpts.Continue = continueToken
	}
	return client.CoreV1().Pods(namespace).List(context.TODO(), listOpts)
}

// GetPodByName
//
//	@Description: 获取pod
//	@param client
//	@param namespace
//	@param name
//	@return *corev1.Pod
//	@return error
func GetPodByName(client *kubernetes.Clientset, namespace, name string) (*corev1.Pod, error) {
	pod, err := client.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return pod, nil
}

// GetPodYaml
//
//	@Description: 获取pod yaml
//	@param client
//	@param namespace
//	@param name
//	@return string
//	@return error
func GetPodYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	pod, err := client.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	yamlStr, err := yamlutil.MarshalWithoutManagedFields(pod)
	if err != nil {
		return "", err
	}
	return yamlStr, nil
}

// CreatePod
//
//	@Description: 创建pod
//	@param client
//	@param namespace 以请求参数为准,避免 YAML 内 metadata.namespace 与之不符时静默落到别处
//	@param podYaml
//	@return error
func CreatePod(client *kubernetes.Clientset, namespace, podYaml string) error {
	pod := &corev1.Pod{}
	if err := yaml.Unmarshal([]byte(podYaml), pod); err != nil {
		return fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	pod.Namespace = namespace
	_, err := client.CoreV1().Pods(namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("创建pod资源失败:%s", err.Error())
	}
	return nil
}

// UpdatePod
//
//	@Description: 更新pod
//	@param client
//	@param namespace 以请求参数为准
//	@param name 校验与 YAML 中名称一致,避免误更新同名空间下的其他资源
//	@param podYaml
//	@return error
func UpdatePod(client *kubernetes.Clientset, namespace, name, podYaml string) error {
	pod := &corev1.Pod{}
	if err := yaml.Unmarshal([]byte(podYaml), pod); err != nil {
		return fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	if pod.Name != name {
		return fmt.Errorf("资源名称不匹配: 请求指定 %s, YAML 中为 %s", name, pod.Name)
	}
	pod.Namespace = namespace
	// 冲突时自动重试(参照 deployment restart/scale 的 RetryOnConflict 模式)。
	// 闭包内 re-Get 最新对象(带新 resourceVersion),再用用户 YAML 的 spec 覆盖后 Update,
	// 否则重试会发同一个过期 resourceVersion 持续 409 直到 backoff 耗尽(死重试)。
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := client.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("获取pod资源失败:%s", err.Error())
		}
		// 用用户 YAML 的 spec 与可变 metadata 覆盖最新对象,保留最新 resourceVersion
		latest.Spec = pod.Spec
		latest.Labels = pod.Labels
		latest.Annotations = pod.Annotations
		_, err = client.CoreV1().Pods(namespace).Update(context.TODO(), latest, metav1.UpdateOptions{})
		return err
	}); err != nil {
		return fmt.Errorf("更新pod资源失败:%s", err.Error())
	}
	return nil
}

// DeletePodByName
//
//	@Description: 删除pod根据名称
//	@param client
//	@param namespace
//	@param name
//	@param force 强制删除：GracePeriodSeconds=0，PropagationPolicy=Background
//	@return error
func DeletePodByName(client *kubernetes.Clientset, namespace, name string, force bool) error {
	deleteOpts := metav1.DeleteOptions{}
	if force {
		gracePeriod := int64(0)
		deleteOpts.GracePeriodSeconds = &gracePeriod
		prop := metav1.DeletePropagationBackground
		deleteOpts.PropagationPolicy = &prop
	}
	err := client.CoreV1().Pods(namespace).Delete(context.TODO(), name, deleteOpts)
	if err != nil {
		return fmt.Errorf("删除pod资源失败:%s", err.Error())
	}
	return nil
}
