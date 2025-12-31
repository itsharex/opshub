package model

import (
	"time"
)

// K8sUserRoleBinding 用户K8s角色绑定
type K8sUserRoleBinding struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	ClusterID     uint64    `gorm:"not null;index:idx_cluster_id;index:idx_cluster_user_role" json:"clusterId"`
	UserID        uint64    `gorm:"not null;index:idx_user_id;index:idx_cluster_user_role" json:"userId"`
	RoleName      string    `gorm:"size:255;not null;index:idx_cluster_user_role" json:"roleName"`
	RoleNamespace string    `gorm:"size:255;default:'';index:idx_cluster_user_role" json:"roleNamespace"`
	RoleType      string    `gorm:"size:50;not null" json:"roleType"` // ClusterRole 或 Role
	BoundBy       uint64    `gorm:"not null" json:"boundBy"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// TableName 指定表名
func (K8sUserRoleBinding) TableName() string {
	return "k8s_user_role_bindings"
}
