package params

type PvListParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
}

type PvQueryByName struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Name        string `form:"name" json:"name" label:"名称"`
}



type PvCreateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Yaml        string `form:"yaml" json:"yaml" label:"Yaml"`
}

type PvUpdateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Yaml        string `form:"yaml" json:"yaml" label:"Yaml"`
}

