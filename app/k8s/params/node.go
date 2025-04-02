package params

type NodeQueryParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	NodeName    string `form:"nodeName" json:"nodeName" label:"节点名称"`
}

type NodeEvictParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	PodName     string `form:"podName" json:"podName" label:"Pod名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}
