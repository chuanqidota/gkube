package model

import (
	"gorm.io/gorm"
	"time"
)

// Permission
// @Description: 权限表
type Permission struct {
	ID          uint           `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	Resource    string         `gorm:"column:resource;type:varchar(100);not null;comment:资源" json:"resource"`
	Action      string         `gorm:"column:action;type:varchar(50);not null;comment:操作" json:"action"`
	ClusterID   *uint          `gorm:"column:cluster_id;comment:集群ID" json:"cluster_id"`
	Namespace   string         `gorm:"column:namespace;type:varchar(100);comment:命名空间" json:"namespace"`
	Description string         `gorm:"column:description;type:text;comment:描述" json:"description"`
	CreatedAt   time.Time      `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

func (Permission) TableName() string {
	return "permission"
}
