package daemonset

import (
	"context"
	apperr "gkube/pkg/errors"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/yaml"

	k8sEvent "gkube/pkg/k8s/event"
	"gkube/pkg/yamlutil"
)

// GetDaemonSetList
//
//	@Description: 获取daemonSet列表
//	@param client
//	@param namespace
//	@return []appsv1.DaemonSet
//	@return error
func GetDaemonSetList(ctx context.Context, client *kubernetes.Clientset, namespace string, labelSelector string) ([]appsv1.DaemonSet, error) {
	listOpts := metav1.ListOptions{ResourceVersion: "0"}
	if labelSelector != "" {
		listOpts.LabelSelector = labelSelector
	}
	daemonSetList, err := client.AppsV1().DaemonSets(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, err
	}
	return daemonSetList.Items, nil
}

// ListDaemonSets returns a paginated daemonset list with metadata
func ListDaemonSets(ctx context.Context, client *kubernetes.Clientset, namespace string, limit int64, continueToken string, labelSelector string) (*appsv1.DaemonSetList, error) {
	listOpts := metav1.ListOptions{ResourceVersion: "0"}
	if limit > 0 {
		listOpts.Limit = limit
	}
	if continueToken != "" {
		listOpts.Continue = continueToken
	}
	if labelSelector != "" {
		listOpts.LabelSelector = labelSelector
	}
	return client.AppsV1().DaemonSets(namespace).List(ctx, listOpts)
}

// GetDaemonSetByName
//
//	@Description: 获取daemonSet
//	@param client
//	@param namespace
//	@param name
//	@return *appsv1.DaemonSet
//	@return error
func GetDaemonSetByName(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (*appsv1.DaemonSet, error) {
	daemonSet, err := client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return daemonSet, nil
}

// GetDaemonSetYaml
//
//	@Description: 获取daemonSetYaml
//	@param client
//	@param namespace
//	@param name
//	@return string
//	@return error
func GetDaemonSetYaml(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (string, error) {
	daemonSet, err := client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return yamlutil.MarshalWithoutManagedFields(daemonSet)
}

// GetDaemonSetByField
//
//	@Description: 根据字段查询daemonSet
//	@param client
//	@param namespace
//	@param fieldMap
//	@return []appsv1.DaemonSet
//	@return error
func GetDaemonSetByField(ctx context.Context, client *kubernetes.Clientset, namespace string, fieldMap map[string]string) ([]appsv1.DaemonSet, error) {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	daemonSetList, err := client.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{
		FieldSelector:  fieldSelector.String(),
		ResourceVersion: "0",
	})
	if err != nil {
		return nil, err
	}
	return daemonSetList.Items, nil
}

// GetDaemonSetByLabel
//
//	@Description: 根据标签查询daemonSet
//	@param client
//	@param namespace
//	@param labelMap
//	@return []appsv1.DaemonSet
//	@return error
func GetDaemonSetByLabel(ctx context.Context, client *kubernetes.Clientset, namespace string, labelMap map[string]string) ([]appsv1.DaemonSet, error) {
	labelSelector := labels.Set(labelMap).AsSelectorPreValidated()
	daemonSetList, err := client.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{
		LabelSelector:  labelSelector.String(),
		ResourceVersion: "0",
	})
	if err != nil {
		return nil, err
	}
	return daemonSetList.Items, nil
}

// CreateDaemonSet
//
//	@Description: 创建daemonSet
//	@param client
//	@param namespace
//	@param daemonSetYaml
//	@return error
func CreateDaemonSet(ctx context.Context, client *kubernetes.Clientset, namespace, daemonSetYaml string) error {
	daemonSet := &appsv1.DaemonSet{}
	err := yaml.Unmarshal([]byte(daemonSetYaml), daemonSet)
	if err != nil {
		return apperr.BadRequest("yaml解析失败", err)
	}
	daemonSet.Namespace = namespace
	_, err = client.AppsV1().DaemonSets(namespace).Create(ctx, daemonSet, metav1.CreateOptions{})
	if err != nil {
		return apperr.K8sAPIFail("创建DaemonSet资源失败", err)
	}
	return nil
}

// UpdateDaemonSet
//
//	@Description: 更新daemonSet
//	@param client
//	@param namespace
//	@param name
//	@param daemonSetYaml
//	@return error
func UpdateDaemonSet(ctx context.Context, client *kubernetes.Clientset, namespace, name, daemonSetYaml string) error {
	daemonSet := &appsv1.DaemonSet{}
	err := yaml.Unmarshal([]byte(daemonSetYaml), daemonSet)
	if err != nil {
		return apperr.BadRequest("yaml解析失败", err)
	}
	// 校验 YAML 中的名称与请求指定的一致，避免误更新同名空间下的其他资源
	if daemonSet.Name != name {
		return apperr.BadRequest("资源名称不匹配", fmt.Errorf("请求指定 %s, YAML 中为 %s", name, daemonSet.Name))
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return apperr.K8sAPIFail("获取daemonSet资源失败", err)
		}
		latest.Spec = daemonSet.Spec
		latest.Labels = daemonSet.Labels
		latest.Annotations = daemonSet.Annotations
		_, err = client.AppsV1().DaemonSets(namespace).Update(ctx, latest, metav1.UpdateOptions{})
		return err
	})
}

// DeleteDaemonSetByName
//
//	@Description: 删除daemonSet
//	@param client
//	@param namespace
//	@param name
//	@return error
func DeleteDaemonSetByName(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error {
	propagation := metav1.DeletePropagationForeground
	err := client.AppsV1().DaemonSets(namespace).Delete(ctx, name, metav1.DeleteOptions{
		PropagationPolicy: &propagation,
	})
	if err != nil {
		return apperr.K8sAPIFail("删除daemonSet资源失败", err)
	}
	return nil
}

// DeleteDaemonSetByField
//
//	@Description: 根据字段删除daemonSet
//	@param client
//	@param namespace
//	@param fieldMap
//	@return error
func DeleteDaemonSetByField(ctx context.Context, client *kubernetes.Clientset, namespace string, fieldMap map[string]string) error {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	err := client.AppsV1().DaemonSets(namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return apperr.K8sAPIFail("删除daemonSet资源失败", err)
	}
	return nil
}

// DeleteDaemonSetByLabel
//
//	@Description: 根据标签删除daemonSet
//	@param client
//	@param namespace
//	@param labelMap
//	@return error
func DeleteDaemonSetByLabel(ctx context.Context, client *kubernetes.Clientset, namespace string, labelMap map[string]string) error {
	labelSelector := labels.Set(labelMap).AsSelectorPreValidated()
	err := client.AppsV1().DaemonSets(namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return apperr.K8sAPIFail("删除daemonSet资源失败", err)
	}
	return nil
}

// DaemonSetPodList
//
//	@Description: 获取daemonSet关联的pod列表
//	@param client
//	@param namespace
//	@param name
//	@return *corev1.PodList
//	@return error
func DaemonSetPodList(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (*corev1.PodList, error) {
	daemonSet, err := client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, apperr.K8sAPIFail("获取daemonSet资源失败", err)
	}
	// 使用完整 selector(matchLabels + matchExpressions),与 deployment 行为对齐
	if daemonSet.Spec.Selector == nil {
		return &corev1.PodList{Items: []corev1.Pod{}}, nil
	}
	selector, err := metav1.LabelSelectorAsSelector(daemonSet.Spec.Selector)
	if err != nil {
		return nil, apperr.K8sAPIFail("解析daemonSet selector失败", err)
	}
	if selector.Empty() {
		return &corev1.PodList{Items: []corev1.Pod{}}, nil
	}
	podList, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector:  selector.String(),
		ResourceVersion: "0",
	})
	if err != nil {
		return nil, apperr.K8sAPIFail("获取pod资源失败", err)
	}
	return podList, nil
}

// RestartDaemonSet
//
//	@Description: 重启daemonSet
//	@param client
//	@param namespace
//	@param name
//	@return bool
//	@return error
func RestartDaemonSet(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (bool, error) {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		daemonSet, err := client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return apperr.K8sAPIFail("获取daemonSet资源失败", err)
		}
		if daemonSet.Spec.Template.Annotations == nil {
			daemonSet.Spec.Template.Annotations = make(map[string]string)
		}
		daemonSet.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.DateTime)
		_, err = client.AppsV1().DaemonSets(namespace).Update(ctx, daemonSet, metav1.UpdateOptions{})
		return err
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func UpdateDaemonSetImage(ctx context.Context, client *kubernetes.Clientset, namespace, name, containerName, image string) (*appsv1.DaemonSet, error) {
	var latest *appsv1.DaemonSet
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		ds, err := client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		for i, c := range ds.Spec.Template.Spec.Containers {
			if c.Name == containerName {
				ds.Spec.Template.Spec.Containers[i].Image = image
				latest, err = client.AppsV1().DaemonSets(namespace).Update(ctx, ds, metav1.UpdateOptions{})
				return err
			}
		}
		return apperr.BadRequest("容器不存在", fmt.Errorf("容器 %s 不存在", containerName))
	})
	return latest, err
}

// RollbackDaemonSet
//
//	@Description: 回滚daemonSet到指定revision
//	@param client
//	@param namespace
//	@param name
//	@param revision
//	@return *appsv1.DaemonSet
//	@return error
func RollbackDaemonSet(ctx context.Context, client *kubernetes.Clientset, namespace, name string, revision int64) (*appsv1.DaemonSet, error) {
	if revision <= 0 {
		return nil, apperr.BadRequest("revision必须为正整数", fmt.Errorf("revision must be a positive integer, got %d", revision))
	}
	ds, err := client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, apperr.K8sAPIFail("获取DaemonSet资源失败", err)
	}
	revisions, err := client.AppsV1().ControllerRevisions(namespace).List(ctx, metav1.ListOptions{
		LabelSelector:  metav1.FormatLabelSelector(ds.Spec.Selector),
		ResourceVersion: "0",
	})
	if err != nil {
		return nil, apperr.K8sAPIFail("获取ControllerRevision列表失败", err)
	}
	var targetData []byte
	for _, rev := range revisions.Items {
		if rev.Revision == revision {
			targetData = rev.Data.Raw
			break
		}
	}
	if len(targetData) == 0 {
		return nil, apperr.BadRequest("revision不存在", fmt.Errorf("revision %d 不存在", revision))
	}
	var restored appsv1.DaemonSet
	if err := json.Unmarshal(targetData, &restored); err != nil {
		return nil, apperr.BadRequest("解析revision数据失败", err)
	}
	var latest *appsv1.DaemonSet
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current, err := client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return apperr.K8sAPIFail("获取DaemonSet资源失败", err)
		}
		// 只回滚 pod template（与 Deployment/StatefulSet 回滚逻辑一致），
		// 不覆盖 selector 等不可变字段，避免触发 Forbidden。
		current.Spec.Template = restored.Spec.Template
		latest, err = client.AppsV1().DaemonSets(namespace).Update(ctx, current, metav1.UpdateOptions{})
		return err
	})
	return latest, err
}

// GetDaemonSetRollbacks returns all ControllerRevision entries for a DaemonSet,
// sorted by revision descending, for UI rollback selection.
func GetDaemonSetRollbacks(ctx context.Context, client *kubernetes.Clientset, namespace, name string) ([]map[string]any, error) {
	ds, err := client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, apperr.K8sAPIFail("获取DaemonSet资源失败", err)
	}
	revisions, err := client.AppsV1().ControllerRevisions(namespace).List(ctx, metav1.ListOptions{
		LabelSelector:  metav1.FormatLabelSelector(ds.Spec.Selector),
		ResourceVersion: "0",
	})
	if err != nil {
		return nil, apperr.K8sAPIFail("获取ControllerRevision列表失败", err)
	}
	// 反序列化 ControllerRevision.Data.Raw 取出容器镜像，方便回滚时确认目标版本。
	// DaemonSet（与 StatefulSet 不同）的 status 不暴露 currentRevision，
	// 所以通过对比 ControllerRevision 中存储的 PodTemplateSpec 与当前 DaemonSet
	// 的 Spec.Template 来判定哪个是“当前”版本，供前端标记“当前”并隐藏其回滚按钮。
	type revEntry struct {
		Revision  int64    `json:"revision"`
		Name      string   `json:"name"`
		Images    []string `json:"images"`
		CreatedAt string   `json:"createdAt"`
		IsCurrent bool     `json:"isCurrent"`
	}
	var entries []revEntry
	for _, rev := range revisions.Items {
		var restored appsv1.DaemonSet
		var images []string
		isCurrent := false
		if err := json.Unmarshal(rev.Data.Raw, &restored); err == nil {
			for _, c := range restored.Spec.Template.Spec.Containers {
				images = append(images, c.Image)
			}
			isCurrent = reflect.DeepEqual(restored.Spec.Template, ds.Spec.Template)
		}
		entries = append(entries, revEntry{
			Revision:  rev.Revision,
			Name:      rev.Name,
			Images:    images,
			CreatedAt: rev.CreationTimestamp.Time.Format(time.RFC3339),
			IsCurrent: isCurrent,
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[j].Revision > entries[i].Revision
	})
	result := make([]map[string]any, len(entries))
	for i, e := range entries {
		result[i] = map[string]any{
			"revision":  e.Revision,
			"name":      e.Name,
			"images":    e.Images,
			"createdAt": e.CreatedAt,
			"isCurrent": e.IsCurrent,
		}
	}
	return result, nil
}

// GetDaemonSetEvents returns the events associated with a DaemonSet.
// 用 fields.Selector 防注入。
func GetDaemonSetEvents(ctx context.Context, client *kubernetes.Clientset, namespace, name string) ([]k8sEvent.KubeEvent, error) {
	selector := fields.AndSelectors(
		fields.OneTermEqualSelector("involvedObject.name", name),
		fields.OneTermEqualSelector("involvedObject.kind", "DaemonSet"),
	).String()
	events, _, _, err := k8sEvent.ListEvents(ctx, client, namespace, selector, 0, "")
	if err != nil {
		return nil, apperr.K8sAPIFail("获取daemonset事件失败", err)
	}
	return events, nil
}
