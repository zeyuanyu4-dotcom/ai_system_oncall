package repository

import (
	"ai_system_oncall/internal/model"

	"gorm.io/gorm"
)

type ProjectMemberRepository struct {
	db *gorm.DB
}

func NewProjectMemberRepository(db *gorm.DB) *ProjectMemberRepository {
	return &ProjectMemberRepository{db: db}
}

// Create creates a new project member
func (r *ProjectMemberRepository) Create(member *model.ProjectMember) error {
	return r.db.Create(member).Error
}

// Update updates a project member
func (r *ProjectMemberRepository) Update(member *model.ProjectMember) error {
	return r.db.Save(member).Error
}

// Delete deletes a project member
func (r *ProjectMemberRepository) Delete(id uint64) error {
	return r.db.Delete(&model.ProjectMember{}, id).Error
}

// DeleteByProjectAndUser deletes a project member by project ID and user ID
func (r *ProjectMemberRepository) DeleteByProjectAndUser(projectID, userID uint64) error {
	return r.db.Where("project_id = ? AND user_id = ?", projectID, userID).Delete(&model.ProjectMember{}).Error
}

// FindByID finds a project member by ID
func (r *ProjectMemberRepository) FindByID(id uint64) (*model.ProjectMember, error) {
	var member model.ProjectMember
	err := r.db.First(&member, id).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// FindByProjectAndUser finds a project member by project ID and user ID
func (r *ProjectMemberRepository) FindByProjectAndUser(projectID, userID uint64) (*model.ProjectMember, error) {
	var member model.ProjectMember
	err := r.db.Where("project_id = ? AND user_id = ?", projectID, userID).First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// ExistsByProjectAndUser checks if user is in project
func (r *ProjectMemberRepository) ExistsByProjectAndUser(projectID, userID uint64) (bool, error) {
	var count int64
	err := r.db.Model(&model.ProjectMember{}).Where("project_id = ? AND user_id = ?", projectID, userID).Count(&count).Error
	return count > 0, err
}

// ListByProjectID lists members of a project
func (r *ProjectMemberRepository) ListByProjectID(projectID uint64) ([]model.ProjectMember, error) {
	var members []model.ProjectMember
	err := r.db.Where("project_id = ?", projectID).Preload("User").Order("joined_at ASC").Find(&members).Error
	return members, err
}

// ListByUserID lists projects a user belongs to
func (r *ProjectMemberRepository) ListByUserID(userID uint64) ([]model.ProjectMember, error) {
	var members []model.ProjectMember
	err := r.db.Where("user_id = ?", userID).Preload("Project").Order("joined_at ASC").Find(&members).Error
	return members, err
}

// GetMemberRole gets a user's role in a project
func (r *ProjectMemberRepository) GetMemberRole(projectID, userID uint64) (string, error) {
	var member model.ProjectMember
	err := r.db.Where("project_id = ? AND user_id = ?", projectID, userID).First(&member).Error
	if err != nil {
		return "", err
	}
	return member.ProjectRole, nil
}

// IsProjectAdmin checks if user is project admin
func (r *ProjectMemberRepository) IsProjectAdmin(projectID, userID uint64) (bool, error) {
	member, err := r.FindByProjectAndUser(projectID, userID)
	if err != nil {
		return false, err
	}
	return member.IsAdmin(), nil
}
