package repository

import (
	"ai_system_oncall/internal/model"

	"gorm.io/gorm"
)

type ServiceAPIRepository struct {
	db *gorm.DB
}

func NewServiceAPIRepository(db *gorm.DB) *ServiceAPIRepository {
	return &ServiceAPIRepository{db: db}
}

// Create creates a new service API
func (r *ServiceAPIRepository) Create(api *model.ServiceAPI) error {
	return r.db.Create(api).Error
}

// Update updates a service API
func (r *ServiceAPIRepository) Update(api *model.ServiceAPI) error {
	return r.db.Save(api).Error
}

// Delete deletes a service API
func (r *ServiceAPIRepository) Delete(id uint64) error {
	return r.db.Delete(&model.ServiceAPI{}, id).Error
}

// FindByID finds a service API by ID
func (r *ServiceAPIRepository) FindByID(id uint64) (*model.ServiceAPI, error) {
	var api model.ServiceAPI
	err := r.db.First(&api, id).Error
	if err != nil {
		return nil, err
	}
	return &api, nil
}

// FindByServiceMethodPath finds a service API by service ID, method, and path
func (r *ServiceAPIRepository) FindByServiceMethodPath(serviceID uint64, method, path string) (*model.ServiceAPI, error) {
	var api model.ServiceAPI
	err := r.db.Where("service_id = ? AND method = ? AND path = ?", serviceID, method, path).First(&api).Error
	if err != nil {
		return nil, err
	}
	return &api, nil
}

// ExistsByServiceMethodPath checks if API exists
func (r *ServiceAPIRepository) ExistsByServiceMethodPath(serviceID uint64, method, path string) (bool, error) {
	var count int64
	err := r.db.Model(&model.ServiceAPI{}).Where("service_id = ? AND method = ? AND path = ?", serviceID, method, path).Count(&count).Error
	return count > 0, err
}

// ListByServiceID lists APIs of a service
func (r *ServiceAPIRepository) ListByServiceID(serviceID uint64) ([]model.ServiceAPI, error) {
	var apis []model.ServiceAPI
	err := r.db.Where("service_id = ?", serviceID).Order("created_at DESC").Find(&apis).Error
	return apis, err
}
