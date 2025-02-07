package k8s

import (
	corev1 "k8s.io/api/core/v1"
)

type NodeInfo struct {
	Name             string               `json:"name" label:"名称"`
	Labels           map[string]string    `json:"labels" label:"标签"`
	IsReady          bool                 `json:"is_ready" label:"是否就绪"`
	Addresses        []corev1.NodeAddress `json:"addresses" label:"地址"`
	CapacityCPU      string               `json:"capacity_cpu" label:"总cpu"`
	CapacityMemory   string               `json:"capacity_memory" label:"总内存"`
	AllocatableCPU   string               `json:"allocatable_cpu" label:"可用cpu"`
	AllocatableMem   string               `json:"allocatable_mem" label:"可用内存"`
	OSImage          string               `json:"os_image" label:"操作系统"`
	KernelVersion    string               `json:"kernel_version" label:"内核版本"`
	ContainerRuntime string               `json:"container_runtime" label:"容器运行时"`
}
