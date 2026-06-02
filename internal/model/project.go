package model

import (
	"time"

	"gorm.io/gorm"
)

// Project 项目模型
type Project struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string         `gorm:"type:varchar(128);not null" json:"name"`
	Code        string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"code"`
	Description string         `gorm:"type:text" json:"description"`
	OwnerID     uint64         `gorm:"index" json:"owner_id"`
	Status      int8           `gorm:"type:tinyint;not null;default:1" json:"status"` // 1:正常, 0:停用
	CreatedBy   uint64         `json:"created_by"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Owner   *User  `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Creator *User  `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Members []ProjectMember `gorm:"foreignKey:ProjectID" json:"members,omitempty"`
}

// TableName specifies the table name for Project model
func (Project) TableName() string {
	return "projects"
}

// IsEnabled checks if project is enabled
func (p *Project) IsEnabled() bool {
	return p.Status == 1
}

// ProjectMember 项目成员模型
type ProjectMember struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID   uint64    `gorm:"not null;uniqueIndex:idx_project_user" json:"project_id"`
	UserID      uint64    `gorm:"not null;uniqueIndex:idx_project_user;index" json:"user_id"`
	ProjectRole string    `gorm:"type:varchar(32);not null;default:member" json:"project_role"`
	JoinedAt    time.Time `gorm:"autoCreateTime" json:"joined_at"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Project *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName specifies the table name for ProjectMember model
func (ProjectMember) TableName() string {
	return "project_members"
}

// IsAdmin checks if member is project admin or owner
func (pm *ProjectMember) IsAdmin() bool {
	return pm.ProjectRole == "admin" || pm.ProjectRole == "owner"
}
