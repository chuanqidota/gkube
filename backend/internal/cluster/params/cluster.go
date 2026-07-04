package params

type CreateClusterParams struct {
	ClusterName string            `json:"clusterName" binding:"required" label:"集群名称"`
	DisplayName string            `json:"displayName" label:"显示名称"`
	Description string            `json:"description" label:"描述"`
	KubeConfig  string            `json:"kubeConfig" binding:"required" label:"KubeConfig"`
	Labels      map[string]string `json:"labels" label:"标签"`
}

type UpdateClusterParams struct {
	ID          uint              `json:"id" binding:"required" label:"集群ID"`
	DisplayName string            `json:"displayName" label:"显示名称"`
	Description string            `json:"description" label:"描述"`
	Labels      map[string]string `json:"labels" label:"标签"`
}

type ClusterQueryParams struct {
	Page    int    `form:"page" json:"page" label:"页码"`
	Size    int    `form:"size" json:"size" label:"每页数量"`
	Status  string `form:"status" json:"status" label:"状态"`
	Keyword string `form:"keyword" json:"keyword" label:"关键词"`
}

type ClusterIDParams struct {
	ID uint `uri:"id" binding:"required" label:"集群ID"`
}
