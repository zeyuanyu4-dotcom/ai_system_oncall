package dto

import "time"

// GenerateDailyReportRequest 手动生成日报请求
type GenerateDailyReportRequest struct {
	Date string `json:"date" binding:"required"` // 格式: 2026-06-04，默认今天
}

// GenerateWeeklyReportRequest 手动生成周报请求
type GenerateWeeklyReportRequest struct {
	WeekStart string `json:"week_start" binding:"required"` // 格式: 2026-06-01 (周一)
}

// GenerateIncidentReviewRequest 生成事件复盘请求
type GenerateIncidentReviewRequest struct {
	IssueIDs []uint64 `json:"issue_ids" binding:"required,min=1"` // 问题单ID列表
}

// ReportInfo 报告信息
type ReportInfo struct {
	ID         uint64     `json:"id"`
	ReportType string     `json:"report_type"`
	ReportDate string     `json:"report_date"`
	ReportWeek string     `json:"report_week"`
	Title      string     `json:"title"`
	Summary    string     `json:"summary"`
	Content    string     `json:"content"`
	IssueCount int        `json:"issue_count"`
	IsAuto     bool       `json:"is_auto"`
	Status     string     `json:"status"`
	CreatorID  uint64     `json:"creator_id"`
	Creator    *UserInfo  `json:"creator,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// DailyReportContent 每日报告内容
type DailyReportContent struct {
	Date       string                `json:"date"`
	Stats      DailyReportStatsDTO   `json:"stats"`
	Issues     []IssueBriefDTO       `json:"issues"`
	Analysis   string                `json:"analysis"`
}

// DailyReportStatsDTO 每日报告统计
type DailyReportStatsDTO struct {
	TotalIssues     int64 `json:"total_issues"`
	PendingIssues  int64 `json:"pending_issues"`
	ResolvedIssues int64 `json:"resolved_issues"`
	P0Issues       int64 `json:"p0_issues"`
	P1Issues       int64 `json:"p1_issues"`
	P2Issues       int64 `json:"p2_issues"`
}

// WeeklyReportContent 周报内容
type WeeklyReportContent struct {
	WeekStart      string                  `json:"week_start"`
	WeekEnd        string                  `json:"week_end"`
	Stats          WeeklyReportStatsDTO    `json:"stats"`
	DailySummaries []DailyReportSummaryDTO `json:"daily_summaries"` // 每日汇总
	Issues         []IssueBriefDTO         `json:"issues"`
	TopServices    []ServiceStatDTO        `json:"top_services"`
	Analysis       string                  `json:"analysis"`
}

// DailyReportSummaryDTO 每日报告摘要
type DailyReportSummaryDTO struct {
	Date           string `json:"date"`
	TotalIssues    int64  `json:"total_issues"`
	P0Issues       int64  `json:"p0_issues"`
	P1Issues       int64  `json:"p1_issues"`
	P2Issues       int64  `json:"p2_issues"`
	ResolvedIssues int64  `json:"resolved_issues"`
}

// WeeklyReportStatsDTO 周报统计
type WeeklyReportStatsDTO struct {
	TotalIssues      int64  `json:"total_issues"`
	ResolvedIssues  int64  `json:"resolved_issues"`
	UnresolvedIssues int64  `json:"unresolved_issues"`
	CriticalIssues   int64  `json:"critical_issues"`
	Trend            string `json:"trend"`
}

// ServiceStatDTO 服务统计
type ServiceStatDTO struct {
	ServiceName string `json:"service_name"`
	IssueCount int    `json:"issue_count"`
}

// IncidentReviewContent 事件复盘内容
type IncidentReviewContent struct {
	IssueCount       int                  `json:"issue_count"`
	Reviews          []IncidentReviewDTO  `json:"reviews"`
	Summary          string               `json:"summary"`
}

// IncidentReviewDTO 事件复盘详情
type IncidentReviewDTO struct {
	IssueID           uint64   `json:"issue_id"`
	IssueNo           string   `json:"issue_no"`
	Title             string   `json:"title"`
	Priority          string   `json:"priority"`
	IncidentTime      string   `json:"incident_time"`
	Duration          int      `json:"duration_minutes"`
	RootCause         string   `json:"root_cause"`
	AffectedServices  []string `json:"affected_services"`
	Impact            string   `json:"impact"`
	Resolution        string   `json:"resolution"`
	LessonsLearned    []string `json:"lessons_learned"`
	ActionItems       []string `json:"action_items"`
}

// IssueBriefDTO 简要问题信息
type IssueBriefDTO struct {
	ID          uint64 `json:"id"`
	IssueNo     string `json:"issue_no"`
	Title       string `json:"title"`
	Priority    string `json:"priority"`
	Status      string `json:"status"`
	ServiceName string `json:"service_name"`
	CreatedAt   string `json:"created_at"`
}

// ReportListRequest 报告列表请求
type ReportListRequest struct {
	ReportType string `form:"report_type"`
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"page_size,default=10"`
}

// ReportListResponse 报告列表响应
type ReportListResponse struct {
	Reports   []ReportInfo `json:"reports"`
	Total     int64        `json:"total"`
	Page      int          `json:"page"`
	PageSize  int          `json:"page_size"`
}

// ReportResponse 报告生成响应
type ReportResponse struct {
	ID        uint64 `json:"id"`
	ReportType string `json:"report_type"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}