package dto

import (
	"time"

	"ai_system_oncall/internal/model"
)

// CreateIssueRequest 创建问题请求
type CreateIssueRequest struct {
	Title        string  `json:"title" binding:"required,min=2,max=255"`
	Description  string  `json:"description"`
	ProjectID    uint64  `json:"project_id" binding:"required"`
	ServiceID    *uint64 `json:"service_id"`
	IssueType    string  `json:"issue_type"`
	Priority     string  `json:"priority"`
	Environment  string  `json:"environment"`
	ImpactScope  string  `json:"impact_scope"`
	ErrorMessage string  `json:"error_message"`
	LogExcerpt   string  `json:"log_excerpt"`
}

// UpdateIssueRequest 更新问题请求
type UpdateIssueRequest struct {
	Title        string  `json:"title" binding:"omitempty,min=2,max=255"`
	Description  string  `json:"description"`
	ServiceID    *uint64 `json:"service_id"`
	IssueType    string  `json:"issue_type"`
	Priority     string  `json:"priority"`
	Environment  string  `json:"environment"`
	ImpactScope  string  `json:"impact_scope"`
	ErrorMessage string  `json:"error_message"`
	LogExcerpt   string  `json:"log_excerpt"`
	RootCause    string  `json:"root_cause"`
	Solution     string  `json:"solution"`
}

// AssignIssueRequest 分配问题请求
type AssignIssueRequest struct {
	AssigneeID uint64 `json:"assignee_id" binding:"required"`
}

// ChangeStatusRequest 修改状态请求
type ChangeStatusRequest struct {
	Status string `json:"status" binding:"required"`
	Reason string `json:"reason"`
}

// ResolveIssueRequest 解决问题请求
type ResolveIssueRequest struct {
	RootCause string `json:"root_cause" binding:"required"`
	Solution  string `json:"solution" binding:"required"`
}

// IssueInfo 问题信息
type IssueInfo struct {
	ID           uint64     `json:"id"`
	IssueNo      string     `json:"issue_no"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	ProjectID    uint64     `json:"project_id"`
	ProjectName  string     `json:"project_name,omitempty"`
	ServiceID    *uint64    `json:"service_id"`
	ServiceName  string     `json:"service_name,omitempty"`
	IssueType    string     `json:"issue_type"`
	Priority     string     `json:"priority"`
	Environment  string     `json:"environment"`
	Status       string     `json:"status"`
	ImpactScope  string     `json:"impact_scope"`
	ErrorMessage string     `json:"error_message"`
	LogExcerpt   string     `json:"log_excerpt"`
	CreatorID    uint64     `json:"creator_id"`
	CreatorName  string     `json:"creator_name,omitempty"`
	AssigneeID   *uint64    `json:"assignee_id"`
	AssigneeName string     `json:"assignee_name,omitempty"`
	AISummary    string     `json:"ai_summary"`
	AIAnalysis   string     `json:"ai_analysis"`
	RootCause    string     `json:"root_cause"`
	Solution     string     `json:"solution"`
	ResolvedAt   *time.Time `json:"resolved_at"`
	ClosedAt     *time.Time `json:"closed_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// IssueListRequest 问题列表请求
type IssueListRequest struct {
	Page        int    `form:"page" binding:"omitempty,min=1"`
	PageSize    int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	ProjectID   uint64 `form:"project_id"`
	ServiceID   uint64 `form:"service_id"`
	Status      string `form:"status"`
	IssueType   string `form:"issue_type"`
	Priority    string `form:"priority"`
	Environment string `form:"environment"`
	CreatorID   uint64 `form:"creator_id"`
	AssigneeID  uint64 `form:"assignee_id"`
	Keyword     string `form:"keyword"`
}

// IssueListResponse 问题列表响应
type IssueListResponse struct {
	Total int64        `json:"total"`
	List  []*IssueInfo `json:"list"`
}

// ToIssueInfo converts Issue model to IssueInfo
func ToIssueInfo(issue *model.Issue) *IssueInfo {
	if issue == nil {
		return nil
	}
	info := &IssueInfo{
		ID:           issue.ID,
		IssueNo:      issue.IssueNo,
		Title:        issue.Title,
		Description:  issue.Description,
		ProjectID:    issue.ProjectID,
		ServiceID:    issue.ServiceID,
		IssueType:    issue.IssueType,
		Priority:     issue.Priority,
		Environment:  issue.Environment,
		Status:       issue.Status,
		ImpactScope:  issue.ImpactScope,
		ErrorMessage: issue.ErrorMessage,
		LogExcerpt:   issue.LogExcerpt,
		CreatorID:    issue.CreatorID,
		AssigneeID:   issue.AssigneeID,
		AISummary:    issue.AISummary,
		AIAnalysis:   issue.AIAnalysis,
		RootCause:    issue.RootCause,
		Solution:     issue.Solution,
		ResolvedAt:   issue.ResolvedAt,
		ClosedAt:     issue.ClosedAt,
		CreatedAt:    issue.CreatedAt,
		UpdatedAt:    issue.UpdatedAt,
	}
	if issue.Project != nil {
		info.ProjectName = issue.Project.Name
	}
	if issue.Service != nil {
		info.ServiceName = issue.Service.Name
	}
	if issue.Creator != nil {
		info.CreatorName = issue.Creator.Username
	}
	if issue.Assignee != nil {
		info.AssigneeName = issue.Assignee.Username
	}
	return info
}

// StatusLogInfo 状态日志信息
type StatusLogInfo struct {
	ID         uint64    `json:"id"`
	IssueID    uint64    `json:"issue_id"`
	FromStatus string    `json:"from_status"`
	ToStatus   string    `json:"to_status"`
	OperatorID uint64    `json:"operator_id"`
	OperatorName string  `json:"operator_name,omitempty"`
	Reason     string    `json:"reason"`
	CreatedAt  time.Time `json:"created_at"`
}

// ToStatusLogInfo converts IssueStatusLog to StatusLogInfo
func ToStatusLogInfo(log *model.IssueStatusLog) *StatusLogInfo {
	if log == nil {
		return nil
	}
	info := &StatusLogInfo{
		ID:         log.ID,
		IssueID:    log.IssueID,
		FromStatus: log.FromStatus,
		ToStatus:   log.ToStatus,
		OperatorID: log.OperatorID,
		Reason:     log.Reason,
		CreatedAt:  log.CreatedAt,
	}
	if log.Operator != nil {
		info.OperatorName = log.Operator.Username
	}
	return info
}

// OperationLogInfo 操作日志信息
type OperationLogInfo struct {
	ID               uint64    `json:"id"`
	IssueID          uint64    `json:"issue_id"`
	OperatorID       uint64    `json:"operator_id"`
	OperatorName     string    `json:"operator_name,omitempty"`
	OperationType    string    `json:"operation_type"`
	OperationContent string    `json:"operation_content"`
	CreatedAt        time.Time `json:"created_at"`
}

// ToOperationLogInfo converts IssueOperationLog to OperationLogInfo
func ToOperationLogInfo(log *model.IssueOperationLog) *OperationLogInfo {
	if log == nil {
		return nil
	}
	info := &OperationLogInfo{
		ID:               log.ID,
		IssueID:          log.IssueID,
		OperatorID:       log.OperatorID,
		OperationType:    log.OperationType,
		OperationContent: log.OperationContent,
		CreatedAt:        log.CreatedAt,
	}
	if log.Operator != nil {
		info.OperatorName = log.Operator.Username
	}
	return info
}
