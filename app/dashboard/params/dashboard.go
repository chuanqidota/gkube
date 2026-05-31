package params

type EventQueryParams struct {
	ClusterID *uint  `form:"clusterId" json:"clusterId" label:"集群ID"`
	Type      string `form:"type" json:"type" label:"事件类型"`
	Limit     int    `form:"limit" json:"limit" label:"数量限制"`
}
