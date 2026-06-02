package service

import (
	"errors"

	"ai_system_oncall/internal/constant"
	"ai_system_oncall/internal/dto"
	"ai_system_oncall/internal/model"
	"ai_system_oncall/internal/repository"

	"gorm.io/gorm"
)

type CommentService struct {
	commentRepo      *repository.CommentRepository
	issueRepo        *repository.IssueRepository
	operationLogRepo *repository.OperationLogRepository
}

func NewCommentService(
	commentRepo *repository.CommentRepository,
	issueRepo *repository.IssueRepository,
	operationLogRepo *repository.OperationLogRepository,
) *CommentService {
	return &CommentService{
		commentRepo:      commentRepo,
		issueRepo:        issueRepo,
		operationLogRepo: operationLogRepo,
	}
}

// CreateComment creates a new comment
func (s *CommentService) CreateComment(issueID, userID uint64, req *dto.CreateCommentRequest) (*dto.CommentInfo, error) {
	// Check if issue exists
	_, err := s.issueRepo.FindByID(issueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("问题不存在")
		}
		return nil, err
	}

	// Set defaults
	commentType := req.CommentType
	if commentType == "" {
		commentType = constant.CommentTypeComment
	}
	visibility := req.Visibility
	if visibility == "" {
		visibility = constant.VisibilityPublic
	}

	// Create comment
	comment := &model.IssueComment{
		IssueID:     issueID,
		UserID:      &userID,
		CommentType: commentType,
		Content:     req.Content,
		Visibility:  visibility,
	}

	if err := s.commentRepo.Create(comment); err != nil {
		return nil, err
	}

	// Create operation log
	opLog := &model.IssueOperationLog{
		IssueID:          issueID,
		OperatorID:       userID,
		OperationType:    constant.OperationAddComment,
		OperationContent: req.Content,
	}
	_ = s.operationLogRepo.Create(opLog)

	// Reload with relations
	comment, _ = s.commentRepo.FindByID(comment.ID)
	return dto.ToCommentInfo(comment), nil
}

// GetComment gets a comment by ID
func (s *CommentService) GetComment(id uint64) (*dto.CommentInfo, error) {
	comment, err := s.commentRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("评论不存在")
		}
		return nil, err
	}
	return dto.ToCommentInfo(comment), nil
}

// ListComments lists comments of an issue
func (s *CommentService) ListComments(issueID uint64, commentType string) (*dto.CommentListResponse, error) {
	// Check if issue exists
	_, err := s.issueRepo.FindByID(issueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("问题不存在")
		}
		return nil, err
	}

	comments, err := s.commentRepo.ListByIssueID(issueID, commentType)
	if err != nil {
		return nil, err
	}

	list := make([]*dto.CommentInfo, 0, len(comments))
	for _, comment := range comments {
		list = append(list, dto.ToCommentInfo(&comment))
	}

	return &dto.CommentListResponse{
		Total: int64(len(list)),
		List:  list,
	}, nil
}

// DeleteComment deletes a comment
func (s *CommentService) DeleteComment(id, userID uint64, globalRole string) error {
	comment, err := s.commentRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("评论不存在")
		}
		return err
	}

	// Check permission: only comment author or system admin can delete
	if globalRole != constant.RoleSystemAdmin && (comment.UserID == nil || *comment.UserID != userID) {
		return errors.New("无权限删除评论")
	}

	return s.commentRepo.Delete(id)
}
