package repository

import (
	"ai_system_oncall/internal/model"
	"time"

	"gorm.io/gorm"
)

type IssueRepository struct {
	db *gorm.DB
}

func NewIssueRepository(db *gorm.DB) *IssueRepository {
	return &IssueRepository{db: db}
}

// Create creates a new issue
func (r *IssueRepository) Create(issue *model.Issue) error {
	return r.db.Create(issue).Error
}

// Update updates an issue
func (r *IssueRepository) Update(issue *model.Issue) error {
	return r.db.Save(issue).Error
}

// Delete soft deletes an issue
func (r *IssueRepository) Delete(id uint64) error {
	return r.db.Delete(&model.Issue{}, id).Error
}

// FindByID finds an issue by ID
func (r *IssueRepository) FindByID(id uint64) (*model.Issue, error) {
	var issue model.Issue
	err := r.db.Preload("Project").Preload("Service").Preload("Creator").Preload("Assignee").First(&issue, id).Error
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

// FindByIssueNo finds an issue by issue number
func (r *IssueRepository) FindByIssueNo(issueNo string) (*model.Issue, error) {
	var issue model.Issue
	err := r.db.Where("issue_no = ?", issueNo).First(&issue).Error
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

// List lists issues with pagination and filters
func (r *IssueRepository) List(page, pageSize int, filters map[string]interface{}) ([]model.Issue, int64, error) {
	var issues []model.Issue
	var total int64

	query := r.db.Model(&model.Issue{})

	if projectID, ok := filters["project_id"]; ok && projectID != nil {
		query = query.Where("project_id = ?", projectID)
	}
	if serviceID, ok := filters["service_id"]; ok && serviceID != nil {
		query = query.Where("service_id = ?", serviceID)
	}
	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if issueType, ok := filters["issue_type"]; ok && issueType != "" {
		query = query.Where("issue_type = ?", issueType)
	}
	if priority, ok := filters["priority"]; ok && priority != "" {
		query = query.Where("priority = ?", priority)
	}
	if environment, ok := filters["environment"]; ok && environment != "" {
		query = query.Where("environment = ?", environment)
	}
	if creatorID, ok := filters["creator_id"]; ok && creatorID != nil {
		query = query.Where("creator_id = ?", creatorID)
	}
	if assigneeID, ok := filters["assignee_id"]; ok && assigneeID != nil {
		query = query.Where("assignee_id = ?", assigneeID)
	}
	if keyword, ok := filters["keyword"]; ok && keyword != "" {
		query = query.Where("title LIKE ? OR issue_no LIKE ?", "%"+keyword.(string)+"%", "%"+keyword.(string)+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Project").Preload("Service").Preload("Creator").Preload("Assignee").
		Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&issues).Error; err != nil {
		return nil, 0, err
	}

	return issues, total, nil
}

// UpdateStatus updates issue status
func (r *IssueRepository) UpdateStatus(id uint64, status string) error {
	return r.db.Model(&model.Issue{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateAssignee updates issue assignee
func (r *IssueRepository) UpdateAssignee(id uint64, assigneeID uint64) error {
	return r.db.Model(&model.Issue{}).Where("id = ?", id).Update("assignee_id", assigneeID).Error
}

// UpdateResolution updates issue resolution info
func (r *IssueRepository) UpdateResolution(id uint64, rootCause, solution string, resolvedAt time.Time) error {
	return r.db.Model(&model.Issue{}).Where("id = ?", id).Updates(map[string]interface{}{
		"root_cause":  rootCause,
		"solution":    solution,
		"resolved_at": resolvedAt,
	}).Error
}

// UpdateClosedAt updates issue closed time
func (r *IssueRepository) UpdateClosedAt(id uint64, closedAt time.Time) error {
	return r.db.Model(&model.Issue{}).Where("id = ?", id).Update("closed_at", closedAt).Error
}

// GetTodayIssueCount gets the count of issues created today for generating issue number
func (r *IssueRepository) GetTodayIssueCount() (int64, error) {
	var count int64
	today := time.Now().Format("2006-01-02")
	err := r.db.Model(&model.Issue{}).Where("DATE(created_at) = ?", today).Count(&count).Error
	return count, err
}

// ListByProjectID lists issues of a project
func (r *IssueRepository) ListByProjectID(projectID uint64) ([]model.Issue, error) {
	var issues []model.Issue
	err := r.db.Where("project_id = ?", projectID).Order("created_at DESC").Find(&issues).Error
	return issues, err
}
