package deployment

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/yaml"

	k8sEvent "gkube/pkg/k8s/event"
	"gkube/pkg/yamlutil"
)

// ListDeployments returns a paginated deployment list with metadata
func ListDeployments(client *kubernetes.Clientset, namespace string, limit int64, continueToken string) (*appsv1.DeploymentList, error) {
	listOpts := metav1.ListOptions{}
	if limit > 0 {
		listOpts.Limit = limit
	}
	if continueToken != "" {
		listOpts.Continue = continueToken
	}
	return client.AppsV1().Deployments(namespace).List(context.TODO(), listOpts)
}

// GetDeploymentYaml
//
//	@Description: 获取deployment yaml
//	@param client
//	@param namespace
//	@param name
//	@return string
//	@return error
func GetDeploymentYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	deployment, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	yamlStr, err := yamlutil.MarshalWithoutManagedFields(deployment)
	if err != nil {
		return "", err
	}
	return yamlStr, nil
}

// CreateDeployment
//
//	@Description: 创建deployment
//	@param client
//	@param namespace
//	@param deploymentYaml
//	@return error
func CreateDeployment(client *kubernetes.Clientset, namespace string, deploymentYaml string) error {
	var deployment *appsv1.Deployment
	if err := yaml.Unmarshal([]byte(deploymentYaml), &deployment); err != nil {
		return fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	// 以 handler 传入的 namespace 为准，避免 YAML 内 metadata.namespace 与之不符时静默落到别处
	deployment.Namespace = namespace
	_, err := client.AppsV1().Deployments(namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("创建deployment资源失败:%s", err.Error())
	}
	return nil
}

// UpdateDeployment
//
//	@Description: 更新deployment
//	@param client
//	@param namespace
//	@param name
//	@param deploymentYaml
//	@return error
func UpdateDeployment(client *kubernetes.Clientset, namespace, name, deploymentYaml string) error {
	var deployment *appsv1.Deployment
	if err := yaml.Unmarshal([]byte(deploymentYaml), &deployment); err != nil {
		return fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	// 校验 YAML 中的名称与请求指定的一致，避免误更新同名空间下的其他资源
	if deployment.Name != name {
		return fmt.Errorf("资源名称不匹配: 请求指定 %s, YAML 中为 %s", name, deployment.Name)
	}
	deployment.Namespace = namespace
	// 冲突时自动重试(参照 pod UpdatePod / deployment restart/scale 的 RetryOnConflict 模式)。
	// 闭包内 re-Get 最新对象(带新 resourceVersion),再用用户 YAML 的 spec 覆盖后 Update,
	// 否则重试会发同一个过期 resourceVersion 持续 409 直到 backoff 耗尽(死重试)。
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("获取deployment资源失败:%s", err.Error())
		}
		// 用用户 YAML 的 spec 与可变 metadata 覆盖最新对象,保留最新 resourceVersion
		latest.Spec = deployment.Spec
		latest.Labels = deployment.Labels
		latest.Annotations = deployment.Annotations
		_, err = client.AppsV1().Deployments(namespace).Update(context.TODO(), latest, metav1.UpdateOptions{})
		return err
	}); err != nil {
		return fmt.Errorf("更新deployment资源失败:%s", err.Error())
	}
	return nil
}

// DeleteDeployment
//
//	@Description: 删除deployment
//	@param client
//	@param namespace
//	@param name
//	@return error
func DeleteDeployment(client *kubernetes.Clientset, namespace, name string) error {
	err := client.AppsV1().Deployments(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("删除deployment资源失败:%s", err.Error())
	}
	return nil
}

// ScaleDeployment
//
//	@Description: 扩所容deployment
//	@param client
//	@param namespace
//	@param name
//	@param replicas
//	@return error
func ScaleDeployment(client *kubernetes.Clientset, namespace, name string, replicas *int32) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		deployment, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("获取deployment资源失败:%s", err.Error())
		}
		deployment.Spec.Replicas = replicas
		_, err = client.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
		// 返回原始 err，以便 RetryOnConflict 识别 409 Conflict 自动重试
		return err
	})
}

// RestartDeployment
//
//	@Description: 重启deployment
//	@param client
//	@param namespace
//	@param name
//	@return error
func RestartDeployment(client *kubernetes.Clientset, namespace, name string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		deployment, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("获取deployment资源失败:%s", err.Error())
		}
		// 合并而非覆盖：保留 pod 模板上已有的运维注解
		if deployment.Spec.Template.Annotations == nil {
			deployment.Spec.Template.Annotations = map[string]string{}
		}
		deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.DateTime)
		_, err = client.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
		return err
	})
}

// UpdateDeploymentImage
//
//	@Description: 更新deployment的容器镜像
//	@param client
//	@param namespace
//	@param name
//	@param containerName
//	@param image
//	@return error
func UpdateDeploymentImage(client *kubernetes.Clientset, namespace, name, containerName, image string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		deployment, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("获取deployment资源失败:%s", err.Error())
		}
		found := false
		for i, container := range deployment.Spec.Template.Spec.Containers {
			if container.Name == containerName {
				deployment.Spec.Template.Spec.Containers[i].Image = image
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("容器 %s 不存在", containerName)
		}
		_, err = client.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
		return err
	})
}

// GetDeploymentDetail
//
//	@Description: 获取deployment详情
//	@param client
//	@param namespace
//	@param name
//	@return *appsv1.Deployment
//	@return error
func GetDeploymentDetail(client *kubernetes.Clientset, namespace, name string) (*appsv1.Deployment, error) {
	deployment, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取deployment详情失败:%s", err.Error())
	}
	return deployment, nil
}

// RollbackDeployment
//
//	@Description: 回滚deployment到指定revision
//	@param client
//	@param namespace
//	@param name
//	@param revision
//	@return error
func RollbackDeployment(client *kubernetes.Clientset, namespace, name string, revision int64) error {
	// 先取 deployment 的 selector 用于查找 ReplicaSet（不参与冲突重试）
	deployment, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取deployment资源失败:%s", err.Error())
	}

	// 用完整 selector(matchLabels + matchExpressions) 查找 RS,nil/空 selector 时中止避免误列
	if deployment.Spec.Selector == nil {
		return fmt.Errorf("deployment selector 为空,无法定位关联 ReplicaSet")
	}
	selector, err := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
	if err != nil {
		return fmt.Errorf("解析deployment selector失败:%s", err.Error())
	}
	if selector.Empty() {
		return fmt.Errorf("deployment selector 为空,无法定位关联 ReplicaSet")
	}
	rsList, err := client.AppsV1().ReplicaSets(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return fmt.Errorf("获取ReplicaSet列表失败:%s", err.Error())
	}

	// Find the target ReplicaSet with matching revision
	var targetRS *appsv1.ReplicaSet
	for i, rs := range rsList.Items {
		revStr := rs.Annotations["deployment.kubernetes.io/revision"]
		if revStr == fmt.Sprintf("%d", revision) {
			targetRS = &rsList.Items[i]
			break
		}
	}

	if targetRS == nil {
		return fmt.Errorf("未找到 revision %d 对应的ReplicaSet", revision)
	}

	// 用目标 RS 的模板覆盖 deployment 模板，冲突时自动重试
	newTemplate := targetRS.Spec.Template
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		d, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("获取deployment资源失败:%s", err.Error())
		}
		d.Spec.Template = newTemplate
		_, err = client.AppsV1().Deployments(namespace).Update(context.TODO(), d, metav1.UpdateOptions{})
		return err
	})
	if err != nil {
		return fmt.Errorf("回滚deployment失败:%s", err.Error())
	}
	return nil
}

// GetDeploymentPods
//
//	@Description: 获取deployment关联的pod
//	@param client
//	@param namespace
//	@param name
//	@return *corev1.PodList
//	@return error
func GetDeploymentPods(client *kubernetes.Clientset, namespace, name string) (*corev1.PodList, error) {
	deployment, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取deployment资源失败:%s", err.Error())
	}
	// 用完整 selector(matchLabels + matchExpressions),nil/空 selector 时返回空列表避免误列全部 Pod
	if deployment.Spec.Selector == nil {
		return &corev1.PodList{Items: []corev1.Pod{}}, nil
	}
	selector, err := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
	if err != nil {
		return nil, fmt.Errorf("解析deployment selector失败:%s", err.Error())
	}
	if selector.Empty() {
		return &corev1.PodList{Items: []corev1.Pod{}}, nil
	}
	podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("获取pod资源失败:%s", err.Error())
	}
	return podList, nil
}

// GetDeploymentReplicaSets returns all ReplicaSets owned by a Deployment
func GetDeploymentReplicaSets(client *kubernetes.Clientset, namespace, name string) ([]appsv1.ReplicaSet, error) {
	deploy, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取deployment资源失败:%s", err.Error())
	}

	// 用完整 selector(matchLabels + matchExpressions),nil/空 selector 时返回空列表避免误列全部 RS
	if deploy.Spec.Selector == nil {
		return []appsv1.ReplicaSet{}, nil
	}
	selector, err := metav1.LabelSelectorAsSelector(deploy.Spec.Selector)
	if err != nil {
		return nil, fmt.Errorf("解析deployment selector失败:%s", err.Error())
	}
	if selector.Empty() {
		return []appsv1.ReplicaSet{}, nil
	}
	rsList, err := client.AppsV1().ReplicaSets(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("获取ReplicaSet列表失败:%s", err.Error())
	}

	// Sort by revision annotation (descending)
	sort.Slice(rsList.Items, func(i, j int) bool {
		revI := getRevision(&rsList.Items[i])
		revJ := getRevision(&rsList.Items[j])
		return revI > revJ
	})

	return rsList.Items, nil
}

// getRevision extracts the revision number from ReplicaSet annotations
func getRevision(rs *appsv1.ReplicaSet) int64 {
	if rs.Annotations == nil {
		return 0
	}
	revStr := rs.Annotations["deployment.kubernetes.io/revision"]
	rev, _ := strconv.ParseInt(revStr, 10, 64)
	return rev
}

// GetDeploymentEvents returns the events associated with a Deployment.
// 复用 event 包返回结构化 KubeEvent(与 Pod events 形态一致),用 fields.Selector 防注入。
func GetDeploymentEvents(client *kubernetes.Clientset, namespace, name string) ([]k8sEvent.KubeEvent, error) {
	selector := fields.AndSelectors(
		fields.OneTermEqualSelector("involvedObject.name", name),
		fields.OneTermEqualSelector("involvedObject.kind", "Deployment"),
	).String()
	events, _, _, err := k8sEvent.ListEvents(client, namespace, selector, 0, "")
	if err != nil {
		return nil, fmt.Errorf("获取deployment事件失败:%s", err.Error())
	}
	return events, nil
}
