package service

import (
	"errors"

	"ai_system_oncall/internal/constant"
	"ai_system_oncall/internal/dto"
	"ai_system_oncall/internal/model"
	"ai_system_oncall/internal/repository"

	"gorm.io/gorm"
)

type ServiceAPIService struct {
	apiRepo     *repository.ServiceAPIRepository
	serviceRepo *repository.ServiceRepository
}

func NewServiceAPIService(apiRepo *repository.ServiceAPIRepository, serviceRepo *repository.ServiceRepository) *ServiceAPIService {
	return &ServiceAPIService{
		apiRepo:     apiRepo,
		serviceRepo: serviceRepo,
	}
}

// CreateAPI creates a new service API
func (s *ServiceAPIService) CreateAPI(serviceID uint64, req *dto.CreateServiceAPIRequest) (*dto.ServiceAPIInfo, error) {
	// Check if service exists
	_, err := s.serviceRepo.FindByID(serviceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("服务不存在")
		}
		return nil, err
	}

	// Check if API already exists
	exists, err := s.apiRepo.ExistsByServiceMethodPath(serviceID, req.Method, req.Path)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("接口已存在")
	}

	// Create API
	api := &model.ServiceAPI{
		ServiceID:   serviceID,
		Method:      req.Method,
		Path:        req.Path,
		Name:        req.Name,
		Description: req.Description,
		Status:      1,
	}

	if err := s.apiRepo.Create(api); err != nil {
		return nil, err
	}

	return dto.ToServiceAPIInfo(api), nil
}

// GetAPI gets an API by ID
func (s *ServiceAPIService) GetAPI(id uint64) (*dto.ServiceAPIInfo, error) {
	api, err := s.apiRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("接口不存在")
		}
		return nil, err
	}
	return dto.ToServiceAPIInfo(api), nil
}

// ListAPIs lists APIs of a service
func (s *ServiceAPIService) ListAPIs(serviceID uint64) ([]*dto.ServiceAPIInfo, error) {
	apis, err := s.apiRepo.ListByServiceID(serviceID)
	if err != nil {
		return nil, err
	}

	list := make([]*dto.ServiceAPIInfo, 0, len(apis))
	for _, api := range apis {
		list = append(list, dto.ToServiceAPIInfo(&api))
	}

	return list, nil
}

// UpdateAPI updates an API
func (s *ServiceAPIService) UpdateAPI(id uint64, req *dto.UpdateServiceAPIRequest) (*dto.ServiceAPIInfo, error) {
	api, err := s.apiRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("接口不存在")
		}
		return nil, err
	}

	if req.Method != "" {
		// Check if new method+path combination exists
		if req.Path != "" {
			exists, err := s.apiRepo.ExistsByServiceMethodPath(api.ServiceID, req.Method, req.Path)
			if err != nil {
				return nil, err
			}
			if exists {
				return nil, errors.New("接口已存在")
			}
		}
		api.Method = req.Method
	}
	if req.Path != "" {
		api.Path = req.Path
	}
	if req.Name != "" {
		api.Name = req.Name
	}
	if req.Description != "" {
		api.Description = req.Description
	}
	if req.Status != nil {
		api.Status = *req.Status
	}

	if err := s.apiRepo.Update(api); err != nil {
		return nil, err
	}

	return dto.ToServiceAPIInfo(api), nil
}

// DeleteAPI deletes an API
func (s *ServiceAPIService) DeleteAPI(id uint64) error {
	_, err := s.apiRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("接口不存在")
		}
		return err
	}
	return s.apiRepo.Delete(id)
}

type ServiceDependencyService struct {
	depRepo     *repository.ServiceDependencyRepository
	serviceRepo *repository.ServiceRepository
}

func NewServiceDependencyService(depRepo *repository.ServiceDependencyRepository, serviceRepo *repository.ServiceRepository) *ServiceDependencyService {
	return &ServiceDependencyService{
		depRepo:     depRepo,
		serviceRepo: serviceRepo,
	}
}

// CreateDependency creates a new service dependency
func (s *ServiceDependencyService) CreateDependency(serviceID uint64, req *dto.CreateServiceDependencyRequest) (*dto.ServiceDependencyInfo, error) {
	// Check if service exists
	_, err := s.serviceRepo.FindByID(serviceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("服务不存在")
		}
		return nil, err
	}

	// Check if depends_on service exists
	_, err = s.serviceRepo.FindByID(req.DependsOnServiceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("依赖的服务不存在")
		}
		return nil, err
	}

	// Cannot depend on self
	if serviceID == req.DependsOnServiceID {
		return nil, errors.New("服务不能依赖自身")
	}

	// Check if dependency already exists
	exists, err := s.depRepo.Exists(serviceID, req.DependsOnServiceID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("依赖关系已存在")
	}

	// Set default dependency type
	depType := req.DependencyType
	if depType == "" {
		depType = constant.DependencyTypeHttp
	}

	// Create dependency
	dep := &model.ServiceDependency{
		ServiceID:          serviceID,
		DependsOnServiceID: req.DependsOnServiceID,
		DependencyType:     depType,
		Description:        req.Description,
	}

	if err := s.depRepo.Create(dep); err != nil {
		return nil, err
	}

	// Reload with relations
	dep, _ = s.depRepo.FindByID(dep.ID)
	return dto.ToServiceDependencyInfo(dep), nil
}

// ListDependencies lists dependencies of a service
func (s *ServiceDependencyService) ListDependencies(serviceID uint64) ([]*dto.ServiceDependencyInfo, error) {
	deps, err := s.depRepo.ListByServiceID(serviceID)
	if err != nil {
		return nil, err
	}

	list := make([]*dto.ServiceDependencyInfo, 0, len(deps))
	for _, dep := range deps {
		list = append(list, dto.ToServiceDependencyInfo(&dep))
	}

	return list, nil
}

// DeleteDependency deletes a dependency
func (s *ServiceDependencyService) DeleteDependency(id uint64) error {
	_, err := s.depRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("依赖关系不存在")
		}
		return err
	}
	return s.depRepo.Delete(id)
}
