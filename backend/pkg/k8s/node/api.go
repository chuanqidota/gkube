package node

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gkube/pkg/yamlutil"

	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// requestTimeout 单次 K8s API 调用超时。
// drain 是长操作，整体不设超时（逐次驱逐受 rest.Config.Timeout 约束）。
const requestTimeout = 30 * time.Second

// ErrYamlParse YAML 解析失败，handler 据此返回 400（客户端错误）而非 502。
var ErrYamlParse = errors.New("YAML解析失败")

const roleLabelPrefix = "node-role.kubernetes.io/"

// protectedLabelPrefixes 下的标签受保护：替换式更新标签时不会被删除或覆盖，
// 避免 frontend 漏传导致 kubernetes.io/hostname 等节点身份标签丢失、调度异常。
var protectedLabelPrefixes = []string{
	"kubernetes.io/",
	"node-role.kubernetes.io/",
	"node.kubernetes.io/",
	"k8s.io/",
}

func isProtectedLabel(key string) bool {
	for _, p := range protectedLabelPrefixes {
		if strings.HasPrefix(key, p) {
			return true
		}
	}
	return false
}

// NodeStatus 从 conditions 中解析节点就绪状态。
// 返回 "Ready" / "NotReady" / "Unknown" 以及 isReady。
func NodeStatus(conditions []corev1.NodeCondition) (status string, isReady bool) {
	status = "Unknown"
	for _, cond := range conditions {
		if cond.Type != corev1.NodeReady {
			continue
		}
		if cond.Status == corev1.ConditionTrue {
			status = "Ready"
			isReady = true
		} else {
			status = "NotReady"
		}
		return
	}
	return
}

// NodeRoles 从标签中提取角色名（node-role.kubernetes.io/<role>），逗号分隔。
func NodeRoles(labels map[string]string) string {
	var roles []string
	for label := range labels {
		if !strings.HasPrefix(label, roleLabelPrefix) {
			continue
		}
		if role := label[len(roleLabelPrefix):]; role != "" {
			roles = append(roles, role)
		}
	}
	return strings.Join(roles, ", ")
}

// TaintView / ConditionView / EventView：节点详情对外暴露的视图结构，
// 字段与 frontend NodeDetail interface 对齐。
type TaintView struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Effect string `json:"effect"`
}

type ConditionView struct {
	Type               string `json:"type"`
	Status             string `json:"status"`
	Reason             string `json:"reason"`
	Message            string `json:"message"`
	LastTransitionTime string `json:"lastTransitionTime"`
}

type EventView struct {
	Type     string `json:"type"`
	Reason   string `json:"reason"`
	Message  string `json:"message"`
	LastSeen string `json:"last_seen"`
}

// PodConditionView / ContainerStatusView：PodView.Status 下的命名子结构，
// 避免匿名 struct 在 podToView 里重复声明类型。
type PodConditionView struct {
	Type   string `json:"type"`
	Status string `json:"status"`
}

type ContainerStatusView struct {
	RestartCount int32 `json:"restartCount"`
}

// PodView 节点上 Pod 的视图，仅暴露前端所需字段，避免序列化完整 corev1.Pod。
// CreationTimestamp 保留 ISO 原始字符串，前端用 formatAge 计算相对年龄（与全站一致）。
// 字段与 frontend K8sPod interface 对齐。
type PodView struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Status    struct {
		Phase             string                 `json:"phase"`
		PodIP             string                 `json:"podIP"`
		Conditions        []PodConditionView     `json:"conditions"`
		ContainerStatuses []ContainerStatusView  `json:"containerStatuses"`
	} `json:"status"`
	CreationTimestamp string `json:"creationTimestamp"`
}

// podToView 将 corev1.Pod 投影为 PodView。
func podToView(pod corev1.Pod) PodView {
	v := PodView{
		Name:              pod.Name,
		Namespace:         pod.Namespace,
		CreationTimestamp: pod.CreationTimestamp.Format(time.RFC3339),
	}
	v.Status.Phase = string(pod.Status.Phase)
	v.Status.PodIP = pod.Status.PodIP
	for _, c := range pod.Status.Conditions {
		v.Status.Conditions = append(v.Status.Conditions, PodConditionView{
			Type:   string(c.Type),
			Status: string(c.Status),
		})
	}
	for _, cs := range pod.Status.ContainerStatuses {
		v.Status.ContainerStatuses = append(v.Status.ContainerStatuses, ContainerStatusView{
			RestartCount: cs.RestartCount,
		})
	}
	return v
}

// Detail 节点详情，字段与 frontend NodeDetail interface 对齐。
type Detail struct {
	Name             string              `json:"name"`
	Status           string              `json:"status"`
	Roles            string              `json:"roles"`
	Version          string              `json:"version"`
	OS               string              `json:"os"`
	Kernel           string              `json:"kernel"`
	ContainerRuntime string              `json:"container_runtime"`
	Architecture     string              `json:"architecture"`
	InternalIP       string              `json:"internal_ip"`
	ExternalIP       string              `json:"external_ip"`
	Hostname         string              `json:"hostname"`
	Unschedulable    bool                `json:"unschedulable"`
	Labels           map[string]string   `json:"labels"`
	Taints           []TaintView         `json:"taints"`
	Conditions       []ConditionView     `json:"conditions"`
	Capacity         corev1.ResourceList `json:"capacity"`
	Allocatable      corev1.ResourceList `json:"allocatable"`
	CreationTimestamp string             `json:"creationTimestamp"`
}

// GetNodeYaml 获取 node 的 yaml。
func GetNodeYaml(client *kubernetes.Clientset, nodeName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	node.TypeMeta = metav1.TypeMeta{APIVersion: "v1", Kind: "Node"}
	return yamlutil.MarshalWithoutManagedFields(node)
}

// UpdateNodeYaml 通过 YAML 更新节点：保留服务端 resourceVersion 防止版本冲突。
func UpdateNodeYaml(client *kubernetes.Clientset, nodeName, yamlStr string) error {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	current, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	var nodeObj corev1.Node
	if err := yaml.Unmarshal([]byte(yamlStr), &nodeObj); err != nil {
		return fmt.Errorf("%w: %s", ErrYamlParse, err.Error())
	}
	nodeObj.ResourceVersion = current.ResourceVersion
	_, err = client.CoreV1().Nodes().Update(ctx, &nodeObj, metav1.UpdateOptions{})
	return err
}

// GetNodePods 获取 node 上的非终态 pod，投影为 PodView 仅暴露前端所需字段。
func GetNodePods(client *kubernetes.Clientset, nodeName string) ([]PodView, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	// 用 OneTermEqualSelector 构造，避免 nodeName 含特殊字符时拼出非法 field selector
	selector := fields.OneTermEqualSelector("spec.nodeName", nodeName)
	podList, err := client.CoreV1().Pods(corev1.NamespaceAll).List(ctx, metav1.ListOptions{
		FieldSelector: selector.String(),
	})
	if err != nil {
		return nil, err
	}
	var pods []PodView
	for _, pod := range podList.Items {
		if pod.Status.Phase != corev1.PodSucceeded && pod.Status.Phase != corev1.PodFailed {
			pods = append(pods, podToView(pod))
		}
	}
	return pods, nil
}

// CordonNode 封锁或解除封锁节点。返回当前封锁状态。
func CordonNode(client *kubernetes.Clientset, nodeName string, cordon bool) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	node.Spec.Unschedulable = cordon
	if _, err := client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{}); err != nil {
		return false, err
	}
	return cordon, nil
}

// DrainOptions 驱逐选项
type DrainOptions struct {
	IgnoreDaemonSets bool `json:"ignoreDaemonSets"` // 是否忽略 DaemonSet 管理的 Pod
	DeleteLocalData  bool `json:"deleteLocalData"`  // 是否删除使用 emptyDir 等本地存储的 Pod
	GracePeriod      int  `json:"gracePeriod"`      // 优雅终止超时秒数，-1 使用 Pod 默认值
	Force            bool `json:"force"`            // 是否强制驱逐 kube-system 下的 Pod
}

// DrainNode 驱逐节点上的所有 pod（先封锁再驱逐）。
// 返回被驱逐/被跳过/驱逐失败的 pod 列表。
// cordon 与 list pods 是短调用，用显式 30s 超时；evict 循环是长操作，
// 保留 context.TODO()，逐次驱逐受 rest.Config.Timeout 约束。
// 与 kubectl drain 一致：单个 pod 驱逐失败不中断，继续尝试其余 pod，最后汇总失败列表。
func DrainNode(client *kubernetes.Clientset, nodeName string, opts DrainOptions) (evicted []string, skipped []string, failed []string, err error) {
	// Step 1: Cordon the node first（短调用，显式超时）
	cordonCtx, cordonCancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cordonCancel()
	node, err := client.CoreV1().Nodes().Get(cordonCtx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("获取节点失败:%s", err.Error())
	}
	if !node.Spec.Unschedulable {
		node.Spec.Unschedulable = true
		if _, err := client.CoreV1().Nodes().Update(cordonCtx, node, metav1.UpdateOptions{}); err != nil {
			return nil, nil, nil, fmt.Errorf("封锁节点失败:%s", err.Error())
		}
	}

	// Step 2: List all pods on the node（短调用，显式超时）
	selector := fields.OneTermEqualSelector("spec.nodeName", nodeName)
	pods, err := client.CoreV1().Pods(corev1.NamespaceAll).List(cordonCtx, metav1.ListOptions{
		FieldSelector: selector.String(),
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("获取节点pod列表失败:%s", err.Error())
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

		// Evict the pod — 失败不中断，记录到 failed 列表后继续下一个
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
			failed = append(failed, fmt.Sprintf("%s/%s (%s)", pod.Namespace, pod.Name, err.Error()))
			continue
		}
		evicted = append(evicted, fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))
	}
	return evicted, skipped, failed, nil
}

// DeleteNode 从集群中删除节点。
func DeleteNode(client *kubernetes.Clientset, nodeName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	return client.CoreV1().Nodes().Delete(ctx, nodeName, metav1.DeleteOptions{})
}

// UpdateNodeLabels 更新节点标签：保留受保护的系统标签，仅替换用户标签。
// 传入的 labels 作为用户标签的完整期望集——不在其中的用户标签会被删除，
// 受保护前缀（kubernetes.io/ 等）下的标签保持原值不被改动。
func UpdateNodeLabels(client *kubernetes.Clientset, nodeName string, labels map[string]string) error {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	merged := make(map[string]string)
	// 保留受保护的系统标签（节点身份/角色等）
	for k, v := range node.Labels {
		if isProtectedLabel(k) {
			merged[k] = v
		}
	}
	// 合并用户标签：非受保护的 key 按传入 map 覆盖/新增
	for k, v := range labels {
		if !isProtectedLabel(k) {
			merged[k] = v
		}
	}
	node.Labels = merged
	_, err = client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	return err
}

// UpdateNodeTaints 替换式更新节点污点（传入完整污点列表）。
func UpdateNodeTaints(client *kubernetes.Clientset, nodeName string, taints []corev1.Taint) error {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	node.Spec.Taints = taints
	_, err = client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	return err
}

// GetNodeDetail 获取节点详情（conditions / addresses / labels / taints / capacity 等）。
func GetNodeDetail(client *kubernetes.Clientset, nodeName string) (*Detail, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	status, _ := NodeStatus(node.Status.Conditions)

	var conditions []ConditionView
	for _, cond := range node.Status.Conditions {
		conditions = append(conditions, ConditionView{
			Type:               string(cond.Type),
			Status:             string(cond.Status),
			Reason:             cond.Reason,
			Message:            cond.Message,
			LastTransitionTime: cond.LastTransitionTime.Time.Format("2006-01-02 15:04:05"),
		})
	}

	var internalIP, externalIP, hostname string
	for _, addr := range node.Status.Addresses {
		switch addr.Type {
		case corev1.NodeInternalIP:
			internalIP = addr.Address
		case corev1.NodeExternalIP:
			externalIP = addr.Address
		case corev1.NodeHostName:
			hostname = addr.Address
		}
	}

	var taints []TaintView
	for _, t := range node.Spec.Taints {
		taints = append(taints, TaintView{
			Key:    t.Key,
			Value:  t.Value,
			Effect: string(t.Effect),
		})
	}

	labels := node.Labels
	if labels == nil {
		labels = map[string]string{}
	}

	return &Detail{
		Name:             node.Name,
		Status:           status,
		Roles:            NodeRoles(node.Labels),
		Version:          node.Status.NodeInfo.KubeletVersion,
		OS:               node.Status.NodeInfo.OSImage,
		Kernel:           node.Status.NodeInfo.KernelVersion,
		ContainerRuntime: node.Status.NodeInfo.ContainerRuntimeVersion,
		Architecture:     node.Status.NodeInfo.Architecture,
		InternalIP:       internalIP,
		ExternalIP:       externalIP,
		Hostname:         hostname,
		Unschedulable:    node.Spec.Unschedulable,
		Labels:           labels,
		Taints:           taints,
		Conditions:       conditions,
		Capacity:         node.Status.Capacity,
		Allocatable:      node.Status.Allocatable,
		CreationTimestamp: node.CreationTimestamp.Format(time.RFC3339),
	}, nil
}

// GetNodeEvents 获取节点相关事件。
func GetNodeEvents(client *kubernetes.Clientset, nodeName string) ([]EventView, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	// 用 fields.AndSelectors 构造，避免 nodeName 注入非法字段选择器语法
	selector := fields.AndSelectors(
		fields.OneTermEqualSelector("involvedObject.name", nodeName),
		fields.OneTermEqualSelector("involvedObject.kind", "Node"),
	)
	events, err := client.CoreV1().Events(corev1.NamespaceAll).List(ctx, metav1.ListOptions{
		FieldSelector: selector.String(),
	})
	if err != nil {
		return nil, err
	}

	var result []EventView
	for _, event := range events.Items {
		lastSeen := ""
		if !event.LastTimestamp.IsZero() {
			lastSeen = event.LastTimestamp.Time.Format("2006-01-02 15:04:05")
		}
		result = append(result, EventView{
			Type:     event.Type,
			Reason:   event.Reason,
			Message:  event.Message,
			LastSeen: lastSeen,
		})
	}
	return result, nil
}
