package repository

import (
	"ai_system_oncall/internal/model"

	"gorm.io/gorm"
)

type AIAnalysisTaskRepository struct {
	db *gorm.DB
}

func NewAIAnalysisTaskRepository(db *gorm.DB) *AIAnalysisTaskRepository {
	return &AIAnalysisTaskRepository{db: db}
}

// Create creates a new task
func (r *AIAnalysisTaskRepository) Create(task *model.AIAnalysisTask) error {
	return r.db.Create(task).Error
}

// Update updates a task
func (r *AIAnalysisTaskRepository) Update(task *model.AIAnalysisTask) error {
	return r.db.Save(task).Error
}

// FindByID finds a task by ID
func (r *AIAnalysisTaskRepository) FindByID(id uint64) (*model.AIAnalysisTask, error) {
	var task model.AIAnalysisTask
	err := r.db.Preload("Issue").First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// FindByIssueID finds tasks by issue ID
func (r *AIAnalysisTaskRepository) FindByIssueID(issueID uint64, limit int) ([]model.AIAnalysisTask, error) {
	var tasks []model.AIAnalysisTask
	query := r.db.Where("issue_id = ?", issueID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&tasks).Error
	return tasks, err
}

// FindLatestByIssueID finds the latest task by issue ID
func (r *AIAnalysisTaskRepository) FindLatestByIssueID(issueID uint64) (*model.AIAnalysisTask, error) {
	var task model.AIAnalysisTask
	err := r.db.Where("issue_id = ?", issueID).Order("created_at DESC").First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// UpdateStatus updates task status
func (r *AIAnalysisTaskRepository) UpdateStatus(id uint64, status string) error {
	return r.db.Model(&model.AIAnalysisTask{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateProgress updates task progress
func (r *AIAnalysisTaskRepository) UpdateProgress(id uint64, progress, currentStep string) error {
	return r.db.Model(&model.AIAnalysisTask{}).Where("id = ?", id).Updates(map[string]interface{}{
		"progress":     progress,
		"current_step": currentStep,
	}).Error
}

// UpdateResult updates task result
func (r *AIAnalysisTaskRepository) UpdateResult(id uint64, result string) error {
	return r.db.Model(&model.AIAnalysisTask{}).Where("id = ?", id).Update("result", result).Error
}

// CancelTask cancels a task
func (r *AIAnalysisTaskRepository) CancelTask(id uint64) error {
	return r.db.Model(&model.AIAnalysisTask{}).Where("id = ? AND status IN ?", id, []string{model.TaskStatusPending, model.TaskStatusRunning}).
		Update("status", model.TaskStatusCancelled).Error
}

// GetRunningTasks gets all running tasks
func (r *AIAnalysisTaskRepository) GetRunningTasks() ([]model.AIAnalysisTask, error) {
	var tasks []model.AIAnalysisTask
	err := r.db.Where("status = ?", model.TaskStatusRunning).Find(&tasks).Error
	return tasks, err
}
