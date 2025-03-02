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

func GetJobList(client *kubernetes.Clientset, namespace string) ([]batchv1.Job, error) {
	jobList, err := client.BatchV1().Jobs(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return jobList.Items, nil
}

func GetJobByName(client *kubernetes.Clientset, namespace string, name string) (*batchv1.Job, error) {
	job, err := client.BatchV1().Jobs(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return job, nil
}

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

func GetJobYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	job, err := client.BatchV1().Jobs(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return job.String(), nil
}

func CreateJob(client *kubernetes.Clientset, jobYaml string) (bool, error) {
	var job batchv1.Job
	if err := yaml.Unmarshal([]byte(jobYaml), &job); err != nil {
		return false, fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	_, err := client.BatchV1().Jobs(job.Namespace).Create(context.Background(), &job, metav1.CreateOptions{})
	if err != nil {
		return false, fmt.Errorf("创建job资源失败:%s", err.Error())
	}
	return true, nil
}

func UpdateJob(client *kubernetes.Clientset, jobYaml string) (bool, error) {
	var job batchv1.Job
	if err := yaml.Unmarshal([]byte(jobYaml), &job); err != nil {
		return false, fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	_, err := client.BatchV1().Jobs(job.Namespace).Update(context.Background(), &job, metav1.UpdateOptions{})
	if err != nil {
		return false, fmt.Errorf("更新失败:%s", err.Error())
	}
	return true, nil
}

func DeleteJob(client *kubernetes.Clientset, namespace, name string) (bool, error) {
	err := client.BatchV1().Jobs(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return false, fmt.Errorf("删除job资源失败:%s", err.Error())
	}
	return true, nil
}

func DeleteJobByField(client *kubernetes.Clientset, namespace string, fieldMap map[string]string) (bool, error) {
	fieldSelector := fields.SelectorFromSet(fieldMap)
	err := client.BatchV1().Jobs(namespace).DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
	if err != nil {
		return false, fmt.Errorf("删除job资源失败:%s", err.Error())
	}
	return true, nil
}

func DeleteJobByLabel(client *kubernetes.Clientset, namespace string, labelMap map[string]string) (bool, error) {
	labelSelector := labels.SelectorFromSet(labelMap)
	err := client.BatchV1().Jobs(namespace).DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return false, fmt.Errorf("删除job资源失败:%s", err.Error())
	}
	return true, nil
}
