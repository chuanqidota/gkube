package model

import (
	"time"

	authmodel "gkube/internal/auth/model"
	clustermodel "gkube/internal/cluster/model"
)

// PermissionBinding 权限绑定表
type PermissionBinding struct {
	ID          uint               `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint               `gorm:"not null;uniqueIndex:idx_user_cluster_ns" json:"userId"`
	RoleID      uint               `gorm:"not null" json:"roleId"`
	ClusterID   uint               `gorm:"not null;uniqueIndex:idx_user_cluster_ns" json:"clusterId"`
	Namespace   string             `gorm:"type:varchar(100);not null;default:'';uniqueIndex:idx_user_cluster_ns;comment:空串=集群级" json:"namespace"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
	User        authmodel.User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Role        Role               `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Cluster     clustermodel.K8SCluster `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
}

func (PermissionBinding) TableName() string { return "permission_binding" }
