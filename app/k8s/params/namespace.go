package params

type NamespaceCreateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"名称"`
}
