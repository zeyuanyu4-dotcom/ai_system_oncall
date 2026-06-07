package repository

import (
	"context"

	"ai_system_oncall/internal/cache"
	"ai_system_oncall/internal/model"

	"gorm.io/gorm"
)

type ServiceRepository struct {
	db  *gorm.DB
	sf  *cache.SingleflightCache
}

func NewServiceRepository(db *gorm.DB) *ServiceRepository {
	return &ServiceRepository{
		db:  db,
		sf:  cache.GetSingleflightCache(),
	}
}

// Create creates a new service
func (r *ServiceRepository) Create(service *model.Service) error {
	if err := r.db.Create(service).Error; err != nil {
		return err
	}
	// 失效服务列表缓存
	ctx := context.Background()
	r.sf.InvalidateByPattern(ctx, "service:list:*")
	if service.ProjectID > 0 {
		r.sf.Invalidate(ctx, cache.KeyProjectServices, service.ProjectID)
	}
	return nil
}

// Update updates a service
func (r *ServiceRepository) Update(service *model.Service) error {
	if err := r.db.Save(service).Error; err != nil {
		return err
	}
	// 失效服务详情和列表缓存
	ctx := context.Background()
	r.sf.Invalidate(ctx, cache.KeyServiceDetail, service.ID)
	r.sf.InvalidateByPattern(ctx, "service:list:*")
	if service.ProjectID > 0 {
		r.sf.Invalidate(ctx, cache.KeyProjectServices, service.ProjectID)
	}
	return nil
}

// Delete soft deletes a service
func (r *ServiceRepository) Delete(id uint64) error {
	// 先获取服务信息用于失效缓存
	var service model.Service
	if err := r.db.First(&service, id).Error; err != nil {
		return err
	}

	if err := r.db.Delete(&model.Service{}, id).Error; err != nil {
		return err
	}

	// 失效所有相关缓存
	ctx := context.Background()
	r.sf.Invalidate(ctx, cache.KeyServiceDetail, id)
	r.sf.InvalidateByPattern(ctx, "service:list:*")
	if service.ProjectID > 0 {
		r.sf.Invalidate(ctx, cache.KeyProjectServices, service.ProjectID)
	}
	return nil
}

// FindByID finds a service by ID
func (r *ServiceRepository) FindByID(id uint64) (*model.Service, error) {
	ctx := context.Background()
	var service model.Service

	err := r.sf.GetWithLoad(ctx, cache.KeyServiceDetail, &service, []interface{}{id}, func() (interface{}, error) {
		var s model.Service
		if err := r.db.Preload("Owner").Preload("Project").First(&s, id).Error; err != nil {
			return nil, err
		}
		return &s, nil
	})

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
	// 先获取服务信息用于失效缓存
	var service model.Service
	if err := r.db.First(&service, id).Error; err != nil {
		// 如果找不到服务，只更新状态
		return r.db.Model(&model.Service{}).Where("id = ?", id).Update("status", status).Error
	}

	if err := r.db.Model(&model.Service{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return err
	}

	// 失效服务详情和列表缓存
	ctx := context.Background()
	r.sf.Invalidate(ctx, cache.KeyServiceDetail, id)
	r.sf.InvalidateByPattern(ctx, "service:list:*")
	if service.ProjectID > 0 {
		r.sf.Invalidate(ctx, cache.KeyProjectServices, service.ProjectID)
	}
	return nil
}
