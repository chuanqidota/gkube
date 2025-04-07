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

type PodQueryByLabelParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	LabelMap    map[string]string `form:"labelMap" json:"labelMap" label:"标签"`
	Namespace   string            `form:"namespace" json:"namespace" label:"命名空间"`
}

type PodQueryByFiledParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	FieldMap    map[string]string `form:"fieldMap" json:"fieldMap" label:"字段"`
	Namespace   string            `form:"namespace" json:"namespace" label:"命名空间"`
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

type PodDeleteByLabelParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	LabelMap    map[string]string `form:"labelMap" json:"labelMap" label:"标签"`
	Namespace   string            `form:"namespace" json:"namespace" label:"命名空间"`
}

type PodDeleteByFieldParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	FieldMap    map[string]string `form:"fieldMap" json:"fieldMap" label:"字段"`
	Namespace   string            `form:"namespace" json:"namespace" label:"命名空间"`
}

type PodEventQueryParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	PodName     string `form:"podName" json:"podName" label:"Pod名称"`
}
