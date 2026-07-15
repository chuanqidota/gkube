package model

import (
	"gorm.io/gorm"
	"time"
)

// User
// @Description: 用户表
type User struct {
	ID           uint           `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	Username     string         `gorm:"column:username;type:varchar(100);uniqueIndex;not null;comment:用户名" json:"username"`
	PasswordHash string         `gorm:"column:password_hash;type:varchar(255);not null;comment:密码哈希" json:"-"`
	Email        string         `gorm:"column:email;type:varchar(200);comment:邮箱" json:"email"`
	DisplayName  string         `gorm:"column:display_name;type:varchar(100);comment:显示名称" json:"display_name"`
	Status       int            `gorm:"column:status;type:tinyint;default:1;comment:状态 1:启用 0:禁用" json:"status"`
	CreatedAt    time.Time      `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;comment:更新时间" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

func (User) TableName() string {
	return "user"
}
