package job

import (
	"context"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// GetJobList
//
//	@Description: 获取job列表
//	@param client
//	@param namespace
//	@return []batchv1.Job
//	@return error
func GetJobList(client *kubernetes.Clientset, namespace string) ([]batchv1.Job, error) {
	jobList, err := client.BatchV1().Jobs(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return jobList.Items, nil
}

// GetJobByName
//
//	@Description: 根据名称获取job
//	@param client
//	@param namespace
//	@param name
//	@return *batchv1.Job
//	@return error
func GetJobByName(client *kubernetes.Clientset, namespace string, name string) (*batchv1.Job, error) {
	job, err := client.BatchV1().Jobs(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return job, nil
}

// GetJobByFiled
//
//	@Description: 根据字段获取job
//	@param client
//	@param namespace
//	@param fieldMap
//	@return []batchv1.Job
//	@return error
func GetJobByFiled(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) ([]batchv1.Job, error) {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	jobList, err := client.BatchV1().Jobs(namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return nil, err
	}
	return jobList.Items, nil
}

// GetJobByLabel
//
//	@Description: 根据标签获取job
//	@param client
//	@param namespace
//	@param labelMap
//	@return []batchv1.Job
//	@return error
func GetJobByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) ([]batchv1.Job, error) {
	labelSelector := labels.SelectorFromSet(labelMap)
	jobList, err := client.BatchV1().Jobs(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return nil, err
	}
	return jobList.Items, nil
}

// GetJobYaml
//
//	@Description: 获取job的yaml文件
//	@param client
//	@param namespace
//	@param name
//	@return string
//	@return error
func GetJobYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	job, err := client.BatchV1().Jobs(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return job.String(), nil
}

// CreateJob
//
//	@Description: 创建job
//	@param client
//	@param namespace
//	@param jobYaml
//	@return error
func CreateJob(client *kubernetes.Clientset, jobYaml string) error {
	var job batchv1.Job
	if err := yaml.Unmarshal([]byte(jobYaml), &job); err != nil {
		return fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	_, err := client.BatchV1().Jobs(job.Namespace).Create(context.Background(), &job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("创建job资源失败:%s", err.Error())
	}
	return nil
}

// UpdateJob
//
//	@Description: 更新job
//	@param client
//	@param jobYaml
//	@return error
func UpdateJob(client *kubernetes.Clientset, jobYaml string) error {
	var job batchv1.Job
	if err := yaml.Unmarshal([]byte(jobYaml), &job); err != nil {
		return fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	_, err := client.BatchV1().Jobs(job.Namespace).Update(context.Background(), &job, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("更新失败:%s", err.Error())
	}
	return nil
}

// DeleteJob
//
//	@Description: 删除job
//	@param client
//	@param namespace
//	@param name
//	@return error
func DeleteJob(client *kubernetes.Clientset, namespace, name string) error {
	err := client.BatchV1().Jobs(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("删除job资源失败:%s", err.Error())
	}
	return nil
}

// DeleteJobByField
//
//	@Description: 根据字段删除job
//	@param client
//	@param namespace
//	@param fieldMap
//	@return error
func DeleteJobByField(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) error {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	err := client.BatchV1().Jobs(namespace).DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return fmt.Errorf("删除job资源失败:%s", err.Error())
	}
	return nil
}

// DeleteJobByLabel
//
//	@Description: 根据标签删除job
//	@param client
//	@param namespace
//	@param labelMap
//	@return error
func DeleteJobByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) error {
	labelSelector := labels.SelectorFromSet(labelMap)
	err := client.BatchV1().Jobs(namespace).DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return fmt.Errorf("删除job资源失败:%s", err.Error())
	}
	return nil
}
