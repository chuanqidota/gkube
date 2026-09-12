package k8s

import k8sLabels "gkube/pkg/k8s/labels"

// ---------------------------------------------------------------------------
// 共享参数结构体 — 替代各 handler 文件中重复定义的 per-resource param structs
// ---------------------------------------------------------------------------

// ListParams 用于分页列表请求（GET/POST，绑定 query + json body）
// 适用：Deployment, Pod, ConfigMap, Secret, PVC, ReplicaSet, StatefulSet,
//
//	DaemonSet, Job, CronJob, Service, Ingress, NetworkPolicy,
//	HPA, ResourceQuota, LimitRange, VolumeSnapshot 等
type ListParams struct {
	ClusterName  string                  `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace    string                  `form:"namespace" json:"namespace" label:"命名空间"`
	Limit        int64                   `form:"limit" json:"limit" label:"每页条数"`
	Continue     string                  `form:"continue" json:"continue" label:"分页标记"`
	LabelFilters []k8sLabels.LabelFilter `json:"labelFilters" form:"labelFilters" label:"标签过滤"`
}

// NamespacedParams 按 namespace + name 定位资源（detail / yaml / events / delete / pods）
// GET 请求用 ShouldBindQuery，DELETE 请求用 ShouldBindJSON。
// 适用：Deployment, Pod, Service, StatefulSet, DaemonSet, Job, CronJob,
//
//	Ingress, PVC, HPA, ReplicaSet, NetworkPolicy, VolumeSnapshot
type NamespacedParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
}

// NamespacedRequiredParams 与 NamespacedParams 相同，但 Namespace 为必填。
// 适用：ConfigMap、Secret 的 create/update（后端 pkg 层要求 namespace 非空）
type NamespacedRequiredParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" binding:"required" label:"命名空间"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
}

// ClusterScopedParams 按 name 定位集群级资源（无 namespace）
// 适用：PV, StorageClass, Node, VolumeSnapshotClass
type ClusterScopedParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
}

// CreateParams 创建 namespaced 资源（从 YAML）
// 适用：Deployment, Pod, Service, StatefulSet, DaemonSet, Job, CronJob,
//
//	Ingress, PVC, HPA, NetworkPolicy, VolumeSnapshot, ResourceQuota, LimitRange
type CreateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Yaml        string `form:"yaml" json:"yaml" binding:"required" label:"YAML"`
}

// UpdateParams 更新 namespaced 资源（从 YAML，需要 name）
// 适用：Deployment, StatefulSet, DaemonSet, Pod (PatchPodMetadata)
type UpdateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" binding:"required" label:"名称"`
	Yaml        string `form:"yaml" json:"yaml" binding:"required" label:"YAML"`
}

// ClusterCreateParams 创建/更新集群级资源（从 YAML，无 namespace）
// 适用：PV, StorageClass, VolumeSnapshotClass
type ClusterCreateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Yaml        string `form:"yaml" json:"yaml" binding:"required" label:"YAML"`
}

// ClusterListParams 集群级资源的分页列表参数
// 适用：PV, StorageClass, Node, VolumeSnapshotClass
type ClusterListParams struct {
	ClusterName  string                  `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
	Limit        int64                   `form:"limit" json:"limit" label:"每页条数"`
	Continue     string                  `form:"continue" json:"continue" label:"分页标记"`
	LabelFilters []k8sLabels.LabelFilter `json:"labelFilters" form:"labelFilters" label:"标签过滤"`
}
