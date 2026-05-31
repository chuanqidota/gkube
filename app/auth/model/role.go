package model

import (
	"gorm.io/gorm"
	"time"
)

// Role
// @Description: 角色表
type Role struct {
	ID          uint           `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	Name        string         `gorm:"column:name;type:varchar(100);uniqueIndex;not null;comment:角色名称" json:"name"`
	Description string         `gorm:"column:description;type:text;comment:描述" json:"description"`
	Permissions []Permission   `gorm:"many2many:role_permission;comment:权限列表" json:"permissions"`
	CreatedAt   time.Time      `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

func (Role) TableName() string {
	return "role"
}
