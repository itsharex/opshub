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

package repository

import (
	"time"

	"github.com/ydcloud-dy/opshub/plugins/nginx/model"
	"gorm.io/gorm"
)

// NginxRepository Nginx 数据仓库
type NginxRepository struct {
	db *gorm.DB
}

// NewNginxRepository 创建仓库实例
func NewNginxRepository(db *gorm.DB) *NginxRepository {
	return &NginxRepository{db: db}
}

// ============== 数据源操作 ==============

// CreateSource 创建数据源
func (r *NginxRepository) CreateSource(source *model.NginxSource) error {
	return r.db.Create(source).Error
}

// UpdateSource 更新数据源
func (r *NginxRepository) UpdateSource(source *model.NginxSource) error {
	return r.db.Save(source).Error
}

// DeleteSource 删除数据源
func (r *NginxRepository) DeleteSource(id uint) error {
	return r.db.Delete(&model.NginxSource{}, id).Error
}

// GetSourceByID 根据ID获取数据源
func (r *NginxRepository) GetSourceByID(id uint) (*model.NginxSource, error) {
	var source model.NginxSource
	err := r.db.First(&source, id).Error
	if err != nil {
		return nil, err
	}
	return &source, nil
}

// ListSources 获取数据源列表
func (r *NginxRepository) ListSources(page, pageSize int, sourceType string, status *int) ([]model.NginxSource, int64, error) {
	var sources []model.NginxSource
	var total int64

	query := r.db.Model(&model.NginxSource{})

	if sourceType != "" {
		query = query.Where("type = ?", sourceType)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&sources).Error
	if err != nil {
		return nil, 0, err
	}

	return sources, total, nil
}

// GetActiveSources 获取活跃数据源
func (r *NginxRepository) GetActiveSources() ([]model.NginxSource, error) {
	var sources []model.NginxSource
	err := r.db.Where("status = ?", 1).Find(&sources).Error
	return sources, err
}

// ============== 访问日志操作 ==============

// CreateAccessLog 创建访问日志
func (r *NginxRepository) CreateAccessLog(log *model.NginxAccessLog) error {
	return r.db.Create(log).Error
}

// BatchCreateAccessLogs 批量创建访问日志
func (r *NginxRepository) BatchCreateAccessLogs(logs []model.NginxAccessLog) error {
	if len(logs) == 0 {
		return nil
	}
	return r.db.CreateInBatches(logs, 1000).Error
}

// ListAccessLogs 获取访问日志列表
func (r *NginxRepository) ListAccessLogs(sourceID uint, page, pageSize int, startTime, endTime *time.Time, filters map[string]interface{}) ([]model.NginxAccessLog, int64, error) {
	var logs []model.NginxAccessLog
	var total int64

	query := r.db.Model(&model.NginxAccessLog{}).Where("source_id = ?", sourceID)

	if startTime != nil {
		query = query.Where("timestamp >= ?", startTime)
	}
	if endTime != nil {
		query = query.Where("timestamp <= ?", endTime)
	}

	// 应用过滤条件
	if filters != nil {
		if ip, ok := filters["remoteAddr"]; ok && ip != "" {
			query = query.Where("remote_addr LIKE ?", "%"+ip.(string)+"%")
		}
		if uri, ok := filters["uri"]; ok && uri != "" {
			query = query.Where("uri LIKE ?", "%"+uri.(string)+"%")
		}
		if status, ok := filters["status"]; ok && status != 0 {
			query = query.Where("status = ?", status)
		}
		if method, ok := filters["method"]; ok && method != "" {
			query = query.Where("method = ?", method)
		}
		if host, ok := filters["host"]; ok && host != "" {
			query = query.Where("host LIKE ?", "%"+host.(string)+"%")
		}
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("timestamp DESC").Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// DeleteOldAccessLogs 删除过期访问日志
func (r *NginxRepository) DeleteOldAccessLogs(sourceID uint, beforeTime time.Time) error {
	return r.db.Where("source_id = ? AND timestamp < ?", sourceID, beforeTime).Delete(&model.NginxAccessLog{}).Error
}

// ============== 日统计操作 ==============

// CreateOrUpdateDailyStats 创建或更新日统计
func (r *NginxRepository) CreateOrUpdateDailyStats(stats *model.NginxDailyStats) error {
	return r.db.Save(stats).Error
}

// GetDailyStats 获取日统计数据
func (r *NginxRepository) GetDailyStats(sourceID uint, date time.Time) (*model.NginxDailyStats, error) {
	var stats model.NginxDailyStats
	err := r.db.Where("source_id = ? AND date = ?", sourceID, date).First(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// ListDailyStats 获取日统计列表
func (r *NginxRepository) ListDailyStats(sourceID uint, startDate, endDate time.Time) ([]model.NginxDailyStats, error) {
	var stats []model.NginxDailyStats
	err := r.db.Where("source_id = ? AND date >= ? AND date <= ?", sourceID, startDate, endDate).
		Order("date DESC").Find(&stats).Error
	return stats, err
}

// GetDailyStatsRange 获取日期范围内的统计数据（所有数据源）
func (r *NginxRepository) GetDailyStatsRange(startDate, endDate time.Time) ([]model.NginxDailyStats, error) {
	var stats []model.NginxDailyStats
	err := r.db.Where("date >= ? AND date <= ?", startDate, endDate).
		Order("date DESC").Find(&stats).Error
	return stats, err
}

// ============== 小时统计操作 ==============

// CreateOrUpdateHourlyStats 创建或更新小时统计
func (r *NginxRepository) CreateOrUpdateHourlyStats(stats *model.NginxHourlyStats) error {
	return r.db.Save(stats).Error
}

// GetHourlyStats 获取小时统计数据
func (r *NginxRepository) GetHourlyStats(sourceID uint, hour time.Time) (*model.NginxHourlyStats, error) {
	var stats model.NginxHourlyStats
	err := r.db.Where("source_id = ? AND hour = ?", sourceID, hour).First(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// ListHourlyStats 获取小时统计列表
func (r *NginxRepository) ListHourlyStats(sourceID uint, startHour, endHour time.Time) ([]model.NginxHourlyStats, error) {
	var stats []model.NginxHourlyStats
	err := r.db.Where("source_id = ? AND hour >= ? AND hour <= ?", sourceID, startHour, endHour).
		Order("hour DESC").Find(&stats).Error
	return stats, err
}

// DeleteOldHourlyStats 删除过期小时统计
func (r *NginxRepository) DeleteOldHourlyStats(sourceID uint, beforeTime time.Time) error {
	return r.db.Where("source_id = ? AND hour < ?", sourceID, beforeTime).Delete(&model.NginxHourlyStats{}).Error
}

// ============== 统计查询 ==============

// GetTodayOverview 获取今日概况
func (r *NginxRepository) GetTodayOverview() (*model.OverviewStats, error) {
	overview := &model.OverviewStats{
		StatusDistribution: make(map[string]int64),
	}

	// 获取数据源统计
	r.db.Model(&model.NginxSource{}).Count(&overview.TotalSources)
	r.db.Model(&model.NginxSource{}).Where("status = ?", 1).Count(&overview.ActiveSources)

	// 获取今日统计
	today := time.Now().Truncate(24 * time.Hour)
	var dailyStats []model.NginxDailyStats
	r.db.Where("date = ?", today).Find(&dailyStats)

	for _, stats := range dailyStats {
		overview.TodayRequests += stats.TotalRequests
		overview.TodayVisitors += stats.UniqueVisitors
		overview.TodayBandwidth += stats.TotalBandwidth
		overview.StatusDistribution["2xx"] += stats.Status2xx
		overview.StatusDistribution["3xx"] += stats.Status3xx
		overview.StatusDistribution["4xx"] += stats.Status4xx
		overview.StatusDistribution["5xx"] += stats.Status5xx
	}

	// 计算错误率
	totalStatus := overview.StatusDistribution["2xx"] + overview.StatusDistribution["3xx"] +
		overview.StatusDistribution["4xx"] + overview.StatusDistribution["5xx"]
	if totalStatus > 0 {
		errorCount := overview.StatusDistribution["4xx"] + overview.StatusDistribution["5xx"]
		overview.TodayErrorRate = float64(errorCount) / float64(totalStatus) * 100
	}

	return overview, nil
}

// GetRequestsTrend 获取请求趋势
func (r *NginxRepository) GetRequestsTrend(sourceID *uint, hours int) ([]model.TrendPoint, error) {
	var trend []model.TrendPoint

	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(hours) * time.Hour)

	query := r.db.Model(&model.NginxHourlyStats{}).
		Select("hour as time, SUM(total_requests) as value").
		Where("hour >= ? AND hour <= ?", startTime, endTime).
		Group("hour").
		Order("hour ASC")

	if sourceID != nil {
		query = query.Where("source_id = ?", *sourceID)
	}

	type Result struct {
		Time  time.Time
		Value int64
	}
	var results []Result
	err := query.Find(&results).Error
	if err != nil {
		return nil, err
	}

	for _, r := range results {
		trend = append(trend, model.TrendPoint{
			Time:  r.Time.Format("15:04"),
			Value: r.Value,
		})
	}

	return trend, nil
}

// GetTopURIs 获取 Top URI
func (r *NginxRepository) GetTopURIs(sourceID uint, startTime, endTime time.Time, limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	type URICount struct {
		URI   string
		Count int64
	}
	var uriCounts []URICount

	err := r.db.Model(&model.NginxAccessLog{}).
		Select("uri, COUNT(*) as count").
		Where("source_id = ? AND timestamp >= ? AND timestamp <= ?", sourceID, startTime, endTime).
		Group("uri").
		Order("count DESC").
		Limit(limit).
		Find(&uriCounts).Error

	if err != nil {
		return nil, err
	}

	for _, uc := range uriCounts {
		results = append(results, map[string]interface{}{
			"uri":   uc.URI,
			"count": uc.Count,
		})
	}

	return results, nil
}

// GetTopIPs 获取 Top IP
func (r *NginxRepository) GetTopIPs(sourceID uint, startTime, endTime time.Time, limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	type IPCount struct {
		RemoteAddr string
		Count      int64
	}
	var ipCounts []IPCount

	err := r.db.Model(&model.NginxAccessLog{}).
		Select("remote_addr, COUNT(*) as count").
		Where("source_id = ? AND timestamp >= ? AND timestamp <= ?", sourceID, startTime, endTime).
		Group("remote_addr").
		Order("count DESC").
		Limit(limit).
		Find(&ipCounts).Error

	if err != nil {
		return nil, err
	}

	for _, ic := range ipCounts {
		results = append(results, map[string]interface{}{
			"ip":    ic.RemoteAddr,
			"count": ic.Count,
		})
	}

	return results, nil
}
