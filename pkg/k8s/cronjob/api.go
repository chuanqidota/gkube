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

// GetCronJob
//
//	@Description: 获取cronjob
//	@param client
//	@param namespace
//	@param name
//	@return *batchv1.CronJob
//	@return error
func GetCronJob(client *kubernetes.Clientset, namespace, name string) (*batchv1.CronJob, error) {
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
	return cronJob.String(), nil
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

// DeleteCronJob
//
//	@Description: 删除cronjob
//	@param client
//	@param namespace
//	@param name
//	@return bool
//	@return error
func DeleteCronJob(client *kubernetes.Clientset, namespace, name string) error {
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
