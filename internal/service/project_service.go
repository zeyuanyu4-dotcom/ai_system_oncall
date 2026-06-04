package service

import (
	"errors"

	"ai_system_oncall/internal/constant"
	"ai_system_oncall/internal/dto"
	"ai_system_oncall/internal/model"
	"ai_system_oncall/internal/repository"

	"gorm.io/gorm"
)

type ProjectService struct {
	projectRepo       *repository.ProjectRepository
	projectMemberRepo *repository.ProjectMemberRepository
	userRepo          *repository.UserRepository
}

func NewProjectService(projectRepo *repository.ProjectRepository, projectMemberRepo *repository.ProjectMemberRepository, userRepo *repository.UserRepository) *ProjectService {
	return &ProjectService{
		projectRepo:       projectRepo,
		projectMemberRepo: projectMemberRepo,
		userRepo:          userRepo,
	}
}

// CreateProject creates a new project
func (s *ProjectService) CreateProject(creatorID uint64, req *dto.CreateProjectRequest) (*dto.ProjectInfo, error) {
	// Check if code exists
	exists, err := s.projectRepo.ExistsByCode(req.Code)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("项目标识已被使用")
	}

	// Create project
	project := &model.Project{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		OwnerID:     creatorID,
		Status:      1,
		CreatedBy:   creatorID,
	}

	if err := s.projectRepo.Create(project); err != nil {
		return nil, err
	}

	// Add creator as owner member
	member := &model.ProjectMember{
		ProjectID:   project.ID,
		UserID:      creatorID,
		ProjectRole: constant.ProjectRoleOwner,
	}
	_ = s.projectMemberRepo.Create(member)

	// Add additional members from request
	for _, m := range req.Members {
		// Validate user exists and has developer or tester role
		user, err := s.userRepo.FindByID(m.UserID)
		if err != nil {
			continue // Skip invalid users
		}
		// Only allow adding developer or tester role users
		if user.Role != constant.RoleDeveloper && user.Role != constant.RoleTester {
			continue
		}

		// Check if already a member
		exists, _ := s.projectMemberRepo.ExistsByProjectAndUser(project.ID, m.UserID)
		if exists {
			continue
		}

		member := &model.ProjectMember{
			ProjectID:   project.ID,
			UserID:      m.UserID,
			ProjectRole: m.ProjectRole,
		}
		_ = s.projectMemberRepo.Create(member)
	}

	return dto.ToProjectInfo(project), nil
}

// GetProject gets a project by ID
func (s *ProjectService) GetProject(id uint64) (*dto.ProjectInfo, error) {
	project, err := s.projectRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("项目不存在")
		}
		return nil, err
	}
	return dto.ToProjectInfo(project), nil
}

// ListProjects lists projects
func (s *ProjectService) ListProjects(req *dto.ProjectListRequest) (*dto.ProjectListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	projects, total, err := s.projectRepo.List(req.Page, req.PageSize, req.Keyword, req.Status)
	if err != nil {
		return nil, err
	}

	list := make([]*dto.ProjectInfo, 0, len(projects))
	for _, project := range projects {
		list = append(list, dto.ToProjectInfo(&project))
	}

	return &dto.ProjectListResponse{
		Total: total,
		List:  list,
	}, nil
}

// UpdateProject updates a project
func (s *ProjectService) UpdateProject(id uint64, req *dto.UpdateProjectRequest) (*dto.ProjectInfo, error) {
	project, err := s.projectRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("项目不存在")
		}
		return nil, err
	}

	if req.Name != "" {
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = req.Description
	}
	if req.Status != nil {
		project.Status = *req.Status
	}

	if err := s.projectRepo.Update(project); err != nil {
		return nil, err
	}

	return dto.ToProjectInfo(project), nil
}

// DeleteProject deletes a project
func (s *ProjectService) DeleteProject(id uint64) error {
	_, err := s.projectRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("项目不存在")
		}
		return err
	}
	return s.projectRepo.Delete(id)
}

// CheckUserProjectAccess checks if user has access to a project
func (s *ProjectService) CheckUserProjectAccess(userID, projectID uint64) (bool, error) {
	return s.projectMemberRepo.ExistsByProjectAndUser(projectID, userID)
}

// GetUserProjectRole gets user's role in a project
func (s *ProjectService) GetUserProjectRole(userID, projectID uint64) (string, error) {
	return s.projectMemberRepo.GetMemberRole(projectID, userID)
}
