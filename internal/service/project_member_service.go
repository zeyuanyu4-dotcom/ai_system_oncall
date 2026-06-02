package service

import (
	"errors"

	"ai_system_oncall/internal/constant"
	"ai_system_oncall/internal/dto"
	"ai_system_oncall/internal/model"
	"ai_system_oncall/internal/repository"

	"gorm.io/gorm"
)

type ProjectMemberService struct {
	projectMemberRepo *repository.ProjectMemberRepository
	projectRepo       *repository.ProjectRepository
	userRepo          *repository.UserRepository
}

func NewProjectMemberService(projectMemberRepo *repository.ProjectMemberRepository, projectRepo *repository.ProjectRepository, userRepo *repository.UserRepository) *ProjectMemberService {
	return &ProjectMemberService{
		projectMemberRepo: projectMemberRepo,
		projectRepo:       projectRepo,
		userRepo:          userRepo,
	}
}

// AddMember adds a member to a project
func (s *ProjectMemberService) AddMember(projectID, operatorID uint64, req *dto.AddMemberRequest) error {
	// Check if project exists
	_, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("项目不存在")
		}
		return err
	}

	// Check if user exists
	user, err := s.userRepo.FindByID(req.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return err
	}

	// Check if user is already a member
	exists, err := s.projectMemberRepo.ExistsByProjectAndUser(projectID, req.UserID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("用户已是项目成员")
	}

	// Validate project role
	validRoles := map[string]bool{
		constant.ProjectRoleMember:   true,
		constant.ProjectRoleDeveloper: true,
		constant.ProjectRoleTester:    true,
		constant.ProjectRoleAdmin:     true,
		constant.ProjectRoleOwner:     true,
	}
	if !validRoles[req.ProjectRole] {
		return errors.New("无效的项目角色")
	}

	// Create member
	member := &model.ProjectMember{
		ProjectID:   projectID,
		UserID:      req.UserID,
		ProjectRole: req.ProjectRole,
	}

	_ = user // avoid unused variable warning
	return s.projectMemberRepo.Create(member)
}

// ListMembers lists members of a project
func (s *ProjectMemberService) ListMembers(projectID uint64) (*dto.MemberListResponse, error) {
	members, err := s.projectMemberRepo.ListByProjectID(projectID)
	if err != nil {
		return nil, err
	}

	list := make([]*dto.MemberInfo, 0, len(members))
	for _, member := range members {
		list = append(list, dto.ToMemberInfo(&member))
	}

	return &dto.MemberListResponse{
		Total: int64(len(list)),
		List:  list,
	}, nil
}

// UpdateMemberRole updates a member's role
func (s *ProjectMemberService) UpdateMemberRole(projectID, userID uint64, req *dto.UpdateMemberRoleRequest) error {
	// Validate project role
	validRoles := map[string]bool{
		constant.ProjectRoleMember:   true,
		constant.ProjectRoleDeveloper: true,
		constant.ProjectRoleTester:    true,
		constant.ProjectRoleAdmin:     true,
		constant.ProjectRoleOwner:     true,
	}
	if !validRoles[req.ProjectRole] {
		return errors.New("无效的项目角色")
	}

	member, err := s.projectMemberRepo.FindByProjectAndUser(projectID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("成员不存在")
		}
		return err
	}

	member.ProjectRole = req.ProjectRole
	return s.projectMemberRepo.Update(member)
}

// RemoveMember removes a member from a project
func (s *ProjectMemberService) RemoveMember(projectID, userID uint64) error {
	// Check if member exists
	member, err := s.projectMemberRepo.FindByProjectAndUser(projectID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("成员不存在")
		}
		return err
	}

	// Cannot remove owner
	if member.ProjectRole == constant.ProjectRoleOwner {
		return errors.New("不能移除项目所有者")
	}

	return s.projectMemberRepo.DeleteByProjectAndUser(projectID, userID)
}
