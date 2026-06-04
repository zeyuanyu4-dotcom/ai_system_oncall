package model

import "time"

// SimulatedLog 模拟日志模型
type SimulatedLog struct {
	ID           uint64     `gorm:"primaryKey" json:"id"`
	ProjectID    uint64     `gorm:"index;not null" json:"project_id"`
	ServiceID    uint64     `gorm:"index;not null" json:"service_id"`
	ServiceName  string     `gorm:"size:128;not null" json:"service_name"`
	Environment  string     `gorm:"size:32;not null;index" json:"environment"`
	LogLevel     string     `gorm:"size:16;not null;index" json:"log_level"`
	TraceID      string     `gorm:"size:64;index" json:"trace_id"`
	RequestPath  string     `gorm:"size:256" json:"request_path"`
	ErrorCode    string     `gorm:"size:32" json:"error_code"`
	LogContent   string     `gorm:"type:text;not null" json:"log_content"`
	StackTrace   string     `gorm:"type:text" json:"stack_trace"`
	OccurredAt   time.Time  `gorm:"index;not null" json:"occurred_at"`
	IssueID      *uint64    `gorm:"index" json:"issue_id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	// 关联
	Project *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Service *Service `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	Issue   *Issue   `gorm:"foreignKey:IssueID" json:"issue,omitempty"`
}

// TableName specifies the table name for SimulatedLog
func (SimulatedLog) TableName() string {
	return "simulated_logs"
}
