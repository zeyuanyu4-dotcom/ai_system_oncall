package repository

import (
	"ai_system_oncall/internal/model"

	"gorm.io/gorm"
)

type ServiceRepository struct {
	db *gorm.DB
}

func NewServiceRepository(db *gorm.DB) *ServiceRepository {
	return &ServiceRepository{db: db}
}

// Create creates a new service
func (r *ServiceRepository) Create(service *model.Service) error {
	return r.db.Create(service).Error
}

// Update updates a service
func (r *ServiceRepository) Update(service *model.Service) error {
	return r.db.Save(service).Error
}

// Delete soft deletes a service
func (r *ServiceRepository) Delete(id uint64) error {
	return r.db.Delete(&model.Service{}, id).Error
}

// FindByID finds a service by ID
func (r *ServiceRepository) FindByID(id uint64) (*model.Service, error) {
	var service model.Service
	err := r.db.Preload("Owner").Preload("Project").First(&service, id).Error
	if err != nil {
		return nil, err
	}
	return &service, nil
}

// FindByCode finds a service by code within a project
func (r *ServiceRepository) FindByCode(projectID uint64, code string) (*model.Service, error) {
	var service model.Service
	err := r.db.Where("project_id = ? AND code = ?", projectID, code).First(&service).Error
	if err != nil {
		return nil, err
	}
	return &service, nil
}

// ExistsByCode checks if code exists within a project
func (r *ServiceRepository) ExistsByCode(projectID uint64, code string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Service{}).Where("project_id = ? AND code = ?", projectID, code).Count(&count).Error
	return count > 0, err
}

// List lists services with pagination
func (r *ServiceRepository) List(page, pageSize int, projectID uint64, serviceType, keyword string, status *int8) ([]model.Service, int64, error) {
	var services []model.Service
	var total int64

	query := r.db.Model(&model.Service{})

	if projectID > 0 {
		query = query.Where("project_id = ?", projectID)
	}
	if serviceType != "" {
		query = query.Where("service_type = ?", serviceType)
	}
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Owner").Preload("Project").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&services).Error; err != nil {
		return nil, 0, err
	}

	return services, total, nil
}

// ListByProjectID lists services of a project
func (r *ServiceRepository) ListByProjectID(projectID uint64) ([]model.Service, error) {
	var services []model.Service
	err := r.db.Where("project_id = ?", projectID).Order("created_at DESC").Find(&services).Error
	return services, err
}

// UpdateStatus updates service status
func (r *ServiceRepository) UpdateStatus(id uint64, status int8) error {
	return r.db.Model(&model.Service{}).Where("id = ?", id).Update("status", status).Error
}
