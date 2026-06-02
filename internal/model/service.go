package model

import (
	"time"

	"gorm.io/gorm"
)

// Service 服务模型
type Service struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID    uint64         `gorm:"not null;uniqueIndex:idx_project_code;index" json:"project_id"`
	Name         string         `gorm:"type:varchar(128);not null" json:"name"`
	Code         string         `gorm:"type:varchar(64);not null;uniqueIndex:idx_project_code" json:"code"`
	Description  string         `gorm:"type:text" json:"description"`
	ServiceType  string         `gorm:"type:varchar(32);not null;default:backend" json:"service_type"`
	OwnerID      uint64         `gorm:"index" json:"owner_id"`
	Language     string         `gorm:"type:varchar(32)" json:"language"`
	RepoURL      string         `gorm:"type:varchar(255)" json:"repo_url"`
	DeployEnv    string         `gorm:"type:varchar(64)" json:"deploy_env"`
	Status       int8           `gorm:"type:tinyint;not null;default:1" json:"status"` // 1:正常, 0:下线
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Project      *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Owner        *User    `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	APIs         []ServiceAPI `gorm:"foreignKey:ServiceID" json:"apis,omitempty"`
	Dependencies []ServiceDependency `gorm:"foreignKey:ServiceID" json:"dependencies,omitempty"`
}

// TableName specifies the table name for Service model
func (Service) TableName() string {
	return "services"
}

// IsEnabled checks if service is enabled
func (s *Service) IsEnabled() bool {
	return s.Status == 1
}

// ServiceAPI 服务接口模型
type ServiceAPI struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ServiceID   uint64    `gorm:"not null;uniqueIndex:idx_service_method_path;index" json:"service_id"`
	Method      string    `gorm:"type:varchar(16);not null;uniqueIndex:idx_service_method_path" json:"method"`
	Path        string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_service_method_path" json:"path"`
	Name        string    `gorm:"type:varchar(128)" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Status      int8      `gorm:"type:tinyint;not null;default:1" json:"status"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Service *Service `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
}

// TableName specifies the table name for ServiceAPI model
func (ServiceAPI) TableName() string {
	return "service_apis"
}

// ServiceDependency 服务依赖模型
type ServiceDependency struct {
	ID                uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ServiceID         uint64    `gorm:"not null;index" json:"service_id"`
	DependsOnServiceID uint64   `gorm:"not null;index" json:"depends_on_service_id"`
	DependencyType    string    `gorm:"type:varchar(32);not null;default:http" json:"dependency_type"`
	Description       string    `gorm:"type:text" json:"description"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Service         *Service `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	DependsOnService *Service `gorm:"foreignKey:DependsOnServiceID" json:"depends_on_service,omitempty"`
}

// TableName specifies the table name for ServiceDependency model
func (ServiceDependency) TableName() string {
	return "service_dependencies"
}
