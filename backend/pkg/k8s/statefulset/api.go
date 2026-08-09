package statefulset

import (
	"encoding/json"
	"gkube/pkg/yamlutil"
	"context"
	"fmt"
	"strings"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/yaml"
	"time"
)

// GetStatefulSetList
//
//	@Description: 获取statefulSet列表
//	@param client
//	@param namespace
//	@return []appsv1.StatefulSet
//	@return error
func GetStatefulSetList(client *kubernetes.Clientset, namespace string) ([]appsv1.StatefulSet, error) {
	statefulSetList, err := client.AppsV1().StatefulSets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return statefulSetList.Items, nil
}

// ListStatefulSets returns a paginated statefulset list with metadata
func ListStatefulSets(client *kubernetes.Clientset, namespace string, limit int64, continueToken string) (*appsv1.StatefulSetList, error) {
	listOpts := metav1.ListOptions{}
	if limit > 0 {
		listOpts.Limit = limit
	}
	if continueToken != "" {
		listOpts.Continue = continueToken
	}
	return client.AppsV1().StatefulSets(namespace).List(context.Background(), listOpts)
}

// GetStatefulSetByName
//
//	@Description: 获取statefulSet
//	@param client
//	@param namespace
//	@param name
//	@return *appsv1.StatefulSet
//	@return error
func GetStatefulSetByName(client *kubernetes.Clientset, namespace, name string) (*appsv1.StatefulSet, error) {
	statefulSet, err := client.AppsV1().StatefulSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return statefulSet, nil
}

// GetStatefulSetYaml
//
//	@Description: 获取statefulSet的yaml
//	@param client
//	@param namespace
//	@param name
//	@return string
//	@return error
func GetStatefulSetYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	statefulSet, err := client.AppsV1().StatefulSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	yamlStr, err := yamlutil.MarshalWithoutManagedFields(statefulSet)
	if err != nil {
		return "", err
	}
	return yamlStr, nil
}

// GetStatefulSetByField
//
//	@Description: 根据字段查询statefulSet
//	@param client
//	@param namespace
//	@param fieldMap
//	@return []appsv1.StatefulSet
//	@return error
func GetStatefulSetByField(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) ([]appsv1.StatefulSet, error) {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	statefulSetList, err := client.AppsV1().StatefulSets(namespace).List(context.Background(), metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return nil, err
	}
	return statefulSetList.Items, nil
}

// GetStatefulSetByLabel
//
//	@Description: 根据标签查询statefulSet
//	@param client
//	@param namespace
//	@param labelMap
//	@return []appsv1.StatefulSet
//	@return error
func GetStatefulSetByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) ([]appsv1.StatefulSet, error) {
	labelSelector := labels.Set(labelMap).AsSelectorPreValidated()
	statefulSetList, err := client.AppsV1().StatefulSets(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return nil, err
	}
	return statefulSetList.Items, nil
}

// CreateStatefulSet
//
//	@Description: 创建statefulSet
//	@param client
//	@param namespace
//	@param statefulSetYaml
//	@return error
func CreateStatefulSet(client *kubernetes.Clientset, namespace, statefulSetYaml string) error {
	var statefulSet appsv1.StatefulSet
	if err := yaml.Unmarshal([]byte(statefulSetYaml), &statefulSet); err != nil {
		return fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	statefulSet.Namespace = namespace
	_, err := client.AppsV1().StatefulSets(namespace).Create(context.Background(), &statefulSet, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("创建statefulSet资源失败:%s", err.Error())
	}
	return nil
}

// UpdateStatefulSet
//
//	@Description: 更新statefulset
//	@param client
//	@param namespace
//	@param name
//	@param statefulSetYaml
//	@return error
func UpdateStatefulSet(client *kubernetes.Clientset, namespace, name, statefulSetYaml string) error {
	var statefulSet appsv1.StatefulSet
	if err := yaml.Unmarshal([]byte(statefulSetYaml), &statefulSet); err != nil {
		return fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	// 校验 YAML 中的名称与请求指定的一致，避免误更新同名空间下的其他资源
	if statefulSet.Name != name {
		return fmt.Errorf("资源名称不匹配: 请求指定 %s, YAML 中为 %s", name, statefulSet.Name)
	}
	statefulSet.Namespace = namespace
	// 冲突时自动重试(与 Deployment Update 保持一致的 RetryOnConflict 模式)
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := client.AppsV1().StatefulSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("获取statefulset资源失败:%s", err.Error())
		}
		latest.Spec = statefulSet.Spec
		latest.Labels = statefulSet.Labels
		latest.Annotations = statefulSet.Annotations
		_, err = client.AppsV1().StatefulSets(namespace).Update(context.Background(), latest, metav1.UpdateOptions{})
		return err
	}); err != nil {
		return fmt.Errorf("更新statefulset资源失败:%s", err.Error())
	}
	return nil
}

// DeleteStatefulSetByName
//
//	@Description: 删除statefulSet
//	@param client
//	@param namespace
//	@param name
//	@return error
func DeleteStatefulSetByName(client *kubernetes.Clientset, namespace, name string) error {
	err := client.AppsV1().StatefulSets(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("删除statefulSet资源失败:%s", err.Error())
	}
	return nil
}

// DeleteStatefulSetByLabel
//
//	@Description: 根据标签删除statefulSet
//	@param client
//	@param namespace
//	@param labelMap
//	@return error
func DeleteStatefulSetByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) error {
	labelSelector := labels.Set(labelMap).AsSelectorPreValidated()
	err := client.AppsV1().StatefulSets(namespace).DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return fmt.Errorf("删除statefulSet资源失败:%s", err.Error())
	}
	return nil
}

// DeleteStatefulSetByField
//
//	@Description: 根据字段删除statefulSet
//	@param client
//	@param namespace
//	@param fieldMap
//	@return error
func DeleteStatefulSetByField(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) error {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	err := client.AppsV1().StatefulSets(namespace).DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return fmt.Errorf("删除statefulSet资源失败:%s", err.Error())
	}
	return nil
}

// StatefulSetPodList
//
//	@Description: 获取statefulSet关联的pod列表
//	@param client
//	@param namespace
//	@param name
//	@return *corev1.PodList
//	@return error
func StatefulSetPodList(client *kubernetes.Clientset, namespace, name string) (*corev1.PodList, error) {
	statefulSet, err := client.AppsV1().StatefulSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取statefulSet资源失败:%s", err.Error())
	}
	selector := labels.Set(statefulSet.Spec.Selector.MatchLabels).AsSelectorPreValidated()
	podList, err := client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("获取pod资源失败:%s", err.Error())
	}
	return podList, nil
}

// ScaleStatefulSet
//
//	@Description: 扩缩容statefulSet
//	@param client
//	@param namespace
//	@param name
//	@param replicas
//	@return bool
//	@return error
func ScaleStatefulSet(client *kubernetes.Clientset, namespace, name string, replicas int32) (bool, error) {
	statefulSet, err := client.AppsV1().StatefulSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("获取statefulSet资源失败:%s", err.Error())
	}
	statefulSet.Spec.Replicas = &replicas
	_, err = client.AppsV1().StatefulSets(namespace).Update(context.Background(), statefulSet, metav1.UpdateOptions{})
	if err != nil {
		return false, fmt.Errorf("更新statefulSet资源失败:%s", err.Error())
	}
	return true, nil
}

// RestartStatefulSet
//
//	@Description: 重启statefulSet
//	@param client
//	@param namespace
//	@param name
//	@return bool
//	@return error
func RestartStatefulSet(client *kubernetes.Clientset, namespace, name string) (bool, error) {
	statefulSet, err := client.AppsV1().StatefulSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("获取statefulSet资源失败:%s", err.Error())
	}
	if statefulSet.Spec.Template.Annotations == nil {
		statefulSet.Spec.Template.Annotations = make(map[string]string)
	}
	statefulSet.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.DateTime)
	_, err = client.AppsV1().StatefulSets(namespace).Update(context.Background(), statefulSet, metav1.UpdateOptions{})
	if err != nil {
		return false, fmt.Errorf("更新statefulSet资源失败:%s", err.Error())
	}
	return true, nil
}

func UpdateStatefulSetImage(client *kubernetes.Clientset, namespace, name, containerName, image string) (*appsv1.StatefulSet, error) {
	ctx := context.Background()
	// 冲突时自动重试(与 UpdateStatefulSet 保持一致)：每次重试都重新获取最新对象再改镜像。
	// RetryOnConflict 只对 409 重试，"容器不存在"这类错误会立即返回。
	var result *appsv1.StatefulSet
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("获取statefulset资源失败:%s", err.Error())
		}
		found := false
		for i, c := range latest.Spec.Template.Spec.Containers {
			if c.Name == containerName {
				latest.Spec.Template.Spec.Containers[i].Image = image
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("容器 %s 不存在", containerName)
		}
		result, err = client.AppsV1().StatefulSets(namespace).Update(ctx, latest, metav1.UpdateOptions{})
		return err
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func RollbackStatefulSet(client *kubernetes.Clientset, namespace, name string, revision int64) (*appsv1.StatefulSet, error) {
	ctx := context.Background()
	sts, err := client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	revisions, err := client.AppsV1().ControllerRevisions(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(sts.Spec.Selector),
	})
	if err != nil {
		return nil, err
	}
	var restored appsv1.StatefulSet
	found := false
	for _, rev := range revisions.Items {
		if rev.Revision == revision {
			if err := json.Unmarshal(rev.Data.Raw, &restored); err != nil {
				return nil, fmt.Errorf("解析 revision %d 失败: %s", revision, err.Error())
			}
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("revision %d 不存在", revision)
	}
	// 冲突时自动重试(与 UpdateStatefulSet 保持一致的 RetryOnConflict 模式)：
	// 重新获取最新 resourceVersion 后再用 revision 的 spec 覆盖，避免并发编辑导致 409。
	var result *appsv1.StatefulSet
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("获取statefulset资源失败:%s", err.Error())
		}
		latest.Spec = restored.Spec
		result, err = client.AppsV1().StatefulSets(namespace).Update(ctx, latest, metav1.UpdateOptions{})
		return err
	}); err != nil {
		return nil, fmt.Errorf("回滚statefulset失败:%s", err.Error())
	}
	return result, nil
}

// GetStatefulSetRollbacks returns all ControllerRevision entries for a StatefulSet,
// sorted by revision descending, for UI rollback selection.
func GetStatefulSetRollbacks(client *kubernetes.Clientset, namespace, name string) ([]map[string]any, error) {
	ctx := context.Background()
	sts, err := client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	revisions, err := client.AppsV1().ControllerRevisions(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(sts.Spec.Selector),
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
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].Revision > entries[i].Revision {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
	result := make([]map[string]any, len(entries))
	for i, e := range entries {
		result[i] = map[string]any{
			"revision": e.Revision,
			"name":     e.Name,
		}
	}
	return result, nil
}

// GetStatefulSetPVs returns PVCs created by the StatefulSet's volumeClaimTemplates.
//
// K8s does NOT label these PVCs with the StatefulSet name; it names them
// `<volumeClaimTemplate.name>-<sts.name>-<ordinal>`. So we list all PVCs in the
// namespace and filter by that naming convention for each template.
func GetStatefulSetPVs(client *kubernetes.Clientset, namespace, name string) (*corev1.PersistentVolumeClaimList, error) {
	sts, err := client.AppsV1().StatefulSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取StatefulSet失败:%s", err.Error())
	}
	allPvcs, err := client.CoreV1().PersistentVolumeClaims(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取PVC列表失败:%s", err.Error())
	}

	// Build prefixes: <vctName>-<stsName>- for each volumeClaimTemplate
	prefixes := make([]string, 0, len(sts.Spec.VolumeClaimTemplates))
	for _, vct := range sts.Spec.VolumeClaimTemplates {
		prefixes = append(prefixes, vct.Name+"-"+name+"-")
	}

	filtered := &corev1.PersistentVolumeClaimList{}
	for _, pvc := range allPvcs.Items {
		pvcName := pvc.Name
		for _, p := range prefixes {
			if strings.HasPrefix(pvcName, p) {
				filtered.Items = append(filtered.Items, pvc)
				break
			}
		}
	}
	return filtered, nil
}
