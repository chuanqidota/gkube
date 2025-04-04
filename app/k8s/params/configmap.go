package params

type ConfigMapQueryListParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type ConfigMapQueryByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Name        string `form:"name" json:"name" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type ConfigMapCreateParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string            `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string            `form:"name" json:"name" label:"名称"`
	Data        map[string]string `form:"data" json:"data" label:"Yaml"`
}

type ConfigMapUpdateParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string            `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string            `form:"name" json:"name" label:"名称"`
	Data        map[string]string `form:"data" json:"data" label:"Yaml"`
}

type ConfigMapDeleteByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Name        string `form:"name" json:"name" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}
