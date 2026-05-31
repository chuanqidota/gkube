package pod

import (
	"context"
	"errors"
	"sort"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// GetPodList
//
//	@Description: 获取pod列表
//	@param client
//	@param namespace
//	@return []corev1.Pod
//	@return error
func GetPodList(client *kubernetes.Clientset, namespace string) ([]corev1.Pod, error) {
	podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
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
	return pod.String(), nil
}

// GetPodByField
//
//	@Description: 通过字段查询pod
//	@param client
//	@param namespace
//	@param fieldMap
//	@return []corev1.Pod
//	@return error
func GetPodByField(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) ([]corev1.Pod, error) {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}

// GetPodByLabel
//
//	@Description: 通过标签查询pod
//	@param client
//	@param namespace
//	@param labelMap
//	@return []corev1.Pod
//	@return error
func GetPodByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) ([]corev1.Pod, error) {
	labelSelector := labels.SelectorFromSet(labelMap)
	podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}

// CreatePod
//
//	@Description: 创建pod
//	@param client
//	@param namespace
//	@param podYaml
//	@return error
func CreatePod(client *kubernetes.Clientset, podYaml string) error {
	pod := &corev1.Pod{}
	err := yaml.Unmarshal([]byte(podYaml), pod)
	if err != nil {
		return err
	}
	_, err = client.CoreV1().Pods(pod.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// UpdatePod
//
//	@Description: 更新pod
//	@param client
//	@param podYaml
//	@return error
func UpdatePod(client *kubernetes.Clientset, podYaml string) error {
	pod := &corev1.Pod{}
	err := yaml.Unmarshal([]byte(podYaml), pod)
	if err != nil {
		return err
	}
	_, err = client.CoreV1().Pods(pod.Namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// DeletePodByName
//
//	@Description: 删除pod根据名称
//	@param client
//	@param namespace
//	@param name
//	@return error
func DeletePodByName(client *kubernetes.Clientset, namespace, name string) error {
	err := client.CoreV1().Pods(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

// DeletePodByField
//
//	@Description: 删除pod根据字段
//	@param client
//	@param namespace
//	@param fieldMap
//	@return error
func DeletePodByField(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) error {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	err := client.CoreV1().Pods(namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return err
	}
	return nil
}

// DeletePodByLabel
//
//	@Description: 删除pod根据标签
//	@param client
//	@param namespace
//	@param labelMap
//	@return error
func DeletePodByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) error {
	labelSelector := labels.SelectorFromSet(labelMap)
	err := client.CoreV1().Pods(namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return err
	}
	return nil
}

// GetLatestPodList 获取最新的pod列表

func GetLatestPodList(client *kubernetes.Clientset, namespace, deployment string) ([]corev1.Pod, error) {
	deploy, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), deployment, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	// 获取该 Deployment 关联的所有 ReplicaSet
	rsList, err := client.AppsV1().ReplicaSets(namespace).List(
		context.TODO(), metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(deploy.Spec.Selector),
		},
	)
	if err != nil {
		return nil, err
	}
	// 过滤出属于该 Deployment 的 ReplicaSet
	var ownedRS []appsv1.ReplicaSet
	for _, rs := range rsList.Items {
		for _, ownerRef := range rs.OwnerReferences {
			if ownerRef.UID == deploy.UID {
				ownedRS = append(ownedRS, rs)
				break
			}
		}
	}
	if len(ownedRS) == 0 {
		return nil, errors.New("未找到属于该 Deployment 的 ReplicaSet")
	}

	getRevision := func(rs appsv1.ReplicaSet) int64 {
		revision := rs.Annotations["deployment.kubernetes.io/revision"]
		if revision == "" {
			return 0
		}
		revisionInt, _ := strconv.Atoi(revision)
		return int64(revisionInt)
	}

	// 按 Revision 排序，获取最新的 ReplicaSet
	sort.Slice(ownedRS, func(i, j int) bool {
		return getRevision(ownedRS[i]) > getRevision(ownedRS[j])
	})
	latestRS := ownedRS[0]

	// 获取该 ReplicaSet 管理的 Pod 的 IP 地址
	podList, err := client.CoreV1().Pods(namespace).List(
		context.TODO(), metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(latestRS.Spec.Selector),
		},
	)
	return podList.Items, err
}
