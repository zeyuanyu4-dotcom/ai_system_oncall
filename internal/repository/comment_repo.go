package repository

import (
	"ai_system_oncall/internal/model"

	"gorm.io/gorm"
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// Create creates a new comment
func (r *CommentRepository) Create(comment *model.IssueComment) error {
	return r.db.Create(comment).Error
}

// Update updates a comment
func (r *CommentRepository) Update(comment *model.IssueComment) error {
	return r.db.Save(comment).Error
}

// Delete soft deletes a comment
func (r *CommentRepository) Delete(id uint64) error {
	return r.db.Delete(&model.IssueComment{}, id).Error
}

// FindByID finds a comment by ID
func (r *CommentRepository) FindByID(id uint64) (*model.IssueComment, error) {
	var comment model.IssueComment
	err := r.db.First(&comment, id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// ListByIssueID lists comments of an issue
func (r *CommentRepository) ListByIssueID(issueID uint64, commentType string) ([]model.IssueComment, error) {
	var comments []model.IssueComment
	query := r.db.Where("issue_id = ?", issueID)
	if commentType != "" {
		query = query.Where("comment_type = ?", commentType)
	}
	err := query.Preload("User").Order("created_at ASC").Find(&comments).Error
	return comments, err
}

// CountByIssueID counts comments of an issue
func (r *CommentRepository) CountByIssueID(issueID uint64) (int64, error) {
	var count int64
	err := r.db.Model(&model.IssueComment{}).Where("issue_id = ?", issueID).Count(&count).Error
	return count, err
}
