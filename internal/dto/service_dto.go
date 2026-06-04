package dto

import (
	"time"

	"ai_system_oncall/internal/model"
)

// CreateServiceRequest 创建服务请求
type CreateServiceRequest struct {
	ProjectID   uint64  `json:"project_id"`
	Name        string  `json:"name" binding:"required,min=2,max=128"`
	Code        string  `json:"code" binding:"required,min=2,max=64"`
	Description string  `json:"description"`
	ServiceType string  `json:"service_type"`
	OwnerID     *uint64 `json:"owner_id"`
	Language    string  `json:"language"`
	RepoURL     string  `json:"repo_url"`
	DeployEnv   string  `json:"deploy_env"`
}

// UpdateServiceRequest 更新服务请求
type UpdateServiceRequest struct {
	Name        string  `json:"name" binding:"omitempty,min=2,max=128"`
	Description string  `json:"description"`
	ServiceType string  `json:"service_type"`
	OwnerID     *uint64 `json:"owner_id"`
	Language    string  `json:"language"`
	RepoURL     string  `json:"repo_url"`
	DeployEnv   string  `json:"deploy_env"`
	Status      *int8   `json:"status"`
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	ID          uint64    `json:"id"`
	ProjectID   uint64    `json:"project_id"`
	ProjectName string    `json:"project_name,omitempty"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	ServiceType string    `json:"service_type"`
	OwnerID     *uint64   `json:"owner_id"`
	OwnerName   string    `json:"owner_name,omitempty"`
	Language    string    `json:"language"`
	RepoURL     string    `json:"repo_url"`
	DeployEnv   string    `json:"deploy_env"`
	Status      int8      `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ServiceListRequest 服务列表请求
type ServiceListRequest struct {
	Page        int    `form:"page" binding:"omitempty,min=1"`
	PageSize    int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	ProjectID   uint64 `form:"project_id"`
	ServiceType string `form:"service_type"`
	Keyword     string `form:"keyword"`
	Status      *int8  `form:"status"`
}

// ServiceListResponse 服务列表响应
type ServiceListResponse struct {
	Total int64          `json:"total"`
	List  []*ServiceInfo `json:"list"`
}

// ToServiceInfo converts Service model to ServiceInfo
func ToServiceInfo(service *model.Service) *ServiceInfo {
	if service == nil {
		return nil
	}
	info := &ServiceInfo{
		ID:          service.ID,
		ProjectID:   service.ProjectID,
		Name:        service.Name,
		Code:        service.Code,
		Description: service.Description,
		ServiceType: service.ServiceType,
		OwnerID:     service.OwnerID,
		Language:    service.Language,
		RepoURL:     service.RepoURL,
		DeployEnv:   service.DeployEnv,
		Status:      service.Status,
		CreatedAt:   service.CreatedAt,
		UpdatedAt:   service.UpdatedAt,
	}
	if service.Owner != nil {
		info.OwnerName = service.Owner.Username
	}
	if service.Project != nil {
		info.ProjectName = service.Project.Name
	}
	return info
}

// CreateServiceAPIRequest 创建服务接口请求
type CreateServiceAPIRequest struct {
	Method      string `json:"method" binding:"required"`
	Path        string `json:"path" binding:"required"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateServiceAPIRequest 更新服务接口请求
type UpdateServiceAPIRequest struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      *int8  `json:"status"`
}

// ServiceAPIInfo 服务接口信息
type ServiceAPIInfo struct {
	ID          uint64    `json:"id"`
	ServiceID   uint64    `json:"service_id"`
	Method      string    `json:"method"`
	Path        string    `json:"path"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      int8      `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToServiceAPIInfo converts ServiceAPI to ServiceAPIInfo
func ToServiceAPIInfo(api *model.ServiceAPI) *ServiceAPIInfo {
	if api == nil {
		return nil
	}
	return &ServiceAPIInfo{
		ID:          api.ID,
		ServiceID:   api.ServiceID,
		Method:      api.Method,
		Path:        api.Path,
		Name:        api.Name,
		Description: api.Description,
		Status:      api.Status,
		CreatedAt:   api.CreatedAt,
		UpdatedAt:   api.UpdatedAt,
	}
}

// CreateServiceDependencyRequest 创建服务依赖请求
type CreateServiceDependencyRequest struct {
	DependsOnServiceID uint64 `json:"depends_on_service_id" binding:"required"`
	DependencyType     string `json:"dependency_type"`
	Description        string `json:"description"`
}

// ServiceDependencyInfo 服务依赖信息
type ServiceDependencyInfo struct {
	ID                 uint64    `json:"id"`
	ServiceID          uint64    `json:"service_id"`
	DependsOnServiceID uint64    `json:"depends_on_service_id"`
	DependencyName     string    `json:"dependency_name,omitempty"`
	DependencyType     string    `json:"dependency_type"`
	Description        string    `json:"description"`
	CreatedAt          time.Time `json:"created_at"`
}

// ToServiceDependencyInfo converts ServiceDependency to ServiceDependencyInfo
func ToServiceDependencyInfo(dep *model.ServiceDependency) *ServiceDependencyInfo {
	if dep == nil {
		return nil
	}
	info := &ServiceDependencyInfo{
		ID:                 dep.ID,
		ServiceID:          dep.ServiceID,
		DependsOnServiceID: dep.DependsOnServiceID,
		DependencyType:     dep.DependencyType,
		Description:        dep.Description,
		CreatedAt:          dep.CreatedAt,
	}
	if dep.DependsOnService != nil {
		info.DependencyName = dep.DependsOnService.Name
	}
	return info
}
