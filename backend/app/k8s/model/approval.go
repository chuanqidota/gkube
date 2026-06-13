package model

import (
	"time"

	"gorm.io/gorm"
)

// ApprovalRequest represents an approval workflow request
type ApprovalRequest struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	RequestID   string         `json:"requestId" gorm:"uniqueIndex;size:64"`
	Type        string         `json:"type" gorm:"size:32;not null"` // deploy, scale, delete, config
	Resource    string         `json:"resource" gorm:"size:256;not null"`
	Namespace   string         `json:"namespace" gorm:"size:64"`
	Cluster     string         `json:"cluster" gorm:"size:64"`
	RequestedBy uint           `json:"requestedBy" gorm:"not null"`
	RequestedAt time.Time      `json:"requestedAt" gorm:"not null"`
	Status      string         `json:"status" gorm:"size:32;not null;default:pending"` // pending, approved, rejected
	ReviewedBy  *uint          `json:"reviewedBy"`
	ReviewedAt  *time.Time     `json:"reviewedAt"`
	Reason      string         `json:"reason" gorm:"size:1024"`
	Details     string         `json:"details" gorm:"type:text"` // JSON string
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (ApprovalRequest) TableName() string {
	return "approval_requests"
}
