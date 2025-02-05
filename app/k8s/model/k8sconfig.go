package model

type K8SCluster struct {
	BaseModel
	ClusterName    string `json:"clusterName" gorm:"column:cluster_name;type:string;size:100;comment:集群名称" binding:"required"`
	KubeConfig     string `json:"kubeConfig" gorm:"column:kube_config;type:string;size:12800;comment:集群凭证" binding:"required"`
	ClusterVersion string `json:"clusterVersion" gorm:"comment:集群版本"`
}

func (K8SCluster) TableName() string {
	return "k8s_cluster"
}
