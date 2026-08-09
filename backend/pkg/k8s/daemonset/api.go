package daemonset

import (
	"context"
	"encoding/json"
	"fmt"
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

	"gkube/pkg/yamlutil"
)

// GetDaemonSetList
//
//	@Description: 获取daemonSet列表
//	@param client
//	@param namespace
//	@return []appsv1.DaemonSet
//	@return error
func GetDaemonSetList(client *kubernetes.Clientset, namespace string) ([]appsv1.DaemonSet, error) {
	daemonSetList, err := client.AppsV1().DaemonSets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return daemonSetList.Items, nil
}

// ListDaemonSets returns a paginated daemonset list with metadata
func ListDaemonSets(client *kubernetes.Clientset, namespace string, limit int64, continueToken string) (*appsv1.DaemonSetList, error) {
	listOpts := metav1.ListOptions{}
	if limit > 0 {
		listOpts.Limit = limit
	}
	if continueToken != "" {
		listOpts.Continue = continueToken
	}
	return client.AppsV1().DaemonSets(namespace).List(context.Background(), listOpts)
}

// GetDaemonSetByName
//
//	@Description: 获取daemonSet
//	@param client
//	@param namespace
//	@param name
//	@return *appsv1.DaemonSet
//	@return error
func GetDaemonSetByName(client *kubernetes.Clientset, namespace, name string) (*appsv1.DaemonSet, error) {
	daemonSet, err := client.AppsV1().DaemonSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
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
func GetDaemonSetYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	daemonSet, err := client.AppsV1().DaemonSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
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
func GetDaemonSetByField(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) ([]appsv1.DaemonSet, error) {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	daemonSetList, err := client.AppsV1().DaemonSets(namespace).List(context.Background(), metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
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
func GetDaemonSetByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) ([]appsv1.DaemonSet, error) {
	labelSelector := labels.Set(labelMap).AsSelectorPreValidated()
	daemonSetList, err := client.AppsV1().DaemonSets(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: labelSelector.String(),
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
func CreateDaemonSet(client *kubernetes.Clientset, namespace, daemonSetYaml string) error {
	daemonSet := &appsv1.DaemonSet{}
	err := yaml.Unmarshal([]byte(daemonSetYaml), daemonSet)
	if err != nil {
		return err
	}
	daemonSet.Namespace = namespace
	_, err = client.AppsV1().DaemonSets(namespace).Create(context.Background(), daemonSet, metav1.CreateOptions{})
	if err != nil {
		return err
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
func UpdateDaemonSet(client *kubernetes.Clientset, namespace, name, daemonSetYaml string) error {
	daemonSet := &appsv1.DaemonSet{}
	err := yaml.Unmarshal([]byte(daemonSetYaml), daemonSet)
	if err != nil {
		return fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	// 校验 YAML 中的名称与请求指定的一致，避免误更新同名空间下的其他资源
	if daemonSet.Name != name {
		return fmt.Errorf("资源名称不匹配: 请求指定 %s, YAML 中为 %s", name, daemonSet.Name)
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := client.AppsV1().DaemonSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("获取daemonSet资源失败:%s", err.Error())
		}
		latest.Spec = daemonSet.Spec
		latest.Labels = daemonSet.Labels
		latest.Annotations = daemonSet.Annotations
		_, err = client.AppsV1().DaemonSets(namespace).Update(context.Background(), latest, metav1.UpdateOptions{})
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
func DeleteDaemonSetByName(client *kubernetes.Clientset, namespace, name string) error {
	err := client.AppsV1().DaemonSets(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("删除daemonSet资源失败:%s", err.Error())
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
func DeleteDaemonSetByField(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) error {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	err := client.AppsV1().DaemonSets(namespace).DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return fmt.Errorf("删除daemonSet资源失败:%s", err.Error())
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
func DeleteDaemonSetByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) error {
	labelSelector := labels.Set(labelMap).AsSelectorPreValidated()
	err := client.AppsV1().DaemonSets(namespace).DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return fmt.Errorf("删除daemonSet资源失败:%s", err.Error())
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
func DaemonSetPodList(client *kubernetes.Clientset, namespace, name string) (*corev1.PodList, error) {
	daemonSet, err := client.AppsV1().DaemonSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取daemonSet资源失败:%s", err.Error())
	}
	selector := labels.Set(daemonSet.Spec.Selector.MatchLabels).AsSelectorPreValidated()
	podList, err := client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("获取pod资源失败:%s", err.Error())
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
func RestartDaemonSet(client *kubernetes.Clientset, namespace, name string) (bool, error) {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		daemonSet, err := client.AppsV1().DaemonSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("获取daemonSet资源失败:%s", err.Error())
		}
		if daemonSet.Spec.Template.Annotations == nil {
			daemonSet.Spec.Template.Annotations = make(map[string]string)
		}
		daemonSet.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.DateTime)
		_, err = client.AppsV1().DaemonSets(namespace).Update(context.Background(), daemonSet, metav1.UpdateOptions{})
		return err
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func UpdateDaemonSetImage(client *kubernetes.Clientset, namespace, name, containerName, image string) (*appsv1.DaemonSet, error) {
	var latest *appsv1.DaemonSet
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		ds, err := client.AppsV1().DaemonSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		for i, c := range ds.Spec.Template.Spec.Containers {
			if c.Name == containerName {
				ds.Spec.Template.Spec.Containers[i].Image = image
				latest, err = client.AppsV1().DaemonSets(namespace).Update(context.Background(), ds, metav1.UpdateOptions{})
				return err
			}
		}
		return fmt.Errorf("container %s not found", containerName)
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
func RollbackDaemonSet(client *kubernetes.Clientset, namespace, name string, revision int64) (*appsv1.DaemonSet, error) {
	ds, err := client.AppsV1().DaemonSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	revisions, err := client.AppsV1().ControllerRevisions(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(ds.Spec.Selector),
	})
	if err != nil {
		return nil, err
	}
	var targetData []byte
	for _, rev := range revisions.Items {
		if rev.Revision == revision {
			targetData = rev.Data.Raw
			break
		}
	}
	if len(targetData) == 0 {
		return nil, fmt.Errorf("revision %d not found", revision)
	}
	var restored appsv1.DaemonSet
	if err := json.Unmarshal(targetData, &restored); err != nil {
		return nil, err
	}
	var latest *appsv1.DaemonSet
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current, err := client.AppsV1().DaemonSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		restored.ResourceVersion = current.ResourceVersion
		restored.UID = current.UID
		latest, err = client.AppsV1().DaemonSets(namespace).Update(context.Background(), &restored, metav1.UpdateOptions{})
		return err
	})
	return latest, err
}

// GetDaemonSetRollbacks returns all ControllerRevision entries for a DaemonSet,
// sorted by revision descending, for UI rollback selection.
func GetDaemonSetRollbacks(client *kubernetes.Clientset, namespace, name string) ([]map[string]any, error) {
	ctx := context.Background()
	ds, err := client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	revisions, err := client.AppsV1().ControllerRevisions(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(ds.Spec.Selector),
	})
	if err != nil {
		return nil, err
	}
	type revEntry struct {
		Revision int64  `json:"revision"`
		Name     string `json:"name"`
	}
	var entries []revEntry
	for _, rev := range revisions.Items {
		entries = append(entries, revEntry{
			Revision: rev.Revision,
			Name:     rev.Name,
		})
	}
	// Sort descending by revision
	sort.Slice(entries, func(i, j int) bool {
		return entries[j].Revision > entries[i].Revision
	})
	result := make([]map[string]any, len(entries))
	for i, e := range entries {
		result[i] = map[string]any{
			"revision": e.Revision,
			"name":     e.Name,
		}
	}
	return result, nil
}
