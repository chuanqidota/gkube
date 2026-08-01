package params

// DashboardQueryParams 仪表盘聚合查询参数(overview/resources/workloads 共用)。
// ClusterID 命中时仅查询该集群;为空时聚合所有 online 集群(向后兼容)。
type DashboardQueryParams struct {
	ClusterID *uint `form:"clusterId" json:"clusterId" label:"集群ID"`
}

type EventQueryParams struct {
	ClusterID     *uint  `form:"clusterId" json:"clusterId" label:"集群ID"`
	Namespace     string `form:"namespace" json:"namespace" label:"命名空间"`
	Type          string `form:"type" json:"type" label:"事件类型"`
	FieldSelector string `form:"fieldSelector" json:"fieldSelector" label:"字段选择器"`
	Limit         int    `form:"limit" json:"limit" label:"数量限制"`
	Continue      string `form:"continue" json:"continue" label:"分页标记"`
}
