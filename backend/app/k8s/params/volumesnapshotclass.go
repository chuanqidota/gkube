package params

type VolumeSnapshotClassQueryListParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
}

type VolumeSnapshotClassQueryByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Name        string `form:"name" json:"name" label:"名称"`
}

type VolumeSnapshotClassCreateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Yaml        string `form:"yaml" json:"yaml" label:"Yaml"`
}

type VolumeSnapshotClassUpdateParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Yaml        string `form:"yaml" json:"yaml" label:"Yaml"`
}

type VolumeSnapshotClassDeleteByNameParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Name        string `form:"name" json:"name" label:"名称"`
}
