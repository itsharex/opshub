# Kubernetes 集群管理 - Client-Go 集成

## 功能特性

### 已实现功能

1. **集群管理**
   - ✅ 集群注册（支持 KubeConfig 和 Token 两种认证方式）
   - ✅ 集群列表查询
   - ✅ 集群连接测试
   - ✅ 集群删除
   - ✅ KubeConfig 加密存储（AES-GCM）

2. **资源管理**（使用 client-go 真实连接 K8S 集群）
   - ✅ 节点（Nodes）列表查询
   - ✅ 命名空间（Namespaces）列表查询
   - ✅ Pod 列表查询
   - ✅ Deployment 列表查询
   - ✅ Clientset 缓存（提升性能）

## 架构设计

### 后端架构

```
plugins/kubernetes/
├── plugin.go                    # 插件入口
├── biz/cluster.go              # 业务逻辑层（加密/解密）
├── data/
│   ├── models/cluster.go       # 数据模型
│   └── repository/cluster_repository.go  # 数据访问层
├── service/cluster_service.go  # 服务层（缓存管理）
└── server/
    ├── cluster_handler.go      # 集群管理 HTTP 处理器
    ├── resource_handler.go     # 资源查询 HTTP 处理器（新增）
    ├── router.go               # 路由注册
    └── utils.go                # 工具函数
```

### 前端架构

```
web/src/
├── api/kubernetes.ts           # API 接口定义（集群 + 资源）
└── views/kubernetes/
    ├── Clusters.vue            # 集群管理页面
    └── Nodes.vue               # 节点管理页面（已实现）
```

## 核心功能实现

### 1. Clientset 缓存机制

在 `service/cluster_service.go` 中实现了并发安全的缓存：

```go
type ClusterService struct {
    clientsetCache map[uint]*kubernetes.Clientset
    cacheMutex     sync.RWMutex
}

func (s *ClusterService) GetCachedClientset(ctx context.Context, id uint) (*kubernetes.Clientset, error) {
    // 先读缓存
    s.cacheMutex.RLock()
    clientset, exists := s.clientsetCache[id]
    s.cacheMutex.RUnlock()

    if exists {
        return clientset, nil
    }

    // 缓存未命中，创建新连接
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
```

### 2. KubeConfig 加密存储

使用 AES-GCM 加密算法：

```go
const encryptionKey = "opshub-k8s-encrypt-key-32bytes!!"

// 加密
func encryptKubeConfig(plainText string) (string, error) {
    // 使用 AES-GCM 加密
    // 返回 Base64 编码的密文
}

// 解密
func decryptKubeConfig(cipherText string) (string, error) {
    // 解密 Base64 编码的密文
    // 返回原始 KubeConfig
}
```

### 3. 资源查询 API

#### 节点列表
```
GET /api/v1/plugins/kubernetes/resources/nodes?clusterId={clusterId}
```

返回字段：
- name: 节点名称
- status: 状态（Ready/NotReady）
- roles: 角色（master/control-plane/worker）
- internalIP: 内部IP地址
- version: Kubelet版本
- osImage: 操作系统镜像
- kernelVersion: 内核版本
- containerRuntime: 容器运行时
- age: 年龄（自动计算）
- labels: 标签

#### 命名空间列表
```
GET /api/v1/plugins/kubernetes/resources/namespaces?clusterId={clusterId}
```

#### Pod 列表
```
GET /api/v1/plugins/kubernetes/resources/pods?clusterId={clusterId}&namespace={namespace}
```

#### Deployment 列表
```
GET /api/v1/plugins/kubernetes/resources/deployments?clusterId={clusterId}&namespace={namespace}
```

## 使用示例

### 1. 注册集群

```typescript
import { createCluster } from '@/api/kubernetes'

// 使用 KubeConfig
await createCluster({
  name: 'my-cluster',
  alias: '生产集群',
  kubeConfig: 'apiVersion: v1\nkind: Config\n...',
  provider: 'aliyun',
  region: 'cn-beijing'
})

// 使用 Token
await createCluster({
  name: 'my-cluster',
  alias: '生产集群',
  apiEndpoint: 'https://k8s-api.example.com:6443',
  kubeConfig: '...', // 前端会自动构建 KubeConfig
  provider: 'aliyun'
})
```

### 2. 查询节点列表

```typescript
import { getNodes } from '@/api/kubernetes'

const nodes = await getNodes(clusterId)
console.log(nodes)
// [
//   {
//     name: 'node-1',
//     status: 'Ready',
//     roles: 'worker',
//     internalIP: '192.168.1.10',
//     version: 'v1.28.0',
//     ...
//   }
// ]
```

### 3. 查询 Pod 列表

```typescript
import { getPods } from '@/api/kubernetes'

// 查询所有命名空间的 Pod
const pods = await getPods(clusterId)

// 查询特定命名空间的 Pod
const pods = await getPods(clusterId, 'default')
```

## 安全特性

1. **KubeConfig 加密存储**：使用 AES-GCM 加密算法
2. **软删除机制**：集群数据不会真正删除
3. **连接测试**：注册时自动测试连接
4. **Clientset 缓存**：避免重复创建连接

## 性能优化

1. **Clientset 缓存**：每个集群的 clientset 只创建一次，后续从缓存读取
2. **并发安全**：使用 RWMutex 保证并发安全
3. **缓存清理**：更新或删除集群时自动清理对应缓存

## 扩展开发

### 添加新的资源查询

1. 在 `server/resource_handler.go` 添加新的 handler 方法
2. 在 `server/router.go` 注册路由
3. 在 `web/src/api/kubernetes.ts` 添加前端 API
4. 创建对应的 Vue 页面

示例：添加 Service 列表查询

```go
// server/resource_handler.go
func (h *ResourceHandler) ListServices(c *gin.Context) {
    // 实现逻辑
}

// server/router.go
clusters.GET("/resources/services", resourceHandler.ListServices)
```

## 注意事项

1. **加密密钥**：当前硬编码在代码中，生产环境应从配置中心读取
2. **连接池**：Clientset 会复用 HTTP 连接，注意连接池配置
3. **超时设置**：资源查询应设置合理的超时时间
4. **权限控制**：应根据用户权限过滤可见资源

## 依赖版本

- client-go: v0.28.0
- Kubernetes API: v1.28.0
- Go: 1.25
