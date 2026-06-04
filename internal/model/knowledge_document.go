package model

import (
	"time"
)

// KnowledgeDocument 知识库文档
type KnowledgeDocument struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID     *uint64        `gorm:"index" json:"project_id,omitempty"`
	ServiceID     *uint64        `gorm:"index" json:"service_id,omitempty"`
	Title         string         `gorm:"type:varchar(255);not null" json:"title"`
	Content       string         `gorm:"type:text" json:"content"`
	DocType       string         `gorm:"type:varchar(50);index;not null" json:"doc_type"`
	CreatorID     uint64         `gorm:"not null" json:"creator_id"`
	UpdaterID     uint64         `gorm:"not null" json:"updater_id"`
	VectorStatus  string         `gorm:"type:varchar(20);index;default:pending" json:"vector_status"`
	CurrentVersion int           `gorm:"default:1" json:"current_version"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     *time.Time     `gorm:"index" json:"-"`

	// 关联关系
	Project     *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Service     *Service `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	Creator     *User    `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	Updater     *User    `gorm:"foreignKey:UpdaterID" json:"updater,omitempty"`
	Versions    []KnowledgeDocVersion `gorm:"foreignKey:DocumentID" json:"versions,omitempty"`
	Attachments []KnowledgeDocAttachment `gorm:"foreignKey:DocumentID" json:"attachments,omitempty"`
}

func (KnowledgeDocument) TableName() string {
	return "knowledge_documents"
}