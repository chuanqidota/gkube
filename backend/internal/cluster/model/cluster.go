package model

import (
	"time"

	"gorm.io/gorm"
)

// K8SCluster
// @Description: k8s集群配置表
type K8SCluster struct {
	ID              uint           `json:"id" gorm:"primarykey"`
	ClusterName     string         `json:"clusterName" gorm:"column:cluster_name;type:string;size:100;uniqueIndex;comment:集群名称" binding:"required"`
	DisplayName     string         `json:"displayName" gorm:"column:display_name;type:string;size:200;comment:集群显示名称"`
	Description     string         `json:"description" gorm:"column:description;type:string;size:500;comment:集群描述"`
	KubeConfig      string         `json:"-" gorm:"column:kube_config;type:text;size:12800;comment:集群凭证"`
	Status          string         `json:"status" gorm:"column:status;type:string;size:50;default:online;comment:集群状态"`
	ClusterVersion  string         `json:"clusterVersion" gorm:"column:cluster_version;type:string;size:100;comment:集群版本"`
	NodeCount       int            `json:"nodeCount" gorm:"column:node_count;comment:节点数量"`
	LastHealthCheck time.Time      `json:"lastHealthCheck" gorm:"column:last_health_check;comment:最后健康检查时间"`
	Labels          string         `json:"labels" gorm:"column:labels;type:text;comment:标签(JSON格式)"`
	CreatedAt       time.Time      `json:"createdAt" gorm:"column:created_at;comment:创建时间"`
	UpdatedAt       time.Time      `json:"updatedAt" gorm:"column:updated_at;comment:更新时间"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index;comment:删除时间"`
}

func (K8SCluster) TableName() string {
	return "k8s_cluster"
}
