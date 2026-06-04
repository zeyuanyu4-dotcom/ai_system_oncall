package service

import (
	"encoding/json"
	"errors"
	"time"

	"ai_system_oncall/internal/client"
	"ai_system_oncall/internal/model"
	"ai_system_oncall/internal/repository"

	"gorm.io/gorm"
)

type AIAnalysisTaskService struct {
	taskRepo   *repository.AIAnalysisTaskRepository
	issueRepo  *repository.IssueRepository
	aiClient   *client.AIClient
}

func NewAIAnalysisTaskService(
	taskRepo *repository.AIAnalysisTaskRepository,
	issueRepo *repository.IssueRepository,
	aiClient *client.AIClient,
) *AIAnalysisTaskService {
	return &AIAnalysisTaskService{
		taskRepo:  taskRepo,
		issueRepo: issueRepo,
		aiClient:  aiClient,
	}
}

// CreateTask creates a new analysis task
func (s *AIAnalysisTaskService) CreateTask(issueID uint64) (*model.AIAnalysisTask, error) {
	// Check if issue exists
	_, err := s.issueRepo.FindByID(issueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("问题不存在")
		}
		return nil, err
	}

	// Check if there's already a running task
	tasks, err := s.taskRepo.FindByIssueID(issueID, 1)
	if err == nil && len(tasks) > 0 {
		latestTask := tasks[0]
		if latestTask.Status == model.TaskStatusPending || latestTask.Status == model.TaskStatusRunning {
			return nil, errors.New("该问题已有分析任务在进行中")
		}
	}

	task := &model.AIAnalysisTask{
		IssueID: issueID,
		Status:  model.TaskStatusPending,
	}

	if err := s.taskRepo.Create(task); err != nil {
		return nil, err
	}

	return task, nil
}

// GetTask gets a task by ID
func (s *AIAnalysisTaskService) GetTask(id uint64) (*model.AIAnalysisTask, error) {
	return s.taskRepo.FindByID(id)
}

// GetTasksByIssueID gets tasks by issue ID
func (s *AIAnalysisTaskService) GetTasksByIssueID(issueID uint64) ([]model.AIAnalysisTask, error) {
	return s.taskRepo.FindByIssueID(issueID, 10)
}

// CancelTask cancels a task
func (s *AIAnalysisTaskService) CancelTask(id uint64) error {
	return s.taskRepo.CancelTask(id)
}

// ExecuteTask executes the analysis task (called by async worker)
func (s *AIAnalysisTaskService) ExecuteTask(taskID uint64) error {
	task, err := s.taskRepo.FindByID(taskID)
	if err != nil {
		return err
	}

	if task.Status != model.TaskStatusPending {
		return errors.New("任务状态不正确")
	}

	// Update status to running
	now := time.Now()
	task.Status = model.TaskStatusRunning
	task.StartedAt = &now
	task.Progress = "0/8"
	task.CurrentStep = "正在启动分析..."
	if err := s.taskRepo.Update(task); err != nil {
		return err
	}

	// Get issue info
	issue, err := s.issueRepo.FindByID(task.IssueID)
	if err != nil {
		s.markTaskFailed(task, err.Error())
		return err
	}

	// Call Python Agent service
	req := &client.AgentAnalysisRequest{
		TaskID:       int(taskID),
		IssueID:      int(issue.ID),
		IssueNo:      issue.IssueNo,
		Title:        issue.Title,
		Description:  issue.Description,
		ErrorMessage: issue.ErrorMessage,
		LogExcerpt:   issue.LogExcerpt,
		Environment:  issue.Environment,
		ProjectID:    issue.ProjectID,
		ProjectName:  "",
		ServiceName:  "",
		ImpactScope:  issue.ImpactScope,
	}

	if issue.Project != nil {
		req.ProjectName = issue.Project.Name
	}
	if issue.Service != nil {
		req.ServiceName = issue.Service.Name
	}

	result, err := s.aiClient.AgentAnalyze(req)
	if err != nil {
		s.markTaskFailed(task, err.Error())
		return err
	}

	// Update task with result
	task.Status = model.TaskStatusCompleted
	completedAt := time.Now()
	task.CompletedAt = &completedAt
	task.Progress = "8/8"
	task.CurrentStep = "分析完成"

	resultJSON, _ := json.Marshal(result)
	task.Result = string(resultJSON)

	toolCallsJSON, _ := json.Marshal(result.ToolCalls)
	task.ToolCalls = string(toolCallsJSON)

	if err := s.taskRepo.Update(task); err != nil {
		return err
	}

	// Update issue with latest AI analysis
	s.issueRepo.UpdateAIAnalysis(issue.ID, result.Summary, string(resultJSON))

	return nil
}

func (s *AIAnalysisTaskService) markTaskFailed(task *model.AIAnalysisTask, errMsg string) {
	task.Status = model.TaskStatusFailed
	task.ErrorMessage = errMsg
	now := time.Now()
	task.CompletedAt = &now
	s.taskRepo.Update(task)
}

// UpdateTaskProgress updates task progress (called by Python agent)
func (s *AIAnalysisTaskService) UpdateTaskProgress(id uint64, progress, currentStep string) error {
	return s.taskRepo.UpdateProgress(id, progress, currentStep)
}
