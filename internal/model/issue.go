package model

import (
	"time"

	"gorm.io/gorm"
)

// Issue 问题单模型
type Issue struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	IssueNo      string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"issue_no"`
	Title        string         `gorm:"type:varchar(255);not null" json:"title"`
	Description  string         `gorm:"type:text" json:"description"`
	ProjectID    uint64         `gorm:"not null;index" json:"project_id"`
	ServiceID    *uint64        `gorm:"index" json:"service_id"`
	IssueType    string         `gorm:"type:varchar(32);not null;default:other" json:"issue_type"`
	Priority     string         `gorm:"type:varchar(16);not null;default:P2" json:"priority"`
	Environment  string         `gorm:"type:varchar(32)" json:"environment"`
	Status       string         `gorm:"type:varchar(32);not null;default:pending_analysis;index" json:"status"`
	ImpactScope  string         `gorm:"type:varchar(255)" json:"impact_scope"`
	ErrorMessage string         `gorm:"type:text" json:"error_message"`
	LogExcerpt   string         `gorm:"type:text" json:"log_excerpt"`
	CreatorID    uint64         `gorm:"not null;index" json:"creator_id"`
	AssigneeID   *uint64        `gorm:"index" json:"assignee_id"`
	AISummary    string         `gorm:"type:text" json:"ai_summary"`
	AIAnalysis   string         `gorm:"type:text" json:"ai_analysis"`
	RootCause    string         `gorm:"type:text" json:"root_cause"`
	Solution     string         `gorm:"type:text" json:"solution"`
	ResolvedAt   *time.Time     `json:"resolved_at"`
	ClosedAt     *time.Time     `json:"closed_at"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Project  *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Service  *Service `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	Creator  *User    `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	Assignee *User    `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	Comments []IssueComment `gorm:"foreignKey:IssueID" json:"comments,omitempty"`
}

// TableName specifies the table name for Issue model
func (Issue) TableName() string {
	return "issues"
}

// IssueComment 问题评论模型
type IssueComment struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	IssueID     uint64         `gorm:"not null;index" json:"issue_id"`
	UserID      *uint64        `gorm:"index" json:"user_id"`
	CommentType string         `gorm:"type:varchar(32);not null;default:comment" json:"comment_type"`
	Content     string         `gorm:"type:text;not null" json:"content"`
	Visibility  string         `gorm:"type:varchar(32);not null;default:public" json:"visibility"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Issue *Issue `gorm:"foreignKey:IssueID" json:"issue,omitempty"`
	User  *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName specifies the table name for IssueComment model
func (IssueComment) TableName() string {
	return "issue_comments"
}

// IssueStatusLog 状态流转日志模型
type IssueStatusLog struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	IssueID    uint64    `gorm:"not null;index" json:"issue_id"`
	FromStatus string    `gorm:"type:varchar(32)" json:"from_status"`
	ToStatus   string    `gorm:"type:varchar(32);not null" json:"to_status"`
	OperatorID uint64    `gorm:"not null" json:"operator_id"`
	Reason     string    `gorm:"type:varchar(255)" json:"reason"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Issue    *Issue `gorm:"foreignKey:IssueID" json:"issue,omitempty"`
	Operator *User  `gorm:"foreignKey:OperatorID" json:"operator,omitempty"`
}

// TableName specifies the table name for IssueStatusLog model
func (IssueStatusLog) TableName() string {
	return "issue_status_logs"
}

// IssueOperationLog 操作日志模型
type IssueOperationLog struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	IssueID          uint64    `gorm:"not null;index" json:"issue_id"`
	OperatorID       uint64    `gorm:"not null" json:"operator_id"`
	OperationType    string    `gorm:"type:varchar(64);not null" json:"operation_type"`
	OperationContent string    `gorm:"type:text" json:"operation_content"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Issue    *Issue `gorm:"foreignKey:IssueID" json:"issue,omitempty"`
	Operator *User  `gorm:"foreignKey:OperatorID" json:"operator,omitempty"`
}

// TableName specifies the table name for IssueOperationLog model
func (IssueOperationLog) TableName() string {
	return "issue_operation_logs"
}
