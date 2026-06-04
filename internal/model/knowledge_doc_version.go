package model

import (
	"time"
)

// KnowledgeDocVersion 文档版本历史
type KnowledgeDocVersion struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	DocumentID    uint64     `gorm:"index;not null" json:"document_id"`
	Version       int        `gorm:"not null" json:"version"`
	Title         string     `gorm:"type:varchar(255);not null" json:"title"`
	Content       string     `gorm:"type:text" json:"content"`
	DocType       string     `gorm:"type:varchar(50);not null" json:"doc_type"`
	EditorID      uint64     `gorm:"not null" json:"editor_id"`
	ChangeSummary string     `gorm:"type:varchar(500)" json:"change_summary"`
	CreatedAt     time.Time  `json:"created_at"`

	// 关联关系
	Document *KnowledgeDocument `gorm:"foreignKey:DocumentID" json:"document,omitempty"`
	Editor   *User              `gorm:"foreignKey:EditorID" json:"editor,omitempty"`
}

func (KnowledgeDocVersion) TableName() string {
	return "knowledge_doc_versions"
}