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
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ydcloud-dy/opshub/pkg/response"
	"github.com/ydcloud-dy/opshub/plugins/nginx/model"
	"github.com/ydcloud-dy/opshub/plugins/nginx/repository"
	"gorm.io/gorm"
)

type Handler struct {
	db   *gorm.DB
	repo *repository.NginxRepository
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		db:   db,
		repo: repository.NewNginxRepository(db),
	}
}

// ==================== 数据源管理 ====================

// ListSources 获取数据源列表
// @Summary 获取数据源列表
// @Description 分页获取Nginx数据源列表
// @Tags Nginx统计-数据源
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param type query string false "数据源类型"
// @Param status query int false "状态"
// @Success 200 {object} response.Response "获取成功"
// @Router /nginx/sources [get]
func (h *Handler) ListSources(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	sourceType := c.Query("type")
	statusStr := c.Query("status")

	var status *int
	if statusStr != "" {
		s, _ := strconv.Atoi(statusStr)
		status = &s
	}

	sources, total, err := h.repo.ListSources(page, pageSize, sourceType, status)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "获取数据源列表失败")
		return
	}

	response.Success(c, gin.H{
		"list":     sources,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetSource 获取数据源详情
// @Summary 获取数据源详情
// @Description 获取指定数据源的详细信息
// @Tags Nginx统计-数据源
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "数据源ID"
// @Success 200 {object} response.Response "获取成功"
// @Failure 404 {object} response.Response "数据源不存在"
// @Router /nginx/sources/{id} [get]
func (h *Handler) GetSource(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	source, err := h.repo.GetSourceByID(uint(id))
	if err != nil {
		response.ErrorCode(c, http.StatusNotFound, "数据源不存在")
		return
	}
	response.Success(c, source)
}

// CreateSource 创建数据源
// @Summary 创建数据源
// @Description 创建新的Nginx数据源
// @Tags Nginx统计-数据源
// @Accept json
// @Produce json
// @Security Bearer
// @Param body body model.NginxSource true "数据源信息"
// @Success 200 {object} response.Response "创建成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /nginx/sources [post]
func (h *Handler) CreateSource(c *gin.Context) {
	var source model.NginxSource
	if err := c.ShouldBindJSON(&source); err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	if err := h.repo.CreateSource(&source); err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "创建数据源失败: "+err.Error())
		return
	}
	response.Success(c, source)
}

// UpdateSource 更新数据源
// @Summary 更新数据源
// @Description 更新指定数据源的信息
// @Tags Nginx统计-数据源
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "数据源ID"
// @Param body body model.NginxSource true "数据源信息"
// @Success 200 {object} response.Response "更新成功"
// @Failure 404 {object} response.Response "数据源不存在"
// @Router /nginx/sources/{id} [put]
func (h *Handler) UpdateSource(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	source, err := h.repo.GetSourceByID(uint(id))
	if err != nil {
		response.ErrorCode(c, http.StatusNotFound, "数据源不存在")
		return
	}

	if err := c.ShouldBindJSON(source); err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	if err := h.repo.UpdateSource(source); err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "更新数据源失败: "+err.Error())
		return
	}
	response.Success(c, source)
}

// DeleteSource 删除数据源
// @Summary 删除数据源
// @Description 删除指定的数据源
// @Tags Nginx统计-数据源
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "数据源ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /nginx/sources/{id} [delete]
func (h *Handler) DeleteSource(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.repo.DeleteSource(uint(id)); err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "删除数据源失败")
		return
	}
	response.Success(c, nil)
}

// ==================== 概况统计 ====================

// GetOverview 获取概况统计
// @Summary 获取概况统计
// @Description 获取今日概况统计数据
// @Tags Nginx统计-概况
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response "获取成功"
// @Router /nginx/overview [get]
func (h *Handler) GetOverview(c *gin.Context) {
	overview, err := h.repo.GetTodayOverview()
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "获取概况失败")
		return
	}

	// 获取请求趋势（最近24小时）
	trend, err := h.repo.GetRequestsTrend(nil, 24)
	if err == nil {
		overview.RequestsTrend = trend
	}

	response.Success(c, overview)
}

// GetRequestsTrend 获取请求趋势
// @Summary 获取请求趋势
// @Description 获取指定时间范围的请求趋势数据
// @Tags Nginx统计-概况
// @Accept json
// @Produce json
// @Security Bearer
// @Param sourceId query int false "数据源ID"
// @Param hours query int false "小时数" default(24)
// @Success 200 {object} response.Response "获取成功"
// @Router /nginx/overview/trend [get]
func (h *Handler) GetRequestsTrend(c *gin.Context) {
	hours, _ := strconv.Atoi(c.DefaultQuery("hours", "24"))
	sourceIDStr := c.Query("sourceId")

	var sourceID *uint
	if sourceIDStr != "" {
		id, _ := strconv.ParseUint(sourceIDStr, 10, 32)
		sid := uint(id)
		sourceID = &sid
	}

	trend, err := h.repo.GetRequestsTrend(sourceID, hours)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "获取趋势数据失败")
		return
	}
	response.Success(c, trend)
}

// ==================== 数据日报 ====================

// GetDailyReport 获取日报数据
// @Summary 获取日报数据
// @Description 获取指定日期范围的统计日报
// @Tags Nginx统计-数据日报
// @Accept json
// @Produce json
// @Security Bearer
// @Param sourceId query int true "数据源ID"
// @Param startDate query string false "开始日期"
// @Param endDate query string false "结束日期"
// @Success 200 {object} response.Response "获取成功"
// @Router /nginx/daily-report [get]
func (h *Handler) GetDailyReport(c *gin.Context) {
	sourceID, _ := strconv.ParseUint(c.Query("sourceId"), 10, 32)
	startDateStr := c.Query("startDate")
	endDateStr := c.Query("endDate")

	// 默认最近7天
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -7)

	if startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = t
		}
	}
	if endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = t
		}
	}

	var stats []model.NginxDailyStats
	var err error

	if sourceID > 0 {
		stats, err = h.repo.ListDailyStats(uint(sourceID), startDate, endDate)
	} else {
		stats, err = h.repo.GetDailyStatsRange(startDate, endDate)
	}

	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "获取日报数据失败")
		return
	}
	response.Success(c, stats)
}

// ==================== 实时统计 ====================

// GetRealTimeStats 获取实时统计
// @Summary 获取实时统计
// @Description 获取最近N小时的小时级统计数据
// @Tags Nginx统计-实时
// @Accept json
// @Produce json
// @Security Bearer
// @Param sourceId query int true "数据源ID"
// @Param hours query int false "小时数" default(6)
// @Success 200 {object} response.Response "获取成功"
// @Router /nginx/realtime [get]
func (h *Handler) GetRealTimeStats(c *gin.Context) {
	sourceID, _ := strconv.ParseUint(c.Query("sourceId"), 10, 32)
	hours, _ := strconv.Atoi(c.DefaultQuery("hours", "6"))

	if sourceID == 0 {
		response.ErrorCode(c, http.StatusBadRequest, "请指定数据源ID")
		return
	}

	endHour := time.Now().Truncate(time.Hour)
	startHour := endHour.Add(-time.Duration(hours) * time.Hour)

	stats, err := h.repo.ListHourlyStats(uint(sourceID), startHour, endHour)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "获取实时数据失败")
		return
	}
	response.Success(c, stats)
}

// ==================== 访问明细 ====================

// ListAccessLogs 获取访问日志列表
// @Summary 获取访问日志列表
// @Description 分页获取访问日志列表
// @Tags Nginx统计-访问明细
// @Accept json
// @Produce json
// @Security Bearer
// @Param sourceId query int true "数据源ID"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param startTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Param remoteAddr query string false "客户端IP"
// @Param uri query string false "请求URI"
// @Param status query int false "状态码"
// @Param method query string false "请求方法"
// @Param host query string false "请求主机"
// @Success 200 {object} response.Response "获取成功"
// @Router /nginx/access-logs [get]
func (h *Handler) ListAccessLogs(c *gin.Context) {
	sourceID, _ := strconv.ParseUint(c.Query("sourceId"), 10, 32)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	if sourceID == 0 {
		response.ErrorCode(c, http.StatusBadRequest, "请指定数据源ID")
		return
	}

	// 解析时间参数
	var startTime, endTime *time.Time
	if st := c.Query("startTime"); st != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", st); err == nil {
			startTime = &t
		}
	}
	if et := c.Query("endTime"); et != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", et); err == nil {
			endTime = &t
		}
	}

	// 构建过滤条件
	filters := make(map[string]interface{})
	if v := c.Query("remoteAddr"); v != "" {
		filters["remoteAddr"] = v
	}
	if v := c.Query("uri"); v != "" {
		filters["uri"] = v
	}
	if v := c.Query("status"); v != "" {
		if s, err := strconv.Atoi(v); err == nil {
			filters["status"] = s
		}
	}
	if v := c.Query("method"); v != "" {
		filters["method"] = v
	}
	if v := c.Query("host"); v != "" {
		filters["host"] = v
	}

	logs, total, err := h.repo.ListAccessLogs(uint(sourceID), page, pageSize, startTime, endTime, filters)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "获取访问日志失败")
		return
	}

	response.Success(c, gin.H{
		"list":     logs,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetTopURIs 获取 Top URI
// @Summary 获取 Top URI
// @Description 获取访问量最高的URI列表
// @Tags Nginx统计-访问明细
// @Accept json
// @Produce json
// @Security Bearer
// @Param sourceId query int true "数据源ID"
// @Param startTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Param limit query int false "数量限制" default(10)
// @Success 200 {object} response.Response "获取成功"
// @Router /nginx/access-logs/top-uris [get]
func (h *Handler) GetTopURIs(c *gin.Context) {
	sourceID, _ := strconv.ParseUint(c.Query("sourceId"), 10, 32)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if sourceID == 0 {
		response.ErrorCode(c, http.StatusBadRequest, "请指定数据源ID")
		return
	}

	// 默认最近24小时
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	if st := c.Query("startTime"); st != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", st); err == nil {
			startTime = t
		}
	}
	if et := c.Query("endTime"); et != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", et); err == nil {
			endTime = t
		}
	}

	results, err := h.repo.GetTopURIs(uint(sourceID), startTime, endTime, limit)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "获取Top URI失败")
		return
	}
	response.Success(c, results)
}

// GetTopIPs 获取 Top IP
// @Summary 获取 Top IP
// @Description 获取访问量最高的IP列表
// @Tags Nginx统计-访问明细
// @Accept json
// @Produce json
// @Security Bearer
// @Param sourceId query int true "数据源ID"
// @Param startTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Param limit query int false "数量限制" default(10)
// @Success 200 {object} response.Response "获取成功"
// @Router /nginx/access-logs/top-ips [get]
func (h *Handler) GetTopIPs(c *gin.Context) {
	sourceID, _ := strconv.ParseUint(c.Query("sourceId"), 10, 32)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if sourceID == 0 {
		response.ErrorCode(c, http.StatusBadRequest, "请指定数据源ID")
		return
	}

	// 默认最近24小时
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	if st := c.Query("startTime"); st != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", st); err == nil {
			startTime = t
		}
	}
	if et := c.Query("endTime"); et != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", et); err == nil {
			endTime = t
		}
	}

	results, err := h.repo.GetTopIPs(uint(sourceID), startTime, endTime, limit)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "获取Top IP失败")
		return
	}
	response.Success(c, results)
}
