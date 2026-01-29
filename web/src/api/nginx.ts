import request from '@/utils/request'

// ==================== 类型定义 ====================

export interface NginxSource {
  id?: number
  name: string
  type: 'host' | 'k8s_ingress'
  description?: string
  status: number
  hostId?: number
  logPath?: string
  logFormat?: string
  clusterId?: number
  namespace?: string
  ingressName?: string
  collectInterval: number
  retentionDays: number
  createdAt?: string
  updatedAt?: string
}

export interface NginxAccessLog {
  id: number
  sourceId: number
  timestamp: string
  remoteAddr: string
  remoteUser?: string
  request: string
  method: string
  uri: string
  protocol: string
  status: number
  bodyBytesSent: number
  httpReferer?: string
  httpUserAgent?: string
  requestTime: number
  upstreamTime?: number
  host: string
  ingressName?: string
  serviceName?: string
  createdAt: string
}

export interface NginxDailyStats {
  id: number
  sourceId: number
  date: string
  totalRequests: number
  uniqueVisitors: number
  totalBandwidth: number
  avgResponseTime: number
  status2xx: number
  status3xx: number
  status4xx: number
  status5xx: number
  topURIs?: string
  topIPs?: string
  topReferers?: string
  topUserAgents?: string
  createdAt: string
  updatedAt: string
}

export interface NginxHourlyStats {
  id: number
  sourceId: number
  hour: string
  totalRequests: number
  uniqueVisitors: number
  totalBandwidth: number
  avgResponseTime: number
  status2xx: number
  status3xx: number
  status4xx: number
  status5xx: number
  createdAt: string
}

export interface OverviewStats {
  totalSources: number
  activeSources: number
  todayRequests: number
  todayVisitors: number
  todayBandwidth: number
  todayErrorRate: number
  requestsTrend?: TrendPoint[]
  bandwidthTrend?: TrendPoint[]
  statusDistribution: Record<string, number>
}

export interface TrendPoint {
  time: string
  value: number
}

// ==================== 数据源管理 ====================

// 获取数据源列表
export const getNginxSources = (params?: { page?: number; pageSize?: number; type?: string; status?: number }) => {
  return request.get('/api/v1/plugins/nginx/sources', { params })
}

// 获取数据源详情
export const getNginxSource = (id: number) => {
  return request.get(`/api/v1/plugins/nginx/sources/${id}`)
}

// 创建数据源
export const createNginxSource = (data: NginxSource) => {
  return request.post('/api/v1/plugins/nginx/sources', data)
}

// 更新数据源
export const updateNginxSource = (id: number, data: NginxSource) => {
  return request.put(`/api/v1/plugins/nginx/sources/${id}`, data)
}

// 删除数据源
export const deleteNginxSource = (id: number) => {
  return request.delete(`/api/v1/plugins/nginx/sources/${id}`)
}

// ==================== 概况统计 ====================

// 获取概况统计
export const getNginxOverview = () => {
  return request.get('/api/v1/plugins/nginx/overview')
}

// 获取请求趋势
export const getNginxRequestsTrend = (params?: { sourceId?: number; hours?: number }) => {
  return request.get('/api/v1/plugins/nginx/overview/trend', { params })
}

// ==================== 数据日报 ====================

// 获取日报数据
export const getNginxDailyReport = (params: { sourceId?: number; startDate?: string; endDate?: string }) => {
  return request.get('/api/v1/plugins/nginx/daily-report', { params })
}

// ==================== 实时统计 ====================

// 获取实时统计
export const getNginxRealTimeStats = (params: { sourceId: number; hours?: number }) => {
  return request.get('/api/v1/plugins/nginx/realtime', { params })
}

// ==================== 访问明细 ====================

// 获取访问日志列表
export const getNginxAccessLogs = (params: {
  sourceId: number
  page?: number
  pageSize?: number
  startTime?: string
  endTime?: string
  remoteAddr?: string
  uri?: string
  status?: number
  method?: string
  host?: string
}) => {
  return request.get('/api/v1/plugins/nginx/access-logs', { params })
}

// 获取 Top URI
export const getNginxTopURIs = (params: { sourceId: number; startTime?: string; endTime?: string; limit?: number }) => {
  return request.get('/api/v1/plugins/nginx/access-logs/top-uris', { params })
}

// 获取 Top IP
export const getNginxTopIPs = (params: { sourceId: number; startTime?: string; endTime?: string; limit?: number }) => {
  return request.get('/api/v1/plugins/nginx/access-logs/top-ips', { params })
}

// ==================== 日志采集 ====================

// 手动触发日志采集
export const collectNginxLogs = (sourceId?: number) => {
  const params = sourceId ? { sourceId } : {}
  return request.post('/api/v1/plugins/nginx/collect', null, { params })
}
