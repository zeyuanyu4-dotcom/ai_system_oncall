package repository

import (
	"ai_system_oncall/internal/model"

	"gorm.io/gorm"
)

type ServiceDependencyRepository struct {
	db *gorm.DB
}

func NewServiceDependencyRepository(db *gorm.DB) *ServiceDependencyRepository {
	return &ServiceDependencyRepository{db: db}
}

// Create creates a new service dependency
func (r *ServiceDependencyRepository) Create(dep *model.ServiceDependency) error {
	return r.db.Create(dep).Error
}

// Delete deletes a service dependency
func (r *ServiceDependencyRepository) Delete(id uint64) error {
	return r.db.Delete(&model.ServiceDependency{}, id).Error
}

// FindByID finds a service dependency by ID
func (r *ServiceDependencyRepository) FindByID(id uint64) (*model.ServiceDependency, error) {
	var dep model.ServiceDependency
	err := r.db.First(&dep, id).Error
	if err != nil {
		return nil, err
	}
	return &dep, nil
}

// Exists checks if dependency exists
func (r *ServiceDependencyRepository) Exists(serviceID, dependsOnServiceID uint64) (bool, error) {
	var count int64
	err := r.db.Model(&model.ServiceDependency{}).Where("service_id = ? AND depends_on_service_id = ?", serviceID, dependsOnServiceID).Count(&count).Error
	return count > 0, err
}

// ListByServiceID lists dependencies of a service
func (r *ServiceDependencyRepository) ListByServiceID(serviceID uint64) ([]model.ServiceDependency, error) {
	var deps []model.ServiceDependency
	err := r.db.Where("service_id = ?", serviceID).Preload("DependsOnService").Order("created_at DESC").Find(&deps).Error
	return deps, err
}

// ListDependents lists services that depend on a service
func (r *ServiceDependencyRepository) ListDependents(serviceID uint64) ([]model.ServiceDependency, error) {
	var deps []model.ServiceDependency
	err := r.db.Where("depends_on_service_id = ?", serviceID).Preload("Service").Order("created_at DESC").Find(&deps).Error
	return deps, err
}
