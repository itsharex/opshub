package model

import (
	"time"
)

// UserKubeConfig 用户K8s凭据表
type UserKubeConfig struct {
	ID             uint64    `gorm:"primaryKey" json:"id"`
	ClusterID      uint64    `gorm:"not null;index" json:"clusterId"`
	UserID         uint64    `gorm:"not null;index" json:"userId"`
	ServiceAccount string    `gorm:"size:255;not null;index" json:"serviceAccount"`
	Namespace      string    `gorm:"size:255;default:'default'" json:"namespace"`
	IsActive       bool      `gorm:"default:1" json:"isActive"`
	CreatedBy      uint64    `gorm:"not null" json:"createdBy"`
	CreatedAt      time.Time `json:"createdAt"`
	RevokedAt      *time.Time `json:"revokedAt"`
}

// TableName 指定表名
func (UserKubeConfig) TableName() string {
	return "k8s_user_kube_configs"
}
