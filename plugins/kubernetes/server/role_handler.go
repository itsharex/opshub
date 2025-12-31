package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"gorm.io/gorm"

	"github.com/ydcloud-dy/opshub/plugins/kubernetes/service"
)

// RoleHandler 角色处理器
type RoleHandler struct {
	clusterService *service.ClusterService
	db             *gorm.DB
}

// NewRoleHandler 创建角色处理器
func NewRoleHandler(db *gorm.DB) *RoleHandler {
	return &RoleHandler{
		clusterService: service.NewClusterService(db),
		db:             db,
	}
}

// ListClusterRoles 获取集群角色列表
// @Summary 获取集群角色列表
// @Description 获取所有 Kubernetes 集群级别的角色
// @Tags Kubernetes/Role
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Success 200 {object} Response
// @Router /api/v1/plugins/kubernetes/roles/cluster [get]
func (h *RoleHandler) ListClusterRoles(c *gin.Context) {
	clusterIdStr := c.Query("clusterId")
	if clusterIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "缺少集群ID参数",
		})
		return
	}

	clusterId, err := strconv.ParseUint(clusterIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的集群ID",
		})
		return
	}

	// 获取集群的 clientset
	clientset, err := h.clusterService.GetCachedClientset(c.Request.Context(), uint(clusterId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取集群连接失败",
		})
		return
	}

	// 获取所有集群角色
	roles, err := clientset.RbacV1().ClusterRoles().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取集群角色失败: " + err.Error(),
		})
		return
	}

	// 转换为前端格式
	roleList := make([]map[string]interface{}, 0)
	for _, role := range roles.Items {
		roleList = append(roleList, convertClusterRole(role))
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    roleList,
	})
}

// ListNamespaces 获取命名空间列表
// @Summary 获取命名空间列表
// @Description 获取所有命名空间
// @Tags Kubernetes/Role
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Success 200 {object} Response
// @Router /api/v1/plugins/kubernetes/roles/namespaces [get]
func (h *RoleHandler) ListNamespaces(c *gin.Context) {
	clusterIdStr := c.Query("clusterId")
	if clusterIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "缺少集群ID参数",
		})
		return
	}

	clusterId, err := strconv.ParseUint(clusterIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的集群ID",
		})
		return
	}

	// 获取集群的 clientset
	clientset, err := h.clusterService.GetCachedClientset(c.Request.Context(), uint(clusterId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取集群连接失败",
		})
		return
	}

	// 获取所有命名空间
	namespaces, err := clientset.CoreV1().Namespaces().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取命名空间失败: " + err.Error(),
		})
		return
	}

	// 转换为前端格式
	nsList := make([]map[string]interface{}, 0)
	for _, ns := range namespaces.Items {
		// 获取每个命名空间的 pod 数量（可选）
		pods, _ := clientset.CoreV1().Pods(ns.Name).List(c.Request.Context(), metav1.ListOptions{})
		podCount := 0
		if pods != nil {
			podCount = len(pods.Items)
		}

		nsList = append(nsList, map[string]interface{}{
			"name":      ns.Name,
			"podCount":  podCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    nsList,
	})
}

// ListNamespaceRoles 获取命名空间角色列表
// @Summary 获取命名空间角色列表
// @Description 获取指定命名空间的所有角色
// @Tags Kubernetes/Role
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace query string true "命名空间"
// @Success 200 {object} Response
// @Router /api/v1/plugins/kubernetes/roles/namespace [get]
func (h *RoleHandler) ListNamespaceRoles(c *gin.Context) {
	clusterIdStr := c.Query("clusterId")
	namespace := c.Query("namespace")

	if clusterIdStr == "" || namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "缺少必需参数",
		})
		return
	}

	clusterId, err := strconv.ParseUint(clusterIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的集群ID",
		})
		return
	}

	// 获取集群的 clientset
	clientset, err := h.clusterService.GetCachedClientset(c.Request.Context(), uint(clusterId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取集群连接失败",
		})
		return
	}

	// 获取指定命名空间的角色
	roles, err := clientset.RbacV1().Roles(namespace).List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取命名空间角色失败: " + err.Error(),
		})
		return
	}

	// 转换为前端格式
	roleList := make([]map[string]interface{}, 0)
	for _, role := range roles.Items {
		roleList = append(roleList, convertNamespaceRole(role, namespace))
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
			"data":    roleList,
	})
}

// GetRoleDetail 获取角色详情
// @Summary 获取角色详情
// @Description 获取角色的详细信息，包括权限规则
// @Tags Kubernetes/Role
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace path string true "命名空间"
// @param name path string true "角色名"
// @Success 200 {object} Response
// @Router /api/v1/plugins/kubernetes/roles/{namespace}/{name} [get]
func (h *RoleHandler) GetRoleDetail(c *gin.Context) {
	clusterIdStr := c.Query("clusterId")
	namespace := c.Param("namespace")
	name := c.Param("name")

	if clusterIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "缺少集群ID参数",
		})
		return
	}

	clusterId, err := strconv.ParseUint(clusterIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的集群ID",
		})
		return
	}

	// 获取集群的 clientset
	clientset, err := h.clusterService.GetCachedClientset(c.Request.Context(), uint(clusterId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取集群连接失败",
		})
		return
	}

	// 获取角色详情
	var detail map[string]interface{}
	if namespace == "" || namespace == "cluster" {
		// 集群角色
		clusterRole, err := clientset.RbacV1().ClusterRoles().Get(c.Request.Context(), name, metav1.GetOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取集群角色失败: " + err.Error(),
			})
			return
		}
		detail = convertClusterRoleDetail(*clusterRole)
	} else {
		// 命名空间角色
		nsRole, err := clientset.RbacV1().Roles(namespace).Get(c.Request.Context(), name, metav1.GetOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取命名空间角色失败: " + err.Error(),
			})
			return
		}
		detail = convertNamespaceRoleDetail(*nsRole, namespace)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    detail,
	})
}

// DeleteRole 删除角色
// @Summary 删除角色
// @Description 删除指定的角色
// @Tags Kubernetes/Role
// @Accept json
// @Produce json
// @Param clusterId query int true "集群ID"
// @Param namespace path string true "命名空间"
// @param name path string true "角色名"
// @Success 200 {object} Response
// @Router /api/v1/plugins/kubernetes/roles/{namespace}/{name} [delete]
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	clusterIdStr := c.Query("clusterId")
	namespace := c.Param("namespace")
	name := c.Param("name")

	if clusterIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "缺少集群ID参数",
		})
		return
	}

	clusterId, err := strconv.ParseUint(clusterIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的集群ID",
		})
		return
	}

	// 获取集群的 clientset
	clientset, err := h.clusterService.GetCachedClientset(c.Request.Context(), uint(clusterId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取集群连接失败",
		})
		return
	}

	// 删除角色
	if namespace == "" {
		// 删除集群角色
		err = clientset.RbacV1().ClusterRoles().Delete(c.Request.Context(), name, metav1.DeleteOptions{})
	} else {
		// 删除命名空间角色
		err = clientset.RbacV1().Roles(namespace).Delete(c.Request.Context(), name, metav1.DeleteOptions{})
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除角色失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除成功",
	})
}

// convertClusterRole 转换集群角色为前端格式
func convertClusterRole(role rbacv1.ClusterRole) map[string]interface{} {
	// 计算 age
	age := role.CreationTimestamp.Format("2006-01-02 15:04:05")

	// 转换 labels
	labels := make(map[string]string)
	for key, value := range role.Labels {
		labels[key] = value
	}

	return map[string]interface{}{
		"name":      role.Name,
		"namespace": "",
		"labels":    labels,
		"age":       age,
		"rules":     role.Rules,
	}
}

// convertNamespaceRole 转换命名空间角色为前端格式
func convertNamespaceRole(role rbacv1.Role, namespace string) map[string]interface{} {
	// 计算 age
	age := role.CreationTimestamp.Format("2006-01-02 15:04:05")

	// 转换 labels
	labels := make(map[string]string)
	for key, value := range role.Labels {
		labels[key] = value
	}

	return map[string]interface{}{
		"name":      role.Name,
		"namespace": namespace,
		"labels":    labels,
		"age":       age,
		"rules":     role.Rules,
	}
}

// convertClusterRoleDetail 转换集群角色详情
func convertClusterRoleDetail(role rbacv1.ClusterRole) map[string]interface{} {
	detail := convertClusterRole(role)

	// 添加权限规则的详细信息
	rules := make([]map[string]interface{}, 0)
	for _, rule := range role.Rules {
		ruleDetail := map[string]interface{}{
			"apiGroups":        rule.APIGroups,
			"resources":        rule.Resources,
			"verbs":            rule.Verbs,
		}

		if len(rule.ResourceNames) > 0 {
			ruleDetail["resourceNames"] = rule.ResourceNames
		}

		if len(rule.NonResourceURLs) > 0 {
			ruleDetail["nonResourceURLs"] = rule.NonResourceURLs
		}

		rules = append(rules, ruleDetail)
	}

	detail["rules"] = rules

	return detail
}

// convertNamespaceRoleDetail 转换命名空间角色详情
func convertNamespaceRoleDetail(role rbacv1.Role, namespace string) map[string]interface{} {
	detail := convertNamespaceRole(role, namespace)

	// 添加权限规则的详细信息
	rules := make([]map[string]interface{}, 0)
	for _, rule := range role.Rules {
		ruleDetail := map[string]interface{}{
			"apiGroups":        rule.APIGroups,
			"resources":        rule.Resources,
			"verbs":            rule.Verbs,
		}

		if len(rule.ResourceNames) > 0 {
			ruleDetail["resourceNames"] = rule.ResourceNames
		}

		if len(rule.NonResourceURLs) > 0 {
			ruleDetail["nonResourceURLs"] = rule.NonResourceURLs
		}

		rules = append(rules, ruleDetail)
	}

	detail["rules"] = rules

	return detail
}
