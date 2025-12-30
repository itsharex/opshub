package main

import (
	"fmt"
	"os"

	"github.com/ydcloud-dy/opshub/cmd/root"
	_ "github.com/ydcloud-dy/opshub/cmd/config"  // 注册配置命令
	_ "github.com/ydcloud-dy/opshub/cmd/server"  // 注册服务命令
	_ "github.com/ydcloud-dy/opshub/cmd/version" // 注册版本命令
	_ "github.com/ydcloud-dy/opshub/docs"        // 导入 Swagger 生成的文档
)

// @title           OpsHub API
// @version         1.0
// @description     运维管理平台 API 文档
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}
