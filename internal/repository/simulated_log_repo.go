package repository

import (
	"time"

	"ai_system_oncall/internal/model"

	"gorm.io/gorm"
)

type SimulatedLogRepository struct {
	db *gorm.DB
}

func NewSimulatedLogRepository(db *gorm.DB) *SimulatedLogRepository {
	return &SimulatedLogRepository{db: db}
}

// Create creates a new simulated log
func (r *SimulatedLogRepository) Create(log *model.SimulatedLog) error {
	return r.db.Create(log).Error
}

// CreateBatch creates multiple simulated logs
func (r *SimulatedLogRepository) CreateBatch(logs []*model.SimulatedLog) error {
	return r.db.Create(&logs).Error
}

// Update updates a simulated log
func (r *SimulatedLogRepository) Update(log *model.SimulatedLog) error {
	return r.db.Save(log).Error
}

// Delete soft deletes a simulated log
func (r *SimulatedLogRepository) Delete(id uint64) error {
	return r.db.Delete(&model.SimulatedLog{}, id).Error
}

// FindByID finds a simulated log by ID
func (r *SimulatedLogRepository) FindByID(id uint64) (*model.SimulatedLog, error) {
	var log model.SimulatedLog
	err := r.db.Preload("Project").Preload("Service").Preload("Issue").
		First(&log, id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// FindByTraceID finds logs by trace ID
func (r *SimulatedLogRepository) FindByTraceID(traceID string) ([]model.SimulatedLog, error) {
	var logs []model.SimulatedLog
	err := r.db.Where("trace_id = ?", traceID).
		Preload("Project").Preload("Service").Preload("Issue").
		Order("occurred_at DESC").
		Find(&logs).Error
	return logs, err
}

// List lists simulated logs with filters
func (r *SimulatedLogRepository) List(page, pageSize int, projectID, serviceID, issueID uint64, traceID, logLevel, environment, keyword string, startTime, endTime *time.Time) ([]model.SimulatedLog, int64, error) {
	var logs []model.SimulatedLog
	var total int64

	query := r.db.Model(&model.SimulatedLog{})

	if projectID > 0 {
		query = query.Where("project_id = ?", projectID)
	}
	if serviceID > 0 {
		query = query.Where("service_id = ?", serviceID)
	}
	if issueID > 0 {
		query = query.Where("issue_id = ?", issueID)
	}
	if traceID != "" {
		query = query.Where("trace_id = ?", traceID)
	}
	if logLevel != "" {
		query = query.Where("log_level = ?", logLevel)
	}
	if environment != "" {
		query = query.Where("environment = ?", environment)
	}
	if keyword != "" {
		query = query.Where("log_content LIKE ? OR stack_trace LIKE ? OR request_path LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if startTime != nil {
		query = query.Where("occurred_at >= ?", startTime)
	}
	if endTime != nil {
		query = query.Where("occurred_at <= ?", endTime)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Project").Preload("Service").Preload("Issue").
		Offset(offset).Limit(pageSize).Order("occurred_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// ListByServiceID lists logs by service ID
func (r *SimulatedLogRepository) ListByServiceID(serviceID uint64, page, pageSize int) ([]model.SimulatedLog, int64, error) {
	var logs []model.SimulatedLog
	var total int64

	query := r.db.Model(&model.SimulatedLog{}).Where("service_id = ?", serviceID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Project").Preload("Service").Preload("Issue").
		Offset(offset).Limit(pageSize).Order("occurred_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// ListByIssueID lists logs by issue ID
func (r *SimulatedLogRepository) ListByIssueID(issueID uint64) ([]model.SimulatedLog, error) {
	var logs []model.SimulatedLog
	err := r.db.Where("issue_id = ?", issueID).
		Preload("Project").Preload("Service").
		Order("occurred_at DESC").
		Find(&logs).Error
	return logs, err
}

// UpdateIssueID updates the issue ID of a log
func (r *SimulatedLogRepository) UpdateIssueID(id uint64, issueID *uint64) error {
	return r.db.Model(&model.SimulatedLog{}).Where("id = ?", id).Update("issue_id", issueID).Error
}

// CountByProject counts logs by project ID
func (r *SimulatedLogRepository) CountByProject(projectID uint64) (int64, error) {
	var count int64
	err := r.db.Model(&model.SimulatedLog{}).Where("project_id = ?", projectID).Count(&count).Error
	return count, err
}
