package params

type NamespaceCreateParams struct {
	ClusterName string            `json:"clusterName" label:"集群名称"`
	Namespace   string            `json:"namespace" label:"名称"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

type NamespaceLabelsParams struct {
	ClusterName string            `json:"clusterName" label:"集群名称"`
	Namespace   string            `json:"namespace" label:"名称"`
	Labels      map[string]string `json:"labels"`
}
