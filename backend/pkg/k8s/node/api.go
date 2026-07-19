package node

import (
	"context"
	"fmt"

	"gkube/pkg/yamlutil"

	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
)

// GetNodeYaml
//
//	@Description: 获取node的yaml
//	@param client
//	@param nodeName
//	@return string
//	@return error
func GetNodeYaml(client *kubernetes.Clientset, nodeName string) (string, error) {
	node, err := client.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	node.TypeMeta = metav1.TypeMeta{APIVersion: "v1", Kind: "Node"}
	nodeYAML, err := yamlutil.MarshalWithoutManagedFields(node)
	if err != nil {
		return "", err
	}
	return nodeYAML, nil
}

// GetNodePods
//
//	@Description: 获取node的pods
//	@param client
//	@param nodeName
//	@return []corev1.Pod
//	@return error
func GetNodePods(client *kubernetes.Clientset, nodeName string) ([]corev1.Pod, error) {
	fieldSelector := fields.SelectorFromSet(fields.Set{"spec.nodeName": nodeName})
	podList, err := client.CoreV1().Pods(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return nil, err
	}
	// Filter out Succeeded and Failed pods
	var pods []corev1.Pod
	for _, pod := range podList.Items {
		if pod.Status.Phase != corev1.PodSucceeded && pod.Status.Phase != corev1.PodFailed {
			pods = append(pods, pod)
		}
	}
	return pods, nil
}

// CordonNode
//
//	@Description: 封锁或解除封锁节点
//	@param client
//	@param nodeName
//	@param cordon true=封锁, false=解除封锁
//	@return bool 返回当前封锁状态
//	@return error
func CordonNode(client *kubernetes.Clientset, nodeName string, cordon bool) (bool, error) {
	node, err := client.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	node.Spec.Unschedulable = cordon
	_, err = client.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})
	if err != nil {
		return false, err
	}
	return cordon, nil
}

// DrainOptions 驱逐选项
type DrainOptions struct {
	IgnoreDaemonSets bool `json:"ignoreDaemonSets"` // 是否忽略 DaemonSet 管理的 Pod
	DeleteLocalData  bool `json:"deleteLocalData"`  // 是否删除使用 emptyDir 等本地存储的 Pod
	GracePeriod      int  `json:"gracePeriod"`       // 优雅终止超时秒数，-1 使用 Pod 默认值
	Force            bool `json:"force"`             // 是否强制驱逐（即使 Pod 未就绪）
}

// DrainNode
//
//	@Description: 驱逐节点上的所有 pod（先封锁再驱逐）
//	@param client
//	@param nodeName
//	@param opts 驱逐选项
//	@return []string 被驱逐的 pod 列表
//	@return []string 被跳过的 pod 列表
//	@return error
func DrainNode(client *kubernetes.Clientset, nodeName string, opts DrainOptions) (evicted []string, skipped []string, err error) {
	// Step 1: Cordon the node first
	node, err := client.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("获取节点失败:%s", err.Error())
	}
	if !node.Spec.Unschedulable {
		node.Spec.Unschedulable = true
		if _, err := client.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{}); err != nil {
			return nil, nil, fmt.Errorf("封锁节点失败:%s", err.Error())
		}
	}

	// Step 2: List all pods on the node
	pods, err := client.CoreV1().Pods(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("获取节点pod列表失败:%s", err.Error())
	}

	// Step 3: Filter and evict
	const systemNamespace = "kube-system"
	for _, pod := range pods.Items {
		// Skip mirror pods (static pods)
		if _, isMirror := pod.Annotations["kubernetes.io/config.mirror"]; isMirror {
			skipped = append(skipped, fmt.Sprintf("%s/%s (static pod)", pod.Namespace, pod.Name))
			continue
		}

		// Skip completed pods
		if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
			skipped = append(skipped, fmt.Sprintf("%s/%s (completed)", pod.Namespace, pod.Name))
			continue
		}

		// Skip kube-system pods (unless forced)
		if pod.Namespace == systemNamespace && !opts.Force {
			skipped = append(skipped, fmt.Sprintf("%s/%s (kube-system)", pod.Namespace, pod.Name))
			continue
		}

		// Skip DaemonSet-managed pods if option is set
		if opts.IgnoreDaemonSets {
			isDaemonSet := false
			for _, ownerRef := range pod.OwnerReferences {
				if ownerRef.Kind == "DaemonSet" {
					isDaemonSet = true
					break
				}
			}
			if isDaemonSet {
				skipped = append(skipped, fmt.Sprintf("%s/%s (DaemonSet)", pod.Namespace, pod.Name))
				continue
			}
		}

		// Skip pods with local storage unless DeleteLocalData is set
		if !opts.DeleteLocalData {
			hasLocalStorage := false
			for _, vol := range pod.Spec.Volumes {
				if vol.EmptyDir != nil {
					hasLocalStorage = true
					break
				}
			}
			if hasLocalStorage {
				skipped = append(skipped, fmt.Sprintf("%s/%s (local storage)", pod.Namespace, pod.Name))
				continue
			}
		}

		// Evict the pod
		eviction := &policyv1.Eviction{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			},
		}
		if opts.GracePeriod >= 0 {
			gracePeriod := int64(opts.GracePeriod)
			eviction.DeleteOptions = &metav1.DeleteOptions{
				GracePeriodSeconds: &gracePeriod,
			}
		}
		if err := client.CoreV1().Pods(pod.Namespace).EvictV1(context.TODO(), eviction); err != nil {
			return evicted, skipped, fmt.Errorf("驱逐pod %s/%s 失败:%s", pod.Namespace, pod.Name, err.Error())
		}
		evicted = append(evicted, fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))
	}
	return evicted, skipped, nil
}

// EvictsNodeSinglePod
//
//	@Description: 驱逐节点上一个pod
//	@param client
//	@param namespace
//	@param podName
//	@return error
func EvictsNodeSinglePod(client *kubernetes.Clientset, namespace, podName string) error {
	if err := client.CoreV1().Pods(namespace).EvictV1(context.TODO(), &policyv1.Eviction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
		},
	}); err != nil {
		return fmt.Errorf("驱逐失败:%s", err.Error())
	}
	return nil
}

// SetTaintNode
//
//	@Description: 追加单个污点
//	@param client
//	@param nodeName
//	@param key
//	@param value
//	@param effect NoSchedule PreferNoSchedule NoExecute
//	@return error
func SetTaintNode(client *kubernetes.Clientset, nodeName, key, value string, effect corev1.TaintEffect) error {
	node, err := client.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	node.Spec.Taints = append(node.Spec.Taints, corev1.Taint{
		Key:    key,
		Value:  value,
		Effect: effect,
	})
	_, err = client.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// DeleteNode
//
//	@Description: 删除节点（先解除封锁再从集群移除）
//	@param client
//	@param nodeName
//	@return error
func DeleteNode(client *kubernetes.Clientset, nodeName string) error {
	return client.CoreV1().Nodes().Delete(context.TODO(), nodeName, metav1.DeleteOptions{})
}

// UpdateNodeLabels
//
//	@Description: 替换式更新节点标签（传入完整标签 map）
//	@param client
//	@param nodeName
//	@param labels 完整的标签 map，将替换现有所有标签
//	@return error
func UpdateNodeLabels(client *kubernetes.Clientset, nodeName string, labels map[string]string) error {
	node, err := client.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	node.Labels = labels
	_, err = client.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// UpdateNodeTaints
//
//	@Description: 替换式更新节点污点（传入完整污点列表）
//	@param client
//	@param nodeName
//	@param taints 完整的污点列表，将替换现有所有污点
//	@return error
func UpdateNodeTaints(client *kubernetes.Clientset, nodeName string, taints []corev1.Taint) error {
	node, err := client.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	node.Spec.Taints = taints
	_, err = client.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}
