package statefulset

import (
	"encoding/json"
	"gkube/pkg/yamlutil"
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
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
	labelSelector := fields.SelectorFromSet(labelMap)
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
	_, err := client.AppsV1().StatefulSets(namespace).Create(context.Background(), &statefulSet, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("创建statefulSet资源失败:%s", err.Error())
	}
	return nil
}

// UpdateStatefulSet
//
//	@Description: 更新statefulSet
//	@param client
//	@param namespace
//	@param statefulSetYaml
//	@return bool
//	@return error
func UpdateStatefulSet(client *kubernetes.Clientset, namespace, statefulSetYaml string) error {
	var statefulSet appsv1.StatefulSet
	if err := yaml.Unmarshal([]byte(statefulSetYaml), &statefulSet); err != nil {
		return fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	_, err := client.AppsV1().StatefulSets(namespace).Update(context.Background(), &statefulSet, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("更新statefulSet资源失败:%s", err.Error())
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
	labelSelector := fields.SelectorFromSet(labelMap)
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
	statefulSet, err := client.AppsV1().StatefulSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("获取statefulSet资源失败:%s", err.Error())
	}
	statefulSet.Spec.Replicas = &replicas
	_, err = client.AppsV1().StatefulSets(namespace).Update(context.TODO(), statefulSet, metav1.UpdateOptions{})
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
	statefulSet, err := client.AppsV1().StatefulSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("获取statefulSet资源失败:%s", err.Error())
	}
	statefulSet.Spec.Template.Annotations = map[string]string{
		"kubectl.kubernetes.io/restartedAt": time.Now().Format(time.DateTime),
	}
	_, err = client.AppsV1().StatefulSets(namespace).Update(context.TODO(), statefulSet, metav1.UpdateOptions{})
	if err != nil {
		return false, fmt.Errorf("更新statefulSet资源失败:%s", err.Error())
	}
	return true, nil
}

func UpdateStatefulSetImage(client *kubernetes.Clientset, namespace, name, containerName, image string) (*appsv1.StatefulSet, error) {
	ctx := context.TODO()
	sts, err := client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	for i, c := range sts.Spec.Template.Spec.Containers {
		if c.Name == containerName {
			sts.Spec.Template.Spec.Containers[i].Image = image
			return client.AppsV1().StatefulSets(namespace).Update(ctx, sts, metav1.UpdateOptions{})
		}
	}
	return nil, fmt.Errorf("container %s not found", containerName)
}

func RollbackStatefulSet(client *kubernetes.Clientset, namespace, name string, revision int64) (*appsv1.StatefulSet, error) {
	ctx := context.TODO()
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
	for _, rev := range revisions.Items {
		if rev.Revision == revision {
			var restored appsv1.StatefulSet
			if err := json.Unmarshal(rev.Data.Raw, &restored); err != nil {
				continue
			}
			restored.ResourceVersion = sts.ResourceVersion
			restored.UID = sts.UID
			return client.AppsV1().StatefulSets(namespace).Update(ctx, &restored, metav1.UpdateOptions{})
		}
	}
	return nil, fmt.Errorf("revision %d not found", revision)
}
