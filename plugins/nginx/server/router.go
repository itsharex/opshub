// Copyright (c) 2026 DYCloud J.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package server

import (
	"github.com/gin-gonic/gin"
	"github.com/ydcloud-dy/opshub/plugins/nginx/model"
	"gorm.io/gorm"
)

// RegisterRoutes 注册路由
func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB) {
	handler := NewHandler(db)

	// Nginx 统计插件路由组 - 使用 /nginx 前缀
	nginxGroup := router.Group("/nginx")
	{
		// 数据源管理
		sources := nginxGroup.Group("/sources")
		{
			sources.GET("", handler.ListSources)        // 获取数据源列表
			sources.GET("/:id", handler.GetSource)      // 获取数据源详情
			sources.POST("", handler.CreateSource)      // 创建数据源
			sources.PUT("/:id", handler.UpdateSource)   // 更新数据源
			sources.DELETE("/:id", handler.DeleteSource) // 删除数据源
		}

		// 概况统计
		nginxGroup.GET("/overview", handler.GetOverview)            // 获取概况
		nginxGroup.GET("/overview/trend", handler.GetRequestsTrend) // 获取请求趋势

		// 数据日报
		nginxGroup.GET("/daily-report", handler.GetDailyReport) // 获取日报数据

		// 实时统计
		nginxGroup.GET("/realtime", handler.GetRealTimeStats) // 获取实时数据

		// 日志采集
		nginxGroup.POST("/collect", handler.CollectLogs) // 手动触发日志采集

		// 访问明细
		accessLogs := nginxGroup.Group("/access-logs")
		{
			accessLogs.GET("", handler.ListAccessLogs)          // 获取访问日志列表
			accessLogs.GET("/top-uris", handler.GetTopURIs)     // 获取 Top URI
			accessLogs.GET("/top-ips", handler.GetTopIPs)       // 获取 Top IP
		}
	}
}

// AutoMigrate 自动迁移表结构
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.NginxSource{},
		&model.NginxAccessLog{},
		&model.NginxDailyStats{},
		&model.NginxHourlyStats{},
	)
}
