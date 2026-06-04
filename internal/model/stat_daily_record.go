package model

import (
	"time"
)

// StatDailyRecord 每日统计快照
type StatDailyRecord struct {
	ID                uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	StatDate          string    `gorm:"type:varchar(16);not null;index" json:"stat_date"` // 统计日期 YYYY-MM-DD
	ProjectID         *uint64   `gorm:"index" json:"project_id"`                           // NULL 表示全局统计

	// 问题统计
	TotalIssues       int64 `gorm:"default:0" json:"total_issues"`         // 问题总数
	NewIssues         int64 `gorm:"default:0" json:"new_issues"`           // 新增问题
	ResolvedIssues    int64 `gorm:"default:0" json:"resolved_issues"`      // 已解决问题
	UnresolvedIssues  int64 `gorm:"default:0" json:"unresolved_issues"`    // 未闭环问题
	P0P1Issues        int64 `gorm:"default:0" json:"p0_p1_issues"`         // 高优问题(P0+P1)
	AvgResolveMinutes int64 `gorm:"default:0" json:"avg_resolve_minutes"`  // 平均处理时长(分钟)
	RepeatedIssues    int64 `gorm:"default:0" json:"repeated_issues"`      // 重复问题数

	// AI 统计
	AIAnalysisCount int64 `gorm:"default:0" json:"ai_analysis_count"` // AI分析次数
	AIAdoptedCount  int64 `gorm:"default:0" json:"ai_adopted_count"`  // AI建议采纳数

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name
func (StatDailyRecord) TableName() string {
	return "stat_daily_records"
}

// IssueTypeStat 问题类型分布
type IssueTypeStat struct {
	IssueType string `json:"issue_type"`
	Count     int64  `json:"count"`
}

// ServiceIssueStat 服务问题排行
type ServiceIssueStat struct {
	ServiceID   uint64 `json:"service_id"`
	ServiceName string `json:"service_name"`
	Count       int64  `json:"count"`
}

// ProjectIssueStat 项目问题排行
type ProjectIssueStat struct {
	ProjectID   uint64 `json:"project_id"`
	ProjectName string `json:"project_name"`
	Count       int64  `json:"count"`
}
