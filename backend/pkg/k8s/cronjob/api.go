package cronjob

import (
	"context"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// CronJobJobsList
//
//	@Description: 获取cronjob关联的job列表
//	@param client
//	@param namespace
//	@param name
//	@return []batchv1.Job
//	@return error
func CronJobJobsList(client *kubernetes.Clientset, namespace, name string) ([]batchv1.Job, error) {
	cronJob, err := client.BatchV1().CronJobs(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取cronjob资源失败:%s", err.Error())
	}
	// CronJob doesn't have a selector like Deployment/StatefulSet
	// We need to find jobs owned by this CronJob using ownerReferences
	jobList, err := client.BatchV1().Jobs(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取job列表失败:%s", err.Error())
	}
	var ownedJobs []batchv1.Job
	for _, job := range jobList.Items {
		for _, ref := range job.OwnerReferences {
			if ref.Kind == "CronJob" && ref.Name == cronJob.Name {
				ownedJobs = append(ownedJobs, job)
				break
			}
		}
	}
	return ownedJobs, nil
}

// GetCronJobList
//
//	@Description: 获取cronjob列表
//	@param client
//	@param namespace
//	@return []batchv1.CronJob
//	@return error
func GetCronJobList(client *kubernetes.Clientset, namespace string) ([]batchv1.CronJob, error) {
	cronJobList, err := client.BatchV1().CronJobs(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return cronJobList.Items, nil
}

// ListCronJobs returns a paginated cronjob list with metadata
func ListCronJobs(client *kubernetes.Clientset, namespace string, limit int64, continueToken string) (*batchv1.CronJobList, error) {
	listOpts := metav1.ListOptions{}
	if limit > 0 {
		listOpts.Limit = limit
	}
	if continueToken != "" {
		listOpts.Continue = continueToken
	}
	return client.BatchV1().CronJobs(namespace).List(context.TODO(), listOpts)
}

// GetCronJobByName
//
//	@Description: 获取cronjob
//	@param client
//	@param namespace
//	@param name
//	@return *batchv1.CronJob
//	@return error
func GetCronJobByName(client *kubernetes.Clientset, namespace, name string) (*batchv1.CronJob, error) {
	cronJob, err := client.BatchV1().CronJobs(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return cronJob, nil
}

// GetCronJobYaml
//
//	@Description: 获取cronjob的yaml
//	@param client
//	@param namespace
//	@param name
//	@return string
//	@return error
func GetCronJobYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	cronJob, err := client.BatchV1().CronJobs(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	yamlBytes, err := yaml.Marshal(cronJob)
	if err != nil {
		return "", err
	}
	return string(yamlBytes), nil
}

// GetCronJobByLabel
//
//	@Description: 根据label获取cronjob列表
//	@param client
//	@param namespace
//	@param labelMap
//	@return []batchv1.CronJob
//	@return error
func GetCronJobByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) ([]batchv1.CronJob, error) {
	labelSelector := labels.SelectorFromSet(labelMap)
	cronJobList, err := client.BatchV1().CronJobs(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return nil, err
	}
	return cronJobList.Items, nil
}

// GetCronJobByField
//
//	@Description: 根据field获取cronjob列表
//	@param client
//	@param namespace
//	@param fieldMap
//	@return []batchv1.CronJob
//	@return error
func GetCronJobByField(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) ([]batchv1.CronJob, error) {
	fieldSelector := labels.SelectorFromSet(fieldMap)
	cronJobList, err := client.BatchV1().CronJobs(namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return nil, err
	}
	return cronJobList.Items, nil
}

// CreateCronJob
//
//	@Description: 创建cronjob
//	@param client
//	@param namespace
//	@param cronJobYaml
//	@return bool
//	@return error
func CreateCronJob(client *kubernetes.Clientset, namespace, cronJobYaml string) error {
	var cronJob batchv1.CronJob
	if err := yaml.Unmarshal([]byte(cronJobYaml), &cronJob); err != nil {
		return fmt.Errorf("yaml文件错误:%s", err.Error())
	}

	_, err := client.BatchV1().CronJobs(namespace).Create(context.TODO(), &cronJob, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("创建cronjob资源失败:%s", err.Error())
	}
	return nil
}

// UpdateCronJob
//
//	@Description: 更新cronjob
//	@param client
//	@param namespace
//	@param cronJobYaml
//	@return bool
//	@return error
func UpdateCronJob(client *kubernetes.Clientset, namespace, cronJobYaml string) error {
	var cronJob batchv1.CronJob
	if err := yaml.Unmarshal([]byte(cronJobYaml), &cronJob); err != nil {
		return fmt.Errorf("yaml文件错误:%s", err.Error())
	}

	_, err := client.BatchV1().CronJobs(namespace).Update(context.TODO(), &cronJob, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("更新cronjob资源失败:%s", err.Error())
	}
	return nil
}

// DeleteCronJobByName
//
//	@Description: 删除cronjob
//	@param client
//	@param namespace
//	@param name
//	@return bool
//	@return error
func DeleteCronJobByName(client *kubernetes.Clientset, namespace, name string) error {
	err := client.BatchV1().CronJobs(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("删除cronjob资源失败:%s", err.Error())
	}
	return nil
}

// DeleteCronJobByField
//
//	@Description: 删除根据field删除cronjob
//	@param client
//	@param namespace
//	@param fieldMap
//	@return bool
//	@return error
func DeleteCronJobByField(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) error {
	fieldSelector := labels.SelectorFromSet(fieldMap)
	err := client.BatchV1().CronJobs(namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return fmt.Errorf("删除cronjob资源失败:%s", err.Error())
	}
	return nil
}

// DeleteCronJobByLabel
//
//	@Description: 删除根据label删除cronjob
//	@param client
//	@param namespace
//	@param labelMap
//	@return bool
//	@return error
func DeleteCronJobByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) error {
	labelSelector := labels.SelectorFromSet(labelMap)
	err := client.BatchV1().CronJobs(namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return fmt.Errorf("删除cronjob资源失败:%s", err.Error())
	}
	return nil
}
