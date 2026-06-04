package dto

import (
	"time"

	"ai_system_oncall/internal/model"
)

// CreateSimulatedLogRequest 创建模拟日志请求
type CreateSimulatedLogRequest struct {
	ProjectID   uint64    `json:"project_id" binding:"required"`
	ServiceID   uint64    `json:"service_id" binding:"required"`
	ServiceName string    `json:"service_name" binding:"required"`
	Environment string    `json:"environment" binding:"required,oneof=dev test staging prod"`
	LogLevel    string    `json:"log_level" binding:"required,oneof=ERROR WARN INFO DEBUG"`
	TraceID     string    `json:"trace_id"`
	RequestPath string    `json:"request_path"`
	ErrorCode   string    `json:"error_code"`
	LogContent  string    `json:"log_content" binding:"required"`
	StackTrace  string    `json:"stack_trace"`
	OccurredAt  string `json:"occurred_at" binding:"required"`
}

// BatchCreateSimulatedLogRequest 批量创建模拟日志请求
type BatchCreateSimulatedLogRequest struct {
	Logs []CreateSimulatedLogRequest `json:"logs" binding:"required,min=1,max=100"`
}

// UpdateSimulatedLogRequest 更新模拟日志请求
type UpdateSimulatedLogRequest struct {
	LogLevel    string    `json:"log_level" binding:"omitempty,oneof=ERROR WARN INFO DEBUG"`
	RequestPath string    `json:"request_path"`
	ErrorCode   string    `json:"error_code"`
	LogContent  string    `json:"log_content"`
	StackTrace  string    `json:"stack_trace"`
	OccurredAt  *string `json:"occurred_at"`
}

// LinkIssueRequest 关联问题单请求
type LinkIssueRequest struct {
	IssueID *uint64 `json:"issue_id"` // null 表示取消关联
}

// SimulatedLogListRequest 日志列表查询请求
type SimulatedLogListRequest struct {
	Page        int       `form:"page" binding:"omitempty,min=1"`
	PageSize    int       `form:"page_size" binding:"omitempty,min=1,max=100"`
	ProjectID   uint64    `form:"project_id"`
	ServiceID   uint64    `form:"service_id"`
	TraceID     string    `form:"trace_id"`
	LogLevel    string    `form:"log_level" binding:"omitempty,oneof=ERROR WARN INFO DEBUG"`
	Environment string    `form:"environment" binding:"omitempty,oneof=dev test staging prod"`
	Keyword     string    `form:"keyword"`
	StartTime   *time.Time `form:"start_time"`
	EndTime     *time.Time `form:"end_time"`
	IssueID     uint64    `form:"issue_id"`
}

// SimulatedLogInfo 日志信息
type SimulatedLogInfo struct {
	ID          uint64     `json:"id"`
	ProjectID   uint64     `json:"project_id"`
	ProjectName string     `json:"project_name,omitempty"`
	ServiceID   uint64     `json:"service_id"`
	ServiceName string     `json:"service_name"`
	Environment string     `json:"environment"`
	LogLevel    string     `json:"log_level"`
	TraceID     string     `json:"trace_id"`
	RequestPath string     `json:"request_path"`
	ErrorCode   string     `json:"error_code"`
	LogContent  string     `json:"log_content"`
	StackTrace  string     `json:"stack_trace"`
	OccurredAt  time.Time  `json:"occurred_at"`
	IssueID     *uint64    `json:"issue_id"`
	IssueNo     string     `json:"issue_no,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// SimulatedLogListResponse 日志列表响应
type SimulatedLogListResponse struct {
	Total int64               `json:"total"`
	List  []*SimulatedLogInfo `json:"list"`
}

// ToSimulatedLogInfo converts SimulatedLog to SimulatedLogInfo
func ToSimulatedLogInfo(log *model.SimulatedLog) *SimulatedLogInfo {
	if log == nil {
		return nil
	}
	info := &SimulatedLogInfo{
		ID:          log.ID,
		ProjectID:   log.ProjectID,
		ServiceID:   log.ServiceID,
		ServiceName: log.ServiceName,
		Environment: log.Environment,
		LogLevel:    log.LogLevel,
		TraceID:     log.TraceID,
		RequestPath: log.RequestPath,
		ErrorCode:   log.ErrorCode,
		LogContent:  log.LogContent,
		StackTrace:  log.StackTrace,
		OccurredAt:  log.OccurredAt,
		IssueID:     log.IssueID,
		CreatedAt:   log.CreatedAt,
		UpdatedAt:   log.UpdatedAt,
	}
	if log.Project != nil {
		info.ProjectName = log.Project.Name
	}
	if log.Issue != nil {
		info.IssueNo = log.Issue.IssueNo
	}
	return info
}
