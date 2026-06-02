package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ai_system_oncall/internal/constant"
	"ai_system_oncall/internal/dto"
	"ai_system_oncall/internal/model"
	"ai_system_oncall/internal/repository"

	"gorm.io/gorm"
)

type IssueService struct {
	issueRepo       *repository.IssueRepository
	projectRepo     *repository.ProjectRepository
	projectMemberRepo *repository.ProjectMemberRepository
	serviceRepo     *repository.ServiceRepository
	statusLogRepo   *repository.StatusLogRepository
	operationLogRepo *repository.OperationLogRepository
}

func NewIssueService(
	issueRepo *repository.IssueRepository,
	projectRepo *repository.ProjectRepository,
	projectMemberRepo *repository.ProjectMemberRepository,
	serviceRepo *repository.ServiceRepository,
	statusLogRepo *repository.StatusLogRepository,
	operationLogRepo *repository.OperationLogRepository,
) *IssueService {
	return &IssueService{
		issueRepo:        issueRepo,
		projectRepo:      projectRepo,
		projectMemberRepo: projectMemberRepo,
		serviceRepo:      serviceRepo,
		statusLogRepo:    statusLogRepo,
		operationLogRepo: operationLogRepo,
	}
}

// generateIssueNo generates a unique issue number
func (s *IssueService) generateIssueNo() (string, error) {
	today := time.Now().Format("20060102")
	count, err := s.issueRepo.GetTodayIssueCount()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("ISSUE-%s-%04d", today, count+1), nil
}

// CreateIssue creates a new issue
func (s *IssueService) CreateIssue(creatorID uint64, req *dto.CreateIssueRequest) (*dto.IssueInfo, error) {
	// Check if project exists
	project, err := s.projectRepo.FindByID(req.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("项目不存在")
		}
		return nil, err
	}
	if !project.IsEnabled() {
		return nil, errors.New("项目已停用")
	}

	// Check if user is project member
	isMember, err := s.projectMemberRepo.ExistsByProjectAndUser(req.ProjectID, creatorID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("用户不属于该项目")
	}

	// Check if service belongs to project (if specified)
	if req.ServiceID != nil && *req.ServiceID > 0 {
		service, err := s.serviceRepo.FindByID(*req.ServiceID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("服务不存在")
			}
			return nil, err
		}
		if service.ProjectID != req.ProjectID {
			return nil, errors.New("服务不属于该项目")
		}
	}

	// Generate issue number
	issueNo, err := s.generateIssueNo()
	if err != nil {
		return nil, err
	}

	// Set defaults
	issueType := req.IssueType
	if issueType == "" {
		issueType = constant.IssueTypeOther
	}
	priority := req.Priority
	if priority == "" {
		priority = constant.PriorityP2
	}

	// Create issue
	issue := &model.Issue{
		IssueNo:      issueNo,
		Title:        req.Title,
		Description:  req.Description,
		ProjectID:    req.ProjectID,
		ServiceID:    req.ServiceID,
		IssueType:    issueType,
		Priority:     priority,
		Environment:  req.Environment,
		Status:       constant.StatusPendingAnalysis,
		ImpactScope:  req.ImpactScope,
		ErrorMessage: req.ErrorMessage,
		LogExcerpt:   req.LogExcerpt,
		CreatorID:    creatorID,
	}

	if err := s.issueRepo.Create(issue); err != nil {
		return nil, err
	}

	// Create status log
	statusLog := &model.IssueStatusLog{
		IssueID:    issue.ID,
		FromStatus: "",
		ToStatus:   constant.StatusPendingAnalysis,
		OperatorID: creatorID,
		Reason:     "创建问题",
	}
	_ = s.statusLogRepo.Create(statusLog)

	// Create operation log
	opLog := &model.IssueOperationLog{
		IssueID:          issue.ID,
		OperatorID:       creatorID,
		OperationType:    constant.OperationCreateIssue,
		OperationContent: fmt.Sprintf("创建问题: %s", issue.Title),
	}
	_ = s.operationLogRepo.Create(opLog)

	// Reload with relations
	issue, _ = s.issueRepo.FindByID(issue.ID)
	return dto.ToIssueInfo(issue), nil
}

// GetIssue gets an issue by ID
func (s *IssueService) GetIssue(id uint64) (*dto.IssueInfo, error) {
	issue, err := s.issueRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("问题不存在")
		}
		return nil, err
	}
	return dto.ToIssueInfo(issue), nil
}

// ListIssues lists issues
func (s *IssueService) ListIssues(req *dto.IssueListRequest) (*dto.IssueListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	filters := make(map[string]interface{})
	if req.ProjectID > 0 {
		filters["project_id"] = req.ProjectID
	}
	if req.ServiceID > 0 {
		filters["service_id"] = req.ServiceID
	}
	if req.Status != "" {
		filters["status"] = req.Status
	}
	if req.IssueType != "" {
		filters["issue_type"] = req.IssueType
	}
	if req.Priority != "" {
		filters["priority"] = req.Priority
	}
	if req.Environment != "" {
		filters["environment"] = req.Environment
	}
	if req.CreatorID > 0 {
		filters["creator_id"] = req.CreatorID
	}
	if req.AssigneeID > 0 {
		filters["assignee_id"] = req.AssigneeID
	}
	if req.Keyword != "" {
		filters["keyword"] = req.Keyword
	}

	issues, total, err := s.issueRepo.List(req.Page, req.PageSize, filters)
	if err != nil {
		return nil, err
	}

	list := make([]*dto.IssueInfo, 0, len(issues))
	for _, issue := range issues {
		list = append(list, dto.ToIssueInfo(&issue))
	}

	return &dto.IssueListResponse{
		Total: total,
		List:  list,
	}, nil
}

// UpdateIssue updates an issue
func (s *IssueService) UpdateIssue(id, operatorID uint64, req *dto.UpdateIssueRequest) (*dto.IssueInfo, error) {
	issue, err := s.issueRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("问题不存在")
		}
		return nil, err
	}

	changes := make(map[string]interface{})

	if req.Title != "" && req.Title != issue.Title {
		issue.Title = req.Title
		changes["title"] = req.Title
	}
	if req.Description != "" {
		issue.Description = req.Description
	}
	if req.ServiceID != nil {
		issue.ServiceID = req.ServiceID
	}
	if req.IssueType != "" {
		issue.IssueType = req.IssueType
	}
	if req.Priority != "" {
		issue.Priority = req.Priority
	}
	if req.Environment != "" {
		issue.Environment = req.Environment
	}
	if req.ImpactScope != "" {
		issue.ImpactScope = req.ImpactScope
	}
	if req.ErrorMessage != "" {
		issue.ErrorMessage = req.ErrorMessage
	}
	if req.LogExcerpt != "" {
		issue.LogExcerpt = req.LogExcerpt
	}
	if req.RootCause != "" {
		issue.RootCause = req.RootCause
	}
	if req.Solution != "" {
		issue.Solution = req.Solution
	}

	if err := s.issueRepo.Update(issue); err != nil {
		return nil, err
	}

	// Create operation log if there are changes
	if len(changes) > 0 {
		content, _ := json.Marshal(changes)
		opLog := &model.IssueOperationLog{
			IssueID:          issue.ID,
			OperatorID:       operatorID,
			OperationType:    constant.OperationUpdateIssue,
			OperationContent: string(content),
		}
		_ = s.operationLogRepo.Create(opLog)
	}

	return dto.ToIssueInfo(issue), nil
}

// DeleteIssue deletes an issue
func (s *IssueService) DeleteIssue(id uint64) error {
	_, err := s.issueRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("问题不存在")
		}
		return err
	}
	return s.issueRepo.Delete(id)
}

// GetOperationLogs gets operation logs of an issue
func (s *IssueService) GetOperationLogs(issueID uint64) ([]*dto.OperationLogInfo, error) {
	logs, err := s.operationLogRepo.ListByIssueID(issueID)
	if err != nil {
		return nil, err
	}

	list := make([]*dto.OperationLogInfo, 0, len(logs))
	for _, log := range logs {
		list = append(list, dto.ToOperationLogInfo(&log))
	}

	return list, nil
}
