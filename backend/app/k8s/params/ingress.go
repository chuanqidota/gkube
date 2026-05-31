package params

type IngressQueryListParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type IngressQueryByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Name        string `form:"name" json:"name" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type IngressQueryByLabelParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	LabelMap    map[string]string `form:"labelMap" json:"labelMap" label:"标签"`
	Namespace   string            `form:"namespace" json:"namespace" label:"命名空间"`
}

type IngressQueryByFieldParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	FieldMap    map[string]string `form:"fieldMap" json:"fieldMap" label:"字段"`
	Namespace   string            `form:"namespace" json:"namespace" label:"命名空间"`
}

type IngressCreateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	IngressYaml string `form:"ingressYaml" json:"ingressYaml" label:"Ingress Yaml"`
}

type IngressUpdateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	IngressYaml string `form:"ingressYaml" json:"ingressYaml" label:"Ingress Yaml"`
}

type IngressDeleteByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Name        string `form:"name" json:"name" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type IngressDeleteByLabelParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	LabelMap    map[string]string `form:"labelMap" json:"labelMap" label:"标签"`
	Namespace   string            `form:"namespace" json:"namespace" label:"命名空间"`
}

type IngressDeleteByFieldParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	FieldMap    map[string]string `form:"fieldMap" json:"fieldMap" label:"字段"`
	Namespace   string            `form:"namespace" json:"namespace" label:"命名空间"`
}
