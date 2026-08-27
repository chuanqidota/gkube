package pod

import (
	"context"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"

	"gkube/pkg/yamlutil"
)

// ListPods returns a paginated pod list with metadata.
// 传 limit<=0 且 continueToken="" 时等价于全量列举(不分页)。
func ListPods(client *kubernetes.Clientset, namespace string, limit int64, continueToken string) (*corev1.PodList, error) {
	listOpts := metav1.ListOptions{ResourceVersion: "0"}
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

// PatchPodMetadata 仅更新 Pod 的 labels 和 annotations（metadata patch）。
// Pod spec 创建后基本不可变,全量 Update 会被 K8s API 拒绝,
// 因此这里用 Strategic Merge Patch 只修补可变的 metadata 字段。
func PatchPodMetadata(client *kubernetes.Clientset, namespace, name, podYaml string) error {
	pod := &corev1.Pod{}
	if err := yaml.Unmarshal([]byte(podYaml), pod); err != nil {
		return fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	if pod.Name != name {
		return fmt.Errorf("资源名称不匹配: 请求指定 %s, YAML 中为 %s", name, pod.Name)
	}

	// 构造只包含 metadata 的 patch 对象
	patchObj := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      pod.Labels,
			Annotations: pod.Annotations,
		},
	}
	patchBytes, err := json.Marshal(patchObj)
	if err != nil {
		return fmt.Errorf("序列化patch失败:%s", err.Error())
	}

	_, err = client.CoreV1().Pods(namespace).Patch(
		context.TODO(), name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{},
	)
	if err != nil {
		return fmt.Errorf("更新pod元数据失败:%s", err.Error())
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
