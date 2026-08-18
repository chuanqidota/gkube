package service

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// EndpointSubset represents a simplified view of an Endpoints subset.
type EndpointSubset struct {
	Ports    []EndpointPort    `json:"ports"`
	Addresses []EndpointAddress `json:"addresses"`
	NotReady []EndpointAddress `json:"not_ready_addresses"`
}

// EndpointPort represents a single port in an endpoint.
type EndpointPort struct {
	Name     string `json:"name"`
	Port     int32  `json:"port"`
	Protocol string `json:"protocol"`
}

// EndpointAddress represents a single backend address.
type EndpointAddress struct {
	IP       string  `json:"ip"`
	NodeName *string `json:"node_name"`
	PodName  string  `json:"pod_name"`
}

// GetServiceEndpoints returns the Endpoints object for a given Service,
// transformed into a simplified structure for the frontend.
func GetServiceEndpoints(client *kubernetes.Clientset, namespace, name string) ([]EndpointSubset, error) {
	ep, err := client.CoreV1().Endpoints(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Service Endpoints失败:%s", err.Error())
	}

	var subsets []EndpointSubset
	for _, ss := range ep.Subsets {
		subset := EndpointSubset{
			Ports:     make([]EndpointPort, 0, len(ss.Ports)),
			Addresses: make([]EndpointAddress, 0, len(ss.Addresses)),
			NotReady:  make([]EndpointAddress, 0, len(ss.NotReadyAddresses)),
		}

		for _, p := range ss.Ports {
			subset.Ports = append(subset.Ports, EndpointPort{
				Name:     p.Name,
				Port:     p.Port,
				Protocol: string(p.Protocol),
			})
		}

		for _, addr := range ss.Addresses {
			subset.Addresses = append(subset.Addresses, transformAddress(addr))
		}

		for _, addr := range ss.NotReadyAddresses {
			subset.NotReady = append(subset.NotReady, transformAddress(addr))
		}

		subsets = append(subsets, subset)
	}

	// Return empty slice instead of nil for consistent JSON output
	if subsets == nil {
		subsets = []EndpointSubset{}
	}
	return subsets, nil
}

func transformAddress(addr corev1.EndpointAddress) EndpointAddress {
	a := EndpointAddress{
		IP:       addr.IP,
		NodeName: addr.NodeName,
	}
	if addr.TargetRef != nil {
		a.PodName = addr.TargetRef.Name
	}
	return a
}
