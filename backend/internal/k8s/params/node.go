package params

type NodeQueryParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	NodeName    string `form:"nodeName" json:"nodeName" label:"节点名称"`
	Name        string `form:"name" json:"name" label:"节点名称"` // 兼容前端 name 参数
}

// CordonNodeParams 封锁/解除封锁节点参数
type CordonNodeParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	Name        string `form:"name" json:"name" binding:"required" label:"节点名称"`
	Cordon      *bool  `json:"cordon" binding:"required" label:"是否封锁"`
}

// TaintNodeParams 单个污点参数（保留兼容）
type TaintNodeParams struct {
	ClusterName string `form:"clusterName" json:"clusterName" label:"集群名称"`
	NodeName    string `form:"nodeName" json:"nodeName" label:"节点名称"`
	Key         string `form:"key" json:"key" label:"污点key"`
	Value       string `form:"value" json:"value" label:"污点value"`
	Effect      string `form:"effect" json:"effect" label:"污点effect"`
}

// UpdateNodeTaintsParams 批量更新污点参数（替换式）
type UpdateNodeTaintsParams struct {
	ClusterName string      `json:"clusterName"`
	Name        string      `json:"name" binding:"required" label:"节点名称"`
	Taints      []TaintItem `json:"taints" label:"污点列表"`
}

// TaintItem 污点项
type TaintItem struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Effect string `json:"effect"`
}

// DrainNodeParams 驱逐节点参数
type DrainNodeParams struct {
	ClusterName     string `json:"clusterName"`
	Name            string `json:"name" binding:"required" label:"节点名称"`
	IgnoreDaemonSets bool  `json:"ignoreDaemonSets"` // 是否忽略 DaemonSet
	DeleteLocalData  bool  `json:"deleteLocalData"`  // 是否删除本地数据
	GracePeriod      int   `json:"gracePeriod"`       // 优雅终止秒数，-1=默认
	Force            bool  `json:"force"`             // 是否强制驱逐
}

// DeleteNodeParams 删除节点参数
type DeleteNodeParams struct {
	ClusterName string `json:"clusterName"`
	Name        string `json:"name" binding:"required" label:"节点名称"`
}

// UpdateNodeLabelsParams 更新节点标签参数
type UpdateNodeLabelsParams struct {
	ClusterName string            `json:"clusterName"`
	Name        string            `json:"name" binding:"required" label:"节点名称"`
	Labels      map[string]string `json:"labels" label:"标签"`
}
