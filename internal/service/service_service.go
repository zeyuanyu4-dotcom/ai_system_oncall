package service

import (
	"errors"

	"ai_system_oncall/internal/constant"
	"ai_system_oncall/internal/dto"
	"ai_system_oncall/internal/model"
	"ai_system_oncall/internal/repository"

	"gorm.io/gorm"
)

type ServiceService struct {
	serviceRepo *repository.ServiceRepository
	projectRepo *repository.ProjectRepository
}

func NewServiceService(serviceRepo *repository.ServiceRepository, projectRepo *repository.ProjectRepository) *ServiceService {
	return &ServiceService{
		serviceRepo: serviceRepo,
		projectRepo: projectRepo,
	}
}

// CreateService creates a new service
func (s *ServiceService) CreateService(req *dto.CreateServiceRequest) (*dto.ServiceInfo, error) {
	// Check if project exists
	_, err := s.projectRepo.FindByID(req.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("项目不存在")
		}
		return nil, err
	}

	// Check if code exists within project
	exists, err := s.serviceRepo.ExistsByCode(req.ProjectID, req.Code)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("服务标识在该项目中已存在")
	}

	// Set default service type
	serviceType := req.ServiceType
	if serviceType == "" {
		serviceType = constant.ServiceTypeBackend
	}

	// Create service
	service := &model.Service{
		ProjectID:   req.ProjectID,
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		ServiceType: serviceType,
		OwnerID:     req.OwnerID,
		Language:    req.Language,
		RepoURL:     req.RepoURL,
		DeployEnv:   req.DeployEnv,
		Status:      1,
	}

	if err := s.serviceRepo.Create(service); err != nil {
		return nil, err
	}

	return dto.ToServiceInfo(service), nil
}

// GetService gets a service by ID
func (s *ServiceService) GetService(id uint64) (*dto.ServiceInfo, error) {
	service, err := s.serviceRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("服务不存在")
		}
		return nil, err
	}
	return dto.ToServiceInfo(service), nil
}

// ListServices lists services
func (s *ServiceService) ListServices(req *dto.ServiceListRequest) (*dto.ServiceListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	services, total, err := s.serviceRepo.List(req.Page, req.PageSize, req.ProjectID, req.ServiceType, req.Keyword, req.Status)
	if err != nil {
		return nil, err
	}

	list := make([]*dto.ServiceInfo, 0, len(services))
	for _, service := range services {
		list = append(list, dto.ToServiceInfo(&service))
	}

	return &dto.ServiceListResponse{
		Total: total,
		List:  list,
	}, nil
}

// UpdateService updates a service
func (s *ServiceService) UpdateService(id uint64, req *dto.UpdateServiceRequest) (*dto.ServiceInfo, error) {
	service, err := s.serviceRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("服务不存在")
		}
		return nil, err
	}

	if req.Name != "" {
		service.Name = req.Name
	}
	if req.Description != "" {
		service.Description = req.Description
	}
	if req.ServiceType != "" {
		service.ServiceType = req.ServiceType
	}
	if req.OwnerID > 0 {
		service.OwnerID = req.OwnerID
	}
	if req.Language != "" {
		service.Language = req.Language
	}
	if req.RepoURL != "" {
		service.RepoURL = req.RepoURL
	}
	if req.DeployEnv != "" {
		service.DeployEnv = req.DeployEnv
	}
	if req.Status != nil {
		service.Status = *req.Status
	}

	if err := s.serviceRepo.Update(service); err != nil {
		return nil, err
	}

	return dto.ToServiceInfo(service), nil
}

// DeleteService deletes a service
func (s *ServiceService) DeleteService(id uint64) error {
	_, err := s.serviceRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("服务不存在")
		}
		return err
	}
	return s.serviceRepo.Delete(id)
}
