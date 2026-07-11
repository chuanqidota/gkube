package deployment

import (
	"gkube/pkg/yamlutil"
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// GetDeploymentList
//
//	@Description: 获取deployment列表
//	@param client
//	@param namespace
//	@return []appsv1.Deployment
//	@return error
func GetDeploymentList(client *kubernetes.Clientset, namespace string) ([]appsv1.Deployment, error) {
	deploymentList, err := client.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return deploymentList.Items, nil
}

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

// GetDeploymentByFiled
//
//	@Description: 根据字段获取deployment列表
//	@param client
//	@param namespace
//	@param fieldMap
//	@return []appsv1.Deployment
//	@return error
func GetDeploymentByFiled(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) ([]appsv1.Deployment, error) {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	deploymentList, err := client.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return nil, err
	}
	return deploymentList.Items, nil
}

// GetDeploymentByLabel
//
//	@Description: 根据标签获取deployment列表
//	@param client
//	@param namespace
//	@param labelMap
//	@return []appsv1.Deployment
//	@return error
func GetDeploymentByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) ([]appsv1.Deployment, error) {
	labelSelector := fields.SelectorFromSet(labelMap)
	deploymentList, err := client.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return nil, err
	}
	return deploymentList.Items, nil
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
//	@param cronJobYaml
//	@return bool
//	@return error
func CreateDeployment(client *kubernetes.Clientset, namespace string, cronJobYaml string) (bool, error) {
	var deployment *appsv1.Deployment
	if err := yaml.Unmarshal([]byte(cronJobYaml), &deployment); err != nil {
		return false, fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	_, err := client.AppsV1().Deployments(namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		return false, fmt.Errorf("创建deployment资源失败:%s", err.Error())
	}
	return true, nil
}

// UpdateDeployment
//
//	@Description: 更新deployment
//	@param client
//	@param namespace
//	@param cronJobYaml
//	@return bool
//	@return error
func UpdateDeployment(client *kubernetes.Clientset, namespace, cronJobYaml string) (bool, error) {
	var deployment *appsv1.Deployment
	if err := yaml.Unmarshal([]byte(cronJobYaml), &deployment); err != nil {
		return false, fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	_, err := client.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return false, fmt.Errorf("更新deployment资源失败:%s", err.Error())
	}
	return true, nil
}

// DeleteDeployment
//
//	@Description: 删除deployment
//	@param client
//	@param namespace
//	@param name
//	@return bool
//	@return error
func DeleteDeployment(client *kubernetes.Clientset, namespace, name string) (bool, error) {
	err := client.AppsV1().Deployments(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return false, fmt.Errorf("删除deployment资源失败:%s", err.Error())
	}
	return true, nil
}

// DeleteDeploymentByField
//
//	@Description: 根据字段删除deployment
//	@param client
//	@param namespace
//	@param fieldMap
//	@return bool
//	@return error
func DeleteDeploymentByField(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) (bool, error) {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	err := client.AppsV1().Deployments(namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return false, fmt.Errorf("删除deployment资源失败:%s", err.Error())
	}
	return true, nil
}

// DeleteDeploymentByLabel
//
//	@Description: 根据标签删除deployment
//	@param client
//	@param namespace
//	@param labelMap
//	@return bool
//	@return error
func DeleteDeploymentByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) (bool, error) {
	labelSelector := fields.SelectorFromSet(labelMap)
	err := client.AppsV1().Deployments(namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return false, fmt.Errorf("删除deployment资源失败:%s", err.Error())
	}
	return true, nil
}

// ScaleDeployment
//
//	@Description: 扩所容deployment
//	@param client
//	@param namespace
//	@param name
//	@param replicas
//	@return bool
//	@return error
func ScaleDeployment(client *kubernetes.Clientset, namespace, name string, replicas int32) (bool, error) {
	deployment, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("获取deployment资源失败:%s", err.Error())
	}
	deployment.Spec.Replicas = &replicas
	_, err = client.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return false, fmt.Errorf("更新deployment资源失败:%s", err.Error())
	}
	return true, nil
}

// RestartDeployment
//
//	@Description: 重启deployment
//	@param client
//	@param namespace
//	@param name
//	@return bool
//	@return error
func RestartDeployment(client *kubernetes.Clientset, namespace, name string) (bool, error) {
	deployment, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("获取deployment资源失败:%s", err.Error())
	}
	deployment.Spec.Template.Annotations = map[string]string{
		"kubectl.kubernetes.io/restartedAt": time.Now().Format(time.DateTime),
	}
	_, err = client.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return false, fmt.Errorf("更新deployment资源失败:%s", err.Error())
	}
	return true, nil
}

// UpdateDeploymentImage
//
//	@Description: 更新deployment的容器镜像
//	@param client
//	@param namespace
//	@param name
//	@param containerName
//	@param image
//	@return bool
//	@return error
func UpdateDeploymentImage(client *kubernetes.Clientset, namespace, name, containerName, image string) (bool, error) {
	deployment, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("获取deployment资源失败:%s", err.Error())
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
		return false, fmt.Errorf("容器 %s 不存在", containerName)
	}
	_, err = client.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return false, fmt.Errorf("更新deployment镜像失败:%s", err.Error())
	}
	return true, nil
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
//	@return bool
//	@return error
func RollbackDeployment(client *kubernetes.Clientset, namespace, name string, revision int64) (bool, error) {
	// Get ReplicaSets for this deployment
	deployment, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("获取deployment资源失败:%s", err.Error())
	}

	labelSelector := labels.Set(deployment.Spec.Selector.MatchLabels).String()
	rsList, err := client.AppsV1().ReplicaSets(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return false, fmt.Errorf("获取ReplicaSet列表失败:%s", err.Error())
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
		return false, fmt.Errorf("未找到 revision %d 对应的ReplicaSet", revision)
	}

	// Apply the template from the target ReplicaSet
	deployment.Spec.Template = targetRS.Spec.Template
	_, err = client.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return false, fmt.Errorf("回滚deployment失败:%s", err.Error())
	}
	return true, nil
}

// DpPodList
//
//	@Description: deployment关联的pod
//	@param client
//	@param namespace
//	@param name
//	@return *corev1.PodList
//	@return error
func DpPodList(client *kubernetes.Clientset, namespace, name string) (*corev1.PodList, error) {
	deployment, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取deployment资源失败:%s", err.Error())
	}
	selector := labels.Set(deployment.Spec.Selector.MatchLabels).AsSelectorPreValidated()
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
	// Get the deployment to find its label selector
	deploy, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment %s/%s: %v", namespace, name, err)
	}

	// Build label selector from deployment's spec.selector
	selector := labels.SelectorFromSet(deploy.Spec.Selector.MatchLabels)
	listOpts := metav1.ListOptions{
		LabelSelector: selector.String(),
	}

	// List ReplicaSets
	rsList, err := client.AppsV1().ReplicaSets(namespace).List(context.TODO(), listOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list replicasets for deployment %s/%s: %v", namespace, name, err)
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
