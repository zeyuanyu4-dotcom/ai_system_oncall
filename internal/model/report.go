package model

import (
	"time"

	"gorm.io/gorm"
)

// Report 报告模型
type Report struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	ReportType  string        `gorm:"type:varchar(32);not null;index" json:"report_type"` // daily, weekly, incident
	ReportDate  string        `gorm:"type:varchar(16);index" json:"report_date"`          // 日期: 2026-06-04
	ReportWeek  string        `gorm:"type:varchar(16);index" json:"report_week"`          // 周: 2026-W23
	Title       string        `gorm:"type:varchar(255);not null" json:"title"`
	Summary     string        `gorm:"type:text" json:"summary"`
	Content     string        `gorm:"type:longtext" json:"content"` // JSON格式的报告内容
	IssueCount  int           `gorm:"default:0" json:"issue_count"`
	CreatorID   uint64        `gorm:"not null;index" json:"creator_id"`
	IsAuto      bool          `gorm:"default:false" json:"is_auto"` // 是否自动生成
	Status      string        `gorm:"type:varchar(32);not null;default:generated" json:"status"` // generating, generated, failed
	ErrorMsg    string        `gorm:"type:text" json:"error_msg"`
	CreatedAt   time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Creator *User `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
}

// TableName specifies the table name for Report model
func (Report) TableName() string {
	return "reports"
}

// ReportType constants
const (
	ReportTypeDaily    = "daily"
	ReportTypeWeekly   = "weekly"
	ReportTypeIncident = "incident"
)

// ReportStatus constants
const (
	ReportStatusGenerating = "generating"
	ReportStatusGenerated = "generated"
	ReportStatusFailed    = "failed"
)

// DailyReportStats 每日报告统计数据
type DailyReportStats struct {
	TotalIssues     int64 `json:"total_issues"`
	PendingIssues  int64 `json:"pending_issues"`
	ResolvedIssues int64 `json:"resolved_issues"`
	P0Issues       int64 `json:"p0_issues"`
	P1Issues       int64 `json:"p1_issues"`
	P2Issues       int64 `json:"p2_issues"`
	AvgResolveTime int   `json:"avg_resolve_time_minutes"` // 分钟
}

// WeeklyReportStats 周报统计数据
type WeeklyReportStats struct {
	TotalIssues      int64         `json:"total_issues"`
	ResolvedIssues  int64         `json:"resolved_issues"`
	UnresolvedIssues int64         `json:"unresolved_issues"`
	CriticalIssues   int64         `json:"critical_issues"`
	AvgResolveTime   int           `json:"avg_resolve_time_minutes"`
	TopServices      []ServiceStat `json:"top_services"`
	Trend            string        `json:"trend"` // improving, worsening, stable
}

// ServiceStat 服务统计
type ServiceStat struct {
	ServiceName string `json:"service_name"`
	IssueCount int    `json:"issue_count"`
}

// IncidentReviewStats 事件复盘统计数据
type IncidentReviewStats struct {
	IssueID       uint64   `json:"issue_id"`
	IssueNo       string   `json:"issue_no"`
	Title         string   `json:"title"`
	Priority      string   `json:"priority"`
	IncidentTime  string   `json:"incident_time"`
	Duration      int      `json:"duration_minutes"`
	RootCause     string   `json:"root_cause"`
	AffectedServices []string `json:"affected_services"`
	Impact        string   `json:"impact"`
	Resolution    string   `json:"resolution"`
	LessonsLearned []string `json:"lessons_learned"`
	ActionItems   []string `json:"action_items"`
}