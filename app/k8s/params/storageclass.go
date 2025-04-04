package params

type StorageClassQueryListParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
}

type StorageClassQueryByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Name        string `form:"name" json:"name" label:"名称"`
}

type StorageClassQueryByFieldParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	FieldMap    map[string]string `form:"fieldMap" json:"fieldMap" label:"字段"`
}

type StorageClassQueryByLabelParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	LabelMap    map[string]string `form:"labelMap" json:"labelMap" label:"标签"`
}

type StorageClassCreateParams struct {
	ClusterName      string `form:"clusterName" json:"clusterName" label:"集群名称"`
	StorageClassYaml string `form:"storageClassYaml" json:"storageClassYaml" label:"StorageClass Yaml"`
}

type StorageClassUpdateParams struct {
	ClusterName      string `form:"clusterName" json:"clusterName" label:"集群名称"`
	StorageClassYaml string `form:"storageClassYaml" json:"storageClassYaml" label:"StorageClass Yaml"`
}

type StorageClassDeleteByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Name        string `form:"name" json:"name" label:"名称"`
}

type StorageClassDeleteByFieldParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	FieldMap    map[string]string `form:"fieldMap" json:"fieldMap" label:"字段"`
}

type StorageClassDeleteByLabelParams struct {
	ClusterName string            `form:"clusterName" json:"clusterName" label:"集群名称"`
	LabelMap    map[string]string `form:"labelMap" json:"labelMap" label:"标签"`
}
