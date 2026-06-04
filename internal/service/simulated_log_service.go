package service

import (
	"errors"
	"time"

	"ai_system_oncall/internal/dto"
	"ai_system_oncall/internal/model"
	"ai_system_oncall/internal/repository"

	"gorm.io/gorm"
)

type SimulatedLogService struct {
	logRepo          *repository.SimulatedLogRepository
	projectRepo      *repository.ProjectRepository
	serviceRepo      *repository.ServiceRepository
	issueRepo        *repository.IssueRepository
	projectMemberRepo *repository.ProjectMemberRepository
}

func NewSimulatedLogService(
	logRepo *repository.SimulatedLogRepository,
	projectRepo *repository.ProjectRepository,
	serviceRepo *repository.ServiceRepository,
	issueRepo *repository.IssueRepository,
	projectMemberRepo *repository.ProjectMemberRepository,
) *SimulatedLogService {
	return &SimulatedLogService{
		logRepo:          logRepo,
		projectRepo:      projectRepo,
		serviceRepo:      serviceRepo,
		issueRepo:        issueRepo,
		projectMemberRepo: projectMemberRepo,
	}
}

// CreateLog creates a new simulated log
func (s *SimulatedLogService) CreateLog(req *dto.CreateSimulatedLogRequest) (*dto.SimulatedLogInfo, error) {
	// Validate project exists
	_, err := s.projectRepo.FindByID(req.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("项目不存在")
		}
		return nil, err
	}

	// Validate service exists
	_, err = s.serviceRepo.FindByID(req.ServiceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("服务不存在")
		}
		return nil, err
	}

	occurredAt, err := time.Parse("2006-01-02T15:04", req.OccurredAt)
	if err != nil {
		return nil, errors.New("发生时间格式错误，请使用 YYYY-MM-DDTHH:MM 格式")
	}

	log := &model.SimulatedLog{
		ProjectID:   req.ProjectID,
		ServiceID:   req.ServiceID,
		ServiceName: req.ServiceName,
		Environment: req.Environment,
		LogLevel:    req.LogLevel,
		TraceID:     req.TraceID,
		RequestPath: req.RequestPath,
		ErrorCode:   req.ErrorCode,
		LogContent:  req.LogContent,
		StackTrace:  req.StackTrace,
		OccurredAt:  occurredAt,
	}

	if err := s.logRepo.Create(log); err != nil {
		return nil, err
	}

	return s.getLogInfo(log.ID)
}

// BatchCreateLogs creates multiple simulated logs
func (s *SimulatedLogService) BatchCreateLogs(req *dto.BatchCreateSimulatedLogRequest) (int, error) {
	logs := make([]*model.SimulatedLog, 0, len(req.Logs))
	for _, logReq := range req.Logs {
		occurredAt, err := time.Parse("2006-01-02T15:04", logReq.OccurredAt)
		if err != nil {
			return 0, errors.New("发生时间格式错误，请使用 YYYY-MM-DDTHH:MM 格式")
		}
		logs = append(logs, &model.SimulatedLog{
			ProjectID:   logReq.ProjectID,
			ServiceID:   logReq.ServiceID,
			ServiceName: logReq.ServiceName,
			Environment: logReq.Environment,
			LogLevel:    logReq.LogLevel,
			TraceID:     logReq.TraceID,
			RequestPath: logReq.RequestPath,
			ErrorCode:   logReq.ErrorCode,
			LogContent:  logReq.LogContent,
			StackTrace:  logReq.StackTrace,
			OccurredAt:  occurredAt,
		})
	}

	if err := s.logRepo.CreateBatch(logs); err != nil {
		return 0, err
	}

	return len(logs), nil
}

// GetLog gets a log by ID
func (s *SimulatedLogService) GetLog(id uint64) (*dto.SimulatedLogInfo, error) {
	return s.getLogInfo(id)
}

// ListLogs lists logs with filters
func (s *SimulatedLogService) ListLogs(req *dto.SimulatedLogListRequest) (*dto.SimulatedLogListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	logs, total, err := s.logRepo.List(
		req.Page, req.PageSize,
		req.ProjectID, req.ServiceID, req.IssueID,
		req.TraceID, req.LogLevel, req.Environment, req.Keyword,
		req.StartTime, req.EndTime,
	)
	if err != nil {
		return nil, err
	}

	list := make([]*dto.SimulatedLogInfo, 0, len(logs))
	for _, log := range logs {
		list = append(list, dto.ToSimulatedLogInfo(&log))
	}

	return &dto.SimulatedLogListResponse{
		Total: total,
		List:  list,
	}, nil
}

// GetLogsByTraceID gets logs by trace ID
func (s *SimulatedLogService) GetLogsByTraceID(traceID string) ([]*dto.SimulatedLogInfo, error) {
	logs, err := s.logRepo.FindByTraceID(traceID)
	if err != nil {
		return nil, err
	}

	list := make([]*dto.SimulatedLogInfo, 0, len(logs))
	for _, log := range logs {
		list = append(list, dto.ToSimulatedLogInfo(&log))
	}

	return list, nil
}

// GetLogsByService gets logs by service ID
func (s *SimulatedLogService) GetLogsByService(serviceID uint64, page, pageSize int) (*dto.SimulatedLogListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	logs, total, err := s.logRepo.ListByServiceID(serviceID, page, pageSize)
	if err != nil {
		return nil, err
	}

	list := make([]*dto.SimulatedLogInfo, 0, len(logs))
	for _, log := range logs {
		list = append(list, dto.ToSimulatedLogInfo(&log))
	}

	return &dto.SimulatedLogListResponse{
		Total: total,
		List:  list,
	}, nil
}

// GetLogsByIssue gets logs linked to an issue
func (s *SimulatedLogService) GetLogsByIssue(issueID uint64) ([]*dto.SimulatedLogInfo, error) {
	logs, err := s.logRepo.ListByIssueID(issueID)
	if err != nil {
		return nil, err
	}

	list := make([]*dto.SimulatedLogInfo, 0, len(logs))
	for _, log := range logs {
		list = append(list, dto.ToSimulatedLogInfo(&log))
	}

	return list, nil
}

// LinkIssue links a log to an issue
func (s *SimulatedLogService) LinkIssue(logID uint64, issueID *uint64) error {
	// Validate log exists
	_, err := s.logRepo.FindByID(logID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("日志不存在")
		}
		return err
	}

	// Validate issue exists if issueID is provided
	if issueID != nil {
		_, err = s.issueRepo.FindByID(*issueID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("问题单不存在")
			}
			return err
		}
	}

	return s.logRepo.UpdateIssueID(logID, issueID)
}

// DeleteLog deletes a log
func (s *SimulatedLogService) DeleteLog(id uint64) error {
	_, err := s.logRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("日志不存在")
		}
		return err
	}

	return s.logRepo.Delete(id)
}

// CheckUserProjectAccess checks if user has access to a project
func (s *SimulatedLogService) CheckUserProjectAccess(userID, projectID uint64) (bool, error) {
	return s.projectMemberRepo.ExistsByProjectAndUser(projectID, userID)
}

// getLogInfo helper to get log info
func (s *SimulatedLogService) getLogInfo(id uint64) (*dto.SimulatedLogInfo, error) {
	log, err := s.logRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("日志不存在")
		}
		return nil, err
	}
	return dto.ToSimulatedLogInfo(log), nil
}
