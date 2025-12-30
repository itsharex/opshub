package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/ydcloud-dy/opshub/internal/conf"
	"github.com/ydcloud-dy/opshub/internal/service"
	"github.com/ydcloud-dy/opshub/pkg/middleware"
	appLogger "github.com/ydcloud-dy/opshub/pkg/logger"
	"go.uber.org/zap"
)

// HTTPServer HTTP服务器
type HTTPServer struct {
	server *http.Server
	conf   *conf.Config
	svc    *service.Service
}

// NewHTTPServer 创建HTTP服务器
func NewHTTPServer(conf *conf.Config, svc *service.Service) *HTTPServer {
	// 设置Gin模式
	gin.SetMode(conf.Server.Mode)

	// 创建路由
	router := gin.New()

	// 使用中间件
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())

	// 注册路由
	s := &HTTPServer{
		conf: conf,
		svc:  svc,
	}

	s.registerRoutes(router)

	// 创建HTTP服务器
	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", conf.Server.HttpPort),
		Handler:      router,
		ReadTimeout:  time.Duration(conf.Server.ReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(conf.Server.WriteTimeout) * time.Millisecond,
	}

	return s
}

// registerRoutes 注册路由
func (s *HTTPServer) registerRoutes(router *gin.Engine) {
	// Swagger 文档
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 健康检查
	router.GET("/health", s.svc.Health)

	// API v1
	v1 := router.Group("/api/v1")
	{
		// 示例接口
		v1.GET("/example", s.svc.Example)

		// 在这里添加更多路由
		// v1.POST("/users", s.svc.CreateUser)
		// v1.GET("/users/:id", s.svc.GetUser)
	}
}

// Start 启动服务器
func (s *HTTPServer) Start() error {
	appLogger.Info("HTTP服务器启动",
		zap.String("addr", s.server.Addr),
		zap.String("mode", s.conf.Server.Mode),
	)

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP服务器启动失败: %w", err)
	}

	return nil
}

// Stop 停止服务器
func (s *HTTPServer) Stop(ctx context.Context) error {
	appLogger.Info("HTTP服务器停止中...")
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("HTTP服务器停止失败: %w", err)
	}
	appLogger.Info("HTTP服务器已停止")
	return nil
}
