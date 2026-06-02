package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string         `gorm:"type:varchar(64);not null" json:"username"`
	Email        string         `gorm:"type:varchar(128);uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	Phone        string         `gorm:"type:varchar(32)" json:"phone"`
	Role         string         `gorm:"type:varchar(32);not null;default:normal_user" json:"role"`
	Status       int8           `gorm:"type:tinyint;not null;default:1" json:"status"` // 1:正常, 0:禁用
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

// IsEnabled checks if user is enabled
func (u *User) IsEnabled() bool {
	return u.Status == 1
}

// IsSystemAdmin checks if user is system admin
func (u *User) IsSystemAdmin() bool {
	return u.Role == "system_admin"
}
