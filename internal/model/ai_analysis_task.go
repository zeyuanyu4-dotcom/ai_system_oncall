package model

import (
	"time"
)

// AIAnalysisTask AI分析任务模型
type AIAnalysisTask struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	IssueID      uint64    `gorm:"not null;index" json:"issue_id"`
	Status       string    `gorm:"type:varchar(32);not null;default:pending;index" json:"status"` // pending, running, completed, failed, cancelled
	Progress     string    `gorm:"type:varchar(32)" json:"progress"`                               // "3/8"
	CurrentStep  string    `gorm:"type:varchar(255)" json:"current_step"`                          // 当前执行步骤描述
	ToolCalls    string    `gorm:"type:longtext" json:"tool_calls"`                                // JSON，工具调用记录
	Result       string    `gorm:"type:longtext" json:"result"`                                    // JSON，最终分析结果
	ErrorMessage string    `gorm:"type:text" json:"error_message"`                                 // 错误信息
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Issue *Issue `gorm:"foreignKey:IssueID" json:"issue,omitempty"`
}

// TableName specifies the table name for AIAnalysisTask model
func (AIAnalysisTask) TableName() string {
	return "ai_analysis_tasks"
}

// Task status constants
const (
	TaskStatusPending    = "pending"
	TaskStatusRunning    = "running"
	TaskStatusCompleted  = "completed"
	TaskStatusFailed     = "failed"
	TaskStatusCancelled  = "cancelled"
)

// ToolCallRecord 工具调用记录
type ToolCallRecord struct {
	Step        int    `json:"step"`
	ToolName    string `json:"tool_name"`
	Input       string `json:"input"`
	Output      string `json:"output"`
	Thought     string `json:"thought"`
	ExecutedAt  string `json:"executed_at"`
	DurationMs  int64  `json:"duration_ms"`
}

// AgentResult Agent分析结果
type AgentResult struct {
	Summary         string            `json:"summary"`
	IssueType       string            `json:"issue_type"`
	RelatedServices []string          `json:"related_services"`
	SuspectedCause  string            `json:"suspected_cause"`
	Evidence        []EvidenceItem    `json:"evidence"`
	Suggestions     []string          `json:"suggestions"`
	MissingInfo     []string          `json:"missing_info"`
	NextSteps       []string          `json:"next_steps"`
	Confidence      float64           `json:"confidence"`
	ToolCalls       []ToolCallRecord  `json:"tool_calls"`
}

// EvidenceItem 证据项
type EvidenceItem struct {
	Source   string `json:"source"`   // 来源工具
	Content  string `json:"content"`  // 证据内容
	Relevance string `json:"relevance"` // 相关性说明
}
