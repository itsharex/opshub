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

package nginx

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/ydcloud-dy/opshub/internal/plugin"
	"github.com/ydcloud-dy/opshub/plugins/nginx/model"
	"github.com/ydcloud-dy/opshub/plugins/nginx/server"
)

// Plugin Nginx统计插件实现
type Plugin struct {
	db        *gorm.DB
	name      string
	ctx       context.Context
	cancelCtx context.CancelFunc
}

// New 创建插件实例
func New() *Plugin {
	return &Plugin{
		name: "nginx",
	}
}

// Name 返回插件名称
func (p *Plugin) Name() string {
	return "nginx"
}

// Description 返回插件描述
func (p *Plugin) Description() string {
	return "Nginx统计插件 - 支持主机Nginx和K8s Ingress-Nginx的访问日志分析和统计"
}

// Version 返回插件版本
func (p *Plugin) Version() string {
	return "1.0.0"
}

// Author 返回插件作者
func (p *Plugin) Author() string {
	return "J"
}

// Enable 启用插件
func (p *Plugin) Enable(db *gorm.DB) error {
	p.db = db

	// 自动迁移所有插件相关的表
	models := []interface{}{
		&model.NginxSource{},
		&model.NginxAccessLog{},
		&model.NginxDailyStats{},
		&model.NginxHourlyStats{},
	}

	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			return err
		}
	}

	// 启动上下文
	p.ctx, p.cancelCtx = context.WithCancel(context.Background())

	// TODO: 启动日志采集调度器
	// go p.startLogCollectorScheduler()

	return nil
}

// Disable 禁用插件
func (p *Plugin) Disable(db *gorm.DB) error {
	// 停止定时任务
	if p.cancelCtx != nil {
		p.cancelCtx()
	}
	return nil
}

// RegisterRoutes 注册路由
func (p *Plugin) RegisterRoutes(router *gin.RouterGroup, db *gorm.DB) {
	server.RegisterRoutes(router, db)
}

// GetMenus 获取插件菜单配置
func (p *Plugin) GetMenus() []plugin.MenuConfig {
	parentPath := "/nginx"

	return []plugin.MenuConfig{
		{
			Name:       "Nginx统计",
			Path:       parentPath,
			Icon:       "DataLine",
			Sort:       50,
			Hidden:     false,
			ParentPath: "",
		},
		{
			Name:       "概况",
			Path:       "/nginx/overview",
			Icon:       "PieChart",
			Sort:       1,
			Hidden:     false,
			ParentPath: parentPath,
		},
		{
			Name:       "数据日报",
			Path:       "/nginx/daily-report",
			Icon:       "Calendar",
			Sort:       2,
			Hidden:     false,
			ParentPath: parentPath,
		},
		{
			Name:       "实时",
			Path:       "/nginx/realtime",
			Icon:       "Timer",
			Sort:       3,
			Hidden:     false,
			ParentPath: parentPath,
		},
		{
			Name:       "访问明细",
			Path:       "/nginx/access-logs",
			Icon:       "List",
			Sort:       4,
			Hidden:     false,
			ParentPath: parentPath,
		},
		{
			Name:       "功能配置",
			Path:       "/nginx/config",
			Icon:       "Setting",
			Sort:       5,
			Hidden:     false,
			ParentPath: parentPath,
		},
	}
}
