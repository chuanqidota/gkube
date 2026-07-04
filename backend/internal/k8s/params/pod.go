package params

type PodQueryListParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type PodQueryByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Name        string `form:"name" json:"name" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}



type PodCreateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	PodYaml     string `form:"podYaml" json:"podYaml" label:"Pod Yaml"`
}

type PodUpdateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	PodYaml     string `form:"podYaml" json:"podYaml" label:"Pod Yaml"`
}

type PodDeleteByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Name        string `form:"name" json:"name" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}



type PodEventQueryParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	PodName     string `form:"podName" json:"podName" label:"Pod名称"`
}
