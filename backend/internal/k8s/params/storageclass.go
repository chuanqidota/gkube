package params

type StorageClassQueryListParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
}

type StorageClassQueryByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Name        string `form:"name" json:"name" label:"名称"`
}



type StorageClassCreateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Yaml        string `form:"yaml" json:"yaml" label:"Yaml"`
}

type StorageClassUpdateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Yaml        string `form:"yaml" json:"yaml" label:"Yaml"`
}

type StorageClassDeleteByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Name        string `form:"name" json:"name" label:"名称"`
}


