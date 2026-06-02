package repository

import (
	"ai_system_oncall/internal/model"

	"gorm.io/gorm"
)

type StatusLogRepository struct {
	db *gorm.DB
}

func NewStatusLogRepository(db *gorm.DB) *StatusLogRepository {
	return &StatusLogRepository{db: db}
}

// Create creates a new status log
func (r *StatusLogRepository) Create(log *model.IssueStatusLog) error {
	return r.db.Create(log).Error
}

// ListByIssueID lists status logs of an issue
func (r *StatusLogRepository) ListByIssueID(issueID uint64) ([]model.IssueStatusLog, error) {
	var logs []model.IssueStatusLog
	err := r.db.Where("issue_id = ?", issueID).Preload("Operator").Order("created_at ASC").Find(&logs).Error
	return logs, err
}

type OperationLogRepository struct {
	db *gorm.DB
}

func NewOperationLogRepository(db *gorm.DB) *OperationLogRepository {
	return &OperationLogRepository{db: db}
}

// Create creates a new operation log
func (r *OperationLogRepository) Create(log *model.IssueOperationLog) error {
	return r.db.Create(log).Error
}

// ListByIssueID lists operation logs of an issue
func (r *OperationLogRepository) ListByIssueID(issueID uint64) ([]model.IssueOperationLog, error) {
	var logs []model.IssueOperationLog
	err := r.db.Where("issue_id = ?", issueID).Preload("Operator").Order("created_at ASC").Find(&logs).Error
	return logs, err
}

// ListByIssueIDAndType lists operation logs by issue ID and type
func (r *OperationLogRepository) ListByIssueIDAndType(issueID uint64, operationType string) ([]model.IssueOperationLog, error) {
	var logs []model.IssueOperationLog
	err := r.db.Where("issue_id = ? AND operation_type = ?", issueID, operationType).Order("created_at ASC").Find(&logs).Error
	return logs, err
}
