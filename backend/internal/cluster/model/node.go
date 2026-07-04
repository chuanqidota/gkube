package model

import (
	corev1 "k8s.io/api/core/v1"
)

type NodeInfo struct {
	Name             string               `json:"name" label:"名称"`
	Status           string               `json:"status" label:"状态"`
	Roles            string               `json:"roles" label:"角色"`
	Version          string               `json:"version" label:"版本"`
	InternalIP       string               `json:"internal_ip" label:"内部IP"`
	ExternalIP       string               `json:"external_ip" label:"外部IP"`
	Architecture     string               `json:"architecture" label:"架构"`
	Unschedulable    bool                 `json:"unschedulable" label:"不可调度"`
	PodCount         int                  `json:"pod_count" label:"Pod数量"`
	Labels           map[string]string    `json:"labels" label:"标签"`
	Taints           []corev1.Taint       `json:"taints" label:"污点"`
	IsReady          bool                 `json:"is_ready" label:"是否就绪"`
	Addresses        []corev1.NodeAddress `json:"addresses" label:"地址"`
	CapacityCPU      string               `json:"capacity_cpu" label:"总cpu"`
	CapacityMemory   string               `json:"capacity_memory" label:"总内存"`
	AllocatableCPU   string               `json:"allocatable_cpu" label:"可用cpu"`
	AllocatableMem   string               `json:"allocatable_mem" label:"可用内存"`
	OSImage          string               `json:"os_image" label:"操作系统"`
	KernelVersion    string               `json:"kernel_version" label:"内核版本"`
	ContainerRuntime string               `json:"container_runtime" label:"容器运行时"`
	Age              string               `json:"age" label:"年龄"`
}
