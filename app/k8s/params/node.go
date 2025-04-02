package params

type NodeQueryParams struct {
    ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	NodeName string `form:"nodeName" json:"nodeName" label:"节点名称"`
}