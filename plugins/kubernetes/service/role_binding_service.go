package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ydcloud-dy/opshub/plugins/kubernetes/biz"
	"github.com/ydcloud-dy/opshub/plugins/kubernetes/model"
)

// RoleBindingService 角色绑定服务
type RoleBindingService struct {
	db         *gorm.DB
	clusterBiz *biz.ClusterBiz
}

// NewRoleBindingService 创建角色绑定服务
func NewRoleBindingService(db *gorm.DB) *RoleBindingService {
	return &RoleBindingService{
		db:         db,
		clusterBiz: biz.NewClusterBiz(db),
	}
}

// BindUserRole 绑定用户到K8s角色
func (s *RoleBindingService) BindUserRole(ctx context.Context, clusterID, userID uint64, roleName, roleNamespace, roleType string, boundBy uint64) error {
	// 检查是否已经绑定
	var existing model.K8sUserRoleBinding
	err := s.db.Where("cluster_id = ? AND user_id = ? AND role_name = ? AND role_namespace = ?",
		clusterID, userID, roleName, roleNamespace).First(&existing).Error

	if err == nil {
		return errors.New("用户已绑定该角色")
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	// 创建绑定
	binding := model.K8sUserRoleBinding{
		ClusterID:     clusterID,
		UserID:        userID,
		RoleName:      roleName,
		RoleNamespace: roleNamespace,
		RoleType:      roleType,
		BoundBy:       boundBy,
	}

	if err := s.db.Create(&binding).Error; err != nil {
		return fmt.Errorf("创建角色绑定失败: %w", err)
	}

	return nil
}

// UnbindUserRole 解绑用户K8s角色
func (s *RoleBindingService) UnbindUserRole(ctx context.Context, clusterID, userID uint64, roleName, roleNamespace string) error {
	result := s.db.Where("cluster_id = ? AND user_id = ? AND role_name = ? AND role_namespace = ?",
		clusterID, userID, roleName, roleNamespace).Delete(&model.K8sUserRoleBinding{})

	if result.Error != nil {
		return fmt.Errorf("解绑角色失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("绑定关系不存在")
	}

	return nil
}

// GetRoleBoundUsers 获取角色已绑定的用户列表
func (s *RoleBindingService) GetRoleBoundUsers(ctx context.Context, clusterID uint64, roleName, roleNamespace string) ([]map[string]interface{}, error) {
	type Result struct {
		UserID    uint64 `json:"userId"`
		Username  string `json:"username"`
		RealName  string `json:"realName"`
		BoundAt   string `json:"boundAt"`
	}

	var results []Result

	// 查询绑定关系及用户信息
	err := s.db.Table("k8s_user_role_bindings as b").
		Select("b.user_id as user_id, u.username, u.real_name, b.created_at as bound_at").
		Joins("LEFT JOIN sys_user u ON u.id = b.user_id").
		Where("b.cluster_id = ? AND b.role_name = ? AND b.role_namespace = ?", clusterID, roleName, roleNamespace).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// 转换为返回格式
	users := make([]map[string]interface{}, 0, len(results))
	for _, r := range results {
		users = append(users, map[string]interface{}{
			"userId":    r.UserID,
			"username":  r.Username,
			"realName":  r.RealName,
			"boundAt":   r.BoundAt,
		})
	}

	return users, nil
}

// GetUserClusterRoles 获取用户在指定集群的角色列表
func (s *RoleBindingService) GetUserClusterRoles(ctx context.Context, clusterID, userID uint64) ([]model.K8sUserRoleBinding, error) {
	var bindings []model.K8sUserRoleBinding
	err := s.db.Where("cluster_id = ? AND user_id = ?", clusterID, userID).
		Order("created_at DESC").
		Find(&bindings).Error

	return bindings, err
}

// GetUserRoleForCluster 获取用户在指定集群的所有角色
func (s *RoleBindingService) GetUserRoleForCluster(clusterID, userID uint64) ([]model.K8sUserRoleBinding, error) {
	var bindings []model.K8sUserRoleBinding
	err := s.db.Where("cluster_id = ? AND user_id = ?", clusterID, userID).
		Find(&bindings).Error

	return bindings, err
}

// GetAvailableUsers 获取可绑定的用户列表
func (s *RoleBindingService) GetAvailableUsers(ctx context.Context, keyword string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	type UserResult struct {
		ID       uint64 `json:"id"`
		Username string `json:"username"`
		RealName string `json:"realName"`
		Email    string `json:"email"`
	}

	var results []UserResult
	var total int64

	query := s.db.Table("sys_user").Select("id, username, real_name, email").Where("deleted_at IS NULL")

	if keyword != "" {
		keywordLike := "%" + keyword + "%"
		query = query.Where("username LIKE ? OR real_name LIKE ? OR email LIKE ?", keywordLike, keywordLike, keywordLike)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Scan(&results).Error; err != nil {
		return nil, 0, err
	}

	// 转换为返回格式
	users := make([]map[string]interface{}, 0, len(results))
	for _, r := range results {
		users = append(users, map[string]interface{}{
			"id":       r.ID,
			"username": r.Username,
			"realName": r.RealName,
			"email":    r.Email,
		})
	}

	return users, total, nil
}

// GetClusterCredentialUsers 获取集群的凭据用户列表（返回所有opshub开头的ServiceAccount）
func (s *RoleBindingService) GetClusterCredentialUsers(ctx context.Context, clusterID uint64, currentUserID uint64) ([]map[string]interface{}, error) {
	// 获取集群信息
	cluster, err := s.clusterBiz.GetCluster(ctx, uint(clusterID))
	if err != nil {
		return nil, fmt.Errorf("获取集群信息失败: %w", err)
	}

	// 获取 kubernetes clientset
	clientset, _, err := s.clusterBiz.GetRepo().GetClientset(cluster)
	if err != nil {
		return nil, fmt.Errorf("获取K8s客户端失败: %w", err)
	}

	// 列出 default 命名空间中的所有 ServiceAccount
	sas, err := clientset.CoreV1().ServiceAccounts("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("列出ServiceAccount失败: %w", err)
	}

	// 过滤出 opshub- 开头的 ServiceAccount，并从数据库查询用户信息
	credentialUsers := make([]map[string]interface{}, 0)

	for _, sa := range sas.Items {
		saName := sa.Name
		// 检查是否是 opshub- 开头的 ServiceAccount
		if !strings.HasPrefix(saName, "opshub-") {
			continue
		}

		// 解析格式: opshub-{username}
		// 例如: opshub-admin, opshub-dujie
		parts := strings.SplitN(saName, "-", 2)
		if len(parts) != 2 {
			continue
		}

		// 提取 username (opshub 之后的部分)
		username := parts[1]

		// 从数据库查询用户信息
		var user struct {
			ID       uint64
			Username string
			RealName string
		}
		err = s.db.Table("sys_user").
			Select("id, username, real_name").
			Where("username = ?", username).
			First(&user).Error

		// 如果查询不到用户信息，跳过
		if err != nil {
			continue
		}

		// 获取用户在该集群的凭据记录
		var kubeConfig model.UserKubeConfig
		err = s.db.Where("cluster_id = ? AND user_id = ? AND service_account = ?", clusterID, user.ID, saName).
			First(&kubeConfig).Error

		// 如果没有凭据记录，也显示（可能是直接在K8s中创建的）
		createdAt := sa.CreationTimestamp.Format("2006-01-02 15:04:05")
		if err == nil {
			// 如果有记录，使用数据库中的创建时间
			createdAt = kubeConfig.CreatedAt.Format("2006-01-02 15:04:05")
		}

		credentialUsers = append(credentialUsers, map[string]interface{}{
			"username":        username,         // 平台用户名
			"realName":        user.RealName,    // 真实姓名
			"serviceAccount":  saName,           // K8s ServiceAccount 完整名称
			"namespace":       sa.Namespace,     // 命名空间
			"userId":          user.ID,          // 平台用户ID
			"createdAt":       createdAt,        // 创建时间
		})
	}

	return credentialUsers, nil
}
