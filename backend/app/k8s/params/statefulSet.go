package params

type StatefulSetQueryListParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type StatefulSetQueryByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" label:"名称"`
}

type StatefulSetQueryByLabelParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	LabelMap    map[string]string `form:"labelMap" json:"labelMap" label:"标签"`
	Namespace   string            `form:"namespace" json:"namespace" label:"命名空间"`
}

type StatefulSetQueryByFieldParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	FieldMap    map[string]string `form:"fieldMap" json:"fieldMap" label:"字段"`
	Namespace   string            `form:"namespace" json:"namespace" label:"命名空间"`
}

type StatefulSetCreateParams struct {
	ClusterName     string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace       string `form:"namespace" json:"namespace" label:"命名空间"`
	StatefulSetYaml string `form:"statefulSetYaml" json:"statefulSet" label:"StatefulSet"`
}

type StatefulSetUpdateParams struct {
	ClusterName     string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace       string `form:"namespace" json:"namespace" label:"命名空间"`
	StatefulSetYaml string `form:"statefulSetYaml" json:"statefulSet" label:"StatefulSet"`
}

type StatefulSetDeleteByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" label:"名称"`
}

type StatefulSetDeleteByLabelParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	LabelMap    map[string]string `form:"labelMap" json:"labelMap" label:"标签"`
	Namespace   string            `form:"namespace" json:"namespace" label:"命名空间"`
}

type StatefulSetDeleteByFieldParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	FieldMap    map[string]string `form:"fieldMap" json:"fieldMap" label:"字段"`
	Namespace   string            `form:"namespace" json:"namespace" label:"命名空间"`
}
