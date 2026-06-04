package dto

import (
	"time"

	"ai_system_oncall/internal/model"
)

// CreateProjectRequest 创建项目请求
type CreateProjectRequest struct {
	Name        string               `json:"name" binding:"required,min=2,max=128"`
	Code        string               `json:"code" binding:"required,min=2,max=64"`
	Description string               `json:"description"`
	Members     []ProjectMemberInput `json:"members"`
}

// ProjectMemberInput 项目成员输入
type ProjectMemberInput struct {
	UserID      uint64 `json:"user_id" binding:"required"`
	ProjectRole string `json:"project_role" binding:"required,oneof=developer tester"`
}

// UpdateProjectRequest 更新项目请求
type UpdateProjectRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=128"`
	Description string `json:"description"`
	Status      *int8  `json:"status"`
}

// ProjectInfo 项目信息
type ProjectInfo struct {
	ID          uint64     `json:"id"`
	Name        string     `json:"name"`
	Code        string     `json:"code"`
	Description string     `json:"description"`
	OwnerID     uint64     `json:"owner_id"`
	OwnerName   string     `json:"owner_name,omitempty"`
	Status      int8       `json:"status"`
	CreatedBy   uint64     `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ProjectListRequest 项目列表请求
type ProjectListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Keyword  string `form:"keyword"`
	Status   *int8  `form:"status"`
}

// ProjectListResponse 项目列表响应
type ProjectListResponse struct {
	Total int64          `json:"total"`
	List  []*ProjectInfo `json:"list"`
}

// ToProjectInfo converts Project model to ProjectInfo
func ToProjectInfo(project *model.Project) *ProjectInfo {
	if project == nil {
		return nil
	}
	info := &ProjectInfo{
		ID:          project.ID,
		Name:        project.Name,
		Code:        project.Code,
		Description: project.Description,
		OwnerID:     project.OwnerID,
		Status:      project.Status,
		CreatedBy:   project.CreatedBy,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}
	if project.Owner != nil {
		info.OwnerName = project.Owner.Username
	}
	return info
}

// AddMemberRequest 添加成员请求
type AddMemberRequest struct {
	UserID      uint64 `json:"user_id" binding:"required"`
	ProjectRole string `json:"project_role" binding:"required"`
}

// UpdateMemberRoleRequest 更新成员角色请求
type UpdateMemberRoleRequest struct {
	ProjectRole string `json:"project_role" binding:"required"`
}

// MemberInfo 成员信息
type MemberInfo struct {
	ID          uint64    `json:"id"`
	UserID      uint64    `json:"user_id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	GlobalRole  string    `json:"global_role"`
	ProjectRole string    `json:"project_role"`
	JoinedAt    time.Time `json:"joined_at"`
}

// MemberListResponse 成员列表响应
type MemberListResponse struct {
	Total int64         `json:"total"`
	List  []*MemberInfo `json:"list"`
}

// ToMemberInfo converts ProjectMember to MemberInfo
func ToMemberInfo(member *model.ProjectMember) *MemberInfo {
	if member == nil {
		return nil
	}
	info := &MemberInfo{
		ID:          member.ID,
		UserID:      member.UserID,
		ProjectRole: member.ProjectRole,
		JoinedAt:    member.JoinedAt,
	}
	if member.User != nil {
		info.Username = member.User.Username
		info.Email = member.User.Email
		info.GlobalRole = member.User.Role
	}
	return info
}
