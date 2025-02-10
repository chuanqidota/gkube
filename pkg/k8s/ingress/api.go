package ingress

import (
	"context"
	"encoding/json"
	"fmt"

	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

func GetIngressList(client *kubernetes.Clientset, namespace string) ([]netv1.Ingress, error) {
	ingress, err := client.NetworkingV1().Ingresses(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return ingress.Items, nil
}

func GetIngressYaml(client *kubernetes.Clientset, namespace, name string) (string, error) {
	ingress, err := client.NetworkingV1().Ingresses(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return "", nil
	}
	ingressJSON, err := json.Marshal(ingress)
	if err != nil {
		return "", err
	}
	configmapYAML, err := yaml.JSONToYAML(ingressJSON)
	if err != nil {
		return "", err
	}
	return string(configmapYAML), nil
}

func CreateIngress(client *kubernetes.Clientset, namespace, name, host, path, serviceName, servicePort string) (bool, error) {
	ingressYAML := fmt.Sprintf(
		`
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: %s
  namespace: %s
spec:
  rules:
  - host: %s
    http:
      paths:
      - path: %s
        pathType: Prefix
        backend:
          service:
            name: %s
            port:
              number: %s
        
        `, name, namespace, host, path, serviceName, servicePort)
	var ingress netv1.Ingress
	if err := yaml.Unmarshal([]byte(ingressYAML), &ingress); err != nil {
		return false, fmt.Errorf("yaml文件错误:%s", err.Error())
	}
	_, err := client.NetworkingV1().Ingresses(namespace).Create(context.TODO(), &ingress, metav1.CreateOptions{})
	if err != nil {
		return false, fmt.Errorf("创建ingress资源失败:%s",err.Error())
	}
    return true,nil

}
