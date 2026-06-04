package dto

import "time"

// DashboardStatsResponse 看板统计响应
type DashboardStatsResponse struct {
	// 概览数据
	TotalIssues      int64   `json:"total_issues"`
	NewIssues        int64   `json:"new_issues"`
	ResolvedIssues   int64   `json:"resolved_issues"`
	UnresolvedIssues int64   `json:"unresolved_issues"`
	P0P1Issues       int64   `json:"p0_p1_issues"`
	AvgResolveTime   float64 `json:"avg_resolve_time_hours"` // 小时
	RepeatedIssues   int64   `json:"repeated_issues"`

	// AI 数据
	AIAnalysisCount int64   `json:"ai_analysis_count"`
	AIAdoptedCount  int64   `json:"ai_adopted_count"`
	AIAdoptRate     float64 `json:"ai_adopt_rate"` // 采纳率

	// 分布数据
	IssueTypeDistribution []IssueTypeDistItem `json:"issue_type_distribution"`
	ServiceIssueRanking   []ServiceRankItem   `json:"service_issue_ranking"`

	// 项目问题排行（仅系统管理员可见）
	ProjectIssueRanking []ProjectRankItem `json:"project_issue_ranking"`
}

// IssueTypeDistItem 问题类型分布项
type IssueTypeDistItem struct {
	IssueType string `json:"issue_type"`
	Count     int64  `json:"count"`
	Percent   string `json:"percent"` // 百分比字符串
}

// ServiceRankItem 服务排行项
type ServiceRankItem struct {
	ServiceID   uint64 `json:"service_id"`
	ServiceName string `json:"service_name"`
	Count       int64  `json:"count"`
}

// ProjectRankItem 项目排行项
type ProjectRankItem struct {
	ProjectID   uint64 `json:"project_id"`
	ProjectName string `json:"project_name"`
	Count       int64  `json:"count"`
}

// DashboardTrendResponse 趋势图响应
type DashboardTrendResponse struct {
	Dates  []string         `json:"dates"`
	Series []TrendSeriesItem `json:"series"`
}

// TrendSeriesItem 趋势系列
type TrendSeriesItem struct {
	Name string  `json:"name"`
	Data []int64 `json:"data"`
}

// GenerateDailyStatRequest 生成每日统计请求
type GenerateDailyStatRequest struct {
	Date string `json:"date"` // 日期，为空则统计昨天
}

// DailyStatRecord 每日统计记录
type DailyStatRecord struct {
	ID                uint64    `json:"id"`
	StatDate          string    `json:"stat_date"`
	TotalIssues       int64     `json:"total_issues"`
	NewIssues         int64     `json:"new_issues"`
	ResolvedIssues    int64     `json:"resolved_issues"`
	UnresolvedIssues  int64     `json:"unresolved_issues"`
	P0P1Issues        int64     `json:"p0_p1_issues"`
	AvgResolveMinutes int64     `json:"avg_resolve_minutes"`
	RepeatedIssues    int64     `json:"repeated_issues"`
	AIAnalysisCount   int64     `json:"ai_analysis_count"`
	AIAdoptedCount    int64     `json:"ai_adopted_count"`
	CreatedAt         time.Time `json:"created_at"`
}

// TrendQueryRequest 趋势查询请求
type TrendQueryRequest struct {
	StartDate string `form:"start_date" binding:"required"` // 开始日期
	EndDate   string `form:"end_date" binding:"required"`   // 结束日期
	ProjectID uint64 `form:"project_id"`                    // 项目ID（可选）
}
