# OpsHub 插件使用指南

## 概述

OpsHub 插件系统支持前后端完全可插拔的架构设计，可以通过简单的操作实现插件的启用和禁用，无需修改核心代码。

## 插件可插拔机制

### 什么是可插拔

插件可插拔指的是：
- **安装/启用**: 动态加载插件功能，包括菜单、路由、API接口等
- **卸载/禁用**: 动态移除插件功能，清理相关资源
- **无侵入**: 插件的安装和卸载不影响系统核心功能
- **独立性**: 每个插件独立运行，互不干扰

### 架构设计

```
┌─────────────────────────────────────────────────────┐
│                   OpsHub 系统                        │
├─────────────────────────────────────────────────────┤
│                                                      │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐   │
│  │ 插件管理器  │  │ 菜单管理器  │  │ 路由管理器  │   │
│  └────────────┘  └────────────┘  └────────────┘   │
│         │               │               │           │
│         └───────────────┴───────────────┘           │
│                         │                           │
│         ┌───────────────┴───────────────┐           │
│         │                               │           │
│  ┌──────▼──────┐              ┌─────────▼────────┐ │
│  │ Kubernetes  │              │  其他插件...      │ │
│  │   插件      │              │                   │ │
│  └─────────────┘              └──────────────────┘ │
│                                                      │
└─────────────────────────────────────────────────────┘
```

## Kubernetes 插件的安装和卸载

### 方式一：通过代码配置（默认）

#### 1. 后端启用 Kubernetes 插件

**文件路径**: `internal/server/http.go`

```go
// 注册插件
if err := pluginMgr.Register(k8splugin.New()); err != nil {
    appLogger.Error("注册Kubernetes插件失败", zap.Error(err))
}
```

#### 2. 前端启用 Kubernetes 插件

**文件路径**: `web/src/main.ts`

```typescript
// 导入插件
import '@/plugins/kubernetes'
```

#### 3. 禁用插件

**禁用后端插件**: 注释掉 `internal/server/http.go` 中的注册代码
```go
// if err := pluginMgr.Register(k8splugin.New()); err != nil {
//     appLogger.Error("注册Kubernetes插件失败", zap.Error(err))
// }
```

**禁用前端插件**: 注释掉 `web/src/main.ts` 中的导入代码
```typescript
// import '@/plugins/kubernetes'
```

### 方式二：通过 API 动态控制（推荐）

系统提供了插件管理 API，可以在运行时动态启用或禁用插件。

#### 1. 启用插件

```bash
# 启用 Kubernetes 插件
curl -X POST http://localhost:9876/api/v1/plugins/kubernetes/enable \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### 2. 禁用插件

```bash
# 禁用 Kubernetes 插件
curl -X POST http://localhost:9876/api/v1/plugins/kubernetes/disable \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### 3. 查看插件列表

```bash
# 获取所有插件列表及状态
curl http://localhost:9876/api/v1/plugins \
  -H "Authorization: Bearer YOUR_TOKEN"
```

响应示例：
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "name": "kubernetes",
      "description": "Kubernetes容器管理平台,提供集群管理、节点管理、工作负载、命名空间等完整功能",
      "version": "1.0.0",
      "author": "OpsHub Team",
      "enabled": true
    }
  ]
}
```

#### 4. 查看插件详情

```bash
# 获取 Kubernetes 插件详情
curl http://localhost:9876/api/v1/plugins/kubernetes \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 方式三：通过前端界面控制

你可以在系统管理界面创建一个插件管理页面，通过调用前端 API 来控制插件：

```typescript
import { enablePlugin, disablePlugin, listPlugins } from '@/api/plugin'

// 启用插件
await enablePlugin('kubernetes')

// 禁用插件
await disablePlugin('kubernetes')

// 获取插件列表
const plugins = await listPlugins()
```

## 插件状态持久化

插件的启用/禁用状态会自动保存到数据库的 `plugin_states` 表中：

```sql
CREATE TABLE plugin_states (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(100) NOT NULL UNIQUE,
  enabled BOOLEAN NOT NULL DEFAULT FALSE,
  created_at DATETIME,
  updated_at DATETIME
);
```

- 当插件被启用时，`enabled` 字段会被设置为 `true`
- 当插件被禁用时，`enabled` 字段会被设置为 `false`
- 系统重启后会自动从数据库读取插件状态

## 插件生命周期

### 1. 注册阶段
- 后端：在 `http.go` 中调用 `pluginMgr.Register()`
- 前端：在 `main.ts` 中 import 插件模块
- 系统会自动在数据库中创建插件状态记录（默认禁用）

### 2. 启用阶段
- 调用插件的 `Enable()` 方法（后端）或 `install()` 方法（前端）
- 注册路由和菜单
- 更新数据库状态为 `enabled=true`

### 3. 运行阶段
- 插件的路由、菜单、功能正常工作
- 用户可以访问插件提供的所有功能

### 4. 禁用阶段
- 调用插件的 `Disable()` 方法（后端）或 `uninstall()` 方法（前端）
- 清理插件资源（注意：不会删除数据库表）
- 更新数据库状态为 `enabled=false`
- 需要刷新页面才能完全移除前端路由

## Kubernetes 插件功能清单

Kubernetes 插件提供了以下功能模块：

1. **集群管理** - 管理 Kubernetes 集群的接入和配置
2. **节点管理** - 查看和管理集群节点
3. **工作负载** - 管理 Deployments、StatefulSets、DaemonSets、Pods
4. **命名空间** - 管理 Kubernetes 命名空间
5. **网络管理** - 管理 Services、Ingress、NetworkPolicies
6. **配置管理** - 管理 ConfigMaps 和 Secrets
7. **存储管理** - 管理 PV、PVC、StorageClasses
8. **访问控制** - 管理 ServiceAccounts、Roles、RoleBindings
9. **终端审计** - 记录和审计终端会话
10. **应用诊断** - 应用健康检查和诊断工具

## 插件目录结构

```
opshub/
├── internal/
│   └── plugins/
│       └── kubernetes/          # 后端插件
│           └── plugin.go        # 插件实现
│
└── web/
    ├── src/
    │   ├── plugins/
    │   │   └── kubernetes/      # 前端插件
    │   │       └── index.ts     # 插件入口
    │   │
    │   ├── views/
    │   │   └── kubernetes/      # 页面组件
    │   │       ├── Index.vue
    │   │       ├── Clusters.vue
    │   │       ├── Nodes.vue
    │   │       └── ...
    │   │
    │   └── api/
    │       └── plugin.ts        # 插件管理API
    │
    └── docs/
        ├── PLUGIN_DEVELOPMENT.md  # 插件开发指南
        └── PLUGIN_USAGE.md        # 插件使用指南（本文档）
```

## 常见问题

### Q1: 启用/禁用插件后需要重启服务器吗？

A:
- **后端**: 不需要重启服务器，但建议刷新页面以确保前端同步更新
- **前端**: 禁用插件后需要刷新页面才能完全移除路由

### Q2: 禁用插件会删除数据吗？

A: 不会。禁用插件只会清理运行时资源（如缓存），不会删除数据库表和数据。如需删除数据，需要手动执行 SQL。

### Q3: 如何一键卸载插件？

A:
1. 通过 API 调用禁用插件
2. 从数据库中删除插件状态记录
3. 删除插件相关的代码文件
4. 手动清理插件创建的数据库表（如需要）

```bash
# 1. 禁用插件
curl -X POST http://localhost:9876/api/v1/plugins/kubernetes/disable

# 2. 删除插件状态
mysql> DELETE FROM plugin_states WHERE name = 'kubernetes';

# 3. 删除代码文件
rm -rf internal/plugins/kubernetes
rm -rf web/src/plugins/kubernetes
rm -rf web/src/views/kubernetes

# 4. 清理数据库表（可选）
mysql> DROP TABLE plugin_kubernetes_clusters;
```

### Q4: 可以同时运行多个插件吗？

A: 可以。插件系统支持同时运行多个插件，它们之间相互独立，互不影响。

### Q5: 如何查看插件的日志？

A: 插件的日志会记录在系统日志文件 `logs/app.log` 中，可以通过搜索插件名称来过滤日志：

```bash
# 查看 Kubernetes 插件日志
tail -f logs/app.log | grep kubernetes
```

## 最佳实践

1. **开发阶段**: 使用代码配置方式，便于调试
2. **生产环境**: 使用 API 动态控制，避免修改代码
3. **测试插件**: 先在测试环境启用，确认无误后再在生产环境启用
4. **备份数据**: 在禁用或卸载插件前，先备份相关数据
5. **版本控制**: 记录插件的版本号，便于升级和回滚

## 技术支持

如有问题，请联系：
- Issue: https://github.com/ydcloud-dy/opshub/issues
- 文档: https://github.com/ydcloud-dy/opshub/wiki
