# Kubernetes 集群管理功能

## 功能特性

### 集群管理
- ✅ 集群列表展示
- ✅ 注册新集群
- ✅ 两种认证方式：KubeConfig / Token
- ✅ 测试集群连接
- ✅ 删除集群
- ✅ 集群状态监控

### 支持的云服务商
- 自建集群 (Native)
- 阿里云 ACK
- 腾讯云 TKE
- AWS EKS

## 使用说明

### 1. 通过 KubeConfig 注册集群

1. 点击"注册集群"按钮
2. 选择认证方式为"KubeConfig"
3. 输入集群名称和别名
4. 粘贴 KubeConfig 文件内容
5. 选择服务商和区域
6. 点击"确定"

### 2. 通过 Token 注册集群

1. 点击"注册集群"按钮
2. 选择认证方式为"Token"
3. 输入集群名称和别名
4. 输入 API Endpoint（例如：https://k8s-api.example.com）
5. 输入 Service Account Token
6. 选择服务商和区域
7. 点击"确定"

### 3. 测试集群连接

- 在集群列表中点击"测试连接"按钮
- 系统会自动连接集群并获取版本信息
- 连接成功后会显示 Kubernetes 版本

## API 接口

```typescript
// 获取集群列表
getClusterList()

// 创建集群
createCluster(data: CreateClusterParams)

// 删除集群
deleteCluster(id: number)

// 测试连接
testClusterConnection(id: number)
```

## 数据模型

```typescript
interface Cluster {
  id: number              // 集群ID
  name: string           // 集群名称（唯一）
  alias: string          // 集群别名
  apiEndpoint: string    // API Server 地址
  version: string        // Kubernetes 版本
  status: number         // 状态：1-正常 2-失败 3-不可用
  provider: string       // 服务商
  region: string         // 区域
  description: string    // 描述
  createdAt: string      // 创建时间
  updatedAt: string      // 更新时间
}
```

## 安全特性

- KubeConfig 加密存储（AES-GCM）
- Token 自动转换为 KubeConfig
- 连接状态实时监控
- 软删除机制

## 插件架构

前端插件完全独立，不与主应用耦合：

```
web/src/
├── plugins/              # 插件目录
│   ├── kubernetes/      # Kubernetes 插件
│   │   └── index.ts     # 插件入口
│   ├── types.ts         # 插件类型定义
│   └── manager.ts       # 插件管理器
├── views/               # 视图目录（插件视图）
│   └── kubernetes/      # Kubernetes 插件视图
│       ├── Clusters.vue # 集群管理
│       └── Index.vue    # 插件入口页
└── api/                 # API 目录
    └── kubernetes.ts    # Kubernetes API
```

## 开发指南

### 添加新的 Kubernetes 功能

1. 在 `web/src/views/kubernetes/` 创建新页面
2. 在 `web/src/api/kubernetes.ts` 添加 API 接口
3. 在后端 `plugins/kubernetes/` 实现对应功能
4. 在插件配置中注册路由和菜单

### 插件完全解耦

- ✅ 所有 Kubernetes 相关代码都在插件目录
- ✅ 可以独立开发、测试、部署
- ✅ 删除插件不影响主应用
- ✅ 符合开闭原则
