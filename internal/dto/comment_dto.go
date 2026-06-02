package dto

import (
	"time"

	"ai_system_oncall/internal/model"
)

// CreateCommentRequest 创建评论请求
type CreateCommentRequest struct {
	CommentType string `json:"comment_type"`
	Content     string `json:"content" binding:"required"`
	Visibility  string `json:"visibility"`
}

// UpdateCommentRequest 更新评论请求
type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

// CommentInfo 评论信息
type CommentInfo struct {
	ID          uint64    `json:"id"`
	IssueID     uint64    `json:"issue_id"`
	UserID      *uint64   `json:"user_id"`
	Username    string    `json:"username,omitempty"`
	CommentType string    `json:"comment_type"`
	Content     string    `json:"content"`
	Visibility  string    `json:"visibility"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToCommentInfo converts IssueComment to CommentInfo
func ToCommentInfo(comment *model.IssueComment) *CommentInfo {
	if comment == nil {
		return nil
	}
	info := &CommentInfo{
		ID:          comment.ID,
		IssueID:     comment.IssueID,
		UserID:      comment.UserID,
		CommentType: comment.CommentType,
		Content:     comment.Content,
		Visibility:  comment.Visibility,
		CreatedAt:   comment.CreatedAt,
		UpdatedAt:   comment.UpdatedAt,
	}
	if comment.User != nil {
		info.Username = comment.User.Username
	}
	return info
}

// CommentListResponse 评论列表响应
type CommentListResponse struct {
	Total int64          `json:"total"`
	List  []*CommentInfo `json:"list"`
}
