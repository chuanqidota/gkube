package params


type ClusterQueryParams struct {
    ClusterName string `form:"clusterName" json:"clusterName" binding:"required" label:"集群名称"`
}