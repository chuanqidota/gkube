package model

import (
	"time"

	"gorm.io/gorm"
)

// RBACMatrix stores role-resource-verb permission mappings
type RBACMatrix struct {
	ID        uint           `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	RoleID    uint           `gorm:"column:role_id;index;not null;comment:角色ID" json:"roleId"`
	Resource  string         `gorm:"column:resource;size:64;not null;comment:资源类型" json:"resource"`
	Verb      string         `gorm:"column:verb;size:32;not null;comment:操作类型" json:"verb"`
	Allowed   bool           `gorm:"column:allowed;default:false;comment:是否允许" json:"allowed"`
	CreatedAt time.Time      `gorm:"column:created_at;comment:创建时间" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"column:updated_at;comment:更新时间" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

func (RBACMatrix) TableName() string {
	return "rbac_matrix"
}
