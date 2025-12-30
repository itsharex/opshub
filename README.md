# OpsHub 运维管理平台

基于 Gin + Cobra + Viper 的运维管理平台后端脚手架

## 项目结构

```
.
├── bin/              # 编译输出目录
├── cmd/              # 命令行工具
│   ├── root/        # 根命令
│   ├── server/      # 服务启动命令
│   ├── config/      # 配置管理命令
│   └── version/     # 版本信息命令
├── config/           # 配置文件
│   └── config.yaml  # 主配置文件
├── docs/             # Swagger 文档
├── internal/         # 内部代码
│   ├── biz/         # 业务逻辑层
│   ├── conf/        # 配置管理
│   ├── data/        # 数据层(MySQL/Redis)
│   ├── server/      # HTTP 服务器
│   └── service/     # 服务层(HTTP handlers)
├── pkg/              # 公共包
│   ├── error/       # 错误处理
│   ├── logger/      # 日志
│   ├── middleware/  # 中间件
│   └── response/    # 响应封装
├── logs/             # 日志文件目录
├── main.go          # 主入口
├── Makefile         # 构建脚本
├── go.mod
└── go.sum
```

## 功能特性

- ✅ 基于 Gin 框架的 HTTP 服务器
- ✅ Cobra 命令行工具框架
- ✅ Viper 配置管理(支持环境变量、命令行参数)
- ✅ Swagger API 文档
- ✅ 统一的错误处理和响应格式
- ✅ 结构化日志(支持文件轮转)
- ✅ MySQL 数据库支持(GORM)
- ✅ Redis 缓存支持
- ✅ 中间件支持(日志、恢复、CORS)
- ✅ 优雅关闭

## 快速开始

### 1. 配置

编辑 `config/config.yaml` 文件,配置数据库和 Redis 连接信息:

```yaml
server:
  mode: debug
  http_port: 8080

database:
  host: 127.0.0.1
  port: 3306
  database: opshub
  username: root
  password: ""

redis:
  host: 127.0.0.1
  port: 6379
  password: ""
```

### 2. 运行

```bash
# 安装依赖
go mod tidy

# 生成 Swagger 文档
make swagger

# 运行服务
make run
# 或
go run main.go server

# 指定配置文件
./bin/opshub server -c config/config.yaml

# 覆盖配置参数
./bin/opshub server -m debug -l info --server.http-port 9090

# 使用环境变量
export OPSHUB_SERVER_MODE=release
export OPSHUB_DATABASE_HOST=192.168.1.100
./bin/opshub server

# 编译
make build
```

### 3. 测试

```bash
# 查看帮助
./bin/opshub -h
./bin/opshub server -h

# 查看版本
./bin/opshub version

# 验证配置
./bin/opshub config validate

# 打印配置
./bin/opshub config print

# 健康检查
curl http://localhost:8080/health

# 示例接口
curl http://localhost:8080/api/v1/example

# Swagger 文档
# 浏览器访问: http://localhost:8080/swagger/index.html
```

## 命令行工具

### 全局选项

```
-c, --config string        配置文件路径 (默认为 config/config.yaml)
-m, --mode string          运行模式: debug, release, test
-l, --log-level string     日志级别: debug, info, warn, error
```

### 子命令

- `server` - 启动 HTTP 服务器
- `config` - 配置管理
  - `config validate` - 验证配置文件
  - `config print` - 打印配置内容
- `version` - 显示版本信息

## 配置优先级

1. 命令行参数 (最高优先级)
2. 环境变量 (OPSHUB_ 前缀)
3. 配置文件
4. 默认值 (最低优先级)

## 开发指南

### 添加新的 API

1. 在 `internal/service/` 中添加处理函数并添加 Swagger 注释
2. 在 `internal/server/http.go` 中注册路由
3. 在 `internal/biz/` 中实现业务逻辑
4. 在 `internal/data/` 中添加数据访问方法
5. 运行 `make swagger` 重新生成文档

### Swagger 注释示例

```go
// @Summary 接口标题
// @Description 接口描述
// @Tags 分类
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} response.Response
// @Router /api/v1/users/{id} [get]
func (s *Service) GetUser(c *gin.Context) {
    // ...
}
```

### 错误处理

```go
import (
    appError "github.com/ydcloud-dy/opshub/pkg/error"
    "github.com/ydcloud-dy/opshub/pkg/response"
)

// 返回错误
return appError.New(appError.ErrValidation, "参数验证失败")

// 在 handler 中使用
response.Error(c, err)
```

### 日志记录

```go
import "github.com/ydcloud-dy/opshub/pkg/logger"

logger.Info("信息日志", zap.String("key", "value"))
logger.Error("错误日志", zap.Error(err))
```

## 技术栈

- **Web框架**: Gin
- **CLI框架**: Cobra
- **配置管理**: Viper
- **API文档**: Swagger
- **ORM**: GORM
- **数据库**: MySQL
- **缓存**: Redis
- **日志**: Zap + Lumberjack

## 许可证

MIT License
