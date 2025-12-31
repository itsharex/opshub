package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"

	"gorm.io/gorm"
	"k8s.io/client-go/kubernetes"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/api/core/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/ydcloud-dy/opshub/plugins/kubernetes/biz"
	"github.com/ydcloud-dy/opshub/plugins/kubernetes/data/models"
	"github.com/ydcloud-dy/opshub/plugins/kubernetes/model"
)

// ClusterService 集群服务层
type ClusterService struct {
	clusterBiz *biz.ClusterBiz
	db         *gorm.DB

	// 缓存已连接的集群 clientset
	clientsetCache map[uint]*kubernetes.Clientset
	metricsCache   map[uint]*metricsv.Clientset
	cacheMutex     sync.RWMutex
}

// NewClusterService 创建集群服务
func NewClusterService(db *gorm.DB) *ClusterService {
	return &ClusterService{
		clusterBiz:     biz.NewClusterBiz(db),
		db:             db,
		clientsetCache: make(map[uint]*kubernetes.Clientset),
		metricsCache:   make(map[uint]*metricsv.Clientset),
	}
}

// CreateClusterRequest 创建集群请求
type CreateClusterRequest struct {
	Name        string `json:"name" binding:"required"`
	Alias       string `json:"alias"`
	APIEndpoint string `json:"apiEndpoint"` // 移除 required，因为 KubeConfig 中已经包含
	KubeConfig  string `json:"kubeConfig" binding:"required"`
	Region      string `json:"region"`
	Provider    string `json:"provider"`
	Description string `json:"description"`
	UserID      uint   `json:"userId"`
}

// UpdateClusterRequest 更新集群请求
type UpdateClusterRequest struct {
	Name        string `json:"name"`
	Alias       string `json:"alias"`
	APIEndpoint string `json:"apiEndpoint"`
	KubeConfig  string `json:"kubeConfig"`
	Region      string `json:"region"`
	Provider    string `json:"provider"`
	Description string `json:"description"`
}

// ClusterDetailResponse 集群详情响应
type ClusterDetailResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Alias       string `json:"alias"`
	APIEndpoint string `json:"apiEndpoint"`
	Version     string `json:"version"`
	Status      int    `json:"status"`
	NodeCount   int    `json:"nodeCount"`   // 节点数量
	Region      string `json:"region"`
	Provider    string `json:"provider"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// CreateCluster 创建集群
func (s *ClusterService) CreateCluster(ctx context.Context, req *CreateClusterRequest) (*ClusterDetailResponse, error) {
	bizReq := &biz.CreateClusterRequest{
		Name:        req.Name,
		Alias:       req.Alias,
		APIEndpoint: req.APIEndpoint,
		KubeConfig:  req.KubeConfig,
		Region:      req.Region,
		Provider:    req.Provider,
		Description: req.Description,
		CreatedBy:   req.UserID,
	}

	cluster, err := s.clusterBiz.CreateCluster(ctx, bizReq)
	if err != nil {
		return nil, err
	}

	return s.toClusterResponse(cluster), nil
}

// UpdateCluster 更新集群
func (s *ClusterService) UpdateCluster(ctx context.Context, id uint, req *UpdateClusterRequest) (*ClusterDetailResponse, error) {
	bizReq := &biz.UpdateClusterRequest{
		Name:        req.Name,
		Alias:       req.Alias,
		APIEndpoint: req.APIEndpoint,
		KubeConfig:  req.KubeConfig,
		Region:      req.Region,
		Provider:    req.Provider,
		Description: req.Description,
	}

	cluster, err := s.clusterBiz.UpdateCluster(ctx, id, bizReq)
	if err != nil {
		return nil, err
	}

	// 清除缓存
	s.clearClientsetCache(id)

	return s.toClusterResponse(cluster), nil
}

// DeleteCluster 删除集群
func (s *ClusterService) DeleteCluster(ctx context.Context, id uint) error {
	// 清除缓存
	s.clearClientsetCache(id)

	return s.clusterBiz.DeleteCluster(ctx, id)
}

// GetCluster 获取集群详情
func (s *ClusterService) GetCluster(ctx context.Context, id uint) (*ClusterDetailResponse, error) {
	cluster, err := s.clusterBiz.GetCluster(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toClusterResponse(cluster), nil
}

// ListClusters 获取集群列表
func (s *ClusterService) ListClusters(ctx context.Context) ([]ClusterDetailResponse, error) {
	clusters, err := s.clusterBiz.ListClusters(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]ClusterDetailResponse, 0, len(clusters))
	for _, cluster := range clusters {
		responses = append(responses, *s.toClusterResponse(&cluster))
	}

	return responses, nil
}

// TestClusterConnection 测试集群连接
func (s *ClusterService) TestClusterConnection(ctx context.Context, id uint) (string, error) {
	// 清除缓存，强制重新连接
	s.clearClientsetCache(id)

	return s.clusterBiz.TestClusterConnection(ctx, id)
}

// GetCachedClientset 获取缓存的 clientset
func (s *ClusterService) GetCachedClientset(ctx context.Context, id uint) (*kubernetes.Clientset, error) {
	s.cacheMutex.RLock()
	clientset, exists := s.clientsetCache[id]
	s.cacheMutex.RUnlock()

	if exists {
		return clientset, nil
	}

	// 缓存不存在，创建新的
	clientset, err := s.clusterBiz.GetClusterClientset(ctx, id)
	if err != nil {
		return nil, err
	}

	// 存入缓存
	s.cacheMutex.Lock()
	s.clientsetCache[id] = clientset
	s.cacheMutex.Unlock()

	return clientset, nil
}

// GetClusterKubeConfig 获取集群的 KubeConfig（解密后的）
func (s *ClusterService) GetClusterKubeConfig(ctx context.Context, id uint) (string, error) {
	cluster, err := s.clusterBiz.GetCluster(ctx, id)
	if err != nil {
		return "", err
	}

	// 调用 biz 层的解密方法
	kubeConfig, err := biz.DecryptKubeConfig(cluster.KubeConfig)
	if err != nil {
		return "", err
	}

	return kubeConfig, nil
}

// ClearClientsetCache 清除指定集群的 clientset 缓存
func (s *ClusterService) ClearClientsetCache(id uint) {
	s.clearClientsetCache(id)
}

// clearClientsetCache 内部方法：清除缓存
func (s *ClusterService) clearClientsetCache(id uint) {
	s.cacheMutex.Lock()
	delete(s.clientsetCache, id)
	s.cacheMutex.Unlock()
}

// toClusterResponse 转换为响应对象
func (s *ClusterService) toClusterResponse(cluster *models.Cluster) *ClusterDetailResponse {
	if cluster == nil {
		return nil
	}

	// 获取节点数量
	nodeCount := 0
	if cluster.Status == models.ClusterStatusNormal {
		// 只在集群正常时获取节点数
		clientset, err := s.GetCachedClientset(context.Background(), cluster.ID)
		if err == nil {
			nodes, _ := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
			if nodes != nil {
				nodeCount = len(nodes.Items)
			}
		}
	}

	return &ClusterDetailResponse{
		ID:          cluster.ID,
		Name:        cluster.Name,
		Alias:       cluster.Alias,
		APIEndpoint: cluster.APIEndpoint,
		Version:     cluster.Version,
		Status:      cluster.Status,
		NodeCount:   nodeCount,
		Region:      cluster.Region,
		Provider:    cluster.Provider,
		Description: cluster.Description,
		CreatedAt:   cluster.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   cluster.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// GetCachedMetricsClientset 获取缓存的 metrics clientset
func (s *ClusterService) GetCachedMetricsClientset(ctx context.Context, id uint) (*metricsv.Clientset, error) {
	s.cacheMutex.RLock()
	metricsClientset, exists := s.metricsCache[id]
	s.cacheMutex.RUnlock()

	if exists {
		return metricsClientset, nil
	}

	// 缓存不存在，创建新的
	// 先获取集群信息
	cluster, err := s.clusterBiz.GetCluster(ctx, id)
	if err != nil {
		return nil, err
	}

	// 使用 repository 的方法获取 config 和 clientset
	_, config, err := s.clusterBiz.GetRepo().GetClientset(cluster)
	if err != nil {
		return nil, err
	}

	// 创建 metrics clientset
	metricsClientset, err = metricsv.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// 存入缓存
	s.cacheMutex.Lock()
	s.metricsCache[id] = metricsClientset
	s.cacheMutex.Unlock()

	return metricsClientset, nil
}

// GetClusterAPIEndpoint 获取集群的 API Endpoint
func (s *ClusterService) GetClusterAPIEndpoint(ctx context.Context, id uint) (string, error) {
	cluster, err := s.clusterBiz.GetCluster(ctx, id)
	if err != nil {
		return "", err
	}

	// 如果数据库中已存储 API Endpoint，直接返回
	if cluster.APIEndpoint != "" {
		return cluster.APIEndpoint, nil
	}

	// 否则从 KubeConfig 中解析
	kubeConfig, err := biz.DecryptKubeConfig(cluster.KubeConfig)
	if err != nil {
		return "", err
	}

	// 从 KubeConfig 中提取 server 地址
	lines := strings.Split(kubeConfig, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "server:") {
			server := strings.TrimPrefix(line, "server:")
			server = strings.TrimSpace(server)
			// 去除引号
			server = strings.Trim(server, "\"")
			server = strings.Trim(server, "'")
			return server, nil
		}
	}

	return "", nil
}

// GetClusterConfig 获取集群的 KubeConfig（解密后的）
func (s *ClusterService) GetClusterConfig(ctx context.Context, id uint) (string, error) {
	cluster, err := s.clusterBiz.GetCluster(ctx, id)
	if err != nil {
		return "", err
	}

	// 调用 biz 层的解密方法
	kubeConfig, err := biz.DecryptKubeConfig(cluster.KubeConfig)
	if err != nil {
		return "", err
	}
	return kubeConfig, nil
}

// GenerateKubeConfigRequest 生成 KubeConfig 请求
type GenerateKubeConfigRequest struct {
	ClusterID uint   `json:"clusterId" binding:"required"`
	Username string `json:"username" binding:"required"`
}

// GenerateUserKubeConfig 为指定用户生成 KubeConfig
func (s *ClusterService) GenerateUserKubeConfig(ctx context.Context, clusterID uint, username string, userID uint) (string, string, error) {
	// 获取集群信息
	cluster, err := s.clusterBiz.GetCluster(ctx, clusterID)
	if err != nil {
		return "", "", err
	}

	// 获取集群的 clientset
	clientset, err := s.clusterBiz.GetClusterClientset(ctx, clusterID)
	if err != nil {
		return "", "", err
	}

	// 生成 KubeConfig
	kubeConfig, uniqueUsername, err := s.createKubeConfigForUser(clientset, cluster, username)
	if err != nil {
		return "", "", err
	}

	// 保存或更新凭据记录到数据库
	kubeConfigRecord := &model.UserKubeConfig{
		ClusterID:      uint64(clusterID),
		UserID:         uint64(userID),
		ServiceAccount: uniqueUsername,
		Namespace:      "default",
		IsActive:       true,
		CreatedBy:      uint64(userID),
	}

	// 使用 ON DUPLICATE KEY UPDATE 处理重复记录
	err = s.db.Where("cluster_id = ? AND user_id = ?", clusterID, userID).
		Assign(kubeConfigRecord).
		FirstOrCreate(kubeConfigRecord).Error

	if err != nil {
		return "", "", fmt.Errorf("保存凭据记录失败: %w", err)
	}

	return kubeConfig, uniqueUsername, nil
}

// GetUserExistingKubeConfig 获取用户现有的KubeConfig（如果存在）
func (s *ClusterService) GetUserExistingKubeConfig(ctx context.Context, clusterID uint, username string, userID uint) (string, string, error) {
	// 从数据库查询用户在该集群的激活凭据
	var config model.UserKubeConfig
	err := s.db.Where("cluster_id = ? AND user_id = ? AND is_active = 1", clusterID, userID).
		Order("created_at DESC").
		First(&config).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", "", fmt.Errorf("用户尚未申请凭据")
		}
		return "", "", fmt.Errorf("查询凭据失败: %w", err)
	}

	// 获取集群信息
	cluster, err := s.clusterBiz.GetCluster(ctx, clusterID)
	if err != nil {
		return "", "", err
	}

	// 获取集群的 clientset
	clientset, err := s.clusterBiz.GetClusterClientset(ctx, clusterID)
	if err != nil {
		return "", "", err
	}

	// 为现有的ServiceAccount生成kubeconfig
	kubeConfig, err := s.generateKubeConfigForServiceAccount(clientset, cluster, config.ServiceAccount)
	if err != nil {
		return "", "", fmt.Errorf("生成KubeConfig失败: %w", err)
	}

	return kubeConfig, config.ServiceAccount, nil
}

// GenerateKubeConfigForServiceAccount 为现有的ServiceAccount生成kubeconfig（公开方法）
func (s *ClusterService) GenerateKubeConfigForServiceAccount(ctx context.Context, clientset *kubernetes.Clientset, cluster *models.Cluster, saName string) (string, error) {
	return s.generateKubeConfigForServiceAccount(clientset, cluster, saName)
}

// GenerateKubeConfigForSA 为现有的ServiceAccount生成kubeconfig（不需要context的版本）
func (s *ClusterService) GenerateKubeConfigForSA(clientset *kubernetes.Clientset, cluster *models.Cluster, saName string) (string, error) {
	return s.generateKubeConfigForServiceAccount(clientset, cluster, saName)
}

// generateKubeConfigForServiceAccount 为现有的ServiceAccount生成kubeconfig
func (s *ClusterService) generateKubeConfigForServiceAccount(clientset *kubernetes.Clientset, cluster *models.Cluster, saName string) (string, error) {
	// 使用 ServiceAccount 的 TokenRequest API 获取临时 token
	expiration := int64(86400 * 365) // 1年有效期
	tr, err := clientset.CoreV1().ServiceAccounts("default").CreateToken(context.TODO(), saName, &authenticationv1.TokenRequest{
		Spec: authenticationv1.TokenRequestSpec{
			ExpirationSeconds: &expiration,
		},
	}, metav1.CreateOptions{})

	if err != nil {
		// 如果 TokenRequest 失败，尝试查找现有的 Secret
		var secretName string
		secrets, err := clientset.CoreV1().Secrets("default").List(context.TODO(), metav1.ListOptions{})
		if err == nil {
			for _, secret := range secrets.Items {
				if strings.HasPrefix(secret.Name, saName+"-token") {
					secretName = secret.Name
					break
				}
			}
		}

		if secretName == "" {
			return "", fmt.Errorf("获取 Token 失败且未找到现有 Secret: %w", err)
		}

		secret, err := clientset.CoreV1().Secrets("default").Get(context.TODO(), secretName, metav1.GetOptions{})
		if err != nil {
			return "", fmt.Errorf("获取 Secret 失败: %w", err)
		}

		token, ok := secret.Data["token"]
		if !ok {
			return "", fmt.Errorf("Secret 中缺少 Token")
		}

		kubeConfig, err := s.generateKubeConfigContent(clientset, cluster, saName, string(token))
		if err != nil {
			return "", err
		}
		return kubeConfig, nil
	}

	// 使用 TokenRequest 返回的 token
	token := tr.Status.Token
	kubeConfig, err := s.generateKubeConfigContent(clientset, cluster, saName, token)
	if err != nil {
		return "", err
	}
	return kubeConfig, nil
}

// RevokeUserKubeConfig 吊销用户的 KubeConfig 凭据
func (s *ClusterService) RevokeUserKubeConfig(ctx context.Context, clusterID uint, username string) error {
	// 获取集群的 clientset
	clientset, err := s.clusterBiz.GetClusterClientset(ctx, clusterID)
	if err != nil {
		return err
	}

	// 删除 ClusterRoleBinding
	crbName := username + "-binding"
	err = clientset.RbacV1().ClusterRoleBindings().Delete(ctx, crbName, metav1.DeleteOptions{})
	if err != nil {
		// ClusterRoleBinding 可能不存在，继续删除 ServiceAccount
	}

	// 删除 ServiceAccount
	err = clientset.CoreV1().ServiceAccounts("default").Delete(ctx, username, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("删除 ServiceAccount 失败: %w", err)
	}

	return nil
}

// createKubeConfigForUser 创建用户专用的 KubeConfig
func (s *ClusterService) createKubeConfigForUser(clientset *kubernetes.Clientset, cluster *models.Cluster, username string) (string, string, error) {
	// ServiceAccount名称直接使用 opshub-{username}
	saName := fmt.Sprintf("opshub-%s", username)

	// 检查ServiceAccount是否已存在
	_, err := clientset.CoreV1().ServiceAccounts("default").Get(context.TODO(), saName, metav1.GetOptions{})
	if err != nil {
		// ServiceAccount不存在，创建新的
		if k8serrors.IsNotFound(err) {
			sa := &v1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name: saName,
				},
			}
			_, err = clientset.CoreV1().ServiceAccounts("default").Create(context.TODO(), sa, metav1.CreateOptions{})
			if err != nil {
				return "", "", fmt.Errorf("创建 ServiceAccount 失败: %w", err)
			}

			// 只有 admin 用户才创建 cluster-admin 权限的 ClusterRoleBinding
			// 其他用户创建的凭据没有任何权限，需要通过"角色授权"功能来分配
			if username == "admin" {
				rbName := saName + "-binding"
				crb := &rbacv1.ClusterRoleBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name: rbName,
					},
					Subjects: []rbacv1.Subject{
						{
							Kind:      "ServiceAccount",
							Name:      saName,
							Namespace: "default",
						},
					},
					RoleRef: rbacv1.RoleRef{
						Kind:     "ClusterRole",
						Name:     "cluster-admin",
						APIGroup: "rbac.authorization.k8s.io",
					},
				}
				_, err = clientset.RbacV1().ClusterRoleBindings().Create(context.TODO(), crb, metav1.CreateOptions{})
				if err != nil {
					return "", "", fmt.Errorf("创建 ClusterRoleBinding 失败: %w", err)
				}
			}
		} else {
			return "", "", fmt.Errorf("查询 ServiceAccount 失败: %w", err)
		}
	}
	// 如果ServiceAccount已存在，直接使用，不需要重新创建权限绑定

	// 使用 ServiceAccount 的 Token 创建请求
	// 通过创建 TokenRequest API 获取临时 token
	expiration := int64(86400 * 365) // 1年有效期
	tr, err := clientset.CoreV1().ServiceAccounts("default").CreateToken(context.TODO(), saName, &authenticationv1.TokenRequest{
		Spec: authenticationv1.TokenRequestSpec{
			ExpirationSeconds: &expiration,
		},
	}, metav1.CreateOptions{})

	if err != nil {
		// 如果 TokenRequest 失败，尝试查找现有的 Secret
		var secretName string
		secrets, err := clientset.CoreV1().Secrets("default").List(context.TODO(), metav1.ListOptions{})
		if err == nil {
			for _, secret := range secrets.Items {
				if strings.HasPrefix(secret.Name, saName+"-token") {
					secretName = secret.Name
					break
				}
			}
		}

		if secretName == "" {
			return "", "", fmt.Errorf("获取 Token 失败且未找到现有 Secret: %w", err)
		}

		secret, err := clientset.CoreV1().Secrets("default").Get(context.TODO(), secretName, metav1.GetOptions{})
		if err != nil {
			return "", "", fmt.Errorf("获取 Secret 失败: %w", err)
		}

		token, ok := secret.Data["token"]
		if !ok {
			return "", "", fmt.Errorf("Secret 中缺少 Token")
		}

		kubeConfig, err := s.generateKubeConfigContent(clientset, cluster, saName, string(token))
		if err != nil {
			return "", "", err
		}
		return kubeConfig, saName, nil
	}

	// 使用 TokenRequest 返回的 token
	token := tr.Status.Token
	kubeConfig, err := s.generateKubeConfigContent(clientset, cluster, saName, token)
	if err != nil {
		return "", "", err
	}
	return kubeConfig, saName, nil
}

// generateKubeConfigContent 生成 kubeconfig 内容
func (s *ClusterService) generateKubeConfigContent(clientset *kubernetes.Clientset, cluster *models.Cluster, username, token string) (string, error) {
	// 优先从集群获取 CA 证书
	caCert, err := s.getClusterCACert(clientset)
	if err != nil {
		// 如果从集群获取失败，尝试从原始 kubeconfig 中提取
		originalKubeConfig, decryptErr := biz.DecryptKubeConfig(cluster.KubeConfig)
		if decryptErr == nil {
			caCert = extractCAFromKubeconfig(originalKubeConfig)
			if caCert == "" {
				return "", fmt.Errorf("无法从原始 kubeconfig 中提取 CA 证书")
			}
		} else {
			return "", fmt.Errorf("解密集群 kubeconfig 失败: %w", decryptErr)
		}
	}

	// 验证 CA 证书不为空
	if caCert == "" {
		return "", fmt.Errorf("CA 证书为空")
	}

	// 解密集群的原始 kubeconfig 获取 server 地址
	serverURL := cluster.APIEndpoint
	if serverURL == "" {
		originalKubeConfig, err := biz.DecryptKubeConfig(cluster.KubeConfig)
		if err != nil {
			return "", fmt.Errorf("解密集群 kubeconfig 失败: %w", err)
		}

		lines := strings.Split(originalKubeConfig, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "server:") {
				serverURL = strings.TrimPrefix(line, "server:")
				serverURL = strings.TrimSpace(serverURL)
				serverURL = strings.Trim(serverURL, "\"")
				serverURL = strings.Trim(serverURL, "'")
				break
			}
		}
	}

	if serverURL == "" {
		return "", fmt.Errorf("无法确定集群的 API Server 地址")
	}

	// 生成 KubeConfig 内容（使用标准的 kubectl 格式）
	kubeConfigContent := fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
  - name: %s
    cluster:
      certificate-authority-data: %s
      server: %s
contexts:
  - name: %s
    context:
      cluster: %s
      user: %s
current-context: %s
preferences: {}
users:
  - name: %s
    user:
      token: %s
`,
		cluster.Name,
		caCert,
		serverURL,
		cluster.Name+"-context",
		cluster.Name,
		username,
		cluster.Name+"-context",
		username,
		token,
	)

	return kubeConfigContent, nil
}

// extractCAFromKubeconfig 从 kubeconfig 内容中提取 CA 证书
func extractCAFromKubeconfig(kubeconfig string) string {
	lines := strings.Split(kubeconfig, "\n")

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// 查找 certificate-authority-data 字段
		if strings.HasPrefix(trimmedLine, "certificate-authority-data:") {
			// 尝试获取同一行的值
			parts := strings.SplitN(trimmedLine, ":", 2)
			if len(parts) == 2 {
				data := strings.TrimSpace(parts[1])
				// 如果同一行有数据且不是其他字段
				if data != "" && !strings.HasPrefix(data, "server:") && !strings.HasPrefix(data, "client-") {
					return data
				}
			}

			// 尝试获取下一行的值（多行格式）
			if i+1 < len(lines) {
				nextLine := strings.TrimSpace(lines[i+1])
				// 下一行应该是纯数据，不包含冒号（不是新的字段）
				if nextLine != "" && !strings.Contains(nextLine, ":") {
					return nextLine
				}
			}
		}
	}

	// 如果都找不到，返回空字符串
	return ""
}

// getClusterCACert 获取集群的 CA 证书
func (s *ClusterService) getClusterCACert(clientset *kubernetes.Clientset) (string, error) {
	// 尝试从 kube-system 命名空间的 ServiceAccount Secret 获取
	secrets, err := clientset.CoreV1().Secrets("kube-system").List(context.TODO(), metav1.ListOptions{})
	if err == nil {
		for _, secret := range secrets.Items {
			if strings.HasPrefix(secret.Name, "default-token-") || strings.HasPrefix(secret.Name, "coredns-token-") {
				if caCert, ok := secret.Data["ca.crt"]; ok {
					// CA 证书已经是 DER 格式，需要 base64 编码
					// k8s 存储的 ca.crt 已经是 PEM 格式，需要转换为 base64
					return base64.StdEncoding.EncodeToString(caCert), nil
				}
			}
		}
	}

	// 如果 kube-system 中没找到，尝试从所有命名空间查找任何包含 ca.crt 的 Secret
	allSecrets, err := clientset.CoreV1().Secrets("").List(context.TODO(), metav1.ListOptions{})
	if err == nil {
		for _, secret := range allSecrets.Items {
			if caCert, ok := secret.Data["ca.crt"]; ok && len(caCert) > 0 {
				return base64.StdEncoding.EncodeToString(caCert), nil
			}
		}
	}

	// 如果还没找到，尝试从 ConfigMap 获取
	cm, err := clientset.CoreV1().ConfigMaps("kube-public").Get(context.TODO(), "cluster-info", metav1.GetOptions{})
	if err == nil {
		if kubeconfig, ok := cm.Data["kubeconfig"]; ok {
			// 从 kubeconfig 中提取 CA 证书
			lines := strings.Split(kubeconfig, "\n")
			inCA := false
			var caData strings.Builder
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "certificate-authority-data:") {
					inCA = true
					// 尝试获取同一行的数据
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						data := strings.TrimSpace(parts[1])
						if data != "" && !strings.HasPrefix(data, "server:") {
							return data, nil
						}
					}
					continue
				}
				if inCA {
					// 如果遇到新的字段（包含冒号且不是续行），说明 CA 数据结束
					if strings.Contains(line, ":") {
						break
					}
					if line != "" {
						caData.WriteString(line)
					}
				}
			}
			if caData.String() != "" {
				return caData.String(), nil
			}
		}
	}

	return "", fmt.Errorf("无法获取集群 CA 证书")
}

