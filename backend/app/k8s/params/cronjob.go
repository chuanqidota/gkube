package params

type CronjobListParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type CronJobQueryByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Name        string `form:"name" json:"name" label:"名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
}

type CronJobQueryByLabelParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	LabelMap    map[string]string `form:"labelMap" json:"labelMap" label:"标签"`
	Namespace   string            `form:"namespace" json:"namespace" label:"命名空间"`
}

type CronJobQueryByFieldParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	FieldMap    map[string]string `form:"fieldMap" json:"fieldMap" label:"字段"`
	Namespace   string            `form:"namespace" json:"namespace" label:"命名空间"`
}

type CronJobCreateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Yaml        string `form:"yaml" json:"yaml" label:"Yaml"`
}

type CronJobUpdateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Yaml        string `form:"yaml" json:"yaml" label:"Yaml"`
}

type CronJobDeleteByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Namespace   string `form:"namespace" json:"namespace" label:"命名空间"`
	Name        string `form:"name" json:"name" label:"名称"`
}

type CronJobDeleteByLabelParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	LabelMap    map[string]string `form:"labelMap" json:"labelMap" label:"标签"`
	Namespace   string            `form:"namespace" json:"namespace" label:"命名空间"`
}

type CronJobDeleteByFieldParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	FieldMap    map[string]string `form:"fieldMap" json:"fieldMap" label:"字段"`
	Namespace   string            `form:"namespace" json:"namespace" label:"命名空间"`
}
