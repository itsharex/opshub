package server

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ydcloud-dy/opshub/plugins/kubernetes/service"
)

// ResourceHandler Kubernetes资源处理器
type ResourceHandler struct {
	clusterService *service.ClusterService
}

// NewResourceHandler 创建资源处理器
func NewResourceHandler(clusterService *service.ClusterService) *ResourceHandler {
	return &ResourceHandler{
		clusterService: clusterService,
	}
}

// NodeInfo 节点信息
type NodeInfo struct {
	Name            string   `json:"name"`
	Status          string   `json:"status"`
	Roles           string   `json:"roles"`
	Age             string   `json:"age"`
	Version         string   `json:"version"`
	InternalIP      string   `json:"internalIP"`
	OSImage         string   `json:"osImage"`
	KernelVersion   string   `json:"kernelVersion"`
	ContainerRuntime string  `json:"containerRuntime"`
	Labels          map[string]string `json:"labels"`
}

// PodInfo Pod信息
type PodInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Ready     string `json:"ready"`
	Status    string `json:"status"`
	Restarts  int32  `json:"restarts"`
	Age       string `json:"age"`
	IP        string `json:"ip"`
	Node      string `json:"node"`
	Labels    map[string]string `json:"labels"`
}

// NamespaceInfo 命名空间信息
type NamespaceInfo struct {
	Name   string            `json:"name"`
	Status string            `json:"status"`
	Age    string            `json:"age"`
	Labels map[string]string `json:"labels"`
}

// DeploymentInfo Deployment信息
type DeploymentInfo struct {
	Name             string `json:"name"`
	Namespace        string `json:"namespace"`
	Ready            string `json:"ready"`
	UpToDate         int32  `json:"upToDate"`
	Available        int32  `json:"available"`
	Age              string `json:"age"`
	Replicas         int32  `json:"replicas"`
	Selector         map[string]string `json:"selector"`
	Labels           map[string]string `json:"labels"`
}

// DaemonSetInfo DaemonSet信息
type DaemonSetInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Ready     string `json:"ready"`
	Age       string `json:"age"`
	Labels    map[string]string `json:"labels"`
}

// StatefulSetInfo StatefulSet信息
type StatefulSetInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Ready     string `json:"ready"`
	Age       string `json:"age"`
	Labels    map[string]string `json:"labels"`
}

// JobInfo Job信息
type JobInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Ready     string `json:"ready"`
	Age       string `json:"age"`
	Labels    map[string]string `json:"labels"`
}

// ClusterStats 集群统计信息
type ClusterStats struct {
	NodeCount        int     `json:"nodeCount"`
	WorkloadCount    int     `json:"workloadCount"`    // Deployment + DaemonSet + StatefulSet + Job
	PodCount         int     `json:"podCount"`
	CPUUsage         float64 `json:"cpuUsage"`         // CPU使用率百分比
	MemoryUsage      float64 `json:"memoryUsage"`      // 内存使用率百分比
	CPUCapacity      float64 `json:"cpuCapacity"`      // CPU总核数
	MemoryCapacity   float64 `json:"memoryCapacity"`   // 内存总容量(字节)
	CPUAllocatable   float64 `json:"cpuAllocatable"`   // CPU可分配量
	MemoryAllocatable float64 `json:"memoryAllocatable"` // 内存可分配量(字节)
	CPUUsed          float64 `json:"cpuUsed"`          // CPU已使用量
	MemoryUsed       float64 `json:"memoryUsed"`       // 内存已使用量(字节)
}

// ClusterNetworkInfo 集群网络信息
type ClusterNetworkInfo struct {
	ServiceCIDR         string            `json:"serviceCIDR"`         // Service CIDR
	PodCIDR             string            `json:"podCIDR"`             // Pod CIDR
	APIServerAddress    string            `json:"apiServerAddress"`    // API Server 地址
	NetworkPlugin       string            `json:"networkPlugin"`       // 网络插件
	ProxyMode           string            `json:"proxyMode"`           // 服务转发模式
	DNSService          string            `json:"dnsService"`          // DNS 服务
}

// ClusterComponentInfo 集群组件信息
type ClusterComponentInfo struct {
	Components []ComponentInfo `json:"components"`  // 控制平面组件
	Runtime    RuntimeInfo     `json:"runtime"`     // 运行时信息
	Storage    []StorageInfo   `json:"storage"`     // 存储信息
}

// ComponentInfo 组件信息
type ComponentInfo struct {
	Name    string `json:"name"`    // 组件名称
	Version string `json:"version"` // 版本
	Status  string `json:"status"`  // 状态
}

// RuntimeInfo 运行时信息
type RuntimeInfo struct {
	ContainerRuntime string `json:"containerRuntime"` // 容器运行时
	Version          string `json:"version"`          // 版本
}

// StorageInfo 存储信息
type StorageInfo struct {
	Name       string `json:"name"`       // 存储名称
	Provisioner string `json:"provisioner"` // Provisioner
	ReclaimPolicy string `json:"reclaimPolicy"` // 回收策略
}

// ListNodes 获取节点列表
func (h *ResourceHandler) ListNodes(c *gin.Context) {
	clusterIDStr := c.Query("clusterId")
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的集群ID",
		})
		return
	}

	clientset, err := h.clusterService.GetCachedClientset(c.Request.Context(), uint(clusterID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取集群连接失败: " + err.Error(),
		})
		return
	}

	nodes, err := clientset.CoreV1().Nodes().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取节点列表失败: " + err.Error(),
		})
		return
	}

	nodeInfos := make([]NodeInfo, 0, len(nodes.Items))
	for _, node := range nodes.Items {
		nodeInfo := NodeInfo{
			Name:            node.Name,
			Version:         node.Status.NodeInfo.KubeletVersion,
			OSImage:         node.Status.NodeInfo.OSImage,
			KernelVersion:   node.Status.NodeInfo.KernelVersion,
			ContainerRuntime: node.Status.NodeInfo.ContainerRuntimeVersion,
			Labels:          node.Labels,
		}

		// 获取节点状态
		for _, condition := range node.Status.Conditions {
			if condition.Type == v1.NodeReady {
				if condition.Status == v1.ConditionTrue {
					nodeInfo.Status = "Ready"
				} else {
					nodeInfo.Status = "NotReady"
				}
				break
			}
		}

		// 获取InternalIP
		for _, addr := range node.Status.Addresses {
			if addr.Type == v1.NodeInternalIP {
				nodeInfo.InternalIP = addr.Address
				break
			}
		}

		// 计算节点年龄
		nodeInfo.Age = calculateAge(node.CreationTimestamp.Time)

		// 获取角色（从Label中推断）
		if _, ok := node.Labels["node-role.kubernetes.io/master"]; ok {
			nodeInfo.Roles = "master"
		} else if _, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok {
			nodeInfo.Roles = "control-plane"
		} else {
			nodeInfo.Roles = "worker"
		}

		nodeInfos = append(nodeInfos, nodeInfo)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    nodeInfos,
	})
}

// ListNamespaces 获取命名空间列表
func (h *ResourceHandler) ListNamespaces(c *gin.Context) {
	clusterIDStr := c.Query("clusterId")
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的集群ID",
		})
		return
	}

	clientset, err := h.clusterService.GetCachedClientset(c.Request.Context(), uint(clusterID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取集群连接失败: " + err.Error(),
		})
		return
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取命名空间列表失败: " + err.Error(),
		})
		return
	}

	namespaceInfos := make([]NamespaceInfo, 0, len(namespaces.Items))
	for _, ns := range namespaces.Items {
		nsInfo := NamespaceInfo{
			Name:   ns.Name,
			Labels: ns.Labels,
			Age:    calculateAge(ns.CreationTimestamp.Time),
		}

		// 获取状态
		if ns.Status.Phase == v1.NamespaceActive {
			nsInfo.Status = "Active"
		} else {
			nsInfo.Status = string(ns.Status.Phase)
		}

		namespaceInfos = append(namespaceInfos, nsInfo)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    namespaceInfos,
	})
}

// ListPods 获取Pod列表
func (h *ResourceHandler) ListPods(c *gin.Context) {
	clusterIDStr := c.Query("clusterId")
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的集群ID",
		})
		return
	}

	namespace := c.Query("namespace")
	if namespace == "" {
		namespace = v1.NamespaceAll
	}

	clientset, err := h.clusterService.GetCachedClientset(c.Request.Context(), uint(clusterID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取集群连接失败: " + err.Error(),
		})
		return
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取Pod列表失败: " + err.Error(),
		})
		return
	}

	podInfos := make([]PodInfo, 0, len(pods.Items))
	for _, pod := range pods.Items {
		podInfo := PodInfo{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Labels:    pod.Labels,
			Age:       calculateAge(pod.CreationTimestamp.Time),
			IP:        pod.Status.PodIP,
			Node:      pod.Spec.NodeName,
		}

		// 计算Ready状态
		readyContainers := 0
		totalContainers := len(pod.Spec.Containers)
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.Ready {
				readyContainers++
			}
			podInfo.Restarts += cs.RestartCount
		}
		podInfo.Ready = strconv.Itoa(readyContainers) + "/" + strconv.Itoa(totalContainers)

		// 获取Pod状态
		podInfo.Status = string(pod.Status.Phase)

		podInfos = append(podInfos, podInfo)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    podInfos,
	})
}

// ListDeployments 获取Deployment列表
func (h *ResourceHandler) ListDeployments(c *gin.Context) {
	clusterIDStr := c.Query("clusterId")
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的集群ID",
		})
		return
	}

	namespace := c.Query("namespace")
	if namespace == "" {
		namespace = v1.NamespaceAll
	}

	clientset, err := h.clusterService.GetCachedClientset(c.Request.Context(), uint(clusterID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取集群连接失败: " + err.Error(),
		})
		return
	}

	deployments, err := clientset.AppsV1().Deployments(namespace).List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取Deployment列表失败: " + err.Error(),
		})
		return
	}

	deploymentInfos := make([]DeploymentInfo, 0, len(deployments.Items))
	for _, deploy := range deployments.Items {
		deployInfo := DeploymentInfo{
			Name:      deploy.Name,
			Namespace: deploy.Namespace,
			UpToDate:  deploy.Status.UpdatedReplicas,
			Available: deploy.Status.AvailableReplicas,
			Age:       calculateAge(deploy.CreationTimestamp.Time),
			Replicas:  *deploy.Spec.Replicas,
			Selector:  deploy.Spec.Selector.MatchLabels,
			Labels:    deploy.Labels,
		}

		// 计算Ready状态
		readyReplicas := deploy.Status.ReadyReplicas
		totalReplicas := *deploy.Spec.Replicas
		deployInfo.Ready = strconv.Itoa(int(readyReplicas)) + "/" + strconv.Itoa(int(totalReplicas))

		deploymentInfos = append(deploymentInfos, deployInfo)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    deploymentInfos,
	})
}

// GetClusterStats 获取集群统计信息
func (h *ResourceHandler) GetClusterStats(c *gin.Context) {
	clusterIDStr := c.Query("clusterId")
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的集群ID",
		})
		return
	}

	clientset, err := h.clusterService.GetCachedClientset(c.Request.Context(), uint(clusterID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取集群连接失败: " + err.Error(),
		})
		return
	}

	// 获取 metrics clientset
	metricsClient, err := h.clusterService.GetCachedMetricsClientset(c.Request.Context(), uint(clusterID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取 metrics client 失败: " + err.Error(),
		})
		return
	}

	stats := ClusterStats{}

	// 获取节点信息
	nodes, err := clientset.CoreV1().Nodes().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取节点列表失败: " + err.Error(),
		})
		return
	}
	stats.NodeCount = len(nodes.Items)

	// 计算CPU和内存总量及可分配量
	var totalCPUCapacity, totalMemoryCapacity float64
	var totalCPUAllocatable, totalMemoryAllocatable float64

	for _, node := range nodes.Items {
		cpuCapacity := node.Status.Capacity.Cpu().AsApproximateFloat64()
		memoryCapacity := float64(node.Status.Capacity.Memory().Value())
		cpuAllocatable := node.Status.Allocatable.Cpu().AsApproximateFloat64()
		memoryAllocatable := float64(node.Status.Allocatable.Memory().Value())

		totalCPUCapacity += cpuCapacity
		totalMemoryCapacity += memoryCapacity
		totalCPUAllocatable += cpuAllocatable
		totalMemoryAllocatable += memoryAllocatable
	}

	stats.CPUCapacity = totalCPUCapacity
	stats.MemoryCapacity = totalMemoryCapacity
	stats.CPUAllocatable = totalCPUAllocatable
	stats.MemoryAllocatable = totalMemoryAllocatable

	// 获取节点指标（Metrics API）
	nodeMetrics, err := metricsClient.MetricsV1beta1().NodeMetricses().List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取节点指标失败: " + err.Error(),
		})
		return
	}

	// 计算实际使用的CPU和内存
	var totalCPUUsed, totalMemoryUsed float64
	for _, nodeMetric := range nodeMetrics.Items {
		cpuUsed := nodeMetric.Usage.Cpu().AsApproximateFloat64()
		memoryUsed := float64(nodeMetric.Usage.Memory().Value())
		totalCPUUsed += cpuUsed
		totalMemoryUsed += memoryUsed
	}

	// 设置已使用量
	stats.CPUUsed = totalCPUUsed
	stats.MemoryUsed = totalMemoryUsed

	// 计算使用率百分比（基于 Allocatable）
	if totalCPUAllocatable > 0 {
		stats.CPUUsage = (totalCPUUsed / totalCPUAllocatable) * 100
	}
	if totalMemoryAllocatable > 0 {
		stats.MemoryUsage = (totalMemoryUsed / totalMemoryAllocatable) * 100
	}

	// 获取Pod数量
	pods, err := clientset.CoreV1().Pods("").List(c.Request.Context(), metav1.ListOptions{})
	if err == nil {
		stats.PodCount = len(pods.Items)
	}

	// 获取Deployment数量
	deployments, err := clientset.AppsV1().Deployments("").List(c.Request.Context(), metav1.ListOptions{})
	deploymentCount := 0
	if err == nil {
		deploymentCount = len(deployments.Items)
	}

	// 获取DaemonSet数量
	daemonsets, err := clientset.AppsV1().DaemonSets("").List(c.Request.Context(), metav1.ListOptions{})
	daemonsetCount := 0
	if err == nil {
		daemonsetCount = len(daemonsets.Items)
	}

	// 获取StatefulSet数量
	statefulsets, err := clientset.AppsV1().StatefulSets("").List(c.Request.Context(), metav1.ListOptions{})
	statefulsetCount := 0
	if err == nil {
		statefulsetCount = len(statefulsets.Items)
	}

	// 获取Job数量
	jobs, err := clientset.BatchV1().Jobs("").List(c.Request.Context(), metav1.ListOptions{})
	jobCount := 0
	if err == nil {
		jobCount = len(jobs.Items)
	}

	// 工作负载总数 = Deployment + DaemonSet + StatefulSet + Job
	stats.WorkloadCount = deploymentCount + daemonsetCount + statefulsetCount + jobCount

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    stats,
	})
}

// GetClusterNetworkInfo 获取集群网络信息
func (h *ResourceHandler) GetClusterNetworkInfo(c *gin.Context) {
	clusterIDStr := c.Query("clusterId")
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的集群ID",
		})
		return
	}

	clientset, err := h.clusterService.GetCachedClientset(c.Request.Context(), uint(clusterID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取集群连接失败: " + err.Error(),
		})
		return
	}

	networkInfo := ClusterNetworkInfo{}

	// 获取集群的 API Endpoint
	apiEndpoint, err := h.clusterService.GetClusterAPIEndpoint(c.Request.Context(), uint(clusterID))
	if err == nil && apiEndpoint != "" {
		networkInfo.APIServerAddress = apiEndpoint
	}

	// 获取节点信息来推断网络配置
	nodes, err := clientset.CoreV1().Nodes().List(c.Request.Context(), metav1.ListOptions{})
	if err == nil && len(nodes.Items) > 0 {
		node := nodes.Items[0]

		// 获取 Pod CIDR
		if podCIDR := node.Spec.PodCIDR; podCIDR != "" {
			networkInfo.PodCIDR = podCIDR
		}
	}

	// 获取 CNI 网络插件（从 kube-system 命名空间的 DaemonSet 中检测）
	daemonSets, err := clientset.AppsV1().DaemonSets("kube-system").List(c.Request.Context(), metav1.ListOptions{})
	if err == nil {
		// 常见的 CNI 插件标识
		cniPlugins := map[string]string{
			"calico":           "Calico",
			"flannel":          "Flannel",
			"weave":            "Weave",
			"canal":            "Canal",
			"cilium":           "Cilium",
			"contiv":           "Contiv",
			"kube-router":      "Kube-Router",
			"amazon-vpc-cni":   "AWS VPC CNI",
			"azure-cniplugin":   "Azure CNI",
			"vsphere-cni":      "vSphere CNI",
			"tke-cni":          "TKE CNI",
			"tke-bridge":       "TKE Bridge",
			"networkpolicy":    "TKE NetworkPolicy",
		}

		for _, ds := range daemonSets.Items {
			dsName := strings.ToLower(ds.Name)
			for key, name := range cniPlugins {
				if strings.Contains(dsName, key) {
					networkInfo.NetworkPlugin = name
					break
				}
			}
			if networkInfo.NetworkPlugin != "" {
				break
			}
		}
	}

	// 获取 kube-proxy 的 proxy 模式（从 DaemonSet 的命令行参数、环境变量或 ConfigMap 中获取）
	kubeProxyDS, err := clientset.AppsV1().DaemonSets("kube-system").Get(c.Request.Context(), "kube-proxy", metav1.GetOptions{})
	if err == nil && len(kubeProxyDS.Spec.Template.Spec.Containers) > 0 {
		container := kubeProxyDS.Spec.Template.Spec.Containers[0]

		// 1. 从命令行参数中查找（优先级最高）
		for _, arg := range container.Command {
			if strings.Contains(arg, "--proxy-mode=") {
				mode := strings.TrimPrefix(arg, "--proxy-mode=")
				networkInfo.ProxyMode = mode
				break
			}
		}

		// 2. 从命令行参数中查找（空格分隔）
		if networkInfo.ProxyMode == "" && len(container.Command) > 0 {
			for i, arg := range container.Command {
				if arg == "--proxy-mode" && i+1 < len(container.Command) {
					networkInfo.ProxyMode = container.Command[i+1]
					break
				}
			}
		}

		// 3. 从环境变量中查找
		if networkInfo.ProxyMode == "" {
			for _, env := range container.Env {
				if env.Name == "KUBE_PROXY_MODE" {
					networkInfo.ProxyMode = env.Value
					break
				}
			}
		}
	}

	// 如果没找到，从 ConfigMap 中查找
	if networkInfo.ProxyMode == "" {
		kubeProxyCM, err := clientset.CoreV1().ConfigMaps("kube-system").Get(c.Request.Context(), "kube-proxy", metav1.GetOptions{})
		if err == nil {
			// 检查 config.yaml
			if config, ok := kubeProxyCM.Data["config.yaml"]; ok {
				// 查找 proxyMode
				if idx := strings.Index(config, "proxyMode:"); idx >= 0 {
					start := idx + 10 // 跳过 "proxyMode:"
					remaining := config[start:]
					// 提取到行尾或注释
					if end := strings.IndexAny(remaining, "\n#"); end > 0 {
						modeStr := strings.TrimSpace(remaining[:end])
						modeStr = strings.Trim(modeStr, `"`)
						networkInfo.ProxyMode = modeStr
					}
				}
			}
			// 检查 config.conf (Kubernetes 1.10+ 使用这个格式)
			if config, ok := kubeProxyCM.Data["config.conf"]; ok {
				if idx := strings.Index(config, "proxyMode"); idx >= 0 {
					start := idx + 10 // 跳过 "proxyMode" 或 "proxyMode:"
					remaining := config[start:]
					// 跳过可能的冒号和等号
					remaining = strings.TrimLeft(remaining, ":=")
					remaining = strings.TrimSpace(remaining)
					// 提取值到行尾或逗号
					if end := strings.IndexAny(remaining, "\n,"); end > 0 {
						modeStr := strings.TrimSpace(remaining[:end])
						modeStr = strings.Trim(modeStr, `"`)
						networkInfo.ProxyMode = modeStr
					}
				}
			}
			if configJSON, ok := kubeProxyCM.Data["config.json"]; ok {
				// JSON 格式配置
				if idx := strings.Index(configJSON, "proxyMode"); idx > 0 {
					start := idx + 11 // 跳过 "proxyMode:"
					if end := strings.Index(configJSON[start:], ","); end > 0 {
						modeStr := strings.TrimSpace(configJSON[start : start+end])
						modeStr = strings.Trim(modeStr, `"`)
						networkInfo.ProxyMode = modeStr
					}
				}
			}
		}
	}

	// 默认值为 ipvs（现代 Kubernetes 的默认模式）
	if networkInfo.ProxyMode == "" {
		// 尝试从节点信息推断（不是100%准确）
		nodes, err := clientset.CoreV1().Nodes().List(c.Request.Context(), metav1.ListOptions{})
		if err == nil && len(nodes.Items) > 0 {
			// 检查内核模块或系统信息来判断
			// 但这比较复杂，这里简单使用默认值
			networkInfo.ProxyMode = "ipvs"
		}
	}


	// 获取 kube-apiserver 服务
	apiServerSvc, err := clientset.CoreV1().Services("default").Get(c.Request.Context(), "kubernetes", metav1.GetOptions{})
	if err == nil && apiServerSvc != nil {
		// 获取 Service CIDR (从 ClusterIPs 推断)
		if len(apiServerSvc.Spec.ClusterIPs) > 0 {
			// 通常是第一个 IP，但我们可以推断 CIDR
			// 例如：10.0.0.1 可能是 10.0.0.0/24 或 10.0.0.0/16
			ip := apiServerSvc.Spec.ClusterIPs[0]
			// 简化处理，直接显示第一个 ClusterIP
			networkInfo.ServiceCIDR = ip
		}
	}

	// 获取 DNS 服务
	_, err = clientset.CoreV1().Services("kube-system").Get(c.Request.Context(), "kube-dns", metav1.GetOptions{})
	if err == nil {
		networkInfo.DNSService = "CoreDNS"
	} else {
		// 尝试获取其他 DNS 实现
		svcs, _ := clientset.CoreV1().Services("kube-system").List(c.Request.Context(), metav1.ListOptions{})
		for _, svc := range svcs.Items {
			if strings.Contains(svc.Name, "dns") {
				networkInfo.DNSService = svc.Name
				break
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    networkInfo,
	})
}

// GetClusterComponentInfo 获取集群组件信息
func (h *ResourceHandler) GetClusterComponentInfo(c *gin.Context) {
	clusterIDStr := c.Query("clusterId")
	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的集群ID",
		})
		return
	}

	clientset, err := h.clusterService.GetCachedClientset(c.Request.Context(), uint(clusterID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取集群连接失败: " + err.Error(),
		})
		return
	}

	componentInfo := ClusterComponentInfo{
		Components: []ComponentInfo{},
	}

	// 获取节点信息来获取运行时
	nodes, err := clientset.CoreV1().Nodes().List(c.Request.Context(), metav1.ListOptions{})
	if err == nil && len(nodes.Items) > 0 {
		node := nodes.Items[0]
		componentInfo.Runtime = RuntimeInfo{
			ContainerRuntime: node.Status.NodeInfo.ContainerRuntimeVersion,
			Version:          node.Status.NodeInfo.KubeletVersion,
		}
	}

	// 获取控制平面组件 Pod
	pods, err := clientset.CoreV1().Pods("kube-system").List(c.Request.Context(), metav1.ListOptions{})
	if err == nil {
		// 常见的控制平面组件（支持多种命名方式）
		controlPlanePatterns := map[string]string{
			"kube-apiserver":          "API Server",
			"kube-apiserver-":         "API Server",
			"apiserver":                "API Server",
			"kube-controller":          "Controller Manager",
			"kube-controller-":         "Controller Manager",
			"kube-controller-manager":  "Controller Manager",
			"cloud-controller":         "Cloud Controller",
			"cloud-controller-":        "Cloud Controller",
			"kube-scheduler":           "Scheduler",
			"kube-scheduler-":          "Scheduler",
			"scheduler":                "Scheduler",
			"etcd":                     "etcd",
			"etcd-":                    "etcd",
			"coredns":                  "CoreDNS",
			"coredns-":                 "CoreDNS",
		}

		componentMap := make(map[string]ComponentInfo)

		// 调试日志：打印 kube-system 命名空间下的所有 Pod
		log.Printf("[GetComponentInfo] kube-system namespace has %d pods", len(pods.Items))
		for _, pod := range pods.Items {
			log.Printf("[GetComponentInfo] Found pod: %s (OwnerReferences: %d)",
				pod.Name, len(pod.OwnerReferences))
		}

		for _, pod := range pods.Items {
			podName := strings.ToLower(pod.Name)
			var componentName string
			var componentKey string

			// 排除非控制平面组件（CNI、网络插件等）
			if strings.Contains(podName, "calico") ||
				strings.Contains(podName, "flannel") ||
				strings.Contains(podName, "kube-proxy") ||
				strings.Contains(podName, "metrics-server") {
				continue
			}

			// 识别组件（支持前缀匹配和包含匹配）
			for pattern, name := range controlPlanePatterns {
				matched := false
				if strings.HasSuffix(pattern, "-") {
					// 前缀匹配模式
					matched = strings.HasPrefix(podName, pattern)
				} else {
					// 精确匹配或包含匹配
					matched = strings.HasPrefix(podName, pattern) ||
						strings.Contains(podName, pattern)
				}

				if matched {
					// 再次检查，确保不是 CNI 组件
					if strings.Contains(podName, "calico") || strings.Contains(podName, "controllers") {
						if !strings.HasPrefix(podName, "kube-controller") {
							continue
						}
					}

					// 使用更具体的key避免重复
					if pattern == "kube-apiserver" || pattern == "kube-apiserver-" {
						componentKey = "kube-apiserver"
					} else if pattern == "kube-controller" || pattern == "kube-controller-" || pattern == "kube-controller-manager" {
						componentKey = "kube-controller"
					} else if pattern == "kube-scheduler" || pattern == "kube-scheduler-" {
						componentKey = "kube-scheduler"
					} else if strings.HasPrefix(pattern, "etcd") {
						componentKey = "etcd"
					} else if strings.HasPrefix(pattern, "coredns") {
						componentKey = "coredns"
					} else if strings.HasPrefix(pattern, "cloud-controller") {
						componentKey = "cloud-controller"
					}
					componentName = name
					log.Printf("[GetComponentInfo] Matched pod %s to component %s (pattern: %s)",
						pod.Name, componentName, pattern)
					break
				}
			}

			if componentName == "" {
				continue
			}

			// 获取版本
			version := "unknown"
			if len(pod.Spec.Containers) > 0 {
				// 尝试从 Image 中提取版本
				image := pod.Spec.Containers[0].Image
				if idx := strings.LastIndex(image, ":"); idx > 0 {
					version = image[idx+1:]
				} else {
					version = image
				}
			}

			// 获取状态
			status := "Running"
			if pod.Status.Phase != v1.PodRunning {
				status = string(pod.Status.Phase)
			}

			componentMap[componentKey] = ComponentInfo{
				Name:    componentName,
				Version: version,
				Status:  status,
			}
			log.Printf("[GetComponentInfo] Added component: %s (version: %s, status: %s)",
				componentName, version, status)
		}

		// 转换为切片
		for _, comp := range componentMap {
			componentInfo.Components = append(componentInfo.Components, comp)
		}
		log.Printf("[GetComponentInfo] Total components found from pods: %d", len(componentInfo.Components))
	} else {
		log.Printf("[GetComponentInfo] Failed to list pods in kube-system: %v", err)
	}

	// 如果没有检测到控制平面组件，可能是二进制部署的集群（systemd 启动）
	// 尝试通过节点标签和版本信息来推断
	log.Printf("[GetComponentInfo] Checking for binary deployment cluster...")

	// 检查是否已经有控制平面组件（API Server, Scheduler, Controller Manager, etcd）
	hasControlPlanePods := false
	for _, comp := range componentInfo.Components {
		if comp.Name == "API Server" || comp.Name == "Scheduler" ||
			comp.Name == "Controller Manager" || comp.Name == "etcd" {
			hasControlPlanePods = true
			break
		}
	}

	if !hasControlPlanePods {
		log.Printf("[GetComponentInfo] No control plane pods found, checking for binary deployment...")

		// 获取集群版本信息
		serverVersion, err := clientset.Discovery().ServerVersion()
		if err == nil {
			k8sVersion := serverVersion.GitVersion
			log.Printf("[GetComponentInfo] Kubernetes version: %s", k8sVersion)

			// 获取所有节点
			nodes, err := clientset.CoreV1().Nodes().List(c.Request.Context(), metav1.ListOptions{})
			if err == nil {
				hasControlPlaneNode := false
				for _, node := range nodes.Items {
					nodeName := strings.ToLower(node.Name)
					log.Printf("[GetComponentInfo] Checking node: %s", node.Name)

					// 检查节点是否是 master/control-plane 节点
					if _, hasControlPlane := node.Labels["node-role.kubernetes.io/control-plane"]; hasControlPlane {
						log.Printf("[GetComponentInfo] Found control-plane node by label: %s", node.Name)
						hasControlPlaneNode = true
						break
					}
					// 兼容旧的标签
					if _, hasMaster := node.Labels["node-role.kubernetes.io/master"]; hasMaster {
						log.Printf("[GetComponentInfo] Found master node by label: %s", node.Name)
						hasControlPlaneNode = true
						break
					}

					// 如果节点名称包含 master/control-plane/mgr 等关键词，也认为是控制平面节点
					if strings.Contains(nodeName, "master") ||
					   strings.Contains(nodeName, "control-plane") ||
					   strings.Contains(nodeName, "control") ||
					   strings.Contains(nodeName, "mgr") {
						log.Printf("[GetComponentInfo] Found control-plane node by name pattern: %s", node.Name)
						hasControlPlaneNode = true
						break
					}
				}

				// 如果检测到控制平面节点但没有找到控制平面 Pod，说明是二进制部署
				if hasControlPlaneNode {
					log.Printf("[GetComponentInfo] Detected binary deployment cluster, adding components...")

					// 添加 API Server
					componentInfo.Components = append(componentInfo.Components, ComponentInfo{
						Name:    "API Server",
						Version: serverVersion.GitVersion,
						Status:  "Running",
					})

					// 添加 Scheduler
					componentInfo.Components = append(componentInfo.Components, ComponentInfo{
						Name:    "Scheduler",
						Version: serverVersion.GitVersion,
						Status:  "Running",
					})

					// 添加 Controller Manager
					componentInfo.Components = append(componentInfo.Components, ComponentInfo{
						Name:    "Controller Manager",
						Version: serverVersion.GitVersion,
						Status:  "Running",
					})

					// 添加 etcd（版本未知）
					componentInfo.Components = append(componentInfo.Components, ComponentInfo{
						Name:    "etcd",
						Version: "unknown",
						Status:  "Running",
					})

					log.Printf("[GetComponentInfo] Added 4 control plane components for binary deployment")
				} else {
					log.Printf("[GetComponentInfo] No control-plane node found, skipping binary deployment detection")
				}
			}
		}
	}

	// 获取存储类
	storageClasses, err := clientset.StorageV1().StorageClasses().List(c.Request.Context(), metav1.ListOptions{})
	if err == nil {
		for _, sc := range storageClasses.Items {
			componentInfo.Storage = append(componentInfo.Storage, StorageInfo{
				Name:           sc.Name,
				Provisioner:    sc.Provisioner,
				ReclaimPolicy:  string(*sc.ReclaimPolicy),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    componentInfo,
	})
}
