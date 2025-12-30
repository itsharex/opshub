package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ydcloud-dy/opshub/cmd/root"
	"github.com/ydcloud-dy/opshub/internal/biz"
	"github.com/ydcloud-dy/opshub/internal/conf"
	dataPkg "github.com/ydcloud-dy/opshub/internal/data"
	"github.com/ydcloud-dy/opshub/internal/server"
	"github.com/ydcloud-dy/opshub/internal/service"
	appLogger "github.com/ydcloud-dy/opshub/pkg/logger"
	"go.uber.org/zap"
)

var Cmd = &cobra.Command{
	Use:   "server",
	Short: "启动服务",
	Long:  `启动 OpsHub HTTP 服务器`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// 从命令行参数覆盖配置
		if mode := viper.GetString("mode"); mode != "" {
			viper.Set("server.mode", mode)
		}
		if logLevel := viper.GetString("log-level"); logLevel != "" {
			viper.Set("log.level", logLevel)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// 加载配置
		cfg, err := runServer()
		if err != nil {
			fmt.Printf("启动服务失败: %v\n", err)
			os.Exit(1)
		}

		// 等待中断信号
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		fmt.Println("\n正在关闭服务...")
		ctx := context.Background()
		if err := stopServer(ctx, cfg); err != nil {
			fmt.Printf("关闭服务失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("服务已关闭")
	},
}

func init() {
	root.Cmd.AddCommand(Cmd)
}

func runServer() (*conf.Config, error) {
	// 加载配置
	cfg, err := conf.Load(root.GetConfigFile())
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	// 初始化日志
	logCfg := &appLogger.Config{
		Level:      cfg.Log.Level,
		Filename:   cfg.Log.Filename,
		MaxSize:    cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxAge,
		Compress:   cfg.Log.Compress,
		Console:    cfg.Log.Console,
	}
	if err := appLogger.Init(logCfg); err != nil {
		return nil, fmt.Errorf("初始化日志失败: %w", err)
	}
	defer appLogger.Sync()

	appLogger.Info("服务启动中...",
		zap.String("version", "1.0.0"),
		zap.String("mode", cfg.Server.Mode),
	)

	// 初始化数据层
	data, err := dataPkg.NewData(cfg)
	if err != nil {
		return nil, fmt.Errorf("初始化数据层失败: %w", err)
	}
	defer data.Close()

	// 初始化Redis
	redis, err := dataPkg.NewRedis(cfg)
	if err != nil {
		return nil, fmt.Errorf("初始化Redis失败: %w", err)
	}
	defer redis.Close()

	// 初始化业务层
	biz := biz.NewBiz(data, redis)

	// 初始化服务层
	svc := service.NewService(biz)

	// 初始化HTTP服务器
	httpServer := server.NewHTTPServer(cfg, svc)

	// 启动服务器
	go func() {
		if err := httpServer.Start(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("HTTP服务器启动失败", zap.Error(err))
		}
	}()

	// 打印启动信息
	printStartupInfo(cfg)

	return cfg, nil
}

func stopServer(ctx context.Context, cfg *conf.Config) error {
	appLogger.Info("服务正在关闭...")
	// 这里可以添加更多的清理逻辑
	return nil
}

func printStartupInfo(cfg *conf.Config) {
	addr := fmt.Sprintf("%s:%d", "0.0.0.0", cfg.Server.HttpPort)

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("       OpsHub 运维管理平台启动成功")
	fmt.Println("========================================")
	fmt.Printf("版本:     1.0.0\n")
	fmt.Printf("模式:     %s\n", cfg.Server.Mode)
	fmt.Printf("监听地址: http://%s\n", addr)
	fmt.Printf("健康检查: http://%s/health\n", addr)
	fmt.Printf("API文档:  http://%s/swagger/index.html\n", addr)
	fmt.Println("========================================")
	fmt.Println()
}
