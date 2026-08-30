package service

import (
	"context"
	"fmt"

	"gkube/pkg/yamlutil"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/yaml"
)

// GetServicesList
//
//	@Description: 获取service列表
//	@param client
//	@param namespace
//	@return []corev1.Service
//	@return error
func GetServicesList(client *kubernetes.Clientset, namespace string) ([]corev1.Service, error) {
	services, err := client.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		return nil, err
	}
	return services.Items, nil
}

// ListServices returns a paginated service list with metadata
func ListServices(client *kubernetes.Clientset, namespace string, limit int64, continueToken string) (*corev1.ServiceList, error) {
	listOpts := metav1.ListOptions{ResourceVersion: "0"}
	if limit > 0 {
		listOpts.Limit = limit
	}
	if continueToken != "" {
		listOpts.Continue = continueToken
	}
	return client.CoreV1().Services(namespace).List(context.TODO(), listOpts)
}

// GetServicesByName
//
//	@Description: 根据名称获取service
//	@param client
//	@param namespace
//	@param name
//	@return *corev1.Service
//	@return error
func GetServicesByName(client *kubernetes.Clientset, namespace, name string) (*corev1.Service, error) {
	service, err := client.CoreV1().Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return service, nil
}

// GetServiceEvents
//
//	@Description: 获取service相关事件
//	@param client
//	@param namespace
//	@param name
//	@return []map[string]any
//	@return error
func GetServiceEvents(client *kubernetes.Clientset, namespace, name string) ([]map[string]any, error) {
	selector := fields.AndSelectors(
		fields.OneTermEqualSelector("involvedObject.name", name),
		fields.OneTermEqualSelector("involvedObject.kind", "Service"),
	).String()
	events, err := client.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector:   selector,
		ResourceVersion: "0",
	})
	if err != nil {
		return nil, fmt.Errorf("获取service事件失败:%s", err.Error())
	}
	var result []map[string]any
	for _, event := range events.Items {
		lastSeen := ""
		if !event.LastTimestamp.IsZero() {
			lastSeen = event.LastTimestamp.Time.Format("2006-01-02 15:04:05")
		}
		result = append(result, map[string]any{
			"type":      event.Type,
			"reason":    event.Reason,
			"message":   event.Message,
			"last_seen": lastSeen,
		})
	}
	return result, nil
}

// GetServicesYaml
//
//	@Description: 根据名称获取service的yaml
//	@param client
//	@param namespace
//	@param name
//	@return string
//	@return error
func GetServicesYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	services, err := client.CoreV1().Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	services.TypeMeta = metav1.TypeMeta{APIVersion: "v1", Kind: "Service"}
	servicesYAML, err := yamlutil.MarshalWithoutManagedFields(services)
	if err != nil {
		return "", err
	}
	return servicesYAML, nil
}

// CreateService
//
//	@Description: 创建service
//	@param client
//	@param namespace
//	@param serviceYAML
//	@return error
func CreateService(client *kubernetes.Clientset, namespace, serviceYAML string) error {
	var service corev1.Service
	if err := yaml.Unmarshal([]byte(serviceYAML), &service); err != nil {
		return fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	_, err := client.CoreV1().Services(namespace).Create(context.TODO(), &service, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("创建service资源失败:%s", err.Error())
	}
	return nil
}

// UpdateService
//
//	@Description: 更新service
//	@param client
//	@param namespace
//	@param serviceYAML
//	@return error
func UpdateService(client *kubernetes.Clientset, namespace, serviceYAML string) error {
	var service corev1.Service
	if err := yaml.Unmarshal([]byte(serviceYAML), &service); err != nil {
		return fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	if service.Namespace == "" {
		service.Namespace = namespace
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest, err := client.CoreV1().Services(service.Namespace).Get(context.TODO(), service.Name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("获取service资源失败:%s", err.Error())
		}
		// 用用户 YAML 的可变字段覆盖最新对象,保留最新 resourceVersion
		// 注意: ClusterIP 在创建后不可变,不覆盖
		// 合并 Ports: 保留服务端分配的 nodePort(用户 YAML 可能未包含)
		latest.Spec.Ports = mergeServicePorts(latest.Spec.Ports, service.Spec.Ports)
		latest.Spec.Selector = service.Spec.Selector
		latest.Spec.Type = service.Spec.Type
		latest.Spec.SessionAffinity = service.Spec.SessionAffinity
		latest.Spec.SessionAffinityConfig = service.Spec.SessionAffinityConfig
		latest.Spec.LoadBalancerIP = service.Spec.LoadBalancerIP
		latest.Spec.LoadBalancerSourceRanges = service.Spec.LoadBalancerSourceRanges
		latest.Spec.ExternalTrafficPolicy = service.Spec.ExternalTrafficPolicy
		latest.Spec.ExternalIPs = service.Spec.ExternalIPs
		latest.Spec.InternalTrafficPolicy = service.Spec.InternalTrafficPolicy
		latest.Spec.AllocateLoadBalancerNodePorts = service.Spec.AllocateLoadBalancerNodePorts
		latest.Spec.IPFamilies = service.Spec.IPFamilies
		latest.Spec.IPFamilyPolicy = service.Spec.IPFamilyPolicy
		latest.Spec.HealthCheckNodePort = service.Spec.HealthCheckNodePort
		latest.Labels = service.Labels
		latest.Annotations = service.Annotations
		_, err = client.CoreV1().Services(service.Namespace).Update(context.TODO(), latest, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("更新service资源失败:%s", err.Error())
		}
		return nil
	})
}

// DeleteService
//
//	@Description: 删除service
//	@param client
//	@param namespace
//	@param name
//	@return error
func DeleteService(client *kubernetes.Clientset, namespace, name string) error {
	propagation := metav1.DeletePropagationForeground
	err := client.CoreV1().Services(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{
		PropagationPolicy: &propagation,
	})
	if err != nil {
		return fmt.Errorf("删除service资源失败:%s", err.Error())
	}
	return nil
}

// ServicePodList
//
//	@Description: 获取service关联的pod列表
//	@param client
//	@param namespace
//	@param name
//	@return *corev1.PodList
//	@return error
func ServicePodList(client *kubernetes.Clientset, namespace, name string) (*corev1.PodList, error) {
	svc, err := client.CoreV1().Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取service资源失败:%s", err.Error())
	}
	if len(svc.Spec.Selector) == 0 {
		return &corev1.PodList{}, nil
	}
	labelSelector := labels.Set(svc.Spec.Selector).AsSelectorPreValidated()
	podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector:  labelSelector.String(),
		ResourceVersion: "0",
	})
	if err != nil {
		return nil, fmt.Errorf("获取service关联pod列表失败:%s", err.Error())
	}
	return podList, nil
}

// mergeServicePorts 合并用户 YAML 的端口与集群中已有的端口。
// 对于 NodePort 类型的服务,如果用户 YAML 中某个端口未指定 nodePort,
// 保留服务端自动分配的值,避免因编辑导致端口变更中断流量。
func mergeServicePorts(existing, userDefined []corev1.ServicePort) []corev1.ServicePort {
	// 构建已有端口的 nodePort 索引
	existingNodePortsByName := make(map[string]int32, len(existing))
	existingNodePortsByPort := make(map[int32]int32, len(existing))
	for _, p := range existing {
		if p.NodePort > 0 {
			existingNodePortsByName[p.Name] = p.NodePort
			existingNodePortsByPort[p.Port] = p.NodePort
		}
	}
	// 对用户定义的端口,补充缺失的 nodePort
	merged := make([]corev1.ServicePort, len(userDefined))
	copy(merged, userDefined)
	for i := range merged {
		if merged[i].NodePort == 0 {
			// 优先按 name 查找,找不到则按 port 数值查找(处理重命名场景)
			if np, ok := existingNodePortsByName[merged[i].Name]; ok {
				merged[i].NodePort = np
			} else if np, ok := existingNodePortsByPort[merged[i].Port]; ok {
				merged[i].NodePort = np
			}
		}
	}
	return merged
}
