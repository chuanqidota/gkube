package k8s

import (
	corev1 "k8s.io/api/core/v1"
)

type NodeInfo struct {
	Name             string
	Labels           map[string]string
	IsReady          bool
	Addresses        []corev1.NodeAddress
	CapacityCPU      string
	CapacityMemory   string
	AllocatableCPU   string
	AllocatableMem   string
	OSImage          string
	KernelVersion    string
	ContainerRuntime string
}
