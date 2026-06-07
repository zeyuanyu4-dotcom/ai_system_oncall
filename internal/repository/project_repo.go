package repository

import (
	"context"

	"ai_system_oncall/internal/cache"
	"ai_system_oncall/internal/model"

	"gorm.io/gorm"
)

type ProjectRepository struct {
	db  *gorm.DB
	sf  *cache.SingleflightCache
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{
		db:  db,
		sf:  cache.GetSingleflightCache(),
	}
}

// Create creates a new project
func (r *ProjectRepository) Create(project *model.Project) error {
	if err := r.db.Create(project).Error; err != nil {
		return err
	}
	// 失效项目列表缓存
	ctx := context.Background()
	r.sf.InvalidateByPattern(ctx, "project:list:*")
	r.sf.InvalidateByPattern(ctx, "project:user:*")
	return nil
}

// Update updates a project
func (r *ProjectRepository) Update(project *model.Project) error {
	if err := r.db.Save(project).Error; err != nil {
		return err
	}
	// 失效项目详情和列表缓存
	ctx := context.Background()
	r.sf.Invalidate(ctx, cache.KeyProjectDetail, project.ID)
	r.sf.InvalidateByPattern(ctx, "project:list:*")
	r.sf.InvalidateByPattern(ctx, "project:user:*")
	return nil
}

// Delete soft deletes a project
func (r *ProjectRepository) Delete(id uint64) error {
	if err := r.db.Delete(&model.Project{}, id).Error; err != nil {
		return err
	}
	// 失效所有相关缓存
	ctx := context.Background()
	r.sf.Invalidate(ctx, cache.KeyProjectDetail, id)
	r.sf.InvalidateByPattern(ctx, "project:list:*")
	r.sf.InvalidateByPattern(ctx, "project:user:*")
	// 同时失效该项目下的服务缓存
	r.sf.InvalidateByPattern(ctx, "service:project:*")
	return nil
}

// FindByID finds a project by ID
func (r *ProjectRepository) FindByID(id uint64) (*model.Project, error) {
	ctx := context.Background()
	var project model.Project

	err := r.sf.GetWithLoad(ctx, cache.KeyProjectDetail, &project, []interface{}{id}, func() (interface{}, error) {
		var p model.Project
		if err := r.db.Preload("Owner").Preload("Creator").First(&p, id).Error; err != nil {
			return nil, err
		}
		return &p, nil
	})

	if err != nil {
		return nil, err
	}
	return &project, nil
}

// FindByCode finds a project by code
func (r *ProjectRepository) FindByCode(code string) (*model.Project, error) {
	var project model.Project
	err := r.db.Where("code = ?", code).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// ExistsByCode checks if code exists
func (r *ProjectRepository) ExistsByCode(code string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Project{}).Where("code = ?", code).Count(&count).Error
	return count > 0, err
}

// List lists projects with pagination
func (r *ProjectRepository) List(page, pageSize int, keyword string, status *int8) ([]model.Project, int64, error) {
	var projects []model.Project
	var total int64

	query := r.db.Model(&model.Project{})

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
	if err := query.Preload("Owner").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&projects).Error; err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

// ListByUserID lists projects that a user belongs to
func (r *ProjectRepository) ListByUserID(userID uint64) ([]model.Project, error) {
	var projects []model.Project
	err := r.db.Table("projects").
		Select("projects.*").
		Joins("JOIN project_members ON project_members.project_id = projects.id").
		Where("project_members.user_id = ?", userID).
		Order("projects.created_at DESC").
		Find(&projects).Error
	return projects, err
}

// UpdateStatus updates project status
func (r *ProjectRepository) UpdateStatus(id uint64, status int8) error {
	if err := r.db.Model(&model.Project{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return err
	}
	// 失效项目详情和列表缓存
	ctx := context.Background()
	r.sf.Invalidate(ctx, cache.KeyProjectDetail, id)
	r.sf.InvalidateByPattern(ctx, "project:list:*")
	r.sf.InvalidateByPattern(ctx, "project:user:*")
	return nil
}
