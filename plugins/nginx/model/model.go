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

package model

import (
	"time"
)

// NginxSource Nginx 数据源类型
type NginxSourceType string

const (
	SourceTypeHost       NginxSourceType = "host"       // 主机上的 Nginx
	SourceTypeK8sIngress NginxSourceType = "k8s_ingress" // K8s Ingress-Nginx
)

// NginxSource Nginx 数据源配置
type NginxSource struct {
	ID          uint            `gorm:"primarykey" json:"id"`
	Name        string          `gorm:"type:varchar(100);not null" json:"name"`               // 数据源名称
	Type        NginxSourceType `gorm:"type:varchar(20);not null" json:"type"`                // 数据源类型
	Description string          `gorm:"type:varchar(500)" json:"description"`                 // 描述
	Status      int             `gorm:"type:tinyint;default:1" json:"status"`                 // 状态 1:启用 0:禁用

	// 主机类型配置
	HostID      *uint   `gorm:"index" json:"hostId"`                                          // 关联主机ID
	LogPath     string  `gorm:"type:varchar(500)" json:"logPath"`                             // 日志文件路径
	LogFormat   string  `gorm:"type:varchar(50);default:'combined'" json:"logFormat"`         // 日志格式

	// K8s Ingress 类型配置
	ClusterID   *uint   `gorm:"index" json:"clusterId"`                                       // 关联集群ID
	Namespace   string  `gorm:"type:varchar(100)" json:"namespace"`                           // Ingress 命名空间
	IngressName string  `gorm:"type:varchar(100)" json:"ingressName"`                         // Ingress 名称

	// 通用配置
	CollectInterval int `gorm:"type:int;default:60" json:"collectInterval"`                   // 采集间隔(秒)
	RetentionDays   int `gorm:"type:int;default:30" json:"retentionDays"`                     // 数据保留天数

	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `gorm:"index" json:"deletedAt,omitempty"`
}

func (NginxSource) TableName() string {
	return "nginx_sources"
}

// NginxAccessLog Nginx 访问日志
type NginxAccessLog struct {
	ID            uint64    `gorm:"primarykey" json:"id"`
	SourceID      uint      `gorm:"index;not null" json:"sourceId"`                           // 数据源ID
	Timestamp     time.Time `gorm:"type:datetime;index;not null" json:"timestamp"`            // 访问时间
	RemoteAddr    string    `gorm:"type:varchar(50);index" json:"remoteAddr"`                 // 客户端IP
	RemoteUser    string    `gorm:"type:varchar(100)" json:"remoteUser"`                      // 用户
	Request       string    `gorm:"type:varchar(2000)" json:"request"`                        // 请求行
	Method        string    `gorm:"type:varchar(10);index" json:"method"`                     // 请求方法
	URI           string    `gorm:"type:varchar(1000)" json:"uri"`                            // 请求URI (不加索引，太长)
	Protocol      string    `gorm:"type:varchar(20)" json:"protocol"`                         // 协议
	Status        int       `gorm:"type:int;index" json:"status"`                             // 状态码
	BodyBytesSent int64     `gorm:"type:bigint" json:"bodyBytesSent"`                         // 响应体大小
	HTTPReferer   string    `gorm:"type:varchar(1000)" json:"httpReferer"`                    // Referer
	HTTPUserAgent string    `gorm:"type:varchar(500)" json:"httpUserAgent"`                   // User-Agent
	RequestTime   float64   `gorm:"type:decimal(10,3)" json:"requestTime"`                    // 请求耗时(秒)
	UpstreamTime  float64   `gorm:"type:decimal(10,3)" json:"upstreamTime"`                   // 上游响应时间
	Host          string    `gorm:"type:varchar(255);index" json:"host"`                      // 请求主机

	// K8s Ingress 特有字段
	IngressName   string    `gorm:"type:varchar(100)" json:"ingressName"`                     // Ingress 名称
	ServiceName   string    `gorm:"type:varchar(100)" json:"serviceName"`                     // 后端服务名称

	CreatedAt     time.Time `json:"createdAt"`
}

func (NginxAccessLog) TableName() string {
	return "nginx_access_logs"
}

// NginxDailyStats Nginx 日统计数据
type NginxDailyStats struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	SourceID        uint      `gorm:"index;not null" json:"sourceId"`                         // 数据源ID
	Date            time.Time `gorm:"type:date;index;not null" json:"date"`                   // 统计日期
	TotalRequests   int64     `gorm:"type:bigint;default:0" json:"totalRequests"`             // 总请求数
	UniqueVisitors  int64     `gorm:"type:bigint;default:0" json:"uniqueVisitors"`            // 独立访客数
	TotalBandwidth  int64     `gorm:"type:bigint;default:0" json:"totalBandwidth"`            // 总带宽(字节)
	AvgResponseTime float64   `gorm:"type:decimal(10,3);default:0" json:"avgResponseTime"`    // 平均响应时间
	Status2xx       int64     `gorm:"type:bigint;default:0" json:"status2xx"`                 // 2xx 状态码数
	Status3xx       int64     `gorm:"type:bigint;default:0" json:"status3xx"`                 // 3xx 状态码数
	Status4xx       int64     `gorm:"type:bigint;default:0" json:"status4xx"`                 // 4xx 状态码数
	Status5xx       int64     `gorm:"type:bigint;default:0" json:"status5xx"`                 // 5xx 状态码数

	// Top 数据 (JSON 格式存储)
	TopURIs         string    `gorm:"type:text" json:"topURIs"`                               // Top URI
	TopIPs          string    `gorm:"type:text" json:"topIPs"`                                // Top IP
	TopReferers     string    `gorm:"type:text" json:"topReferers"`                           // Top Referer
	TopUserAgents   string    `gorm:"type:text" json:"topUserAgents"`                         // Top User-Agent

	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func (NginxDailyStats) TableName() string {
	return "nginx_daily_stats"
}

// NginxHourlyStats Nginx 小时统计数据
type NginxHourlyStats struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	SourceID        uint      `gorm:"index;not null" json:"sourceId"`                         // 数据源ID
	Hour            time.Time `gorm:"type:datetime;index;not null" json:"hour"`               // 统计小时
	TotalRequests   int64     `gorm:"type:bigint;default:0" json:"totalRequests"`             // 总请求数
	UniqueVisitors  int64     `gorm:"type:bigint;default:0" json:"uniqueVisitors"`            // 独立访客数
	TotalBandwidth  int64     `gorm:"type:bigint;default:0" json:"totalBandwidth"`            // 总带宽(字节)
	AvgResponseTime float64   `gorm:"type:decimal(10,3);default:0" json:"avgResponseTime"`    // 平均响应时间
	Status2xx       int64     `gorm:"type:bigint;default:0" json:"status2xx"`                 // 2xx 状态码数
	Status3xx       int64     `gorm:"type:bigint;default:0" json:"status3xx"`                 // 3xx 状态码数
	Status4xx       int64     `gorm:"type:bigint;default:0" json:"status4xx"`                 // 4xx 状态码数
	Status5xx       int64     `gorm:"type:bigint;default:0" json:"status5xx"`                 // 5xx 状态码数

	CreatedAt       time.Time `json:"createdAt"`
}

func (NginxHourlyStats) TableName() string {
	return "nginx_hourly_stats"
}

// NginxRealTimeStats 实时统计数据 (内存中保持，不入库)
type NginxRealTimeStats struct {
	SourceID        uint      `json:"sourceId"`
	Timestamp       time.Time `json:"timestamp"`
	RequestsPerSec  float64   `json:"requestsPerSec"`   // 每秒请求数
	BandwidthPerSec int64     `json:"bandwidthPerSec"`  // 每秒带宽
	ActiveConns     int64     `json:"activeConns"`      // 活跃连接数
	AvgResponseTime float64   `json:"avgResponseTime"`  // 平均响应时间
	Status2xxRate   float64   `json:"status2xxRate"`    // 2xx 比例
	Status4xxRate   float64   `json:"status4xxRate"`    // 4xx 比例
	Status5xxRate   float64   `json:"status5xxRate"`    // 5xx 比例
}

// OverviewStats 概况统计
type OverviewStats struct {
	TotalSources    int64           `json:"totalSources"`    // 数据源总数
	ActiveSources   int64           `json:"activeSources"`   // 活跃数据源数
	TodayRequests   int64           `json:"todayRequests"`   // 今日请求数
	TodayVisitors   int64           `json:"todayVisitors"`   // 今日访客数
	TodayBandwidth  int64           `json:"todayBandwidth"`  // 今日带宽
	TodayErrorRate  float64         `json:"todayErrorRate"`  // 今日错误率
	RequestsTrend   []TrendPoint    `json:"requestsTrend"`   // 请求趋势
	BandwidthTrend  []TrendPoint    `json:"bandwidthTrend"`  // 带宽趋势
	StatusDistribution map[string]int64 `json:"statusDistribution"` // 状态码分布
}

// TrendPoint 趋势数据点
type TrendPoint struct {
	Time  string `json:"time"`
	Value int64  `json:"value"`
}
