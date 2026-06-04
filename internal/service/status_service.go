package service

import (
	"errors"
	"time"

	"ai_system_oncall/internal/constant"
	"ai_system_oncall/internal/dto"
	"ai_system_oncall/internal/model"
	"ai_system_oncall/internal/repository"

	"gorm.io/gorm"
)

// StatusMachine defines valid status transitions
var statusTransitions = map[string][]string{
	constant.StatusPendingAnalysis: {
		constant.StatusPendingAssignment,
		constant.StatusProcessing,
	},
	constant.StatusPendingAssignment: {
		constant.StatusProcessing,
	},
	constant.StatusProcessing: {
		constant.StatusPendingConfirmation,
	},
	constant.StatusPendingConfirmation: {
		constant.StatusResolved,
		constant.StatusReopened,
	},
	constant.StatusResolved: {
		constant.StatusClosed,
		constant.StatusReopened,
	},
	constant.StatusClosed: {
		constant.StatusReopened,
	},
	constant.StatusReopened: {
		constant.StatusProcessing,
	},
}

type StatusService struct {
	issueRepo         *repository.IssueRepository
	projectMemberRepo *repository.ProjectMemberRepository
	statusLogRepo     *repository.StatusLogRepository
	operationLogRepo  *repository.OperationLogRepository
	userRepo          *repository.UserRepository
}

func NewStatusService(
	issueRepo *repository.IssueRepository,
	projectMemberRepo *repository.ProjectMemberRepository,
	statusLogRepo *repository.StatusLogRepository,
	operationLogRepo *repository.OperationLogRepository,
	userRepo *repository.UserRepository,
) *StatusService {
	return &StatusService{
		issueRepo:         issueRepo,
		projectMemberRepo: projectMemberRepo,
		statusLogRepo:     statusLogRepo,
		operationLogRepo:  operationLogRepo,
		userRepo:          userRepo,
	}
}

// isValidTransition checks if status transition is valid
func (s *StatusService) isValidTransition(fromStatus, toStatus string) bool {
	if fromStatus == "" {
		return true // Initial status
	}

	allowedStatuses, ok := statusTransitions[fromStatus]
	if !ok {
		return false
	}

	for _, status := range allowedStatuses {
		if status == toStatus {
			return true
		}
	}
	return false
}

// canUserChangeStatus checks if user has permission to change status
func (s *StatusService) canUserChangeStatus(issue *model.Issue, userID uint64, globalRole, toStatus string) (bool, error) {
	// System admin can do anything
	if globalRole == constant.RoleSystemAdmin {
		return true, nil
	}

	// Get user's project role (empty string if not a member)
	projectRole, _ := s.projectMemberRepo.GetMemberRole(issue.ProjectID, userID)

	// If user is a project member (any role), they can change any status
	if projectRole != "" {
		return true, nil
	}

	// Non-project-member: only creator can confirm resolution or reopen
	if issue.CreatorID == userID {
		if toStatus == constant.StatusResolved || toStatus == constant.StatusReopened {
			return true, nil
		}
	}

	return false, nil
}

// AssignIssue assigns an issue to a user
func (s *StatusService) AssignIssue(issueID, operatorID, assigneeID uint64, globalRole string) error {
	issue, err := s.issueRepo.FindByID(issueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("问题不存在")
		}
		return err
	}

	// Check permission: system_admin or project member can assign
	if globalRole != constant.RoleSystemAdmin {
		projectRole, err := s.projectMemberRepo.GetMemberRole(issue.ProjectID, operatorID)
		if err != nil || projectRole == "" {
			return errors.New("无权限分配问题，只有项目成员可以分配")
		}
	}

	// Assignee must be a project member with developer role
	assigneeProjectRole, err := s.projectMemberRepo.GetMemberRole(issue.ProjectID, assigneeID)
	if err != nil || assigneeProjectRole == "" {
		return errors.New("处理人不是该项目成员")
	}

	// Check if assignee has developer role (global role)
	assignee, err := s.userRepo.FindByID(assigneeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("处理人不存在")
		}
		return err
	}
	if assignee.Role != constant.RoleDeveloper {
		return errors.New("只能分配给研发人员")
	}

	// Update assignee
	if err := s.issueRepo.UpdateAssignee(issueID, assigneeID); err != nil {
		return err
	}

	// Update the issue object's assignee_id for later use
	issue.AssigneeID = &assigneeID

	// If status is pending_analysis or pending_assignment, change to processing
	if issue.Status == constant.StatusPendingAnalysis || issue.Status == constant.StatusPendingAssignment {
		issue.Status = constant.StatusProcessing
		_ = s.issueRepo.Update(issue)

		// Create status log
		statusLog := &model.IssueStatusLog{
			IssueID:    issueID,
			FromStatus: issue.Status,
			ToStatus:   constant.StatusProcessing,
			OperatorID: operatorID,
			Reason:     "分配处理人",
		}
		_ = s.statusLogRepo.Create(statusLog)
	}

	// Create operation log
	opLog := &model.IssueOperationLog{
		IssueID:          issueID,
		OperatorID:       operatorID,
		OperationType:    constant.OperationAssignIssue,
		OperationContent: "分配处理人",
	}
	_ = s.operationLogRepo.Create(opLog)

	return nil
}

// ChangeStatus changes issue status
func (s *StatusService) ChangeStatus(issueID, operatorID uint64, globalRole, toStatus, reason string) error {
	issue, err := s.issueRepo.FindByID(issueID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("问题不存在")
		}
		return err
	}

	// Check if transition is valid
	if !s.isValidTransition(issue.Status, toStatus) {
		return errors.New("状态流转非法")
	}

	// Check permission
	hasPermission, err := s.canUserChangeStatus(issue, operatorID, globalRole, toStatus)
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("无权限修改状态")
	}

	fromStatus := issue.Status
	issue.Status = toStatus

	// Handle special status changes
	now := time.Now()
	switch toStatus {
	case constant.StatusResolved:
		issue.ResolvedAt = &now
	case constant.StatusClosed:
		issue.ClosedAt = &now
	}

	if err := s.issueRepo.Update(issue); err != nil {
		return err
	}

	// Create status log
	statusLog := &model.IssueStatusLog{
		IssueID:    issueID,
		FromStatus: fromStatus,
		ToStatus:   toStatus,
		OperatorID: operatorID,
		Reason:     reason,
	}
	_ = s.statusLogRepo.Create(statusLog)

	// Create operation log
	opLog := &model.IssueOperationLog{
		IssueID:          issueID,
		OperatorID:       operatorID,
		OperationType:    constant.OperationChangeStatus,
		OperationContent: reason,
	}
	_ = s.operationLogRepo.Create(opLog)

	return nil
}

// GetStatusLogs gets status logs of an issue
func (s *StatusService) GetStatusLogs(issueID uint64) ([]*dto.StatusLogInfo, error) {
	logs, err := s.statusLogRepo.ListByIssueID(issueID)
	if err != nil {
		return nil, err
	}

	list := make([]*dto.StatusLogInfo, 0, len(logs))
	for _, log := range logs {
		list = append(list, dto.ToStatusLogInfo(&log))
	}

	return list, nil
}
