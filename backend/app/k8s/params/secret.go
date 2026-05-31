package params

type SecretQueryListParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type SecretQueryByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Name        string `form:"name" json:"name" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type SecretCreateParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string            `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string            `form:"name" json:"name" label:"名称"`
	Data        map[string]string `form:"data" json:"data" label:"data"`
}

type SecretUpdateParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string            `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string            `form:"name" json:"name" label:"名称"`
	Data        map[string]string `form:"data" json:"data" label:"data"`
}

type SecretDeleteParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" label:"名称"`
}
