package model

// 记录
type TerminalRecord struct {
	BaseModel
	Key         string `json:"key" gorm:"column:key;type:string;size:100;comment:唯一标识"`
	ClusterName string `json:"clusterName" gorm:"column:cluster_name;type:string;size:100;comment:集群名称"`
	Namespace   string `json:"namespace" gorm:"column:namespace;type:string;size:100;comment:命名空间"`
	PodName     string `json:"podName" gorm:"column:pod_name;type:string;size:100;comment:pod名称"`
}


func (TerminalRecord) TableName() string {
	return "terminal_record"
}